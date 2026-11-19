package main

import (
	"bytes"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
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

//go:embed testdata/cli.yaml
var cliCasesYAML []byte

//go:embed testdata/cli_manifest.yaml
var cliManifestYAML []byte

func TestReleaseGuardCLI(t *testing.T) {
	if os.Getenv("RELEASE_GUARD_TEST_PROCESS") == "1" {
		for i, arg := range os.Args {
			if arg == "--" {
				os.Args = append([]string{"release-guard"}, os.Args[i+1:]...)
				main()
				os.Exit(0)
			}
		}
		os.Exit(2)
	}
	var fixture struct {
		Cases []struct {
			Name         string   `yaml:"name"`
			Args         []string `yaml:"args"`
			Stdout       string   `yaml:"stdout"`
			Stderr       string   `yaml:"stderr"`
			Exit         int      `yaml:"exit"`
			GitHubOutput string   `yaml:"github_output"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(cliCasesYAML, &fixture); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range fixture.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "release-guard CLI cases", cliManifestYAML, names)
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "github_output")
			if err := os.WriteFile(path, []byte("preexisting=1\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"-test.run=^TestReleaseGuardCLI$", "--"}, c.Args...)
			cmd := exec.Command(os.Args[0], args...)
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "GITHUB_OUTPUT=") && !strings.HasPrefix(entry, "RELEASE_GUARD_TEST_PROCESS=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "RELEASE_GUARD_TEST_PROCESS=1", "GITHUB_OUTPUT="+path)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			exit := 0
			if err != nil {
				failure, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				exit = failure.ExitCode()
			}
			if exit != c.Exit || stdout.String() != c.Stdout || stderr.String() != c.Stderr {
				t.Errorf("exit/stdout/stderr = %d/%q/%q, want %d/%q/%q", exit, stdout.String(), stderr.String(), c.Exit, c.Stdout, c.Stderr)
			}
			output, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(output) != c.GitHubOutput {
				t.Errorf("$GITHUB_OUTPUT = %q, want %q", output, c.GitHubOutput)
			}
		})
	}
}
