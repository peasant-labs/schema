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
	if err := fixtures.CheckMin(23); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Cases) != 23 {
		t.Fatalf("timeline fixture has %d cases, want exactly 23", len(fixtures.Cases))
	}
	for _, fixture := range fixtures.Cases {
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
					if err := fixture.Expected.Repair.Apply(payload); err != nil {
						t.Fatalf("Apply repair: %v", err)
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
