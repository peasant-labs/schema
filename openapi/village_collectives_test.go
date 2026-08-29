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
	Operations        []villageCollectivesOperationFixture `yaml:"operations"`
	ComponentEnums    []villageComponentEnumFixture        `yaml:"component_enums"`
	ComponentRequired []villageComponentRequiredFixture    `yaml:"component_required"`
	ComponentProps    []villageComponentPropertyFixture    `yaml:"component_properties"`
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

type villageComponentRequiredFixture struct {
	Component string   `yaml:"component"`
	Contains  []string `yaml:"contains"`
	Excludes  []string `yaml:"excludes"`
}

type villageComponentPropertyFixture struct {
	Component string `yaml:"component"`
	Property  string `yaml:"property"`
	Required  *bool  `yaml:"required"`
	Nullable  *bool  `yaml:"nullable"`
	ItemsRef  string `yaml:"items_ref"`
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

func TestBuildVillageAPISpec_CollectivesRequiredFields(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	document := currentVillageSpecDocument(t)

	for _, fixture := range fixtures.ComponentRequired {
		if fixture.Component == "" {
			t.Fatalf("required-field fixture has a blank component: %+v", fixture)
		}
		component := villageComponent(t, document, fixture.Component)
		required := villageRequiredSet(t, component, fixture.Component)
		for _, field := range fixture.Contains {
			if _, ok := required[field]; !ok {
				t.Errorf("component %s required set is missing %q", fixture.Component, field)
			}
		}
		for _, field := range fixture.Excludes {
			if _, ok := required[field]; ok {
				t.Errorf("component %s required set includes forbidden %q", fixture.Component, field)
			}
		}
	}
}

func TestBuildVillageAPISpec_CollectivesPropertyContracts(t *testing.T) {
	fixtures := loadVillageCollectivesFixtures(t)
	document := currentVillageSpecDocument(t)

	for _, fixture := range fixtures.ComponentProps {
		if fixture.Component == "" || fixture.Property == "" {
			t.Fatalf("property fixture has a blank key field: %+v", fixture)
		}
		component := villageComponent(t, document, fixture.Component)
		property := villageProperty(t, component, fixture.Component, fixture.Property)
		if fixture.Required != nil {
			required := villageRequiredSet(t, component, fixture.Component)
			_, got := required[fixture.Property]
			if got != *fixture.Required {
				t.Errorf("component %s property %s required = %v, want %v", fixture.Component, fixture.Property, got, *fixture.Required)
			}
		}
		if fixture.Nullable != nil {
			got := schemaAllowsNull(property)
			if got != *fixture.Nullable {
				t.Errorf("component %s property %s nullable = %v, want %v", fixture.Component, fixture.Property, got, *fixture.Nullable)
			}
		}
		if fixture.ItemsRef != "" {
			items, ok := property["items"].(map[string]any)
			if !ok {
				t.Fatalf("component %s property %s declares no array items schema: %s", fixture.Component, fixture.Property, mustJSON(t, property))
			}
			if got, ok := items["$ref"].(string); !ok || got != fixture.ItemsRef {
				t.Errorf("component %s property %s items $ref = %v, want %s", fixture.Component, fixture.Property, items["$ref"], fixture.ItemsRef)
			}
		}
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
	if len(fixtures.ComponentRequired) == 0 {
		t.Fatal("Village collectives operation fixtures must name at least one required-field component")
	}
	if len(fixtures.ComponentProps) == 0 {
		t.Fatal("Village collectives operation fixtures must name at least one property contract")
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

func villageRequiredSet(t *testing.T, component map[string]any, componentName string) map[string]struct{} {
	t.Helper()
	raw, ok := component["required"].([]any)
	if !ok {
		t.Fatalf("component %s declares no required set", componentName)
	}
	required := map[string]struct{}{}
	for _, value := range raw {
		asString, ok := value.(string)
		if !ok {
			t.Fatalf("component %s has non-string required value %v", componentName, value)
		}
		required[asString] = struct{}{}
	}
	return required
}

func villageProperty(t *testing.T, component map[string]any, componentName, propertyName string) map[string]any {
	t.Helper()
	properties, ok := component["properties"].(map[string]any)
	if !ok {
		t.Fatalf("component %s declares no properties object", componentName)
	}
	property, ok := properties[propertyName].(map[string]any)
	if !ok {
		t.Fatalf("component %s declares no %s property", componentName, propertyName)
	}
	return property
}
