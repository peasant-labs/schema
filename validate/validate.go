package validate

import (
	"bytes"
	"encoding/json"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"

	schema "github.com/peasant-labs/schema"
)

// schemaURL is the in-compiler resource name for the PublishRequest schema. It is
// an opaque handle (not fetched over the network), so any stable string works.
const schemaURL = "publish-request.schema.json"

var (
	compiled   *jsonschema.Schema
	compileErr error
)

// init compiles the SINGLE BYTE-SOURCE PublishRequest JSON-Schema — the generated
// artifact returned by schema.PublishRequestSchemaJSON() (extracted from the
// Village API OpenAPI spec by openapi.BuildPublishRequestSchema and committed under
// generated/). It deliberately reads the SAME bytes the village's documented spec
// is derived from, so the enforced schema and the published contract can never
// drift.
//
// This replaced the former hand-maintained validate/schema.json, which had
// diverged from the generated artifact (urn:...:1.0 with a `modelHarness` key and
// no harness-enum tightening on `model.harness`) and would have silently changed
// the village's 422 verdicts on re-pin.
func init() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaURL, bytes.NewReader(schema.PublishRequestSchemaJSON())); err != nil {
		compileErr = err
		return
	}
	compiled, compileErr = compiler.Compile(schemaURL)
}

// ValidatePublishRequest validates a publish-request JSON body against the
// generated PublishRequest JSON-Schema. A non-nil error means the body is invalid
// (a type/enum/pattern violation, or not valid JSON) and the caller must reject it
// (the village maps this to HTTP 422). It returns nil for a conforming body.
func ValidatePublishRequest(data []byte) error {
	if compileErr != nil {
		return compileErr
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	return compiled.Validate(v)
}
