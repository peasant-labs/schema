package schema

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/map/diff_enums.yaml
var diffEnumCasesYAML []byte

//go:embed testdata/map/diff_enums_manifest.yaml
var diffEnumManifestYAML []byte

type diffEnumInput struct {
	Enum  string `yaml:"enum"`
	Value string `yaml:"value"`
}

type diffEnumExpected struct {
	Valid         bool   `yaml:"valid"`
	JSON          string `yaml:"json"`
	ErrorContains string `yaml:"error_contains"`
}

type diffEnumManifestEntry struct {
	Enum           string                  `yaml:"enum"`
	Name           string                  `yaml:"name"`
	Classification testcase.Classification `yaml:"classification"`
}

type diffEnumMutationInput struct {
	Kind                      string                  `yaml:"kind"`
	Target                    string                  `yaml:"target"`
	ReplacementName           string                  `yaml:"replacement_name"`
	ReplacementClassification testcase.Classification `yaml:"replacement_classification"`
}

type diffEnumManifest struct {
	Cases     []diffEnumManifestEntry                      `yaml:"cases"`
	Mutations testcase.Corpus[diffEnumMutationInput, bool] `yaml:"mutations"`
}

func loadDiffEnumFixtures(t *testing.T) (testcase.Corpus[diffEnumInput, diffEnumExpected], diffEnumManifest) {
	t.Helper()
	corpus, err := testcase.LoadCorpus[diffEnumInput, diffEnumExpected](diffEnumCasesYAML)
	if err != nil {
		t.Fatalf("load diff enum corpus: %v", err)
	}
	assert.RequireMin(t, corpus, len(AllFileChangeStatuses)+len(AllDiffLineKinds)+2)
	assert.RequireValid(t, corpus)

	manifest, err := decodeDiffEnumManifest(diffEnumManifestYAML)
	if err != nil {
		t.Fatalf("load independent diff enum manifest: %v", err)
	}
	if err := validateDiffEnumManifest(corpus, manifest); err != nil {
		t.Fatalf("diff enum corpus differs from independent manifest: %v", err)
	}
	return corpus, manifest
}

func decodeDiffEnumManifest(data []byte) (diffEnumManifest, error) {
	var manifest diffEnumManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return diffEnumManifest{}, fmt.Errorf("decode diff enum manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return diffEnumManifest{}, fmt.Errorf("decode trailing diff enum manifest document: %w", err)
		}
		return diffEnumManifest{}, fmt.Errorf("decode diff enum manifest: multiple YAML documents are not allowed")
	}
	if err := manifest.Mutations.Validate(); err != nil {
		return diffEnumManifest{}, fmt.Errorf("validate diff enum manifest mutations: %w", err)
	}
	if len(manifest.Cases) != len(AllFileChangeStatuses)+len(AllDiffLineKinds)+2 {
		return diffEnumManifest{}, fmt.Errorf("diff enum manifest has %d cases, want exactly %d", len(manifest.Cases), len(AllFileChangeStatuses)+len(AllDiffLineKinds)+2)
	}
	if len(manifest.Mutations.Cases) != 2 {
		return diffEnumManifest{}, fmt.Errorf("diff enum manifest has %d mutations, want exactly 2", len(manifest.Mutations.Cases))
	}
	return manifest, nil
}

func validateDiffEnumManifest(corpus testcase.Corpus[diffEnumInput, diffEnumExpected], manifest diffEnumManifest) error {
	if len(corpus.Cases) != len(manifest.Cases) {
		return fmt.Errorf("corpus has %d cases, manifest has %d", len(corpus.Cases), len(manifest.Cases))
	}
	for index, fixture := range corpus.Cases {
		expected := manifest.Cases[index]
		if fixture.Input.Enum != expected.Enum || fixture.Name != expected.Name || fixture.Classification != expected.Classification {
			return fmt.Errorf("case %d identity=(%q,%q,%q), want (%q,%q,%q)", index, fixture.Input.Enum, fixture.Name, fixture.Classification, expected.Enum, expected.Name, expected.Classification)
		}
	}
	return nil
}

func evaluateDiffEnum(input diffEnumInput) (valid bool, rendered string, encoded string, validationErr error) {
	switch input.Enum {
	case "file-change-status":
		value := FileChangeStatus(input.Value)
		payload, err := json.Marshal(value)
		if err != nil {
			return false, "", "", err
		}
		return value.IsValid(), value.String(), string(payload), value.Validate()
	case "diff-line-kind":
		value := DiffLineKind(input.Value)
		payload, err := json.Marshal(value)
		if err != nil {
			return false, "", "", err
		}
		return value.IsValid(), value.String(), string(payload), value.Validate()
	default:
		return false, "", "", fmt.Errorf("unknown diff enum fixture type %q", input.Enum)
	}
}

