package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type releasePRWorkflowFixture struct {
	TriggerType      string `yaml:"trigger_type"`
	ConcurrencyGroup string `yaml:"concurrency_group"`
	CancelInProgress string `yaml:"cancel_in_progress"`
	ValidateIf       string `yaml:"validate_if"`
}

// TestReleasePRWorkflowContract pins release-pr.yml: the merge path runs in
// its own never-cancelled group, edits queue behind running gates, validate
// runs on open PRs only. Expectations are the named fields of the
// testdata/release-pr/release_pr_workflow.yaml fixture, matched by name.
func TestReleasePRWorkflowContract(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	checkReleasePRWorkflowContract(t, filepath.Join(root, ".github", "workflows", "release-pr.yml"), filepath.Join(root, "testdata", "release-pr", "release_pr_workflow.yaml"))
}

func checkReleasePRWorkflowContract(t *testing.T, workflowPath, fixturePath string) {
	t.Helper()
	fixtureData, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture releasePRWorkflowFixture
	decoder := yaml.NewDecoder(bytes.NewReader(fixtureData))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode %s: %v", fixturePath, err)
	}
	for name, value := range map[string]string{"trigger_type": fixture.TriggerType, "concurrency_group": fixture.ConcurrencyGroup, "cancel_in_progress": fixture.CancelInProgress, "validate_if": fixture.ValidateIf} {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("%s declares no %s; every contract field is required by name", fixturePath, name)
		}
	}
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
	// Raw yaml.Node tree so the unquoted top-level "on:" key matches by its
	// literal source text (see workflow_paths_test.go).
	child := func(node *yaml.Node, keys ...string) *yaml.Node {
		for _, key := range keys {
			if node = workflowMappingChild(node, key); node == nil {
				return nil
			}
		}
		return node
	}
	types := child(doc.Content[0], "on", "pull_request", "types")
	if types == nil || types.Kind != yaml.SequenceNode {
		t.Fatalf("%s on.pull_request.types is missing; restore the trigger types (with edited) so the title is still revalidated on edit", workflowPath)
	}
	var haveTrigger bool
	for _, item := range types.Content {
		haveTrigger = haveTrigger || item.Value == fixture.TriggerType
	}
	if !haveTrigger {
		t.Errorf("%s on.pull_request.types is missing %q (required by %s); add it back so the title is still revalidated on edit", workflowPath, fixture.TriggerType, fixturePath)
	}
	for name, want := range map[string]string{"group": fixture.ConcurrencyGroup, "cancel-in-progress": fixture.CancelInProgress} {
		node := child(doc.Content[0], "concurrency", name)
		if node == nil {
			t.Errorf("%s concurrency.%s is missing, want %q", workflowPath, name, want)
		} else if got := normWorkflowExpr(node.Value); got != normWorkflowExpr(want) {
			t.Errorf("%s concurrency.%s = %q, want %q", workflowPath, name, node.Value, want)
		}
	}
	node := child(doc.Content[0], "jobs", "validate", "if")
	if node == nil {
		t.Fatalf("%s jobs.validate.if is missing (required by %s)", workflowPath, fixturePath)
	} else if got := normWorkflowExpr(node.Value); got != normWorkflowExpr(fixture.ValidateIf) {
		t.Errorf("%s jobs.validate.if = %q, want %q", workflowPath, node.Value, fixture.ValidateIf)
	}
}

func normWorkflowExpr(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
