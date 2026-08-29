package openapi_test

import (
	"embed"
	"encoding/json"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_collectives_operations.yaml
var villageCollectivesFixtures embed.FS

type villageCollectivesOperationFixtures struct {
	Operations     []villageCollectivesOperationFixture `yaml:"operations"`
	ComponentEnums []villageComponentEnumFixture        `yaml:"component_enums"`
}

type villageCollectivesOperationFixture struct {
	Path               string   `yaml:"path"`
	Method             string   `yaml:"method"`
	OperationID        string   `yaml:"operation_id"`
	RequestComponents  []string `yaml:"request_components"`
	ResponseComponents []string `yaml:"response_components"`
	Statuses           []string `yaml:"statuses"`
	DescriptionAnchors []string `yaml:"description_anchors"`
}

type villageComponentEnumFixture struct {
	Component string   `yaml:"component"`
	Values    []string `yaml:"values"`
}

func TestBuildVillageAPISpec_CollectivesOperations(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	document := currentVillageSpecDocument(t)

	seen := map[string]struct{}{}
	for _, fixture := range fixtures.Operations {
		if fixture.Path == "" || fixture.Method == "" || fixture.OperationID == "" {
			t.Fatalf("operation fixture has a blank key field: %+v", fixture)
		}
		key := fixture.Method + " " + fixture.Path
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("operation fixture repeats %s", key)
		}
		seen[key] = struct{}{}

		operation := villageOperation(t, document, fixture.Path, fixture.Method)
		if got := operation["operationId"]; got != fixture.OperationID {
			t.Fatalf("%s operationId = %v, want %s", key, got, fixture.OperationID)
		}
		responses, ok := operation["responses"].(map[string]any)
		if !ok {
			t.Fatalf("%s declares no responses object", key)
		}
		assertSameStringSet(t, key+" response statuses", keys(responses), fixture.Statuses)

		operationJSON := mustJSON(t, operation)
		for _, component := range fixture.RequestComponents {
			if !strings.Contains(operationJSON, component) {
				t.Fatalf("%s request does not reference component %s: %s", key, component, operationJSON)
			}
		}
		for _, component := range fixture.ResponseComponents {
			if !strings.Contains(operationJSON, component) {
				t.Fatalf("%s response does not reference component %s: %s", key, component, operationJSON)
			}
		}
		description, ok := operation["description"].(string)
		if !ok || description == "" {
			t.Fatalf("%s declares no operation description", key)
		}
		for _, anchor := range fixture.DescriptionAnchors {
			if !strings.Contains(description, anchor) {
				t.Fatalf("%s description is missing anchor %q: %q", key, anchor, description)
			}
		}
	}
}

func TestBuildVillageAPISpec_CollectivesEnums(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	document := currentVillageSpecDocument(t)

	for _, fixture := range fixtures.ComponentEnums {
		if fixture.Component == "" {
			t.Fatalf("enum fixture has a blank component: %+v", fixture)
		}
		component := villageComponent(t, document, fixture.Component)
		raw, ok := component["enum"].([]any)
		if !ok {
			t.Fatalf("component %s declares no enum", fixture.Component)
		}
		values := make([]string, 0, len(raw))
		for _, value := range raw {
			asString, ok := value.(string)
			if !ok {
				t.Fatalf("component %s has non-string enum value %v", fixture.Component, value)
			}
			values = append(values, asString)
		}
		assertSameStringSet(t, fixture.Component+" enum values", values, fixture.Values)
	}
}

func loadVillageCollectivesFixtures(t *testing.T) villageCollectivesOperationFixtures {
	t.Helper()
	data, err := villageCollectivesFixtures.ReadFile("testdata/village_collectives_operations.yaml")
	if err != nil {
		t.Fatalf("read Village collectives operation fixtures: %v", err)
	}
	var fixtures villageCollectivesOperationFixtures
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("decode Village collectives operation fixtures: %v", err)
	}
	if len(fixtures.Operations) == 0 {
		t.Fatal("Village collectives operation fixtures must name at least one operation")
	}
	if len(fixtures.ComponentEnums) == 0 {
		t.Fatal("Village collectives operation fixtures must name at least one enum")
	}
	return fixtures
}

func currentVillageSpecDocument(t *testing.T) map[string]any {
	t.Helper()
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	raw, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal Village API spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode Village API spec: %v", err)
	}
	return document
}

func villageOperation(t *testing.T, document map[string]any, path, method string) map[string]any {
	t.Helper()
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec has no paths")
	}
	pathItem, ok := paths[path].(map[string]any)
	if !ok {
		t.Fatalf("Village API spec is missing %s", path)
	}
	operation, ok := pathItem[method].(map[string]any)
	if !ok {
		t.Fatalf("Village API spec is missing %s %s", method, path)
	}
	return operation
}
