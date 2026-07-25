package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/rewrites.yaml
var rewritesCasesYAML []byte

type rewriteRepairMutationKind string

const (
	rewriteRepairClearSuccessorHash rewriteRepairMutationKind = "clear_successor_hash"
	rewriteRepairAppendSessionID    rewriteRepairMutationKind = "append_session_id"
)

type rewriteRepairMutation struct {
	Kind  rewriteRepairMutationKind `yaml:"kind"`
	Input string                    `yaml:"input"`
}

type rewriteRepairFixtureInput struct {
	SourceCase string                `yaml:"sourceCase"`
	Mutation   rewriteRepairMutation `yaml:"mutation"`
}

type rewriteRepairFixtureExpected struct {
	OriginalErrorContains string `yaml:"originalErrorContains"`
	PostMutationValid     bool   `yaml:"postMutationValid"`
}

type rewriteRepairManifest struct {
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

type rewriteFixtures struct {
	Cases          []testcase.Case[schema.RewrittenCommit, bool]                            `yaml:"cases"`
	RepairManifest rewriteRepairManifest                                                    `yaml:"repairManifest"`
	Repairs        testcase.Corpus[rewriteRepairFixtureInput, rewriteRepairFixtureExpected] `yaml:"repairs"`
}

func loadRewriteFixtures(t *testing.T) rewriteFixtures {
	t.Helper()
	var fixtures rewriteFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(rewritesCasesYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("decode rewrites fixtures: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			t.Fatalf("decode trailing rewrites fixture document: %v", err)
		}
		t.Fatal("decode rewrites fixtures: multiple YAML documents are not allowed")
	}

	corpus := testcase.Corpus[schema.RewrittenCommit, bool]{Cases: fixtures.Cases}
	assert.RequireMin(t, corpus, 17)
	assert.RequireValid(t, corpus)
	assert.RequireValid(t, fixtures.Repairs)
	requireRewriteRepairInventory(t, fixtures)
	return fixtures
}

func requireRewriteRepairInventory(t *testing.T, fixtures rewriteFixtures) {
	t.Helper()
	manifest := fixtures.RepairManifest
	if manifest.ExpectedCaseCount <= 0 {
		t.Fatalf("repair manifest expectedCaseCount must be positive, got %d", manifest.ExpectedCaseCount)
	}
	if len(manifest.RequiredCaseNames) != manifest.ExpectedCaseCount {
		t.Fatalf("repair manifest has %d requiredCaseNames, want %d", len(manifest.RequiredCaseNames), manifest.ExpectedCaseCount)
	}
	if len(fixtures.Repairs.Cases) != manifest.ExpectedCaseCount {
		t.Fatalf("repair corpus has %d cases, want exactly %d", len(fixtures.Repairs.Cases), manifest.ExpectedCaseCount)
	}

	required := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" {
			t.Fatal("repair manifest contains an empty case name")
		}
		if _, exists := required[name]; exists {
			t.Fatalf("repair manifest repeats case name %q", name)
		}
		required[name] = struct{}{}
	}
	for _, repair := range fixtures.Repairs.Cases {
		if _, exists := required[repair.Name]; !exists {
			t.Fatalf("repair corpus contains unregistered case %q", repair.Name)
		}
		delete(required, repair.Name)
	}
	for name := range required {
		t.Fatalf("repair corpus is missing required case %q", name)
	}
}

