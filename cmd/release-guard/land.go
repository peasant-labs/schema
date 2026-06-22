package main

// land.go implements `release-guard land` — the CI auto-land that reproduces
// GitLab's "squash + merge-commit" triangle on a GitHub PR landing into develop:
//
//	*   M  Merge branch '<feat>' into '<develop>'   (parents [T, S]; --no-ff)
//	|\
//	| * S  <PR title>                               (the PR's commits squashed onto T)
//	|/
//	*   T  (previous develop tip)
//
// Sequence: maintainer+approval preconditions -> reshape (rebase head onto the
// target tip T, soft-reset, commit S via release.SquashMessage, force-push S with
// --force-with-lease) -> F1 poll-to-terminal mergeable_state gate (merge only on a
// settled "clean" with S's OWN re-triggered checks having run) -> F2 base-tip==T
// recheck (re-reshape on drift, bounded) -> gh-api PUT .../merge (merge_method=
// merge, sha=S head-lock) so GitHub authors the --no-ff M and marks the PR MERGED.
//
// ALL git and gh calls route through the package `commandOutput` seam (the same
// var main_test.go fakes), NOT a net/http client, so the orchestration is fully
// unit-testable: poll / stale-clean / blocked-refusal / base-drift / conflict /
// already-merged paths are all driven by faking commandOutput. The core logic
// returns errors (land); only the CLI entry point (runLand) calls fatalf.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/peasant-labs/schema/internal/release"
)

// Tunables, declared as vars so tests can shrink the poll to run instantly.
var (
	// landPollInterval is the wait between mergeable_state polls.
	landPollInterval = 2 * time.Second
	// landPollMax bounds the F1 poll (timeout = landPollMax * landPollInterval).
	landPollMax = 60
	// landReshapeMax bounds the F2 re-reshape retry on repeated base drift.
	landReshapeMax = 5
	// landSleep is the sleep primitive (overridden to a no-op in tests).
	landSleep = time.Sleep
)

// errAlreadyMerged signals an idempotent no-op: the PR is (or became) merged, so
// land succeeds without doing anything.
var errAlreadyMerged = errors.New("pr already merged")

// prMeta is the subset of a GitHub pull request the land flow reads.
type prMeta struct {
	Number         int                    `json:"number"`
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	Merged         bool                   `json:"merged"`
	Mergeable      *bool                  `json:"mergeable"`
	MergeableState release.MergeableState `json:"mergeable_state"`
	Head           prRef                  `json:"head"`
	Base           prRef                  `json:"base"`
}

type prRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// checkRunsResponse is the subset of the commit check-runs API we consume.
type checkRunsResponse struct {
	TotalCount int        `json:"total_count"`
	CheckRuns  []checkRun `json:"check_runs"`
}

type checkRun struct {
	Status string `json:"status"` // "queued" | "in_progress" | "completed"
}

// landOpts is the parsed `land` invocation.
type landOpts struct {
	prNum   int
	target  string
	remote  string
	actor   string
	signKey string
	repo    string
}

// runLand is the CLI boundary: parse, run land, and fatalf on error.
func runLand(args []string) {
	if err := land(args); err != nil {
		if errors.Is(err, errAlreadyMerged) {
			return // idempotent no-op already reported by land
		}
		fatalf("%v", err)
	}
}

