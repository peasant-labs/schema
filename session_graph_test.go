package schema

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	validator "github.com/santhosh-tekuri/jsonschema/v5"
	jsonschema "github.com/swaggest/jsonschema-go"
)

func TestSessionGraphPrimitiveCorpus(t *testing.T) {
	c, err := LoadSessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireMin(t, c.Refs, 15)
	caseassert.RequireValid(t, c.Refs)
	caseassert.RequireMin(t, c.Enums, 13)
	caseassert.RequireValid(t, c.Enums)
	caseassert.RequireMin(t, c.Semantics, 20)
	caseassert.RequireValid(t, c.Semantics)
	requireNames(t, c.Refs.Cases, []string{"source-empty", "submission-invalid-utf8", "revision-one-byte", "source-96-multibyte", "submission-97-multibyte", "revision-preserves-space"})
	for _, x := range c.Refs.Cases {
		got, e := ConstructSessionGraphRef(x.Input)
		assertFixtureError(t, x.Name, x.Classification, x.Expected.ErrorContains, e)
		if e == nil {
			want, _ := base64.StdEncoding.DecodeString(x.Expected.ValueBase64)
			if got != string(want) {
				t.Errorf("%s returned bytes %x want %x", x.Name, []byte(got), want)
			}
			encoded, _ := json.Marshal(got)
			var round string
			if json.Unmarshal(encoded, &round) != nil || round != got {
				t.Errorf("%s JSON roundtrip changed bytes", x.Name)
			}
			assertRefSchema(t, x.Input.Alias, got, true)
		} else if raw, e2 := base64.StdEncoding.DecodeString(x.Input.BytesBase64); e2 == nil && json.Valid([]byte(fmt.Sprintf("%q", string(raw)))) {
			assertRefSchema(t, x.Input.Alias, string(raw), false)
		}
	}
	for _, x := range c.Enums.Cases {
		testEnumFixture(t, x.Input)
	}
	required := []string{"relationship-known-target", "relationship-unknown-forbids-target", "relationship-unique-kinds", "anchor-general", "anchor-exact", "anchor-partial-rejected", "provenance-all-unknown", "provenance-empty-rejected", "navigation-local-resolved", "navigation-unavailable-no-id", "navigation-unavailable-leaks-id", "helper-group-valid", "helper-group-wrong-purpose", "helper-context-valid", "earlier-empty-valid", "earlier-null-invalid"}
	requireNames(t, c.Semantics.Cases, required)
	for _, x := range c.Semantics.Cases {
		e := ValidateSessionGraphFixtureInput(x.Input)
		assertFixtureError(t, x.Name, x.Classification, x.Expected.ErrorContains, e)
	}
}

func assertFixtureError(t *testing.T, name string, class testcase.Classification, needle string, err error) {
	t.Helper()
	if class == testcase.MustPass && err != nil {
		t.Errorf("%s: unexpected error: %v", name, err)
	}
	if class == testcase.MustFail && (err == nil || needle == "" || !strings.Contains(err.Error(), needle)) {
		t.Errorf("%s: error=%v want meaningful contains %q", name, err, needle)
	}
}
func requireNames[I, E any](t *testing.T, cases []testcase.Case[I, E], required []string) {
	t.Helper()
	names := map[string]bool{}
	for _, x := range cases {
		names[x.Name] = true
	}
	for _, n := range required {
		if !names[n] {
			t.Errorf("required fixture %q missing", n)
		}
	}
}

func assertRefSchema(t *testing.T, alias, value string, wantValid bool) {
	t.Helper()
	var s jsonschema.Schema
	var e error
	switch alias {
	case "source":
		s, e = SourceEntryRef("").JSONSchema()
	case "submission":
		s, e = SubmissionRef("").JSONSchema()
	case "revision":
		s, e = PublicRevisionRef("").JSONSchema()
	default:
		t.Fatalf("unknown alias %q", alias)
	}
	if e != nil {
		t.Fatal(e)
	}
	data, _ := json.Marshal(s)
	old, had := validator.Formats[PublicRefUTF8ByteFormat]
	validator.Formats[PublicRefUTF8ByteFormat] = ValidatePublicRefJSONSchemaFormat
	defer func() {
		if had {
			validator.Formats[PublicRefUTF8ByteFormat] = old
		} else {
			delete(validator.Formats, PublicRefUTF8ByteFormat)
		}
	}()
	compiler := validator.NewCompiler()
	compiler.AssertFormat = true
	if e = compiler.AddResource("schema.json", strings.NewReader(string(data))); e != nil {
		t.Fatal(e)
	}
	compiled, e := compiler.Compile("schema.json")
	if e != nil {
		t.Fatal(e)
	}
	e = compiled.Validate(value)
	if (e == nil) != wantValid {
		t.Errorf("%s schema validation=%v want valid=%v schema=%s", alias, e, wantValid, data)
	}
}

type enumAPI struct {
	inventory []string
	construct func(string) error
	valid     func(string) bool
	schema    func() (jsonschema.Schema, error)
}

