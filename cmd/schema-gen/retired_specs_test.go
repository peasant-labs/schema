package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// retiredSpec is a released OpenAPI spec version that is NO LONGER generated (the
// generator now emits a newer version) but is RETAINED in-tree as a frozen golden.
// Each is pinned to the sha256 of its committed bytes so the immutability guard
// fails on BOTH an in-place EDIT (hash mismatch) and a DELETION (missing file).
type retiredSpec struct {
	// name is the version-stamped artifact base (no extension), e.g.
	// "peasantlocal-api-0.1.0". The guard checks both the .json and .yaml copies
	// in the committed spec dir (generated/).
	name string
	// jsonSHA256 / yamlSHA256 are the lowercase-hex sha256 of the frozen .json /
	// .yaml bytes, pinned from the KNOWN-GOOD source (NOT a possibly-mutated
	// working tree):
	//   - peasantlocal-api-0.1.0: github/develop's committed copy (what SLICE-2
	//     froze when it bumped the live version to 0.2.0).
	//   - village-api-0.1.0:      rev 9745a0b0^ — the last rev before the village
	//     0.2.0 bump — which is the SAME source as the W4 restore, so the restored
	//     file's hash equals this pinned constant BY CONSTRUCTION.
	jsonSHA256 string
	yamlSHA256 string
	// jsonOnly marks a retired artifact that has ONLY a .json copy and no .yaml
	// sibling — the extracted publish-request JSON-Schema (publish-request-<v>.schema)
	// is emitted JSON-only by the generator (it is a JSON Schema, not an OpenAPI
	// spec doc), so freezing it must skip the .yaml present/hash check. yamlSHA256
	// is left empty for such entries.
	jsonOnly bool
}

// retiredSpecRegistry is the set of RETIRED spec versions under the generic
// released-versions-immutability guard.
//
// SCOPE (M-A): RETIRED versions ONLY. CURRENT-generated specs
// (peasantlocal-api-0.2.0, village-api-0.4.0, types-0.1.0,
// publish-request-0.4.0) are deliberately EXCLUDED — they stay under the
// codegen-freshness gate (TestCodegenFreshness_SpecsMatchSource), which regenerates
// them from the Go source on every run. Pinning a current version's hash here would
// false-fail `make check` on every legitimate regen. The partition key is simply
// "is this version still generated?": still-generated => freshness (mutable-by-regen);
// retired => this guard (immutable/frozen).
//
// REGISTER-AT-FREEZE-TIME: a version is MOVED into this registry at the moment it is
// frozen (i.e. in the same change that bumps the live const past it), so there is no
// window where a retired spec is mutable-and-unguarded. The 0.2.0 village-api trio
// (village-api-0.2.0 json+yaml + publish-request-0.2.0.schema json-only) was frozen
// here when VillageAPIVersion bumped to 0.3.0 (rc2 #118 required harness+model). The
// 0.3.0 village-api trio (village-api-0.3.0 json+yaml + publish-request-0.3.0.schema
// json-only) was frozen here when VillageAPIVersion bumped to 0.4.0 (added the
// optional top-level license enum).
var retiredSpecRegistry = []retiredSpec{
	{
		name:       "peasantlocal-api-0.1.0",
		jsonSHA256: "40b31eb6816a96c1283b2854e76d94936198b56c6ee78f16f02d774566791aee",
		yamlSHA256: "8c147003402e78f5a964135a689247c4876a2a980fcbd3be2c2bf6a553ce1b23",
	},
	{
		name:       "village-api-0.1.0",
		jsonSHA256: "114ceda1e4f6a22f1d770b0eb7964d6d4c32b0855b339b1892a2f75a88660c6d",
		yamlSHA256: "6e9710cdad192a7ba43355ca92317f2b994e009d7a288fa7d3b49ab11a0133f1",
	},
	// --- 0.2.0 village-api trio: frozen at the 0.3.0 bump (rc2 #118) ---
	{
		name:       "village-api-0.2.0",
		jsonSHA256: "39c89136901ee843daabd8cd25681541705d2e72b98e434c16cbe5aececa36ac",
		yamlSHA256: "8036e5dcc74f557561adea6ca838627bcf82c901a285107a17ad8d5300826aaa",
	},
	{
		// publish-request schema is JSON-only (a JSON Schema, not an OpenAPI doc),
		// so the file is publish-request-0.2.0.schema.json with no .yaml sibling.
		name:       "publish-request-0.2.0.schema",
		jsonSHA256: "2f55f5472ce1a0604ccc3fa41219ed9096a8845f0d84bc6ea5b607986a4f2657",
		jsonOnly:   true,
	},
	// --- 0.3.0 village-api trio: frozen at the 0.4.0 bump (added the license enum) ---
	{
		name:       "village-api-0.3.0",
		jsonSHA256: "3a4a4b5fb893247946dcf1b2ac5c99a30b437fe5070cf2aa54ecc4ff7d2f9344",
		yamlSHA256: "a3ed38c9f7a4e01890eb5ab2186c4945a83174a08d15cf08b45fc4ddc033de97",
	},
	{
		// json-only, like publish-request-0.2.0.schema above.
		name:       "publish-request-0.3.0.schema",
		jsonSHA256: "06eb63fde21cd6b0a2811ff48c3c3fcce7116d14aee9e66afd30f80f14af7285",
		jsonOnly:   true,
	},
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
	if len(retiredSpecRegistry) == 0 {
		t.Fatal("retiredSpecRegistry is empty: every newly-retired version MUST be registered the moment it is frozen (register-at-freeze-time), else it is mutable-and-unguarded")
	}

	for _, dir := range committedSpecDirs(root) {
		for _, rs := range retiredSpecRegistry {
			assertFrozen(t, filepath.Join(dir, rs.name+".json"), rs.jsonSHA256)
			// JSON-only artifacts (the extracted publish-request schema) have no
			// .yaml sibling — skip the yaml present/hash check for them.
			if !rs.jsonOnly {
				assertFrozen(t, filepath.Join(dir, rs.name+".yaml"), rs.yamlSHA256)
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
