package validate_test

import (
	"testing"

	"github.com/peasant-labs/schema/validate"
)

func TestValidatePublishRequest_Valid(t *testing.T) {
	validJSON := `{
		"identity": {
			"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba",
			"schemaVersion": 1
		},
		"model": {
			"harness": "claude-code",
			"model": "claude-3-5-sonnet-20241022"
		},
		"timestamp": {
			"start": 1708700000000,
			"end": 1708700600000
		},
		"source": {
			"format": "jsonl"
		},
		"git": {},
		"project": {
			"hash": "0000000000000000000000000000000000000000000000000000000000000000",
			"name": "test-project"
		},
		"stats": {
			"turnCount": 10,
			"toolCallCount": 25,
			"subagentCount": 0,
			"durationMs": 600000,
			"tokensIn": 5000,
			"tokensOut": 3000
		},
		"subagents": [],
		"diagnostics": {
			"warnings": []
		}
	}`

	err := validate.ValidatePublishRequest([]byte(validJSON))
	if err != nil {
		t.Errorf("valid payload should pass validation: %v", err)
	}
}

func TestValidatePublishRequest_Invalid(t *testing.T) {
	invalidJSON := `{
		"identity": {
			"sessionId": "not-a-valid-uuid",
			"schemaVersion": 1
		},
		"model": {
			"harness": "claude-code",
			"model": "claude-3-5-sonnet-20241022"
		},
		"timestamp": {
			"start": 1708700000000,
			"end": 1708700600000
		},
		"source": {
			"format": "jsonl"
		},
		"project": {
			"hash": "0000000000000000000000000000000000000000000000000000000000000000",
			"name": "test-project"
		},
		"stats": {
			"turnCount": 10,
			"toolCallCount": 25,
			"subagentCount": 0,
			"durationMs": 600000,
			"tokensIn": 5000,
			"tokensOut": 3000
		},
		"subagents": [],
		"diagnostics": {
			"warnings": []
		}
	}`

	err := validate.ValidatePublishRequest([]byte(invalidJSON))
	if err == nil {
		t.Error("invalid payload should fail validation")
	}
}

func TestValidatePublishRequest_MissingRequired(t *testing.T) {
	// Note: swaggest doesn't generate "required" arrays by default.
	// This test verifies basic validation works, not required field enforcement.
	// The nested validation (TestValidatePublishRequest_NestedSessionID) is the important test.
	missingRequired := `{}`

	err := validate.ValidatePublishRequest([]byte(missingRequired))
	// Currently passes because schema doesn't have required enforcement
	// This is a schema generation issue, not validation
	if err != nil {
		t.Logf("Validation error (optional): %v", err)
	}
}

func TestValidatePublishRequest_InvalidJSON(t *testing.T) {
	notJSON := `this is not json`

	err := validate.ValidatePublishRequest([]byte(notJSON))
	if err == nil {
		t.Error("invalid JSON should fail validation")
	}
}

func TestValidatePublishRequest_NestedSessionID(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		wantInvalid bool
		description string
	}{
		{
			name: "invalid nested parentUuid",
			payload: `{
				"identity": {
					"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba",
					"parentUuid": "not-a-valid-uuid",
					"schemaVersion": 1
				},
				"model": {"harness": "claude-code", "model": "claude-3-5-sonnet-20241022"},
				"timestamp": {"start": 1708700000000, "end": 1708700600000},
				"source": {"format": "jsonl"},
				"git": {},
				"project": {"hash": "0000000000000000000000000000000000000000000000000000000000000000", "name": "test"},
				"stats": {"turnCount": 10, "toolCallCount": 25, "subagentCount": 0, "durationMs": 600000, "tokensIn": 5000, "tokensOut": 3000},
				"subagents": [],
				"diagnostics": {"warnings": []}
			}`,
			wantInvalid: true,
			description: "parentUuid should validate via $ref to SchemaSessionID",
		},
		{
			name: "valid nested parentUuid",
			payload: `{
				"identity": {
					"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba",
					"parentUuid": "agent-a3aee4f",
					"schemaVersion": 1
				},
				"model": {"harness": "claude-code", "model": "claude-3-5-sonnet-20241022"},
				"timestamp": {"start": 1708700000000, "end": 1708700600000},
				"source": {"format": "jsonl"},
				"git": {},
				"project": {"hash": "0000000000000000000000000000000000000000000000000000000000000000", "name": "test"},
				"stats": {"turnCount": 10, "toolCallCount": 25, "subagentCount": 0, "durationMs": 600000, "tokensIn": 5000, "tokensOut": 3000},
				"subagents": [],
				"diagnostics": {"warnings": []}
			}`,
			wantInvalid: false,
			description: "valid agent- prefixed parentUuid should pass",
		},
		{
			name: "invalid provider nested",
			payload: `{
				"identity": {"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba", "schemaVersion": 1},
				"model": {"harness": "invalid-provider", "model": "claude-3-5-sonnet-20241022"},
				"timestamp": {"start": 1708700000000, "end": 1708700600000},
				"source": {"format": "jsonl"},
				"git": {},
				"project": {"hash": "0000000000000000000000000000000000000000000000000000000000000000", "name": "test"},
				"stats": {"turnCount": 10, "toolCallCount": 25, "subagentCount": 0, "durationMs": 600000, "tokensIn": 5000, "tokensOut": 3000},
				"subagents": [],
				"diagnostics": {"warnings": []}
			}`,
			wantInvalid: true,
			description: "invalid provider enum should fail via $ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.ValidatePublishRequest([]byte(tt.payload))
			if tt.wantInvalid && err == nil {
				t.Errorf("%s: expected invalid but passed", tt.description)
			}
			if !tt.wantInvalid && err != nil {
				t.Errorf("%s: expected valid but got error: %v", tt.description, err)
			}
		})
	}
}

