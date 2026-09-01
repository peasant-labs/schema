// Command release-guard is the thin CLI the release workflows shell out to. It
// wraps the pure logic in internal/release so the workflows never re-encode the
// title/tag grammar or the final-release guard inline.
//
// Subcommands:
//
//	release-guard parse-title "<pr title>"
//	    Validate a release-PR title and emit version + kind. Exits non-zero with
//	    an actionable message on a malformed title.
//
//	release-guard parse-tag "<git tag>"
//	    Validate a git tag as a schema release reference (rejects the legacy
//	    pkg/schema/v* and any non-release tag) and emit version + kind.
//
//	release-guard check-maintainer --user <login>
//	    Assert a GitHub user is a release maintainer (admin/maintain collaborator
//	    on $GITHUB_REPOSITORY). Reads the collaborator permission via the
//	    GitHubClient and applies release.IsMaintainer — the single source of the
//	    maintainer predicate consumed by both release-pr.yml gates.
//
//	release-guard check-approval --pr <number>
//	    Assert a release PR has at least one standing APPROVED review from a
//	    maintainer. Fetches all review pages via the GitHubClient, reduces each
//	    reviewer's latest non-COMMENTED review, then applies release.IsMaintainer
//	    per approver.
//
//	release-guard check-workflow [--release .github/workflows/release.yml] [--policy .github/release-guard.policy.yml]
//	    Validate that the release workflow satisfies the per-repo policy
//	    (.github/release-guard.policy.yml): every declared job exists with the
//	    required needs-edges (and, for reusable gates, the required uses /
//	    secrets:inherit / no-if shape; for OIDC-authenticated gates, the required
//	    job permissions and GitHub environment binding). Repo-agnostic - the job graph lives in the
//	    policy file, not in this tool.
//
// Every subcommand that derives a (version, kind) writes "version=<v>" and
// "kind=<k>" to the file named by $GITHUB_OUTPUT when set (so workflow steps can
// consume them via steps.<id>.outputs.*), and always echoes them to stdout.
//
// main() is the SOLE composition root: it reads GH_TOKEN + GITHUB_REPOSITORY
// from the environment ONCE, builds the GitHubClient (go-github), and injects it
// DOWN into pure handler funcs that take their deps
// as parameters and never touch os.Getenv. Handlers return an error; main maps a
// non-nil error to a non-zero exit via fatalf, preserving exit-code parity.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/peasant-labs/schema/internal/release"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: release-guard <parse-title|parse-tag|check-workflow|check-maintainer|check-approval|bubble> ...")
	}
	sub := os.Args[1]
	args := os.Args[2:]
	ctx := context.Background()

	var err error
	switch sub {
	case "parse-title":
		err = runParseTitle(args)
	case "parse-tag":
		err = runParseTag(args)
	case "check-workflow":
		err = runCheckWorkflow(args)
	case "check-maintainer":
		gh := mustGitHubClient(sub)
		err = runCheckMaintainer(ctx, gh, mustRepo(sub), args)
	case "check-approval":
		gh := mustGitHubClient(sub)
		err = runCheckApproval(ctx, gh, mustRepo(sub), args)
	case "bubble":
		gh := mustGitHubClient(sub)
		err = runBubble(ctx, gh, mustRepo(sub), args)
	default:
		fatalf("unknown subcommand %q: expected parse-title, parse-tag, check-workflow, check-maintainer, check-approval, or bubble", sub)
	}
	if err != nil {
		fatalf("%v", err)
	}
}

// --- composition root: env read ONCE, deps built here (never in handlers) ---

// gitHubTokenEnv is the SINGLE environment variable release-guard reads for the
// GitHub API token. It is GH_TOKEN — the variable every release workflow exports
// (`env: { GH_TOKEN: ${{ github.token }} }`), which is also what the previous
// `gh` shell-outs consumed — so the go-github seam authenticates with the exact
// credential the workflows already provide (no separate GITHUB_-prefixed name).
const gitHubTokenEnv = "GH_TOKEN"

// readGitHubToken reads the GitHub API token from $GH_TOKEN. It returns an
// actionable error (what/why/where/how) when the token is empty or unset, and is
// the TESTABLE core of the composition root's fail-fast: mustGitHubClient calls
// it and maps a non-nil error to a fatal exit, so both the env-var name read and
// its diagnostic are covered without a test having to trigger os.Exit.
func readGitHubToken(sub string) (string, error) {
	token := os.Getenv(gitHubTokenEnv)
	if token == "" {
		return "", fmt.Errorf("%s: $%s is empty or unset, but this command calls the GitHub API. Export a token with repo read access before running release-guard %s (in GitHub Actions: `env: { %s: ${{ github.token }} }`)", sub, gitHubTokenEnv, sub, gitHubTokenEnv)
	}
	return token, nil
}

// mustGitHubClient reads $GH_TOKEN once (via readGitHubToken) and builds the
// production GitHubClient, FAILING FAST (actionable) on an empty/unset token
// BEFORE any API call. newGitHubClient also rejects an empty token, but checking
// here first yields the what/why/where/how message.
func mustGitHubClient(sub string) GitHubClient {
	token, err := readGitHubToken(sub)
	if err != nil {
		fatalf("%v", err)
	}
	gh, err := newGitHubClient(token)
	if err != nil {
		fatalf("%s: cannot construct the GitHub API client: %v", sub, err)
	}
	return gh
}

// mustRepo reads GITHUB_REPOSITORY once (owner/repo); the wrapper splits it.
func mustRepo(sub string) string {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		fatalf("%s: $GITHUB_REPOSITORY is not set; cannot address the repository for the GitHub API call. Set it to owner/repo (GitHub Actions sets it automatically)", sub)
	}
	return repo
}

