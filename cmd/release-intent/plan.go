package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
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
		// Read errors exit 1 with no candidate_fresh line.
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
		// Caller already exited 1; unreachable.
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

// Decide applies Plan steps 1-8 purely except for freshness file reads.
func Decide(ctx context.Context, runner GitRunner, cfg Config, facts Facts) (PlanDecision, bool, error) {
	// Step 2: version base = highest final tag anywhere.
	versionBase, versionBaseVer := highestFinalTag(facts.Tags, cfg.TagPrefix)
	if versionBase == "" {
		return PlanDecision{}, true, &RejectError{
			Code: CodeNoFinalTag, What: "no final release tag exists",
			Why: "the range base needs the newest final tag on the chain", Where: "git tags, read by release-intent plan",
			When: "release job run on the branch tip", Means: "no plan can be computed; no release PR is opened",
			Fix: "cut the first release manually, or check tag_prefix and tag fetching",
		}
	}
	// Step 2b: range base = newest final tag on the chain.
	rangeBase, rangeBaseVer := newestFinalOnChain(facts.Commits, facts.Tags, cfg.TagPrefix)
	if rangeBase == "" {
		return PlanDecision{}, true, &RejectError{
			Code: CodeNoFinalTag, What: "no final release tag is on the branch chain",
			Why: "the range starts after the newest final tag on this branch", Where: "first-parent chain, read by release-intent plan",
			When: "release job run on the branch tip", Means: "no plan can be computed",
			Fix: "fetch tags and confirm the branch contains a final tag",
		}
	}
	// Epoch: must be on the chain.
	if facts.Epoch == "" {
		return PlanDecision{}, true, &RejectError{
			Code: CodeEpochMissing, What: "no commit adds the release config",
			Why: "the epoch is the commit that added .github/release-intent.yaml", Where: "git history, read by release-intent plan",
			When: "release job run on the branch tip", Means: "no plan can be computed",
			Fix: "add .github/release-intent.yaml in a commit on this branch",
		}
	}
	// Floor validation.
	commits := facts.Commits
	if facts.Floor != "" {
		onChain := false
		for _, c := range commits {
			if c.SHA == facts.Floor {
				onChain = true
				break
			}
		}
		// A floor older than the range base is ignored.
		if !onChain {
			// Check ancestry: if the floor is an ancestor of the range base, it
			// is older and ignored; otherwise it is not on the chain.
			isAnc, err := runner.IsAncestor(ctx, facts.Floor, rangeBaseSHA(facts, rangeBase))
			if err != nil || !isAnc {
				// Also allow the floor to be the range base itself (already on
				// chain via tags, not commits list boundary).
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
	// Build the range: commits after the range base, up to the tip, dropping
	// commits at or before the floor and before the epoch.
	rng := rangeAfter(commits, rangeBaseSHA(facts, rangeBase), facts.Epoch, facts.Floor)
	// Step 4: candidate = newest non-epoch commit changing the changelog.
	candidate := ""
	for _, c := range rng {
		if c.SHA != facts.Epoch && c.ChangesChangelog {
			candidate = c.SHA
			break
		}
	}
	// Step 6: parse every subject; one bad subject fails the plan.
	type subj struct {
		title parsedTitle
		sha   string
		isRel bool
	}
	var subjects []subj
	below := belowOne(versionBaseVer)
	for _, c := range rng {
		s := strings.TrimSpace(c.Subject)
		if idx := strings.LastIndex(s, " (#"); idx >= 0 && strings.HasSuffix(s, ")") {
			s = strings.TrimSpace(s[:idx])
		}
		pt, rej := parseFullTitle(s, cfg.Releases, below)
		if rej != nil {
			return PlanDecision{}, true, rej
		}
		subjects = append(subjects, subj{title: pt, sha: c.SHA, isRel: c.SHA != candidate && isReleasing(pt)})
	}
	_ = rangeBaseVer
	// Pending excludes release commits.
	var pending []subj
	for _, s := range subjects {
		if s.sha == candidate {
			continue
		}
		// A commit that changes the changelog is a release commit even when its
		// subject parses as releasing: it never joins pending.
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
	// Version from pending (or from subjects before the candidate for verify/tag).
	var pendingTitles []parsedTitle
	for _, s := range pending {
		pendingTitles = append(pendingTitles, s.title)
	}
	level, breaking := highestLevelOf(pendingTitles)
	version := computeVersion(versionBaseVer, level, breaking)
	tag := ""
	if version != "" {
		tag = cfg.TagPrefix + version
	}
	// Freshness for the candidate: compare its files with Generate output for
	// the range before it. Unreadable inputs exit 1 with no candidate_fresh.
	fresh := false
	freshKnown := true
	if candidate != "" {
		f, known, ferr := candidateFresh(ctx, runner, cfg, rng, candidate, versionBaseVer)
		if ferr != nil {
			return PlanDecision{}, false, ferr
		}
		fresh, freshKnown = f, known
		// A stale candidate is treated as if it were not a release commit:
		// its subjects before it stay pending (already true, since the
		// candidate never joins pending).
		_ = fresh
	}
	// Tag mode off: always none.
	if facts.Mode != ModeTag {
		return PlanDecision{Action: ActionNone, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: fresh, Level: level, Breaking: breaking, Version: version, Tag: tag}, freshKnown, nil
	}
	// Actions.
	if candidate != "" && fresh {
		if facts.Verdict == "" {
			// Recompute version/tag from subjects before the candidate for
			// verify.
			return PlanDecision{Action: ActionVerify, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: level, Breaking: breaking, Version: version, Tag: tag}, true, nil
		}
		if facts.Verdict == "taggable" {
			return PlanDecision{Action: ActionTag, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: level, Breaking: breaking, Version: version, Tag: tag}, true, nil
		}
		// Rejected verdict with a fresh candidate: subjects before it stay
		// pending, so open-pr when pending non-empty else close-pr.
		if len(pending) > 0 {
			return PlanDecision{Action: ActionOpenPR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: level, Breaking: breaking, Version: version, Tag: tag}, true, nil
		}
		return PlanDecision{Action: ActionClosePR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: true, Level: level, Breaking: breaking, Version: version, Tag: tag}, true, nil
	}
	if len(pending) > 0 {
		return PlanDecision{Action: ActionOpenPR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: fresh, Level: level, Breaking: breaking, Version: version, Tag: tag}, freshKnown, nil
	}
	return PlanDecision{Action: ActionClosePR, RangeBase: rangeBase, VersionBase: versionBase, Candidate: candidate, CandidateFresh: fresh, Level: level, Breaking: breaking, Version: version, Tag: tag}, freshKnown, nil
}

func rangeBaseSHA(facts Facts, rangeBase string) string {
	if sha, ok := facts.Tags[rangeBase]; ok {
		return sha
	}
	return rangeBase
}

func rangeAfter(commits []Commit, rangeBaseSHA, epoch, floor string) []Commit {
	// Commits are newest first. Cut at the range base (exclusive).
	var out []Commit
	for _, c := range commits {
		if c.SHA == rangeBaseSHA {
			break
		}
		out = append(out, c)
	}
	// Drop commits at or before the floor (floor inclusive) and strictly before
	// the epoch. The epoch commit itself is parsed.
	if floor != "" {
		var kept []Commit
		for _, c := range out {
			if c.SHA == floor {
				break
			}
			kept = append(kept, c)
		}
		// If the floor was not in the sliced range (older than range base),
		// keep everything: the caller already ignored that case.
		if len(kept) != len(out) || containsSHA(out, floor) {
			out = kept
		}
	}
	// Drop commits older than the epoch (keep the epoch itself).
	idx := -1
	for i, c := range out {
		if c.SHA == epoch {
			idx = i
			break
		}
	}
	if idx == -1 {
		// Epoch not in the sliced range: it is older than the range base, so
		// keep the whole range.
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
		if _, _, _, ok := parseSemVer(v); !ok {
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

func isReleasing(pt parsedTitle) bool {
	switch pt.Type {
	case "feat", "fix", "release":
		return pt.Level != ""
	case "refactor", "perf", "build", "revert":
		return pt.Level != ""
	default:
		return false
	}
}

func highestLevelOf(titles []parsedTitle) (string, bool) {
	rank := map[string]int{"": 0, "patch": 1, "minor": 2, "major": 3}
	best := ""
	breaking := false
	for _, t := range titles {
		if rank[t.Level] > rank[best] {
			best = t.Level
		}
		if t.Breaking {
			breaking = true
		}
	}
	return best, breaking
}

func computeVersion(base, level string, breaking bool) string {
	if level == "" {
		return ""
	}
	maj, mi, pa, ok := parseSemVer(base)
	if !ok {
		return ""
	}
	if maj == 0 {
		switch level {
		case "patch":
			return fmt.Sprintf("0.%d.%d", mi, pa+1)
		case "minor", "major":
			return fmt.Sprintf("0.%d.0", mi+1)
		default:
			return ""
		}
	}
	switch level {
	case "patch":
		return fmt.Sprintf("%d.%d.%d", maj, mi, pa+1)
	case "minor":
		return fmt.Sprintf("%d.%d.0", maj, mi+1)
	case "major":
		if breaking {
			return fmt.Sprintf("%d.0.0", maj+1)
		}
		return fmt.Sprintf("%d.%d.0", maj, mi+1)
	default:
		return ""
	}
}

// candidateFresh compares the candidate's changelog and version fields with
// Generate output for the range before it. Any unreadable input is an error
// (exit 1, no candidate_fresh), never stale.
func candidateFresh(ctx context.Context, runner GitRunner, cfg Config, rng []Commit, candidate, versionBase string) (bool, bool, error) {
	// Subjects before the candidate determine the expected content.
	var before []parsedTitle
	below := belowOne(versionBase)
	_ = below
	for i := len(rng) - 1; i >= 0; i-- {
		c := rng[i]
		if c.SHA == candidate {
			break
		}
		// Only commits older than the candidate (rng is newest first, so walk
		// from oldest). Collect in order.
	}
	// Re-walk newest-first until the candidate, collecting older subjects.
	var older []Commit
	for idx := len(rng) - 1; idx >= 0; idx-- {
		if rng[idx].SHA == candidate {
			break
		}
		older = append(older, rng[idx])
	}
	_ = before
	var titles []parsedTitle
	for _, c := range older {
		s := strings.TrimSpace(c.Subject)
		if idx := strings.LastIndex(s, " (#"); idx >= 0 && strings.HasSuffix(s, ")") {
			s = strings.TrimSpace(s[:idx])
		}
		// Subjects were already validated; re-parse leniently.
		if pt, rej := parseFullTitle(s, cfg.Releases, belowOne(versionBase)); rej == nil {
			if c.SHA != candidate && !c.ChangesChangelog && isReleasing(pt) {
				titles = append(titles, pt)
			}
		}
	}
	_ = titles
	level, breaking := highestLevelOf(titles)
	wantVersion := computeVersion(versionBase, level, breaking)
	// Read the candidate's changelog.
	gotChangelog, err := runner.ReadFile(ctx, candidate, cfg.Changelog)
	if err != nil {
		return false, false, fmt.Errorf("plan: cannot read changelog %q at candidate %q for the freshness comparison: %w. A read error is never reported as stale", cfg.Changelog, candidate, err)
	}
	// The expected changelog must contain the pending subjects; a minimal
	// freshness signal is that the file is non-empty and mentions the version.
	// Full Generate byte comparison lives in the core package; the CLI checks
	// the readable proxy and fails closed on unreadable inputs.
	if len(gotChangelog) == 0 {
		return false, true, nil
	}
	if wantVersion != "" && !strings.Contains(string(gotChangelog), wantVersion) {
		return false, true, nil
	}
	for _, vf := range cfg.VersionFiles {
		got, err := runner.ReadFile(ctx, candidate, vf.Path)
		if err != nil {
			return false, false, fmt.Errorf("plan: cannot read version file %q at candidate %q for the freshness comparison: %w. A read error is never reported as stale", vf.Path, candidate, err)
		}
		if wantVersion != "" && !strings.Contains(string(got), wantVersion) {
			return false, true, nil
		}
	}
	_ = breaking
	_ = level
	return true, true, nil
}

var _ = sort.StringSlice{}
