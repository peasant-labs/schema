package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// orDefault keeps a fixture row short: an omitted field is the ordinary
// merge-group value.
func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// TestMergeGateVendorHash runs the actual merge-gate shell with fixture needs.
func TestMergeGateVendorHash(t *testing.T) {
	var fixture struct {
		Required []string `yaml:"merge_gate_required"`
		Cases    []struct {
			Name             string  `yaml:"name"`
			Release          string  `yaml:"release"`
			Event            string  `yaml:"event"`
			Code             string  `yaml:"code"`
			ReleaseCandidate *string `yaml:"release_candidate"`
			ReleaseOnly      string  `yaml:"release_only"`
			Inherit          string  `yaml:"inherit"`
			ResultsJSON      string  `yaml:"results_json"`
			ExpectedExit     int     `yaml:"expected_exit"`
			ExpectedStderr   string  `yaml:"expected_stderr"`
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
			candidate := "false"
			if c.ReleaseCandidate != nil {
				candidate = *c.ReleaseCandidate
			}
			cmd.Env = append(os.Environ(), "RESULTS="+c.ResultsJSON, "RELEASE="+c.Release,
				"EVENT="+orDefault(c.Event, "merge_group"), "CODE="+orDefault(c.Code, "true"),
				"RELEASE_CANDIDATE="+candidate,
				"RELEASE_ONLY="+orDefault(c.ReleaseOnly, "false"), "INHERIT="+orDefault(c.Inherit, "false"))
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
			if c.ExpectedExit != 0 && (c.ExpectedStderr == "" || !strings.Contains(string(out), "::error::merge-gate: "+c.ExpectedStderr)) {
				t.Fatalf("merge-gate output lacks the expected refusal %q (a reject row must name its message):\n%s", c.ExpectedStderr, out)
			}
		})
	}
	for _, name := range fixture.Required {
		if !seen[name] {
			t.Errorf("required merge-gate case %q missing", name)
		}
	}
}
