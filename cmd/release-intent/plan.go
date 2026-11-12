package main

import (
	"context"

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
	facts, cfg, err := Gather(ctx, runner, branch, DefaultConfigPath, floor, intent.ParseMode(mode), intent.Verdict(verdict))
	if err != nil {
		return intentErrorToReject(err)
	}
	decision, err := intent.Plan(cfg, facts)
	if err != nil {
		return intentErrorToReject(err)
	}
	lines := []string{
		"action=" + decision.Action.String(),
		"range_base=" + decision.RangeBase,
		"version_base=" + decision.VersionBase,
		"candidate=" + decision.Candidate,
	}
	if decision.Candidate == "" {
		lines = append(lines, "candidate_fresh=")
	} else {
		switch decision.CandidateFresh {
		case intent.FreshnessFresh:
			lines = append(lines, "candidate_fresh=true")
		case intent.FreshnessStale:
			lines = append(lines, "candidate_fresh=false")
		default:
			lines = append(lines, "candidate_fresh=")
		}
	}
	lines = append(lines,
		"level="+decision.Level.String(),
		"breaking="+boolString(decision.Breaking),
		"version="+decision.Version.String(),
		"tag="+decision.Tag,
		"error=",
	)
	return emitLines(lines)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
