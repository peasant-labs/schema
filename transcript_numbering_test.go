package schema_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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
	CapturedFrom string                     `yaml:"capturedFrom"`
	Versions     map[string]numberingCorpus `yaml:"versions"`
}

type numberingExpectation struct {
	Outcome       string   `yaml:"outcome"`
	Stage         string   `yaml:"stage"`
	ErrorContains []string `yaml:"errorContains"`
}

type numberingCase struct {
	testcase.Case[numberingInput, []numberingVerdict] `yaml:",inline"`
	Expect                                            numberingExpectation `yaml:"expect"`
}

type numberingCorpus struct {
	Cases []numberingCase `yaml:"cases"`
}

func cookFixtureTurns(t *testing.T, turns []schema.TurnDetail) []schema.CookedTurn {
	t.Helper()
	raw, err := json.Marshal(struct {
		Turns []schema.TurnDetail `json:"turns"`
	}{Turns: turns})
	if err != nil {
		t.Fatal(err)
	}
	input, err := schema.NewTranscriptCookTurns(raw)
	if err != nil {
		t.Fatal(err)
	}
	verdict, err := schema.CookTranscriptRows(input, schema.HarnessCodex)
	if err != nil {
		t.Fatal(err)
	}
	return verdict
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
	standard := testcase.Corpus[numberingInput, []numberingVerdict]{Cases: make([]testcase.Case[numberingInput, []numberingVerdict], len(corpus.Cases))}
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		standard.Cases[i] = c.Case
		names[i] = c.Name
		if c.Expect.Outcome != "cooked" && c.Expect.Outcome != "refused" {
			t.Fatalf("numbering case %q must declare a cooked or refused outcome", c.Name)
		}
		if (c.Classification == testcase.MustPass) != (c.Expect.Outcome == "cooked") {
			t.Fatalf("numbering case %q has inconsistent classification and outcome", c.Name)
		}
	}
	caseassert.RequireValid(t, standard)
	caseassert.RequireManifestNames(t, "transcript numbering", transcriptNumberingManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			detail := []byte(`{"turns":` + c.Input.TurnsJSON + `}`)
			turns, err := schema.NewTranscriptCookTurns(detail)
			if c.Expect.Stage == "constructor" {
				if err == nil || errors.Is(err, schema.ErrTranscriptUncookable) || turns != nil {
					t.Fatalf("constructor = %v, %v; want an ordinary error and no turns", turns, err)
				}
				for _, fragment := range c.Expect.ErrorContains {
					if !strings.Contains(err.Error(), fragment) {
						t.Fatalf("constructor error %v does not contain %q", err, fragment)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			actual, err := schema.CookTranscriptRows(turns, schema.Harness(c.Input.Harness))
			if c.Expect.Outcome == "refused" {
				if c.Expect.Stage != "cook" || !errors.Is(err, schema.ErrTranscriptUncookable) || actual != nil {
					t.Fatalf("cook = %#v, %v; want ErrTranscriptUncookable and no partial verdict", actual, err)
				}
				for _, fragment := range c.Expect.ErrorContains {
					if !strings.Contains(err.Error(), fragment) {
						t.Fatalf("cook error %v does not contain %q", err, fragment)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
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
