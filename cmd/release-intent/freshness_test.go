package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/freshness.yaml
var freshnessFixtureYAML []byte

type freshInput struct {
	Style    string `yaml:"style"`
	Scenario string `yaml:"scenario"`
}

type freshExpected struct {
	Exit     int      `yaml:"exit"`
	Action   string   `yaml:"action"`
	Fresh    string   `yaml:"fresh"`
	Error    string   `yaml:"error"`
	Lines    []string `yaml:"lines"`
	// CheckLines is the exact stdout of the release-pr --check that follows
	// the write, for scenarios that run both.
	CheckLines []string `yaml:"check_lines"`
	Absent   []string `yaml:"absent"`
	Sections int      `yaml:"sections"`
	Version  string   `yaml:"version"`
}

type freshFixtures struct {
	RequiredNames []string                                   `yaml:"required_names"`
	Cases         testcase.Corpus[freshInput, freshExpected] `yaml:"cases"`
}

func loadFreshnessFixtures(t *testing.T) freshFixtures {
	t.Helper()
	var f freshFixtures
	if err := yaml.Unmarshal(freshnessFixtureYAML, &f); err != nil {
		t.Fatalf("parse freshness fixture: %v", err)
	}
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("freshness fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "freshness fixtures")
	checkCorpusSiblings(t, freshnessFixtureYAML, "freshness fixtures")
	return f
}

const freshPendingDate = "2020-01-02T00:00:00Z"

func freshBaseChangelog(style string) string {
	switch style {
	case "bare":
		return "# changelog\n\n## 0.1.0\n\n- old\n"
	case "bracket-undated":
		return "# changelog\n\n## [v0.1.0]\n\n- old\n"
	case "bracket-dated":
		return "# changelog\n\n## [v0.1.0] - 2020-01-01\n\n- old\n"
	case "no-section":
		return "# changelog\n"
	default:
		return "# changelog\n"
	}
}

func newFreshRepo(t *testing.T, style string) string {
	t.Helper()
	dir := newBareRepo(t)
	writeFile(t, dir, ".github/release-intent.yaml", checkTitleTestConfig)
	writeFile(t, dir, "CHANGELOG.md", freshBaseChangelog(style))
	writeFile(t, dir, "package.json", "{\"version\": \"0.1.0\"}\n")
	runGit(t, dir, "add", ".")
	runGitDated(t, dir, "2020-01-01T00:00:00Z", "commit", "-q", "-m", "feat(init)[minor]: seed")
	// Annotated, like the fairtrade and schema release tags: the plan path
	// must run on the tag shape with a tag object between the name and the
	// commit.
	runGit(t, dir, "tag", "-a", "v0.1.0", "-m", "release 0.1.0")
	return dir
}

func freshCommitPending(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, dir, "work.txt", "more\n")
	runGit(t, dir, "add", "work.txt")
	runGitDated(t, dir, freshPendingDate, "commit", "-q", "-m", "feat(graph)[patch]: add facet")
}

func freshCommitRelease(t *testing.T, dir, date, subject string) {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGitDated(t, dir, date, "commit", "-q", "-m", subject)
}

func gitTagPointsAt(t *testing.T, dir, rev string) string {
	t.Helper()
	cmd := exec.Command("git", "tag", "--points-at", rev)
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git tag --points-at %q: %v", rev, err)
	}
	return strings.TrimSpace(string(out))
}

func planOutput(t *testing.T, dir string, extra ...string) (int, map[string]string, string) {
	t.Helper()
	argv := append([]string{"plan", "--branch", "HEAD", "--mode", "tag", "--verdict", "taggable"}, extra...)
	exit, stdout, _ := runBinary(t, dir, argv, "", nil)
	out := map[string]string{}
	for _, line := range strings.Split(strings.TrimSuffix(stdout, "\n"), "\n") {
		if line == "" {
			continue
		}
		k, v, _ := strings.Cut(line, "=")
		out[k] = v
	}
	return exit, out, stdout
}

