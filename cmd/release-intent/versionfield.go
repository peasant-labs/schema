package main

import (
	"fmt"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// versionFieldValue reads the protected top-level field from JSON bytes. It
// returns corrupt=true when the bytes do not parse, the field is missing, or
// the field is not a string: every caller fails closed on corrupt input. A
// missing file (nil data) yields ("", false) so an absent version file reads
// as empty. The field rule itself lives in the contract package, parsed once
// here; this only interprets its failure for the caller's fail-closed branch.
func versionFieldValue(data []byte, field string) (value string, corrupt bool) {
	if len(data) == 0 {
		return "", false
	}
	v, err := intent.ReadVersionField(data, field)
	if err != nil {
		return "", true
	}
	return v, false
}

// fileContent assembles the file-backed release content for a stale check:
// the changelog plus the protected field of every version file the generator
// wrote. A missing or corrupt file is an error, never an empty value, so a
// broken tree exits 1 instead of comparing as stale.
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
		val, corrupt := versionFieldValue(data, vf.Field)
		if corrupt {
			return intent.Content{}, fmt.Errorf("version file %q does not parse, so the protected field cannot be compared", vf.Path)
		}
		got.Versions[vf.Path] = val
	}
	return got, nil
}
