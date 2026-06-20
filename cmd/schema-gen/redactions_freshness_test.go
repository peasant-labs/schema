package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema"
)

// TestRedactionsFreshness_MatchesSource regenerates testdata/session-detail/redactions.yaml
// from schema.RedactionExamples (via the SAME buildRedactionsYAML() main() uses —
// not a mock) and asserts the committed file is byte-identical. It FAILS on drift:
// either a hand-edited committed YAML, or an edit to RedactionExamples that was not
// regenerated + committed.
//
// This is a SHAPE/DETERMINISM gate only — it CANNOT verify engine alignment (the
// leaf has no redaction engine). The no-drift guarantee for the redacted OUTPUT is
// the peasant-side behavioural conformance test (pkg/redact/redactconform_test.go).
func TestRedactionsFreshness_MatchesSource(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	want, err := buildRedactionsYAML()
	if err != nil {
		t.Fatalf("generate redactions YAML from Go source: %v", err)
	}
	if len(want) == 0 {
		t.Fatal("generator produced empty redactions YAML")
	}

	path := filepath.Join(root, redactionsYAMLRelPath)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("committed redactions fixture missing: %s\nRegenerate with `%s` and commit.", path, regenCmd)
	}
	if !bytes.Equal(got, want) {
		t.Errorf(
			"STALE committed fixture: %s drifted from RedactionExamples.\n"+
				"  what: the committed redactions.yaml does not match what the generator emits.\n"+
				"  why:  RedactionExamples changed without regen, or the file was hand-edited.\n"+
				"  fix:  run `%s` and commit the result.",
			path, regenCmd)
	}
}

// TestRedactionsFreshness_RoundTrips asserts the generated YAML unmarshals back to
// the source slice via the published loader, so the embedded artifact peasant +
// web consume is faithful to RedactionExamples (catches a yaml-tag/field mismatch
// that the byte gate alone would not explain).
func TestRedactionsFreshness_RoundTrips(t *testing.T) {
	cases, err := schema.LoadRedactionExamples()
	if err != nil {
		t.Fatalf("LoadRedactionExamples: %v", err)
	}
	if len(cases) != len(schema.RedactionExamples) {
		t.Fatalf("round-trip count mismatch: loaded %d, source has %d", len(cases), len(schema.RedactionExamples))
	}
	for i, want := range schema.RedactionExamples {
		if cases[i] != want {
			t.Errorf("round-trip case %d (%q) differs:\n got: %+v\nwant: %+v", i, want.Name, cases[i], want)
		}
	}
}
