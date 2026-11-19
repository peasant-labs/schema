package schema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	schema "github.com/peasant-labs/schema"
)

// TestVillageAPISpecJSON_MatchesCurrentVersion verifies that the public accessor
// serves the frozen numbered document named by VillageAPIVersion.
func TestVillageAPISpecJSON_MatchesCurrentVersion(t *testing.T) {
	got := schema.VillageAPISpecJSON()
	if len(got) == 0 {
		t.Fatal("VillageAPISpecJSON() returned empty bytes")
	}
	if !json.Valid(got) {
		t.Fatal("VillageAPISpecJSON() did not return valid JSON")
	}

	key := "village-api-" + schema.VillageAPIVersion + ".json"
	want, err := os.ReadFile("generated/" + key)
	if err != nil {
		t.Fatalf("read frozen spec %s: %v", key, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf(
			"VillageAPISpecJSON() does not match the frozen %s.\n"+
				"  what: the embedded accessor bytes differ from the numbered document.\n"+
				"  why:  the version constant or embedded document is inconsistent.\n"+
				"  fix:  restore the frozen document named by VillageAPIVersion.",
			key)
	}
}

// TestPublishRequestSchemaJSON_MatchesGenerated checks the numbered frozen
// request schema returned to the exported validator.
func TestPublishRequestSchemaJSON_MatchesGenerated(t *testing.T) {
	got := schema.PublishRequestSchemaJSON()
	if len(got) == 0 {
		t.Fatal("PublishRequestSchemaJSON() returned empty bytes")
	}
	if !json.Valid(got) {
		t.Fatal("PublishRequestSchemaJSON() did not return valid JSON")
	}

	key := "publish-request-" + schema.VillageAPIVersion + ".schema.json"
	want, err := os.ReadFile("generated/" + key)
	if err != nil {
		t.Fatalf("read frozen spec %s: %v", key, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf(
			"PublishRequestSchemaJSON() does not match the frozen %s.\n"+
				"  what: the embedded accessor bytes differ from the numbered request schema.\n"+
				"  why:  the version constant or embedded document is inconsistent.\n"+
				"  fix:  restore the frozen document named by VillageAPIVersion.",
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

	key := "annotation-push-request-" + schema.VillageAPIVersion + ".schema.json"
	want, err := os.ReadFile("generated/" + key)
	if err != nil {
		t.Fatalf("read frozen spec %s: %v", key, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("AnnotationPushRequestSchemaJSON() does not match frozen %s; restore the numbered document", key)
	}
}
