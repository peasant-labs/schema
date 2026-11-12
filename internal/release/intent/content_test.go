package intent_test

// Content-comparison suite: drives intent.CompareContent over the content
// corpus. A mismatch must be the typed stale-content rejection; equality
// must be nil.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/content_corpus.yaml
var contentCorpusYAML []byte

type contentInput struct {
	Want intent.Content `yaml:"want"`
	Got  intent.Content `yaml:"got"`
}

type contentExpected struct {
	Code string `yaml:"code"`
}

func TestCompareContent(t *testing.T) {
	corpus := loadNamedCorpus[contentInput, contentExpected](t, contentCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			err := intent.CompareContent(c.Input.Want, c.Input.Got)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("CompareContent: expected an error, got nil")
				}
				requireSixParts(t, err, c.Expected.Code)
				return
			}
			if err != nil {
				t.Fatalf("CompareContent: unexpected error: %v", err)
			}
		})
	}
}
