package openapi

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"regexp"
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
		t.Fatalf("decode the existing-operations fixture: %v", err)
	}
	if fixture.AgainstVersion == "" {
		t.Fatal("the existing-operations fixture names no against_version; name the frozen Village version the allowed changes are measured from")
	}

	frozen := decodeVillageDocumentShape(t, "the frozen Village document "+VillageAPIVersion, schema.VillageAPISpecJSON())
	spec, err := BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	builtJSON, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal the rebuilt Village document: %v", err)
	}
	built := decodeVillageDocumentShape(t, "the rebuilt Village document", builtJSON)

	frozenComponents := canonicalMembers(t, frozen.Components.Schemas)
	builtComponents := canonicalMembers(t, built.Components.Schemas)
	changedComponent := func(name string) bool {
		builtSchema, ok := builtComponents[name]
		return !ok || frozenComponents[name] != builtSchema
	}

	var operations []string
	for path, frozenItem := range frozen.Paths {
		builtItem, ok := built.Paths[path]
		if !ok {
			t.Errorf("the frozen document declares %s, which the rebuilt document no longer declares; existing routes stay declared", path)
			continue
		}
		frozenMembers := canonicalMembers(t, frozenItem)
		builtMembers := canonicalMembers(t, builtItem)
		for member, frozenJSON := range frozenMembers {
			if !villageOperationMethods[member] {
				if builtMembers[member] != frozenJSON {
					t.Errorf("path %s member %s differs from the frozen document; existing path items keep their members", path, member)
				}
				continue
			}
			var operation struct {
				OperationID string `json:"operationId"`
			}
			if err := json.Unmarshal(frozenItem[member], &operation); err != nil {
				t.Fatalf("decode the frozen %s %s operation: %v", member, path, err)
			}
			builtJSON, ok := builtMembers[member]
			if !ok {
				t.Errorf("the frozen document declares %s %s (%s), which the rebuilt document no longer declares", member, path, operation.OperationID)
				continue
			}
			changed := frozenJSON != builtJSON
			for name := range referencedComponents(frozenJSON, frozenComponents) {
				changed = changed || changedComponent(name)
			}
			for name := range referencedComponents(builtJSON, builtComponents) {
				_, existed := frozenComponents[name]
				changed = changed || (existed && changedComponent(name))
			}
			if changed {
				operations = append(operations, operation.OperationID)
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
		t.Errorf("existing Village operations that differ from the frozen %s document = %v, want %v; an existing operation changes only deliberately: list it under allowed_changes in the existing-operations fixture with against_version %s, or restore its declaration", VillageAPIVersion, operations, wantOperations, VillageAPIVersion)
	}
	if strings.Join(components, ",") != strings.Join(wantComponents, ",") {
		t.Errorf("existing Village components that differ from the frozen %s document = %v, want %v; an existing component changes only deliberately: list it under allowed_changes in the existing-operations fixture with against_version %s, or restore its declaration", VillageAPIVersion, components, wantComponents, VillageAPIVersion)
	}
}

// villageOperationMethods are the path item members that are operations.
var villageOperationMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// villageDocumentShape is the part of a Village document this comparison
// reads: each path item's members and each component schema, as raw JSON.
type villageDocumentShape struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]json.RawMessage `json:"schemas"`
	} `json:"components"`
}

func decodeVillageDocumentShape(t *testing.T, label string, raw []byte) villageDocumentShape {
	t.Helper()
	var document villageDocumentShape
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode %s: %v", label, err)
	}
	if len(document.Paths) == 0 || len(document.Components.Schemas) == 0 {
		t.Fatalf("%s declares no paths or no component schemas; the comparison would be vacuous", label)
	}
	return document
}

// canonicalMembers returns each raw JSON member as canonical JSON, with
// object keys sorted, so two equal values give equal text.
func canonicalMembers(t *testing.T, members map[string]json.RawMessage) map[string]string {
	t.Helper()
	canonical := make(map[string]string, len(members))
	for name, raw := range members {
		var generic any
		if err := json.Unmarshal(raw, &generic); err != nil {
			t.Fatalf("decode member %s for comparison: %v", name, err)
		}
		text, err := json.Marshal(generic)
		if err != nil {
			t.Fatalf("re-marshal member %s for comparison: %v", name, err)
		}
		canonical[name] = string(text)
	}
	return canonical
}

// componentReference matches one component reference in canonical JSON. A
// reference written inside a string value would have escaped quotes, so only
// real $ref members match.
var componentReference = regexp.MustCompile(`[{,]"\$ref":"#/components/schemas/([^"]+)"`)

// referencedComponents returns every component the canonical JSON references,
// directly or through the components it references.
func referencedComponents(canonical string, components map[string]string) map[string]bool {
	seen := map[string]bool{}
	pending := []string{canonical}
	for len(pending) > 0 {
		next := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, match := range componentReference.FindAllStringSubmatch(next, -1) {
			name := match[1]
			if seen[name] {
				continue
			}
			seen[name] = true
			pending = append(pending, components[name])
		}
	}
	return seen
}
