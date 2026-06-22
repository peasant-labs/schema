package release

import (
	"fmt"
	"regexp"
	"strings"
)

// This file holds the PURE message-construction helpers for the CI auto-land
// triangle (SLICE-G). They produce the two commit messages of GitLab's
// "squash + merge-commit" shape reproduced on GitHub:
//
//	*   M  Merge branch '<feat>' into '<target>'   (parents [T, S]; --no-ff)
//	|\
//	| * S  <PR title>                              (the PR's commits squashed)
//	|/
//	*   T  (previous target tip)
//
// S's message is built by SquashMessage; M's title/body by MergeCommitTitle /
// MergeCommitBody. Keeping them pure (no git/gh I/O) makes them golden-testable
// and keeps the `release-guard land` orchestration (cmd/release-guard/land.go)
// a thin wire-up over the subprocess seam.
//
// NOTE (internal/landing extraction): these helpers currently live in
// internal/release alongside the other release-grammar helpers. If a second
// consumer ever needs them they generalise cleanly into an internal/landing
// package; today there is exactly one consumer (release-guard land), so they
// stay here rather than forcing a one-file package.

// SquashMessage builds the commit message for the single squashed commit S that
// `release-guard land` creates on the target tip before the merge API authors M.
//
// It mirrors the message construction of ~/dotfiles/scripts/git-prmerge (the
// manual reference flow the user already uses) EXACTLY: the PR title, a blank
// line, the PR body, then a "---" rule and an oldest-first list of the squashed
// commit headlines. The git-prmerge construction (step 2 of that script) is:
//
//	gh pr view --jq '.title + "\n\n" + (.body // "")'         > msg   # title\n\nbody\n
//	printf '\n---\nSquashed commits:\n'                       >> msg   # blank line, rule, header
//	git log --reverse --format='- %s' <target>..<feature>     >> msg   # - headline (oldest first)
//
// BREADCRUMB (G-DRY) ↔ ~/dotfiles/scripts/git-prmerge, step 2.
// git-prmerge is an EXTERNAL dotfiles tool and is NEVER run by schema CI, so this
// is a ONE-TIME authoring-parity SNAPSHOT, not continuous parity. The golden test
// (landing_test.go, TestSquashMessageGolden) pins OUR format against
// testdata/squash_message.golden. If you intentionally change this format, update
// BOTH the golden fixture AND git-prmerge's step 2 in the same change so the
// snapshot keeps documenting a real correspondence.
func SquashMessage(prTitle, prBody string, squashedHeadlines []string) string {
	var b strings.Builder
	b.WriteString(prTitle)
	b.WriteString("\n\n")
	b.WriteString(prBody)
	b.WriteString("\n\n---\nSquashed commits:\n")
	for _, headline := range squashedHeadlines {
		b.WriteString("- ")
		b.WriteString(headline)
		b.WriteString("\n")
	}
	return b.String()
}

// MergeCommitTitle is the title of the merge commit M the GitHub merge API
// authors. It mirrors GitLab's default ("Merge branch '<source>' into
// '<target>'") so the auto-landed history reads identically to the sfurs
// reference the user asked to reproduce.
func MergeCommitTitle(headRef, target string) string {
	return fmt.Sprintf("Merge branch '%s' into '%s'", headRef, target)
}

// MergeCommitBody is the body of the merge commit M. It carries the GitLab-style
// "Closes #<issue>" + "See PR #<number>" trailer (the GitHub analogue of the
// sfurs "Closes #942 / See merge request …!770" example). The "Closes #" line is
// emitted only when an issue reference was found in the PR body; "See PR #" always
// references the landed PR.
func MergeCommitBody(closesIssue string, prNumber int) string {
	see := fmt.Sprintf("See PR #%d", prNumber)
	if closesIssue == "" {
		return see
	}
	return fmt.Sprintf("Closes #%s\n\n%s", closesIssue, see)
}

// MergeableState is GitHub's computed `.mergeable_state` for a pull request —
// the AUTHORITATIVE single-field gate for the auto-land merge. It is typed (per
// the repo's no-stringly-typed rule) so the F1 gate compares against named
// constants, not bare strings.
//
// CRITICAL: `mergeable_state` is NOT the same as the `.mergeable` BOOLEAN.
// `.mergeable` is TRUE whenever there is no merge CONFLICT — it stays TRUE even
// when required checks are red/pending or required reviews are missing (i.e.
// when `mergeable_state == "blocked"`). Gating on `.mergeable` would wrongly
// auto-land a red PR; the gate MUST be `mergeable_state == "clean"`.
type MergeableState string

