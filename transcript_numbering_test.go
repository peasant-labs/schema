package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_numbering.yaml
var transcriptNumberingYAML []byte

//go:embed testdata/transcript_numbering_manifest.yaml
var transcriptNumberingManifestYAML []byte

type numberingInput struct {
	Harness   string `yaml:"harness"`
	TurnsJSON string `yaml:"turnsJSON"`
}

type numberingVerdict struct {
	Position       int    `yaml:"position"`
	Keep           bool   `yaml:"keep"`
	ServedPosition int    `yaml:"servedPosition"`
	DisplayNumber  string `yaml:"displayNumber"`
}

type numberingFixture struct {
	CapturedFrom string                                                         `yaml:"capturedFrom"`
	Versions     map[string]testcase.Corpus[numberingInput, []numberingVerdict] `yaml:"versions"`
}

func TestCookTranscriptRows(t *testing.T) {
	fixture := decodeStrictYAML[numberingFixture](t, transcriptNumberingYAML, "transcript numbering fixture")
	if fixture.CapturedFrom != "bb908530f72bbdc44e4bc6c85582afe3668ffa74" {
		t.Fatalf("numbering fixture capture = %q; preserve the pre-change browser oracle", fixture.CapturedFrom)
	}
	corpus, ok := fixture.Versions[schema.CutVersion]
	if !ok {
		t.Fatalf("numbering fixture lacks expectations for cut version %q", schema.CutVersion)
	}
	if _, ok := fixture.Versions["c1"]; !ok {
		t.Fatal("numbering fixture lost the original version's expectations")
	}
	caseassert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "transcript numbering", transcriptNumberingManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var turns []schema.TurnDetail
			if err := json.Unmarshal([]byte(c.Input.TurnsJSON), &turns); err != nil {
				t.Fatal(err)
			}
			actual := schema.CookTranscriptRows(turns, schema.Harness(c.Input.Harness))
			got := make([]numberingVerdict, len(actual))
			for i, row := range actual {
				got[i] = numberingVerdict{row.Position, row.Keep, row.ServedPosition, row.DisplayNumber}
			}
			if !reflect.DeepEqual(got, c.Expected) {
				t.Fatalf("cooked rows = %#v, want %#v", got, c.Expected)
			}
		})
	}
}
