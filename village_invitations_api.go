package schema

import (
	"fmt"
	"strings"
	"time"
)

// Collective invitations. An owner adding a person to a collective creates an
// invitation; the person becomes a member only by accepting it. A membership
// the person consented to (their own join or an accepted invitation) is what
// lets the collective's owners see their profile, so an owner can no longer
// widen who sees a person without that person's consent.
//
// Inviting never confirms that an account exists. The invite operation
// answers the same way for an unknown name, a person who is not discoverable,
// and a discoverable person, and an owner sees an invitation only as the
// handle they typed. A declined invitation looks exactly like a pending one to
// the owner until it expires, so declining reveals nothing.

// VillageCollectiveInvitationRequest invites a person to a collective by their
// Village username. The username is what the owner typed; the server answers
// the same way whether or not an account holds it.
type VillageCollectiveInvitationRequest struct {
	Username string `json:"username" minLength:"1"`
}

// VillageCollectiveInvitationSent is one open invitation as the collective's
// owners see it. InvitedUsername is the handle the inviting owner typed,
// lowercased; it is never resolved to a person, so the row says nothing about
// whether an account holds that handle. There is deliberately no status: an
// invitation the person declined is listed exactly like a pending one until it
// expires. InvitedBy is the owner who sent it as this viewer may see them.
type VillageCollectiveInvitationSent struct {
	ID              VillageUUID        `json:"id"`
	InvitedUsername string             `json:"invited_username" minLength:"1"`
	InvitedBy       VillagePublicOwner `json:"invited_by"`
	CreatedAt       time.Time          `json:"created_at"`
	ExpiresAt       time.Time          `json:"expires_at"`
}

// Validate refuses a row a producer must not emit.
func (s VillageCollectiveInvitationSent) Validate() error {
	const at = "schema.VillageCollectiveInvitationSent.Validate"
	if err := validateInvitationIdentity(at, s.ID, s.CreatedAt, s.ExpiresAt); err != nil {
		return err
	}
	if strings.TrimSpace(s.InvitedUsername) == "" {
		return fmt.Errorf("collective invitation validation failed at %s: invited_username is empty; send the handle the owner typed", at)
	}
	if err := s.InvitedBy.Validate(); err != nil {
		return fmt.Errorf("collective invitation validation failed at %s: invited_by: %w", at, err)
	}
	return nil
}

// VillageCollectiveInvitationsSentResponse lists a collective's open
// invitations for its owners: pending and declined ones that have not expired,
// in one indistinguishable list.
type VillageCollectiveInvitationsSentResponse struct {
	Invitations []VillageCollectiveInvitationSent `json:"invitations" nullable:"false"`
}

// Validate refuses a list a producer must not emit.
func (r VillageCollectiveInvitationsSentResponse) Validate() error {
	if r.Invitations == nil {
		return fmt.Errorf("collective invitation validation failed at schema.VillageCollectiveInvitationsSentResponse.Validate: invitations is null; emit [] when the collective has no open invitation")
	}
	seen := make(map[VillageUUID]struct{}, len(r.Invitations))
	for _, invitation := range r.Invitations {
		if err := invitation.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[invitation.ID]; duplicate {
			return fmt.Errorf("collective invitation validation failed at schema.VillageCollectiveInvitationsSentResponse.Validate: invitation %s is listed twice; list each invitation once", invitation.ID)
		}
		seen[invitation.ID] = struct{}{}
	}
	return nil
}

// VillageInvitedCollective is the collective an invitation names, as its
// invitee sees it. An invitation discloses this summary even when the
// collective is visible only to its members: deciding needs it.
type VillageInvitedCollective struct {
	ID             VillageUUID                     `json:"id"`
	Name           string                          `json:"name" minLength:"1"`
	Description    *string                         `json:"description"`
	AcceptanceMode VillageCollectiveAcceptanceMode `json:"acceptance_mode"`
}

// VillageCollectiveInvitation is one invitation as the invited person sees
// it. InvitedBy is the inviting owner as this viewer may see them: an owner
// who is not discoverable is anonymous here, and a client says "an owner of
// <collective>" instead of a handle.
type VillageCollectiveInvitation struct {
	ID         VillageUUID              `json:"id"`
	Collective VillageInvitedCollective `json:"collective"`
	InvitedBy  VillagePublicOwner       `json:"invited_by"`
	CreatedAt  time.Time                `json:"created_at"`
	ExpiresAt  time.Time                `json:"expires_at"`
}

