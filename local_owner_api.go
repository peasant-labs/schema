package schema

import (
	"fmt"
	"strings"
	"time"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// LocalOwnerRecentClaimLimit is the most recent claim entries a status read
// returns. The log itself is append-only and keeps every entry; the wire bounds
// one response so a session-end harvest that claims for years cannot grow it
// without limit. A client that needs an older page asks again before
// LocalOwnerStatus.NextBefore.
const LocalOwnerRecentClaimLimit = 100

// LocalOwner is the one permanent owner of a peasant data directory. It is not
// a Village account, a username, or an email address: it is created once for
// the directory and survives every sign-out and sign-in. A store holds exactly
// one local owner for its whole life.
type LocalOwner struct {
	// ID is the stable local owner identifier generated once per data directory.
	ID string `json:"id" minLength:"1" maxLength:"64"`
	// CreatedAt is when this directory's local owner was first created.
	CreatedAt time.Time `json:"createdAt"`
}

// Validate checks the local owner identity and creation time.
func (o LocalOwner) Validate() error {
	if strings.TrimSpace(o.ID) == "" || len(o.ID) > 64 {
		return fmt.Errorf("local owner validation failed at schema.LocalOwner.Validate: id %q is empty or over 64 bytes; emit the stable identifier generated once for this data directory", o.ID)
	}
	if o.CreatedAt.IsZero() {
		return fmt.Errorf("local owner validation failed at schema.LocalOwner.Validate: createdAt is the zero time; emit the time this data directory's local owner was created")
	}
	return nil
}

// LocalClaimSource says why one claim entry exists.
type LocalClaimSource string

const (
	// LocalClaimSourceManual: the person ran an explicit claim command.
	LocalClaimSourceManual LocalClaimSource = "manual"
	// LocalClaimSourceSignInPrompt: the person answered the sign-in claim
	// prompt.
	LocalClaimSourceSignInPrompt LocalClaimSource = "sign_in_prompt"
	// LocalClaimSourceAutomatic: a harvest claimed the sessions it stored for
	// the first time while signed in. An automatic claim reuses the account's
	// latest claim entry and appends only when the account has none, so a
	// session-end hook cannot append one entry per run. Updated, reindexed, and
	// rebuilt rows are never claimed automatically.
	LocalClaimSourceAutomatic LocalClaimSource = "automatic"
)

// AllLocalClaimSources is the closed set of claim-entry sources.
var AllLocalClaimSources = []LocalClaimSource{
	LocalClaimSourceManual,
	LocalClaimSourceSignInPrompt,
	LocalClaimSourceAutomatic,
}

// IsValid reports whether s is a member of the claim-source set.
func (s LocalClaimSource) IsValid() bool { return inSet(s, AllLocalClaimSources) }

// String returns the wire representation of the claim source.
func (s LocalClaimSource) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (LocalClaimSource) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Local Claim Source",
		"Why one claim entry exists: an explicit command, the sign-in prompt, or an automatic claim of newly stored sessions",
		AllLocalClaimSources,
	), nil
}

// LocalClaimEntry is one entry in the local claim log. The account identity
// reuses CollectiveObservationScope's field names and grammar: a normalized
// full Village base URL plus a stable Village user UUID. Username is a display
// snapshot only, never identity, so a rename or a reused username never adds a
// second account.
//
// Entries are ordered by ClaimedAt, then by insertion order (the store's
// monotonically increasing rowid), never by the random entry id. A status read
// returns only the most recent LocalOwnerRecentClaimLimit entries.
type LocalClaimEntry struct {
	// ID is the stable claim entry identifier.
	ID string `json:"id" minLength:"1" maxLength:"64"`
	// LocalOwnerID is the local owner the claim belongs to.
	LocalOwnerID string `json:"localOwnerId" minLength:"1" maxLength:"64"`
	// VillageURL is the normalized full Village base URL, using the same
	// grammar as CollectiveObservationScope.villageUrl.
	VillageURL string `json:"villageUrl" maxLength:"1024"`
	// UserID is the stable Village user UUID, never the username. It is the
	// same value CollectiveObservationScope.userId carries.
	UserID VillageUUID `json:"userId"`
	// Username is the display snapshot recorded at claim time. It is never
	// identity and may differ between entries for one Village user UUID. It
	// must name a person, so an empty value is refused rather than shown.
	Username string `json:"username" minLength:"1"`
	// Source says why this entry exists.
	Source LocalClaimSource `json:"source"`
	// ClaimedAt is when this entry was recorded. It is one part of the claim
	// order key; the other is insertion order.
	ClaimedAt time.Time `json:"claimedAt"`
}

