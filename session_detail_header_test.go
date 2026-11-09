package schema_test

import (
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
)

// TestSessionDetailHeaderMirrorsDurableMetadata keeps the metadata-only header
// a projection of the durable payload: every durable field that is not content
// must exist on the header, and every header field must exist on the durable
// payload. A new durable metadata field therefore cannot be added without
// deciding whether the header carries it.
func TestSessionDetailHeaderMirrorsDurableMetadata(t *testing.T) {
	durable := jsonFieldNames(t, reflect.TypeOf(schema.SessionDetailPayload{}))
	header := jsonFieldNames(t, reflect.TypeOf(schema.SessionDetailHeader{}))
	payload := jsonFieldNames(t, reflect.TypeOf(schema.SessionDetailHeaderPayload{}))

	// Content that cannot live on a metadata-only projection: the turn list and
	// the per-turn native metadata that is validated against those turns. The
	// earlier-history key survives as section state only.
	contentOnly := map[string]bool{"turns": true, "nativeMetadata": true}
	for name := range durable {
		if contentOnly[name] {
			continue
		}
		if !header[name] {
			t.Errorf("SessionDetailHeader is missing durable metadata field %q; add it to the header or record why it is content-only", name)
		}
	}
	for name := range header {
		if !durable[name] {
			t.Errorf("SessionDetailHeader carries %q, which the durable SessionDetailPayload does not; a header field needs a durable source", name)
		}
	}
	for name := range header {
		if !payload[name] {
			t.Errorf("SessionDetailHeaderPayload is missing promoted header field %q", name)
		}
	}
	for _, name := range []string{"turns", "nativeMetadata"} {
		if header[name] {
			t.Errorf("SessionDetailHeader carries %q; the metadata-only frame must not ship turn content or turn-coupled native metadata", name)
		}
	}
	for _, name := range []string{"relationshipNavigation", "teammate"} {
		if !payload[name] {
			t.Errorf("SessionDetailHeaderPayload lacks local frame field %q", name)
		}
	}
}

// jsonFieldNames returns the JSON property names a struct exposes, following
// anonymous embeddings because they promote their fields onto the wire.
func jsonFieldNames(t *testing.T, typ reflect.Type) map[string]bool {
	t.Helper()
	names := map[string]bool{}
	var walk func(reflect.Type)
	walk = func(current reflect.Type) {
		for current.Kind() == reflect.Pointer {
			current = current.Elem()
		}
		if current.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < current.NumField(); i++ {
			field := current.Field(i)
			if field.Anonymous {
				walk(field.Type)
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "-" || name == "" {
				continue
			}
			names[name] = true
		}
	}
	walk(typ)
	return names
}