// Validate refuses an invitation a producer must not emit.
func (i VillageCollectiveInvitation) Validate() error {
	const at = "schema.VillageCollectiveInvitation.Validate"
	if err := validateInvitationIdentity(at, i.ID, i.CreatedAt, i.ExpiresAt); err != nil {
		return err
	}
	if i.Collective.ID == "" {
		return fmt.Errorf("collective invitation validation failed at %s: collective.id is empty; name the collective the invitation is for", at)
	}
	if strings.TrimSpace(i.Collective.Name) == "" {
		return fmt.Errorf("collective invitation validation failed at %s: collective.name is empty; the invitee decides by the collective's name", at)
	}
	if !knownAcceptanceMode(i.Collective.AcceptanceMode) {
		return fmt.Errorf("collective invitation validation failed at %s: collective.acceptance_mode %q is not in the closed menu", at, i.Collective.AcceptanceMode)
	}
	if err := i.InvitedBy.Validate(); err != nil {
		return fmt.Errorf("collective invitation validation failed at %s: invited_by: %w", at, err)
	}
	return nil
}

// VillageMyInvitationsResponse lists the caller's own pending invitations that
// have not expired. It never lists anyone else's.
type VillageMyInvitationsResponse struct {
	Invitations []VillageCollectiveInvitation `json:"invitations" nullable:"false"`
}

// Validate refuses a list a producer must not emit.
func (r VillageMyInvitationsResponse) Validate() error {
	if r.Invitations == nil {
		return fmt.Errorf("collective invitation validation failed at schema.VillageMyInvitationsResponse.Validate: invitations is null; emit [] when the caller has no pending invitation")
	}
	seen := make(map[VillageUUID]struct{}, len(r.Invitations))
	for _, invitation := range r.Invitations {
		if err := invitation.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[invitation.ID]; duplicate {
			return fmt.Errorf("collective invitation validation failed at schema.VillageMyInvitationsResponse.Validate: invitation %s is listed twice; list each invitation once", invitation.ID)
		}
		seen[invitation.ID] = struct{}{}
	}
	return nil
}

func knownAcceptanceMode(mode VillageCollectiveAcceptanceMode) bool {
	for _, known := range AllVillageGroupAcceptanceModes {
		if mode == known {
			return true
		}
	}
	return false
}

func validateInvitationIdentity(at string, id VillageUUID, createdAt, expiresAt time.Time) error {
	if id == "" {
		return fmt.Errorf("collective invitation validation failed at %s: id is empty; every invitation has an id", at)
	}
	if createdAt.IsZero() || expiresAt.IsZero() {
		return fmt.Errorf("collective invitation validation failed at %s: created_at and expires_at are required", at)
	}
	if !expiresAt.After(createdAt) {
		return fmt.Errorf("collective invitation validation failed at %s: expires_at %s is not after created_at %s", at, expiresAt.Format(time.RFC3339), createdAt.Format(time.RFC3339))
	}
	return nil
}

// Notification email. A person may give Village one address to be emailed
// when a collective invites them. The address is personal contact data: it is
// returned only in the person's own settings (VillageUserSettings) and never
// in any response another viewer can receive, including the invitation an
// owner sees.

// VillageNotificationEmail is the caller's own notification address. Verified
// is false until the caller confirms the address with the token Village
// emailed to it; Village sends invitation emails only to a verified address.
type VillageNotificationEmail struct {
	Address  string `json:"address" minLength:"3" maxLength:"254"`
	Verified bool   `json:"verified"`
}

// VillageSetNotificationEmailRequest sets the caller's notification address,
// replacing any earlier one, and makes Village email a verification token to
// it. The new address is unverified until that token is confirmed.
type VillageSetNotificationEmailRequest struct {
	Address string `json:"address" minLength:"3" maxLength:"254"`
}

// VillageVerifyNotificationEmailRequest confirms the caller's notification
// address with the token Village emailed to it.
type VillageVerifyNotificationEmailRequest struct {
	Token string `json:"token" minLength:"1"`
}

// validateNotificationEmailAddress is the shape rule every address shares: one
// @ with text on both sides and no whitespace. Deliverability is proved by the
// verification token, not by this check.
func validateNotificationEmailAddress(at, address string) error {
	if len(address) < 3 || len(address) > 254 {
		return fmt.Errorf("notification email validation failed at %s: address must be 3 to 254 characters", at)
	}
	if strings.ContainsAny(address, " \t\r\n") {
		return fmt.Errorf("notification email validation failed at %s: address contains whitespace", at)
	}
	local, domain, found := strings.Cut(address, "@")
	if !found || local == "" || domain == "" || strings.Contains(domain, "@") {
		return fmt.Errorf("notification email validation failed at %s: address must have exactly one @ with text on both sides", at)
	}
	return nil
}

// Validate refuses an address a producer must not emit.
func (e VillageNotificationEmail) Validate() error {
	return validateNotificationEmailAddress("schema.VillageNotificationEmail.Validate", e.Address)
}

// Validate refuses an address a client must not send.
func (r VillageSetNotificationEmailRequest) Validate() error {
	return validateNotificationEmailAddress("schema.VillageSetNotificationEmailRequest.Validate", r.Address)
}