// Validate checks the entry identifier, the owner link, the account scope, the
// display snapshot, the source, and the claim time.
func (e LocalClaimEntry) Validate() error {
	if strings.TrimSpace(e.ID) == "" || len(e.ID) > 64 {
		return fmt.Errorf("local claim entry validation failed at schema.LocalClaimEntry.Validate: id %q is empty or over 64 bytes; emit the stable claim entry identifier", e.ID)
	}
	if strings.TrimSpace(e.LocalOwnerID) == "" || len(e.LocalOwnerID) > 64 {
		return fmt.Errorf("local claim entry validation failed for %q at schema.LocalClaimEntry.Validate: localOwnerId %q is empty or over 64 bytes; emit the local owner this claim belongs to", e.ID, e.LocalOwnerID)
	}
	scope, err := NormalizeCollectiveObservationScope(e.VillageURL, e.UserID)
	if err != nil {
		return fmt.Errorf("local claim entry validation failed for %q at schema.LocalClaimEntry.Validate: villageUrl/userId do not form the established account-scope identity: %w", e.ID, err)
	}
	if scope.VillageURL != e.VillageURL || scope.UserID != e.UserID {
		return fmt.Errorf("local claim entry validation failed for %q at schema.LocalClaimEntry.Validate: villageUrl %q and userId %q are not normalized; emit the canonical %q/%q so one account cannot be recorded under two spellings", e.ID, e.VillageURL, e.UserID, scope.VillageURL, scope.UserID)
	}
	if strings.TrimSpace(e.Username) == "" {
		return fmt.Errorf("local claim entry validation failed for %q at schema.LocalClaimEntry.Validate: username is empty; emit the display snapshot recorded at claim time, because an empty name would be shown as a blank owner", e.ID)
	}
	if !e.Source.IsValid() {
		return fmt.Errorf("local claim entry validation failed for %q at schema.LocalClaimEntry.Validate: source %q is outside the closed set; emit a member of schema.AllLocalClaimSources", e.ID, e.Source)
	}
	if e.ClaimedAt.IsZero() {
		return fmt.Errorf("local claim entry validation failed for %q at schema.LocalClaimEntry.Validate: claimedAt is the zero time; emit when this entry was recorded", e.ID)
	}
	return nil
}

// LocalOwnerClaimRequest is the claim control's body. all=false claims only
// unclaimed sessions and annotations; all=true claims every local session and
// annotation. An explicit claim appends a new entry and never rewrites history.
type LocalOwnerClaimRequest struct {
	All bool `json:"all"`
}

// LocalOwnerPromptState is the server-computed claim prompt state for the
// signed-in account.
type LocalOwnerPromptState string

const (
	// LocalOwnerPromptStateNotApplicable: no Village account is selected, so
	// there is no account scope to prompt for.
	LocalOwnerPromptStateNotApplicable LocalOwnerPromptState = "not_applicable"
	// LocalOwnerPromptStateDue: an account is selected and no decision exists
	// for this local owner plus account scope.
	LocalOwnerPromptStateDue LocalOwnerPromptState = "due"
	// LocalOwnerPromptStateDeclined: the person chose "not now" for this local
	// owner plus account scope. A decline writes no claim entry.
	LocalOwnerPromptStateDeclined LocalOwnerPromptState = "declined"
	// LocalOwnerPromptStateCompleted: a claim decision was completed for this
	// local owner plus account scope, including a claim action that found
	// nothing left to claim.
	LocalOwnerPromptStateCompleted LocalOwnerPromptState = "completed"
)

