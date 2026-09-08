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
	requireNames(t, c.Refs.Cases, []string{"source-empty", "submission-invalid-utf8", "revision-one-byte", "source-96-multibyte", "submission-97-multibyte", "revision-preserves-space", "source-invalid-bytes", "source-one-bytes", "source-97-bytes", "submission-empty-bytes", "submission-one-bytes", "submission-96-bytes", "revision-empty-bytes", "revision-invalid-bytes", "revision-96-bytes", "revision-97-bytes"})
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
		} else if raw, e2 := base64.StdEncoding.DecodeString(x.Input.BytesBase64); e2 == nil {
			assertRefSchema(t, x.Input.Alias, string(raw), false)
		}
	}
	for _, x := range c.Enums.Cases {
		testEnumFixture(t, x.Input)
	}
	requireNames(t, c.Enums.Cases, []string{"enum-relationship", "enum-target", "enum-evidence", "enum-purpose", "enum-origin", "enum-actor", "enum-delivery", "enum-ownership", "enum-modality", "enum-anchor", "enum-earlier", "enum-navigation", "enum-list_item"})
	runPrimitiveArm(t, c.Relationships, 12, []string{"relationship-known-target", "relationship-unknown-forbids-target", "anchor-general", "anchor-exact", "anchor-partial-rejected", "relationship-known-missing-target", "relationship-known-invalid-target", "relationship-explicit-none-forbids-target", "relationship-conflict-forbids-target", "anchor-forbidden-started-by", "anchor-forbidden-nonknown", "anchor-general-forbids-refs"}, func(v SessionRelationship) error { return v.Validate() })
	runPrimitiveArm(t, c.RelationshipSets, 2, []string{"relationship-unique-kinds", "relationship-legal-unique-pair"}, func(v SessionRelationshipsInput) error { return ValidateSessionRelationships(v.Relationships) })
	runPrimitiveArm(t, c.Provenance, 10, []string{"provenance-all-unknown", "provenance-empty-rejected", "provenance-invalid-origin", "provenance-invalid-actor", "provenance-invalid-delivery", "provenance-invalid-ownership", "provenance-invalid-evidence", "provenance-invalid-modality", "provenance-invalid-submission", "provenance-overlong-submission"}, func(v ContentProvenance) error { return v.Validate() })
	runPrimitiveArm(t, c.Navigation, 9, []string{"navigation-local-resolved", "navigation-unavailable-no-id", "navigation-unavailable-leaks-id", "navigation-public-resolved", "navigation-resolved-missing-target", "navigation-resolved-dual-target", "navigation-general-link-absent-anchor", "navigation-general-link-general-anchor", "navigation-general-link-exact-anchor"}, func(v SessionRelationshipNavigation) error { return v.Validate() })
	runPrimitiveArm(t, c.HelperGroups, 5, []string{"helper-group-valid", "helper-group-wrong-purpose", "helper-negative-count", "helper-empty-group", "helper-empty-scope"}, func(v HelperGroupSummary) error { return v.Validate() })
	runPrimitiveArm(t, c.HelperContexts, 2, []string{"helper-context-valid", "helper-context-invalid-status"}, func(v HelperContextSummary) error { return v.Validate() })
	runPrimitiveArm(t, c.EarlierHistory, 2, []string{"earlier-empty-valid", "earlier-null-invalid"}, func(v EarlierHistorySection) error { return v.Validate() })
	runRawRelationshipArm(t, c.RawRelationships)
	runRawNavigationArm(t, c.RawNavigation)
}

func TestSessionGraphRecursiveAndRawBoundaries(t *testing.T) {
	c, err := LoadSessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireMin(t, c.Recursive, 10)
	caseassert.RequireValid(t, c.Recursive)
	caseassert.RequireMin(t, c.RawDurable, 11)
	caseassert.RequireValid(t, c.RawDurable)
	for _, arm := range []struct {
		corpus   testcase.Corpus[SessionGraphRawInput, SessionGraphFixtureExpected]
		semantic bool
	}{{c.Recursive, true}, {c.RawDurable, false}} {
		for _, x := range arm.corpus.Cases {
			t.Run(x.Name, func(t *testing.T) {
				raw := completeRawDetailFixture(t, x.Input.RawJSON)
				_, detailErr := DecodeSessionDetailPayloadRaw(raw)
				assertFixtureError(t, x.Name+" detail", x.Classification, x.Expected.ErrorContains, detailErr)
				envelope := `{"kind":"session_detail","sessionDetail":` + string(raw) + `}`
				_, envelopeErr := DecodeTranscriptContentRaw([]byte(envelope))
				assertFixtureError(t, x.Name+" transcript", x.Classification, x.Expected.ErrorContains, envelopeErr)
				if arm.semantic {
					var detail SessionDetailPayload
					if err := json.Unmarshal(raw, &detail); err != nil {
						t.Fatal(err)
					}
					assertFixtureError(t, x.Name+" typed detail", x.Classification, x.Expected.ErrorContains, ValidateSessionDetailPayload(detail))
					assertFixtureError(t, x.Name+" typed transcript", x.Classification, x.Expected.ErrorContains, ValidateTranscriptContent(TranscriptContent{Kind: ContentKindSessionDetail, SessionDetail: &detail}))
					assertFixtureError(t, x.Name+" native validator", x.Classification, x.Expected.ErrorContains, ValidateNativeMetadata(detail))
				}
			})
		}
	}
}

