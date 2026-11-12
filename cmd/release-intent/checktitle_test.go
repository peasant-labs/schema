package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/check-title.yaml
var checkTitleFixtureYAML []byte

type checkTitleInput struct {
	Title      string `yaml:"title"`
	Author     string `yaml:"author"`
	HeadBranch string `yaml:"headBranch"`
	Setup      string `yaml:"setup"`
}

type checkTitleExpected struct {
	Exit   int      `yaml:"exit"`
	Code   string   `yaml:"code"`
	Stdout []string `yaml:"stdout"`
	Absent []string `yaml:"absent"`
}

type checkTitleFixtures struct {
	RequiredNames []string                                             `yaml:"required_names"`
	Cases         testcase.Corpus[checkTitleInput, checkTitleExpected] `yaml:"cases"`
}

func loadCheckTitleFixtures(t *testing.T) checkTitleFixtures {
	t.Helper()
	var f checkTitleFixtures
	if err := yaml.Unmarshal(checkTitleFixtureYAML, &f); err != nil {
		t.Fatalf("parse check-title fixture: %v", err)
	}
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("check-title fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "check-title fixtures")
	checkCorpusSiblings(t, checkTitleFixtureYAML, "check-title fixtures")
	return f
}

// mockGitRunner returns programmed git results without touching git. It mocks
// the git dependency, never checkTitle: every method returns the scripted
// outcome for the scenario under test.
type mockGitRunner struct {
	firstParent func(ctx context.Context, ref, changelog string) ([]intent.Commit, error)
	tags        func(ctx context.Context, prefix string) (map[string]string, error)
	firstAdding func(ctx context.Context, branch, path string) (string, error)
	changed     func(ctx context.Context, base, head string) (string, []string, error)
	readFile    func(ctx context.Context, rev, path string) ([]byte, error)
	mergeTree   func(ctx context.Context, tip, current string) error
}

func (m *mockGitRunner) FirstParent(ctx context.Context, ref, changelog string) ([]intent.Commit, error) {
	return m.firstParent(ctx, ref, changelog)
}
func (m *mockGitRunner) Tags(ctx context.Context, prefix string) (map[string]string, error) {
	return m.tags(ctx, prefix)
}
func (m *mockGitRunner) FirstCommitAdding(ctx context.Context, branch, path string) (string, error) {
	return m.firstAdding(ctx, branch, path)
}
func (m *mockGitRunner) ChangedPaths(ctx context.Context, base, head string) (string, []string, error) {
	return m.changed(ctx, base, head)
}
func (m *mockGitRunner) ReadFile(ctx context.Context, rev, path string) ([]byte, error) {
	return m.readFile(ctx, rev, path)
}
func (m *mockGitRunner) MergeTree(ctx context.Context, tip, current string) error {
	if m.mergeTree != nil {
		return m.mergeTree(ctx, tip, current)
	}
	return fmt.Errorf("mock runner: MergeTree is not scripted for this scenario")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe for stdout capture: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func checkTitleExit(err error) int {
	if err == nil {
		return 0
	}
	if _, ok := asReject(err); ok {
		return 2
	}
	return 1
}

// TestCheckTitleTempRepos drives checkTitle on temp repositories described by
// the YAML fixture: one row per protected-path and base shape. Direct calls
// never print the machine line (that happens in exitWith); accept rows assert
// a nil error and reject rows assert the typed code.
func TestCheckTitleTempRepos(t *testing.T) {
	f := loadCheckTitleFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			var base, head string
			var runner GitRunner
			if c.Input.Setup == "mock-base-failure" {
				// The base config blob read fails (a failed on-demand fetch,
				// not an absent file) while the head config reads fine.
				// checkTitle must fail closed: a base read error is never the
				// adopting PR.
				fetchErr := &BlobReadError{Rev: "base", Path: DefaultConfigPath, Err: fmt.Errorf("simulated on-demand fetch failure")}
				runner = &mockGitRunner{
					readFile: func(ctx context.Context, rev, path string) ([]byte, error) {
						if rev == "base" {
							return nil, fetchErr
						}
						return []byte(checkTitleTestConfig), nil
					},
				}
				base, head = "base", "head"
			} else {
				switch c.Input.Setup {
				case "protection-removed":
					base, head, runner = setupProtectionRemoved(t)
				case "pr-behind-base":
					base, head, runner = setupPRBehindBase(t)
				case "scripts-only-after-release":
					base, head, runner = setupScriptsOnlyAfterRelease(t)
				case "version-field-edit":
					base, head, runner = setupVersionFieldEdit(t)
				case "group-shaped-base":
					base, head, runner = setupGroupShapedBase(t)
				case "ordinary-scripts-edit":
					base, head, runner = setupOrdinaryScriptsEdit(t)
				default:
					t.Fatalf("unknown check-title setup %q", c.Input.Setup)
				}
			}
			var runErr error
			out := captureStdout(t, func() {
				runErr = checkTitle(context.Background(), runner, base, head, c.Input.Author, c.Input.HeadBranch, c.Input.Title)
			})
			assertExactStdout(t, out, c.Expected.Stdout, c.Name)
			for _, absent := range c.Expected.Absent {
				if strings.Contains(out, absent) {
					t.Fatalf("%s: stdout %q must not contain %q", c.Name, out, absent)
				}
			}
			if exit := checkTitleExit(runErr); exit != c.Expected.Exit {
				t.Fatalf("%s: exit = %d, want %d (err %v)", c.Name, exit, c.Expected.Exit, runErr)
			}
			if c.Expected.Code != "" {
				rej, ok := asReject(runErr)
				if !ok {
					t.Fatalf("%s: expected rejection with code %q, got %v", c.Name, c.Expected.Code, runErr)
				}
				if string(rej.Err.Code) != c.Expected.Code {
					t.Fatalf("%s: rejection code = %q, want %q", c.Name, rej.Err.Code, c.Expected.Code)
				}
			} else if c.Expected.Exit == 0 && runErr != nil {
				t.Fatalf("%s: expected accept, got err %v", c.Name, runErr)
			}
		})
	}
}