// AllLocalOwnerPromptStates is the closed set of claim prompt states.
var AllLocalOwnerPromptStates = []LocalOwnerPromptState{
	LocalOwnerPromptStateNotApplicable,
	LocalOwnerPromptStateDue,
	LocalOwnerPromptStateDeclined,
	LocalOwnerPromptStateCompleted,
}

// IsValid reports whether s is a member of the prompt-state set.
func (s LocalOwnerPromptState) IsValid() bool { return inSet(s, AllLocalOwnerPromptStates) }

// String returns the wire representation of the prompt state.
func (s LocalOwnerPromptState) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (LocalOwnerPromptState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Local Owner Prompt State",
		"The server-computed claim prompt state: no account selected, due, declined, or completed for one local owner plus account scope",
		AllLocalOwnerPromptStates,
	), nil
}

// LocalOwnerClaimPrompt is the claim prompt's per-account state. The server
// computes it from the selected account and the decision recorded for this
// local owner plus account scope, so two surfaces never ask differently.
//
// Account is present exactly when a Village account is selected, and absent for
// not_applicable. DecidedAt is present exactly when the state is settled
// (declined or completed) and absent for not_applicable and due.
type LocalOwnerClaimPrompt struct {
	// State is the computed prompt state.
	State LocalOwnerPromptState `json:"state"`
	// Account is the selected account scope: a normalized full Village base URL
	// plus a stable Village user UUID. It is present exactly when signed in.
	Account *CollectiveObservationScope `json:"account,omitempty" nullable:"false"`
	// DecidedAt is when the prompt was declined or a claim decision completed.
	// It is absent while the prompt is not applicable or still due.
	DecidedAt *time.Time `json:"decidedAt,omitempty" nullable:"false"`
}

// Validate checks the closed state, the account scope, and the decided time
// against the state.
func (p LocalOwnerClaimPrompt) Validate() error {
	const where = "local owner claim prompt validation failed at schema.LocalOwnerClaimPrompt.Validate: "
	if !p.State.IsValid() {
		return fmt.Errorf("%sstate %q is outside the closed set; emit a member of schema.AllLocalOwnerPromptStates", where, p.State)
	}
	switch p.State {
	case LocalOwnerPromptStateNotApplicable:
		if p.Account != nil {
			return fmt.Errorf("%sstate not_applicable carries an account; no account is selected, so clear account rather than name one the prompt does not apply to", where)
		}
	case LocalOwnerPromptStateDue:
		if p.Account == nil {
			return fmt.Errorf("%sstate due has no account; emit the selected account scope so the client knows which account the prompt is for", where)
		}
		if p.DecidedAt != nil {
			return fmt.Errorf("%sstate due carries decidedAt; a due prompt has no decision yet, so omit decidedAt", where)
		}
	case LocalOwnerPromptStateDeclined, LocalOwnerPromptStateCompleted:
		if p.Account == nil {
			return fmt.Errorf("%sstate %s has no account; emit the account scope the decision was recorded for", where, p.State)
		}
		if p.DecidedAt == nil {
			return fmt.Errorf("%sstate %s has no decidedAt; emit when the decision was recorded", where, p.State)
		}
	}
	if p.Account != nil {
		scope, err := NormalizeCollectiveObservationScope(p.Account.VillageURL, p.Account.UserID)
		if err != nil {
			return fmt.Errorf("%saccount does not form the established account-scope identity: %w", where, err)
		}
		if scope.VillageURL != p.Account.VillageURL || scope.UserID != p.Account.UserID {
			return fmt.Errorf("%saccount %q/%q is not normalized; emit the canonical %q/%q so one account cannot be recorded under two spellings", where, p.Account.VillageURL, p.Account.UserID, scope.VillageURL, scope.UserID)
		}
	}
	if p.DecidedAt != nil && p.DecidedAt.IsZero() {
		return fmt.Errorf("%sdecidedAt is the zero time; emit when the decision was recorded", where)
	}
	return nil
}

