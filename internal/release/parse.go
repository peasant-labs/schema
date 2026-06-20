package release

import (
	"fmt"
	"regexp"
)

// titlePattern matches a release-PR title: "release(<version>): <subject>".
//
// The captured group is the version, which is then re-validated through
// NewVersion so the title grammar and the version grammar cannot drift apart.
// The trailing ": " (colon + single space) separator is REQUIRED, as is a
// non-empty subject implied by it. This mirrors bestiary's tag-on-release-merge
// discipline (conventional-commit scope carries the version).
var titlePattern = regexp.MustCompile(`^release\((v\d+\.\d+\.\d+(?:-rc\d+)?)\): `)

// ParseReleaseTitle parses a release-PR title of the form
// "release(vX.Y.Z): subject" or "release(vX.Y.Z-rcN): subject" and returns the
// validated version and its kind.
//
// On any grammar violation it returns (KindInvalid, error) with an actionable
// message describing the expected shape and a concrete example. This is the
// single implementation of the title grammar; the workflows must NOT re-encode
// it inline.
func ParseReleaseTitle(title string) (Version, ReleaseKind, error) {
	m := titlePattern.FindStringSubmatch(title)
	if m == nil {
		return "", KindInvalid, fmt.Errorf(
			"invalid release PR title %q: expected \"release(vX.Y.Z): subject\" or \"release(vX.Y.Z-rcN): subject\" (e.g. \"release(v0.1.0-rc1): first release candidate\"); the version must carry a leading \"v\" and the \": \" separator is required",
			title,
		)
	}
	v, err := NewVersion(m[1])
	if err != nil {
		return "", KindInvalid, fmt.Errorf("release PR title %q has an invalid version: %w", title, err)
	}
	return v, v.Kind(), nil
}

// ParseTag parses a git tag as a schema release reference. It accepts ONLY bare
// "vX.Y.Z[-rcN]" tags; namespaced tags such as the legacy "pkg/schema/v1.2.3"
// (retained from when this module was nested in peasant) are rejected because
// they are not releases of THIS module and must never trigger the release
// pipeline. (The release.yml trigger filter `v*` already excludes pkg/schema/v*
// by name; this parse is the defense-in-depth guard inside the workflow.)
func ParseTag(tag string) (Version, ReleaseKind, error) {
	v, err := NewVersion(tag)
	if err != nil {
		return "", KindInvalid, fmt.Errorf("tag %q is not a release tag (release tags look like v0.1.0 or v0.1.0-rc1): %w", tag, err)
	}
	return v, v.Kind(), nil
}
