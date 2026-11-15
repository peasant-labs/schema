package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"reflect"
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

type v2SegmentInput struct {
	Segment  int `yaml:"segment"`
	Total    int `yaml:"total"`
	Start    int `yaml:"start"`
	Count    int `yaml:"count"`
	GapAt    int `yaml:"gapAt"`
	RefFirst int `yaml:"refFirst"`
	RefLast  int `yaml:"refLast"`
}

type v2SegmentExpected struct {
	SegmentValid  bool   `yaml:"segmentValid"`
	BoundaryValid bool   `yaml:"boundaryValid"`
	ErrorContains string `yaml:"errorContains"`
}

type v2DirectoryInput struct {
	TotalTurns        int  `yaml:"totalTurns"`
	TotalSegments     int  `yaml:"totalSegments"`
	Page              *int `yaml:"page"`
	PerPage           *int `yaml:"perPage"`
	RefCount          int  `yaml:"refCount"`
	RefStart          int  `yaml:"refStart"`
	PositionStart     int  `yaml:"positionStart"`
	BreakPosition     *int `yaml:"breakPosition"`
	FirstTurn         *int `yaml:"firstTurn"`
	LastTurn          *int `yaml:"lastTurn"`
	TotalConversation *int `yaml:"totalConversation"`
	SpanBefore        *int `yaml:"spanBefore"`
	SpanCount         *int `yaml:"spanCount"`
}

type v2DirectoryExpected struct {
	ErrorContains string `yaml:"errorContains"`
}

type v2WireInput struct {
	Subject string `yaml:"subject"`
	JSON    string `yaml:"json"`
}

type v2WireExpected struct {
	ErrorContains       string  `yaml:"errorContains"`
	CheckContentSession bool    `yaml:"checkContentSession"`
	ContentSessionID    *string `yaml:"contentSessionId"`
}

type v2RawInput struct {
	Mode string `yaml:"mode"`
	Row  int    `yaml:"row"`
}

type v2RawExpected struct {
	Raw     string   `yaml:"raw"`
	Members []string `yaml:"members"`
}

type segmentV2Cases struct {
	Turns       testcase.Corpus[v2TurnInput, v2TurnExpected]           `yaml:"turns"`
	Capture     testcase.Corpus[v2CaptureInput, v2CaptureExpected]     `yaml:"capture"`
	Evidence    testcase.Corpus[string, string]                        `yaml:"evidence"`
	Segments    testcase.Corpus[v2SegmentInput, v2SegmentExpected]     `yaml:"segments"`
	Directories testcase.Corpus[v2DirectoryInput, v2DirectoryExpected] `yaml:"directories"`
	Wire        testcase.Corpus[v2WireInput, v2WireExpected]           `yaml:"wire"`
	Raw         testcase.Corpus[v2RawInput, v2RawExpected]             `yaml:"raw"`
}

func loadSegmentV2Cases(t *testing.T) segmentV2Cases {
	t.Helper()
	fixture := decodeStrictYAML[segmentV2Cases](t, transcriptSegmentsV2YAML, "v2 segment fixture")
	caseassert.RequireValid(t, fixture.Turns)
	caseassert.RequireValid(t, fixture.Capture)
	caseassert.RequireValid(t, fixture.Evidence)
	caseassert.RequireValid(t, fixture.Segments)
	caseassert.RequireValid(t, fixture.Directories)
	caseassert.RequireValid(t, fixture.Wire)
	caseassert.RequireValid(t, fixture.Raw)
	names := make([]string, 0, len(fixture.Turns.Cases)+len(fixture.Capture.Cases)+len(fixture.Evidence.Cases)+len(fixture.Segments.Cases)+len(fixture.Directories.Cases)+len(fixture.Wire.Cases)+len(fixture.Raw.Cases))
	for _, c := range fixture.Turns.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Capture.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Evidence.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Segments.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Directories.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Wire.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixture.Raw.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "v2 segment", transcriptSegmentsV2ManifestYAML, names)
	return fixture
}

