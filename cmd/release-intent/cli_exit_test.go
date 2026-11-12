package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var testBinaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "release-intent-bin")
	if err != nil {
		os.Stderr.WriteString("TestMain: cannot create temp dir for the test binary: " + err.Error() + "\n")
		os.Exit(1)
	}
	bin := filepath.Join(dir, "release-intent")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Stderr.WriteString("TestMain: cannot build the release-intent binary once for process tests: " + err.Error() + "\n" + string(out))
		os.RemoveAll(dir)
		os.Exit(1)
	}
	testBinaryPath = bin
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// childEnv builds the hermetic environment for the child process: it drops
// ambient GITHUB_OUTPUT (unless the row sets it through extraEnv), GIT_DIR and
// GIT_WORK_TREE so a caller running under git exec cannot redirect the binary
// at the wrong repository, and pins the git identity the temp repos need.
func childEnv(t *testing.T, extra []string) []string {
	t.Helper()
	home := t.TempDir()
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "GITHUB_OUTPUT", "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_PREFIX", "GIT_NAMESPACE", "HOME", "XDG_CONFIG_HOME":
			continue
		}
		if strings.HasPrefix(key, "GIT_") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env,
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, "xdg"),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
	return append(env, extra...)
}

func runBinary(t *testing.T, dir string, argv []string, stdin string, extraEnv []string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(testBinaryPath, argv...)
	cmd.Dir = dir
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	} else {
		cmd.Stdin = strings.NewReader("")
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = childEnv(t, extraEnv)
	err := cmd.Run()
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			t.Fatalf("run binary %v: %v\nstderr: %s", argv, err, stderr.String())
		}
	}
	return exit, stdout.String(), stderr.String()
}

func setupProcessCase(t *testing.T, setup, queueFile string) (string, map[string]string) {
	t.Helper()
	replace := map[string]string{}
	switch setup {
	case "none":
		dir := t.TempDir()
		return dir, replace
	case "config-valid":
		dir := t.TempDir()
		cfg := "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\n"
		writeSetupFile(t, dir, "config.yaml", cfg)
		replace["SETUP"] = dir
		return dir, replace
	case "config-unknown-key":
		dir := t.TempDir()
		cfg := "releases: true\ntag_prefix: v\nbogus_key: true\n"
		writeSetupFile(t, dir, "bad.yaml", cfg)
		replace["SETUP"] = dir
		return dir, replace
	case "queue":
		repo := initIntentRepo(t)
		data, err := queueJSONFS.ReadFile("testdata/queue/" + queueFile)
		if err != nil {
			t.Fatalf("read embedded queue file %q: %v", queueFile, err)
		}
		writeSetupFile(t, repo, "queue.json", string(data))
		replace["SETUP"] = repo
		return repo, replace
	case "git-basic":
		repo := initIntentRepo(t)
		base := gitRev(t, repo, "HEAD~1")
		head := gitRev(t, repo, "HEAD")
		replace["SETUP_BASE"] = base
		replace["SETUP_HEAD"] = head
		replace["SETUP"] = repo
		return repo, replace
	case "git-bad-subject":
		repo := newBareRepo(t)
		writeFile(t, repo, ".github/release-intent.yaml", checkTitleTestConfig)
		writeFile(t, repo, "CHANGELOG.md", "# changelog\n")
		writeFile(t, repo, "package.json", "{\"version\": \"0.1.0\"}\n")
		runGit(t, repo, "add", ".")
		runGit(t, repo, "commit", "-q", "-m", "feat(init)[minor]: seed")
		runGit(t, repo, "tag", "v0.1.0")
		writeFile(t, repo, "work.txt", "try\n")
		runGit(t, repo, "add", "work.txt")
		runGit(t, repo, "commit", "-q", "-m", "wip(x): try")
		replace["SETUP"] = repo
		return repo, replace
	case "classify-repo":
		repo := newBareRepo(t)
		writeFile(t, repo, ".github/release-intent.yaml", "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: automation/release\n  vendor-hash: automation/vendor-hash\n  cut-release: automation/cut-release\n")
		runGit(t, repo, "add", ".")
		runGit(t, repo, "commit", "-q", "-m", "feat(init)[minor]: seed")
		replace["SETUP_REV"] = gitRev(t, repo, "HEAD")
		replace["SETUP"] = repo
		return repo, replace
	case "classify-base":
		repo := newBareRepo(t)
		writeFile(t, repo, ".github/release-intent.yaml", "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: automation/release\n")
		writeFile(t, repo, "CHANGELOG.md", "# changelog\n")
		writeFile(t, repo, "package.json", "{\"version\": \"0.1.0\"}\n")
		runGit(t, repo, "add", ".")
		runGit(t, repo, "commit", "-q", "-m", "feat(init)[minor]: seed")
		replace["SETUP_BASE"] = gitRev(t, repo, "HEAD")
		writeFile(t, repo, ".github/release-intent.yaml", "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: feature/work\n")
		runGit(t, repo, "add", ".github/release-intent.yaml")
		runGit(t, repo, "commit", "-q", "-m", "feat(graph)[minor]: tamper")
		replace["SETUP_HEAD"] = gitRev(t, repo, "HEAD")
		replace["SETUP"] = repo
		return repo, replace
	case "workflows-accept":
		dir := writeWorkflowsSetup(t, false)
		replace["SETUP"] = dir
		return dir, replace
	case "workflows-violation", "policy-only":
		dir := writeWorkflowsSetup(t, setup == "workflows-violation")
		replace["SETUP"] = dir
		return dir, replace
	default:
		t.Fatalf("unknown process setup %q", setup)
		return "", nil
	}
}

