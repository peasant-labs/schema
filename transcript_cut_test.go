package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/transcript_cut.yaml
var transcriptCutYAML []byte

//go:embed testdata/transcript_cut_manifest.yaml
var transcriptCutManifestYAML []byte

//go:embed testdata/transcript_windows.yaml
var transcriptWindowsYAML []byte

//go:embed testdata/transcript_windows_manifest.yaml
var transcriptWindowsManifestYAML []byte

type cutToolCallSpec struct {
	Arguments string `yaml:"arguments"`
	Result    string `yaml:"result"`
}

type cutTurnSpec struct {
	Index          int               `yaml:"index"`
	Role           string            `yaml:"role"`
	EntryType      string            `yaml:"entryType,omitempty"`
	ContentBytes   int               `yaml:"contentBytes,omitempty"`
	Content        string            `yaml:"content,omitempty"`
	ContentHex     string            `yaml:"contentHex,omitempty"`
	ArgumentsBytes []int             `yaml:"argumentsBytes,omitempty"`
	ResultBytes    []int             `yaml:"resultBytes,omitempty"`
	ToolCalls      []cutToolCallSpec `yaml:"toolCalls,omitempty"`
	Parts          []cutPartSpec     `yaml:"parts,omitempty"`
}

type cutImageSpec struct {
	Ref           string `yaml:"ref,omitempty"`
	MediaType     string `yaml:"mediaType,omitempty"`
	ByteSize      int    `yaml:"byteSize,omitempty"`
	Width         int    `yaml:"width,omitempty"`
	Height        int    `yaml:"height,omitempty"`
	OmittedReason string `yaml:"omittedReason,omitempty"`
}

type cutPartSpec struct {
	Type  string        `yaml:"type"`
	Text  string        `yaml:"text,omitempty"`
	Image *cutImageSpec `yaml:"image,omitempty"`
}

func (s cutPartSpec) part() (schema.TurnContentPart, error) {
	part := schema.TurnContentPart{Type: schema.TurnPartType(s.Type), Text: s.Text}
	if !part.Type.IsValid() {
		return part, fmt.Errorf("cut fixture part names type %q; use text or image", s.Type)
	}
	if s.Image != nil {
		part.Image = &schema.ImagePart{
			Ref:           s.Image.Ref,
			MediaType:     schema.ImageMediaType(s.Image.MediaType),
			ByteSize:      s.Image.ByteSize,
			Width:         s.Image.Width,
			Height:        s.Image.Height,
			OmittedReason: schema.OmittedReason(s.Image.OmittedReason),
		}
	}
	return part, nil
}

func (s cutTurnSpec) turn() (schema.TurnDetail, error) {
	role := schema.Role(s.Role)
	if !slices.Contains(schema.AllRoles, role) {
		return schema.TurnDetail{}, fmt.Errorf("cut fixture turn %d names role %q; use a member of schema.AllRoles", s.Index, s.Role)
	}
	entryType := schema.EntryTypeText
	if s.EntryType != "" {
		entryType = schema.EntryType(s.EntryType)
		if !slices.Contains(schema.AllEntryTypes, entryType) {
			return schema.TurnDetail{}, fmt.Errorf("cut fixture turn %d names entry type %q; use a member of schema.AllEntryTypes", s.Index, s.EntryType)
		}
	}
	stated := 0
	content := ""
	if s.ContentBytes != 0 {
		stated++
		content = strings.Repeat("x", s.ContentBytes)
	}
	if s.Content != "" {
		stated++
		content = s.Content
	}
	if s.ContentHex != "" {
		stated++
		raw, err := hex.DecodeString(s.ContentHex)
		if err != nil {
			return schema.TurnDetail{}, fmt.Errorf("cut fixture turn %d contentHex: %w", s.Index, err)
		}
		content = string(raw)
	}
	if stated > 1 {
		return schema.TurnDetail{}, fmt.Errorf("cut fixture turn %d states its text %d ways; use one of contentBytes, content and contentHex", s.Index, stated)
	}
	if len(s.ArgumentsBytes) != len(s.ResultBytes) {
		return schema.TurnDetail{}, fmt.Errorf("cut fixture turn %d states %d tool arguments and %d results", s.Index, len(s.ArgumentsBytes), len(s.ResultBytes))
	}
	turn := schema.TurnDetail{Index: s.Index, Role: role, EntryType: entryType, Content: content}
	for _, spec := range s.Parts {
		part, err := spec.part()
		if err != nil {
			return schema.TurnDetail{}, fmt.Errorf("cut fixture turn %d: %w", s.Index, err)
		}
		turn.ContentParts = append(turn.ContentParts, part)
	}
	for i := range s.ArgumentsBytes {
		turn.ToolCalls = append(turn.ToolCalls, schema.ToolCallDetail{
			Arguments: strings.Repeat("x", s.ArgumentsBytes[i]),
			Result:    strings.Repeat("x", s.ResultBytes[i]),
		})
	}
	for _, call := range s.ToolCalls {
		turn.ToolCalls = append(turn.ToolCalls, schema.ToolCallDetail{Arguments: call.Arguments, Result: call.Result})
	}
	return turn, nil
}

