package schema

import jsonschema "github.com/swaggest/jsonschema-go"

// closedStringEnumSchema builds the OpenAPI schema for a closed string type
// from its canonical Go inventory. Keeping the values in the All* collection
// makes the Go API and every generated language binding consume one source.
func closedStringEnumSchema[T ~string](title, description string, values []T) jsonschema.Schema {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle(title)
	s.WithDescription(description)
	enum := make([]any, len(values))
	for i, value := range values {
		enum[i] = string(value)
	}
	s.WithEnum(enum...)
	s.WithExamples(enum...)
	return s
}
