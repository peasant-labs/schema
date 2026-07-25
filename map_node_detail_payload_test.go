package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
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
	payloadMutationRepairCommitRefShape       payloadMutationKind = "repair_commit_ref_shape"
	payloadMutationClearClassification        payloadMutationKind = "clear_classification"
	payloadMutationNilInsights                payloadMutationKind = "nil_insights"
	payloadMutationInitializeInsights         payloadMutationKind = "initialize_insights"
	payloadMutationInitializeRewrittenCommits payloadMutationKind = "initialize_rewritten_commits"
)

type payloadValidationFixtureInput struct {
	Payload            payloadValidationKind    `yaml:"payload"`
	Path               string                   `yaml:"path"`
	RecentCommitHashes []string                 `yaml:"recentCommitHashes"`
	RecentCommits      []schema.CommitRef       `yaml:"recentCommits"`
	RewrittenCommits   []schema.RewrittenCommit `yaml:"rewrittenCommits"`
	UnrecordedCommits  []schema.CommitRef       `yaml:"unrecordedCommits"`
	Insights           []schema.SessionInsight  `yaml:"insights"`
	NilSlice           payloadNilSlice          `yaml:"nilSlice"`
}

type payloadValidationFixtureExpected struct {
	Valid                bool                `yaml:"valid"`
	ErrorContains        string              `yaml:"errorContains"`
	Mutation             payloadMutationKind `yaml:"mutation"`
	MutatedValid         bool                `yaml:"mutatedValid"`
	MutatedErrorContains string              `yaml:"mutatedErrorContains"`
}

type payloadValidationManifest struct {
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

type payloadValidationSubject struct {
	mapNodeDetail *schema.MapNodeDetailPayload
	changeDetail  *schema.ChangeDetailPayload
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
	return fmt.Errorf("payload validation fixture did not construct a payload")
}

func clearInsightClassifications(insights []schema.SessionInsight) {
	for index := range insights {
		insights[index].Classification = nil
	}
}

func (s *payloadValidationSubject) mutate(kind payloadMutationKind) error {
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
	case payloadMutationRepairCommitRefShape:
		if s.mapNodeDetail != nil {
			if len(s.mapNodeDetail.RecentCommits) == 0 {
				return fmt.Errorf("mutation %q requires a map node detail payload with a recent commit", kind)
			}
			commit := s.mapNodeDetail.RecentCommits[0]
			s.mapNodeDetail.RecentCommits[0] = schema.NewCommitRef(commit.Hash, commit.Subject)
		} else if s.changeDetail != nil {
			if len(s.changeDetail.UnrecordedCommits) == 0 {
				return fmt.Errorf("mutation %q requires a change detail payload with an unrecorded commit", kind)
			}
			commit := s.changeDetail.UnrecordedCommits[0]
			s.changeDetail.UnrecordedCommits[0] = schema.NewCommitRef(commit.Hash, commit.Subject)
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

// TestMapNodeDetailPayload_FixtureContract drives the MapNodeDetailPayload and
// ChangeDetailPayload validators through the typed corpus. Every case also
// applies its fixture-declared mutation and requires the expected validity
// transition, proving the positive and negative assertions are non-vacuous.
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
			assertPayloadValidationResult(t, "original", subject.validate(), fixture.Expected.Valid, fixture.Expected.ErrorContains)
			if err := subject.mutate(fixture.Expected.Mutation); err != nil {
				t.Fatalf("apply mutation: %v", err)
			}
			assertPayloadValidationResult(t, "mutated", subject.validate(), fixture.Expected.MutatedValid, fixture.Expected.MutatedErrorContains)
		})
	}
}
