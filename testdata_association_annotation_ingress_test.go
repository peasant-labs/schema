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
	assert.RequireMin(t, cases, 9)
	assert.RequireValid(t, cases)
	assert.RequireMin(t, fixtures.AnnotationRequestShapes, 2)
	assert.RequireValid(t, fixtures.AnnotationRequestShapes)
	assert.RequireMin(t, fixtures.StrictDecoding, 4)
	assert.RequireValid(t, fixtures.StrictDecoding)

	requireAssociationAnnotationIngressCase(t, fixtures, "durable association and association annotation are valid")
	requireAssociationAnnotationIngressCase(t, fixtures, "duplicate association ID is rejected within one request")
	requireAssociationAnnotationIngressCase(t, fixtures, "duplicate observed hash rejects a durable ID alias")
	requireAssociationAnnotationIngressCase(t, fixtures, "empty observed commit hash is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "association annotation missing target ID is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "association annotation mixed with session target is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "malformed association annotation target is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "exact durable replay is represented by one canonical association")
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
	return item
}
