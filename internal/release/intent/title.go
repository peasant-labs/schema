package intent

import (
	"regexp"
	"strconv"
	"strings"
)

// The intent title grammar:
//
//	title  = type ["(" scopes ")"] ["!"] ["[" bump "]"] ": " subject
//	landed = title [" (#" digits ")"]
//
// Scopes are free-form lowercase words. A bracket is required on the
// must-release types, optional on the may-release types, and forbidden on the
// never-release types. The landed suffix is the squash-commit PR number.
var (
	oldReleasePattern = regexp.MustCompile(`^release\(v?\d+\.\d+\.\d+[^)]*\)`)
	landedSuffix      = regexp.MustCompile(` \(#(\d+)\)$`)
	titlePattern      = regexp.MustCompile(`^([a-z]+)(\(([^()]*)\))?(!)?(\[(major|minor|patch)\])?: (\S.*)$`)
	scopePattern      = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

// SplitLandedSubject splits a landed subject into its title text and PR
// number. A PR title without the squash suffix returns pr 0. It is the one
// rule for the suffix: callers never re-derive it.
func SplitLandedSubject(raw string) (text string, pr int) {
	if m := landedSuffix.FindStringSubmatch(raw); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return raw, 0
		}
		return raw[:len(raw)-len(m[0])], n
	}
	return raw, 0
}

// ParseTitle parses one intent title against the release config and the
// repo's current version, which decides what a breaking mark may pair with.
// An empty current reads as below 1.0. It accepts both PR titles and landed
// subjects; a landed subject sets Title.PR to its PR number.
func ParseTitle(cfg Config, current SemVer, raw string) (Title, error) {
	trimmed := strings.TrimSpace(raw)
	if oldReleasePattern.MatchString(trimmed) {
		return Title{}, rejection(ErrOldReleaseTitle, raw,
			"the title uses the old release form",
			"release titles carry a level, never a number, since the release job computes every version",
			"the title-guard check is red, so the merge queue will not accept this PR",
			"retitle with a level, for example \"release[minor]: publish the pending changes\", and drop any version or CHANGELOG edit from the PR")
	}

	text, pr := SplitLandedSubject(trimmed)

	m := titlePattern.FindStringSubmatch(text)
	if m == nil {
		return Title{}, rejection(ErrMalformed, raw,
			"the title does not match the intent grammar",
			"a title is a type, an optional scope list, an optional breaking mark, an optional level, then \": \" and a subject",
			"the title-guard check is red, so the merge queue will not accept this PR",
			"retitle in the grammar, for example \"feat(graph)[minor]: add a facet\"; a GitHub Revert title needs a fresh title in this form")
	}

	typ := Type(m[1])
	if !typ.IsValid() {
		return Title{}, rejection(ErrUnknownType, raw,
			"type "+strconv.Quote(m[1])+" is not a known title type",
			"the type set is closed so every repo judges the same verbs",
			"the title-guard check is red, so the merge queue will not accept this PR",
			"pick a type from the closed set, for example feat, fix, release, refactor, docs, test, ci, chore or style")
	}

	scopes, ok := parseScopes(m[3], m[2] != "")
	if !ok {
		return Title{}, rejection(ErrMalformed, raw,
			"the scope list "+strconv.Quote(m[2])+" does not parse",
			"scopes are comma-separated lowercase words with no spaces, for example \"(graph)\" or \"(graph,table)\"",
			"the title-guard check is red, so the merge queue will not accept this PR",
			"rename each scope to lowercase letters, digits and hyphens with no spaces, or drop the list")
	}

	breaking := m[4] == "!"
	level, _ := parseLevel(m[6])

	if !cfg.Releases && (level != LevelNone || breaking) {
		return Title{}, rejection(ErrReleasesDisabled, raw,
			"the title carries a release marker in a repo that does not release",
			"this repo has releases off, so no title may carry a level or a breaking mark",
			"the title-guard check is red, so the merge queue will not accept this PR",
			"drop the bracket and the breaking mark, for example \"feat(graph): add a facet\"")
	}

	switch typ.Class() {
	case ClassMustRelease:
		if cfg.Releases && level == LevelNone {
			return Title{}, rejection(ErrMissingBump, raw,
				"type "+typ.String()+" releases but the title has no level",
				"feat, fix and release always release, and nothing is inferred from the type",
				"the title-guard check is red, so the merge queue will not accept this PR",
				"add a level, for example \""+typ.String()+"[minor]: ...\" (or [patch]); if the change ships nothing, use refactor, test, ci, chore or docs with no bracket")
		}
	case ClassNeverRelease:
		if level != LevelNone || breaking {
			return Title{}, rejection(ErrBumpOnNonReleasing, raw,
				"type "+typ.String()+" never releases but the title carries a release marker",
				"docs, test, ci, chore and style ship no release, so a bracket or a breaking mark on them contradicts the type",
				"the title-guard check is red, so the merge queue will not accept this PR",
				"drop the bracket and the breaking mark, or change the type to one that releases")
		}
	}

	if breaking {
		below10 := true
		if current != "" {
			major, _, _, splitErr := splitSemVer(current)
			if splitErr != nil {
				return Title{}, inputError(
					"invalid current version "+strconv.Quote(string(current)),
					"the current version is the version base and must be a bare X.Y.Z value",
					"intent.ParseTitle",
					"checking what a breaking mark may pair with",
					"the breaking rule cannot be judged from this version",
					"pass the highest final release tag without its prefix, for example \"0.1.0\"",
				)
			}
			below10 = major == 0
		}
		want := LevelMinor
		if !below10 {
			want = LevelMajor
		}
		if level != want {
			return Title{}, rejection(ErrBangBumpMismatch, raw,
				"the breaking mark pairs with ["+level.String()+"] but this repo needs ["+want.String()+"]",
				bangWhy(below10),
				"the title-guard check is red, so the merge queue will not accept this PR",
				"write the break as \"!["+want.String()+"]\", for example \"feat(graph)!["+want.String()+"]: drop the legacy field\"")
		}
	} else if level == LevelMajor {
		return Title{}, rejection(ErrBangBumpMismatch, raw,
			"the title carries [major] with no breaking mark",
			"a major level only exists alongside a declared break; no title mints the next major by itself",
			"the title-guard check is red, so the merge queue will not accept this PR",
			"add the breaking mark for the repo's form, or lower the level to [minor]")
	}

	return Title{
		Type:     typ,
		Scopes:   scopes,
		Breaking: breaking,
		Level:    level,
		Subject:  m[7],
		PR:       pr,
	}, nil
}

// bangWhy states the breaking rule for the repo's side of 1.0.
func bangWhy(below10 bool) string {
	if below10 {
		return "below 1.0 a break is written ![minor] and moves the middle number; ![patch], ![major] and [major] are rejected"
	}
	return "from 1.0 a break is written ![major]; ! with [minor] or [patch], and [major] with no !, are rejected"
}

// parseScopes validates one scope list. present reports whether the title
// carried parentheses at all; empty parentheses are malformed.
func parseScopes(inner string, present bool) ([]string, bool) {
	if !present {
		return []string{}, true
	}
	parts := strings.Split(inner, ",")
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		if !scopePattern.MatchString(part) {
			return nil, false
		}
		scopes = append(scopes, part)
	}
	return scopes, true
}
