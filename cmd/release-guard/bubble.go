package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/peasant-labs/schema/internal/release"
)

// --- bubble subcommand: API-only squash-merge bubbler --------------------------
//
// `release-guard bubble --pr <n> [--max-attempts N] [--boundary <sha>]` is invoked by
// bubble-merge.yml on a merged pull_request:closed event. It reads develop's
// tip, drain-all-bubbles every pending squash into a signed merge-commit
// triangle (M parents [T, S], M.tree == S.tree), and advances develop via a
// single fast-forward compare-and-swap — failing loud (non-zero exit + a GitHub
// Actions ::error::) if it cannot, leaving develop linear and unchanged.
//
// It composes the canonical GitHubClient seam (SLICE-1) directly in package main
// — the repo's sole composition root — so there is no internal/bubble or
// internal/githubapi package: the orchestration reads its own-types (release.*)
// off the one go-github wrapper.

// bubble retry / walk bounds.
const (
	defaultBubbleMaxAttempts = 5
	// defaultWalkBound caps the first-parent drain-all walk: collectPending
	// examines at most this many commits (the tip plus up to walkBound-1
	// ancestors) before giving up, so a develop with an unexpectedly long run of
	// single-parent commits (no merge boundary) fails loud instead of walking
	// unbounded (R-A).
	defaultWalkBound = 256
)

// TerminalError is returned when the bounded retry loop is exhausted without a
// successful bubble (persistent non-fast-forward / tip churn). The branch is
// left linear and unchanged; the CLI maps this to a non-zero exit with an
// actionable ::error:: (B10).
type TerminalError struct {
	Attempts int
}

func (e *TerminalError) Error() string {
	return fmt.Sprintf(
		"bubble: gave up after %d attempt(s): develop kept advancing during the fast-forward compare-and-swap. "+
			"develop was left linear and unchanged (no force, no partial state). Re-run the bubble workflow once develop settles.",
		e.Attempts,
	)
}

// bubbler drives one bubble run against the injected GitHubClient seam: read tip
// -> decide -> (drain-all + resolve provenance -> create M-chain -> one FF-CAS)
// with bounded retry and fail-loud terminal behavior. All decision/topology
// logic defers to internal/release (DecideBubble, BubbleFactsFromCommits,
// BubbleMergeMessage); this type is only I/O + control flow.
type bubbler struct {
	gh          GitHubClient
	repo        string
	branch      string // e.g. "develop"
	ref         string // unprefixed ref name, e.g. "heads/develop"
	maxAttempts int
	walkBound   int
	// boundary, when non-empty, is the operator-supplied drain floor SHA (from
	// --boundary): the first-parent walk stops when it reaches this commit and
	// anchors the bubble's first parent T there. It is the explicit override for a
	// deliberate first-run backfill that the released-history guard would
	// otherwise refuse.
	boundary string
	logf     func(format string, args ...any)
}