func assertExactStdout(t *testing.T, out string, want []string, name string) {
	t.Helper()
	trimmed := strings.TrimSuffix(out, "\n")
	var lines []string
	if trimmed != "" {
		lines = strings.Split(trimmed, "\n")
	}
	if len(lines) != len(want) {
		t.Fatalf("%s: stdout has %d lines %q, want %d %q (exact match)", name, len(lines), out, len(want), want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("%s: stdout line %d = %q, want %q (exact match)", name, i, lines[i], want[i])
		}
	}
}

const checkTitleTestConfig = "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files:\n  - {path: package.json, field: version}\nprotected_paths: []\nbot_branches:\n  release: automation/release\n"

func setupOrdinaryScriptsEdit(t *testing.T) (string, string, GitRunner) {
	t.Helper()
	dir := initIntentRepo(t)
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\", \"scripts\": {\"lint\": \"quiet\"}}\n")
	runGit(t, dir, "add", "package.json")
	runGit(t, dir, "commit", "-q", "-m", "fix(scripts)[patch]: quiet lint")
	return gitRev(t, dir, "HEAD~1"), gitRev(t, dir, "HEAD"), mustRunner(t, dir)
}

func setupProtectionRemoved(t *testing.T) (string, string, GitRunner) {
	t.Helper()
	dir := newBareRepo(t)
	writeFile(t, dir, ".github/release-intent.yaml", checkTitleTestConfig)
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(init)[minor]: seed")
	base := gitRev(t, dir, "HEAD")
	writeFile(t, dir, ".github/release-intent.yaml", "releases: true\ntag_prefix: v\nautomation_author: peasant-labs-releaser[bot]\nchangelog: CHANGELOG.md\nversion_files: []\nprotected_paths: []\nbot_branches:\n  release: automation/release\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.2.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: drop protection and bump")
	return base, gitRev(t, dir, "HEAD"), mustRunner(t, dir)
}

func setupPRBehindBase(t *testing.T) (string, string, GitRunner) {
	t.Helper()
	dir := newBareRepo(t)
	writeFile(t, dir, ".github/release-intent.yaml", checkTitleTestConfig)
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	writeFile(t, dir, "work.txt", "base\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(init)[minor]: seed")
	runGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "feature.txt", "work\n")
	runGit(t, dir, "add", "feature.txt")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[minor]: branch work")
	head := gitRev(t, dir, "HEAD")
	runGit(t, dir, "checkout", "-q", "master")
	writeFile(t, dir, "later.txt", "later\n")
	runGit(t, dir, "add", "later.txt")
	runGit(t, dir, "commit", "-q", "-m", "docs(readme)[patch]: later words")
	return gitRev(t, dir, "HEAD"), head, mustRunner(t, dir)
}

func setupScriptsOnlyAfterRelease(t *testing.T) (string, string, GitRunner) {
	t.Helper()
	dir := newBareRepo(t)
	writeFile(t, dir, ".github/release-intent.yaml", checkTitleTestConfig)
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(init)[minor]: seed")
	runGit(t, dir, "tag", "v0.1.0")
	runGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\", \"scripts\": {\"lint\": \"quiet\"}}\n")
	runGit(t, dir, "add", "package.json")
	runGit(t, dir, "commit", "-q", "-m", "fix(scripts)[patch]: quiet lint")
	head := gitRev(t, dir, "HEAD")
	runGit(t, dir, "checkout", "-q", "master")
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n\n## 0.2.0\n\n- release one (#1)\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.2.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "release[patch]: v0.2.0 (#1)")
	return gitRev(t, dir, "HEAD"), head, mustRunner(t, dir)
}

func setupVersionFieldEdit(t *testing.T) (string, string, GitRunner) {
	t.Helper()
	dir := newBareRepo(t)
	writeFile(t, dir, ".github/release-intent.yaml", checkTitleTestConfig)
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(init)[minor]: seed")
	base := gitRev(t, dir, "HEAD")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.1\"}\n")
	runGit(t, dir, "add", "package.json")
	runGit(t, dir, "commit", "-q", "-m", "feat(graph)[patch]: bump")
	return base, gitRev(t, dir, "HEAD"), mustRunner(t, dir)
}

func setupGroupShapedBase(t *testing.T) (string, string, GitRunner) {
	t.Helper()
	dir := newBareRepo(t)
	writeFile(t, dir, ".github/release-intent.yaml", checkTitleTestConfig)
	writeFile(t, dir, "CHANGELOG.md", "# changelog\n")
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat(init)[minor]: seed")
	runGit(t, dir, "checkout", "-q", "-b", "queue-a")
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-q", "-m", "feat(a)[minor]: entry a")
	base := gitRev(t, dir, "HEAD")
	runGit(t, dir, "checkout", "-q", "master")
	runGit(t, dir, "checkout", "-q", "-b", "queue-b")
	writeFile(t, dir, "b.txt", "b\n")
	runGit(t, dir, "add", "b.txt")
	runGit(t, dir, "commit", "-q", "-m", "feat(b)[minor]: entry b")
	return base, gitRev(t, dir, "HEAD"), mustRunner(t, dir)
}

func newBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "master")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t@t")
	return dir
}

func mustRunner(t *testing.T, dir string) GitRunner {
	t.Helper()
	r, err := newExecGitIn(dir)
	if err != nil {
		t.Fatalf("newExecGitIn: %v", err)
	}
	return r
}
