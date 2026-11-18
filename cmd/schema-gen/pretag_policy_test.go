package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestPreTagPolicyResolvesChecks mirrors the policy checker's release-job view:
// steps in its directly needed jobs, then steps in the credentialed job itself.
func TestPreTagPolicyResolvesChecks(t *testing.T) {
	var fixture struct {
		Commands map[string]string `yaml:"pre_tag_commands"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/release-queue.policy.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Queue struct {
			Release struct {
				Workflow string `yaml:"workflow"`
				Job      string `yaml:"job"`
				Checks   []struct {
					ID    string `yaml:"id"`
					Check string `yaml:"check"`
				} `yaml:"pre_tag_steps"`
			} `yaml:"release_job"`
		} `yaml:"queue"`
	}
	if err := yaml.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	release := policy.Queue.Release
	if release.Workflow == "" || release.Job == "" || len(release.Checks) == 0 {
		t.Fatal("release policy must name its workflow, job, and pre-tag checks")
	}
	data, err = os.ReadFile(filepath.Join(root, ".github/workflows", release.Workflow))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	jobs := workflowMappingChild(doc.Content[0], "jobs")
	job := workflowMappingChild(jobs, release.Job)
	needsNode := workflowMappingChild(job, "needs")
	if job == nil || needsNode == nil {
		t.Fatalf("release job %q or its dependencies are missing", release.Job)
	}
	var needs []string
	if err := needsNode.Decode(&needs); err != nil {
		t.Fatal(err)
	}
	needs = append(needs, release.Job)
	stepsByID := make(map[string][]*yaml.Node)
	for _, name := range needs {
		steps := workflowMappingChild(workflowMappingChild(jobs, name), "steps")
		if steps == nil {
			t.Fatalf("release dependency %q has no steps", name)
		}
		for _, step := range steps.Content {
			if id := workflowMappingChild(step, "id"); id != nil {
				stepsByID[id.Value] = append(stepsByID[id.Value], step)
			}
		}
	}
	for _, check := range release.Checks {
		expected, ok := fixture.Commands[check.Check]
		if !ok || expected == "" {
			t.Fatalf("policy check %q has no expected command", check.Check)
		}
		steps := stepsByID[check.ID]
		if len(steps) != 1 {
			t.Fatalf("pre-tag step %q occurs %d times in the release job sequence", check.ID, len(steps))
		}
		run := workflowMappingChild(steps[0], "run")
		if run == nil || strings.TrimSpace(run.Value) != expected {
			t.Errorf("pre-tag step %q must run exactly %q", check.ID, expected)
		}
	}
}
