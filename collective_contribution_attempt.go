package schema

import (
	"fmt"
	"time"
)

// LocalRemoteContributionAttempt is retained owner-history evidence. Attempt is
// a projection of a server attempt, not an immutable status-change event.
// The enclosing storage scope supplies the Village URL and authenticated user.
type LocalRemoteContributionAttempt struct {
	CollectiveID VillageUUID       `json:"collectiveId"`
	TranscriptID TranscriptID      `json:"transcriptId"`
	Attempt      VillageShareEvent `json:"attempt"`
	FirstSeenAt  time.Time         `json:"firstSeenAt"`
}

// VillageContributionAttemptReconciliation describes the effect on retained
// evidence. Superseded is a valid older pending observation, not an error.
type VillageContributionAttemptReconciliation string

const (
	VillageContributionAttemptInserted   VillageContributionAttemptReconciliation = "inserted"
	VillageContributionAttemptUnchanged  VillageContributionAttemptReconciliation = "unchanged"
	VillageContributionAttemptUpdated    VillageContributionAttemptReconciliation = "updated"
	VillageContributionAttemptSuperseded VillageContributionAttemptReconciliation = "superseded"
)

// VillageContributionAttemptActor returns the status's actor capacity, not a
// moderator identity or proof that manual review occurred. Initial automatic
// approval has moderator capacity even without a decision timestamp.
func VillageContributionAttemptActor(status VillageShareStatus) (VillageShareEventActor, error) {
	switch status {
	case VillageShareStatusPending:
		return VillageShareEventActorNone, nil
	case VillageShareStatusApproved, VillageShareStatusRejected:
		return VillageShareEventActorModerator, nil
	case VillageShareStatusRetracted:
		return VillageShareEventActorOwner, nil
	case VillageShareStatusRevoked:
		return VillageShareEventActorCollective, nil
	default:
		return "", attemptValidationError("status is outside the share-attempt status menu")
	}
}

// Validate checks the historical wire payload before reconciliation or emission.
// Approved attempts may have a null decided_at when initially auto-approved.
func (event VillageShareEvent) Validate() error {
	actor, err := VillageContributionAttemptActor(event.Status)
	if err != nil {
		return err
	}
	if event.EventNum < 1 {
		return attemptValidationError("event_num must be a positive int32 ordinal")
	}
	if !attemptUTCTime(event.RecordedAt) {
		return attemptValidationError("recorded_at must be a nonzero UTC timestamp")
	}
	if event.DecidedByActor != actor {
		return attemptValidationError("decided_by_actor does not match the status's actor capacity")
	}
	if event.Status == VillageShareStatusPending {
		if event.DecidedAt != nil {
			return attemptValidationError("pending attempts must have null decided_at")
		}
		return nil
	}
	if event.DecidedAt == nil {
		if event.Status != VillageShareStatusApproved {
			return attemptValidationError("this terminal status requires decided_at")
		}
	} else if !attemptUTCTime(*event.DecidedAt) {
		return attemptValidationError("decided_at must be a nonzero UTC timestamp when present")
	}
	return nil
}

func (attempt LocalRemoteContributionAttempt) Validate() error {
	if !transcriptIDUUID.MatchString(string(attempt.CollectiveID)) || !transcriptIDUUID.MatchString(string(attempt.TranscriptID)) {
		return attemptValidationError("collectiveId and transcriptId must be canonical lowercase UUIDs")
	}
	if !attemptUTCTime(attempt.FirstSeenAt) {
		return attemptValidationError("firstSeenAt must be a nonzero UTC timestamp")
	}
	return attempt.Attempt.Validate()
}

// ReconcileVillageContributionAttempt reconciles one identity against the latest
// retained value. Callers must select the authenticated storage scope first and
// call this under their write transaction, not against an earlier network-read
// snapshot. A new ordinal is a separate identity and must be passed with nil
// saved evidence; gaps do not imply missing attempts.
//
// On conflict the saved evidence is returned unchanged along with an error. The
// result never aliases either input's decision timestamp. FirstSeenAt belongs
// to local evidence and is preserved on every existing-identity reconciliation.
func ReconcileVillageContributionAttempt(saved *LocalRemoteContributionAttempt, incoming LocalRemoteContributionAttempt) (LocalRemoteContributionAttempt, VillageContributionAttemptReconciliation, error) {
	retained := LocalRemoteContributionAttempt{}
	if saved != nil {
		retained = cloneContributionAttempt(*saved)
	}
	if err := incoming.Validate(); err != nil {
		return retained, "", err
	}
	if saved == nil {
		return cloneContributionAttempt(incoming), VillageContributionAttemptInserted, nil
	}
	if err := saved.Validate(); err != nil {
		return retained, "", err
	}
	old, next := saved.Attempt, incoming.Attempt
	if saved.CollectiveID != incoming.CollectiveID || saved.TranscriptID != incoming.TranscriptID || old.EventNum != next.EventNum {
		return retained, "", attemptValidationError("the incoming attempt has a different collective, transcript, or ordinal identity")
	}
	if !old.RecordedAt.Equal(next.RecordedAt) {
		return retained, "", attemptValidationError("recorded_at changed for an existing attempt identity")
	}
	if old.Status == next.Status && old.DecidedByActor == next.DecidedByActor && equalAttemptDecisionTime(old.DecidedAt, next.DecidedAt) {
		return retained, VillageContributionAttemptUnchanged, nil
	}
	if old.Status == VillageShareStatusPending {
		if next.DecidedAt == nil {
			return retained, "", attemptValidationError("an in-place pending decision requires a newly recorded decided_at")
		}
		result := cloneContributionAttempt(incoming)
		result.FirstSeenAt = saved.FirstSeenAt
		return result, VillageContributionAttemptUpdated, nil
	}
	if next.Status == VillageShareStatusPending {
		return retained, VillageContributionAttemptSuperseded, nil
	}
	return retained, "", attemptValidationError("terminal status or decision evidence changed for an existing attempt identity")
}

func cloneContributionAttempt(attempt LocalRemoteContributionAttempt) LocalRemoteContributionAttempt {
	if attempt.Attempt.DecidedAt != nil {
		value := *attempt.Attempt.DecidedAt
		attempt.Attempt.DecidedAt = &value
	}
	return attempt
}

func equalAttemptDecisionTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

func attemptUTCTime(value time.Time) bool {
	_, offset := value.Zone()
	return !value.IsZero() && offset == 0
}

func attemptValidationError(reason string) error {
	return fmt.Errorf("collective contribution attempt validation failed at schema.ReconcileVillageContributionAttempt during history validation: %s; retain prior evidence without installing this observation and fetch validated owner history before retrying", reason)
}
