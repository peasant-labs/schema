package schema

import (
	"fmt"
	"strings"
	"time"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- Publication state ---

// LocalPublicationState is whether a local session has a publication on Village.
type LocalPublicationState string

const (
	// LocalPublicationUnpublished: no publication of this session is recorded.
	LocalPublicationUnpublished LocalPublicationState = "unpublished"
	// LocalPublicationPublished: Village holds a transcript of this session.
	LocalPublicationPublished LocalPublicationState = "published"
)

// AllLocalPublicationStates is the canonical publication state menu.
var AllLocalPublicationStates = []LocalPublicationState{LocalPublicationUnpublished, LocalPublicationPublished}

func (s LocalPublicationState) IsValid() bool  { return inSet(s, AllLocalPublicationStates) }
func (s LocalPublicationState) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (LocalPublicationState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Publication State", "Whether a local session has a publication on Village", AllLocalPublicationStates), nil
}

// LocalPublicationAttemptFailure is the most recent failed publish attempt recorded
// for a session. A client compares AttemptedAt with the publication's
// PublishedAt to tell whether the failure is newer than the last success.
type LocalPublicationAttemptFailure struct {
	AttemptedAt time.Time `json:"attemptedAt"`
	Message     string    `json:"message"`
}

// LocalPublicationAudienceMember is one collective a published transcript is shared
// with. Status is the collective's current share status: approved means its
// members can read the transcript, pending means the collective's owner has
// not approved it yet.
type LocalPublicationAudienceMember struct {
	CollectiveID VillageUUID        `json:"collectiveId"`
	Name         string             `json:"name"`
	Status       VillageShareStatus `json:"status"`
}

// LocalPublication is the durable publication state of one local session for
// the Village account this computer is signed in to: the stored credential's
// Village and user. A publication recorded for another account is not
// reported, and a signed-out computer reports every session unpublished.
type LocalPublication struct {
	SessionID string                `json:"sessionId"`
	State     LocalPublicationState `json:"state"`
	// TranscriptID, TranscriptURL, and PublishedAt are present exactly when
	// State is published. PublishedAt is when Village last accepted this
	// session's content, by a first publish or an update; a client counts the
	// turns recorded after it to say about how many turns are new. The count
	// is approximate: it compares this computer's turn times with Village's
	// accept time.
	TranscriptID  *TranscriptID `json:"transcriptId,omitempty" nullable:"false"`
	TranscriptURL string        `json:"transcriptUrl,omitempty"`
	PublishedAt   *time.Time    `json:"publishedAt,omitempty" nullable:"false"`
	// LastAttempt is the most recent failed attempt, absent when none is
	// recorded.
	LastAttempt *LocalPublicationAttemptFailure `json:"lastAttempt,omitempty" nullable:"false"`
	// AutoPublish reports that a Peasant-managed hook is installed in this
	// session's repository, for a rule or from the terminal, so a commit or
	// push publishes the session without a click. A hook outlives the rule it
	// was installed for. It is false when the saved selection leaves the
	// session out, because the hook's push applies the selection.
	AutoPublish bool `json:"autoPublish" description:"A Peasant-managed git hook is installed in this session's repository, for an auto-publish rule or from the terminal, so a commit or push publishes the session without a click. False when the saved selection leaves the session out."`
	// OutsideSelection reports that the saved selection leaves this session
	// out of the local lists. The read ignores the selection, so such a
	// session is still returned.
	OutsideSelection bool `json:"outsideSelection"`
	// Audience lists the collectives the transcript is shared with. It is
	// present on a published row exactly when the request asked for
	// include=audience, as [] when the transcript is shared with no
	// collective, and never on an unpublished row.
	Audience *[]LocalPublicationAudienceMember `json:"audience,omitempty" nullable:"false"`
}

// Validate checks the state pairing and the audience of one publication row.
func (p LocalPublication) Validate() error {
	if strings.TrimSpace(p.SessionID) == "" {
		return fmt.Errorf("publication validation failed at schema.LocalPublication.Validate: sessionId is empty; the caller cannot match the row to a session; emit the requested session ID")
	}
	if !p.State.IsValid() {
		return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: state %q is outside the closed set; emit a member of schema.AllLocalPublicationStates", p.SessionID, p.State)
	}
	published := p.State == LocalPublicationPublished
	if published != (p.TranscriptID != nil) || published != (p.TranscriptURL != "") || published != (p.PublishedAt != nil) {
		return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: state %q disagrees with transcriptId, transcriptUrl, or publishedAt; a published row names its transcript and an unpublished row names none", p.SessionID, p.State)
	}
	if p.PublishedAt != nil && p.PublishedAt.IsZero() {
		return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: publishedAt is the zero time; a client cannot count new turns from it; emit the time Village accepted the content", p.SessionID)
	}
	if p.LastAttempt != nil && (p.LastAttempt.AttemptedAt.IsZero() || strings.TrimSpace(p.LastAttempt.Message) == "") {
		return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: lastAttempt lacks its time or message; the caller cannot say what failed or when; emit both", p.SessionID)
	}
	if !published && p.Audience != nil {
		return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: an unpublished row carries an audience; nothing is shared before publication; omit audience", p.SessionID)
	}
	if p.Audience == nil {
		return nil
	}
	seen := make(map[VillageUUID]struct{}, len(*p.Audience))
	for _, member := range *p.Audience {
		if member.Status != VillageShareStatusApproved && member.Status != VillageShareStatusPending {
			return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: audience member %q has status %q; the audience lists only approved and pending shares", p.SessionID, member.CollectiveID, member.Status)
		}
		if _, duplicate := seen[member.CollectiveID]; duplicate {
			return fmt.Errorf("publication validation failed for %q at schema.LocalPublication.Validate: collective %q is listed twice in the audience; list each collective once", p.SessionID, member.CollectiveID)
		}
		seen[member.CollectiveID] = struct{}{}
	}
	return nil
}

// LocalPublicationsResponse is the response of GET /api/v1/publications. It
// holds one row per requested session that exists on this computer; an
// identifier that names no session is omitted.
type LocalPublicationsResponse struct {
	Publications []LocalPublication `json:"publications" nullable:"false"`
}

// Validate checks every row and that each session appears once.
func (r LocalPublicationsResponse) Validate() error {
	if r.Publications == nil {
		return fmt.Errorf("publications validation failed at schema.LocalPublicationsResponse.Validate: publications is null; emit [] when no requested session exists")
	}
	seen := make(map[string]struct{}, len(r.Publications))
	for _, publication := range r.Publications {
		if err := publication.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[publication.SessionID]; duplicate {
			return fmt.Errorf("publications validation failed at schema.LocalPublicationsResponse.Validate: session %q has two rows; emit one row per session", publication.SessionID)
		}
		seen[publication.SessionID] = struct{}{}
	}
	return nil
}

// --- Collectives for publishing ---

// LocalCollectiveSuggestionReason is why the server suggests a collective for a
// session.
type LocalCollectiveSuggestionReason string

const (
	// LocalCollectiveSuggestionLinkedRepository: the collective links the session's
	// repository.
	LocalCollectiveSuggestionLinkedRepository LocalCollectiveSuggestionReason = "linked_repository"
	// LocalCollectiveSuggestionLinkedGithubOrg: the collective links the GitHub
	// organization that owns the session's repository.
	LocalCollectiveSuggestionLinkedGithubOrg LocalCollectiveSuggestionReason = "linked_github_org"
)

// AllLocalCollectiveSuggestionReasons is the canonical suggestion reason menu.
var AllLocalCollectiveSuggestionReasons = []LocalCollectiveSuggestionReason{LocalCollectiveSuggestionLinkedRepository, LocalCollectiveSuggestionLinkedGithubOrg}

func (r LocalCollectiveSuggestionReason) IsValid() bool {
	return inSet(r, AllLocalCollectiveSuggestionReasons)
}
func (r LocalCollectiveSuggestionReason) String() string { return string(r) }

// JSONSchema implements jsonschema.Exposer.
func (LocalCollectiveSuggestionReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Collective Suggestion Reason", "Why the local server suggests a collective for a session", AllLocalCollectiveSuggestionReasons), nil
}

// LocalCollectiveSuggestion is the server's reason to suggest one collective.
// Match is what matched: the linked repository's schema.RemoteLabel for a
// linked_repository suggestion, or the organization login for a
// linked_github_org suggestion.
type LocalCollectiveSuggestion struct {
	Reason LocalCollectiveSuggestionReason `json:"reason"`
	Match  string                          `json:"match"`
}

// LocalVillageCollective is one Village collective the signed-in user belongs
// to, as Village returns it, with the suggestion the local server computed.
type LocalVillageCollective struct {
	Group VillageUserGroup `json:"group"`
	// Suggestion is present when the server suggests this collective for the
	// requested session.
	Suggestion *LocalCollectiveSuggestion `json:"suggestion,omitempty" nullable:"false"`
}

// LocalVillageCollectivesResponse is the response of GET
// /api/v1/village/collectives.
type LocalVillageCollectivesResponse struct {
	Collectives []LocalVillageCollective `json:"collectives" nullable:"false"`
}

// Validate checks every suggestion and that each collective appears once.
func (r LocalVillageCollectivesResponse) Validate() error {
	if r.Collectives == nil {
		return fmt.Errorf("collectives validation failed at schema.LocalVillageCollectivesResponse.Validate: collectives is null; emit [] when the user belongs to no collective")
	}
	seen := make(map[VillageUUID]struct{}, len(r.Collectives))
	for _, collective := range r.Collectives {
		if _, duplicate := seen[collective.Group.ID]; duplicate {
			return fmt.Errorf("collectives validation failed at schema.LocalVillageCollectivesResponse.Validate: collective %q is listed twice; list each collective once", collective.Group.ID)
		}
		seen[collective.Group.ID] = struct{}{}
		if s := collective.Suggestion; s != nil && (!s.Reason.IsValid() || strings.TrimSpace(s.Match) == "") {
			return fmt.Errorf("collectives validation failed at schema.LocalVillageCollectivesResponse.Validate: collective %q has a suggestion with reason %q and match %q; a suggestion names a known reason and what matched", collective.Group.ID, s.Reason, s.Match)
		}
	}
	return nil
}
