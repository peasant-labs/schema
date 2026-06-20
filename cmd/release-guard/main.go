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
//	release-guard check-final --tag vX.Y.Z [--workflow release.yml]
//	    Guard a FINAL release: require a same-version release candidate whose
//	    release run is green AND whose tag is an ancestor of the final commit.
//	    Queries git (tag list, rev-list, merge-base) and gh (workflow run
//	    conclusions), then defers the decision to release.CheckFinal.
//
//	release-guard check-maintainer --user <login>
//	    Assert a GitHub user is a release maintainer (admin/maintain collaborator
//	    on $GITHUB_REPOSITORY). Queries gh for the collaborator permission and
//	    applies release.IsMaintainer — the single source of the maintainer
//	    predicate consumed by both release-pr.yml gates.
//
//	release-guard check-approval --pr <number>
//	    Assert a release PR has at least one standing APPROVED review from a
//	    maintainer. Fetches all review pages, reduces each reviewer's latest
//	    non-COMMENTED review, then applies release.IsMaintainer per approver.
//
//	release-guard check-workflow [--release .github/workflows/release.yml]
//	    Validate that the release workflow keeps the GitHub-Release publish job
//	    behind the guard, the Nix vendorHash gate, and the contract gates
//	    (oasdiff / go-apidiff / vacuum). The schema-repo analogue of peasant's
//	    goreleaser-gate check — it asserts the SUBTRACTED (binary-free) pipeline's
//	    own publication gates via release.CheckReleaseWorkflowFile.
//
// Every subcommand that derives a (version, kind) writes "version=<v>" and
// "kind=<k>" to the file named by $GITHUB_OUTPUT when set (so workflow steps can
// consume them via steps.<id>.outputs.*), and always echoes them to stdout.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peasant-labs/schema/internal/release"
)

var commandOutput = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: release-guard <parse-title|parse-tag|check-final|check-workflow|check-maintainer|check-approval> ...")
	}
	sub := os.Args[1]
	args := os.Args[2:]

	switch sub {
	case "parse-title":
		runParseTitle(args)
	case "parse-tag":
		runParseTag(args)
	case "check-final":
		runCheckFinal(args)
	case "check-workflow":
		runCheckWorkflow(args)
	case "check-maintainer":
		runCheckMaintainer(args)
	case "check-approval":
		runCheckApproval(args)
	default:
		fatalf("unknown subcommand %q: expected parse-title, parse-tag, check-final, check-workflow, check-maintainer, or check-approval", sub)
	}
}

// runCheckWorkflow validates the schema repo's release.yml: the GitHub-Release
// publish job must sit behind the guard, the Nix vendorHash gate, and the
// contract gates. This is the (binary-free) schema analogue of peasant's
// goreleaser-gate check; the grammar lives in release.CheckReleaseWorkflow.
func runCheckWorkflow(args []string) {
	releaseWorkflow := ".github/workflows/release.yml"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--release":
			i++
			if i >= len(args) {
				fatalf("--release requires a value")
			}
			releaseWorkflow = args[i]
		default:
			fatalf("unknown flag %q for check-workflow", args[i])
		}
	}
	if err := release.CheckReleaseWorkflowFile(releaseWorkflow); err != nil {
		fatalf("%v", err)
	}
	fmt.Printf("release workflow gates are valid: %s keeps the GitHub-Release publish behind guard, nix-vendor-hash, and contract-gates\n", releaseWorkflow)
}

func runParseTitle(args []string) {
	if len(args) != 1 {
		fatalf("usage: release-guard parse-title \"<pr title>\"")
	}
	v, kind, err := release.ParseReleaseTitle(args[0])
	if err != nil {
		fatalf("%v", err)
	}
	emit(v, kind)
}

func runParseTag(args []string) {
	if len(args) != 1 {
		fatalf("usage: release-guard parse-tag \"<git tag>\"")
	}
	v, kind, err := release.ParseTag(args[0])
	if err != nil {
		fatalf("%v", err)
	}
	emit(v, kind)
}

