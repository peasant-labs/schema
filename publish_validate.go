package schema

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// legacyPublishRequestSchemaURL is the in-compiler resource name for the frozen
// Village 0.10.0 publish-request schema retained for rc11 compatibility.
const legacyPublishRequestSchemaURL = "publish-request-0.10.0.schema.json"

var (
	publishRequestSchema     *jsonschema.Schema
	publishRequestSchemaErr  error
	publishRequestSchemaOnce sync.Once
)

//go:embed generated/publish-request-0.10.0.schema.json
var releasedPublishRequestSchema []byte

// compilePublishRequestSchema compiles the single byte-frozen schema used by
// the released rc11 compatibility validator. The successor validator has a
// separate source type and must not silently change this legacy boundary.
func compilePublishRequestSchema() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(legacyPublishRequestSchemaURL, bytes.NewReader(releasedPublishRequestSchema)); err != nil {
		publishRequestSchemaErr = err
		return
	}
	publishRequestSchema, publishRequestSchemaErr = compiler.Compile(legacyPublishRequestSchemaURL)
}

// ValidatePublishRequest validates a legacy rc11 publish-request JSON body
// against the byte-frozen Village 0.10.0 schema. It intentionally does not
// follow the current Village API version: successor authoritative multipart
// metadata is validated by DecodeAuthoritativePublishRequest. A non-nil error means the body is
// invalid — a type/enum/pattern violation, the rc2 (#118) required model object or
// its harness/model fields missing (e.g. "missing properties: 'model'"/'harness'),
// or not valid JSON — and the caller must reject it (the village maps this to HTTP
// 422). It returns nil for a conforming body, or the compile error if the embedded
// schema fails to compile on first use.
func ValidatePublishRequest(data []byte) error {
	publishRequestSchemaOnce.Do(compilePublishRequestSchema)
	if publishRequestSchemaErr != nil {
		return publishRequestSchemaErr
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if err := publishRequestSchema.Validate(v); err != nil {
		return err
	}

	var request PublishRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return fmt.Errorf("decode publish request after generated schema validation: %w", err)
	}
	return request.Validate()
}
