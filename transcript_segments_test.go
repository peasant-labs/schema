package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/transcript_segments.yaml
var transcriptSegmentsYAML []byte

//go:embed testdata/transcript_sparse_session.yaml
var transcriptSparseSessionYAML []byte

type transcriptEntryFixture struct {
	Name  string `yaml:"name"`
	Bytes int    `yaml:"bytes"`
}

type transcriptCutTurn struct {
	Index   int                      `yaml:"index"`
	Role    string                   `yaml:"role"`
	Entries []transcriptEntryFixture `yaml:"entries"`
}

// size is the pre-bound measure the cut applies: every entry of the turn.
func (t transcriptCutTurn) size() int {
	total := 0
	for _, entry := range t.Entries {
		total += entry.Bytes
	}
	return total
}

func (t transcriptCutTurn) hasEntry(name string) bool {
	for _, entry := range t.Entries {
		if entry.Name == name {
			return true
		}
	}
	return false
}

type transcriptCutSegment struct {
	Index     int `yaml:"index"`
	FirstTurn int `yaml:"firstTurn"`
	LastTurn  int `yaml:"lastTurn"`
}

type transcriptCrossing struct {
	Turn  int    `yaml:"turn"`
	Entry string `yaml:"entry"`
}

type transcriptCutExpected struct {
	Segments []transcriptCutSegment `yaml:"segments"`
	Crossing transcriptCrossing     `yaml:"crossing"`
}

type transcriptCutCase struct {
	Name        string                `yaml:"name"`
	BudgetBytes int                   `yaml:"budgetBytes"`
	Turns       []transcriptCutTurn   `yaml:"turns"`
	Expected    transcriptCutExpected `yaml:"expected"`
}

type transcriptRefFixture struct {
	Index      int `yaml:"index"`
	FirstTurn  int `yaml:"firstTurn"`
	LastTurn   int `yaml:"lastTurn"`
	ByteLength int `yaml:"byteLength"`
}

func (r transcriptRefFixture) toRef() schema.TranscriptSegmentRef {
	return schema.TranscriptSegmentRef{Index: r.Index, FirstTurn: r.FirstTurn, LastTurn: r.LastTurn, ByteLength: r.ByteLength}
}

type transcriptShapeInput struct {
	Schema        string                 `yaml:"schema"`
	Payload       map[string]any         `yaml:"payload"`
	DirectoryRefs []transcriptRefFixture `yaml:"directoryRefs"`
}

type transcriptDirectoryCap struct {
	RefCount         int   `yaml:"ref_count"`
	PerPage          int   `yaml:"per_page"`
	TotalSegments    int   `yaml:"total_segments"`
	Pages            []int `yaml:"pages"`
	LastPageRefRange []int `yaml:"last_page_ref_range"`
}

type transcriptSegmentsFixtures struct {
	RequiredNames []string                                    `yaml:"required_names"`
	ShapeCases    testcase.Corpus[transcriptShapeInput, bool] `yaml:"shape_cases"`
	CutCases      []transcriptCutCase                         `yaml:"cut_cases"`
	DirectoryCap  transcriptDirectoryCap                      `yaml:"directory_cap"`
}

