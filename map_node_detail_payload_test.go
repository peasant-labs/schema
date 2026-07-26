package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/map_node_detail_payload.yaml
var mapNodeDetailPayloadCasesYAML []byte

//go:embed testdata/local-api/map_node_detail_payload_manifest.yaml
var mapNodeDetailPayloadManifestYAML []byte

type payloadValidationKind string

const (
	payloadValidationMapNodeDetail payloadValidationKind = "map_node_detail"
	payloadValidationChangeDetail  payloadValidationKind = "change_detail"
	payloadValidationReviewList    payloadValidationKind = "review_list"
)

type payloadNilSlice string

const (
	payloadNilSliceNone             payloadNilSlice = "none"
	payloadNilSliceInsights         payloadNilSlice = "insights"
	payloadNilSliceRewrittenCommits payloadNilSlice = "rewritten_commits"
)

type payloadMutationKind string

const (
	payloadMutationRemoveSuccessorCommit      payloadMutationKind = "remove_successor_commit"
	payloadMutationAddSuccessorCommit         payloadMutationKind = "add_successor_commit"
	payloadMutationRemoveDuplicateSuccessor   payloadMutationKind = "remove_duplicate_successor_commit"
	payloadMutationRepairSuccessorAssociation payloadMutationKind = "repair_successor_association_mirror"
	payloadMutationDriftSuccessorAssociation  payloadMutationKind = "drift_successor_association_conclusion"
	payloadMutationRepairCommitRefShape       payloadMutationKind = "repair_commit_ref_shape"
	payloadMutationClearClassification        payloadMutationKind = "clear_classification"
	payloadMutationNilInsights                payloadMutationKind = "nil_insights"
	payloadMutationInitializeInsights         payloadMutationKind = "initialize_insights"
	payloadMutationInitializeRewrittenCommits payloadMutationKind = "initialize_rewritten_commits"
)

type payloadValidationFixtureInput struct {
	Payload            payloadValidationKind       `yaml:"payload"`
	Path               string                      `yaml:"path"`
	RecentCommitHashes []string                    `yaml:"recentCommitHashes"`
	RecentCommits      []schema.CommitRef          `yaml:"recentCommits"`
	RewrittenCommits   []schema.RewrittenCommit    `yaml:"rewrittenCommits"`
	Sessions           []schema.TimelineSessionRef `yaml:"sessions"`
	UnrecordedCommits  []schema.CommitRef          `yaml:"unrecordedCommits"`
	Insights           []schema.SessionInsight     `yaml:"insights"`
	NilSlice           payloadNilSlice             `yaml:"nilSlice"`
}

type payloadValidationFixtureExpected struct {
	Valid                          bool                         `yaml:"valid"`
	ErrorContains                  string                       `yaml:"errorContains"`
	Mutation                       payloadMutationKind          `yaml:"mutation"`
	MutatedValid                   bool                         `yaml:"mutatedValid"`
	MutatedErrorContains           string                       `yaml:"mutatedErrorContains"`
	Repair                         payloadMutationKind          `yaml:"repair"`
	RepairedValid                  bool                         `yaml:"repairedValid"`
	SuccessorAssociationID         schema.AssociationID         `yaml:"successorAssociationId"`
	SuccessorAssociationConclusion schema.AssociationConclusion `yaml:"successorAssociationConclusion"`
	DuplicateSuccessorIndex        int                          `yaml:"duplicateSuccessorIndex"`
}

