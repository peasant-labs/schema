package schema_test

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

func validSuccessorHash() *string {
	h := "successor-1"
	return &h
}

func TestMapNodeDetailPayload_Validate_GhostSuccessorMustBeInRecentCommits(t *testing.T) {
	payload := schema.NewMapNodeDetailPayload("internal/api")
	payload.RecentCommits = []schema.CommitRef{schema.NewCommitRef("successor-1", "fix")}
	payload.RewrittenCommits = []schema.RewrittenCommit{{
		GhostHash:     "ghost-1",
		SessionIDs:    []schema.SessionID{},
		SuccessorHash: validSuccessorHash(),
		Resolution:    schema.RewriteResolutionRewritten,
		Method:        schema.RewriteMethodPatchID,
		Confidence:    schema.ConfidenceHigh,
	}}
	if err := payload.Validate(); err != nil {
		t.Fatalf("Validate() with successor present in RecentCommits: %v", err)
	}

	payload.RecentCommits = []schema.CommitRef{}
	err := payload.Validate()
	if err == nil || !strings.Contains(err.Error(), "not present in recentCommits") {
		t.Fatalf("Validate() with successor absent from RecentCommits: error=%v, want a not-present-in-recentCommits rejection", err)
	}
}

func TestMapNodeDetailPayload_Validate_InsightClassificationMustBeNil(t *testing.T) {
	payload := schema.NewMapNodeDetailPayload("internal/api")
	category := "architecture"
	payload.Insights = []schema.SessionInsight{{
		Kind:       schema.InsightKindDecision,
		Provenance: schema.InsightProvenanceMechanical,
		Confidence: schema.ConfidenceHigh,
		Title:      "example",
		Subjects:   []string{"internal/api"},
		Evidence:   []schema.InsightEvidence{{SessionID: "session-a"}},
		Classification: &schema.InsightClassification{
			Category: category,
		},
	}}
	err := payload.Validate()
	if err == nil || !strings.Contains(err.Error(), "classification is non-nil") {
		t.Fatalf("Validate() with a populated Classification: error=%v, want a classification-is-non-nil rejection", err)
	}
}

func TestMapNodeDetailPayload_Validate_RejectsNilSlices(t *testing.T) {
	payload := schema.NewMapNodeDetailPayload("internal/api")
	payload.RewrittenCommits = nil
	err := payload.Validate()
	if err == nil || !strings.Contains(err.Error(), "must be arrays") {
		t.Fatalf("Validate() with nil RewrittenCommits: error=%v, want a must-be-arrays rejection", err)
	}
}

func TestChangeDetailPayload_Validate_InsightClassificationMustBeNil(t *testing.T) {
	payload := schema.NewChangeDetailPayload("feat/x")
	payload.Insights = []schema.SessionInsight{{
		Kind:       schema.InsightKindFriction,
		Provenance: schema.InsightProvenanceMined,
		Confidence: schema.ConfidenceMedium,
		Title:      "example",
		Subjects:   []string{"internal/api"},
		Evidence:   []schema.InsightEvidence{},
		Classification: &schema.InsightClassification{
			Category: "process",
		},
	}}
	err := payload.Validate()
	if err == nil || !strings.Contains(err.Error(), "classification is non-nil") {
		t.Fatalf("Validate() with a populated Classification: error=%v, want a classification-is-non-nil rejection", err)
	}

	payload.Insights[0].Classification = nil
	if err := payload.Validate(); err != nil {
		t.Fatalf("Validate() with Classification cleared: %v", err)
	}
}

func TestChangeDetailPayload_Validate_RejectsNilInsights(t *testing.T) {
	payload := schema.NewChangeDetailPayload("feat/x")
	payload.Insights = nil
	err := payload.Validate()
	if err == nil || !strings.Contains(err.Error(), "insights must be an array") {
		t.Fatalf("Validate() with nil Insights: error=%v, want a must-be-an-array rejection", err)
	}
}
