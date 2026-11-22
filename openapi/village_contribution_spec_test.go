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
			if name, ok := v["name"].(string); ok { // unstring-allow: go-json-key-assert — this walk visits every node of the whole document by design
				found["parameter"][name] = true
			}
		}
		if properties, ok := v["properties"].(map[string]any); ok { // unstring-allow: go-json-key-assert — this walk visits every node of the whole document by design
			for name := range properties {
				found["property"][name] = true
			}
		}
		if values, ok := v["enum"].([]any); ok { // unstring-allow: go-json-key-assert — this walk visits every node of the whole document by design
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
		operation := decodeVillageSpecOperation(t, raw, fixture.Path, fixture.Method)
		arms := operation.Responses[fixture.Status].Content[villageMediaTypeJSON].Schema.OneOf
		declared := make([]string, 0, len(arms))
		for _, arm := range arms {
			declared = append(declared, strings.TrimPrefix(arm.Ref, "#/components/schemas/"))
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
// deprecated twin of each canonical operation, the one whose operationId adds
// Deprecated, to the same description anchors, so the alias cannot drift from
// the prose it copies.
func TestBuildVillageAPISpec_DeprecatedAliasDescriptionsKeepAnchors(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	raw := currentVillageSpecJSON(t)
	byID := villageSpecOperationsByID(t, raw)
	checked := 0
	for _, fixture := range fixtures.Operations {
		alias, ok := byID[fixture.OperationID+"Deprecated"]
		if !ok {
			continue
		}
		checked++
		for _, anchor := range fixture.DescriptionAnchors {
			if !strings.Contains(alias.Description, anchor) {
				t.Errorf("%s description is missing anchor %q: %q", alias.OperationID, anchor, alias.Description)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no deprecated alias was checked; the test would be vacuous")
	}
}

// villageMediaTypeJSON is the media type of every Village JSON body.
const villageMediaTypeJSON = "application/json"

// villageSpecOperation is the part of one operation these tests read.
type villageSpecOperation struct {
	OperationID string `json:"operationId"`
	Description string `json:"description"`
	Responses   map[string]struct {
		Content map[string]struct {
			Schema struct {
				OneOf []struct {
					Ref string `json:"$ref"`
				} `json:"oneOf"`
			} `json:"schema"`
		} `json:"content"`
	} `json:"responses"`
}

// villageSpecPaths keeps each path item's members raw, because a path item
// also carries members that are not operations, such as parameters.
type villageSpecPaths struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

func currentVillageSpecJSON(t *testing.T) []byte {
	t.Helper()
	raw, err := json.Marshal(currentVillageSpecDocument(t))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func decodeVillageSpecPaths(t *testing.T, raw []byte) villageSpecPaths {
	t.Helper()
	var document villageSpecPaths
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode Village API spec paths: %v", err)
	}
	if len(document.Paths) == 0 {
		t.Fatal("Village API spec has no paths")
	}
	return document
}

func decodeVillageSpecOperation(t *testing.T, raw []byte, path, method string) villageSpecOperation {
	t.Helper()
	member, ok := decodeVillageSpecPaths(t, raw).Paths[path][method]
	if !ok {
		t.Fatalf("Village API spec is missing %s %s", method, path)
	}
	var operation villageSpecOperation
	if err := json.Unmarshal(member, &operation); err != nil {
		t.Fatalf("decode %s %s: %v", method, path, err)
	}
	return operation
}

// villageSpecOperationsByID returns every operation of the document by its
// operationId. A path item member that is not an operation has none and is
// skipped.
func villageSpecOperationsByID(t *testing.T, raw []byte) map[string]villageSpecOperation {
	t.Helper()
	byID := map[string]villageSpecOperation{}
	for path, item := range decodeVillageSpecPaths(t, raw).Paths {
		for method, member := range item {
			var operation villageSpecOperation
			if json.Unmarshal(member, &operation) != nil || operation.OperationID == "" {
				continue
			}
			if _, dup := byID[operation.OperationID]; dup {
				t.Fatalf("operationId %s is declared twice, again at %s %s", operation.OperationID, method, path)
			}
			byID[operation.OperationID] = operation
		}
	}
	return byID
}
