package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_entry_locate.yaml
var entryLocateYAML []byte

//go:embed testdata/transcript_entry_locate_manifest.yaml
var entryLocateManifestYAML []byte

type entryLocateInput struct {
	Detail     string   `yaml:"detail"`
	Harness    string   `yaml:"harness"`
	References []string `yaml:"references"`
	Subset     []int    `yaml:"subset"`
	Protect    bool     `yaml:"protect"`
	Strict     bool     `yaml:"strict"`
}

type entryLocateHit struct {
	Found         bool   `yaml:"found"`
	Position      int    `yaml:"position"`
	Kind          string `yaml:"kind"`
	RecordedIndex *int   `yaml:"recordedIndex"`
}

type entryLocateExpected struct {
	Hits       []entryLocateHit `yaml:"hits"`
	LegacyKept []int            `yaml:"legacyKept"`
	PiKept     []int            `yaml:"piKept"`
	Referenced []int            `yaml:"referenced"`
}

type entryLocateCases struct {
	Scan testcase.Corpus[entryLocateInput, entryLocateExpected] `yaml:"scan"`
	Wire testcase.Corpus[string, []string]                      `yaml:"wire"`
}

func loadEntryLocateCases(t *testing.T) entryLocateCases {
	t.Helper()
	fixtures := decodeStrictYAML[entryLocateCases](t, entryLocateYAML, "transcript entry locator fixture")
	caseassert.RequireValid(t, fixtures.Scan)
	caseassert.RequireValid(t, fixtures.Wire)
	var names []string
	for _, c := range fixtures.Scan.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.Wire.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "transcript entry locator", entryLocateManifestYAML, names)
	return fixtures
}

func TestLocateTranscriptEntry(t *testing.T) {
	for _, c := range loadEntryLocateCases(t).Scan.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Input.Strict {
				if _, err := schema.DecodeSessionDetailPayloadRaw([]byte(c.Input.Detail)); err != nil {
					t.Fatalf("strict fixture is not a valid served detail: %v", err)
				}
			}
			turns, err := schema.NewTranscriptCookTurns(json.RawMessage(c.Input.Detail))
			if err != nil {
				t.Fatal(err)
			}
			if c.Input.Protect {
				for _, harness := range []schema.Harness{"", schema.HarnessPi} {
					verdicts, err := schema.CookTranscriptRows(turns, harness)
					if err != nil {
						t.Fatal(err)
					}
					var kept []int
					for _, verdict := range verdicts {
						if verdict.Keep {
							kept = append(kept, verdict.Position)
						}
					}
					want := c.Expected.LegacyKept
					if harness == schema.HarnessPi {
						want = c.Expected.PiKept
					}
					if !slices.Equal(kept, want) {
						t.Fatalf("harness %q kept %v, want %v", harness, kept, want)
					}
					for _, position := range c.Expected.Referenced {
						if !verdicts[position].Keep {
							t.Fatalf("harness %q removed reference-bearing row %d", harness, position)
						}
					}
				}
				return
			}
			positions := c.Input.Subset
			if positions == nil {
				verdicts, err := schema.CookTranscriptRows(turns, schema.Harness(c.Input.Harness))
				if err != nil {
					t.Fatal(err)
				}
				for _, verdict := range verdicts {
					if verdict.Keep {
						positions = append(positions, verdict.Position)
					}
				}
			}
			rows := make([]schema.TranscriptEntryScanRow, 0, len(positions))
			for _, position := range positions {
				state, recorded := schema.DecodeTranscriptRecordedIndex(turns[position].Raw())
				rows = append(rows, schema.TranscriptEntryScanRow{Turn: turns[position], State: state, Recorded: recorded})
			}
			if len(c.Input.References) != len(c.Expected.Hits) {
				t.Fatal("each requested reference needs a finding")
			}
			for i, reference := range c.Input.References {
				hit, found, err := schema.LocateTranscriptEntry(rows, reference)
				if err != nil {
					t.Fatal(err)
				}
				got := entryLocateHit{Found: found}
				if found {
					got.Position, got.Kind, got.RecordedIndex = hit.Position, string(hit.Kind), hit.RecordedIndex
				}
				if !reflect.DeepEqual(got, c.Expected.Hits[i]) {
					t.Fatalf("reference %d: got %+v, want %+v", i, got, c.Expected.Hits[i])
				}
			}
		})
	}
}

