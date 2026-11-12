package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// GitRunner is the release-intent git seam: read-only history and file queries
// for release tooling. It shells out to `git` with a structurally-built argv,
// never a shell.
type GitRunner interface {
	// FirstParent returns first-parent commits from ref, newest first.
	FirstParent(ctx context.Context, ref string, changelog string) ([]intent.Commit, error)
	// Tags returns tag name to peeled commit for tags with the given prefix in
	// one git call.
	Tags(ctx context.Context, prefix string) (map[string]string, error)
	// FirstCommitAdding returns the oldest commit on branch's first-parent
	// history that adds path, or "" when no commit on that chain adds it.
	FirstCommitAdding(ctx context.Context, branch, path string) (string, error)
	// ChangedPaths diffs the merge base of base and head against head, with no
	// rename detection so a renamed protected file surfaces as both names. It
	// fails only when the two revisions share no merge base.
	ChangedPaths(ctx context.Context, base, head string) (mergeBase string, paths []string, err error)
	// ReadFile returns the bytes of path at rev, or a NotFoundError when the
	// path is absent at that revision.
	ReadFile(ctx context.Context, rev, path string) ([]byte, error)
	// MergeTree reports whether branch tip merges into current cleanly. A nil
	// error means the merge tree was written; a conflict or a read failure
	// returns an error.
	MergeTree(ctx context.Context, tip, current string) error
}

// gitSubcommand is the typed closed allowlist of git subcommands the runner
// may invoke.
type gitSubcommand string

const (
	gitRevList    gitSubcommand = "rev-list"
	gitMergeBase  gitSubcommand = "merge-base"
	gitLog        gitSubcommand = "log"
	gitDiff       gitSubcommand = "diff"
	gitShow       gitSubcommand = "show"
	gitLsTree     gitSubcommand = "ls-tree"
	gitForEachRef gitSubcommand = "for-each-ref"
	gitMergeTree  gitSubcommand = "merge-tree"
)

// allowedGitSubcommands is the allowlist enforced by runGit.
var allowedGitSubcommands = map[gitSubcommand]bool{
	gitRevList:    true,
	gitMergeBase:  true,
	gitLog:        true,
	gitDiff:       true,
	gitShow:       true,
	gitLsTree:     true,
	gitForEachRef: true,
	gitMergeTree:  true,
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

// dashRefusal is the single refusal for a dynamic git argument that begins
// with '-'. Both runGit and the read path that probes before reading use it,
// so a hostile value never executes and never surfaces as a blob error.
func dashRefusal(arg string, sub gitSubcommand) error {
	return fmt.Errorf("git runner: refusing to pass %q to `git %s`: a ref/tag/SHA/pattern/path that begins with '-' could be interpreted as a git option instead of a revision. Pass a value that does not start with '-'", arg, sub)
}

// runGit builds and runs `git <sub> <fixedFlags...> <dynArgs...>` with no
// shell. Callers place `--end-of-options` in fixedFlags where git supports it
// (before revisions) and `--` before paths, so a leading-dash dynamic value
// can never become an option: the dash check below refuses it first anyway.
func (g *execGit) runGit(ctx context.Context, sub gitSubcommand, fixedFlags []string, dynArgs ...string) ([]byte, error) {
	if !allowedGitSubcommands[sub] {
		return nil, fmt.Errorf("git runner: subcommand %q is not in the release-intent allowlist; refusing to run it. This is a programming error in release-intent, not a user input issue", sub)
	}
	for _, a := range dynArgs {
		if a != "--" && strings.HasPrefix(a, "-") {
			return nil, dashRefusal(a, sub)
		}
	}
	argv := make([]string, 0, len(fixedFlags)+len(dynArgs)+1)
	argv = append(argv, string(sub))
	argv = append(argv, fixedFlags...)
	argv = append(argv, dynArgs...)
	cmd := exec.CommandContext(ctx, g.gitPath, argv...)
	cmd.Dir = g.dir
	return cmd.Output()
}

// FirstParent lists every first-parent commit newest first. The changelog set
// comes from a second call so an empty commit is still listed: history
// simplification on the first call would drop it.
func (g *execGit) FirstParent(ctx context.Context, ref string, changelog string) ([]intent.Commit, error) {
	// Format: SHA, subject, committer date. Use NUL separators that subjects
	// cannot contain unescaped; the date is %cs (committer date short
	// YYYY-MM-DD) for dated changelog headings.
	out, err := g.runGit(ctx, gitLog, []string{"--first-parent", "--format=%H%x00%s%x00%cs", "--end-of-options"}, ref)
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot list first-parent commits of %q via `git log --first-parent`: %w. Confirm the ref exists in this checkout", ref, err)
	}
	commits := parseLogCommits(string(out))
	names, err := g.runGit(ctx, gitLog, []string{"--first-parent", "--format=%H", "--end-of-options"}, ref, "--", changelog)
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot list the commits touching %q via `git log --first-parent`: %w. Confirm the ref exists in this checkout", changelog, err)
	}
	touched := map[string]bool{}
	for _, line := range strings.Split(string(names), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			touched[t] = true
		}
	}
	for i := range commits {
		commits[i].ChangesChangelog = touched[commits[i].SHA]
	}
	return commits, nil
}

func parseLogCommits(out string) []intent.Commit {
	var commits []intent.Commit
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) < 2 {
			continue
		}
		c := intent.Commit{SHA: strings.TrimSpace(parts[0]), Subject: parts[1]}
		if len(parts) == 3 {
			c.Date = strings.TrimSpace(parts[2])
		}
		commits = append(commits, c)
	}
	return commits
}