type payloadValidationManifest struct {
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

type payloadValidationSubject struct {
	mapNodeDetail *schema.MapNodeDetailPayload
	changeDetail  *schema.ChangeDetailPayload
	reviewList    *schema.ReviewListPayload
}

func decodePayloadValidationManifest(data []byte) (payloadValidationManifest, error) {
	var manifest payloadValidationManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return payloadValidationManifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return payloadValidationManifest{}, fmt.Errorf("decode trailing manifest document: %w", err)
		}
		return payloadValidationManifest{}, fmt.Errorf("decode manifest: multiple YAML documents are not allowed")
	}
	if manifest.ExpectedCaseCount <= 0 {
		return payloadValidationManifest{}, fmt.Errorf("expectedCaseCount must be positive, got %d", manifest.ExpectedCaseCount)
	}
	if len(manifest.RequiredCaseNames) != manifest.ExpectedCaseCount {
		return payloadValidationManifest{}, fmt.Errorf("manifest has %d requiredCaseNames, want expectedCaseCount %d", len(manifest.RequiredCaseNames), manifest.ExpectedCaseCount)
	}
	return manifest, nil
}

func loadPayloadValidationFixtures(t *testing.T) (testcase.Corpus[payloadValidationFixtureInput, payloadValidationFixtureExpected], payloadValidationManifest) {
	t.Helper()
	corpus, err := testcase.LoadCorpus[payloadValidationFixtureInput, payloadValidationFixtureExpected](mapNodeDetailPayloadCasesYAML)
	if err != nil {
		t.Fatalf("load map node detail payload corpus: %v", err)
	}
	manifest, err := decodePayloadValidationManifest(mapNodeDetailPayloadManifestYAML)
	if err != nil {
		t.Fatalf("load map node detail payload manifest: %v", err)
	}
	assert.RequireMin(t, corpus, manifest.ExpectedCaseCount)
	assert.RequireValid(t, corpus)
	if len(corpus.Cases) != manifest.ExpectedCaseCount {
		t.Fatalf("map node detail payload corpus has %d cases, want exactly %d", len(corpus.Cases), manifest.ExpectedCaseCount)
	}

	requiredNames := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" {
			t.Fatal("map node detail payload manifest contains an empty required case name")
		}
		if _, exists := requiredNames[name]; exists {
			t.Fatalf("map node detail payload manifest repeats required case name %q", name)
		}
		requiredNames[name] = struct{}{}
	}
	for _, fixture := range corpus.Cases {
		if _, exists := requiredNames[fixture.Name]; !exists {
			t.Fatalf("map node detail payload corpus contains unregistered case %q", fixture.Name)
		}
		delete(requiredNames, fixture.Name)
	}
	for missing := range requiredNames {
		t.Fatalf("map node detail payload corpus is missing required case %q", missing)
	}
	return corpus, manifest
}

func newPayloadValidationSubject(input payloadValidationFixtureInput) (*payloadValidationSubject, error) {
	switch input.Payload {
	case payloadValidationMapNodeDetail:
		payload := schema.NewMapNodeDetailPayload(input.Path)
		if input.RecentCommits != nil {
			payload.RecentCommits = append([]schema.CommitRef{}, input.RecentCommits...)
		} else {
			payload.RecentCommits = make([]schema.CommitRef, 0, len(input.RecentCommitHashes))
			for _, hash := range input.RecentCommitHashes {
				payload.RecentCommits = append(payload.RecentCommits, schema.NewCommitRef(hash, "fixture successor"))
			}
		}
		payload.RewrittenCommits = append([]schema.RewrittenCommit{}, input.RewrittenCommits...)
		payload.Insights = append([]schema.SessionInsight{}, input.Insights...)
		switch input.NilSlice {
		case "", payloadNilSliceNone:
		case payloadNilSliceInsights:
			payload.Insights = nil
		case payloadNilSliceRewrittenCommits:
			payload.RewrittenCommits = nil
		default:
			return nil, fmt.Errorf("map_node_detail fixture has unknown nilSlice %q", input.NilSlice)
		}
		return &payloadValidationSubject{mapNodeDetail: payload}, nil
	case payloadValidationChangeDetail:
		payload := schema.NewChangeDetailPayload(input.Path)
		if input.UnrecordedCommits != nil {
			payload.UnrecordedCommits = append([]schema.CommitRef{}, input.UnrecordedCommits...)
		}
		payload.Insights = append([]schema.SessionInsight{}, input.Insights...)
		switch input.NilSlice {
		case "", payloadNilSliceNone:
		case payloadNilSliceInsights:
			payload.Insights = nil
		default:
			return nil, fmt.Errorf("change_detail fixture has unsupported nilSlice %q", input.NilSlice)
		}
		return &payloadValidationSubject{changeDetail: payload}, nil
	case payloadValidationReviewList:
		payload := schema.NewReviewListPayload(payloadValidationProjectHash)
		payload.Sessions = append([]schema.TimelineSessionRef{}, input.Sessions...)
		if input.RecentCommits != nil {
			payload.RecentCommits = append([]schema.CommitRef{}, input.RecentCommits...)
		}
		payload.RewrittenCommits = append([]schema.RewrittenCommit{}, input.RewrittenCommits...)
		return &payloadValidationSubject{reviewList: payload}, nil
	default:
		return nil, fmt.Errorf("fixture has unknown payload %q", input.Payload)
	}
}

