package openapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type projectResolverContractFixture struct {
	Contract struct {
		Path                 string   `yaml:"path"`
		Method               string   `yaml:"method"`
		OperationID          string   `yaml:"operation_id"`
		QueryName            string   `yaml:"query_name"`
		QueryType            string   `yaml:"query_type"`
		ResponseStatus       string   `yaml:"response_status"`
		MediaType            string   `yaml:"media_type"`
		ResponseRef          string   `yaml:"response_ref"`
		ResponseComponent    string   `yaml:"response_component"`
		RequiredProperties   []string `yaml:"required_properties"`
		ProjectProperty      string   `yaml:"project_property"`
		ProjectType          string   `yaml:"project_type"`
		ProjectHashProperty  string   `yaml:"project_hash_property"`
		ProjectHashRef       string   `yaml:"project_hash_ref"`
		ProjectHashComponent string   `yaml:"project_hash_component"`
		ProjectHashPattern   string   `yaml:"project_hash_pattern"`
		ProjectHashMinLength int      `yaml:"project_hash_min_length"`
		ProjectHashMaxLength int      `yaml:"project_hash_max_length"`
		ExampleProjectHash   string   `yaml:"example_project_hash"`
	} `yaml:"contract"`
	Mutations testcase.Corpus[projectResolverMutation, bool] `yaml:"mutations"`
}

type projectResolverMutation struct {
	Kind   string `yaml:"kind"`
	Target string `yaml:"target"`
}

// ============================================================================
// BuildPeasantLocalAPISpec tests (V14, V15a, V15b)
// ============================================================================

// TestBuildPeasantLocalAPISpec_OpenAPI31 verifies the local spec is valid OpenAPI 3.1 (V14).
func TestBuildPeasantLocalAPISpec_OpenAPI31(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}
	if spec == nil {
		t.Fatal("BuildPeasantLocalAPISpec() returned nil")
	}

	if spec.Openapi != "3.1.0" {
		t.Errorf("spec.Openapi = %q; want %q", spec.Openapi, "3.1.0")
	}
}

// TestBuildPeasantLocalAPISpec_Version verifies info.version tracks the
// single-source PeasantLocalAPIVersion const, so the assertion is not stranded on a stale
// literal at the next bump.
func TestBuildPeasantLocalAPISpec_Version(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	if spec.Info.Version != specpkg.PeasantLocalAPIVersion {
		t.Errorf("info.version = %q; want %q", spec.Info.Version, specpkg.PeasantLocalAPIVersion)
	}
}

// TestBuildPeasantLocalAPISpec_RESTRoutes verifies REST paths are present (V15a).
func TestBuildPeasantLocalAPISpec_RESTRoutes(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	yamlStr := specYAML(t, spec)

	routes := []string{
		"/api/v1/health",
		"/api/v1/sessions",
		"/api/v1/sessions/{id}",
		"/api/v1/config/mock",
		"/api/v1/shutdown",
	}
	for _, route := range routes {
		if !strings.Contains(yamlStr, route) {
			t.Errorf("local spec missing route %q", route)
		}
	}
}

func TestBuildPeasantLocalAPISpec_ProjectResolverSurface(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}
	fixture := loadProjectResolverContractFixture(t)
	document := specDocument(t, spec)
	if err := validateProjectResolverContract(document, fixture); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range fixture.Mutations.Cases {
		t.Run(mutation.Name, func(t *testing.T) {
			mutated := cloneSpecDocument(t, document)
			applyProjectResolverMutation(t, mutated, fixture, mutation.Input)
			accepted := validateProjectResolverContract(mutated, fixture) == nil
			if accepted != mutation.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, mutation.Expected)
			}
		})
	}
}

