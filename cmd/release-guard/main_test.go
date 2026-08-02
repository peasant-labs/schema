package main

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// --- interface mocks (mock the DEPENDENCY, never the SUT) -------------------

type mockGitHubClient struct {
	collaboratorPermissionFn func(ctx context.Context, repo, user string) (release.CollaboratorPermission, error)
	workflowRunsForCommitFn  func(ctx context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error)
	releaseExistsFn          func(ctx context.Context, repo, tag string) (bool, error)
	pullReviewsFn            func(ctx context.Context, repo string, prNumber int) ([]release.Review, error)
	refFn                    func(ctx context.Context, repo, ref string) (release.GitRef, error)
	commitFn                 func(ctx context.Context, repo, sha string) (release.GitCommit, error)
	pullFn                   func(ctx context.Context, repo string, number int) (release.Pull, error)
	createCommitFn           func(ctx context.Context, repo string, in release.NewCommit) (release.GitCommit, error)
	updateRefFastForwardFn   func(ctx context.Context, repo, ref, newSHA string) error
	tagsFn                   func(ctx context.Context, repo string) ([]release.TagRef, error)
}

func (m *mockGitHubClient) CollaboratorPermission(ctx context.Context, repo, user string) (release.CollaboratorPermission, error) {
	return m.collaboratorPermissionFn(ctx, repo, user)
}
func (m *mockGitHubClient) WorkflowRunsForCommit(ctx context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error) {
	return m.workflowRunsForCommitFn(ctx, repo, workflowFile, commitSHA)
}
func (m *mockGitHubClient) ReleaseExists(ctx context.Context, repo, tag string) (bool, error) {
	return m.releaseExistsFn(ctx, repo, tag)
}
func (m *mockGitHubClient) PullReviews(ctx context.Context, repo string, prNumber int) ([]release.Review, error) {
	return m.pullReviewsFn(ctx, repo, prNumber)
}
func (m *mockGitHubClient) Ref(ctx context.Context, repo, ref string) (release.GitRef, error) {
	return m.refFn(ctx, repo, ref)
}
func (m *mockGitHubClient) Commit(ctx context.Context, repo, sha string) (release.GitCommit, error) {
	return m.commitFn(ctx, repo, sha)
}
func (m *mockGitHubClient) Pull(ctx context.Context, repo string, number int) (release.Pull, error) {
	return m.pullFn(ctx, repo, number)
}
func (m *mockGitHubClient) CreateCommit(ctx context.Context, repo string, in release.NewCommit) (release.GitCommit, error) {
	return m.createCommitFn(ctx, repo, in)
}
func (m *mockGitHubClient) UpdateRefFastForward(ctx context.Context, repo, ref, newSHA string) error {
	return m.updateRefFastForwardFn(ctx, repo, ref, newSHA)
}
func (m *mockGitHubClient) Tags(ctx context.Context, repo string) ([]release.TagRef, error) {
	if m.tagsFn == nil {
		return nil, nil
	}
	return m.tagsFn(ctx, repo)
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

type initialFinalCLIInput struct {
	Requested               string   `yaml:"requested"`
	Configured              string   `yaml:"configured"`
	ProductTags             []string `yaml:"productTags"`
	ProductTagError         string   `yaml:"productTagError"`
	RCTags                  []string `yaml:"rcTags"`
	GreenRC                 bool     `yaml:"greenRC"`
	PublicationCompleted    bool     `yaml:"publicationCompleted"`
	PublicationError        string   `yaml:"publicationError"`
	ForbidPublicationLookup bool     `yaml:"forbidPublicationLookup"`
}

type initialFinalCLIExpected struct {
	ErrorContains string `yaml:"errorContains"`
}

//go:embed testdata/initial-final/cases.yaml
var initialFinalCLICasesYAML []byte

func TestRunCheckFinal_InitialFinalCases(t *testing.T) {
	corpus, err := testcase.LoadCorpus[initialFinalCLIInput, initialFinalCLIExpected](initialFinalCLICasesYAML)
	if err != nil {
		t.Fatalf("load initial-final CLI corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 5)
	assert.RequireValid(t, corpus)

	for _, c := range corpus.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			git := &mockGitRunner{
				revParseFn: func(_ context.Context, ref string) (string, error) {
					if ref == c.Input.Requested {
						return finalCommit, nil
					}
					for _, rcTag := range c.Input.RCTags {
						if ref == rcTag {
							return rcCommit, nil
						}
					}
					return "", errors.New("unexpected ref " + ref)
				},
				listTagsFn: func(_ context.Context, pattern string) ([]string, error) {
					switch pattern {
					case "v*":
						if c.Input.ProductTagError != "" {
							return nil, errors.New(c.Input.ProductTagError)
						}
						return c.Input.ProductTags, nil
					case c.Input.Requested + "-rc*":
						return c.Input.RCTags, nil
					default:
						return nil, errors.New("unexpected pattern " + pattern)
					}
				},
				isAncestorFn: func(_ context.Context, ancestor, descendant string) (bool, error) {
					return ancestor != "" && descendant == finalCommit, nil
				},
			}
			gh := &mockGitHubClient{workflowRunsForCommitFn: func(_ context.Context, _, _ string, commit string) ([]release.WorkflowRun, error) {
				if c.Input.GreenRC && commit == rcCommit {
					return []release.WorkflowRun{{HeadSHA: commit, Status: release.WorkflowRunCompleted, Conclusion: release.WorkflowRunSuccess}}, nil
				}
				return nil, nil
			}, releaseExistsFn: func(_ context.Context, repo, tag string) (bool, error) {
				if c.Input.ForbidPublicationLookup {
					t.Fatal("later final with prior release history must not depend on the initial final's publication lookup")
				}
				if repo != testRepo || tag != c.Input.Configured {
					return false, errors.New("unexpected publication evidence request")
				}
				if c.Input.PublicationError != "" {
					return false, errors.New(c.Input.PublicationError)
				}
				return c.Input.PublicationCompleted, nil
			}}
			err := runCheckFinal(context.Background(), gh, git, testRepo, []string{
				"--tag", c.Input.Requested, "--initial-final", c.Input.Configured,
			})
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("runCheckFinal error = %v, want substring %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("runCheckFinal returned unexpected error: %v", err)
			}
		})
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

// --- composition root: the token is read from the workflows' $GH_TOKEN --------
//
// The release workflows export $GH_TOKEN (`env: { GH_TOKEN: ${{ github.token }} }`);
// the tool must read that exact variable or every API-backed gate authenticates
// with an empty token and fails. readGitHubToken is the testable core of the
// composition root's fail-fast (mustGitHubClient maps its error to os.Exit).
//
// This test uses t.Setenv (process-global env), so it must NOT run in parallel.
func TestReadGitHubToken(t *testing.T) {
	// Pin the variable name the composition root reads to the one the workflows
	// export — the whole point of the BLOCKER fix. Setting the RAW literal below
	// (not the const) then independently proves readGitHubToken reads THIS name.
	if gitHubTokenEnv != "GH_TOKEN" {
		t.Fatalf("gitHubTokenEnv = %q, want \"GH_TOKEN\" (the variable every release workflow exports)", gitHubTokenEnv)
	}

	t.Run("reads the value of the exported GH_TOKEN", func(t *testing.T) {
		t.Setenv("GH_TOKEN", "ghp_a_real_looking_token")
		token, err := readGitHubToken("check-final")
		if err != nil {
			t.Fatalf("readGitHubToken with $GH_TOKEN set = %v, want nil", err)
		}
		if token != "ghp_a_real_looking_token" {
			t.Fatalf("readGitHubToken = %q, want the $GH_TOKEN value", token)
		}
	})

	t.Run("empty token yields an actionable fail-fast", func(t *testing.T) {
		t.Setenv("GH_TOKEN", "")
		_, err := readGitHubToken("check-approval")
		if err == nil {
			t.Fatal("readGitHubToken with an empty $GH_TOKEN = nil error, want an actionable fail-fast")
		}
		// what/why/where/how + it names the subcommand and the env var. Requiring
		// "GH_TOKEN" ALSO guards the regression: the pre-swap name is not a
		// superstring of "GH_TOKEN", so a revert to it would fail this assertion.
		for _, want := range []string{"GH_TOKEN", "check-approval", "empty or unset", "token with repo read access"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q is not actionable (missing %q)", err.Error(), want)
			}
		}
	})
}

// --- $GITHUB_OUTPUT: emit APPENDS version=/kind= to the step-output file -------
//
// Exit-code mapping is proven-by-construction, not unit-tested here: every
// subcommand funnels through the uniform `if err != nil { fatalf(...) }` in main()
// (a one-line os.Exit(1) wrapper), and each handler's error vs nil-error is
// covered by the per-gate `wantErr` tests above. What IS observable without
// re-exec'ing the process is emit()'s file effect, asserted here.
//
// Uses t.Setenv (process-global $GITHUB_OUTPUT), so it must NOT run in parallel.
func TestEmit_AppendsToGitHubOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "github_output")
	// Seed a prior line so the test proves emit APPENDS (O_APPEND), never
	// truncates — the file accumulates outputs across a job's steps.
	if err := os.WriteFile(path, []byte("preexisting=1\n"), 0o644); err != nil {
		t.Fatalf("seed $GITHUB_OUTPUT: %v", err)
	}
	t.Setenv("GITHUB_OUTPUT", path)

	v, kind, err := release.ParseTag("v1.2.3-rc4")
	if err != nil {
		t.Fatalf("ParseTag: %v", err)
	}
	if err := emit(v, kind); err != nil {
		t.Fatalf("emit: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read $GITHUB_OUTPUT: %v", err)
	}
	want := "preexisting=1\nversion=v1.2.3-rc4\nkind=rc\n"
	if string(got) != want {
		t.Fatalf("$GITHUB_OUTPUT = %q, want %q (emit appends version=/kind=)", string(got), want)
	}
}

