package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func unexecutableRunner() *execGit {
	return &execGit{gitPath: filepath.Join("/nonexistent", "definitely-not-git")}
}

func TestGitRunner_RejectsLeadingDashArgs(t *testing.T) {
	t.Parallel()
	g := unexecutableRunner()
	ctx := context.Background()
	hostile := []string{"-e", "--upload-pack=touch /tmp/pwned", "--output=x", "-"}
	for _, arg := range hostile {
		arg := arg
		t.Run("IsAncestor/"+arg, func(t *testing.T) {
			t.Parallel()
			_, err := g.IsAncestor(ctx, arg, "HEAD")
			if err == nil {
				t.Fatalf("expected refusal for hostile arg %q, got nil", arg)
			}
			if !strings.Contains(err.Error(), "begins with '-'") {
				t.Fatalf("error for %q is not the leading-dash refusal: %q", arg, err.Error())
			}
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				t.Fatalf("hostile arg %q reached exec", arg)
			}
		})
		t.Run("ReadFile/"+arg, func(t *testing.T) {
			t.Parallel()
			_, err := g.ReadFile(ctx, arg, "CHANGELOG.md")
			if err == nil {
				t.Fatalf("expected refusal for hostile rev %q, got nil", arg)
			}
		})
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

func TestGitRunner_IsAncestorDistinguishesNegative(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	r, err := newExecGitIn(dir)
	if err != nil {
		t.Fatalf("newExecGitIn: %v", err)
	}
	ctx := context.Background()
	yes, err := r.IsAncestor(ctx, "HEAD~1", "HEAD")
	if err != nil || !yes {
		t.Fatalf("IsAncestor(HEAD~1, HEAD) = %v, %v; want true, nil", yes, err)
	}
	no, err := r.IsAncestor(ctx, "HEAD", "HEAD~1")
	if err != nil {
		t.Fatalf("IsAncestor(HEAD, HEAD~1) returned error for a definitive negative: %v", err)
	}
	if no {
		t.Fatal("IsAncestor(HEAD, HEAD~1) = true, want false")
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
	paths, err := r.ChangedPaths(context.Background(), base, head)
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
	paths, err := r.ChangedPaths(context.Background(), base, head)
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

func TestGitRunner_NonAncestorBaseIsDistinct(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	runGit(t, dir, "checkout", "-q", "-b", "side", "HEAD~1")
	runGit(t, dir, "commit", "-q", "--allow-empty", "-m", "side work")
	side := gitRev(t, dir, "side")
	runGit(t, dir, "checkout", "-q", "master")
	tip := gitRev(t, dir, "HEAD")
	r, _ := newExecGitIn(dir)
	_, err := r.ChangedPaths(context.Background(), side, tip)
	if err == nil {
		t.Fatal("ChangedPaths with a non-ancestor base = nil, want BaseNotAncestorError")
	}
	var notAnc *BaseNotAncestorError
	if !errors.As(err, &notAnc) {
		t.Fatalf("ChangedPaths(non-ancestor) = %T %v, want *BaseNotAncestorError", err, err)
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
	runGit(t, dir, "config", "user.name", "t")
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
	if string(rej.Code) != "protected-path" {
		t.Fatalf("invalid version JSON code = %q, want protected-path", rej.Code)
	}
}

func TestCheckTitle_FailedBlobReadIsInternal(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := initIntentRepo(t)
	r, _ := newExecGitIn(dir)
	ctx := context.Background()
	// Read a blob that does not exist at a valid rev: a BlobReadError, which
	// checkTitle surfaces as an internal error (exit 1), never an accept.
	_, err := r.ReadFile(ctx, "HEAD", "does-not-exist-xyz.txt")
	var blob *BlobReadError
	if !errors.As(err, &blob) {
		t.Fatalf("ReadFile(missing) = %T %v, want *BlobReadError", err, err)
	}
}

// --- temp repo helpers ---

func initIntentRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "master")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
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
