package main

import (
	"encoding/json"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// versionFieldValue reads the protected top-level field from JSON bytes. It
// returns corrupt=true only when the bytes do not parse: callers fail closed
// on corrupt input. A missing file (nil data), a missing field, or a
// non-string value yields ("", false) so unchanged files compare equal and an
// added or removed value compares as changed. The field rule itself lives in
// the contract package; this only interprets its failure for the caller's
// fail-closed branch.
func versionFieldValue(data []byte, field string) (value string, corrupt bool) {
	if len(data) == 0 {
		return "", false
	}
	if !json.Valid(data) {
		return "", true
	}
	v, err := intent.ReadVersionField(data, field)
	if err != nil {
		return "", false
	}
	return v, false
}
