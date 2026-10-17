package openapi

import (
	"reflect"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// Go aliases do not retain their names in reflect.Type. Bind this concrete
// instantiation to its explicit catalog identity before any references are made.
// Its properties still come exclusively from reflecting schema.Page[T].
func registerSyncPageName(r *openapi31.Reflector) {
	r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
		jsonschema.InterceptDefName(func(t reflect.Type, name string) string {
			if t == reflect.TypeFor[schema.LocalSyncSessionPage]() {
				return "LocalSyncSessionPage"
			}
			return name
		}))
}
