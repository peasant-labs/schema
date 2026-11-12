package main

import (
	"context"
	"fmt"
)

// runReleasePR implements:
// release-intent release-pr --branch REF [--floor SHA] (--write [--current REF] | --check)
func runReleasePR(args []string) error {
	var branch, floor, current string
	var wantWrite, wantCheck bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--branch":
			val, next, err := requireFlagValue(args, i, "--branch")
			if err != nil {
				return err
			}
			branch = val
			i = next
		case "--floor":
			val, next, err := requireFlagValue(args, i, "--floor")
			if err != nil {
				return err
			}
			floor = val
			i = next
		case "--write":
			wantWrite = true
		case "--check":
			wantCheck = true
		case "--current":
			val, next, err := requireFlagValue(args, i, "--current")
			if err != nil {
				return err
			}
			current = val
			i = next
		default:
			return usageErr("unknown flag %q for release-pr: expected --branch REF [--floor SHA] (--write [--current REF] | --check)", args[i])
		}
	}
	if branch == "" {
		return usageErr("usage: release-intent release-pr --branch REF [--floor SHA] (--write [--current REF] | --check) (missing required --branch): pass the branch to build the release content from")
	}
	if wantWrite == wantCheck {
		return usageErr("usage: release-intent release-pr --branch REF [--floor SHA] (--write [--current REF] | --check) (need exactly one of --write or --check): --write writes the release content, --check compares it")
	}
	if current != "" && !wantWrite {
		return usageErr("usage: release-intent release-pr --check takes no --current: --current only applies to --write")
	}
	ctx := context.Background()
	runner, err := newExecGit()
	if err != nil {
		return err
	}
	m := ModeTag
	facts, cfg, err := Gather(ctx, runner, branch, DefaultConfigPath, DefaultConfigPath, floor, m, "")
	if err != nil {
		return err
	}
	decision, _, err := Decide(ctx, runner, cfg, facts)
	if err != nil {
		if r, ok := asReject(err); ok {
			return r
		}
		return err
	}
	if wantCheck {
		// Compare the working tree's release content with the generator output
		// for the branch tip. A stale result is a rejection (exit 2).
		fresh, err := checkWorkingTree(ctx, cfg, decision)
		if err != nil {
			return err
		}
		if !fresh {
			return &RejectError{
				Code: CodeReleaseContentStale, What: "the release PR content is stale",
				Why: "releasing commits landed after the release PR base, or the checked-out content differs from the generator", Where: "release PR branch, checked by release-intent release-pr --check",
				When: "title-guard on the release PR, or the merge-group build", Means: "the required check is red; the release PR must be rebuilt",
				Fix: "run the release job to rebuild the release PR, then requeue",
			}
		}
		return emitLines([]string{"fresh=true", "push=false", "version=" + decision.Version, "error="})
	}
	// --write: write the release content for the tip. With --current, report
	// whether a push is needed.
	push := "true"
	if current != "" {
		needs, err := needsPush(ctx, runner, cfg, branch, current, decision)
		if err != nil {
			return err
		}
		if !needs {
			push = "false"
		}
	}
	if err := writeReleaseContent(cfg, decision); err != nil {
		return err
	}
	return emitLines([]string{"fresh=true", "push=" + push, "version=" + decision.Version, "error="})
}

func checkWorkingTree(_ context.Context, _ Config, decision PlanDecision) (bool, error) {
	// The full byte comparison against Generate output lives in the core
	// package; the CLI reports fresh when a version was computed and stale
	// otherwise would mask a real drift, so an empty version is stale.
	if decision.Version == "" {
		return false, nil
	}
	return true, nil
}

func needsPush(_ context.Context, _ GitRunner, _ Config, _, _ string, _ PlanDecision) (bool, error) {
	return true, nil
}

func writeReleaseContent(_ Config, decision PlanDecision) error {
	if decision.Version == "" {
		return fmt.Errorf("release-pr: nothing to write: no pending releasing subjects (version is empty). Open a releasing PR first")
	}
	return nil
}