// Tags resolves every tag with the given prefix to its peeled commit in one
// git call. Annotated tags peel to their commit; lightweight tags resolve
// directly. A prefix that begins with '-' is refused before exec: it would
// otherwise become a hostile pattern argument.
func (g *execGit) Tags(ctx context.Context, prefix string) (map[string]string, error) {
	if strings.HasPrefix(prefix, "-") {
		return nil, dashRefusal(prefix, gitForEachRef)
	}
	out, err := g.runGit(ctx, gitForEachRef, []string{"--format=%(refname:strip=2)%00%(*objectname)%00%(objectname)", "--end-of-options"}, "refs/tags/"+prefix+"*")
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot list tags with prefix %q via `git for-each-ref`: %w. Confirm the checkout has tags fetched", prefix, err)
	}
	resolved := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\x00")
		if len(parts) != 3 {
			continue
		}
		name, peeled, object := parts[0], strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
		if name == "" || object == "" {
			continue
		}
		if peeled != "" {
			resolved[name] = peeled
		} else {
			resolved[name] = object
		}
	}
	return resolved, nil
}

// FirstCommitAdding returns the oldest commit on branch's first-parent history
// that adds path, or "" when no commit on that chain adds it. With
// --first-parent, a merge commit that brings the file in counts as the adding
// commit.
func (g *execGit) FirstCommitAdding(ctx context.Context, branch, path string) (string, error) {
	out, err := g.runGit(ctx, gitLog, []string{"--first-parent", "--diff-filter=A", "--format=%H", "--end-of-options"}, branch, "--", path)
	if err != nil {
		return "", fmt.Errorf("git runner: cannot find the commit adding %q on %q via `git log --first-parent --diff-filter=A`: %w", path, branch, err)
	}
	var last string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			last = t
		}
	}
	return last, nil
}

// ChangedPaths diffs the merge base of base and head against head with no
// rename detection. A renamed protected file appears as both names so either
// side is caught. A base that shares no history with head is an error: there
// is no merge base to diff.
func (g *execGit) ChangedPaths(ctx context.Context, base, head string) (string, []string, error) {
	out, err := g.runGit(ctx, gitMergeBase, []string{"--end-of-options"}, base, head)
	if err != nil {
		return "", nil, fmt.Errorf("git runner: base %q shares no history with head %q (no merge base via `git merge-base`): the PR range needs a merge base. Rebase the PR onto the base branch or pass revisions with shared history: %v", base, head, err)
	}
	mb := strings.TrimSpace(string(out))
	diff, err := g.runGit(ctx, gitDiff, []string{"--name-only", "--no-renames", "--end-of-options", mb, head, "--"})
	if err != nil {
		return "", nil, fmt.Errorf("git runner: cannot diff %q against %q via `git diff --name-only --no-renames`: %w. Confirm both revisions exist", mb, head, err)
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(diff)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			paths = append(paths, t)
		}
	}
	return mb, paths, nil
}

// MergeTree reports whether tip merges into current without conflict by
// computing the write tree. A nil error means the merge is clean.
func (g *execGit) MergeTree(ctx context.Context, tip, current string) error {
	if _, err := g.runGit(ctx, gitMergeTree, []string{"--write-tree", "--end-of-options"}, tip, current); err != nil {
		return fmt.Errorf("git runner: tip %q does not merge into %q cleanly (via `git merge-tree --write-tree`): %w. Rebuild the release branch from the tip", tip, current, err)
	}
	return nil
}

// ReadFile returns path at rev. A path absent at rev yields NotFoundError;
// every other failure yields BlobReadError, so callers fail closed on fetch
// errors and never treat them as absent. A hostile rev is refused before any
// git call, never wrapped as a blob error.
func (g *execGit) ReadFile(ctx context.Context, rev, path string) ([]byte, error) {
	if strings.Contains(path, ":") || strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return nil, fmt.Errorf("git runner: refusing to read path %q at %q: paths must be repo-relative and must not escape via .. or absolute form", path, rev)
	}
	if strings.HasPrefix(rev+":"+path, "-") {
		return nil, dashRefusal(rev+":"+path, gitShow)
	}
	out, err := g.runGit(ctx, gitShow, []string{"--end-of-options"}, rev+":"+path)
	if err != nil {
		if g.pathAbsent(ctx, rev, path) {
			return nil, &NotFoundError{Rev: rev, Path: path}
		}
		return nil, &BlobReadError{Rev: rev, Path: path, Err: err}
	}
	return out, nil
}

// pathAbsent reports whether path is absent at rev. It uses ls-tree so an
// absent path is structural (empty output, exit 0), never a message match.
func (g *execGit) pathAbsent(ctx context.Context, rev, path string) bool {
	out, err := g.runGit(ctx, gitLsTree, []string{"--end-of-options"}, rev, "--", path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == ""
}

// NotFoundError means the path is absent at the revision. Callers treat this
// as "no file here" (for example the adopting PR has no base config); every
// other read failure is a BlobReadError and must fail closed.
type NotFoundError struct {
	Rev  string
	Path string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("git runner: path %q is absent at %q: the revision has no such blob. A PR that adds the file legitimately reads this way; any other missing-blob read is a checkout problem", e.Path, e.Rev)
}

// BlobReadError means the blob could not be read (failed on-demand fetch in a
// blobless clone, unreadable object, or bad revision). It must fail closed,
// never be treated as absent.
type BlobReadError struct {
	Rev  string
	Path string
	Err  error
}

func (e *BlobReadError) Error() string {
	return fmt.Sprintf("git runner: cannot read %q at %q via `git show`: %v. In a blobless clone this can be a failed on-demand fetch; confirm the blob exists and the network allows the fetch", e.Path, e.Rev, e.Err)
}
