package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/peasant-labs/schema/internal/release"
)

// newTestGitHubClient returns a githubClient whose go-github *Client is pointed
// at a local httptest server, so the thin wrapper is exercised over the real
// go-github request/decode/pagination machinery without any network access.
func newTestGitHubClient(t *testing.T, handler http.Handler) *githubClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	base := server.URL + "/"
	gh, err := github.NewClient(github.WithURLs(&base, &base))
	if err != nil {
		t.Fatalf("build test go-github client: %v", err)
	}
	return &githubClient{gh: gh}
}

// assertErrContains fails unless err is non-nil and its message contains every
// listed substring.
func assertErrContains(t *testing.T, err error, wants []string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error containing %v, got nil", wants)
	}
	for _, want := range wants {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing actionable substring %q", err.Error(), want)
		}
	}
}

// TestGitHubClient_SeamCases drives the single-response seam cases from
// testdata/github/cases.yaml: happy path, non-2xx error-wrap (404/500),
// empty-result fail-safety, and the [null]-element nil-guards, across all three
// wrapper methods. The canned bytes live in fixtures; the httptest server still
// drives the REAL production githubClient (the SUT is never mocked).
func TestGitHubClient_SeamCases(t *testing.T) {
	t.Parallel()

	cases := loadGithubSeamCases(t)
	if len(cases) != 8 {
		t.Fatalf("github seam fixture has %d cases, want 8 (fixture truncated?)", len(cases))
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			body := readGithubFixture(t, tc.Response)
			var gotPath string
			c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				if tc.Status != 0 {
					w.WriteHeader(tc.Status)
				}
				_, _ = w.Write(body)
			}))

			switch tc.Method {
			case "collaborator_permission":
				perm, err := c.CollaboratorPermission(context.Background(), "peasant-labs/schema", tc.User)
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("CollaboratorPermission: %v", err)
				}
				if perm != release.CollaboratorPermission(tc.WantPermission) {
					t.Fatalf("permission = %q, want %q", perm, tc.WantPermission)
				}
				if tc.WantPath != "" && gotPath != tc.WantPath {
					t.Fatalf("request path = %q, want %q (owner/repo split)", gotPath, tc.WantPath)
				}

			case "workflow_runs":
				runs, err := c.WorkflowRunsForCommit(context.Background(), "peasant-labs/schema", "release.yml", tc.Commit)
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("WorkflowRunsForCommit: %v", err)
				}
				if len(runs) != tc.WantRunCount {
					t.Fatalf("collected %d runs, want %d: %+v", len(runs), tc.WantRunCount, runs)
				}
				if got := release.RunGreenForCommit(runs, tc.Commit); got != tc.WantGreen {
					t.Fatalf("RunGreenForCommit(%s) = %v, want %v", tc.Commit, got, tc.WantGreen)
				}

			case "pull_reviews":
				reviews, err := c.PullReviews(context.Background(), "peasant-labs/schema", tc.PR)
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("PullReviews: %v", err)
				}
				if len(reviews) != tc.WantReviewCount {
					t.Fatalf("collected %d reviews, want %d: %+v", len(reviews), tc.WantReviewCount, reviews)
				}
				approvers := release.LatestApprovers(reviews)
				if len(approvers) != len(tc.WantApprovers) {
					t.Fatalf("LatestApprovers = %v, want %v", approvers, tc.WantApprovers)
				}
				for i := range approvers {
					if approvers[i] != tc.WantApprovers[i] {
						t.Fatalf("LatestApprovers[%d] = %q, want %q", i, approvers[i], tc.WantApprovers[i])
					}
				}

			default:
				t.Fatalf("unknown seam method %q in fixture case %q", tc.Method, tc.Name)
			}
		})
	}
}

