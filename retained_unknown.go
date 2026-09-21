package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// RetainedUnknownRecord preserves an uninterpreted source record or block after
// ordinary producer redaction. Payload is complete JSON TEXT, not a parsed JSON
// value: this preserves numeric literals outside JavaScript's number domain.
// SourceRef identifies a source stream, never a filesystem path. RecordIndex
// counts every source record, including known records, starting at zero; it is
// assigned by traversal and does not depend on a native ordinal. Pointer is an
// RFC 6901 pointer within that record (empty for the whole record). Position is
// the zero-based traversal position of the record/block within the source stream.
// Namespace disambiguates native discriminators such as record and message.block.
// Neither Namespace nor Kind is a closed vocabulary.
type RetainedUnknownRecord struct {
	SourceRef   string `json:"sourceRef" minLength:"1"`
	RecordIndex int64  `json:"recordIndex" minimum:"0" maximum:"9007199254740991"`
	Position    int64  `json:"position" minimum:"0" maximum:"9007199254740991"`
	Pointer     string `json:"pointer" pattern:"^(?:/(?:[^~]|~[01])*)*$"`
	Namespace   string `json:"namespace" minLength:"1"`
	Kind        string `json:"kind" minLength:"1"`
	Payload     string `json:"payload" minLength:"1"`
}

// InterpretationDiagnostics is durable interpretation state, independent of
// conversational outcome. Partial means retained evidence was not interpreted;
// it does not authorize loss of that evidence or imply damaged input is valid.
type InterpretationDiagnostics struct {
	Partial bool `json:"partial"`
}

// ValidateRetainedUnknown checks complete retained evidence and its source order.
// The normal outer transport limits apply, never native-metadata-only budgets.
func ValidateRetainedUnknown(value SessionDetailPayload) error {
	fail := func(reason string) error {
		return fmt.Errorf("retained unknown validation failed at schema.ValidateRetainedUnknown during detail validation: %s; source evidence cannot be reconstructed safely; retain complete redacted JSON with unique ordered source positions and diagnostics.partial=true", reason)
	}
	if len(value.RetainedUnknown) > 0 && (value.Diagnostics == nil || !value.Diagnostics.Partial) {
		return fail("retainedUnknown requires diagnostics.partial=true")
	}
	type cursor struct {
		position, record int64
		pointers         []string
	}
	sources := map[string]cursor{}
	for i, record := range value.RetainedUnknown {
		if record.SourceRef == "" || record.Namespace == "" || record.Kind == "" || !utf8.ValidString(record.SourceRef) || !utf8.ValidString(record.Namespace) || !utf8.ValidString(record.Kind) || !utf8.ValidString(record.Pointer) || !utf8.ValidString(record.Payload) {
			return fail(fmt.Sprintf("retainedUnknown[%d] has empty identity or invalid Unicode", i))
		}
		if record.RecordIndex < 0 || record.RecordIndex > 9007199254740991 || record.Position < 0 || record.Position > 9007199254740991 || record.Position < record.RecordIndex {
			return fail(fmt.Sprintf("retainedUnknown[%d] has invalid recordIndex or position", i))
		}
		if !validUnknownPointer(record.Pointer) {
			return fail(fmt.Sprintf("retainedUnknown[%d] has an invalid RFC 6901 pointer", i))
		}
		previous, exists := sources[record.SourceRef]
		if exists && (record.Position <= previous.position || record.RecordIndex < previous.record) {
			return fail(fmt.Sprintf("retainedUnknown[%d] repeats or reverses a source position", i))
		}
		if !exists || record.RecordIndex != previous.record {
			previous.pointers = nil
		}
		for _, pointer := range previous.pointers {
			if pointer == record.Pointer || strings.HasPrefix(pointer, record.Pointer+"/") || strings.HasPrefix(record.Pointer, pointer+"/") {
				return fail(fmt.Sprintf("retainedUnknown[%d] duplicates or overlaps a source pointer", i))
			}
		}
		if err := ScanRawJSONDocument([]byte(record.Payload), RawJSONPathPolicy{MaxDocumentBytes: 8 << 20, MaxDocumentDepth: 64}); err != nil {
			return fail(fmt.Sprintf("retainedUnknown[%d].payload is not complete valid JSON: %v", i, err))
		}
		sources[record.SourceRef] = cursor{record.Position, record.RecordIndex, append(previous.pointers, record.Pointer)}
	}
	if len(value.RetainedUnknown) > 0 {
		wire, err := json.Marshal(value)
		if err != nil {
			return fail("detail cannot be serialized: " + err.Error())
		}
		if len(wire) > 8<<20 {
			return fail("detail exceeds the existing 8 MiB outer transport limit")
		}
	}
	return nil
}

func validUnknownPointer(pointer string) bool {
	if pointer == "" {
		return true
	}
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	for i := 0; i < len(pointer); i++ {
		if pointer[i] == '~' {
			i++
			if i == len(pointer) || (pointer[i] != '0' && pointer[i] != '1') {
				return false
			}
		}
	}
	return true
}
