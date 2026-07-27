package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// annotationPushRequestSchemaURL is the in-compiler handle for the standalone
// annotation-push request schema. It is never fetched over the network.
const annotationPushRequestSchemaURL = "annotation-push-request.schema.json"

var (
	annotationPushRequestSchema     *jsonschema.Schema
	annotationPushRequestSchemaErr  error
	annotationPushRequestSchemaOnce sync.Once
)

func compileAnnotationPushRequestSchema() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(annotationPushRequestSchemaURL, bytes.NewReader(AnnotationPushRequestSchemaJSON())); err != nil {
		annotationPushRequestSchemaErr = err
		return
	}
	annotationPushRequestSchema, annotationPushRequestSchemaErr = compiler.Compile(annotationPushRequestSchemaURL)
}

// ValidateAnnotationPushRequest validates an annotation-push JSON body against
// the generated operation schema, then decodes AnnotationPushRequest and calls
// its typed validator. The schema enforces the documented structural contract;
// the typed validator additionally enforces relational rules such as
// entryTarget.endIndex being greater than entryTarget.entryIndex.
func ValidateAnnotationPushRequest(data []byte) error {
	annotationPushRequestSchemaOnce.Do(compileAnnotationPushRequestSchema)
	if annotationPushRequestSchemaErr != nil {
		return annotationPushRequestSchemaErr
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := annotationPushRequestSchema.Validate(value); err != nil {
		return err
	}

	var request AnnotationPushRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return fmt.Errorf("decode annotation push request after generated schema validation: %w", err)
	}
	return request.Validate()
}
