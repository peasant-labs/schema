package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// retiredSpec is a released OpenAPI spec version that is NO LONGER generated (the
// generator now emits a newer version) but is RETAINED in-tree as a frozen golden.
// Each is pinned to the sha256 of its committed bytes so the immutability guard
// fails on BOTH an in-place EDIT (hash mismatch) and a DELETION (missing file).
type retiredSpec struct {
	// name is the version-stamped artifact base (no extension), e.g.
	// "peasantlocal-api-0.1.0". The guard checks both the .json and .yaml copies
	// in the committed spec dir (generated/).
	Name string `yaml:"name"`
	// jsonSHA256 / yamlSHA256 are the lowercase-hex sha256 of the frozen .json /
	// .yaml bytes, pinned from the KNOWN-GOOD source (NOT a possibly-mutated
	// working tree):
	//   - peasantlocal-api-0.1.0: github/develop's committed copy from immediately
	//     before the live version was bumped to 0.2.0.
	//   - village-api-0.1.0:      rev 9745a0b0^ — the last rev before the village
	//     0.2.0 bump — which is the SAME source as the W4 restore, so the restored
	//     file's hash equals this pinned constant BY CONSTRUCTION.
	JSONSHA256 string `yaml:"json_sha256"`
	YAMLSHA256 string `yaml:"yaml_sha256"`
	// jsonOnly marks a retired artifact that has ONLY a .json copy and no .yaml
	// sibling — the extracted publish-request JSON-Schema (publish-request-<v>.schema)
	// is emitted JSON-only by the generator (it is a JSON Schema, not an OpenAPI
	// spec doc), so freezing it must skip the .yaml present/hash check. yamlSHA256
	// is left empty for such entries.
	JSONOnly bool `yaml:"json_only"`
}

