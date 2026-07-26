package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// assertStringsEqual fails unless got and want are the same length with equal
// elements in order.
func assertStringsEqual(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v (length %d != %d)", label, got, want, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %q, want %q", label, i, got[i], want[i])
		}
	}
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
	if len(cases) != 20 {
		t.Fatalf("github seam fixture has %d cases, want 20 (fixture truncated?)", len(cases))
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

			case "ref":
				ref, err := c.Ref(context.Background(), "peasant-labs/schema", tc.Ref)
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("Ref: %v", err)
				}
				if ref.SHA != tc.WantSHA {
					t.Fatalf("ref SHA = %q, want %q", ref.SHA, tc.WantSHA)
				}
				if tc.WantPath != "" && gotPath != tc.WantPath {
					t.Fatalf("request path = %q, want %q (owner/repo split + singular git/ref)", gotPath, tc.WantPath)
				}

			case "commit":
				commit, err := c.Commit(context.Background(), "peasant-labs/schema", tc.SHA)
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("Commit: %v", err)
				}
				if commit.SHA != tc.WantSHA {
					t.Fatalf("commit SHA = %q, want %q", commit.SHA, tc.WantSHA)
				}
				if commit.TreeSHA != tc.WantTreeSHA {
					t.Fatalf("commit TreeSHA = %q, want %q", commit.TreeSHA, tc.WantTreeSHA)
				}
				if tc.WantMessage != "" && commit.Message != tc.WantMessage {
					t.Fatalf("commit Message = %q, want %q", commit.Message, tc.WantMessage)
				}
				assertStringsEqual(t, "commit ParentSHAs", commit.ParentSHAs, tc.WantParents)

			case "pull":
				pull, err := c.Pull(context.Background(), "peasant-labs/schema", tc.PR)
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("Pull: %v", err)
				}
				if pull.Number != tc.WantNumber {
					t.Fatalf("pull Number = %d, want %d", pull.Number, tc.WantNumber)
				}
				if pull.Title != tc.WantTitle {
					t.Fatalf("pull Title = %q, want %q", pull.Title, tc.WantTitle)
				}

			case "create_commit":
				commit, err := c.CreateCommit(context.Background(), "peasant-labs/schema", release.NewCommit{
					Message:    tc.Message,
					TreeSHA:    tc.Tree,
					ParentSHAs: tc.Parents,
				})
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("CreateCommit: %v", err)
				}
				if commit.SHA != tc.WantSHA {
					t.Fatalf("created commit SHA = %q, want %q", commit.SHA, tc.WantSHA)
				}
				if commit.TreeSHA != tc.WantTreeSHA {
					t.Fatalf("created commit TreeSHA = %q, want %q", commit.TreeSHA, tc.WantTreeSHA)
				}
				assertStringsEqual(t, "created commit ParentSHAs", commit.ParentSHAs, tc.WantParents)

			case "update_ref":
				err := c.UpdateRefFastForward(context.Background(), "peasant-labs/schema", tc.Ref, tc.NewSHA)
				if tc.WantNotFastForward {
					if !errors.Is(err, release.ErrNotFastForward) {
						t.Fatalf("UpdateRefFastForward error = %v, want it to wrap release.ErrNotFastForward", err)
					}
					return
				}
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					// A non-fast-forward-reason 422 must NOT be misreported as ErrNotFastForward.
					if errors.Is(err, release.ErrNotFastForward) {
						t.Fatalf("a non-FF-reason 422 was misreported as release.ErrNotFastForward: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("UpdateRefFastForward: %v", err)
				}

			case "tags":
				tags, err := c.Tags(context.Background(), "peasant-labs/schema")
				if len(tc.WantErrContains) > 0 {
					assertErrContains(t, err, tc.WantErrContains)
					return
				}
				if err != nil {
					t.Fatalf("Tags: %v", err)
				}
				got := make([]string, len(tags))
				for i, tag := range tags {
					got[i] = tag.Name + "=" + tag.CommitSHA
				}
				assertStringsEqual(t, "tags", got, tc.WantTags)

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

// TestGitHubClient_CreateCommit_RequestShape pins the write-path request body:
// POST git/commits must carry the message, the tree SHA, and the parents in
// order [T, S] (first-parent T). The response is decoded through the real
// wrapper; here the assertion is on what leaves the client.
func TestGitHubClient_CreateCommit_RequestShape(t *testing.T) {
	t.Parallel()

	respBody := readGithubFixture(t, "create-commit.json")
	var (
		gotMethod, gotPath string
		gotReqBody         []byte
	)
	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotReqBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(respBody)
	}))

	// These mirror what the bubbler actually POSTed to mint M2 during the live
	// bubble seed run, so the request asserted here and the harvested response in
	// create-commit.json are a matched real pair (see cases.yaml PROVENANCE).
	const (
		tip    = "245069a38ab8d4b1e78f43c1fc24788bbd87769a" // T (first parent): the prior bubble M1
		squash = "cb200266fec5cf07f041222002f364be3c1e85d2" // S (second parent): the squash of PR #38
		tree   = "e2b98895746fb17be48ec4c12f39dead5b3eab0a"
		msg    = "Merge PR #38: ci(release-guard): manual workflow_dispatch seed trigger for bubble --boundary"
	)
	commit, err := c.CreateCommit(context.Background(), "peasant-labs/schema", release.NewCommit{
		Message:    msg,
		TreeSHA:    tree,
		ParentSHAs: []string{tip, squash},
	})
	if err != nil {
		t.Fatalf("CreateCommit: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if want := "/repos/peasant-labs/schema/git/commits"; gotPath != want {
		t.Fatalf("path = %q, want %q", gotPath, want)
	}
	// go-github's Commit.MarshalJSON emits parents as a bare SHA array, which is
	// exactly GitHub's documented POST git/commits request shape.
	var sent struct {
		Message string   `json:"message"`
		Tree    string   `json:"tree"`
		Parents []string `json:"parents"`
	}
	if err := json.Unmarshal(gotReqBody, &sent); err != nil {
		t.Fatalf("decode request body %q: %v", gotReqBody, err)
	}
	if sent.Message != msg {
		t.Fatalf("request message = %q, want %q", sent.Message, msg)
	}
	if sent.Tree != tree {
		t.Fatalf("request tree = %q, want %q", sent.Tree, tree)
	}
	if len(sent.Parents) != 2 || sent.Parents[0] != tip || sent.Parents[1] != squash {
		t.Fatalf("request parents = %v, want [%s %s] (first-parent T)", sent.Parents, tip, squash)
	}
	// The decoded response still maps to the own-type merge commit.
	if commit.SHA != "2b6a8d8821bb368dffa449c7e723e46c0d36b499" {
		t.Fatalf("created commit SHA = %q, want the fixture merge SHA", commit.SHA)
	}
}

