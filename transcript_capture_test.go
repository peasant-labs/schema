package schema_test

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/transcript_capture.yaml
var transcriptCaptureYAML []byte

//go:embed testdata/transcript_capture_manifest.yaml
var transcriptCaptureManifestYAML []byte

// captureInput builds one capture document: literal text, a Go-quoted byte
// literal for bytes YAML cannot carry, or a generated nesting depth.
type captureInput struct {
	Detail *string `yaml:"detail"`
	Bytes  *string `yaml:"bytes"`
	Nest   *struct {
		Depth int    `yaml:"depth"`
		Where string `yaml:"where"`
	} `yaml:"nest"`
	Kills      captureMutant `yaml:"kills"`
	Arch64Only bool          `yaml:"arch64Only"`
}

// captureMutant names a capture mutation the fixture's mutant cases must kill.
type captureMutant string

const (
	captureMutantLastDuplicateWins  captureMutant = "last-duplicate-wins"
	captureMutantEscapedQuoteOffset captureMutant = "offset-after-escaped-quote"
	captureMutantMultibyteOffset    captureMutant = "offset-after-multibyte-rune"
	captureMutantStoredLength       captureMutant = "stored-length-measure"
	captureMutantCutOffByOne        captureMutant = "cut-off-by-one"
)

var allCaptureMutants = []captureMutant{
	captureMutantLastDuplicateWins,
	captureMutantEscapedQuoteOffset,
	captureMutantMultibyteOffset,
	captureMutantStoredLength,
	captureMutantCutOffByOne,
}

func (in captureInput) document(t testing.TB) []byte {
	t.Helper()
	switch {
	case in.Detail != nil:
		return []byte(*in.Detail)
	case in.Bytes != nil:
		s, err := strconv.Unquote(*in.Bytes)
		if err != nil {
			t.Fatalf("bytes %q is not a Go double-quoted literal: %v", *in.Bytes, err)
		}
		return []byte(s)
	case in.Nest != nil:
		switch in.Nest.Where {
		case "turn":
			n := in.Nest.Depth - 3
			return []byte(`{"turns":[{"x":` + strings.Repeat("[", n) + strings.Repeat("]", n) + `}]}`)
		case "outside":
			n := in.Nest.Depth - 1
			return []byte(`{"x":` + strings.Repeat("[", n) + strings.Repeat("]", n) + `,"turns":[]}`)
		}
		t.Fatalf("nest where %q is not turn or outside", in.Nest.Where)
	}
	t.Fatal("capture input has no detail, bytes or nest")
	return nil
}

// captureExpected is the observation the baseline build recorded. Its fields
// and YAML form match tools/capture-differential's observation.
type captureExpected struct {
	Refusal schema.TranscriptCaptureRefusal `yaml:"refusal,omitempty"`
	Turn    *int                            `yaml:"turn,omitempty"`
	Turns   []captureTurnExpected           `yaml:"turns,omitempty"`
	Cook    *struct {
		ClaudeCode []string `yaml:"claudeCode,flow"`
		Pi         []string `yaml:"pi,flow"`
	} `yaml:"cook,omitempty"`
	Cuts *struct {
		Budget   []string `yaml:"budget,flow"`
		One      []string `yaml:"one,flow"`
		Fit      []string `yaml:"fit,omitempty,flow"`
		FitBelow []string `yaml:"fitBelow,omitempty,flow"`
	} `yaml:"cuts,omitempty"`
	Kept *struct {
		Budget []string `yaml:"budget,flow"`
		One    []string `yaml:"one,flow"`
	} `yaml:"kept,omitempty"`
}

type captureTurnExpected struct {
	Index         int    `yaml:"index"`
	Raw           string `yaml:"raw"`
	RecordedIndex string `yaml:"recordedIndex"`
	ByteLength    int    `yaml:"byteLength"`
	Detail        string `yaml:"detail"`
	Role          string `yaml:"role"`
	EntryType     string `yaml:"entryType"`
	Depth         int    `yaml:"depth"`
	Usage         bool   `yaml:"usage"`
	Conversation  bool   `yaml:"conversation"`
}

type captureCorpus = testcase.Corpus[captureInput, captureExpected]

