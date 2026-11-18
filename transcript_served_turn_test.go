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
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_positioned_segments.yaml
var transcriptPositionedSegmentsYAML []byte

//go:embed testdata/transcript_positioned_segments_manifest.yaml
var transcriptPositionedSegmentsManifestYAML []byte

type positionedTurnInput struct {
	Raw    string `yaml:"raw"`
	RawHex string `yaml:"rawHex"`
}

type positionedTurnExpected struct {
	State string `yaml:"state"`
	Index *int   `yaml:"index"`
}

type positionedCaptureInput struct {
	Detail string `yaml:"detail"`
}

type positionedCaptureExpected struct {
	Turns         []string `yaml:"turns"`
	ErrorContains string   `yaml:"errorContains"`
}

type positionedSegmentInput struct {
	Segment  int `yaml:"segment"`
	Total    int `yaml:"total"`
	Start    int `yaml:"start"`
	Count    int `yaml:"count"`
	GapAt    int `yaml:"gapAt"`
	RefFirst int `yaml:"refFirst"`
	RefLast  int `yaml:"refLast"`
}

type positionedSegmentExpected struct {
	SegmentValid  bool   `yaml:"segmentValid"`
	BoundaryValid bool   `yaml:"boundaryValid"`
	ErrorContains string `yaml:"errorContains"`
}

type positionedDirectoryInput struct {
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

type positionedDirectoryExpected struct {
	ErrorContains string `yaml:"errorContains"`
}

// positionedWireSubject selects which positioned-read shape a fixture tests.
// It is independent of command names, diff kinds, and preview event kinds.
type positionedWireSubject string

const (
	positionedWireTransport positionedWireSubject = "transport"
	positionedWireContext   positionedWireSubject = "context"
	positionedWirePartition positionedWireSubject = "partition"
	positionedWireRow       positionedWireSubject = "row"
	positionedWireDirectory positionedWireSubject = "directory"
	positionedWireSegment   positionedWireSubject = "segment"
	positionedWireWindow    positionedWireSubject = "window"
	positionedWireRefSchema positionedWireSubject = "ref-schema"
)

type positionedWireInput struct {
	Subject positionedWireSubject `yaml:"subject"`
	JSON    string                `yaml:"json"`
}

type positionedWireExpected struct {
	ErrorContains       string   `yaml:"errorContains"`
	CheckContentSession bool     `yaml:"checkContentSession"`
	ContentSessionID    *string  `yaml:"contentSessionId"`
	CheckHarness        bool     `yaml:"checkHarness"`
	Harness             *string  `yaml:"harness"`
	RecordedIndex       *int     `yaml:"recordedIndex"`
	DisplayNumber       *string  `yaml:"displayNumber"`
	RequiredMembers     []string `yaml:"requiredMembers"`
	ForbiddenMembers    []string `yaml:"forbiddenMembers"`
	Contains            string   `yaml:"contains"`
	NotContains         string   `yaml:"notContains"`
}

type positionedRawInput struct {
	Mode string `yaml:"mode"`
	Row  int    `yaml:"row"`
}

type positionedRawExpected struct {
	Raw     string   `yaml:"raw"`
	Members []string `yaml:"members"`
}

type positionedServedInput struct {
	Detail  string `yaml:"detail"`
	Harness string `yaml:"harness"`
}

type positionedServedRow struct {
	Position      int    `yaml:"position"`
	DisplayNumber string `yaml:"displayNumber"`
	Turn          string `yaml:"turn"`
}

type positionedSegmentCases struct {
	Turns       testcase.Corpus[positionedTurnInput, positionedTurnExpected]           `yaml:"turns"`
	Capture     testcase.Corpus[positionedCaptureInput, positionedCaptureExpected]     `yaml:"capture"`
	Evidence    testcase.Corpus[string, string]                                        `yaml:"evidence"`
	Segments    testcase.Corpus[positionedSegmentInput, positionedSegmentExpected]     `yaml:"segments"`
	Directories testcase.Corpus[positionedDirectoryInput, positionedDirectoryExpected] `yaml:"directories"`
	Wire        testcase.Corpus[positionedWireInput, positionedWireExpected]           `yaml:"wire"`
	Raw         testcase.Corpus[positionedRawInput, positionedRawExpected]             `yaml:"raw"`
	Served      testcase.Corpus[positionedServedInput, []positionedServedRow]          `yaml:"served"`
}

func loadPositionedSegmentCases(t *testing.T) positionedSegmentCases {
	t.Helper()
	fixture := decodeStrictYAML[positionedSegmentCases](t, transcriptPositionedSegmentsYAML, "positioned segment fixture")
	caseassert.RequireValid(t, fixture.Turns)
	caseassert.RequireValid(t, fixture.Capture)
	caseassert.RequireValid(t, fixture.Evidence)
	caseassert.RequireValid(t, fixture.Segments)
	caseassert.RequireValid(t, fixture.Directories)
	caseassert.RequireValid(t, fixture.Wire)
	caseassert.RequireValid(t, fixture.Raw)
	caseassert.RequireValid(t, fixture.Served)
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
	for _, c := range fixture.Served.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "positioned segment", transcriptPositionedSegmentsManifestYAML, names)
	return fixture
}

