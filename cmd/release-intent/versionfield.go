package main

import (
	"fmt"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// versionFieldValue reads the protected top-level field from JSON bytes. A
// missing file (nil data) yields ("", nil) so an absent version file reads as
// empty. Every other fault — unparseable bytes, a missing field, a non-string
// value — returns the contract reader's error, which names the exact fault;
// callers fail closed on it. The field rule itself lives in the contract
// package, parsed once here.
func versionFieldValue(data []byte, field string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	return intent.ReadVersionField(data, field)
}

// firstLine keeps the first line of a possibly multi-line error for the
// single-line What of a rejection.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// fileContent assembles the file-backed release content for a stale check:
// the changelog plus the protected field of every version file the generator
// wrote. A missing or corrupt file is an error, never an empty value, so a
// broken tree exits 1 instead of comparing as stale. The fault text matches
// the protected-path rejection: the contract reader names the exact fault in
// both.
func fileContent(read func(path string) ([]byte, error), cfg intent.Config, want intent.Content) (intent.Content, error) {
	got := intent.Content{Versions: map[string]string{}}
	changelog, err := read(cfg.Changelog)
	if err != nil {
		return intent.Content{}, fmt.Errorf("cannot read changelog %q: %w", cfg.Changelog, err)
	}
	got.Changelog = string(changelog)
	for _, vf := range cfg.VersionFiles {
		if _, ok := want.Versions[vf.Path]; !ok {
			continue
		}
		data, err := read(vf.Path)
		if err != nil {
			return intent.Content{}, fmt.Errorf("cannot read version file %q: %w", vf.Path, err)
		}
		val, ferr := versionFieldValue(data, vf.Field)
		if ferr != nil {
			return intent.Content{}, fmt.Errorf("cannot read the protected field from version file %q: %w", vf.Path, ferr)
		}
		got.Versions[vf.Path] = val
	}
	return got, nil
}
