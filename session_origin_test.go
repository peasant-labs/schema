package schema_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

func loadSessionOriginFixtures(t *testing.T) *schema.SessionOriginFixtures {
	t.Helper()
	fixtures, err := schema.LoadSessionOriginFixtures()
	if err != nil {
		t.Fatalf("LoadSessionOriginFixtures: %v", err)
	}
	return fixtures
}

// TestSessionOriginValidity drives the wire predicate over the corpus. The menu
// is derived from AllSessionOrigins by the loader, so widening the enum without
// a row fails before this test runs.
func TestSessionOriginValidity(t *testing.T) {
	fixtures := loadSessionOriginFixtures(t)
	for _, c := range fixtures.Validities.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := schema.SessionOrigin(c.Input).IsValid()
			if got != c.Expected.Valid {
				t.Fatalf("SessionOrigin(%q).IsValid() = %t, want %t (%s)", c.Input, got, c.Expected.Valid, c.Expected.Behaviour)
			}
			if got != (c.Classification == testcase.MustPass) {
				t.Fatalf("case %q classification %q disagrees with the verdict %t", c.Name, c.Classification, got)
			}
			if s := schema.SessionOrigin(c.Input).String(); s != c.Input {
				t.Fatalf("SessionOrigin(%q).String() = %q, want the token unchanged", c.Input, s)
			}
		})
	}
}

// TestSessionOriginRoundTrips proves the field independently on BOTH payloads:
// the two placements are separate contract surfaces, and a single assertion
// would be satisfied by one of them being right while the other was wrong.
func TestSessionOriginRoundTrips(t *testing.T) {
	fixtures := loadSessionOriginFixtures(t)
	for _, c := range fixtures.RoundTrips.Cases {
		t.Run(c.Name, func(t *testing.T) {
			encoded, decoded := roundTripSessionOrigin(t, c.Input.Payload, schema.SessionOrigin(c.Input.Origin))
			if decoded != schema.SessionOrigin(c.Expected.Decoded) {
				t.Fatalf("decoded sessionOrigin = %q, want %q", decoded, c.Expected.Decoded)
			}
			present := bytes.Contains(encoded, []byte(`"sessionOrigin"`))
			if present != c.Expected.FieldPresent {
				t.Fatalf("sessionOrigin present on the wire = %t, want %t; encoded body: %s", present, c.Expected.FieldPresent, encoded)
			}
			if c.Expected.EncodedJSON != "" && !bytes.Contains(encoded, []byte(c.Expected.EncodedJSON)) {
				t.Fatalf("encoded body does not contain %q; body: %s", c.Expected.EncodedJSON, encoded)
			}
		})
	}
}

// roundTripSessionOrigin encodes the named payload carrying origin, then decodes
// it back, returning the wire bytes and the recovered declaration.
func roundTripSessionOrigin(t *testing.T, payload schema.SessionOriginPayloadKind, origin schema.SessionOrigin) ([]byte, schema.SessionOrigin) {
	t.Helper()
	switch payload {
	case schema.SessionOriginPayloadDetail:
		encoded, err := json.Marshal(schema.SessionDetailPayload{ID: "s-1", SessionOrigin: origin})
		if err != nil {
			t.Fatalf("marshal SessionDetailPayload: %v", err)
		}
		var back schema.SessionDetailPayload
		if err := json.Unmarshal(encoded, &back); err != nil {
			t.Fatalf("unmarshal SessionDetailPayload: %v", err)
		}
		return encoded, back.SessionOrigin
	case schema.SessionOriginPayloadSummary:
		encoded, err := json.Marshal(schema.SessionSummary{ID: "s-1", SessionOrigin: origin})
		if err != nil {
			t.Fatalf("marshal SessionSummary: %v", err)
		}
		var back schema.SessionSummary
		if err := json.Unmarshal(encoded, &back); err != nil {
			t.Fatalf("unmarshal SessionSummary: %v", err)
		}
		return encoded, back.SessionOrigin
	}
	t.Fatalf("payload kind %q has no round-trip implementation", payload)
	return nil, ""
}

