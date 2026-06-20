package openapi

import (
	"fmt"
	"strings"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// BuildTypesSpec builds an OpenAPI 3.1 specification registering foundational shared
// domain types as reusable components. No paths — this is a pure type catalog
// for SDK code generators and type-sharing across API boundaries.
//
// Registered entry-point types (7 top-level components):
//   - Provider, Role, SessionID, ModelID (primitive domain types)
//   - Visibility (access control)
//   - SessionEntry (content layer)
//   - QualityMetrics (analytics)
//
// Transitively referenced types (ToolCallKind, StopReason, EntryType, SessionOutcome, etc.)
// are collected via CollectDefinitions and registered as separate top-level components.
// All component names and $ref paths are normalized to plain names (no "Schema" prefix).
func BuildTypesSpec() (*openapi31.Spec, error) {
	r := openapi31.NewReflector()
	registerHarnessSchema(r)
	r.Spec.Info.
		WithTitle("Peasant Types").
		// Derived from TypesVersion (single source — see package doc).
		WithVersion(TypesVersion).
		WithDescription("Shared domain type catalog for Peasant APIs. No operations — register these " +
			"components for SDK code generation and cross-API type sharing.")

	entryPoints := []struct {
		name string
		val  interface{}
	}{
		{"Provider", new(schema.Harness)},
		{"Role", new(schema.Role)},
		{"SessionID", new(schema.SessionID)},
		{"ModelID", new(schema.ModelID)},
		{"Visibility", new(schema.Visibility)},
		{"SessionEntry", new(schema.SessionEntry)},
		{"QualityMetrics", new(schema.QualityMetrics)},
	}

	jsr := r.JSONSchemaReflector()
	for _, t := range entryPoints {
		s, err := jsr.Reflect(t.val, jsonschema.CollectDefinitions(
			func(name string, defSchema jsonschema.Schema) {
				// Strip the "Schema" prefix that swaggest derives from the package alias,
				// so transitive types are stored as "Provider", "Role", etc.
				plainName := strings.TrimPrefix(name, "Schema")
				sm, smErr := defSchema.ToSchemaOrBool().ToSimpleMap()
				if smErr != nil {
					return
				}
				r.SpecEns().ComponentsEns().WithSchemasItem(plainName, sm)
			},
		))
		if err != nil {
			return nil, fmt.Errorf("reflect shared type %s: %w", t.name, err)
		}
		sm, err := s.ToSchemaOrBool().ToSimpleMap()
		if err != nil {
			return nil, fmt.Errorf("marshal shared type %s: %w", t.name, err)
		}
		r.SpecEns().ComponentsEns().WithSchemasItem(t.name, sm)
	}

	// Fix $ref values: JSON Schema Draft 4 CollectDefinitions uses "#/definitions/"
	// but OAS 3.1 components live under "#/components/schemas/". Walk all component
	// schemas and rewrite any "#/definitions/SchemaX" ref to "#/components/schemas/X"
	// (stripping the "Schema" prefix to match the plain component names above).
	if comps := r.SpecEns().Components; comps != nil {
		for _, schemaMap := range comps.Schemas {
			fixSharedDefinitionRefs(schemaMap)
		}
	}

	return r.Spec, nil
}

// fixSharedDefinitionRefs is like fixDefinitionRefs but also strips the "Schema"
// prefix from definition names, normalizing "#/definitions/SchemaX" to
// "#/components/schemas/X". Used by BuildTypesSpec where all components
// are stored under plain names (no "Schema" prefix).
func fixSharedDefinitionRefs(m map[string]interface{}) {
	const oldPrefix = "#/definitions/"
	const newPrefix = "#/components/schemas/"
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if k == "$ref" && strings.HasPrefix(val, oldPrefix) {
				name := strings.TrimPrefix(val, oldPrefix)
				name = strings.TrimPrefix(name, "Schema")
				m[k] = newPrefix + name
			}
		case map[string]interface{}:
			fixSharedDefinitionRefs(val)
		case []interface{}:
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					fixSharedDefinitionRefs(itemMap)
				}
			}
		}
	}
}

// fixDefinitionRefs recursively walks a schema map and rewrites any $ref values
// that use the JSON Schema Draft 4 "#/definitions/" prefix to the OAS 3.1
// "#/components/schemas/" prefix.
func fixDefinitionRefs(m map[string]interface{}) {
	const oldPrefix = "#/definitions/"
	const newPrefix = "#/components/schemas/"
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if k == "$ref" && strings.HasPrefix(val, oldPrefix) {
				m[k] = newPrefix + strings.TrimPrefix(val, oldPrefix)
			}
		case map[string]interface{}:
			fixDefinitionRefs(val)
		case []interface{}:
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					fixDefinitionRefs(itemMap)
				}
			}
		}
	}
}

// addComponentSchema reflects a Go type and registers it as a named component schema
// in the OpenAPI spec. Use this for types not naturally referenced by any operation.
func addComponentSchema(r *openapi31.Reflector, name string, val interface{}) error {
	jsr := r.JSONSchemaReflector()
	s, err := jsr.Reflect(val)
	if err != nil {
		return fmt.Errorf("reflect component schema %s: %w", name, err)
	}
	sm, err := s.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return fmt.Errorf("marshal component schema %s: %w", name, err)
	}
	r.SpecEns().ComponentsEns().WithSchemasItem(name, sm)
	return nil
}

// addRESTOp adds a REST operation to the reflector.
func addRESTOp(r *openapi31.Reflector, method, path, opID, desc string, tags []string, req, resp interface{}) error {
	oc, err := r.NewOperationContext(method, path)
	if err != nil {
		return fmt.Errorf("new operation %s %s: %w", method, path, err)
	}
	if req != nil {
		oc.AddReqStructure(req)
	}
	if resp != nil {
		oc.AddRespStructure(resp)
	}
	oc.SetDescription(desc)
	oc.SetID(opID)
	for _, tag := range tags {
		oc.SetTags(tag)
	}
	if err := r.AddOperation(oc); err != nil {
		return fmt.Errorf("add operation %s %s: %w", method, path, err)
	}
	return nil
}
