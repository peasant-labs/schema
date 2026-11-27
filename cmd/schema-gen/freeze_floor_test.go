package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/freeze_floor.yaml
var freezeFloorYAML []byte

//go:embed testdata/freeze_floor_manifest.yaml
var freezeFloorManifestYAML []byte

type freezeFloorEntry struct {
	Label  string  `yaml:"label"`
	Config *string `yaml:"config"`
	Side   bool    `yaml:"side"`
	Merge  bool    `yaml:"merge"`
}

type freezeFloorInput struct {
	History []freezeFloorEntry `yaml:"history"`
}

type freezeFloorExpected struct {
	Floor         string `yaml:"floor"`
	ErrorContains string `yaml:"error_contains"`
}

// freezeFloorScript returns the run text of the release job's freeze-floor
// step, so the test executes the workflow's own search.
func freezeFloorScript(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release-batch.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	steps := workflowMappingChild(workflowMappingChild(workflowMappingChild(doc.Content[0], "jobs"), "plan"), "steps")
	if steps == nil {
		t.Fatal("release-batch.yml plan job has no steps")
	}
	for _, step := range steps.Content {
		if id := workflowMappingChild(step, "id"); id != nil && id.Value == "freeze-floor" {
			if run := workflowMappingChild(step, "run"); run != nil {
				return run.Value
			}
		}
	}
	t.Fatal("release-batch.yml plan job has no freeze-floor step")
	return ""
}

func TestFreezeFloorSearch(t *testing.T) {
	corpus, err := testcase.LoadCorpus[freezeFloorInput, freezeFloorExpected](freezeFloorYAML)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "freeze floor", freezeFloorManifestYAML, names)
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	script := freezeFloorScript(t, root)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			env := append(os.Environ(),
				"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
				"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
			git := func(args ...string) string {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
				}
				return strings.TrimSpace(string(out))
			}
			write := func(path, content string) {
				t.Helper()
				full := filepath.Join(dir, path)
				if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q", "-b", "develop")
			commits := make(map[string]string)
			sideOpen := false
			for _, entry := range c.Input.History {
				switch {
				case entry.Merge:
					if !sideOpen {
						t.Fatalf("entry %q merges without a side branch", entry.Label)
					}
					git("merge", "-q", "--no-ff", "-m", entry.Label, "side")
					sideOpen = false
				default:
					if entry.Side && !sideOpen {
						git("checkout", "-q", "-b", "side")
						sideOpen = true
					}
					if entry.Config != nil {
						write(".github/release-intent.yaml", *entry.Config)
					} else {
						write("content.txt", entry.Label+"\n")
					}
					git("add", "-A")
					git("commit", "-q", "-m", entry.Label)
					if entry.Side {
						commits[entry.Label] = git("rev-parse", "HEAD")
						git("checkout", "-q", "develop")
						continue
					}
				}
				commits[entry.Label] = git("rev-parse", "HEAD")
			}
			git("update-ref", "refs/remotes/origin/develop", "develop")
			output := filepath.Join(t.TempDir(), "output")
			cmd := exec.Command("bash", "-c", script)
			cmd.Dir = dir
			cmd.Env = append(env, "GITHUB_OUTPUT="+output)
			out, err := cmd.CombinedOutput()
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(string(out), c.Expected.ErrorContains) {
					t.Fatalf("expected refusal containing %q; err=%v output=%s", c.Expected.ErrorContains, err, out)
				}
				return
			}
			if err != nil {
				t.Fatalf("freeze-floor step: %v\n%s", err, out)
			}
			recorded, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			want, ok := commits[c.Expected.Floor]
			if !ok {
				t.Fatalf("expected floor %q names no history entry", c.Expected.Floor)
			}
			if got := strings.TrimSpace(string(recorded)); got != "freeze_floor="+want {
				t.Fatalf("freeze-floor output = %q, want freeze_floor=%s (%s)", got, want, c.Expected.Floor)
			}
		})
	}
}