// run performs the bubble. It returns nil on success or a no-op skip
// (already-bubbled / not-a-squash), a *TerminalError when the bounded retry is
// exhausted, or a wrapped error on an unrecoverable failure (e.g. no merge
// boundary within the walk bound, or an API error).
func (b *bubbler) run(ctx context.Context) error {
	for attempt := 1; attempt <= b.maxAttempts; attempt++ {
		ref, err := b.gh.Ref(ctx, b.repo, b.ref)
		if err != nil {
			return fmt.Errorf("read %s tip: %w", b.branch, err)
		}
		tip, err := b.gh.Commit(ctx, b.repo, ref.SHA)
		if err != nil {
			return fmt.Errorf("read %s tip commit %s: %w", b.branch, ref.SHA, err)
		}

		facts, err := b.factsFromTip(ctx, tip)
		if err != nil {
			return err
		}

		switch decision := release.DecideBubble(facts); decision {
		case release.BubbleSkipNotSquash:
			b.logf("notice: %s tip %s is not a single-parent squash (%d parents); nothing to bubble", b.branch, tip.SHA, len(tip.ParentSHAs))
			return nil
		case release.BubbleSkipAlreadyBubbled:
			b.logf("notice: %s tip %s is already a bubble merge commit; nothing to do", b.branch, tip.SHA)
			return nil
		case release.BubbleRetryTipAdvanced:
			// Defensive / unreachable by construction: factsFromTip derives the
			// candidate squash S FROM the current tip (tip == S for the 1-parent
			// case), so it never produces the tip-advanced fact-set that yields
			// RetryTipAdvanced. A genuine develop-advanced race is caught by the
			// FF-CAS ErrNotFastForward path below (re-read + retry), NOT by a
			// DecideBubble decision. If this arm ever fires, a fact-construction
			// invariant was violated — fail loud rather than silently looping to a
			// TerminalError. (DecideBubble's RetryTipAdvanced value and its B-CORE
			// unit tests stay intact; only this orchestration arm is defensive.)
			return fmt.Errorf(
				"bubble: unexpected RetryTipAdvanced from factsFromTip for %s tip %s — "+
					"tip-advance must be caught by the FF-CAS retry, not a DecideBubble decision; "+
					"this indicates a fact-construction invariant was violated",
				b.branch, tip.SHA,
			)
		case release.BubbleProceed:
			items, err := b.buildItems(ctx, tip)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				b.logf("notice: no pending squashes above the merge boundary on %s; nothing to bubble", b.branch)
				return nil
			}
			if err := b.bubble(ctx, items); err != nil {
				if errors.Is(err, release.ErrNotFastForward) {
					b.logf("notice: %s advanced during the bubble fast-forward; retrying (attempt %d/%d)", b.branch, attempt, b.maxAttempts)
					continue
				}
				return fmt.Errorf("bubble %d squash(es) onto %s: %w", len(items), b.branch, err)
			}
			b.logf("notice: bubbled %d squash(es) onto %s via one fast-forward update", len(items), b.branch)
			return nil
		default:
			return fmt.Errorf("bubble: unhandled decision %s for %s tip %s", decision, b.branch, tip.SHA)
		}
	}
	return &TerminalError{Attempts: b.maxAttempts}
}

// factsFromTip derives release.BubbleFacts from the branch tip commit, choosing
// the candidate squash S as follows:
//   - tip is single-parent: the tip itself is a fresh squash S -> Proceed
//     (drain-all then gathers any stacked older squashes too).
//   - tip is a two-parent merge whose tree equals its second parent's tree and
//     whose second parent is single-parent: our bubble M (parents [T, S]) ->
//     AlreadyBubbled (no-op).
//   - any OTHER two-parent tip (a regular non-bubble merge, or a bubble of a
//     merge): not a fresh single-parent squash -> NotSquash (skip, no-op). A
//     stable non-bubble merge tip is classified as a SKIP, NOT a retry, so it
//     never false-fails the bounded loop.
//   - tip has any other parent count (root / octopus): not a squash -> skip.
func (b *bubbler) factsFromTip(ctx context.Context, tip release.GitCommit) (release.BubbleFacts, error) {
	switch len(tip.ParentSHAs) {
	case 1:
		// S = tip (the just-fast-forwarded squash) -> Proceed.
		return release.BubbleFactsFromCommits(tip, tip), nil
	case 2:
		// Compare the merge tip against its second parent S (the bubbled squash).
		sSHA := tip.ParentSHAs[1]
		s, err := b.gh.Commit(ctx, b.repo, sSHA)
		if err != nil {
			return release.BubbleFacts{}, fmt.Errorf("read second parent %s of %s tip %s: %w", sSHA, b.branch, tip.SHA, err)
		}
		if tip.TreeSHA == s.TreeSHA && len(s.ParentSHAs) == 1 {
			// Our bubble M: parents [T, S], M.tree == S.tree, S single-parent
			// (BubbleFactsFromCommits yields TipParentCount=2, TipTreeEqualsS,
			// !TipEqualsS -> AlreadyBubbled).
			return release.BubbleFactsFromCommits(s, tip), nil
		}
		// Any other two-parent merge tip is not a fresh squash to bubble. Report
		// SParentCount=2 (the tip is a two-parent merge, not a single-parent
		// squash) so DecideBubble returns NotSquash -> skip no-op, rather than a
		// RetryTipAdvanced that would false-fail on a stable tip.
		return release.BubbleFacts{
			SParentCount:   2,
			TipParentCount: 2,
			TipTreeEqualsS: tip.TreeSHA == s.TreeSHA,
			TipEqualsS:     false,
		}, nil
	default:
		// 0 parents (root) or >2 (octopus): SParentCount != 1 -> NotSquash.
		return release.BubbleFactsFromCommits(tip, tip), nil
	}
}

