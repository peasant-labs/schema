package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestReleaseSetupGates(t *testing.T) {
	var fixture struct {
		Precheck struct {
			ID         string   `yaml:"id"`
			SetupIf    string   `yaml:"setup_if"`
			SetupSteps []string `yaml:"setup_steps"`
		} `yaml:"precheck"`
		CutIf   string   `yaml:"cut_default_branch_if"`
		CutJobs []string `yaml:"cut_jobs"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	readJobs := func(workflow string) *yaml.Node {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, ".github/workflows", workflow))
		if err != nil {
			t.Fatal(err)
		}
		var doc yaml.Node
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Fatal(err)
		}
		jobs := workflowMappingChild(doc.Content[0], "jobs")
		if jobs == nil {
			t.Fatalf("%s has no jobs", workflow)
		}
		return jobs
	}
	jobs := readJobs("release-batch.yml")
	steps := workflowMappingChild(workflowMappingChild(jobs, "plan"), "steps")
	if steps == nil {
		t.Fatal("release plan has no steps")
	}
	precheck := -1
	positions := make(map[string]int)
	for index, step := range steps.Content {
		if id := workflowMappingChild(step, "id"); id != nil && id.Value == fixture.Precheck.ID {
			precheck = index
		}
		if name := workflowMappingChild(step, "name"); name != nil {
			positions[name.Value] = index
		}
	}
	if precheck < 0 {
		t.Fatal("scheduled precheck is missing")
	}
	for _, name := range fixture.Precheck.SetupSteps {
		index, ok := positions[name]
		if !ok || index <= precheck {
			t.Fatalf("setup step %q does not follow the precheck", name)
		}
		condition := workflowMappingChild(steps.Content[index], "if")
		if condition == nil || condition.Value != fixture.Precheck.SetupIf {
			t.Errorf("setup step %q is not gated by the precheck", name)
		}
	}
	for i := 1; i < len(jobs.Content); i += 2 {
		steps := workflowMappingChild(jobs.Content[i], "steps")
		if steps == nil {
			continue
		}
		for _, step := range steps.Content {
			if run := workflowMappingChild(step, "run"); run != nil &&
				(strings.Contains(run.Value, "vars.RELEASE_BATCH_MODE") || strings.Contains(run.Value, "vars.RELEASE_BATCH_FLOOR")) {
				t.Error("release mode and floor must enter shell steps through env")
			}
		}
	}
	cutJobs := readJobs("cut-release.yml")
	for _, name := range fixture.CutJobs {
		condition := workflowMappingChild(workflowMappingChild(cutJobs, name), "if")
		if condition == nil || condition.Value != fixture.CutIf {
			t.Errorf("cut-release job %q is not restricted to the default branch and tag mode", name)
		}
	}
}
