package schema

import (
	"fmt"
	"time"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// This file declares the local-account wire surface: the account, session, and
// admin operations a self-hosted Village serves without a software-forge
// provider. It defines shapes and statuses only; the handlers ship in the
// village repository. The server route names are the source of truth for the
// paths declared in openapi/village.go.

// VillageRegistrationMode is the closed set of account-creation policies a
// deployment can run under. closed (the default) refuses self-serve signup;
// invite accepts only single-use invitation links; email enables self-serve
// signup with an email verification step.
type VillageRegistrationMode string

const (
	VillageRegistrationModeClosed VillageRegistrationMode = "closed"
	VillageRegistrationModeInvite VillageRegistrationMode = "invite"
	VillageRegistrationModeEmail  VillageRegistrationMode = "email"
)

// AllVillageRegistrationModes is the canonical order of the registration-mode
// menu.
var AllVillageRegistrationModes = []VillageRegistrationMode{
	VillageRegistrationModeClosed,
	VillageRegistrationModeInvite,
	VillageRegistrationModeEmail,
}

func (VillageRegistrationMode) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Village Registration Mode",
		"Account-creation policy: closed refuses self-serve signup, invite accepts single-use invitation links, and email enables signup with verification",
		AllVillageRegistrationModes,
	), nil
}

func (m VillageRegistrationMode) IsValid() bool {
	for _, candidate := range AllVillageRegistrationModes {
		if candidate == m {
			return true
		}
	}
	return false
}

// VillageProvisioningStatus is the closed set of provisioning states a
// provisioned account moves through. An account is pending until it first
// authenticates, and activated from the first successful sign-in or password
// change onward.
type VillageProvisioningStatus string

const (
	VillageProvisioningStatusPending   VillageProvisioningStatus = "pending"
	VillageProvisioningStatusActivated VillageProvisioningStatus = "activated"
)

// AllVillageProvisioningStatuses is the canonical order of the provisioning
// status menu.
var AllVillageProvisioningStatuses = []VillageProvisioningStatus{
	VillageProvisioningStatusPending,
	VillageProvisioningStatusActivated,
}

func (VillageProvisioningStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Village Provisioning Status",
		"Provisioning state of a local account: pending before first sign-in, activated afterward",
		AllVillageProvisioningStatuses,
	), nil
}

func (s VillageProvisioningStatus) IsValid() bool {
	for _, candidate := range AllVillageProvisioningStatuses {
		if candidate == s {
			return true
		}
	}
	return false
}

