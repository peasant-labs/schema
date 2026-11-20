package openapi

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// villageOperationExpectation is the declared shape of one Village operation
// as a fixture states it. Statuses, media types, headers and parameters are
// compared as exact sets: the document may not carry one the fixture does not
// list, and may not miss one it lists.
type villageOperationExpectation struct {
	OperationID string `yaml:"operation_id"`
	Tag         string `yaml:"tag"`
	// Servers is the path item's servers override; empty means none.
	Servers    []string                      `yaml:"servers"`
	Parameters []villageParameterExpectation `yaml:"parameters"`
	// RequestBody is nil when the operation declares no request body.
	RequestBody        *villageRequestBodyExpectation        `yaml:"request_body"`
	Responses          map[string]villageResponseExpectation `yaml:"responses"`
	DescriptionAnchors []string                              `yaml:"description_anchors"`
}

type villageParameterExpectation struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	// Schema, when set, is the parameter schema's descriptor (see
	// villageSchemaDescriptor).
	Schema string `yaml:"schema"`
}

type villageRequestBodyExpectation struct {
	Required bool `yaml:"required"`
	// Media maps each media type to its schema descriptor.
	Media map[string]string `yaml:"media"`
}

type villageResponseExpectation struct {
	// Media maps each media type to its schema descriptor; empty means the
	// response has no body.
	Media map[string]string `yaml:"media"`
	// Headers maps each header name to whether it is required.
	Headers map[string]bool `yaml:"headers"`
	// DescriptionAnchors are phrases the response description must contain.
	DescriptionAnchors []string `yaml:"description_anchors"`
}

// villageDocument marshals a built document to its JSON form, which is what
// every expectation is compared with.
func villageDocument(t *testing.T, spec interface{ MarshalJSON() ([]byte, error) }) map[string]any {
	t.Helper()
	raw, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal Village document: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode Village document: %v", err)
	}
	return document
}

// checkVillageOperationExpectation returns every way the document's operation
// at method and path differs from the expectation. method is lower case.
func checkVillageOperationExpectation(document map[string]any, method, path string, want villageOperationExpectation) []string {
	var problems []string
	fail := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}
	paths, _ := document["paths"].(map[string]any)
	item, ok := paths[path].(map[string]any)
	if !ok {
		return []string{fmt.Sprintf("the document has no path %s", path)}
	}
	operation, ok := item[method].(map[string]any)
	if !ok {
		return []string{fmt.Sprintf("the document has no %s operation on %s", method, path)}
	}

	if got, _ := operation["operationId"].(string); got != want.OperationID {
		fail("operationId = %q, want %q", got, want.OperationID)
	}
	tags, _ := operation["tags"].([]any)
	if len(tags) != 1 || tags[0] != want.Tag {
		fail("tags = %v, want [%s]", tags, want.Tag)
	}

	var servers []string
	if raw, ok := item["servers"].([]any); ok {
		for _, server := range raw {
			entry, _ := server.(map[string]any)
			url, _ := entry["url"].(string)
			servers = append(servers, url)
		}
	}
	if strings.Join(servers, ",") != strings.Join(want.Servers, ",") {
		fail("path item servers = %v, want %v", servers, want.Servers)
	}

	gotParameters := map[string]villageParameterExpectation{}
	if raw, ok := operation["parameters"].([]any); ok {
		for _, parameter := range raw {
			entry, _ := parameter.(map[string]any)
			name, _ := entry["name"].(string)
			in, _ := entry["in"].(string)
			required, _ := entry["required"].(bool)
			schemaMap, _ := entry["schema"].(map[string]any)
			gotParameters[in+" "+name] = villageParameterExpectation{Name: name, In: in, Required: required, Schema: villageSchemaDescriptor(schemaMap)}
		}
	}
	wantParameters := map[string]bool{}
	for _, parameter := range want.Parameters {
		key := parameter.In + " " + parameter.Name
		wantParameters[key] = true
		got, ok := gotParameters[key]
		switch {
		case !ok:
			fail("parameter %s is not declared", key)
		case got.Required != parameter.Required:
			fail("parameter %s required = %t, want %t", key, got.Required, parameter.Required)
		case parameter.Schema != "" && got.Schema != parameter.Schema:
			fail("parameter %s schema = %s, want %s", key, got.Schema, parameter.Schema)
		}
	}
	for key := range gotParameters {
		if !wantParameters[key] {
			fail("parameter %s is declared but the fixture does not list it", key)
		}
	}

	body, hasBody := operation["requestBody"].(map[string]any)
	switch {
	case want.RequestBody == nil && hasBody:
		fail("the operation declares a request body the fixture does not list")
	case want.RequestBody != nil && !hasBody:
		fail("the operation declares no request body; the fixture lists one")
	case want.RequestBody != nil:
		required, _ := body["required"].(bool)
		if required != want.RequestBody.Required {
			fail("request body required = %t, want %t", required, want.RequestBody.Required)
		}
		content, _ := body["content"].(map[string]any)
		problems = append(problems, compareVillageMedia("request body", content, want.RequestBody.Media)...)
	}

	responses, _ := operation["responses"].(map[string]any)
	for status := range responses {
		if _, ok := want.Responses[status]; !ok {
			fail("status %s is declared but the fixture does not list it", status)
		}
	}
	for status, wantResponse := range want.Responses {
		response, ok := responses[status].(map[string]any)
		if !ok {
			fail("status %s is not declared", status)
			continue
		}
		content, _ := response["content"].(map[string]any)
		problems = append(problems, compareVillageMedia("status "+status, content, wantResponse.Media)...)
		headers, _ := response["headers"].(map[string]any)
		for name, header := range headers {
			wantRequired, ok := wantResponse.Headers[name]
			if !ok {
				fail("status %s declares header %s, which the fixture does not list", status, name)
				continue
			}
			entry, _ := header.(map[string]any)
			required, _ := entry["required"].(bool)
			if required != wantRequired {
				fail("status %s header %s required = %t, want %t", status, name, required, wantRequired)
			}
		}
		for name := range wantResponse.Headers {
			if _, ok := headers[name]; !ok {
				fail("status %s does not declare header %s", status, name)
			}
		}
		description, _ := response["description"].(string)
		for _, anchor := range wantResponse.DescriptionAnchors {
			if !strings.Contains(description, anchor) {
				fail("status %s description %q is missing anchor %q", status, description, anchor)
			}
		}
	}

	description, _ := operation["description"].(string)
	for _, anchor := range want.DescriptionAnchors {
		if !strings.Contains(description, anchor) {
			fail("operation description %q is missing anchor %q", description, anchor)
		}
	}
	sort.Strings(problems)
	return problems
}