type cutSpanExpectation struct {
	Before int `yaml:"before"`
	Count  int `yaml:"count"`
}

type cutRefExpectation struct {
	Index             int                `yaml:"index"`
	FirstTurn         int                `yaml:"firstTurn"`
	LastTurn          int                `yaml:"lastTurn"`
	ByteLength        int                `yaml:"byteLength"`
	ConversationTurns cutSpanExpectation `yaml:"conversationTurns"`
}

type cutExpectation struct {
	Segments []cutRefExpectation `yaml:"segments"`
	// CrossingTurn is the turn that did not fit and opened a later segment.
	CrossingTurn *int `yaml:"crossingTurn,omitempty"`
	// UncountedSegments is the least number of segments that must hold a system
	// or tool row, so a position case cannot pass on a cut with none.
	UncountedSegments int `yaml:"uncountedSegments,omitempty"`
	// TurnByteLengths is TranscriptTurnByteLength of every turn, in order.
	TurnByteLengths []int `yaml:"turnByteLengths,omitempty"`
}

type cutCase struct {
	Name        string         `yaml:"name"`
	BudgetBytes int            `yaml:"budgetBytes"`
	Turns       []cutTurnSpec  `yaml:"turns"`
	Expected    cutExpectation `yaml:"expected"`
}

type cutFixture struct {
	Cases []cutCase `yaml:"cases"`
}

type requiredNamesManifest struct {
	RequiredNames []string `yaml:"requiredNames"`
}

func decodeStrictYAML[T any](t *testing.T, raw []byte, label string) T {
	t.Helper()
	var value T
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode %s: %v", label, err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		t.Fatalf("%s must contain exactly one YAML document: %v", label, err)
	}
	return value
}

// requireManifestNames fails when a required case is missing, a case is not in
// the manifest, or a name repeats, so a row cannot be deleted or added silently.
func requireManifestNames(t *testing.T, label string, manifestRaw []byte, names []string) {
	t.Helper()
	manifest := decodeStrictYAML[requiredNamesManifest](t, manifestRaw, label+" manifest")
	if len(manifest.RequiredNames) == 0 {
		t.Fatalf("%s manifest names no cases", label)
	}
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] {
			t.Fatalf("%s fixture has an empty or repeated case name %q", label, name)
		}
		seen[name] = true
	}
	required := map[string]bool{}
	for _, name := range manifest.RequiredNames {
		if required[name] {
			t.Fatalf("%s manifest repeats %q", label, name)
		}
		required[name] = true
		if !seen[name] {
			t.Fatalf("%s fixture is missing required case %q", label, name)
		}
	}
	for _, name := range names {
		if !required[name] {
			t.Fatalf("%s case %q is not in the manifest; add it there", label, name)
		}
	}
}

func loadCutFixture(t *testing.T) cutFixture {
	t.Helper()
	fixture := decodeStrictYAML[cutFixture](t, transcriptCutYAML, "transcript cut fixture")
	names := make([]string, len(fixture.Cases))
	for i, c := range fixture.Cases {
		names[i] = c.Name
		if c.BudgetBytes <= 0 {
			t.Fatalf("transcript cut case %q states no budget", c.Name)
		}
	}
	requireManifestNames(t, "transcript cut", transcriptCutManifestYAML, names)
	return fixture
}

