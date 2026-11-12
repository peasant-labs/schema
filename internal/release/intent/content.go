package intent

import (
	"strconv"
)

// This file holds the one release-content comparison: whether file-backed
// content matches what the generator produces now. The release-pr check and
// the freshness check both call it, so a stale tree gets the same verdict
// and the same code everywhere.

// CompareContent reports whether got matches the generated want. Only the
// file-backed content compares: the changelog text and one value per version
// file path. The section is rendering detail of the changelog, not separate
// file content. A mismatch is the typed stale-content rejection; equality
// is nil.
func CompareContent(want, got Content) error {
	if want.Changelog != got.Changelog {
		return staleContent("the changelog text differs from what the generator produces now")
	}
	if len(want.Versions) != len(got.Versions) {
		return staleContent("the version file set differs from what the generator produces now")
	}
	for path, wantVersion := range want.Versions {
		gotVersion, ok := got.Versions[path]
		if !ok {
			return staleContent("version file " + strconv.Quote(path) + " is missing from the compared content")
		}
		if gotVersion != wantVersion {
			return staleContent("version file " + strconv.Quote(path) + " holds " +
				strconv.Quote(gotVersion) + " instead of " + strconv.Quote(wantVersion))
		}
	}
	return nil
}

// staleContent builds the typed stale-content rejection for one difference.
func staleContent(detail string) *Error {
	return &Error{
		Code:  ErrReleaseContentStale,
		What:  "release content is stale: " + detail,
		Why:   "the release PR must match what the generator produces for the branch tip",
		Where: "release content, compared by intent.CompareContent",
		When:  "checking a release PR or a landed candidate against the generator",
		Means: "a stale release PR fails its check and a stale candidate is never tagged",
		Fix:   "rebuild the release PR from the branch tip; never hand-edit generated content",
	}
}