// --- findMaintainerApprover: first maintainer in order; none; error-wrap -------

func TestFindMaintainerApprover_ReturnsFirstMaintainerInOrder(t *testing.T) {
	t.Parallel()

	// carol is not a maintainer (skipped); dave and erin both are — the FIRST in
	// slice order (dave) must win, so approval is attributed deterministically.
	perms := map[string]release.CollaboratorPermission{
		"carol": release.PermWrite,
		"dave":  release.PermMaintain,
		"erin":  release.PermAdmin,
	}
	gh := &mockGitHubClient{
		collaboratorPermissionFn: func(_ context.Context, _, user string) (release.CollaboratorPermission, error) {
			return perms[user], nil
		},
	}
	approver, perm, err := findMaintainerApprover(context.Background(), gh, testRepo, []string{"carol", "dave", "erin"})
	if err != nil {
		t.Fatalf("findMaintainerApprover: %v", err)
	}
	if approver != "dave" || perm != release.PermMaintain {
		t.Fatalf("findMaintainerApprover = (%q,%q), want (dave, maintain) — the first maintainer in order", approver, perm)
	}
}

func TestFindMaintainerApprover_NoneAreMaintainers(t *testing.T) {
	t.Parallel()

	gh := &mockGitHubClient{
		collaboratorPermissionFn: func(_ context.Context, _, _ string) (release.CollaboratorPermission, error) {
			return release.PermWrite, nil
		},
	}
	approver, perm, err := findMaintainerApprover(context.Background(), gh, testRepo, []string{"carol", "dave"})
	if err != nil {
		t.Fatalf("findMaintainerApprover: %v", err)
	}
	if approver != "" || perm != "" {
		t.Fatalf("findMaintainerApprover = (%q,%q), want empty (no approver is a maintainer)", approver, perm)
	}
}

