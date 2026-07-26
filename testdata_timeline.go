package schema

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

const (
	timelineFixtureCaseCount                       = 24
	timelineSuccessorAssociationMirrorFixtureCount = 5
)

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
	// timelineRepairReplaceSuccessorAssociation restores one displayed
	// successor association from its authoritative rewrite-ledger counterpart.
	timelineRepairReplaceSuccessorAssociation timelineFixtureRepairKind = "replace_successor_association"
)

// timelineFixtureRepair records the in-place repair expected for a rejected
// timeline input.
type timelineFixtureRepair struct {
	Kind              timelineFixtureRepairKind `json:"kind" yaml:"kind"`
	SessionID         SessionID                 `json:"sessionId" yaml:"sessionId"`
	GhostHash         string                    `json:"ghostHash" yaml:"ghostHash"`
	SuccessorHash     string                    `json:"successorHash" yaml:"successorHash"`
	AssociationID     AssociationID             `json:"associationId" yaml:"associationId"`
	PostMutationValid bool                      `json:"postMutationValid" yaml:"postMutationValid"`
}

// Apply repairs the payload in place by restoring the authoritative session
// binding truth required by the repair fixture.
func (r *timelineFixtureRepair) Apply(payload *ReviewListPayload) error {
	if r == nil {
		return fmt.Errorf("timeline fixture repair apply: repair is nil; call Apply only when a repair is declared")
	}
	switch r.Kind {
	case timelineRepairSetSessionBindingTrue:
		for index := range payload.Sessions {
			if payload.Sessions[index].SessionID == r.SessionID {
				payload.Sessions[index].HasCommitBinding = true
				return nil
			}
		}
		return fmt.Errorf("timeline fixture repair apply: sessionId %q not found in payload sessions; the declared repair cannot be applied", r.SessionID)
	case timelineRepairReplaceSuccessorAssociation:
		return r.replaceSuccessorAssociation(payload)
	default:
		return fmt.Errorf("timeline fixture repair apply: unknown repair kind %q", r.Kind)
	}
}

func (r *timelineFixtureRepair) replaceSuccessorAssociation(payload *ReviewListPayload) error {
	ledgerIndex := -1
	for index := range payload.RewrittenCommits {
		if payload.RewrittenCommits[index].GhostHash == r.GhostHash {
			ledgerIndex = index
			break
		}
	}
	if ledgerIndex < 0 {
		return fmt.Errorf("timeline fixture repair apply: ghostHash %q is not present in rewrittenCommits; the declared successor-association repair cannot be applied", r.GhostHash)
	}
	ledger := payload.RewrittenCommits[ledgerIndex]
	if ledger.SuccessorHash == nil || *ledger.SuccessorHash != r.SuccessorHash {
		return fmt.Errorf("timeline fixture repair apply: ghostHash %q does not name successorHash %q; the declared successor-association repair cannot be applied", r.GhostHash, r.SuccessorHash)
	}
	ledgerAssociationIndex := -1
	for index, association := range ledger.Associations {
		if association.ID == r.AssociationID {
			ledgerAssociationIndex = index
			break
		}
	}
	if ledgerAssociationIndex < 0 {
		return fmt.Errorf("timeline fixture repair apply: rewrite ledger ghostHash %q has no associationId %q; the declared successor-association repair cannot be applied", r.GhostHash, r.AssociationID)
	}
	ledgerAssociation := ledger.Associations[ledgerAssociationIndex]
	for commitIndex := range payload.RecentCommits {
		commit := &payload.RecentCommits[commitIndex]
		if commit.Hash != r.SuccessorHash {
			continue
		}
		for associationIndex, association := range commit.Associations {
			if association.ID != ledgerAssociation.ID && association.SessionID != ledgerAssociation.SessionID {
				continue
			}
			commit.Associations[associationIndex] = ledgerAssociation
			commit.SessionIDs[associationIndex] = ledgerAssociation.SessionID
			return nil
		}
		return fmt.Errorf("timeline fixture repair apply: successorHash %q has no association sharing associationId %q or sessionId %q with rewrite ledger ghostHash %q; the declared successor-association repair cannot be applied", r.SuccessorHash, ledgerAssociation.ID, ledgerAssociation.SessionID, r.GhostHash)
	}
	return fmt.Errorf("timeline fixture repair apply: successorHash %q is not present in recentCommits; the declared successor-association repair cannot be applied", r.SuccessorHash)
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
	Cases                           []TimelineFixtureCase `json:"cases" yaml:"cases"`
	SuccessorAssociationMirrorCases []TimelineFixtureCase `json:"successorAssociationMirrorCases" yaml:"successorAssociationMirrorCases"`
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
	if len(fixtures.Cases) != timelineFixtureCaseCount {
		return fmt.Errorf("load timeline fixtures: corpus has %d cases, want exactly %d canonical relationship cases", len(fixtures.Cases), timelineFixtureCaseCount)
	}
	if err := fixtures.CheckMin(timelineFixtureCaseCount); err != nil {
		return fmt.Errorf("load timeline fixtures: %w", err)
	}
	passCount, failCount, err := validateTimelineFixtureCases(fixtures.Cases)
	if err != nil {
		return fmt.Errorf("load timeline fixtures: canonical relationship cases: %w", err)
	}
	if passCount != 7 || failCount != 17 {
		return fmt.Errorf("load timeline fixtures: canonical outcome coverage changed; got %d must-pass and %d must-fail cases, want 7 and 17", passCount, failCount)
	}
	if len(fixtures.SuccessorAssociationMirrorCases) != timelineSuccessorAssociationMirrorFixtureCount {
		return fmt.Errorf("load timeline fixtures: successor association mirror corpus has %d cases, want exactly %d field mutations", len(fixtures.SuccessorAssociationMirrorCases), timelineSuccessorAssociationMirrorFixtureCount)
	}
	mirrorPassCount, mirrorFailCount, err := validateTimelineFixtureCases(fixtures.SuccessorAssociationMirrorCases)
	if err != nil {
		return fmt.Errorf("load timeline fixtures: successor association mirror cases: %w", err)
	}
	if mirrorPassCount != 0 || mirrorFailCount != timelineSuccessorAssociationMirrorFixtureCount {
		return fmt.Errorf("load timeline fixtures: successor association mirror outcome coverage changed; got %d must-pass and %d must-fail cases, want 0 and %d", mirrorPassCount, mirrorFailCount, timelineSuccessorAssociationMirrorFixtureCount)
	}
	return nil
}

