package schema_test

import (
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
	assert.RequireMin(t, fixtures.StrictDecoding, 4)
	assert.RequireValid(t, fixtures.StrictDecoding)

	requireAssociationAnnotationIngressCase(t, fixtures, "durable association and association annotation are valid")
	requireAssociationAnnotationIngressCase(t, fixtures, "duplicate association ID is rejected within one request")
	requireAssociationAnnotationIngressCase(t, fixtures, "duplicate observed hash rejects a durable ID alias")
	requireAssociationAnnotationIngressCase(t, fixtures, "empty observed commit hash is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "association annotation missing target ID is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "association annotation mixed with session target is rejected")
	requireAssociationAnnotationIngressCase(t, fixtures, "malformed association annotation target is rejected")
}

func TestAssociationAnnotationIngressTypedValidatorsAndHashes(t *testing.T) {
	fixtures := loadAssociationAnnotationIngressFixtures(t)
	for _, fixture := range fixtures.CaseCorpus().Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			item := annotationPushItem(fixture.Input.Annotation)
			publishErr := (schema.PublishRequest{Git: schema.GitContext{Associations: publishedAssociations(fixture.Input.Associations)}}).Validate()
			annotationErr := (schema.AnnotationPushRequest{Annotations: []schema.AnnotationPushItem{item}}).Validate()
			gotValid := publishErr == nil && annotationErr == nil
			if gotValid != fixture.Expected {
				t.Fatalf("PublishRequest.Validate() error=%v, AnnotationPushRequest.Validate() error=%v, combined valid=%t, want %t", publishErr, annotationErr, gotValid, fixture.Expected)
			}
			if (fixture.Classification == testcase.MustPass) != fixture.Expected {
				t.Fatalf("classification %q does not agree with expected validity %t", fixture.Classification, fixture.Expected)
			}
			if fixture.Input.HashComparison == nil {
				return
			}
			comparison := item
			alternate := schema.AssociationID(fixture.Input.HashComparison.AlternateTargetAssociationID)
			comparison.TargetAssociationID = &alternate
			gotDistinct := item.ComputeContentHash() != comparison.ComputeContentHash()
			if gotDistinct != fixture.Input.HashComparison.Distinct {
				t.Fatalf("ComputeContentHash target distinction=%t, want %t", gotDistinct, fixture.Input.HashComparison.Distinct)
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

func requireAssociationAnnotationIngressCase(t *testing.T, fixtures *testutil.AssociationAnnotationIngressFixtures, name string) {
	t.Helper()
	for _, fixture := range fixtures.CaseCorpus().Cases {
		if fixture.Name == name {
			return
		}
	}
	t.Fatalf("association annotation ingress corpus is missing required case %q; the relevant validator behavior would be untested", name)
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
