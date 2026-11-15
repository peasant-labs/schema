package schema_test

import (
	_ "embed"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_window_v2.yaml
var transcriptWindowV2YAML []byte

//go:embed testdata/transcript_window_v2_manifest.yaml
var transcriptWindowV2ManifestYAML []byte

type windowTurnV2Spec struct {
	Position int    `yaml:"position"`
	Role     string `yaml:"role"`
	Depth    int    `yaml:"depth"`
	Index    *int   `yaml:"index"`
}

type windowRefV2Spec struct {
	Index int `yaml:"index"`
	First int `yaml:"first"`
	Last  int `yaml:"last"`
}

type windowV2Input struct {
	TotalSegments int                `yaml:"totalSegments"`
	TurnCount     int                `yaml:"turnCount"`
	Span          int                `yaml:"span"`
	Anchor        *int               `yaml:"anchor"`
	Turns         []windowTurnV2Spec `yaml:"turns"`
	Refs          []windowRefV2Spec  `yaml:"refs"`
}

type windowV2Expected struct {
	AnchorPosition    int    `yaml:"anchorPosition"`
	AnchorTurn        *int   `yaml:"anchorTurn"`
	RequestedPosition *int   `yaml:"requestedPosition"`
	AnchorSegment     int    `yaml:"anchorSegment"`
	FirstPosition     int    `yaml:"firstPosition"`
	LastPosition      int    `yaml:"lastPosition"`
	Segments          []int  `yaml:"segments"`
	HasOlder          bool   `yaml:"hasOlder"`
	HasNewer          bool   `yaml:"hasNewer"`
	ErrorContains     string `yaml:"errorContains"`
}

func TestResolveTranscriptWindowV2(t *testing.T) {
	corpus := decodeStrictYAML[testcase.Corpus[windowV2Input, windowV2Expected]](t, transcriptWindowV2YAML, "v2 window fixture")
	caseassert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "v2 window", transcriptWindowV2ManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sessionID := schema.SessionID("ses_test")
			turns := make([]schema.TranscriptWindowTurnV2, c.Input.TurnCount)
			for i := range turns {
				turns[i] = schema.TranscriptWindowTurnV2{Position: i, Role: schema.RoleSystem}
			}
			for _, row := range c.Input.Turns {
				if row.Position < 0 || row.Position >= len(turns) {
					t.Fatalf("fixture position %d is outside the row array", row.Position)
				}
				turns[row.Position] = schema.TranscriptWindowTurnV2{Position: row.Position, Role: schema.Role(row.Role), Depth: row.Depth, RecordedIndex: row.Index}
			}
			refs := make([]schema.TranscriptSegmentRefV2, len(c.Input.Refs))
			for i, ref := range c.Input.Refs {
				refs[i] = schema.TranscriptSegmentRefV2{Index: ref.Index, SourceFirstPosition: ref.First, SourceLastPosition: ref.Last}
			}
			input := schema.TranscriptWindowV2Input{
				Context: schema.TranscriptReadContext{
					Transport: schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:  "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: len(turns), TotalSegments: c.Input.TotalSegments,
				},
				Partition:       schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain},
				RequestedAnchor: c.Input.Anchor, Span: c.Input.Span, Turns: turns, Refs: refs,
			}
			got, err := schema.ResolveTranscriptWindowV2(input)
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("window error = %v, want text %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := c.Expected
			if got.AnchorPosition != want.AnchorPosition || !reflect.DeepEqual(got.AnchorTurn, want.AnchorTurn) ||
				!reflect.DeepEqual(got.RequestedPosition, want.RequestedPosition) || got.AnchorSegment != want.AnchorSegment ||
				got.FirstPosition != want.FirstPosition || got.LastPosition != want.LastPosition ||
				!reflect.DeepEqual(got.Segments, want.Segments) || got.HasOlder != want.HasOlder || got.HasNewer != want.HasNewer {
				t.Fatalf("window = %+v, want %+v", got, want)
			}
			if err := schema.ValidateTranscriptWindowV2Coverage(got, refs); err != nil {
				t.Fatalf("resolved window fails coverage: %v", err)
			}
		})
	}
}
