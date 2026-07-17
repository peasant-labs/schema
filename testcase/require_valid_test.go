package testcase_test

// Positive and negative controls for assert.RequireValid, the loud *testing.T
// wrapper around the pure Corpus.Validate. They are symmetric to the RequireMin /
// CheckMin controls in corpus_test.go: RequireValid passes on a fully-populated
// fixture, and the pure Corpus.Validate it wraps rejects a deliberately vacuous
// fixture (so RequireValid reddens on that fixture). The cases live in testdata
// YAML fixtures, never inline.

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/vacuous_corpus.yaml
var vacuousCorpusYAML []byte

// TestRequireValid_PopulatedFixturePasses is the positive control: RequireValid
// accepts a fully-populated corpus loaded from a fixture (the same example corpus
// the other controls use), and does not fail the test.
func TestRequireValid_PopulatedFixturePasses(t *testing.T) {
	corpus, err := testcase.LoadCorpus[string, bool](exampleCorpusYAML)
	if err != nil {
		t.Fatalf("LoadCorpus(testdata/example.yaml): %v", err)
	}
	assert.RequireMin(t, corpus, 1)
	assert.RequireValid(t, corpus)
}

// TestRequireValid_VacuousFixtureRejected is the fixture-driven negative control.
// RequireValid is exactly t.Fatalf(Corpus.Validate error), so proving the pure
// seam rejects the vacuous fixture proves RequireValid reddens on it, without
// aborting this suite (the same pure/wrapper split TestCheckMin_NegativeControl
// uses). Each fixture case is vacuous in one distinct way, so each must fail
// Case.Validate and the whole corpus must fail Corpus.Validate.
func TestRequireValid_VacuousFixtureRejected(t *testing.T) {
	// Decode directly to exercise the standalone validator used by programmatic
	// corpora. LoadCorpus intentionally validates at its trust boundary and would
	// reject this negative-control document before the standalone seam is reached.
	var corpus testcase.Corpus[string, bool]
	decoder := yaml.NewDecoder(bytes.NewReader(vacuousCorpusYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&corpus); err != nil {
		t.Fatalf("decode testdata/vacuous_corpus.yaml: %v", err)
	}
	// Four distinct vacuity kinds: out-of-set classification, out-of-set
	// provenance source, empty ref, empty mutation description.
	if got := len(corpus.Cases); got != 4 {
		t.Fatalf("vacuous fixture has %d cases, want the four vacuity kinds", got)
	}
	for _, c := range corpus.Cases {
		if err := c.Validate(); err == nil {
			t.Errorf("case %q: Case.Validate accepted a vacuous case; RequireValid would not redden on it", c.Name)
		}
	}
	// The pure seam RequireValid wraps rejects the corpus, so RequireValid(t,
	// corpus) would t.Fatalf here. Verified green->red->green by pointing
	// RequireValid at this fixture; the committed control asserts the seam so the
	// suite stays green.
	if err := corpus.Validate(); err == nil {
		t.Fatal("Corpus.Validate accepted the vacuous fixture; RequireValid would not fire")
	}
}
