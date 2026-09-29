package openapi_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
)

// TestBuildVillageAPISpec_RequestBodies validates each fixture body against
// the request body schema the operation declares, as Village does when it
// enforces the served document.
func TestBuildVillageAPISpec_RequestBodies(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	document := currentVillageSpecDocument(t)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	compiler := schema.NewJSONSchemaCompiler()
	if err := compiler.AddResource("village.json", bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, fixture := range fixtures.RequestBodies {
		names[fixture.Name] = true
	}
	for _, required := range fixtures.RequestBodyNames {
		if !names[required] {
			t.Fatalf("request body case %q named in request_body_names is missing", required)
		}
	}
	if len(names) != len(fixtures.RequestBodyNames) || len(fixtures.RequestBodies) != len(fixtures.RequestBodyNames) {
		t.Fatalf("request_bodies has %d cases and request_body_names names %d; the manifest is the exact case set", len(fixtures.RequestBodies), len(fixtures.RequestBodyNames))
	}
	seen := map[string]bool{}
	for _, fixture := range fixtures.RequestBodies {
		t.Run(fixture.Name, func(t *testing.T) {
			if fixture.Name == "" || seen[fixture.Name] {
				t.Fatalf("request body fixture name %q is blank or repeated", fixture.Name)
			}
			seen[fixture.Name] = true
			operation := villageOperation(t, document, fixture.Path, fixture.Method)
			body, _ := operation["requestBody"].(map[string]any)
			content, _ := body["content"].(map[string]any)
			media, _ := content["application/json"].(map[string]any)
			schemaMap, _ := media["schema"].(map[string]any)
			ref, _ := schemaMap["$ref"].(string)
			if !strings.HasPrefix(ref, "#/components/schemas/") {
				t.Fatalf("%s %s declares no JSON request body component: %v", fixture.Method, fixture.Path, schemaMap)
			}
			contract, err := compiler.Compile("village.json" + ref)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(fixture.Body)
			if err != nil {
				t.Fatal(err)
			}
			var instance any
			if err := json.Unmarshal(encoded, &instance); err != nil {
				t.Fatal(err)
			}
			err = contract.Validate(instance)
			if (err == nil) != fixture.Valid {
				t.Fatalf("body %s accepted=%v, want %v: %v", encoded, err == nil, fixture.Valid, err)
			}
		})
	}
}
