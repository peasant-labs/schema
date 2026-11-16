package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// TranscriptCookTurn pairs one served turn's original bytes with its own
// decoded value. The cooking rule uses the exact-name member view, not detail.
// Construct turns with NewTranscriptCookTurns; the zero value is not a turn.
type TranscriptCookTurn struct {
	raw     TranscriptServedTurn
	members map[string]json.RawMessage
	detail  TurnDetail
}

// Raw returns the preserved turn object for the segment writer.
func (t TranscriptCookTurn) Raw() TranscriptServedTurn { return t.raw }

// Detail returns the value decoded from the same turn bytes for a host index.
func (t TranscriptCookTurn) Detail() TurnDetail { return t.detail }

// NewTranscriptCookTurns captures one served session detail's turns in stored
// order. It never unwraps an envelope and never pairs two separate decodes.
func NewTranscriptCookTurns(sessionDetail json.RawMessage) ([]TranscriptCookTurn, error) {
	const where = "transcript turn capture failed at schema.NewTranscriptCookTurns: "
	if !utf8.Valid(sessionDetail) || !json.Valid(sessionDetail) {
		return nil, fmt.Errorf("%ssession detail is not one valid UTF-8 JSON object; no turns can be paired; pass the served session detail object", where)
	}
	detail, err := transcriptRawMembers(sessionDetail)
	if err != nil {
		return nil, fmt.Errorf("%ssession detail is not one valid JSON object: %w; pass the served session detail object", where, err)
	}
	if _, ok := detail[transcriptWireSessionDetail]; ok {
		return nil, fmt.Errorf("%san envelope was supplied instead of its session detail; no turns can be paired; select sessionDetail before calling", where)
	}
	raw, ok := detail[transcriptWireTurns]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return []TranscriptCookTurn{}, nil
	}
	var source []json.RawMessage
	if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '[' {
		return nil, fmt.Errorf("%sturns is not an array; no turns can be paired; pass the served turn array", where)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		return nil, fmt.Errorf("%sthe turns array cannot be decoded: %w; pass a valid served turn array", where, err)
	}
	turns := make([]TranscriptCookTurn, len(source))
	for i, rawTurn := range source {
		if err := ScanRawJSONDocument(rawTurn, RawJSONPathPolicy{}); err != nil {
			return nil, fmt.Errorf("%sturn %d has invalid or repeated members: %w; repair the stored turn before retrying", where, i, err)
		}
		members, err := transcriptRawMembers(rawTurn)
		if err != nil {
			return nil, fmt.Errorf("%sturn %d is not one JSON object: %w; repair the stored turn before retrying", where, i, err)
		}
		turn, err := NewTranscriptServedTurn(rawTurn)
		if err != nil {
			return nil, fmt.Errorf("%sturn %d cannot be captured: %w", where, i, err)
		}
		var decoded TurnDetail
		if err := json.Unmarshal(rawTurn, &decoded); err != nil {
			return nil, fmt.Errorf("%sturn %d cannot be decoded as a turn: %w; repair the stored member value before retrying", where, i, err)
		}
		turns[i] = TranscriptCookTurn{raw: turn, members: members, detail: decoded}
	}
	return turns, nil
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
