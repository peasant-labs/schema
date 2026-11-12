package intent

import (
	"path"
	"strconv"
)

// ProtectedChange is one changed path and, for a version file, whether the
// protected field moved. The command layer reads the field at both ends of
// the change from git; this predicate only judges the answer.
type ProtectedChange struct {
	Path                  string `yaml:"path"`
	ProtectedFieldChanged bool   `yaml:"protected_field_changed"`
}

// CheckProtectedPaths rejects a non-release change to a protected path or to
// a protected version field. The release PR may change anything; every other
// PR, including any other bot PR, may not. The config always comes from the
// base revision, so a PR's own edit to the config cannot unprotect its path.
func CheckProtectedPaths(cfg Config, changes []ProtectedChange, isReleasePR bool) error {
	if isReleasePR {
		return nil
	}
	for _, change := range changes {
		if change.Path == cfg.Changelog {
			return protectedError(change.Path, "the changelog is written only by the release PR")
		}
		for _, glob := range cfg.ProtectedPaths {
			if change.Path == glob {
				return protectedError(change.Path, "the path is listed in protected_paths and only the release PR may change it")
			}
			if matched, err := path.Match(glob, change.Path); err == nil && matched {
				return protectedError(change.Path, "the path matches protected_paths entry "+strconv.Quote(glob)+" and only the release PR may change it")
			}
		}
		for _, vf := range cfg.VersionFiles {
			if change.Path == vf.Path && change.ProtectedFieldChanged {
				return protectedError(change.Path, "the protected field "+strconv.Quote(vf.Field)+" is written only by the release PR")
			}
		}
	}
	return nil
}

// protectedError builds the typed protected-path rejection for one path.
func protectedError(changePath, why string) *Error {
	return &Error{
		Code:  ErrProtectedPath,
		What:  "protected path " + strconv.Quote(changePath) + " changed outside the release PR",
		Why:   why,
		Where: "PR change, judged by intent.CheckProtectedPaths against the base release config",
		When:  "checking a PR before it can be enqueued",
		Means: "the title-guard check is red, so the merge queue will not accept this PR",
		Fix:   "drop the protected edit from the PR; the release PR writes that path when the batch lands",
	}
}