func refExpectation(ref schema.TranscriptSegmentRef) (cutRefExpectation, bool) {
	if ref.ConversationTurns == nil {
		return cutRefExpectation{}, false
	}
	return cutRefExpectation{
		Index:             ref.Index,
		FirstTurn:         ref.FirstTurn,
		LastTurn:          ref.LastTurn,
		ByteLength:        ref.ByteLength,
		ConversationTurns: cutSpanExpectation{Before: ref.ConversationTurns.Before, Count: ref.ConversationTurns.Count},
	}, true
}

// TestCutTranscriptSegments replays the cut cases. The refs are compared whole,
// so a moved boundary, a changed byte length or a changed position each fail.
func TestCutTranscriptSegments(t *testing.T) {
	for _, c := range loadCutFixture(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			turns := make([]schema.TurnDetail, len(c.Turns))
			for i, spec := range c.Turns {
				turn, err := spec.turn()
				if err != nil {
					t.Fatal(err)
				}
				turns[i] = turn
			}
			if c.Expected.TurnByteLengths != nil {
				got := make([]int, len(turns))
				for i := range turns {
					got[i] = schema.TranscriptTurnByteLength(turns[i])
				}
				if !slices.Equal(got, c.Expected.TurnByteLengths) {
					t.Fatalf("turn byte lengths = %v, want %v", got, c.Expected.TurnByteLengths)
				}
			}

			refs := schema.CutTranscriptSegments(turns, c.BudgetBytes)
			if refs == nil {
				t.Fatal("the cut returned a nil slice; a directory encodes segments as an array, never null")
			}
			got := make([]cutRefExpectation, len(refs))
			for i, ref := range refs {
				row, ok := refExpectation(ref)
				if !ok {
					t.Fatalf("segment %d carries no conversationTurns span", ref.Index)
				}
				got[i] = row
			}
			want := c.Expected.Segments
			if want == nil {
				want = []cutRefExpectation{}
			}
			if !slices.Equal(got, want) {
				t.Fatalf("refs = %+v\nwant   %+v", got, want)
			}

			// The directory the cut produces is one the wire validator accepts.
			total := 0
			for _, ref := range refs {
				total += ref.ConversationTurns.Count
			}
			directory := schema.TranscriptSegmentDirectory{
				SessionID: "ses_cut", Revision: "1", TotalTurns: len(turns), TotalSegments: len(refs),
				Segments: refs, TotalConversationTurns: &total,
			}
			if err := directory.Validate(); err != nil {
				t.Fatalf("the cut's directory does not validate: %v", err)
			}

			if c.Expected.CrossingTurn != nil {
				opened := false
				for _, ref := range refs {
					if ref.Index > 0 && ref.FirstTurn == *c.Expected.CrossingTurn {
						opened = true
					}
				}
				if !opened {
					t.Fatalf("turn %d did not open a later segment: %+v", *c.Expected.CrossingTurn, got)
				}
			}
			if c.Expected.UncountedSegments > 0 {
				holding := 0
				for _, ref := range refs {
					for _, turn := range turns {
						if turn.Index >= ref.FirstTurn && turn.Index <= ref.LastTurn && (turn.Role == schema.RoleSystem || turn.Role == schema.RoleTool) {
							holding++
							break
						}
					}
				}
				if holding < c.Expected.UncountedSegments {
					t.Fatalf("%d segments hold a system or tool row, want at least %d; the case would pass on a cut with no uncounted rows", holding, c.Expected.UncountedSegments)
				}
			}
		})
	}
}

// TestTranscriptJSONStringByteLength holds the exported string measure equal to
// the encoder it stands for and to the turn measure built from it.
func TestTranscriptJSONStringByteLength(t *testing.T) {
	for _, c := range loadCutFixture(t).Cases {
		for _, spec := range c.Turns {
			turn, err := spec.turn()
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(turn.Content)
			if err != nil {
				t.Fatal(err)
			}
			if got := schema.TranscriptJSONStringByteLength(turn.Content); got != len(encoded) {
				t.Fatalf("%s turn %d: string measure = %d, the JSON encoder writes %d bytes", c.Name, turn.Index, got, len(encoded))
			}
			if len(turn.ContentParts) == 0 && len(turn.ToolCalls) == 0 {
				if got := schema.TranscriptTurnByteLength(turn); got != len(encoded) {
					t.Fatalf("%s turn %d: turn measure = %d, want its content's %d", c.Name, turn.Index, got, len(encoded))
				}
			}
		}
	}
}

