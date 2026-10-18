package openapi

import (
	"fmt"
	schema "github.com/peasant-labs/schema"
)

func constrainContributionHistory(components map[string]map[string]interface{}) error {
	base := components["LocalContributionHistoryResponse"]
	props, ok := base["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("history union generation lost reflected properties; restore the canonical history response before generating")
	}
	// Counts are JSON-safe numbers, not int64 transport values or bigint output.
	if total, ok := props["total"].(map[string]interface{}); ok {
		delete(total, "format")
	}
	var variants []interface{}
	for _, kind := range schema.AllLocalContributionEvidenceKinds {
		p := map[string]interface{}{}
		for key, value := range props {
			if key != "remoteAttempts" && key != "receipts" && key != "operationDiagnostics" {
				p[key] = value
			}
		}
		key := "remoteAttempts"
		if kind == schema.LocalContributionEvidenceReceipts {
			key = "receipts"
		} else if kind == schema.LocalContributionEvidenceOperationDiagnostics {
			key = "operationDiagnostics"
		}
		p[key] = props[key]
		p["kind"] = map[string]interface{}{"type": "string", "const": string(kind)}
		variants = append(variants, map[string]interface{}{"type": "object", "additionalProperties": false, "properties": p, "required": []string{"scope", "kind", "traversalRevision", "page", "limit", "total", "coverage", key}})
	}
	delete(base, "properties")
	delete(base, "required")
	delete(base, "type")
	base["oneOf"] = variants
	return nil
}
