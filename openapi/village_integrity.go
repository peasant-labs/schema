package openapi

import (
	"fmt"

	"github.com/swaggest/openapi-go/openapi31"
)

// contentHashHeader names the response header that tells a reader which copy
// of a transcript the content read served.
const contentHashHeader = "Village-Content-Hash"

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
