package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// TranscriptCookTurn pairs one served turn's original bytes with its own
// decoded value. The cooking rule uses the exact-name member view, not detail.
// Construct turns with NewTranscriptCookTurns; the zero value is not a turn.
type TranscriptCookTurn struct {
	raw     TranscriptServedTurn
	members []transcriptMember
	detail  TurnDetail
}

// Raw returns the preserved turn object for the segment writer.
func (t TranscriptCookTurn) Raw() TranscriptServedTurn { return t.raw }

// Detail returns the value decoded from the same turn bytes for a host index.
func (t TranscriptCookTurn) Detail() TurnDetail { return t.detail }

// RecordedIndex reads the turn's recorded index from the member view the
// capture already holds. It applies the same rule as
// DecodeTranscriptRecordedIndex on the turn's bytes and gives the same result.
func (t TranscriptCookTurn) RecordedIndex() (TranscriptRecordedIndexState, *int) {
	if !t.constructed() {
		return TranscriptRecordedIndexInvalid, nil
	}
	raw, present := t.memberView().member(transcriptWireIndex)
	return transcriptRecordedIndex(raw, present)
}

// constructed reports whether the turn came from NewTranscriptCookTurns: a
// captured turn always holds at least the two bytes of an object.
func (t TranscriptCookTurn) constructed() bool { return len(t.raw.raw) != 0 }

func (t TranscriptCookTurn) memberView() transcriptTurnMembers {
	return transcriptTurnMembers{raw: t.raw.raw, table: t.members}
}

// transcriptRawMembers reads decoded object keys with exact case. The decoder
// owns JSON string and escape handling; duplicate decoded names are refused.
func transcriptRawMembers(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if token != json.Delim('{') {
		return nil, fmt.Errorf("a JSON object is required")
	}
	members := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("an object member name is required")
		}
		if _, exists := members[name]; exists {
			return nil, fmt.Errorf("duplicate object key %q", name)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		members[name] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("an object must be the only JSON value")
	}
	return members, nil
}
