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

// mutatedSubject applies one field mutation to a fresh valid subject.
func mutatedSubject(mutate func(*testcase.Case[string, bool])) testcase.Case[string, bool] {
	c := validSubject()
	mutate(&c)
	return c
}

// validateCase is a dogfooded case ABOUT Case.Validate: its Input is the subject
// case under test, its Expected is whether Validate should accept that subject.
type validateCase = testcase.Case[testcase.Case[string, bool], bool]

// TestCheckMin_NegativeControl proves the minimum-size floor actually fires: a
// corpus below the floor errors, a corpus at or above it does not.
func TestCheckMin_NegativeControl(t *testing.T) {
	one := testcase.Corpus[string, bool]{Cases: make([]testcase.Case[string, bool], 1)}
	if err := one.CheckMin(2); err == nil {
		t.Fatal("CheckMin(2) on a 1-case corpus returned nil; the size floor does not fire")
	}
	if err := one.CheckMin(1); err != nil {
		t.Errorf("CheckMin(1) on a 1-case corpus errored: %v", err)
	}

	var empty testcase.Corpus[string, bool]
	if err := empty.CheckMin(1); err == nil {
		t.Fatal("CheckMin(1) on an empty corpus returned nil; the size floor does not fire")
	}

	// The loud wrapper agrees with the pure guard on the satisfied path.
	assert.RequireMin(t, one, 1)
}

// TestCaseValidate_PopulatedPassesEmptyFieldFails is the positive+negative
// control on Case.Validate, dogfooded: the five subject cases are carried as a
// testcase.Corpus that itself must pass RequireMin and Corpus.Validate.
func TestCaseValidate_PopulatedPassesEmptyFieldFails(t *testing.T) {
	corpus := testcase.Corpus[testcase.Case[string, bool], bool]{Cases: []validateCase{
		{
			Name:           "fully populated case passes",
			Input:          validSubject(),
			Expected:       true,
			Classification: testcase.MustPass,
			Provenance:     testcase.Provenance{Source: testcase.SourceManual, Ref: "a case with all four field-groups populated must pass Validate"},
			Mutation:       testcase.Mutation{Description: "the unmodified valid subject case"},
		},
		{
			Name:           "empty provenance ref fails",
			Input:          mutatedSubject(func(c *testcase.Case[string, bool]) { c.Provenance.Ref = "" }),
			Expected:       false,
			Classification: testcase.MustFail,
			Provenance:     testcase.Provenance{Source: testcase.SourceBoundary, Ref: "Validate must reject a case that cites no reason to exist"},
			Mutation:       testcase.Mutation{Description: "cleared the subject's provenance ref"},
		},
		{
			Name:           "empty mutation description fails",
			Input:          mutatedSubject(func(c *testcase.Case[string, bool]) { c.Mutation.Description = "" }),
			Expected:       false,
			Classification: testcase.MustFail,
			Provenance:     testcase.Provenance{Source: testcase.SourceBoundary, Ref: "Validate must reject a case with no change under test"},
			Mutation:       testcase.Mutation{Description: "cleared the subject's mutation description"},
		},
		{
			Name:           "out-of-set classification fails",
			Input:          mutatedSubject(func(c *testcase.Case[string, bool]) { c.Classification = "sometimes" }),
			Expected:       false,
			Classification: testcase.MustFail,
			Provenance:     testcase.Provenance{Source: testcase.SourceBoundary, Ref: "Validate must reject a classification outside must-pass/must-fail"},
			Mutation:       testcase.Mutation{Description: "set the classification off the closed set"},
		},
		{
			Name:           "invalid provenance source fails",
			Input:          mutatedSubject(func(c *testcase.Case[string, bool]) { c.Provenance.Source = "folklore" }),
			Expected:       false,
			Classification: testcase.MustFail,
			Provenance:     testcase.Provenance{Source: testcase.SourceBoundary, Ref: "Validate must reject a provenance source outside the closed set"},
			Mutation:       testcase.Mutation{Description: "set the provenance source off the closed set"},
		},
	}}

	// Row-count guard (dogfood): the one positive and all four field-failures.
	assert.RequireMin(t, corpus, 5)
	// The dogfood corpus is itself non-vacuous.
	if err := corpus.Validate(); err != nil {
		t.Fatalf("dogfood corpus is under-populated: %v", err)
	}

	for _, mc := range corpus.Cases {
		// The case's own classification must agree with its expected outcome.
		if want := mc.Classification == testcase.MustPass; want != mc.Expected {
			t.Errorf("%s: classification %q disagrees with expected=%v", mc.Name, mc.Classification, mc.Expected)
		}
		gotOK := mc.Input.Validate() == nil
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
	assert.RequireMin(t, matrix, 12)
	assert.RequireValid(t, matrix)
	if len(matrix.Cases) != 12 {
		t.Fatalf("strict loader matrix has %d cases, want 12", len(matrix.Cases))
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
