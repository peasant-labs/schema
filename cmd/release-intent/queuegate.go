package main

import (
	"context"
	"fmt"
	"os"

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
		return rejectAs("queue-gate", err)
	}
	cfgData, err := os.ReadFile(DefaultConfigPath)
	if err != nil {
		return configReadError(DefaultConfigPath, "", err)
	}
	cfg, err := intent.ParseConfig(cfgData)
	if err != nil {
		return rejectAs("queue-gate", configInvalidError(DefaultConfigPath, err))
	}
	current, err := queueCurrent(cfg)
	if err != nil {
		return err
	}
	gate, err := intent.QueueGate(cfg, current, snapshot, releasePR)
	if err != nil {
		return rejectAs("queue-gate", err)
	}
	return emitLines([]string{"gate=" + gate.String()})
}

// queueCurrent reads the current version from git tags for the hold rule. A
// read failure exits 1 with no gate line: nothing is armed or disarmed.
func queueCurrent(cfg intent.Config) (intent.SemVer, error) {
	ctx := context.Background()
	runner, err := newExecGit()
	if err != nil {
		return "", err
	}
	tags, err := runner.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return "", fmt.Errorf("queue-gate: cannot list tags with prefix %q: %w", cfg.TagPrefix, err)
	}
	_, version, ok := intent.HighestFinalTag(tags, cfg.TagPrefix)
	if !ok {
		return "", fmt.Errorf("queue-gate: no final release tag found for the hold rule")
	}
	return version, nil
}