// land is the testable core of the auto-land flow. It returns an actionable
// error on any refusal/failure; errAlreadyMerged is an idempotent success.
func land(args []string) error {
	opts, err := parseLandArgs(args)
	if err != nil {
		return err
	}

	pr, err := fetchPRMeta(opts.repo, opts.prNum)
	if err != nil {
		return fmt.Errorf("land: cannot fetch PR #%d on %s (gh api): %w", opts.prNum, opts.repo, err)
	}

	// Idempotent: an already-merged PR is a no-op success (a re-fired trigger or
	// a retried run must not error).
	if pr.Merged {
		fmt.Printf("land: PR #%d is already merged; nothing to do\n", opts.prNum)
		return errAlreadyMerged
	}

	// Re-verify the gate at run time (do not trust the workflow trigger): a
	// maintainer APPROVED review, and — when the triggering actor is known — that
	// the actor is a maintainer. Reuses the existing check-approval / check-
	// maintainer helpers so the predicate is defined exactly once.
	if err := requireMaintainerApproval(opts.repo, opts.prNum); err != nil {
		return err
	}
	if opts.actor != "" {
		if err := requireMaintainer(opts.repo, opts.actor); err != nil {
			return err
		}
	}

	// F2 outer loop: reshape onto the current target tip, gate, recheck the tip,
	// and re-reshape if the base drifted under us. Bounded.
	for attempt := 1; attempt <= landReshapeMax; attempt++ {
		t, err := targetTip(opts.remote, opts.target)
		if err != nil {
			return fmt.Errorf("land: cannot resolve %s/%s tip: %w", opts.remote, opts.target, err)
		}

		s, err := reshape(opts, pr, t)
		if err != nil {
			return err
		}

		// F1 gate: poll mergeable_state to terminal; proceed only on settled clean.
		decision, err := pollMergeGate(opts.repo, opts.prNum, s)
		if err != nil {
			if errors.Is(err, errAlreadyMerged) {
				return errAlreadyMerged
			}
			return err
		}
		if decision == release.MergeGateReshape {
			fmt.Printf("land: PR #%d base advanced during the poll; re-reshaping (attempt %d/%d)\n", opts.prNum, attempt, landReshapeMax)
			continue
		}

		// F2 base-tip recheck immediately before the merge: the merge API `sha`
		// param locks the HEAD (S) only and does NOT 409 when the base advances,
		// so an explicit develop-tip==T recheck is required to avoid an M off a
		// stale base. On drift, re-reshape onto the new tip.
		t2, err := targetTip(opts.remote, opts.target)
		if err != nil {
			return fmt.Errorf("land: cannot re-resolve %s/%s tip before merge: %w", opts.remote, opts.target, err)
		}
		if t2 != t {
			fmt.Printf("land: %s/%s advanced %s->%s between gate and merge; re-reshaping (attempt %d/%d)\n", opts.remote, opts.target, short(t), short(t2), attempt, landReshapeMax)
			continue
		}

		if err := merge(opts.repo, pr, s, opts.target); err != nil {
			return err
		}
		fmt.Printf("land: PR #%d auto-landed — S=%s merged into %s as a --no-ff triangle; the PR now shows MERGED\n", opts.prNum, short(s), opts.target)
		return nil
	}

	return fmt.Errorf("land: PR #%d could not be landed after %d reshape attempts — %s/%s kept advancing under the merge. Re-run the auto-land, or land manually with git-prmerge", opts.prNum, landReshapeMax, opts.remote, opts.target)
}

func parseLandArgs(args []string) (landOpts, error) {
	opts := landOpts{target: "develop", remote: "origin"}
	var prArg string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--pr":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("land: --pr requires a value")
			}
			prArg = args[i]
		case "--target":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("land: --target requires a value")
			}
			opts.target = args[i]
		case "--remote":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("land: --remote requires a value")
			}
			opts.remote = args[i]
		case "--actor":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("land: --actor requires a value")
			}
			opts.actor = args[i]
		case "--sign-key":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("land: --sign-key requires a value")
			}
			opts.signKey = args[i]
		default:
			return opts, fmt.Errorf("land: unknown flag %q", args[i])
		}
	}
	if prArg == "" {
		return opts, fmt.Errorf("usage: release-guard land --pr <number> --target develop [--remote origin] [--actor <login>] [--sign-key <gpg-id>]")
	}
	n, err := strconv.Atoi(prArg)
	if err != nil || n <= 0 {
		return opts, fmt.Errorf("land: --pr must be a positive PR number, got %q", prArg)
	}
	opts.prNum = n
	opts.repo = os.Getenv("GITHUB_REPOSITORY")
	if opts.repo == "" {
		return opts, fmt.Errorf("land: $GITHUB_REPOSITORY is not set; cannot resolve PR #%d", n)
	}
	return opts, nil
}