func TestPositionedSegmentServesCookedRows(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Served.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turns, err := schema.NewTranscriptCookTurns(json.RawMessage(c.Input.Detail))
			if err != nil {
				t.Fatal(err)
			}
			verdicts, err := schema.CookTranscriptRows(turns, schema.Harness(c.Input.Harness))
			if err != nil {
				t.Fatal(err)
			}
			rows := make([]schema.TranscriptSegmentRow, 0, len(turns))
			for i, verdict := range verdicts {
				if !verdict.Keep {
					continue
				}
				turn := turns[i].Raw()
				state, index := schema.DecodeTranscriptRecordedIndex(turn)
				rows = append(rows, schema.TranscriptSegmentRow{
					Position: verdict.ServedPosition, RecordedIndexState: state, RecordedIndex: index,
					DisplayNumber: verdict.DisplayNumber, Turn: turn,
				})
			}
			sessionID := schema.SessionID("ses_test")
			segment := schema.PositionedSegment{
				Context: schema.TranscriptReadContext{
					ProtocolVersion: schema.SegmentedReadAPIVersion,
					Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:        "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: len(rows), TotalSegments: 1,
				},
				Partition: schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain}, Rows: rows,
			}
			wire, err := schema.MarshalPositionedSegment(segment)
			if err != nil {
				t.Fatal(err)
			}
			var read schema.PositionedSegment
			if err := json.Unmarshal(wire, &read); err != nil {
				t.Fatal(err)
			}
			got := make([]positionedServedRow, len(read.Rows))
			for i, row := range read.Rows {
				got[i] = positionedServedRow{Position: row.Position, DisplayNumber: row.DisplayNumber, Turn: string(row.Turn.Raw())}
			}
			if !reflect.DeepEqual(got, c.Expected) {
				t.Fatalf("served rows = %#v, want %#v", got, c.Expected)
			}
		})
	}
}

