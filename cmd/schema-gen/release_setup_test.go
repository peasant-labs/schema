package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The provenance shell is tested separately. These job and step wiring checks
// establish that its failure fails the plan and prevents the release job.
func TestReleaseSetupGates(t *testing.T) {
	var fixture struct {
		Precheck struct {
			ID         string   `yaml:"id"`
			SetupIf    string   `yaml:"setup_if"`
			SetupSteps []string `yaml:"setup_steps"`
		} `yaml:"precheck"`
		Provenance struct {
			CandidateOutput string              `yaml:"candidate_output"`
			JobConditions   map[string]string   `yaml:"job_conditions"`
			JobNeeds        map[string][]string `yaml:"job_needs"`
			PullRequests    string              `yaml:"pull_requests"`
		} `yaml:"provenance_gates"`
		PublicationRequired []string            `yaml:"publication_required"`
		PublicationNeeds    map[string][]string `yaml:"publication_needs"`
		CutIf               string              `yaml:"cut_default_branch_if"`
		CutJobs             []string            `yaml:"cut_jobs"`
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
	plan := workflowMappingChild(jobs, "plan")
	if workflowMappingChild(plan, "continue-on-error") != nil {
		t.Error("release plan must not set continue-on-error")
	}
	permissions := workflowMappingChild(plan, "permissions")
	if value := workflowMappingChild(permissions, "pull-requests"); value == nil || fixture.Provenance.PullRequests == "" || value.Value != fixture.Provenance.PullRequests {
		t.Error("release plan must have pull-requests: read")
	}
	if value := workflowMappingChild(workflowMappingChild(plan, "outputs"), "candidate"); value == nil || value.Value != fixture.Provenance.CandidateOutput {
		t.Errorf("release plan candidate output must equal %q", fixture.Provenance.CandidateOutput)
	}
	if _, ok := fixture.Provenance.JobConditions["verify"]; !ok {
		t.Error("provenance fixture must require a verify job condition")
	}
	if _, ok := fixture.Provenance.JobConditions["release"]; !ok {
		t.Error("provenance fixture must require a release job condition")
	}
	for job, expected := range fixture.Provenance.JobConditions {
		if value := workflowMappingChild(workflowMappingChild(jobs, job), "if"); value == nil || expected == "" || value.Value != expected {
			t.Errorf("%s job condition must equal %q", job, expected)
		}
	}
	if _, ok := fixture.Provenance.JobNeeds["verify"]; !ok {
		t.Error("provenance fixture must require verify job dependencies")
	}
	if _, ok := fixture.Provenance.JobNeeds["release"]; !ok {
		t.Error("provenance fixture must require release job dependencies")
	}
	for job, expected := range fixture.Provenance.JobNeeds {
		needs := workflowMappingChild(workflowMappingChild(jobs, job), "needs")
		var actual []string
		if needs == nil || needs.Decode(&actual) != nil {
			t.Errorf("%s job must declare its dependencies", job)
			continue
		}
		want := slices.Clone(expected)
		slices.Sort(want)
		slices.Sort(actual)
		if !slices.Equal(actual, want) {
			t.Errorf("%s job dependencies = %v, want %v", job, actual, want)
		}
	}
	steps := workflowMappingChild(plan, "steps")
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
	publishJobs := readJobs("release.yml")
	for _, name := range fixture.PublicationRequired {
		expected, ok := fixture.PublicationNeeds[name]
		if !ok {
			t.Errorf("publication fixture must require %s job dependencies", name)
			continue
		}
		needs := workflowMappingChild(workflowMappingChild(publishJobs, name), "needs")
		var actual []string
		if needs == nil || needs.Decode(&actual) != nil {
			t.Errorf("%s publication job must declare its dependencies", name)
			continue
		}
		want := slices.Clone(expected)
		slices.Sort(want)
		slices.Sort(actual)
		if !slices.Equal(actual, want) {
			t.Errorf("%s publication job dependencies = %v, want %v", name, actual, want)
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
