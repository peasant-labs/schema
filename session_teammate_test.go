package schema_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// TestSessionTeammatePublishedEvidenceComponent proves the generated Types
// document publishes the teammate evidence as a reference to its own one-value
// closed set. An open string would accept every evidence kind a producer sends,
// including the shared session graph kinds this local set deliberately excludes,
// and would leave the Go predicate as the only thing refusing them.
func TestSessionTeammatePublishedEvidenceComponent(t *testing.T) {
	fixtures, err := schema.LoadSessionTeammateFixtures()
	if err != nil {
		t.Fatalf("LoadSessionTeammateFixtures: %v", err)
	}

	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec: %v", err)
	}
	teammate, ok := spec.Components.Schemas[fixtures.Component]
	if !ok {
		t.Fatalf("%s is absent from the Types catalog; the identity object would degrade to an unnamed shape in every generated binding", fixtures.Component)
	}
	properties, ok := teammate["properties"].(map[string]any)
	if !ok {
		t.Fatalf("%s declares no properties map", fixtures.Component)
	}
	evidenceField, ok := properties[fixtures.EvidenceField].(map[string]any)
	if !ok {
		t.Fatalf("%s.%s is absent; the evidence kind would leave the published contract", fixtures.Component, fixtures.EvidenceField)
	}
	if open, present := evidenceField["type"]; present && open != nil {
		t.Fatalf("%s.%s declares a bare type %v instead of referencing %s; an open string accepts every evidence kind", fixtures.Component, fixtures.EvidenceField, open, fixtures.EvidenceRef)
	}
	if got := evidenceField["$ref"]; got != fixtures.EvidenceRef {
		t.Fatalf("%s.%s $ref = %v, want %q; the published field must reference the closed set, not describe an open string", fixtures.Component, fixtures.EvidenceField, got, fixtures.EvidenceRef)
	}

	evidenceComponent, ok := spec.Components.Schemas[fixtures.EvidenceComponent]
	if !ok {
		t.Fatalf("%s is absent from the Types catalog; the reference would dangle and generated bindings would fall back to a plain string", fixtures.EvidenceComponent)
	}
	published := make(map[string]struct{})
	rawEnum, ok := evidenceComponent["enum"].([]any)
	if !ok {
		t.Fatalf("%s publishes no enum; a referenced component without an enum is still an open string", fixtures.EvidenceComponent)
	}
	for _, raw := range rawEnum {
		token, ok := raw.(string)
		if !ok {
			t.Fatalf("%s enum carries a non-string member %v", fixtures.EvidenceComponent, raw)
		}
		published[token] = struct{}{}
	}
	for _, want := range fixtures.RequiredMenu {
		if _, ok := published[want.String()]; !ok {
			t.Fatalf("%s does not publish %q; every declared teammate evidence member must reach the generated binding", fixtures.EvidenceComponent, want)
		}
	}
	if len(published) != len(fixtures.RequiredMenu) {
		t.Fatalf("%s publishes %d members, but the corpus declares %d; an extra member would accept an evidence kind no recorded spawn produces", fixtures.EvidenceComponent, len(published), len(fixtures.RequiredMenu))
	}

	// The identity rules are proven on the published object itself: its bound
	// must be the Go constant, not a duplicated literal, and its required set
	// must name every member a teammate object cannot omit.
	if want := fixtures.IdentityMaxLength; want != schema.SessionTeammateMaxIdentityRunes {
		t.Fatalf("identity_max_length is %d, but schema.SessionTeammateMaxIdentityRunes is %d; the corpus pins the published bound to the Go rule", want, schema.SessionTeammateMaxIdentityRunes)
	}
	for _, field := range fixtures.IdentityFields {
		property, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("%s.%s is absent; the identity bound would apply to no published property", fixtures.Component, field)
		}
		if got := schemaInteger(t, fixtures.Component+"."+field, property, "minLength"); got != 1 {
			t.Fatalf("%s.%s minLength = %d, want 1; an empty identity cannot be matched to a recorded spawn", fixtures.Component, field, got)
		}
		if got := schemaInteger(t, fixtures.Component+"."+field, property, "maxLength"); got != schema.SessionTeammateMaxIdentityRunes {
			t.Fatalf("%s.%s maxLength = %d, but schema.SessionTeammateMaxIdentityRunes is %d; the published bound must be pinned to the Go rule, not restated as a literal that can drift", fixtures.Component, field, got, schema.SessionTeammateMaxIdentityRunes)
		}
	}
	required := schemaStringList(t, fixtures.Component, teammate, "required")
	for _, name := range fixtures.IdentityRequired {
		if _, ok := required[name]; !ok {
			t.Fatalf("%s does not require %q; a generated binding would treat a member the identity cannot omit as optional", fixtures.Component, name)
		}
		if _, ok := properties[name]; !ok {
			t.Fatalf("%s requires %q but publishes no such property", fixtures.Component, name)
		}
	}

	// Compile the whole Types document so the evidence $ref resolves exactly as
	// a consumer's validator resolves it, then validate each fragment.
	document, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Types document: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("types.json", bytes.NewReader(document)); err != nil {
		t.Fatalf("add Types document resource: %v", err)
	}
	evidenceSchema := compilePublishedComponent(t, compiler, "types.json", fixtures.EvidenceComponent)
	identitySchema := compilePublishedComponent(t, compiler, "types.json", fixtures.Component)

	for _, c := range fixtures.SchemaProbes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requirePublishedVerdict(t, evidenceSchema, c.Input.Evidence, c.Expected.Accepted, c.Expected.ErrorContains)
		})
	}
	for _, c := range fixtures.IdentityProbes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			instance, err := json.Marshal(map[string]any{
				"team":     c.Input.Team,
				"name":     c.Input.Name,
				"evidence": c.Input.Evidence,
			})
			if err != nil {
				t.Fatalf("marshal identity probe: %v", err)
			}
			var value any
			if err := json.Unmarshal(instance, &value); err != nil {
				t.Fatalf("unmarshal identity probe: %v", err)
			}
			requirePublishedVerdict(t, identitySchema, value, c.Expected.Accepted, c.Expected.ErrorContains)
		})
	}
}

