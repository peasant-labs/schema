package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// runQueueGate implements: release-intent queue-gate --queue FILE --release-pr N.
func runQueueGate(args []string) error {
	var queuePath string
	var releasePR int
	var haveQueue, havePR bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--queue":
			val, next, err := requireFlagValue(args, i, "--queue")
			if err != nil {
				return err
			}
			queuePath = val
			haveQueue = true
			i = next
		case "--release-pr":
			val, next, err := requireFlagValue(args, i, "--release-pr")
			if err != nil {
				return err
			}
			var n int
			if _, err := fmt.Sscanf(val, "%d", &n); err != nil || n <= 0 {
				return usageErr("usage: release-intent queue-gate --queue FILE --release-pr N (bad --release-pr %q): pass the release PR number as a positive integer", val)
			}
			releasePR = n
			havePR = true
			i = next
		default:
			return usageErr("unknown flag %q for queue-gate: expected --queue FILE --release-pr N", args[i])
		}
	}
	if !haveQueue {
		return usageErr("usage: release-intent queue-gate --queue FILE --release-pr N (missing required --queue): pass the file holding the raw queue response")
	}
	if !havePR {
		return usageErr("usage: release-intent queue-gate --queue FILE --release-pr N (missing required --release-pr): pass the release PR number")
	}
	data, err := os.ReadFile(queuePath)
	if err != nil {
		return fmt.Errorf("queue-gate: cannot read queue file %q: %v. The workflow must write the raw queue response to this path before calling queue-gate", queuePath, err)
	}
	snapshot, err := intent.DecodeQueueResponse(data)
	if err != nil {
		return intentErrorToReject(err)
	}
	cfg, err := intent.LoadConfig(DefaultConfigPath)
	if err != nil {
		return fmt.Errorf("queue-gate: cannot read release config %q for the hold rule: %w", DefaultConfigPath, err)
	}
	current, err := queueCurrent()
	if err != nil {
		return err
	}
	gate, err := intent.QueueGate(cfg, current, snapshot.Entries, releasePR)
	if err != nil {
		return intentErrorToReject(err)
	}
	return emitLines([]string{"gate=" + gate.String()})
}

// queueCurrent reads the current version from git tags for the hold rule. A
// read failure exits 1 with no gate line: nothing is armed or disarmed.
func queueCurrent() (intent.SemVer, error) {
	ctx := context.Background()
	runner, err := newExecGit()
	if err != nil {
		return "", err
	}
	cfg, err := intent.LoadConfig(DefaultConfigPath)
	if err != nil {
		return "", fmt.Errorf("queue-gate: cannot read release config %q for the hold rule: %w", DefaultConfigPath, err)
	}
	tags, err := runner.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return "", fmt.Errorf("queue-gate: cannot list tags with prefix %q: %w", cfg.TagPrefix, err)
	}
	_, ver := highestFinalTag(tags, cfg.TagPrefix)
	if ver == "" {
		return "", fmt.Errorf("queue-gate: no final release tag found for the hold rule")
	}
	sv, err := intent.ParseSemVer(ver)
	if err != nil {
		return "", fmt.Errorf("queue-gate: highest tag version %q does not parse: %w", ver, err)
	}
	return sv, nil
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
		v, ok := strings.CutPrefix(t, prefix)
		if !ok {
			continue
		}
		if strings.Contains(v, "-rc") {
			continue
		}
		var maj, mi, pa int
		if _, err := fmt.Sscanf(v, "%d.%d.%d", &maj, &mi, &pa); err != nil {
			continue
		}
		if _, err := intent.ParseSemVer(v); err != nil {
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
