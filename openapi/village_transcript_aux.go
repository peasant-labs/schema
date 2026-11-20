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
			method: http.MethodDelete, path: "/api/v1/transcripts/{id}", id: "deleteTranscript", tag: "transcripts",
			description: "Delete an owned transcript and its stored content. A missing blob store also returns 503; Retry-After is present only for a session lookup failure.",
			requests:    []interface{}{transcriptPath}, session: villageSessionRequired,
			response:      new(schema.VillageStatusResponse),
			errorStatuses: []int{400, 403, 404, 500, 503},
		},
		{
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/attestations", id: "listTranscriptAttestations", tag: "transcripts",
			description: "List attestations for a visible transcript; inaccessible transcripts return 404.",
			requests:    []interface{}{transcriptPath}, session: villageSessionOptional,
			response: new([]schema.VillageTranscriptAttestation), errorStatuses: []int{400, 404, 500},
		},
		{
			method: http.MethodPost, path: "/api/v1/transcripts/{id}/attestations", id: "createTranscriptAttestation", tag: "transcripts",
			description: "Attest to a visible transcript as a verified organization member. Duplicate errors can return 409; errors containing violates return 403 first.",
			requests:    []interface{}{transcriptPath, new(schema.VillageCreateAttestationRequest)}, session: villageSessionRequired,
			response: new(schema.VillageTranscriptAttestation), successStatus: http.StatusCreated,
			errorStatuses: []int{400, 403, 404, 409, 500},
		},
		{
			method: http.MethodDelete, path: "/api/v1/transcripts/{id}/attestations/{attestationID}", id: "deleteTranscriptAttestation", tag: "transcripts",
			description: "Delete the caller's attestation; also returns deleted when none matched. The transcript id is not checked.",
			requests:    []interface{}{attestationPath}, session: villageSessionRequired,
			response: new(schema.VillageStatusResponse), errorStatuses: []int{400, 500},
		},
		{
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/annotations", id: "listTranscriptAnnotations", tag: "transcripts",
			description: "List annotations on a visible transcript; inaccessible transcripts return 404.",
			requests:    []interface{}{transcriptPath}, session: villageSessionOptional,
			response: new(schema.VillageTranscriptAnnotationsResponse), errorStatuses: []int{400, 404, 500},
		},
		{
			method: http.MethodPost, path: "/api/v1/transcripts/{id}/annotations", id: "createTranscriptAnnotation", tag: "transcripts",
			description: "Create a manual entry annotation on a visible transcript.",
			requests:    []interface{}{transcriptPath, new(schema.VillageManualAnnotationRequest)}, session: villageSessionRequired,
			response: new(schema.AnnotationSummary), successStatus: http.StatusCreated, errorStatuses: []int{400, 404, 500},
		},
		{
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/commits", id: "listTranscriptCommits", tag: "transcripts",
			description: "List persisted commits on a visible transcript in payload order; inaccessible transcripts return 404.",
			requests:    []interface{}{transcriptPath}, session: villageSessionOptional,
			response: new(schema.VillageTranscriptCommitsResponse), errorStatuses: []int{400, 404, 500},
		},
		{
			method: http.MethodPost, path: "/api/v1/transcripts/publish/batch", id: "publishTranscriptBatch", tag: "transcripts",
			description: "Batch publish is not implemented. No request body is declared; the body and success shapes arrive with the implementation in a later version.",
			session:     villageSessionRequired, declaredStub: true, errorStatuses: []int{http.StatusNotImplemented},
		},
		{
			method: http.MethodGet, path: "/api/v1/openapi.json", id: "getVillageAPIDocument", tag: "operations",
			description: "Return the frozen Village API document this server pins, which declares this operation.",
			response:    new(villageAPIDocument),
		},
		{
			method: http.MethodGet, path: "/health", id: "getHealth", tag: "operations",
			description: "Liveness probe. Answers 200 even if the migration registry cannot be read, when migrations is absent.",
			servers:     []string{"/"},
			responses: []villageResponse{{status: http.StatusOK, description: "Process is alive.",
				headers: []villageHeader{{name: "Cache-Control", required: true, description: "no-store"}},
				media:   []villageMedia{{mediaType: villageMediaJSON, body: new(schema.VillageHealthResponse)}}}},
		},
	}
	if err := addVillageOperations(r, ops); err != nil {
		return err
	}
	if err := requireRequestBody(r.Spec, http.MethodPost, "/api/v1/transcripts/{id}/attestations"); err != nil {
		return err
	}
	return requireRequestBody(r.Spec, http.MethodPost, "/api/v1/transcripts/{id}/annotations")
}
