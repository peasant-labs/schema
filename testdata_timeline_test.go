package schema_test

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
)

const canonicalTimelineProjectHash schema.ProjectHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestTimelineFixturesValidateRelationships(t *testing.T) {
	fixtures, err := schema.LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	if err := fixtures.CheckMin(24); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Cases) != 24 {
		t.Fatalf("timeline fixture has %d cases, want exactly 24", len(fixtures.Cases))
	}
	if len(fixtures.SuccessorAssociationMirrorCases) != 5 {
		t.Fatalf("timeline successor association mirror fixture has %d cases, want exactly 5", len(fixtures.SuccessorAssociationMirrorCases))
	}
	validateTimelineFixtureRelationships(t, fixtures.Cases)
	validateTimelineFixtureRelationships(t, fixtures.SuccessorAssociationMirrorCases)
}

func validateTimelineFixtureRelationships(t *testing.T, fixtures []schema.TimelineFixtureCase) {
	t.Helper()
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			payload := reviewListPayloadFromTimelineFixture(fixture)
			err := payload.Validate()
			switch fixture.Classification {
			case testcase.MustPass:
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
			case testcase.MustFail:
				if err == nil || !strings.Contains(err.Error(), fixture.Expected.ErrorContains) {
					t.Fatalf("Validate error = %v, want containing %q", err, fixture.Expected.ErrorContains)
				}
				if fixture.Expected.Repair != nil {
					requireActionableValidationError(t, err, fixture.Expected.ErrorContains)
					var ledgerTargetSessionID schema.SessionID
					if fixture.Name == "rewrite_ledger_reference_requires_binding_truth" {
						ledgerTargetSessionID = requireSingleRewriteLedgerSessionID(t, fixture.Input.RewrittenCommits)
						if fixture.Expected.Repair.SessionID != ledgerTargetSessionID {
							t.Fatalf("repair sessionId=%q, want ledger-referenced sessionId %q", fixture.Expected.Repair.SessionID, ledgerTargetSessionID)
						}
						before := requireTimelineSession(t, payload.Sessions, ledgerTargetSessionID)
						if before.HasCommitBinding {
							t.Fatalf("ledger-referenced session %q hasCommitBinding=true before repair, want false", ledgerTargetSessionID)
						}
					}
					if err := fixture.Expected.Repair.Apply(payload); err != nil {
						t.Fatalf("Apply repair: %v", err)
					}
					if ledgerTargetSessionID != "" {
						after := requireTimelineSession(t, payload.Sessions, ledgerTargetSessionID)
						if !after.HasCommitBinding {
							t.Fatalf("ledger-referenced session %q hasCommitBinding=false after repair, want true", ledgerTargetSessionID)
						}
					}
					repairedErr := payload.Validate()
					if (repairedErr == nil) != fixture.Expected.Repair.PostMutationValid {
						t.Fatalf("repaired Validate error = %v, want valid=%v", repairedErr, fixture.Expected.Repair.PostMutationValid)
					}
				}
				if fixture.Name == "rewrite_ledger_references_missing_session" {
					requireActionableValidationError(t, err, fixture.Expected.ErrorContains)
				}
				if fixture.Name == "null_associations" {
					assertSinglePayloadContextPrefix(t, err, "review list validation:")
				}
			default:
				t.Fatalf("unsupported classification %q", fixture.Classification)
			}
		})
	}
}

func reviewListPayloadFromTimelineFixture(fixture schema.TimelineFixtureCase) *schema.ReviewListPayload {
	payload := schema.NewReviewListPayload(canonicalTimelineProjectHash)
	payload.Sessions = fixture.Input.Sessions
	payload.RecentCommits = fixture.Input.Commits
	if fixture.Input.RewrittenCommits != nil {
		payload.RewrittenCommits = fixture.Input.RewrittenCommits
	}
	return payload
}

func requireSingleRewriteLedgerSessionID(t *testing.T, rewrittenCommits []schema.RewrittenCommit) schema.SessionID {
	t.Helper()
	var sessionID schema.SessionID
	count := 0
	for _, rewrittenCommit := range rewrittenCommits {
		for _, referencedSessionID := range rewrittenCommit.SessionIDs {
			sessionID = referencedSessionID
			count++
		}
	}
	if count != 1 {
		t.Fatalf("rewrite ledger references %d session IDs, want exactly one repair target", count)
	}
	return sessionID
}

func requireTimelineSession(t *testing.T, sessions []schema.TimelineSessionRef, sessionID schema.SessionID) schema.TimelineSessionRef {
	t.Helper()
	for _, session := range sessions {
		if session.SessionID == sessionID {
			return session
		}
	}
	t.Fatalf("timeline session %q not found", sessionID)
	return schema.TimelineSessionRef{}
}
