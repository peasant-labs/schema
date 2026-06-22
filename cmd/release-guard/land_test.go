package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// landFake is a stateful fake for the commandOutput seam that drives the whole
// `land` flow: the git plumbing (fetch/checkout/rebase/log/reset/commit/push/
// rev-list/diff) and the gh-api calls (PR fetch, reviews, collaborator
// permission, check-runs, merge). Each test wires up the responses for its
// scenario; the fake records every call so assertions can inspect the sequence.
type landFake struct {
	t       *testing.T
	repo    string
	prNum   int
	headRef string
	title   string
	body    string

	// initial PR fetch (before reshape)
	initialMerged bool
	initialHead   string

	// approval (reviews + collaborator permission)
	approver   string            // login of the approving maintainer
	approver2  string            // collaborator permission for the approver (default "maintain")
	perms      map[string]string // per-login permission override (login -> "admin"/"maintain"/"write"/...)
	noApprover bool              // reviews carry no standing approval

	// push failure injection
	pushErr error

	// git
	rebaseConflict bool
	conflictFiles  []string
	headlines      []string
	revHEAD        []string // S sha returned by `rev-list -n 1 HEAD`, popped per reshape
	tips           []string // origin/<target> tip returned by `rev-list -n 1 origin/<target>`, popped per call

	// poll PR responses (every PR fetch AFTER the initial one)
	pollPRs []pollPR

	// check-runs settled? popped each time checksSettled is queried
	checks []bool

	mergeErr error

	// recording
	calls         [][]string
	prFetchi      int
	revHEADi      int
	tipsi         int
	checksi       int
	pollPRi       int
	pushCount     int
	mergeCalls    [][]string
	commitCalls   [][]string
	squashMessage string
}

type pollPR struct {
	head   string
	state  string
	merged bool
}

func (f *landFake) pullsPath() string { return fmt.Sprintf("repos/%s/pulls/%d", f.repo, f.prNum) }
func (f *landFake) mergePath() string { return fmt.Sprintf("repos/%s/pulls/%d/merge", f.repo, f.prNum) }
func (f *landFake) reviewsPath() string {
	return fmt.Sprintf("repos/%s/pulls/%d/reviews", f.repo, f.prNum)
}

func hasArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func (f *landFake) prJSON(merged bool, head, state string) []byte {
	return []byte(fmt.Sprintf(
		`{"number":%d,"title":%q,"body":%q,"merged":%v,"mergeable":true,"mergeable_state":%q,"head":{"ref":%q,"sha":%q},"base":{"ref":"develop","sha":"basesha"}}`,
		f.prNum, f.title, f.body, merged, state, f.headRef, head))
}

func (f *landFake) dispatch(name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string{name}, args...))

	switch name {
	case "gh":
		return f.dispatchGH(args)
	case "git":
		return f.dispatchGit(args)
	}
	return nil, fmt.Errorf("unexpected command %s %v", name, args)
}

func (f *landFake) dispatchGH(args []string) ([]byte, error) {
	if len(args) < 2 || args[0] != "api" {
		return nil, fmt.Errorf("unexpected gh args %v", args)
	}
	// merge: gh api --method PUT .../merge -f ...
	if args[1] == "--method" {
		f.mergeCalls = append(f.mergeCalls, append([]string{"gh"}, args...))
		if f.mergeErr != nil {
			return []byte("merge failed"), f.mergeErr
		}
		return []byte(`{"merged":true,"sha":"mergesha"}`), nil
	}
	path := args[1]
	switch {
	case path == f.pullsPath() && len(args) == 2:
		// PR metadata fetch
		if f.prFetchi == 0 {
			f.prFetchi++
			return f.prJSON(f.initialMerged, f.initialHead, "clean"), nil
		}
		i := f.pollPRi
		if i >= len(f.pollPRs) {
			i = len(f.pollPRs) - 1
		}
		p := f.pollPRs[i]
		f.pollPRi++
		f.prFetchi++
		return f.prJSON(p.merged, p.head, p.state), nil
	case path == f.reviewsPath() && hasArg(args, "--paginate") && hasArg(args, "--slurp"):
		if f.noApprover {
			return []byte(`[[{"user":{"login":"alice"},"state":"CHANGES_REQUESTED"}]]`), nil
		}
		return []byte(fmt.Sprintf(`[[{"user":{"login":%q},"state":"APPROVED"}]]`, f.approver)), nil
	case strings.HasSuffix(path, "/permission"):
		// path: repos/<repo>/collaborators/<login>/permission
		login := path
		if i := strings.Index(path, "/collaborators/"); i >= 0 {
			login = strings.TrimSuffix(path[i+len("/collaborators/"):], "/permission")
		}
		if f.perms != nil {
			if perm, ok := f.perms[login]; ok {
				return []byte(perm + "\n"), nil
			}
		}
		perm := f.approver2
		if perm == "" {
			perm = "maintain"
		}
		return []byte(perm + "\n"), nil
	case strings.Contains(path, "/commits/") && strings.HasSuffix(path, "/check-runs"):
		settled := false
		if f.checksi < len(f.checks) {
			settled = f.checks[f.checksi]
		}
		f.checksi++
		if settled {
			return []byte(`{"total_count":1,"check_runs":[{"status":"completed","conclusion":"success"}]}`), nil
		}
		return []byte(`{"total_count":1,"check_runs":[{"status":"in_progress"}]}`), nil
	}
	return nil, fmt.Errorf("unexpected gh api path %q (args %v)", path, args)
}