func loadTranscriptSegmentsFixtures(t *testing.T) transcriptSegmentsFixtures {
	t.Helper()
	var fixtures transcriptSegmentsFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(transcriptSegmentsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("load transcript segments contract: %v", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("transcript segments fixture must contain one document: %v", err)
	}
	assertcase.RequireMin(t, fixtures.ShapeCases, 8)
	assertcase.RequireValid(t, fixtures.ShapeCases)
	if len(fixtures.CutCases) < 3 {
		t.Fatalf("transcript segments fixture has %d cut cases; at least three arms are required", len(fixtures.CutCases))
	}
	seen := map[string]bool{}
	for _, c := range fixtures.ShapeCases.Cases {
		if seen[c.Name] {
			t.Fatalf("transcript segments fixture repeats case %q", c.Name)
		}
		seen[c.Name] = true
	}
	for _, c := range fixtures.CutCases {
		if seen[c.Name] {
			t.Fatalf("transcript segments fixture repeats case %q", c.Name)
		}
		seen[c.Name] = true
	}
	for _, name := range fixtures.RequiredNames {
		if !seen[name] {
			t.Fatalf("required transcript segment case %q is missing", name)
		}
	}
	if len(seen) != len(fixtures.RequiredNames) {
		t.Fatalf("transcript segment required-name manifest has %d names for %d cases", len(fixtures.RequiredNames), len(seen))
	}
	return fixtures
}

// TestTranscriptShapeValidation drives the wire validators and the generated
// Types schema over one case table. Where the Go validator accepts a payload the
// emitted schema must accept it too, so the two language surfaces cannot drift.
func TestTranscriptShapeValidation(t *testing.T) {
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Types spec: %v", err)
	}
	compiler := schema.NewJSONSchemaCompiler()
	if err := compiler.AddResource("types.json", bytes.NewReader(raw)); err != nil {
		t.Fatalf("add Types schema resource: %v", err)
	}
	contracts := map[string]*jsonschema.Schema{}
	for _, name := range []string{"TranscriptSegment", "TranscriptSegmentDirectory", "TranscriptWindow"} {
		contract, err := compiler.Compile("types.json#/components/schemas/" + name)
		if err != nil {
			t.Fatalf("compile %s contract: %v", name, err)
		}
		contracts[name] = contract
	}

	for _, c := range loadTranscriptSegmentsFixtures(t).ShapeCases.Cases {
		t.Run(c.Name, func(t *testing.T) {
			value, decodeErr := decodeTranscriptShape(c.Input)
			valid := decodeErr == nil
			if valid != c.Expected {
				t.Fatalf("valid=%v, want %v: %v", valid, c.Expected, decodeErr)
			}
			contract, ok := contracts[c.Input.Schema]
			if !ok {
				t.Fatalf("fixture names unknown schema %q", c.Input.Schema)
			}
			schemaErr := contract.Validate(c.Input.Payload)
			if valid && schemaErr != nil {
				t.Fatalf("Go validator accepts a payload the Types schema rejects: %v", schemaErr)
			}
			if c.Expected {
				if schemaErr != nil {
					t.Fatalf("must-pass payload fails the Types schema: %v", schemaErr)
				}
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var output any
				if err := json.Unmarshal(encoded, &output); err != nil {
					t.Fatal(err)
				}
				if err := requireJSONSubset(c.Input.Payload, output, "$"); err != nil {
					t.Fatalf("round trip changed values: %v", err)
				}
			}
		})
	}
}