func loadProjectResolverContractFixture(t *testing.T) projectResolverContractFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/project_resolver_contract.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var fixture projectResolverContractFixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Mutations.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Mutations.CheckMin(8); err != nil {
		t.Fatal(err)
	}
	projectHash, err := schema.NewProjectHash(fixture.Contract.ExampleProjectHash)
	if err != nil {
		t.Fatalf("resolver contract example_project_hash is not canonical: %v", err)
	}
	if err := projectHash.Validate(); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func specDocument(t *testing.T, spec any) map[string]any {
	t.Helper()
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func cloneSpecDocument(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func validateProjectResolverContract(document map[string]any, fixture projectResolverContractFixture) error {
	contract := fixture.Contract
	paths, err := requiredMap(document, "paths", "OpenAPI document")
	if err != nil {
		return err
	}
	path, err := requiredMap(paths, contract.Path, "paths")
	if err != nil {
		return err
	}
	operation, err := requiredMap(path, contract.Method, contract.Path)
	if err != nil {
		return err
	}
	if operation["operationId"] != contract.OperationID {
		return fmt.Errorf("project resolver contract: %s %s operationId=%v, want %q; restore the canonical resolver operation", contract.Method, contract.Path, operation["operationId"], contract.OperationID)
	}
	parameters, ok := operation["parameters"].([]any)
	if !ok {
		return fmt.Errorf("project resolver contract: %s %s parameters are missing or not an array; restore the required name query", contract.Method, contract.Path)
	}
	foundQuery := false
	for _, candidate := range parameters {
		parameter, ok := candidate.(map[string]any)
		if ok && parameter["in"] == "query" && parameter["name"] == contract.QueryName {
			if parameter["required"] != true {
				return fmt.Errorf("project resolver contract: query parameter %q is optional; saved routes must supply an exact display identity", contract.QueryName)
			}
			querySchema, err := requiredMap(parameter, "schema", "resolver query "+contract.QueryName)
			if err != nil {
				return err
			}
			if querySchema["type"] != contract.QueryType {
				return fmt.Errorf("project resolver contract: query parameter %q schema.type=%v, want %q; restore the string display-name contract", contract.QueryName, querySchema["type"], contract.QueryType)
			}
			foundQuery = true
		}
	}
	if !foundQuery {
		return fmt.Errorf("project resolver contract: required query parameter %q is absent; saved routes must supply an exact display identity", contract.QueryName)
	}
	responses, err := requiredMap(operation, "responses", contract.OperationID)
	if err != nil {
		return err
	}
	response, err := requiredMap(responses, contract.ResponseStatus, "responses")
	if err != nil {
		return err
	}
	content, err := requiredMap(response, "content", "resolver response")
	if err != nil {
		return err
	}
	media, err := requiredMap(content, contract.MediaType, "resolver response content")
	if err != nil {
		return err
	}
	schemaValue, err := requiredMap(media, "schema", "resolver response media")
	if err != nil {
		return err
	}
	if schemaValue["$ref"] != contract.ResponseRef {
		return fmt.Errorf("project resolver contract: 200 JSON response ref=%v, want %q; restore the resolver payload", schemaValue["$ref"], contract.ResponseRef)
	}
	components, err := requiredMap(document, "components", "OpenAPI document")
	if err != nil {
		return err
	}
	schemas, err := requiredMap(components, "schemas", "components")
	if err != nil {
		return err
	}
	resolver, err := requiredMap(schemas, contract.ResponseComponent, "components.schemas")
	if err != nil {
		return err
	}
	required, ok := resolver["required"].([]any)
	if !ok || len(required) != len(contract.RequiredProperties) {
		return fmt.Errorf("project resolver contract: %s.required=%v, want exact properties %v", contract.ResponseComponent, resolver["required"], contract.RequiredProperties)
	}
	for index, name := range contract.RequiredProperties {
		if required[index] != name {
			return fmt.Errorf("project resolver contract: %s.required[%d]=%v, want %q", contract.ResponseComponent, index, required[index], name)
		}
	}
	properties, err := requiredMap(resolver, "properties", contract.ResponseComponent)
	if err != nil {
		return err
	}
	projectProperty, err := requiredMap(properties, contract.ProjectProperty, contract.ResponseComponent+".properties")
	if err != nil {
		return err
	}
	if projectProperty["type"] != contract.ProjectType {
		return fmt.Errorf("project resolver contract: response property %q type=%v, want %q; restore the string project display identity", contract.ProjectProperty, projectProperty["type"], contract.ProjectType)
	}
	hashProperty, err := requiredMap(properties, contract.ProjectHashProperty, contract.ResponseComponent+".properties")
	if err != nil {
		return err
	}
	if hashProperty["$ref"] != contract.ProjectHashRef {
		return fmt.Errorf("project resolver contract: projectHash ref=%v, want %q", hashProperty["$ref"], contract.ProjectHashRef)
	}
	hashSchema, err := requiredMap(schemas, contract.ProjectHashComponent, "components.schemas")
	if err != nil {
		return err
	}
	if hashSchema["pattern"] != contract.ProjectHashPattern || hashSchema["minLength"] != float64(contract.ProjectHashMinLength) || hashSchema["maxLength"] != float64(contract.ProjectHashMaxLength) {
		return fmt.Errorf("project resolver contract: project hash constraints pattern=%v minLength=%v maxLength=%v, want %q/%d/%d", hashSchema["pattern"], hashSchema["minLength"], hashSchema["maxLength"], contract.ProjectHashPattern, contract.ProjectHashMinLength, contract.ProjectHashMaxLength)
	}
	return nil
}

func requiredMap(parent map[string]any, key, location string) (map[string]any, error) {
	value, ok := parent[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("project resolver contract: %s.%s is missing or not an object", location, key)
	}
	return value, nil
}

func applyProjectResolverMutation(t *testing.T, document map[string]any, fixture projectResolverContractFixture, mutation projectResolverMutation) {
	t.Helper()
	paths, _ := requiredMap(document, "paths", "document")
	path, _ := requiredMap(paths, fixture.Contract.Path, "paths")
	operation, _ := requiredMap(path, fixture.Contract.Method, fixture.Contract.Path)
	components, _ := requiredMap(document, "components", "document")
	schemas, _ := requiredMap(components, "schemas", "components")
	resolver, _ := requiredMap(schemas, fixture.Contract.ResponseComponent, "schemas")
	switch mutation.Kind {
	case "remove-operation":
		delete(path, fixture.Contract.Method)
	case "redirect-operation-id":
		operation["operationId"] = mutation.Target
	case "make-query-optional":
		for _, candidate := range operation["parameters"].([]any) {
			parameter := candidate.(map[string]any)
			if parameter["name"] == fixture.Contract.QueryName {
				parameter["required"] = false
			}
		}
	case "change-query-type":
		for _, candidate := range operation["parameters"].([]any) {
			parameter := candidate.(map[string]any)
			if parameter["name"] == fixture.Contract.QueryName {
				querySchema, _ := requiredMap(parameter, "schema", "query")
				querySchema["type"] = mutation.Target
			}
		}
	case "redirect-response":
		responses, _ := requiredMap(operation, "responses", "operation")
		response, _ := requiredMap(responses, fixture.Contract.ResponseStatus, "responses")
		content, _ := requiredMap(response, "content", "response")
		media, _ := requiredMap(content, fixture.Contract.MediaType, "content")
		schemaValue, _ := requiredMap(media, "schema", "media")
		schemaValue["$ref"] = mutation.Target
	case "remove-required-property":
		required := resolver["required"].([]any)
		filtered := make([]any, 0, len(required))
		for _, value := range required {
			if value != mutation.Target {
				filtered = append(filtered, value)
			}
		}
		resolver["required"] = filtered
	case "change-project-type":
		properties, _ := requiredMap(resolver, "properties", "resolver")
		projectProperty, _ := requiredMap(properties, fixture.Contract.ProjectProperty, "properties")
		projectProperty["type"] = mutation.Target
	case "change-hash-pattern":
		hashSchema, _ := requiredMap(schemas, fixture.Contract.ProjectHashComponent, "schemas")
		hashSchema["pattern"] = mutation.Target
	default:
		t.Fatalf("unknown project resolver mutation kind %q", mutation.Kind)
	}
}

// TestBuildPeasantLocalAPISpec_WSChannelSchemas verifies WS message schemas exist as components (V15a).
func TestBuildPeasantLocalAPISpec_WSChannelSchemas(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// Each WS payload type should appear as a component schema.
	payloads := []string{
		"DashboardPayload",
		"SessionsPayload",
		"SessionDetailPayload",
		"TrendsPayload",
		"QualityPayload",
	}
	for _, p := range payloads {
		if !strings.Contains(jsonStr, p) {
			t.Errorf("local spec missing WS message schema component %q", p)
		}
	}
}

// TestBuildPeasantLocalAPISpec_HealthResponse verifies health endpoint response schema (V15b).
func TestBuildPeasantLocalAPISpec_HealthResponse(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !strings.Contains(jsonStr, "HealthResponse") {
		t.Errorf("local spec missing HealthResponse component")
	}
}

// TestBuildPeasantLocalAPISpec_SessionSummary verifies SessionSummary appears (V15b).
func TestBuildPeasantLocalAPISpec_SessionSummary(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !strings.Contains(jsonStr, "SessionSummary") {
		t.Errorf("local spec missing SessionSummary component")
	}
}

// TestBuildPeasantLocalAPISpec_ValidJSON verifies the output is valid JSON.
func TestBuildPeasantLocalAPISpec_ValidJSON(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	data, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Error("local spec MarshalJSON produced invalid JSON")
	}
}

// TestBuildPeasantLocalAPISpec_MockConfigResponse verifies MockConfigResponse is in local spec (V15b).
func TestBuildPeasantLocalAPISpec_MockConfigResponse(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !strings.Contains(jsonStr, "MockConfigResponse") {
		t.Errorf("local spec missing MockConfigResponse component")
	}
}

// ============================================================================
// Annotation endpoint tests (S9)
// ============================================================================

// TestBuildPeasantLocalAPISpec_AnnotationRoutes verifies annotation REST paths are registered.
func TestBuildPeasantLocalAPISpec_AnnotationRoutes(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	yamlStr := specYAML(t, spec)

	routes := []string{
		"/api/v1/annotations",
		"/api/v1/annotation-types",
	}
	for _, route := range routes {
		if !strings.Contains(yamlStr, route) {
			t.Errorf("local spec missing annotation route %q", route)
		}
	}

	// Verify session_id query parameter is present for GET /annotations.
	jsonStr := specJSON(t, spec)
	if !strings.Contains(jsonStr, "session_id") {
		t.Error("GET /api/v1/annotations spec missing session_id query parameter")
	}
}

// TestBuildPeasantLocalAPISpec_AnnotationOperationIDs verifies annotation operation IDs are registered.
func TestBuildPeasantLocalAPISpec_AnnotationOperationIDs(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	opIDs := []string{
		"listAnnotations",
		"createAnnotation",
		"listAnnotationTypes",
	}
	for _, id := range opIDs {
		if !strings.Contains(jsonStr, id) {
			t.Errorf("local spec missing annotation operationId %q", id)
		}
	}
}

// TestBuildPeasantLocalAPISpec_AnnotationComponents verifies annotation schema components are registered.
func TestBuildPeasantLocalAPISpec_AnnotationComponents(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	components := []string{
		"AnnotationsPayload",
		"AnnotationSummary",
		"AnnotationTypeSummary",
		"CreateAnnotationRequest",
		"CreateAnnotationResponse",
	}
	for _, c := range components {
		if !strings.Contains(jsonStr, c) {
			t.Errorf("local spec missing annotation component %q", c)
		}
	}
}

// TestBuildPeasantLocalAPISpec_CreateAnnotation201 verifies POST /api/v1/annotations returns 201.
func TestBuildPeasantLocalAPISpec_CreateAnnotation201(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// The spec should contain "201" as a response status for the POST endpoint.
	if !strings.Contains(jsonStr, `"201"`) {
		t.Error("local spec missing 201 response status for POST /api/v1/annotations")
	}
}