func TestPositionedSegmentWireValidation(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Wire.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			var encoded []byte
			switch c.Input.Subject {
			case positionedWireTransport:
				var transport schema.TranscriptReadTransport
				if err := json.Unmarshal([]byte(c.Input.JSON), &transport); err != nil {
					t.Fatal(err)
				}
				err = transport.Validate()
			case positionedWireContext:
				var context schema.TranscriptReadContext
				if err := json.Unmarshal([]byte(c.Input.JSON), &context); err != nil {
					t.Fatal(err)
				}
				err = context.Validate()
				if c.Expected.CheckContentSession && !reflect.DeepEqual(context.ContentSessionID, c.Expected.ContentSessionID) {
					t.Fatalf("contentSessionId = %v, want %v", context.ContentSessionID, c.Expected.ContentSessionID)
				}
				if c.Expected.CheckHarness {
					if !reflect.DeepEqual(context.Harness, c.Expected.Harness) {
						t.Fatalf("harness = %v, want %v", context.Harness, c.Expected.Harness)
					}
					wire, marshalErr := json.Marshal(context)
					if marshalErr != nil {
						t.Fatal(marshalErr)
					}
					var object map[string]json.RawMessage
					if err := json.Unmarshal(wire, &object); err != nil {
						t.Fatal(err)
					}
					raw, present := object["harness"]
					if present != (c.Expected.Harness != nil) {
						t.Fatalf("marshaled harness presence = %v, want %v: %s", present, c.Expected.Harness != nil, wire)
					}
					if present {
						var value string
						if err := json.Unmarshal(raw, &value); err != nil || value != *c.Expected.Harness {
							t.Fatalf("marshaled harness = %s, want %q: %v", raw, *c.Expected.Harness, err)
						}
					}
				}
			case positionedWirePartition:
				var partition schema.TranscriptSegmentPartition
				if err := json.Unmarshal([]byte(c.Input.JSON), &partition); err != nil {
					t.Fatal(err)
				}
				err = partition.Validate()
			case positionedWireRow:
				var row schema.TranscriptSegmentRow
				if err := json.Unmarshal([]byte(c.Input.JSON), &row); err != nil {
					t.Fatal(err)
				}
				err = row.Validate()
				if c.Expected.RecordedIndex != nil && !reflect.DeepEqual(row.RecordedIndex, c.Expected.RecordedIndex) {
					t.Fatalf("recordedIndex = %v, want %v", row.RecordedIndex, c.Expected.RecordedIndex)
				}
				if c.Expected.DisplayNumber != nil && row.DisplayNumber != *c.Expected.DisplayNumber {
					t.Fatalf("displayNumber = %q, want %q", row.DisplayNumber, *c.Expected.DisplayNumber)
				}
			case positionedWireDirectory:
				var directory schema.PositionedDirectory
				if err := json.Unmarshal([]byte(c.Input.JSON), &directory); err != nil {
					t.Fatal(err)
				}
				err = directory.Validate()
				if err == nil {
					encoded, err = json.Marshal(directory)
				}
			case positionedWireSegment:
				var segment schema.PositionedSegment
				if err := json.Unmarshal([]byte(c.Input.JSON), &segment); err != nil {
					t.Fatal(err)
				}
				encoded, err = schema.MarshalPositionedSegment(segment)
			case positionedWireWindow:
				var window schema.PositionedWindow
				if err := json.Unmarshal([]byte(c.Input.JSON), &window); err != nil {
					t.Fatal(err)
				}
				err = window.Validate()
				if err == nil {
					encoded, err = json.Marshal(window)
				}
			case positionedWireRefSchema:
				spec, buildErr := openapi.BuildTypesSpec()
				if buildErr != nil {
					t.Fatal(buildErr)
				}
				component := spec.Components.Schemas["PositionedSegmentRef"]
				if component == nil {
					t.Fatal("positioned segment ref is absent from the Types catalog")
				}
				raw, marshalErr := json.Marshal(component)
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				var shape struct {
					Required []string `json:"required"`
				}
				if err := json.Unmarshal(raw, &shape); err != nil {
					t.Fatal(err)
				}
				required := make(map[string]bool, len(shape.Required))
				for _, name := range shape.Required {
					required[name] = true
				}
				encoded, err = json.Marshal(required)
			default:
				t.Fatalf("unknown fixture subject %q", c.Input.Subject)
			}
			if c.Classification == testcase.MustPass {
				if err != nil {
					t.Fatal(err)
				}
				if len(c.Expected.RequiredMembers)+len(c.Expected.ForbiddenMembers) > 0 {
					var members map[string]json.RawMessage
					if err := json.Unmarshal(encoded, &members); err != nil {
						t.Fatal(err)
					}
					for _, name := range c.Expected.RequiredMembers {
						if _, ok := members[name]; !ok {
							t.Fatalf("wire has no %q member: %s", name, encoded)
						}
					}
					for _, name := range c.Expected.ForbiddenMembers {
						if _, ok := members[name]; ok {
							t.Fatalf("wire unexpectedly carries %q: %s", name, encoded)
						}
					}
				}
				if c.Expected.Contains != "" && !strings.Contains(string(encoded), c.Expected.Contains) {
					t.Fatalf("wire %s does not contain %q", encoded, c.Expected.Contains)
				}
				if c.Expected.NotContains != "" && strings.Contains(string(encoded), c.Expected.NotContains) {
					t.Fatalf("wire %s unexpectedly contains %q", encoded, c.Expected.NotContains)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
				t.Fatalf("validation = %v, want error containing %q", err, c.Expected.ErrorContains)
			}
		})
	}
}

