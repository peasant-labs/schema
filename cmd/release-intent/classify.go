package main

import (
	"github.com/peasant-labs/schema/internal/release/intent"
)

// runClassify implements: release-intent classify --author A --head-branch B.
func runClassify(args []string) error {
	var author, headBranch, configPath string
	configPath = DefaultConfigPath
	var haveAuthor, haveBranch bool
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
		case "--config":
			val, next, err := requireFlagValue(args, i, "--config")
			if err != nil {
				return err
			}
			configPath = val
			i = next
		default:
			return usageErr("unknown flag %q for classify: expected --author LOGIN --head-branch BRANCH [--config F]", args[i])
		}
	}
	if !haveAuthor {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH [--config F] (missing required --author): pass the PR author login")
	}
	if !haveBranch {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH [--config F] (missing required --head-branch): pass the PR head branch name")
	}
	cfg, err := intent.LoadConfig(configPath)
	if err != nil {
		return err
	}
	kind := intent.ClassifyPR(cfg, author, headBranch)
	if err := emitLines([]string{"kind=" + kind.String()}); err != nil {
		return err
	}
	return nil
}