// compilePublishedComponent compiles one named component of a published
// document, so a $ref to a sibling component resolves as a consumer's validator
// would resolve it.
func compilePublishedComponent(t *testing.T, compiler *jsonschema.Compiler, document, component string) *jsonschema.Schema {
	t.Helper()
	compiled, err := compiler.Compile(document + "#/components/schemas/" + component)
	if err != nil {
		t.Fatalf("compile %s from %s: %v", component, document, err)
	}
	return compiled
}

// requirePublishedVerdict runs one fixture-driven probe against a compiled
// component and reports the verdict it actually gave.
func requirePublishedVerdict(t *testing.T, compiled *jsonschema.Schema, instance any, accepted bool, errorContains string) {
	t.Helper()
	verr := compiled.Validate(instance)
	if accepted {
		if verr != nil {
			t.Fatalf("published component refused %v: %v", instance, verr)
		}
		return
	}
	if verr == nil {
		t.Fatalf("published component accepted %v, but the row expects a refusal naming %q", instance, errorContains)
	}
	if !strings.Contains(verr.Error(), errorContains) {
		t.Fatalf("refusal message does not contain the pinned fragment.\n  want substring: %q\n  got message:    %s\n  fix: re-pin error_contains to the message the validator really emits (do not guess).", errorContains, verr.Error())
	}
}

// schemaInteger reads one integer keyword from a published property.
func schemaInteger(t *testing.T, where string, property map[string]any, keyword string) int {
	t.Helper()
	raw, ok := property[keyword]
	if !ok {
		t.Fatalf("%s publishes no %s", where, keyword)
	}
	switch value := raw.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		t.Fatalf("%s %s is %T, want an integer", where, keyword, raw)
		return 0
	}
}

// schemaStringList reads a published string-list keyword such as required.
func schemaStringList(t *testing.T, component string, schemaMap map[string]any, keyword string) map[string]struct{} {
	t.Helper()
	values := map[string]struct{}{}
	switch raw := schemaMap[keyword].(type) {
	case []any:
		for _, entry := range raw {
			name, ok := entry.(string)
			if !ok {
				t.Fatalf("%s %s carries a non-string member %v", component, keyword, entry)
			}
			values[name] = struct{}{}
		}
	case []string:
		for _, name := range raw {
			values[name] = struct{}{}
		}
	case nil:
		t.Fatalf("%s publishes no %s", component, keyword)
	default:
		t.Fatalf("%s %s is %T, want a string list", component, keyword, raw)
	}
	return values
}
