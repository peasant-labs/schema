package openapi

import (
	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
	"net/http"
)

func addVillageAccountOperations(r *openapi31.Reflector) error {
	cookie := villageHeader{name: "Set-Cookie", required: true, description: "Clears peasant_token with Path=/, Max-Age=0 and HttpOnly."}
	ops := []villageOperationSpec{
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodPost, path: "/api/v1/auth/logout", id: "logOut", tag: "auth", description: "Clear the session cookie; bearer tokens are not revoked.", responses: []villageResponse{{status: 200, description: "Logged out.", headers: []villageHeader{cookie}, media: []villageMedia{{mediaType: villageMediaJSON, body: new(schema.VillageStatusResponse)}}}}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodDelete, path: "/api/v1/auth/me", id: "deleteMyAccount", tag: "auth", description: "Permanently delete the authenticated account and clear the session cookie.", session: villageSessionRequired, responses: []villageResponse{{status: 200, description: "Deleted.", headers: []villageHeader{cookie}, media: []villageMedia{{mediaType: villageMediaJSON, body: new(schema.VillageStatusResponse)}}}}, errorStatuses: []int{500}},
		{
			method: http.MethodPatch,
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			path:          "/api/v1/auth/me/settings",
			id:            "setMyDiscoverability",
			requiredBody:  true,
			deprecated:    true,
			tag:           "auth",
			description:   "Deprecated: use PATCH /api/v1/users/me/settings instead; this route is expected to stop answering after its Sunset date. Update discoverability; is_discoverable must be present and non-null.",
			session:       villageSessionRequired,
			requests:      []interface{}{new(schema.VillageDiscoverabilityRequest)},
			response:      new(schema.VillageUser),
			errorStatuses: []int{400, 500},
			responseHeaders: []villageHeader{
				{name: "Deprecation", required: true, description: "RFC 9745 deprecation date as an RFC 9651 structured Date (for example @1796083200), indicating when this route was deprecated."},
				{name: "Sunset", required: true, description: "RFC 8594 HTTP-date after which this route is expected to stop answering."},
			},
		},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodPatch, path: "/api/v1/auth/me/username", id: "setMyUsername", requiredBody: true, tag: "auth", description: "Set a username after trimming and lowercasing; invalid or reserved names are refused and taken names conflict.", session: villageSessionRequired, requests: []interface{}{new(schema.VillageUsernameRequest)}, response: new(schema.VillageUser), errorStatuses: []int{400, 409, 500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodPost, path: "/api/v1/auth/api-keys", id: "createMyAPIKey", tag: "auth", description: "Create an API key; plaintext is returned only here. A missing or invalid body gives an empty label.", session: villageSessionRequired, requests: []interface{}{new(schema.VillageCreateAPIKeyRequest)}, response: new(schema.VillageCreatedAPIKey), successStatus: 201, errorStatuses: []int{500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/auth/api-keys", id: "listMyAPIKeys", tag: "auth", description: "List API keys newest first, including revoked keys, without plaintext or hashes; an empty list is [].", session: villageSessionRequired, response: new([]schema.VillageAPIKey), errorStatuses: []int{500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodDelete, path: "/api/v1/auth/api-keys/{id}", id: "revokeMyAPIKey", tag: "auth", description: "Revoke the caller's API key; a nonmatching key also returns revoked.", session: villageSessionRequired, requests: []interface{}{new(struct {
			ID schema.VillageUUID `path:"id"`
		})}, response: new(schema.VillageStatusResponse), errorStatuses: []int{400, 500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodPatch, path: "/api/v1/auth/orgs/{orgLogin}/visibility", id: "setMyOrganizationVisibility", requiredBody: true, tag: "auth", description: "Set organization visibility; omitted visible means false and a nonmatching organization still returns updated.", session: villageSessionRequired, requests: []interface{}{new(struct {
			OrgLogin string `path:"orgLogin" minLength:"1"`
		}), new(schema.VillageOrganizationVisibilityRequest)}, response: new(schema.VillageOrganizationVisibilityResponse), errorStatuses: []int{400, 500}},
	}
	return addVillageOperations(r, ops)
}