func TestTranscriptEntryLocatorWire(t *testing.T) {
	fixtures := loadEntryLocateCases(t)
	types, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	readComponent := func(name string) map[string]any {
		t.Helper()
		raw, err := json.Marshal(types.Components.Schemas[name])
		if err != nil {
			t.Fatal(err)
		}
		var result map[string]any
		if err := json.Unmarshal(raw, &result); err != nil || result == nil {
			t.Fatalf("component %s = %s: %v", name, raw, err)
		}
		return result
	}
	request, response := readComponent("TranscriptEntryLocateRequest"), readComponent("TranscriptEntryMatch")
	keys := func(value map[string]any) []string {
		out := make([]string, 0, len(value))
		for key := range value {
			out = append(out, key)
		}
		slices.Sort(out)
		return out
	}
	for _, c := range fixtures.Wire.Cases {
		t.Run(c.Name, func(t *testing.T) {
			want := slices.Clone(c.Expected)
			slices.Sort(want)
			switch c.Input {
			case "kind-set":
				var got []string
				for _, v := range schema.AllTranscriptEntryReferenceKinds {
					if !v.IsValid() {
						t.Fatalf("known kind %q invalid", v)
					}
					got = append(got, string(v))
				}
				slices.Sort(got)
				if !slices.Equal(got, want) {
					t.Fatalf("kinds = %v, want %v", got, want)
				}
				kindSchema := readComponent("TranscriptEntryReferenceKind")
				var reflected []string
				for _, v := range kindSchema["enum"].([]any) {
					reflected = append(reflected, v.(string))
				}
				slices.Sort(reflected)
				if !slices.Equal(reflected, want) {
					t.Fatalf("reflected kinds = %v, want %v", reflected, want)
				}
			case "kind-unknown":
				if schema.TranscriptEntryReferenceKind("unexpected").IsValid() {
					t.Fatal("unknown kind accepted")
				}
			case "request-required":
				var got []string
				for _, v := range request["required"].([]any) {
					got = append(got, v.(string))
				}
				slices.Sort(got)
				if !slices.Equal(got, want) {
					t.Fatalf("required = %v, want %v", got, want)
				}
			case "request-properties", "response-properties":
				object := request
				if c.Input == "response-properties" {
					object = response
				}
				got := keys(object["properties"].(map[string]any))
				if !slices.Equal(got, want) {
					t.Fatalf("properties = %v, want %v", got, want)
				}
			case "request-min":
				ref := request["properties"].(map[string]any)["reference"].(map[string]any)
				if ref["minLength"] != float64(1) {
					t.Fatalf("reference schema = %v", ref)
				}
				if _, _, err := schema.LocateTranscriptEntry(nil, ""); err == nil {
					t.Fatal("empty scan reference accepted")
				}
			case "constant-size":
				locate := func(reference string) []byte {
					t.Helper()
					quoted, err := json.Marshal(reference)
					if err != nil {
						t.Fatal(err)
					}
					turns, err := schema.NewTranscriptCookTurns(json.RawMessage(`{"turns":[{"index":7,"sourceEntryRef":` + string(quoted) + `}]}`))
					if err != nil {
						t.Fatal(err)
					}
					state, recorded := schema.DecodeTranscriptRecordedIndex(turns[0].Raw())
					hit, found, err := schema.LocateTranscriptEntry([]schema.TranscriptEntryScanRow{{Turn: turns[0], State: state, Recorded: recorded}}, reference)
					if err != nil || !found {
						t.Fatalf("locate reference: found=%v, err=%v", found, err)
					}
					result, err := json.Marshal(schema.TranscriptEntryMatch{Kind: hit.Kind, Position: hit.Position, RecordedIndex: hit.RecordedIndex})
					if err != nil {
						t.Fatal(err)
					}
					return result
				}
				a, b := locate("short"), locate(strings.Repeat("z", 96))
				if len(a) != len(b) || strings.Contains(string(a), "short") || strings.Contains(string(b), strings.Repeat("z", 96)) {
					t.Fatalf("location response carries reference: %s %s", a, b)
				}
			case "statuses":
				spec, err := openapi.BuildVillageAPISpec()
				if err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(spec)
				if err != nil {
					t.Fatal(err)
				}
				var doc map[string]any
				if err := json.Unmarshal(raw, &doc); err != nil {
					t.Fatal(err)
				}
				path := doc["paths"].(map[string]any)["/api/v1/transcripts/{id}/segments/locate"].(map[string]any)
				op := path["post"].(map[string]any)
				if op["operationId"] != "locateTranscriptEntry" || op["requestBody"].(map[string]any)["required"] != true {
					t.Fatalf("locator operation = %v", op)
				}
				responses := op["responses"].(map[string]any)
				if got := keys(responses); !slices.Equal(got, want) {
					t.Fatalf("statuses = %v, want %v", got, want)
				}
				for code, raw := range responses {
					headers := raw.(map[string]any)["headers"].(map[string]any)
					if headers["Cache-Control"] == nil || headers["Vary"] == nil || (code == "503" && headers["Retry-After"] == nil) {
						t.Fatalf("response %s headers = %v", code, headers)
					}
				}
			default:
				t.Fatalf("unknown wire check %q", c.Input)
			}
		})
	}
}
