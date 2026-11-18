package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/release_adoption.yaml
var releaseAdoptionFixture []byte

func TestReleaseAdoptionWorkflowWiring(t *testing.T) {
	var fixture struct {
		ToolPin  string   `yaml:"tool_pin"`
		Required []string `yaml:"required"`
		Cases    []struct {
			Name             string            `yaml:"name"`
			Workflow         string            `yaml:"workflow"`
			Job              string            `yaml:"job"`
			Step             string            `yaml:"step"`
			ID               string            `yaml:"id"`
			Uses             string            `yaml:"uses"`
			After            string            `yaml:"after"`
			RunContains      []string          `yaml:"run_contains"`
			RunExcludes      []string          `yaml:"run_excludes"`
			JobRunExcludes   []string          `yaml:"job_run_excludes"`
			JobRunOnly       map[string]string `yaml:"job_run_only"`
			ContinueContains []string          `yaml:"continue_contains"`
			NoContinue       bool              `yaml:"no_continue"`
			NoJobContinue    bool              `yaml:"no_job_continue"`
			IfContains       []string          `yaml:"if_contains"`
			IfValue          string            `yaml:"if_value"`
			EnvValues        map[string]string `yaml:"env_values"`
			WithValues       map[string]string `yaml:"with_values"`
			WithExcludes     []string          `yaml:"with_excludes"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", c.Workflow))
			if err != nil {
				t.Fatal(err)
			}
			var doc yaml.Node
			if err := yaml.Unmarshal(data, &doc); err != nil {
				t.Fatal(err)
			}
			job := workflowMappingChild(workflowMappingChild(doc.Content[0], "jobs"), c.Job)
			if c.NoJobContinue && workflowMappingChild(job, "continue-on-error") != nil {
				t.Error("job must not set continue-on-error")
			}
			steps := workflowMappingChild(job, "steps")
			if steps == nil {
				t.Fatalf("job %s has no steps", c.Job)
			}
			for _, step := range steps.Content {
				if run := workflowMappingChild(step, "run"); run != nil {
					for _, excluded := range c.JobRunExcludes {
						if strings.Contains(run.Value, excluded) {
							t.Errorf("job %s unexpectedly runs %q", c.Job, excluded)
						}
					}
					for fragment, allowedStep := range c.JobRunOnly {
						if name := workflowMappingChild(step, "name"); strings.Contains(run.Value, fragment) && (name == nil || name.Value != allowedStep) {
							t.Errorf("job %s permits %q only in step %q", c.Job, fragment, allowedStep)
						}
					}
				}
			}
			for index, step := range steps.Content {
				name := workflowMappingChild(step, "name")
				if name == nil || name.Value != c.Step {
					continue
				}
				if c.After != "" {
					foundEarlier := false
					for _, earlier := range steps.Content[:index] {
						if name := workflowMappingChild(earlier, "name"); name != nil && name.Value == c.After {
							foundEarlier = true
							break
						}
					}
					if !foundEarlier {
						t.Errorf("step %q must follow %q", c.Step, c.After)
					}
				}
				if c.ID != "" {
					id := workflowMappingChild(step, "id")
					if id == nil || id.Value != c.ID {
						t.Errorf("step id is not %q", c.ID)
					}
				}
				if c.Uses != "" {
					uses := workflowMappingChild(step, "uses")
					if uses == nil || uses.Value != c.Uses {
						t.Errorf("step uses is not %q", c.Uses)
					}
				}
				checkWorkflowFragments(t, step, "run", c.RunContains)
				if run := workflowMappingChild(step, "run"); run != nil {
					for _, fragment := range c.RunExcludes {
						if strings.Contains(run.Value, fragment) {
							t.Errorf("run unexpectedly contains %q", fragment)
						}
					}
				}
				checkWorkflowFragments(t, step, "continue-on-error", c.ContinueContains)
				if c.NoContinue && workflowMappingChild(step, "continue-on-error") != nil {
					t.Error("step must not set continue-on-error")
				}
				checkWorkflowFragments(t, step, "if", c.IfContains)
				if c.IfValue != "" {
					condition := workflowMappingChild(step, "if")
					if condition == nil || condition.Value != c.IfValue {
						t.Errorf("step if must equal %q", c.IfValue)
					}
				}
				for key, expected := range c.EnvValues {
					value := workflowMappingChild(workflowMappingChild(step, "env"), key)
					if value == nil || value.Value != expected {
						t.Errorf("step env %s must equal %q", key, expected)
					}
				}
				for key, expected := range c.WithValues {
					value := workflowMappingChild(workflowMappingChild(step, "with"), key)
					if value == nil || value.Value != expected {
						t.Errorf("step input %s must equal %q", key, expected)
					}
				}
				for _, key := range c.WithExcludes {
					if workflowMappingChild(workflowMappingChild(step, "with"), key) != nil {
						t.Errorf("step must not set input %s", key)
					}
				}
				return
			}
			t.Fatalf("step %q not found in %s", c.Step, c.Workflow)
		})
	}
	for _, required := range fixture.Required {
		if !seen[required] {
			t.Errorf("required workflow case %q missing", required)
		}
	}

	paths, err := filepath.Glob(filepath.Join(root, ".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var doc yaml.Node
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Fatal(err)
		}
		jobs := workflowMappingChild(doc.Content[0], "jobs")
		if jobs == nil {
			continue
		}
		for i := 1; i < len(jobs.Content); i += 2 {
			steps := workflowMappingChild(jobs.Content[i], "steps")
			if steps == nil {
				continue
			}
			for _, step := range steps.Content {
				with := workflowMappingChild(step, "with")
				repo := workflowMappingChild(with, "repository")
				if repo == nil || repo.Value != "peasant-labs/infra" {
					continue
				}
				found = true
				ref := workflowMappingChild(with, "ref")
				if ref == nil || ref.Value != fixture.ToolPin {
					t.Errorf("%s: release tool checkout is not pinned to %s", path, fixture.ToolPin)
				}
			}
		}
	}
	if !found {
		t.Fatal("no release-tool checkout found")
	}
}

func checkWorkflowFragments(t *testing.T, step *yaml.Node, key string, fragments []string) {
	t.Helper()
	if len(fragments) == 0 {
		return
	}
	value := workflowMappingChild(step, key)
	if value == nil {
		t.Fatalf("step has no %s", key)
	}
	for _, fragment := range fragments {
		if !strings.Contains(value.Value, fragment) {
			t.Errorf("%s does not contain %q", key, fragment)
		}
	}
}

// TestReleaseTagVersionGuard runs the jq-based workflow step. A matching
// manifest must pass and a mismatched one must fail.
func TestReleaseTagVersionGuard(t *testing.T) {
	var fixture struct {
		Required []string `yaml:"version_guard_required"`
		Cases    []struct {
			Name           string `yaml:"name"`
			Tag            string `yaml:"tag"`
			PackageVersion string `yaml:"package_version"`
			ExpectedExit   int    `yaml:"expected_exit"`
			ExpectedOutput string `yaml:"expected_output"`
		} `yaml:"version_guard_cases"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
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
	job := workflowMappingChild(workflowMappingChild(doc.Content[0], "jobs"), "npm-publish")
	steps := workflowMappingChild(job, "steps")
	if steps == nil {
		t.Fatal("npm-publish has no steps")
	}
	var run string
	for _, step := range steps.Content {
		if name := workflowMappingChild(step, "name"); name != nil && name.Value == "Guard tag and package version consistency" {
			if script := workflowMappingChild(step, "run"); script != nil {
				run = script.Value
			}
		}
	}
	if run == "" {
		t.Fatal("npm tag version guard has no script")
	}
	if !strings.Contains(run, "jq -r '.version' typescript/package.json") {
		t.Fatal("npm tag version guard must read the manifest with jq")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "typescript"), 0o755); err != nil {
				t.Fatal(err)
			}
			manifest, err := json.Marshal(map[string]string{"version": c.PackageVersion})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "typescript", "package.json"), manifest, 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", "-e", "-c", run)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "TAG="+c.Tag)
			output, err := cmd.CombinedOutput()
			exit := 0
			if err != nil {
				if failure, ok := err.(*exec.ExitError); ok {
					exit = failure.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if exit != c.ExpectedExit || !strings.Contains(string(output), c.ExpectedOutput) {
				t.Fatalf("guard exit %d, want %d with %q; output:\n%s", exit, c.ExpectedExit, c.ExpectedOutput, output)
			}
		})
	}
	for _, required := range fixture.Required {
		if !seen[required] {
			t.Errorf("required version guard case %q missing", required)
		}
	}
}
