package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/internal/testutil"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

func TestAssociationAnnotationIngressFixtureContract(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	cases := fixtures.CaseCorpus()
	assert.RequireMin(t, cases, 11)
	assert.RequireValid(t, cases)
	assert.RequireMin(t, fixtures.AnnotationRequestShapes, 2)
	assert.RequireValid(t, fixtures.AnnotationRequestShapes)
	assert.RequireMin(t, fixtures.StrictDecoding, 8)
	assert.RequireValid(t, fixtures.StrictDecoding)

	requireAssociationAnnotationIngressCase(t, fixtures, "durable association and association annotation are valid")
	requireAssociationAnnotationIngressCase(t, fixtures, "duplicate association ID is rejected within one request")
	requireAssociationAnnotationIngressCase(t, fixtures, "duplicate observed hash rejects a durable ID alias")
	requireAssociationAnnotationIngressCase(t, fixtures, "empty observed commit hash is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "association annotation missing target ID is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "association annotation mixed with session target is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "malformed association annotation target is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "entry annotation rejects empty session ID")
	requireAssociationAnnotationIngressCase(t, fixtures, "entry annotation rejects equal range at the runtime boundary")
	requireAssociationAnnotationIngressCase(t, fixtures, "entry annotation rejects reversed range at the runtime boundary")
	requireAssociationAnnotationIngressCase(t, fixtures, "exact durable replay is represented by one canonical association")
}

func TestAssociationAnnotationIngressStrictDecodingInventory(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	want := map[string]struct{}{
		"canonical ingress corpus shape is accepted":   {},
		"unknown ingress corpus field is rejected":     {},
		"duplicate ingress corpus field is rejected":   {},
		"unknown annotation field is rejected":         {},
		"duplicate annotation field is rejected":       {},
		"unknown entry target field is rejected":       {},
		"duplicate entry target field is rejected":     {},
		"trailing ingress corpus document is rejected": {},
	}
	if len(fixtures.StrictDecoding.Cases) != len(want) {
		t.Fatalf("strict decoding corpus has %d cases, want exact inventory of %d", len(fixtures.StrictDecoding.Cases), len(want))
	}
	for _, fixture := range fixtures.StrictDecoding.Cases {
		if _, exists := want[fixture.Name]; !exists {
			t.Fatalf("strict decoding corpus has unexpected case %q", fixture.Name)
		}
		delete(want, fixture.Name)
	}
	for name := range want {
		t.Fatalf("strict decoding corpus is missing required case %q", name)
	}
}

func TestAssociationAnnotationIngressTypedAndPublishBoundaryValidators(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	for _, fixture := range fixtures.CaseCorpus().Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			item := annotationPushItem(fixture.Input.Annotation)
			publishRequest := schema.PublishRequest{Git: schema.GitContext{Associations: publishedAssociations(fixture.Input.Associations)}}
			publishErr := publishRequest.Validate()
			annotationErr := (schema.AnnotationPushRequest{Annotations: []schema.AnnotationPushItem{item}}).Validate()
			if got := publishErr == nil; got != fixture.Expected.PublishRequestValid {
				t.Fatalf("PublishRequest.Validate() error=%v, valid=%t, want %t", publishErr, got, fixture.Expected.PublishRequestValid)
			}
			if got := annotationErr == nil; got != fixture.Expected.AnnotationRequestValid {
				t.Fatalf("AnnotationPushRequest.Validate() error=%v, valid=%t, want %t", annotationErr, got, fixture.Expected.AnnotationRequestValid)
			}
			if fixture.Expected.AnnotationEntryTargetValid != nil {
				if item.EntryTarget == nil {
					t.Fatal("fixture expects a direct AnnotationEntryTarget verdict but has no entryTarget")
				}
				if got := item.EntryTarget.Validate() == nil; got != *fixture.Expected.AnnotationEntryTargetValid {
					t.Fatalf("AnnotationEntryTarget.Validate() valid=%t, want %t", got, *fixture.Expected.AnnotationEntryTargetValid)
				}
			}
			annotationBoundaryErr := annotationPushRequestBoundaryError(t, fixture.Input.Annotation)
			if got := annotationBoundaryErr == nil; got != fixture.Expected.AnnotationRequestValid {
				t.Fatalf("ValidateAnnotationPushRequest() error=%v, valid=%t, want %t", annotationBoundaryErr, got, fixture.Expected.AnnotationRequestValid)
			}
			boundaryErr := publishRequestBoundaryError(t, fixture.Input)
			if got := boundaryErr == nil; got != fixture.Expected.PublishRequestValid {
				t.Fatalf("ValidatePublishRequest() error=%v, valid=%t, want %t", boundaryErr, got, fixture.Expected.PublishRequestValid)
			}
			wantCombined := fixture.Expected.PublishRequestValid && fixture.Expected.AnnotationRequestValid
			if (fixture.Classification == testcase.MustPass) != wantCombined {
				t.Fatalf("classification %q does not agree with combined expected validity %t", fixture.Classification, wantCombined)
			}
		})
	}
}