func decodeTranscriptShape(input transcriptShapeInput) (any, error) {
	raw, err := json.Marshal(input.Payload)
	if err != nil {
		return nil, err
	}
	switch input.Schema {
	case "TranscriptSegment":
		value := &schema.TranscriptSegment{}
		if err := json.Unmarshal(raw, value); err != nil {
			return nil, err
		}
		return value, value.Validate()
	case "TranscriptSegmentDirectory":
		value := &schema.TranscriptSegmentDirectory{}
		if err := json.Unmarshal(raw, value); err != nil {
			return nil, err
		}
		return value, value.Validate()
	case "TranscriptWindow":
		value := &schema.TranscriptWindow{}
		if err := json.Unmarshal(raw, value); err != nil {
			return nil, err
		}
		if err := value.Validate(); err != nil {
			return nil, err
		}
		if len(input.DirectoryRefs) > 0 {
			refs := make([]schema.TranscriptSegmentRef, len(input.DirectoryRefs))
			for i, ref := range input.DirectoryRefs {
				refs[i] = ref.toRef()
			}
			if err := schema.ValidateTranscriptWindowCoverage(*value, refs); err != nil {
				return nil, err
			}
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unknown transcript shape fixture schema %q", input.Schema)
	}
}

// TestTranscriptCutExpectations asserts the recorded segmentation expectations
// on the wire shapes: whole turns only, the crossing entry deferred to the next
// segment, no segment past the budget except one holding an oversized turn, and
// one directory ref per segment.
func TestTranscriptCutExpectations(t *testing.T) {
	const sessionID = schema.SessionID("ses_3cd91f52effeXd3QAJ54jOyzv5")
	const revision = "4213"
	timestamp := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)

	for _, c := range loadTranscriptSegmentsFixtures(t).CutCases {
		t.Run(c.Name, func(t *testing.T) {
			if c.BudgetBytes <= 0 {
				t.Fatalf("cut case declares budget %d; the budget must be positive", c.BudgetBytes)
			}
			if len(c.Turns) == 0 || len(c.Expected.Segments) == 0 {
				t.Fatal("cut case must declare turns and expected segments")
			}
			owner := map[int]int{}
			for i, segment := range c.Expected.Segments {
				if segment.Index != i {
					t.Fatalf("expected segment %d declares index %d; segment ordinals are contiguous from zero", i, segment.Index)
				}
				var turns []transcriptCutTurn
				for _, turn := range c.Turns {
					if segment.FirstTurn <= turn.Index && turn.Index <= segment.LastTurn {
						turns = append(turns, turn)
					}
				}
				if len(turns) == 0 {
					t.Fatalf("segment %d declares range [%d, %d] with no turns", segment.Index, segment.FirstTurn, segment.LastTurn)
				}
				if turns[0].Index != segment.FirstTurn || turns[len(turns)-1].Index != segment.LastTurn {
					t.Fatalf("segment %d declares [%d, %d] but its turns span [%d, %d]", segment.Index, segment.FirstTurn, segment.LastTurn, turns[0].Index, turns[len(turns)-1].Index)
				}
				total := 0
				for _, turn := range turns {
					if previous, exists := owner[turn.Index]; exists {
						t.Fatalf("turn %d belongs to segments %d and %d; a turn must belong to exactly one segment", turn.Index, previous, segment.Index)
					}
					owner[turn.Index] = segment.Index
					total += turn.size()
				}
				if i > 0 && c.Expected.Segments[i-1].LastTurn >= segment.FirstTurn {
					t.Fatalf("segment %d starts at turn %d, not after segment %d ending at %d", segment.Index, segment.FirstTurn, c.Expected.Segments[i-1].Index, c.Expected.Segments[i-1].LastTurn)
				}
				oversized := len(turns) == 1 && turns[0].size() > c.BudgetBytes
				if !oversized && total > c.BudgetBytes {
					t.Fatalf("segment %d carries %d bytes above the %d-byte budget without holding one oversized turn", segment.Index, total, c.BudgetBytes)
				}
			}
			if len(owner) != len(c.Turns) {
				t.Fatalf("expected segments cover %d of %d turns", len(owner), len(c.Turns))
			}
			for _, turn := range c.Turns {
				if turn.Role == "user" {
					if _, ok := owner[turn.Index]; !ok {
						t.Fatalf("user turn %d is not contained by one segment", turn.Index)
					}
				}
			}

			// The crossing entry is absent from the earlier segment and first in the
			// segment that starts with the turn that did not fit.
			crossingSegment := -1
			for i, segment := range c.Expected.Segments {
				if segment.FirstTurn <= c.Expected.Crossing.Turn && c.Expected.Crossing.Turn <= segment.LastTurn {
					crossingSegment = i
				}
			}
			if crossingSegment <= 0 {
				t.Fatalf("crossing turn %d must start a segment after the first; got segment %d", c.Expected.Crossing.Turn, crossingSegment)
			}
			for _, turn := range c.Turns {
				previous := c.Expected.Segments[crossingSegment-1]
				if previous.FirstTurn <= turn.Index && turn.Index <= previous.LastTurn && turn.hasEntry(c.Expected.Crossing.Entry) {
					t.Fatalf("crossing entry %q appears in segment %d; the entry that did not fit starts the next segment", c.Expected.Crossing.Entry, previous.Index)
				}
			}
			var crossingTurn *transcriptCutTurn
			for i := range c.Turns {
				if c.Turns[i].Index == c.Expected.Crossing.Turn {
					crossingTurn = &c.Turns[i]
				}
			}
			if crossingTurn == nil || len(crossingTurn.Entries) == 0 || crossingTurn.Entries[0].Name != c.Expected.Crossing.Entry {
				t.Fatalf("crossing turn %d does not start with entry %q", c.Expected.Crossing.Turn, c.Expected.Crossing.Entry)
			}

			// The same expectations hold on the wire shape, not only on the fixture.
			refs := make([]schema.TranscriptSegmentRef, 0, len(c.Expected.Segments))
			for _, segment := range c.Expected.Segments {
				var turns []schema.TranscriptTurn
				total := 0
				for _, turn := range c.Turns {
					if turn.Index < segment.FirstTurn || segment.LastTurn < turn.Index {
						continue
					}
					total += turn.size()
					turns = append(turns, schema.TranscriptTurn{
						TurnDetail: schema.TurnDetail{
							Index:     turn.Index,
							Role:      schema.Role(turn.Role),
							Content:   "recorded fixture turn",
							Timestamp: timestamp,
							Depth:     0,
						},
						Truncated:  len(turns) == 0 && turn.size() > c.BudgetBytes,
						ByteLength: turn.size(),
					})
				}
				value := schema.TranscriptSegment{
					Index:      segment.Index,
					Revision:   revision,
					FirstTurn:  segment.FirstTurn,
					LastTurn:   segment.LastTurn,
					ByteLength: total,
					HasOlder:   segment.Index > 0,
					HasNewer:   segment.Index < len(c.Expected.Segments)-1,
					Turns:      turns,
				}
				if err := value.Validate(); err != nil {
					t.Fatalf("recorded segment %d is not a valid wire segment: %v", segment.Index, err)
				}
				refs = append(refs, schema.TranscriptSegmentRef{Index: segment.Index, FirstTurn: segment.FirstTurn, LastTurn: segment.LastTurn, ByteLength: total})
			}
			directory := schema.TranscriptSegmentDirectory{
				SessionID:     sessionID,
				Revision:      revision,
				TotalTurns:    len(c.Turns),
				TotalSegments: len(refs),
				Segments:      refs,
			}
			if err := directory.Validate(); err != nil {
				t.Fatalf("directory derived from the cut is invalid: %v", err)
			}
			if len(directory.Segments) != len(c.Expected.Segments) {
				t.Fatalf("directory carries %d refs for %d segments", len(directory.Segments), len(c.Expected.Segments))
			}
			for i, ref := range directory.Segments {
				want := c.Expected.Segments[i]
				if ref.Index != want.Index || ref.FirstTurn != want.FirstTurn || ref.LastTurn != want.LastTurn {
					t.Fatalf("directory ref %d is %+v, want extent [%d, %d] at index %d", i, ref, want.FirstTurn, want.LastTurn, want.Index)
				}
			}
		})
	}
}

