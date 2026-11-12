package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// runPlan implements: release-intent plan --branch REF --mode MODE [--floor SHA] [--verdict taggable|rejected].
func runPlan(args []string) error {
	var branch, mode, floor, verdict string
	var haveBranch, haveMode bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--branch":
			val, next, err := requireFlagValue(args, i, "--branch")
			if err != nil {
				return err
			}
			branch, haveBranch = val, true
			i = next
		case "--mode":
			val, next, err := requireFlagValue(args, i, "--mode")
			if err != nil {
				return err
			}
			mode, haveMode = val, true
			i = next
		case "--floor":
			val, next, err := requireFlagValue(args, i, "--floor")
			if err != nil {
				return err
			}
			floor = val
			i = next
		case "--verdict":
			val, next, err := requireFlagValue(args, i, "--verdict")
			if err != nil {
				return err
			}
			if val != "taggable" && val != "rejected" {
				return usageErr("usage: release-intent plan --branch REF --mode MODE [--floor SHA] [--verdict taggable|rejected] (bad --verdict %q): pass taggable or rejected", val)
			}
			verdict = val
			i = next
		default:
			return usageErr("unknown flag %q for plan: expected --branch REF --mode MODE [--floor SHA] [--verdict taggable|rejected]", args[i])
		}
	}
	if !haveBranch {
		return usageErr("usage: release-intent plan --branch REF --mode MODE [--floor SHA] [--verdict taggable|rejected] (missing required --branch): pass the branch to plan from, e.g. origin/develop")
	}
	if !haveMode {
		return usageErr("usage: release-intent plan --branch REF --mode MODE [--floor SHA] [--verdict taggable|rejected] (missing required --mode): pass tag or plan-only")
	}
	ctx := context.Background()
	runner, err := newExecGit()
	if err != nil {
		return err
	}
	m := ParseMode(mode)
	facts, cfg, err := Gather(ctx, runner, branch, DefaultConfigPath, DefaultConfigPath, floor, m, verdict)
	if err != nil {
		return err
	}
	decision, freshKnown, ferr := Decide(ctx, runner, cfg, facts)
	if ferr != nil {
		if r, ok := asReject(ferr); ok {
			return r
		}
		return ferr
	}
	lines := []string{
		"action=" + string(decision.Action),
		"range_base=" + decision.RangeBase,
		"version_base=" + decision.VersionBase,
		"candidate=" + decision.Candidate,
	}
	if decision.Candidate == "" {
		lines = append(lines, "candidate_fresh=")
	} else if !freshKnown {
		lines = append(lines, "candidate_fresh=")
	} else {
		lines = append(lines, "candidate_fresh="+boolString(decision.CandidateFresh))
	}
	lines = append(lines,
		"level="+decision.Level,
		"breaking="+boolString(decision.Breaking),
		"version="+decision.Version,
		"tag="+decision.Tag,
		"error="+decision.ErrorCode,
	)
	return emitLines(lines)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Decision is the planner output.
type PlanDecision struct {
	Action         Action
	RangeBase      string
	VersionBase    string
	Candidate      string
	CandidateFresh bool
	Level          string
	Breaking       bool
	Version        string
	Tag            string
	ErrorCode      string
}

// Action is the closed set of plan actions.
type Action string

const (
	ActionNone    Action = "none"
	ActionVerify  Action = "verify"
	ActionTag     Action = "tag"
	ActionOpenPR  Action = "open-pr"
	ActionClosePR Action = "close-pr"
)

// AllActions is the canonical closed set.
var AllActions = []Action{ActionNone, ActionVerify, ActionTag, ActionOpenPR, ActionClosePR}