// requireMaintainerApproval refuses unless the PR has ≥1 standing APPROVED review
// from an admin/maintain collaborator (reuses fetchReviews + LatestApprovers +
// findMaintainerApprover — the same code path as `check-approval`).
func requireMaintainerApproval(repo string, prNum int) error {
	reviews, err := fetchReviews(repo, strconv.Itoa(prNum))
	if err != nil {
		return fmt.Errorf("land: cannot read reviews for PR #%d on %s: %w", prNum, repo, err)
	}
	approvers := release.LatestApprovers(reviews)
	if len(approvers) == 0 {
		return fmt.Errorf("land: PR #%d has no standing APPROVED reviews; a maintainer (%s or %s) must approve before auto-land", prNum, release.PermAdmin, release.PermMaintain)
	}
	approver, perm, err := findMaintainerApprover(repo, approvers)
	if err != nil {
		return fmt.Errorf("land: %w", err)
	}
	if approver == "" {
		return fmt.Errorf("land: PR #%d has standing approvals, but none from an %s/%s maintainer; a maintainer must approve before auto-land", prNum, release.PermAdmin, release.PermMaintain)
	}
	fmt.Printf("land: PR #%d has a standing maintainer approval from %s (%s)\n", prNum, approver, perm)
	return nil
}

// requireMaintainer refuses unless the triggering actor is an admin/maintain
// collaborator (reuses collaboratorPermission + IsMaintainer — the `check-
// maintainer` code path).
func requireMaintainer(repo, user string) error {
	perm, err := collaboratorPermission(repo, user)
	if err != nil {
		return fmt.Errorf("land: cannot read the collaborator permission of %q on %s (gh api): %w", user, repo, err)
	}
	if !release.IsMaintainer(perm) {
		return fmt.Errorf("land: actor %q has permission %q on %s, but auto-land requires a maintainer (%s or %s); ask a maintainer to trigger it", user, perm, repo, release.PermAdmin, release.PermMaintain)
	}
	fmt.Printf("land: trigger actor %s is a maintainer (%s)\n", user, perm)
	return nil
}

// reshape rebases the PR head onto the target tip T, squashes it into ONE commit
// S whose message is built by release.SquashMessage, and force-pushes S to the
// PR branch with --force-with-lease. It returns S's sha. A rebase conflict ends
// the auto-land with an actionable, file-naming error (the run-it-locally path).
func reshape(opts landOpts, pr prMeta, t string) (string, error) {
	headRef := pr.Head.Ref
	if headRef == "" {
		return "", fmt.Errorf("land: PR #%d has no head ref; cannot reshape", pr.Number)
	}

	// Fetch ONLY the PR head, then put it on a local working branch so the
	// rebase/soft-reset/commit operate on a known ref. We fetch the head alone
	// (not "<target> <headRef>" together) so FETCH_HEAD is unambiguous — a
	// multi-ref fetch leaves `git checkout FETCH_HEAD` resolving to the first
	// fetched ref, not the head. The target tip T is already in the object store
	// (targetTip fetched it) and the rebase below targets T's sha directly.
	if _, err := runGit("fetch", opts.remote, headRef); err != nil {
		return "", fmt.Errorf("land: cannot fetch the PR head %s/%s: %w", opts.remote, headRef, gitErr(err))
	}
	const work = "release-guard-land"
	if _, err := runGit("checkout", "-B", work, "FETCH_HEAD"); err != nil {
		return "", fmt.Errorf("land: cannot create the working branch from the PR head: %w", gitErr(err))
	}

	// Rebase onto T. Intermediate commits are squashed away in the soft-reset, so
	// they need not be signed.
	if _, err := runGit("-c", "commit.gpgsign=false", "rebase", t); err != nil {
		conflicts := conflictedFiles()
		_, _ = runGit("rebase", "--abort")
		filesMsg := "the conflicting files could not be enumerated"
		if len(conflicts) > 0 {
			filesMsg = "conflicting files:\n  - " + strings.Join(conflicts, "\n  - ")
		}
		return "", fmt.Errorf("land: PR #%d cannot be auto-landed — rebasing the PR head onto %s/%s conflicts.\n%s\nResolve it locally and re-land: check out the PR branch and run `git prmerge %s`, push --force-with-lease, then re-trigger the auto-land",
			pr.Number, opts.remote, opts.target, filesMsg, opts.target)
	}

	// Collect the squashed headlines (oldest-first) AFTER the rebase, exactly as
	// git-prmerge does (`git log --reverse --format=%s <T>..HEAD`).
	headlines, err := commitHeadlines(t)
	if err != nil {
		return "", err
	}

	msg := release.SquashMessage(pr.Title, pr.Body, headlines)
	if _, err := runGit("reset", "--soft", t); err != nil {
		return "", fmt.Errorf("land: cannot soft-reset onto %s before squashing: %w", short(t), gitErr(err))
	}

	commitArgs := []string{}
	if opts.signKey != "" {
		// Optional bot-GPG path for a develop that requires signed commits.
		commitArgs = append(commitArgs, "-c", "commit.gpgsign=true", "-c", "user.signingkey="+opts.signKey)
	}
	commitArgs = append(commitArgs, "commit", "--cleanup=verbatim", "-m", msg)
	if opts.signKey != "" {
		commitArgs = append(commitArgs, "-S")
	}
	if _, err := runGit(commitArgs...); err != nil {
		return "", fmt.Errorf("land: cannot create the squashed commit S: %w", gitErr(err))
	}

	s, err := revParse("HEAD")
	if err != nil {
		return "", fmt.Errorf("land: cannot resolve the squashed commit S: %w", err)
	}

	// Force-push S to the PR branch. --force-with-lease refuses if someone else
	// advanced the branch since we fetched it.
	if _, err := runGit("push", "--force-with-lease", opts.remote, "HEAD:"+headRef); err != nil {
		return "", fmt.Errorf("land: cannot force-push the squashed commit S to %s/%s: %w\nSomeone may have pushed to the PR branch; re-trigger the auto-land", opts.remote, headRef, gitErr(err))
	}
	return s, nil
}

