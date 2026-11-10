package schema_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

func TestAttachmentRouteRefFixtures(t *testing.T) {
	fixtures := loadAttachmentRouteFixtures(t)
	assert.RequireMin(t, fixtures.Refs, 8)
	assert.RequireValid(t, fixtures.Refs)
	if fixtures.Operation.Pattern != schema.AttachmentRefPattern {
		t.Fatalf("fixture pattern %q disagrees with schema.AttachmentRefPattern %q", fixtures.Operation.Pattern, schema.AttachmentRefPattern)
	}
	pattern, err := regexp.Compile(fixtures.Operation.Pattern)
	if err != nil {
		t.Fatalf("compile attachment ref pattern: %v", err)
	}
	for _, fixtureCase := range fixtures.Refs.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			err := schema.ValidateAttachmentRef(fixtureCase.Input.Ref)
			accepted := err == nil
			if accepted != fixtureCase.Expected.Accepted {
				t.Fatalf("ValidateAttachmentRef accepted=%v, want %v; err=%v", accepted, fixtureCase.Expected.Accepted, err)
			}
			if pattern.MatchString(fixtureCase.Input.Ref) != accepted {
				t.Fatalf("OpenAPI pattern acceptance %v disagrees with ValidateAttachmentRef acceptance %v for %q", pattern.MatchString(fixtureCase.Input.Ref), accepted, fixtureCase.Input.Ref)
			}
			if accepted && fixtureCase.Classification != testcase.MustPass {
				t.Fatalf("case classification %q disagrees with an accepted expectation", fixtureCase.Classification)
			}
			if !accepted && fixtureCase.Classification != testcase.MustFail {
				t.Fatalf("case classification %q disagrees with a rejected expectation", fixtureCase.Classification)
			}
			if err != nil && !strings.Contains(err.Error(), fixtureCase.Expected.ErrorContains) {
				t.Fatalf("validation error %q does not contain %q", err, fixtureCase.Expected.ErrorContains)
			}
		})
	}
}

func TestAttachmentRouteResponseFixtures(t *testing.T) {
	fixtures := loadAttachmentRouteFixtures(t)
	assert.RequireMin(t, fixtures.Responses, 5)
	assert.RequireValid(t, fixtures.Responses)
	for _, fixtureCase := range fixtures.Responses.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			got, err := schema.AttachmentRouteResponseFor(schema.AttachmentRouteOutcome(fixtureCase.Input.Outcome))
			if fixtureCase.Expected.Accepted == false && fixtureCase.Expected.ErrorContains != "" && fixtureCase.Expected.Status == 0 {
				if err == nil {
					t.Fatalf("AttachmentRouteResponseFor accepted outcome %q", fixtureCase.Input.Outcome)
				}
				if !strings.Contains(err.Error(), fixtureCase.Expected.ErrorContains) {
					t.Fatalf("response error %q does not contain %q", err, fixtureCase.Expected.ErrorContains)
				}
				if fixtureCase.Classification != testcase.MustFail {
					t.Fatalf("case classification %q disagrees with a rejected expectation", fixtureCase.Classification)
				}
				return
			}
			if err != nil {
				t.Fatalf("AttachmentRouteResponseFor: %v", err)
			}
			if fixtureCase.Classification != testcase.MustPass {
				t.Fatalf("case classification %q disagrees with an accepted expectation", fixtureCase.Classification)
			}
			if got.Status != fixtureCase.Expected.Status || string(got.MediaType) != fixtureCase.Expected.MediaType || got.CacheControl != fixtureCase.Expected.CacheControl || got.Code != fixtureCase.Expected.Code {
				t.Fatalf("response status=%d media=%q cache=%q code=%q, want status=%d media=%q cache=%q code=%q", got.Status, got.MediaType, got.CacheControl, got.Code, fixtureCase.Expected.Status, fixtureCase.Expected.MediaType, fixtureCase.Expected.CacheControl, fixtureCase.Expected.Code)
			}
			if got.Status == fixtures.Operation.SuccessStatus {
				if got.CacheControl != fixtures.Operation.CacheControl || !attachmentMediaListed(fixtures.Operation.SuccessMediaTypes, string(got.MediaType)) || got.Code != "" {
					t.Fatalf("success response %q does not match the operation fixture", fixtureCase.Name)
				}
			}
			if got.Status == fixtures.Operation.MalformedStatus && (got.Code != fixtures.Operation.MalformedCode || got.MediaType != "" || got.CacheControl != "") {
				t.Fatalf("malformed response carries media or cache, or the wrong code")
			}
			if got.Status == fixtures.Operation.MissingStatus && (got.Code != fixtures.Operation.MissingCode || got.MediaType != "" || got.CacheControl != "") {
				t.Fatalf("missing response carries media or cache, or the wrong code")
			}
		})
	}
}

