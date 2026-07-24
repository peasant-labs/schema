package release

import (
	"fmt"
	"strings"
)

// BubbleDecision is the typed outcome of evaluating whether (and how) a
// freshly squash-merged commit should be "bubbled" into the develop history as
// a merge-commit triangle (T -> M, parents [T, S], M.tree == S.tree).
//
// It is a strongly-typed enum (per the repo's no-stringly-typed-API rule) so
// the orchestrator branches on named constants and never compares bare strings.
// All safety-relevant branching lives in DecideBubble — never in workflow bash.
type BubbleDecision int

const (
	// BubbleProceed: develop's tip is exactly the squash S (a single-parent
	// commit at S.tree). The bubble merge-commit M should be built and the ref
	// fast-forwarded S -> M.
	BubbleProceed BubbleDecision = iota
	// BubbleSkipNotSquash: the target commit is not a single-parent squash
	// (S.parentCount != 1), so it is not an eligible squash-merge. Skip, exit 0.
	BubbleSkipNotSquash
	// BubbleSkipAlreadyBubbled: develop's tip is already the bubble merge-commit
	// for S (a two-parent merge whose tree equals S.tree, distinct from S).
	// Re-running on an already-bubbled tip is a no-op. Skip, exit 0.
	BubbleSkipAlreadyBubbled
	// BubbleRetryTipAdvanced: develop's tip has advanced to some other commit
	// (its tree differs from S.tree and it is not S), so the read is stale. The
	// orchestrator must re-read the tip and re-decide (bounded retry).
	BubbleRetryTipAdvanced
)

// String renders the decision for CLI output, logs, and error messages.
func (d BubbleDecision) String() string {
	switch d {
	case BubbleProceed:
		return "proceed"
	case BubbleSkipNotSquash:
		return "skip-not-squash"
	case BubbleSkipAlreadyBubbled:
		return "skip-already-bubbled"
	case BubbleRetryTipAdvanced:
		return "retry-tip-advanced"
	default:
		return fmt.Sprintf("BubbleDecision(%d)", int(d))
	}
}

// BubbleFacts are the side-effect-free topology facts the orchestrator reads
// from the GitHub Git Data API (the squash S and develop's current tip) and
// feeds to DecideBubble. Keeping the decision a pure function of these facts is
// what makes every branch (B12 precedence) unit-testable BEFORE any protected
// branch is written.
type BubbleFacts struct {
	// SParentCount is the number of parents of the target squash commit S. A
	// native squash-merge produces a single-parent commit, so the only eligible
	// value is 1; anything else means S is not a squash (B4 / R-B).
	SParentCount int
	// TipParentCount is the number of parents of develop's current tip. A bubble
	// merge-commit M has two parents [T, S]; this distinguishes a genuine bubble
	// tip from an unrelated same-tree commit when classifying AlreadyBubbled.
	TipParentCount int
	// TipTreeEqualsS reports whether develop's tip has the same tree as S. A
	// bubble M is built with M.tree == S.tree, so a two-parent tip at S.tree is
	// the already-bubbled state.
	TipTreeEqualsS bool
	// TipEqualsS reports whether develop's tip is the squash S itself (the
	// just-fast-forwarded, ready-to-bubble state).
	TipEqualsS bool
}

// BubbleFactsFromCommits derives the pure BubbleFacts that DecideBubble consumes
// from the SLICE-1 git-data own-types: a candidate squash commit S and develop's
// current tip commit. It is the mechanical realignment from the GitHub Git Data
// projection (release.GitCommit) onto the decision's topology facts, so callers
// resolve facts from own-types instead of hand-assembling a BubbleFacts literal.
//
// It is a PURE projection of the two commits. Choosing WHICH commit is the
// candidate S for a given tip (e.g. a two-parent merge's second parent) and any
// safety reclassification of a stable non-bubble merge tip remain the
// orchestrator's responsibility — this function only maps (S, tip) -> facts.
func BubbleFactsFromCommits(squash, tip GitCommit) BubbleFacts {
	return BubbleFacts{
		SParentCount:   len(squash.ParentSHAs),
		TipParentCount: len(tip.ParentSHAs),
		TipTreeEqualsS: tip.TreeSHA == squash.TreeSHA,
		TipEqualsS:     tip.SHA == squash.SHA,
	}
}

// DecideBubble maps topology facts to a typed decision with the PINNED
// precedence:
//
//	NotSquash(SParentCount != 1) > AlreadyBubbled > TipAdvanced > Proceed
//
// Precedence only matters where facts overlap; the canonical overlaps are:
//   - NotSquash dominates everything: a non-single-parent S is never bubbled,
//     regardless of what the tip looks like.
//   - AlreadyBubbled vs TipAdvanced: both have tip != S. A tip that is a genuine
//     two-parent merge at S.tree is the already-built bubble (no-op), NOT a
//     "tip advanced" retry (B12).
func DecideBubble(f BubbleFacts) BubbleDecision {
	switch {
	case f.SParentCount != 1:
		return BubbleSkipNotSquash
	// Exactly 2 parents is the bubble shape M = merge[T, S]; a >2-parent octopus
	// at S.tree is NOT a bubble and correctly falls through to RetryTipAdvanced
	// (bounded retry -> fail-loud), never a silent Proceed. Do not relax to >=2.
	case f.TipParentCount == 2 && f.TipTreeEqualsS && !f.TipEqualsS:
		return BubbleSkipAlreadyBubbled
	case f.TipEqualsS:
		return BubbleProceed
	default:
		return BubbleRetryTipAdvanced
	}
}

