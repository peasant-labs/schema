package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// parsedTitle is the machine-readable form of an intent title.
type parsedTitle struct {
	Type     string
	Scopes   []string
	Breaking bool
	Level    string
	Subject  string
}

var (
	typeSet = map[string]bool{
		"feat": true, "fix": true, "release": true,
		"refactor": true, "perf": true, "build": true, "revert": true,
		"docs": true, "test": true, "ci": true, "chore": true, "style": true,
	}
	mustRelease = map[string]bool{"feat": true, "fix": true, "release": true}
	mayRelease  = map[string]bool{"refactor": true, "perf": true, "build": true, "revert": true}
	neverRel    = map[string]bool{"docs": true, "test": true, "ci": true, "chore": true, "style": true}

	fullTitleRe = regexp.MustCompile(`^([a-z]+)(\(([^)]*)\))?(!)?(\[(major|minor|patch)\])?: (\S.*)$`)
	scopeRe     = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

// parseFullTitle validates the grammar and returns the typed code on failure.
func parseFullTitle(raw string, releases bool, belowOne bool) (parsedTitle, *RejectError) {
	if strings.HasPrefix(raw, "release(v") || strings.HasPrefix(raw, "release (v") {
		return parsedTitle{}, &RejectError{
			Code: CodeOldReleaseTitle, What: fmt.Sprintf("title %q uses the old release(vX.Y.Z) form", raw),
			Why: "the old form carries a version number and versions are now computed", Where: "PR title, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red, so the merge queue will not accept this PR",
			Fix: "retitle as release[patch]: ... or release[minor]: ... with no version; drop any package.json version or CHANGELOG.md edit from the PR",
		}
	}
	if strings.HasPrefix(raw, "Revert \"") {
		return parsedTitle{}, &RejectError{
			Code: CodeMalformed, What: fmt.Sprintf("title %q is a Revert-button title", raw),
			Why: "the Revert button writes Revert \"<old title>\" which is not in the grammar", Where: "PR title, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red, so the merge queue will not accept this PR",
			Fix: "retitle manually in the grammar, e.g. revert(<scope>)[patch]: revert ...; no push is needed",
		}
	}
	m := fullTitleRe.FindStringSubmatch(raw)
	if m == nil {
		return parsedTitle{}, &RejectError{
			Code: CodeMalformed, What: fmt.Sprintf("title %q does not match type[(scopes)][!][[level]]: subject", raw),
			Why: "the title grammar requires type, optional scopes, optional !, optional [level], then : and a subject", Where: "PR title, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red, so the merge queue will not accept this PR",
			Fix: "retitle as feat(<scope>)[minor]: ... (or [patch]); scopes are free-form lowercase; no push is needed",
		}
	}
	typ, scopesRaw, bang, level := m[1], m[3], m[4] == "!", m[6]
	if !typeSet[typ] {
		return parsedTitle{}, &RejectError{
			Code: CodeUnknownType, What: fmt.Sprintf("type %q is not in the closed set", typ),
			Why: "titles must use feat, fix, release, refactor, perf, build, revert, docs, test, ci, chore or style", Where: "PR title, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red, so the merge queue will not accept this PR",
			Fix: "change the type (e.g. feat, fix, refactor, docs); scopes stay free-form",
		}
	}
	var scopes []string
	if m[2] != "" {
		if scopesRaw == "" {
			return parsedTitle{}, &RejectError{
				Code: CodeMalformed, What: fmt.Sprintf("title %q has empty parentheses", raw),
				Why: "scopes are free-form but parentheses must hold at least one scope", Where: "PR title, checked by release-intent check-title",
				When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
				Fix: "drop the empty () or add a scope, e.g. feat(transcript)[minor]: ...",
			}
		}
		for _, s := range strings.Split(scopesRaw, ",") {
			if !scopeRe.MatchString(s) {
				return parsedTitle{}, &RejectError{
					Code: CodeMalformed, What: fmt.Sprintf("scope %q is not lowercase alphanumerics with dashes", s),
					Why: "scopes are free-form lowercase; no spaces", Where: "PR title, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "use lowercase scopes with dashes, e.g. (transcript) or (local-api,transcript)",
				}
			}
			scopes = append(scopes, s)
		}
	}
	pt := parsedTitle{Type: typ, Scopes: scopes, Breaking: bang, Level: level, Subject: m[7]}
	if !releases && (level != "" || bang) {
		return parsedTitle{}, &RejectError{
			Code: CodeReleasesDisabled, What: fmt.Sprintf("title %q carries a bump in a repo with releases off", raw),
			Why: "village never releases, so no title may carry [level] or !", Where: "PR title, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
			Fix: "drop the bracket and ! (e.g. feat(transcript): ...)",
		}
	}
	if mustRelease[typ] && releases && level == "" {
		return parsedTitle{}, &RejectError{
			Code: CodeMissingBump, What: fmt.Sprintf("title %q has no bump bracket", raw),
			Why: "feat, fix and release always release, and nothing is inferred from the type", Where: "PR title, checked by release-intent check-title",
			When: "pull_request edited, before the PR can be enqueued", Means: "the title-guard check is red, so the merge queue will not accept this PR",
			Fix: fmt.Sprintf("retitle as %s%s%s[minor]: ... (or [patch]); if the change ships nothing, use refactor, test, ci, chore or docs with no bracket. No push is needed", typ, scopeSuffix(scopes), bangMark(bang)),
		}
	}
	if neverRel[typ] && (level != "" || bang) {
		return parsedTitle{}, &RejectError{
			Code: CodeBumpOnNonReleasing, What: fmt.Sprintf("type %q never releases but title %q carries a bump", typ, raw),
			Why: "docs, test, ci, chore and style never release by definition", Where: "PR title, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
			Fix: "drop the bracket and !, or change the type to feat, fix, refactor, perf, build or revert",
		}
	}
	if bang || level != "" {
		if belowOne {
			if bang && level != "minor" {
				return parsedTitle{}, &RejectError{
					Code: CodeBangBumpMismatch, What: fmt.Sprintf("title %q uses ! with [%s] below 1.0", raw, level),
					Why: "below 1.0 a breaking change is ![minor] only: ![major] and ![patch] are rejected", Where: "PR title, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "use ![minor] for a break below 1.0, or drop ! for a non-breaking [minor]/[patch]",
				}
			}
			if !bang && level == "major" {
				return parsedTitle{}, &RejectError{
					Code: CodeBangBumpMismatch, What: fmt.Sprintf("title %q uses [major] with no ! below 1.0", raw),
					Why: "a non-breaking major does not exist, and no title mints 1.0.0", Where: "PR title, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "use [minor] or [patch], or ![minor] for a break",
				}
			}
		} else {
			if bang && level != "major" {
				return parsedTitle{}, &RejectError{
					Code: CodeBangBumpMismatch, What: fmt.Sprintf("title %q uses ! with [%s] at 1.0 or above", raw, level),
					Why: "from 1.0, ! requires [major]", Where: "PR title, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "use ![major] for a break, or drop ! for [minor]/[patch]",
				}
			}
			if !bang && level == "major" {
				return parsedTitle{}, &RejectError{
					Code: CodeBangBumpMismatch, What: fmt.Sprintf("title %q uses [major] with no !", raw),
					Why: "from 1.0, [major] with no ! is rejected: a non-breaking major does not exist", Where: "PR title, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "use ![major] for a break, or [minor]/[patch] otherwise",
				}
			}
		}
	}
	return pt, nil
}

func scopeSuffix(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	return "(" + strings.Join(scopes, ",") + ")"
}

func bangMark(b bool) string {
	if b {
		return "!"
	}
	return ""
}

// parseSemVer parses X.Y.Z.
func parseSemVer(s string) (major, minor, patch int, ok bool) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	pat, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return maj, min, pat, true
}