// pendingCommit is an internal carrier pairing a squash with its full commit
// message (needed for co-author / issue trailers) during the drain-all walk.
type pendingCommit struct {
	squash  release.Squash
	message string
}

// buildItems performs the drain-all walk from the tip back to the merge
// boundary, resolves each pending squash's provenance, and renders its bubble
// merge message, returning the oldest-first []release.BubbleItem.
func (b *bubbler) buildItems(ctx context.Context, tip release.GitCommit) ([]release.BubbleItem, error) {
	pending, err := b.collectPending(ctx, tip)
	if err != nil {
		return nil, err
	}
	if err := b.guardReleasedHistory(ctx, pending); err != nil {
		return nil, err
	}
	items := make([]release.BubbleItem, 0, len(pending))
	for _, pc := range pending {
		prov, err := b.resolveProvenance(ctx, pc)
		if err != nil {
			return nil, err
		}
		items = append(items, release.BubbleItem{
			Squash:  pc.squash,
			Message: renderBubbleMessage(prov, pc.squash.SHA),
		})
	}
	return items, nil
}

// collectPending walks first-parents from the tip, collecting the run of
// single-parent commits (the un-bubbled squashes) until it reaches the merge
// boundary T (a commit whose parent count != 1). It returns them oldest-first.
// If no boundary is found within walkBound, it fails loud (R-A) rather than
// guessing a boundary.
func (b *bubbler) collectPending(ctx context.Context, tip release.GitCommit) ([]pendingCommit, error) {
	var pending []pendingCommit
	c := tip
	for steps := 0; steps < b.walkBound; steps++ {
		if b.boundary != "" && c.SHA == b.boundary {
			// Reached the operator-supplied drain floor: stop here and anchor the
			// bubble's first parent T at this commit (it is NOT collected, so the
			// last pending commit's ParentSHA is this floor). This is the explicit
			// override that scopes a first-run backfill above the released backlog.
			reversePending(pending)
			return pending, nil
		}
		if len(c.ParentSHAs) >= 2 {
			// Reached the merge boundary T: collected commits are the pending
			// squashes. Reverse to oldest-first.
			reversePending(pending)
			return pending, nil
		}
		if len(c.ParentSHAs) == 0 {
			// A root commit is NOT a merge boundary. R-A requires a genuine merge
			// boundary to anchor the first bubble's first parent T; fail loud
			// rather than treating the root as T.
			return nil, fmt.Errorf(
				"drain-all: reached root commit %s with no merge boundary above it on %s; "+
					"refusing to bubble (R-A) — cannot anchor the bubble's first parent; investigate manually",
				c.SHA, b.branch,
			)
		}
		pending = append(pending, pendingCommit{
			squash: release.Squash{
				SHA:       c.SHA,
				ParentSHA: c.ParentSHAs[0],
				TreeSHA:   c.TreeSHA,
			},
			message: c.Message,
		})
		next, err := b.gh.Commit(ctx, b.repo, c.ParentSHAs[0])
		if err != nil {
			return nil, fmt.Errorf("drain-all walk: read commit %s: %w", c.ParentSHAs[0], err)
		}
		c = next
	}
	return nil, fmt.Errorf(
		"drain-all: examined the walk bound of %d first-parent commits from %s tip %s without reaching a merge boundary; "+
			"refusing to bubble (R-A) — develop has more than %d single-parent commits above any merge boundary; investigate manually (or pass --boundary <sha> to set an explicit drain floor)",
		b.walkBound, b.branch, tip.SHA, b.walkBound,
	)
}

// releaseTagNameRE recognises a release tag by name: the canonical `vMAJOR.MINOR…`
// form (v1.2.3, v1.2.3-rc4) and the legacy `pkg/schema/vMAJOR.MINOR…` published
// form. Non-release tags (arbitrary markers) are deliberately excluded so the
// first-run guard cannot false-fire on a normal post-install squash that merely
// happens to carry an unrelated tag.
var releaseTagNameRE = regexp.MustCompile(`^(?:pkg/schema/)?v\d+\.\d+`)

// isReleaseTagName reports whether a tag name is a release tag the first-run
// guard must protect against draining across.
func isReleaseTagName(name string) bool { return releaseTagNameRE.MatchString(name) }

