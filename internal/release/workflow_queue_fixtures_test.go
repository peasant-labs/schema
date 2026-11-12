package release

// Test-only fixture loader for the merge-queue workflow policy corpus. The
// //go:embed lives in a _test.go file so the testdata/workflow_queue tree
// compiles ONLY into the test binary, never into production code. Each case
// overlays changed files onto the shared base tree; the harness materializes
// the result into a temp dir and runs the production CheckWorkflows over it.

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/workflow_queue
var workflowQueueFixtureFS embed.FS

// queueCaseInput names the case overlay directory under
// testdata/workflow_queue/overlays. Empty means the base tree as-is.
type queueCaseInput struct {
	Overlay string `yaml:"overlay"`
}

// queueCaseExpected states the rule a reject case must trip (and only that
// rule) with the message fragments its diagnostics must carry. An accept case
// carries its sampled rule and no fragments; the run must yield no diagnostics
// at all.
type queueCaseExpected struct {
	Rule         string   `yaml:"rule"`
	WantContains []string `yaml:"want_contains"`
}

// queueCaseRow is one corpus row: the standard case fields plus the accept
// sibling a must-fail row proves itself against.
type queueCaseRow struct {
	Name           string                  `yaml:"name"`
	Input          queueCaseInput          `yaml:"input"`
	Expected       queueCaseExpected       `yaml:"expected"`
	Classification testcase.Classification `yaml:"classification"`
	Provenance     testcase.Provenance     `yaml:"provenance"`
	Mutation       testcase.Mutation       `yaml:"mutation"`
	Sibling        string                  `yaml:"sibling"`
}

// queueWorkflowFixtures is the parsed testdata/workflow_queue/cases.yaml
// corpus with its required-name manifest.
type queueWorkflowFixtures struct {
	RequiredNames []string       `yaml:"required_names"`
	Rows          []queueCaseRow `yaml:"cases"`
}

// corpus wraps the rows in the canonical Corpus for non-vacuity validation.
func (f queueWorkflowFixtures) corpus() testcase.Corpus[queueCaseInput, queueCaseExpected] {
	cases := make([]testcase.Case[queueCaseInput, queueCaseExpected], 0, len(f.Rows))
	for _, row := range f.Rows {
		cases = append(cases, testcase.Case[queueCaseInput, queueCaseExpected]{
			Name:           row.Name,
			Input:          row.Input,
			Expected:       row.Expected,
			Classification: row.Classification,
			Provenance:     row.Provenance,
			Mutation:       row.Mutation,
		})
	}
	return testcase.Corpus[queueCaseInput, queueCaseExpected]{Cases: cases}
}

// loadQueueWorkflowFixtures parses and boundary-validates the corpus: strict
// fields, non-vacuous cases, an exact match between the required-name manifest
// and the case names, and an existing must-pass sibling for every must-fail
// row.
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
	if err := f.corpus().Validate(); err != nil {
		t.Fatalf("queue workflow corpus is under-populated: %v", err)
	}
	byName := make(map[string]queueCaseRow, len(f.Rows))
	for _, row := range f.Rows {
		if _, dup := byName[row.Name]; dup {
			t.Fatalf("queue workflow corpus names case %q twice", row.Name)
		}
		if row.Classification == testcase.MustFail {
			if row.Sibling == "" {
				t.Fatalf("reject case %q names no accept sibling", row.Name)
			}
		} else if row.Sibling != "" {
			t.Fatalf("accept case %q must not name a sibling", row.Name)
		}
		if row.Expected.Rule == "" || !QueueRule(row.Expected.Rule).IsValid() {
			t.Fatalf("case %q names unknown rule %q (must be a closed QueueRule member)", row.Name, row.Expected.Rule)
		}
		byName[row.Name] = row
	}
	for _, row := range f.Rows {
		if row.Classification != testcase.MustFail {
			continue
		}
		sib, ok := byName[row.Sibling]
		if !ok {
			t.Fatalf("reject case %q names missing sibling %q", row.Name, row.Sibling)
		}
		if sib.Classification != testcase.MustPass {
			t.Fatalf("reject case %q names sibling %q which is not must-pass", row.Name, row.Sibling)
		}
	}
	seen := make(map[string]bool, len(f.Rows))
	for name := range byName {
		seen[name] = true
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

// materializeQueueFixture overlays the case files onto the shared base tree in
// a temp dir and returns its path, so the test drives the production
// directory reader over the effective tree.
func materializeQueueFixture(t *testing.T, overlay string) string {
	t.Helper()
	dst := t.TempDir()
	copyEmbeddedDir(t, "testdata/workflow_queue/base", dst)
	if overlay != "" {
		copyEmbeddedDir(t, "testdata/workflow_queue/overlays/"+overlay, dst)
	}
	return dst
}

// copyEmbeddedDir copies one embedded tree over dst, creating directories.
func copyEmbeddedDir(t *testing.T, src, dst string) {
	t.Helper()
	err := fs.WalkDir(workflowQueueFixtureFS, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
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
		t.Fatalf("materialize queue fixture %q: %v", src, err)
	}
}

// loadBaseConfig parses the shared base release config, the one source of
// the bot branches the checker resolves.
func loadBaseConfig(t *testing.T) intent.Config {
	t.Helper()
	data, err := workflowQueueFixtureFS.ReadFile("testdata/workflow_queue/base/release-intent.yaml")
	if err != nil {
		t.Fatalf("read base release config: %v", err)
	}
	cfg, err := intent.ParseConfig(data)
	if err != nil {
		t.Fatalf("parse base release config: %v", err)
	}
	return cfg
}
