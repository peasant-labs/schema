package crosslang_test

import (
	"os"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/crosslang"
	"github.com/peasant-labs/schema/testcase"
)

// TestProofCorpus_GoConsumes proves the Go side consumes the proof corpus: it
// loads the source-of-truth file through the shared schema/testcase loader,
// validates every case is fully populated, and runs each case against the trivial
// "non-empty string is accepted" rule so the expected outcomes are self-consistent.
// The TS side runs the same rule over a byte-identical vendored copy, so a
// divergence in either language surfaces.
func TestProofCorpus_GoConsumes(t *testing.T) {
	data, err := os.ReadFile(crosslang.ProofCorpusPath)
	if err != nil {
		t.Fatalf("read proof corpus %s: %v", crosslang.ProofCorpusPath, err)
	}
	corpus, err := testcase.LoadCorpus[string, bool](data)
	if err != nil {
		t.Fatalf("load proof corpus: %v", err)
	}
	if err := corpus.Validate(); err != nil {
		t.Fatalf("proof corpus has a vacuous case: %v", err)
	}
	if err := corpus.CheckMin(1); err != nil {
		t.Fatalf("proof corpus is empty: %v", err)
	}
	for _, c := range corpus.Cases {
		accepted := len(c.Input) > 0
		if accepted != c.Expected {
			t.Errorf("case %q: (len(input)>0)=%v, want expected=%v", c.Name, accepted, c.Expected)
		}
	}
}

// TestProofCorpus_SourceHashFresh is the schema-side half of the committed
// source-hash anchor: the committed source-hash must be the sha256 of the proof
// corpus bytes. It reddens if
// the corpus content changes without the hash refreshed, so a source drift cannot
// silently break the cross-language anchor the TS side compares against.
func TestProofCorpus_SourceHashFresh(t *testing.T) {
	data, err := os.ReadFile(crosslang.ProofCorpusPath)
	if err != nil {
		t.Fatalf("read proof corpus %s: %v", crosslang.ProofCorpusPath, err)
	}
	committed, err := os.ReadFile(crosslang.SourceHashPath)
	if err != nil {
		t.Fatalf("read committed source-hash %s: %v", crosslang.SourceHashPath, err)
	}
	want := strings.TrimSpace(string(committed))
	got := crosslang.ComputeSourceHash(data)
	if got != want {
		t.Errorf("STALE source-hash: %s does not match the proof corpus.\n"+
			"  what: the committed hash is not the sha256 of %s.\n"+
			"  why:  the corpus content changed without the hash refreshed.\n"+
			"  fix:  run `go generate ./...` and commit %s.",
			crosslang.SourceHashPath, crosslang.ProofCorpusPath, crosslang.SourceHashPath)
	}
}
