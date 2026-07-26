package schema

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

const timelineFixtureCaseCount = 23

// TimelineFixtureInput is one normalized session and commit relationship.
type TimelineFixtureInput struct {
	Sessions         []TimelineSessionRef `json:"sessions" yaml:"sessions"`
	Commits          []CommitRef          `json:"commits" yaml:"commits"`
	RewrittenCommits []RewrittenCommit    `json:"rewrittenCommits,omitempty" yaml:"rewrittenCommits,omitempty"`
}

// timelineFixtureRepairKind names a fixture-owned repair for one rejected
// timeline input.
type timelineFixtureRepairKind string

const (
	// timelineRepairSetSessionBindingTrue flips one timeline session's
	// HasCommitBinding flag in place.
	timelineRepairSetSessionBindingTrue timelineFixtureRepairKind = "set_session_binding_true"
)

// timelineFixtureRepair records the in-place repair expected for a rejected
// timeline input.
type timelineFixtureRepair struct {
	Kind              timelineFixtureRepairKind `json:"kind" yaml:"kind"`
	SessionID         SessionID                 `json:"sessionId" yaml:"sessionId"`
	PostMutationValid bool                      `json:"postMutationValid" yaml:"postMutationValid"`
}

// Apply repairs the payload in place by restoring the authoritative session
// binding truth required by the repair fixture.
func (r *timelineFixtureRepair) Apply(payload *ReviewListPayload) error {
	if r == nil {
		return fmt.Errorf("timeline fixture repair apply: repair is nil; call Apply only when a repair is declared")
	}
	for index := range payload.Sessions {
		if payload.Sessions[index].SessionID == r.SessionID {
			payload.Sessions[index].HasCommitBinding = true
			return nil
		}
	}
	return fmt.Errorf("timeline fixture repair apply: sessionId %q not found in payload sessions; the declared repair cannot be applied", r.SessionID)
}

// TimelineFixtureExpected records the validation error for a rejected input.
// A must-pass case leaves ErrorContains empty.
type TimelineFixtureExpected struct {
	ErrorContains string                 `json:"errorContains,omitempty" yaml:"error_contains"`
	Repair        *timelineFixtureRepair `json:"repair,omitempty" yaml:"repair,omitempty"`
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
		if fixture.Expected.Repair != nil {
			if fixture.Classification != testcase.MustFail {
				return fmt.Errorf("load timeline fixtures: case %q declares a repair but is not must-fail; repairs must start from a rejected input", fixture.Name)
			}
			if fixture.Expected.Repair.Kind != timelineRepairSetSessionBindingTrue {
				return fmt.Errorf("load timeline fixtures: case %q declares unknown repair kind %q", fixture.Name, fixture.Expected.Repair.Kind)
			}
			if fixture.Expected.Repair.SessionID == "" {
				return fmt.Errorf("load timeline fixtures: case %q repair has an empty sessionId", fixture.Name)
			}
			if !fixture.Expected.Repair.PostMutationValid {
				return fmt.Errorf("load timeline fixtures: case %q repair must declare postMutationValid=true", fixture.Name)
			}
		}
	}
	if err := generic.Validate(); err != nil {
		return fmt.Errorf("load timeline fixtures: %w", err)
	}
	if passCount != 6 || failCount != 17 {
		return fmt.Errorf("load timeline fixtures: canonical outcome coverage changed; got %d must-pass and %d must-fail cases, want 6 and 17", passCount, failCount)
	}
	return nil
}
