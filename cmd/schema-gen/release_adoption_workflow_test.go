package main

import (
	_ "embed"
	"os"
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
			RunContains      []string          `yaml:"run_contains"`
			RunExcludes      []string          `yaml:"run_excludes"`
			ContinueContains []string          `yaml:"continue_contains"`
			IfContains       []string          `yaml:"if_contains"`
			IfValue          string            `yaml:"if_value"`
			EnvValues        map[string]string `yaml:"env_values"`
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
			steps := workflowMappingChild(job, "steps")
			if steps == nil {
				t.Fatalf("job %s has no steps", c.Job)
			}
			for _, step := range steps.Content {
				name := workflowMappingChild(step, "name")
				if name == nil || name.Value != c.Step {
					continue
				}
				if c.ID != "" {
					id := workflowMappingChild(step, "id")
					if id == nil || id.Value != c.ID {
						t.Errorf("step id is not %q", c.ID)
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

func TestReleaseTitleComesFromPreparedCommit(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release-batch.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	jobs := workflowMappingChild(doc.Content[0], "jobs")
	outputs := workflowMappingChild(workflowMappingChild(jobs, "plan"), "outputs")
	title := workflowMappingChild(outputs, "title")
	if title == nil || title.Value != "${{ steps.write.outputs.title }}" {
		t.Fatal("the plan job must expose the title from release-pr --commit")
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
