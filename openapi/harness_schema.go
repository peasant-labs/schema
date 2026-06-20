package openapi

import (
	"reflect"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// harnessType is the reflect.Type of schema.Harness (an alias for bestiary.Harness).
var harnessType = reflect.TypeFor[schema.Harness]()

// registerHarnessSchema teaches an OpenAPI reflector to emit the Harness enum
// schema for every schema.Harness field it encounters — top-level entry points
// (the "Provider" shared component) and fields nested inside reflected structs
// (SessionEntry, metadata, local-API payloads, …).
//
// Why this exists: schema.Harness is a type ALIAS for bestiary.Harness, an
// external type. Go forbids defining methods on it, so — unlike the sibling enum
// types (Visibility, Role, SourceFormat, SessionOutcome), which each implement
// jsonschema.Exposer via a JSONSchema() method — Harness cannot expose its enum
// to the reflector through the interface. Without this hook the reflector emits a
// bare {"type":"string"} for every Harness field and the enum vanishes from the
// generated specs. We inject the authoritative schema (schema.HarnessJSONSchema)
// via an InterceptSchema hook keyed on the Harness type.
//
// TEMPORARY (bestiary-migration scaffolding): delete this hook and its callers
// once bestiary.Harness gains a native JSONSchema() Exposer, or once peasant
// stops aliasing it in favour of a local newtype carrying the method.
func registerHarnessSchema(r *openapi31.Reflector) {
	jsr := r.JSONSchemaReflector()
	jsr.DefaultOptions = append(jsr.DefaultOptions, jsonschema.InterceptSchema(
		func(params jsonschema.InterceptSchemaParams) (bool, error) {
			// The interceptor fires twice per value (before and after default
			// processing). Apply only on the post-processing pass, once default
			// reflection has set type:string — this avoids writing the enum twice.
			if !params.Processed || !isHarnessValue(params.Value) {
				return false, nil
			}
			hs := schema.HarnessJSONSchema()
			params.Schema.Enum = hs.Enum
			params.Schema.Title = hs.Title
			params.Schema.Description = hs.Description
			params.Schema.Examples = hs.Examples
			// Return false so we don't suppress any remaining default processing.
			return false, nil
		},
	))
}

// isHarnessValue reports whether v reflects the Harness type (or a pointer to it).
func isHarnessValue(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	t := v.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t == harnessType
}
