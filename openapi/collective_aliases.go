package openapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/swaggest/openapi-go/openapi31"
)

// Expand after all response/body corrections. Unchanged nested payload schemas
// are shared; operation objects are copied before changing deprecation or IDs.
// The released grouped contributable envelope is the one response exception.
// Helper transcript-groups and share operations deliberately stay unchanged.
func addVillageCollectiveAliases(spec *openapi31.Spec) error {
	paths := spec.Paths.MapOfPathItemValues
	for name, item := range paths {
		if name != "/api/v1/groups" && !strings.HasPrefix(name, "/api/v1/groups/") {
			continue
		}
		canonical := strings.Replace(name, "/api/v1/groups", "/api/v1/collectives", 1)
		if _, exists := paths[canonical]; exists {
			return fmt.Errorf("canonical collective route %s already exists; declare one source path before alias expansion", canonical)
		}
		paths[canonical] = item
		for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"} {
			op, err := item.Operation(method)
			if err != nil {
				return err
			}
			if op == nil {
				continue
			}
			deprecated := *op
			if name == "/api/v1/groups/{id}/contributable" && method == "GET" {
				// This one legacy response retains groupId. Copy its nested response
				// before replacing the union so the canonical operation stays group_id.
				data, err := json.Marshal(op)
				if err != nil {
					return fmt.Errorf("copy deprecated grouped contributable operation before response selection: %w", err)
				}
				deprecated = openapi31.Operation{}
				if err := json.Unmarshal(data, &deprecated); err != nil {
					return fmt.Errorf("decode deprecated grouped contributable operation before response selection: %w", err)
				}
			}
			deprecated.WithID(*op.ID + "Deprecated").WithDeprecated(true)
			if err := item.SetOperation(method, &deprecated); err != nil {
				return err
			}
		}
		paths[name] = item
	}
	return setVillageGroupedResponse(spec, "/api/v1/groups/{id}/contributable", map[string]interface{}{"$ref": "#/components/schemas/SchemaVillageContributableResponse"}, "SchemaVillageDeprecatedGroupedContributableResponse")
}
