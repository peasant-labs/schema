package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/projectlabel/remote-label.yaml
var remoteLabelCasesYAML []byte

//go:embed testdata/projectlabel/remote_label_manifest.yaml
var remoteLabelManifestYAML []byte

// remoteLabelExpected is the expected two-valued RemoteLabel result for one
// fixture case.
type remoteLabelExpected struct {
	Label string `json:"label" yaml:"label"`
	OK    bool   `json:"ok" yaml:"ok"`
}

// remoteLabelFixtureManifest is the deletion-protection inventory for the
// remote-label corpus: every name it lists must be present in the corpus, so
// a deleted row (rather than a renamed or replaced one) is caught even though
// the corpus is otherwise free to grow. It deliberately carries no case-count
// field: count/minimum guards are out of scope for this fixture family.
type remoteLabelFixtureManifest struct {
	RequiredCaseNames []string
}

type remoteLabelFixtureManifestYAML struct {
	RequiredCaseNames []remoteLabelManifestCaseName `yaml:"requiredCaseNames"`
}

// remoteLabelManifestCaseName rejects a non-string requiredCaseNames entry at
// decode time instead of silently coercing it.
type remoteLabelManifestCaseName string

func (name *remoteLabelManifestCaseName) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
		return fmt.Errorf("required case name must be a YAML string")
	}
	*name = remoteLabelManifestCaseName(value.Value)
	return nil
}

func loadRemoteLabelFixtures(t *testing.T) testcase.Corpus[string, remoteLabelExpected] {
	t.Helper()
	fixtures, _, err := loadRemoteLabelFixturesFromYAML(remoteLabelCasesYAML, remoteLabelManifestYAML)
	if err != nil {
		t.Fatalf("load remote label fixtures and inventory: %v", err)
	}
	return fixtures
}

func loadRemoteLabelFixturesFromYAML(corpusYAML, manifestYAML []byte) (testcase.Corpus[string, remoteLabelExpected], remoteLabelFixtureManifest, error) {
	fixtures, err := testcase.LoadCorpus[string, remoteLabelExpected](corpusYAML)
	if err != nil {
		return testcase.Corpus[string, remoteLabelExpected]{}, remoteLabelFixtureManifest{}, fmt.Errorf("decode remote label corpus: %w", err)
	}
	manifest, err := decodeRemoteLabelFixtureManifest(manifestYAML)
	if err != nil {
		return testcase.Corpus[string, remoteLabelExpected]{}, remoteLabelFixtureManifest{}, err
	}
	if err := validateRemoteLabelFixtureInventory(fixtures, manifest); err != nil {
		return testcase.Corpus[string, remoteLabelExpected]{}, remoteLabelFixtureManifest{}, err
	}
	return fixtures, manifest, nil
}

func decodeRemoteLabelFixtureManifest(data []byte) (remoteLabelFixtureManifest, error) {
	var decoded remoteLabelFixtureManifestYAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&decoded); err != nil {
		return remoteLabelFixtureManifest{}, fmt.Errorf("decode remote label manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return remoteLabelFixtureManifest{}, fmt.Errorf("decode trailing remote label manifest document: %w", err)
		}
		return remoteLabelFixtureManifest{}, fmt.Errorf("decode remote label manifest: multiple YAML documents are not allowed")
	}
	manifest := remoteLabelFixtureManifest{
		RequiredCaseNames: make([]string, len(decoded.RequiredCaseNames)),
	}
	for index, name := range decoded.RequiredCaseNames {
		manifest.RequiredCaseNames[index] = string(name)
	}
	if len(manifest.RequiredCaseNames) == 0 {
		return remoteLabelFixtureManifest{}, fmt.Errorf("remote label manifest has no requiredCaseNames; name at least one case the corpus must never drop")
	}
	seen := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for index, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" {
			return remoteLabelFixtureManifest{}, fmt.Errorf("remote label manifest requiredCaseNames[%d] is blank", index)
		}
		if _, exists := seen[name]; exists {
			return remoteLabelFixtureManifest{}, fmt.Errorf("remote label manifest repeats required case name %q", name)
		}
		seen[name] = struct{}{}
	}
	return manifest, nil
}

