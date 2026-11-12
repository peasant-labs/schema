package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GitRunner is the release-intent git seam: read-only history and file queries
// for release tooling. It shells out to `git` with a structurally-built argv,
// never a shell.
type GitRunner interface {
	// FirstParent returns first-parent commits from ref, newest first.
	FirstParent(ctx context.Context, ref string, changelog string) ([]Commit, error)
	// Tags returns tag name to peeled commit for tags with the given prefix.
	Tags(ctx context.Context, prefix string) (map[string]string, error)
	// FirstCommitAdding returns the oldest commit that adds path, or "".
	FirstCommitAdding(ctx context.Context, path string) (string, error)
	// ChangedPaths returns paths differing between the merge base of base and
	// head against head, with no rename detection so a renamed protected file
	// surfaces as both names.
	ChangedPaths(ctx context.Context, base, head string) ([]string, error)
	// ReadFile returns the bytes of path at rev.
	ReadFile(ctx context.Context, rev, path string) ([]byte, error)
	// IsAncestor reports whether ancestor is an ancestor of descendant.
	IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error)
}

// Commit is one first-parent commit: its SHA, its subject line, whether it
// changes the changelog file, and its commit date as YYYY-MM-DD for dated
// changelog headings.
type Commit struct {
	SHA              string
	Subject          string
	ChangesChangelog bool
	Date             string
}

// gitSubcommand is the typed closed allowlist of git subcommands the runner
// may invoke.
type gitSubcommand string

const (
	gitRevList   gitSubcommand = "rev-list"
	gitTag       gitSubcommand = "tag"
	gitMergeBase gitSubcommand = "merge-base"
	gitLog       gitSubcommand = "log"
	gitDiff      gitSubcommand = "diff"
	gitShow      gitSubcommand = "show"
	gitCatFile   gitSubcommand = "cat-file"
)

// allowedGitSubcommands is the allowlist enforced by runGit.
var allowedGitSubcommands = map[gitSubcommand]bool{
	gitRevList:   true,
	gitTag:       true,
	gitMergeBase: true,
	gitLog:       true,
	gitDiff:      true,
	gitShow:      true,
	gitCatFile:   true,
}

// execGit is the production GitRunner.
type execGit struct {
	gitPath string
	dir     string
}

// newExecGit resolves `git` on PATH once and returns the hardened runner.
func newExecGit() (GitRunner, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot find a `git` executable on PATH for release-intent history checks: %w. Install git or enter the nix dev shell (which provides it) before running release-intent git commands", err)
	}
	return &execGit{gitPath: gitPath}, nil
}

// newExecGitIn returns the runner rooted at dir so tests can point it at a
// throwaway repo without mutating the process-global cwd.
func newExecGitIn(dir string) (GitRunner, error) {
	r, err := newExecGit()
	if err != nil {
		return nil, err
	}
	g := r.(*execGit)
	g.dir = dir
	return g, nil
}

// runGit builds and runs `git <sub> <fixedFlags...> [--end-of-options|-- <dynArgs...>]`.
func (g *execGit) runGit(ctx context.Context, sub gitSubcommand, fixedFlags []string, dynArgs ...string) ([]byte, error) {
	if !allowedGitSubcommands[sub] {
		return nil, fmt.Errorf("git runner: subcommand %q is not in the release-intent allowlist; refusing to run it. This is a programming error in release-intent, not a user input issue", sub)
	}
	for _, a := range dynArgs {
		if strings.HasPrefix(a, "-") {
			return nil, fmt.Errorf("git runner: refusing to pass %q to `git %s`: a ref/tag/SHA/pattern/path that begins with '-' could be interpreted as a git option instead of a revision. Pass a value that does not start with '-'", a, sub)
		}
	}
	argv := make([]string, 0, len(fixedFlags)+len(dynArgs)+2)
	argv = append(argv, string(sub))
	argv = append(argv, fixedFlags...)
	// Revision-safe sentinel for rev-list/merge-base/log; path separator for
	// diff/show/cat-file paths is added by the caller via "--".
	if sub == gitDiff || sub == gitShow || sub == gitCatFile || sub == gitLog {
		argv = append(argv, dynArgs...)
	} else {
		argv = append(argv, "--end-of-options")
		argv = append(argv, dynArgs...)
	}
	cmd := exec.CommandContext(ctx, g.gitPath, argv...)
	cmd.Dir = g.dir
	return cmd.Output()
}

func (g *execGit) RevParse(ctx context.Context, ref string) (string, error) {
	out, err := g.runGit(ctx, gitRevList, []string{"-n", "1"}, ref)
	if err != nil {
		return "", fmt.Errorf("git runner: cannot resolve %q via `git rev-list -n 1`: %w. Confirm the ref exists in this checkout", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *execGit) ListTags(ctx context.Context, pattern string) ([]string, error) {
	out, err := g.runGit(ctx, gitTag, []string{"--list"}, pattern)
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot list tags matching %q via `git tag --list`: %w. Confirm the checkout has tags fetched", pattern, err)
	}
	var tags []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// IsAncestor distinguishes a definitive negative (exit 1) from a real failure.
func (g *execGit) IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error) {
	_, err := g.runGit(ctx, gitMergeBase, []string{"--is-ancestor"}, ancestor, descendant)
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("git runner: cannot determine whether %q is an ancestor of %q via `git merge-base --is-ancestor`: %w. Treat this as not-an-ancestor and investigate the git error", ancestor, descendant, err)
}

// FirstParent lists first-parent commits newest first with changelog touch info
// and commit dates.
func (g *execGit) FirstParent(ctx context.Context, ref string, changelog string) ([]Commit, error) {
	// Format: SHA, subject, date, then names touched. Use NUL separators that
	// subjects cannot contain unescaped; the date is %cs (committer date
	// short YYYY-MM-DD) for dated changelog headings.
	out, err := g.runGit(ctx, gitLog, []string{"--first-parent", "--format=%H%x00%s%x00%cs", "--name-only", ref, "--", changelog, "."})
	if err != nil {
		out2, err2 := g.runGit(ctx, gitLog, []string{"--first-parent", "--format=%H%x00%s%x00%cs", ref, "--"})
		if err2 != nil {
			return nil, fmt.Errorf("git runner: cannot list first-parent commits of %q via `git log --first-parent`: %w. Confirm the ref exists", ref, err)
		}
		return parseLogCommits(string(out2)), nil
	}
	return parseLogCommitsWithNames(string(out), changelog), nil
}

func parseLogCommits(out string) []Commit {
	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) < 2 {
			continue
		}
		c := Commit{SHA: strings.TrimSpace(parts[0]), Subject: parts[1]}
		if len(parts) == 3 {
			c.Date = strings.TrimSpace(parts[2])
		}
		commits = append(commits, c)
	}
	return commits
}

