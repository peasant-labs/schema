package schema_test

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/associations.yaml
var associationsCasesYAML []byte

// TestSessionAssociation_FixtureContract drives every
// SessionAssociation case in testdata/local-api/associations.yaml against
// the real SessionAssociation.Validate, and confirms every member of the
// three closed sets it composes (AssociationKind, Confidence,
// AssociationEvidence) has at least one must-pass covering case.
func TestSessionAssociation_FixtureContract(t *testing.T) {
	corpus, err := testcase.LoadCorpus[schema.SessionAssociation, bool](associationsCasesYAML)
	if err != nil {
		t.Fatalf("load associations corpus: %v", err)
	}
	assert.RequireMin(t, corpus, len(schema.AllAssociationKinds)+1)
	assert.RequireValid(t, corpus)

	coveredKinds := map[schema.AssociationKind]bool{}
	coveredConfidences := map[schema.Confidence]bool{}
	coveredEvidence := map[schema.AssociationEvidence]bool{}

	for _, c := range corpus.Cases {
		err := c.Input.Validate()
		valid := err == nil
		if valid != c.Expected {
			t.Errorf("%s: Validate() error=%v (valid=%v), want valid=%v", c.Name, err, valid, c.Expected)
			continue
		}
		if c.Classification == testcase.MustFail {
			requireActionableValidationError(t, err)
			continue
		}
		coveredKinds[c.Input.Kind] = true
		coveredConfidences[c.Input.Confidence] = true
		coveredEvidence[c.Input.Evidence] = true
	}

	for _, kind := range schema.AllAssociationKinds {
		if !coveredKinds[kind] {
			t.Errorf("AssociationKind member %q has no must-pass fixture case", kind)
		}
	}
	for _, confidence := range schema.AllConfidences {
		if !coveredConfidences[confidence] {
			t.Errorf("Confidence member %q has no must-pass fixture case", confidence)
		}
	}
	for _, evidence := range schema.AllAssociationEvidences {
		if !coveredEvidence[evidence] {
			t.Errorf("AssociationEvidence member %q has no must-pass fixture case", evidence)
		}
	}
}
