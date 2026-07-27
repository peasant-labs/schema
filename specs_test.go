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
// schema.VillageAPIVersion — so village can serve
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
// schema.ValidatePublishRequest compiles and the village enforces through.
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

// TestPublishRequestSchemaJSON_IDMatchesVersion is the DEFENSE-IN-DEPTH $id guard
// for the version-aware accessor (the standalone's analogue of #118's
// TestPublishRequestSchema_EmbedMatchesVillageAPIVersion). The accessor reads
// generated/publish-request-<VillageAPIVersion>.schema.json by FILENAME, so a
// stale-by-construction read is already impossible (the only file it can name is
// the current version's). This asserts the same property from the CONTENT side:
// the bytes' $id must carry the CURRENT VillageAPIVersion. It catches a mismatch
// the filename indirection cannot — e.g. a hand-edited committed schema whose $id
// was not regenerated, or a generator change that stopped stamping $id from the
// version const — before the validator enforces against a wrongly-stamped schema.
func TestPublishRequestSchemaJSON_IDMatchesVersion(t *testing.T) {
	var doc struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(schema.PublishRequestSchemaJSON(), &doc); err != nil {
		t.Fatalf("PublishRequestSchemaJSON() is not valid JSON: %v", err)
	}
	want := "urn:peasant:publish-request:" + schema.VillageAPIVersion
	if doc.ID != want {
		t.Errorf("PublishRequestSchemaJSON() $id = %q; want %q.\n"+
			"  what: the version-aware accessor returns a schema whose $id differs from VillageAPIVersion.\n"+
			"  why:  the committed generated/publish-request-<version>.schema.json was hand-edited or not "+
			"regenerated after a VillageAPIVersion bump — the validator would enforce against a wrongly-stamped schema.\n"+
			"  fix:  run `go run ./cmd/schema-gen` and commit generated/, or correct VillageAPIVersion in versions.go.",
			doc.ID, want)
	}
}

// TestAnnotationPushRequestSchemaJSON_MatchesGenerated proves the canonical
// annotation request boundary compiles the generated operation-compatible
// artifact, rather than a hand-maintained surrogate schema.
func TestAnnotationPushRequestSchemaJSON_MatchesGenerated(t *testing.T) {
	got := schema.AnnotationPushRequestSchemaJSON()
	if len(got) == 0 {
		t.Fatal("AnnotationPushRequestSchemaJSON() returned empty bytes")
	}
	var document struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(got, &document); err != nil {
		t.Fatalf("AnnotationPushRequestSchemaJSON() is not valid JSON: %v", err)
	}
	wantID := "urn:peasant:annotation-push-request:" + schema.VillageAPIVersion
	if document.ID != wantID {
		t.Fatalf("AnnotationPushRequestSchemaJSON() $id=%q, want %q", document.ID, wantID)
	}

	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts: %v", err)
	}
	key := "annotation-push-request-" + schema.VillageAPIVersion + ".schema.json"
	want, ok := artifacts[key]
	if !ok {
		t.Fatalf("generator did not produce %q; cannot confirm the accessor is current", key)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("AnnotationPushRequestSchemaJSON() does not match freshly generated %s; run `go run ./cmd/schema-gen` and commit generated/", key)
	}
}