func parseLogCommitsWithNames(out, changelog string) []Commit {
	var commits []Commit
	var cur *Commit
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "\x00") {
			parts := strings.SplitN(line, "\x00", 3)
			sha := ""
			subj := ""
			date := ""
			if len(parts) > 0 {
				sha = strings.TrimSpace(parts[0])
			}
			if len(parts) > 1 {
				subj = parts[1]
			}
			if len(parts) > 2 {
				date = strings.TrimSpace(parts[2])
			}
			commits = append(commits, Commit{SHA: sha, Subject: subj, Date: date})
			cur = &commits[len(commits)-1]
			continue
		}
		if cur != nil && strings.TrimSpace(line) == changelog {
			cur.ChangesChangelog = true
		}
	}
	return commits
}

// Tags resolves every tag matching prefix to its peeled commit.
func (g *execGit) Tags(ctx context.Context, prefix string) (map[string]string, error) {
	tags, err := g.ListTags(ctx, prefix+"*")
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(tags))
	for _, t := range tags {
		sha, err := g.RevParse(ctx, t+"^{commit}")
		if err != nil {
			// Annotated vs lightweight: fall back to the ref itself.
			sha, err = g.RevParse(ctx, t)
			if err != nil {
				return nil, fmt.Errorf("git runner: cannot peel tag %q to a commit: %w. Confirm the tag exists", t, err)
			}
		}
		out[t] = sha
	}
	return out, nil
}

// FirstCommitAdding returns the oldest commit adding path, or "" when none.
func (g *execGit) FirstCommitAdding(ctx context.Context, path string) (string, error) {
	out, err := g.runGit(ctx, gitLog, []string{"--diff-filter=A", "--format=%H", "--"}, path)
	if err != nil {
		return "", fmt.Errorf("git runner: cannot find the commit adding %q via `git log --diff-filter=A`: %w", path, err)
	}
	var last string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			last = t
		}
	}
	return last, nil
}

// ChangedPaths diffs merge-base(base,head) against head with no rename
// detection. A renamed protected file appears as both names so either side is
// caught. A base that is not an ancestor of head is a distinct typed failure.
func (g *execGit) ChangedPaths(ctx context.Context, base, head string) ([]string, error) {
	isAnc, err := g.IsAncestor(ctx, base, head)
	if err != nil {
		return nil, err
	}
	if !isAnc {
		// Fall back to a direct merge-base probe: when base and head share no
		// history, report base-not-on-chain distinctly from a git failure.
		return nil, &BaseNotAncestorError{Base: base, Head: head}
	}
	mb, err := g.mergeBase(ctx, base, head)
	if err != nil {
		return nil, err
	}
	out, err := g.runGit(ctx, gitDiff, []string{"--name-only", "--no-renames", mb, head, "--"})
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot diff %q against %q via `git diff --name-only --no-renames`: %w. Confirm both revisions exist", mb, head, err)
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			paths = append(paths, t)
		}
	}
	return paths, nil
}

func (g *execGit) mergeBase(ctx context.Context, a, b string) (string, error) {
	out, err := g.runGit(ctx, gitMergeBase, []string{}, a, b)
	if err != nil {
		return "", fmt.Errorf("git runner: cannot find merge base of %q and %q via `git merge-base`: %w", a, b, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ReadFile returns path at rev via `git show rev:path`.
func (g *execGit) ReadFile(ctx context.Context, rev, path string) ([]byte, error) {
	if strings.Contains(path, ":") || strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return nil, fmt.Errorf("git runner: refusing to read path %q at %q: paths must be repo-relative and must not escape via .. or absolute form", path, rev)
	}
	out, err := g.runGit(ctx, gitShow, []string{}, rev+":"+path)
	if err != nil {
		return nil, &BlobReadError{Rev: rev, Path: path, Err: err}
	}
	return out, nil
}

// BaseNotAncestorError means base is not on head's history.
type BaseNotAncestorError struct {
	Base string
	Head string
}

func (e *BaseNotAncestorError) Error() string {
	return fmt.Sprintf("git runner: base %q is not an ancestor of head %q (checked via `git merge-base --is-ancestor`): the PR range has no merge base. Rebase the PR onto the base branch or pass a base on its history", e.Base, e.Head)
}

// BlobReadError means the blob could not be read (missing file, failed fetch).
type BlobReadError struct {
	Rev  string
	Path string
	Err  error
}

func (e *BlobReadError) Error() string {
	return fmt.Sprintf("git runner: cannot read %q at %q via `git show`: %v. In a blobless clone this can be a failed on-demand fetch; confirm the blob exists and the network allows the fetch", e.Path, e.Rev, e.Err)
}