func validateTimelineFixtureCases(fixtures []TimelineFixtureCase) (int, int, error) {
	generic := testcase.Corpus[TimelineFixtureInput, TimelineFixtureExpected]{
		Cases: make([]testcase.Case[TimelineFixtureInput, TimelineFixtureExpected], len(fixtures)),
	}
	families := make(map[string]struct{}, len(fixtures))
	passCount, failCount := 0, 0
	for index, fixture := range fixtures {
		generic.Cases[index] = testcase.Case[TimelineFixtureInput, TimelineFixtureExpected]{
			Name:           fixture.Name,
			Input:          fixture.Input,
			Expected:       fixture.Expected,
			Classification: fixture.Classification,
			Provenance:     fixture.Provenance,
			Mutation:       fixture.Mutation,
		}
		if strings.TrimSpace(fixture.Family) == "" {
			return 0, 0, fmt.Errorf("case %q has an empty family; identify the behavior this row protects", fixture.Name)
		}
		if _, duplicate := families[fixture.Family]; duplicate {
			return 0, 0, fmt.Errorf("duplicate family %q; each canonical behavior must have one row", fixture.Family)
		}
		families[fixture.Family] = struct{}{}
		if len(fixture.Input.Sessions) == 0 && len(fixture.Input.Commits) == 0 {
			return 0, 0, fmt.Errorf("case %q has no sessions or commits; add relationship data so the case cannot pass vacuously", fixture.Name)
		}
		switch fixture.Classification {
		case testcase.MustPass:
			passCount++
			if strings.TrimSpace(fixture.Expected.ErrorContains) != "" {
				return 0, 0, fmt.Errorf("must-pass case %q unexpectedly declares error_contains; remove the contradictory error expectation", fixture.Name)
			}
		case testcase.MustFail:
			failCount++
			if strings.TrimSpace(fixture.Expected.ErrorContains) == "" {
				return 0, 0, fmt.Errorf("must-fail case %q has no error_contains; name the validation failure the mutation must trigger", fixture.Name)
			}
		}
		if fixture.Expected.Repair != nil {
			if fixture.Classification != testcase.MustFail {
				return 0, 0, fmt.Errorf("case %q declares a repair but is not must-fail; repairs must start from a rejected input", fixture.Name)
			}
			if err := validateTimelineFixtureRepair(fixture.Name, fixture.Expected.Repair); err != nil {
				return 0, 0, err
			}
			if !fixture.Expected.Repair.PostMutationValid {
				return 0, 0, fmt.Errorf("case %q repair must declare postMutationValid=true", fixture.Name)
			}
		}
	}
	if err := generic.Validate(); err != nil {
		return 0, 0, err
	}
	return passCount, failCount, nil
}

func validateTimelineFixtureRepair(name string, repair *timelineFixtureRepair) error {
	switch repair.Kind {
	case timelineRepairSetSessionBindingTrue:
		if repair.SessionID == "" {
			return fmt.Errorf("case %q repair has an empty sessionId", name)
		}
	case timelineRepairReplaceSuccessorAssociation:
		if strings.TrimSpace(repair.GhostHash) == "" || strings.TrimSpace(repair.SuccessorHash) == "" {
			return fmt.Errorf("case %q successor-association repair must name non-empty ghostHash and successorHash values", name)
		}
		if err := repair.AssociationID.Validate(); err != nil {
			return fmt.Errorf("case %q successor-association repair has invalid associationId: %w", name, err)
		}
	default:
		return fmt.Errorf("case %q declares unknown repair kind %q", name, repair.Kind)
	}
	return nil
}
