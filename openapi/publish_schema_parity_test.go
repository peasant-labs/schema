package openapi_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
)

// legacyPublishSchema is a FROZEN copy of the now-retired BuildJSONSchema output
// (BuildJSONSchema("0.1.0"), byte-identical to what the village vendored before
// this migration). It is committed so the enforcement-PARITY gate below keeps
// pinning equivalence even after BuildJSONSchema is deleted — the new extracted
// schema must accept/reject the same bodies the old one did.
//
//go:embed testdata/legacy-publish-request.schema.json
var legacyPublishSchema []byte

// compileSchema compiles raw JSON-Schema bytes with santhosh-tekuri/jsonschema —
// the SAME validator the village uses in production to enforce publish bodies
// (backend/internal/handler/openapi.go newSchemaValidator). Using the real
// validator (not a hand-rolled checker) is what makes this a meaningful parity
// gate: it proves the village's enforcement verdict is unchanged.
func compileSchema(t *testing.T, name string, raw []byte) *jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	if err := c.AddResource(name, bytes.NewReader(raw)); err != nil {
		t.Fatalf("add %s schema resource: %v", name, err)
	}
	sch, err := c.Compile(name)
	if err != nil {
		t.Fatalf("compile %s schema (the extracted schema must use a draft the village "+
			"validator supports — 2020-12): %v", name, err)
	}
	return sch
}

// accepts reports whether sch accepts the given JSON body (nil error == accept).
func accepts(t *testing.T, sch *jsonschema.Schema, body []byte) bool {
	t.Helper()
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("corpus body is not valid JSON: %v\nbody: %s", err, body)
	}
	return sch.Validate(doc) == nil
}

// TestPublishSchema_EnforcementParity is the RETIREMENT GATE for BuildJSONSchema.
// It compiles BOTH the frozen legacy schema and the freshly-extracted
// village-api-derived schema with the village's real validator and asserts they
// return IDENTICAL accept/reject verdicts across the corpus. Equivalence here is
// the precondition for deleting BuildJSONSchema and migrating the village to
// enforce against the extracted schema. The committed legacy copy keeps this
// pinned after deletion.
func TestPublishSchema_EnforcementParity(t *testing.T) {
	newRaw, err := os.ReadFile("../generated/publish-request-0.10.0.schema.json")
	if err != nil {
		t.Fatalf("read frozen extracted publish schema: %v", err)
	}

	legacy := compileSchema(t, "legacy.json", legacyPublishSchema)
	extracted := compileSchema(t, "extracted.json", newRaw)
	fixtures, err := schema.LoadPublishVerdictFixtures()
	if err != nil {
		t.Fatalf("LoadPublishVerdictFixtures: %v", err)
	}

	for _, tc := range fixtures.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			body := []byte(tc.Body)
			gotLegacy := accepts(t, legacy, body)
			gotExtracted := accepts(t, extracted, body)
			wantLegacy := tc.Expect.SchemaAccepts
			if tc.Expect.LegacyAccepts != nil {
				wantLegacy = *tc.Expect.LegacyAccepts
			}

			if tc.Expect.LegacyAccepts == nil && gotLegacy != gotExtracted {
				t.Fatalf("PARITY BROKEN for %q: legacy accepts=%v, extracted accepts=%v.\n"+
					"  what: the village-api-derived publish schema disagrees with the retired "+
					"BuildJSONSchema output on this body.\n"+
					"  why:  a semantic difference (e.g. a tightened enum or type) changes the "+
					"village's 422 verdict.\n"+
					"  fix:  reconcile BuildPublishRequestSchema with the verdict fixture, or surface "+
					"the intended divergence for sign-off before retiring the legacy builder.\n"+
					"  body: %s", tc.Name, gotLegacy, gotExtracted, tc.Body)
			}
			if gotLegacy != wantLegacy {
				t.Errorf("%q: legacy schema accept=%v; want %v", tc.Name, gotLegacy, wantLegacy)
			}
			if gotExtracted != tc.Expect.SchemaAccepts {
				t.Errorf("%q: extracted schema accept=%v; want %v", tc.Name, gotExtracted, tc.Expect.SchemaAccepts)
			}
		})
	}
}

// TestPublishSchema_ExtractedShape pins the structural contract of the extracted
// schema: 2020-12 dialect, urn $id derived from VillageAPIVersion, hoisted
// PublishRequest properties at the root, and a self-contained $defs with no stray
// OpenAPI components/schemas pointers.
func TestPublishSchema_ExtractedShape(t *testing.T) {
	raw, err := specpkg.BuildPublishRequestSchema(specpkg.VillageAPIVersion)
	if err != nil {
		t.Fatalf("BuildPublishRequestSchema: %v", err)
	}

	if bytes.Contains(raw, []byte("#/components/schemas/")) {
		t.Error("extracted schema still contains OpenAPI '#/components/schemas/' pointers; " +
			"refs were not rewritten to '#/$defs/' and the schema is not self-contained")
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("extracted schema is not valid JSON: %v", err)
	}

	if got := doc["$schema"]; got != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("$schema = %v; want the JSON Schema 2020-12 dialect URI", got)
	}
	if got, want := doc["$id"], "urn:peasant:publish-request:"+specpkg.VillageAPIVersion; got != want {
		t.Errorf("$id = %v; want %q (derived from VillageAPIVersion)", got, want)
	}
	if _, ok := doc["$defs"].(map[string]any); !ok {
		t.Error("extracted schema missing a $defs object of bundled component definitions")
	}
	props, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatal("extracted schema missing root 'properties' (PublishRequest body not hoisted)")
	}
	for _, field := range []string{"identity", "model", "source", "entries", "timestamp"} {
		if _, ok := props[field]; !ok {
			t.Errorf("extracted schema root properties missing PublishRequest field %q", field)
		}
	}
}
