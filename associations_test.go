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
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/associations.yaml
var associationsCasesYAML []byte

type associationRepairMutationKind string

const (
	associationRepairReplaceID            associationRepairMutationKind = "replace_id"
	associationRepairDropLastEvidence     associationRepairMutationKind = "drop_last_evidence"
	associationRepairClearTouchedFilePath associationRepairMutationKind = "clear_touched_file_path"
	associationRepairSwapEvidence         associationRepairMutationKind = "swap_evidence"
)

type associationRepairMutation struct {
	Kind  associationRepairMutationKind `yaml:"kind"`
	Input string                        `yaml:"input"`
}

type associationRepairInput struct {
	SourceCase string                    `yaml:"sourceCase"`
	Mutation   associationRepairMutation `yaml:"mutation"`
}

type associationRepairExpected struct {
	OriginalErrorContains string `yaml:"originalErrorContains"`
	PostMutationValid     bool   `yaml:"postMutationValid"`
}

type associationFixtureManifest struct {
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

type associationFixtures struct {
	Cases          []testcase.Case[schema.SessionAssociation, bool]                   `yaml:"cases"`
	CaseManifest   associationFixtureManifest                                         `yaml:"caseManifest"`
	RepairManifest associationFixtureManifest                                         `yaml:"repairManifest"`
	Repairs        testcase.Corpus[associationRepairInput, associationRepairExpected] `yaml:"repairs"`
}

func loadAssociationFixtures(t *testing.T) associationFixtures {
	t.Helper()
	var fixtures associationFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(associationsCasesYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("decode associations fixtures: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			t.Fatalf("decode trailing associations fixture document: %v", err)
		}
		t.Fatal("decode associations fixtures: multiple YAML documents are not allowed")
	}
	corpus := testcase.Corpus[schema.SessionAssociation, bool]{Cases: fixtures.Cases}
	assert.RequireMin(t, corpus, len(schema.AllAssociationConclusions)+len(schema.AllAssociationEvidenceKinds)+10)
	assert.RequireValid(t, corpus)
	requireAssociationFixtureInventory(t, "case", fixtures.CaseManifest, associationCaseNames(fixtures.Cases))
	assert.RequireValid(t, fixtures.Repairs)
	requireAssociationFixtureInventory(t, "repair", fixtures.RepairManifest, associationRepairNames(fixtures.Repairs.Cases))
	return fixtures
}

func associationCaseNames(cases []testcase.Case[schema.SessionAssociation, bool]) []string {
	names := make([]string, len(cases))
	for index, fixture := range cases {
		names[index] = fixture.Name
	}
	return names
}

func associationRepairNames(cases []testcase.Case[associationRepairInput, associationRepairExpected]) []string {
	names := make([]string, len(cases))
	for index, repair := range cases {
		names[index] = repair.Name
	}
	return names
}

func requireAssociationFixtureInventory(t *testing.T, inventory string, manifest associationFixtureManifest, names []string) {
	t.Helper()
	if manifest.ExpectedCaseCount <= 0 || len(manifest.RequiredCaseNames) != manifest.ExpectedCaseCount || len(names) != manifest.ExpectedCaseCount {
		t.Fatalf("association %s manifest and corpus must contain exactly %d cases", inventory, manifest.ExpectedCaseCount)
	}
	required := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" {
			t.Fatalf("association %s manifest contains an empty case name", inventory)
		}
		if _, exists := required[name]; exists {
			t.Fatalf("association %s manifest repeats case name %q", inventory, name)
		}
		required[name] = struct{}{}
	}
	for _, name := range names {
		if _, exists := required[name]; !exists {
			t.Fatalf("association %s corpus contains unregistered case %q", inventory, name)
		}
		delete(required, name)
	}
	for name := range required {
		t.Fatalf("association %s corpus is missing required case %q", inventory, name)
	}
}

// TestSessionAssociation_FixtureContract drives the public association validator
// and covers every conclusion, confidence, and atomic evidence kind.
func TestSessionAssociation_FixtureContract(t *testing.T) {
	fixtures := loadAssociationFixtures(t)
	coveredConclusions := map[schema.AssociationConclusion]bool{}
	coveredConfidences := map[schema.Confidence]bool{}
	coveredEvidenceKinds := map[schema.AssociationEvidenceKind]bool{}
	for _, c := range fixtures.Cases {
		err := c.Input.Validate()
		valid := err == nil
		if valid != c.Expected {
			t.Errorf("%s: Validate() error=%v (valid=%v), want valid=%v", c.Name, err, valid, c.Expected)
			continue
		}
		if c.Classification == testcase.MustFail {
			requireActionableValidationError(t, err)
			continue
		}
		coveredConclusions[c.Input.Conclusion] = true
		coveredConfidences[c.Input.Confidence] = true
		for _, observation := range c.Input.Evidence {
			coveredEvidenceKinds[observation.Kind] = true
		}
	}
	for _, conclusion := range schema.AllAssociationConclusions {
		if !coveredConclusions[conclusion] {
			t.Errorf("AssociationConclusion member %q has no must-pass fixture case", conclusion)
		}
	}
	for _, confidence := range schema.AllConfidences {
		if !coveredConfidences[confidence] {
			t.Errorf("Confidence member %q has no must-pass fixture case", confidence)
		}
	}
	for _, kind := range schema.AllAssociationEvidenceKinds {
		if !coveredEvidenceKinds[kind] {
			t.Errorf("AssociationEvidenceKind member %q has no must-pass fixture case", kind)
		}
	}
}