const (
	// MergeClean: no conflict AND required checks pass AND required reviews
	// satisfied. The ONLY state the auto-land merge proceeds on.
	MergeClean MergeableState = "clean"
	// MergeBlocked: required checks failing/pending or required reviews missing
	// (no conflict, so the `.mergeable` boolean is still TRUE — the trap).
	MergeBlocked MergeableState = "blocked"
	// MergeDirty: the merge would conflict.
	MergeDirty MergeableState = "dirty"
	// MergeBehind: the head is behind the base branch (the base advanced) — the
	// auto-land signal to re-reshape S onto the new tip.
	MergeBehind MergeableState = "behind"
	// MergeUnstable: mergeable, but a non-required check is failing. Refused
	// (the gate proceeds ONLY on a settled clean).
	MergeUnstable MergeableState = "unstable"
	// MergeHasHooks: mergeable + clean, but the repo has pre-receive hooks.
	MergeHasHooks MergeableState = "has_hooks"
	// MergeDraft: the PR is a draft.
	MergeDraft MergeableState = "draft"
	// MergeUnknown: GitHub has not finished (re)computing mergeability — the
	// NON-TERMINAL state the poll waits out. Right after a force-push GitHub
	// resets to this while it recomputes for the new head.
	MergeUnknown MergeableState = "unknown"
)

// MergeGateDecision is the action the F1 poll loop takes from a SINGLE
// observation of a PR's mergeability after force-pushing the squashed commit S.
type MergeGateDecision int

const (
	// MergeGateWait: the observation is not yet trustworthy (GitHub still
	// recomputing, the push not yet reflected, or S's own checks not yet
	// settled) — keep polling.
	MergeGateWait MergeGateDecision = iota
	// MergeGateProceed: settled clean with S's checks having run — merge.
	MergeGateProceed
	// MergeGateReshape: the base advanced (behind) — re-reshape S onto the new
	// tip and re-enter the poll.
	MergeGateReshape
	// MergeGateRefuse: a terminal non-clean state — refuse the auto-land.
	MergeGateRefuse
)

// ClassifyMergeGate is the PURE F1 decision: given one mergeability observation,
// decide whether to wait, proceed, re-reshape, or refuse. This is the
// load-bearing safety logic, isolated from the gh/git I/O so it is directly
// unit-testable.
//
//   - headMatches: the PR's reported head sha == the S we just force-pushed. If
//     false, the API has not yet reflected our push (or someone else pushed) — a
//     reported state would be STALE, so we wait.
//   - checksSettled: every check-run on S has completed. This defeats the
//     stale-pre-push-"clean" race: right after the force-push GitHub can briefly
//     report the PRE-push head's "clean" before S's re-triggered checks even
//     start. Requiring S's OWN checks to have run before trusting "clean" means a
//     stale "clean" (checks not yet run for S) is treated as WAIT, not proceed.
//
// REFUSE covers blocked/dirty/unstable/has_hooks/draft (anything terminal that
// is not clean and not the re-reshape signal `behind`).
func ClassifyMergeGate(state MergeableState, headMatches, checksSettled bool) MergeGateDecision {
	if !headMatches {
		return MergeGateWait
	}
	switch state {
	case MergeUnknown, "":
		return MergeGateWait
	case MergeBehind:
		return MergeGateReshape
	case MergeClean:
		if !checksSettled {
			return MergeGateWait
		}
		return MergeGateProceed
	default: // blocked, dirty, unstable, has_hooks, draft
		return MergeGateRefuse
	}
}

// closesIssuePattern matches a GitHub closing keyword followed by an issue
// number (e.g. "Closes #114", "fixes #7", "Resolved #42"). Case-insensitive,
// matching GitHub's own closing-keyword grammar.
var closesIssuePattern = regexp.MustCompile(`(?i)\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s+#(\d+)`)

// ExtractClosesIssue returns the FIRST issue number referenced by a GitHub
// closing keyword in the PR body, or "" if none is present. It lets `land`
// propagate the PR's "Closes #N" intent onto the merge commit M without a
// separate flag, mirroring how git-prmerge inherits the PR body verbatim.
func ExtractClosesIssue(prBody string) string {
	m := closesIssuePattern.FindStringSubmatch(prBody)
	if m == nil {
		return ""
	}
	return m[1]
}
