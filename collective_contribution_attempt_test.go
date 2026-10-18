package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/collective_contribution_attempts.yaml
var contributionAttemptFixtureYAML []byte

type contributionAttemptInput struct {
	CollectiveID   schema.VillageUUID            `yaml:"collective_id"`
	TranscriptID   schema.TranscriptID           `yaml:"transcript_id"`
	EventNum       int32                         `yaml:"event_num"`
	Status         schema.VillageShareStatus     `yaml:"status"`
	RecordedAt     time.Time                     `yaml:"recorded_at"`
	DecidedAt      *time.Time                    `yaml:"decided_at"`
	DecidedByActor schema.VillageShareEventActor `yaml:"decided_by_actor"`
	FirstSeenAt    time.Time                     `yaml:"first_seen_at"`
}

func (i contributionAttemptInput) value() schema.LocalRemoteContributionAttempt {
	return schema.LocalRemoteContributionAttempt{
		CollectiveID: i.CollectiveID, TranscriptID: i.TranscriptID, FirstSeenAt: i.FirstSeenAt,
		Attempt: schema.VillageShareEvent{EventNum: i.EventNum, Status: i.Status, RecordedAt: i.RecordedAt, DecidedAt: i.DecidedAt, DecidedByActor: i.DecidedByActor},
	}
}

type contributionReconciliationInput struct {
	Saved    *contributionAttemptInput `yaml:"saved"`
	Incoming contributionAttemptInput  `yaml:"incoming"`
}

type contributionReconciliationExpected struct {
	Disposition schema.VillageContributionAttemptReconciliation `yaml:"disposition"`
	Status      schema.VillageShareStatus                       `yaml:"status"`
}

func loadContributionAttemptFixtures(t *testing.T) testcase.Corpus[contributionReconciliationInput, contributionReconciliationExpected] {
	t.Helper()
	var fixtures struct {
		RequiredNames []string                                                                             `yaml:"required_names"`
		Corpus        testcase.Corpus[contributionReconciliationInput, contributionReconciliationExpected] `yaml:"corpus"`
	}
	if err := yaml.Unmarshal(contributionAttemptFixtureYAML, &fixtures); err != nil {
		t.Fatal(err)
	}
	assertcase.RequireValid(t, fixtures.Corpus)
	assertcase.RequireMin(t, fixtures.Corpus, 1)
	requireNames(t, "contribution attempts", fixtures.RequiredNames, namesOf(fixtures.Corpus))
	// Guard the manifest itself so removing a case and its YAML manifest entry
	// together cannot silently remove a required behavior.
	requireNames(t, "required contribution attempt behaviors", []string{
		"attempt-new-pending", "attempt-pending-approved", "attempt-pending-rejected",
		"attempt-pending-retracted", "attempt-pending-revoked", "attempt-late-pending-after-terminal",
		"attempt-terminal-exact-replay", "attempt-recorded-at-conflict", "attempt-terminal-conflict",
		"attempt-terminal-decision-conflict", "attempt-approved-withdrawal-new-ordinal",
		"attempt-initial-auto-approved-null-decision", "attempt-moderator-deletion-wire-unchanged",
		"attempt-identity-conflict", "attempt-transcript-identity-conflict", "attempt-ordinal-identity-conflict",
		"attempt-pending-decision-missing", "attempt-malformed-stale-pending", "attempt-invalid-actor",
		"attempt-invalid-status", "attempt-invalid-ordinal", "attempt-invalid-recorded-time",
		"attempt-invalid-first-seen-time", "attempt-invalid-collective-id", "attempt-terminal-decision-required",
		"attempt-invalid-saved-evidence", "attempt-equivalent-poll-retains-first-seen",
		"attempt-terminal-null-decision-conflict", "attempt-pending-exact-replay",
		"attempt-terminal-submillisecond-conflict", "attempt-terminal-fraction-equivalent-replay",
	}, namesOf(fixtures.Corpus))
	return fixtures.Corpus
}

func TestReconcileVillageContributionAttempt(t *testing.T) {
	for _, c := range loadContributionAttemptFixtures(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var saved *schema.LocalRemoteContributionAttempt
			if c.Input.Saved != nil {
				value := c.Input.Saved.value()
				saved = &value
			}
			incoming := c.Input.Incoming.value()
			before, err := json.Marshal(struct {
				Saved    *schema.LocalRemoteContributionAttempt
				Incoming schema.LocalRemoteContributionAttempt
			}{saved, incoming})
			if err != nil {
				t.Fatal(err)
			}
			got, disposition, err := schema.ReconcileVillageContributionAttempt(saved, incoming)
			valid := c.Classification == testcase.MustPass
			if (err == nil) != valid {
				t.Fatalf("valid=%v disposition=%s error=%v", valid, disposition, err)
			}
			if got.Attempt.Status != c.Expected.Status || disposition != c.Expected.Disposition {
				t.Fatalf("status=%s disposition=%s, want %+v", got.Attempt.Status, disposition, c.Expected)
			}
			after, marshalErr := json.Marshal(struct {
				Saved    *schema.LocalRemoteContributionAttempt
				Incoming schema.LocalRemoteContributionAttempt
			}{saved, incoming})
			if marshalErr != nil || string(before) != string(after) {
				t.Fatalf("reconciliation mutated inputs: %v", marshalErr)
			}
			if saved != nil && (disposition != schema.VillageContributionAttemptUpdated) && !reflect.DeepEqual(got, *saved) {
				t.Fatalf("retained evidence changed: got %+v want %+v", got, *saved)
			}
			if valid {
				want := incoming
				if saved != nil {
					want.FirstSeenAt = saved.FirstSeenAt
					if disposition != schema.VillageContributionAttemptUpdated {
						want = *saved
					}
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("result lost evidence: got %+v want %+v", got, want)
				}
				if got.Attempt.DecidedAt != nil {
					if got.Attempt.DecidedAt == incoming.Attempt.DecidedAt || (saved != nil && got.Attempt.DecidedAt == saved.Attempt.DecidedAt) {
						t.Fatal("result aliases an input decision pointer")
					}
				}
			}
		})
	}
}
