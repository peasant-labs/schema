package openapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
)

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
// single-source PeasantLocalAPIVersion const (bumped to 0.2.0 for the
// Map/Review/Search surface), so the assertion is not stranded on a stale
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
