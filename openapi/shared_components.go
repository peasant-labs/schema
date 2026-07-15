package openapi

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/swaggest/openapi-go/openapi31"
)

// harmonizeSharedTypeComponents replaces every API component whose canonical
// name exists in the Types document with that exact schema. Operation-specific
// components keep distinct names and are therefore left untouched.
func harmonizeSharedTypeComponents(spec *openapi31.Spec) error {
	typesSpec, err := BuildTypesSpec()
	if err != nil {
		return fmt.Errorf("harmonize shared API components: build canonical Types document: %w", err)
	}
	apiComponents := spec.ComponentsEns().Schemas
	typeComponents := typesSpec.ComponentsEns().Schemas
	if len(apiComponents) == 0 || len(typeComponents) == 0 {
		return fmt.Errorf("harmonize shared API components: API or Types components are empty; generation cannot establish canonical schema identity")
	}

	rawNames := make([]string, 0, len(apiComponents))
	for rawName := range apiComponents {
		rawNames = append(rawNames, rawName)
	}
	sort.Strings(rawNames)
	rawByCanonical := make(map[string]string, len(rawNames))
	for _, rawName := range rawNames {
		canonical := canonicalTypeName(rawName)
		prior, exists := rawByCanonical[canonical]
		if !exists || componentNamePriority(rawName, canonical) < componentNamePriority(prior, canonical) {
			rawByCanonical[canonical] = rawName
		}
	}

	for _, rawName := range rawNames {
		canonical := canonicalTypeName(rawName)
		typeSchema, shared := typeComponents[canonical]
		if !shared {
			continue
		}
		clone, err := cloneSchemaMap(typeSchema)
		if err != nil {
			return fmt.Errorf("harmonize shared API component %q from canonical type %q: %w", rawName, canonical, err)
		}
		rewriteCanonicalComponentRefs(clone, rawByCanonical)
		apiComponents[rawName] = clone
	}
	return nil
}

func componentNamePriority(rawName, canonical string) int {
	switch rawName {
	case canonical:
		return 0
	case "Schema" + canonical:
		return 1
	default:
		return 2
	}
}

func cloneSchemaMap(source map[string]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("encode canonical schema: %w", err)
	}
	var clone map[string]interface{}
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, fmt.Errorf("decode canonical schema clone: %w", err)
	}
	return clone, nil
}

func rewriteCanonicalComponentRefs(node any, rawByCanonical map[string]string) {
	switch value := node.(type) {
	case map[string]interface{}:
		for key, child := range value {
			if key == "$ref" {
				if ref, ok := child.(string); ok {
					const prefix = "#/components/schemas/"
					if strings.HasPrefix(ref, prefix) {
						canonical := strings.TrimPrefix(ref, prefix)
						if rawName, exists := rawByCanonical[canonical]; exists {
							value[key] = prefix + rawName
						}
					}
				}
				continue
			}
			rewriteCanonicalComponentRefs(child, rawByCanonical)
		}
	case []interface{}:
		for _, child := range value {
			rewriteCanonicalComponentRefs(child, rawByCanonical)
		}
	}
}