type captureFixture struct {
	RecordedIndex captureCorpus `yaml:"recordedIndex"`
	Verdicts      captureCorpus `yaml:"verdicts"`
	Mutants       captureCorpus `yaml:"mutants"`
}

func loadCaptureFixture(t testing.TB) captureFixture {
	t.Helper()
	var f captureFixture
	decoder := yaml.NewDecoder(bytes.NewReader(transcriptCaptureYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		t.Fatalf("decode capture fixture: %v", err)
	}
	var names []string
	for _, arm := range []captureCorpus{f.RecordedIndex, f.Verdicts, f.Mutants} {
		if err := arm.Validate(); err != nil {
			t.Fatal(err)
		}
		for _, c := range arm.Cases {
			names = append(names, c.Name)
		}
	}
	if tt, ok := t.(*testing.T); ok {
		caseassert.RequireManifestNames(tt, "transcript capture", transcriptCaptureManifestYAML, names)
	}
	return f
}

func captureSHA(b []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

// observeCapture records the candidate's public consequences of one capture
// in the fixture's form.
func observeCapture(t *testing.T, doc []byte) captureExpected {
	t.Helper()
	turns, err := schema.NewTranscriptCookTurns(doc)
	if err != nil {
		refusal := requireCaptureError(t, err)
		turn := refusal.Turn
		return captureExpected{Refusal: refusal.Refusal, Turn: &turn}
	}
	var o captureExpected
	details := make([]schema.TurnDetail, len(turns))
	for i, turn := range turns {
		d := turn.Detail()
		details[i] = d
		raw, err := turn.Raw().MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		state, index := turn.RecordedIndex()
		wantState, wantIndex := schema.DecodeTranscriptRecordedIndex(turn.Raw())
		if state != wantState || !reflect.DeepEqual(index, wantIndex) {
			t.Fatalf("turn %d RecordedIndex() = %s %v, DecodeTranscriptRecordedIndex(Raw()) = %s %v", i, state, index, wantState, wantIndex)
		}
		ri := string(state)
		if index != nil {
			ri = fmt.Sprintf("%s:%d", state, *index)
		}
		f := turn.Facts()
		requireFactsMatchDetail(t, i, f, d)
		detail, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		o.Turns = append(o.Turns, captureTurnExpected{
			Index: f.Index, Raw: captureSHA(raw), RecordedIndex: ri, ByteLength: f.ByteLength,
			Detail: captureSHA(detail), Role: string(f.Role), EntryType: string(f.EntryType), Depth: f.Depth,
			Usage: f.UsagePresent, Conversation: f.ConversationTurn,
		})
	}
	o.Cook = &struct {
		ClaudeCode []string `yaml:"claudeCode,flow"`
		Pi         []string `yaml:"pi,flow"`
	}{captureCook(turns, schema.HarnessClaudeCode), captureCook(turns, schema.HarnessPi)}
	o.Cuts = &struct {
		Budget   []string `yaml:"budget,flow"`
		One      []string `yaml:"one,flow"`
		Fit      []string `yaml:"fit,omitempty,flow"`
		FitBelow []string `yaml:"fitBelow,omitempty,flow"`
	}{Budget: captureCuts(details, schema.TranscriptSegmentBudgetBytes), One: captureCuts(details, 1)}
	if len(details) >= 2 {
		fit := schema.TranscriptTurnByteLength(details[0]) + schema.TranscriptTurnByteLength(details[1])
		o.Cuts.Fit, o.Cuts.FitBelow = captureCuts(details, fit), captureCuts(details, fit-1)
	}
	if verdicts, err := schema.CookTranscriptRows(turns, schema.HarnessClaudeCode); err == nil {
		kept := make([]schema.TurnDetail, 0, len(turns))
		for i, v := range verdicts {
			if v.Keep {
				kept = append(kept, details[i])
			}
		}
		o.Kept = &struct {
			Budget []string `yaml:"budget,flow"`
			One    []string `yaml:"one,flow"`
		}{captureCookCut(t, turns, verdicts, kept, schema.TranscriptSegmentBudgetBytes), captureCookCut(t, turns, verdicts, kept, 1)}
	}
	return o
}

// captureCookCut cuts the kept turns and holds the result equal to
// CutTranscriptSegments over their details.
func captureCookCut(t *testing.T, turns []schema.TranscriptCookTurn, verdicts []schema.CookedTurn, kept []schema.TurnDetail, budget int) []string {
	t.Helper()
	refs, err := schema.CutTranscriptCookTurns(turns, verdicts, budget)
	if err != nil {
		t.Fatal(err)
	}
	if want := schema.CutTranscriptSegments(kept, budget); !reflect.DeepEqual(refs, want) {
		t.Fatalf("CutTranscriptCookTurns at budget %d = %+v, CutTranscriptSegments over the kept details = %+v", budget, refs, want)
	}
	return captureRefStrings(refs)
}

// requireFactsMatchDetail holds Facts() equal to the values Detail() gives.
func requireFactsMatchDetail(t testing.TB, i int, f schema.TranscriptCookTurnFacts, d schema.TurnDetail) {
	t.Helper()
	usage := d.Usage != nil
	for _, tool := range d.ToolCalls {
		usage = usage || tool.Usage != nil
	}
	want := schema.TranscriptCookTurnFacts{
		Index: d.Index, Role: d.Role, EntryType: d.EntryType, Depth: d.Depth, UsagePresent: usage,
		ConversationTurn: schema.IsConversationTurn(d.Role, d.EntryType, d.Depth), ByteLength: schema.TranscriptTurnByteLength(d),
	}
	if f != want {
		t.Fatalf("turn %d Facts() = %+v, Detail() gives %+v", i, f, want)
	}
}

func captureCook(turns []schema.TranscriptCookTurn, harness schema.Harness) []string {
	rows, err := schema.CookTranscriptRows(turns, harness)
	if err != nil {
		return []string{"uncookable"}
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		if r.Keep {
			out[i] = fmt.Sprintf("keep %d %s", r.ServedPosition, r.DisplayNumber)
		} else {
			out[i] = fmt.Sprintf("drop %d", r.ServedPosition)
		}
	}
	return out
}

func captureCuts(details []schema.TurnDetail, budget int) []string {
	return captureRefStrings(schema.CutTranscriptSegments(details, budget))
}

func captureRefStrings(refs []schema.TranscriptSegmentRef) []string {
	out := make([]string, len(refs))
	for i, r := range refs {
		span := "none"
		if r.ConversationTurns != nil {
			span = fmt.Sprintf("%d+%d", r.ConversationTurns.Before, r.ConversationTurns.Count)
		}
		out[i] = fmt.Sprintf("%d %d-%d %dB %s", r.Index, r.FirstTurn, r.LastTurn, r.ByteLength, span)
	}
	return out
}

// documentCaptureRefusals are the refusals about the whole document; every
// other class names a refusing turn.
var documentCaptureRefusals = []schema.TranscriptCaptureRefusal{
	schema.TranscriptCaptureDocumentInvalid,
	schema.TranscriptCaptureDocumentNotObject,
	schema.TranscriptCaptureDocumentDuplicateMember,
	schema.TranscriptCaptureDocumentEnvelope,
	schema.TranscriptCaptureTurnsNotArray,
}

func requireCaptureError(t testing.TB, err error) *schema.TranscriptCaptureError {
	t.Helper()
	var refusal *schema.TranscriptCaptureError
	if !errors.As(err, &refusal) || !refusal.Refusal.IsValid() {
		t.Fatalf("capture refusal %v is not a *TranscriptCaptureError with a valid class", err)
	}
	documentLevel := false
	for _, r := range documentCaptureRefusals {
		documentLevel = documentLevel || r == refusal.Refusal
	}
	if documentLevel != (refusal.Turn == -1) || refusal.Turn < -1 {
		t.Fatalf("refusal %s names turn %d", refusal.Refusal, refusal.Turn)
	}
	if errors.Unwrap(err) != refusal.Err || refusal.Error() != refusal.Err.Error() {
		t.Fatalf("refusal %v does not wrap its message error", err)
	}
	return refusal
}

func TestTranscriptCaptureRefusalSet(t *testing.T) {
	want := []schema.TranscriptCaptureRefusal{
		"document-invalid", "document-not-object", "document-duplicate-member", "document-envelope",
		"turns-not-array", "turn-unicode-escape", "turn-duplicate-member", "turn-not-object", "turn-type-mismatch",
	}
	if !reflect.DeepEqual(schema.AllTranscriptCaptureRefusals, want) {
		t.Fatalf("AllTranscriptCaptureRefusals = %v, want %v", schema.AllTranscriptCaptureRefusals, want)
	}
	for _, r := range want {
		if !r.IsValid() {
			t.Fatalf("%s is not valid", r)
		}
	}
	if schema.TranscriptCaptureRefusal("turn-invalid").IsValid() {
		t.Fatal("an unknown refusal is valid")
	}
}

func TestTranscriptCaptureFixture(t *testing.T) {
	f := loadCaptureFixture(t)
	killed := map[captureMutant]bool{}
	for _, arm := range []struct {
		name   string
		corpus captureCorpus
	}{{"recordedIndex", f.RecordedIndex}, {"verdicts", f.Verdicts}, {"mutants", f.Mutants}} {
		for _, c := range arm.corpus.Cases {
			t.Run(arm.name+"/"+c.Name, func(t *testing.T) {
				if c.Input.Arch64Only && strconv.IntSize == 32 {
					t.Skip("the recorded expectation holds a value only a 64-bit int can decode; internal/recordedindex covers 32-bit")
				}
				if (c.Classification == testcase.MustFail) != (c.Expected.Refusal != "") {
					t.Fatalf("classification %s disagrees with the recorded refusal %q", c.Classification, c.Expected.Refusal)
				}
				switch arm.name {
				case "recordedIndex":
					if c.Classification != testcase.MustPass {
						t.Fatal("a recorded-index case must be accepted")
					}
				case "mutants":
					if !inCaptureMutants(c.Input.Kills) {
						t.Fatalf("mutant case kills %q, outside the closed mutant list", c.Input.Kills)
					}
					killed[c.Input.Kills] = true
				default:
					if c.Input.Kills != "" {
						t.Fatal("only mutant cases name a mutant")
					}
				}
				got := observeCapture(t, c.Input.document(t))
				gotYAML, err := yaml.Marshal(got)
				if err != nil {
					t.Fatal(err)
				}
				wantYAML, err := yaml.Marshal(c.Expected)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(gotYAML, wantYAML) {
					t.Fatalf("capture observation differs from the baseline recording\n--- got\n%s--- want\n%s", gotYAML, wantYAML)
				}
			})
		}
	}
	for _, m := range allCaptureMutants {
		if !killed[m] {
			t.Errorf("no mutant case kills %s", m)
		}
	}
}

func inCaptureMutants(m captureMutant) bool {
	for _, known := range allCaptureMutants {
		if m == known {
			return true
		}
	}
	return false
}

// TestTranscriptCaptureCopiesItsInput holds the private-copy promise: the
// returned turns do not alias the caller's buffer.
func TestTranscriptCaptureCopiesItsInput(t *testing.T) {
	doc := []byte(`{"turns":[{"index":1,"role":"user","content":"abc"}]}`)
	turns, err := schema.NewTranscriptCookTurns(doc)
	if err != nil {
		t.Fatal(err)
	}
	before := turns[0].Raw().Raw()
	for i := range doc {
		doc[i] = ' '
	}
	if after := turns[0].Raw().Raw(); !bytes.Equal(before, after) {
		t.Fatalf("captured turn changed with the caller's buffer: %q", after)
	}
	if state, index := turns[0].RecordedIndex(); state != schema.TranscriptRecordedIndexPresent || *index != 1 {
		t.Fatalf("RecordedIndex() = %s %v after the caller's buffer changed", state, index)
	}
}

func TestTranscriptCookTurnZeroValueHasNoRecordedIndex(t *testing.T) {
	var turn schema.TranscriptCookTurn
	state, index := turn.RecordedIndex()
	wantState, wantIndex := schema.DecodeTranscriptRecordedIndex(turn.Raw())
	if state != wantState || index != nil || wantIndex != nil {
		t.Fatalf("zero turn RecordedIndex() = %s %v, DecodeTranscriptRecordedIndex = %s %v", state, index, wantState, wantIndex)
	}
}

// FuzzTranscriptCapture checks properties that need no baseline: no panic,
// classified refusals, document validity matching encoding/json, recorded
// indexes matching DecodeTranscriptRecordedIndex, and stable re-capture.
// Parity with the released capture is the differential's job.
func FuzzTranscriptCapture(f *testing.F) {
	fixture := loadCaptureFixture(f)
	for _, arm := range []captureCorpus{fixture.RecordedIndex, fixture.Verdicts, fixture.Mutants} {
		for _, c := range arm.Cases {
			f.Add(c.Input.document(f))
		}
	}
	f.Fuzz(func(t *testing.T, doc []byte) {
		turns, err := schema.NewTranscriptCookTurns(doc)
		valid := utf8.Valid(doc) && json.Valid(doc)
		if err != nil {
			refusal := requireCaptureError(t, err)
			if (refusal.Refusal == schema.TranscriptCaptureDocumentInvalid) == valid {
				t.Fatalf("refusal %s for a document encoding/json calls valid=%t", refusal.Refusal, valid)
			}
			return
		}
		if !valid {
			t.Fatal("capture accepted a document encoding/json calls invalid")
		}
		var again bytes.Buffer
		again.WriteString(`{"turns":[`)
		for i, turn := range turns {
			state, index := turn.RecordedIndex()
			wantState, wantIndex := schema.DecodeTranscriptRecordedIndex(turn.Raw())
			if state != wantState || !reflect.DeepEqual(index, wantIndex) {
				t.Fatalf("turn %d RecordedIndex() = %s %v, DecodeTranscriptRecordedIndex(Raw()) = %s %v", i, state, index, wantState, wantIndex)
			}
			requireFactsMatchDetail(t, i, turn.Facts(), turn.Detail())
			var decoded schema.TurnDetail
			if err := json.Unmarshal(turn.Raw().Raw(), &decoded); err != nil || !reflect.DeepEqual(decoded, turn.Detail()) {
				t.Fatalf("turn %d Detail() = %+v, json.Unmarshal(Raw()) = %+v, %v", i, turn.Detail(), decoded, err)
			}
			if i > 0 {
				again.WriteByte(',')
			}
			again.Write(turn.Raw().Raw())
		}
		again.WriteString(`]}`)
		recaptured, err := schema.NewTranscriptCookTurns(again.Bytes())
		if err != nil {
			t.Fatalf("re-capturing the captured turns failed: %v", err)
		}
		if len(recaptured) != len(turns) {
			t.Fatalf("re-capture gave %d turns, want %d", len(recaptured), len(turns))
		}
		for i := range turns {
			if !bytes.Equal(recaptured[i].Raw().Raw(), turns[i].Raw().Raw()) || !reflect.DeepEqual(recaptured[i].Detail(), turns[i].Detail()) {
				t.Fatalf("re-captured turn %d differs", i)
			}
		}
	})
}

func TestCutTranscriptCookTurnsRefusesMismatchedVerdicts(t *testing.T) {
	turns, err := schema.NewTranscriptCookTurns([]byte(`{"turns":[{"index":0,"role":"user","content":"a"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, verdicts := range [][]schema.CookedTurn{nil, {{Position: 1, Keep: true}}} {
		if _, err := schema.CutTranscriptCookTurns(turns, verdicts, 1); err == nil {
			t.Fatalf("verdicts %+v were accepted", verdicts)
		}
	}
	if _, err := schema.CutTranscriptCookTurns(make([]schema.TranscriptCookTurn, 1), []schema.CookedTurn{{Keep: true}}, 1); err == nil {
		t.Fatal("a zero-value turn was cut")
	}
}

// FuzzTranscriptJSONStringByteLength holds the counted measure equal to the
// bytes encoding/json writes.
func FuzzTranscriptJSONStringByteLength(f *testing.F) {
	for _, s := range []string{"", "a", "<>&", "\u2028\u2029", "\"\\/", "\b\f\n\r\t\x00\x1f\x7f", "\xff\xc3(", "é中😀", "\ufffd"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		encoded, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		if got := schema.TranscriptJSONStringByteLength(s); got != len(encoded) {
			t.Fatalf("TranscriptJSONStringByteLength(%q) = %d, encoder writes %d bytes", s, got, len(encoded))
		}
	})
}
