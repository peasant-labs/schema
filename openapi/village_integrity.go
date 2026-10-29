package openapi

import (
	"fmt"
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// contentHashHeader names the response header that tells a reader which copy
// of a transcript the content read served.
const contentHashHeader = "Village-Content-Hash"

// addVillageIntegrityOperations declares the owner's integrity check. The
// answer is only a result and the owner's remaining daily allowance; it never
// carries turns, a rule or a field.
func addVillageIntegrityOperations(r *openapi31.Reflector) error {
	transcriptPath := new(struct {
		ID schema.TranscriptID `path:"id" description:"Transcript identifier"`
	})
	return addVillageOperations(r, []villageOperationSpec{
		{
			method:      http.MethodPost,
			path:        "/api/v1/transcripts/{id}/integrity-check",
			id:          "checkTranscriptIntegrity",
			tag:         "transcripts",
			description: "Ask village to check one of the caller's own transcripts for integrity. Only the signed-in owner may call it. The call is synchronous: the response carries the answer. Statuses are decided in this order: 401 when the caller is signed out; 400 for a malformed id; 404 when the caller cannot read the transcript; 403 when the caller can read it but does not own it; 409 when the owner's transcript is public; 503 when the check is not available; 429 when the owner's daily allowance is used up; otherwise village runs the check, and answers 503 if it does not complete, or 200. 409 means a public transcript, which is checked automatically. 429 means the owner's daily checks for the UTC day are used up. 503 means the check is not available now. A check that does not complete spends nothing, and an unchecked result spends nothing. The answer is only flagged, not_flagged or unchecked, with no turns, rule or field. remaining_today is the number of checks left after this one. A requested check is never shown publicly and never affects the public status.",
			requests:    []interface{}{transcriptPath},
			response:    new(schema.VillageIntegrityCheckResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusConflict,
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
				http.StatusServiceUnavailable,
			},
		},
	})
}

// setTranscriptContentHashHeader declares, on the 200 response of the
// transcript content read, the optional header carrying the bare content hash
// of the stored copy the response was served from.
func setTranscriptContentHashHeader(spec *openapi31.Spec) error {
	const path = "/api/v1/transcripts/{id}/content"
	item, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok || item.Get == nil || item.Get.Responses == nil {
		return fmt.Errorf("declare %s: GET %s is not registered; readers could not tell which copy they were served; register the content read first", contentHashHeader, path)
	}
	response, ok := item.Get.Responses.MapOfResponseOrReferenceValues["200"]
	if !ok || response.Response == nil {
		return fmt.Errorf("declare %s: GET %s has no 200 response", contentHashHeader, path)
	}
	if _, ok := spec.Components.Schemas["SchemaTranscriptContentHash"]; !ok {
		return fmt.Errorf("declare %s: the TranscriptContentHash component is missing, so the header could not reference it", contentHashHeader)
	}
	if response.Response.Headers == nil {
		response.Response.Headers = map[string]openapi31.HeaderOrReference{}
	}
	description := "The bare content_hash, without quotes, of the stored copy this response was served from; it is not a hash of the served bytes. A rewrite on read sends the new hash, and withholding repository fields for an anonymous owner does not change it. Absent while village holds no hash for the served copy. A reader highlights flagged turns only when it equals transcript.content_hash from the metadata read. A plain header, not an ETag."
	response.Response.Headers[contentHashHeader] = openapi31.HeaderOrReference{Header: &openapi31.Header{
		Description: &description,
		Schema:      schemaRef("SchemaTranscriptContentHash"),
	}}
	item.Get.Responses.MapOfResponseOrReferenceValues["200"] = response
	spec.Paths.MapOfPathItemValues[path] = item
	return nil
}
