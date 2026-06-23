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

func TestGitHubClient_CollaboratorPermission(t *testing.T) {
	t.Parallel()

	var gotPath string
	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"permission":"admin"}`)
	}))

	perm, err := c.CollaboratorPermission(context.Background(), "peasant-labs/schema", "octocat")
	if err != nil {
		t.Fatalf("CollaboratorPermission: %v", err)
	}
	if perm != release.PermAdmin {
		t.Fatalf("permission = %q, want %q", perm, release.PermAdmin)
	}
	// owner/repo split: the wrapper must address /repos/<owner>/<repo>/...
	if want := "/repos/peasant-labs/schema/collaborators/octocat/permission"; gotPath != want {
		t.Fatalf("request path = %q, want %q (owner/repo split)", gotPath, want)
	}
}

func TestGitHubClient_WorkflowRunsForCommit(t *testing.T) {
	t.Parallel()

	const commit = "deadbeefcafe"
	var gotHeadSHA, gotPath string
	var page1Hits int

	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeadSHA = r.URL.Query().Get("head_sha")
		gotPath = r.URL.Path
		if r.URL.Query().Get("page") == "2" {
			// Final page: a green run on the commit (plus a nil-safe decoy).
			fmt.Fprintf(w, `{"total_count":3,"workflow_runs":[{"head_sha":%q,"status":"completed","conclusion":"success"}]}`, commit)
			return
		}
		page1Hits++
		// First page: an in-progress decoy + a completed-failure decoy; link to p2.
		w.Header().Set("Link", fmt.Sprintf("<http://%s%s?page=2>; rel=\"next\"", r.Host, r.URL.Path))
		fmt.Fprintf(w, `{"total_count":3,"workflow_runs":[`+
			`{"head_sha":%q,"status":"in_progress","conclusion":null},`+
			`{"head_sha":%q,"status":"completed","conclusion":"failure"}]}`, commit, commit)
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

func TestGitHubClient_PullReviews_PaginatesInOrder(t *testing.T) {
	t.Parallel()

	var gotPath string
	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Query().Get("page") == "2" {
			// Page 2: bob approves; a null-user review must be nil-guarded.
			fmt.Fprint(w, `[{"user":{"login":"bob"},"state":"APPROVED"},{"user":null,"state":"COMMENTED"}]`)
			return
		}
		w.Header().Set("Link", fmt.Sprintf("<http://%s%s?page=2>; rel=\"next\"", r.Host, r.URL.Path))
		fmt.Fprint(w, `[{"user":{"login":"alice"},"state":"COMMENTED"},{"user":{"login":"alice"},"state":"APPROVED"}]`)
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
