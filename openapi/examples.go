package openapi

import (
	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// exampleRef creates an ExampleOrReference with a summary and value for use in
// operation-level examples (request bodies, responses).
func exampleRef(summary string, value any) openapi31.ExampleOrReference {
	var eor openapi31.ExampleOrReference
	eor.WithExample(openapi31.Example{})
	eor.Example.WithSummary(summary)
	eor.Example.WithValue(value)
	return eor
}

// setResponseExample injects a named example into a path's response media type.
// method should be "get" or "post" (determines which Operation field to use).
func setResponseExample(spec *openapi31.Spec, path, method, statusCode, key, summary string, value any) {
	pi, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok {
		return
	}

	var op *openapi31.Operation
	switch method {
	case "get":
		op = pi.Get
	case "post":
		op = pi.Post
	default:
		return
	}
	if op == nil || op.Responses == nil {
		return
	}

	ror, ok := op.Responses.MapOfResponseOrReferenceValues[statusCode]
	if !ok || ror.Response == nil || ror.Response.Content == nil {
		return
	}

	mt, ok := ror.Response.Content["application/json"]
	if !ok {
		return
	}
	mt.WithExamplesItem(key, exampleRef(summary, value))
	ror.Response.Content["application/json"] = mt
	op.Responses.MapOfResponseOrReferenceValues[statusCode] = ror

	switch method {
	case "get":
		pi.Get = op
	case "post":
		pi.Post = op
	}
	spec.Paths.MapOfPathItemValues[path] = pi
}

// setRequestBodyExample injects a named example into a path's request body media type.
func setRequestBodyExample(spec *openapi31.Spec, path, method, key, summary string, value any) {
	pi, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok {
		return
	}

	var op *openapi31.Operation
	switch method {
	case "post":
		op = pi.Post
	default:
		return
	}
	if op == nil || op.RequestBody == nil || op.RequestBody.RequestBody == nil {
		return
	}

	rb := op.RequestBody.RequestBody
	mt, ok := rb.Content["application/json"]
	if !ok {
		return
	}
	mt.WithExamplesItem(key, exampleRef(summary, value))
	rb.Content["application/json"] = mt
	pi.Post = op
	spec.Paths.MapOfPathItemValues[path] = pi
}

// AddVillageExamples injects operation-level examples into the village API spec.
// Covers POST /api/v1/transcripts/publish request body and 200 response.
func AddVillageExamples(spec *openapi31.Spec) {
	const publishPath = "/api/v1/transcripts/publish"

	// Request body examples.
	setRequestBodyExample(spec, publishPath, "post", "minimal", "Minimal publish request", map[string]any{
		"identity": map[string]any{
			"sessionId":     "99d59925-36bc-424c-a789-8be54d9702ba",
			"schemaVersion": schema.MetadataSchemaVersion,
		},
		"model": map[string]any{
			"harness":  schema.HarnessClaudeCode.String(),
			"model":    "claude-opus-4-6",
			"hostSlug": "github.com--user--repo",
		},
		"timestamp": map[string]any{
			"start": 1709136000000,
			"end":   1709139600000,
		},
		"source": map[string]any{
			"format": schema.SourceFormatJSONL.String(),
		},
		"git": map[string]any{
			"branch": "main",
			"remote": "https://github.com/user/repo.git",
		},
		"project": map[string]any{
			"hash": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			"name": "repo",
		},
		"stats": map[string]any{
			"turnCount":     12,
			"toolCallCount": 8,
			"subagentCount": 0,
			"durationMs":    3600000,
			"tokensIn":      15000,
			"tokensOut":     8000,
		},
		"diagnostics": map[string]any{
			"warnings": []any{},
		},
	})

	setRequestBodyExample(spec, publishPath, "post", "full", "Full publish request with entries and quality", map[string]any{
		"identity": map[string]any{
			"sessionId":     "99d59925-36bc-424c-a789-8be54d9702ba",
			"schemaVersion": schema.MetadataSchemaVersion,
		},
		"model": map[string]any{
			"harness":  schema.HarnessClaudeCode.String(),
			"model":    "claude-opus-4-6",
			"version":  "2.1.47",
			"hostSlug": "github.com--user--repo",
		},
		"timestamp": map[string]any{
			"start": 1709136000000,
			"end":   1709139600000,
		},
		"source": map[string]any{
			"format": schema.SourceFormatJSONL.String(),
		},
		"git": map[string]any{
			"branch": "feat/openapi-docs",
			"remote": "https://github.com/user/repo.git",
		},
		"project": map[string]any{
			"hash": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			"name": "repo",
		},
		"stats": map[string]any{
			"turnCount":     24,
			"toolCallCount": 16,
			"subagentCount": 1,
			"durationMs":    7200000,
			"tokensIn":      45000,
			"tokensOut":     22000,
		},
		"quality": map[string]any{
			"turnCount":   24,
			"totalTokens": 67000,
			"outcome":     schema.OutcomeResolved.String(),
		},
		"entries": []any{
			map[string]any{
				"sessionId":   "99d59925-36bc-424c-a789-8be54d9702ba",
				"entryIndex":  0,
				"harness":     schema.HarnessClaudeCode.String(),
				"entryType":   schema.EntryTypeText.String(),
				"role":        schema.RoleUser.String(),
				"hasToolUse":  false,
				"hasThinking": false,
				"isError":     false,
				"depth":       0,
			},
		},
		"diagnostics": map[string]any{
			"warnings": []any{},
		},
	})

	// Response example.
	setResponseExample(spec, publishPath, "post", "200", "newTranscript", "Newly created transcript", map[string]any{
		"transcriptId":  "e7f8a9b0-1234-5678-9abc-def012345678",
		"blobKey":       "transcripts/2024/02/28/e7f8a9b0-1234-5678-9abc-def012345678.jsonl",
		"blobSizeBytes": 142857,
		"publishedAt":   1709139700000,
		"updatedAt":     1709139700000,
		"created":       true,
	})

	// --- Auth endpoints ---
	const exchangePath = "/api/v1/auth/cli/exchange"

	// POST /api/v1/auth/cli/exchange — request body examples.
	setRequestBodyExample(spec, exchangePath, "post", "codeExchange", "Exchange OAuth code for credentials", map[string]any{
		"code":  "a1b2c3d4e5f6g7h8i9j0",
		"state": "f47ac10b58cc4372a5670e02b2c3d479",
	})

	// POST /api/v1/auth/cli/exchange — response example.
	setResponseExample(spec, exchangePath, "post", "200", "credentials", "Successful credential exchange", map[string]any{
		"api_key":  "ak_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"key_id":   "key_01HQXYZ1234567890ABCDEF",
		"user_id":  "usr_01HQXYZ1234567890ABCDEF",
		"username": "alice",
	})
}

// AddPeasantLocalExamples injects operation-level response examples into the local dashboard API spec.
func AddPeasantLocalExamples(spec *openapi31.Spec) {
	// GET /api/v1/health
	setResponseExample(spec, "/api/v1/health", "get", "200", "healthy", "Healthy server", map[string]any{
		"status": "ok",
	})

	// GET /api/v1/sessions
	setResponseExample(spec, "/api/v1/sessions", "get", "200", "twoSessions", "Two session summaries", map[string]any{
		"sessions": []any{
			map[string]any{
				"id":            "99d59925-36bc-424c-a789-8be54d9702ba",
				"harness":       schema.HarnessClaudeCode.String(),
				"startTime":     "2024-02-28T10:00:00Z",
				"durationMins":  45.2,
				"totalTokens":   23000,
				"turnCount":     12,
				"toolCallCount": 8,
				"project":       "myapp",
			},
			map[string]any{
				"id":            "b2c3d4e5-6789-0abc-def1-234567890abc",
				"harness":       schema.HarnessOpenCode.String(),
				"startTime":     "2024-02-28T14:30:00Z",
				"durationMins":  22.7,
				"totalTokens":   11500,
				"turnCount":     6,
				"toolCallCount": 4,
				"project":       "another-project",
			},
		},
	})

	// GET /api/v1/sessions/{id}
	setResponseExample(spec, "/api/v1/sessions/{id}", "get", "200", "detailedSession", "Session with turns", map[string]any{
		"id":            "99d59925-36bc-424c-a789-8be54d9702ba",
		"harness":       schema.HarnessClaudeCode.String(),
		"startTime":     "2024-02-28T10:00:00Z",
		"endTime":       "2024-02-28T10:45:12Z",
		"durationMins":  45.2,
		"totalTokens":   23000,
		"tokensIn":      15000,
		"tokensOut":     8000,
		"turnCount":     2,
		"toolCallCount": 1,
		"turns": []any{
			map[string]any{
				"index":     0,
				"role":      schema.RoleUser.String(),
				"content":   "Add OpenAPI examples to the schema types",
				"timestamp": "2024-02-28T10:00:00Z",
				"depth":     0,
			},
			map[string]any{
				"index":     1,
				"role":      schema.RoleAssistant.String(),
				"content":   "I'll add WithExamples calls to each JSONSchema method...",
				"timestamp": "2024-02-28T10:00:15Z",
				"depth":     0,
				"toolCalls": []any{
					map[string]any{
						"id":        "tool_01",
						"name":      "Read",
						"arguments": "{\"path\":\"pkg/schema/types.go\"}",
						"result":    "file contents...",
					},
				},
			},
		},
	})

	// GET /api/v1/config/mock
	setResponseExample(spec, "/api/v1/config/mock", "get", "200", "mocksEnabled", "Mock data enabled for web", map[string]any{
		"enabled": true,
		"web":     []any{"dashboard", "sessions", "trends"},
	})

	// POST /api/v1/shutdown
	setResponseExample(spec, "/api/v1/shutdown", "post", "200", "shuttingDown", "Server shutting down", map[string]any{
		"status": "shutting_down",
	})
}