// Squash identifies a single-parent squash-merge commit S that is a candidate
// for bubbling. TreeSHA is set explicitly on the bubble merge-commit M so that
// M.tree == S.tree without a content merge (the conflict-free, signed-M path).
type Squash struct {
	SHA       string
	ParentSHA string
	TreeSHA   string
}

// BubbleItem is a fully pre-resolved unit of work for the Bubbler: the squash to
// bubble plus its rendered merge-commit Message. The orchestrator resolves
// provenance and renders Message via BubbleMergeMessage so that the Bubbler
// (the I/O layer) carries no provenance logic and creates exactly one M per
// item in a drain-all chain, each M carrying its own PR's trailers.
type BubbleItem struct {
	Squash  Squash
	Message string
}

// PRProvenance is the per-PR attribution resolved by the orchestrator from the
// GitHub API and rendered into a bubble merge-commit's trailers by
// BubbleMergeMessage. It carries the data for the subject and the
// Closes/Approved-by/Reviewed-by/Co-authored-by trailers — NOT the PR body
// (the squash commit already carries the title + body).
type PRProvenance struct {
	// Number is the pull-request number for the "Merge PR #<n>" subject.
	Number int
	// Title is the PR title; when empty the subject falls back to "Merge PR #<n>".
	Title string
	// ClosesIssues renders one "Closes #<issue>" trailer per entry, in order.
	ClosesIssues []int
	// ApprovedBy renders one "Approved-by: <value>" trailer per entry, in order.
	// Sourced from LatestApprovers over the PR's reviews.
	ApprovedBy []string
	// ReviewedBy renders one "Reviewed-by: <value>" trailer per entry, in order.
	ReviewedBy []string
	// CoAuthoredBy renders one "Co-authored-by: <value>" trailer per entry; each
	// value is already in "Name <email>" form.
	CoAuthoredBy []string
}

// BubbleMergeMessage renders the merge-commit message for a bubble M from the
// resolved PR provenance: a single-line subject, then (if any trailers exist) a
// blank line and the git trailers in fixed order — Closes, Approved-by,
// Reviewed-by, Co-authored-by. It is deliberately NOT the PR body; the squash
// commit already carries the title + body, so M records only the merge
// provenance.
//
// Both the subject and every trailer value are forced single-line (any CR/LF/tab
// is collapsed to a space) so neither a multi-line title nor a multi-line trailer
// value can break the subject/trailer structure or inject a forged trailer. When
// the title is empty/unresolved the subject falls back to "Merge PR #<n>".
func BubbleMergeMessage(p PRProvenance) string {
	title := strings.TrimSpace(flattenLineBreaks(p.Title))

	var subject string
	if title == "" {
		subject = fmt.Sprintf("Merge PR #%d", p.Number)
	} else {
		subject = fmt.Sprintf("Merge PR #%d: %s", p.Number, title)
	}

	return AssembleMessage(subject, p)
}

// AssembleMessage joins a caller-provided subject with the provenance's git
// trailers in the fixed order Closes / Approved-by / Reviewed-by /
// Co-authored-by (a blank line separates the subject from the trailer block; a
// subject with no trailers is returned as-is). Every trailer value is
// line-flattened so it cannot inject extra lines or a forged trailer.
//
// It is the SINGLE SOURCE of the bubble trailer grammar: BubbleMergeMessage uses
// it (with the canonical "Merge PR #<n>: <title>" subject), and the bubble
// orchestrator's no-PR fallback reuses it (with its own "Merge commit <sha>"
// subject) so the trailer format is never re-implemented across the boundary.
func AssembleMessage(subject string, p PRProvenance) string {
	trailers := messageTrailers(p)
	if len(trailers) == 0 {
		return subject
	}
	return subject + "\n\n" + strings.Join(trailers, "\n")
}

// messageTrailers builds the ordered, line-flattened trailer lines for a bubble
// merge message.
func messageTrailers(p PRProvenance) []string {
	trailers := make([]string, 0, len(p.ClosesIssues)+len(p.ApprovedBy)+len(p.ReviewedBy)+len(p.CoAuthoredBy))
	for _, issue := range p.ClosesIssues {
		trailers = append(trailers, fmt.Sprintf("Closes #%d", issue))
	}
	for _, approver := range p.ApprovedBy {
		trailers = append(trailers, "Approved-by: "+flattenLineBreaks(approver))
	}
	for _, reviewer := range p.ReviewedBy {
		trailers = append(trailers, "Reviewed-by: "+flattenLineBreaks(reviewer))
	}
	for _, coAuthor := range p.CoAuthoredBy {
		trailers = append(trailers, "Co-authored-by: "+flattenLineBreaks(coAuthor))
	}
	return trailers
}

// flattenLineBreaks replaces every CR/LF/tab run with a single space so a title
// or trailer value can never inject extra lines into the subject or fake a
// trailer block.
func flattenLineBreaks(s string) string {
	return strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t'
	}), " ")
}
