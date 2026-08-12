package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/session-detail/transcripts/turn_models.yaml
var turnModelFixtureYAML []byte

//go:embed testdata/session-detail/transcripts/turn_models_manifest.yaml
var turnModelManifestYAML []byte

type turnModelFixtureInput struct {
	Index         int         `yaml:"index"`
	Role          schema.Role `yaml:"role"`
	Content       string      `yaml:"content"`
	Timestamp     time.Time   `yaml:"timestamp"`
	Depth         int         `yaml:"depth"`
	ObservedModel yaml.Node   `yaml:"observedModel,omitempty"`
}

type turnModelFixtureExpected struct {
	Accepted             bool   `yaml:"accepted"`
	ObservedModelPresent bool   `yaml:"observedModelPresent"`
	ObservedModel        string `yaml:"observedModel,omitempty"`
	ErrorContains        string `yaml:"errorContains,omitempty"`
}

type turnModelFixtureManifest struct {
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

func TestTurnDetailObservedModelFixture(t *testing.T) {
	corpus, _ := loadTurnModelFixtures(t)
	for _, fixtureCase := range corpus.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			if !fixtureCase.Input.Role.IsValid() {
				t.Fatalf("fixture role %q is not a known role", fixtureCase.Input.Role)
			}
			if turnModelNodePresent(fixtureCase.Input.ObservedModel) != fixtureCase.Expected.ObservedModelPresent {
				t.Fatalf("observedModel presence disagrees with expected presence=%v", fixtureCase.Expected.ObservedModelPresent)
			}

			var raw string
			constructorValid := !fixtureCase.Expected.ObservedModelPresent
			var constructorErr error
			if fixtureCase.Expected.ObservedModelPresent && fixtureCase.Input.ObservedModel.Tag == "!!str" {
				raw = fixtureCase.Input.ObservedModel.Value
				observed, err := schema.NewObservedModelID(raw)
				constructorErr = err
				constructorValid = err == nil
				if constructorValid && observed.String() != raw {
					t.Fatalf("constructor changed accepted bytes: got %q, want %q", observed, raw)
				}
			}
			if constructorValid != fixtureCase.Expected.Accepted {
				t.Fatalf("NewObservedModelID validity=%v, want accepted=%v", constructorValid, fixtureCase.Expected.Accepted)
			}
			if !fixtureCase.Expected.Accepted {
				if fixtureCase.Input.ObservedModel.Tag == "!!str" && constructorErr == nil {
					t.Fatal("must-fail string row did not receive a constructor validation error")
				}
				return
			}

			turn := schema.TurnDetail{
				Index:         fixtureCase.Input.Index,
				Role:          fixtureCase.Input.Role,
				Content:       fixtureCase.Input.Content,
				Timestamp:     fixtureCase.Input.Timestamp,
				Depth:         fixtureCase.Input.Depth,
				ObservedModel: schema.ObservedModelID(raw),
			}
			wire, err := json.Marshal(turn)
			if err != nil {
				t.Fatalf("marshal TurnDetail after constructor validation: %v", err)
			}
			var object map[string]json.RawMessage
			if err := json.Unmarshal(wire, &object); err != nil {
				t.Fatalf("decode marshaled TurnDetail: %v", err)
			}
			rawObserved, emitted := object["observedModel"]
			if emitted != fixtureCase.Expected.ObservedModelPresent {
				t.Fatalf("observedModel emitted=%v, expected=%v; wire=%s", emitted, fixtureCase.Expected.ObservedModelPresent, wire)
			}
			if !emitted {
				return
			}
			var got string
			if err := json.Unmarshal(rawObserved, &got); err != nil {
				t.Fatalf("decode emitted observedModel: %v", err)
			}
			if got != fixtureCase.Expected.ObservedModel || got != raw {
				t.Fatalf("observedModel bytes=%q, want fixture=%q and source=%q", got, fixtureCase.Expected.ObservedModel, raw)
			}
		})
	}
}