func TestAttachmentRouteOpenAPI(t *testing.T) {
	fixtures := loadAttachmentRouteFixtures(t)
	op := fixtures.Operation
	if op.Path != schema.AttachmentRoutePath || op.OperationID != schema.AttachmentRouteOperationID || op.CacheControl != schema.AttachmentCacheControl {
		t.Fatalf("operation fixture disagrees with the exported route constants")
	}
	if op.SuccessStatus != schema.AttachmentRouteStatusOK || op.MalformedStatus != schema.AttachmentRouteStatusMalformed || op.MissingStatus != schema.AttachmentRouteStatusNotFound {
		t.Fatalf("operation fixture statuses disagree with the exported route statuses")
	}
	if op.MalformedCode != schema.AttachmentRouteCodeMalformed || op.MissingCode != schema.AttachmentRouteCodeNotFound {
		t.Fatalf("operation fixture codes disagree with the exported route codes")
	}
	spec, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("BuildPeasantLocalAPISpec: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal local API spec: %v", err)
	}
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode local API spec: %v", err)
	}
	methods := document.Paths[op.Path]
	raw, ok := methods[op.Method]
	if !ok {
		t.Fatalf("local API does not declare %s %s", op.Method, op.Path)
	}
	var operation map[string]any
	if err := json.Unmarshal(raw, &operation); err != nil {
		t.Fatalf("decode attachment operation: %v", err)
	}
	if operation["operationId"] != op.OperationID {
		t.Fatalf("operationId = %v, want %q", operation["operationId"], op.OperationID)
	}
	description, _ := operation["description"].(string)
	for _, clause := range op.DescriptionClauses {
		if !strings.Contains(description, clause) {
			t.Fatalf("operation description is missing %q", clause)
		}
	}
	parameters, _ := operation["parameters"].([]any)
	var parameterFound bool
	for _, item := range parameters {
		parameter, _ := item.(map[string]any)
		if parameter["in"] != "path" || parameter["name"] != op.Parameter {
			continue
		}
		parameterFound = true
		schemaMap, _ := parameter["schema"].(map[string]any)
		if parameter["required"] != true || schemaMap["type"] != "string" || schemaMap["pattern"] != op.Pattern {
			t.Fatalf("ref parameter = %#v", parameter)
		}
		if intFromJSON(schemaMap["minLength"]) != op.MinLength || intFromJSON(schemaMap["maxLength"]) != op.MaxLength {
			t.Fatalf("ref parameter length min=%v max=%v, want %d and %d", schemaMap["minLength"], schemaMap["maxLength"], op.MinLength, op.MaxLength)
		}
	}
	if !parameterFound {
		t.Fatalf("local API path %s has no %q path parameter", op.Path, op.Parameter)
	}
	responses, _ := operation["responses"].(map[string]any)
	success, _ := responses[itoa(op.SuccessStatus)].(map[string]any)
	content, _ := success["content"].(map[string]any)
	if len(content) != len(op.SuccessMediaTypes) {
		t.Fatalf("200 content types = %#v, want %v", content, op.SuccessMediaTypes)
	}
	for _, media := range op.SuccessMediaTypes {
		body, _ := content[media].(map[string]any)
		schemaMap, _ := body["schema"].(map[string]any)
		if schemaMap["type"] != "string" || schemaMap["format"] != "binary" || schemaMap["contentMediaType"] != media || schemaMap["contentEncoding"] != "binary" {
			t.Fatalf("%s schema = %#v, want a binary string of that media type", media, schemaMap)
		}
	}
	headers, _ := success["headers"].(map[string]any)
	header, _ := headers[op.CacheHeader].(map[string]any)
	headerSchema, _ := header["schema"].(map[string]any)
	if header["required"] != true || headerSchema["type"] != "string" || headerSchema["const"] != op.CacheControl {
		t.Fatalf("cache header = %#v", header)
	}
	malformed, ok := responses[itoa(op.MalformedStatus)].(map[string]any)
	if !ok {
		t.Fatalf("local API attachment route is missing status %d", op.MalformedStatus)
	}
	missing, ok := responses[itoa(op.MissingStatus)].(map[string]any)
	if !ok {
		t.Fatalf("local API attachment route is missing status %d", op.MissingStatus)
	}
	if !strings.Contains(stringField(malformed, "description"), op.MalformedCode) {
		t.Fatalf("400 description does not name %s: %v", op.MalformedCode, malformed["description"])
	}
	if !strings.Contains(stringField(missing, "description"), op.MissingCode) {
		t.Fatalf("404 description does not name %s: %v", op.MissingCode, missing["description"])
	}
}

func loadAttachmentRouteFixtures(t *testing.T) schema.AttachmentRouteFixtures {
	t.Helper()
	fixtures, err := schema.LoadAttachmentRouteFixtures()
	if err != nil {
		t.Fatalf("load attachment route fixtures: %v", err)
	}
	return fixtures
}

func attachmentMediaListed(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func intFromJSON(value any) int {
	number, ok := value.(float64)
	if !ok {
		return 0
	}
	return int(number)
}

func stringField(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return value
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