func TestTranscriptV2WireValidation(t *testing.T) {
	for _, c := range loadSegmentV2Cases(t).Wire.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			switch c.Input.Subject {
			case "transport":
				var transport schema.TranscriptReadTransport
				if err := json.Unmarshal([]byte(c.Input.JSON), &transport); err != nil {
					t.Fatal(err)
				}
				err = transport.Validate()
			case "context":
				var context schema.TranscriptReadContext
				if err := json.Unmarshal([]byte(c.Input.JSON), &context); err != nil {
					t.Fatal(err)
				}
				err = context.Validate()
				if c.Expected.CheckContentSession && !reflect.DeepEqual(context.ContentSessionID, c.Expected.ContentSessionID) {
					t.Fatalf("contentSessionId = %v, want %v", context.ContentSessionID, c.Expected.ContentSessionID)
				}
			case "partition":
				var partition schema.TranscriptSegmentPartition
				if err := json.Unmarshal([]byte(c.Input.JSON), &partition); err != nil {
					t.Fatal(err)
				}
				err = partition.Validate()
			case "row":
				var row schema.TranscriptSegmentRow
				if err := json.Unmarshal([]byte(c.Input.JSON), &row); err != nil {
					t.Fatal(err)
				}
				err = row.Validate()
			default:
				t.Fatalf("unknown fixture subject %q", c.Input.Subject)
			}
			if c.Classification == testcase.MustPass {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
				t.Fatalf("validation = %v, want error containing %q", err, c.Expected.ErrorContains)
			}
		})
	}
}

func TestTranscriptSegmentDirectoryV2(t *testing.T) {
	for _, c := range loadSegmentV2Cases(t).Directories.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sessionID := schema.SessionID("ses_test")
			refs := make([]schema.TranscriptSegmentRefV2, c.Input.RefCount)
			for i := range refs {
				position := c.Input.PositionStart + i
				if c.Input.BreakPosition != nil && i == *c.Input.BreakPosition {
					position++
				}
				refs[i] = schema.TranscriptSegmentRefV2{
					Index: c.Input.RefStart + i, SourceFirstPosition: position, SourceLastPosition: position,
				}
			}
			if len(refs) > 0 {
				refs[0].FirstTurn, refs[0].LastTurn = c.Input.FirstTurn, c.Input.LastTurn
				if c.Input.SpanBefore != nil && c.Input.SpanCount != nil {
					refs[0].ConversationTurns = &schema.ConversationTurnSpan{Before: *c.Input.SpanBefore, Count: *c.Input.SpanCount}
				}
			}
			dir := schema.TranscriptSegmentDirectoryV2{
				Context: schema.TranscriptReadContext{
					Transport: schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:  "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: c.Input.TotalTurns,
					TotalSegments: c.Input.TotalSegments, TotalConversationTurns: c.Input.TotalConversation,
				},
				Partition:    schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain},
				UsagePresent: schema.TranscriptUsageUnknown, Segments: refs,
				DirectoryPage: c.Input.Page, DirectoryPerPage: c.Input.PerPage,
			}
			err := dir.Validate()
			if c.Classification == testcase.MustPass {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
				t.Fatalf("directory validation = %v, want error containing %q", err, c.Expected.ErrorContains)
			}
		})
	}
}

func TestTranscriptSegmentV2Boundaries(t *testing.T) {
	for _, c := range loadSegmentV2Cases(t).Segments.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sessionID := schema.SessionID("ses_test")
			turn, err := schema.NewTranscriptV2Turn([]byte(`{"index":0}`))
			if err != nil {
				t.Fatal(err)
			}
			rows := make([]schema.TranscriptSegmentRow, c.Input.Count)
			for i := range rows {
				position := c.Input.Start + i
				if c.Input.GapAt > 0 && i >= c.Input.GapAt {
					position++
				}
				index := 0
				rows[i] = schema.TranscriptSegmentRow{Position: position, RecordedIndexState: schema.TranscriptRecordedIndexPresent, RecordedIndex: &index, Turn: turn}
			}
			segment := schema.TranscriptSegmentV2{
				Context: schema.TranscriptReadContext{
					Transport: schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:  "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: 100, TotalSegments: c.Input.Total,
				},
				Partition: schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain},
				Index:     c.Input.Segment, Rows: rows, HasOlder: c.Input.Segment > 0, HasNewer: c.Input.Segment < c.Input.Total-1,
			}
			ref := schema.TranscriptSegmentRefV2{Index: c.Input.Segment, SourceFirstPosition: c.Input.RefFirst, SourceLastPosition: c.Input.RefLast}
			segmentErr := segment.Validate()
			boundaryErr := schema.ValidateTranscriptSegmentV2Boundaries(segment, ref)
			if (segmentErr == nil) != c.Expected.SegmentValid || (boundaryErr == nil) != c.Expected.BoundaryValid {
				t.Fatalf("segment error = %v, boundary error = %v, want %+v", segmentErr, boundaryErr, c.Expected)
			}
			if c.Expected.ErrorContains != "" {
				err := boundaryErr
				if !c.Expected.SegmentValid {
					err = segmentErr
				}
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("error = %v, want text %q", err, c.Expected.ErrorContains)
				}
			}
			if c.Expected.SegmentValid && c.Expected.BoundaryValid {
				if _, err := schema.MarshalTranscriptSegmentV2(segment); err != nil {
					t.Fatalf("write valid segment: %v", err)
				}
			}
		})
	}
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