func TestFindMaintainerApprover_PermissionErrorIsWrapped(t *testing.T) {
	t.Parallel()

	gh := &mockGitHubClient{
		collaboratorPermissionFn: func(_ context.Context, _, _ string) (release.CollaboratorPermission, error) {
			return "", errors.New("boom: 500 from GitHub")
		},
	}
	_, _, err := findMaintainerApprover(context.Background(), gh, testRepo, []string{"carol"})
	if err == nil {
		t.Fatal("findMaintainerApprover with a permission error = nil, want a wrapped error")
	}
	for _, want := range []string{"collaborator permission of approving reviewer", "carol", testRepo} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing actionable substring %q", err.Error(), want)
		}
	}
}

// --- check-workflow handler flag parsing (--policy / --release / unknown) ------

func TestRunCheckWorkflow_FlagParse(t *testing.T) {
	t.Parallel()

	t.Run("unknown flag is rejected by name", func(t *testing.T) {
		t.Parallel()
		err := runCheckWorkflow([]string{"--nope", "x"})
		if err == nil || !strings.Contains(err.Error(), `unknown flag "--nope"`) {
			t.Fatalf("runCheckWorkflow(--nope) = %v, want an unknown-flag error naming it", err)
		}
	})

	t.Run("--policy requires a value", func(t *testing.T) {
		t.Parallel()
		err := runCheckWorkflow([]string{"--policy"})
		if err == nil || !strings.Contains(err.Error(), "--policy requires a value") {
			t.Fatalf("runCheckWorkflow(--policy) = %v, want a missing-value error", err)
		}
	})

	t.Run("--release requires a value", func(t *testing.T) {
		t.Parallel()
		err := runCheckWorkflow([]string{"--release"})
		if err == nil || !strings.Contains(err.Error(), "--release requires a value") {
			t.Fatalf("runCheckWorkflow(--release) = %v, want a missing-value error", err)
		}
	})

	t.Run("--policy value flows into the loader", func(t *testing.T) {
		t.Parallel()
		// A custom --policy path is threaded to LoadWorkflowPolicy; a nonexistent
		// one surfaces a read error naming that exact path (proves the wiring).
		missing := filepath.Join(t.TempDir(), "custom.policy.yml")
		err := runCheckWorkflow([]string{"--policy", missing})
		if err == nil || !strings.Contains(err.Error(), missing) {
			t.Fatalf("runCheckWorkflow(--policy %s) = %v, want a load error naming the custom path", missing, err)
		}
	})
}