func runCheckFinal(args []string) {
	var tag, workflow string
	workflow = "release.yml"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--tag":
			i++
			if i >= len(args) {
				fatalf("--tag requires a value")
			}
			tag = args[i]
		case "--workflow":
			i++
			if i >= len(args) {
				fatalf("--workflow requires a value")
			}
			workflow = args[i]
		default:
			fatalf("unknown flag %q for check-final", args[i])
		}
	}
	if tag == "" {
		fatalf("usage: release-guard check-final --tag vX.Y.Z [--workflow release.yml]")
	}

	final, kind, err := release.ParseTag(tag)
	if err != nil {
		fatalf("%v", err)
	}
	if kind != release.KindFinal {
		fatalf("check-final: %s is a release candidate, not a final release; the rc→final guard only applies to final tags", final)
	}

	finalCommit, err := revParse(string(final))
	if err != nil {
		fatalf("check-final: cannot resolve the commit for final tag %s: %v", final, err)
	}

	rcTags, err := listRCTags(final.Base())
	if err != nil {
		fatalf("check-final: cannot list release-candidate tags for %s: %v", final.Base(), err)
	}

	statuses := make([]release.RCStatus, 0, len(rcTags))
	for _, rcTag := range rcTags {
		rcVer, rcKind, perr := release.ParseTag(rcTag)
		if perr != nil || rcKind != release.KindRC || rcVer.Base() != final.Base() {
			continue
		}
		ancestor := isAncestor(string(rcVer), finalCommit)
		green, gerr := runGreen(workflow, string(rcVer))
		if gerr != nil {
			fatalf("check-final: cannot determine the release-run status of %s: %v", rcVer, gerr)
		}
		statuses = append(statuses, release.RCStatus{Tag: rcVer, RunGreen: green, IsAncestor: ancestor})
	}

	if err := release.CheckFinal(final, statuses); err != nil {
		fatalf("%v", err)
	}
	fmt.Printf("final release %s is permitted: a same-version rc is green and an ancestor of the final commit\n", final)
}

func runCheckMaintainer(args []string) {
	var user string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user":
			i++
			if i >= len(args) {
				fatalf("--user requires a value")
			}
			user = args[i]
		default:
			fatalf("unknown flag %q for check-maintainer", args[i])
		}
	}
	if user == "" {
		fatalf("usage: release-guard check-maintainer --user <login>")
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		fatalf("check-maintainer: $GITHUB_REPOSITORY is not set; cannot resolve the collaborator permission for %q", user)
	}

	perm, err := collaboratorPermission(repo, user)
	if err != nil {
		fatalf("check-maintainer: cannot read the collaborator permission of %q on %s (gh api): %v", user, repo, err)
	}
	if !release.IsMaintainer(perm) {
		fatalf("user %q has permission %q on %s, but a release action (open or approve a release PR) requires a maintainer (%s or %s); ask a maintainer to perform it",
			user, perm, repo, release.PermAdmin, release.PermMaintain)
	}
	fmt.Printf("user %s is a maintainer (%s)\n", user, perm)
}

func runCheckApproval(args []string) {
	var pr string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--pr":
			i++
			if i >= len(args) {
				fatalf("--pr requires a value")
			}
			pr = args[i]
		default:
			fatalf("unknown flag %q for check-approval", args[i])
		}
	}
	if pr == "" {
		fatalf("usage: release-guard check-approval --pr <number>")
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		fatalf("check-approval: $GITHUB_REPOSITORY is not set; cannot fetch reviews for PR #%s", pr)
	}

	reviews, err := fetchReviews(repo, pr)
	if err != nil {
		fatalf("check-approval: cannot read reviews for release PR #%s on %s: %v", pr, repo, err)
	}
	approvers := release.LatestApprovers(reviews)
	if len(approvers) == 0 {
		fatalf("release PR #%s has no standing APPROVED reviews; a maintainer (%s or %s) must approve before merge",
			pr, release.PermAdmin, release.PermMaintain)
	}

	approver, perm, err := findMaintainerApprover(repo, approvers)
	if err != nil {
		fatalf("check-approval: %v", err)
	}
	if approver != "" {
		fmt.Printf("release PR #%s has a standing maintainer approval from %s (%s)\n", pr, approver, perm)
		return
	}
	fatalf("release PR #%s has standing approvals, but none from an %s/%s maintainer; a maintainer must approve the release",
		pr, release.PermAdmin, release.PermMaintain)
}

