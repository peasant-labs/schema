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
	if err := fixtures.CheckMin(20); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Cases) != 20 {
		t.Fatalf("timeline fixture has %d cases, want exactly 20", len(fixtures.Cases))
	}
	for _, fixture := range fixtures.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			payload := schema.NewReviewListPayload(canonicalTimelineProjectHash)
			payload.Sessions = fixture.Input.Sessions
			payload.RecentCommits = fixture.Input.Commits
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
			default:
				t.Fatalf("unsupported classification %q", fixture.Classification)
			}
		})
	}
}
