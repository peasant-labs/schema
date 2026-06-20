package openapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

// publishRequestComponent is the components/schemas key the village-api spec
// emits for schema.PublishRequest (the openapi31 reflector prefixes "Schema").
const publishRequestComponent = "SchemaPublishRequest"

// componentsRefPrefix is the $ref prefix the village-api spec uses for its
// reusable component schemas.
const componentsRefPrefix = "#/components/schemas/"

// defsRefPrefix is the $ref prefix the extracted standalone JSON Schema uses for
// its bundled definitions. Draft 2020-12 names this keyword "$defs".
const defsRefPrefix = "#/$defs/"

// jsonSchema2020Dialect is the JSON Schema dialect the extracted artifact pins
// explicitly via "$schema". OpenAPI 3.1 component schemas are aligned with JSON
// Schema 2020-12, and the village's validator (santhosh-tekuri/jsonschema v5)
// auto-detects this dialect from the keyword and supports it.
const jsonSchema2020Dialect = "https://json-schema.org/draft/2020-12/schema"

// BuildPublishRequestSchema EXTRACTS the PublishRequest component — with all of
// its transitive $ref dependencies, bundled and self-contained — out of the
// village-api OpenAPI 3.1 spec into a standalone JSON Schema 2020-12 document.
//
// This is the SINGLE SOURCE OF TRUTH for the village's publish-body enforcement:
// the village vendors the bytes returned here as
// backend/internal/handler/openapispec/publish-request.schema.json and compiles
// them with santhosh-tekuri/jsonschema to reject malformed publish bodies (422).
// Deriving it from the village-api spec (rather than a separate reflector) means
// the documented HTTP surface and the enforced wire schema can never drift — they
// come from one Build*Spec function (GH #53: peasant is the source of truth, the
// village vendors + enforces).
//
// The returned document:
//   - sets "$schema" to the JSON Schema 2020-12 dialect (pinned explicitly so the
//     village validator selects the right draft);
//   - sets "$id" to urn:peasant:publish-request:<VillageAPIVersion> (derived from
//     the single-source version const, never retyped);
//   - hoists SchemaPublishRequest's own keywords (type, properties) to the root;
//   - bundles every transitively-referenced component under "$defs", with all
//     "#/components/schemas/X" $refs rewritten to "#/$defs/X".
//
// version is the doc-surface version stamped into "$id" (callers pass
// VillageAPIVersion).
func BuildPublishRequestSchema(version string) ([]byte, error) {
	spec, err := BuildVillageAPISpec()
	if err != nil {
		return nil, fmt.Errorf("build village-api spec for PublishRequest extraction: %w", err)
	}

	// Marshal the spec to JSON and walk it as a generic tree. Extracting from the
	// emitted bytes (rather than the typed openapi31.Spec) guarantees the standalone
	// schema is byte-faithful to what the spec documents — including the harness
	// enum injected via registerHarnessSchema — with zero reflection drift.
	specJSON, err := spec.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal village-api spec: %w", err)
	}
	var specDoc map[string]any
	if err := json.Unmarshal(specJSON, &specDoc); err != nil {
		return nil, fmt.Errorf("unmarshal village-api spec: %w", err)
	}

	components, err := componentSchemas(specDoc)
	if err != nil {
		return nil, err
	}

	root, ok := components[publishRequestComponent].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"village-api spec is missing the %q component in components/schemas; "+
				"cannot extract the PublishRequest schema — verify BuildVillageAPISpec still "+
				"registers schema.PublishRequest as the publish request body",
			publishRequestComponent)
	}

	// Collect the transitive closure of components referenced from PublishRequest.
	closure := map[string]struct{}{}
	if err := collectRefs(root, components, closure); err != nil {
		return nil, err
	}

	// Build the bundled $defs from the closure, rewriting every internal ref from
	// the OpenAPI components pointer to the standalone $defs pointer.
	defs := make(map[string]any, len(closure))
	for name := range closure {
		comp, ok := components[name].(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"PublishRequest references component %q which is not present (or not an object) "+
					"in the village-api spec's components/schemas; the spec is internally inconsistent",
				name)
		}
		defs[name] = rewriteRefs(comp)
	}

	// Assemble the standalone schema: PublishRequest's own keywords at the root,
	// plus the bundled $defs. The root's own refs are rewritten too.
	out := map[string]any{}
	for k, v := range rewriteRefs(root).(map[string]any) {
		out[k] = v
	}
	out["$schema"] = jsonSchema2020Dialect
	out["$id"] = "urn:peasant:publish-request:" + version
	out["title"] = "Publish Request Schema"
	out["description"] = "JSON Schema for peasant push PublishRequest wire format, " +
		"extracted from the village-api OpenAPI spec (peasant is the source of truth; " +
		"the village vendors + enforces this schema)."
	out["$defs"] = defs

	// Marshal with sorted keys (encoding/json sorts map keys) and indentation that
	// matches the rest of the generated artifacts.
	pretty, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal extracted PublishRequest schema: %w", err)
	}
	return pretty, nil
}

// componentSchemas pulls components/schemas out of a marshaled spec document.
func componentSchemas(specDoc map[string]any) (map[string]any, error) {
	comps, ok := specDoc["components"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("village-api spec has no components object; cannot extract schemas")
	}
	schemas, ok := comps["schemas"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("village-api spec has no components/schemas object; cannot extract schemas")
	}
	return schemas, nil
}

// collectRefs walks node, adding every "#/components/schemas/X" target it finds to
// acc, then recursing into each newly-discovered component so the closure is
// transitive. components is the spec's components/schemas map.
func collectRefs(node any, components map[string]any, acc map[string]struct{}) error {
	switch n := node.(type) {
	case map[string]any:
		for k, v := range n {
			if k == "$ref" {
				ref, ok := v.(string)
				if !ok {
					return fmt.Errorf("$ref value is not a string: %v", v)
				}
				name, ok := componentName(ref)
				if !ok {
					// A non-component ref (should not occur in these specs). Skip.
					continue
				}
				if _, seen := acc[name]; seen {
					continue
				}
				acc[name] = struct{}{}
				child, ok := components[name].(map[string]any)
				if !ok {
					return fmt.Errorf(
						"component %q referenced via %q is absent from components/schemas",
						name, ref)
				}
				if err := collectRefs(child, components, acc); err != nil {
					return err
				}
				continue
			}
			if err := collectRefs(v, components, acc); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range n {
			if err := collectRefs(item, components, acc); err != nil {
				return err
			}
		}
	}
	return nil
}

// rewriteRefs deep-copies node, rewriting every "#/components/schemas/X" $ref to
// "#/$defs/X" so the extracted schema is self-contained.
func rewriteRefs(node any) any {
	switch n := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(n))
		for k, v := range n {
			if k == "$ref" {
				if ref, ok := v.(string); ok {
					if name, ok := componentName(ref); ok {
						out[k] = defsRefPrefix + name
						continue
					}
				}
			}
			out[k] = rewriteRefs(v)
		}
		return out
	case []any:
		out := make([]any, len(n))
		for i, item := range n {
			out[i] = rewriteRefs(item)
		}
		return out
	default:
		return n
	}
}

// componentName returns the trailing component name of a components/schemas $ref
// (e.g. "#/components/schemas/SchemaModelInfo" -> "SchemaModelInfo", true). It
// reports false for refs that do not point into components/schemas.
func componentName(ref string) (string, bool) {
	if !strings.HasPrefix(ref, componentsRefPrefix) {
		return "", false
	}
	return strings.TrimPrefix(ref, componentsRefPrefix), true
}
