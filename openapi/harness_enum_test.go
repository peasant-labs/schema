package openapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
)

// These tests guard the core behavior of the B' InterceptSchema hook
// (harness_schema.go): that the Harness enum actually appears in the generated
// OpenAPI specs. Because schema.Harness is an alias for bestiary.Harness, the
// hook injects the enum via a runtime reflect.Type comparison rather than an
// Exposer method. A subtle type mismatch (or a missing registerHarnessSchema
// call on a future reflector) would silently drop the enum to a bare
// {"type":"string"} with NO compile error — undetected until API consumers
// notice. These assertions fail loudly if that regresses.

func assertHarnessEnumPresent(t *testing.T, specJSON string) {
	t.Helper()
	// The expected values come from the same source the hook uses
	// (schema.Harnesses), so the test tracks the hook automatically. If the hook
	// fails to emit the enum, none of these values appear and the test fails.
	for _, h := range schema.Harnesses() {
		if !strings.Contains(specJSON, `"`+string(h)+`"`) {
			t.Errorf("generated spec is missing harness value %q — the InterceptSchema hook may not be emitting the Harness enum", h)
		}
	}
}

func TestBuildVillageAPISpec_HarnessEnum(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec(): %v", err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal village spec: %v", err)
	}
	assertHarnessEnumPresent(t, string(data))
}
