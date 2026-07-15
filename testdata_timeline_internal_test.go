package schema

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
)

func TestTimelineFixtureValidationRejectsVacuousMutation(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases[0].Mutation.Description = " "
	if err := fixtures.Validate(); err == nil || !strings.Contains(err.Error(), "mutation description is empty") {
		t.Fatalf("vacuous mutation error = %v, want mutation-description rejection", err)
	}
}

func TestTimelineFixtureValidationRejectsContradictoryOutcome(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases[0].Classification = testcase.MustFail
	if err := validateTimelineFixtures(fixtures); err == nil || !strings.Contains(err.Error(), "has no error_contains") {
		t.Fatalf("contradictory outcome error = %v, want missing error expectation rejection", err)
	}
}

func TestTimelineFixtureValidationRejectsMissingCanonicalCase(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases = fixtures.Cases[:len(fixtures.Cases)-1]
	if err := validateTimelineFixtures(fixtures); err == nil || !strings.Contains(err.Error(), "want exactly 14") {
		t.Fatalf("missing-case error = %v, want exact-count rejection", err)
	}
}
