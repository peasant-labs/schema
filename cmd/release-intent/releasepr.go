package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		fresh, err := checkWorkingTree(cfg, decision)
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
	push := "true"
	if current != "" {
		needs, err := needsPush(ctx, runner, cfg, current, decision)
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

// checkWorkingTree compares the working tree's release content with the
// computed version: every version file must carry the version in its field,
// and the changelog must mention it. An empty version is stale.
func checkWorkingTree(cfg intent.Config, decision PlanDecision) (bool, error) {
	if decision.Version == "" {
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
		if v, _ := obj[vf.Field].(string); v != decision.Version {
			return false, nil
		}
	}
	changelog, err := os.ReadFile(cfg.Changelog)
	if err != nil {
		return false, fmt.Errorf("release-pr --check: cannot read working-tree file %q: %w. Run from the repository root with the release PR checked out", cfg.Changelog, err)
	}
	if !strings.Contains(string(changelog), decision.Version) {
		return false, nil
	}
	return true, nil
}

// needsPush reports whether the open release branch needs an update: false
// when that branch already holds exactly the computed content.
func needsPush(ctx context.Context, runner GitRunner, cfg intent.Config, current string, decision PlanDecision) (bool, error) {
	if decision.Version == "" {
		return true, nil
	}
	for _, vf := range cfg.VersionFiles {
		data, err := runner.ReadFile(ctx, current, vf.Path)
		if err != nil {
			return true, nil
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return true, nil
		}
		if v, _ := obj[vf.Field].(string); v != decision.Version {
			return true, nil
		}
	}
	changelog, err := runner.ReadFile(ctx, current, cfg.Changelog)
	if err != nil {
		return true, nil
	}
	if !strings.Contains(string(changelog), decision.Version) {
		return true, nil
	}
	return false, nil
}

// writeReleaseContent writes the computed version into the working tree: the
// protected field of every version file and a changelog entry mentioning the
// version. The full section rendering moves to the contract package with
// Generate; the command layer writes the versioned fields and the version
// marker so the content is real and checkable.
func writeReleaseContent(cfg intent.Config, decision PlanDecision) error {
	if decision.Version == "" {
		return fmt.Errorf("release-pr: nothing to write: no pending releasing subjects (version is empty). Open a releasing PR first")
	}
	for _, vf := range cfg.VersionFiles {
		data, err := os.ReadFile(vf.Path)
		if err != nil {
			return fmt.Errorf("release-pr --write: cannot read %q: %w. Run from the repository root", vf.Path, err)
		}
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("release-pr --write: cannot parse %q as JSON: %w. Fix the file or drop the edit", vf.Path, err)
		}
		obj[vf.Field] = decision.Version
		out, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return fmt.Errorf("release-pr --write: cannot encode %q: %w", vf.Path, err)
		}
		out = append(out, '\n')
		if err := os.WriteFile(vf.Path, out, 0o644); err != nil {
			return fmt.Errorf("release-pr --write: cannot write %q: %w", vf.Path, err)
		}
	}
	changelog, err := os.ReadFile(cfg.Changelog)
	if err != nil {
		return fmt.Errorf("release-pr --write: cannot read %q: %w. Run from the repository root", cfg.Changelog, err)
	}
	if !strings.Contains(string(changelog), decision.Version) {
		entry := "## " + decision.Tag + "\n\nPending release content.\n\n"
		updated := append([]byte(entry), changelog...)
		if err := os.WriteFile(cfg.Changelog, updated, 0o644); err != nil {
			return fmt.Errorf("release-pr --write: cannot write %q: %w", cfg.Changelog, err)
		}
	}
	return nil
}
