package openapi_test

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
)

// TestVillageSpec_SelectiveRequiredScope is the rc2 (#118) SELECTIVE-SCOPE guard.
// It parses the REGENERATED village-api spec (the exact bytes the generator emits)
// and asserts that EXACTLY the SchemaModelInfo and SchemaPublishRequest component
// schemas carry a non-empty `required` array — with the EXACT contents
// [harness,model] and [model]. This proves the required strictening is selective:
// no other component schema gains a required array (no global/blanket flip), and
// no `required:"true"` tag leaks onto another struct.
//
// Out of scope by construction: operation/path parameter `required: true` booleans
// live under paths/.../parameters, NOT components/schemas, so examining only
// components/schemas excludes them (a parameter's required is a bool, not the
// object-schema required ARRAY this guard inspects).
func TestVillageSpec_SelectiveRequiredScope(t *testing.T) {
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

	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Required []string `json:"required"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	if len(doc.Components.Schemas) == 0 {
		t.Fatalf("%s has no components/schemas — spec shape changed unexpectedly", key)
	}

	got := map[string][]string{}
	for name, s := range doc.Components.Schemas {
		if len(s.Required) > 0 {
			r := append([]string(nil), s.Required...)
			sort.Strings(r) // order is not semantically meaningful in JSON Schema required
			got[name] = r
		}
	}

	want := map[string][]string{
		"SchemaModelInfo":      {"harness", "model"},
		"SchemaPublishRequest": {"model"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SELECTIVE-required scope drift in %s.\n"+
			"  what: the set/contents of component schemas carrying a non-empty `required` array changed.\n"+
			"  why:  rc2 #118 requires EXACTLY SchemaModelInfo=[harness,model] + SchemaPublishRequest=[model] — no other schema, no global flip.\n"+
			"  got:  %v\n"+
			"  want: %v\n"+
			"  fix:  a new required:\"true\" tag leaked onto another struct, or a tag was dropped — reconcile struct tags (model.go/publish.go) + `make schema`.",
			key, got, want)
	}
}

// TestPublishVerdicts_ErrorContainsPinnedStrings LOCKS every verdict row's pinned
// `error_contains` substring to the REAL santhosh-tekuri/jsonschema/v5 message the
// validator actually emits for that body against the extracted 0.3.0 publish
// schema. Without this, a wrong/guessed pinned string would silently pass (the
// parity test only checks accept/reject, not the message) and mislead any consumer
// that asserts on these substrings.
func TestPublishVerdicts_ErrorContainsPinnedStrings(t *testing.T) {
	newRaw, err := specpkg.BuildPublishRequestSchema(specpkg.VillageAPIVersion)
	if err != nil {
		t.Fatalf("BuildPublishRequestSchema: %v", err)
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