// guardReleasedHistory refuses a first-run drain that would bubble merge-commits
// over ALREADY-RELEASED history (the ratified B11 "surface before the first
// protected-develop write" case).
//
// Mechanism — release-tag reachability, exact for the linear first-parent drain:
// the pending set spans the first-parent range [merge boundary, tip], so any
// release tag whose commit lies in that range appears AS one of the pending
// commits. Reachability therefore reduces to SHA membership — no ancestor walk
// is needed. The newest pending commit is the just-merged triggering squash
// (never released), so only the commits BELOW it are checked. A release tag
// below the boundary (already merged/bubbled) or the triggering squash itself
// never trips the guard, so a normal post-install stacked-squash run — every
// pending commit newer than every tag — proceeds untouched (contract (b)).
//
// On a hit it fails loud and leaves develop unchanged (contract (a)): the
// operator decides intentionally, and can re-run with --boundary <sha> set to
// the drain floor above the released backlog to proceed (contract (c)).
func (b *bubbler) guardReleasedHistory(ctx context.Context, pending []pendingCommit) error {
	if len(pending) < 2 {
		// Only the triggering squash (or nothing): no drained commit BELOW it to
		// cross released history.
		return nil
	}
	tags, err := b.gh.Tags(ctx, b.repo)
	if err != nil {
		return fmt.Errorf("bubble first-run guard: cannot list release tags on %s to check the drain does not cross released history: %w", b.repo, err)
	}
	released := make(map[string]string, len(tags)) // commit SHA -> release tag name
	for _, t := range tags {
		if t.CommitSHA != "" && isReleaseTagName(t.Name) {
			released[t.CommitSHA] = t.Name
		}
	}
	// pending is oldest-first; the last element is the triggering squash (the tip)
	// — skip it and check only the commits below it.
	for _, pc := range pending[:len(pending)-1] {
		if tag, ok := released[pc.squash.SHA]; ok {
			return fmt.Errorf(
				"bubble: refusing to drain across already-released history on %s: pending squash %s is the release tag %q. "+
					"Bubbling would insert merge commits over released commits and reshape the first-parent mainline. "+
					"develop is left unchanged (no write). If this is a deliberate first-run backfill, re-run with --boundary <sha> "+
					"set to the commit ABOVE the released backlog (the drain floor); otherwise investigate why released history is being drained",
				b.branch, pc.squash.SHA, tag,
			)
		}
	}
	return nil
}

// resolveProvenance resolves a squash's PR attribution: the PR number/title (via
// the "(#n)" squash-subject suffix, then PullRequests.Get for the authoritative
// title/body), the standing approvers + reviewers (via the PR's reviews), the
// co-authors (from the squash commit's trailers), and the closing issues (from
// the PR body + commit message).
//
// Provenance is BEST-EFFORT (cosmetic vs. the bubble itself): if the "(#n)" is
// missing OR the Pull/PullReviews lookups hard-error (e.g. the "(#n)" is an issue
// ref, a deleted/fork PR, or a coincidental "(#n)"), it LOGS a warning and
// degrades to the no-PR "Merge commit <sha>" provenance so the bubble still
// advances — it never returns an error and never wedges develop. Fail-loud is
// reserved for the merge / FF-CAS write path, not a trailer lookup.
func (b *bubbler) resolveProvenance(ctx context.Context, pc pendingCommit) (release.PRProvenance, error) {
	subject := firstLine(pc.message)

	n, ok := parseTrailingPRNumber(subject)
	if !ok {
		// No "(#n)" suffix: degenerate, bubble best-effort with the subject as the
		// title (number stays 0 -> "Merge commit <sha>" subject).
		b.logf("notice: squash %s has no (#n) PR suffix; bubbling without a PR number", pc.squash.SHA)
		return b.noPRProvenance(subject, pc), nil
	}

	pull, err := b.gh.Pull(ctx, b.repo, n)
	if err != nil {
		// Unresolvable "(#n)": degrade rather than wedge develop on every re-run.
		b.logf("warning: cannot resolve PR #%d for squash %s (%v); bubbling with the no-PR \"Merge commit <sha>\" message instead", n, pc.squash.SHA, err)
		return b.noPRProvenance(stripTrailingPRNumber(subject), pc), nil
	}
	// Prefer the authoritative PR title; fall back to the "(#n)"-stripped subject
	// when the PR title is empty/unresolved.
	title := stripTrailingPRNumber(subject)
	if t := strings.TrimSpace(pull.Title); t != "" {
		title = pull.Title
	}

	reviews, err := b.gh.PullReviews(ctx, b.repo, n)
	if err != nil {
		b.logf("warning: cannot resolve reviews for PR #%d (squash %s) (%v); bubbling with the no-PR \"Merge commit <sha>\" message instead", n, pc.squash.SHA, err)
		return b.noPRProvenance(title, pc), nil
	}
	approvedBy := release.LatestApprovers(reviews)

	return release.PRProvenance{
		Number:       n,
		Title:        title,
		ClosesIssues: parseClosesIssues(pull.Body + "\n" + pc.message),
		ApprovedBy:   approvedBy,
		// Exclude standing approvers from Reviewed-by so each person is attributed
		// once under their most-specific role (Approved-by wins over Reviewed-by).
		ReviewedBy:   reviewersExcludingApprovers(reviews, approvedBy),
		CoAuthoredBy: parseCoAuthors(pc.message),
	}, nil
}