func TestSessionGraphNativeAggregateBudget(t *testing.T) {
	c, err := LoadSessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireMin(t, c.NativeLimits, 2)
	caseassert.RequireValid(t, c.NativeLimits)
	for _, x := range c.NativeLimits.Cases {
		t.Run(x.Name, func(t *testing.T) {
			detail := SessionDetailPayload{ID: "ses_native_budget", Harness: HarnessPi, Turns: []TurnDetail{}, EarlierHistory: []EarlierHistorySection{{State: EarlierHistoryUncertainMigrated, Turns: []TurnDetail{}}}}
			makeRecords := func(n, offset int) []NativeMetadataRecord {
				out := make([]NativeMetadataRecord, n)
				for i := range out {
					suffix := fmt.Sprintf("%d", i+offset)
					chunks := []string{}
					for remaining := x.Input.DataBytes; remaining > 0; {
						size := remaining
						if size > 15000 {
							size = 15000
						}
						chunks = append(chunks, strings.Repeat("x", size))
						remaining -= size
					}
					data, err := json.Marshal(chunks)
					if err != nil {
						t.Fatal(err)
					}
					out[i] = NativeMetadataRecord{ID: "m_" + suffix, Kind: NativeMetadataPiCustomData, Source: NativeSourceRef{EntryRef: SourceEntryRef("source_" + suffix), SourceType: NativeSourcePiCustom}, CustomType: "fixture", Data: data}
				}
				return out
			}
			detail.NativeMetadata = makeRecords(x.Input.MainRecords, 0)
			detail.EarlierHistory[0].NativeMetadata = makeRecords(x.Input.EarlierRecords, x.Input.MainRecords)
			assertFixtureError(t, x.Name+" typed", x.Classification, x.Expected.ErrorContains, ValidateSessionDetailPayload(detail))
			raw, err := json.Marshal(detail)
			if err != nil {
				t.Fatal(err)
			}
			_, err = DecodeSessionDetailPayloadRaw(raw)
			assertFixtureError(t, x.Name+" raw", x.Classification, x.Expected.ErrorContains, err)
		})
	}
}

func completeRawDetailFixture(t *testing.T, fragment string) []byte {
	t.Helper()
	base, err := json.Marshal(SessionDetailPayload{ID: "ses_fixture", Harness: HarnessClaudeCode, Turns: []TurnDetail{}, ChildSessions: []ChildSessionRef{}})
	if err != nil {
		t.Fatal(err)
	}
	var complete, overrides map[string]json.RawMessage
	if json.Unmarshal(base, &complete) != nil || json.Unmarshal([]byte(fragment), &overrides) != nil {
		t.Fatalf("invalid raw detail fixture %q", fragment)
	}
	for key, value := range overrides {
		complete[key] = value
	}
	out, err := json.Marshal(complete)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func runPrimitiveArm[I any](t *testing.T, c testcase.Corpus[I, SessionGraphFixtureExpected], min int, names []string, validate func(I) error) {
	t.Helper()
	caseassert.RequireMin(t, c, min)
	caseassert.RequireValid(t, c)
	requireNames(t, c.Cases, names)
	for _, x := range c.Cases {
		assertFixtureError(t, x.Name, x.Classification, x.Expected.ErrorContains, validate(x.Input))
	}
}
func runRawRelationshipArm(t *testing.T, c testcase.Corpus[SessionGraphRawInput, SessionGraphRawExpected]) {
	t.Helper()
	caseassert.RequireMin(t, c, 1)
	caseassert.RequireValid(t, c)
	requireNames(t, c.Cases, []string{"raw-relationship-wrong-target-type"})
	for _, x := range c.Cases {
		_, err := ValidateRawRelationshipFixture(x.Input)
		assertFixtureError(t, x.Name, x.Classification, x.Expected.ErrorContains, err)
	}
}
func runRawNavigationArm(t *testing.T, c testcase.Corpus[SessionGraphRawInput, SessionGraphRawExpected]) {
	t.Helper()
	caseassert.RequireMin(t, c, 1)
	caseassert.RequireValid(t, c)
	requireNames(t, c.Cases, []string{"raw-navigation-roundtrip-shape"})
	for _, x := range c.Cases {
		value, encoded, err := ValidateRawNavigationFixture(x.Input)
		assertFixtureError(t, x.Name, x.Classification, x.Expected.ErrorContains, err)
		if err == nil {
			var expected SessionRelationshipNavigation
			if e := json.Unmarshal([]byte(x.Expected.RawJSON), &expected); e != nil {
				t.Fatalf("%s expected raw JSON: %v", x.Name, e)
			}
			if !reflect.DeepEqual(value, expected) || encoded != x.Expected.RawJSON {
				t.Errorf("%s roundtrip value=%#v JSON=%s want %#v JSON=%s", x.Name, value, encoded, expected, x.Expected.RawJSON)
			}
		}
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
	for _, bad := range []string{i.Empty, i.Unknown} {
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
