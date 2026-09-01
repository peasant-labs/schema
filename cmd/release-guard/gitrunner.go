package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GitRunner is the release-guard's git-lineage seam: read-only history queries
// for release tooling. It is hardened (typed structured argv, no
// shell, leading-dash rejection, an --end-of-options sentinel, and a subcommand
// allowlist) but still SHELLS `git` — go-git adoption is out of scope.
type GitRunner interface {
	// RevParse resolves ref to a single commit SHA (git rev-list -n 1).
	RevParse(ctx context.Context, ref string) (string, error)
	// ListTags returns the tags matching pattern (git tag --list <pattern>),
	// one per line, blanks dropped.
	ListTags(ctx context.Context, pattern string) ([]string, error)
	// IsAncestor reports whether ancestor is an ancestor of descendant
	// (git merge-base --is-ancestor). A definitive "no" returns (false, nil); a
	// genuine git failure returns (false, err) so the caller can distinguish a
	// real negative from a broken invocation.
	IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error)
}

// gitSubcommand is the typed, closed allowlist of git subcommands the lineage
// seam may invoke (blast-radius cap — a stringly-typed subcommand could be
// any git porcelain; this enum is exactly the three read-only history queries).
type gitSubcommand string

const (
	gitRevList   gitSubcommand = "rev-list"
	gitTag       gitSubcommand = "tag"
	gitMergeBase gitSubcommand = "merge-base"
)

// allowedGitSubcommands is the allowlist enforced by runGit.
var allowedGitSubcommands = map[gitSubcommand]bool{
	gitRevList:   true,
	gitTag:       true,
	gitMergeBase: true,
}

// execGit is the production GitRunner. It runs a `git` binary resolved ONCE on
// PATH with a structurally-built argv — never a shell, never string
// concatenation.
type execGit struct {
	gitPath string
	// dir is the working directory for git. The zero value ("") inherits the
	// process cwd — exactly the behaviour of the old inline helpers, which ran
	// git at the workflow's repository root. It exists so tests can point the
	// runner at a throwaway repo WITHOUT mutating the process-global cwd (which
	// would forbid t.Parallel()).
	dir string
}

// newExecGit resolves `git` on PATH and returns the hardened runner.
//
// PATH safety: we use the stdlib exec.LookPath rather than cli/safeexec.
// Since Go 1.19, os/exec refuses to resolve an executable
// from a cwd-relative PATH entry (it returns exec.ErrDot), which closes the
// PATH/cwd hijack hole cli/safeexec was created to patch. On the Linux GitHub
// Actions runners + nix dev shell where release-guard runs, PATH never contains
// ".", so the stdlib fully covers us and we avoid adding a dependency to this
// contract-only leaf. Resolving once here (not per call) also fail-fasts with an
// actionable error if git is absent.
func newExecGit() (GitRunner, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git runner: cannot find a `git` executable on PATH for release-guard lineage checks: %w. Install git or enter the nix dev shell (which provides it) before running release-guard history checks", err)
	}
	return &execGit{gitPath: gitPath}, nil
}

// runGit builds and runs `git <sub> <fixedFlags...> --end-of-options <dynArgs...>`.
// fixedFlags are trusted, structurally-defined flags (e.g. "-n", "1",
// "--is-ancestor"); dynArgs are UNTRUSTED dynamic refs/tags/SHAs/patterns.
//
// Hardening:
//   - subcommand allowlist: reject anything outside {rev-list, tag, merge-base}.
//   - leading-dash rejection: a dynamic arg that begins with '-' could be
//     misread as a git option (the -e / --upload-pack=… class of injection), so
//     it is refused before exec.
//   - --end-of-options sentinel: everything after it is parsed positionally, so
//     even a value that slipped past the dash check cannot become an option. We
//     use --end-of-options rather than the literal `--` because for rev-list /
//     merge-base, `--` separates revisions from PATHSPECS — it would misparse a
//     ref as a path. --end-of-options is the revision-safe equivalent.
//   - no shell: exec.CommandContext runs git directly with an argv slice.
func (g *execGit) runGit(ctx context.Context, sub gitSubcommand, fixedFlags []string, dynArgs ...string) ([]byte, error) {
	if !allowedGitSubcommands[sub] {
		return nil, fmt.Errorf("git runner: subcommand %q is not in the release-guard allowlist {%s, %s, %s}; refusing to run it. This is a programming error in release-guard, not a user input issue", sub, gitRevList, gitTag, gitMergeBase)
	}
	for _, a := range dynArgs {
		if strings.HasPrefix(a, "-") {
			return nil, fmt.Errorf("git runner: refusing to pass %q to `git %s`: a ref/tag/SHA/pattern that begins with '-' could be interpreted as a git option (e.g. -e, --upload-pack=…) instead of a revision. Pass a value that does not start with '-' (valid git refs never do)", a, sub)
		}
	}

	argv := make([]string, 0, len(fixedFlags)+len(dynArgs)+2)
	argv = append(argv, string(sub))
	argv = append(argv, fixedFlags...)
	argv = append(argv, "--end-of-options")
	argv = append(argv, dynArgs...)

	cmd := exec.CommandContext(ctx, g.gitPath, argv...)
	cmd.Dir = g.dir
	return cmd.Output()
}

func (g *execGit) RevParse(ctx context.Context, ref string) (string, error) {
	out, err := g.runGit(ctx, gitRevList, []string{"-n", "1"}, ref)
	if err != nil {
		return "", fmt.Errorf("git runner: cannot resolve %q to a commit via `git rev-list -n 1`: %w. Confirm the ref/tag exists in this checkout", ref, err)
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

func (g *execGit) IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error) {
	// merge-base --is-ancestor exits 0 (ancestor), 1 (not), or >1 (error).
	_, err := g.runGit(ctx, gitMergeBase, []string{"--is-ancestor"}, ancestor, descendant)
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		// Exit 1 is git's definitive "not an ancestor" — not a failure.
		return false, nil
	}
	// Exit >1 (or a non-exit error, e.g. leading-dash rejection / missing git):
	// a genuine failure. The old inline isAncestor swallowed this as a bare
	// false (blocking the guard); we surface it so the caller can block AND
	// report why.
	return false, fmt.Errorf("git runner: cannot determine whether %q is an ancestor of %q via `git merge-base --is-ancestor`: %w. Treat this as not-an-ancestor (the guard blocks) and investigate the git error", ancestor, descendant, err)
}
