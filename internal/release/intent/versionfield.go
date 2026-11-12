package intent

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

// This file holds the byte-exact version-field reader and writer for JSON
// version files. Only the release PR writes the protected field, and it must
// write nothing else: decoding into a map reorders keys and re-escapes text,
// so the writer replaces only the value bytes and keeps every other byte.

// versionFieldFailure builds a fail-closed version-field error.
func versionFieldFailure(what, why, fix string) error {
	return inputError(
		what, why,
		"a JSON version file, read by intent.ReadVersionField or intent.WriteVersionField",
		"reading or writing the protected version field",
		"the version field cannot be trusted from this read",
		fix,
	)
}

// topLevelStringFields walks the top-level object of data, calling found for
// each key with the raw bytes of its value, whether the value is a string,
// and the value's byte offsets. Nested objects never match: only depth-1 keys
// are visited. A non-object document, a duplicate top-level key, or trailing
// garbage fails closed.
func topLevelStringFields(data []byte, found func(key string, value []byte, isString bool, start, end int) error) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	tok, err := decoder.Token()
	if err != nil {
		return versionFieldFailure(
			"the version file is not valid JSON: "+err.Error(),
			"the version field is read from JSON text",
			"leave the file valid JSON with one top-level object",
		)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return versionFieldFailure(
			"the version file holds no top-level object",
			"the version field is a top-level key of one JSON object",
			"keep the file a single JSON object with the version field at its top level",
		)
	}
	seen := map[string]int{}
	for decoder.More() {
		keyTok, err := decoder.Token()
		if err != nil {
			return versionFieldFailure(
				"the version file breaks off inside its top-level object: "+err.Error(),
				"every top-level key needs a value",
				"leave the file valid JSON with one top-level object",
			)
		}
		key, ok := keyTok.(string)
		if !ok {
			return versionFieldFailure(
				"the version file has a non-string top-level key",
				"JSON object keys are strings",
				"leave the file valid JSON with one top-level object",
			)
		}
		seen[key]++
		start := valueStart(data, int(decoder.InputOffset()))
		if start < 0 {
			return versionFieldFailure(
				"the version file breaks off after key "+strconv.Quote(key),
				"every top-level key needs a value",
				"leave the file valid JSON with one top-level object",
			)
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return versionFieldFailure(
				"the version file breaks off inside the value of key "+strconv.Quote(key)+": "+err.Error(),
				"every top-level value must parse",
				"leave the file valid JSON with one top-level object",
			)
		}
		end := int(decoder.InputOffset())
		var asString string
		isString := json.Unmarshal(raw, &asString) == nil
		if err := found(key, raw, isString, start, end); err != nil {
			return err
		}
	}
	tok, err = decoder.Token()
	if err != nil {
		return versionFieldFailure(
			"the version file breaks off before its closing brace: "+err.Error(),
			"the top-level object must close",
			"leave the file valid JSON with one top-level object",
		)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return versionFieldFailure(
			"the version file has no closing brace where one belongs",
			"the top-level object must close exactly once",
			"leave the file valid JSON with one top-level object",
		)
	}
	if decoder.More() {
		return versionFieldFailure(
			"the version file holds a second top-level value",
			"the file is one JSON object",
			"keep a single object in the file",
		)
	}
	for key, count := range seen {
		if count > 1 {
			return versionFieldFailure(
				"the version file repeats top-level key "+strconv.Quote(key),
				"a repeated key has no one value to read or write",
				"keep one entry per top-level key",
			)
		}
	}
	return nil
}

// valueStart finds the first byte of the value that follows a key ending at
// offset off: optional whitespace, the colon, optional whitespace. It returns
// -1 when the shape is not there.
func valueStart(data []byte, off int) int {
	i := off
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	if i >= len(data) || data[i] != ':' {
		return -1
	}
	i++
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	if i >= len(data) {
		return -1
	}
	return i
}

// ReadVersionField returns the string value of the top-level field in a JSON
// document. A missing field, a non-string value, or unparseable input is an
// error, never a guess.
func ReadVersionField(data []byte, field string) (string, error) {
	value := ""
	found := false
	err := topLevelStringFields(data, func(key string, raw []byte, isString bool, _, _ int) error {
		if key != field {
			return nil
		}
		if !isString {
			return versionFieldFailure(
				"the version file holds a non-string value for field "+strconv.Quote(field),
				"the protected field carries the version as a JSON string",
				"keep the field a quoted string, for example \"version\": \"0.1.0\"",
			)
		}
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return versionFieldFailure(
				"the version file holds an unreadable string for field "+strconv.Quote(field)+": "+err.Error(),
				"the protected field carries the version as a JSON string",
				"keep the field a quoted string, for example \"version\": \"0.1.0\"",
			)
		}
		value, found = s, true
		return nil
	})
	if err != nil {
		return "", err
	}
	if !found {
		return "", versionFieldFailure(
			"the version file has no top-level field "+strconv.Quote(field),
			"only the named field is protected and written",
			"add the field at the file's top level, or fix the field name in the release config",
		)
	}
	return value, nil
}

// WriteVersionField replaces the string value of the top-level field with the
// new value and returns the rewritten document. Only the value bytes change:
// key order, whitespace, and escaping everywhere else stay byte-identical. A
// missing or ambiguous field fails without writing anything.
func WriteVersionField(data []byte, field, value string) ([]byte, error) {
	if !validVersionFieldChar(value) {
		return nil, versionFieldFailure(
			"version "+strconv.Quote(value)+" carries characters outside the version grammar",
			"a version is three dot-separated decimal numbers",
			"pass the computed version, never a hand-written one",
		)
	}
	start, end := -1, -1
	matches := 0
	err := topLevelStringFields(data, func(key string, _ []byte, isString bool, s, e int) error {
		if key != field {
			return nil
		}
		if !isString {
			return versionFieldFailure(
				"the version file holds a non-string value for field "+strconv.Quote(field),
				"the protected field carries the version as a JSON string",
				"keep the field a quoted string, for example \"version\": \"0.1.0\"",
			)
		}
		matches++
		start, end = s, e
		return nil
	})
	if err != nil {
		return nil, err
	}
	if matches == 0 {
		return nil, versionFieldFailure(
			"the version file has no top-level field "+strconv.Quote(field),
			"only the named field is protected and written",
			"add the field at the file's top level, or fix the field name in the release config",
		)
	}
	out := make([]byte, 0, len(data)+len(value))
	out = append(out, data[:start]...)
	out = append(out, strconv.Quote(value)...)
	out = append(out, data[end:]...)
	if !json.Valid(out) {
		return nil, versionFieldFailure(
			"rewriting field "+strconv.Quote(field)+" broke the document",
			"a rewrite must keep valid JSON",
			"report this with the file content; do not hand-edit the version",
		)
	}
	return out, nil
}

// validVersionFieldChar reports whether a version value holds only version
// characters. The rewrite already validated the document; this keeps a stray
// value from ever landing in a version file.
func validVersionFieldChar(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) && r != '.' {
			return false
		}
	}
	return true
}