// TestTranscriptDirectoryCap asserts the one paged directory shape: ten thousand
// refs in a page, the true whole-session total retained, and a short last page.
func TestTranscriptDirectoryCap(t *testing.T) {
	fixtures := loadTranscriptSegmentsFixtures(t)
	capCase := fixtures.DirectoryCap
	if capCase.RefCount != schema.MaxTranscriptDirectoryRefs {
		t.Fatalf("directory cap case serves %d refs, want the %d-ref contract cap", capCase.RefCount, schema.MaxTranscriptDirectoryRefs)
	}
	if capCase.PerPage != schema.MaxTranscriptDirectoryRefs {
		t.Fatalf("directory cap case declares perPage %d, want %d", capCase.PerPage, schema.MaxTranscriptDirectoryRefs)
	}
	if len(capCase.Pages) == 0 {
		t.Fatal("directory cap case names no pages")
	}
	for _, page := range capCase.Pages {
		start := (page - 1) * capCase.PerPage
		count := capCase.RefCount
		if start+count > capCase.TotalSegments {
			count = capCase.TotalSegments - start
		}
		refs := make([]schema.TranscriptSegmentRef, 0, count)
		for i := 0; i < count; i++ {
			index := start + i
			refs = append(refs, schema.TranscriptSegmentRef{Index: index, FirstTurn: index * 2, LastTurn: index*2 + 1, ByteLength: 10})
		}
		pageNumber := page
		perPage := capCase.PerPage
		directory := schema.TranscriptSegmentDirectory{
			SessionID:        "ses_3cd91f52effeXd3QAJ54jOyzv5",
			Revision:         "4213",
			TotalTurns:       capCase.TotalSegments,
			TotalSegments:    capCase.TotalSegments,
			Segments:         refs,
			DirectoryPage:    &pageNumber,
			DirectoryPerPage: &perPage,
		}
		if err := directory.Validate(); err != nil {
			t.Fatalf("directory page %d is invalid: %v", page, err)
		}
		if page == len(capCase.Pages) && len(capCase.LastPageRefRange) == 2 {
			first, last := refs[0].Index, refs[len(refs)-1].Index
			if first != capCase.LastPageRefRange[0] || last != capCase.LastPageRefRange[1] {
				t.Fatalf("last page holds refs %d..%d, want %d..%d", first, last, capCase.LastPageRefRange[0], capCase.LastPageRefRange[1])
			}
		}
	}
	oversized := make([]schema.TranscriptSegmentRef, 0, capCase.RefCount+1)
	for i := 0; i <= capCase.RefCount; i++ {
		oversized = append(oversized, schema.TranscriptSegmentRef{Index: i, FirstTurn: i * 2, LastTurn: i*2 + 1, ByteLength: 10})
	}
	unpaged := schema.TranscriptSegmentDirectory{
		SessionID:     "ses_3cd91f52effeXd3QAJ54jOyzv5",
		Revision:      "4213",
		TotalTurns:    len(oversized),
		TotalSegments: len(oversized),
		Segments:      oversized,
	}
	if err := unpaged.Validate(); err == nil {
		t.Fatalf("a %d-ref unpaged directory is accepted; the cap is %d", len(oversized), schema.MaxTranscriptDirectoryRefs)
	}
}

