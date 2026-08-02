package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

// --- R6 hardening: leading-dash rejection (BDD #6) --------------------------

// A runner whose gitPath cannot execute: if a dynamic arg were NOT rejected
// before exec, the call would fail with an exec error instead of the
// leading-dash refusal — so asserting the refusal message proves the guard
// short-circuits BEFORE any process is spawned.
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
		t.Run("RevParse/"+arg, func(t *testing.T) {
			t.Parallel()
			_, err := g.RevParse(ctx, arg)
			assertLeadingDashRefusal(t, err, arg)
		})
		t.Run("ListTags/"+arg, func(t *testing.T) {
			t.Parallel()
			_, err := g.ListTags(ctx, arg)
			assertLeadingDashRefusal(t, err, arg)
		})
		t.Run("IsAncestor/"+arg, func(t *testing.T) {
			t.Parallel()
			_, err := g.IsAncestor(ctx, arg, "HEAD")
			assertLeadingDashRefusal(t, err, arg)
		})
	}
}

func assertLeadingDashRefusal(t *testing.T, err error, arg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected refusal for hostile arg %q, got nil", arg)
	}
	msg := err.Error()
	if !strings.Contains(msg, "begins with '-'") || !strings.Contains(msg, arg) {
		t.Fatalf("error for %q is not the actionable leading-dash refusal: %q", arg, msg)
	}
	// It must NOT have reached exec (no ExitError) — rejection happens first.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		t.Fatalf("hostile arg %q reached exec (got ExitError) — rejection must happen first", arg)
	}
}

// --- R6 hardening: subcommand allowlist -------------------------------------

func TestGitRunner_SubcommandAllowlist(t *testing.T) {
	t.Parallel()

	g := unexecutableRunner()
	ctx := context.Background()

	// An off-allowlist subcommand is refused before exec.
	_, err := g.runGit(ctx, gitSubcommand("push"), nil, "origin")
	if err == nil {
		t.Fatal("expected refusal for off-allowlist subcommand 'push', got nil")
	}
	if !strings.Contains(err.Error(), "not in the release-guard allowlist") {
		t.Fatalf("allowlist error not actionable: %q", err.Error())
	}

	// The three allowed subcommands pass the allowlist gate (they then fail at
	// exec because gitPath is unexecutable — proving the gate let them through).
	for _, sub := range []gitSubcommand{gitRevList, gitTag, gitMergeBase} {
		_, err := g.runGit(ctx, sub, nil, "HEAD")
		if err == nil {
			t.Fatalf("expected exec failure for %q with an unexecutable git, got nil", sub)
		}
		if strings.Contains(err.Error(), "not in the release-guard allowlist") {
			t.Fatalf("allowed subcommand %q was wrongly rejected by the allowlist", sub)
		}
	}
}

// --- PATH resolution safety -------------------------------------------------

func TestNewExecGit_ResolvesAbsoluteGitPath(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH; skipping PATH-resolution check")
	}
	r, err := newExecGit()
	if err != nil {
		t.Fatalf("newExecGit: %v", err)
	}
	g, ok := r.(*execGit)
	if !ok {
		t.Fatalf("newExecGit returned %T, want *execGit", r)
	}
	// The stdlib resolves git to an ABSOLUTE path off PATH (never a cwd-relative
	// one — the Go 1.19 exec.ErrDot fix), which is the safety property that lets
	// us skip cli/safeexec.
	if !filepath.IsAbs(g.gitPath) {
		t.Fatalf("resolved git path %q is not absolute (cwd-relative resolution is the hole safeexec patches)", g.gitPath)
	}
}

// --- Real-git semantics: byte-equivalent to the old inline helpers ----------