// --- git / gh helpers ---

func revParse(ref string) (string, error) {
	out, err := commandOutput("git", "rev-list", "-n", "1", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func listRCTags(base release.Version) ([]string, error) {
	out, err := commandOutput("git", "tag", "--list", string(base)+"-rc*")
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

func isAncestor(rcRef, finalCommit string) bool {
	// Exit 0 = ancestor; exit 1 = not; >1 = error (treated as not-ancestor, the
	// safe default — the guard then blocks and the maintainer investigates).
	return exec.Command("git", "merge-base", "--is-ancestor", rcRef, finalCommit).Run() == nil
}

// ghRun is the subset of `gh run list --json` output we consume.
type ghRun struct {
	HeadSHA    string `json:"headSha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// runGreen reports whether the given rc tag's commit has a completed, successful
// run of the named workflow. It matches on the commit SHA (a tag run's headSha
// is the tagged commit) so a later run on the same commit is also accepted.
func runGreen(workflow, rcTag string) (bool, error) {
	rcCommit, err := revParse(rcTag)
	if err != nil {
		return false, fmt.Errorf("resolving commit for %s: %w", rcTag, err)
	}
	out, err := commandOutput(
		"gh", "run", "list",
		"--workflow", workflow,
		"--commit", rcCommit,
		"--limit", "100",
		"--json", "headSha,status,conclusion",
	)
	if err != nil {
		return false, fmt.Errorf("gh run list: %w", err)
	}
	var runs []ghRun
	if err := json.Unmarshal(out, &runs); err != nil {
		return false, fmt.Errorf("parsing gh run list output: %w", err)
	}
	for _, r := range runs {
		if r.HeadSHA == rcCommit && r.Status == "completed" && r.Conclusion == "success" {
			return true, nil
		}
	}
	return false, nil
}

func fetchReviews(repo, pr string) ([]release.Review, error) {
	out, err := commandOutput(
		"gh", "api",
		fmt.Sprintf("repos/%s/pulls/%s/reviews", repo, pr),
		"--paginate",
		"--slurp",
	)
	if err != nil {
		return nil, err
	}
	return parseReviews(out)
}

func parseReviews(data []byte) ([]release.Review, error) {
	var pages [][]release.Review
	if err := json.Unmarshal(data, &pages); err == nil {
		var reviews []release.Review
		for _, page := range pages {
			reviews = append(reviews, page...)
		}
		return reviews, nil
	}

	var reviews []release.Review
	if err := json.Unmarshal(data, &reviews); err != nil {
		return nil, fmt.Errorf("parsing gh reviews JSON: %w", err)
	}
	return reviews, nil
}

func collaboratorPermission(repo, user string) (release.CollaboratorPermission, error) {
	out, err := commandOutput(
		"gh", "api",
		fmt.Sprintf("repos/%s/collaborators/%s/permission", repo, user),
		"--jq", ".permission",
	)
	if err != nil {
		return "", err
	}
	return release.CollaboratorPermission(strings.TrimSpace(string(out))), nil
}

func findMaintainerApprover(repo string, approvers []string) (string, release.CollaboratorPermission, error) {
	for _, approver := range approvers {
		perm, err := collaboratorPermission(repo, approver)
		if err != nil {
			return "", "", fmt.Errorf("cannot read the collaborator permission of approving reviewer %q on %s (gh api): %w", approver, repo, err)
		}
		if release.IsMaintainer(perm) {
			return approver, perm, nil
		}
	}
	return "", "", nil
}

// --- output helpers ---

// emit writes version/kind to $GITHUB_OUTPUT (when set) and stdout.
func emit(v release.Version, kind release.ReleaseKind) {
	line := fmt.Sprintf("version=%s\nkind=%s\n", v, kind)
	if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fatalf("cannot open $GITHUB_OUTPUT (%s): %v", path, err)
		}
		defer f.Close()
		if _, err := f.WriteString(line); err != nil {
			fatalf("cannot write to $GITHUB_OUTPUT (%s): %v", path, err)
		}
	}
	fmt.Print(line)
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "release-guard: "+format+"\n", a...)
	os.Exit(1)
}
