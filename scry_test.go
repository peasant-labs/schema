package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/integrity/scry_turn_bounds.yaml
var scryTurnBoundsYAML []byte

//go:embed testdata/integrity/transcript_integrity.yaml
var transcriptIntegrityYAML []byte

type scryTurnBoundsInput struct {
	TurnCount int `yaml:"turn_count"`
	Response  any `yaml:"response"`
}

type scryTurnBoundsExpected struct {
	Valid bool `yaml:"valid"`
}

func TestScryVerdictIsClosed(t *testing.T) {
	var want []string
	for _, enum := range loadClosedEnumFixture(t).Enums {
		if enum.Name == "ScryVerdict" {
			want = enum.AllValues
		}
	}
	if len(want) == 0 {
		t.Fatal("enums.yaml does not define ScryVerdict")
	}
	got := make([]string, 0, len(schema.AllScryVerdicts))
	for _, verdict := range schema.AllScryVerdicts {
		if !verdict.IsValid() {
			t.Fatalf("%q is listed but not valid", verdict)
		}
		got = append(got, verdict.String())
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AllScryVerdicts = %v, want %v", got, want)
	}
	if schema.ScryVerdict("flagged").IsValid() || schema.ScryVerdict("").IsValid() {
		t.Fatal("a verdict outside the closed set is valid")
	}
}

func TestScryCheckResponseValidateBoundsTurns(t *testing.T) {
	corpus, err := testcase.LoadCorpus[scryTurnBoundsInput, scryTurnBoundsExpected](scryTurnBoundsYAML)
	if err != nil {
		t.Fatal(err)
	}
	assertcase.RequireMin(t, corpus, 5)
	assertcase.RequireValid(t, corpus)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			raw, err := json.Marshal(c.Input.Response)
			if err != nil {
				t.Fatal(err)
			}
			var response schema.ScryCheckResponse
			if err := json.Unmarshal(raw, &response); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := response.Validate(c.Input.TurnCount) == nil; got != c.Expected.Valid {
				t.Fatalf("Validate(%d) valid=%v, want %v", c.Input.TurnCount, got, c.Expected.Valid)
			}
		})
	}
}

func TestScryCheckResponseRoundTripIsByteStable(t *testing.T) {
	fixture := loadBoundaryFixture(t, transcriptIntegrityYAML)
	for _, c := range fixture.Cases {
		if c.Name != "scry-response-checked-flagged-valid" {
			continue
		}
		raw, err := json.Marshal(c.Input.Payload)
		if err != nil {
			t.Fatal(err)
		}
		var response schema.ScryCheckResponse
		if err := json.Unmarshal(raw, &response); err != nil {
			t.Fatal(err)
		}
		out, err := json.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		const want = `{"scryVersion":"1.0.0","verdict":"checked","flaggedTurns":[0,1]}`
		if string(out) != want {
			t.Fatalf("marshalled %s, want %s", out, want)
		}
		return
	}
	t.Fatal("scry-response-checked-flagged-valid is not in the corpus")
}
