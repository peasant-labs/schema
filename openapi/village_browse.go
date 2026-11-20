package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

func addVillageBrowseOperations(r *openapi31.Reflector) error {
	username := new(struct {
		Username string `path:"username"`
	})
	project := new(struct {
		ProjectHash schema.ProjectHash `path:"projectHash"`
	})
	userProject := new(struct {
		Username    string             `path:"username"`
		ProjectHash schema.ProjectHash `path:"projectHash"`
	})
	search := new(struct {
		Q     string `query:"q" description:"Trimmed search text; empty returns an empty list."`
		Limit string `query:"limit" description:"Positive integer; defaults to 10 and is capped at 50. Invalid values are ignored."`
	})
	limit := new(struct {
		Limit string `query:"limit" description:"Positive integer; defaults to 10. Invalid values are ignored."`
	})
	ops := []villageOperationSpec{
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/orgs/search", id: "searchOrganizations", tag: "organizations", description: "Search organizations by login. Results are ordered by member count then transcript count, descending; an empty query returns an empty list.", session: villageSessionOptional, requests: []interface{}{search}, response: new(schema.VillageOrganizationSearchResponse), errorStatuses: []int{500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/tags", id: "listTags", tag: "tags", description: "List tags ordered by name, including tags with no transcripts.", response: new([]schema.VillageTagUsage), errorStatuses: []int{500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/tags/popular", id: "listPopularTags", tag: "tags", description: "List tags with transcripts by descending usage. The limit defaults to 10 and has no cap.", requests: []interface{}{limit}, response: new([]schema.VillageTagUsage), errorStatuses: []int{500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/users/{username}", id: "getUserProfile", tag: "users", description: "Read a public profile. An unknown or undiscoverable user answers 404 unless the caller owns the hidden profile.", session: villageSessionOptional, requests: []interface{}{username}, response: new(schema.VillagePublicOwner), errorStatuses: []int{404}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/users/{username}/orgs", id: "listUserOrganizations", tag: "users", description: "List the visible organizations of a discoverable user, or the caller's own hidden profile. Unknown and hidden users answer 404.", session: villageSessionOptional, requests: []interface{}{username}, response: new([]schema.VillageListUserOrganization), errorStatuses: []int{404, 500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodGet, path: "/api/v1/users/{username}/projects/{projectHash}", id: "getUserProject", tag: "users", description: "Read a user's project and the transcripts and collectives visible to the caller. Unknown users, hidden users and unowned hashes share one 404.", session: villageSessionOptional, requests: []interface{}{userProject}, response: new(schema.VillageUserProjectResponse), errorStatuses: []int{400, 404, 500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodPatch, path: "/api/v1/users/me/projects/{projectHash}", id: "setMyProjectDisplayName", requiredBody: true, tag: "users", description: "Set the display name of the caller's project. The name is trimmed and must be non-empty and at most 255 bytes.", session: villageSessionRequired, requests: []interface{}{project, new(schema.VillageProjectDisplayNameRequest)}, response: new(schema.VillageResolvedProject), errorStatuses: []int{400, 404, 500}},
		// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
		{method: http.MethodDelete, path: "/api/v1/users/me/projects/{projectHash}/display-name", id: "clearMyProjectDisplayName", tag: "users", description: "Remove the caller's project display-name correction. A missing correction answers 404.", session: villageSessionRequired, requests: []interface{}{project}, response: new(schema.VillageResolvedProject), errorStatuses: []int{400, 404, 500}},
	}
	return addVillageOperations(r, ops)
}