func TestMapDiffEnumsFixtureContract(t *testing.T) {
	corpus, _ := loadDiffEnumFixtures(t)
	coveredStatuses := map[FileChangeStatus]bool{}
	coveredKinds := map[DiffLineKind]bool{}

	for _, fixture := range corpus.Cases {
		valid, rendered, encoded, validationErr := evaluateDiffEnum(fixture.Input)
		if valid != fixture.Expected.Valid {
			t.Errorf("%s: IsValid=%v, want %v", fixture.Name, valid, fixture.Expected.Valid)
		}
		if rendered != fixture.Input.Value {
			t.Errorf("%s: String=%q, want %q", fixture.Name, rendered, fixture.Input.Value)
		}
		if encoded != fixture.Expected.JSON {
			t.Errorf("%s: JSON=%q, want %q", fixture.Name, encoded, fixture.Expected.JSON)
		}
		if fixture.Expected.Valid {
			if validationErr != nil {
				t.Errorf("%s: Validate returned %v", fixture.Name, validationErr)
			}
			switch fixture.Input.Enum {
			case "file-change-status":
				coveredStatuses[FileChangeStatus(fixture.Input.Value)] = true
			case "diff-line-kind":
				coveredKinds[DiffLineKind(fixture.Input.Value)] = true
			}
			continue
		}
		if validationErr == nil || !strings.Contains(validationErr.Error(), fixture.Expected.ErrorContains) {
			t.Errorf("%s: validation error=%v, want text containing %q", fixture.Name, validationErr, fixture.Expected.ErrorContains)
			continue
		}
		for _, required := range []string{" at schema.", "during wire-boundary validation", "callers cannot", "use a member of schema.All"} {
			if !strings.Contains(validationErr.Error(), required) {
				t.Errorf("%s: validation error %q is not actionable; missing %q", fixture.Name, validationErr, required)
			}
		}
	}

	for _, status := range AllFileChangeStatuses {
		if !coveredStatuses[status] {
			t.Errorf("AllFileChangeStatuses member %q has no must-pass fixture", status)
		}
	}
	for _, kind := range AllDiffLineKinds {
		if !coveredKinds[kind] {
			t.Errorf("AllDiffLineKinds member %q has no must-pass fixture", kind)
		}
	}
}

func TestMapDiffEnumPayloadFieldsPreserveJSON(t *testing.T) {
	corpus, _ := loadDiffEnumFixtures(t)
	for _, fixture := range corpus.Cases {
		if !fixture.Expected.Valid {
			continue
		}
		switch fixture.Input.Enum {
		case "file-change-status":
			status := FileChangeStatus(fixture.Input.Value)
			for _, payload := range []any{
				FileChange{Path: "src/example.go", Status: status},
				ChangeDiffPayload{Branch: "feature/diff", File: "src/example.go", Status: status, Hunks: []DiffHunk{}},
			} {
				encoded, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("%s: marshal payload: %v", fixture.Name, err)
				}
				if !bytes.Contains(encoded, []byte(`"status":`+fixture.Expected.JSON)) {
					t.Errorf("%s: payload JSON %s does not preserve status token", fixture.Name, encoded)
				}
			}
		case "diff-line-kind":
			line := DiffLine{Kind: DiffLineKind(fixture.Input.Value), Text: "example"}
			encoded, err := json.Marshal(line)
			if err != nil {
				t.Fatalf("%s: marshal diff line: %v", fixture.Name, err)
			}
			if !bytes.Contains(encoded, []byte(`"kind":`+fixture.Expected.JSON)) {
				t.Errorf("%s: diff line JSON %s does not preserve kind token", fixture.Name, encoded)
			}
		}
	}

	statusType := reflect.TypeOf(FileChangeStatus(""))
	for _, contract := range []struct {
		Owner reflect.Type
		Field string
	}{
		{Owner: reflect.TypeOf(FileChange{}), Field: "Status"},
		{Owner: reflect.TypeOf(ChangeDiffPayload{}), Field: "Status"},
	} {
		field, ok := contract.Owner.FieldByName(contract.Field)
		if !ok || field.Type != statusType {
			t.Errorf("%s.%s type=%v, want schema.FileChangeStatus", contract.Owner.Name(), contract.Field, field.Type)
		}
	}
	field, ok := reflect.TypeOf(DiffLine{}).FieldByName("Kind")
	if !ok || field.Type != reflect.TypeOf(DiffLineKind("")) {
		t.Errorf("DiffLine.Kind type=%v, want schema.DiffLineKind", field.Type)
	}
}

func TestMapDiffEnumManifestRejectsCountPreservingMutations(t *testing.T) {
	corpus, manifest := loadDiffEnumFixtures(t)
	for _, mutation := range manifest.Mutations.Cases {
		mutated := corpus
		mutated.Cases = append([]testcase.Case[diffEnumInput, diffEnumExpected](nil), corpus.Cases...)
		target := -1
		for index := range mutated.Cases {
			if mutated.Cases[index].Name == mutation.Input.Target {
				target = index
				break
			}
		}
		if target < 0 {
			t.Fatalf("%s: target %q is absent", mutation.Name, mutation.Input.Target)
		}
		switch mutation.Input.Kind {
		case "rename":
			mutated.Cases[target].Name = mutation.Input.ReplacementName
		case "replace":
			mutated.Cases[target].Name = mutation.Input.ReplacementName
			mutated.Cases[target].Classification = mutation.Input.ReplacementClassification
		default:
			t.Fatalf("%s: unknown mutation kind %q", mutation.Name, mutation.Input.Kind)
		}
		accepted := validateDiffEnumManifest(mutated, manifest) == nil
		if accepted != mutation.Expected {
			t.Errorf("%s: accepted=%v, want %v", mutation.Name, accepted, mutation.Expected)
		}
	}
}

func TestMapDiffEnumManifestIsStrict(t *testing.T) {
	unknown := bytes.Replace(diffEnumManifestYAML, []byte("mutations:\n"), []byte("unknown: true\nmutations:\n"), 1)
	if _, err := decodeDiffEnumManifest(unknown); err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("unknown manifest field error=%v, want strict-field rejection", err)
	}
	if _, err := decodeDiffEnumManifest(append(append([]byte(nil), diffEnumManifestYAML...), []byte("\n---\n{}\n")...)); err == nil || !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Fatalf("trailing manifest document error=%v, want multiple-document rejection", err)
	}
}
