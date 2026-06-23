package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

// --- interface mocks (mock the DEPENDENCY, never the SUT) -------------------

type mockGitHubClient struct {
	collaboratorPermissionFn func(ctx context.Context, repo, user string) (release.CollaboratorPermission, error)
	workflowRunsForCommitFn  func(ctx context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error)
	pullReviewsFn            func(ctx context.Context, repo string, prNumber int) ([]release.Review, error)
}

func (m *mockGitHubClient) CollaboratorPermission(ctx context.Context, repo, user string) (release.CollaboratorPermission, error) {
	return m.collaboratorPermissionFn(ctx, repo, user)
}
func (m *mockGitHubClient) WorkflowRunsForCommit(ctx context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error) {
	return m.workflowRunsForCommitFn(ctx, repo, workflowFile, commitSHA)
}
func (m *mockGitHubClient) PullReviews(ctx context.Context, repo string, prNumber int) ([]release.Review, error) {
	return m.pullReviewsFn(ctx, repo, prNumber)
}

type mockGitRunner struct {
	revParseFn   func(ctx context.Context, ref string) (string, error)
	listTagsFn   func(ctx context.Context, pattern string) ([]string, error)
	isAncestorFn func(ctx context.Context, ancestor, descendant string) (bool, error)
}

func (m *mockGitRunner) RevParse(ctx context.Context, ref string) (string, error) {
	return m.revParseFn(ctx, ref)
}
func (m *mockGitRunner) ListTags(ctx context.Context, pattern string) ([]string, error) {
	return m.listTagsFn(ctx, pattern)
}
func (m *mockGitRunner) IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error) {
	return m.isAncestorFn(ctx, ancestor, descendant)
}

const testRepo = "peasant-labs/schema"

// --- BDD #1: check-maintainer (maintainer vs non-maintainer) ----------------

func TestRunCheckMaintainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		perm    release.CollaboratorPermission
		wantErr bool
	}{
		{"admin is a maintainer", release.PermAdmin, false},
		{"maintain is a maintainer", release.PermMaintain, false},
		{"write is not a maintainer", release.PermWrite, true},
		{"triage is not a maintainer", release.PermTriage, true},
		{"read is not a maintainer", release.PermRead, true},
		{"none is not a maintainer", release.PermNone, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gh := &mockGitHubClient{
				collaboratorPermissionFn: func(_ context.Context, repo, user string) (release.CollaboratorPermission, error) {
					if repo != testRepo || user != "octocat" {
						t.Fatalf("CollaboratorPermission called with (%q,%q), want (%q, octocat)", repo, user, testRepo)
					}
					return tc.perm, nil
				},
			}
			err := runCheckMaintainer(context.Background(), gh, testRepo, []string{"--user", "octocat"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("runCheckMaintainer(perm=%s) err = %v, wantErr %v", tc.perm, err, tc.wantErr)
			}
		})
	}
}

// --- BDD #2: check-final (green via server-side HeadSHA; no-green blocks) ----

// finalCheckRunner builds a GitRunner for a v1.2.3 final with one rc (v1.2.3-rc1)
// that is an ancestor; RevParse maps the final tag and the rc tag to distinct
// 40-hex commits.
const (
	finalCommit = "1111111111111111111111111111111111111111"
	rcCommit    = "2222222222222222222222222222222222222222"
)

func finalCheckRunner() *mockGitRunner {
	return &mockGitRunner{
		revParseFn: func(_ context.Context, ref string) (string, error) {
			switch ref {
			case "v1.2.3":
				return finalCommit, nil
			case "v1.2.3-rc1":
				return rcCommit, nil
			default:
				return "", errors.New("unexpected ref " + ref)
			}
		},
		listTagsFn: func(_ context.Context, pattern string) ([]string, error) {
			if pattern != "v1.2.3-rc*" {
				return nil, errors.New("unexpected pattern " + pattern)
			}
			return []string{"v1.2.3-rc1"}, nil
		},
		isAncestorFn: func(_ context.Context, ancestor, descendant string) (bool, error) {
			return ancestor == "v1.2.3-rc1" && descendant == finalCommit, nil
		},
	}
}