func (f *landFake) dispatchGit(args []string) ([]byte, error) {
	switch {
	case args[0] == "fetch", args[0] == "checkout", args[0] == "reset":
		return []byte{}, nil
	case hasArg(args, "rebase") && args[len(args)-1] == "--abort":
		return []byte{}, nil
	case hasArg(args, "rebase"):
		if f.rebaseConflict {
			return nil, errors.New("CONFLICT (content): merge conflict; exit status 1")
		}
		return []byte{}, nil
	case args[0] == "log":
		return []byte(strings.Join(f.headlines, "\n") + "\n"), nil
	case hasArg(args, "commit"):
		f.commitCalls = append(f.commitCalls, append([]string{"git"}, args...))
		// capture the squash message that follows -m
		for i := 0; i < len(args)-1; i++ {
			if args[i] == "-m" {
				f.squashMessage = args[i+1]
			}
		}
		return []byte{}, nil
	case args[0] == "rev-list":
		ref := args[len(args)-1]
		if ref == "HEAD" {
			s := "Sx"
			if f.revHEADi < len(f.revHEAD) {
				s = f.revHEAD[f.revHEADi]
			}
			f.revHEADi++
			return []byte(s + "\n"), nil
		}
		// origin/<target> tip
		tip := "Tx"
		if f.tipsi < len(f.tips) {
			tip = f.tips[f.tipsi]
		} else if len(f.tips) > 0 {
			tip = f.tips[len(f.tips)-1]
		}
		f.tipsi++
		return []byte(tip + "\n"), nil
	case args[0] == "push":
		f.pushCount++
		if f.pushErr != nil {
			return []byte("push rejected"), f.pushErr
		}
		return []byte{}, nil
	case args[0] == "diff":
		return []byte(strings.Join(f.conflictFiles, "\n") + "\n"), nil
	}
	return nil, fmt.Errorf("unexpected git args %v", args)
}

// installLandFake swaps in the fake commandOutput + no-op sleep + tiny poll
// bounds, restoring originals on cleanup.
func installLandFake(t *testing.T, f *landFake) {
	t.Helper()
	f.t = t
	t.Setenv("GITHUB_REPOSITORY", f.repo)

	origCmd := commandOutput
	origSleep := landSleep
	origMax := landPollMax
	commandOutput = f.dispatch
	landSleep = func(time.Duration) {}
	landPollMax = 20
	t.Cleanup(func() {
		commandOutput = origCmd
		landSleep = origSleep
		landPollMax = origMax
	})
}

func baseFake() *landFake {
	return &landFake{
		repo:        "peasant-labs/schema",
		prNum:       207,
		headRef:     "feat--x",
		title:       "feat: add auto-land triangle",
		body:        "Implements the triangle.\n\nCloses #114",
		approver:    "alice",
		approver2:   "maintain",
		initialHead: "OLDHEAD0000000",
		headlines:   []string{"feat: part one", "feat: part two"},
	}
}

