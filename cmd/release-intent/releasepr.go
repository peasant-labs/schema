package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/peasant-labs/schema/internal/release/intent"
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
	facts, cfg, err := Gather(ctx, runner, branch, DefaultConfigPath, floor, intent.ModeTag, intent.Verdict(""))
	if err != nil {
		return intentErrorToReject(err)
	}
	decision, err := intent.Plan(cfg, facts)
	if err != nil {
		return intentErrorToReject(err)
	}
	content, err := releaseContent(ctx, runner, cfg, branch, decision)
	if err != nil {
		return err
	}
	if wantCheck {
		fresh, err := checkWorkingTree(cfg, decision, content)
		if err != nil {
			return err
		}
		if !fresh {
			return &RejectError{
				Code: intent.ErrReleaseContentStale, What: "the release PR content is stale",
				Why: "releasing commits landed after the release PR base, or the checked-out content differs from the generator", Where: "release PR branch, checked by release-intent release-pr --check",
				When: "title-guard on the release PR, or the merge-group build", Means: "the required check is red; the release PR must be rebuilt",
				Fix: "run the release job to rebuild the release PR, then requeue",
			}
		}
		return emitLines([]string{"fresh=true", "push=false", "version=" + decision.Version.String(), "error="})
	}
	push := "true"
	if current != "" {
		needs, err := needsPush(ctx, runner, cfg, current, content)
		if err != nil {
			return err
		}
		if !needs {
			push = "false"
		}
	}
	if err := writeReleaseContent(cfg, content); err != nil {
		return err
	}
	return emitLines([]string{"fresh=true", "push=" + push, "version=" + decision.Version.String(), "error="})
}

// releaseContent renders the generator output for the plan decision. The base
// changelog comes from the branch tip; the date is today since Generate only
// uses it for a dated heading style.
func releaseContent(ctx context.Context, runner GitRunner, cfg intent.Config, branch string, decision intent.Decision) (intent.Content, error) {
	if len(decision.Pending) == 0 || decision.Version == "" || decision.Tag == "" {
		return intent.Content{}, fmt.Errorf("release-pr: nothing to write: no pending releasing subjects (version is empty). Open a releasing PR first")
	}
	base, err := runner.ReadFile(ctx, branch, cfg.Changelog)
	if err != nil {
		// A missing changelog at the tip starts from an empty base.
		base = nil
	}
	today := time.Now().Format("2006-01-02")
	return intent.Generate(cfg, intent.ContentInput{Pending: decision.Pending, BaseChangelog: string(base), Version: decision.Version, Tag: decision.Tag, Date: today})
}

// checkWorkingTree compares the working tree with generator output.
func checkWorkingTree(cfg intent.Config, decision intent.Decision, content intent.Content) (bool, error) {
	if len(decision.Pending) == 0 {
		return false, nil
	}
	for _, vf := range cfg.VersionFiles {
		data, err := os.ReadFile(vf.Path)
		if err != nil {
			return false, fmt.Errorf("release-pr --check: cannot read working-tree file %q: %w. Run from the repository root with the release PR checked out", vf.Path, err)
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return false, nil
		}
		if v, _ := obj[vf.Field].(string); v != content.Versions[vf.Path] {
			return false, nil
		}
	}
	changelog, err := os.ReadFile(cfg.Changelog)
	if err != nil {
		return false, fmt.Errorf("release-pr --check: cannot read working-tree file %q: %w. Run from the repository root with the release PR checked out", cfg.Changelog, err)
	}
	if string(changelog) != content.Changelog {
		return false, nil
	}
	return true, nil
}

// needsPush reports whether the open release branch needs an update.
func needsPush(ctx context.Context, runner GitRunner, cfg intent.Config, current string, content intent.Content) (bool, error) {
	for _, vf := range cfg.VersionFiles {
		data, err := runner.ReadFile(ctx, current, vf.Path)
		if err != nil {
			return true, nil
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return true, nil
		}
		if v, _ := obj[vf.Field].(string); v != content.Versions[vf.Path] {
			return true, nil
		}
	}
	changelog, err := runner.ReadFile(ctx, current, cfg.Changelog)
	if err != nil {
		return true, nil
	}
	if string(changelog) != content.Changelog {
		return true, nil
	}
	return false, nil
}

// writeReleaseContent writes generator output into the working tree: the
// protected field of every version file and the full changelog text.
func writeReleaseContent(cfg intent.Config, content intent.Content) error {
	for _, vf := range cfg.VersionFiles {
		want, ok := content.Versions[vf.Path]
		if !ok {
			continue
		}
		data, err := os.ReadFile(vf.Path)
		if err != nil {
			return fmt.Errorf("release-pr --write: cannot read %q: %w. Run from the repository root", vf.Path, err)
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("release-pr --write: cannot parse %q as JSON: %w. Fix the file or drop the edit", vf.Path, err)
		}
		obj[vf.Field] = want
		out, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return fmt.Errorf("release-pr --write: cannot encode %q: %w", vf.Path, err)
		}
		out = append(out, '\n')
		if err := os.WriteFile(vf.Path, out, 0o644); err != nil {
			return fmt.Errorf("release-pr --write: cannot write %q: %w", vf.Path, err)
		}
	}
	if err := os.WriteFile(cfg.Changelog, []byte(content.Changelog), 0o644); err != nil {
		return fmt.Errorf("release-pr --write: cannot write %q: %w", cfg.Changelog, err)
	}
	return nil
}