// noPRProvenance builds the best-effort provenance for a squash with no resolved
// PR: Number 0 (renders "Merge commit <sha>" via renderBubbleMessage), the given
// title, and the trailers recoverable from the commit message alone (no PR body
// or reviews are available).
func (b *bubbler) noPRProvenance(title string, pc pendingCommit) release.PRProvenance {
	return release.PRProvenance{
		Number:       0,
		Title:        title,
		ClosesIssues: parseClosesIssues(pc.message),
		CoAuthoredBy: parseCoAuthors(pc.message),
	}
}

// bubble creates the merge-commit chain M_1..M_k and advances the branch to M_k.
//
// For items [S_1, ..., S_k] (oldest-first), with T = S_1.ParentSHA (the merge
// boundary the orchestrator walked back to):
//
//	M_1 = commit{tree: S_1.tree, parents: [T,       S_1]}
//	M_2 = commit{tree: S_2.tree, parents: [M_1,     S_2]}
//	...
//	M_k = commit{tree: S_k.tree, parents: [M_{k-1}, S_k]}
//
// so each M_i has first-parent the previous M (first-parent chain M_k->...->T)
// and M_i.tree == S_i.tree. The branch is then advanced T...->M_k via ONE
// fast-forward compare-and-swap (UpdateRefFastForward has no force option). A
// 422 non-FF rejection is surfaced (wrapping release.ErrNotFastForward) so run's
// bounded retry re-reads the tip.
func (b *bubbler) bubble(ctx context.Context, items []release.BubbleItem) error {
	if len(items) == 0 {
		return nil
	}

	// prev starts at the merge boundary T = the oldest squash's parent, so M_1's
	// first parent is T (NOT the current branch tip, which is the newest squash).
	prev := items[0].Squash.ParentSHA
	for _, item := range items {
		m, err := b.gh.CreateCommit(ctx, b.repo, release.NewCommit{
			Message:    item.Message,
			TreeSHA:    item.Squash.TreeSHA,
			ParentSHAs: []string{prev, item.Squash.SHA},
		})
		if err != nil {
			return fmt.Errorf("create bubble merge commit for squash %s: %w", item.Squash.SHA, err)
		}
		prev = m.SHA
	}

	// If the ref update below fails (non-FF or otherwise), the M commits created
	// above are unreferenced and dangle harmlessly — GitHub garbage-collects
	// unreachable objects. Nothing leaks and the branch is left untouched.
	if err := b.gh.UpdateRefFastForward(ctx, b.repo, b.ref, prev); err != nil {
		// UpdateRefFastForward already wraps release.ErrNotFastForward on a 422
		// non-FF; re-wrap with %w so run's errors.Is retry check still fires while
		// keeping the branch-advance context.
		return fmt.Errorf("advance %s to bubbled tip %s: %w", b.branch, prev, err)
	}
	return nil
}