func (s *payloadValidationSubject) validate() error {
	if s.mapNodeDetail != nil {
		return s.mapNodeDetail.Validate()
	}
	if s.changeDetail != nil {
		return s.changeDetail.Validate()
	}
	if s.reviewList != nil {
		return s.reviewList.Validate()
	}
	return fmt.Errorf("payload validation fixture did not construct a payload")
}

func clearInsightClassifications(insights []schema.SessionInsight) {
	for index := range insights {
		insights[index].Classification = nil
	}
}

func (s *payloadValidationSubject) mutate(kind payloadMutationKind, associationID schema.AssociationID, conclusion schema.AssociationConclusion, duplicateSuccessorIndex int) error {
	switch kind {
	case payloadMutationRemoveSuccessorCommit:
		if s.mapNodeDetail == nil {
			return fmt.Errorf("mutation %q requires a map node detail payload", kind)
		}
		s.mapNodeDetail.RecentCommits = []schema.CommitRef{}
	case payloadMutationAddSuccessorCommit:
		if s.mapNodeDetail == nil || len(s.mapNodeDetail.RewrittenCommits) == 0 || s.mapNodeDetail.RewrittenCommits[0].SuccessorHash == nil {
			return fmt.Errorf("mutation %q requires a map node detail payload with a successor hash", kind)
		}
		hash := *s.mapNodeDetail.RewrittenCommits[0].SuccessorHash
		s.mapNodeDetail.RecentCommits = append(s.mapNodeDetail.RecentCommits, schema.NewCommitRef(hash, "fixture successor"))
	case payloadMutationRemoveDuplicateSuccessor:
		recentCommits, err := s.successorCommits()
		if err != nil {
			return fmt.Errorf("mutation %q: %w", kind, err)
		}
		if err := removeDuplicateSuccessorCommit(recentCommits, duplicateSuccessorIndex); err != nil {
			return fmt.Errorf("mutation %q: %w", kind, err)
		}
	case payloadMutationRepairSuccessorAssociation:
		recentCommits, err := s.successorCommits()
		if err != nil {
			return fmt.Errorf("mutation %q: %w", kind, err)
		}
		if err := repairSuccessorAssociationMirror(recentCommits, s.rewrittenCommits()); err != nil {
			return fmt.Errorf("mutation %q: %w", kind, err)
		}
	case payloadMutationDriftSuccessorAssociation:
		recentCommits, err := s.successorCommits()
		if err != nil {
			return fmt.Errorf("mutation %q: %w", kind, err)
		}
		if err := driftSuccessorAssociationConclusion(recentCommits, associationID, conclusion); err != nil {
			return fmt.Errorf("mutation %q: %w", kind, err)
		}
	case payloadMutationRepairCommitRefShape:
		if s.mapNodeDetail != nil {
			if len(s.mapNodeDetail.RecentCommits) == 0 {
				return fmt.Errorf("mutation %q requires a map node detail payload with a recent commit", kind)
			}
			repairCommitRefShape(&s.mapNodeDetail.RecentCommits[0])
		} else if s.changeDetail != nil {
			if len(s.changeDetail.UnrecordedCommits) == 0 {
				return fmt.Errorf("mutation %q requires a change detail payload with an unrecorded commit", kind)
			}
			repairCommitRefShape(&s.changeDetail.UnrecordedCommits[0])
		} else {
			return fmt.Errorf("mutation %q requires a payload", kind)
		}
	case payloadMutationClearClassification:
		if s.mapNodeDetail != nil {
			clearInsightClassifications(s.mapNodeDetail.Insights)
		} else if s.changeDetail != nil {
			clearInsightClassifications(s.changeDetail.Insights)
		} else {
			return fmt.Errorf("mutation %q requires a payload", kind)
		}
	case payloadMutationNilInsights:
		if s.mapNodeDetail != nil {
			s.mapNodeDetail.Insights = nil
		} else if s.changeDetail != nil {
			s.changeDetail.Insights = nil
		} else {
			return fmt.Errorf("mutation %q requires a payload", kind)
		}
	case payloadMutationInitializeInsights:
		if s.mapNodeDetail != nil {
			s.mapNodeDetail.Insights = []schema.SessionInsight{}
		} else if s.changeDetail != nil {
			s.changeDetail.Insights = []schema.SessionInsight{}
		} else {
			return fmt.Errorf("mutation %q requires a payload", kind)
		}
	case payloadMutationInitializeRewrittenCommits:
		if s.mapNodeDetail == nil {
			return fmt.Errorf("mutation %q requires a map node detail payload", kind)
		}
		s.mapNodeDetail.RewrittenCommits = []schema.RewrittenCommit{}
	default:
		return fmt.Errorf("unknown payload mutation %q", kind)
	}
	return nil
}

