package intent

import (
	"fmt"
	"regexp"
	"strconv"
)

// semVerPattern matches a bare X.Y.Z value: three dot-separated decimal
// numbers and nothing else. No leading "v", no prerelease suffix: titles
// never mint those, and tags carry them outside this type.
var semVerPattern = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

// ParseSemVer validates raw as a bare X.Y.Z value and returns it typed.
func ParseSemVer(raw string) (SemVer, error) {
	if _, _, _, err := splitSemVer(SemVer(raw)); err != nil {
		return "", inputError(
			"invalid version "+strconv.Quote(raw),
			"a version is three dot-separated decimal numbers with no prefix or suffix",
			"intent.ParseSemVer",
			"parsing a version value",
			"the caller has no version to compute from",
			"pass a value like \"0.1.0\" (tags add the repo prefix outside this type)",
		)
	}
	return SemVer(raw), nil
}

// splitSemVer breaks a validated SemVer into its three numbers.
func splitSemVer(s SemVer) (major, minor, patch int, err error) {
	m := semVerPattern.FindStringSubmatch(string(s))
	if m == nil {
		return 0, 0, 0, fmt.Errorf("invalid SemVer %q", string(s))
	}
	numbers := make([]int, 3)
	for i, part := range m[1:] {
		n, convErr := strconv.Atoi(part)
		if convErr != nil {
			return 0, 0, 0, fmt.Errorf("invalid SemVer %q: %w", string(s), convErr)
		}
		numbers[i] = n
	}
	return numbers[0], numbers[1], numbers[2], nil
}

// semVerOf renders three numbers as a bare X.Y.Z value.
func semVerOf(major, minor, patch int) SemVer {
	return SemVer(strconv.Itoa(major) + "." + strconv.Itoa(minor) + "." + strconv.Itoa(patch))
}

// NextVersion computes the next version from a base, the highest pending
// level, and whether any pending subject is breaking.
//
// Below 1.0 a break advances the middle number and is listed as breaking in
// the notes; from 1.0 a break advances the first number. No input mints
// 1.0.0: graduation to 1.0 is a separate maintainer decision, so a major
// level below 1.0 advances the middle number instead. LevelNone returns the
// base unchanged; the planner only calls this when something is pending.
func NextVersion(base SemVer, level Level, breaking bool) (SemVer, error) {
	major, minor, patch, err := splitSemVer(base)
	if err != nil {
		return "", inputError(
			"invalid base version "+strconv.Quote(string(base)),
			"the base is the highest final release tag and must be a bare X.Y.Z value",
			"intent.NextVersion",
			"computing the next version",
			"no next version can be computed from this base",
			"pass the version base the planner derived from the repository tags",
		)
	}
	if breaking {
		if major == 0 {
			return semVerOf(0, minor+1, 0), nil
		}
		return semVerOf(major+1, 0, 0), nil
	}
	switch level {
	case LevelPatch:
		return semVerOf(major, minor, patch+1), nil
	case LevelMinor:
		return semVerOf(major, minor+1, 0), nil
	case LevelMajor:
		if major == 0 {
			return semVerOf(0, minor+1, 0), nil
		}
		return semVerOf(major+1, 0, 0), nil
	default:
		return base, nil
	}
}

// compareSemVer orders two validated versions: -1 when a is older, 0 when
// equal, 1 when a is newer. Callers pass values already validated through
// ParseSemVer; an invalid value is an error, never a guess.
func compareSemVer(a, b SemVer) (int, error) {
	amaj, amin, apat, err := splitSemVer(a)
	if err != nil {
		return 0, err
	}
	bmaj, bmin, bpat, err := splitSemVer(b)
	if err != nil {
		return 0, err
	}
	for _, pair := range [][2]int{{amaj, bmaj}, {amin, bmin}, {apat, bpat}} {
		if pair[0] < pair[1] {
			return -1, nil
		}
		if pair[0] > pair[1] {
			return 1, nil
		}
	}
	return 0, nil
}
