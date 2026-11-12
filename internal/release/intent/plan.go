package intent

import (
	"strings"
)

// This file holds the release planner: the pure reduction of branch facts to
// the one action plus every value the workflow needs to carry it out.
//
// The planner reads the range base (the newest final tag on the first-parent
// chain) and the version base (the highest final tag anywhere) from git
// through facts the command layer filled in. Only final tags count; release
// candidates and unrelated tags never base arithmetic.

// splitReleaseTag splits a release tag into its bare version. Anything the
// prefix does not open, or the version grammar does not accept, is not a
// release tag and is ignored.
func splitReleaseTag(prefix, tag string) (SemVer, bool) {
	rest, ok := strings.CutPrefix(tag, prefix)
	if !ok {
		return "", false
	}
	v, err := ParseSemVer(rest)
	if err != nil {
		return "", false
	}
	return v, true
}

// HighestFinalTag returns the highest final release tag in tags: a tag the
// prefix opens whose rest parses as a bare X.Y.Z value. Release candidates
// and unrelated tags, such as a bare word with the prefix or another tool's
// namespaced tag, never count. It is the one rule for the version base;
// Plan uses it so both serve the same answer.
func HighestFinalTag(tags map[string]string, prefix string) (tag string, version SemVer, ok bool) {
	for name := range tags {
		v, isRelease := splitReleaseTag(prefix, name)
		if !isRelease {
			continue
		}
		if !ok {
			tag, version, ok = name, v, true
			continue
		}
		newer, err := compareSemVer(v, version)
		if err != nil || newer <= 0 {
			continue
		}
		tag, version = name, v
	}
	return tag, version, ok
}

// planError builds a typed planner rejection.
func planError(code ErrorCode, what, why, means, fix string) *Error {
	return &Error{
		Code:  code,
		What:  what,
		Why:   why,
		Where: "release history, read by intent.Plan",
		When:  "planning the release action for the branch tip",
		Means: means,
		Fix:   fix,
	}
}

// Plan reduces branch facts to the one release decision. Every subject in the
// range is parsed; one unparseable subject fails the whole plan with its code.
// A stale or unreadable candidate never verifies or tags: the action falls
// back to open-pr or close-pr, and freshness reports unknown when no entry
// was recorded for the candidate.
func Plan(cfg Config, facts Facts) (Decision, error) {
	tagsBySHA := map[string][]SemVer{}
	for tag, sha := range facts.Tags {
		v, ok := splitReleaseTag(cfg.TagPrefix, tag)
		if !ok {
			continue
		}
		tagsBySHA[sha] = append(tagsBySHA[sha], v)
	}
	versionBaseTag, versionBase, found := HighestFinalTag(facts.Tags, cfg.TagPrefix)
	if !found {
		return Decision{}, planError(ErrNoFinalTag,
			"no final release tag in the repository",
			"arithmetic needs a base, and only final tags base it",
			"no release action can be planned and no version computed",
			"cut the first release under the previous ceremony so the planner has a base")
	}

	rangeBaseIdx := -1
	for i, commit := range facts.Commits {
		if len(tagsBySHA[commit.SHA]) > 0 {
			rangeBaseIdx = i
			break
		}
	}
	if rangeBaseIdx < 0 {
		return Decision{}, planError(ErrNoFinalTag,
			"no final release tag on the branch chain",
			"the range starts after the newest final tag on the chain",
			"no release action can be planned and no version computed",
			"cut the first release under the previous ceremony so the planner has a base")
	}
	end := rangeBaseIdx
	if facts.Floor != "" {
		floorIdx := -1
		for i, commit := range facts.Commits {
			if commit.SHA == facts.Floor {
				floorIdx = i
				break
			}
		}
		if floorIdx < 0 {
			return Decision{}, planError(ErrFloorNotOnChain,
				"the recovery floor "+quote(facts.Floor)+" is not on the branch chain",
				"the floor drops a bad subject and everything before it, so it must name a chain commit",
				"no release action can be planned while the floor points nowhere",
				"set the floor to the bad commit itself, or clear it once a newer final tag exists")
		}
		if floorIdx < end {
			end = floorIdx
		}
	}
	if facts.Epoch == "" {
		return Decision{}, planError(ErrEpochMissing,
			"no grammar epoch recorded for this branch",
			"subjects before the config landed are not parsed, so the epoch bounds the range",
			"no release action can be planned without the range bound",
			"record the commit that added the release config as the epoch and plan again")
	}
	epochIdx := -1
	for i, commit := range facts.Commits {
		if commit.SHA == facts.Epoch {
			epochIdx = i
			break
		}
	}
	if epochIdx < 0 {
		return Decision{}, planError(ErrEpochMissing,
			"the grammar epoch "+quote(facts.Epoch)+" is not on the branch chain",
			"subjects before the config landed are not parsed, so the epoch must name a chain commit",
			"no release action can be planned without the range bound",
			"record the commit that added the release config as the epoch and plan again")
	}
	if epochIdx < end {
		end = epochIdx + 1
	}
	rng := facts.Commits[:end]

	candidate := ""
	freshness := FreshnessUnknown
	for _, commit := range rng {
		if commit.SHA == facts.Epoch || !commit.ChangesChangelog {
			continue
		}
		candidate = commit.SHA
		if fresh, ok := facts.Fresh[commit.SHA]; ok {
			if fresh {
				freshness = FreshnessFresh
			} else {
				freshness = FreshnessStale
			}
		}
		break
	}

	pending := []Title{}
	level := LevelNone
	breaking := false
	for _, commit := range rng {
		title, err := ParseTitle(cfg, versionBase, commit.Subject)
		if err != nil {
			return Decision{}, err
		}
		if commit.ChangesChangelog || title.Level == LevelNone {
			continue
		}
		pending = append(pending, title)
		if title.Level > level {
			level = title.Level
		}
		if title.Breaking {
			breaking = true
		}
	}

	version := SemVer("")
	tag := ""
	if len(pending) > 0 {
		next, err := NextVersion(versionBase, level, breaking)
		if err != nil {
			return Decision{}, err
		}
		version = next
		tag = cfg.TagPrefix + string(next)
	}

	rangeBase := ""
	for tagName, sha := range facts.Tags {
		if sha != facts.Commits[rangeBaseIdx].SHA {
			continue
		}
		if _, ok := splitReleaseTag(cfg.TagPrefix, tagName); ok {
			rangeBase = tagName
			break
		}
	}

	action := ActionClosePR
	if len(pending) > 0 {
		action = ActionOpenPR
	}
	if facts.Mode == ModeTag && candidate != "" && freshness == FreshnessFresh {
		switch facts.Verdict {
		case VerdictTaggable:
			action = ActionTag
		case VerdictRejected:
			action = ActionOpenPR
		default:
			action = ActionVerify
		}
	}
	if facts.Mode != ModeTag {
		action = ActionNone
	}

	return Decision{
		Action:         action,
		RangeBase:      rangeBase,
		VersionBase:    versionBaseTag,
		Candidate:      candidate,
		CandidateFresh: freshness,
		Pending:        pending,
		Level:          level,
		Breaking:       breaking,
		Version:        version,
		Tag:            tag,
	}, nil
}

func quote(s string) string {
	return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
}
