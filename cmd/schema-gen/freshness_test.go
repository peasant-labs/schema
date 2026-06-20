package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/openapi"
)

// regenCmd is the canonical regeneration command — it matches the //go:generate
// directive in generate.go and is exactly what cmd/schema-gen's main() runs.
// Always regenerate with `go run ./cmd/schema-gen` and commit generated/ so the
// committed specs stay byte-identical to the Go source.
const regenCmd = "go run ./cmd/schema-gen"

// committedSpecDirs is the location the generator writes each spec to; it must
// stay byte-identical to the Go source. The module root is the single source of
// truth (specs live in generated/ at the schema repo root).
func committedSpecDirs(root string) []string {
	return []string{
		filepath.Join(root, "generated"),
	}
}

// TestCodegenFreshness_SpecsMatchSource regenerates the OpenAPI specs from the Go
// schema source (via the SAME generator cmd/schema-gen uses — not a mock) and
// asserts every committed JSON/YAML copy is byte-identical. It FAILS on drift:
// either a hand-edited committed spec, or a Go schema change (e.g. a key reverted
// to provider/modelHarness) that was not regenerated + committed.
//
// This gate covers THIS repo's committed generated/ specs — the schema module is
// the single source of truth that consumers (peasant, village) import. Regenerate
// with `go run ./cmd/schema-gen` and commit generated/ when the Go schema changes.
func TestCodegenFreshness_SpecsMatchSource(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts from Go source: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatal("generator produced no artifacts")
	}

	for _, dir := range committedSpecDirs(root) {
		for filename, want := range artifacts {
			path := filepath.Join(dir, filename)
			got, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("committed spec missing: %s\nRegenerate with `%s` and commit.", path, regenCmd)
				continue
			}
			if !bytes.Equal(got, want) {
				t.Errorf(
					"STALE committed spec: %s drifted from the Go schema source.\n"+
						"  what: the committed OpenAPI file does not match what the generator emits.\n"+
						"  why:  a schema change was not regenerated, or the file was hand-edited.\n"+
						"  fix:  run `%s` and commit the result.",
					path, regenCmd)
			}
		}
	}
}

// TestCodegenFreshness_UnifiedHarnessKey asserts the generated specs carry the
// unified `harness` wire key and NOT the legacy `modelHarness` key (the SLICE-B1
// emit-side flip). This makes a Go-source revert to the old key fail loudly here
// even before the byte-diff, with a clear cause.
func TestCodegenFreshness_UnifiedHarnessKey(t *testing.T) {
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts: %v", err)
	}

	// The village publish spec carries ModelInfo/SessionEntry — the flipped types.
	// Key is DERIVED from the single-source version const, so a version bump in
	// artifacts.go does not strand this assertion on a stale filename.
	villageKey := "village-api-" + openapi.VillageAPIVersion + ".json"
	village, ok := artifacts[villageKey]
	if !ok {
		t.Fatalf("%s not generated", villageKey)
	}
	if !bytes.Contains(village, []byte(`"harness"`)) {
		t.Error("generated village spec missing unified \"harness\" key")
	}
	if bytes.Contains(village, []byte(`"modelHarness"`)) {
		t.Error("generated village spec still contains legacy \"modelHarness\" key — emit-side flip incomplete/reverted")
	}
}