func TestRunCheckFinal_GreenViaHeadSHA(t *testing.T) {
	t.Parallel()

	var gotCommitSHA string
	gh := &mockGitHubClient{
		workflowRunsForCommitFn: func(_ context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error) {
			gotCommitSHA = commitSHA
			// Decoys (other commit / in-progress / failure) + the green run on
			// the rc commit. The server-side HeadSHA filter is modelled by the
			// handler passing the rc commit; RunGreenForCommit re-checks it.
			return []release.WorkflowRun{
				{HeadSHA: "deadbeef", Status: release.WorkflowRunCompleted, Conclusion: release.WorkflowRunSuccess},
				{HeadSHA: commitSHA, Status: release.WorkflowRunInProgress, Conclusion: release.WorkflowRunNoConclusion},
				{HeadSHA: commitSHA, Status: release.WorkflowRunCompleted, Conclusion: release.WorkflowRunFailure},
				{HeadSHA: commitSHA, Status: release.WorkflowRunCompleted, Conclusion: release.WorkflowRunSuccess},
			}, nil
		},
	}
	err := runCheckFinal(context.Background(), gh, finalCheckRunner(), testRepo, []string{"--tag", "v1.2.3"})
	if err != nil {
		t.Fatalf("runCheckFinal (green rc) = %v, want nil", err)
	}
	// The run lookup must use the rc's COMMIT (resolved via RevParse), not the tag.
	if gotCommitSHA != rcCommit {
		t.Fatalf("WorkflowRunsForCommit got commit %q, want the rc commit %q", gotCommitSHA, rcCommit)
	}
}

func TestRunCheckFinal_NoGreenRunBlocks(t *testing.T) {
	t.Parallel()

	gh := &mockGitHubClient{
		workflowRunsForCommitFn: func(_ context.Context, _, _, commitSHA string) ([]release.WorkflowRun, error) {
			// No completed+success run on the rc commit → not green.
			return []release.WorkflowRun{
				{HeadSHA: commitSHA, Status: release.WorkflowRunCompleted, Conclusion: release.WorkflowRunFailure},
			}, nil
		},
	}
	err := runCheckFinal(context.Background(), gh, finalCheckRunner(), testRepo, []string{"--tag", "v1.2.3"})
	if err == nil {
		t.Fatal("runCheckFinal (no green rc) = nil, want a blocking error")
	}
}

// PARITY: a genuine git lineage failure (IsAncestor exit >1 → error) must be
// treated as not-an-ancestor and BLOCK — never silently pass.
func TestRunCheckFinal_IsAncestorErrorBlocks(t *testing.T) {
	t.Parallel()

	git := finalCheckRunner()
	git.isAncestorFn = func(_ context.Context, _, _ string) (bool, error) {
		return false, errors.New("merge-base: exit status 128 (bad object)")
	}
	gh := &mockGitHubClient{
		workflowRunsForCommitFn: func(_ context.Context, _, _, commitSHA string) ([]release.WorkflowRun, error) {
			// The rc IS green — only the lineage error must keep it from passing.
			return []release.WorkflowRun{
				{HeadSHA: commitSHA, Status: release.WorkflowRunCompleted, Conclusion: release.WorkflowRunSuccess},
			}, nil
		},
	}
	err := runCheckFinal(context.Background(), gh, git, testRepo, []string{"--tag", "v1.2.3"})
	if err == nil {
		t.Fatal("runCheckFinal must BLOCK when IsAncestor errors (green-but-lineage-broken rc), got nil")
	}
}

// --- BDD #3: check-approval (paginated approvals; maintainer gate) ----------

