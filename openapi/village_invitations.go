package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// addVillageInvitationOperations declares collective invitations and the
// caller's notification email.
//
// An owner adding a person creates an invitation the person accepts or
// declines; there is no operation that makes someone a member without their
// consent. Inviting never confirms that an account exists, and the owner sees
// an invitation only as the handle they typed. The notification address is
// returned only in the caller's own settings.
func addVillageInvitationOperations(r *openapi31.Reflector) error {
	groupPath := new(struct {
		ID schema.VillageUUID `path:"id" description:"Collective identifier"`
	})
	groupInvitationPath := new(struct {
		ID           schema.VillageUUID `path:"id" description:"Collective identifier"`
		InvitationID schema.VillageUUID `path:"invitationID" description:"Invitation identifier"`
	})
	myInvitationPath := new(struct {
		InvitationID schema.VillageUUID `path:"invitationID" description:"Invitation identifier"`
	})
	unsubscribeQuery := new(struct {
		Token string `query:"token" required:"true" minLength:"1" description:"The signed token from the email's unsubscribe link"`
	})

	operations := []villageOperationSpec{
		{
			method:        http.MethodPost,
			path:          "/api/v1/groups/{id}/invitations",
			id:            "createGroupInvitation",
			tag:           "collectives",
			description:   "Invite a person to an owned collective by username. The person becomes a member only by accepting. The answer is 202 with status invited for every username, whether no account holds it, its account is not discoverable, or it is discoverable, so inviting never confirms that an account exists. Every invited name gets an invitation that the owners' list shows until it expires or is revoked, whether or not an account holds the name. Any invitation email is sent after this response, off the request path, so the response's timing never depends on the invitee. Inviting the same username again while an invitation is open is the same 202 and notifies no one. An invitation the person declined stays open to the owner, unchanged, until it expires. 409 means the person is already a member who consented; a member whose membership predates invitations may be invited again so that they can consent. An owner's invitation admits the person whatever the collective's acceptance mode, including curated and verified_only. 429 means the inviter or the collective has reached its invitation limit.",
			requests:      []interface{}{groupPath, new(schema.VillageCollectiveInvitationRequest)},
			response:      new(schema.VillageStatusResponse),
			successStatus: http.StatusAccepted,
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusConflict,
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/{id}/invitations",
			id:          "listGroupInvitations",
			tag:         "collectives",
			description: "List an owned collective's open invitations, each as the handle the inviting owner typed. Pending and declined invitations that have not expired are listed alike, with no status, so a decline is not revealed. Every invited name is listed until its invitation expires or is revoked, whether or not an account holds the name. An invitation whose invitee deletes their account stays listed, unchanged, until it expires, so a deletion is not revealed either. An accepted invitation leaves the list and its person appears in the roster; a revoked or expired one leaves the list.",
			requests:    []interface{}{groupPath},
			response:    new(schema.VillageCollectiveInvitationsSentResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/groups/{id}/invitations/{invitationID}",
			id:          "revokeGroupInvitation",
			tag:         "collectives",
			description: "Revoke an open invitation of an owned collective. The invitation leaves the invitee's list and nothing is recorded. Revoking any listed invitation answers 200, including one for a name no account holds or one whose invitee has deleted their account, so revoking reveals nothing about the invitee. 404 means the collective has no open invitation with this id.",
			requests:    []interface{}{groupInvitationPath},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/users/me/invitations",
			id:          "listMyInvitations",
			tag:         "users",
			description: "List the caller's own pending invitations that have not expired, each with the collective's summary and the inviting owner as the caller may see them. The summary is disclosed even for a collective visible only to its members. Another person's invitations are never listed.",
			response:    new(schema.VillageMyInvitationsResponse),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/users/me/invitations/{invitationID}/accept",
			id:          "acceptMyInvitation",
			tag:         "users",
			description: "Accept one of the caller's pending invitations and join the collective as a member. Accepting is consent: the collective's owners may then see the caller's profile, while other members still see the caller as anonymous. Accepting when the caller is already a member who consented joins nothing new and answers 200. 404 means the caller holds no pending, unexpired invitation with this id, including one addressed to someone else. The operation takes no body; a later choice made at the moment of joining, such as how the caller is credited, would be an optional body.",
			requests:    []interface{}{myInvitationPath},
			response:    new(schema.VillageCollectiveStatusRoleResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/users/me/invitations/{invitationID}/decline",
			id:          "declineMyInvitation",
			tag:         "users",
			description: "Decline one of the caller's pending invitations. The owners are not told: the invitation stays in their list unchanged until it expires, and inviting the caller again from the same collective before then notifies no one. 404 means the caller holds no pending, unexpired invitation with this id, including one addressed to someone else.",
			requests:    []interface{}{myInvitationPath},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPut,
			path:        "/api/v1/users/me/email",
			id:          "setMyNotificationEmail",
			tag:         "users",
			description: "Set the caller's notification email address, replacing any earlier one, and email a verification token to it. The address is unverified until the token is confirmed, and Village emails invitations only to a verified address. The answer is the caller's own settings. 429 means too many verification emails were requested; 501 means this Village sends no email.",
			requests:    []interface{}{new(schema.VillageSetNotificationEmailRequest)},
			response:    new(schema.VillageUserSettings),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
				http.StatusNotImplemented,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/users/me/email",
			id:          "removeMyNotificationEmail",
			tag:         "users",
			description: "Remove the caller's notification email address. Village keeps no copy of it. The answer is the caller's own settings.",
			response:    new(schema.VillageUserSettings),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/users/me/email/verify",
			id:          "verifyMyNotificationEmail",
			tag:         "users",
			description: "Confirm the caller's notification email address with the token Village emailed to it. The token is bound to the caller's account and to that address, so it confirms nothing for anyone else. 400 means the token is wrong, expired, or for an address the caller has since replaced; the answer does not say which. The answer is the caller's own settings.",
			requests:    []interface{}{new(schema.VillageVerifyNotificationEmailRequest)},
			response:    new(schema.VillageUserSettings),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/email/unsubscribe",
			id:          "unsubscribeInvitationEmails",
			tag:         "users",
			description: "Turn off invitation emails for the account a signed unsubscribe token names, without signing in. This is the one-click unsubscribe target (RFC 8058) that every invitation email links to; a request body, such as List-Unsubscribe=One-Click, is ignored. The token can only turn this one setting off. 400 means the token is malformed, forged, or expired.",
			requests:    []interface{}{unsubscribeQuery},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			},
		},
	}

	if err := addVillageOperations(r, operations); err != nil {
		return err
	}
	for _, body := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/groups/{id}/invitations"},
		{http.MethodPut, "/api/v1/users/me/email"},
		{http.MethodPost, "/api/v1/users/me/email/verify"},
	} {
		if err := requireRequestBody(r.Spec, body.method, body.path); err != nil {
			return err
		}
	}
	return nil
}
