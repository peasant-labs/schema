package openapi

import (
	"fmt"
	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
	"net/http"
)

func addCollectiveObservationOperations(r *openapi31.Reflector) error {
	const base = "/api/v1/village/collective-observations"
	path := new(struct {
		CollectiveID schema.VillageUUID `path:"collectiveId"`
	})
	query := new(struct {
		Kind         schema.LocalContributionEvidenceKind `query:"kind" default:"remote-attempts"`
		CollectiveID schema.VillageUUID                   `query:"collectiveId"`
		Page         int                                  `query:"page" minimum:"1" maximum:"9007199254740991" default:"1"`
		Limit        int                                  `query:"limit" minimum:"1" maximum:"200" default:"50"`
		Revision     schema.ContributionTraversalRevision `query:"revision"`
	})
	ops := []localOperationSpec{
		{method: http.MethodGet, path: base, id: "listCollectiveObservations", tag: "village", description: "Read saved account-scoped relevant metadata without contacting Village. Never-fetched and known-empty remain distinct. Signed-out private reads never use a last-account fallback.", response: new(schema.CollectiveObservationResponse), errorStatuses: []int{400, 401, 500}, errorResponse: new(LocalErrorResponse)},
		{method: http.MethodGet, path: "/api/v1/village/contributions", id: "listLocalContributionEvidence", tag: "village", description: "Read exactly one bounded evidence category offline. Default remote-attempts, page 1, limit 50; limit 1..200. Page >1 requires revision. Malformed or missing required token is 400; valid selected scope/incarnation/kind/generation mismatch is 409 with currentTraversalRevision and restart guidance. Filter BEFORE total/page. Read token, coverage, filtered total and rows in one storage snapshot. Order remote attempts by collectiveId/transcriptId/event_num ascending, receipts by confirmedAt/receiptId ascending, diagnostics by attemptedAt/operationId ascending. No overall cache revision, joined catalog labels or poll timestamps participate in traversal.", requests: []interface{}{query}, response: new(schema.LocalContributionHistoryResponse), errorStatuses: []int{400, 401, 500}, errorResponse: new(LocalErrorResponse)},
		{method: http.MethodPost, path: base + "/fetch", id: "fetchCollectiveObservations", tag: "village", description: "Synchronously fetch metadata only, with no request body. Complete catalog/history install independently; complete association sets install per transcript. Never transfer bodies, publish, join or alter consent. Typed reports on non-2xx preserve every committed effect and prior completion date. Status precedence: persistence > auth > deadline/cancellation > remote/contract > superseded. 200 only complete; initial auth failure commits nothing. No detached job or refresh alias.", response: new(schema.CollectiveFetchReport), errorStatuses: []int{401, 409, 500, 502, 504}, errorResponse: new(schema.CollectiveFetchReport)},
		{method: http.MethodPut, path: base + "/{collectiveId}/tracking", id: "trackCollectiveObservation", tag: "village", description: "Verify selected collective visibility online before storing explicit tracking. No body; tracking grants no membership or publication consent. Verification failure writes no preference; unavailable resources use privacy-safe 404.", requests: []interface{}{path}, response: new(schema.CollectiveObservationResponse), errorStatuses: []int{400, 401, 404, 502, 504}, errorResponse: new(LocalErrorResponse)},
		{method: http.MethodDelete, path: base + "/{collectiveId}/tracking", id: "untrackCollectiveObservation", tag: "village", description: "Idempotently remove an explicit tracking preference using saved identity offline. No remote mutation or publication consent change.", requests: []interface{}{path}, response: new(schema.CollectiveObservationResponse), errorStatuses: []int{400, 401, 404, 502, 504}, errorResponse: new(LocalErrorResponse)},
	}
	ops[1].alternateResponses = map[int]interface{}{409: new(schema.LocalContributionTraversalConflict)}
	// Tracking still has a local persistence step even when remote verification
	// succeeds (or, for untracking, no remote request is made at all).
	ops[3].alternateResponses = map[int]interface{}{500: new(LocalErrorResponse)}
	ops[4].alternateResponses = map[int]interface{}{500: new(LocalErrorResponse)}
	if err := addLocalOperations(r, ops); err != nil {
		return err
	}
	// Signed-out callers have no authenticated UUID with which to construct a
	// truthful scoped report. A mid-invocation account change still returns the
	// full old-scope report, including already committed effects.
	item := r.Spec.Paths.MapOfPathItemValues[base+"/fetch"]
	if item.Post == nil || item.Post.Responses == nil {
		return fmt.Errorf("configure collective fetch auth alternatives: POST fetch is missing; restore the operation before schema generation")
	}
	response := item.Post.Responses.MapOfResponseOrReferenceValues["401"]
	if response.Response == nil {
		return fmt.Errorf("configure collective fetch auth alternatives: the 401 response is missing; restore its typed report before schema generation")
	}
	media, ok := response.Response.Content["application/json"]
	if !ok {
		return fmt.Errorf("configure collective fetch auth alternatives: JSON media is missing; restore the typed JSON response before generation")
	}
	media.Schema = map[string]interface{}{"oneOf": []interface{}{map[string]interface{}{"$ref": "#/components/schemas/SchemaCollectiveFetchReport"}, map[string]interface{}{"$ref": "#/components/schemas/OpenapiLocalErrorResponse"}}}
	response.Response.Content["application/json"] = media
	item.Post.Responses.MapOfResponseOrReferenceValues["401"] = response
	r.Spec.Paths.MapOfPathItemValues[base+"/fetch"] = item
	return nil
}