func TestRunCheckApproval(t *testing.T) {
	t.Parallel()

	// reviews modelling a multi-page result already flattened by PullReviews:
	// alice comments then approves; bob approves.
	reviews := []release.Review{
		{User: &release.ReviewUser{Login: "alice"}, State: release.ReviewStateCommented},
		{User: &release.ReviewUser{Login: "alice"}, State: release.ReviewStateApproved},
		{User: &release.ReviewUser{Login: "bob"}, State: release.ReviewStateApproved},
	}

	tests := []struct {
		name    string
		reviews []release.Review
		perms   map[string]release.CollaboratorPermission
		wantErr bool
	}{
		{
			name:    "standing maintainer approval (bob is maintain)",
			reviews: reviews,
			perms:   map[string]release.CollaboratorPermission{"alice": release.PermWrite, "bob": release.PermMaintain},
			wantErr: false,
		},
		{
			name:    "approvals but no maintainer among them",
			reviews: reviews,
			perms:   map[string]release.CollaboratorPermission{"alice": release.PermWrite, "bob": release.PermWrite},
			wantErr: true,
		},
		{
			name:    "no standing approvals",
			reviews: []release.Review{{User: &release.ReviewUser{Login: "alice"}, State: release.ReviewStateChangesRequested}},
			perms:   map[string]release.CollaboratorPermission{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var gotPR int
			gh := &mockGitHubClient{
				pullReviewsFn: func(_ context.Context, repo string, prNumber int) ([]release.Review, error) {
					gotPR = prNumber
					return tc.reviews, nil
				},
				collaboratorPermissionFn: func(_ context.Context, _, user string) (release.CollaboratorPermission, error) {
					return tc.perms[user], nil
				},
			}
			err := runCheckApproval(context.Background(), gh, testRepo, []string{"--pr", "94"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("runCheckApproval err = %v, wantErr %v", err, tc.wantErr)
			}
			if gotPR != 94 {
				t.Fatalf("PullReviews got PR #%d, want #94 (--pr parsed to int)", gotPR)
			}
		})
	}
}

func TestRunCheckApproval_RejectsNonNumericPR(t *testing.T) {
	t.Parallel()
	gh := &mockGitHubClient{} // never called
	err := runCheckApproval(context.Background(), gh, testRepo, []string{"--pr", "not-a-number"})
	if err == nil || !strings.Contains(err.Error(), "is not a number") {
		t.Fatalf("runCheckApproval(--pr not-a-number) = %v, want a 'not a number' error", err)
	}
}

// --- carry-note #5: $GITHUB_OUTPUT byte-parity ------------------------------

func TestFormatOutput_ByteParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tag      string
		wantLine string
	}{
		{"v1.2.3", "version=v1.2.3\nkind=final\n"},
		{"v1.2.3-rc4", "version=v1.2.3-rc4\nkind=rc\n"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			v, kind, err := release.ParseTag(tc.tag)
			if err != nil {
				t.Fatalf("ParseTag(%q): %v", tc.tag, err)
			}
			// Byte-identical to the pre-swap emit() format string.
			if got := formatOutput(v, kind); got != tc.wantLine {
				t.Fatalf("formatOutput(%q) = %q, want byte-identical %q", tc.tag, got, tc.wantLine)
			}
		})
	}
}

// --- BDD #1 mechanism: empty token fails fast at construction ---------------

func TestNewGitHubClient_EmptyTokenErrors(t *testing.T) {
	t.Parallel()
	if _, err := newGitHubClient(""); err == nil {
		t.Fatal("newGitHubClient(\"\") = nil error, want a fail-fast empty-token error")
	}
}

// --- parse handlers reject malformed input ----------------------------------

func TestRunParseTag_RejectsNamespacedTag(t *testing.T) {
	t.Parallel()
	if err := runParseTag([]string{"pkg/schema/v1.2.3"}); err == nil {
		t.Fatal("runParseTag(namespaced tag) = nil, want rejection")
	}
}