// TestLandHappyPathTriangle exercises the full sequence: reshape -> poll settled
// clean -> base-tip recheck -> merge. It asserts the merge call inputs that make
// GitHub author M=[T,S] and mark the PR MERGED: merge_method=merge + sha=S
// head-lock + the GitLab-style title/body. (The actual M parents and the MERGED
// status are GitHub-enforced from these inputs; they are on the live-verify
// checklist, not unit-testable here.)
func TestLandHappyPathTriangle(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000", "T0000000"} // targetTip before reshape, then F2 recheck — stable
	f.pollPRs = []pollPR{{head: "S1111111", state: "clean"}}
	f.checks = []bool{true}
	installLandFake(t, f)

	if err := land([]string{"--pr", "207", "--target", "develop"}); err != nil {
		t.Fatalf("land happy path returned error: %v", err)
	}

	if len(f.mergeCalls) != 1 {
		t.Fatalf("expected exactly 1 merge API call, got %d (%v)", len(f.mergeCalls), f.mergeCalls)
	}
	mc := strings.Join(f.mergeCalls[0], " ")
	for _, want := range []string{
		"--method PUT",
		f.mergePath(),
		"merge_method=merge",
		"sha=S1111111", // head sha-lock on S
		"commit_title=Merge branch 'feat--x' into 'develop'",
		"commit_message=Closes #114\n\nSee PR #207",
	} {
		if !strings.Contains(mc, want) {
			t.Fatalf("merge call missing %q\n got: %s", want, mc)
		}
	}
	// The squashed commit S carries the SquashMessage (PR title + headlines).
	if !strings.Contains(f.squashMessage, "feat: add auto-land triangle") ||
		!strings.Contains(f.squashMessage, "Squashed commits:") ||
		!strings.Contains(f.squashMessage, "- feat: part one") {
		t.Fatalf("squash commit message not built from SquashMessage:\n%s", f.squashMessage)
	}
	if f.pushCount != 1 {
		t.Fatalf("expected 1 force-push of S, got %d", f.pushCount)
	}
}

// TestLandRefusesBlockedState is the DISCRIMINATING safety test: a PR whose
// mergeable_state=="blocked" (so the `.mergeable` boolean is TRUE — no conflict)
// MUST be refused. Reading the boolean alone would wrongly proceed.
func TestLandRefusesBlockedState(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000"}
	f.pollPRs = []pollPR{{head: "S1111111", state: "blocked"}}
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil {
		t.Fatal("land must REFUSE a mergeable_state==blocked PR, got nil error")
	}
	if !strings.Contains(err.Error(), "not auto-landable") || !strings.Contains(err.Error(), "mergeable_state=\"blocked\"") {
		t.Fatalf("refusal error not actionable: %v", err)
	}
	if len(f.mergeCalls) != 0 {
		t.Fatalf("a blocked PR must NOT be merged, got %d merge calls", len(f.mergeCalls))
	}
}

// TestLandWaitsForChecksThenMerges is the poll-to-terminal test: right after the
// force-push the PR reports "clean" but S's re-triggered checks are still
// in_progress (the stale pre-push clean). land must NOT merge on that read — it
// polls until S's checks complete, then merges.
func TestLandWaitsForChecksThenMerges(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000", "T0000000"}
	// both polls report clean; checks are in_progress first, completed second.
	f.pollPRs = []pollPR{
		{head: "S1111111", state: "clean"},
		{head: "S1111111", state: "clean"},
	}
	f.checks = []bool{false, true}
	installLandFake(t, f)

	if err := land([]string{"--pr", "207", "--target", "develop"}); err != nil {
		t.Fatalf("land returned error: %v", err)
	}
	if f.checksi != 2 {
		t.Fatalf("expected 2 check-runs polls (waited for S's checks to settle), got %d", f.checksi)
	}
	if len(f.mergeCalls) != 1 {
		t.Fatalf("expected exactly 1 merge after checks settled, got %d", len(f.mergeCalls))
	}
	if !strings.Contains(strings.Join(f.mergeCalls[0], " "), "sha=S1111111") {
		t.Fatalf("merge did not lock on S: %v", f.mergeCalls[0])
	}
}

// TestLandReshapesOnBaseDrift is the F2 test: develop advances (T -> T') between
// the gate and the merge, so the base recheck must re-reshape onto T' rather
// than merge an M off the stale base.
func TestLandReshapesOnBaseDrift(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111", "S2222222"}
	// attempt1: targetTip=T1, F2 recheck=T1' (drift!) -> re-reshape
	// attempt2: targetTip=T2, F2 recheck=T2 (stable) -> merge
	f.tips = []string{"T1111111", "T1prime0", "T2222222", "T2222222"}
	f.pollPRs = []pollPR{
		{head: "S1111111", state: "clean"}, // attempt1 poll
		{head: "S2222222", state: "clean"}, // attempt2 poll
	}
	f.checks = []bool{true, true}
	installLandFake(t, f)

	if err := land([]string{"--pr", "207", "--target", "develop"}); err != nil {
		t.Fatalf("land returned error: %v", err)
	}
	if f.pushCount != 2 {
		t.Fatalf("expected 2 reshapes (force-pushes) across the base-drift retry, got %d", f.pushCount)
	}
	if len(f.mergeCalls) != 1 {
		t.Fatalf("expected exactly 1 merge after the re-reshape, got %d", len(f.mergeCalls))
	}
	if !strings.Contains(strings.Join(f.mergeCalls[0], " "), "sha=S2222222") {
		t.Fatalf("merge must lock on the RE-RESHAPED S2 (not the stale S1): %v", f.mergeCalls[0])
	}
}

