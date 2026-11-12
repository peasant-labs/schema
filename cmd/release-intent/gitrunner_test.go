package main

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/git-hostile-args.yaml
var hostileArgsYAML []byte

type hostileArgInput struct {
	Arg string `yaml:"arg"`
}

type hostileArgExpected struct {
	Refusal string `yaml:"refusal"`
}

type hostileArgFixtures struct {
	RequiredNames []string                                             `yaml:"required_names"`
	Cases         testcase.Corpus[hostileArgInput, hostileArgExpected] `yaml:"cases"`
}

func loadHostileArgFixtures(t *testing.T) hostileArgFixtures {
	t.Helper()
	var f hostileArgFixtures
	if err := yaml.Unmarshal(hostileArgsYAML, &f); err != nil {
		t.Fatalf("parse hostile args fixture: %v", err)
	}
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("hostile args fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "hostile args fixtures")
	// Every hostile row must name the dash refusal: the guard it pins.
	for _, c := range f.Cases.Cases {
		if c.Expected.Refusal == "" {
			t.Fatalf("hostile args fixture %q declares no refusal substring", c.Name)
		}
	}
	return f
}

func realGitRunner(t *testing.T, dir string) GitRunner {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	r, err := newExecGitIn(dir)
	if err != nil {
		t.Fatalf("newExecGitIn: %v", err)
	}
	return r
}

// The dash guard must fire before exec for every runner method, including
// git show. A real git path proves the refusal happens first: without the
// guard the call would reach git and fail as a blob error instead.
func TestGitRunner_RejectsLeadingDashArgs(t *testing.T) {
	t.Parallel()
	f := loadHostileArgFixtures(t)
	dir := t.TempDir()
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			g := realGitRunner(t, dir)
			ctx := context.Background()
			assertDashRefusal(t, func() error {
				_, err := g.ReadFile(ctx, c.Input.Arg, "CHANGELOG.md")
				return err
			}, c.Input.Arg, c.Expected.Refusal)
			assertDashRefusal(t, func() error {
				_, _, err := g.ChangedPaths(ctx, c.Input.Arg, "HEAD")
				return err
			}, c.Input.Arg, c.Expected.Refusal)
			assertDashRefusal(t, func() error {
				_, err := g.FirstParent(ctx, c.Input.Arg, "CHANGELOG.md")
				return err
			}, c.Input.Arg, c.Expected.Refusal)
			assertDashRefusal(t, func() error {
				_, err := g.Tags(ctx, c.Input.Arg)
				return err
			}, c.Input.Arg, c.Expected.Refusal)
		})
	}
}

func assertDashRefusal(t *testing.T, call func() error, arg, want string) {
	t.Helper()
	err := call()
	if err == nil {
		t.Fatalf("expected refusal for hostile arg %q, got nil", arg)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error for %q is not the leading-dash refusal naming %q: %q", arg, want, err.Error())
	}
	var blob *BlobReadError
	if errors.As(err, &blob) {
		t.Fatalf("hostile arg %q reached git and failed as a blob read; the dash guard must refuse first", arg)
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		t.Fatalf("hostile arg %q reached exec (got ExitError); the dash guard must refuse first", arg)
	}
}

func TestGitRunner_SubcommandAllowlist(t *testing.T) {
	t.Parallel()
	g := unexecutableRunner()
	ctx := context.Background()
	_, err := g.runGit(ctx, gitSubcommand("push"), nil, "origin")
	if err == nil {
		t.Fatal("expected refusal for off-allowlist subcommand 'push', got nil")
	}
	if !strings.Contains(err.Error(), "not in the release-intent allowlist") {
		t.Fatalf("allowlist error not actionable: %q", err.Error())
	}
}

func unexecutableRunner() *execGit {
	return &execGit{gitPath: filepath.Join("/nonexistent", "definitely-not-git")}
}

