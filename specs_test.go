package schema_test

import (
	"bytes"
	"encoding/json"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
)

// TestVillageAPISpecJSON_MatchesCurrentVersion is the conformance test for the
// W9 version-aware accessor (IP3). It proves schema.VillageAPISpecJSON() returns
// the CURRENT Village API spec bytes — exactly what the generator emits for
// schema.VillageAPIVersion — so village (the consumer, SLICE-D) can serve
// GET /openapi.json from it without vendoring or hard-coding a version filename.
//
// It is an external (schema_test) test so it may import both schema and openapi
// without the import cycle that forbids the root package from importing openapi.
func TestVillageAPISpecJSON_MatchesCurrentVersion(t *testing.T) {
	got := schema.VillageAPISpecJSON()
	if len(got) == 0 {
		t.Fatal("VillageAPISpecJSON() returned empty bytes")
	}
	if !json.Valid(got) {
		t.Fatal("VillageAPISpecJSON() did not return valid JSON")
	}

	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts: %v", err)
	}
	key := "village-api-" + schema.VillageAPIVersion + ".json"
	want, ok := artifacts[key]
	if !ok {
		t.Fatalf("generator did not produce %q; cannot confirm the accessor is current", key)
	}
	if !bytes.Equal(got, want) {
		t.Errorf(
			"VillageAPISpecJSON() does not match the freshly generated %s.\n"+
				"  what: the embedded accessor bytes differ from the generator output for the current version.\n"+
				"  why:  generated/ is stale or VillageAPIVersion drifted from the committed artifact.\n"+
				"  fix:  run `go run ./cmd/schema-gen` and commit generated/.",
			key)
	}
}

// TestPublishRequestSchemaJSON_MatchesGenerated is the conformance test for the
// W9 single-byte-source PublishRequest accessor (IP3). It proves
// schema.PublishRequestSchemaJSON() returns exactly the standalone PublishRequest
// JSON-Schema the generator emits for schema.VillageAPIVersion — the same bytes
// validate.ValidatePublishRequest compiles and the village enforces through.
// Coupling the accessor, the validator, and the documented spec to one generated
// artifact is what guarantees the village's 422 behavior cannot silently drift on
// re-pin (it is what retired the divergent hand-maintained validate/schema.json).
func TestPublishRequestSchemaJSON_MatchesGenerated(t *testing.T) {
	got := schema.PublishRequestSchemaJSON()
	if len(got) == 0 {
		t.Fatal("PublishRequestSchemaJSON() returned empty bytes")
	}
	if !json.Valid(got) {
		t.Fatal("PublishRequestSchemaJSON() did not return valid JSON")
	}

	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts: %v", err)
	}
	key := "publish-request-" + schema.VillageAPIVersion + ".schema.json"
	want, ok := artifacts[key]
	if !ok {
		t.Fatalf("generator did not produce %q; cannot confirm the accessor is current", key)
	}
	if !bytes.Equal(got, want) {
		t.Errorf(
			"PublishRequestSchemaJSON() does not match the freshly generated %s.\n"+
				"  what: the embedded accessor bytes differ from the generator output for the current version.\n"+
				"  why:  generated/ is stale or VillageAPIVersion drifted from the committed artifact.\n"+
				"  fix:  run `go run ./cmd/schema-gen` and commit generated/.",
			key)
	}
}
