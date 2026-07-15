package testcase_test

// Dogfooded non-vacuity tests for the test-case corpus tooling. These tests use
// github.com/peasant-labs/schema/testcase to carry their OWN inputs, complete
// with provenance + mutation metadata and a minimum-size guard, so the discipline
// the corpus enforces on every other suite is enforced on the corpus itself. They
// live in an external test package (and import testcase/assert) so the pure-data
// package stays testing-free.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/example.yaml
var exampleCorpusYAML []byte

//go:embed testdata/load_cases.yaml
var loadCasesYAML []byte

//go:embed testdata/check_min_cases.yaml
var checkMinCasesYAML []byte

//go:embed testdata/validate_cases.yaml
var validateCasesYAML []byte

type checkMinInput struct {
	Size    int `yaml:"size"`
	Minimum int `yaml:"minimum"`
}

// validSubject returns a fully-populated Case that Case.Validate must accept: the
// baseline the field-failure cases below each mutate one field of.
func validSubject() testcase.Case[string, bool] {
	return testcase.Case[string, bool]{
		Name:           "subject",
		Input:          "x",
		Expected:       true,
		Classification: testcase.MustPass,
		Provenance:     testcase.Provenance{Source: testcase.SourceManual, Ref: "a case with all four field-groups populated"},
		Mutation:       testcase.Mutation{Description: "the unmodified valid subject case"},
	}
}

type validateMutation string

const (
	validateUnmodified              validateMutation = "unmodified"
	validateEmptyProvenanceRef      validateMutation = "empty-provenance-ref"
	validateEmptyMutationDesc       validateMutation = "empty-mutation-description"
	validateUnknownClassification   validateMutation = "unknown-classification"
	validateUnknownProvenanceSource validateMutation = "unknown-provenance-source"
)

type validateInput struct {
	Mutation validateMutation `yaml:"mutation"`
}

func subjectForMutation(t *testing.T, mutation validateMutation) testcase.Case[string, bool] {
	t.Helper()
	subject := validSubject()
	switch mutation {
	case validateUnmodified:
	case validateEmptyProvenanceRef:
		subject.Provenance.Ref = ""
	case validateEmptyMutationDesc:
		subject.Mutation.Description = ""
	case validateUnknownClassification:
		subject.Classification = "sometimes"
	case validateUnknownProvenanceSource:
		subject.Provenance.Source = "folklore"
	default:
		t.Fatalf("validate fixture selected unknown mutation %q", mutation)
	}
	return subject
}

// TestCheckMin_NegativeControl proves the minimum-size floor actually fires: a
// corpus below the floor errors, a corpus at or above it does not.
func TestCheckMin_NegativeControl(t *testing.T) {
	matrix, err := testcase.LoadCorpus[checkMinInput, bool](checkMinCasesYAML)
	if err != nil {
		t.Fatalf("load minimum matrix: %v", err)
	}
	assert.RequireMin(t, matrix, 4)
	assert.RequireValid(t, matrix)
	for _, tc := range matrix.Cases {
		corpus := testcase.Corpus[string, bool]{Cases: make([]testcase.Case[string, bool], tc.Input.Size)}
		if got := corpus.CheckMin(tc.Input.Minimum) == nil; got != tc.Expected {
			t.Errorf("%s: accepted=%v, want %v", tc.Name, got, tc.Expected)
		}
	}
}