// TestTranscriptContractCarriesNoPageCoordinate asserts that the transcript
// surfaces name segments and turn identities only: no page, page size, position,
// page count, or per-turn entry cursor is reintroduced, and the directory and
// window carry no turn content.
func TestTranscriptContractCarriesNoPageCoordinate(t *testing.T) {
	values := map[string]any{
		"TranscriptSegmentDirectory": schema.TranscriptSegmentDirectory{},
		"TranscriptSegmentRef":       schema.TranscriptSegmentRef{},
		"TranscriptSegment":          schema.TranscriptSegment{},
		"TranscriptTurn":             schema.TranscriptTurn{},
		"TranscriptWindow":           schema.TranscriptWindow{},
		"SessionDetailHeader":        schema.SessionDetailHeader{},
		"SessionDetailHeaderPayload": schema.SessionDetailHeaderPayload{},
	}
	forbidden := []string{"page", "perPage", "pageCount", "position", "startEntry"}
	for name, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		var tree any
		if err := json.Unmarshal(encoded, &tree); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		for _, key := range forbidden {
			if hasJSONKey(tree, key) {
				t.Fatalf("%s carries %q; the transcript has no page and no ordinal position", name, key)
			}
		}
		if name == "TranscriptSegmentDirectory" || name == "TranscriptWindow" {
			if hasJSONKey(tree, "turns") {
				t.Fatalf("%s carries turn content; only the segment read carries turns", name)
			}
		}
	}
}

type transcriptSparseTurn struct {
	Index     int    `yaml:"index"`
	Role      string `yaml:"role"`
	ToolCalls int    `yaml:"toolCalls"`
}

type transcriptSparseSegment struct {
	Index      int `yaml:"index"`
	FirstTurn  int `yaml:"firstTurn"`
	LastTurn   int `yaml:"lastTurn"`
	ByteLength int `yaml:"byteLength"`
}

type transcriptSparseWindow struct {
	RequestedAnchor int   `yaml:"requestedAnchor"`
	AnchorTurn      int   `yaml:"anchorTurn"`
	Span            int   `yaml:"span"`
	FirstTurn       int   `yaml:"firstTurn"`
	LastTurn        int   `yaml:"lastTurn"`
	Segments        []int `yaml:"segments"`
}

type transcriptSparseSession struct {
	SessionID       string                    `yaml:"session_id"`
	Revision        string                    `yaml:"revision"`
	TotalTurns      int                       `yaml:"total_turns"`
	ToolCallCount   int                       `yaml:"tool_call_count"`
	GroupIdentities []int                     `yaml:"group_identities"`
	Turns           []transcriptSparseTurn    `yaml:"turns"`
	Segments        []transcriptSparseSegment `yaml:"segments"`
	Window          transcriptSparseWindow    `yaml:"window"`
}

