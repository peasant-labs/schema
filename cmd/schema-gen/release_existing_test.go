package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type releaseViewCase struct {
	Name           string `yaml:"name"`
	Response       string `yaml:"response"`
	APIExit        int    `yaml:"api_exit"`
	ExpectedExit   int    `yaml:"expected_exit"`
	ExpectedOutput string `yaml:"expected_output"`
}

// TestReleaseViewChecks executes each release-state step of the publish job
// and its actual jq predicate, with the GitHub CLI's response isolated by a
// fixture: the rerun completeness check, the draft check before publishing,
// and the published check after it.
func TestReleaseViewChecks(t *testing.T) {
	var fixture struct {
		ExistingRequired  []string          `yaml:"existing_release_required"`
		ExistingCases     []releaseViewCase `yaml:"existing_release_cases"`
		DraftRequired     []string          `yaml:"draft_release_required"`
		DraftCases        []releaseViewCase `yaml:"draft_release_cases"`
		PublishedRequired []string          `yaml:"published_release_required"`
		PublishedCases    []releaseViewCase `yaml:"published_release_cases"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		step, fields string
		required     []string
		cases        []releaseViewCase
	}{
		{"Verify the existing release is complete", "isDraft,assets", fixture.ExistingRequired, fixture.ExistingCases},
		{"Verify the draft before publishing", "isDraft,assets", fixture.DraftRequired, fixture.DraftCases},
		{"Verify the published release", "isDraft,isImmutable,assets", fixture.PublishedRequired, fixture.PublishedCases},
	} {
		t.Run(check.step, func(t *testing.T) {
			runReleaseViewStep(t, check.step, check.fields, check.required, check.cases)
		})
	}
}

func runReleaseViewStep(t *testing.T, stepName, fields string, required []string, cases []releaseViewCase) {
	t.Helper()
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	job := workflowMappingChild(workflowMappingChild(doc.Content[0], "jobs"), "release")
	steps := workflowMappingChild(job, "steps")
	if steps == nil {
		t.Fatal("release job has no steps")
	}
	var script string
	for _, step := range steps.Content {
		if name := workflowMappingChild(step, "name"); name != nil && name.Value == stepName {
			if run := workflowMappingChild(step, "run"); run != nil {
				script = run.Value
			}
		}
	}
	if script == "" {
		t.Fatalf("step %q has no script", stepName)
	}
	seen := make(map[string]bool)
	for _, c := range cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate release view case %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			gh := `#!/bin/sh
[ "$#" -eq 9 ] && [ "$1" = release ] && [ "$2" = view ] && [ "$3" = "$TAG" ] && [ "$4" = --repo ] && [ "$5" = "$GITHUB_REPOSITORY" ] && [ "$6" = --json ] && [ "$7" = "$TEST_FIELDS" ] && [ "$8" = --jq ] || exit 92
if [ "$TEST_API_EXIT" -ne 0 ]; then
  printf 'mock API unavailable\n' >&2
  exit "$TEST_API_EXIT"
fi
printf '%s\n' "$TEST_RELEASE_JSON" | jq -e -r "$9" || exit 1
`
			if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(gh), 0o755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", "-e", "-c", script)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"TAG=v1.2.3", "GITHUB_REPOSITORY=peasant-labs/schema", "TEST_FIELDS="+fields,
				"TEST_RELEASE_JSON="+c.Response, "TEST_API_EXIT="+strconv.Itoa(c.APIExit))
			output, err := cmd.CombinedOutput()
			exit := 0
			if err != nil {
				if failure, ok := err.(*exec.ExitError); ok {
					exit = failure.ExitCode()
				} else {
					t.Fatalf("execute step: %v\n%s", err, output)
				}
			}
			if exit != c.ExpectedExit || !strings.Contains(string(output), c.ExpectedOutput) {
				t.Fatalf("step exit %d, want %d with %q:\n%s", exit, c.ExpectedExit, c.ExpectedOutput, output)
			}
		})
	}
	for _, name := range required {
		if !seen[name] {
			t.Errorf("required release view case %q missing", name)
		}
	}
}