// LocalOwnerOwnershipCounts is the claimed partition of one local artifact
// kind.
type LocalOwnerOwnershipCounts struct {
	// Total is every artifact of this kind the local store holds.
	Total int `json:"total" minimum:"0"`
	// Claimed is how many of them name the selected account.
	Claimed int `json:"claimed" minimum:"0"`
	// Unclaimed is the remainder. It is the count a prompt shows, so the server
	// owns it with the other two rather than each client subtracting.
	Unclaimed int `json:"unclaimed" minimum:"0"`
}

// Validate checks that the partition is complete and non-negative.
func (c LocalOwnerOwnershipCounts) Validate() error {
	if c.Total < 0 || c.Claimed < 0 || c.Unclaimed < 0 {
		return fmt.Errorf("local owner ownership count validation failed at schema.LocalOwnerOwnershipCounts.Validate: total=%d claimed=%d unclaimed=%d; every count is zero or greater", c.Total, c.Claimed, c.Unclaimed)
	}
	if c.Claimed+c.Unclaimed != c.Total {
		return fmt.Errorf("local owner ownership count validation failed at schema.LocalOwnerOwnershipCounts.Validate: claimed %d plus unclaimed %d does not equal total %d; the three counts must partition the artifacts", c.Claimed, c.Unclaimed, c.Total)
	}
	return nil
}

// LocalOwnerOwnership is the local artifact partition the claim control shows.
type LocalOwnerOwnership struct {
	Sessions    LocalOwnerOwnershipCounts `json:"sessions"`
	Annotations LocalOwnerOwnershipCounts `json:"annotations"`
}

// Validate checks both partitions.
func (o LocalOwnerOwnership) Validate() error {
	if err := o.Sessions.Validate(); err != nil {
		return err
	}
	return o.Annotations.Validate()
}

// LocalOwnerStatus is the local owner read: the one local owner, the most
// recent claim entries, the full claim total, the claim prompt state, and the
// artifact ownership counts. This is a local read and makes no network call.
type LocalOwnerStatus struct {
	Owner LocalOwner `json:"owner"`
	// Claims is the most recent LocalOwnerRecentClaimLimit claim entries,
	// ordered by claimedAt and then insertion order. It is never null; an
	// installation with no claims answers [].
	Claims []LocalClaimEntry `json:"claims" nullable:"false"`
	// TotalClaims is the number of entries in the whole append-only log, not
	// only the page Claims carries.
	TotalClaims int `json:"totalClaims" minimum:"0"`
	// NextBefore is the cursor for the next older page. It is the id of the
	// oldest entry in Claims and is present exactly when TotalClaims exceeds
	// the returned page.
	NextBefore *string `json:"nextBefore,omitempty" minLength:"1" maxLength:"64" nullable:"false"`
	// Prompt is the server-computed claim prompt state for the selected
	// account.
	Prompt LocalOwnerClaimPrompt `json:"prompt"`
	// Ownership is the claimed partition of local sessions and annotations.
	Ownership LocalOwnerOwnership `json:"ownership"`
}