// pollMergeGate is the F1 loop: it polls the PR's mergeable_state to a terminal
// state and returns MergeGateProceed only on a settled "clean" with S's own
// re-triggered checks having run, or MergeGateReshape on a base advance. It
// returns an actionable error on a terminal non-clean state (refuse) or on the
// bounded timeout, and errAlreadyMerged if the PR merged out from under the poll.
func pollMergeGate(repo string, prNum int, s string) (release.MergeGateDecision, error) {
	for i := 0; i < landPollMax; i++ {
		pr, err := fetchPRMeta(repo, prNum)
		if err != nil {
			return release.MergeGateRefuse, fmt.Errorf("land: cannot poll PR #%d mergeability (gh api): %w", prNum, err)
		}
		if pr.Merged {
			fmt.Printf("land: PR #%d became merged during the poll; nothing to do\n", prNum)
			return release.MergeGateRefuse, errAlreadyMerged
		}

		headMatches := pr.Head.SHA == s
		settled := false
		if headMatches && pr.MergeableState == release.MergeClean {
			// Only worth the extra API call once the reported state is the clean
			// we might act on: confirm S's OWN checks have run (defeats the stale
			// pre-push clean).
			settled, err = checksSettled(repo, s)
			if err != nil {
				return release.MergeGateRefuse, fmt.Errorf("land: cannot read check-runs for S=%s on PR #%d: %w", short(s), prNum, err)
			}
		}

		switch release.ClassifyMergeGate(pr.MergeableState, headMatches, settled) {
		case release.MergeGateProceed:
			return release.MergeGateProceed, nil
		case release.MergeGateReshape:
			return release.MergeGateReshape, nil
		case release.MergeGateRefuse:
			return release.MergeGateRefuse, fmt.Errorf("land: PR #%d is not auto-landable — mergeable_state=%q (required checks failing/pending, reviews missing, a conflict, or a draft). The `mergeable` boolean alone is not the gate. Make the PR green+approved and re-trigger, or land manually with git-prmerge", prNum, pr.MergeableState)
		case release.MergeGateWait:
			landSleep(landPollInterval)
		}
	}
	return release.MergeGateRefuse, fmt.Errorf("land: timed out after %d polls waiting for PR #%d to settle to a terminal mergeable_state with S's checks complete. The required checks may be slow or stuck; re-trigger once the PR is green", landPollMax, prNum)
}