// VillageAccountProfile is the self projection returned by GET /api/v1/auth/me.
// It is self-only: email and must_change_password describe the caller's own
// account, and the projection is not the broader VillageUser row served on
// public surfaces. username is the neutral handle; the deprecated
// github_username stays populated for forge-created accounts until callers
// migrate to the neutral handle.
type VillageAccountProfile struct {
	ID          VillageUUID `json:"id"`
	Username    string      `json:"username"`
	Email       *string     `json:"email"`
	DisplayName *string     `json:"display_name"`
	AvatarURL   *string     `json:"avatar_url"`
	IsAdmin     bool        `json:"is_admin"`
	Provider    string      `json:"provider"`
	// Deprecated: read username instead. github_username stays populated for
	// accounts created through a software-forge provider.
	GithubUsername     string     `json:"github_username" deprecated:"true"`
	MustChangePassword bool       `json:"must_change_password"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	PasswordChangedAt  *time.Time `json:"password_changed_at"`
}

// VillageAuthProvidersResponse is the sign-in-method discovery response
// returned by GET /api/v1/auth/providers. It tells the frontend which doors are
// open without attempting a login: the local username-and-password door, the
// deployment's registration mode, and the enabled forge providers. A providers
// array is always present (possibly empty) so a zero-provider instance is
// explicit rather than ambiguous.
type VillageAuthProvidersResponse struct {
	Local            bool                    `json:"local"`
	RegistrationMode VillageRegistrationMode `json:"registration_mode"`
	Providers        []string                `json:"providers" nullable:"false"`
}

// VillageAccountSession is one server-side session record in the account's session
// list. It carries no token or digest: the raw session value lives only in the
// HttpOnly cookie. current marks the session that served the listing request.
type VillageAccountSession struct {
	ID         VillageUUID `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	LastSeenAt time.Time   `json:"last_seen_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	Current    bool        `json:"current"`
	UserAgent  *string     `json:"user_agent"`
	IPAddress  *string     `json:"ip_address"`
}

// VillageAccountSessionListResponse is the response of GET /api/v1/auth/sessions.
// sessions is always a concrete array.
type VillageAccountSessionListResponse struct {
	Sessions []VillageAccountSession `json:"sessions" nullable:"false"`
}

// VillageBootstrapClaimRequest is the body of POST /api/v1/auth/bootstrap/claim,
// the one-time claim that creates the first administrator on a fresh instance.
// Unknown fields are rejected so a caller cannot silently ask for state the
// server will not set.
type VillageBootstrapClaimRequest struct {
	Token    string `json:"token" required:"true" description:"One-time install claim token minted at startup or pinned through the environment"`
	Username string `json:"username" required:"true" description:"Neutral handle for the first administrator"`
	Password string `json:"password" required:"true" description:"Initial administrator password; never logged or echoed"`
}

// VillageLocalLoginRequest is the body of POST /api/v1/auth/local/login. A wrong
// password and an unknown handle share one generic refusal, so the request and
// its refusals do not reveal whether an account exists.
type VillageLocalLoginRequest struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

// VillageAcceptInviteRequest is the body of
// POST /api/v1/auth/local/accept-invite. The consume is transactional: a handle
// collision rolls the whole consume back and leaves the invitation usable.
type VillageAcceptInviteRequest struct {
	Token    string `json:"token" required:"true" description:"Single-use invitation token"`
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

// VillageChangePasswordRequest is the body of POST /api/v1/auth/me/password.
// A successful change rotates the current session, revokes every other session,
// and clears any temporary-credential expiry atomically.
type VillageChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" required:"true"`
	NewPassword     string `json:"new_password" required:"true"`
}