// Validate checks the owner, the bounded non-null claim page, the total and
// cursor, the prompt, the ownership counts, and that every returned claim
// belongs to this owner.
func (s LocalOwnerStatus) Validate() error {
	if err := s.Owner.Validate(); err != nil {
		return err
	}
	if s.Claims == nil {
		return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: claims is null; emit [] when this data directory has no claims")
	}
	if len(s.Claims) > LocalOwnerRecentClaimLimit {
		return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: claims carries %d entries; one response returns at most %d, so page the rest behind nextBefore", len(s.Claims), LocalOwnerRecentClaimLimit)
	}
	if s.TotalClaims < len(s.Claims) {
		return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: totalClaims %d is less than the %d returned entries; totalClaims counts the whole log and cannot omit a returned entry", s.TotalClaims, len(s.Claims))
	}
	more := s.TotalClaims > len(s.Claims)
	if more && s.NextBefore == nil {
		return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: totalClaims %d exceeds the %d returned entries but nextBefore is absent; emit the cursor for the older page", s.TotalClaims, len(s.Claims))
	}
	if !more && s.NextBefore != nil {
		return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: nextBefore %q is set but totalClaims %d equals the %d returned entries; omit the cursor when no entries remain", *s.NextBefore, s.TotalClaims, len(s.Claims))
	}
	if s.NextBefore != nil {
		if strings.TrimSpace(*s.NextBefore) == "" || len(*s.NextBefore) > 64 {
			return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: nextBefore %q is empty or over 64 bytes; emit the id of the oldest returned entry", *s.NextBefore)
		}
		if len(s.Claims) == 0 || s.Claims[len(s.Claims)-1].ID != *s.NextBefore {
			return fmt.Errorf("local owner status validation failed at schema.LocalOwnerStatus.Validate: nextBefore %q does not name the oldest returned entry; emit that entry's id so the next page is contiguous", *s.NextBefore)
		}
	}
	for _, claim := range s.Claims {
		if claim.LocalOwnerID != s.Owner.ID {
			return fmt.Errorf("local owner status validation failed for claim %q at schema.LocalOwnerStatus.Validate: localOwnerId %q does not name this directory's owner %q; emit only claims appended for this local owner", claim.ID, claim.LocalOwnerID, s.Owner.ID)
		}
		if err := claim.Validate(); err != nil {
			return err
		}
	}
	if err := s.Prompt.Validate(); err != nil {
		return err
	}
	return s.Ownership.Validate()
}

// LocalOwnerClaimResult is the claim route's answer: the same status the read
// and decline routes return, plus what this claim did.
type LocalOwnerClaimResult struct {
	// Status is the recomputed local owner status after the claim.
	Status LocalOwnerStatus `json:"status"`
	// Claimed is how many artifacts this action claimed.
	Claimed int `json:"claimed" minimum:"0"`
	// AlreadyClaimed is how many requested artifacts already named the
	// selected account.
	AlreadyClaimed int `json:"alreadyClaimed" minimum:"0"`
	// ClaimID is the entry this action appended. Explicit claims append, so an
	// explicit claim that claimed nothing still records its decision. It is
	// absent only when the claim wrote no entry.
	ClaimID *string `json:"claimId,omitempty" minLength:"1" maxLength:"64" nullable:"false"`
}

// Validate checks the nested status, the non-negative counts, and the entry id.
func (r LocalOwnerClaimResult) Validate() error {
	if err := r.Status.Validate(); err != nil {
		return err
	}
	if r.Claimed < 0 || r.AlreadyClaimed < 0 {
		return fmt.Errorf("local owner claim result validation failed at schema.LocalOwnerClaimResult.Validate: claimed=%d alreadyClaimed=%d; every count is zero or greater", r.Claimed, r.AlreadyClaimed)
	}
	if r.ClaimID != nil {
		if strings.TrimSpace(*r.ClaimID) == "" || len(*r.ClaimID) > 64 {
			return fmt.Errorf("local owner claim result validation failed at schema.LocalOwnerClaimResult.Validate: claimId %q is empty or over 64 bytes; emit the appended entry id or omit the field", *r.ClaimID)
		}
	}
	return nil
}

// LocalArtifactIdentityKind says whose artifact a local read row describes.
type LocalArtifactIdentityKind string

