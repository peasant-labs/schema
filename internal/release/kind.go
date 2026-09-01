// Package release holds the pure, table-driven-testable logic that the release
// CI workflows depend on: the release-PR title grammar, the git-tag grammar, the
// typed release-kind derivation, and release workflow policy checks.
//
// Everything here is deliberately side-effect-free so the negative paths (a
// malformed title, a non-release tag, or a broken workflow graph) are
// unit-testable BEFORE they run in a workflow. The thin CLI in
// cmd/release-guard wires these functions to real workflow files; the workflows
// shell out to that CLI.
package release

// ReleaseKind classifies a release reference (a PR title or a git tag) as a
// release candidate, a final release, or an invalid/unrecognized reference.
//
// It is a strongly-typed enum (per the repo's no-stringly-typed-API rule) so
// that workflow steps and Go callers compare against named constants rather than
// bare strings.
type ReleaseKind string

const (
	// KindInvalid is returned together with an error whenever a reference does
	// not parse as a release.
	KindInvalid ReleaseKind = "invalid"
	// KindRC is a release candidate: a version carrying an -rcN prerelease
	// suffix (e.g. v0.1.0-rc1). RCs publish as prereleases and use the npm
	// `next` dist-tag.
	KindRC ReleaseKind = "rc"
	// KindFinal is a final, non-prerelease version (e.g. v0.1.0).
	KindFinal ReleaseKind = "final"
)

// String renders the kind for CLI output and workflow consumption.
func (k ReleaseKind) String() string { return string(k) }
