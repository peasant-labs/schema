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

// TranscriptServedTurn holds one served turn's JSON object without interpreting or
// rewriting it. The row's recorded-index state and value are the only
// statement of its recorded index.
type TranscriptServedTurn struct {
	raw []byte
}

// NewTranscriptServedTurn copies one complete, valid UTF-8 JSON object. Other JSON
// values are refused because a segment row must carry a turn object.
func NewTranscriptServedTurn(raw []byte) (TranscriptServedTurn, error) {
	const where = "served turn validation failed at schema.NewTranscriptServedTurn: "
	trimmed := bytes.TrimSpace(raw)
	if !utf8.Valid(raw) {
		return TranscriptServedTurn{}, fmt.Errorf("%sthe turn contains invalid UTF-8; its bytes cannot be served as JSON; encode the turn as valid UTF-8", where)
	}
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return TranscriptServedTurn{}, fmt.Errorf("%sthe turn is not one complete JSON object; a row cannot carry it; provide a valid turn object", where)
	}
	return TranscriptServedTurn{raw: bytes.Clone(raw)}, nil
}

// MarshalJSON returns a copy of the stored bytes when called directly. A positioned
// segment must use MarshalPositionedSegment: encoding/json compacts nested
// marshaler output and does not preserve those bytes.
func (t TranscriptServedTurn) MarshalJSON() ([]byte, error) {
	if len(t.raw) == 0 {
		return nil, fmt.Errorf("served turn encoding failed at schema.TranscriptServedTurn.MarshalJSON: the turn is empty; no JSON object can be emitted; construct it with schema.NewTranscriptServedTurn")
	}
	return bytes.Clone(t.raw), nil
}

// UnmarshalJSON copies the bytes of one turn object.
func (t *TranscriptServedTurn) UnmarshalJSON(raw []byte) error {
	turn, err := NewTranscriptServedTurn(raw)
	if err != nil {
		return err
	}
	*t = turn
	return nil
}

// Raw returns a copy of the retained turn bytes.
func (t TranscriptServedTurn) Raw() []byte {
	return bytes.Clone(t.raw)
}

// JSONSchema publishes an open object. Its index can be absent or invalid on a
// legacy turn, so it cannot use TurnDetail's required integer schema.
func (TranscriptServedTurn) JSONSchema() (jsonschema.Schema, error) {
	var s jsonschema.Schema
	s.AddType(jsonschema.Object)
	return s, nil
}

// DecodeTranscriptRecordedIndex reads the recorded index without fabricating a
// value for an absent or invalid member. The accepted range is the nonnegative
// JavaScript-safe integer range that the platform's int can also represent.
func DecodeTranscriptRecordedIndex(turn TranscriptServedTurn) (TranscriptRecordedIndexState, *int) {
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
