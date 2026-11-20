package openapi

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_existing_operations.yaml
var villageExistingOperationsYAML []byte

type villageExistingOperationsFixture struct {
	AgainstVersion string `yaml:"against_version"`
	AllowedChanges struct {
		Operations []string `yaml:"operations"`
		Components []string `yaml:"components"`
	} `yaml:"allowed_changes"`
}

var villageOperationMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// TestVillageExistingOperationsAreUnchanged rebuilds the Village document and
// compares every operation of the last frozen document with its rebuilt
// declaration, including every component the operation references. Declaring
// new routes must not move an existing one; the only operations and
// components allowed to differ are the ones the fixture lists, and while the
// fixture's version is current they must be exactly those.
func TestVillageExistingOperationsAreUnchanged(t *testing.T) {
	var fixture villageExistingOperationsFixture
	decoder := yaml.NewDecoder(bytes.NewReader(villageExistingOperationsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode testdata/village_existing_operations.yaml: %v", err)
	}
	if fixture.AgainstVersion == "" {
		t.Fatal("testdata/village_existing_operations.yaml names no against_version; name the frozen Village version the allowed changes are measured from")
	}

	var frozen map[string]any
	if err := json.Unmarshal(schema.VillageAPISpecJSON(), &frozen); err != nil {
		t.Fatalf("decode the frozen Village document %s: %v", VillageAPIVersion, err)
	}
	spec, err := BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	built := villageDocument(t, spec)

	frozenComponents := villageComponentSchemas(frozen)
	builtComponents := villageComponentSchemas(built)
	changedComponent := func(name string) bool {
		builtSchema, ok := builtComponents[name]
		return !ok || !reflect.DeepEqual(frozenComponents[name], builtSchema)
	}

	var operations []string
	frozenPaths, _ := frozen["paths"].(map[string]any)
	builtPaths, _ := built["paths"].(map[string]any)
	for path, rawItem := range frozenPaths {
		frozenItem, _ := rawItem.(map[string]any)
		builtItem, ok := builtPaths[path].(map[string]any)
		if !ok {
			t.Errorf("the frozen document declares %s, which the rebuilt document no longer declares; existing routes stay declared", path)
			continue
		}
		for key, frozenValue := range frozenItem {
			if !villageOperationMethods[key] {
				if !reflect.DeepEqual(frozenValue, builtItem[key]) {
					t.Errorf("path %s member %s differs from the frozen document; existing path items keep their members", path, key)
				}
				continue
			}
			frozenOperation, _ := frozenValue.(map[string]any)
			id, _ := frozenOperation["operationId"].(string)
			builtOperation, ok := builtItem[key].(map[string]any)
			if !ok {
				t.Errorf("the frozen document declares %s %s (%s), which the rebuilt document no longer declares", key, path, id)
				continue
			}
			changed := !reflect.DeepEqual(frozenOperation, builtOperation)
			for name := range villageReferencedComponents(frozenOperation, frozenComponents) {
				changed = changed || changedComponent(name)
			}
			for name := range villageReferencedComponents(builtOperation, builtComponents) {
				_, existed := frozenComponents[name]
				changed = changed || (existed && changedComponent(name))
			}
			if changed {
				operations = append(operations, id)
			}
		}
	}
	var components []string
	for name := range frozenComponents {
		if changedComponent(name) {
			components = append(components, name)
		}
	}
	sort.Strings(operations)
	sort.Strings(components)

	wantOperations := []string{}
	wantComponents := []string{}
	if fixture.AgainstVersion == VillageAPIVersion {
		wantOperations = append(wantOperations, fixture.AllowedChanges.Operations...)
		wantComponents = append(wantComponents, fixture.AllowedChanges.Components...)
	}
	sort.Strings(wantOperations)
	sort.Strings(wantComponents)
	if strings.Join(operations, ",") != strings.Join(wantOperations, ",") {
		t.Errorf("existing Village operations that differ from the frozen %s document = %v, want %v; an existing operation changes only deliberately: list it under allowed_changes in testdata/village_existing_operations.yaml with against_version %s, or restore its declaration", VillageAPIVersion, operations, wantOperations, VillageAPIVersion)
	}
	if strings.Join(components, ",") != strings.Join(wantComponents, ",") {
		t.Errorf("existing Village components that differ from the frozen %s document = %v, want %v; an existing component changes only deliberately: list it under allowed_changes in testdata/village_existing_operations.yaml with against_version %s, or restore its declaration", VillageAPIVersion, components, wantComponents, VillageAPIVersion)
	}
}

func villageComponentSchemas(document map[string]any) map[string]any {
	components, _ := document["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	return schemas
}

// villageReferencedComponents returns every component schema value
// references, directly or through the components it references.
func villageReferencedComponents(value any, components map[string]any) map[string]bool {
	seen := map[string]bool{}
	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			if ref, ok := typed["$ref"].(string); ok {
				name := strings.TrimPrefix(ref, "#/components/schemas/")
				if name != ref && !seen[name] {
					seen[name] = true
					walk(components[name])
				}
			}
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return seen
}
