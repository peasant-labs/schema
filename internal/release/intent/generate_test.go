package intent_test

// Generator suite: drives intent.Generate over the generate corpus.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/generate_corpus.yaml
var generateCorpusYAML []byte

type generateInput struct {
	Config        intent.Config  `yaml:"config"`
	Pending       []intent.Title `yaml:"pending"`
	BaseChangelog string         `yaml:"base_changelog"`
	Version       string         `yaml:"version"`
	Tag           string         `yaml:"tag"`
	Date          string         `yaml:"date"`
}

type generateExpected struct {
	Changelog string            `yaml:"changelog"`
	Section   string            `yaml:"section"`
	Versions  map[string]string `yaml:"versions"`
}

func TestGenerate(t *testing.T) {
	corpus := loadNamedCorpus[generateInput, generateExpected](t, generateCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			in := intent.ContentInput{
				Pending:       c.Input.Pending,
				BaseChangelog: c.Input.BaseChangelog,
				Version:       intent.SemVer(c.Input.Version),
				Tag:           c.Input.Tag,
				Date:          c.Input.Date,
			}
			got, err := intent.Generate(c.Input.Config, in)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("Generate: expected an error, got %+v", got)
				}
				requireSixLabels(t, err)
				return
			}
			if err != nil {
				t.Fatalf("Generate: unexpected error: %v", err)
			}
			if got.Changelog != c.Expected.Changelog {
				t.Errorf("Changelog = %q, want %q", got.Changelog, c.Expected.Changelog)
			}
			if got.Section != c.Expected.Section {
				t.Errorf("Section = %q, want %q", got.Section, c.Expected.Section)
			}
			if len(got.Versions) != len(c.Expected.Versions) {
				t.Fatalf("Versions = %v, want %v", got.Versions, c.Expected.Versions)
			}
			for path, want := range c.Expected.Versions {
				if got.Versions[path] != want {
					t.Errorf("Versions[%q] = %q, want %q", path, got.Versions[path], want)
				}
			}
		})
	}
}
