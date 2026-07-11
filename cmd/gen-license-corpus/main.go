// Command gen-license-corpus writes the enum-exhaustion license corpus to the
// committed artifact. It is invoked by the go:generate directive in the
// licensecorpus package (which sets the working directory to that package), so it
// writes CommittedCorpusPath relative to the package directory.
//
// Regenerate with `go generate ./...` (or `go run ./cmd/gen-license-corpus` from
// the licensecorpus directory) whenever schema.AllLicenses changes, then commit
// the result. The freshness gate fails if the committed file drifts from a fresh
// render.
package main

import (
	"log"
	"os"

	"github.com/peasant-labs/schema/licensecorpus"
)

func main() {
	data, err := licensecorpus.RenderCorpus()
	if err != nil {
		log.Fatalf("gen-license-corpus: render corpus: %v", err)
	}
	if err := os.WriteFile(licensecorpus.CommittedCorpusPath, data, 0o644); err != nil {
		log.Fatalf("gen-license-corpus: write %s: %v", licensecorpus.CommittedCorpusPath, err)
	}
}