func TestGitRunner_FirstParentListsEmptyCommits(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "commit", "-q", "--allow-empty", "-m", "release[patch]: empty cut (#3)")
	r, _ := newExecGitIn(dir)
	commits, err := r.FirstParent(context.Background(), "HEAD", "CHANGELOG.md")
	if err != nil {
		t.Fatalf("FirstParent: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("FirstParent listed %d commits, want 3 (empty commits are not dropped)", len(commits))
	}
	if commits[0].Subject != "release[patch]: empty cut (#3)" {
		t.Fatalf("newest commit = %q, want the empty release commit", commits[0].Subject)
	}
	if commits[0].ChangesChangelog {
		t.Fatalf("empty commit must not be marked as changing the changelog")
	}
}

func TestGitRunner_FirstCommitAddingFollowsFirstParent(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "master")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t@t")
	writeFile(t, dir, "work.txt", "base\n")
	runGit(t, dir, "add", "work.txt")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: base work")
	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, ".github/release-intent.yaml", "releases: true\n")
	runGit(t, dir, "add", ".github/release-intent.yaml")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: adopt config")
	runGit(t, dir, "checkout", "-q", "master")
	runGit(t, dir, "merge", "-q", "--no-ff", "-m", "merge adopt config", "side")
	r, _ := newExecGitIn(dir)
	ctx := context.Background()
	mergeSHA := gitRev(t, dir, "HEAD")
	got, err := r.FirstCommitAdding(ctx, "master", ".github/release-intent.yaml")
	if err != nil {
		t.Fatalf("FirstCommitAdding: %v", err)
	}
	if got != mergeSHA {
		t.Fatalf("FirstCommitAdding = %q, want the no-ff merge commit %q", got, mergeSHA)
	}
}

func TestGitRunner_TagsResolveAnnotatedAndLightweight(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "tag", "-a", "v0.2.0", "-m", "annotated")
	r, _ := newExecGitIn(dir)
	tags, err := r.Tags(context.Background(), "v")
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("Tags = %v, want exactly v0.1.0 and v0.2.0 in one git call", tags)
	}
	for tag, sha := range tags {
		if len(sha) != 40 {
			t.Fatalf("tag %q resolved to %q, want a 40-char peeled commit", tag, sha)
		}
	}
}

func TestGitRunner_ChangedPathsSeesRenamedBothNames(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "mv", "protected.txt", "renamed.txt")
	runGit(t, dir, "commit", "-q", "-m", "rename protected")
	base := gitRev(t, dir, "HEAD~1")
	head := gitRev(t, dir, "HEAD")
	r, _ := newExecGitIn(dir)
	_, paths, err := r.ChangedPaths(context.Background(), base, head)
	if err != nil {
		t.Fatalf("ChangedPaths: %v", err)
	}
	seen := map[string]bool{}
	for _, p := range paths {
		seen[p] = true
	}
	if !seen["protected.txt"] || !seen["renamed.txt"] {
		t.Fatalf("renamed file must surface as both names without rename detection, got %v", paths)
	}
}

func TestGitRunner_ChangedPathsSeesDeleted(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "rm", "-q", "protected.txt")
	runGit(t, dir, "commit", "-q", "-m", "delete protected")
	base := gitRev(t, dir, "HEAD~1")
	head := gitRev(t, dir, "HEAD")
	r, _ := newExecGitIn(dir)
	_, paths, err := r.ChangedPaths(context.Background(), base, head)
	if err != nil {
		t.Fatalf("ChangedPaths: %v", err)
	}
	found := false
	for _, p := range paths {
		if p == "protected.txt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("deleted protected file must be listed, got %v", paths)
	}
}

func TestGitRunner_ChangedPathsFailsWithoutMergeBase(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "checkout", "-q", "--orphan", "side")
	writeFile(t, dir, "other.txt", "unrelated\n")
	runGit(t, dir, "add", "other.txt")
	runGit(t, dir, "commit", "-q", "-m", "unrelated root")
	side := gitRev(t, dir, "side")
	runGit(t, dir, "checkout", "-q", "master")
	tip := gitRev(t, dir, "HEAD")
	r, _ := newExecGitIn(dir)
	_, _, err := r.ChangedPaths(context.Background(), side, tip)
	if err == nil {
		t.Fatal("ChangedPaths with no merge base = nil, want an error naming the missing merge base")
	}
	if !strings.Contains(err.Error(), "no merge base") {
		t.Fatalf("ChangedPaths error %q does not name the missing merge base", err.Error())
	}
}

