package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Mode is the closed set of plan modes.
type Mode string

const (
	ModePlanOnly Mode = "plan-only"
	ModeTag      Mode = "tag"
)

// ParseMode maps anything but "tag" to ModePlanOnly.
func ParseMode(s string) Mode {
	if s == "tag" {
		return ModeTag
	}
	return ModePlanOnly
}

// Facts are the git inputs Plan needs.
type Facts struct {
	Commits   []Commit
	Tags      map[string]string
	Epoch     string
	Floor     string
	Mode      Mode
	Verdict   string
	TagPrefix string
}

// Gather collects commits, tags and the epoch through the runner. A read
// failure returns an error so the caller exits 1 with no candidate_fresh line:
// a read error is never reported as stale.
func Gather(ctx context.Context, r GitRunner, branch, changelog, configPath, floor string, mode Mode, verdict string) (Facts, Config, error) {
	cfgData, err := r.ReadFile(ctx, branch, configPath)
	if err != nil {
		return Facts{}, Config{}, fmt.Errorf("plan: cannot read release config %q at %q: %w. Confirm the branch has the config file", configPath, branch, err)
	}
	cfg, err := ParseConfig(cfgData, configPath)
	if err != nil {
		return Facts{}, Config{}, err
	}
	commits, err := r.FirstParent(ctx, branch, cfg.Changelog)
	if err != nil {
		return Facts{}, Config{}, fmt.Errorf("plan: cannot list first-parent commits of %q: %w. Confirm the branch exists", branch, err)
	}
	tags, err := r.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return Facts{}, Config{}, fmt.Errorf("plan: cannot list tags with prefix %q: %w", cfg.TagPrefix, err)
	}
	epoch, err := r.FirstCommitAdding(ctx, configPath)
	if err != nil {
		return Facts{}, Config{}, fmt.Errorf("plan: cannot find the commit adding %q: %w", configPath, err)
	}
	return Facts{Commits: commits, Tags: tags, Epoch: epoch, Floor: floor, Mode: mode, Verdict: verdict, TagPrefix: cfg.TagPrefix}, cfg, nil
}

// highestFinalTag returns the highest final (no -rcN) tag and its version.
func highestFinalTag(tags map[string]string, prefix string) (tag, version string) {
	type tv struct {
		tag         string
		version     string
		maj, mi, pa int
	}
	var finals []tv
	for t := range tags {
		v := strings.TrimPrefix(t, prefix)
		if v == t {
			continue
		}
		if strings.Contains(v, "-rc") {
			continue
		}
		maj, mi, pa, ok := parseSemVer(v)
		if !ok {
			continue
		}
		finals = append(finals, tv{t, v, maj, mi, pa})
	}
	if len(finals) == 0 {
		return "", ""
	}
	sort.Slice(finals, func(i, j int) bool {
		if finals[i].maj != finals[j].maj {
			return finals[i].maj > finals[j].maj
		}
		if finals[i].mi != finals[j].mi {
			return finals[i].mi > finals[j].mi
		}
		return finals[i].pa > finals[j].pa
	})
	return finals[0].tag, finals[0].version
}

// belowOne reports whether the version base is below 1.0.
func belowOne(version string) bool {
	maj, _, _, ok := parseSemVer(version)
	if !ok {
		return true
	}
	return maj == 0
}
