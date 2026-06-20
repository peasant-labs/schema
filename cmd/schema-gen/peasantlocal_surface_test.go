package main

import (
	"bytes"
	"testing"

	"github.com/peasant-labs/schema/openapi"
)

// TestPeasantLocalSpec_ContainsMapSurface is the POSITIVE surface assertion (W6a):
// it proves the GENERATED peasantlocal-api-<current> spec actually CONTAINS the
// Map/Review/Search surface + the FrictionCluster schema. The codegen-freshness gate
// only proves the committed bytes match the generator; it would stay green even if a
// future edit silently DROPPED an op from BuildPeasantLocalAPISpec (the committed copy
// would just regenerate without it). This bytes.Contains spot-check — the analog of the
// village-api `"harness"` key check in freshness_test.go — catches that dropped surface.
//
// It is a SPOT-CHECK over the named ops: it pins exactly these surfaces, so a NEW 0.2.x
// op must be added to this list. That is inherent to a spot-check and acceptable (a
// dropped *existing* surface is the failure mode we guard).
func TestPeasantLocalSpec_ContainsMapSurface(t *testing.T) {
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts: %v", err)
	}

	// Key derived from the single-source version const, so a future bump does not
	// strand this assertion on a stale filename.
	name := "peasantlocal-api-" + openapi.PeasantLocalAPIVersion + ".json"
	spec, ok := artifacts[name]
	if !ok {
		t.Fatalf("%s not generated; got artifacts %v", name, keys(artifacts))
	}

	wantSurfaces := []string{
		// The 8 Map/Review/Search ops added in the 0.2.0 bump.
		"/api/v1/map/{projectHash}",
		"/api/v1/map/{projectHash}/node",
		"/api/v1/map/{projectHash}/tasks",
		"/api/v1/projects/summary",
		"/api/v1/review/{projectHash}",
		"/api/v1/review/{projectHash}/change",
		"/api/v1/review/{projectHash}/diff",
		"/api/v1/search",
		// The FrictionCluster schema (reflected transitively via the Review payload).
		"FrictionCluster",
	}
	for _, want := range wantSurfaces {
		if !bytes.Contains(spec, []byte(want)) {
			t.Errorf(
				"generated %s is MISSING surface %q.\n"+
					"  what: a Map/Review/Search op or the FrictionCluster schema was dropped from BuildPeasantLocalAPISpec.\n"+
					"  why:  codegen-freshness cannot catch a dropped surface (the committed copy just regenerates without it).\n"+
					"  fix:  restore the op/schema in openapi/peasantlocal.go and `go run ./cmd/schema-gen`.",
				name, want)
		}
	}
}

// keys returns the artifact filenames, for a helpful failure message.
func keys(m openapi.SpecArtifacts) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
