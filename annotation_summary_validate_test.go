package schema_test

import (
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// TestAnnotationSummary_TargetFixtureContract exercises the shared response
// validator over YAML-owned target arm cases, including the association ID arm.
func TestAnnotationSummary_TargetFixtureContract(t *testing.T) {
	fixtures, err := schema.LoadAnnotationFixtures()
	if err != nil {
		t.Fatalf("LoadAnnotationFixtures: %v", err)
	}
	assert.RequireMin(t, fixtures.TargetValidations, 9)
	assert.RequireValid(t, fixtures.TargetValidations)
	for _, fixture := range fixtures.TargetValidations.Cases {
		err := fixture.Input.Validate()
		if (err == nil) != fixture.Expected {
			t.Fatalf("%s: Validate() error=%v, want valid=%v", fixture.Name, err, fixture.Expected)
		}
		if fixture.Classification == testcase.MustFail {
			requireActionableValidationError(t, err)
		}
	}
}

// TestAnnotationSummary_TargetRepairCorpus proves a mixed target can be
// repaired through AnnotationSummary.Validate without a test-only path.
func TestAnnotationSummary_TargetRepairCorpus(t *testing.T) {
	fixtures, err := schema.LoadAnnotationFixtures()
	if err != nil {
		t.Fatalf("LoadAnnotationFixtures: %v", err)
	}
	assert.RequireMin(t, fixtures.TargetRepairs, 2)
	assert.RequireValid(t, fixtures.TargetRepairs)
	sources := make(map[string]testcase.Case[schema.AnnotationSummary, bool], len(fixtures.TargetValidations.Cases))
	for _, fixture := range fixtures.TargetValidations.Cases {
		sources[fixture.Name] = fixture
	}
	for _, repair := range fixtures.TargetRepairs.Cases {
		t.Run(repair.Name, func(t *testing.T) {
			source, exists := sources[repair.Input.SourceCase]
			if !exists || source.Classification != testcase.MustFail || source.Expected {
				t.Fatalf("repair source %q must be a rejected target fixture", repair.Input.SourceCase)
			}
			input := source.Input
			originalErr := input.Validate()
			requireActionableValidationError(t, originalErr, repair.Expected.OriginalErrorContains)
			switch string(repair.Input.Kind) {
			case "clear_target_session_id":
				if repair.Input.TargetAssociationID != nil {
					t.Fatalf("repair kind %q must not declare targetAssociationId", repair.Input.Kind)
				}
				input.TargetSessionID = nil
			case "set_target_association_id":
				if repair.Input.TargetAssociationID == nil {
					t.Fatalf("repair kind %q requires targetAssociationId", repair.Input.Kind)
				}
				associationID := *repair.Input.TargetAssociationID
				input.TargetAssociationID = &associationID
			default:
				t.Fatalf("unknown annotation target repair kind %q", repair.Input.Kind)
			}
			if err := input.Validate(); (err == nil) != repair.Expected.PostMutationValid {
				t.Fatalf("post-mutation Validate() error=%v, want valid=%v", err, repair.Expected.PostMutationValid)
			}
		})
	}
}