// TestRewrittenCommit_FixtureContract drives every
// RewrittenCommit case in testdata/local-api/rewrites.yaml against the real
// RewrittenCommit.Validate, and confirms every member of the three closed
// sets it composes (RewriteResolution, RewriteMethod, Confidence) has at
// least one must-pass covering case.
func TestRewrittenCommit_FixtureContract(t *testing.T) {
	corpus := testcase.Corpus[schema.RewrittenCommit, bool]{Cases: loadRewriteFixtures(t).Cases}

	coveredResolutions := map[schema.RewriteResolution]bool{}
	coveredMethods := map[schema.RewriteMethod]bool{}
	coveredConfidences := map[schema.Confidence]bool{}

	for _, c := range corpus.Cases {
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
		coveredResolutions[c.Input.Resolution] = true
		coveredMethods[c.Input.Method] = true
		coveredConfidences[c.Input.Confidence] = true
	}

	for _, resolution := range schema.AllRewriteResolutions {
		if !coveredResolutions[resolution] {
			t.Errorf("RewriteResolution member %q has no must-pass fixture case", resolution)
		}
	}
	for _, method := range schema.AllRewriteMethods {
		if !coveredMethods[method] {
			t.Errorf("RewriteMethod member %q has no must-pass fixture case", method)
		}
	}
	for _, confidence := range schema.AllConfidences {
		if !coveredConfidences[confidence] {
			t.Errorf("Confidence member %q has no must-pass fixture case", confidence)
		}
	}
}

func cloneRewrittenCommit(input schema.RewrittenCommit) schema.RewrittenCommit {
	cloned := input
	if input.SessionIDs != nil {
		cloned.SessionIDs = append(make([]schema.SessionID, 0, len(input.SessionIDs)), input.SessionIDs...)
	}
	if input.SuccessorHash != nil {
		successor := *input.SuccessorHash
		cloned.SuccessorHash = &successor
	}
	return cloned
}

func applyRewriteRepair(input *schema.RewrittenCommit, mutation rewriteRepairMutation) error {
	switch mutation.Kind {
	case rewriteRepairClearSuccessorHash:
		if mutation.Input != "" {
			return fmt.Errorf("repair %q does not accept input, got %q", mutation.Kind, mutation.Input)
		}
		if input.SuccessorHash == nil {
			return fmt.Errorf("repair %q requires a successorHash", mutation.Kind)
		}
		input.SuccessorHash = nil
	case rewriteRepairAppendSessionID:
		if strings.TrimSpace(mutation.Input) == "" {
			return fmt.Errorf("repair %q requires a non-empty session ID input", mutation.Kind)
		}
		input.SessionIDs = append(input.SessionIDs, schema.SessionID(mutation.Input))
	default:
		return fmt.Errorf("unknown rewrite repair mutation %q", mutation.Kind)
	}
	return nil
}

// TestRewrittenCommit_RepairCorpus executes every fixture-declared repair
// through one mutation path. The exact manifest prevents a regression repair
// from disappearing or being replaced without an explicit fixture change.
func TestRewrittenCommit_RepairCorpus(t *testing.T) {
	fixtures := loadRewriteFixtures(t)
	sources := make(map[string]testcase.Case[schema.RewrittenCommit, bool], len(fixtures.Cases))
	for _, fixture := range fixtures.Cases {
		sources[fixture.Name] = fixture
	}

	for _, repair := range fixtures.Repairs.Cases {
		t.Run(repair.Name, func(t *testing.T) {
			source, exists := sources[repair.Input.SourceCase]
			if !exists {
				t.Fatalf("source case %q does not exist", repair.Input.SourceCase)
			}
			if source.Classification != testcase.MustFail || source.Expected {
				t.Fatalf("source case %q must describe a rejected input", repair.Input.SourceCase)
			}
			if repair.Classification != testcase.MustPass || !repair.Expected.PostMutationValid {
				t.Fatal("repair must be must-pass with postMutationValid=true")
			}
			if strings.TrimSpace(repair.Expected.OriginalErrorContains) == "" {
				t.Fatal("repair must declare originalErrorContains")
			}

			input := cloneRewrittenCommit(source.Input)
			originalErr := input.Validate()
			requireValidationErrorContains(t, originalErr, repair.Expected.OriginalErrorContains)
			requireActionableValidationError(t, originalErr)
			if err := applyRewriteRepair(&input, repair.Input.Mutation); err != nil {
				t.Fatalf("apply repair mutation: %v", err)
			}

			postMutationErr := input.Validate()
			if (postMutationErr == nil) != repair.Expected.PostMutationValid {
				t.Fatalf("post-mutation Validate() error=%v, want valid=%v", postMutationErr, repair.Expected.PostMutationValid)
			}
		})
	}
}
