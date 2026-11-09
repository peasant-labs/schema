package openapi

import (
	"reflect"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// Go aliases do not retain their names in reflect.Type. Bind each concrete
// instantiation of the generic offset page to its explicit catalog identity
// before any references are made. Their properties still come exclusively from
// reflecting schema.OffsetPage[T].
func registerSyncPageName(r *openapi31.Reflector) {
	r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
		jsonschema.InterceptDefName(func(t reflect.Type, name string) string {
			switch t {
			case reflect.TypeFor[schema.LocalSyncSessionPage]():
				return "LocalSyncSessionPage"
			case reflect.TypeFor[schema.LocalSessionPage]():
				return "LocalSessionPage"
			}
			return name
		}))
}
