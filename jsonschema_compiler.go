package schema

import jsonschema "github.com/santhosh-tekuri/jsonschema/v5"

// NewJSONSchemaCompiler returns the canonical compiler for schema-owned
// validation assets. Custom wire formats are assertions, not annotations, so
// callers cannot accidentally compile a schema that silently ignores them.
func NewJSONSchemaCompiler() *jsonschema.Compiler {
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat = true
	compiler.Formats[PublicRefUTF8ByteFormat] = ValidatePublicRefJSONSchemaFormat
	return compiler
}