// writeWorkflowsSetup writes a minimal policy plus workflows dir. With queue
// true the policy names two required contexts (reversed, to prove sorting)
// that the empty workflows dir cannot satisfy; without queue the policy is
// jobs-only and the checker accepts without reading the dir.
func writeWorkflowsSetup(t *testing.T, queue bool) string {
	t.Helper()
	dir := t.TempDir()
	policy := "jobs:\n  - name: guard\n"
	if queue {
		policy = "jobs:\n  - name: guard\nqueue:\n  required:\n    - {context: title-guard, workflow: title-guard.yml, job: title-guard}\n    - {context: lib, workflow: ci.yml, job: lib}\n"
	}
	writeSetupFile(t, dir, "policy.yml", policy)
	if err := os.MkdirAll(filepath.Join(dir, "workflows"), 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	return dir
}

func writeSetupFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir for %q: %v", name, err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write setup %q: %v", name, err)
	}
}

func TestCLIExitCodes(t *testing.T) {
	f := loadCLIFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			dir, replace := setupProcessCase(t, c.Input.Setup, c.Input.QueueFile)
			ordered := make([]string, 0, len(replace))
			for k := range replace {
				ordered = append(ordered, k)
			}
			sort.Slice(ordered, func(i, j int) bool { return len(ordered[i]) > len(ordered[j]) })
			argv := make([]string, len(c.Input.Argv))
			for i, a := range c.Input.Argv {
				for _, k := range ordered {
					a = strings.ReplaceAll(a, k, replace[k])
				}
				argv[i] = a
			}
			var extra []string
			var githubOutput string
			if c.Input.GithubOutput {
				githubOutput = filepath.Join(t.TempDir(), "github_output")
				if err := os.WriteFile(githubOutput, []byte("seed=1\n"), 0o644); err != nil {
					t.Fatalf("seed $GITHUB_OUTPUT: %v", err)
				}
				extra = []string{"GITHUB_OUTPUT=" + githubOutput}
			}
			exit, stdout, stderr := runBinary(t, dir, argv, c.Input.Stdin, extra)
			if exit != c.Expected.Exit {
				t.Fatalf("%s: exit = %d, want %d\nstdout: %q\nstderr: %q", c.Name, exit, c.Expected.Exit, stdout, stderr)
			}
			lines := splitStdoutLines(stdout)
			if len(lines) != len(c.Expected.Stdout) {
				t.Fatalf("%s: stdout has %d lines %q, want %d %q (exact match, in order)", c.Name, len(lines), stdout, len(c.Expected.Stdout), c.Expected.Stdout)
			}
			for i := range c.Expected.Stdout {
				if lines[i] != c.Expected.Stdout[i] {
					t.Fatalf("%s: stdout line %d = %q, want %q (exact match)", c.Name, i, lines[i], c.Expected.Stdout[i])
				}
			}
			for _, absent := range c.Expected.Absent {
				if strings.Contains(stdout, absent) {
					t.Fatalf("%s: stdout %q must not contain %q", c.Name, stdout, absent)
				}
			}
			for _, want := range c.Expected.Stderr {
				if !strings.Contains(stderr, want) {
					t.Fatalf("%s: stderr %q does not contain %q", c.Name, stderr, want)
				}
			}
			if c.Input.GithubOutput {
				got, err := os.ReadFile(githubOutput)
				if err != nil {
					t.Fatalf("read $GITHUB_OUTPUT: %v", err)
				}
				if string(got) != "seed=1\n" {
					t.Fatalf("$GITHUB_OUTPUT = %q, want the seed unchanged (print-required must not append step outputs)", string(got))
				}
			}
		})
	}
}

func splitStdoutLines(stdout string) []string {
	trimmed := strings.TrimSuffix(stdout, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// TestGITHUBOutputAppend proves key=value lines append to $GITHUB_OUTPUT.
// The output path travels through the child's environment, never t.Setenv.
func TestGITHUBOutputAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(path, []byte("preexisting=1\n"), 0o644); err != nil {
		t.Fatalf("seed $GITHUB_OUTPUT: %v", err)
	}
	exit, stdout, _ := runBinary(t, t.TempDir(), []string{"group-pr", "--ref", "refs/heads/gh-readonly-queue/main/pr-12-abc123"}, "", []string{"GITHUB_OUTPUT=" + path})
	if exit != 0 {
		t.Fatalf("group-pr exit = %d, want 0", exit)
	}
	if !strings.Contains(stdout, "pr=12") {
		t.Fatalf("stdout %q missing pr=12", stdout)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read $GITHUB_OUTPUT: %v", err)
	}
	if string(got) != "preexisting=1\npr=12\n" {
		t.Fatalf("$GITHUB_OUTPUT = %q, want %q (append, never truncate)", string(got), "preexisting=1\npr=12\n")
	}
}