func TestAssociationAnnotationIngressTargetKindCoverage(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	expected := make(map[string]struct{}, len(schema.AllTargetKinds))
	coverage := make(map[string]annotationTargetKindCoverage, len(schema.AllTargetKinds))
	for _, kind := range schema.AllTargetKinds {
		expected[string(kind)] = struct{}{}
	}
	for _, fixture := range fixtures.CaseCorpus().Cases {
		kind := fixture.Input.Annotation.TargetKind
		if _, known := expected[kind]; !known {
			t.Fatalf("fixture %q has unexpected target kind %q", fixture.Name, kind)
		}
		observed := coverage[kind]
		if fixture.Expected.AnnotationRequestValid {
			observed.valid++
		} else {
			observed.invalid++
		}
		if hasExplicitNullInactiveTarget(fixture.Input.Annotation) {
			observed.explicitNull++
		}
		coverage[kind] = observed
	}
	if len(coverage) != len(expected) {
		t.Fatalf("target-kind fixture coverage has %d kinds, want exactly %d", len(coverage), len(expected))
	}
	for _, kind := range schema.AllTargetKinds {
		observed, present := coverage[string(kind)]
		if !present {
			t.Fatalf("target-kind fixture coverage is missing %q", kind)
		}
		if kind == schema.TargetFileVersion {
			if observed.valid != 0 || observed.invalid == 0 {
				t.Fatalf("file_version fixtures must prove unsupported push rejection, got valid=%d invalid=%d", observed.valid, observed.invalid)
			}
			continue
		}
		if observed.valid == 0 || observed.invalid == 0 || observed.explicitNull == 0 {
			t.Fatalf("target-kind %q fixtures must include valid, invalid, and explicit-null inactive-arm rows, got valid=%d invalid=%d explicit-null=%d", kind, observed.valid, observed.invalid, observed.explicitNull)
		}
	}
}

type annotationTargetKindCoverage struct {
	valid        int
	invalid      int
	explicitNull int
}

func hasExplicitNullInactiveTarget(annotation testutil.AssociationAnnotationIngressAnnotation) bool {
	for _, field := range []string{"targetAssociationId", "sessionId", "entryTarget", "annotationId", "projectHash"} {
		if annotation.HasExplicitNullTargetField(field) {
			return true
		}
	}
	return false
}

func TestAssociationAnnotationIngressHashDistinctTargetIsMandatory(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	fixture := requireAssociationAnnotationIngressCase(t, fixtures, "durable association and association annotation are valid")
	if fixture.Input.HashComparison == nil {
		t.Fatal("hash-distinct association target fixture is missing hashComparison")
	}
	if !fixture.Input.HashComparison.Distinct {
		t.Fatal("hash-distinct association target fixture must require a distinct hash")
	}
	item := annotationPushItem(fixture.Input.Annotation)
	comparison := item
	alternate := schema.AssociationID(fixture.Input.HashComparison.AlternateTargetAssociationID)
	comparison.TargetAssociationID = &alternate
	if item.ComputeContentHash() == comparison.ComputeContentHash() {
		t.Fatal("ComputeContentHash did not distinguish the mandatory alternate association target")
	}
}

func TestAssociationAnnotationIngressAnnotationRequestNullability(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	for _, fixture := range fixtures.AnnotationRequestShapes.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			var annotations []schema.AnnotationPushItem
			if fixture.Input.Annotations != nil {
				annotations = make([]schema.AnnotationPushItem, len(fixture.Input.Annotations))
			}
			err := (schema.AnnotationPushRequest{Annotations: annotations}).Validate()
			if got := err == nil; got != fixture.Expected {
				t.Fatalf("AnnotationPushRequest.Validate() error=%v, valid=%t, want %t", err, got, fixture.Expected)
			}
			boundaryErr := annotationRequestShapeBoundaryError(t, fixture.Input)
			if got := boundaryErr == nil; got != fixture.Expected {
				t.Fatalf("ValidateAnnotationPushRequest() error=%v, valid=%t, want %t", boundaryErr, got, fixture.Expected)
			}
			if (fixture.Classification == testcase.MustPass) != fixture.Expected {
				t.Fatalf("classification %q does not agree with expected validity %t", fixture.Classification, fixture.Expected)
			}
		})
	}
}

func TestAssociationAnnotationIngressFixtureStrictDecoding(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	for _, fixture := range fixtures.StrictDecoding.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			_, err := testutil.DecodeAssociationAnnotationIngressFixtures([]byte(fixture.Input))
			if (err == nil) != fixture.Expected {
				t.Fatalf("decodeAssociationAnnotationIngressFixtures() error=%v, want valid=%t", err, fixture.Expected)
			}
			if (fixture.Classification == testcase.MustPass) != fixture.Expected {
				t.Fatalf("classification %q does not agree with expected validity %t", fixture.Classification, fixture.Expected)
			}
		})
	}
}