// TestLandConflictNamesFiles asserts the rebase-conflict path yields an
// actionable error naming the conflicting files + the resolve-locally
// instruction, and never merges.
func TestLandConflictNamesFiles(t *testing.T) {
	f := baseFake()
	f.tips = []string{"T0000000"}
	f.rebaseConflict = true
	f.conflictFiles = []string{"openapi/schema.yaml", "internal/store/migrations.go"}
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil {
		t.Fatal("land must fail on a rebase conflict, got nil error")
	}
	msg := err.Error()
	for _, want := range []string{"openapi/schema.yaml", "internal/store/migrations.go", "git prmerge"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("conflict error missing %q:\n%s", want, msg)
		}
	}
	if len(f.mergeCalls) != 0 {
		t.Fatalf("a conflicting PR must NOT be merged, got %d merge calls", len(f.mergeCalls))
	}
}

// TestLandAlreadyMergedNoOp asserts idempotence: an already-merged PR is a no-op
// success (no reshape, no merge call).
func TestLandAlreadyMergedNoOp(t *testing.T) {
	f := baseFake()
	f.initialMerged = true
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if !errors.Is(err, errAlreadyMerged) {
		t.Fatalf("expected errAlreadyMerged, got %v", err)
	}
	if f.pushCount != 0 || len(f.mergeCalls) != 0 {
		t.Fatalf("already-merged PR must do nothing, got pushes=%d merges=%d", f.pushCount, len(f.mergeCalls))
	}
}

// TestLandRefusesWithoutMaintainerApproval asserts the run-time approval re-check
// refuses a PR with no standing maintainer approval before any reshape.
func TestLandRefusesWithoutMaintainerApproval(t *testing.T) {
	f := baseFake()
	f.noApprover = true
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil || !strings.Contains(err.Error(), "no standing APPROVED reviews") {
		t.Fatalf("expected an approval-refusal error, got %v", err)
	}
	if f.pushCount != 0 || len(f.mergeCalls) != 0 {
		t.Fatalf("unapproved PR must not reshape or merge, got pushes=%d merges=%d", f.pushCount, len(f.mergeCalls))
	}
}

// TestLandTimesOutOnPersistentlyUnsettledChecks asserts the bounded-timeout fail:
// if S's checks never settle (clean reported but checks forever in_progress),
// land fails rather than merging.
func TestLandTimesOutOnPersistentlyUnsettledChecks(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000"}
	f.pollPRs = []pollPR{{head: "S1111111", state: "clean"}}
	f.checks = []bool{false} // pop past end -> always false (unsettled)
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected a bounded-timeout error, got %v", err)
	}
	if len(f.mergeCalls) != 0 {
		t.Fatalf("timeout must NOT merge, got %d merge calls", len(f.mergeCalls))
	}
}

// TestLandSurfacesMergeAPIError (T-M1) covers the merge-API-error branch: the
// gate passes and land reaches the merge call, but the merge API itself fails
// (e.g. a 409 from a head move, or the lost micro-TOCTOU window). land must
// surface the actionable error and NOT claim success.
func TestLandSurfacesMergeAPIError(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000", "T0000000"}
	f.pollPRs = []pollPR{{head: "S1111111", state: "clean"}}
	f.checks = []bool{true}
	f.mergeErr = errors.New("HTTP 409: Head branch was modified")
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil {
		t.Fatal("land must return an error when the merge API fails, got nil (false success)")
	}
	msg := err.Error()
	for _, want := range []string{"merge API call for PR #207 failed", "re-trigger the auto-land", "git-prmerge"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("merge-error message missing %q:\n%s", want, msg)
		}
	}
	// The merge WAS attempted (gate passed) but did not succeed.
	if len(f.mergeCalls) != 1 {
		t.Fatalf("expected exactly 1 (failed) merge attempt, got %d", len(f.mergeCalls))
	}
}

