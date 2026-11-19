package main

import (
	_ "embed"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

//go:embed testdata/output.yaml
var outputCasesYAML []byte

//go:embed testdata/output_manifest.yaml
var outputManifestYAML []byte

func TestFormatOutput_ByteParity(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Name string `yaml:"name"`
			Tag  string `yaml:"tag"`
			Want string `yaml:"want"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(outputCasesYAML, &fixture); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(fixture.Cases))
	for _, c := range fixture.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "output cases", outputManifestYAML, names)
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			v, kind, err := release.ParseTag(c.Tag)
			if err != nil {
				t.Fatal(err)
			}
			if got := formatOutput(v, kind); got != c.Want {
				t.Fatalf("formatOutput(%q) = %q, want %q", c.Tag, got, c.Want)
			}
		})
	}
}

func TestRunParseTag_RejectsNamespacedTag(t *testing.T) {
	t.Parallel()
	if err := runParseTag([]string{"pkg/schema/v1.2.3"}); err == nil {
		t.Fatal("runParseTag(namespaced tag) = nil, want rejection")
	}
}

func TestEmit_AppendsToGitHubOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "github_output")
	// Seed a prior line so the test proves emit APPENDS (O_APPEND), never
	// truncates — the file accumulates outputs across a job's steps.
	if err := os.WriteFile(path, []byte("preexisting=1\n"), 0o644); err != nil {
		t.Fatalf("seed $GITHUB_OUTPUT: %v", err)
	}
	t.Setenv("GITHUB_OUTPUT", path)

	v, kind, err := release.ParseTag("v1.2.3-rc4")
	if err != nil {
		t.Fatalf("ParseTag: %v", err)
	}
	if err := emit(v, kind); err != nil {
		t.Fatalf("emit: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read $GITHUB_OUTPUT: %v", err)
	}
	want := "preexisting=1\nversion=v1.2.3-rc4\nkind=rc\n"
	if string(got) != want {
		t.Fatalf("$GITHUB_OUTPUT = %q, want %q (emit appends version=/kind=)", string(got), want)
	}
}
