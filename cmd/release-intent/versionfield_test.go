package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// fairtradeShapedPackageJSON is deliberately hostile to whole-file rewrites:
// unsorted keys, scripts with shell operators the JSON escaper would mangle,
// and the protected field in the middle.
const fairtradeShapedPackageJSON = `{"name": "fairtrade",
  "scripts": {"lint": "a && b > c", "build": "x || y"},
  "version": "0.2.0",
  "dependencies": {"left-pad": "1.3.0"}
}
`

func TestWriteReleaseContent_OnlyVersionLineChanges(t *testing.T) {
	// No t.Parallel: the test changes the process-global working directory
	// because writeReleaseContent resolves repo-relative paths like the binary
	// does when run from the repository root.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "master")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t@t")
	writeFile(t, dir, "package.json", fairtradeShapedPackageJSON)
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "seed")
	cfg := intent.Config{
		Changelog:    "CHANGELOG.md",
		VersionFiles: []intent.VersionFile{{Path: "package.json", Field: "version"}},
	}
	content := intent.Content{
		Changelog: "# changelog\n",
		Section:   "## 0.2.0\n",
		Versions:  map[string]string{"package.json": "0.2.1"},
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir("/"); err != nil {
			t.Fatalf("chdir back: %v", err)
		}
	}()
	if err := writeReleaseContent(cfg, content); err != nil {
		t.Fatalf("writeReleaseContent: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		t.Fatalf("read rewritten package.json: %v", err)
	}
	if strings.Contains(string(got), "\\u0026") || strings.Contains(string(got), "\\u003e") {
		t.Fatalf("rewrote scripts with escapes, want byte-identical surrounding text:\n%s", got)
	}
	if !strings.Contains(string(got), `"lint": "a && b > c"`) {
		t.Fatalf("scripts line changed, want byte-identical:\n%s", got)
	}
	cmd := exec.Command("git", "diff", "--numstat", "--", "package.json")
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git diff --numstat: %v", err)
	}
	if strings.TrimSpace(string(out)) != "1\t1\tpackage.json" {
		t.Fatalf("git diff --numstat = %q, want one added and one deleted line", out)
	}
}