// TestGitHubClient_UpdateRefFastForward_RequestShape pins the never-clobber
// write path: PATCH git/refs/{ref} must send the new SHA with force=false so a
// forced (history-rewriting) ref update is not even representable.
func TestGitHubClient_UpdateRefFastForward_RequestShape(t *testing.T) {
	t.Parallel()

	respBody := readGithubFixture(t, "update-ref-ok.json")
	var (
		gotMethod, gotPath string
		gotReqBody         []byte
	)
	c := newTestGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotReqBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write(respBody)
	}))

	// The M2 sha the live bubble's fast-forward compare-and-swap actually advanced
	// develop to, matching the harvested update-ref-ok.json body.
	const newSHA = "2b6a8d8821bb368dffa449c7e723e46c0d36b499"
	if err := c.UpdateRefFastForward(context.Background(), "peasant-labs/schema", "heads/develop", newSHA); err != nil {
		t.Fatalf("UpdateRefFastForward: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Fatalf("method = %q, want PATCH", gotMethod)
	}
	if want := "/repos/peasant-labs/schema/git/refs/heads/develop"; gotPath != want {
		t.Fatalf("path = %q, want %q", gotPath, want)
	}
	var sent struct {
		SHA   string `json:"sha"`
		Force *bool  `json:"force"`
	}
	if err := json.Unmarshal(gotReqBody, &sent); err != nil {
		t.Fatalf("decode request body %q: %v", gotReqBody, err)
	}
	if sent.SHA != newSHA {
		t.Fatalf("request sha = %q, want %q", sent.SHA, newSHA)
	}
	if sent.Force == nil || *sent.Force {
		t.Fatalf("request force = %v, want false (never-clobber fast-forward)", sent.Force)
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
