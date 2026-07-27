package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// publishRequestSchemaURL is the in-compiler resource name for the PublishRequest
// schema. It is an opaque handle (not fetched over the network), so any stable
// string works.
const publishRequestSchemaURL = "publish-request.schema.json"

var (
	publishRequestSchema     *jsonschema.Schema
	publishRequestSchemaErr  error
	publishRequestSchemaOnce sync.Once
)

// compilePublishRequestSchema compiles the SINGLE BYTE-SOURCE PublishRequest JSON-Schema returned by the
// version-aware accessor PublishRequestSchemaJSON() — the generated artifact
// committed under generated/publish-request-<VillageAPIVersion>.schema.json,
// extracted from the Village API spec by openapi.BuildPublishRequestSchema. It
// deliberately reads the SAME bytes the village's documented spec is derived from,
// so the enforced schema and the published contract can never drift.
//
// rc2 (#118): this validator moved from the retired `validate` subpackage into the
// root `schema` package — the symbol a re-pinned consumer expects
// (schema.ValidatePublishRequest, matching peasant pkg/schema@v1.4.0). It KEEPS the
// standalone's version-aware FUNCTION byte-source (PublishRequestSchemaJSON())
// rather than a literal //go:embed var: a VillageAPIVersion bump re-points it in
// lockstep with no embed path to hand-edit, structurally eliminating the staleness
// class #118's embed-guard test must police (TestPublishRequestSchemaJSON_IDMatchesVersion
// keeps a defense-in-depth $id assertion).
func compilePublishRequestSchema() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(publishRequestSchemaURL, bytes.NewReader(PublishRequestSchemaJSON())); err != nil {
		publishRequestSchemaErr = err
		return
	}
	publishRequestSchema, publishRequestSchemaErr = compiler.Compile(publishRequestSchemaURL)
}

// ValidatePublishRequest validates a publish-request JSON body against the
// generated PublishRequest JSON-Schema (the bytes PublishRequestSchemaJSON()
// returns for the current VillageAPIVersion). A non-nil error means the body is
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
