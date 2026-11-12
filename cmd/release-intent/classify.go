package main

import (
	"context"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// runClassify implements: release-intent classify --author A --head-branch B --rev REV.
func runClassify(args []string) error {
	var author, headBranch, rev string
	var haveAuthor, haveBranch, haveRev bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--author":
			val, next, err := requireFlagValue(args, i, "--author")
			if err != nil {
				return err
			}
			author = val
			haveAuthor = true
			i = next
		case "--head-branch":
			val, next, err := requireFlagValue(args, i, "--head-branch")
			if err != nil {
				return err
			}
			headBranch = val
			haveBranch = true
			i = next
		case "--rev":
			val, next, err := requireFlagValue(args, i, "--rev")
			if err != nil {
				return err
			}
			rev = val
			haveRev = true
			i = next
		default:
			return usageErr("unknown flag %q for classify: expected --author LOGIN --head-branch BRANCH --rev REV", args[i])
		}
	}
	if !haveAuthor {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH --rev REV (missing required --author): pass the PR author login")
	}
	if !haveBranch {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH --rev REV (missing required --head-branch): pass the PR head branch name")
	}
	if !haveRev {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH --rev REV (missing required --rev): pass the base revision whose release config judges the PR")
	}
	ctx := context.Background()
	runner, err := newExecGit()
	if err != nil {
		return err
	}
	cfgData, err := runner.ReadFile(ctx, rev, DefaultConfigPath)
	if err != nil {
		return configReadError(DefaultConfigPath, rev, err)
	}
	cfg, err := intent.ParseConfig(cfgData)
	if err != nil {
		return configInvalid("classify", "release config "+DefaultConfigPath+" at "+rev, err)
	}
	kind := intent.ClassifyPR(cfg, author, headBranch)
	if err := emitLines([]string{"kind=" + kind.String()}); err != nil {
		return err
	}
	return nil
}