type windowTurnSpec struct {
	Index     int    `yaml:"index"`
	Role      string `yaml:"role"`
	ToolCalls int    `yaml:"toolCalls"`
}

type windowExpectation struct {
	AnchorTurn int  `yaml:"anchorTurn"`
	FirstTurn  int  `yaml:"firstTurn"`
	LastTurn   int  `yaml:"lastTurn"`
	HasOlder   bool `yaml:"hasOlder"`
	HasNewer   bool `yaml:"hasNewer"`
	// Error names the refusal a case expects: no-user-turn or span-out-of-range.
	Error string `yaml:"error,omitempty"`
}

type windowCase struct {
	Name   string `yaml:"name"`
	Anchor int    `yaml:"anchor"`
	Span   int    `yaml:"span"`
	// Turns replaces the sparse session for this case.
	Turns    []windowTurnSpec  `yaml:"turns,omitempty"`
	Expected windowExpectation `yaml:"expected"`
}

type windowFixture struct {
	Cases                  []windowCase `yaml:"cases"`
	CrossesSegmentBoundary []string     `yaml:"crossesSegmentBoundary"`
}

type sparseSessionFixture struct {
	SessionID string           `yaml:"session_id"`
	Revision  string           `yaml:"revision"`
	Turns     []windowTurnSpec `yaml:"turns"`
	Segments  []struct {
		Index      int `yaml:"index"`
		FirstTurn  int `yaml:"firstTurn"`
		LastTurn   int `yaml:"lastTurn"`
		ByteLength int `yaml:"byteLength"`
	} `yaml:"segments"`
	Window struct {
		RequestedAnchor int   `yaml:"requestedAnchor"`
		AnchorTurn      int   `yaml:"anchorTurn"`
		Span            int   `yaml:"span"`
		FirstTurn       int   `yaml:"firstTurn"`
		LastTurn        int   `yaml:"lastTurn"`
		Segments        []int `yaml:"segments"`
	} `yaml:"window"`
}

func loadSparseSession(t *testing.T) sparseSessionFixture {
	t.Helper()
	var fixture sparseSessionFixture
	if err := yaml.Unmarshal(transcriptSparseSessionYAML, &fixture); err != nil {
		t.Fatalf("decode sparse session fixture: %v", err)
	}
	if len(fixture.Turns) == 0 || fixture.Revision == "" {
		t.Fatal("sparse session fixture is incomplete")
	}
	return fixture
}

// windowSessionTurns builds wire turns for a window case: 64 KiB of content per
// turn and two small tool calls per stated call, so the 55-turn session cuts
// into several segments under the contract budget.
func windowSessionTurns(t *testing.T, specs []windowTurnSpec) ([]schema.TurnDetail, []schema.TranscriptWindowTurn) {
	t.Helper()
	turns := make([]schema.TurnDetail, len(specs))
	meta := make([]schema.TranscriptWindowTurn, len(specs))
	for i, spec := range specs {
		role := schema.Role(spec.Role)
		if !slices.Contains(schema.AllRoles, role) {
			t.Fatalf("window fixture turn %d names role %q; use a member of schema.AllRoles", spec.Index, spec.Role)
		}
		turn := schema.TurnDetail{Index: spec.Index, Role: role, EntryType: schema.EntryTypeText, Content: strings.Repeat("s", 64<<10)}
		for call := 0; call < spec.ToolCalls; call++ {
			turn.ToolCalls = append(turn.ToolCalls, schema.ToolCallDetail{Arguments: `{"command":"echo"}`, Result: "ok"})
		}
		turns[i] = turn
		meta[i] = schema.TranscriptWindowTurn{Index: spec.Index, Role: role}
	}
	return turns, meta
}