const payloadValidationProjectHash schema.ProjectHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func (s *payloadValidationSubject) successorCommits() (*[]schema.CommitRef, error) {
	switch {
	case s.mapNodeDetail != nil:
		return &s.mapNodeDetail.RecentCommits, nil
	case s.reviewList != nil:
		return &s.reviewList.RecentCommits, nil
	default:
		return nil, fmt.Errorf("requires a map node detail or review list payload")
	}
}

func (s *payloadValidationSubject) rewrittenCommits() []schema.RewrittenCommit {
	if s.mapNodeDetail != nil {
		return s.mapNodeDetail.RewrittenCommits
	}
	if s.reviewList != nil {
		return s.reviewList.RewrittenCommits
	}
	return nil
}

func removeDuplicateSuccessorCommit(recentCommits *[]schema.CommitRef, duplicateSuccessorIndex int) error {
	if duplicateSuccessorIndex < 0 || duplicateSuccessorIndex >= len(*recentCommits) {
		return fmt.Errorf("duplicate successor index %d is outside %d recentCommits", duplicateSuccessorIndex, len(*recentCommits))
	}
	targetHash := (*recentCommits)[duplicateSuccessorIndex].Hash
	if targetHash == "" {
		return fmt.Errorf("recentCommits[%d] has an empty hash and is not a duplicate successor authority", duplicateSuccessorIndex)
	}
	occurrences := 0
	for _, commit := range *recentCommits {
		if commit.Hash == targetHash {
			occurrences++
		}
	}
	if occurrences < 2 {
		return fmt.Errorf("recentCommits[%d] hash %q is not duplicated", duplicateSuccessorIndex, targetHash)
	}
	*recentCommits = append((*recentCommits)[:duplicateSuccessorIndex], (*recentCommits)[duplicateSuccessorIndex+1:]...)
	return nil
}

