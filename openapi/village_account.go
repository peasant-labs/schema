package openapi

import (
	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
	"net/http"
)

func addVillageAccountOperations(r *openapi31.Reflector) error {
	cookie := villageHeader{name: "Set-Cookie", required: true, description: "Clears peasant_token with Path=/, Max-Age=0 and HttpOnly."}
	ops := []villageOperationSpec{
		{method: http.MethodPost, path: "/api/v1/auth/logout", id: "logOut", tag: "auth", description: "Clear the session cookie; bearer tokens are not revoked.", responses: []villageResponse{{status: 200, description: "Logged out.", headers: []villageHeader{cookie}, media: []villageMedia{{mediaType: villageMediaJSON, body: new(schema.VillageStatusResponse)}}}}},
		{method: http.MethodDelete, path: "/api/v1/auth/me", id: "deleteMyAccount", tag: "auth", description: "Permanently delete the authenticated account and clear the session cookie.", session: villageSessionRequired, responses: []villageResponse{{status: 200, description: "Deleted.", headers: []villageHeader{cookie}, media: []villageMedia{{mediaType: villageMediaJSON, body: new(schema.VillageStatusResponse)}}}}, errorStatuses: []int{500}},
		{method: http.MethodPatch, path: "/api/v1/auth/me/settings", id: "updateMyDiscoverability", tag: "auth", description: "Update discoverability; is_discoverable must be present and non-null.", session: villageSessionRequired, requests: []interface{}{new(schema.VillageDiscoverabilityRequest)}, response: new(schema.VillageUser), errorStatuses: []int{400, 500}},
		{method: http.MethodPatch, path: "/api/v1/auth/me/username", id: "setMyUsername", tag: "auth", description: "Set a username after trimming and lowercasing; invalid or reserved names are refused and taken names conflict.", session: villageSessionRequired, requests: []interface{}{new(schema.VillageUsernameRequest)}, response: new(schema.VillageUser), errorStatuses: []int{400, 409, 500}},
		{method: http.MethodPost, path: "/api/v1/auth/api-keys", id: "createMyAPIKey", tag: "auth", description: "Create an API key; plaintext is returned only here. A missing or invalid body gives an empty label.", session: villageSessionRequired, requests: []interface{}{new(schema.VillageCreateAPIKeyRequest)}, response: new(schema.VillageCreatedAPIKey), successStatus: 201, errorStatuses: []int{500}},
		{method: http.MethodDelete, path: "/api/v1/auth/api-keys/{id}", id: "revokeMyAPIKey", tag: "auth", description: "Revoke the caller's API key; a nonmatching key also returns revoked.", session: villageSessionRequired, requests: []interface{}{new(struct {
			ID schema.VillageUUID `path:"id"`
		})}, response: new(schema.VillageStatusResponse), errorStatuses: []int{400, 500}},
		{method: http.MethodPatch, path: "/api/v1/auth/orgs/{orgLogin}/visibility", id: "setMyOrganizationVisibility", tag: "auth", description: "Set organization visibility; omitted visible means false and a nonmatching organization still returns updated.", session: villageSessionRequired, requests: []interface{}{new(struct {
			OrgLogin string `path:"orgLogin" minLength:"1"`
		}), new(schema.VillageOrganizationVisibilityRequest)}, response: new(schema.VillageOrganizationVisibilityResponse), errorStatuses: []int{400, 500}},
	}
	if err := addVillageOperations(r, ops); err != nil {
		return err
	}
	for _, path := range []string{"/api/v1/auth/me/settings", "/api/v1/auth/me/username", "/api/v1/auth/orgs/{orgLogin}/visibility"} {
		if err := requireRequestBody(r.Spec, http.MethodPatch, path); err != nil {
			return err
		}
	}
	return nil
}