func compareVillageMedia(where string, content map[string]any, want map[string]string) []string {
	var problems []string
	for mediaType, raw := range content {
		wantSchema, ok := want[mediaType]
		if !ok {
			problems = append(problems, fmt.Sprintf("%s declares media type %s, which the fixture does not list", where, mediaType))
			continue
		}
		medium, _ := raw.(map[string]any)
		schemaMap, _ := medium["schema"].(map[string]any)
		if got := villageSchemaDescriptor(schemaMap); got != wantSchema {
			problems = append(problems, fmt.Sprintf("%s %s schema = %s, want %s", where, mediaType, got, wantSchema))
		}
	}
	for mediaType := range want {
		if _, ok := content[mediaType]; !ok {
			problems = append(problems, fmt.Sprintf("%s does not declare media type %s", where, mediaType))
		}
	}
	return problems
}

// villageSchemaDescriptor names a schema compactly: the component name of a
// $ref, []Name for an array of a component ([]Name|null when the array admits
// null), the type of a schema that carries only a type, and its JSON otherwise.
func villageSchemaDescriptor(schema map[string]any) string {
	if schema == nil {
		return ""
	}
	if _, ok := schema["description"]; ok {
		// A description is prose, not shape.
		trimmed := make(map[string]any, len(schema))
		for key, value := range schema {
			if key != "description" {
				trimmed[key] = value
			}
		}
		schema = trimmed
	}
	if ref, ok := schema["$ref"].(string); ok && len(schema) == 1 {
		return strings.TrimPrefix(ref, "#/components/schemas/")
	}
	if items, ok := schema["items"].(map[string]any); ok && len(schema) == 2 {
		switch kind := schema["type"].(type) {
		case string:
			if kind == "array" {
				return "[]" + villageSchemaDescriptor(items)
			}
		case []any:
			if len(kind) == 2 && ((kind[0] == "null" && kind[1] == "array") || (kind[0] == "array" && kind[1] == "null")) {
				return "[]" + villageSchemaDescriptor(items) + "|null"
			}
		}
	}
	if kind, ok := schema["type"].(string); ok && len(schema) == 1 {
		return kind
	}
	raw, _ := json.Marshal(schema)
	return string(raw)
}
