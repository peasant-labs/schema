package schema_test

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/association_parity.yaml
var associationParityCasesYAML []byte

func loadAssociationParityFixtures(t *testing.T) testcase.Corpus[schema.SessionAssociation, bool] {
	t.Helper()
	fixtures, err := testcase.LoadCorpus[schema.SessionAssociation, bool](associationParityCasesYAML)
	if err != nil {
		t.Fatalf("load association parity fixtures: %v", err)
	}
	assert.RequireMin(t, fixtures, len(associationParityRequiredCaseNames))
	assert.RequireValid(t, fixtures)
	requireAssociationParityFixtureInventory(t, fixtures)
	return fixtures
}

var associationParityRequiredCaseNames = []string{
	"go_valid_non_empty_session_id_is_accepted",
	"distinct_same_kind_recorded_commit_observations_in_canonical_order_are_accepted",
	"unicode_e000_then_10000_utf8_order_is_accepted",
	"unicode_10000_then_e000_utf8_order_is_rejected",
}

func requireAssociationParityFixtureInventory(t *testing.T, fixtures testcase.Corpus[schema.SessionAssociation, bool]) {
	t.Helper()
	if len(fixtures.Cases) != len(associationParityRequiredCaseNames) {
		t.Fatalf("association parity corpus has %d cases, want exactly %d", len(fixtures.Cases), len(associationParityRequiredCaseNames))
	}
	names := make(map[string]struct{}, len(fixtures.Cases))
	for _, fixture := range fixtures.Cases {
		if _, exists := names[fixture.Name]; exists {
			t.Fatalf("association parity corpus repeats case %q", fixture.Name)
		}
		names[fixture.Name] = struct{}{}
	}
	for _, name := range associationParityRequiredCaseNames {
		if _, exists := names[name]; !exists {
			t.Fatalf("association parity corpus is missing required case %q", name)
		}
	}
}

func TestSessionAssociation_SharedParityCorpus(t *testing.T) {
	for _, fixture := range loadAssociationParityFixtures(t).Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			valid := fixture.Input.Validate() == nil
			if valid != fixture.Expected {
				t.Errorf("Validate() valid=%t, want %t", valid, fixture.Expected)
			}
			wantClassification := testcase.MustFail
			if fixture.Expected {
				wantClassification = testcase.MustPass
			}
			if fixture.Classification != wantClassification {
				t.Errorf("classification=%q, want %q for expected=%t", fixture.Classification, wantClassification, fixture.Expected)
			}
		})
	}
}