func repairSuccessorAssociationMirror(recentCommits *[]schema.CommitRef, rewrittenCommits []schema.RewrittenCommit) error {
	successors := indexPayloadSuccessorAssociations(*recentCommits)
	for rewrittenIndex := range rewrittenCommits {
		rewritten := rewrittenCommits[rewrittenIndex]
		if rewritten.SuccessorHash == nil {
			continue
		}
		successor, exists := successors[*rewritten.SuccessorHash]
		if !exists {
			return fmt.Errorf("rewrittenCommits[%d] successorHash %q is absent from recentCommits; cannot restore the declared mirror", rewrittenIndex, *rewritten.SuccessorHash)
		}
		for _, ledgerAssociation := range rewritten.Associations {
			associationIndex, hasAssociationID := successor.byID[ledgerAssociation.ID]
			sessionAssociationIndex, hasSessionID := successor.bySessionID[ledgerAssociation.SessionID]
			if hasAssociationID && hasSessionID && associationIndex != sessionAssociationIndex {
				displacedAssociation := successor.commit.Associations[associationIndex]
				displacedSessionID := displacedAssociation.SessionID
				successor.commit.Associations[associationIndex] = ledgerAssociation
				successor.commit.SessionIDs[associationIndex] = ledgerAssociation.SessionID
				displacedSessionAssociation := successor.commit.Associations[sessionAssociationIndex]
				displacedSessionAssociation.SessionID = displacedSessionID
				successor.commit.Associations[sessionAssociationIndex] = displacedSessionAssociation
				successor.commit.SessionIDs[sessionAssociationIndex] = displacedSessionID
				return nil
			}
			if !hasAssociationID {
				associationIndex = sessionAssociationIndex
			}
			exists := hasAssociationID || hasSessionID
			if !exists {
				return fmt.Errorf("recentCommits successorHash %q has no association sharing associationId %q or sessionId %q with rewrittenCommits[%d]; cannot restore the declared mirror", successor.commit.Hash, ledgerAssociation.ID, ledgerAssociation.SessionID, rewrittenIndex)
			}
			if reflect.DeepEqual(successor.commit.Associations[associationIndex], ledgerAssociation) {
				continue
			}
			successor.commit.Associations[associationIndex] = ledgerAssociation
			successor.commit.SessionIDs[associationIndex] = ledgerAssociation.SessionID
			return nil
		}
	}
	return fmt.Errorf("payload has no rewritten commit with a displayed successor; cannot restore a successor association mirror")
}

type payloadSuccessorAssociationIndex struct {
	commit      *schema.CommitRef
	byID        map[schema.AssociationID]int
	bySessionID map[schema.SessionID]int
}

func indexPayloadSuccessorAssociations(commits []schema.CommitRef) map[string]payloadSuccessorAssociationIndex {
	indexed := make(map[string]payloadSuccessorAssociationIndex, len(commits))
	for commitIndex := range commits {
		commit := &commits[commitIndex]
		byID := make(map[schema.AssociationID]int, len(commit.Associations))
		bySessionID := make(map[schema.SessionID]int, len(commit.Associations))
		for associationIndex, association := range commit.Associations {
			byID[association.ID] = associationIndex
			bySessionID[association.SessionID] = associationIndex
		}
		indexed[commit.Hash] = payloadSuccessorAssociationIndex{commit: commit, byID: byID, bySessionID: bySessionID}
	}
	return indexed
}

func driftSuccessorAssociationConclusion(recentCommits *[]schema.CommitRef, associationID schema.AssociationID, conclusion schema.AssociationConclusion) error {
	if err := associationID.Validate(); err != nil {
		return fmt.Errorf("successor association ID: %w", err)
	}
	if err := conclusion.Validate(); err != nil {
		return fmt.Errorf("successor association conclusion: %w", err)
	}
	for _, successor := range indexPayloadSuccessorAssociations(*recentCommits) {
		if associationIndex, exists := successor.byID[associationID]; exists {
			successor.commit.Associations[associationIndex].Conclusion = conclusion
			return nil
		}
	}
	return fmt.Errorf("recentCommits has no associationId %q to drift", associationID)
}