func loadTranscriptSparseSession(t *testing.T) transcriptSparseSession {
	t.Helper()
	var fixture transcriptSparseSession
	decoder := yaml.NewDecoder(bytes.NewReader(transcriptSparseSessionYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("load sparse transcript session: %v", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("sparse transcript fixture must contain one document: %v", err)
	}
	return fixture
}

// TestTranscriptSparseSessionKeepsIdentity asserts that turn identities stay
// sparse on the wire: a directory addresses segments by their own ordinal and
// turns by their identity, and no position or page arithmetic appears anywhere.
func TestTranscriptSparseSessionKeepsIdentity(t *testing.T) {
	fixture := loadTranscriptSparseSession(t)
	if len(fixture.Turns) != fixture.TotalTurns {
		t.Fatalf("fixture declares %d turns but carries %d", fixture.TotalTurns, len(fixture.Turns))
	}
	previous := -1
	assistants := map[int]bool{}
	toolCalls := 0
	role := map[int]string{}
	for _, turn := range fixture.Turns {
		if turn.Index <= previous {
			t.Fatalf("turn identity %d does not follow %d; identities are strictly increasing", turn.Index, previous)
		}
		previous = turn.Index
		role[turn.Index] = turn.Role
		if turn.Role == "assistant" {
			assistants[turn.Index] = true
		}
		toolCalls += turn.ToolCalls
	}
	if toolCalls != fixture.ToolCallCount {
		t.Fatalf("fixture turns carry %d tool calls, want %d", toolCalls, fixture.ToolCallCount)
	}
	if fixture.Turns[len(fixture.Turns)-1].Index <= fixture.TotalTurns {
		t.Fatalf("last identity %d is not sparse against %d turns; the fixture must exercise a gap", fixture.Turns[len(fixture.Turns)-1].Index, fixture.TotalTurns)
	}
	for _, identity := range fixture.GroupIdentities {
		if _, ok := role[identity]; !ok {
			t.Fatalf("group identity %d is absent from the recorded turns", identity)
		}
	}

	refs := make([]schema.TranscriptSegmentRef, len(fixture.Segments))
	for i, segment := range fixture.Segments {
		refs[i] = schema.TranscriptSegmentRef{Index: segment.Index, FirstTurn: segment.FirstTurn, LastTurn: segment.LastTurn, ByteLength: segment.ByteLength}
		if role[segment.FirstTurn] == "" {
			t.Fatalf("segment %d starts at identity %d, which is not a recorded turn", segment.Index, segment.FirstTurn)
		}
		if role[segment.LastTurn] == "" {
			t.Fatalf("segment %d ends at identity %d, which is not a recorded turn", segment.Index, segment.LastTurn)
		}
	}
	directory := schema.TranscriptSegmentDirectory{
		SessionID:     schema.SessionID(fixture.SessionID),
		Revision:      fixture.Revision,
		TotalTurns:    fixture.TotalTurns,
		TotalSegments: len(refs),
		Segments:      refs,
	}
	if err := directory.Validate(); err != nil {
		t.Fatalf("sparse directory is invalid: %v", err)
	}

	window := schema.TranscriptWindow{
		SessionID:       schema.SessionID(fixture.SessionID),
		Revision:        fixture.Revision,
		RequestedAnchor: fixture.Window.RequestedAnchor,
		AnchorTurn:      fixture.Window.AnchorTurn,
		Span:            fixture.Window.Span,
		FirstTurn:       fixture.Window.FirstTurn,
		LastTurn:        fixture.Window.LastTurn,
		Segments:        fixture.Window.Segments,
		HasOlder:        fixture.Window.FirstTurn > fixture.Turns[0].Index,
		HasNewer:        fixture.Window.LastTurn < fixture.Turns[len(fixture.Turns)-1].Index,
	}
	if err := window.Validate(); err != nil {
		t.Fatalf("sparse window descriptor is invalid: %v", err)
	}
	if err := schema.ValidateTranscriptWindowCoverage(window, refs); err != nil {
		t.Fatalf("sparse window does not cover its declared segment run: %v", err)
	}
	before, after := 0, 0
	for _, turn := range fixture.Turns {
		if turn.Index < window.AnchorTurn && window.FirstTurn <= turn.Index && assistants[turn.Index] {
			before++
		}
		if window.AnchorTurn < turn.Index && turn.Index <= window.LastTurn && assistants[turn.Index] {
			after++
		}
	}
	if before != window.Span || after != window.Span {
		t.Fatalf("window collects %d assistant turns before and %d after the anchor, want %d per side", before, after, window.Span)
	}

	encoded, err := json.Marshal(window)
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &keys); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"turns", "content", "toolCalls", "position", "page", "perPage"} {
		if _, present := keys[forbidden]; present {
			t.Fatalf("window descriptor carries %q; the window path must carry no content and no page coordinate", forbidden)
		}
	}
}
