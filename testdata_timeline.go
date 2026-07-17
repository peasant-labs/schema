package schema

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

const timelineFixtureCaseCount = 16

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

// TimelineFixtureCase is one public relationship case. Family is the stable
// behavioral identity; Name remains the executable testcase identity.
type TimelineFixtureCase struct {
	Family         string                  `json:"family" yaml:"family"`
	Name           string                  `json:"name" yaml:"name"`
	Input          TimelineFixtureInput    `json:"input" yaml:"input"`
	Expected       TimelineFixtureExpected `json:"expected" yaml:"expected"`
	Classification testcase.Classification `json:"classification" yaml:"classification"`
	Provenance     testcase.Provenance     `json:"provenance" yaml:"provenance"`
	Mutation       testcase.Mutation       `json:"mutation" yaml:"mutation"`
}

// TimelineFixtureCorpus is the project Git timeline validation corpus.
type TimelineFixtureCorpus struct {
	Cases []TimelineFixtureCase `json:"cases" yaml:"cases"`
}

// CheckMin rejects a timeline corpus smaller than the required behavioral floor.
func (c TimelineFixtureCorpus) CheckMin(minimum int) error {
	if len(c.Cases) < minimum {
		return fmt.Errorf("timeline corpus has %d cases, want at least %d", len(c.Cases), minimum)
	}
	return nil
}

// LoadTimelineFixtures parses and validates the shared public timeline corpus.
func LoadTimelineFixtures() (TimelineFixtureCorpus, error) {
	var fixtures TimelineFixtureCorpus
	decoder := yaml.NewDecoder(bytes.NewReader(TimelineYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		return TimelineFixtureCorpus{}, fmt.Errorf("load timeline fixtures: decode document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return TimelineFixtureCorpus{}, fmt.Errorf("load timeline fixtures: decode trailing document: %w", err)
		}
		return TimelineFixtureCorpus{}, fmt.Errorf("load timeline fixtures: multiple YAML documents are not allowed")
	}
	if err := validateTimelineFixtures(fixtures); err != nil {
		return TimelineFixtureCorpus{}, err
	}
	return fixtures, nil
}

func validateTimelineFixtures(fixtures TimelineFixtureCorpus) error {
	if err := fixtures.CheckMin(timelineFixtureCaseCount); err != nil {
		return fmt.Errorf("load timeline fixtures: %w", err)
	}
	if len(fixtures.Cases) != timelineFixtureCaseCount {
		return fmt.Errorf("load timeline fixtures: corpus has %d cases, want exactly %d canonical relationship cases", len(fixtures.Cases), timelineFixtureCaseCount)
	}
	generic := testcase.Corpus[TimelineFixtureInput, TimelineFixtureExpected]{
		Cases: make([]testcase.Case[TimelineFixtureInput, TimelineFixtureExpected], len(fixtures.Cases)),
	}
	families := make(map[string]struct{}, len(fixtures.Cases))
	passCount, failCount := 0, 0
	for index, fixture := range fixtures.Cases {
		generic.Cases[index] = testcase.Case[TimelineFixtureInput, TimelineFixtureExpected]{
			Name:           fixture.Name,
			Input:          fixture.Input,
			Expected:       fixture.Expected,
			Classification: fixture.Classification,
			Provenance:     fixture.Provenance,
			Mutation:       fixture.Mutation,
		}
		if strings.TrimSpace(fixture.Family) == "" {
			return fmt.Errorf("load timeline fixtures: case %q has an empty family; identify the behavior this row protects", fixture.Name)
		}
		if _, duplicate := families[fixture.Family]; duplicate {
			return fmt.Errorf("load timeline fixtures: duplicate family %q; each canonical behavior must have one row", fixture.Family)
		}
		families[fixture.Family] = struct{}{}
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
	if err := generic.Validate(); err != nil {
		return fmt.Errorf("load timeline fixtures: %w", err)
	}
	if passCount != 5 || failCount != 11 {
		return fmt.Errorf("load timeline fixtures: canonical outcome coverage changed; got %d must-pass and %d must-fail cases, want 5 and 11", passCount, failCount)
	}
	return nil
}