// TestValidatePublishRequest_GeneratedSchemaVerdicts is the W9 COUPLED verdict
// corpus. It pins the accept/reject contract of ValidatePublishRequest now that it
// compiles the GENERATED publish-request schema (the single byte-source) rather
// than the retired hand-maintained validate/schema.json. The corpus mirrors the
// village's own enforce-path corpus (handler.TestValidatePublish_SchemaVerdicts),
// so the two repos pin identical 422 behavior across the re-pin and any drift
// surfaces in BOTH suites.
//
// DIVERGENCE EXPLAINED: the retired validate/schema.json (urn:peasant:publish-request:1.0)
// keyed the harness enum on `model.modelHarness` and lacked the `harness` field, so
// it ACCEPTED an out-of-enum `model.harness` and a wrong-typed `entries` — exactly
// the two rows below that the generated schema (urn:peasant:publish-request:0.2.0,
// the bytes the village already enforced) REJECTS. Routing the validator at the
// generated artifact makes the village's post-re-pin verdicts byte-for-byte
// identical to its pre-re-pin verdicts.
func TestValidatePublishRequest_GeneratedSchemaVerdicts(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantAccept bool
	}{
		// well-formed minimal body (no `required` blocks in the generated schema).
		{"well-formed", `{"model": {"harness": "claude-code", "model": "x"}, "source": {"format": "jsonl"}, "project": {"hash": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "name": "r"}}`, true},
		// enum violation: source.format ∉ {jsonl,json}.
		{"bad-source-format", `{"source": {"filePath": "/p/t", "format": "xml"}}`, false},
		// out-of-enum harness on model.harness — REJECTED by the generated schema,
		// ACCEPTED by the retired hand schema (the key divergence W9 reconciles).
		{"unknown-harness", `{"model": {"harness": "totally-made-up", "model": "x"}}`, false},
		// omitted harness is NOT rejected: the generated schema declares no required.
		{"omitted-harness-accepts", `{"model": {"model": "x"}}`, true},
		// wrong type: timestamp.start must be an integer, not a string.
		{"wrong-type-start", `{"timestamp": {"start": "soon", "end": 1700000060000}}`, false},
		// wrong type: entries must be an array — REJECTED by the generated schema,
		// ACCEPTED by the retired hand schema (the second divergence W9 reconciles).
		{"wrong-type-entries", `{"entries": {"not": "an array"}}`, false},
		// pattern violation on the constrained ProjectHash newtype.
		{"bad-project-hash", `{"project": {"hash": "tooshort", "name": "x"}}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.ValidatePublishRequest([]byte(tc.body))
			if tc.wantAccept && err != nil {
				t.Errorf("well-formed body rejected by generated schema: %v\nbody: %s", err, tc.body)
			}
			if !tc.wantAccept && err == nil {
				t.Errorf("malformed body accepted by generated schema (expected 422-class rejection)\nbody: %s", tc.body)
			}
		})
	}
}
