// Command gen-source-hash writes the committed source-hash of the cross-language
// proof corpus. It is invoked by the go:generate directive in the crosslang
// package (which sets the working directory there), so it reads and writes paths
// relative to that package. Refresh with `go generate ./...` after an intentional
// change to the proof corpus, then commit the result; the freshness gate fails if
// the committed hash is stale.
package main

import (
	"log"
	"os"

	"github.com/peasant-labs/schema/crosslang"
)

func main() {
	corpus, err := os.ReadFile(crosslang.ProofCorpusPath)
	if err != nil {
		log.Fatalf("gen-source-hash: read %s: %v", crosslang.ProofCorpusPath, err)
	}
	line := crosslang.ComputeSourceHash(corpus) + "\n"
	if err := os.WriteFile(crosslang.SourceHashPath, []byte(line), 0o644); err != nil {
		log.Fatalf("gen-source-hash: write %s: %v", crosslang.SourceHashPath, err)
	}
}
