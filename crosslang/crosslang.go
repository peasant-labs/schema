// Package crosslang holds the single cross-language proof corpus: one
// schema/testcase corpus consumed byte-identically by both this module's Go test
// (via testcase.LoadCorpus) and the transcript-browser TypeScript side (a vendored
// copy parsed with yaml).
//
// The cross-language anchor is a committed source-hash. This Go side owns the
// source-of-truth corpus, the committed hex sha256 of its bytes, and the freshness
// gate that reddens if the corpus content changes without the hash refreshed. The
// TS side vendors a byte-identical copy plus the same hash and checks its copy
// against it. Both halves are per-repo and offline: no cross-repo CI, network, or
// auth. The durable typed accessor and cross-repo publishing are out of scope here
// (tracked as peasant#125).
package crosslang

import (
	"crypto/sha256"
	"encoding/hex"
)

//go:generate go run github.com/peasant-labs/schema/cmd/gen-source-hash

// ProofCorpusPath is the source-of-truth cross-language corpus, relative to this
// package directory (where go:generate and go test both run).
const ProofCorpusPath = "testdata/proof_corpus.yaml"

// SourceHashPath is the committed hex sha256 of the proof corpus bytes: the anchor
// the TS side compares its vendored copy against.
const SourceHashPath = "testdata/proof_corpus.sha256"

// ComputeSourceHash returns the lowercase hex sha256 of the corpus bytes. The
// generator (which writes SourceHashPath), the Go freshness gate, and the TS
// compare all hash the same corpus bytes, so the committed hash pins the content
// across languages.
func ComputeSourceHash(corpus []byte) string {
	sum := sha256.Sum256(corpus)
	return hex.EncodeToString(sum[:])
}