// VillageSignupRequest is the body of POST /api/v1/auth/local/signup. It is
// accepted only when the deployment's registration mode is email; closed and
// invite modes refuse it. Responses do not reveal whether an account already
// exists.
type VillageSignupRequest struct {
	Email    string `json:"email" required:"true"`
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

// VillageVerifyEmailRequest is the body of POST /api/v1/auth/local/verify. The
// verification token is single use and expiring; only its digest is stored.
type VillageVerifyEmailRequest struct {
	Token string `json:"token" required:"true"`
}

// VillageSessionIssuedResponse is returned by every request that mints a
// server-side session: local login, invite acceptance, bootstrap claim, signup
// verification, and the forge callbacks. The session itself travels in an
// HttpOnly cookie, so the body carries no token; the account projection lets the
// caller route without a second read. must_change_password is true while the
// credential is a temporary one, and a true value means the forced-change gate
// only answers allowlisted surfaces until it is cleared.
type VillageSessionIssuedResponse struct {
	Account            VillageAccountProfile `json:"account"`
	MustChangePassword bool                  `json:"must_change_password"`
}

// VillageAdminCreateUserRequest is the body of POST /api/v1/admin/users.
// Administrators provision an account with a one-time temporary password whose
// expiry the server sets; email is optional and never an authentication factor.
type VillageAdminCreateUserRequest struct {
	Username string  `json:"username" required:"true"`
	Email    *string `json:"email"`
	IsAdmin  bool    `json:"is_admin,omitempty"`
}

// VillageTemporaryCredentialResponse is the one-time receipt for a freshly
// issued temporary password, from admin account creation or credential re-issue.
// The plaintext password is shown once and never stored; the server keeps only
// an Argon2id hash. temporary_password_expires_at is the deadline after which
// the credential can no longer sign in.
type VillageTemporaryCredentialResponse struct {
	Account                    VillageAccountProfile `json:"account"`
	TemporaryPassword          string                `json:"temporary_password"`
	TemporaryPasswordExpiresAt time.Time             `json:"temporary_password_expires_at"`
}

// VillageAdminUser is one row in the read-only admin member list. It carries no
// management affordances: handle, creation time, admin flag, and provisioning
// status only.
type VillageAdminUser struct {
	ID        VillageUUID               `json:"id"`
	Username  string                    `json:"username"`
	Email     *string                   `json:"email"`
	IsAdmin   bool                      `json:"is_admin"`
	Status    VillageProvisioningStatus `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
}

// VillageAdminUserListResponse is the response of GET /api/v1/admin/users.
// users is always a concrete array.
type VillageAdminUserListResponse struct {
	Users []VillageAdminUser `json:"users" nullable:"false"`
}

// VillageCreateInviteRequest is the body of POST /api/v1/admin/invites. Email is
// optional: an invitation may be issued as a bare link. expires_in_hours
// defaults to the server's invitation lifetime when omitted.
type VillageCreateInviteRequest struct {
	Email          *string `json:"email"`
	IsAdmin        bool    `json:"is_admin,omitempty"`
	ExpiresInHours *int32  `json:"expires_in_hours" minimum:"1" maximum:"8760"`
}

// VillageInvite is one invitation row. It never carries the raw token: only the
// digest is stored, and the plaintext link is returned once at creation.
type VillageInvite struct {
	ID         VillageUUID `json:"id"`
	Email      *string     `json:"email"`
	IsAdmin    bool        `json:"is_admin"`
	CreatedBy  VillageUUID `json:"created_by"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	AcceptedAt *time.Time  `json:"accepted_at"`
	RevokedAt  *time.Time  `json:"revoked_at"`
}

// VillageInviteListResponse is the response of GET /api/v1/admin/invites.
// invites is always a concrete array.
type VillageInviteListResponse struct {
	Invites []VillageInvite `json:"invites" nullable:"false"`
}

// VillageCreateInviteResponse is the one-time receipt for a newly created
// invitation. The plaintext token is returned once; later reads through
// VillageInviteListResponse expose only the metadata.
type VillageCreateInviteResponse struct {
	Invite VillageInvite `json:"invite"`
	Token  string        `json:"token"`
}

// Validate enforces the collection invariants the server guarantees: a sessions
// listing is never null. It mirrors the sibling Village envelopes so a consumer
// cannot mistake an omitted collection for an empty one.
func (r VillageAccountSessionListResponse) Validate() error {
	if r.Sessions == nil {
		return fmt.Errorf("Village session list validation failed at schema.VillageAccountSessionListResponse.Validate: sessions is null; the server always emits a concrete collection; emit [] when the account has no other sessions")
	}
	return nil
}

func (r VillageAdminUserListResponse) Validate() error {
	if r.Users == nil {
		return fmt.Errorf("Village admin user list validation failed at schema.VillageAdminUserListResponse.Validate: users is null; the server always emits a concrete collection; emit [] when the instance has no accounts")
	}
	return nil
}

func (r VillageInviteListResponse) Validate() error {
	if r.Invites == nil {
		return fmt.Errorf("Village invite list validation failed at schema.VillageInviteListResponse.Validate: invites is null; the server always emits a concrete collection; emit [] when no invitations are outstanding")
	}
	return nil
}

func (r VillageAuthProvidersResponse) Validate() error {
	if !r.RegistrationMode.IsValid() {
		return fmt.Errorf("Village auth providers validation failed at schema.VillageAuthProvidersResponse.Validate: registration_mode %q is outside the declared menu %v; a deployment must report a policy the contract describes", r.RegistrationMode, AllVillageRegistrationModes)
	}
	if r.Providers == nil {
		return fmt.Errorf("Village auth providers validation failed at schema.VillageAuthProvidersResponse.Validate: providers is null; the server always emits a concrete collection; emit [] when no forge provider is configured")
	}
	return nil
}