// retiredSpecRegistry is the set of RETIRED spec versions under the generic
// released-versions-immutability guard.
//
// This registry covers RETIRED versions ONLY. CURRENT-generated specs (the
// artifacts derived from PeasantLocalAPIVersion, VillageAPIVersion, and
// TypesVersion, plus the publish-request schema derived from VillageAPIVersion)
// are deliberately EXCLUDED — they stay under the codegen-freshness gate
// (TestCodegenFreshness_SpecsMatchSource), which regenerates them from the Go
// source on every run. Pinning a current version's hash here would false-fail
// `make check` on every legitimate regen. The partition key is simply "is this
// version still generated?": still-generated => freshness (mutable-by-regen);
// retired => this guard (immutable/frozen).
//
// REGISTER-AT-FREEZE-TIME: a version is MOVED into this registry at the moment it is
// frozen (i.e. in the same change that bumps the live const past it), so there is no
// window where a retired spec is mutable-and-unguarded. The 0.2.0 village-api trio
// (village-api-0.2.0 json+yaml + publish-request-0.2.0.schema json-only) was frozen
// here when VillageAPIVersion bumped to 0.3.0 (rc2 #118 required harness+model). Each
// row now lives in testdata/retired_specs.yaml so adding a newly frozen version
// extends the fixture rather than an inline test matrix.
func loadRetiredSpecRegistry(t *testing.T, root string) []retiredSpec {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "cmd", "schema-gen", "testdata", "retired_specs.yaml"))
	if err != nil {
		t.Fatalf("read retired spec registry fixture: %v", err)
	}
	var fixture struct {
		Specs []retiredSpec `yaml:"specs"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode retired spec registry fixture: %v", err)
	}
	if len(fixture.Specs) == 0 {
		t.Fatal("retired spec registry fixture is empty: register every frozen version")
	}
	seen := make(map[string]struct{}, len(fixture.Specs))
	for _, spec := range fixture.Specs {
		if spec.Name == "" || spec.JSONSHA256 == "" || (!spec.JSONOnly && spec.YAMLSHA256 == "") {
			t.Fatalf("retired spec registry row %q is incomplete; provide every required frozen hash", spec.Name)
		}
		if _, duplicate := seen[spec.Name]; duplicate {
			t.Fatalf("retired spec registry repeats %q", spec.Name)
		}
		seen[spec.Name] = struct{}{}
	}
	return fixture.Specs
}

// TestRetiredSpecsImmutable is the generic released-versions-immutability guard. It
// asserts every RETIRED spec version is PRESENT in the committed spec dir
// (generated/) and byte-frozen to its pinned sha256. It fails loudly on:
//   - an in-place EDIT of a released spec (the original peasantlocal-api-0.1.0
//     mutation bug — adding new surface onto an already-shipped version instead of a
//     new one); and
//   - a DELETION of a released spec (the present-check).
//
// The codegen-freshness gate cannot catch either, because it only diffs the
// CURRENTLY-generated artifacts against source — a retired version it no longer emits
// is invisible to it. This guard closes exactly that gap. CURRENT-generated specs are
// intentionally out of scope here (see retiredSpecRegistry SCOPE note).
func TestRetiredSpecsImmutable(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	retiredSpecRegistry := loadRetiredSpecRegistry(t, root)

	for _, dir := range committedSpecDirs(root) {
		for _, rs := range retiredSpecRegistry {
			assertFrozen(t, filepath.Join(dir, rs.Name+".json"), rs.JSONSHA256)
			// JSON-only artifacts (the extracted publish-request schema) have no
			// .yaml sibling — skip the yaml present/hash check for them.
			if !rs.JSONOnly {
				assertFrozen(t, filepath.Join(dir, rs.Name+".yaml"), rs.YAMLSHA256)
			}
		}
	}
}

// assertFrozen reads the committed file and asserts it is PRESENT and its sha256
// matches the pinned constant. The byte check itself lives in checkFrozen (returns
// an error) so it can be exercised directly by the negative-control self-test.
func assertFrozen(t *testing.T, path, wantSHA string) {
	t.Helper()
	if err := checkFrozen(path, wantSHA); err != nil {
		t.Error(err)
	}
}

// checkFrozen is the pure core of the immutability guard: it returns a non-nil,
// actionable error if the committed file at path is ABSENT or its sha256 does not
// match wantSHA, and nil only when the file is present AND byte-frozen to wantSHA.
// Returning an error (rather than calling t.Errorf) is what lets TestCheckFrozen_
// NegativeControl prove the guard actually FAILS on a wrong-hash / missing input —
// locking the guard's own correctness against a regression that made it always-pass.
func checkFrozen(path, wantSHA string) error {
	got, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(
			"RETIRED spec MISSING: %s\n"+
				"  what: a frozen, released OpenAPI spec is absent (deleted or moved).\n"+
				"  why:  retired versions are immutable goldens and must remain committed in generated/.\n"+
				"  fix:  restore it from git history (it is no longer regenerated) and re-run; do not delete retired specs.",
			path)
	}

	sum := sha256.Sum256(got)
	if gotSHA := hex.EncodeToString(sum[:]); gotSHA != wantSHA {
		return fmt.Errorf(
			"RETIRED spec MUTATED: %s drifted from its pinned hash.\n"+
				"  what: a RETIRED (frozen) OpenAPI spec was edited in place.\n"+
				"  why:  released versions are immutable; NEW API surface goes on a NEW version (bump the version const + `go run ./cmd/schema-gen`), never onto an already-shipped spec.\n"+
				"  got:  %s\n"+
				"  want: %s\n"+
				"  fix:  revert the edit. If you intended new surface, bump the live version const and regenerate — the new version is emitted, this old one stays frozen.",
			path, gotSHA, wantSHA)
	}
	return nil
}

// TestCheckFrozen_NegativeControl is a PERMANENT self-test of the immutability
// guard's own correctness. It proves checkFrozen RETURNS AN ERROR on (1) a
// wrong-hash (mutated) input and (2) a missing file, and returns NIL on a
// matching-hash input. Without this control, a regression that made checkFrozen
// unconditionally pass (e.g. comparing a value to itself) would silently disable
// TestRetiredSpecsImmutable while still showing green.
//
// It is hermetic: it writes a SYNTHETIC fixture under t.TempDir() and never reads
// or mutates the real committed specs.
func TestCheckFrozen_NegativeControl(t *testing.T) {
	dir := t.TempDir()
	synthetic := []byte("synthetic retired-spec bytes — not a real spec\n")
	path := filepath.Join(dir, "synthetic-api-0.0.0.json")
	if err := os.WriteFile(path, synthetic, 0o644); err != nil {
		t.Fatalf("write synthetic fixture: %v", err)
	}

	// True hash of the synthetic bytes — the positive control.
	sum := sha256.Sum256(synthetic)
	trueSHA := hex.EncodeToString(sum[:])

	// (1) wrong-hash MUST fail — else the guard would not catch an in-place edit.
	const wrongSHA = "0000000000000000000000000000000000000000000000000000000000000000"
	if err := checkFrozen(path, wrongSHA); err == nil {
		t.Fatal("checkFrozen returned nil for a WRONG hash: the immutability guard would NOT catch an in-place edit of a retired spec")
	}

	// (2) matching-hash MUST pass — proving (1)'s failure is attributable to the
	// hash mismatch, not an unconditional error.
	if err := checkFrozen(path, trueSHA); err != nil {
		t.Fatalf("checkFrozen returned an error for a MATCHING hash: %v", err)
	}

	// (3) missing file MUST fail — else the guard would not catch a deletion.
	missing := filepath.Join(dir, "does-not-exist-0.0.0.json")
	if err := checkFrozen(missing, wrongSHA); err == nil {
		t.Fatal("checkFrozen returned nil for a MISSING file: the immutability guard would NOT catch a deletion of a retired spec")
	}
}