func TestTurnModelFixtureInventoryRejectsCountPreservingRename(t *testing.T) {
	_, manifest := loadTurnModelFixtures(t)
	old := "name: " + manifest.RequiredCaseNames[0]
	mutated := replaceTurnModelFixtureText(t, string(turnModelFixtureYAML), old, "name: unregistered-observed-model-case")
	corpus, err := testcase.LoadCorpus[turnModelFixtureInput, turnModelFixtureExpected]([]byte(mutated))
	if err != nil {
		t.Fatalf("load count-preserving rename mutation before inventory validation: %v", err)
	}
	if err := validateTurnModelFixtureInventory(corpus, manifest); err == nil {
		t.Fatal("count-preserving fixture rename passed the independent required-name inventory")
	}
}

func loadTurnModelFixtures(t *testing.T) (testcase.Corpus[turnModelFixtureInput, turnModelFixtureExpected], turnModelFixtureManifest) {
	t.Helper()
	corpus, err := testcase.LoadCorpus[turnModelFixtureInput, turnModelFixtureExpected](turnModelFixtureYAML)
	if err != nil {
		t.Fatalf("load turn model corpus: %v", err)
	}
	manifest, err := decodeTurnModelFixtureManifest(turnModelManifestYAML)
	if err != nil {
		t.Fatalf("load turn model manifest: %v", err)
	}
	if err := validateTurnModelFixtureInventory(corpus, manifest); err != nil {
		t.Fatalf("validate turn model fixture inventory: %v", err)
	}
	return corpus, manifest
}

func decodeTurnModelFixtureManifest(data []byte) (turnModelFixtureManifest, error) {
	var manifest turnModelFixtureManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return turnModelFixtureManifest{}, fmt.Errorf("decode turn model manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return turnModelFixtureManifest{}, fmt.Errorf("decode trailing turn model manifest document: %w", err)
		}
		return turnModelFixtureManifest{}, fmt.Errorf("decode turn model manifest: multiple YAML documents are not allowed")
	}
	if manifest.ExpectedCaseCount <= 0 {
		return turnModelFixtureManifest{}, fmt.Errorf("turn model manifest expectedCaseCount must be positive, got %d", manifest.ExpectedCaseCount)
	}
	if len(manifest.RequiredCaseNames) != manifest.ExpectedCaseCount {
		return turnModelFixtureManifest{}, fmt.Errorf("turn model manifest has %d required names, want exactly %d", len(manifest.RequiredCaseNames), manifest.ExpectedCaseCount)
	}
	seen := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for index, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" || name != strings.TrimSpace(name) {
			return turnModelFixtureManifest{}, fmt.Errorf("turn model manifest requiredCaseNames[%d]=%q must be non-empty without edge whitespace", index, name)
		}
		if _, duplicate := seen[name]; duplicate {
			return turnModelFixtureManifest{}, fmt.Errorf("turn model manifest repeats required case name %q", name)
		}
		seen[name] = struct{}{}
	}
	return manifest, nil
}

func validateTurnModelFixtureInventory(corpus testcase.Corpus[turnModelFixtureInput, turnModelFixtureExpected], manifest turnModelFixtureManifest) error {
	if len(corpus.Cases) != manifest.ExpectedCaseCount {
		return fmt.Errorf("turn model corpus has %d cases, want exactly %d", len(corpus.Cases), manifest.ExpectedCaseCount)
	}
	required := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		required[name] = struct{}{}
	}
	actual := make(map[string]struct{}, len(corpus.Cases))
	for _, fixtureCase := range corpus.Cases {
		if _, duplicate := actual[fixtureCase.Name]; duplicate {
			return fmt.Errorf("turn model corpus repeats case name %q", fixtureCase.Name)
		}
		actual[fixtureCase.Name] = struct{}{}
		if _, registered := required[fixtureCase.Name]; !registered {
			return fmt.Errorf("turn model corpus contains unregistered case %q", fixtureCase.Name)
		}
	}
	for name := range required {
		if _, present := actual[name]; !present {
			return fmt.Errorf("turn model corpus is missing required case %q", name)
		}
	}
	return nil
}

func turnModelNodePresent(node yaml.Node) bool {
	return node.Kind != 0
}

func replaceTurnModelFixtureText(t *testing.T, source, old, replacement string) string {
	t.Helper()
	if count := strings.Count(source, old); count != 1 {
		t.Fatalf("fixture mutation target %q occurs %d times, want exactly once", old, count)
	}
	return strings.Replace(source, old, replacement, 1)
}