// TestLandFailsOnPersistentBaseDrift (T-M2) covers the F2 reshape-exhaustion
// bound: develop's tip advances at EVERY pre-merge recheck, so land re-reshapes
// up to landReshapeMax times and then fails loudly rather than looping forever or
// merging onto a stale base.
func TestLandFailsOnPersistentBaseDrift(t *testing.T) {
	origMax := landReshapeMax
	landReshapeMax = 3
	t.Cleanup(func() { landReshapeMax = origMax })

	f := baseFake()
	// One reshaped S + one clean poll per attempt; the targetTip pair each
	// attempt differs (pre-reshape tip != F2-recheck tip) so every attempt drifts.
	f.revHEAD = []string{"S1", "S2", "S3"}
	f.tips = []string{
		"a1", "b1", // attempt 1: reshape onto a1, recheck sees b1 -> drift
		"a2", "b2", // attempt 2
		"a3", "b3", // attempt 3
	}
	f.pollPRs = []pollPR{
		{head: "S1", state: "clean"},
		{head: "S2", state: "clean"},
		{head: "S3", state: "clean"},
	}
	f.checks = []bool{true, true, true}
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil || !strings.Contains(err.Error(), "after 3 reshape attempts") {
		t.Fatalf("expected a bounded reshape-exhaustion error, got %v", err)
	}
	if len(f.mergeCalls) != 0 {
		t.Fatalf("persistent base drift must NOT merge, got %d merge calls", len(f.mergeCalls))
	}
	if f.pushCount != 3 {
		t.Fatalf("expected %d reshapes (one per bounded attempt), got %d", landReshapeMax, f.pushCount)
	}
}

// TestLandRefusesNonMaintainerActor (T-M3a) covers the run-time maintainer
// re-verify on --actor: a non-maintainer trigger actor is refused even when a
// (separate) maintainer approval stands, and before any reshape.
func TestLandRefusesNonMaintainerActor(t *testing.T) {
	f := baseFake()
	f.approver = "alice"
	// alice is a maintainer (the standing approval); mallory triggered but is not.
	f.perms = map[string]string{"alice": "maintain", "mallory": "write"}
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop", "--actor", "mallory"})
	if err == nil || !strings.Contains(err.Error(), "actor \"mallory\"") || !strings.Contains(err.Error(), "requires a maintainer") {
		t.Fatalf("expected a non-maintainer-actor refusal, got %v", err)
	}
	if f.pushCount != 0 || len(f.mergeCalls) != 0 {
		t.Fatalf("a non-maintainer actor must not reshape or merge, got pushes=%d merges=%d", f.pushCount, len(f.mergeCalls))
	}
}

// TestLandSignKeyPath (T-M3b) covers the optional --sign-key bot-GPG reshape
// path: the squashed commit S is created with the signing config wired
// (commit.gpgsign=true + user.signingkey=<id> + -S), and the land still completes
// the triangle.
func TestLandSignKeyPath(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000", "T0000000"}
	f.pollPRs = []pollPR{{head: "S1111111", state: "clean"}}
	f.checks = []bool{true}
	installLandFake(t, f)

	if err := land([]string{"--pr", "207", "--target", "develop", "--sign-key", "BOTKEY123"}); err != nil {
		t.Fatalf("land --sign-key returned error: %v", err)
	}
	if len(f.commitCalls) != 1 {
		t.Fatalf("expected exactly 1 squash commit, got %d", len(f.commitCalls))
	}
	cc := strings.Join(f.commitCalls[0], " ")
	for _, want := range []string{"commit.gpgsign=true", "user.signingkey=BOTKEY123", "-S"} {
		if !strings.Contains(cc, want) {
			t.Fatalf("signed commit missing %q\n got: %s", want, cc)
		}
	}
	if len(f.mergeCalls) != 1 {
		t.Fatalf("expected the signed reshape to still merge, got %d merge calls", len(f.mergeCalls))
	}
}

// TestLandFailsOnForcePushReject (T-M3b, cheap) covers the force-push-failure
// branch: --force-with-lease is rejected (someone advanced the PR branch), so
// land fails actionably and never merges.
func TestLandFailsOnForcePushReject(t *testing.T) {
	f := baseFake()
	f.revHEAD = []string{"S1111111"}
	f.tips = []string{"T0000000"}
	f.pushErr = errors.New("stale info: remote ref changed")
	installLandFake(t, f)

	err := land([]string{"--pr", "207", "--target", "develop"})
	if err == nil || !strings.Contains(err.Error(), "force-push the squashed commit S") {
		t.Fatalf("expected a force-push-reject error, got %v", err)
	}
	if len(f.mergeCalls) != 0 {
		t.Fatalf("a failed force-push must NOT merge, got %d merge calls", len(f.mergeCalls))
	}
}
