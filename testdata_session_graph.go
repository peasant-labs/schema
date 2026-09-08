package schema

import (
	"fmt"

	"github.com/peasant-labs/schema/testcase"
)

type SessionGraphFixtureInput struct {
	Operation    string                         `yaml:"operation"`
	Value        string                         `yaml:"value,omitempty"`
	Relationship *SessionRelationship           `yaml:"relationship,omitempty"`
	Relations    []SessionRelationship          `yaml:"relations,omitempty"`
	Provenance   *ContentProvenance             `yaml:"provenance,omitempty"`
	Navigation   *SessionRelationshipNavigation `yaml:"navigation,omitempty"`
	Group        *HelperGroupSummary            `yaml:"group,omitempty"`
	Context      *HelperContextSummary          `yaml:"context,omitempty"`
	Earlier      *EarlierHistorySection         `yaml:"earlier,omitempty"`
}
type SessionGraphFixtureExpected struct {
	ErrorContains string `yaml:"error_contains,omitempty"`
}
type SessionGraphFixtureCorpus = testcase.Corpus[SessionGraphFixtureInput, SessionGraphFixtureExpected]

func LoadSessionGraphFixtures() (SessionGraphFixtureCorpus, error) {
	c, err := testcase.LoadCorpus[SessionGraphFixtureInput, SessionGraphFixtureExpected](SessionGraphProvenanceYAML)
	if err != nil {
		return c, fmt.Errorf("load session graph fixtures: %w", err)
	}
	return c, nil
}

func ValidateSessionGraphFixtureInput(i SessionGraphFixtureInput) error {
	switch i.Operation {
	case "source_ref":
		_, e := NewSourceEntryRef(i.Value)
		return e
	case "submission_ref":
		_, e := NewSubmissionRef(i.Value)
		return e
	case "revision_ref":
		_, e := NewPublicRevisionRef(i.Value)
		return e
	case "relationship":
		if i.Relationship == nil {
			return fmt.Errorf("missing relationship")
		}
		return i.Relationship.Validate()
	case "relations":
		return ValidateSessionRelationships(i.Relations)
	case "provenance":
		if i.Provenance == nil {
			return fmt.Errorf("missing provenance")
		}
		return i.Provenance.Validate()
	case "navigation":
		if i.Navigation == nil {
			return fmt.Errorf("missing navigation")
		}
		return i.Navigation.Validate()
	case "group":
		if i.Group == nil {
			return fmt.Errorf("missing group")
		}
		return i.Group.Validate()
	case "context":
		if i.Context == nil {
			return fmt.Errorf("missing context")
		}
		return i.Context.Validate()
	case "earlier":
		if i.Earlier == nil {
			return fmt.Errorf("missing earlier")
		}
		return i.Earlier.Validate()
	default:
		return fmt.Errorf("unknown fixture operation %q", i.Operation)
	}
}
