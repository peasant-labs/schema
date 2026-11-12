package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// Gather collects commits, tags and the epoch through the runner and fills
// per-commit freshness through intent.Generate. A read failure returns an
// error so the caller exits 1 with no candidate_fresh line: a read error is
// never reported as stale.
func Gather(ctx context.Context, r GitRunner, branch, configPath, floor string, mode intent.Mode, verdict intent.Verdict) (intent.Facts, intent.Config, error) {
	cfgData, err := r.ReadFile(ctx, branch, configPath)
	if err != nil {
		return intent.Facts{}, intent.Config{}, fmt.Errorf("plan: cannot read release config %q at %q: %w. Confirm the branch has the config file", configPath, branch, err)
	}
	cfg, err := loadConfigBytes(cfgData)
	if err != nil {
		return intent.Facts{}, intent.Config{}, err
	}
	commits, err := r.FirstParent(ctx, branch, cfg.Changelog)
	if err != nil {
		return intent.Facts{}, intent.Config{}, fmt.Errorf("plan: cannot list first-parent commits of %q: %w. Confirm the branch exists", branch, err)
	}
	tags, err := r.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return intent.Facts{}, intent.Config{}, fmt.Errorf("plan: cannot list tags with prefix %q: %w", cfg.TagPrefix, err)
	}
	epoch, err := r.FirstCommitAdding(ctx, configPath)
	if err != nil {
		return intent.Facts{}, intent.Config{}, fmt.Errorf("plan: cannot find the commit adding %q: %w", configPath, err)
	}
	intentCommits := make([]intent.Commit, 0, len(commits))
	for _, c := range commits {
		intentCommits = append(intentCommits, intent.Commit{SHA: c.SHA, Subject: c.Subject, ChangesChangelog: c.ChangesChangelog})
	}
	base := intent.Facts{Commits: intentCommits, Tags: tags, Epoch: epoch, Floor: floor, Mode: mode, Verdict: verdict, Fresh: map[string]bool{}}
	// Find the candidate with a first pass that carries no freshness, then
	// record freshness for it through Generate. Plan only reads the entry for
	// the commit it picks, so one entry suffices.
	dec, err := intent.Plan(cfg, base)
	if err != nil {
		// A planner rejection (bad subject, missing base, bad floor) returns
		// as-is; the caller maps it to exit 2 with error=<code>.
		return intent.Facts{}, intent.Config{}, err
	}
	if dec.Candidate == "" {
		return base, cfg, nil
	}
	fresh, ferr := candidateFresh(ctx, r, cfg, commits, tags, epoch, floor, dec.Candidate)
	if ferr != nil {
		return intent.Facts{}, intent.Config{}, ferr
	}
	base.Fresh[dec.Candidate] = fresh
	return base, cfg, nil
}

// candidateFresh compares the candidate's changelog and version fields with
// what Generate produces for the range ending just before it. Any unreadable
// input is an error (exit 1, no candidate_fresh), never stale.
func candidateFresh(ctx context.Context, r GitRunner, cfg intent.Config, commits []Commit, tags map[string]string, epoch, floor, candidate string) (bool, error) {
	idx := -1
	for i, c := range commits {
		if c.SHA == candidate {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false, fmt.Errorf("plan: cannot find candidate %q in the branch history for the freshness comparison", candidate)
	}
	parent := ""
	if idx+1 < len(commits) {
		parent = commits[idx+1].SHA
	}
	// Sub-range before the candidate, reusing Plan for range, parsing and
	// arithmetic instead of re-deriving them here.
	before := append([]intent.Commit(nil), intentCommitsBefore(commits, idx)...)
	sub, err := intent.Plan(cfg, intent.Facts{Commits: before, Tags: tags, Epoch: epoch, Floor: floor, Mode: intent.ModePlanOnly, Fresh: map[string]bool{}})
	if err != nil {
		// The full range already parsed, so a sub-range failure means the
		// candidate has no generatable base: treat as stale, not a read error.
		// A genuine read error surfaces below when files are read.
		_ = err
		return false, nil
	}
	if len(sub.Pending) == 0 || sub.Version == "" || sub.Tag == "" {
		return false, nil
	}
	baseChangelog := ""
	if parent != "" {
		data, err := r.ReadFile(ctx, parent, cfg.Changelog)
		if err != nil {
			return false, fmt.Errorf("plan: cannot read changelog %q at %q for the freshness comparison: %w. A read error is never reported as stale", cfg.Changelog, parent, err)
		}
		baseChangelog = string(data)
	}
	date := newestPendingDate(commits, idx, sub.Pending)
	content, err := intent.Generate(cfg, intent.ContentInput{Pending: sub.Pending, BaseChangelog: baseChangelog, Version: sub.Version, Tag: sub.Tag, Date: date})
	if err != nil {
		return false, nil
	}
	gotChangelog, err := r.ReadFile(ctx, candidate, cfg.Changelog)
	if err != nil {
		return false, fmt.Errorf("plan: cannot read changelog %q at candidate %q for the freshness comparison: %w. A read error is never reported as stale", cfg.Changelog, candidate, err)
	}
	if string(gotChangelog) != content.Changelog {
		return false, nil
	}
	for _, vf := range cfg.VersionFiles {
		data, err := r.ReadFile(ctx, candidate, vf.Path)
		if err != nil {
			return false, fmt.Errorf("plan: cannot read version file %q at candidate %q for the freshness comparison: %w. A read error is never reported as stale", vf.Path, candidate, err)
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return false, nil
		}
		if v, _ := obj[vf.Field].(string); v != content.Versions[vf.Path] {
			return false, nil
		}
	}
	_ = tags
	return true, nil
}

func intentCommitsBefore(commits []Commit, idx int) []intent.Commit {
	out := make([]intent.Commit, 0, len(commits)-idx-1)
	for _, c := range commits[idx+1:] {
		out = append(out, intent.Commit{SHA: c.SHA, Subject: c.Subject, ChangesChangelog: c.ChangesChangelog})
	}
	return out
}

func newestPendingDate(commits []Commit, idx int, pending []intent.Title) string {
	// The newest pending subject in commit order carries the date for a dated
	// heading. Match by subject since titles carry no SHA.
	bySubject := map[string]int{}
	for _, t := range pending {
		bySubject[t.Subject]++
	}
	for _, c := range commits[idx+1:] {
		if c.ChangesChangelog || c.Date == "" {
			continue
		}
		for _, t := range pending {
			if t.Subject == strings.TrimSpace(trimLandedSuffix(c.Subject)) && bySubject[t.Subject] > 0 {
				bySubject[t.Subject]--
				return c.Date
			}
		}
	}
	return ""
}

func trimLandedSuffix(s string) string {
	t := strings.TrimSpace(s)
	if idx := strings.LastIndex(t, " (#"); idx >= 0 && strings.HasSuffix(t, ")") {
		t = strings.TrimSpace(t[:idx])
	}
	return t
}
