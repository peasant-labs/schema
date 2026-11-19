package main

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/freeze_contract_titles.yaml
var freezeContractTitles []byte

//go:embed testdata/freeze_contract_titles_manifest.yaml
var freezeContractTitlesManifest []byte

type freezeTitleInput struct {
	Event   string       `yaml:"event"`
	Title   string       `yaml:"title"`
	Local   reportChange `yaml:"local"`
	Ahead   bool         `yaml:"ahead"`
	Sibling string       `yaml:"sibling"`
}

type freezeTitleExpected struct {
	Exit int    `yaml:"exit"`
	Code string `yaml:"code"`
}

func TestFreezeContractTitles(t *testing.T) {
	corpus, err := testcase.LoadCorpus[freezeTitleInput, freezeTitleExpected](freezeContractTitles)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	var names []string
	accepts := make(map[string]bool)
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
		accepts[c.Name] = c.Classification == testcase.MustPass
	}
	caseassert.RequireManifestNames(t, "freeze contract title", freezeContractTitlesManifest, names)
	var tool struct {
		Pin string `yaml:"tool_pin"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &tool); err != nil || len(tool.Pin) != 40 {
		t.Fatalf("release tool pin is invalid: %q, %v", tool.Pin, err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	binDir := t.TempDir()
	install := exec.CommandContext(ctx, "go", "install", "github.com/peasant-labs/infra/tools/release/cmd/release-intent@"+tool.Pin)
	install.Dir = root
	install.Env = append(os.Environ(), "GOBIN="+binDir, "GOWORK=off", "GOFLAGS=")
	if output, err := install.CombinedOutput(); err != nil {
		t.Fatalf("build pinned release-intent: %v\n%s", err, output)
	}
	dir := t.TempDir()
	env := append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir, cmd.Env = dir, env
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "-q")
	if err := os.Mkdir(filepath.Join(dir, ".github"), 0o755); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(root, ".github/release-intent.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github/release-intent.yaml"), config, 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-qm", "ci: configure release")
	git("tag", "v0.1.0")
	head := git("rev-parse", "HEAD")
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Classification == testcase.MustFail && !accepts[c.Input.Sibling] {
				t.Fatal("title refusal has no accepting sibling")
			}
			if c.Input.Event != "pr" && c.Input.Event != "queue" {
				t.Fatalf("unknown title event %q", c.Input.Event)
			}
			report := contractReport{Report: 1, Base: head, Head: head}
			for _, surface := range reportSurfaces {
				row := surfaceReport{Name: surface.name, Change: reportNone}
				if surface.name == "local-api" {
					row.Change, row.AheadOnBase = c.Input.Local, c.Input.Ahead
				}
				report.Surfaces = append(report.Surfaces, row)
			}
			data, err := yaml.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			reportFile := filepath.Join(t.TempDir(), "report.yaml")
			if err := os.WriteFile(reportFile, data, 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctx, filepath.Join(binDir, "release-intent"), "check-contract", "--report", reportFile)
			cmd.Dir, cmd.Env, cmd.Stdin = dir, env, bytes.NewBufferString(c.Input.Title+"\n")
			output, err := cmd.CombinedOutput()
			exit := 0
			if err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					exit = e.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if exit != c.Expected.Exit || (c.Expected.Code != "" && !strings.Contains(string(output), "error="+c.Expected.Code)) {
				t.Fatalf("check-contract exit=%d, want %d and %q:\n%s", exit, c.Expected.Exit, c.Expected.Code, output)
			}
		})
	}
}