func enumFor(name string) enumAPI {
	switch name {
	case "relationship":
		return makeEnumAPI(AllSessionRelationshipKinds, func(v string) error { _, e := NewSessionRelationshipKind(v); return e }, func(v string) bool { return SessionRelationshipKind(v).IsValid() }, SessionRelationshipKind("").JSONSchema)
	case "target":
		return makeEnumAPI(AllRelationshipTargetStates, func(v string) error { _, e := NewRelationshipTargetState(v); return e }, func(v string) bool { return RelationshipTargetState(v).IsValid() }, RelationshipTargetState("").JSONSchema)
	case "evidence":
		return makeEnumAPI(AllEvidenceKinds, func(v string) error { _, e := NewEvidenceKind(v); return e }, func(v string) bool { return EvidenceKind(v).IsValid() }, EvidenceKind("").JSONSchema)
	case "purpose":
		return makeEnumAPI(AllSessionPurposes, func(v string) error { _, e := NewSessionPurpose(v); return e }, func(v string) bool { return SessionPurpose(v).IsValid() }, SessionPurpose("").JSONSchema)
	case "origin":
		return makeEnumAPI(AllContentOrigins, func(v string) error { _, e := NewContentOrigin(v); return e }, func(v string) bool { return ContentOrigin(v).IsValid() }, ContentOrigin("").JSONSchema)
	case "actor":
		return makeEnumAPI(AllActorOrigins, func(v string) error { _, e := NewActorOrigin(v); return e }, func(v string) bool { return ActorOrigin(v).IsValid() }, ActorOrigin("").JSONSchema)
	case "delivery":
		return makeEnumAPI(AllDeliveryOrigins, func(v string) error { _, e := NewDeliveryOrigin(v); return e }, func(v string) bool { return DeliveryOrigin(v).IsValid() }, DeliveryOrigin("").JSONSchema)
	case "ownership":
		return makeEnumAPI(AllContentOwnerships, func(v string) error { _, e := NewContentOwnership(v); return e }, func(v string) bool { return ContentOwnership(v).IsValid() }, ContentOwnership("").JSONSchema)
	case "modality":
		return makeEnumAPI(AllInputModalities, func(v string) error { _, e := NewInputModality(v); return e }, func(v string) bool { return InputModality(v).IsValid() }, InputModality("").JSONSchema)
	case "anchor":
		return makeEnumAPI(AllPublicSourceAnchorKinds, func(v string) error { _, e := NewPublicSourceAnchorKind(v); return e }, func(v string) bool { return PublicSourceAnchorKind(v).IsValid() }, PublicSourceAnchorKind("").JSONSchema)
	case "earlier":
		return makeEnumAPI(AllEarlierHistoryStates, func(v string) error { _, e := NewEarlierHistoryState(v); return e }, func(v string) bool { return EarlierHistoryState(v).IsValid() }, EarlierHistoryState("").JSONSchema)
	case "navigation":
		return makeEnumAPI(AllRelationshipNavigationStatuses, func(v string) error { _, e := NewRelationshipNavigationStatus(v); return e }, func(v string) bool { return RelationshipNavigationStatus(v).IsValid() }, RelationshipNavigationStatus("").JSONSchema)
	case "list_item":
		return makeEnumAPI(AllSessionListItemKinds, func(v string) error { _, e := NewSessionListItemKind(v); return e }, func(v string) bool { return SessionListItemKind(v).IsValid() }, SessionListItemKind("").JSONSchema)
	default:
		panic("unknown fixture enum " + name)
	}
}
func makeEnumAPI[T ~string](all []T, c func(string) error, v func(string) bool, s func() (jsonschema.Schema, error)) enumAPI {
	members := make([]string, len(all))
	for i, x := range all {
		members[i] = string(x)
	}
	return enumAPI{members, c, v, s}
}
func testEnumFixture(t *testing.T, i SessionGraphEnumInput) {
	t.Helper()
	api := enumFor(i.Enum)
	if !reflect.DeepEqual(api.inventory, i.Members) {
		t.Errorf("enum %s inventory=%v want %v", i.Enum, api.inventory, i.Members)
	}
	s, e := api.schema()
	if e != nil {
		t.Fatal(e)
	}
	data, _ := json.Marshal(s)
	compiler := validator.NewCompiler()
	if e = compiler.AddResource("enum.json", strings.NewReader(string(data))); e != nil {
		t.Fatal(e)
	}
	compiled, e := compiler.Compile("enum.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, member := range i.Members {
		if e := api.construct(member); e != nil || !api.valid(member) || compiled.Validate(member) != nil {
			t.Errorf("enum %s member %q constructor=%v valid=%v", i.Enum, member, e, api.valid(member))
		}
	}
	for _, bad := range []string{"", i.Unknown} {
		if api.construct(bad) == nil || compiled.Validate(bad) == nil {
			t.Errorf("enum %s accepted off-menu %q", i.Enum, bad)
		}
		if i.Enum == "purpose" && bad == "" {
			if !api.valid(bad) {
				t.Errorf("purpose omission should remain valid")
			}
		} else if api.valid(bad) {
			t.Errorf("enum %s IsValid accepted %q", i.Enum, bad)
		}
	}
}