func repairCommitRefShape(commit *schema.CommitRef) {
	commit.HasSession = len(commit.SessionIDs) > 0
	if len(commit.SessionIDs) == 0 || len(commit.Associations) != len(commit.SessionIDs) {
		return
	}
	bySessionID := make(map[schema.SessionID]schema.SessionAssociation, len(commit.Associations))
	for _, association := range commit.Associations {
		bySessionID[association.SessionID] = association
	}
	for index, sessionID := range commit.SessionIDs {
		if association, exists := bySessionID[sessionID]; exists {
			commit.Associations[index] = association
		}
	}
}

func assertSinglePayloadContextPrefix(t *testing.T, err error, prefix string) {
	t.Helper()
	if got := strings.Count(err.Error(), prefix); got != 1 {
		t.Fatalf("Validate() error=%q, want exactly one %q prefix", err, prefix)
	}
}

func assertCommitRefRepairPreservesBindingIdentity(t *testing.T, before schema.CommitRef, beforeSessionIDs []schema.SessionID, beforeAssociations map[schema.SessionID]schema.SessionAssociation, after schema.CommitRef) {
	t.Helper()
	if before.Hash != after.Hash || before.Subject != after.Subject {
		t.Fatalf("repair changed commit identity: before hash=%q subject=%q, after hash=%q subject=%q", before.Hash, before.Subject, after.Hash, after.Subject)
	}
	if !reflect.DeepEqual(beforeSessionIDs, after.SessionIDs) {
		t.Fatalf("repair changed sessionIds: before=%v after=%v", beforeSessionIDs, after.SessionIDs)
	}
	if after.HasSession != (len(after.SessionIDs) > 0) {
		t.Fatalf("repair left hasSession=%v for %d sessionIds", after.HasSession, len(after.SessionIDs))
	}
	afterAssociationsBySessionID := sessionAssociationsBySessionID(after)
	if !reflect.DeepEqual(beforeAssociations, afterAssociationsBySessionID) {
		t.Fatalf("repair changed session associations by sessionId: before=%v after=%v", beforeAssociations, afterAssociationsBySessionID)
	}
	afterAssociationSessionIDs := make([]schema.SessionID, len(after.Associations))
	for index, association := range after.Associations {
		afterAssociationSessionIDs[index] = association.SessionID
	}
	if !reflect.DeepEqual(afterAssociationSessionIDs, after.SessionIDs) {
		t.Fatalf("repair left associations out of sync with sessionIds: associations=%v sessionIds=%v", afterAssociationSessionIDs, after.SessionIDs)
	}
}

func sessionAssociationsBySessionID(commit schema.CommitRef) map[schema.SessionID]schema.SessionAssociation {
	bySessionID := make(map[schema.SessionID]schema.SessionAssociation, len(commit.Associations))
	for _, association := range commit.Associations {
		bySessionID[association.SessionID] = association
	}
	return bySessionID
}

func assertPayloadValidationResult(t *testing.T, label string, err error, valid bool, errorContains string) {
	t.Helper()
	if (err == nil) != valid {
		t.Fatalf("%s: Validate() error=%v, want valid=%v", label, err, valid)
	}
	if valid {
		if errorContains != "" {
			t.Fatalf("%s: valid expectation must not declare errorContains %q", label, errorContains)
		}
		return
	}
	if strings.TrimSpace(errorContains) == "" {
		t.Fatalf("%s: rejected expectation must name an actionable errorContains value", label)
	}
	if !strings.Contains(err.Error(), errorContains) {
		t.Fatalf("%s: Validate() error=%q, want substring %q", label, err, errorContains)
	}
}

