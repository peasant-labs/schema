package main

import (
	_ "embed"
	"os"
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
	runGit(t, dir, "tag", "v0.1.0")
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
				freshCommitRelease(t, dir, "2020-01-03T00:00:00Z", "release[patch]: cut it")
				runGit(t, dir, "tag", "v0.1.1")
				writeFile(t, dir, "later.txt", "later\n")
				runGit(t, dir, "add", "later.txt")
				runGitDated(t, dir, "2020-01-04T00:00:00Z", "commit", "-q", "-m", "feat(graph)[patch]: later facet")
				if exit, stdout, _ := runBinary(t, dir, []string{"release-pr", "--branch", "HEAD", "--write"}, "", nil); exit != 0 {
					t.Fatalf("%s: second release-pr --write exit = %d, want 0\nstdout: %q", c.Name, exit, stdout)
				}
				writeFile(t, dir, "newest.txt", "newest\n")
				runGit(t, dir, "add", "newest.txt")
				runGitDated(t, dir, "2020-01-05T00:00:00Z", "commit", "-q", "-m", "feat(graph)[patch]: newest facet")
				freshCommitRelease(t, dir, "2020-01-06T00:00:00Z", "release[patch]: cut again")
				writeFile(t, dir, "final.txt", "final\n")
				runGit(t, dir, "add", "final.txt")
				runGitDated(t, dir, "2020-01-07T00:00:00Z", "commit", "-q", "-m", "feat(graph)[patch]: final facet")
				exit, got, _ := planOutput(t, dir)
				assertFreshRow(t, c.Name, c.Expected, exit, got)
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