// validateRemoteLabelFixtureInventory fails when a required case name is
// missing from the corpus (a deleted row) or the corpus itself is invalid.
// It deliberately does not compare a case count: the corpus may grow freely
// as long as every named case remains present.
func validateRemoteLabelFixtureInventory(fixtures testcase.Corpus[string, remoteLabelExpected], manifest remoteLabelFixtureManifest) error {
	if err := fixtures.Validate(); err != nil {
		return fmt.Errorf("validate remote label corpus: %w", err)
	}
	present := make(map[string]struct{}, len(fixtures.Cases))
	for _, c := range fixtures.Cases {
		present[c.Name] = struct{}{}
	}
	for _, required := range manifest.RequiredCaseNames {
		if _, ok := present[required]; !ok {
			return fmt.Errorf("remote label corpus is missing required case %q named in the manifest; a case was deleted or renamed", required)
		}
	}
	return nil
}

// TestRemoteLabel drives schema.RemoteLabel over every fixture case and
// asserts the exact two-valued result. This is the production code path: the
// same function peasant and village both call.
func TestRemoteLabel(t *testing.T) {
	fixtures := loadRemoteLabelFixtures(t)
	for _, c := range fixtures.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			label, ok := schema.RemoteLabel(c.Input)
			if label != c.Expected.Label || ok != c.Expected.OK {
				t.Fatalf("RemoteLabel(%q) = (%q, %v), want (%q, %v)", c.Input, label, ok, c.Expected.Label, c.Expected.OK)
			}
			switch c.Classification {
			case testcase.MustPass:
				if !ok {
					t.Fatalf("case %q is classified must-pass but RemoteLabel returned ok=false", c.Name)
				}
			case testcase.MustFail:
				if ok {
					t.Fatalf("case %q is classified must-fail but RemoteLabel returned ok=true", c.Name)
				}
			}
		})
	}
}

// TestRemoteLabelFixtureInventory proves the deletion-protection manifest is
// load-bearing: a manifest naming a case absent from the corpus must fail to
// load, and the shipped corpus + manifest pair must load cleanly together.
func TestRemoteLabelFixtureInventory(t *testing.T) {
	if _, _, err := loadRemoteLabelFixturesFromYAML(remoteLabelCasesYAML, remoteLabelManifestYAML); err != nil {
		t.Fatalf("shipped remote label corpus and manifest must load together: %v", err)
	}

	t.Run("manifest_naming_an_absent_case_is_rejected", func(t *testing.T) {
		mutatedManifest := []byte("requiredCaseNames:\n  - this_case_does_not_exist_in_the_corpus\n")
		if _, _, err := loadRemoteLabelFixturesFromYAML(remoteLabelCasesYAML, mutatedManifest); err == nil {
			t.Fatal("expected an error when the manifest names a case absent from the corpus")
		}
	})

	t.Run("deleting_a_required_row_is_rejected", func(t *testing.T) {
		// Simulate a fixture-row deletion by decoding the real corpus, then
		// dropping the first case named in the manifest, and confirming the
		// manifest that shipped alongside it now rejects the mutated corpus.
		fixtures, err := testcase.LoadCorpus[string, remoteLabelExpected](remoteLabelCasesYAML)
		if err != nil {
			t.Fatalf("decode remote label corpus: %v", err)
		}
		manifest, err := decodeRemoteLabelFixtureManifest(remoteLabelManifestYAML)
		if err != nil {
			t.Fatalf("decode remote label manifest: %v", err)
		}
		if len(manifest.RequiredCaseNames) == 0 {
			t.Fatal("manifest has no required case names to delete")
		}
		victim := manifest.RequiredCaseNames[0]
		mutated := make([]testcase.Case[string, remoteLabelExpected], 0, len(fixtures.Cases))
		for _, c := range fixtures.Cases {
			if c.Name == victim {
				continue
			}
			mutated = append(mutated, c)
		}
		if len(mutated) != len(fixtures.Cases)-1 {
			t.Fatalf("expected to delete exactly one case named %q", victim)
		}
		if err := validateRemoteLabelFixtureInventory(testcase.Corpus[string, remoteLabelExpected]{Cases: mutated}, manifest); err == nil {
			t.Fatalf("expected validateRemoteLabelFixtureInventory to reject a corpus missing required case %q", victim)
		}
	})
}
