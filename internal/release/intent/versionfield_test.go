package intent_test

// Version-field suite: drives intent.ReadVersionField and
// intent.WriteVersionField over the version-field corpus. Writes must be
// byte-exact outside the value; failures must be plain six-part errors.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/versionfield_corpus.yaml
var versionFieldCorpusYAML []byte

type versionFieldInput struct {
	Op       string `yaml:"op"`
	Document string `yaml:"document"`
	Field    string `yaml:"field"`
	Value    string `yaml:"value"`
}

type versionFieldExpected struct {
	Output string `yaml:"output"`
}

func TestVersionField(t *testing.T) {
	corpus := loadNamedCorpus[versionFieldInput, versionFieldExpected](t, versionFieldCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			switch c.Input.Op {
			case "read":
				got, err := intent.ReadVersionField([]byte(c.Input.Document), c.Input.Field)
				if c.Classification == "must-fail" {
					if err == nil {
						t.Fatalf("ReadVersionField: expected an error, got %q", got)
					}
					requireSixLabels(t, err)
					return
				}
				if err != nil {
					t.Fatalf("ReadVersionField: unexpected error: %v", err)
				}
				if got != c.Expected.Output {
					t.Errorf("ReadVersionField = %q, want %q", got, c.Expected.Output)
				}
			case "write":
				got, err := intent.WriteVersionField([]byte(c.Input.Document), c.Input.Field, c.Input.Value)
				if c.Classification == "must-fail" {
					if err == nil {
						t.Fatalf("WriteVersionField: expected an error, got %q", got)
					}
					requireSixLabels(t, err)
					return
				}
				if err != nil {
					t.Fatalf("WriteVersionField: unexpected error: %v", err)
				}
				if string(got) != c.Expected.Output {
					t.Errorf("WriteVersionField = %q, want %q", got, c.Expected.Output)
				}
			default:
				t.Fatalf("unknown op %q: want read or write", c.Input.Op)
			}
		})
	}
}
