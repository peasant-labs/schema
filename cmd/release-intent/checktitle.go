package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// runCheckTitle implements:
// release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH
// with the PR title on stdin.
func runCheckTitle(args []string) error {
	var base, head, author, headBranch string
	var haveBase, haveHead, haveAuthor, haveBranch bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--base":
			val, next, err := requireFlagValue(args, i, "--base")
			if err != nil {
				return err
			}
			base, haveBase = val, true
			i = next
		case "--head":
			val, next, err := requireFlagValue(args, i, "--head")
			if err != nil {
				return err
			}
			head, haveHead = val, true
			i = next
		case "--author":
			val, next, err := requireFlagValue(args, i, "--author")
			if err != nil {
				return err
			}
			author, haveAuthor = val, true
			i = next
		case "--head-branch":
			val, next, err := requireFlagValue(args, i, "--head-branch")
			if err != nil {
				return err
			}
			headBranch, haveBranch = val, true
			i = next
		default:
			return usageErr("unknown flag %q for check-title: expected --base REV --head REV --author LOGIN --head-branch BRANCH", args[i])
		}
	}
	if !haveBase {
		return usageErr("usage: release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH (missing required --base): pass the PR base revision")
	}
	if !haveHead {
		return usageErr("usage: release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH (missing required --head): pass the PR head revision")
	}
	if !haveAuthor {
		return usageErr("usage: release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH (missing required --author): pass the PR author login")
	}
	if !haveBranch {
		return usageErr("usage: release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH (missing required --head-branch): pass the PR head branch name")
	}
	title, err := readTitleStdin()
	if err != nil {
		return err
	}
	ctx := context.Background()
	runner, err := newExecGit()
	if err != nil {
		return err
	}
	return checkTitle(ctx, runner, base, head, author, headBranch, title)
}

func readTitleStdin() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("check-title: cannot stat stdin for the PR title: %v. Pipe the title, e.g. echo 'feat(x)[minor]: ...' | release-intent check-title --base A --head B --author U --head-branch H", err)
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", usageErr("usage: release-intent check-title reads the PR title on stdin (got a terminal): pipe the title, e.g. echo 'feat(x)[minor]: ...' | release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH")
	}
	sc := bufio.NewScanner(os.Stdin)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("check-title: cannot read the PR title on stdin: %v", err)
	}
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			t := strings.TrimSpace(l)
			if idx := strings.LastIndex(t, " (#"); idx >= 0 && strings.HasSuffix(t, ")") {
				t = strings.TrimSpace(t[:idx])
			}
			return t, nil
		}
	}
	return "", usageErr("usage: release-intent check-title reads the PR title on stdin (got empty input): pipe the title, e.g. echo 'feat(x)[minor]: ...' | release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH")
}

func checkTitle(ctx context.Context, runner GitRunner, base, head, author, headBranch, title string) error {
	cfgData, err := runner.ReadFile(ctx, base, DefaultConfigPath)
	adopting := false
	if err != nil {
		headData, herr := runner.ReadFile(ctx, head, DefaultConfigPath)
		if herr != nil {
			return fmt.Errorf("check-title: cannot read release config %q at base %q or head %q: base read: %v; head read: %v. Confirm the base branch carries the config, or that this is the adopting PR adding it", DefaultConfigPath, base, head, err, herr)
		}
		cfgData = headData
		adopting = true
	}
	cfg, err := loadConfigBytes(cfgData)
	if err != nil {
		return fmt.Errorf("check-title: the base release config %w", err)
	}
	tags, err := runner.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return fmt.Errorf("check-title: cannot list tags with prefix %q: %w. Confirm tags are fetched", cfg.TagPrefix, err)
	}
	_, ver := highestFinalTag(tags, cfg.TagPrefix)
	var current intent.SemVer
	if ver != "" {
		if sv, serr := intent.ParseSemVer(ver); serr == nil {
			current = sv
		}
	}
	if _, perr := intent.ParseTitle(cfg, current, title); perr != nil {
		return intentErrorToReject(perr)
	}
	if !adopting {
		paths, err := runner.ChangedPaths(ctx, base, head)
		if err != nil {
			if _, ok := err.(*BaseNotAncestorError); ok {
				return fmt.Errorf("check-title: %v", err)
			}
			return fmt.Errorf("check-title: cannot read the PR change between %q and %q: %w", base, head, err)
		}
		changes, cerr := buildProtectedChanges(ctx, runner, cfg, base, head, paths)
		if cerr != nil {
			return cerr
		}
		isReleasePR := intent.ClassifyPR(cfg, author, headBranch) == intent.PRKindRelease
		if cerr := intent.CheckProtectedPaths(cfg, changes, isReleasePR); cerr != nil {
			return intentErrorToReject(cerr)
		}
	}
	return emitLines([]string{"error="})
}

// buildProtectedChanges reads the protected version field at both ends for
// every touched version file. Unreadable or unparsable JSON fails closed by
// marking the field changed.
func buildProtectedChanges(ctx context.Context, runner GitRunner, cfg intent.Config, base, head string, paths []string) ([]intent.ProtectedChange, error) {
	changed := map[string]bool{}
	for _, p := range paths {
		changed[p] = true
	}
	var out []intent.ProtectedChange
	for p := range changed {
		out = append(out, intent.ProtectedChange{Path: p})
	}
	for _, vf := range cfg.VersionFiles {
		if !changed[vf.Path] {
			continue
		}
		baseField, berr := readVersionField(ctx, runner, base, vf.Path, vf.Field)
		headField, herr := readVersionField(ctx, runner, head, vf.Path, vf.Field)
		fieldChanged := baseField != headField
		if berr != nil || herr != nil {
			if isJSONError(berr) || isJSONError(herr) {
				fieldChanged = true
			} else if berr != nil && herr != nil {
				fieldChanged = baseField != headField
			} else {
				fieldChanged = true
			}
		}
		for i := range out {
			if out[i].Path == vf.Path {
				out[i].ProtectedFieldChanged = fieldChanged
			}
		}
	}
	return out, nil
}

func readVersionField(ctx context.Context, runner GitRunner, rev, path, field string) (string, error) {
	data, err := runner.ReadFile(ctx, rev, path)
	if err != nil {
		return "", err
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", &jsonParseError{Path: path, Rev: rev, Err: err}
	}
	v, _ := obj[field].(string)
	return v, nil
}

type jsonParseError struct {
	Path string
	Rev  string
	Err  error
}

func (e *jsonParseError) Error() string {
	return fmt.Sprintf("cannot parse %q at %q as JSON: %v", e.Path, e.Rev, e.Err)
}

func isJSONError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*jsonParseError)
	return ok
}
