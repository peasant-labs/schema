package openapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
)

// --- helpers ---

// specJSON marshals an openapi31.Spec to JSON via MarshalJSON.
func specJSON(t *testing.T, spec interface{ MarshalJSON() ([]byte, error) }) string {
	t.Helper()
	data, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	return string(data)
}

// specYAML marshals an openapi31.Spec to YAML via MarshalYAML.
func specYAML(t *testing.T, spec interface{ MarshalYAML() ([]byte, error) }) string {
	t.Helper()
	data, err := spec.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}
	return string(data)
}

// ============================================================================
// BuildTypesSpec tests
// ============================================================================

// TestBuildTypesSpec_Title verifies info.title is "Peasant Types".
func TestBuildTypesSpec_Title(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}
	if spec == nil {
		t.Fatal("BuildTypesSpec() returned nil")
	}

	if spec.Info.Title != "Peasant Types" {
		t.Errorf("info.title = %q; want %q", spec.Info.Title, "Peasant Types")
	}
}

// TestBuildTypesSpec_Version verifies info.version is "0.1.0".
func TestBuildTypesSpec_Version(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	if spec.Info.Version != "0.1.0" {
		t.Errorf("info.version = %q; want %q", spec.Info.Version, "0.1.0")
	}
}

// TestBuildTypesSpec_OpenAPI31 verifies the spec is valid OpenAPI 3.1.
func TestBuildTypesSpec_OpenAPI31(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	if spec.Openapi != "3.1.0" {
		t.Errorf("spec.Openapi = %q; want %q", spec.Openapi, "3.1.0")
	}
}

// TestBuildTypesSpec_ZeroPaths verifies no paths are defined (pure type catalog).
func TestBuildTypesSpec_ZeroPaths(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// Paths block should be absent or empty.
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if paths, ok := raw["paths"]; ok {
		pathMap, isMap := paths.(map[string]interface{})
		if !isMap || len(pathMap) > 0 {
			t.Errorf("shared types spec should have zero paths; got: %v", paths)
		}
	}
}

// TestBuildTypesSpec_ProviderEnum verifies Provider enum values appear.
func TestBuildTypesSpec_ProviderEnum(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	for _, provider := range []string{"claude-code", "gemini-cli", "codex", "opencode"} {
		if !containsQuoted(jsonStr, provider) {
			t.Errorf("shared types spec missing Provider enum value %q", provider)
		}
	}
}

// TestBuildTypesSpec_SessionEntryComponent verifies SessionEntry is registered
// as a plain-named component (no "Schema" prefix).
func TestBuildTypesSpec_SessionEntryComponent(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "SessionEntry") {
		t.Errorf("shared types spec missing SessionEntry component key")
	}
}

// TestBuildTypesSpec_QualityMetricsComponent verifies QualityMetrics is registered
// as a plain-named component (no "Schema" prefix).
func TestBuildTypesSpec_QualityMetricsComponent(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "QualityMetrics") {
		t.Errorf("shared types spec missing QualityMetrics component key")
	}
}

// TestBuildTypesSpec_NoDefinitionRefs verifies that no broken "#/definitions/" $refs
// remain in the spec — all internal refs must use "#/components/schemas/" (OAS 3.1 valid).
func TestBuildTypesSpec_NoDefinitionRefs(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if strings.Contains(jsonStr, `"#/definitions/`) {
		t.Errorf("shared types spec contains broken '#/definitions/' $ref values; all refs must use '#/components/schemas/'")
	}
}

// TestBuildTypesSpec_ValidJSON verifies the output is valid JSON.
func TestBuildTypesSpec_ValidJSON(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	data, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Error("shared types spec MarshalJSON produced invalid JSON")
	}
}

// containsQuoted checks that s contains the value wrapped in JSON double-quotes.
// This avoids false positives from matching substrings in description text.
func containsQuoted(s, value string) bool {
	return strings.Contains(s, `"`+value+`"`)
}
