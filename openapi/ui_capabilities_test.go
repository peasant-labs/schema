package openapi_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/ui_capabilities_contract.yaml
var uiCapabilitiesContractFixture []byte

type uiCapabilitiesFixture struct {
	Contract struct {
		Path           string `yaml:"path"`
		Method         string `yaml:"method"`
		OperationID    string `yaml:"operation_id"`
		ResponseStatus string `yaml:"response_status"`
		MediaType      string `yaml:"media_type"`
		ResponseRef    string `yaml:"response_ref"`
		Component      string `yaml:"component"`
		Property       string `yaml:"property"`
	} `yaml:"contract"`
	Cases []struct {
		Name  string `yaml:"name"`
		Body  string `yaml:"body"`
		Valid bool   `yaml:"valid"`
	} `yaml:"cases"`
}

func TestUICapabilitiesContract(t *testing.T) {
	var fixture uiCapabilitiesFixture
	decoder := yaml.NewDecoder(bytes.NewReader(uiCapabilitiesContractFixture))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) < 6 {
		t.Fatalf("UI capabilities fixture has %d cases, want at least 6", len(fixture.Cases))
	}

	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	document := specDocument(t, spec)
	operation, err := requiredMap(document, "paths", "OpenAPI document")
	if err != nil {
		t.Fatal(err)
	}
	operation, err = requiredMap(operation, fixture.Contract.Path, "paths")
	if err != nil {
		t.Fatal(err)
	}
	operation, err = requiredMap(operation, fixture.Contract.Method, fixture.Contract.Path)
	if err != nil {
		t.Fatal(err)
	}
	if operation["operationId"] != fixture.Contract.OperationID {
		t.Fatalf("operationId=%v, want %q", operation["operationId"], fixture.Contract.OperationID)
	}
	responses, err := requiredMap(operation, "responses", fixture.Contract.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	response, err := requiredMap(responses, fixture.Contract.ResponseStatus, "responses")
	if err != nil {
		t.Fatal(err)
	}
	content, err := requiredMap(response, "content", "response")
	if err != nil {
		t.Fatal(err)
	}
	media, err := requiredMap(content, fixture.Contract.MediaType, "content")
	if err != nil {
		t.Fatal(err)
	}
	responseSchema, err := requiredMap(media, "schema", "media")
	if err != nil {
		t.Fatal(err)
	}
	if responseSchema["$ref"] != fixture.Contract.ResponseRef {
		t.Fatalf("response ref=%v, want %q", responseSchema["$ref"], fixture.Contract.ResponseRef)
	}

	components, err := requiredMap(document, "components", "OpenAPI document")
	if err != nil {
		t.Fatal(err)
	}
	components, err = requiredMap(components, "schemas", "components")
	if err != nil {
		t.Fatal(err)
	}
	component, err := requiredMap(components, fixture.Contract.Component, "components.schemas")
	if err != nil {
		t.Fatal(err)
	}
	properties, err := requiredMap(component, "properties", fixture.Contract.Component)
	if err != nil {
		t.Fatal(err)
	}
	property, err := requiredMap(properties, fixture.Contract.Property, fixture.Contract.Component+".properties")
	if err != nil {
		t.Fatal(err)
	}
	items, err := requiredMap(property, "items", fixture.Contract.Property)
	if err != nil {
		t.Fatal(err)
	}
	if items["type"] != "string" {
		t.Fatalf("item type=%v, want string", items["type"])
	}
	if _, closed := items["enum"]; closed {
		t.Fatal("UI capability items must remain open strings, not an enum")
	}

	raw, err := json.Marshal(component)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("ui-capabilities.json", bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile("ui-capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixture.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			var body any
			if err := json.Unmarshal([]byte(tc.Body), &body); err != nil {
				t.Fatal(err)
			}
			valid := compiled.Validate(body) == nil
			if valid != tc.Valid {
				t.Fatalf("schema accepted=%v, want %v for %s", valid, tc.Valid, fmt.Sprintf("%q", tc.Body))
			}
		})
	}
}
