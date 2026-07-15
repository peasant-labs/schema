package openapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
)

// ============================================================================
// Village API — operation-level examples
// ============================================================================

// TestBuildVillageAPISpec_HasRequestBodyExamples verifies that the publish
// endpoint request body contains named examples ("minimal" and "full").
func TestBuildVillageAPISpec_HasRequestBodyExamples(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	for _, key := range []string{"minimal", "full"} {
		if !containsQuoted(jsonStr, key) {
			t.Errorf("village spec missing request body example key %q", key)
		}
	}
}

// TestBuildVillageAPISpec_HasResponseExample verifies that the publish
// endpoint 200 response contains the "newTranscript" named example.
func TestBuildVillageAPISpec_HasResponseExample(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "newTranscript") {
		t.Errorf("village spec missing response example key %q", "newTranscript")
	}
}

// TestBuildVillageAPISpec_ExampleContainsSessionID verifies that example data
// includes a realistic session ID value.
func TestBuildVillageAPISpec_ExampleContainsSessionID(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !strings.Contains(jsonStr, "99d59925-36bc-424c-a789-8be54d9702ba") {
		t.Error("village spec examples missing sample session ID")
	}
}

// TestBuildVillageAPISpec_HasAuthEndpoints verifies that the village spec
// includes the CLI auth endpoints (login and exchange).
func TestBuildVillageAPISpec_HasAuthEndpoints(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	for _, path := range []string{"/api/v1/auth/cli/login", "/api/v1/auth/cli/exchange"} {
		if !strings.Contains(jsonStr, path) {
			t.Errorf("village spec missing auth endpoint path %q", path)
		}
	}

	for _, opID := range []string{"cliLogin", "cliExchangeCode"} {
		if !containsQuoted(jsonStr, opID) {
			t.Errorf("village spec missing auth operation ID %q", opID)
		}
	}
}

// TestBuildVillageAPISpec_HasExchangeExamples verifies that the exchange
// endpoint has request body and response examples.
func TestBuildVillageAPISpec_HasExchangeExamples(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	for _, key := range []string{"codeExchange", "credentials"} {
		if !containsQuoted(jsonStr, key) {
			t.Errorf("village spec missing exchange example key %q", key)
		}
	}
}

// ============================================================================
// Peasant Local API — operation-level examples
// ============================================================================

// TestBuildPeasantLocalAPISpec_HasHealthExample verifies the health endpoint has
// a response example.
func TestBuildPeasantLocalAPISpec_HasHealthExample(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "healthy") {
		t.Errorf("local spec missing health response example key %q", "healthy")
	}
}

// TestBuildPeasantLocalAPISpec_HasSessionsExample verifies the sessions list
// endpoint has a response example.
func TestBuildPeasantLocalAPISpec_HasSessionsExample(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "twoSessions") {
		t.Errorf("local spec missing sessions response example key %q", "twoSessions")
	}
}

// TestBuildPeasantLocalAPISpec_HasSessionDetailExample verifies the session detail
// endpoint has a response example with turns.
func TestBuildPeasantLocalAPISpec_HasSessionDetailExample(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "detailedSession") {
		t.Errorf("local spec missing session detail response example key %q", "detailedSession")
	}
}

// TestBuildPeasantLocalAPISpec_HasMockConfigExample verifies the mock config
// endpoint has a response example.
func TestBuildPeasantLocalAPISpec_HasMockConfigExample(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "mocksEnabled") {
		t.Errorf("local spec missing mock config response example key %q", "mocksEnabled")
	}
}

// TestBuildPeasantLocalAPISpec_HasShutdownExample verifies the shutdown endpoint
// has a response example.
func TestBuildPeasantLocalAPISpec_HasShutdownExample(t *testing.T) {
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	if !containsQuoted(jsonStr, "shuttingDown") {
		t.Errorf("local spec missing shutdown response example key %q", "shuttingDown")
	}
}

// ============================================================================
// Types — schema-level examples (via WithExamples in JSONSchema())
// ============================================================================

// TestBuildTypesSpec_HasSchemaExamples verifies that schema-level examples
// from WithExamples() in JSONSchema() methods appear in the shared types spec.
// Structurally navigates to components.schemas.{Type}.examples to avoid false
// positives from enum values that coincidentally match.
func TestBuildTypesSpec_HasSchemaExamples(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// Parse to navigate the JSON structure.
	var raw map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	comps, ok := raw["components"].(map[string]any)
	if !ok {
		t.Fatal("spec missing 'components' object")
	}
	schemas, ok := comps["schemas"].(map[string]any)
	if !ok {
		t.Fatal("spec missing 'components.schemas' object")
	}

	// Verify specific component schemas have an "examples" array (distinct from "enum").
	checks := []struct {
		component   string
		wantExample string
	}{
		{"Harness", "claude-code"},
		{"Role", "user"},
		{"Visibility", "public"},
	}
	for _, c := range checks {
		schema, ok := schemas[c.component].(map[string]any)
		if !ok {
			t.Errorf("component %q not found in schemas", c.component)
			continue
		}
		examples, ok := schema["examples"]
		if !ok {
			t.Errorf("component %q missing 'examples' field (WithExamples not propagated)", c.component)
			continue
		}
		exArr, ok := examples.([]any)
		if !ok || len(exArr) == 0 {
			t.Errorf("component %q 'examples' is not a non-empty array", c.component)
			continue
		}
		found := false
		for _, ex := range exArr {
			if ex == c.wantExample {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("component %q examples %v does not contain %q", c.component, exArr, c.wantExample)
		}
	}
}
