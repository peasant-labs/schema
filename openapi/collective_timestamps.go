package openapi

import schema "github.com/peasant-labs/schema"

// Apply only to the collective evidence graph, not unrelated legacy contracts.
func constrainCollectiveTimestamps(components map[string]map[string]interface{}) {
	seen := map[string]bool{}
	var walk func(interface{})
	walk = func(value interface{}) {
		switch v := value.(type) {
		case map[string]interface{}:
			if ref, ok := v["$ref"].(string); ok {
				const prefix = "#/components/schemas/"
				if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
					name := ref[len(prefix):]
					if !seen[name] {
						seen[name] = true
						walk(components[name])
					}
				}
			}
			if v["format"] == "date-time" {
				v["pattern"] = schema.CollectiveTimestampPattern
				v["not"] = map[string]interface{}{"pattern": `^0001-01-01T00:00:00(\.0+)?Z$`}
			}
			for key, item := range v {
				if key != "not" {
					walk(item)
				}
			}
		case []interface{}:
			for _, item := range v {
				walk(item)
			}
		case map[string]map[string]interface{}:
			for _, item := range v {
				walk(item)
			}
		}
	}
	for _, name := range []string{"VillageShareEvent", "LocalRemoteContributionAttempt", "CollectiveObservationResponse", "LocalContributionHistoryResponse", "PullAssociationMetadataResult"} {
		walk(components[name])
	}
}