// Decide applies Plan steps 1-8 purely except for freshness file reads. Title
// parsing, version arithmetic and protected checks delegate to the contract
// package; the command layer owns only git fact collection and output mapping.
func Decide(ctx context.Context, runner GitRunner, cfg intent.Config, facts Facts) (PlanDecision, bool, error) {
	versionBase, versionBaseVer := highestFinalTag(facts.Tags, cfg.TagPrefix)
	if versionBase == "" {
		return PlanDecision{}, true, &RejectError{
			Code: CodeNoFinalTag, What: "no final release tag exists",
			Why: "the range base needs the newest final tag on the chain", Where: "git tags, read by release-intent plan",
			When: "release job run on the branch tip", Means: "no plan can be computed; no release PR is opened",
			Fix: "cut the first release manually, or check tag_prefix and tag fetching",
		}
	}
	rangeBase, _ := newestFinalOnChain(facts.Commits, facts.Tags, cfg.TagPrefix)
	if rangeBase == "" {
		return PlanDecision{}, true, &RejectError{
			Code: CodeNoFinalTag, What: "no final release tag is on the branch chain",
			Why: "the range starts after the newest final tag on this branch", Where: "first-parent chain, read by release-intent plan",
			When: "release job run on the branch tip", Means: "no plan can be computed",
			Fix: "fetch tags and confirm the branch contains a final tag",
		}
	}
	if facts.Epoch == "" {
		return PlanDecision{}, true, &RejectError{
			Code: CodeEpochMissing, What: "no commit adds the release config",
			Why: "the epoch is the commit that added .github/release-intent.yaml", Where: "git history, read by release-intent plan",
			When: "release job run on the branch tip", Means: "no plan can be computed",
			Fix: "add .github/release-intent.yaml in a commit on this branch",
		}
	}
	commits := facts.Commits
	if facts.Floor != "" {
		onChain := false
		for _, c := range commits {
			if c.SHA == facts.Floor {
				onChain = true
				break
			}
		}
		if !onChain {
			isAnc, err := runner.IsAncestor(ctx, facts.Floor, rangeBaseSHA(facts, rangeBase))
			if err != nil || !isAnc {
				if facts.Floor != rangeBaseSHA(facts, rangeBase) {
					return PlanDecision{}, true, &RejectError{
						Code: CodeFloorNotOnChain, What: fmt.Sprintf("floor %q is not on the branch chain", facts.Floor),
						Why: "the floor must be a commit on this branch", Where: "git history, read by release-intent plan",
						When: "release job run with RELEASE_BATCH_FLOOR set", Means: "no plan can be computed",
						Fix: "set RELEASE_BATCH_FLOOR to a commit on the branch, or unset it",
					}
				}
			}
		}
	}
	rng := rangeAfter(commits, rangeBaseSHA(facts, rangeBase), facts.Epoch, facts.Floor)
	candidate := ""
	for _, c := range rng {
		if c.SHA != facts.Epoch && c.ChangesChangelog {
			candidate = c.SHA
			break
		}
	}
	var current intent.SemVer
	if versionBaseVer != "" {
		if sv, serr := intent.ParseSemVer(versionBaseVer); serr == nil {
			current = sv
		}
	}
	type subj struct {
		title intent.Title
		sha   string
		isRel bool
	}
	var subjects []subj
	for _, c := range rng {
		s := strings.TrimSpace(c.Subject)
		if idx := strings.LastIndex(s, " (#"); idx >= 0 && strings.HasSuffix(s, ")") {
			s = strings.TrimSpace(s[:idx])
		}
		pt, perr := intent.ParseTitle(cfg, current, s)
		if perr != nil {
			return PlanDecision{}, true, intentErrorToReject(perr)
		}
		subjects = append(subjects, subj{title: pt, sha: c.SHA, isRel: c.SHA != candidate && pt.Level != intent.LevelNone})
	}
	var pending []subj
	for _, s := range subjects {
		if s.sha == candidate {
			continue
		}
		isChangelog := false
		for _, c := range rng {
			if c.SHA == s.sha && c.ChangesChangelog {
				isChangelog = true
				break
			}
		}
		if isChangelog {
			continue
		}
		if s.isRel {
			pending = append(pending, s)
		}
	}
	var pendingTitles []intent.Title
	for _, s := range pending {
		pendingTitles = append(pendingTitles, s.title)
	}
	level, breaking := highestLevelOf(pendingTitles)
	version := ""
	if level != intent.LevelNone {
		if sv, verr := intent.NextVersion(current, level, breaking); verr == nil {
			version = string(sv)
		}
	}
	levelWord := ""
	switch level {
	case intent.LevelPatch:
		levelWord = "patch"
	case intent.LevelMinor:
		levelWord = "minor"
	case intent.LevelMajor:
		levelWord = "major"
	}
	tag := ""
	if version != "" {
		tag = cfg.TagPrefix + version
	}
	fresh := false
	freshKnown := true
	if candidate != "" {
		f, known, ferr := candidateFresh(ctx, runner, cfg, rng, candidate, current)
		if ferr != nil {
			return PlanDecision{}, false, ferr
		}
		fresh, freshKnown = f, known
	}
	if facts.Mode != ModeTag {
		return PlanDecision{Action: ActionNone, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: fresh, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, freshKnown, nil
	}
	if candidate != "" && fresh {
		if facts.Verdict == "" {
			return PlanDecision{Action: ActionVerify, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, true, nil
		}
		if facts.Verdict == "taggable" {
			return PlanDecision{Action: ActionTag, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, true, nil
		}
		if len(pending) > 0 {
			return PlanDecision{Action: ActionOpenPR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, true, nil
		}
		return PlanDecision{Action: ActionClosePR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, true, nil
	}
	if len(pending) > 0 {
		return PlanDecision{Action: ActionOpenPR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: fresh, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, freshKnown, nil
	}
	return PlanDecision{Action: ActionClosePR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: fresh, Level: levelWord, Breaking: breaking, Version: version, Tag: tag}, freshKnown, nil
}

func rangeBaseSHA(facts Facts, rangeBase string) string {
	if sha, ok := facts.Tags[rangeBase]; ok {
		return sha
	}
	return rangeBase
}

func rangeAfter(commits []Commit, rangeBaseSHA, epoch, floor string) []Commit {
	var out []Commit
	for _, c := range commits {
		if c.SHA == rangeBaseSHA {
			break
		}
		out = append(out, c)
	}
	if floor != "" {
		var kept []Commit
		for _, c := range out {
			if c.SHA == floor {
				break
			}
			kept = append(kept, c)
		}
		if len(kept) != len(out) || containsSHA(out, floor) {
			out = kept
		}
	}
	idx := -1
	for i, c := range out {
		if c.SHA == epoch {
			idx = i
			break
		}
	}
	if idx == -1 {
		return out
	}
	return out[:idx+1]
}

func containsSHA(commits []Commit, sha string) bool {
	for _, c := range commits {
		if c.SHA == sha {
			return true
		}
	}
	return false
}

func newestFinalOnChain(commits []Commit, tags map[string]string, prefix string) (string, string) {
	bySHA := map[string]string{}
	for t, sha := range tags {
		v := strings.TrimPrefix(t, prefix)
		if v == t || strings.Contains(v, "-rc") {
			continue
		}
		if _, err := intent.ParseSemVer(v); err != nil {
			continue
		}
		bySHA[sha] = t
	}
	for _, c := range commits {
		if t, ok := bySHA[c.SHA]; ok {
			return t, strings.TrimPrefix(t, prefix)
		}
	}
	return "", ""
}

func highestLevelOf(titles []intent.Title) (intent.Level, bool) {
	best := intent.LevelNone
	breaking := false
	for _, t := range titles {
		if t.Level > best {
			best = t.Level
		}
		if t.Breaking {
			breaking = true
		}
	}
	return best, breaking
}

// candidateFresh compares the candidate's changelog and version fields with
// the expected version for the range before it. Any unreadable input is an
// error (exit 1, no candidate_fresh), never stale. The full byte comparison
// against Generate output moves to the contract package with Plan; the command
// layer checks the readable proxy and fails closed on unreadable inputs.
func candidateFresh(ctx context.Context, runner GitRunner, cfg intent.Config, rng []Commit, candidate string, current intent.SemVer) (bool, bool, error) {
	var older []Commit
	for idx := len(rng) - 1; idx >= 0; idx-- {
		if rng[idx].SHA == candidate {
			break
		}
		older = append(older, rng[idx])
	}
	var titles []intent.Title
	for _, c := range older {
		s := strings.TrimSpace(c.Subject)
		if idx := strings.LastIndex(s, " (#"); idx >= 0 && strings.HasSuffix(s, ")") {
			s = strings.TrimSpace(s[:idx])
		}
		if pt, perr := intent.ParseTitle(cfg, current, s); perr == nil {
			if c.SHA != candidate && !c.ChangesChangelog && pt.Level != intent.LevelNone {
				titles = append(titles, pt)
			}
		}
	}
	level, _ := highestLevelOf(titles)
	var wantVersion intent.SemVer
	if level != intent.LevelNone {
		if sv, verr := intent.NextVersion(current, level, false); verr == nil {
			// Breaking moves the same numbers below 1.0, so the proxy version
			// is sufficient for the freshness read check.
			wantVersion = sv
		} else {
			// Fall back to recomputing with breaking true when the plain
			// arithmetic needs it.
			if sv2, verr2 := intent.NextVersion(current, level, true); verr2 == nil {
				wantVersion = sv2
			}
		}
	}
	gotChangelog, err := runner.ReadFile(ctx, candidate, cfg.Changelog)
	if err != nil {
		return false, false, fmt.Errorf("plan: cannot read changelog %q at candidate %q for the freshness comparison: %w. A read error is never reported as stale", cfg.Changelog, candidate, err)
	}
	if len(gotChangelog) == 0 {
		return false, true, nil
	}
	if wantVersion != "" && !strings.Contains(string(gotChangelog), string(wantVersion)) {
		return false, true, nil
	}
	for _, vf := range cfg.VersionFiles {
		got, err := runner.ReadFile(ctx, candidate, vf.Path)
		if err != nil {
			return false, false, fmt.Errorf("plan: cannot read version file %q at candidate %q for the freshness comparison: %w. A read error is never reported as stale", vf.Path, candidate, err)
		}
		if wantVersion != "" && !strings.Contains(string(got), string(wantVersion)) {
			return false, true, nil
		}
	}
	return true, true, nil
}

var _ = sort.StringSlice{}