const (
	// LocalArtifactIdentityKindLocal: this data directory's local owner.
	LocalArtifactIdentityKindLocal LocalArtifactIdentityKind = "local"
	// LocalArtifactIdentityKindVillage: the Village account that claimed the
	// local artifact. A claim outlives the sign-in that made it, so this is the
	// account that owns the artifact now, not necessarily the account selected
	// in this process.
	LocalArtifactIdentityKindVillage LocalArtifactIdentityKind = "village"
	// LocalArtifactIdentityKindForeign: a pulled artifact's owner, read from the
	// pull manifest. A foreign artifact is never claimed and never becomes
	// local. The identity carries the owner's recorded Village URL, user UUID,
	// and display snapshot; Village transcript reads keep using
	// VillagePublicOwner for the same person.
	LocalArtifactIdentityKindForeign LocalArtifactIdentityKind = "foreign"
	// LocalArtifactIdentityKindUnknown: the owner is projected but withheld, or
	// a pulled row's manifest owner is empty. It carries no identity and is
	// never a local fallback.
	LocalArtifactIdentityKindUnknown LocalArtifactIdentityKind = "unknown"
)

// AllLocalArtifactIdentityKinds is the canonical artifact identity menu.
var AllLocalArtifactIdentityKinds = []LocalArtifactIdentityKind{
	LocalArtifactIdentityKindLocal,
	LocalArtifactIdentityKindVillage,
	LocalArtifactIdentityKindForeign,
	LocalArtifactIdentityKindUnknown,
}

// IsValid reports whether k is a member of the artifact identity menu.
func (k LocalArtifactIdentityKind) IsValid() bool {
	return inSet(k, AllLocalArtifactIdentityKinds)
}

// String returns the wire representation of the artifact identity kind.
func (k LocalArtifactIdentityKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (LocalArtifactIdentityKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Local Artifact Identity Kind",
		"Whose artifact a local read describes: this data directory's local owner, the Village account that claimed it, a foreign pull-manifest owner, or an owner that is projected but withheld",
		AllLocalArtifactIdentityKinds,
	), nil
}

// LocalArtifactIdentity names the owner or author of a local read row without
// merging either role into the other or into the producer.
//
// An absent owner field means this server does not project owners at all.
// {"kind":"unknown"} means the server projects owners but withholds this one; a
// client must not read either as the local holder. The two are deliberately
// distinct wire states.
//
// LocalOwnerID is set exactly for kind local. VillageURL, VillageUserID, and
// ClaimID are set for kind village; ClaimID joins the row to the entry that
// explains it. VillageURL and VillageUserID are set for kind foreign. Username
// is a display snapshot that may be omitted when the Village withheld it, but
// an empty username is refused. Kind unknown carries no identity at all.
type LocalArtifactIdentity struct {
	Kind LocalArtifactIdentityKind `json:"kind"`
	// LocalOwnerID is this data directory's local owner identifier, 1 to 64
	// bytes. It is set exactly for kind local.
	LocalOwnerID *string `json:"localOwnerId,omitempty" minLength:"1" maxLength:"64" nullable:"false"`
	// VillageURL is the normalized full Village base URL, using the same
	// grammar as CollectiveObservationScope.villageUrl. It is set for kind
	// village or foreign.
	VillageURL *string `json:"villageUrl,omitempty" maxLength:"1024" nullable:"false"`
	// VillageUserID is the stable Village user UUID. It is set for kind village
	// or foreign.
	VillageUserID *VillageUUID `json:"villageUserId,omitempty" nullable:"false"`
	// ClaimID is the claim entry that links this artifact to the Village
	// account. It is set exactly for kind village.
	ClaimID *string `json:"claimId,omitempty" minLength:"1" maxLength:"64" nullable:"false"`
	// Username is the display snapshot. It is omitted when the Village
	// withheld it, and it must not be empty when present.
	Username *string `json:"username,omitempty" minLength:"1" nullable:"false"`
}

