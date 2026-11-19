// Package release holds the tag grammar and typed release-kind derivation.
package release

// ReleaseKind classifies a release reference (a git tag) as a
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