func TestGitRunner_RealGitSemantics(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH; skipping real-git semantics check")
	}
	repo := initTestRepo(t)
	r, err := newExecGit()
	if err != nil {
		t.Fatalf("newExecGit: %v", err)
	}
	g := r.(*execGit)
	g.dir = repo // point at the throwaway repo without mutating process cwd
	ctx := context.Background()

	// RevParse resolves a tag to its 40-hex commit SHA.
	sha, err := g.RevParse(ctx, "v1.0.0-rc1")
	if err != nil {
		t.Fatalf("RevParse(v1.0.0-rc1): %v", err)
	}
	if len(sha) != 40 || strings.TrimLeft(sha, "0123456789abcdef") != "" {
		t.Fatalf("RevParse returned %q, want a 40-char hex SHA", sha)
	}

	// ListTags matches the rc glob (and excludes the non-rc tag).
	tags, err := g.ListTags(ctx, "v1.0.0-rc*")
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 2 || tags[0] != "v1.0.0-rc1" || tags[1] != "v1.0.0-rc2" {
		t.Fatalf("ListTags = %v, want [v1.0.0-rc1 v1.0.0-rc2]", tags)
	}

	// IsAncestor: rc1's commit is an ancestor of HEAD; HEAD is not an ancestor
	// of rc1.
	yes, err := g.IsAncestor(ctx, "v1.0.0-rc1", "HEAD")
	if err != nil {
		t.Fatalf("IsAncestor(rc1, HEAD): %v", err)
	}
	if !yes {
		t.Fatal("IsAncestor(rc1, HEAD) = false, want true")
	}
	no, err := g.IsAncestor(ctx, "HEAD", "v1.0.0-rc1")
	if err != nil {
		t.Fatalf("IsAncestor(HEAD, rc1) returned error for a definitive negative: %v", err)
	}
	if no {
		t.Fatal("IsAncestor(HEAD, rc1) = true, want false")
	}
}

func TestRunCheckFinal_InitialFinalThroughRealRepository(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH; skipping real repository initial-final check")
	}
	repo := t.TempDir()
	gitEnv := hermeticGitTestEnv(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = gitEnv
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	run("commit", "-q", "--allow-empty", "-m", "initial final source")
	run("tag", "v1.0.0")

	runner, err := newExecGit()
	if err != nil {
		t.Fatalf("newExecGit: %v", err)
	}
	runner.(*execGit).dir = repo
	completed := false
	gh := &mockGitHubClient{workflowRunsForCommitFn: func(context.Context, string, string, string) ([]release.WorkflowRun, error) {
		t.Fatal("bootstrap without rc tags must not query rc workflow runs")
		return nil, nil
	}, releaseExistsFn: func(context.Context, string, string) (bool, error) {
		return completed, nil
	}}
	if err := runCheckFinal(context.Background(), gh, runner, testRepo, []string{"--tag", "v1.0.0", "--initial-final", "v1.0.0"}); err != nil {
		t.Fatalf("runCheckFinal through a real fresh git repository: %v", err)
	}
	completed = true
	err = runCheckFinal(context.Background(), gh, runner, testRepo, []string{"--tag", "v1.0.0", "--initial-final", "v1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "already has a published repository release") {
		t.Fatalf("second run after repository release completion = %v, want self-disable rejection", err)
	}
}

// initTestRepo builds a throwaway git repo with two commits; v1.0.0-rc1 + the
// non-rc decoy v1.0.0 tag the first commit, v1.0.0-rc2 tags the second (HEAD).
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitEnv := hermeticGitTestEnv(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = gitEnv
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	run("commit", "-q", "--allow-empty", "-m", "c1")
	run("tag", "v1.0.0-rc1")
	run("tag", "v1.0.0") // non-rc decoy: must be excluded by the rc glob
	run("commit", "-q", "--allow-empty", "-m", "c2")
	run("tag", "v1.0.0-rc2")
	return dir
}

func hermeticGitTestEnv(t *testing.T) []string {
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

func TestInitTestRepo_IgnoresAmbientGitConfig(t *testing.T) {
	poison := filepath.Join(t.TempDir(), "poison.gitconfig")
	if err := os.WriteFile(poison, []byte("[commit]\n\tgpgSign = true\n[gpg]\n\tprogram = /definitely-not-a-signer\n"), 0o600); err != nil {
		t.Fatalf("write hostile git config: %v", err)
	}
	t.Setenv("HOME", filepath.Dir(poison))
	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(poison))
	t.Setenv("GIT_CONFIG_SYSTEM", poison)
	t.Setenv("GIT_CONFIG_GLOBAL", poison)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "commit.gpgSign")
	t.Setenv("GIT_CONFIG_VALUE_0", "true")

	initTestRepo(t)
}
