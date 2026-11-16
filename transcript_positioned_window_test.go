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

//go:embed testdata/transcript_positioned_window.yaml
var transcriptPositionedWindowYAML []byte

//go:embed testdata/transcript_positioned_window_manifest.yaml
var transcriptPositionedWindowManifestYAML []byte

type windowPositionedTurnSpec struct {
	Position int    `yaml:"position"`
	Role     string `yaml:"role"`
	Depth    int    `yaml:"depth"`
	Index    *int   `yaml:"index"`
	Content  string `yaml:"content"`
}

type windowPositionedRefSpec struct {
	Index int `yaml:"index"`
	First int `yaml:"first"`
	Last  int `yaml:"last"`
}

type windowPositionedInput struct {
	TotalSegments int                        `yaml:"totalSegments"`
	TurnCount     int                        `yaml:"turnCount"`
	Span          int                        `yaml:"span"`
	Anchor        *int                       `yaml:"anchor"`
	Partition     string                     `yaml:"partition"`
	Section       *int                       `yaml:"section"`
	Mutation      string                     `yaml:"windowMutation"`
	Cook          bool                       `yaml:"cook"`
	CompareOffset int                        `yaml:"compareOffset"`
	CompareCount  int                        `yaml:"compareCount"`
	Turns         []windowPositionedTurnSpec `yaml:"turns"`
	Refs          []windowPositionedRefSpec  `yaml:"refs"`
}

type windowPositionedExpected struct {
	AnchorPosition    int      `yaml:"anchorPosition"`
	AnchorTurn        *int     `yaml:"anchorTurn"`
	RequestedPosition *int     `yaml:"requestedPosition"`
	AnchorSegment     int      `yaml:"anchorSegment"`
	FirstPosition     int      `yaml:"firstPosition"`
	LastPosition      int      `yaml:"lastPosition"`
	Segments          []int    `yaml:"segments"`
	HasOlder          bool     `yaml:"hasOlder"`
	HasNewer          bool     `yaml:"hasNewer"`
	ErrorContains     string   `yaml:"errorContains"`
	ErrorIncludes     []string `yaml:"errorIncludes"`
	Span              *int     `yaml:"span"`
}

func TestResolvePositionedWindow(t *testing.T) {
	corpus := decodeStrictYAML[testcase.Corpus[windowPositionedInput, windowPositionedExpected]](t, transcriptPositionedWindowYAML, "positioned window fixture")
	caseassert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "positioned window", transcriptPositionedWindowManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sessionID := schema.SessionID("ses_test")
			turns := make([]schema.PositionedWindowTurn, c.Input.TurnCount)
			for i := range turns {
				turns[i] = schema.PositionedWindowTurn{Position: i, Role: schema.RoleSystem}
			}
			for _, row := range c.Input.Turns {
				if row.Position < 0 || row.Position >= len(turns) {
					t.Fatalf("fixture position %d is outside the row array", row.Position)
				}
				turns[row.Position] = schema.PositionedWindowTurn{Position: row.Position, Role: schema.Role(row.Role), Depth: row.Depth, RecordedIndex: row.Index}
			}
			if c.Input.Cook {
				source := make([]schema.TurnDetail, len(turns))
				for _, row := range c.Input.Turns {
					source[row.Position] = schema.TurnDetail{Role: schema.Role(row.Role), Depth: row.Depth, Content: row.Content}
				}
				verdict := cookFixtureTurns(t, source)
				served := make([]schema.PositionedWindowTurn, 0, len(turns))
				for i, row := range turns {
					if verdict[i].Keep {
						row.Position = verdict[i].ServedPosition
						served = append(served, row)
					}
				}
				turns = served
			}
			refs := make([]schema.PositionedSegmentRef, len(c.Input.Refs))
			for i, ref := range c.Input.Refs {
				refs[i] = schema.PositionedSegmentRef{Index: ref.Index, SourceFirstPosition: ref.First, SourceLastPosition: ref.Last}
			}
			partition := schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain}
			if c.Input.Partition != "" {
				partition.Kind = schema.TranscriptSegmentPartitionKind(c.Input.Partition)
				partition.SectionIndex = c.Input.Section
			}
			input := schema.PositionedWindowInput{
				Context: schema.TranscriptReadContext{
					ProtocolVersion: schema.SegmentedReadAPIVersion,
					Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &sessionID},
					Revision:        "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
					NativePresence: schema.TranscriptNativeAbsent, TotalTurns: len(turns), TotalSegments: c.Input.TotalSegments,
				},
				Partition:       partition,
				RequestedAnchor: c.Input.Anchor, Span: c.Input.Span, Turns: turns, Refs: refs,
			}
			got, err := schema.ResolvePositionedWindow(input)
			if c.Input.Mutation != "" {
				if err != nil {
					t.Fatalf("resolve before window mutation: %v", err)
				}
				switch c.Input.Mutation {
				case "noncontiguous":
					got.Segments = []int{0, 2}
				case "incorrect-newer":
					got.HasNewer = !got.HasNewer
				case "requested-outside-run":
					found := false
					for _, ref := range refs {
						if ref.Index < got.Segments[0] || ref.Index > got.Segments[len(got.Segments)-1] {
							position := ref.SourceFirstPosition
							got.RequestedPosition = &position
							found = true
							break
						}
					}
					if !found || got.RequestedAnchor == nil {
						t.Fatal("fixture needs an anchored window and a ref outside its resolved run")
					}
				default:
					t.Fatalf("unknown fixture window mutation %q", c.Input.Mutation)
				}
				err = schema.ValidatePositionedWindowCoverage(got, refs)
			}
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("window error = %v, want text %q", err, c.Expected.ErrorContains)
				}
				for _, fragment := range c.Expected.ErrorIncludes {
					if !strings.Contains(err.Error(), fragment) {
						t.Fatalf("window error = %v, missing required guidance %q", err, fragment)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := c.Expected
			if !reflect.DeepEqual(got.Context, input.Context) || !reflect.DeepEqual(got.Partition, input.Partition) ||
				!reflect.DeepEqual(got.RequestedAnchor, input.RequestedAnchor) ||
				(want.Span != nil && got.Span != *want.Span) ||
				got.AnchorPosition != want.AnchorPosition || !reflect.DeepEqual(got.AnchorTurn, want.AnchorTurn) ||
				!reflect.DeepEqual(got.RequestedPosition, want.RequestedPosition) || got.AnchorSegment != want.AnchorSegment ||
				got.FirstPosition != want.FirstPosition || got.LastPosition != want.LastPosition ||
				!reflect.DeepEqual(got.Segments, want.Segments) || got.HasOlder != want.HasOlder || got.HasNewer != want.HasNewer {
				t.Fatalf("window = %+v, want %+v", got, want)
			}
			if err := schema.ValidatePositionedWindowCoverage(got, refs); err != nil {
				t.Fatalf("resolved window fails coverage: %v", err)
			}
			if c.Input.CompareCount > 0 {
				start, end := c.Input.CompareOffset, c.Input.CompareOffset+c.Input.CompareCount
				if start < 0 || end > len(refs) {
					t.Fatal("fixture's comparison interval is outside the refs")
				}
				input.Refs = refs[start:end]
				fromPage, err := schema.ResolvePositionedWindow(input)
				if err != nil || !reflect.DeepEqual(fromPage, got) {
					t.Fatalf("page window = %+v, error %v; want whole-directory window %+v", fromPage, err, got)
				}
			}
		})
	}
}
