package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type workflowPathsFixture struct {
	RequiredPaths []string `yaml:"required_paths"`
}

// TestTestsWorkflowTriggersOnTypeScriptPaths guards that
// the unfiltered workflow delegates the TypeScript path decision to its changes
// job. A future edit must not stop make check for a TypeScript-only change.
func TestTestsWorkflowTriggersOnTypeScriptPaths(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}

	fixturePath := filepath.Join(root, "testdata", "typescript", "workflow_paths.yaml")
	fixtureData, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture workflowPathsFixture
	fixtureDecoder := yaml.NewDecoder(bytes.NewReader(fixtureData))
	fixtureDecoder.KnownFields(true)
	if err := fixtureDecoder.Decode(&fixture); err != nil {
		t.Fatalf("decode %s: %v", fixturePath, err)
	}
	if len(fixture.RequiredPaths) == 0 {
		t.Fatalf("%s declares no required_paths", fixturePath)
	}

	workflowPath := filepath.Join(root, ".github", "workflows", "tests.yml")
	workflowData, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(workflowData, &doc); err != nil {
		t.Fatalf("parse %s as YAML: %v", workflowPath, err)
	}
	if len(doc.Content) == 0 {
		t.Fatalf("%s is empty", workflowPath)
	}

	// Read via the raw yaml.Node tree (not a decoded map/struct) so the
	// unquoted top-level "on:" key is matched by its literal source text
	// rather than resolved to the YAML 1.1 boolean `true`.
	onNode := workflowMappingChild(doc.Content[0], "on")
	if onNode == nil {
		t.Fatalf("%s has no top-level \"on:\" trigger mapping", workflowPath)
	}

	if workflowMappingChild(onNode, "merge_group") == nil || workflowMappingChild(onNode, "pull_request") == nil {
		t.Fatalf("%s must run for pull requests and merge groups", workflowPath)
	}
	jobs := workflowMappingChild(doc.Content[0], "jobs")
	changes := workflowMappingChild(jobs, "changes")
	steps := workflowMappingChild(changes, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		t.Fatal("changes job has no steps")
	}
	script := ""
	for _, step := range steps.Content {
		if id := workflowMappingChild(step, "id"); id != nil && id.Value == "diff" {
			if run := workflowMappingChild(step, "run"); run != nil {
				script = run.Value
			}
		}
	}
	for _, required := range fixture.RequiredPaths {
		if !strings.Contains(script, required) {
			t.Errorf("changes classifier is missing %q from %s", required, fixturePath)
		}
	}
}

func workflowMappingChild(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
