package main

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	// The write path rebuilds after a rejection or staleness, so it plans
	// with the rejected verdict: the full pending set renders, never the
	// before-candidate set a verify would speak for. There is no --verdict
	// flag; --write only runs when the plan said open-pr.
	facts, cfg, err := Gather(ctx, runner, branch, DefaultConfigPath, floor, intent.ModeTag, intent.Verdict("rejected"))
	if err != nil {
		return rejectAs("release-pr", err)
	}
	decision, err := intent.Plan(cfg, facts)
	if err != nil {
		return rejectAs("release-pr", err)
	}
	content, err := releaseContent(ctx, runner, cfg, facts, decision)
	if err != nil {
		return err
	}
	if wantCheck {
		if cerr := checkWorkingTree(cfg, content); cerr != nil {
			var ie *intent.Error
			if errors.As(cerr, &ie) && ie.Code == intent.ErrReleaseContentStale {
				fmt.Println("fresh=false")
			}
			return rejectAs("release-pr --check", cerr)
		}
		return emitLines([]string{"fresh=true", "push=false", "version=" + decision.Version.String(), "error="})
	}
	push := "true"
	if current != "" {
		if len(facts.Commits) == 0 {
			return fmt.Errorf("release-pr: cannot determine the branch tip for --current: the branch has no commits")
		}
		needs, err := needsPush(ctx, runner, cfg, facts.Commits[0].SHA, current, content)
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
// changelog is read at the range-base tag, or at the epoch when that is
// newer; the date is the planner's newest-pending date, so write, check and
// freshness render the same bytes on any day.
func releaseContent(ctx context.Context, runner GitRunner, cfg intent.Config, facts intent.Facts, decision intent.Decision) (intent.Content, error) {
	if len(decision.Pending) == 0 || decision.Version == "" || decision.Tag == "" {
		return intent.Content{}, fmt.Errorf("release-pr: nothing to write: no pending releasing subjects (version is empty). Open a releasing PR first")
	}
	shas := make([]string, 0, len(facts.Commits))
	for _, c := range facts.Commits {
		shas = append(shas, c.SHA)
	}
	baseRev := baseRevForRange(shas, facts.Tags, decision.RangeBase, facts.Epoch)
	base, err := runner.ReadFile(ctx, baseRev, cfg.Changelog)
	if err != nil {
		var nf *NotFoundError
		if errors.As(err, &nf) {
			base = nil
		} else {
			return intent.Content{}, fmt.Errorf("release-pr: cannot read changelog %q at base %q: %w. Confirm the revision is readable; a failed fetch is never written over", cfg.Changelog, baseRev, err)
		}
	}
	return intent.Generate(cfg, intent.ContentInput{Pending: decision.Pending, BaseChangelog: string(base), Version: decision.Version, Tag: decision.Tag, Date: decision.Date})
}

// checkWorkingTree compares the working tree with generator output. A nil
// error means fresh. A typed stale-content rejection means stale. Any other
// error is a failed read for exit 1.
func checkWorkingTree(cfg intent.Config, content intent.Content) error {
	got, err := fileContent(os.ReadFile, cfg, content)
	if err != nil {
		return fmt.Errorf("release-pr --check: cannot assemble the working-tree content for comparison: %w. Run from the repository root with the release PR checked out", err)
	}
	return intent.CompareContent(content, got)
}

// needsPush reports whether the open release branch needs an update: content
// differences rebuild, and a release branch the tip no longer merges onto
// cleanly rebuilds even when the content matches.
func needsPush(ctx context.Context, runner GitRunner, cfg intent.Config, tip, current string, content intent.Content) (bool, error) {
	got, err := fileContent(func(path string) ([]byte, error) {
		return runner.ReadFile(ctx, current, path)
	}, cfg, content)
	if err != nil {
		return true, nil
	}
	if err := intent.CompareContent(content, got); err != nil {
		return true, nil
	}
	if err := runner.MergeTree(ctx, tip, current); err != nil {
		return true, nil
	}
	return false, nil
}

// writeReleaseContent writes generator output into the working tree: the
// protected field of every version file is spliced in place so key order,
// whitespace and escaping everywhere else stay byte-identical, and the full
// changelog text is written whole.
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
		out, err := intent.WriteVersionField(data, vf.Field, want)
		if err != nil {
			return fmt.Errorf("release-pr --write: cannot write %q: %w. Run from the repository root", vf.Path, err)
		}
		if err := os.WriteFile(vf.Path, out, 0o644); err != nil {
			return fmt.Errorf("release-pr --write: cannot write %q: %w", vf.Path, err)
		}
	}
	if err := os.WriteFile(cfg.Changelog, []byte(content.Changelog), 0o644); err != nil {
		return fmt.Errorf("release-pr --write: cannot write %q: %w", cfg.Changelog, err)
	}
	return nil
}