// TestResolveTranscriptWindow replays the window cases against the directory
// CutTranscriptSegments builds under the contract budget.
func TestResolveTranscriptWindow(t *testing.T) {
	sparse := loadSparseSession(t)
	fixture := decodeStrictYAML[windowFixture](t, transcriptWindowsYAML, "transcript windows fixture")
	names := make([]string, len(fixture.Cases))
	for i, c := range fixture.Cases {
		names[i] = c.Name
	}
	requireManifestNames(t, "transcript windows", transcriptWindowsManifestYAML, names)
	for _, name := range fixture.CrossesSegmentBoundary {
		if !slices.Contains(names, name) {
			t.Fatalf("crossesSegmentBoundary names unknown case %q", name)
		}
	}
	if len(fixture.CrossesSegmentBoundary) == 0 {
		t.Fatal("no window case is required to cross a segment boundary")
	}
	sessionID := schema.SessionID(sparse.SessionID)

	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			specs := sparse.Turns
			if c.Turns != nil {
				specs = c.Turns
			}
			turns, meta := windowSessionTurns(t, specs)
			refs := schema.CutTranscriptSegments(turns, schema.TranscriptSegmentBudgetBytes)

			window, err := schema.ResolveTranscriptWindow(sessionID, sparse.Revision, c.Anchor, c.Span, meta, refs)
			switch c.Expected.Error {
			case "no-user-turn":
				if !errors.Is(err, schema.ErrTranscriptWindowNoUserTurn) {
					t.Fatalf("error = %v, want ErrTranscriptWindowNoUserTurn", err)
				}
				for _, fragment := range []string{string(sessionID), "turn " + strconv.Itoa(c.Anchor)} {
					if !strings.Contains(err.Error(), fragment) {
						t.Fatalf("error %q does not name %q", err, fragment)
					}
				}
				return
			case "span-out-of-range":
				if err == nil || errors.Is(err, schema.ErrTranscriptWindowNoUserTurn) || !strings.Contains(err.Error(), fmt.Sprintf("span is %d", c.Span)) {
					t.Fatalf("error = %v, want a refusal naming span %d", err, c.Span)
				}
				return
			case "":
			default:
				t.Fatalf("window case names unknown error %q", c.Expected.Error)
			}
			if err != nil {
				t.Fatalf("resolve window: %v", err)
			}
			got := windowExpectation{AnchorTurn: window.AnchorTurn, FirstTurn: window.FirstTurn, LastTurn: window.LastTurn, HasOlder: window.HasOlder, HasNewer: window.HasNewer}
			if got != c.Expected {
				t.Fatalf("window = %+v, want %+v", got, c.Expected)
			}
			if window.RequestedAnchor != c.Anchor || window.Span != c.Span || window.SessionID != sessionID || window.Revision != sparse.Revision {
				t.Fatalf("window echoes anchor %d span %d session %q revision %q, want %d, %d, %q, %q", window.RequestedAnchor, window.Span, window.SessionID, window.Revision, c.Anchor, c.Span, sessionID, sparse.Revision)
			}
			if slices.Contains(fixture.CrossesSegmentBoundary, c.Name) && len(window.Segments) < 2 {
				t.Fatalf("the window is covered by %d segment(s); this case must cross a segment boundary", len(window.Segments))
			}
			if err := window.Validate(); err != nil {
				t.Fatalf("window does not validate: %v", err)
			}
			if err := schema.ValidateTranscriptWindowCoverage(window, refs); err != nil {
				t.Fatalf("window coverage: %v", err)
			}
			// The descriptor carries no content.
			encoded, err := json.Marshal(window)
			if err != nil {
				t.Fatal(err)
			}
			for _, forbidden := range []string{`"turns"`, `"content"`, `"toolCalls"`} {
				if bytes.Contains(encoded, []byte(forbidden)) {
					t.Fatalf("window body carries %s: %s", forbidden, encoded)
				}
			}
		})
	}

	// The sparse session also records a hand-declared directory and the window
	// it covers. The resolver names that same segment run.
	t.Run("declared-directory", func(t *testing.T) {
		_, meta := windowSessionTurns(t, sparse.Turns)
		refs := make([]schema.TranscriptSegmentRef, len(sparse.Segments))
		for i, ref := range sparse.Segments {
			refs[i] = schema.TranscriptSegmentRef{Index: ref.Index, FirstTurn: ref.FirstTurn, LastTurn: ref.LastTurn, ByteLength: ref.ByteLength}
		}
		want := sparse.Window
		window, err := schema.ResolveTranscriptWindow(sessionID, sparse.Revision, want.RequestedAnchor, want.Span, meta, refs)
		if err != nil {
			t.Fatal(err)
		}
		if window.AnchorTurn != want.AnchorTurn || window.FirstTurn != want.FirstTurn || window.LastTurn != want.LastTurn || !slices.Equal(window.Segments, want.Segments) {
			t.Fatalf("window = anchor %d [%d, %d] segments %v, want anchor %d [%d, %d] segments %v", window.AnchorTurn, window.FirstTurn, window.LastTurn, window.Segments, want.AnchorTurn, want.FirstTurn, want.LastTurn, want.Segments)
		}
	})

	// A window whose edge lies in no ref is refused, not served short.
	t.Run("uncovered-directory", func(t *testing.T) {
		_, meta := windowSessionTurns(t, sparse.Turns)
		refs := []schema.TranscriptSegmentRef{{Index: 0, FirstTurn: 4, LastTurn: 40}}
		_, err := schema.ResolveTranscriptWindow(sessionID, sparse.Revision, 72, 1, meta, refs)
		if err == nil || !strings.Contains(err.Error(), "is not covered by") {
			t.Fatalf("error = %v, want a refusal naming the uncovered window", err)
		}
	})
}

