package schema_test

import (
	"testing"

	"github.com/peasant-labs/schema"
)

// TestValidatePublishRequest_Corpus drives the ROOT client validator
// (schema.ValidatePublishRequest, over the version-aware PublishRequestSchemaJSON()
// byte-source) across the WHOLE shared publish-verdict corpus
// (testdata/publish/verdicts.yaml) via the shared schema.RunPublishVerdicts driver
// — ZERO inline case data. This proves the bytes the client validates against
// enforce the same corpus verdicts the village does (distinct from the parity test,
// which builds the schema in-memory). It is the corpus-driven replacement for the
// retired validate/validate_test.go's TestValidatePublishRequest_GeneratedSchemaVerdicts.
func TestValidatePublishRequest_Corpus(t *testing.T) {
	schema.RunPublishVerdicts(t, schema.ValidatePublishRequest)
}

// TestValidatePublishRequest_InvalidJSON subsumes the retired validate_test.go
// `_InvalidJSON` discrimination (AC5 behavior-preservation): a non-JSON body is
// rejected by the json.Unmarshal pre-pass in ValidatePublishRequest. This is kept
// as an INLINE validator test rather than a corpus row because the verdict corpus
// carries a "every row is a valid-JSON body with a schema verdict" invariant
// (TestPublishVerdictFixtures_Structure + the parity test both json.Unmarshal every
// row); malformed JSON is a validator-level concern (the schema can only be
// evaluated against parsed JSON), not a schema-verdict one — the same shape upstream
// peasant #118 chose.
func TestValidatePublishRequest_InvalidJSON(t *testing.T) {
	if err := schema.ValidatePublishRequest([]byte(`this is not json`)); err == nil {
		t.Error("invalid JSON should fail validation")
	}
}