// TestMapNodeDetailPayload_FixtureContract drives the MapNodeDetailPayload,
// ChangeDetailPayload, and ReviewListPayload validators through the typed
// corpus. Every case also applies its fixture-declared mutation and requires
// the expected validity transition, proving the positive and negative
// assertions are non-vacuous.
func TestMapNodeDetailPayload_FixtureContract(t *testing.T) {
	corpus, _ := loadPayloadValidationFixtures(t)
	for _, fixture := range corpus.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			wantPass := fixture.Classification == testcase.MustPass
			if fixture.Expected.Valid != wantPass {
				t.Fatalf("classification=%q disagrees with expected.valid=%v", fixture.Classification, fixture.Expected.Valid)
			}
			if fixture.Expected.MutatedValid == fixture.Expected.Valid {
				t.Fatalf("mutation %q does not change expected validity; mutation proof would be vacuous", fixture.Expected.Mutation)
			}

			subject, err := newPayloadValidationSubject(fixture.Input)
			if err != nil {
				t.Fatalf("construct fixture payload: %v", err)
			}
			var originalCommit schema.CommitRef
			var originalSessionIDs []schema.SessionID
			var originalAssociations map[schema.SessionID]schema.SessionAssociation
			var prefix string
			if fixture.Expected.Mutation == payloadMutationRepairCommitRefShape {
				switch {
				case subject.mapNodeDetail != nil:
					if len(subject.mapNodeDetail.RecentCommits) == 0 {
						t.Fatal("repair proof requires a recent commit")
					}
					originalCommit = subject.mapNodeDetail.RecentCommits[0]
					originalSessionIDs = append([]schema.SessionID(nil), originalCommit.SessionIDs...)
					originalAssociations = sessionAssociationsBySessionID(originalCommit)
					prefix = "map node detail validation:"
				case subject.changeDetail != nil:
					if len(subject.changeDetail.UnrecordedCommits) == 0 {
						t.Fatal("repair proof requires an unrecorded commit")
					}
					originalCommit = subject.changeDetail.UnrecordedCommits[0]
					originalSessionIDs = append([]schema.SessionID(nil), originalCommit.SessionIDs...)
					originalAssociations = sessionAssociationsBySessionID(originalCommit)
					prefix = "change detail validation:"
				}
			}
			originalErr := subject.validate()
			assertPayloadValidationResult(t, "original", originalErr, fixture.Expected.Valid, fixture.Expected.ErrorContains)
			if fixture.Expected.Mutation == payloadMutationRepairCommitRefShape {
				assertSinglePayloadContextPrefix(t, originalErr, prefix)
			}
			if err := subject.mutate(fixture.Expected.Mutation, fixture.Expected.SuccessorAssociationID, fixture.Expected.SuccessorAssociationConclusion, fixture.Expected.DuplicateSuccessorIndex); err != nil {
				t.Fatalf("apply mutation: %v", err)
			}
			mutatedErr := subject.validate()
			assertPayloadValidationResult(t, "mutated", mutatedErr, fixture.Expected.MutatedValid, fixture.Expected.MutatedErrorContains)
			if fixture.Expected.Repair != "" {
				if !fixture.Expected.Valid || fixture.Expected.MutatedValid || !fixture.Expected.RepairedValid {
					t.Fatalf("repair %q must prove valid -> invalid -> valid, got original=%v mutated=%v repaired=%v", fixture.Expected.Repair, fixture.Expected.Valid, fixture.Expected.MutatedValid, fixture.Expected.RepairedValid)
				}
				if err := subject.mutate(fixture.Expected.Repair, "", "", 0); err != nil {
					t.Fatalf("apply repair %q: %v", fixture.Expected.Repair, err)
				}
				assertPayloadValidationResult(t, "repaired", subject.validate(), fixture.Expected.RepairedValid, "")
			}
			if fixture.Expected.Mutation == payloadMutationRepairCommitRefShape {
				var repairedCommit schema.CommitRef
				switch {
				case subject.mapNodeDetail != nil:
					repairedCommit = subject.mapNodeDetail.RecentCommits[0]
				case subject.changeDetail != nil:
					repairedCommit = subject.changeDetail.UnrecordedCommits[0]
				}
				assertCommitRefRepairPreservesBindingIdentity(t, originalCommit, originalSessionIDs, originalAssociations, repairedCommit)
			}
		})
	}
}
