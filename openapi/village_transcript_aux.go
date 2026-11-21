package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// villageAPIDocument describes the document envelope without recursively
// referring to the document that declares this operation.
type villageAPIDocument struct {
	OpenAPI string                 `json:"openapi" required:"true"`
	Info    map[string]interface{} `json:"info" required:"true" nullable:"false"`
	Paths   map[string]interface{} `json:"paths" required:"true" nullable:"false"`
}

// Description implements jsonschema.Described.
func (villageAPIDocument) Description() string {
	return "An OpenAPI 3.1 document: the frozen Village API document the server pins. Only its envelope is described here; the document itself is the contract for its members."
}

func addVillageTranscriptAuxOperations(r *openapi31.Reflector) error {
	transcriptPath := new(struct {
		ID schema.TranscriptID `path:"id" description:"Transcript identifier"`
	})
	attestationPath := new(struct {
		ID            schema.TranscriptID `path:"id" description:"Transcript identifier (not checked by this operation)"`
		AttestationID schema.VillageUUID  `path:"attestationID" description:"Attestation identifier"`
	})
	ops := []villageOperationSpec{
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodDelete, path: "/api/v1/transcripts/{id}", id: "deleteTranscript", tag: "transcripts",
			description: "Delete an owned transcript and its stored content. A missing blob store also returns 503; Retry-After is present only for a session lookup failure.",
			requests:    []interface{}{transcriptPath}, session: villageSessionRequired,
			response:      new(schema.VillageStatusResponse),
			errorStatuses: []int{400, 403, 404, 500, 503},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/attestations", id: "listTranscriptAttestations", tag: "transcripts",
			description: "List attestations for a visible transcript; inaccessible transcripts return 404.",
			requests:    []interface{}{transcriptPath}, session: villageSessionOptional,
			response: new([]schema.VillageTranscriptAttestation), errorStatuses: []int{400, 404, 500},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodPost, path: "/api/v1/transcripts/{id}/attestations", id: "createTranscriptAttestation", requiredBody: true, tag: "transcripts",
			description: "Attest to a visible transcript as a verified organization member. A caller who is not a verified member of the organization, and a duplicate of an existing attestation, are both refused with 403.",
			requests:    []interface{}{transcriptPath, new(schema.VillageCreateAttestationRequest)}, session: villageSessionRequired,
			response: new(schema.VillageTranscriptAttestation), successStatus: http.StatusCreated,
			errorStatuses: []int{400, 403, 404, 500},
			errorSummaries: map[int]string{
				http.StatusForbidden: "Forbidden. The caller is not a verified member of the organization, or the same attestation already exists; the database refusal is not told apart.",
			},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodDelete, path: "/api/v1/transcripts/{id}/attestations/{attestationID}", id: "deleteTranscriptAttestation", tag: "transcripts",
			description: "Delete the caller's attestation; also returns deleted when none matched. The transcript id is not checked.",
			requests:    []interface{}{attestationPath}, session: villageSessionRequired,
			response: new(schema.VillageStatusResponse), errorStatuses: []int{400, 500},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/annotations", id: "listTranscriptAnnotations", tag: "transcripts",
			description: "List annotations on a visible transcript; inaccessible transcripts return 404.",
			requests:    []interface{}{transcriptPath}, session: villageSessionOptional,
			response: new(schema.VillageTranscriptAnnotationsResponse), errorStatuses: []int{400, 404, 500},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodPost, path: "/api/v1/transcripts/{id}/annotations", id: "createTranscriptAnnotation", requiredBody: true, tag: "transcripts",
			description: "Create a manual entry annotation on a visible transcript.",
			requests:    []interface{}{transcriptPath, new(schema.VillageManualAnnotationRequest)}, session: villageSessionRequired,
			response: new(schema.AnnotationSummary), successStatus: http.StatusCreated, errorStatuses: []int{400, 404, 500},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/commits", id: "listTranscriptCommits", tag: "transcripts",
			description: "List persisted commits on a visible transcript in payload order; inaccessible transcripts return 404. When the transcript's owner is anonymous to the caller, every commit's message, authorName and authorEmail are null.",
			requests:    []interface{}{transcriptPath}, session: villageSessionOptional,
			response: new(schema.VillageTranscriptCommitsResponse), errorStatuses: []int{400, 404, 500},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodPost, path: "/api/v1/transcripts/publish/batch", id: "publishTranscriptBatch", tag: "transcripts",
			description: "A declared stub: batch publish has no behaviour, and every request with a valid session answers 501. This operation declares no request body and no success response.",
			session:     villageSessionRequired, declaredStub: true, errorStatuses: []int{http.StatusNotImplemented},
			errorSummaries: map[int]string{
				http.StatusNotImplemented: "Not Implemented. Batch publish has no behaviour; the request is not read.",
			},
		},
		{
			// unstring-allow: go-endpoint-literal — this declaration defines the route, and the route members are generated from it
			method: http.MethodGet, path: "/api/v1/openapi.json", id: "getVillageAPIDocument", tag: "operations",
			description: "Return the frozen Village API document this server pins, which declares this operation.",
			response:    new(villageAPIDocument),
		},
		{
			method: http.MethodGet, path: "/health", id: "getHealth", tag: "operations",
			description: "Liveness probe. Answers 200 whenever the process runs; migrations is absent when the migration registry could not be read.",
			servers:     []string{"/"},
			responses: []villageResponse{{status: http.StatusOK, description: "Process is alive.",
				headers: []villageHeader{{name: "Cache-Control", required: true, description: "no-store: a conforming cache does not store the response, so each probe is answered by the process."}},
				media:   []villageMedia{{mediaType: villageMediaJSON, body: new(schema.VillageHealthResponse)}}}},
		},
	}
	return addVillageOperations(r, ops)
}
