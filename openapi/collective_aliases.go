package openapi

import (
	"fmt"
	"github.com/swaggest/openapi-go/openapi31"
	"strings"
)

// Expand after all response/body corrections. Immutable nested payload schemas
// are shared; operation objects are copied before changing deprecation or IDs.
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
			deprecated.WithID(*op.ID + "Deprecated").WithDeprecated(true)
			if err := item.SetOperation(method, &deprecated); err != nil {
				return err
			}
		}
		paths[name] = item
	}
	return nil
}
