package openapi

import (
	"encoding/json"
	"fmt"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// publishRequestSchemaArtifactName is the filename of the standalone PublishRequest
// JSON Schema the village vendors + enforces. It is versioned by the village-api
// doc-surface version (the spec it is extracted from), so a VillageAPIVersion bump
// renames it in lockstep with the spec it derives from.
func publishRequestSchemaArtifactName() string {
	return "publish-request-" + VillageAPIVersion + ".schema.json"
}

// Versioned-spec info.version values are re-exported from the root schema package
// (the single source of truth — see versions.go there). The spec builders in this
// package reference them unqualified, and the root accessor schema.VillageAPISpecJSON()
// keys its embedded-bytes lookup off the same constants, so a version bump stays a
// one-line edit at the root and flows here automatically. They are kept exported here
// for backward compatibility with existing openapi.* call sites.
const (
	// VillageAPIVersion is the info.version of the Village API spec.
	VillageAPIVersion = schema.VillageAPIVersion
	// PeasantLocalAPIVersion is the info.version of the local dashboard API spec.
	PeasantLocalAPIVersion = schema.PeasantLocalAPIVersion
	// TypesVersion is the info.version of the types spec.
	TypesVersion = schema.TypesVersion
)

// SpecArtifacts is the set of generated OpenAPI spec files, keyed by filename
// (e.g. "village-api-0.2.0.json"), each mapping to the exact bytes written to
// generated/ at the module root. It is the single source of truth for
// what the generator emits, shared by cmd/schema-gen (which writes them) and the
// codegen-freshness test (which diffs them against the committed copies). It does
// NOT include the HTML docs/CLI reference — those are non-contract artifacts.
type SpecArtifacts map[string][]byte

// GenerateSpecArtifacts builds every committed OpenAPI JSON/YAML spec from the Go
// schema source and returns them as filename->bytes. The marshaling here MUST
// match cmd/schema-gen exactly (both call this function), so a clean tree
// regenerates byte-for-byte and the freshness test only fails on real drift.
func GenerateSpecArtifacts() (SpecArtifacts, error) {
	out := make(SpecArtifacts)

	// --- Standalone PublishRequest JSON Schema (village vendors + enforces) ---
	// Extracted from the village-api spec (single source of truth) rather than
	// reflected separately, so the documented surface and the enforced wire schema
	// cannot drift. This replaced the retired BuildJSONSchema/BuildOpenAPISpec
	// legacy artifacts (GH #53).
	publishSchemaBytes, err := BuildPublishRequestSchema(VillageAPIVersion)
	if err != nil {
		return nil, fmt.Errorf("build PublishRequest schema: %w", err)
	}
	out[publishRequestSchemaArtifactName()] = publishSchemaBytes

	// --- Versioned specs (village, peasant-local, shared types) ---
	specs := []struct {
		name  string
		build func() (*openapi31.Spec, error)
	}{
		{"village-api-" + VillageAPIVersion, BuildVillageAPISpec},
		{"peasantlocal-api-" + PeasantLocalAPIVersion, BuildPeasantLocalAPISpec},
		{"types-" + TypesVersion, BuildTypesSpec},
	}
	for _, s := range specs {
		spec, err := s.build()
		if err != nil {
			return nil, fmt.Errorf("build %s: %w", s.name, err)
		}
		jsonBytes, yamlBytes, err := marshalSpec(spec, s.name)
		if err != nil {
			return nil, err
		}
		out[s.name+".json"] = jsonBytes
		out[s.name+".yaml"] = yamlBytes
	}

	return out, nil
}

// marshalSpec marshals an OpenAPI spec to pretty JSON + YAML exactly as the
// generator writes them.
func marshalSpec(spec *openapi31.Spec, name string) (jsonBytes, yamlBytes []byte, err error) {
	rawJSON, err := spec.MarshalJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("marshal %s JSON: %w", name, err)
	}
	var pretty json.RawMessage = rawJSON
	jsonBytes, err = json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("pretty-print %s JSON: %w", name, err)
	}
	yamlBytes, err = spec.MarshalYAML()
	if err != nil {
		return nil, nil, fmt.Errorf("marshal %s YAML: %w", name, err)
	}
	return jsonBytes, yamlBytes, nil
}