// merge calls the GitHub merge API to author the --no-ff merge commit M (parents
// [T, S]) and mark the PR MERGED. merge_method=merge is what yields the triangle
// (squash would be linear; a hand-pushed M would show "Closed"). sha=S locks the
// HEAD. The App token the workflow runs this under is what RE-TRIGGERS downstream
// workflows (e.g. release-pr.yml Trigger B) — a GITHUB_TOKEN merge would not.
func merge(repo string, pr prMeta, s, target string) error {
	title := release.MergeCommitTitle(pr.Head.Ref, target)
	body := release.MergeCommitBody(release.ExtractClosesIssue(pr.Body), pr.Number)

	// IRREDUCIBLE MICRO-TOCTOU (do NOT attempt to "fix"): an inherent race
	// remains between the F2 base-tip recheck just above and this merge call.
	// GitHub's merge API has NO base-SHA compare-and-swap — the `sha` param locks
	// only the HEAD (S), not the base — so if develop advances in the
	// sub-millisecond window between the recheck and this PUT (with the merge
	// still clean), M could be authored as [T', S]. This is a platform
	// limitation, already minimized by the immediate pre-merge recheck + the
	// `auto-land-develop` concurrency group (serializes auto-lands) + the
	// re-reshape loop. There is no correct further mitigation in the client.
	out, err := commandOutput(
		"gh", "api", "--method", "PUT",
		fmt.Sprintf("repos/%s/pulls/%d/merge", repo, pr.Number),
		"-f", "merge_method=merge",
		"-f", "sha="+s,
		"-f", "commit_title="+title,
		"-f", "commit_message="+body,
	)
	if err != nil {
		return fmt.Errorf("land: the merge API call for PR #%d failed: %w\nIf the head moved (409) or the PR went non-clean, re-trigger the auto-land; otherwise land manually with git-prmerge.\ngh output: %s", pr.Number, gitErr(err), strings.TrimSpace(string(out)))
	}
	return nil
}

// --- gh / git helpers (all via the commandOutput seam) ---

func fetchPRMeta(repo string, prNum int) (prMeta, error) {
	out, err := commandOutput("gh", "api", fmt.Sprintf("repos/%s/pulls/%d", repo, prNum))
	if err != nil {
		return prMeta{}, err
	}
	var pr prMeta
	if err := json.Unmarshal(out, &pr); err != nil {
		return prMeta{}, fmt.Errorf("parsing gh pull JSON: %w", err)
	}
	return pr, nil
}

// checksSettled reports whether EVERY check-run on commit sha has completed (and
// there is at least one). A still-queued/in_progress check (S's re-triggered
// checks not yet done) makes the "clean" we might have read untrustworthy.
func checksSettled(repo, sha string) (bool, error) {
	out, err := commandOutput("gh", "api", fmt.Sprintf("repos/%s/commits/%s/check-runs", repo, sha))
	if err != nil {
		return false, err
	}
	var resp checkRunsResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return false, fmt.Errorf("parsing gh check-runs JSON: %w", err)
	}
	if resp.TotalCount == 0 || len(resp.CheckRuns) == 0 {
		return false, nil
	}
	for _, run := range resp.CheckRuns {
		if run.Status != "completed" {
			return false, nil
		}
	}
	return true, nil
}

// targetTip fetches the target branch and returns its tip sha.
func targetTip(remote, target string) (string, error) {
	if _, err := runGit("fetch", remote, target); err != nil {
		return "", gitErr(err)
	}
	return revParse(remote + "/" + target)
}

// commitHeadlines returns the subjects of the commits in <base>..HEAD,
// oldest-first (the squashed-commit headlines).
func commitHeadlines(base string) ([]string, error) {
	out, err := runGit("log", "--reverse", "--format=%s", base+"..HEAD")
	if err != nil {
		return nil, fmt.Errorf("land: cannot list the squashed commit headlines: %w", gitErr(err))
	}
	var headlines []string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line != "" {
			headlines = append(headlines, line)
		}
	}
	return headlines, nil
}

// conflictedFiles lists the unmerged paths after a failed rebase.
func conflictedFiles() []string {
	out, err := runGit("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f := strings.TrimSpace(line); f != "" {
			files = append(files, f)
		}
	}
	return files
}

func runGit(args ...string) ([]byte, error) {
	return commandOutput("git", args...)
}

// gitErr surfaces a subprocess's stderr (captured by exec.Cmd.Output into
// *ExitError) so the actionable messages name the real failure, not just "exit
// status 1".
func gitErr(err error) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if s := strings.TrimSpace(string(ee.Stderr)); s != "" {
			return fmt.Errorf("%w: %s", err, s)
		}
	}
	return err
}

func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
