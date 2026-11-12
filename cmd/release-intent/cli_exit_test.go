package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
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
		os.Exit(1)
	}
	testBinaryPath = bin
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
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
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, extraEnv...)
	// Hermetic git identity for temp repos; harmless for non-git commands.
	cmd.Env = append(cmd.Env,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
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

func setupProcessCase(t *testing.T, setup string) (string, map[string]string) {
	t.Helper()
	replace := map[string]string{}
	switch setup {
	case "none":
		dir := t.TempDir()
		return dir, replace
	case "config-bot":
		dir := t.TempDir()
		cfg := "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: automation/release\n  vendor-hash: automation/vendor-hash\n  cut-release: automation/cut-release\n"
		writeSetupFile(t, dir, ".github/release-intent.yaml", cfg)
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
	case "queue-empty":
		dir := t.TempDir()
		data, err := queueJSONFS.ReadFile("testdata/queue/empty.json")
		if err != nil {
			t.Fatalf("read embedded empty queue: %v", err)
		}
		writeSetupFile(t, dir, "queue.json", string(data))
		replace["SETUP"] = dir
		return dir, replace
	case "queue-malformed":
		dir := t.TempDir()
		data, err := queueJSONFS.ReadFile("testdata/queue/malformed.json")
		if err != nil {
			t.Fatalf("read embedded malformed queue: %v", err)
		}
		writeSetupFile(t, dir, "queue.json", string(data))
		replace["SETUP"] = dir
		return dir, replace
	case "git-basic":
		repo := initIntentRepo(t)
		base := gitRev(t, repo, "HEAD~1")
		head := gitRev(t, repo, "HEAD")
		replace["SETUP_BASE"] = base
		replace["SETUP_HEAD"] = head
		replace["SETUP"] = repo
		return repo, replace
	default:
		t.Fatalf("unknown process setup %q", setup)
		return "", nil
	}
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
			dir, replace := setupProcessCase(t, c.Input.Setup)
			argv := make([]string, len(c.Input.Argv))
			for i, a := range c.Input.Argv {
				for k, v := range replace {
					a = strings.ReplaceAll(a, k, v)
				}
				argv[i] = a
			}
			exit, stdout, _ := runBinary(t, dir, argv, c.Input.Stdin, nil)
			if exit != c.Expected.Exit {
				t.Fatalf("%s: exit = %d, want %d\nstdout: %q", c.Name, exit, c.Expected.Exit, stdout)
			}
			for _, want := range c.Expected.Stdout {
				if !strings.Contains(stdout, want) {
					t.Fatalf("%s: stdout %q does not contain %q", c.Name, stdout, want)
				}
			}
			if c.Expected.NoGate {
				if strings.Contains(stdout, "gate=") {
					t.Fatalf("%s: stdout %q must not contain a gate line on a failed read", c.Name, stdout)
				}
				if strings.Contains(stdout, "candidate_fresh") {
					t.Fatalf("%s: stdout %q must not contain candidate_fresh on a read error", c.Name, stdout)
				}
			}
		})
	}
}

// TestGITHUBOutputAppend proves key=value lines append to $GITHUB_OUTPUT.
// It uses t.Setenv (process-global), so it must not run in parallel.
func TestGITHUBOutputAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(path, []byte("preexisting=1\n"), 0o644); err != nil {
		t.Fatalf("seed $GITHUB_OUTPUT: %v", err)
	}
	t.Setenv("GITHUB_OUTPUT", path)
	exit, stdout, _ := runBinary(t, t.TempDir(), []string{"group-pr", "--ref", "refs/heads/gh-readonly-queue/main/pr-12-abc123"}, "", nil)
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
