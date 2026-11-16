package schema_test

import (
	_ "embed"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
)

//go:embed testdata/typescript/wire_integer_compatibility.yaml
var wireIntegerCompatibilityYAML []byte

type wireIntegerInput struct {
	Target string `yaml:"target"`
	JSON   string `yaml:"json"`
	Path   string `yaml:"path"`
	Bigint bool   `yaml:"bigint"`
}

type wireIntegerExpected struct {
	Valid          bool   `yaml:"valid"`
	Start          string `yaml:"start"`
	End            string `yaml:"end"`
	NumberPosition bool   `yaml:"numberPosition"`
}

func TestDurableReadIntegerCompatibility(t *testing.T) {
	corpus, err := testcase.LoadCorpus[wireIntegerInput, wireIntegerExpected](wireIntegerCompatibilityYAML)
	if err != nil {
		t.Fatal(err)
	}
	foundValid, foundInvalid := false, false
	for _, c := range corpus.Cases {
		if c.Input.Target != "durable" {
			continue
		}
		if c.Expected.Valid != (c.Classification == testcase.MustPass) {
			t.Fatalf("%s: classification does not match expected verdict", c.Name)
		}
		foundValid = foundValid || c.Expected.Valid
		foundInvalid = foundInvalid || !c.Expected.Valid
		t.Run(c.Name, func(t *testing.T) {
			_, detailErr := schema.DecodeSessionDetailPayloadRaw([]byte(c.Input.JSON))
			envelope := `{"contractVersion":"0.1.1","kind":"session_detail","sessionDetail":` + c.Input.JSON + `}`
			_, envelopeErr := schema.DecodeTranscriptContentRaw([]byte(envelope))
			if (detailErr == nil) != c.Expected.Valid || (envelopeErr == nil) != c.Expected.Valid {
				t.Fatalf("durable detail error = %v; envelope error = %v; want valid=%v", detailErr, envelopeErr, c.Expected.Valid)
			}
		})
	}
	if !foundValid || !foundInvalid {
		t.Fatal("integer compatibility corpus must exercise accepted and refused durable reads")
	}
}
