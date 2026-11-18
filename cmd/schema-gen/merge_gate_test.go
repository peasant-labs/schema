package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestMergeGateVendorHash runs the actual merge-gate shell with fixture needs.
func TestMergeGateVendorHash(t *testing.T) {
	var fixture struct {
		Required []string `yaml:"merge_gate_required"`
		Cases    []struct {
			Name         string `yaml:"name"`
			Release      string `yaml:"release"`
			ResultsJSON  string `yaml:"results_json"`
			ExpectedExit int    `yaml:"expected_exit"`
		} `yaml:"merge_gate_cases"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/workflows/tests.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	job := workflowMappingChild(workflowMappingChild(doc.Content[0], "jobs"), "merge-gate")
	steps := workflowMappingChild(job, "steps")
	if steps == nil {
		t.Fatal("merge-gate has no steps")
	}
	var script string
	for _, step := range steps.Content {
		if name := workflowMappingChild(step, "name"); name != nil && name.Value == "Require successful or intentionally skipped jobs" {
			if run := workflowMappingChild(step, "run"); run != nil {
				script = run.Value
			}
		}
	}
	if script == "" {
		t.Fatal("merge-gate shell step missing")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate merge-gate case %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			cmd := exec.Command("bash", "-c", script)
			cmd.Env = append(os.Environ(), "RESULTS="+c.ResultsJSON, "EVENT=merge_group", "CODE=true", "RELEASE="+c.Release)
			out, err := cmd.CombinedOutput()
			exit := 0
			if err != nil {
				if failure, ok := err.(*exec.ExitError); ok {
					exit = failure.ExitCode()
				} else {
					t.Fatalf("execute merge-gate: %v\n%s", err, out)
				}
			}
			if exit != c.ExpectedExit {
				t.Fatalf("merge-gate exit = %d, want %d:\n%s", exit, c.ExpectedExit, out)
			}
		})
	}
	for _, name := range fixture.Required {
		if !seen[name] {
			t.Errorf("required merge-gate case %q missing", name)
		}
	}
}
