package openapi_test

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
)

// TestVillageSpec_RequirednessMatchesCanonicalTypes asserts that every shared
// Village component has the exact canonical Types required set. The distinct
// operation-only publish body retains its stricter model requirement without
// shadowing the canonical PublishRequest name.
//
// Out of scope by construction: operation/path parameter `required: true` booleans
// live under paths/.../parameters, NOT components/schemas, so examining only
// components/schemas excludes them (a parameter's required is a bool, not the
// object-schema required ARRAY this guard inspects).
func TestVillageSpec_RequirednessMatchesCanonicalTypes(t *testing.T) {
	artifacts, err := specpkg.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("GenerateSpecArtifacts: %v", err)
	}
	// Key derives from the single-source version const, so a version bump does not
	// strand this assertion on a stale filename.
	key := "village-api-" + specpkg.VillageAPIVersion + ".json"
	raw, ok := artifacts[key]
	if !ok {
		t.Fatalf("generator did not emit %s; check VillageAPIVersion / BuildVillageAPISpec", key)
	}

	type componentDocument struct {
		Components struct {
			Schemas map[string]struct {
				Required []string `json:"required"`
			} `json:"schemas"`
		} `json:"components"`
	}
	var doc componentDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	if len(doc.Components.Schemas) == 0 {
		t.Fatalf("%s has no components/schemas — spec shape changed unexpectedly", key)
	}

	typesSpec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec: %v", err)
	}
	typesRaw, err := typesSpec.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal Types spec: %v", err)
	}
	var canonical componentDocument
	if err := json.Unmarshal(typesRaw, &canonical); err != nil {
		t.Fatalf("unmarshal Types spec: %v", err)
	}
	publishRequired := append([]string(nil), doc.Components.Schemas["OpenapiTranscriptPublishRequest"].Required...)
	sort.Strings(publishRequired)
	if !reflect.DeepEqual(publishRequired, []string{"contentHash", "model"}) {
		t.Fatalf("Village successor publish metadata required=%v, want [contentHash model]; visibilityIntent is optional for legacy compatibility", publishRequired)
	}
	for name, component := range doc.Components.Schemas {
		canonicalName := strings.TrimPrefix(name, "Schema")
		if canonicalName == "BestiaryHarness" || canonicalName == "Provider" {
			canonicalName = "Harness"
		}
		root, shared := canonical.Components.Schemas[canonicalName]
		if !shared {
			continue
		}
		got := append([]string(nil), component.Required...)
		want := append([]string(nil), root.Required...)
		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Village component %s required=%v, canonical Types %s required=%v; harmonize the shared component or give a genuinely operation-specific schema a distinct name", name, got, canonicalName, want)
		}
	}
}

// TestPublishVerdicts_ErrorContainsPinnedStrings LOCKS every verdict row's pinned
// `error_contains` substring to the REAL santhosh-tekuri/jsonschema/v5 message the
// validator actually emits for that body against the extracted 0.3.0 publish
// schema. Without this, a wrong/guessed pinned string would silently pass (the
// parity test only checks accept/reject, not the message) and mislead any consumer
// that asserts on these substrings.
func TestPublishVerdicts_ErrorContainsPinnedStrings(t *testing.T) {
	newRaw, err := os.ReadFile("../generated/publish-request-0.10.0.schema.json")
	if err != nil {
		t.Fatalf("read frozen publish schema: %v", err)
	}
	sch := compileSchema(t, "extracted.json", newRaw)

	fixtures, err := schema.LoadPublishVerdictFixtures()
	if err != nil {
		t.Fatalf("LoadPublishVerdictFixtures: %v", err)
	}

	checked := 0
	for _, tc := range fixtures.Cases {
		if tc.Expect.ErrorContains == "" {
			continue
		}
		checked++
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Expect.SchemaAccepts {
				t.Fatalf("row %q sets error_contains but schema_accepts:true — an accepted body has no validator error", tc.Name)
			}
			var body any
			if err := json.Unmarshal([]byte(tc.Body), &body); err != nil {
				t.Fatalf("row %q body is not valid JSON: %v", tc.Name, err)
			}
			verr := sch.Validate(body)
			if verr == nil {
				t.Fatalf("row %q expected a validation error (error_contains=%q) but the 0.3.0 schema ACCEPTED the body", tc.Name, tc.Expect.ErrorContains)
			}
			if !strings.Contains(verr.Error(), tc.Expect.ErrorContains) {
				t.Errorf("row %q: validator message does not contain the pinned error_contains.\n"+
					"  want substring: %q\n"+
					"  got message:    %s\n"+
					"  fix: re-pin error_contains to the REAL santhosh-tekuri/jsonschema/v5 message (do not guess).",
					tc.Name, tc.Expect.ErrorContains, verr.Error())
			}
		})
	}
	if checked == 0 {
		t.Fatal("no verdict rows carried error_contains — the pinned-string lock exercised nothing")
	}
}
