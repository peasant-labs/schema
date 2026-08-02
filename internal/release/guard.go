package release

import (
	"fmt"
	"strings"
)

// RCStatus is the observed state of one release-candidate tag relative to a
// pending final release: whether its release.yml run was green, and whether the
// rc tag is an ancestor of the final release commit (git merge-base
// --is-ancestor). The CLI populates these from real git/gh queries; CheckFinal
// consumes them as pure data.
type RCStatus struct {
	// Tag is the release-candidate tag, e.g. v0.1.0-rc1.
	Tag Version `yaml:"tag"`
	// RunGreen is true when the rc tag's release.yml run completed successfully.
	RunGreen bool `yaml:"runGreen"`
	// IsAncestor is true when the rc tag's commit is an ancestor of the final
	// release commit (the final was tagged on a descendant of the proven rc).
	IsAncestor bool `yaml:"isAncestor"`
}

// FinalPolicy controls the one narrow exception to the normal final-release
// rule. An empty InitialFinal keeps the ordinary same-version ancestor-rc
// requirement. A non-empty value permits only that exact final version, and
// only while the repository has no prior product releases.
type FinalPolicy struct {
	InitialFinal Version
}

// FinalEvidence is the repository evidence used to decide whether a final may
// proceed. PriorReleases excludes the tag currently being checked.
type FinalEvidence struct {
	RCs           []RCStatus
	PriorReleases []Version
}

// CheckFinal decides whether a FINAL release may proceed.
//
// A final release is permitted only if at least one same-base-version release
// candidate exists that is BOTH (a) green — its release.yml run succeeded — AND
// (b) an ancestor of the final commit. This enforces the "final ⇒ proven rc on
// the same commit lineage" invariant inside the pipeline rather than by
// convention. rc-to-final tree identity is impossible (they are different merge
// commits by construction), so ancestry is the strongest cheap invariant.
//
// CheckFinal is pure: the caller supplies the observed rc statuses. It returns
// nil to proceed, or an actionable error naming exactly which condition failed
// for every candidate considered.
func CheckFinal(final Version, evidence FinalEvidence, policy FinalPolicy) error {
	if final.IsRC() {
		return fmt.Errorf(
			"CheckFinal called with a non-final version %q: release candidates do not require a predecessor rc — call CheckFinal only for final (non-rc) releases",
			final,
		)
	}
	if policy.InitialFinal != "" {
		if policy.InitialFinal.Kind() != KindFinal {
			return fmt.Errorf("initial-final policy is invalid: configured version %q is not an exact final release tag. Configure one version in vX.Y.Z form, without an rc suffix", policy.InitialFinal)
		}
		if len(evidence.PriorReleases) == 0 {
			if final != policy.InitialFinal {
				return fmt.Errorf("final release %s is blocked by the initial-final policy: this fresh repository may bootstrap only the exact configured final %s. Request %s, or remove the policy and cut a green same-version ancestor rc", final, policy.InitialFinal, policy.InitialFinal)
			}
			return nil
		}
	}

	base := final.Base()
	var sameBase []RCStatus
	for _, rc := range evidence.RCs {
		if rc.Tag.Base() == base {
			sameBase = append(sameBase, rc)
		}
	}

	if len(sameBase) == 0 {
		return fmt.Errorf(
			"final release %s is blocked: no same-version release candidate (%s-rcN) was found. Cut and merge a release(%s-rc1) PR, let its release run go green, then tag the final %s on a descendant commit",
			final, base, base, final,
		)
	}

	for _, rc := range sameBase {
		if rc.RunGreen && rc.IsAncestor {
			return nil
		}
	}

	// No candidate satisfied both conditions — explain why each was rejected so
	// the maintainer can fix it (per C-actionable-errors: what/why/how-to-fix).
	reasons := make([]string, 0, len(sameBase))
	for _, rc := range sameBase {
		switch {
		case !rc.RunGreen && !rc.IsAncestor:
			reasons = append(reasons, fmt.Sprintf("%s: its release run is not green AND it is not an ancestor of the final commit", rc.Tag))
		case !rc.RunGreen:
			reasons = append(reasons, fmt.Sprintf("%s: its release run is not green (re-run release.yml on the rc tag until it succeeds)", rc.Tag))
		case !rc.IsAncestor:
			reasons = append(reasons, fmt.Sprintf("%s: it is not an ancestor of the final commit (tag the final on a commit that descends from the proven rc)", rc.Tag))
		}
	}
	return fmt.Errorf(
		"final release %s is blocked: no same-version rc is both green and an ancestor of the final commit:\n  - %s",
		final, strings.Join(reasons, "\n  - "),
	)
}
