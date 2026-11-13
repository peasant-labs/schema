package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// addLocalOwnerOperations declares the local owner read and the claim prompt
// controls. None of these routes contacts Village: the owner and the claim log
// are local state. Both POST routes declare the shared cross-origin refusal
// through addLocalOperations.
func addLocalOwnerOperations(r *openapi31.Reflector) error {
	return addLocalOperations(r, []localOperationSpec{
		{
			method: http.MethodGet, path: "/api/v1/local-owner", id: "getLocalOwnerStatus", tag: "local-owner",
			description: "Read this data directory's one permanent local owner, the most recent claim entries, the full claim total, the claim prompt state, and the ownership counts. Entries are ordered by claimedAt then insertion order. Pass before=<nextBefore> to read the next older page. This is a local read: it needs no Village login and makes no network call. claims is [] when this directory has no claims.",
			requests: []interface{}{new(struct {
				Before string `query:"before" minLength:"1" maxLength:"64" description:"Older-page cursor: the nextBefore value from the previous status read. Omit for the most recent page."`
			})},
			response:      new(schema.LocalOwnerStatus),
			errorStatuses: []int{http.StatusInternalServerError},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/local-owner/claim", id: "claimLocalOwner", tag: "local-owner",
			description:   "Claim local sessions and annotations for the signed-in Village account. all=false claims only unclaimed artifacts; all=true claims every local session and annotation. An explicit claim appends one entry and never rewrites an earlier one; an automatic claim reuses the account's latest entry. Pulled transcripts and annotations are never claimed. The answer is the recomputed status plus what this action claimed.",
			requests:      []interface{}{new(schema.LocalOwnerClaimRequest)},
			response:      new(schema.LocalOwnerClaimResult),
			requiredBody:  true,
			errorStatuses: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError, http.StatusGatewayTimeout},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/local-owner/decline", id: "declineLocalOwnerPrompt", tag: "local-owner",
			description:   "Decline the claim prompt for the signed-in Village account. A decline writes no claim entry; it records only that the prompt was dismissed so the product does not ask again for the same local owner plus account scope. The claim control stays available.",
			response:      new(schema.LocalOwnerStatus),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusInternalServerError},
			errorResponse: new(LocalErrorResponse),
		},
	})
}
