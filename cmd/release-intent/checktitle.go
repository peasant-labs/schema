package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
			// Landed subjects carry " (#N)"; strip it so a title edited after
			// merge gives no false hit.
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
	// Config always comes from the base branch. A PR that adds the config has
	// no base file: it is checked for grammar only, with its own file
	// supplying the tag prefix, and no path is protected yet.
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
	cfg, err := ParseConfig(cfgData, DefaultConfigPath)
	if err != nil {
		return fmt.Errorf("check-title: the base release config %w", err)
	}
	// Highest tag for the breaking rule, read at run time.
	tags, err := runner.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return fmt.Errorf("check-title: cannot list tags with prefix %q: %w. Confirm tags are fetched", cfg.TagPrefix, err)
	}
	_, ver := highestFinalTag(tags, cfg.TagPrefix)
	below := true
	if ver != "" {
		below = belowOne(ver)
	}
	pt, rej := parseFullTitle(title, cfg.Releases, below)
	if rej != nil {
		return rej
	}
	_ = pt
	if !adopting {
		paths, err := runner.ChangedPaths(ctx, base, head)
		if err != nil {
			if _, ok := err.(*BaseNotAncestorError); ok {
				return fmt.Errorf("check-title: %v", err)
			}
			return fmt.Errorf("check-title: cannot read the PR change between %q and %q: %w", base, head, err)
		}
		protected, perr := checkProtected(ctx, runner, cfg, base, head, paths, author, headBranch)
		if perr != nil {
			return perr
		}
		if protected != nil {
			return protected
		}
	}
	return emitLines([]string{"error="})
}

// checkProtected enforces whole-file changelog and protected_paths plus the
// protected field of version files read at both ends.
func checkProtected(ctx context.Context, runner GitRunner, cfg Config, base, head string, paths []string, author, headBranch string) (*RejectError, error) {
	isReleasePR := author == cfg.AutomationAuthor && author != "" && cfg.BotBranches["release"] != "" && headBranch == cfg.BotBranches["release"]
	// When the config carries no bot_branches map, the workflows decide the
	// release PR by branch; the exemption then applies to the automation author
	// on any branch would be wrong, so require the branch match only when the
	// map names one. With no map, only the release workflow's own check excuses
	// the author; check-title treats every author equally except the exact
	// automation-author-plus-release-branch pair.
	if len(cfg.BotBranches) == 0 {
		isReleasePR = false
	}
	if isReleasePR {
		return nil, nil
	}
	changed := map[string]bool{}
	for _, p := range paths {
		changed[p] = true
	}
	if changed[cfg.Changelog] {
		return &RejectError{
			Code: CodeProtectedPath, What: fmt.Sprintf("PR changes protected file %q", cfg.Changelog),
			Why: "only the release PR may write the changelog", Where: "PR change, checked by release-intent check-title",
			When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
			Fix: "drop the CHANGELOG.md edit from this PR; the bot release PR is its only writer",
		}, nil
	}
	for _, pat := range cfg.ProtectedPaths {
		for p := range changed {
			if matchProtected(pat, p) {
				return &RejectError{
					Code: CodeProtectedPath, What: fmt.Sprintf("PR changes protected path %q (matched by %q)", p, pat),
					Why: "only the release PR may change protected paths", Where: "PR change, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "drop the protected-path edit from this PR",
				}, nil
			}
		}
	}
	for _, vf := range cfg.VersionFiles {
		if !changed[vf.Path] {
			continue
		}
		baseField, berr := readVersionField(ctx, runner, base, vf)
		headField, herr := readVersionField(ctx, runner, head, vf)
		if berr != nil || herr != nil {
			// Fail closed: unreadable or unparsable JSON blocks the PR. A
			// missing file at one end counts as a change when the other end
			// exists.
			why := "the version file cannot be read at both ends"
			if berr != nil && herr != nil {
				why = fmt.Sprintf("the version file cannot be read: base: %v; head: %v", berr, herr)
			} else if berr != nil {
				why = fmt.Sprintf("the version file cannot be read at the base: %v", berr)
			} else {
				why = fmt.Sprintf("the version file cannot be read at the head: %v", herr)
			}
			// Invalid JSON is a distinct fail-closed from a missing blob: name it.
			if isJSONError(berr) || isJSONError(herr) {
				return &RejectError{
					Code: CodeProtectedPath, What: fmt.Sprintf("version file %q does not parse as JSON", vf.Path),
					Why: "the protected field cannot be compared when the file does not parse, so the PR fails closed", Where: "PR change, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "fix the JSON in " + vf.Path + " or drop the edit",
				}, nil
			}
			_ = why
			// A deleted or renamed file surfaces here: the field cannot be
			// compared, so any version-file touch by a non-release PR is
			// rejected when either end is unreadable.
			if baseField != headField {
				return &RejectError{
					Code: CodeProtectedPath, What: fmt.Sprintf("PR changes version file %q (unreadable at one end)", vf.Path),
					Why: "only the release PR may write the protected version field", Where: "PR change, checked by release-intent check-title",
					When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
					Fix: "drop the " + vf.Path + " edit, or keep every field but " + vf.Field + " unchanged",
				}, nil
			}
			continue
		}
		if baseField != headField {
			return &RejectError{
				Code: CodeProtectedPath, What: fmt.Sprintf("PR changes protected field %q of %q", vf.Field, vf.Path),
				Why: "only the release PR may write the version field; other fields stay editable", Where: "PR change, checked by release-intent check-title",
				When: "pull_request opened, edited, synchronize, reopened, or the merge-group build", Means: "the title-guard check is red",
				Fix: "revert the " + vf.Field + " field in " + vf.Path + "; scripts, dependencies and the schema pin stay editable",
			}, nil
		}
	}
	return nil, nil
}

func readVersionField(ctx context.Context, runner GitRunner, rev string, vf VersionFile) (string, error) {
	data, err := runner.ReadFile(ctx, rev, vf.Path)
	if err != nil {
		return "", err
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", &jsonParseError{Path: vf.Path, Rev: rev, Err: err}
	}
	v, _ := obj[vf.Field].(string)
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

func matchProtected(pattern, path string) bool {
	if strings.HasSuffix(pattern, "/*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	if strings.Contains(pattern, "*") {
		// Minimal glob: * matches any non-separator run.
		parts := strings.Split(pattern, "*")
		idx := 0
		for i, part := range parts {
			if part == "" {
				continue
			}
			found := strings.Index(path[idx:], part)
			if found < 0 {
				return false
			}
			idx += found + len(part)
			if i == 0 && !strings.HasPrefix(path, part) && !strings.HasPrefix(pattern, "*") {
				return false
			}
		}
		if !strings.HasSuffix(pattern, "*") && !strings.HasSuffix(path, parts[len(parts)-1]) {
			return false
		}
		return true
	}
	return pattern == path
}