// TestGitHubClient_WorkflowRunsForCommit covers the multi-page pagination
// boundary + server-side HeadSHA filter. The per-page bodies live in fixtures;
// the Link-header page routing + page-count assertions stay inline (behavioral).
func TestGitHubClient_WorkflowRunsForCommit(t *testing.T) {
	t.Parallel()

	const commit = "deadbeefcafe"
	// Pre-read outside the handler goroutine (t.Fatalf must run on the test goroutine).
	page1 := readGithubFixture(t, "runs-page1.json") // in_progress + failure decoys
	page2 := readGithubFixture(t, "runs-page2.json") // the green run on the commit

	var gotHeadSHA, gotPath string
	var page1Hits int
	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeadSHA = r.URL.Query().Get("head_sha")
		gotPath = r.URL.Path
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write(page2)
			return
		}
		page1Hits++
		w.Header().Set("Link", fmt.Sprintf("<http://%s%s?page=2>; rel=\"next\"", r.Host, r.URL.Path))
		_, _ = w.Write(page1)
	}))

	runs, err := c.WorkflowRunsForCommit(context.Background(), "peasant-labs/schema", "release.yml", commit)
	if err != nil {
		t.Fatalf("WorkflowRunsForCommit: %v", err)
	}

	// HeadSHA must be sent as a SERVER-SIDE filter (not a client-side scan).
	if gotHeadSHA != commit {
		t.Fatalf("head_sha query param = %q, want %q (server-side filter)", gotHeadSHA, commit)
	}
	if want := "/repos/peasant-labs/schema/actions/workflows/release.yml/runs"; gotPath != want {
		t.Fatalf("request path = %q, want %q", gotPath, want)
	}
	// Pagination: both pages collected → 3 runs in order.
	if len(runs) != 3 {
		t.Fatalf("collected %d runs, want 3 (pagination): %+v", len(runs), runs)
	}
	if page1Hits != 1 {
		t.Fatalf("page 1 fetched %d times, want exactly 1", page1Hits)
	}
	// Decode + typed-enum mapping of the green run on the final page.
	green := runs[2]
	if green.HeadSHA != commit || green.Status != release.WorkflowRunCompleted || green.Conclusion != release.WorkflowRunSuccess {
		t.Fatalf("green run mapped to %+v, want {HeadSHA:%s completed success}", green, commit)
	}
	// The mapped runs feed RunGreenForCommit unchanged.
	if !release.RunGreenForCommit(runs, commit) {
		t.Fatalf("RunGreenForCommit(decoded runs) = false, want true")
	}
}

// TestGitHubClient_PullReviews_PaginatesInOrder covers the multi-page review
// pagination + ascending-order preservation + the null-user nil-guard. Per-page
// bodies live in fixtures; the Link-header routing stays inline (behavioral).
func TestGitHubClient_PullReviews_PaginatesInOrder(t *testing.T) {
	t.Parallel()

	page1 := readGithubFixture(t, "reviews-page1.json") // alice comments then approves
	page2 := readGithubFixture(t, "reviews-page2.json") // bob approves; a null-user review

	var gotPath string
	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write(page2)
			return
		}
		w.Header().Set("Link", fmt.Sprintf("<http://%s%s?page=2>; rel=\"next\"", r.Host, r.URL.Path))
		_, _ = w.Write(page1)
	}))

	reviews, err := c.PullReviews(context.Background(), "peasant-labs/schema", 42)
	if err != nil {
		t.Fatalf("PullReviews: %v", err)
	}
	if want := "/repos/peasant-labs/schema/pulls/42/reviews"; gotPath != want {
		t.Fatalf("request path = %q, want %q", gotPath, want)
	}
	// NextPage loop collects ALL pages in ascending order.
	if len(reviews) != 4 {
		t.Fatalf("collected %d reviews, want 4 (both pages): %+v", len(reviews), reviews)
	}
	wantStates := []release.ReviewState{release.ReviewStateCommented, release.ReviewStateApproved, release.ReviewStateApproved, release.ReviewStateCommented}
	for i, want := range wantStates {
		if reviews[i].State != want {
			t.Fatalf("review[%d].State = %q, want %q (order preserved)", i, reviews[i].State, want)
		}
	}
	// nil-guard: the null-user review maps to a nil User.
	if reviews[3].User != nil {
		t.Fatalf("null-user review mapped to %+v, want nil User", reviews[3].User)
	}
	// End-to-end: the decoded reviews feed LatestApprovers in today's order.
	approvers := release.LatestApprovers(reviews)
	if len(approvers) != 2 || approvers[0] != "alice" || approvers[1] != "bob" {
		t.Fatalf("LatestApprovers = %v, want [alice bob]", approvers)
	}
}

func TestSplitRepo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		repo      string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{name: "valid owner/repo", repo: "peasant-labs/schema", wantOwner: "peasant-labs", wantName: "schema"},
		{name: "missing slash", repo: "schema", wantErr: true},
		{name: "empty owner", repo: "/schema", wantErr: true},
		{name: "empty repo", repo: "peasant-labs/", wantErr: true},
		{name: "three segments", repo: "a/b/c", wantErr: true},
		{name: "empty string", repo: "", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			owner, name, err := splitRepo(tc.repo)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("splitRepo(%q) = (%q,%q,nil), want error", tc.repo, owner, name)
				}
				if got := err.Error(); !strings.Contains(got, "owner/repo form") || !strings.Contains(got, tc.repo) {
					t.Fatalf("error %q is not actionable (want it to name the bad value %q and owner/repo form)", got, tc.repo)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitRepo(%q) unexpected error: %v", tc.repo, err)
			}
			if owner != tc.wantOwner || name != tc.wantName {
				t.Fatalf("splitRepo(%q) = (%q,%q), want (%q,%q)", tc.repo, owner, name, tc.wantOwner, tc.wantName)
			}
		})
	}
}