// TestTranscriptWindowSpanConstantsMatchTheDeclaredRange holds every restated
// span bound equal to the constants: the struct tags on TranscriptWindow.Span
// and the span query parameter of the Local and Village window reads.
func TestTranscriptWindowSpanConstantsMatchTheDeclaredRange(t *testing.T) {
	if schema.TranscriptWindowSpanDefault < schema.TranscriptWindowSpanMinimum || schema.TranscriptWindowSpanDefault > schema.TranscriptWindowSpanMaximum {
		t.Fatalf("the default span %d lies outside %d..%d", schema.TranscriptWindowSpanDefault, schema.TranscriptWindowSpanMinimum, schema.TranscriptWindowSpanMaximum)
	}
	field, ok := reflect.TypeOf(schema.TranscriptWindow{}).FieldByName("Span")
	if !ok {
		t.Fatal("TranscriptWindow has no Span field")
	}
	if field.Tag.Get("minimum") != strconv.Itoa(schema.TranscriptWindowSpanMinimum) || field.Tag.Get("maximum") != strconv.Itoa(schema.TranscriptWindowSpanMaximum) {
		t.Fatalf("TranscriptWindow.Span is tagged %s..%s, want the constants %d..%d", field.Tag.Get("minimum"), field.Tag.Get("maximum"), schema.TranscriptWindowSpanMinimum, schema.TranscriptWindowSpanMaximum)
	}

	local, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	village, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	for _, read := range []struct {
		name string
		spec any
		path string
	}{
		{"local", local, "/api/v1/sessions/{id}/transcript/window"},
		{"village", village, "/api/v1/transcripts/{id}/window"},
	} {
		raw, err := json.Marshal(read.spec)
		if err != nil {
			t.Fatal(err)
		}
		var document struct {
			Paths map[string]map[string]struct {
				Parameters []struct {
					Name   string         `json:"name"`
					Schema map[string]any `json:"schema"`
				} `json:"parameters"`
			} `json:"paths"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, parameter := range document.Paths[read.path]["get"].Parameters {
			if parameter.Name != "span" {
				continue
			}
			found = true
			want := map[string]any{
				"minimum": float64(schema.TranscriptWindowSpanMinimum),
				"maximum": float64(schema.TranscriptWindowSpanMaximum),
				"default": float64(schema.TranscriptWindowSpanDefault),
			}
			for key, value := range want {
				if parameter.Schema[key] != value {
					t.Fatalf("%s window read declares span %s=%v, want the constant %v", read.name, key, parameter.Schema[key], value)
				}
			}
		}
		if !found {
			t.Fatalf("%s spec declares no span parameter on GET %s", read.name, read.path)
		}
	}
}
