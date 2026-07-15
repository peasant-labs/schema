package schema

import (
	"fmt"
	"strings"

	"github.com/peasant-labs/schema/testcase"
)

const timelineFixtureCaseCount = 14

// TimelineFixtureInput is one normalized session and commit relationship.
type TimelineFixtureInput struct {
	Sessions []TimelineSessionRef `json:"sessions" yaml:"sessions"`
	Commits  []CommitRef          `json:"commits" yaml:"commits"`
}

// TimelineFixtureExpected records the validation error for a rejected input.
// A must-pass case leaves ErrorContains empty.
type TimelineFixtureExpected struct {
	ErrorContains string `json:"errorContains,omitempty" yaml:"error_contains"`
}

// TimelineFixtureCorpus is the project Git timeline validation corpus.
type TimelineFixtureCorpus = testcase.Corpus[TimelineFixtureInput, TimelineFixtureExpected]

// LoadTimelineFixtures parses and validates the shared timeline corpus.
func LoadTimelineFixtures() (TimelineFixtureCorpus, error) {
	fixtures, err := testcase.LoadCorpus[TimelineFixtureInput, TimelineFixtureExpected](TimelineYAML)
	if err != nil {
		return TimelineFixtureCorpus{}, fmt.Errorf("load timeline fixtures: %w", err)
	}
	if err := validateTimelineFixtures(fixtures); err != nil {
		return TimelineFixtureCorpus{}, err
	}
	return fixtures, nil
}

func validateTimelineFixtures(fixtures TimelineFixtureCorpus) error {
	if len(fixtures.Cases) != timelineFixtureCaseCount {
		return fmt.Errorf("load timeline fixtures: corpus has %d cases, want exactly %d canonical relationship cases", len(fixtures.Cases), timelineFixtureCaseCount)
	}
	passCount, failCount := 0, 0
	for _, fixture := range fixtures.Cases {
		if len(fixture.Input.Sessions) == 0 && len(fixture.Input.Commits) == 0 {
			return fmt.Errorf("load timeline fixtures: case %q has no sessions or commits; add relationship data so the case cannot pass vacuously", fixture.Name)
		}
		switch fixture.Classification {
		case testcase.MustPass:
			passCount++
			if strings.TrimSpace(fixture.Expected.ErrorContains) != "" {
				return fmt.Errorf("load timeline fixtures: must-pass case %q unexpectedly declares error_contains; remove the contradictory error expectation", fixture.Name)
			}
		case testcase.MustFail:
			failCount++
			if strings.TrimSpace(fixture.Expected.ErrorContains) == "" {
				return fmt.Errorf("load timeline fixtures: must-fail case %q has no error_contains; name the validation failure the mutation must trigger", fixture.Name)
			}
		}
	}
	if passCount != 5 || failCount != 9 {
		return fmt.Errorf("load timeline fixtures: canonical outcome coverage changed; got %d must-pass and %d must-fail cases, want 5 and 9", passCount, failCount)
	}
	return nil
}