func identityScopeErr(where, kind string, err error) error {
	return fmt.Errorf("%skind %s villageUrl/villageUserId do not form the established account-scope identity: %w", where, kind, err)
}

// Validate checks the identity fields against the declared kind.
func (i *LocalArtifactIdentity) Validate() error {
	if i == nil {
		return nil
	}
	const where = "local artifact identity validation failed at schema.LocalArtifactIdentity.Validate: "
	if !i.Kind.IsValid() {
		return fmt.Errorf("%skind %q is outside the closed set; emit a member of schema.AllLocalArtifactIdentityKinds", where, i.Kind)
	}
	if i.Username != nil && strings.TrimSpace(*i.Username) == "" {
		return fmt.Errorf("%skind %s carries an empty username; omit the display snapshot when the Village withheld it, because an empty name would be shown as a blank person", where, i.Kind)
	}
	switch i.Kind {
	case LocalArtifactIdentityKindLocal:
		if i.LocalOwnerID == nil || strings.TrimSpace(*i.LocalOwnerID) == "" {
			return fmt.Errorf("%skind local has no localOwnerId; emit this data directory's local owner identifier", where)
		}
		if len(*i.LocalOwnerID) > 64 {
			return fmt.Errorf("%skind local carries a localOwnerId over 64 bytes; emit the stored identifier", where)
		}
		if i.VillageURL != nil {
			return fmt.Errorf("%skind local carries villageUrl %q; a local owner is not a Village account, so clear the Village identity or report kind village", where, *i.VillageURL)
		}
		if i.VillageUserID != nil {
			return fmt.Errorf("%skind local carries villageUserId %q; a local owner is not a Village account, so clear the Village identity or report kind village", where, *i.VillageUserID)
		}
		if i.ClaimID != nil {
			return fmt.Errorf("%skind local carries claimId %q; a local artifact is not linked to a claim entry, so clear it", where, *i.ClaimID)
		}
		if i.Username != nil {
			return fmt.Errorf("%skind local carries username %q; the local owner is not a Village account, so omit the display snapshot or report kind village", where, *i.Username)
		}
	case LocalArtifactIdentityKindVillage:
		if i.ClaimID == nil || strings.TrimSpace(*i.ClaimID) == "" {
			return fmt.Errorf("%skind village has no claimId; emit the claim entry that links this artifact to the account so a row joins the entry that explains it", where)
		}
		if len(*i.ClaimID) > 64 {
			return fmt.Errorf("%skind village carries a claimId over 64 bytes; emit the stored claim entry identifier", where)
		}
		if i.LocalOwnerID != nil {
			return fmt.Errorf("%skind village carries localOwnerId %q; a claimed artifact is owned by the Village account, so clear it", where, *i.LocalOwnerID)
		}
		if i.VillageURL == nil {
			return fmt.Errorf("%skind village has no villageUrl; emit the normalized base URL that scopes the claimed account", where)
		}
		if i.VillageUserID == nil {
			return fmt.Errorf("%skind village has no villageUserId; emit the stable Village user UUID", where)
		}
		scope, err := NormalizeCollectiveObservationScope(*i.VillageURL, *i.VillageUserID)
		if err != nil {
			return identityScopeErr(where, string(i.Kind), err)
		}
		if scope.VillageURL != *i.VillageURL || scope.UserID != *i.VillageUserID {
			return fmt.Errorf("%skind village carries villageUrl %q and villageUserId %q that are not normalized; emit the canonical %q/%q so one account cannot be recorded under two spellings", where, *i.VillageURL, *i.VillageUserID, scope.VillageURL, scope.UserID)
		}
	case LocalArtifactIdentityKindForeign:
		if i.ClaimID != nil {
			return fmt.Errorf("%skind foreign carries claimId %q; a pulled artifact is never claimed, so clear it", where, *i.ClaimID)
		}
		if i.LocalOwnerID != nil {
			return fmt.Errorf("%skind foreign carries localOwnerId %q; a foreign owner is not this directory's local owner, so clear it", where, *i.LocalOwnerID)
		}
		if i.VillageURL == nil {
			return fmt.Errorf("%skind foreign has no villageUrl; emit the normalized base URL the pulled owner was recorded under", where)
		}
		if i.VillageUserID == nil {
			return fmt.Errorf("%skind foreign has no villageUserId; emit the stable Village user UUID recorded by the pull manifest", where)
		}
		scope, err := NormalizeCollectiveObservationScope(*i.VillageURL, *i.VillageUserID)
		if err != nil {
			return identityScopeErr(where, string(i.Kind), err)
		}
		if scope.VillageURL != *i.VillageURL || scope.UserID != *i.VillageUserID {
			return fmt.Errorf("%skind foreign carries villageUrl %q and villageUserId %q that are not normalized; emit the canonical %q/%q", where, *i.VillageURL, *i.VillageUserID, scope.VillageURL, scope.UserID)
		}
	case LocalArtifactIdentityKindUnknown:
		if i.LocalOwnerID != nil || i.VillageURL != nil || i.VillageUserID != nil || i.ClaimID != nil || i.Username != nil {
			return fmt.Errorf("%skind unknown carries an identity; the owner is projected but withheld, so clear localOwnerId, villageUrl, villageUserId, claimId, and username rather than fall back to the local holder", where)
		}
	}
	return nil
}

