package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// TranscriptEntryReferenceKind identifies the member holding a located entry.
// It describes the finding, not the boundary kind carried by a link.
type TranscriptEntryReferenceKind string

const (
	TranscriptEntryReferenceSource     TranscriptEntryReferenceKind = "source"
	TranscriptEntryReferenceToolCall   TranscriptEntryReferenceKind = "tool_call"
	TranscriptEntryReferenceToolResult TranscriptEntryReferenceKind = "tool_result"
)

var AllTranscriptEntryReferenceKinds = []TranscriptEntryReferenceKind{
	TranscriptEntryReferenceSource,
	TranscriptEntryReferenceToolCall,
	TranscriptEntryReferenceToolResult,
}

func (v TranscriptEntryReferenceKind) IsValid() bool {
	return inSet(v, AllTranscriptEntryReferenceKinds)
}

func (TranscriptEntryReferenceKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Entry Reference Kind", AllTranscriptEntryReferenceKinds)
}

// TranscriptEntryLocateRequest identifies a reference and the expected read
// epoch. It contains neither a location kind nor a partition selector.
type TranscriptEntryLocateRequest struct {
	Reference      string `json:"reference" minLength:"1"`
	Revision       string `json:"revision"`
	Representation string `json:"representation"`
	ContentHash    string `json:"contentHash"`
}

// TranscriptEntryMatch reports a location without carrying the reference.
// RecordedIndex is present only when the matched row has a recorded index.
type TranscriptEntryMatch struct {
	Context        TranscriptReadContext        `json:"context"`
	Kind           TranscriptEntryReferenceKind `json:"kind"`
	Position       int                          `json:"position"`
	SegmentOrdinal int                          `json:"segmentOrdinal"`
	RecordedIndex  *int                         `json:"recordedIndex,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
}

// TranscriptEntryScanRow is one served main-partition row and the recorded
// index pair computed by the same pass that captured its turn.
type TranscriptEntryScanRow struct {
	Turn     TranscriptCookTurn
	State    TranscriptRecordedIndexState
	Recorded *int // non-nil exactly when State is TranscriptRecordedIndexPresent
}

// TranscriptEntryHit is a scan finding. Position is its index in the served
// rows and Kind identifies the exact member that held the reference.
type TranscriptEntryHit struct {
	Position      int
	Kind          TranscriptEntryReferenceKind
	RecordedIndex *int
}

// LocateTranscriptEntry searches served main-partition rows in order, reading
// each row's original bytes by decoded exact member name. Within a row it
// checks sourceEntryRef, then each tool call's callEntryRef and resultEntryRef.
// A missing match is not an error. The reference must be non-empty.
func LocateTranscriptEntry(rows []TranscriptEntryScanRow, reference string) (TranscriptEntryHit, bool, error) {
	const where = "transcript entry scan failed at schema.LocateTranscriptEntry: "
	if reference == "" {
		return TranscriptEntryHit{}, false, fmt.Errorf("%sthe reference is empty; no entry can be located; pass the non-empty reference from the link", where)
	}
	for position, row := range rows {
		members, err := transcriptRawMembers(row.Turn.Raw().Raw())
		if err != nil {
			return TranscriptEntryHit{}, false, fmt.Errorf("%sserved row %d is not a valid turn object: %w; capture the served rows again before scanning", where, position, err)
		}
		hit := TranscriptEntryHit{Position: position}
		if row.State == TranscriptRecordedIndexPresent {
			hit.RecordedIndex = row.Recorded
		}
		if transcriptEntryStringEquals(members[transcriptWireSourceEntryRef], reference) {
			hit.Kind = TranscriptEntryReferenceSource
			return hit, true, nil
		}
		rawCalls := bytes.TrimSpace(members[transcriptWireToolCalls])
		if len(rawCalls) == 0 || rawCalls[0] != '[' {
			continue
		}
		var calls []json.RawMessage
		if err := json.Unmarshal(rawCalls, &calls); err != nil {
			return TranscriptEntryHit{}, false, fmt.Errorf("%sserved row %d has malformed toolCalls: %w; capture the served rows again before scanning", where, position, err)
		}
		for i, rawCall := range calls {
			call, err := transcriptRawMembers(rawCall)
			if err != nil {
				return TranscriptEntryHit{}, false, fmt.Errorf("%sserved row %d tool call %d is not a valid object: %w; capture the served rows again before scanning", where, position, i, err)
			}
			if transcriptEntryStringEquals(call[transcriptWireCallEntryRef], reference) {
				hit.Kind = TranscriptEntryReferenceToolCall
				return hit, true, nil
			}
			if transcriptEntryStringEquals(call[transcriptWireResultEntryRef], reference) {
				hit.Kind = TranscriptEntryReferenceToolResult
				return hit, true, nil
			}
		}
	}
	return TranscriptEntryHit{}, false, nil
}

func transcriptEntryStringEquals(raw json.RawMessage, reference string) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '"' {
		return false
	}
	var value string
	return json.Unmarshal(raw, &value) == nil && value == reference
}