// --- handlers (pure: deps injected, never os.Getenv; return error) ----------

func runParseTitle(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: release-guard parse-title \"<pr title>\"")
	}
	v, kind, err := release.ParseReleaseTitle(args[0])
	if err != nil {
		return err
	}
	return emit(v, kind)
}

func runParseTag(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: release-guard parse-tag \"<git tag>\"")
	}
	v, kind, err := release.ParseTag(args[0])
	if err != nil {
		return err
	}
	return emit(v, kind)
}

func runCheckWorkflow(args []string) error {
	releaseWorkflow := ".github/workflows/release.yml"
	policyPath := ".github/release-guard.policy.yml"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--release":
			i++
			if i >= len(args) {
				return fmt.Errorf("--release requires a value")
			}
			releaseWorkflow = args[i]
		case "--policy":
			i++
			if i >= len(args) {
				return fmt.Errorf("--policy requires a value")
			}
			policyPath = args[i]
		default:
			return fmt.Errorf("unknown flag %q for check-workflow", args[i])
		}
	}
	policy, err := release.LoadWorkflowPolicy(policyPath)
	if err != nil {
		return err
	}
	if err := release.CheckReleaseWorkflowFile(releaseWorkflow, policy); err != nil {
		return err
	}
	fmt.Printf("release workflow gates are valid: %s satisfies the release-guard policy in %s\n", releaseWorkflow, policyPath)
	return nil
}

func runCheckMaintainer(ctx context.Context, gh GitHubClient, repo string, args []string) error {
	user := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user":
			i++
			if i >= len(args) {
				return fmt.Errorf("--user requires a value")
			}
			user = args[i]
		default:
			return fmt.Errorf("unknown flag %q for check-maintainer", args[i])
		}
	}
	if user == "" {
		return fmt.Errorf("usage: release-guard check-maintainer --user <login>")
	}

	perm, err := gh.CollaboratorPermission(ctx, repo, user)
	if err != nil {
		return fmt.Errorf("check-maintainer: cannot read the collaborator permission of %q on %s: %v", user, repo, err)
	}
	if !release.IsMaintainer(perm) {
		return fmt.Errorf("user %q has permission %q on %s, but a release action (open or approve a release PR) requires a maintainer (%s or %s); ask a maintainer to perform it",
			user, perm, repo, release.PermAdmin, release.PermMaintain)
	}
	fmt.Printf("user %s is a maintainer (%s)\n", user, perm)
	return nil
}

func runCheckApproval(ctx context.Context, gh GitHubClient, repo string, args []string) error {
	pr := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--pr":
			i++
			if i >= len(args) {
				return fmt.Errorf("--pr requires a value")
			}
			pr = args[i]
		default:
			return fmt.Errorf("unknown flag %q for check-approval", args[i])
		}
	}
	if pr == "" {
		return fmt.Errorf("usage: release-guard check-approval --pr <number>")
	}
	prNumber, err := strconv.Atoi(pr)
	if err != nil {
		return fmt.Errorf("check-approval: --pr value %q is not a number: %v", pr, err)
	}

	reviews, err := gh.PullReviews(ctx, repo, prNumber)
	if err != nil {
		return fmt.Errorf("check-approval: cannot read reviews for release PR #%d on %s: %v", prNumber, repo, err)
	}
	approvers := release.LatestApprovers(reviews)
	if len(approvers) == 0 {
		return fmt.Errorf("release PR #%d has no standing APPROVED reviews; a maintainer (%s or %s) must approve before merge",
			prNumber, release.PermAdmin, release.PermMaintain)
	}

	approver, perm, err := findMaintainerApprover(ctx, gh, repo, approvers)
	if err != nil {
		return fmt.Errorf("check-approval: %v", err)
	}
	if approver != "" {
		fmt.Printf("release PR #%d has a standing maintainer approval from %s (%s)\n", prNumber, approver, perm)
		return nil
	}
	return fmt.Errorf("release PR #%d has standing approvals, but none from an %s/%s maintainer; a maintainer must approve the release",
		prNumber, release.PermAdmin, release.PermMaintain)
}

// findMaintainerApprover returns the first approver who is a maintainer, in the
// given order. ("", "", nil) means none of the approvers is a maintainer.
func findMaintainerApprover(ctx context.Context, gh GitHubClient, repo string, approvers []string) (string, release.CollaboratorPermission, error) {
	for _, approver := range approvers {
		perm, err := gh.CollaboratorPermission(ctx, repo, approver)
		if err != nil {
			return "", "", fmt.Errorf("cannot read the collaborator permission of approving reviewer %q on %s: %w", approver, repo, err)
		}
		if release.IsMaintainer(perm) {
			return approver, perm, nil
		}
	}
	return "", "", nil
}

// --- output helpers ---

// formatOutput renders the version/kind lines exactly as written to
// $GITHUB_OUTPUT and stdout. It is the byte-parity surface: the format is
// "version=<v>\nkind=<k>\n", unchanged across the go-github swap.
func formatOutput(v release.Version, kind release.ReleaseKind) string {
	return fmt.Sprintf("version=%s\nkind=%s\n", v, kind)
}

// emit writes version/kind to $GITHUB_OUTPUT (when set) and stdout.
func emit(v release.Version, kind release.ReleaseKind) error {
	line := formatOutput(v, kind)
	if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("cannot open $GITHUB_OUTPUT (%s): %v", path, err)
		}
		defer f.Close()
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("cannot write to $GITHUB_OUTPUT (%s): %v", path, err)
		}
	}
	fmt.Print(line)
	return nil
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "release-guard: "+format+"\n", a...)
	os.Exit(1)
}