// TestSessionOriginPublishedComponent offers each probe to the SessionOrigin
// component of the generated Types document, so a non-Go consumer inherits the
// same closed set the Go predicate enforces.
func TestSessionOriginPublishedComponent(t *testing.T) {
	fixtures := loadSessionOriginFixtures(t)
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec: %v", err)
	}
	component, ok := spec.Components.Schemas["SessionOrigin"]
	if !ok {
		t.Fatal("SessionOrigin is absent from the Types catalog; the closed set would degrade to an open string in every generated binding")
	}
	raw, err := json.Marshal(component)
	if err != nil {
		t.Fatalf("marshal SessionOrigin component: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("session-origin.json", bytes.NewReader(raw)); err != nil {
		t.Fatalf("add SessionOrigin resource: %v", err)
	}
	compiled, err := compiler.Compile("session-origin.json")
	if err != nil {
		t.Fatalf("compile SessionOrigin component: %v", err)
	}

	for _, c := range fixtures.SchemaProbes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			verr := compiled.Validate(c.Input.Origin)
			if c.Expected.Accepted {
				if verr != nil {
					t.Fatalf("published component refused %q: %v", c.Input.Origin, verr)
				}
				return
			}
			if verr == nil {
				t.Fatalf("published component accepted %q, but the row expects a refusal naming %q", c.Input.Origin, c.Expected.ErrorContains)
			}
			if !strings.Contains(verr.Error(), c.Expected.ErrorContains) {
				t.Fatalf("refusal message does not contain the pinned fragment.\n  want substring: %q\n  got message:    %s\n  fix: re-pin error_contains to the message the validator really emits (do not guess).", c.Expected.ErrorContains, verr.Error())
			}
		})
	}
}

// TestSessionOriginPlacementsAreTyped guards the deliberate decision that BOTH
// placements take the typed enum. SessionSummary.Outcome is a plain string while
// SessionDetailPayload.Outcome is typed; that asymmetry is not repeated here, and
// a future edit that re-introduced it would silently drop the closed set from one
// of the two surfaces.
func TestSessionOriginPlacementsAreTyped(t *testing.T) {
	var detail schema.SessionDetailPayload
	var summary schema.SessionSummary
	detail.SessionOrigin = schema.SessionOriginAgent
	summary.SessionOrigin = schema.SessionOriginAgent
	if detail.SessionOrigin != summary.SessionOrigin {
		t.Fatalf("the two placements disagree: detail=%q summary=%q", detail.SessionOrigin, summary.SessionOrigin)
	}
}

// TestSessionSummariesByIDOperation asserts the by-id link-resolution operation
// is declared as the corpus says, including every stated scope clause. The
// clauses are the contract: a consumer reads them to know that a session hidden
// from every list still resolves through this route.
func TestSessionSummariesByIDOperation(t *testing.T) {
	fixtures := loadSessionOriginFixtures(t)
	contract := fixtures.ByIDOperation

	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Local API spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal Local API spec: %v", err)
	}

	operation := nestedMap(t, document, "paths", contract.Path, contract.Method)
	if operation["operationId"] != contract.OperationID {
		t.Fatalf("operationId = %v, want %q", operation["operationId"], contract.OperationID)
	}

	description, _ := operation["description"].(string)
	if description == "" {
		t.Fatal("the by-id operation carries no description; its scope clauses are the contract a consumer relies on")
	}
	for _, anchor := range contract.DescriptionAnchors {
		if !strings.Contains(description, anchor.Text) {
			t.Errorf("the by-id operation description does not state %q.\n  want substring: %q\n  got: %s", anchor.Name, anchor.Text, description)
		}
	}

	parameters, ok := operation["parameters"].([]any)
	if !ok || len(parameters) == 0 {
		t.Fatalf("the by-id operation declares no parameters; the %q identifier list would be undocumented", contract.QueryParameter)
	}
	found := false
	for _, raw := range parameters {
		parameter, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if parameter["name"] == contract.QueryParameter && parameter["in"] == "query" {
			found = true
			if parameter["required"] != true {
				t.Errorf("query parameter %q is not required; an unfiltered call would return the whole list from a link-resolution route", contract.QueryParameter)
			}
		}
	}
	if !found {
		t.Errorf("the by-id operation declares no %q query parameter", contract.QueryParameter)
	}

	media := nestedMap(t, operation, "responses", contract.ResponseStatus, "content", contract.MediaType)
	responseSchema, ok := media["schema"].(map[string]any)
	if !ok {
		t.Fatalf("the by-id operation's %s response declares no schema", contract.ResponseStatus)
	}
	if responseSchema["$ref"] != contract.ResponseRef {
		t.Fatalf("response ref = %v, want %q", responseSchema["$ref"], contract.ResponseRef)
	}
}

// nestedMap walks a decoded JSON document, failing with the path it could not
// follow rather than panicking on a type assertion.
func nestedMap(t *testing.T, document map[string]any, keys ...string) map[string]any {
	t.Helper()
	current := document
	for i, key := range keys {
		next, ok := current[key]
		if !ok {
			t.Fatalf("document has no %q at path %v", key, keys[:i+1])
		}
		asMap, ok := next.(map[string]any)
		if !ok {
			t.Fatalf("document value at path %v is %T, want an object", keys[:i+1], next)
		}
		current = asMap
	}
	return current
}
