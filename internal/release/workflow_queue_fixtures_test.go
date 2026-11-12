package release

// Test-only fixture loader for the merge-queue workflow policy corpus. The
// //go:embed lives in a _test.go file so the testdata/workflow_queue tree
// compiles ONLY into the test binary, never into production code. Each case
// directory holds the policy.yml and workflows the production CheckWorkflows
// reads from a real directory on disk.

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/workflow_queue
var workflowQueueFixtureFS embed.FS

// queueCaseInput names the case directory under testdata/workflow_queue.
type queueCaseInput struct {
	Dir string `yaml:"dir"`
}

// queueCaseExpected states the rule a reject case must trip and the message
// fragments its diagnostics must carry. An accept case leaves both empty.
type queueCaseExpected struct {
	Rule         string   `yaml:"rule"`
	WantContains []string `yaml:"want_contains"`
}

// queueWorkflowFixtures is the parsed testdata/workflow_queue/cases.yaml
// corpus with its required-name manifest. The case rows decode as a plain
// slice (not the Corpus document shape) and are wrapped for validation.
type queueWorkflowFixtures struct {
	RequiredNames []string                                           `yaml:"required_names"`
	Rows          []testcase.Case[queueCaseInput, queueCaseExpected] `yaml:"cases"`
}

// corpus wraps the rows in the canonical Corpus for validation.
func (f queueWorkflowFixtures) corpus() testcase.Corpus[queueCaseInput, queueCaseExpected] {
	return testcase.Corpus[queueCaseInput, queueCaseExpected]{Cases: f.Rows}
}

// loadQueueWorkflowFixtures parses and boundary-validates the corpus: strict
// fields, non-vacuous cases, and an exact match between the required-name
// manifest and the case names, so a dropped or stray sample fails loudly.
func loadQueueWorkflowFixtures(t *testing.T) queueWorkflowFixtures {
	t.Helper()
	data, err := workflowQueueFixtureFS.ReadFile("testdata/workflow_queue/cases.yaml")
	if err != nil {
		t.Fatalf("read queue workflow cases: %v", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f queueWorkflowFixtures
	if err := dec.Decode(&f); err != nil {
		t.Fatalf("load queue workflow cases (testdata/workflow_queue/cases.yaml): %v", err)
	}
	corpus := f.corpus()
	if err := corpus.Validate(); err != nil {
		t.Fatalf("queue workflow corpus is under-populated: %v", err)
	}
	seen := make(map[string]bool, len(corpus.Cases))
	for _, c := range corpus.Cases {
		seen[c.Name] = true
	}
	for _, name := range f.RequiredNames {
		if !seen[name] {
			t.Fatalf("required queue workflow case %q is missing from the corpus", name)
		}
	}
	if len(seen) != len(f.RequiredNames) {
		t.Fatalf("required_names names %d cases but the corpus carries %d", len(f.RequiredNames), len(seen))
	}
	return f
}

// copyQueueFixture materializes one embedded case directory onto disk and
// returns its path, so the test drives the production directory reader.
func copyQueueFixture(t *testing.T, dir string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), dir)
	err := fs.WalkDir(workflowQueueFixtureFS, "testdata/workflow_queue/"+dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("testdata/workflow_queue/"+dir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := workflowQueueFixtureFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("materialize queue fixture %q: %v", dir, err)
	}
	return dst
}