// TestCaseValidate_PopulatedPassesEmptyFieldFails is the positive+negative
// control on Case.Validate, dogfooded: the five subject cases are carried as a
// testcase.Corpus that itself must pass RequireMin and Corpus.Validate.
func TestCaseValidate_PopulatedPassesEmptyFieldFails(t *testing.T) {
	corpus, err := testcase.LoadCorpus[validateInput, bool](validateCasesYAML)
	if err != nil {
		t.Fatalf("load validation matrix: %v", err)
	}

	// Row-count guard (dogfood): the one positive and all four field-failures.
	assert.RequireMin(t, corpus, 5)
	if len(corpus.Cases) != 5 {
		t.Fatalf("validation matrix has %d rows, want exactly 5", len(corpus.Cases))
	}
	// The dogfood corpus is itself non-vacuous.
	if err := corpus.Validate(); err != nil {
		t.Fatalf("dogfood corpus is under-populated: %v", err)
	}

	for _, mc := range corpus.Cases {
		// The case's own classification must agree with its expected outcome.
		if want := mc.Classification == testcase.MustPass; want != mc.Expected {
			t.Errorf("%s: classification %q disagrees with expected=%v", mc.Name, mc.Classification, mc.Expected)
		}
		gotOK := subjectForMutation(t, mc.Input.Mutation).Validate() == nil
		if gotOK != mc.Expected {
			t.Errorf("%s: Case.Validate accepted=%v, want %v", mc.Name, gotOK, mc.Expected)
		}
	}
}

// TestCorpusFixture_AllFourCriteriaPopulated exercises the all-four criterion end
// to end: a fully-populated YAML fixture loads through LoadCorpus, clears
// Corpus.Validate, and satisfies the size guard; and a corpus with one
// under-populated case is rejected.
func TestCorpusFixture_AllFourCriteriaPopulated(t *testing.T) {
	corpus, err := testcase.LoadCorpus[string, bool](exampleCorpusYAML)
	if err != nil {
		t.Fatalf("LoadCorpus(testdata/example.yaml): %v", err)
	}

	assert.RequireMin(t, corpus, 2)
	if err := corpus.Validate(); err != nil {
		t.Fatalf("the fixture corpus is under-populated: %v", err)
	}

	// Positively confirm every field-group is non-vacuous on each loaded case.
	for _, c := range corpus.Cases {
		if !c.Classification.IsValid() {
			t.Errorf("%s: classification %q not in-set", c.Name, c.Classification)
		}
		if !c.Provenance.Source.IsValid() {
			t.Errorf("%s: provenance source %q not in-set", c.Name, c.Provenance.Source)
		}
		if c.Provenance.Ref == "" {
			t.Errorf("%s: empty provenance ref", c.Name)
		}
		if c.Mutation.Description == "" {
			t.Errorf("%s: empty mutation description", c.Name)
		}
	}

	// Negative: appending one under-populated case (empty ref) must fail Validate.
	bad := testcase.Corpus[string, bool]{Cases: append(append([]testcase.Case[string, bool](nil), corpus.Cases...), testcase.Case[string, bool]{
		Name:           "vacuous case",
		Classification: testcase.MustPass,
		Provenance:     testcase.Provenance{Source: testcase.SourceManual, Ref: ""},
		Mutation:       testcase.Mutation{Description: "an empty-ref case that Validate must reject"},
	})}
	if err := bad.Validate(); err == nil {
		t.Fatal("Corpus.Validate accepted a corpus containing an empty provenance ref")
	}
}

func TestLoadCorpus_StrictSharedCases(t *testing.T) {
	matrix, err := testcase.LoadCorpus[string, bool](loadCasesYAML)
	if err != nil {
		t.Fatalf("LoadCorpus(testdata/load_cases.yaml): %v", err)
	}
	assert.RequireMin(t, matrix, 14)
	assert.RequireValid(t, matrix)
	if len(matrix.Cases) != 14 {
		t.Fatalf("strict loader matrix has %d cases, want 14", len(matrix.Cases))
	}

	for _, testCase := range matrix.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			loaded, err := testcase.LoadCorpus[string, bool]([]byte(testCase.Input))
			if err == nil {
				err = loaded.Validate()
			}
			if got := err == nil; got != testCase.Expected {
				t.Fatalf("accepted=%v, want %v (error: %v)", got, testCase.Expected, err)
			}
		})
	}
}