func TestCheckTitle_MissingBaseConfigAdopts(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	// The adopting PR adds the config: build a repo where the base commit has
	// no config and the head adds it. The base is an ancestor of the head.
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "master")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t@t")
	writeFile(t, dir, "work.txt", "base\n")
	runGit(t, dir, "add", "work.txt")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: base work")
	base := gitRev(t, dir, "HEAD")
	writeFile(t, dir, ".github/release-intent.yaml", "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: automation/release\n")
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: adopt config")
	head := gitRev(t, dir, "HEAD")
	r, _ := newExecGitIn(dir)
	ctx := context.Background()
	// A good title with no protected change must be accepted even though the
	// base has no config: grammar only, no path protection yet.
	if err := checkTitle(ctx, r, base, head, "octocat", "feature/work", "feat(graph)[minor]: add facet"); err != nil {
		t.Fatalf("adopting PR with good title: %v", err)
	}
}

func TestCheckTitle_InvalidJSONFailClosed(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	writeFile(t, dir, "package.json", "{not json")
	runGit(t, dir, "add", "package.json")
	runGit(t, dir, "commit", "-q", "-m", "break json")
	base := gitRev(t, dir, "HEAD~1")
	head := gitRev(t, dir, "HEAD")
	r, _ := newExecGitIn(dir)
	err := checkTitle(context.Background(), r, base, head, "octocat", "feature/work", "feat(graph)[minor]: add facet")
	rej, ok := asReject(err)
	if !ok {
		t.Fatalf("invalid version JSON must fail closed as a rejection, got %T %v", err, err)
	}
	if string(rej.Err.Code) != "protected-path" {
		t.Fatalf("invalid version JSON code = %q, want protected-path", rej.Err.Code)
	}
}

// --- temp repo helpers ---

func initIntentRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "master")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t@t")
	writeFile(t, dir, ".github/release-intent.yaml", "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: automation/release\n")
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	writeFile(t, dir, "protected.txt", "guarded\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(init)[minor]: seed")
	runGit(t, dir, "tag", "v0.1.0")
	writeFile(t, dir, "work.txt", "more\n")
	runGit(t, dir, "add", "work.txt")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: add facet")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func runGitDated(t *testing.T, dir, date string, args ...string) {
	t.Helper()
	env := hermeticGitEnv(t)
	env = append(env, "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func gitRev(t *testing.T, dir, rev string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-list", "-n", "1", rev)
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-list %q: %v", rev, err)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", name, err)
	}
}

func hermeticGitEnv(t *testing.T) []string {
	t.Helper()
	home := t.TempDir()
	env := make([]string, 0, len(os.Environ())+7)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") || key == "HOME" || key == "XDG_CONFIG_HOME" {
			continue
		}
		env = append(env, entry)
	}
	return append(env,
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, "xdg"),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
}

func TestDebugChangedPaths(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "mv", "protected.txt", "renamed.txt")
	runGit(t, dir, "commit", "-q", "-m", "rename protected")
	base := gitRev(t, dir, "HEAD~1")
	head := gitRev(t, dir, "HEAD")
	t.Logf("base=%s head=%s", base, head)
	cmd := exec.Command("git", "log", "--oneline", "-3")
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv(t)
	out, _ := cmd.CombinedOutput()
	t.Logf("log:\n%s", out)
	r, _ := newExecGitIn(dir)
	mb, paths, err := r.ChangedPaths(context.Background(), base, head)
	t.Logf("mb=%q paths=%v err=%v", mb, paths, err)
}

func TestDebugDashDirect(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	g, err := newExecGitIn(dir)
	if err != nil {
		t.Fatalf("newExecGitIn: %v", err)
	}
	out, err := g.(*execGit).runGit(context.Background(), gitShow, []string{"--end-of-options"}, "-e:CHANGELOG.md")
	t.Logf("out=%q err=%v (%T)", string(out), err, err)
	if err == nil || !strings.Contains(err.Error(), "begins with") {
		t.Fatalf("no refusal: out=%q err=%v", string(out), err)
	}
}
