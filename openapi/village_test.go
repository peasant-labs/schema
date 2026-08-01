package openapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
)

// ============================================================================
// BuildVillageAPISpec tests (V12, V13, V15)
// ============================================================================

// TestBuildVillageAPISpec_OpenAPI31 verifies the village spec is valid OpenAPI 3.1 (V13).
func TestBuildVillageAPISpec_OpenAPI31(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}
	if spec == nil {
		t.Fatal("BuildVillageAPISpec() returned nil")
	}

	if spec.Openapi != "3.1.0" {
		t.Errorf("spec.Openapi = %q; want %q", spec.Openapi, "3.1.0")
	}
}

// TestBuildVillageAPISpec_Version verifies info.version is derived from the
// single-source VillageAPIVersion const (so the spec's doc-surface semver and the
// artifact filename never drift). It asserts against the const, not a literal, so
// a future bump updates one place.
func TestBuildVillageAPISpec_Version(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	if spec.Info.Version != specpkg.VillageAPIVersion {
		t.Errorf("info.version = %q; want %q (VillageAPIVersion)", spec.Info.Version, specpkg.VillageAPIVersion)
	}
}

// TestBuildVillageAPISpec_PublishPath verifies POST /api/v1/transcripts/publish exists.
func TestBuildVillageAPISpec_PublishPath(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	yamlStr := specYAML(t, spec)
	if !strings.Contains(yamlStr, "/api/v1/transcripts/publish") {
		t.Errorf("village spec missing /api/v1/transcripts/publish path; got:\n%s", yamlStr)
	}
}

// TestBuildVillageAPISpec_AnnotationManifestPath verifies the village spec carries
// the GET /api/v1/annotations/manifest route + its response schema (GH #69 C3 /
// follow-up o7kz6) — so the vendored village-api spec includes it and a re-vendor
// from peasant's generator won't clobber the canonical schema source.
func TestBuildVillageAPISpec_AnnotationManifestPath(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}
	yamlStr := specYAML(t, spec)
	for _, want := range []string{"/api/v1/annotations/manifest", "AnnotationManifestResponse"} {
		if !strings.Contains(yamlStr, want) {
			t.Errorf("village spec missing %q; got:\n%s", want, yamlStr)
		}
	}
}

func TestBuildVillageAPISpec_PullSkipGatePath(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}
	yamlStr := specYAML(t, spec)
	for _, want := range []string{
		"/api/v1/pull/transcripts/skip-gate",
		"PullSkipGateRequest",
		"PullSkipGateResponse",
	} {
		if !strings.Contains(yamlStr, want) {
			t.Errorf("village spec missing %q; got:\n%s", want, yamlStr)
		}
	}
}

// TestBuildVillageAPISpec_PublishRequestComponent verifies the operation-specific publish body component.
func TestBuildVillageAPISpec_PublishRequestComponent(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	components := spec.ComponentsEns().Schemas
	if _, ok := components["OpenapiTranscriptPublishRequest"]; !ok {
		t.Errorf("village spec missing operation-specific TranscriptPublishRequest component")
	}
	if _, ok := components["SchemaPublishRequest"]; ok {
		t.Errorf("village spec shadows canonical PublishRequest with an operation component")
	}
}

// TestBuildVillageAPISpec_SessionEntryComponent verifies the successor SessionEntry with toolCallId.
func TestBuildVillageAPISpec_SessionEntryComponent(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// The successor SessionEntry must be present as a component schema key (quoted to avoid
	// false positives from matching the word in description text).
	if !strings.Contains(jsonStr, `"SchemaAuthoritativeSessionEntry"`) {
		t.Errorf("village spec missing SchemaAuthoritativeSessionEntry component key in components/schemas")
	}

	// toolCallId field must appear in the spec.
	if !strings.Contains(jsonStr, "toolCallId") {
		t.Errorf("village spec missing toolCallId field")
	}
}

// TestBuildVillageAPISpec_ToolCallKindEnum verifies ToolCallKind enum appears (V15).
func TestBuildVillageAPISpec_ToolCallKindEnum(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// ToolCallKind must be present as a component schema key (quoted to avoid
	// false positives from matching the word in description text).
	if !strings.Contains(jsonStr, `"SchemaToolCallKind"`) {
		t.Errorf("village spec missing SchemaToolCallKind component key in components/schemas")
	}

	// Check that the enum values are present.
	for _, kind := range []string{"read", "edit", "search", "execute", "other"} {
		if !strings.Contains(jsonStr, `"`+kind+`"`) {
			t.Errorf("village spec missing ToolCallKind enum value %q", kind)
		}
	}
}

// TestBuildVillageAPISpec_StopReasonEnum verifies StopReason enum appears (V15).
func TestBuildVillageAPISpec_StopReasonEnum(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	// StopReason must be present as a component schema key (quoted to avoid
	// false positives from matching the word in description text).
	if !strings.Contains(jsonStr, `"SchemaStopReason"`) {
		t.Errorf("village spec missing SchemaStopReason component key in components/schemas")
	}

	for _, reason := range []string{"end_turn", "cancelled", "max_tokens"} {
		if !strings.Contains(jsonStr, `"`+reason+`"`) {
			t.Errorf("village spec missing StopReason enum value %q", reason)
		}
	}
}

// TestBuildVillageAPISpec_VisibilityEnum verifies Visibility enum appears.
func TestBuildVillageAPISpec_VisibilityEnum(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	jsonStr := specJSON(t, spec)

	for _, vis := range []string{"private", "group", "public"} {
		if !strings.Contains(jsonStr, `"`+vis+`"`) {
			t.Errorf("village spec missing Visibility enum value %q", vis)
		}
	}
}

// TestBuildVillageAPISpec_ValidJSON verifies the output is valid JSON.
func TestBuildVillageAPISpec_ValidJSON(t *testing.T) {
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec() error: %v", err)
	}

	data, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Error("village spec MarshalJSON produced invalid JSON")
	}
}