// TestFreshness drives the built binary on temp repositories: write, commit,
// plan and check across heading styles, stale histories, read errors and the
// stale-rebuild section count. Commit dates are pinned away from the test day
// so dated headings are deterministic.
func TestFreshness(t *testing.T) {
	f := loadFreshnessFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			dir := newFreshRepo(t, c.Input.Style)
			switch c.Input.Scenario {
			case "fresh-after-write":
				freshCommitPending(t, dir)
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				freshCommitRelease(t, dir, "2020-01-03T00:00:00Z", "release[patch]: cut it")
				exit, got, _ := planOutput(t, dir)
				assertFreshRow(t, c.Name, c.Expected, exit, got)
			case "stale-new-releasing", "check-stale-tree":
				// A release built from an older tip: write, let one releasing
				// commit land, then commit the stale content as the release.
				freshCommitPending(t, dir)
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				writeFile(t, dir, "later.txt", "later\n")
				runGit(t, dir, "add", "later.txt")
				runGitDated(t, dir, "2020-01-04T00:00:00Z", "commit", "-q", "-m", "feat(graph)[patch]: later facet")
				freshCommitRelease(t, dir, "2020-01-05T00:00:00Z", "release[patch]: cut it")
				if c.Input.Scenario == "check-stale-tree" {
					exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--check"}, "", nil)
					if exit != c.Expected.Exit {
						t.Fatalf("%s: release-pr --check exit = %d, want %d\nstdout: %q", c.Name, exit, c.Expected.Exit, stdout)
					}
					assertExactLines(t, stdout, c.Expected.Lines, c.Name)
					return
				}
				exit, got, _ := planOutput(t, dir)
				assertFreshRow(t, c.Name, c.Expected, exit, got)
			case "older-fresh-newer-stale":
				freshCommitPending(t, dir)
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				// The older release commit is fresh but stays untagged, so it
				// is a candidate, never the range base.
				freshCommitRelease(t, dir, "2020-01-03T00:00:00Z", "release[patch]: cut it")
				older := gitRev(t, dir, "HEAD")
				// A new pending subject, then a hand-built stale release: its
				// bytes do not match the generator, so the newer candidate is
				// stale.
				writeFile(t, dir, "later.txt", "later\n")
				runGit(t, dir, "add", "later.txt")
				runGitDated(t, dir, "2020-01-04T00:00:00Z", "commit", "-q", "-m", "feat(graph)[patch]: later facet")
				writeFile(t, dir, "CHANGELOG.md", "# changelog\n\n## 0.1.1\n\n- hand-written entry\n")
				writeFile(t, dir, "package.json", "{\"version\": \"0.1.1\"}\n")
				runGit(t, dir, "add", "-A")
				runGitDated(t, dir, "2020-01-05T00:00:00Z", "commit", "-q", "-m", "release[patch]: cut again")
				newer := gitRev(t, dir, "HEAD")
				exit, got, _ := planOutput(t, dir)
				assertFreshRow(t, c.Name, c.Expected, exit, got)
				if got["candidate"] != newer {
					t.Fatalf("%s: candidate = %q, want the newer stale release %q", c.Name, got["candidate"], newer)
				}
				if out := gitTagPointsAt(t, dir, older); out != "" {
					t.Fatalf("%s: older release commit carries tags %q, want none", c.Name, out)
				}
			case "freshness-read-error":
				freshCommitPending(t, dir)
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				runGit(t, dir, "rm", "-qf", "package.json")
				freshCommitRelease(t, dir, "2020-01-03T00:00:00Z", "release[patch]: cut it")
				exit, got, stdout := planOutput(t, dir)
				if exit != c.Expected.Exit {
					t.Fatalf("%s: exit = %d, want %d\nstdout: %q", c.Name, exit, c.Expected.Exit, stdout)
				}
				for _, absent := range c.Expected.Absent {
					if strings.Contains(stdout, absent) {
						t.Fatalf("%s: stdout %q must not contain %q on a read error", c.Name, stdout, absent)
					}
				}
				_ = got
			case "rejected-later-minor-writes-full", "fresh-candidate-later-subject-checks-full":
				freshCommitPending(t, dir)
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				freshCommitRelease(t, dir, "2020-01-03T00:00:00Z", "release[patch]: cut it")
				// A later minor that touches no changelog file: the candidate
				// stays fresh and the plan speaks for the before set (0.1.1),
				// yet the release content must be the full set (0.2.0).
				writeFile(t, dir, "later.txt", "later\n")
				runGit(t, dir, "add", "later.txt")
				runGitDated(t, dir, "2020-01-04T00:00:00Z", "commit", "-q", "-m", "feat(graph)[minor]: later facet")
				if exit, got, _ := planOutput(t, dir); exit != 0 || got["candidate_fresh"] != "true" {
					t.Fatalf("%s: candidate must stay fresh, got exit=%d plan=%v", c.Name, exit, got)
				}
				exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil)
				if exit != c.Expected.Exit {
					t.Fatalf("%s: release-pr --write exit = %d, want %d\nstdout: %q", c.Name, exit, c.Expected.Exit, stdout)
				}
				assertExactLines(t, stdout, c.Expected.Lines, c.Name)
				data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
				if err != nil {
					t.Fatalf("%s: read rewritten changelog: %v", c.Name, err)
				}
				if !strings.Contains(string(data), "## 0.2.0") || !strings.Contains(string(data), "later facet") || !strings.Contains(string(data), "add facet") {
					t.Fatalf("%s: rewritten changelog must hold the 0.2.0 section with both subjects:\n%s", c.Name, string(data))
				}
				pkg, err := os.ReadFile(filepath.Join(dir, "package.json"))
				if err != nil {
					t.Fatalf("%s: read rewritten package.json: %v", c.Name, err)
				}
				if !strings.Contains(string(pkg), `"version": "0.2.0"`) {
					t.Fatalf("%s: rewritten package.json must carry 0.2.0:\n%s", c.Name, string(pkg))
				}
				if c.Input.Scenario == "fresh-candidate-later-subject-checks-full" {
					// The check compares the tree with the same full set.
					exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--check"}, "", nil)
					if exit != c.Expected.Exit {
						t.Fatalf("%s: release-pr --check exit = %d, want %d\nstdout: %q", c.Name, exit, c.Expected.Exit, stdout)
					}
					assertExactLines(t, stdout, c.Expected.CheckLines, c.Name)
				}
			case "stale-rebuild-single-section":
				freshCommitPending(t, dir)
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				writeFile(t, dir, "later.txt", "later\n")
				runGit(t, dir, "add", "later.txt")
				runGitDated(t, dir, "2020-01-04T00:00:00Z", "commit", "-q", "-m", "feat(graph)[patch]: later facet")
				freshCommitRelease(t, dir, "2020-01-05T00:00:00Z", "release[patch]: cut it")
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != c.Expected.Exit {
					t.Fatalf("%s: rebuild exit = %d, want %d\nstdout: %q", c.Name, exit, c.Expected.Exit, stdout)
				}
				data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
				if err != nil {
					t.Fatalf("%s: read rebuilt changelog: %v", c.Name, err)
				}
				n := 0
				for _, line := range strings.Split(string(data), "\n") {
					if strings.HasPrefix(strings.TrimSpace(line), "## ") && strings.Contains(line, "0.1.1") {
						n++
					}
				}
				if n != c.Expected.Sections {
					t.Fatalf("%s: rebuilt changelog has %d sections for 0.1.1, want %d (exactly one section per version)\n%s", c.Name, n, c.Expected.Sections, string(data))
				}
			default:
				t.Fatalf("unknown freshness scenario %q", c.Input.Scenario)
			}
		})
	}
}

func assertFreshRow(t *testing.T, name string, want freshExpected, exit int, got map[string]string) {
	t.Helper()
	if exit != want.Exit {
		t.Fatalf("%s: exit = %d, want %d (plan %v)", name, exit, want.Exit, got)
	}
	if want.Action != "" && got["action"] != want.Action {
		t.Fatalf("%s: action = %q, want %q", name, got["action"], want.Action)
	}
	if want.Fresh != "" && got["candidate_fresh"] != want.Fresh {
		t.Fatalf("%s: candidate_fresh = %q, want %q", name, got["candidate_fresh"], want.Fresh)
	}
	if want.Error != "" && got["error"] != want.Error {
		t.Fatalf("%s: error = %q, want %q", name, got["error"], want.Error)
	}
	if want.Version != "" && got["version"] != want.Version {
		t.Fatalf("%s: version = %q, want %q", name, got["version"], want.Version)
	}
}

func assertExactLines(t *testing.T, out string, want []string, name string) {
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
