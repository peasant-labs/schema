package licensecorpus_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/licensecorpus"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// loadCommitted reads and parses the committed generated corpus (relative to this
// package directory, where go test runs).
func loadCommitted(t *testing.T) testcase.Corpus[schema.License, bool] {
	t.Helper()
	data, err := os.ReadFile(licensecorpus.CommittedCorpusPath)
	if err != nil {
		t.Fatalf("read committed corpus %s: %v (regenerate with `go generate ./...`)", licensecorpus.CommittedCorpusPath, err)
	}
	corpus, err := testcase.LoadCorpus[schema.License, bool](data)
	if err != nil {
		t.Fatalf("load committed corpus: %v", err)
	}
	return corpus
}

// TestLicenseCorpus_ExhaustiveCoverage asserts the committed corpus covers every
// schema.AllLicenses member with a must-pass case and carries the must-fail
// negatives, and runs each case against the real License.IsValid so the corpus is
// a live enum-exhaustion test, not just a table. It reddens if AllLicenses widens
// without the corpus regenerated: a new member has no must-pass case in the
// committed file, so the coverage assertion (and the row-count floor) fail.
func TestLicenseCorpus_ExhaustiveCoverage(t *testing.T) {
	corpus := loadCommitted(t)

	// Non-vacuity: every generated case is fully populated (in-set classification,
	// enum provenance, non-empty ref and mutation description).
	if err := corpus.Validate(); err != nil {
		t.Fatalf("corpus has a vacuous case: %v", err)
	}

	// Run every case against the real SUT (License.IsValid) and record must-pass
	// coverage by input value.
	covered := map[schema.License]bool{}
	for _, c := range corpus.Cases {
		if got := c.Input.IsValid(); got != c.Expected {
			t.Errorf("case %q: License(%q).IsValid() = %v, want %v", c.Name, c.Input, got, c.Expected)
		}
		if c.Classification == testcase.MustPass {
			covered[c.Input] = true
		}
	}

	// Exhaustive coverage: every AllLicenses member has a must-pass case.
	for _, l := range schema.AllLicenses {
		if !covered[l] {
			t.Errorf("AllLicenses member %q has no must-pass case in the committed corpus; regenerate with `go generate ./...` after widening the menu", l)
		}
	}

	// The must-fail negatives are present: the non-menu id and the empty license.
	for _, neg := range []schema.License{schema.License("MIT"), schema.License("")} {
		found := false
		for _, c := range corpus.Cases {
			if c.Classification == testcase.MustFail && c.Input == neg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing must-fail negative for %q", neg)
		}
	}

	// Row-count floor (dogfoods the assert seam): at least one case per menu member
	// plus the two negatives. Widening the menu without regenerating drops below it.
	assert.RequireMin(t, corpus, len(schema.AllLicenses)+2)
}

// TestLicenseCorpus_Freshness asserts the committed corpus is byte-identical to a
// fresh render from schema.AllLicenses. It reddens on drift: either a hand-edited
// committed file, or a menu change that was not regenerated and committed. A
// generated artifact must not diverge from its generator.
func TestLicenseCorpus_Freshness(t *testing.T) {
	want, err := os.ReadFile(licensecorpus.CommittedCorpusPath)
	if err != nil {
		t.Fatalf("read committed corpus %s: %v", licensecorpus.CommittedCorpusPath, err)
	}
	got, err := licensecorpus.RenderCorpus()
	if err != nil {
		t.Fatalf("render corpus: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("STALE committed license corpus: %s drifted from a fresh render.\n"+
			"  what: the committed corpus does not match what the generator emits.\n"+
			"  why:  the menu changed without regenerating, or the file was hand-edited.\n"+
			"  fix:  run `go generate ./...` and commit licensecorpus/testdata/license_corpus.yaml.",
			licensecorpus.CommittedCorpusPath)
	}
}
