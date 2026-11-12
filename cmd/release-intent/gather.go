package main

import (
	"context"
	"errors"
	"fmt"

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
	cfg, err := intent.ParseConfig(cfgData)
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
	epoch, err := r.FirstCommitAdding(ctx, branch, configPath)
	if err != nil {
		return intent.Facts{}, intent.Config{}, fmt.Errorf("plan: cannot find the commit adding %q on %q: %w", configPath, branch, err)
	}
	base := intent.Facts{Branch: branch, Commits: commits, Tags: tags, Epoch: epoch, Floor: floor, Mode: mode, Verdict: verdict, Fresh: map[string]bool{}}
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
	fresh, ferr := candidateFresh(ctx, r, cfg, commits, tags, dec, epoch)
	if ferr != nil {
		return intent.Facts{}, intent.Config{}, ferr
	}
	base.Fresh[dec.Candidate] = fresh
	return base, cfg, nil
}

// baseRevForRange returns the revision the release content is generated from:
// the range-base tag commit, or the epoch commit when the epoch is newer than
// the range base on the branch history.
func baseRevForRange(shas []string, tags map[string]string, rangeBaseTag, epoch string) string {
	index := map[string]int{}
	for i, sha := range shas {
		if _, ok := index[sha]; !ok {
			index[sha] = i
		}
	}
	rangeIdx, rok := index[tags[rangeBaseTag]]
	epochIdx, eok := index[epoch]
	if !rok {
		return epoch
	}
	if !eok {
		return tags[rangeBaseTag]
	}
	if epochIdx < rangeIdx {
		return epoch
	}
	return tags[rangeBaseTag]
}

// candidateFresh compares the candidate's changelog and version fields with
// what Generate produces for the candidate's own sub-range, which the planner
// reports in Before. Any unreadable input is an error (exit 1, no
// candidate_fresh), never stale.
func candidateFresh(ctx context.Context, r GitRunner, cfg intent.Config, commits []intent.Commit, tags map[string]string, dec intent.Decision, epoch string) (bool, error) {
	shas := make([]string, 0, len(commits))
	for _, c := range commits {
		shas = append(shas, c.SHA)
	}
	baseRev := baseRevForRange(shas, tags, dec.RangeBase, epoch)
	baseData, err := r.ReadFile(ctx, baseRev, cfg.Changelog)
	if err != nil {
		var nf *NotFoundError
		if errors.As(err, &nf) {
			baseData = nil
		} else {
			return false, fmt.Errorf("plan: cannot read changelog %q at base %q for the freshness comparison: %w. A read error is never reported as stale", cfg.Changelog, baseRev, err)
		}
	}
	if len(dec.Before.Pending) == 0 {
		return false, nil
	}
	content, err := intent.Generate(cfg, intent.ContentInput{Pending: dec.Before.Pending, BaseChangelog: string(baseData), Version: dec.Before.Version, Tag: dec.Before.Tag, Date: dec.Before.Date})
	if err != nil {
		return false, fmt.Errorf("plan: cannot render the expected release content for the freshness comparison: %w. A read error is never reported as stale", err)
	}
	got, err := fileContent(func(path string) ([]byte, error) {
		return r.ReadFile(ctx, dec.Candidate, path)
	}, cfg, content)
	if err != nil {
		return false, fmt.Errorf("plan: cannot assemble the candidate content for the freshness comparison: %w. A read error is never reported as stale", err)
	}
	if err := intent.CompareContent(content, got); err != nil {
		return false, nil
	}
	return true, nil
}