func cloneAssociation(input schema.SessionAssociation) schema.SessionAssociation {
	cloned := input
	cloned.Evidence = append([]schema.AssociationEvidenceObservation(nil), input.Evidence...)
	for index := range cloned.Evidence {
		observation := &cloned.Evidence[index]
		if observation.RecordedCommitHash != nil {
			value := *observation.RecordedCommitHash
			observation.RecordedCommitHash = &value
		}
		if observation.TouchedFilePath != nil {
			value := *observation.TouchedFilePath
			observation.TouchedFilePath = &value
		}
		if observation.BranchName != nil {
			value := *observation.BranchName
			observation.BranchName = &value
		}
		if observation.WindowStartMs != nil {
			value := *observation.WindowStartMs
			observation.WindowStartMs = &value
		}
		if observation.WindowEndMs != nil {
			value := *observation.WindowEndMs
			observation.WindowEndMs = &value
		}
	}
	return cloned
}

func applyAssociationRepair(input *schema.SessionAssociation, mutation associationRepairMutation) error {
	switch mutation.Kind {
	case associationRepairReplaceID:
		if strings.TrimSpace(mutation.Input) == "" {
			return fmt.Errorf("repair %q requires a non-empty association ID", mutation.Kind)
		}
		input.ID = schema.AssociationID(mutation.Input)
	case associationRepairDropLastEvidence:
		if mutation.Input != "" || len(input.Evidence) == 0 {
			return fmt.Errorf("repair %q requires non-empty evidence and no input", mutation.Kind)
		}
		input.Evidence = input.Evidence[:len(input.Evidence)-1]
	case associationRepairClearTouchedFilePath:
		if mutation.Input != "" || len(input.Evidence) != 1 {
			return fmt.Errorf("repair %q requires exactly one observation and no input", mutation.Kind)
		}
		input.Evidence[0].TouchedFilePath = nil
	case associationRepairSwapEvidence:
		if mutation.Input != "" || len(input.Evidence) != 2 {
			return fmt.Errorf("repair %q requires exactly two observations and no input", mutation.Kind)
		}
		input.Evidence[0], input.Evidence[1] = input.Evidence[1], input.Evidence[0]
	default:
		return fmt.Errorf("unknown association repair mutation %q", mutation.Kind)
	}
	return nil
}

// TestSessionAssociation_RepairCorpus proves representative invalid inputs are
// repaired through the same public validator rather than a test-only path.
func TestSessionAssociation_RepairCorpus(t *testing.T) {
	fixtures := loadAssociationFixtures(t)
	sources := make(map[string]testcase.Case[schema.SessionAssociation, bool], len(fixtures.Cases))
	for _, fixture := range fixtures.Cases {
		sources[fixture.Name] = fixture
	}
	for _, repair := range fixtures.Repairs.Cases {
		t.Run(repair.Name, func(t *testing.T) {
			source, exists := sources[repair.Input.SourceCase]
			if !exists || source.Classification != testcase.MustFail || source.Expected {
				t.Fatalf("repair source %q must be a rejected fixture", repair.Input.SourceCase)
			}
			if repair.Classification != testcase.MustPass || !repair.Expected.PostMutationValid {
				t.Fatal("repair must be must-pass with postMutationValid=true")
			}
			if strings.TrimSpace(repair.Expected.OriginalErrorContains) == "" {
				t.Fatal("repair must declare originalErrorContains")
			}
			input := cloneAssociation(source.Input)
			originalErr := input.Validate()
			requireActionableValidationError(t, originalErr, repair.Expected.OriginalErrorContains)
			if err := applyAssociationRepair(&input, repair.Input.Mutation); err != nil {
				t.Fatalf("apply repair mutation: %v", err)
			}
			if err := input.Validate(); (err == nil) != repair.Expected.PostMutationValid {
				t.Fatalf("post-mutation Validate() error=%v, want valid=%v", err, repair.Expected.PostMutationValid)
			}
		})
	}
}