// reversePending reverses in place (tip-first walk order -> oldest-first plan order).
func reversePending(p []pendingCommit) {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

// --- provenance parsing (pure, single-sourced with internal/release) -----------
//
// These parsers live in package main by design, NOT in internal/release: they
// belong to provenance RESOLUTION, which interleaves live GitHub I/O
// (Pull/PullReviews) and is inherently part of the orchestrator's composition
// root. internal/release owns only the pure RENDER grammar (AssembleMessage /
// BubbleMergeMessage) that turns a resolved PRProvenance into the final message.
// Keeping resolution here and grammar there keeps the render layer free of any
// I/O and free of the squash-subject/trailer parsing quirks.

var (
	// trailingPRNumberRE matches GitHub's native squash subject convention
	// "<title> (#123)" — the PR-number source for per-squash provenance.
	trailingPRNumberRE = regexp.MustCompile(`\(#(\d+)\)\s*$`)

	// coAuthorRE matches a "Co-authored-by: Name <email>" trailer line (the
	// GitHub squash convention) anywhere in the commit message.
	coAuthorRE = regexp.MustCompile(`(?im)^\s*co-authored-by:\s*(.+?)\s*$`)

	// closesIssueRE matches a GitHub issue-closing keyword + "#<n>"
	// (close/closes/closed, fix/fixes/fixed, resolve/resolves/resolved).
	closesIssueRE = regexp.MustCompile(`(?i)\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s+#(\d+)\b`)
)

// renderBubbleMessage renders a bubble merge-commit message. With a resolved PR
// (Number > 0) it delegates to the pure release.BubbleMergeMessage
// ("Merge PR #<n>: <title>" + trailers). Without a PR — no "(#n)" subject
// suffix, a rare post-merge edge — it avoids the misleading "Merge PR #0" by
// using a "Merge commit <short-sha>" subject, then reuses release.AssembleMessage
// for the SAME trailer grammar (so Closes / Co-authored-by formatting is
// single-sourced, not re-implemented; there are no approvers/reviewers to
// attribute without a PR).
func renderBubbleMessage(p release.PRProvenance, squashSHA string) string {
	if p.Number > 0 {
		return release.BubbleMergeMessage(p)
	}

	subject := "Merge commit " + shortSHA(squashSHA)
	if title := strings.TrimSpace(p.Title); title != "" {
		subject += ": " + title
	}
	return release.AssembleMessage(subject, p)
}

// shortSHA abbreviates a commit sha to 12 chars (git's default short length).
func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// firstLine returns the first line of a commit message (its subject), trimmed.
func firstLine(message string) string {
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		return strings.TrimSpace(message[:i])
	}
	return strings.TrimSpace(message)
}

