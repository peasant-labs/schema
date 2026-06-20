package release

import (
	"fmt"
	"regexp"
)

// Version is a validated schema release version, always including the leading
// "v" (e.g. "v0.1.0" or "v0.1.0-rc1"). Construct only via NewVersion so the
// invariant (matches the frozen grammar) holds everywhere a Version is seen.
type Version string

// versionPattern matches a schema release version: vMAJOR.MINOR.PATCH with an
// optional -rcN prerelease suffix. It is fully anchored, and the rc suffix
// REQUIRES at least one digit — so "v0.1.0-rc" (no number) is rejected, which is
// one of the must-fail grammar cases.
var versionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+(-rc\d+)?$`)

// rcSuffixPattern matches a trailing -rcN prerelease suffix.
var rcSuffixPattern = regexp.MustCompile(`-rc\d+$`)

// NewVersion validates raw against the frozen release-version grammar and
// returns a typed Version. On failure it returns an actionable error naming the
// expected shape and a concrete example.
func NewVersion(raw string) (Version, error) {
	if !versionPattern.MatchString(raw) {
		return "", fmt.Errorf(
			"invalid release version %q: must match vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-rcN (e.g. v0.1.0 or v0.1.0-rc1)",
			raw,
		)
	}
	return Version(raw), nil
}

// String renders the version (with its leading "v").
func (v Version) String() string { return string(v) }

// IsRC reports whether the version carries an -rcN prerelease suffix.
func (v Version) IsRC() bool { return rcSuffixPattern.MatchString(string(v)) }

// Base returns the version with any -rcN suffix stripped: the final version a
// release candidate is a candidate FOR. For a final version Base is the identity
// (v0.1.0 -> v0.1.0); for an rc it strips the suffix (v0.1.0-rc3 -> v0.1.0).
func (v Version) Base() Version {
	return Version(rcSuffixPattern.ReplaceAllString(string(v), ""))
}

// Kind returns KindRC for prerelease (-rcN) versions and KindFinal otherwise.
// A Version is, by construction, never KindInvalid.
func (v Version) Kind() ReleaseKind {
	if v.IsRC() {
		return KindRC
	}
	return KindFinal
}
