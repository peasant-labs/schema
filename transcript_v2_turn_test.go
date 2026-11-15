package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_segments_v2.yaml
var transcriptSegmentsV2YAML []byte

//go:embed testdata/transcript_segments_v2_manifest.yaml
var transcriptSegmentsV2ManifestYAML []byte

type v2TurnInput struct {
	Raw    string `yaml:"raw"`
	RawHex string `yaml:"rawHex"`
}

type v2TurnExpected struct {
	State string `yaml:"state"`
	Index *int   `yaml:"index"`
}

type v2CaptureInput struct {
	Detail string `yaml:"detail"`
	Want   int    `yaml:"want"`
}

type v2CaptureExpected struct {
	Turns         []string `yaml:"turns"`
	ErrorContains string   `yaml:"errorContains"`
}

type segmentV2Cases struct {
	Turns    testcase.Corpus[v2TurnInput, v2TurnExpected]       `yaml:"turns"`
	Capture  testcase.Corpus[v2CaptureInput, v2CaptureExpected] `yaml:"capture"`
	Evidence testcase.Corpus[string, string]                    `yaml:"evidence"`
}

func loadSegmentV2Cases(t *testing.T) segmentV2Cases {
	t.Helper()
	fixture := decodeStrictYAML[segmentV2Cases](t, transcriptSegmentsV2YAML, "v2 segment fixture")
	caseassert.RequireValid(t, fixture.Turns)
	caseassert.RequireValid(t, fixture.Capture)
	caseassert.RequireValid(t, fixture.Evidence)
	names := make([]string, 0, len(fixture.Turns.Cases)+len(fixture.Capture.Cases)+len(fixture.Evidence.Cases))
	for _, c := range fixture.Turns.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Capture.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Evidence.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "v2 segment", transcriptSegmentsV2ManifestYAML, names)
	return fixture
}

func TestTranscriptV2TurnEvidence(t *testing.T) {
	for _, c := range loadSegmentV2Cases(t).Evidence.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var turn schema.TurnDetail
			if err := json.Unmarshal([]byte(c.Input), &turn); err != nil {
				t.Fatal(err)
			}
			if turn.Provenance == nil {
				t.Fatal("fixture does not carry provenance")
			}
			err := turn.Provenance.Validate()
			if err == nil || !strings.Contains(err.Error(), c.Expected) {
				t.Fatalf("evidence error = %v, want text %q", err, c.Expected)
			}
		})
	}
}

func TestTranscriptV2Turn(t *testing.T) {
	fixture := loadSegmentV2Cases(t)
	for _, c := range fixture.Turns.Cases {
		t.Run(c.Name, func(t *testing.T) {
			raw := []byte(c.Input.Raw)
			if c.Input.RawHex != "" {
				var err error
				raw, err = hex.DecodeString(c.Input.RawHex)
				if err != nil {
					t.Fatal(err)
				}
			}
			turn, err := schema.NewTranscriptV2Turn(raw)
			if c.Classification == testcase.MustFail {
				if err == nil {
					t.Fatal("invalid turn was accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			// Both accessors must copy, so callers cannot modify retained bytes.
			want := bytes.Clone(raw)
			raw[0] ^= 1
			got := turn.Raw()
			if !bytes.Equal(got, want) {
				t.Fatalf("turn bytes = %q, want %q", got, want)
			}
			got[0] ^= 1
			encoded, err := turn.MarshalJSON()
			if err != nil || !bytes.Equal(encoded, want) {
				t.Fatalf("direct MarshalJSON = %q, %v; want %q", encoded, err, want)
			}
			var decoded schema.TranscriptV2Turn
			if err := json.Unmarshal(want, &decoded); err != nil || !bytes.Equal(decoded.Raw(), want) {
				t.Fatalf("UnmarshalJSON = %q, %v; want %q", decoded.Raw(), err, want)
			}
			state, index := schema.DecodeTranscriptRecordedIndex(turn)
			if string(state) != c.Expected.State || (index == nil) != (c.Expected.Index == nil) {
				t.Fatalf("recorded index = (%q, %v), want (%q, %v)", state, index, c.Expected.State, c.Expected.Index)
			}
			if index != nil && *index != *c.Expected.Index {
				t.Fatalf("recorded index = %d, want %d", *index, *c.Expected.Index)
			}
		})
	}
}

func TestTranscriptV2TurnBytes(t *testing.T) {
	for _, c := range loadSegmentV2Cases(t).Capture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turns, err := schema.TranscriptV2TurnBytes(json.RawMessage(c.Input.Detail), c.Input.Want)
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("capture error = %v, want text %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(turns) != len(c.Expected.Turns) {
				t.Fatalf("captured %d turns, want %d", len(turns), len(c.Expected.Turns))
			}
			for i, turn := range turns {
				if string(turn.Raw()) != c.Expected.Turns[i] {
					t.Fatalf("turn %d = %q, want %q", i, turn.Raw(), c.Expected.Turns[i])
				}
			}
		})
	}
}
