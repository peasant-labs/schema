package intent_test

// Highest-final-tag suite: drives intent.HighestFinalTag over the tag corpus.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/highest_tag_corpus.yaml
var highestTagCorpusYAML []byte

type highestTagInput struct {
	Tags   map[string]string `yaml:"tags"`
	Prefix string            `yaml:"prefix"`
}

type highestTagExpected struct {
	Tag     string `yaml:"tag"`
	Version string `yaml:"version"`
	Found   bool   `yaml:"found"`
}

func TestHighestFinalTag(t *testing.T) {
	corpus := loadNamedCorpus[highestTagInput, highestTagExpected](t, highestTagCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			tag, version, found := intent.HighestFinalTag(c.Input.Tags, c.Input.Prefix)
			if found != c.Expected.Found {
				t.Fatalf("Found = %v, want %v", found, c.Expected.Found)
			}
			if tag != c.Expected.Tag {
				t.Errorf("Tag = %q, want %q", tag, c.Expected.Tag)
			}
			if string(version) != c.Expected.Version {
				t.Errorf("Version = %q, want %q", version, c.Expected.Version)
			}
		})
	}
}