// parseTrailingPRNumber extracts the "(#123)" PR number from a squash subject,
// returning (number, true) on a match.
func parseTrailingPRNumber(subject string) (int, bool) {
	m := trailingPRNumberRE.FindStringSubmatch(subject)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// stripTrailingPRNumber removes a trailing "(#123)" suffix from a subject so the
// fallback title does not duplicate the number already in the "Merge PR #n:"
// subject rendered by release.BubbleMergeMessage.
func stripTrailingPRNumber(subject string) string {
	return strings.TrimSpace(trailingPRNumberRE.ReplaceAllString(subject, ""))
}

// parseCoAuthors extracts the values of every "Co-authored-by:" trailer in a
// commit message, in order (each already "Name <email>").
func parseCoAuthors(message string) []string {
	matches := coAuthorRE.FindAllStringSubmatch(message, -1)
	if matches == nil {
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if v := strings.TrimSpace(m[1]); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// parseClosesIssues extracts issue numbers referenced by a closing keyword
// (Closes/Fixes/Resolves #n) in the given text, in order, de-duplicated.
func parseClosesIssues(text string) []int {
	matches := closesIssueRE.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return nil
	}
	seen := make(map[int]bool)
	var out []int
	for _, m := range matches {
		n, err := strconv.Atoi(m[1])
		if err != nil || n <= 0 || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return out
}

// distinctReviewers returns the distinct, non-empty reviewer logins across a
// PR's reviews, in first-seen order. Unlike release.LatestApprovers (which keeps
// only standing approvals) this records everyone who submitted any review.
func distinctReviewers(reviews []release.Review) []string {
	seen := make(map[string]bool)
	var out []string
	for _, r := range reviews {
		if r.User == nil || r.User.Login == "" || seen[r.User.Login] {
			continue
		}
		seen[r.User.Login] = true
		out = append(out, r.User.Login)
	}
	return out
}

// reviewersExcludingApprovers returns the distinct reviewers with the standing
// approvers removed, so a person who both reviewed and approved is attributed
// once — under Approved-by (the most-specific role), not also Reviewed-by.
func reviewersExcludingApprovers(reviews []release.Review, approvers []string) []string {
	approved := make(map[string]bool, len(approvers))
	for _, a := range approvers {
		approved[a] = true
	}
	out := make([]string, 0)
	for _, login := range distinctReviewers(reviews) {
		if !approved[login] {
			out = append(out, login)
		}
	}
	return out
}

// --- CLI entry points ----------------------------------------------------------

// runBubble is the `release-guard bubble` dispatch handler, mirroring
// runCheckFinal's injection (gh + repo from the composition root). It adapts the
// process stdio and delegates to bubbleRun (the writer-injectable seam).
func runBubble(ctx context.Context, gh GitHubClient, repo string, args []string) error {
	return bubbleRun(ctx, gh, repo, os.Stdout, os.Stderr, args)
}

// bubbleRun parses flags, builds the bubbler over the injected GitHubClient, runs
// the bubble, and returns a non-nil error on any failure. User-facing output is
// written to the injected writers: progress/diagnostics to stderr, the terminal
// ::error:: annotation to stdout (so it surfaces as a GitHub Actions run
// annotation). Keeping writers injected makes the assembly + fail-loud path
// unit-testable without a subprocess or capturing os.Stdout.
//
// Unlike the retired stdlib version this seam does NOT read the environment: the
// composition root (main -> mustGitHubClient/mustRepo) already resolved the token
// and repo and injected gh + repo, so bubbleRun is a pure function of its args.
func bubbleRun(ctx context.Context, gh GitHubClient, repo string, stdout, stderr io.Writer, args []string) error {
	var (
		prFlag      string
		boundary    string
		maxAttempts = defaultBubbleMaxAttempts
	)
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--pr":
			i++
			if i >= len(args) {
				return errfTo(stderr, "bubble: --pr requires a value")
			}
			prFlag = args[i]
		case "--max-attempts":
			i++
			if i >= len(args) {
				return errfTo(stderr, "bubble: --max-attempts requires a value")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				return errfTo(stderr, "bubble: --max-attempts must be a positive integer, got %q", args[i])
			}
			maxAttempts = n
		case "--boundary":
			i++
			if i >= len(args) {
				return errfTo(stderr, "bubble: --boundary requires a value")
			}
			boundary = args[i]
		default:
			return errfTo(stderr, "bubble: unknown flag %q", args[i])
		}
	}

	// --pr is informational trigger-context only: bubbling is tip-driven
	// drain-all, not PR-filtered (each squash carries its own "(#n)"). Validate it
	// as a positive int (consistency with --max-attempts) when provided, then use
	// it only for the log line.
	if prFlag != "" {
		if n, err := strconv.Atoi(prFlag); err != nil || n < 1 {
			return errfTo(stderr, "bubble: --pr must be a positive integer, got %q", prFlag)
		}
	}

	const branch = "develop"
	b := &bubbler{
		gh:          gh,
		repo:        repo,
		branch:      branch,
		ref:         "heads/" + branch,
		maxAttempts: maxAttempts,
		walkBound:   defaultWalkBound,
		boundary:    boundary,
		logf: func(format string, a ...any) {
			fmt.Fprintf(stderr, "release-guard bubble: "+format+"\n", a...)
		},
	}

	if prFlag != "" {
		fmt.Fprintf(stderr, "release-guard bubble: triggered by PR #%s on %s\n", prFlag, repo)
	}

	if err := b.run(ctx); err != nil {
		// Fail loud: a GitHub Actions ::error:: on stdout (so it surfaces in the
		// run annotation) plus the returned error (mapped to a non-zero exit by
		// main). develop is left unchanged.
		fmt.Fprintf(stdout, "::error::release-guard bubble failed: %v\n", err)
		fmt.Fprintf(stderr, "release-guard: bubble: %v\n", err)
		return err
	}
	fmt.Fprintf(stdout, "bubble complete on %s\n", repo)
	return nil
}

// errfTo writes an actionable diagnostic to the given writer and returns it as
// an error (the bubble seam's writer-injected analogue of the package fatalf).
func errfTo(stderr io.Writer, format string, a ...any) error {
	err := fmt.Errorf(format, a...)
	fmt.Fprintf(stderr, "release-guard: %v\n", err)
	return err
}
