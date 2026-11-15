package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// TranscriptRecordedIndexState describes whether a served turn carries a
// usable recorded index. An index is not a position in the served partition.
type TranscriptRecordedIndexState string

const (
	TranscriptRecordedIndexPresent TranscriptRecordedIndexState = "present"
	TranscriptRecordedIndexAbsent  TranscriptRecordedIndexState = "absent"
	TranscriptRecordedIndexInvalid TranscriptRecordedIndexState = "invalid"
)

// AllTranscriptRecordedIndexStates is the complete recorded-index state set.
var AllTranscriptRecordedIndexStates = []TranscriptRecordedIndexState{
	TranscriptRecordedIndexPresent,
	TranscriptRecordedIndexAbsent,
	TranscriptRecordedIndexInvalid,
}

// IsValid reports whether v is a recorded-index state.
func (v TranscriptRecordedIndexState) IsValid() bool {
	return inSet(v, AllTranscriptRecordedIndexStates)
}

// JSONSchema publishes the recorded-index state set.
func (TranscriptRecordedIndexState) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Recorded Index State", AllTranscriptRecordedIndexStates)
}

// TranscriptV2Turn holds one served turn's JSON object without interpreting or
// rewriting it. The row's recorded-index state and value are the only v2
// statement of its recorded index.
type TranscriptV2Turn struct {
	raw []byte
}

// NewTranscriptV2Turn copies one complete, valid UTF-8 JSON object. Other JSON
// values are refused because a segment row must carry a turn object.
func NewTranscriptV2Turn(raw []byte) (TranscriptV2Turn, error) {
	const where = "v2 turn validation failed at schema.NewTranscriptV2Turn: "
	trimmed := bytes.TrimSpace(raw)
	if !utf8.Valid(raw) {
		return TranscriptV2Turn{}, fmt.Errorf("%sthe turn contains invalid UTF-8; its bytes cannot be served as JSON; encode the turn as valid UTF-8", where)
	}
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return TranscriptV2Turn{}, fmt.Errorf("%sthe turn is not one complete JSON object; a row cannot carry it; provide a valid turn object", where)
	}
	return TranscriptV2Turn{raw: bytes.Clone(raw)}, nil
}

// MarshalJSON returns a copy of the stored bytes when called directly. A v2
// segment must use MarshalTranscriptSegmentV2: encoding/json compacts nested
// marshaler output and does not preserve those bytes.
func (t TranscriptV2Turn) MarshalJSON() ([]byte, error) {
	if len(t.raw) == 0 {
		return nil, fmt.Errorf("v2 turn encoding failed at schema.TranscriptV2Turn.MarshalJSON: the turn is empty; no JSON object can be emitted; construct it with schema.NewTranscriptV2Turn")
	}
	return bytes.Clone(t.raw), nil
}

// UnmarshalJSON copies the bytes of one turn object.
func (t *TranscriptV2Turn) UnmarshalJSON(raw []byte) error {
	turn, err := NewTranscriptV2Turn(raw)
	if err != nil {
		return err
	}
	*t = turn
	return nil
}

// Raw returns a copy of the retained turn bytes.
func (t TranscriptV2Turn) Raw() []byte {
	return bytes.Clone(t.raw)
}

// JSONSchema publishes an open object. Its index can be absent or invalid on a
// legacy turn, so it cannot use TurnDetail's required integer schema.
func (TranscriptV2Turn) JSONSchema() (jsonschema.Schema, error) {
	var s jsonschema.Schema
	s.AddType(jsonschema.Object)
	return s, nil
}

// DecodeTranscriptRecordedIndex reads the recorded index without fabricating a
// value for an absent or invalid member. The accepted range is the nonnegative
// JavaScript-safe integer range.
func DecodeTranscriptRecordedIndex(turn TranscriptV2Turn) (TranscriptRecordedIndexState, *int) {
	if err := ScanRawJSONDocument(turn.raw, RawJSONPathPolicy{}); err != nil {
		return TranscriptRecordedIndexInvalid, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(turn.raw, &fields); err != nil {
		return TranscriptRecordedIndexInvalid, nil
	}
	raw, ok := fields["index"]
	if !ok {
		return TranscriptRecordedIndexAbsent, nil
	}
	n, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil || n < 0 || n > maxSafeJSONInteger || (strconv.IntSize == 32 && n > 1<<31-1) {
		return TranscriptRecordedIndexInvalid, nil
	}
	index := int(n)
	return TranscriptRecordedIndexPresent, &index
}

// TranscriptV2TurnBytes captures the turns from one served session detail in
// stored order. want is the number of turns read from the same bytes by the
// caller; a different count cannot be paired safely.
func TranscriptV2TurnBytes(sessionDetail json.RawMessage, want int) ([]TranscriptV2Turn, error) {
	const where = "v2 turn capture failed at schema.TranscriptV2TurnBytes: "
	trimmed := bytes.TrimSpace(sessionDetail)
	if !utf8.Valid(sessionDetail) || len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return nil, fmt.Errorf("%ssession detail is not one valid UTF-8 JSON object; turn bytes cannot be paired; pass the served session detail object", where)
	}
	var detail map[string]json.RawMessage
	if err := json.Unmarshal(sessionDetail, &detail); err != nil {
		return nil, fmt.Errorf("%sdecode session detail: %w", where, err)
	}
	var turns []TranscriptV2Turn
	if raw, ok := detail["turns"]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if err := json.Unmarshal(raw, &turns); err != nil {
			return nil, fmt.Errorf("%sdecode turns: %w; pass the served turn array", where, err)
		}
	}
	if len(turns) != want {
		return nil, fmt.Errorf("%sthe served detail contains %d turns but the typed read contains %d; bytes cannot be paired safely; decode both from the same detail", where, len(turns), want)
	}
	return turns, nil
}
