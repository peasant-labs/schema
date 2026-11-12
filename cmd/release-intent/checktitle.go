package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
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
	for sc.Scan() {
		if t := strings.TrimSpace(sc.Text()); t != "" {
			return t, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("check-title: cannot read the PR title on stdin: %v", err)
	}
	return "", usageErr("usage: release-intent check-title reads the PR title on stdin (got empty input): pipe the title, e.g. echo 'feat(x)[minor]: ...' | release-intent check-title --base REV --head REV --author LOGIN --head-branch BRANCH")
}

func checkTitle(ctx context.Context, runner GitRunner, base, head, author, headBranch, title string) error {
	cfgData, err := runner.ReadFile(ctx, base, DefaultConfigPath)
	adopting := false
	if err != nil {
		var nf *NotFoundError
		if !errors.As(err, &nf) {
			return fmt.Errorf("check-title: cannot read release config %q at base %q: %v. Confirm the base revision is readable; only a missing file means the adopting PR", DefaultConfigPath, base, err)
		}
		headData, herr := runner.ReadFile(ctx, head, DefaultConfigPath)
		if herr != nil {
			return fmt.Errorf("check-title: cannot read release config %q at base %q or head %q: base has no file; head read: %v. Confirm the head revision carries the config the adopting PR adds", DefaultConfigPath, base, head, herr)
		}
		cfgData = headData
		adopting = true
	}
	cfg, err := intent.ParseConfig(cfgData)
	if err != nil {
		return rejectAs("check-title", err)
	}
	tags, err := runner.Tags(ctx, cfg.TagPrefix)
	if err != nil {
		return fmt.Errorf("check-title: cannot list tags with prefix %q: %w. Confirm tags are fetched", cfg.TagPrefix, err)
	}
	_, version, ok := intent.HighestFinalTag(tags, cfg.TagPrefix)
	var current intent.SemVer
	if ok {
		current = version
	}
	if _, perr := intent.ParseTitle(cfg, current, title); perr != nil {
		return rejectAs("check-title", perr)
	}
	if !adopting {
		mergeBase, paths, err := runner.ChangedPaths(ctx, base, head)
		if err != nil {
			return fmt.Errorf("check-title: cannot read the PR change between %q and %q: %w", base, head, err)
		}
		changes, cerr := buildProtectedChanges(ctx, runner, cfg, mergeBase, head, paths)
		if cerr != nil {
			return cerr
		}
		isReleasePR := intent.ClassifyPR(cfg, author, headBranch) == intent.PRKindRelease
		if cerr := intent.CheckProtectedPaths(cfg, changes, isReleasePR); cerr != nil {
			return rejectAs("check-title", cerr)
		}
	}
	return emitLines([]string{"error="})
}

// buildProtectedChanges reads the protected version field at the merge base
// and at head for every touched version file. A corrupt file fails closed;
// any other blob read error exits 1 and never becomes an accept or a stale
// verdict.
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
		baseVal, baseCorrupt, berr := readProtectedField(ctx, runner, base, vf.Path, vf.Field)
		if berr != nil {
			return nil, berr
		}
		headVal, headCorrupt, herr := readProtectedField(ctx, runner, head, vf.Path, vf.Field)
		if herr != nil {
			return nil, herr
		}
		if baseCorrupt || headCorrupt {
			return nil, rejectAs("check-title", protectedFieldError(vf.Path))
		}
		for i := range out {
			if out[i].Path == vf.Path {
				out[i].ProtectedFieldChanged = baseVal != headVal
			}
		}
	}
	return out, nil
}

// readProtectedField returns the protected field value and whether the file
// is corrupt. Blob read failures (other than absent, which reads as empty)
// are errors for exit 1.
func readProtectedField(ctx context.Context, runner GitRunner, rev, path, field string) (string, bool, error) {
	data, err := runner.ReadFile(ctx, rev, path)
	if err != nil {
		var nf *NotFoundError
		if errors.As(err, &nf) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("check-title: cannot read version file %q at %q: %v. Confirm the revision is readable; a failed fetch is never an accept", path, rev, err)
	}
	val, corrupt := versionFieldValue(data, field)
	return val, corrupt, nil
}

func protectedFieldError(path string) error {
	return &intent.Error{
		Code:  intent.ErrProtectedPath,
		What:  "version file " + strconv.Quote(path) + " does not parse as JSON",
		Why:   "the protected field cannot be compared when the file does not parse, so the PR fails closed",
		Where: "PR change, checked by release-intent check-title",
		When:  "pull_request opened, edited, synchronize, reopened, or the merge-group build",
		Means: "the title-guard check is red, so the merge queue will not accept this PR",
		Fix:   "fix the JSON in " + path + " or drop the edit",
	}
}