func TestPositionedDirectory(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Directories.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sessionID := schema.SessionID("ses_test")
			refs := make([]schema.PositionedSegmentRef, c.Input.RefCount)
			for i := range refs {
				position := c.Input.PositionStart + i
				if c.Input.BreakPosition != nil && i == *c.Input.BreakPosition {
					position++
				}
				refs[i] = schema.PositionedSegmentRef{
					Index: c.Input.RefStart + i, SourceFirstPosition: position, SourceLastPosition: position,
				}
			}
			if len(refs) > 0 {
				refs[0].FirstTurn, refs[0].LastTurn = c.Input.FirstTurn, c.Input.LastTurn
				if c.Input.SpanBefore != nil && c.Input.SpanCount != nil {
					refs[0].ConversationTurns = &schema.ConversationTurnSpan{Before: *c.Input.SpanBefore, Count: *c.Input.SpanCount}
				}
			}
			dir := schema.PositionedDirectory{
				Context: schema.TranscriptReadContext{
					ProtocolVersion: schema.SegmentedReadAPIVersion,
					Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:        "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
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

func TestPositionedSegmentBoundaries(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Segments.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sessionID := schema.SessionID("ses_test")
			turn, err := schema.NewTranscriptServedTurn([]byte(`{"index":0}`))
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
			segment := schema.PositionedSegment{
				Context: schema.TranscriptReadContext{
					ProtocolVersion: schema.SegmentedReadAPIVersion,
					Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:        "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: 100, TotalSegments: c.Input.Total,
				},
				Partition: schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain},
				Index:     c.Input.Segment, Rows: rows, HasOlder: c.Input.Segment > 0, HasNewer: c.Input.Segment < c.Input.Total-1,
			}
			ref := schema.PositionedSegmentRef{Index: c.Input.Segment, SourceFirstPosition: c.Input.RefFirst, SourceLastPosition: c.Input.RefLast}
			segmentErr := segment.Validate()
			boundaryErr := schema.ValidatePositionedSegmentBoundaries(segment, ref)
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
				if _, err := schema.MarshalPositionedSegment(segment); err != nil {
					t.Fatalf("write valid segment: %v", err)
				}
			}
		})
	}
}

func TestTranscriptServedTurnEvidence(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Evidence.Cases {
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

func TestTranscriptServedTurn(t *testing.T) {
	fixture := loadPositionedSegmentCases(t)
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
			turn, err := schema.NewTranscriptServedTurn(raw)
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
			var decoded schema.TranscriptServedTurn
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
			// Exercise the transport writer even for a syntactically valid turn
			// that the document boundary would reject, such as duplicate keys.
			sessionID := schema.SessionID("ses_test")
			segment := schema.PositionedSegment{
				Context: schema.TranscriptReadContext{
					ProtocolVersion: schema.SegmentedReadAPIVersion,
					Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:        "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: 1, TotalSegments: 1,
				},
				Partition: schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain},
				Rows:      []schema.TranscriptSegmentRow{{Position: 0, RecordedIndexState: state, RecordedIndex: index, Turn: turn}},
			}
			body, err := schema.MarshalPositionedSegment(segment)
			if err != nil {
				t.Fatalf("write preserved turn: %v", err)
			}
			var served schema.PositionedSegment
			if err := json.Unmarshal(body, &served); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(served.Rows[0].Turn.Raw(), want) {
				t.Fatalf("writer changed turn bytes: got %q, want %q", served.Rows[0].Turn.Raw(), want)
			}
		})
	}
}

func TestTranscriptCookTurns(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Capture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turns, err := schema.NewTranscriptCookTurns(json.RawMessage(c.Input.Detail))
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
				if string(turn.Raw().Raw()) != c.Expected.Turns[i] {
					t.Fatalf("turn %d = %q, want %q", i, turn.Raw().Raw(), c.Expected.Turns[i])
				}
				var decoded schema.TurnDetail
				if err := json.Unmarshal(turn.Raw().Raw(), &decoded); err != nil || !reflect.DeepEqual(decoded, turn.Detail()) {
					t.Fatalf("turn %d was not paired with the detail decoded from its bytes: %v", i, err)
				}
			}
		})
	}
}
