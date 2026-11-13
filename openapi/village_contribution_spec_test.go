package openapi_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
)

// TestBuildVillageAPISpec_RetiredNamesStayRetired fails if a name the contract
// dropped returns as a parameter, a property or an enum value anywhere in the
// served document.
func TestBuildVillageAPISpec_RetiredNamesStayRetired(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	if len(fixtures.RetiredNames) == 0 {
		t.Fatal("retired_names is empty; the absence check would be vacuous")
	}
	found := map[string]map[string]bool{"parameter": {}, "property": {}, "enum_value": {}}
	walkVillageNames(currentVillageSpecDocument(t), found)
	for kind, names := range found {
		if len(names) == 0 {
			t.Fatalf("the walk found no %s names; the absence check would be vacuous", kind)
		}
	}
	for _, retired := range fixtures.RetiredNames {
		if retired.Name == "" || retired.Reason == "" {
			t.Fatalf("retired name fixture has a blank field: %+v", retired)
		}
		declared, ok := found[retired.Kind]
		if !ok {
			t.Fatalf("retired name %q has unknown kind %q", retired.Name, retired.Kind)
		}
		if declared[retired.Name] {
			t.Errorf("%s %q was retired (%s) but the contract declares it again", retired.Kind, retired.Name, retired.Reason)
		}
	}
}

// walkVillageNames records every parameter name, property name and enum value
// reachable in a decoded spec.
func walkVillageNames(node any, found map[string]map[string]bool) {
	switch v := node.(type) {
	case map[string]any:
		if _, in := v["in"]; in {
			if name, ok := v["name"].(string); ok {
				found["parameter"][name] = true
			}
		}
		if properties, ok := v["properties"].(map[string]any); ok {
			for name := range properties {
				found["property"][name] = true
			}
		}
		if values, ok := v["enum"].([]any); ok {
			for _, value := range values {
				if name, ok := value.(string); ok {
					found["enum_value"][name] = true
				}
			}
		}
		for _, child := range v {
			walkVillageNames(child, found)
		}
	case []any:
		for _, child := range v {
			walkVillageNames(child, found)
		}
	}
}

// TestBuildVillageAPISpec_ResponseArmsAreExclusive validates one sample of each
// success arm of a oneOf response against the served document: the union
// accepts it and exactly one arm, the named one, does.
func TestBuildVillageAPISpec_ResponseArmsAreExclusive(t *testing.T) {
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
	var sampleNames []string
	for _, fixture := range fixtures.ResponseArmSamples {
		for _, sample := range fixture.Samples {
			sampleNames = append(sampleNames, sample.Name)
		}
	}
	assertSameStringSet(t, "response arm sample names", sampleNames, fixtures.ResponseArmSampleNames)
	if len(sampleNames) != len(fixtures.ResponseArmSampleNames) {
		t.Fatalf("response_arm_samples has %d samples and response_arm_sample_names names %d; the manifest is the exact case set", len(sampleNames), len(fixtures.ResponseArmSampleNames))
	}
	for _, fixture := range fixtures.ResponseArmSamples {
		operation := villageOperation(t, document, fixture.Path, fixture.Method)
		responses, _ := operation["responses"].(map[string]any)
		response, _ := responses[fixture.Status].(map[string]any)
		content, _ := response["content"].(map[string]any)
		media, _ := content["application/json"].(map[string]any)
		union, _ := media["schema"].(map[string]any)
		arms, _ := union["oneOf"].([]any)
		declared := make([]string, 0, len(arms))
		for _, arm := range arms {
			ref, _ := arm.(map[string]any)["$ref"].(string)
			declared = append(declared, strings.TrimPrefix(ref, "#/components/schemas/"))
		}
		if len(declared) != len(fixture.Arms) {
			t.Fatalf("%s %s %s declares arms %v, want %v", fixture.Method, fixture.Path, fixture.Status, declared, fixture.Arms)
		}
		for i := range declared {
			if declared[i] != fixture.Arms[i] {
				t.Fatalf("%s %s %s arm %d = %s, want %s", fixture.Method, fixture.Path, fixture.Status, i, declared[i], fixture.Arms[i])
			}
		}
		if len(fixture.Samples) != len(declared) {
			t.Fatalf("%s %s has %d samples for %d arms; every arm needs one", fixture.Method, fixture.Path, len(fixture.Samples), len(declared))
		}
		pointer := "village.json#/paths/" + pointerEscape(fixture.Path) + "/" + fixture.Method + "/responses/" + fixture.Status + "/content/application~1json/schema"
		unionSchema, err := compiler.Compile(pointer)
		if err != nil {
			t.Fatal(err)
		}
		armSchemas := make([]interface{ Validate(any) error }, len(declared))
		for i := range declared {
			compiled, err := compiler.Compile(pointer + "/oneOf/" + strconv.Itoa(i))
			if err != nil {
				t.Fatal(err)
			}
			armSchemas[i] = compiled
		}
		for _, sample := range fixture.Samples {
			t.Run(sample.Name, func(t *testing.T) {
				encoded, err := json.Marshal(sample.Body)
				if err != nil {
					t.Fatal(err)
				}
				var instance any
				if err := json.Unmarshal(encoded, &instance); err != nil {
					t.Fatal(err)
				}
				if err := unionSchema.Validate(instance); err != nil {
					t.Fatalf("the response union refuses %s: %v", encoded, err)
				}
				var accepted []string
				for i, arm := range armSchemas {
					if arm.Validate(instance) == nil {
						accepted = append(accepted, declared[i])
					}
				}
				if len(accepted) != 1 || accepted[0] != sample.Arm {
					t.Fatalf("%s is accepted by arms %v, want exactly [%s]", encoded, accepted, sample.Arm)
				}
			})
		}
	}
}

func pointerEscape(path string) string {
	return url.PathEscape(strings.ReplaceAll(strings.ReplaceAll(path, "~", "~0"), "/", "~1"))
}

// TestBuildVillageAPISpec_DeprecatedAliasDescriptionsKeepAnchors holds the
// deprecated /api/v1/groups twin of each canonical operation to the same
// description anchors, so the alias cannot drift from the prose it copies.
func TestBuildVillageAPISpec_DeprecatedAliasDescriptionsKeepAnchors(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	document := currentVillageSpecDocument(t)
	paths, _ := document["paths"].(map[string]any)
	checked := 0
	for _, fixture := range fixtures.Operations {
		if fixture.Path != "/api/v1/collectives" && !strings.HasPrefix(fixture.Path, "/api/v1/collectives/") {
			continue
		}
		aliasPath := strings.Replace(fixture.Path, "/api/v1/collectives", "/api/v1/groups", 1)
		item, ok := paths[aliasPath].(map[string]any)
		if !ok {
			continue
		}
		operation, ok := item[fixture.Method].(map[string]any)
		if !ok {
			continue
		}
		checked++
		if got := operation["operationId"]; got != fixture.OperationID+"Deprecated" {
			t.Errorf("%s %s operationId = %v, want %sDeprecated", fixture.Method, aliasPath, got, fixture.OperationID)
		}
		description, _ := operation["description"].(string)
		for _, anchor := range fixture.DescriptionAnchors {
			if !strings.Contains(description, anchor) {
				t.Errorf("%s %s description is missing anchor %q: %q", fixture.Method, aliasPath, anchor, description)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no deprecated alias was checked; the test would be vacuous")
	}
}
