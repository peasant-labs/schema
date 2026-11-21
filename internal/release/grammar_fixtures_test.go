package release_test

// Test-only fixture loader for the release-grammar corpus. The //go:embed lives
// in a _test.go file so testdata/grammar/versions.yaml compiles ONLY into the
// test binary, never into the production release-guard tool.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/grammar/versions.yaml
var grammarFixtureYAML []byte

// newVersionCase is one NewVersion(raw) row: raw input, expected Version, and the
// must-fail flag (which suppresses the want comparison).
type newVersionCase struct {
	Name    string          `yaml:"name"`
	Raw     string          `yaml:"raw"`
	Want    release.Version `yaml:"want"`
	WantErr bool            `yaml:"wantErr"`
}

// grammarFixtures is the parsed testdata/grammar/versions.yaml corpus: one
// section per grammar table-test. The version_kind and parse_tag sections were
// migrated onto the github.com/peasant-labs/schema/testcase corpus standard (see
// grammar_corpus_test.go); this loader serves the version-validation cases.
type grammarFixtures struct {
	NewVersion []newVersionCase `yaml:"new_version"`
}

// loadGrammarFixtures parses the embedded grammar corpus.
func loadGrammarFixtures(t *testing.T) grammarFixtures {
	t.Helper()
	var f grammarFixtures
	if err := yaml.Unmarshal(grammarFixtureYAML, &f); err != nil {
		t.Fatalf("load grammar fixtures (testdata/grammar/versions.yaml): %v", err)
	}
	return f
}
