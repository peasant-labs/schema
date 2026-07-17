package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type workflowPathsFixture struct {
	RequiredPaths []string `yaml:"required_paths"`
}

// TestTestsWorkflowTriggersOnTypeScriptPaths guards that
// .github/workflows/tests.yml's pull_request and push path filters cover every
// path testdata/typescript/workflow_paths.yaml names as required. A future
// workflow edit that narrows those filters would silently stop CI from
// re-running `make check` (and so the TypeScript typecheck/test/package gates)
// on a change that only touches typescript/**, testcase/testdata/**, or
// testdata/**.
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

	for _, trigger := range []string{"pull_request", "push"} {
		triggerNode := workflowMappingChild(onNode, trigger)
		if triggerNode == nil {
			t.Errorf("%s on.%s is missing; the TypeScript-dependent required_paths in %s cannot be enforced for this trigger", workflowPath, trigger, fixturePath)
			continue
		}
		pathsNode := workflowMappingChild(triggerNode, "paths")
		if pathsNode == nil || pathsNode.Kind != yaml.SequenceNode {
			t.Errorf("%s on.%s.paths is missing or not a sequence", workflowPath, trigger)
			continue
		}
		have := make(map[string]bool, len(pathsNode.Content))
		for _, item := range pathsNode.Content {
			have[item.Value] = true
		}
		for _, required := range fixture.RequiredPaths {
			if !have[required] {
				t.Errorf("%s on.%s.paths is missing %q (required by %s); CI would stop re-running make check on a change scoped to this path, add it back to the trigger filter", workflowPath, trigger, required, fixturePath)
			}
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
