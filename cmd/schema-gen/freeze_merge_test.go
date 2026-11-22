package main

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/freeze_merge.yaml
var freezeMergeCases []byte

//go:embed testdata/freeze_merge_manifest.yaml
var freezeMergeManifest []byte

type freezeMergeInput struct {
	DocsEdit    freezeWriterEdit `yaml:"docs_edit"`
	BaseEdit    freezeWriterEdit `yaml:"base_edit"`
	ContentEdit freezeWriterEdit `yaml:"content_edit"`
	Ahead       []string         `yaml:"ahead"`
	DocsMarker  string           `yaml:"docs_marker"`
}

type freezeMergeExpected struct {
	Fresh bool `yaml:"fresh"`
}

func TestFreezeMergeFreshness(t *testing.T) {
	requireReportOasdiff(t)
	corpus, err := testcase.LoadCorpus[freezeMergeInput, freezeMergeExpected](freezeMergeCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "freeze merge", freezeMergeManifest, names)
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "repository")
			copyFreezeSource(t, root, dir)
			deps := filepath.Join(root, "typescript", "node_modules")
			_, depsErr := os.Stat(filepath.Join(deps, ".bin", "openapi-ts"))
			_, pnpmErr := exec.LookPath("pnpm")
			fullFreshness := depsErr == nil && pnpmErr == nil
			if !fullFreshness && os.Getenv("CI") != "" {
				t.Fatal("full merge freshness needs pnpm and the locked TypeScript dependencies")
			}
			env := append(os.Environ(), "GOWORK=off", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
				"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
				"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
			run := func(command string, args ...string) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
				defer cancel()
				cmd := exec.CommandContext(ctx, command, args...)
				cmd.Dir, cmd.Env = dir, env
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s %v: %v\n%s", command, args, err, output)
				}
			}
			git := func(args ...string) {
				t.Helper()
				run("git", append([]string{"-c", "commit.gpgsign=false"}, args...)...)
			}
			if fullFreshness {
				run("pnpm", "--dir", "typescript", "install", "--offline", "--frozen-lockfile", "--ignore-scripts")
			}
			releaseFreezeBaseline(t, dir)
			edit := func(change freezeWriterEdit) {
				t.Helper()
				filename := filepath.Join(dir, change.Path)
				data, err := os.ReadFile(filename)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Count(string(data), change.Find) != 1 {
					t.Fatalf("edit anchor in %s is not unique", change.Path)
				}
				if err := os.WriteFile(filename, []byte(strings.Replace(string(data), change.Find, change.Replace, 1)), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q", "-b", "release-source")
			edit(c.Input.DocsEdit)
			run("go", "run", "./cmd/schema-gen")
			docsState, err := calculateSurfaceStates(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range docsState.Surfaces {
				if slices.Contains(c.Input.Ahead, row.Name) && row.State != surfaceDocsAhead {
					t.Fatalf("description-only %s state = %s, want docs-ahead", row.Name, row.State)
				}
			}
			edit(c.Input.BaseEdit)
			run("go", "run", "./cmd/schema-gen")
			if fullFreshness {
				run("pnpm", "--dir", "typescript", "run", "generate")
			}
			git("add", "-A")
			git("commit", "-qm", "feat(types,local-api)[minor]: base change")
			git("switch", "-qc", "content")
			edit(c.Input.ContentEdit)
			run("go", "run", "./cmd/schema-gen")
			if fullFreshness {
				run("pnpm", "--dir", "typescript", "run", "generate")
			}
			git("add", "-A")
			git("commit", "-qm", "feat(types,local-api)[minor]: pending change")
			git("switch", "-q", "release-source")
			state, err := calculateSurfaceStates(dir)
			if err != nil {
				t.Fatal(err)
			}
			requireAheadSurfaces(t, state, c.Input.Ahead)
			plan := minorFreezePlan(t, state)
			planFile, stateFile := filepath.Join(t.TempDir(), "plan.yaml"), filepath.Join(t.TempDir(), "state.yaml")
			for name, value := range map[string]any{planFile: plan, stateFile: state} {
				data, err := yaml.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			run("go", "run", "./cmd/schema-gen", "freeze", "--plan", planFile, "--state", stateFile)
			for i, surface := range reportSurfaces {
				if plan.Surfaces[i].Freeze && surface.name == "types" {
					frozen, err := os.ReadFile(filepath.Join(dir, "generated", "types-"+plan.Surfaces[i].To+".json"))
					if err != nil || !strings.Contains(string(frozen), c.Input.DocsMarker) {
						t.Fatalf("frozen Types document lacks the pending docs: %v", err)
					}
				}
			}
			git("add", "-A")
			git("commit", "-qm", "release[minor]: freeze")
			git("switch", "-q", "content")
			git("merge", "--no-edit", "release-source")
			// No regeneration is done between the merge and the freshness gate.
			if fullFreshness {
				run("make", "freshness")
			} else {
				run("go", "test", "./cmd/schema-gen", "-run", "^TestCodegenFreshness_SpecsMatchSource$", "-count=1")
			}
		})
	}
}
