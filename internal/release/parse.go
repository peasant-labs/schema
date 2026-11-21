package release

import (
	"fmt"
)

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
