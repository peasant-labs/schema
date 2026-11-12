package intent_test

// Protected-path suite: drives intent.CheckProtectedPaths over the
// protected corpus.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/protected_corpus.yaml
var protectedCorpusYAML []byte

type protectedInput struct {
	Config      intent.Config            `yaml:"config"`
	IsReleasePR bool                     `yaml:"is_release_pr"`
	Changes     []intent.ProtectedChange `yaml:"changes"`
}

type protectedExpected struct {
	Code string `yaml:"code"`
}

func TestCheckProtectedPaths(t *testing.T) {
	corpus := loadNamedCorpus[protectedInput, protectedExpected](t, protectedCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			err := intent.CheckProtectedPaths(c.Input.Config, c.Input.Changes, c.Input.IsReleasePR)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("CheckProtectedPaths: expected an error, got nil")
				}
				requireSixParts(t, err, c.Expected.Code)
				return
			}
			if err != nil {
				t.Fatalf("CheckProtectedPaths: unexpected error: %v", err)
			}
		})
	}
}
