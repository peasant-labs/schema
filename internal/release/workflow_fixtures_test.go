package release

// Test-only fixture loader for the release-workflow policy corpus. The //go:embed
// lives in a _test.go file so the testdata/workflow tree compiles ONLY into the
// test binary, never into the production release-guard tool. The workflow YAML
// files, the peasant policy file, and the case table (cases.yaml) all live under
// testdata/workflow and are read from this single embedded FS.

import (
	"embed"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/workflow
var workflowFixtureFS embed.FS

// readWorkflowFixture returns the bytes of testdata/workflow/<name>.
func readWorkflowFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := workflowFixtureFS.ReadFile("testdata/workflow/" + name)
	if err != nil {
		t.Fatalf("read workflow fixture %q: %v", name, err)
	}
	return data
}

// checkWorkflowCase is one policy-vs-workflow validation case. Policy names a
// WorkflowPolicy Go value (resolved by policyByName); Workflow names an embedded
// testdata/workflow/*.yml file. An empty WantContains means the pair must be
// ACCEPTED (no error); a non-empty WantContains means the resulting error MUST
// contain every listed substring.
type checkWorkflowCase struct {
	Name         string   `yaml:"name"`
	Policy       string   `yaml:"policy"`
	Workflow     string   `yaml:"workflow"`
	WantContains []string `yaml:"wantContains"`
}

// checkWorkflowCases is the parsed testdata/workflow/cases.yaml corpus.
type checkWorkflowCases struct {
	Accept []checkWorkflowCase `yaml:"accept"`
	Reject []checkWorkflowCase `yaml:"reject"`
}

// loadCheckWorkflowCases parses the embedded check-workflow case table.
func loadCheckWorkflowCases(t *testing.T) checkWorkflowCases {
	t.Helper()
	data, err := workflowFixtureFS.ReadFile("testdata/workflow/cases.yaml")
	if err != nil {
		t.Fatalf("read check-workflow cases: %v", err)
	}
	var c checkWorkflowCases
	if err := yaml.Unmarshal(data, &c); err != nil {
		t.Fatalf("load check-workflow cases (testdata/workflow/cases.yaml): %v", err)
	}
	return c
}

// policyByName resolves a case's policy reference to its WorkflowPolicy Go value.
// The two shape policies are the canonical expected values also asserted by the
// LoadWorkflowPolicy round-trip; "empty" is the no-jobs policy.
func policyByName(t *testing.T, name string) WorkflowPolicy {
	t.Helper()
	switch name {
	case "schema":
		return schemaShapePolicy()
	case "peasant":
		return peasantShapePolicy()
	case "empty":
		return WorkflowPolicy{}
	default:
		t.Fatalf("unknown policy ref %q in check-workflow case fixture", name)
		return WorkflowPolicy{}
	}
}
