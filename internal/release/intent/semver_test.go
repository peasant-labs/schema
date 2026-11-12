package intent_test

// Version arithmetic and SemVer parsing suites.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/arithmetic_corpus.yaml
var arithmeticCorpusYAML []byte

//go:embed testdata/semver_corpus.yaml
var semverCorpusYAML []byte

type arithmeticInput struct {
	Base     string `yaml:"base"`
	Level    string `yaml:"level"`
	Breaking bool   `yaml:"breaking"`
}

type arithmeticExpected struct {
	Next string `yaml:"next"`
}

type semverInput struct {
	Raw string `yaml:"raw"`
}

type semverExpected struct {
	Value string `yaml:"value"`
}

func TestNextVersion(t *testing.T) {
	corpus := loadNamedCorpus[arithmeticInput, arithmeticExpected](t, arithmeticCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := intent.NextVersion(intent.SemVer(c.Input.Base), testLevel(c.Input.Level), c.Input.Breaking)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("NextVersion(%q): expected an error, got %q", c.Input.Base, got)
				}
				requireSixLabels(t, err)
				return
			}
			if err != nil {
				t.Fatalf("NextVersion(%q): unexpected error: %v", c.Input.Base, err)
			}
			if string(got) != c.Expected.Next {
				t.Errorf("NextVersion = %q, want %q", got, c.Expected.Next)
			}
		})
	}
}

func TestParseSemVer(t *testing.T) {
	corpus := loadNamedCorpus[semverInput, semverExpected](t, semverCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := intent.ParseSemVer(c.Input.Raw)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("ParseSemVer(%q): expected an error, got %q", c.Input.Raw, got)
				}
				requireSixLabels(t, err)
				return
			}
			if err != nil {
				t.Fatalf("ParseSemVer(%q): unexpected error: %v", c.Input.Raw, err)
			}
			if string(got) != c.Expected.Value {
				t.Errorf("ParseSemVer = %q, want %q", got, c.Expected.Value)
			}
		})
	}
}

func testLevel(word string) intent.Level {
	switch word {
	case "patch":
		return intent.LevelPatch
	case "minor":
		return intent.LevelMinor
	case "major":
		return intent.LevelMajor
	default:
		return intent.LevelNone
	}
}