func requireAssociationAnnotationIngressCase(t *testing.T, fixtures *testutil.AssociationAnnotationIngressFixtures, name string) testcase.Case[testutil.AssociationAnnotationIngressInput, testutil.AssociationAnnotationIngressExpected] {
	t.Helper()
	for _, fixture := range fixtures.CaseCorpus().Cases {
		if fixture.Name == name {
			return fixture
		}
	}
	t.Fatalf("association annotation ingress corpus is missing required case %q; the relevant validator behavior would be untested", name)
	return testcase.Case[testutil.AssociationAnnotationIngressInput, testutil.AssociationAnnotationIngressExpected]{}
}

func publishRequestBoundaryError(t *testing.T, input testutil.AssociationAnnotationIngressInput) error {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"model": map[string]string{"harness": "claude-code", "model": "fixture-model"},
		"git":   map[string]any{"associations": publishedAssociations(input.Associations)},
	})
	if err != nil {
		t.Fatalf("marshal fixture publish body: %v", err)
	}
	return schema.ValidatePublishRequest(body)
}

func annotationPushRequestBoundaryError(t *testing.T, annotation testutil.AssociationAnnotationIngressAnnotation) error {
	t.Helper()
	item := testutil.AnnotationPushItemTargetJSON(annotation)
	item["contentHash"] = "fixture-content-hash"
	item["typeId"] = "fixture.annotation"
	item["value"] = "fixture-value"
	item["isPrimary"] = false
	body, err := json.Marshal(map[string]any{"annotations": []any{item}})
	if err != nil {
		t.Fatalf("marshal fixture annotation body: %v", err)
	}
	return schema.ValidateAnnotationPushRequest(body)
}

func annotationRequestShapeBoundaryError(t *testing.T, input testutil.AnnotationRequestShape) error {
	t.Helper()
	body, err := json.Marshal(map[string]any{"annotations": input.Annotations})
	if err != nil {
		t.Fatalf("marshal fixture annotation request body: %v", err)
	}
	return schema.ValidateAnnotationPushRequest(body)
}

func TestAssociationAnnotationIngressFixtureStrictDecodingErrorsAreActionable(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	for _, fixture := range fixtures.StrictDecoding.Cases {
		if fixture.Expected {
			continue
		}
		t.Run(fixture.Name, func(t *testing.T) {
			_, err := testutil.DecodeAssociationAnnotationIngressFixtures([]byte(fixture.Input))
			if err == nil {
				t.Fatal("strict decoding case unexpectedly succeeded")
			}
			if !strings.Contains(err.Error(), "association annotation ingress") {
				t.Fatalf("strict decoding error lacks fixture-boundary context: %v", err)
			}
		})
	}
}

func loadAssociationAnnotationIngressFixtures(t *testing.T) *testutil.AssociationAnnotationIngressFixtures {
	t.Helper()
	fixtures, err := testutil.DecodeAssociationAnnotationIngressFixtures(schema.AssociationAnnotationIngressYAML)
	if err != nil {
		t.Fatalf("DecodeAssociationAnnotationIngressFixtures: %v", err)
	}
	return fixtures
}

func publishedAssociations(fixtures []testutil.PublishedAssociationFixture) []schema.PublishedAssociation {
	associations := make([]schema.PublishedAssociation, len(fixtures))
	for index, fixture := range fixtures {
		associations[index] = schema.PublishedAssociation{
			ID:                 schema.AssociationID(fixture.ID),
			ObservedCommitHash: fixture.ObservedCommitHash,
		}
	}
	return associations
}

func annotationPushItem(fixture testutil.AssociationAnnotationIngressAnnotation) schema.AnnotationPushItem {
	item := schema.AnnotationPushItem{
		TargetKind:   schema.TargetKind(fixture.TargetKind),
		SessionID:    fixture.SessionID,
		AnnotationID: fixture.AnnotationID,
		TypeID:       "fixture.annotation",
		Value:        "fixture-value",
	}
	if fixture.TargetAssociationID != nil {
		associationID := schema.AssociationID(*fixture.TargetAssociationID)
		item.TargetAssociationID = &associationID
	}
	if fixture.ProjectHash != nil {
		projectHash := schema.ProjectHash(*fixture.ProjectHash)
		item.ProjectHash = &projectHash
	}
	if fixture.EntryTarget != nil {
		item.EntryTarget = &schema.AnnotationEntryTarget{
			SessionID:  fixture.EntryTarget.SessionID,
			EntryIndex: fixture.EntryTarget.EntryIndex,
			EndIndex:   fixture.EntryTarget.EndIndex,
		}
	}
	return item
}
