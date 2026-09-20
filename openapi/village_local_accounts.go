package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// addVillageLocalAccountOperations declares the full local-account wire surface:
// the account operations (bootstrap claim, local login, invite acceptance,
// self-serve signup and verification, logout, password change, profile read,
// sign-in-method discovery), the session operations (listing and revocation),
// and the administrator provisioning surface (account creation, read-only member
// list, invitation management, and credential re-issue).
//
// It declares shapes and statuses only: the handlers ship in the village
// repository. The middleware-produced statuses are declared here rather than left
// to each handler, so a client can read the full refusal set from the contract:
//
//   - 429 on the rate-limited unauthenticated credential entry points (claim,
//     login, invite acceptance, signup, verification), which share one limiter.
//   - 403 on CSRF refusals for cookie-authenticated unsafe requests; bearer and
//     API-key callers are exempt, so this status appears on the state-changing
//     routes rather than on reads.
//   - 403 on the forced-change gate while a password change is required; only
//     the allowlisted surfaces answer, and the request never reaches a handler.
//   - 401 for an anonymous caller and 403 for an authenticated non-admin on the
//     administrator routes, neither of which mutates state or exposes secrets.
//
// The session-bearing responses deliberately carry no token: the credential
// travels in an HttpOnly cookie, so every route that mints a session returns the
// account projection instead.
func addVillageLocalAccountOperations(r *openapi31.Reflector) error {
	sessionPath := new(struct {
		ID schema.VillageUUID `path:"id" description:"Session identifier"`
	})
	invitePath := new(struct {
		ID schema.VillageUUID `path:"id" description:"Invitation identifier"`
	})
	adminUserPath := new(struct {
		Username string `path:"username" description:"Neutral account handle"`
	})

	operations := []villageOperationSpec{
		// --- Account operations ---
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/bootstrap/claim",
			id:            "bootstrapClaimInstance",
			tag:           "auth",
			description:   "Claim a fresh instance once by creating its first administrator. The request carries the one-time install claim token together with the administrator handle and password. Exactly one claim succeeds, including under concurrent attempts; the durable install lock closes after the first success and survives restarts, and a second claim or a bad or expired token fails closed. A handle that already exists returns 409 and leaves the claim usable. No SMTP dependency is involved.",
			requests:      []interface{}{new(schema.VillageBootstrapClaimRequest)},
			response:      new(schema.VillageSessionIssuedResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict, http.StatusTooManyRequests, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/local/login",
			id:            "localLogin",
			tag:           "auth",
			description:   "Sign in with a local handle and password and mint a server-side session returned in an HttpOnly cookie. A wrong password and an unknown handle share one generic refusal that does not reveal whether an account exists. An expired temporary credential returns 401 with a distinct, actionable body. When a password change is required the session is minted with must_change_password true, and the forced-change gate then answers only allowlisted surfaces. Login works with no SMTP configuration.",
			requests:      []interface{}{new(schema.VillageLocalLoginRequest)},
			response:      new(schema.VillageSessionIssuedResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/local/accept-invite",
			id:            "acceptInvite",
			tag:           "auth",
			description:   "Accept an invitation by consuming its single-use token and creating the account. The consume is transactional: a handle collision returns 409 and rolls the whole consume back, leaving the invitation usable so the person can choose another handle. A reused, expired, or revoked invitation fails closed. Invitation tokens are stored only as digests and never appear in logs.",
			requests:      []interface{}{new(schema.VillageAcceptInviteRequest)},
			response:      new(schema.VillageSessionIssuedResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict, http.StatusGone, http.StatusTooManyRequests, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/local/signup",
			id:            "localSignup",
			tag:           "auth",
			description:   "Self-serve registration with email, handle, and password. Accepted only when the deployment's registration mode is email; closed and invite modes refuse it and create no account. The response does not reveal whether an account already exists, so a duplicate email or handle and a successful signup are indistinguishable. The created account is inactive until the verification link is consumed and cannot sign in beforehand. The endpoint is rate limited through the shared limiter.",
			requests:      []interface{}{new(schema.VillageSignupRequest)},
			response:      new(schema.VillageStatusResponse),
			successStatus: http.StatusAccepted,
			errorStatuses: []int{http.StatusBadRequest, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/local/verify",
			id:            "localVerifyEmail",
			tag:           "auth",
			description:   "Activate a self-serve account by consuming its single-use, expiring email verification token. Only the token digest is stored, and a replayed, expired, or invalid token fails closed. The activated account becomes able to sign in and is returned as the neutral account profile.",
			requests:      []interface{}{new(schema.VillageVerifyEmailRequest)},
			response:      new(schema.VillageAccountProfile),
			errorStatuses: []int{http.StatusBadRequest, http.StatusForbidden, http.StatusGone, http.StatusTooManyRequests, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/logout",
			id:            "localLogout",
			tag:           "auth",
			description:   "Revoke the presented session immediately. The session is resolved server-side, so a replayed cookie returns 401 afterward. A cross-site cookie-authenticated request is refused with 403 by CSRF protection before the handler runs.",
			response:      new(schema.VillageStatusResponse),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/auth/me/password",
			id:            "changeOwnPassword",
			tag:           "auth",
			description:   "Change the caller's password. A successful change rotates the current session, revokes every other session so a replayed cookie returns 401, and clears any temporary-credential expiry atomically. The current session's replacement is returned as an issued session. A cross-site cookie-authenticated request is refused with 403 by CSRF protection.",
			requests:      []interface{}{new(schema.VillageChangePasswordRequest)},
			response:      new(schema.VillageSessionIssuedResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodGet,
			path:          "/api/v1/auth/me",
			id:            "getOwnAccount",
			tag:           "auth",
			description:   "Read the caller's own account projection. It carries the neutral username, the self-only fields email and must_change_password, and no other account's data. The deprecated github_username stays populated for forge-created accounts until callers migrate to the neutral handle.",
			response:      new(schema.VillageAccountProfile),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodGet,
			path:          "/api/v1/auth/providers",
			id:            "listAuthProviders",
			tag:           "auth",
			description:   "Discover the enabled sign-in methods without attempting a login: whether the local username-and-password door is open, the deployment's registration mode, and the enabled forge providers. providers is always a concrete array, so a zero-provider instance is explicit rather than ambiguous.",
			response:      new(schema.VillageAuthProvidersResponse),
			errorStatuses: []int{http.StatusInternalServerError},
		},

		// --- Session operations ---
		{
			method:        http.MethodGet,
			path:          "/api/v1/auth/sessions",
			id:            "listOwnSessions",
			tag:           "auth",
			description:   "List the caller's server-side sessions. No token or digest is exposed: the raw session value lives only in the HttpOnly cookie. current marks the session that served the request. sessions is always a concrete array.",
			response:      new(schema.VillageAccountSessionListResponse),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodDelete,
			path:          "/api/v1/auth/sessions/{id}",
			id:            "revokeOwnSession",
			tag:           "auth",
			description:   "Revoke one of the caller's own sessions by identifier. A revoked session fails closed, so its replayed cookie returns 401. A session the caller does not own is not distinguishable from a missing one. A cross-site cookie-authenticated request is refused with 403 by CSRF protection.",
			requests:      []interface{}{sessionPath},
			response:      new(schema.VillageStatusResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError},
		},

		// --- Admin operations ---
		{
			method:        http.MethodPost,
			path:          "/api/v1/admin/users",
			id:            "adminCreateUser",
			tag:           "admin",
			description:   "Provision an account as an administrator with a one-time temporary password whose expiry the server sets and enforces before, at, and after the window. Email is optional and never an authentication factor. The plaintext password is shown once and only its Argon2id hash is stored. A handle that already exists returns 409. Anonymous callers receive 401 and authenticated non-admins receive 403; neither mutates state or exposes secrets.",
			requests:      []interface{}{new(schema.VillageAdminCreateUserRequest)},
			response:      new(schema.VillageTemporaryCredentialResponse),
			successStatus: http.StatusCreated,
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict, http.StatusInternalServerError},
		},
		{
			method:        http.MethodGet,
			path:          "/api/v1/admin/users",
			id:            "adminListUsers",
			tag:           "admin",
			description:   "Read the read-only member list: one row per account with the neutral handle, creation time, admin flag, and provisioning status (pending or activated). It carries no management actions. Anonymous callers receive 401 and authenticated non-admins receive 403.",
			response:      new(schema.VillageAdminUserListResponse),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/admin/invites",
			id:            "adminCreateInvite",
			tag:           "admin",
			description:   "Issue a single-use, expiring invitation as an administrator. Email is optional, so an invitation may be a bare link. The plaintext token is returned once at creation; later reads expose only metadata, and only the digest is stored. Anonymous callers receive 401 and authenticated non-admins receive 403.",
			requests:      []interface{}{new(schema.VillageCreateInviteRequest)},
			response:      new(schema.VillageCreateInviteResponse),
			successStatus: http.StatusCreated,
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodGet,
			path:          "/api/v1/admin/invites",
			id:            "adminListInvites",
			tag:           "admin",
			description:   "List outstanding and historical invitations as an administrator, without raw tokens. Anonymous callers receive 401 and authenticated non-admins receive 403.",
			response:      new(schema.VillageInviteListResponse),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError},
		},
		{
			method:        http.MethodDelete,
			path:          "/api/v1/admin/invites/{id}",
			id:            "adminRevokeInvite",
			tag:           "admin",
			description:   "Revoke an outstanding invitation as an administrator. A revoked invitation fails closed when presented for acceptance. Anonymous callers receive 401 and authenticated non-admins receive 403.",
			requests:      []interface{}{invitePath},
			response:      new(schema.VillageStatusResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/admin/users/{username}/temporary-password",
			id:            "adminReissueTemporaryPassword",
			tag:           "admin",
			description:   "Re-issue a one-time temporary password for an account that has not yet activated. An activated account returns 409, and issuing a new credential invalidates the previous one. The shown-once password carries a fresh, enforced expiry. Anonymous callers receive 401 and authenticated non-admins receive 403.",
			requests:      []interface{}{adminUserPath},
			response:      new(schema.VillageTemporaryCredentialResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
		},
	}
	return addVillageOperations(r, operations)
}