// validateArtifactAdapterVersion applies the positive-when-present producer
// revision rule the local artifact reads share with UnifiedMetadata.
func validateArtifactAdapterVersion(version *int, at string) error {
	if version != nil && *version <= 0 {
		return fmt.Errorf("local artifact read validation failed at %s: adapterVersion %d is not positive; a producer revision is positive when present, so omit the field when unknown", at, *version)
	}
	return nil
}

// equalOptionalArtifactIdentity compares two optional identity projections by
// value, so separately decoded summaries can be checked for the same identity.
func equalOptionalArtifactIdentity(a, b *LocalArtifactIdentity) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Kind == b.Kind &&
		equalOptionalString(a.LocalOwnerID, b.LocalOwnerID) &&
		equalOptionalString(a.VillageURL, b.VillageURL) &&
		equalOptionalVillageUUID(a.VillageUserID, b.VillageUserID) &&
		equalOptionalString(a.ClaimID, b.ClaimID) &&
		equalOptionalString(a.Username, b.Username)
}

func equalOptionalVillageUUID(a, b *VillageUUID) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

// equalOptionalRevision compares two optional positive revisions by value.
func equalOptionalRevision(a, b *int) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

// LocalAnnotationRow is the local annotation read envelope. It embeds the
// shared AnnotationSummary unchanged and adds the local owner and author as
// optional, role-distinct projections. It never changes AnnotationSummary
// itself or the Village pull envelope's PullAnnotation.Author.
type LocalAnnotationRow struct {
	AnnotationSummary
	// Owner is whose artifact this annotation is in the local store. It is
	// absent when this server does not project owners; {"kind":"unknown"} means
	// the owner is projected but withheld.
	Owner *LocalArtifactIdentity `json:"owner,omitempty" nullable:"false"`
	// Author is who wrote the annotation, kept distinct from the owner and
	// from the producer. It uses the same absence rule as Owner.
	Author *LocalArtifactIdentity `json:"author,omitempty" nullable:"false"`
}

// Validate checks the shared annotation summary and both optional identity
// projections against their declared kinds.
func (r LocalAnnotationRow) Validate() error {
	if err := r.AnnotationSummary.Validate(); err != nil {
		return err
	}
	if err := r.Owner.Validate(); err != nil {
		return err
	}
	return r.Author.Validate()
}
