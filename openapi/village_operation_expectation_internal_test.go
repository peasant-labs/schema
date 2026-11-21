package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/swaggest/openapi-go/openapi31"
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

// checkVillageOperationExpectation returns every way the document's operation
// at method and path differs from the expectation. method is an HTTP method
// in any case.
func checkVillageOperationExpectation(spec *openapi31.Spec, method, path string, want villageOperationExpectation) []string {
	var problems []string
	fail := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}
	item, ok := declaredPathItem(spec, path)
	if !ok {
		return []string{fmt.Sprintf("the document has no path %s", path)}
	}
	operation := pathOperation(item, strings.ToUpper(method))
	if operation == nil {
		return []string{fmt.Sprintf("the document has no %s operation on %s", method, path)}
	}

	if got := derefString(operation.ID); got != want.OperationID {
		fail("operationId = %q, want %q", got, want.OperationID)
	}
	if len(operation.Tags) != 1 || operation.Tags[0] != want.Tag {
		fail("tags = %v, want [%s]", operation.Tags, want.Tag)
	}

	var servers []string
	for _, server := range item.Servers {
		servers = append(servers, server.URL)
	}
	if strings.Join(servers, ",") != strings.Join(want.Servers, ",") {
		fail("path item servers = %v, want %v", servers, want.Servers)
	}

	gotParameters := map[string]villageParameterExpectation{}
	for _, reference := range operation.Parameters {
		parameter := reference.Parameter
		if parameter == nil {
			fail("a parameter is a reference; the expectation reads inline parameters only")
			continue
		}
		in := string(parameter.In)
		gotParameters[in+" "+parameter.Name] = villageParameterExpectation{
			Name: parameter.Name, In: in, Required: derefBool(parameter.Required),
			Schema: villageSchemaDescriptor(parameter.Schema),
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

	var body *openapi31.RequestBody
	if operation.RequestBody != nil {
		body = operation.RequestBody.RequestBody
	}
	switch {
	case want.RequestBody == nil && body != nil:
		fail("the operation declares a request body the fixture does not list")
	case want.RequestBody != nil && body == nil:
		fail("the operation declares no request body; the fixture lists one")
	case want.RequestBody != nil:
		if got := derefBool(body.Required); got != want.RequestBody.Required {
			fail("request body required = %t, want %t", got, want.RequestBody.Required)
		}
		problems = append(problems, compareVillageMedia("request body", body.Content, want.RequestBody.Media)...)
	}

	var responses map[string]openapi31.ResponseOrReference
	if operation.Responses != nil {
		responses = operation.Responses.MapOfResponseOrReferenceValues
	}
	for status := range responses {
		if _, ok := want.Responses[status]; !ok {
			fail("status %s is declared but the fixture does not list it", status)
		}
	}
	for status, wantResponse := range want.Responses {
		reference, ok := responses[status]
		if !ok || reference.Response == nil {
			fail("status %s is not declared inline", status)
			continue
		}
		response := reference.Response
		problems = append(problems, compareVillageMedia("status "+status, response.Content, wantResponse.Media)...)
		for name, header := range response.Headers {
			wantRequired, ok := wantResponse.Headers[name]
			if !ok {
				fail("status %s declares header %s, which the fixture does not list", status, name)
				continue
			}
			required := header.Header != nil && derefBool(header.Header.Required)
			if required != wantRequired {
				fail("status %s header %s required = %t, want %t", status, name, required, wantRequired)
			}
		}
		for name := range wantResponse.Headers {
			if _, ok := response.Headers[name]; !ok {
				fail("status %s does not declare header %s", status, name)
			}
		}
		for _, anchor := range wantResponse.DescriptionAnchors {
			if !strings.Contains(response.Description, anchor) {
				fail("status %s description %q is missing anchor %q", status, response.Description, anchor)
			}
		}
	}

	description := derefString(operation.Description)
	for _, anchor := range want.DescriptionAnchors {
		if !strings.Contains(description, anchor) {
			fail("operation description %q is missing anchor %q", description, anchor)
		}
	}
	sort.Strings(problems)
	return problems
}

func compareVillageMedia(where string, content map[string]openapi31.MediaType, want map[string]string) []string {
	var problems []string
	for mediaType, medium := range content {
		wantSchema, ok := want[mediaType]
		if !ok {
			problems = append(problems, fmt.Sprintf("%s declares media type %s, which the fixture does not list", where, mediaType))
			continue
		}
		if got := villageSchemaDescriptor(medium.Schema); got != wantSchema {
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

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefBool(value *bool) bool {
	return value != nil && *value
}

// villageSchemaShape is the part of a schema a descriptor names. A schema with
// any other member is described by its JSON instead.
type villageSchemaShape struct {
	Ref         string             `json:"$ref"`
	Type        villageSchemaTypes `json:"type"`
	Items       json.RawMessage    `json:"items"`
	Minimum     *json.Number       `json:"minimum"`
	Maximum     *json.Number       `json:"maximum"`
	Description string             `json:"description"`
}

// villageSchemaTypes is a schema's type, written as one name or a list.
type villageSchemaTypes []string

func (types *villageSchemaTypes) UnmarshalJSON(raw []byte) error {
	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		*types = villageSchemaTypes{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return err
	}
	sort.Strings(many)
	*types = many
	return nil
}

// villageSchemaDescriptor names a schema compactly: the component name of a
// $ref, []Name for an array of a component ([]Name|null when the array admits
// null), the type of a schema that carries only a type, that type followed by
// MIN..MAX when it carries both bounds, and its JSON otherwise. A description
// is prose, not shape, and is ignored.
func villageSchemaDescriptor(schema map[string]interface{}) string {
	if schema == nil {
		return ""
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return fmt.Sprintf("unencodable schema: %v", err)
	}
	return villageRawSchemaDescriptor(raw)
}

func villageRawSchemaDescriptor(raw []byte) string {
	var shape villageSchemaShape
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shape); err != nil {
		return string(raw)
	}
	types := strings.Join(shape.Type, ",")
	switch {
	case shape.Ref != "" && types == "" && shape.Items == nil:
		return strings.TrimPrefix(shape.Ref, "#/components/schemas/")
	case shape.Ref == "" && shape.Items != nil && types == "array":
		return "[]" + villageRawSchemaDescriptor(shape.Items)
	case shape.Ref == "" && shape.Items != nil && types == "array,null":
		return "[]" + villageRawSchemaDescriptor(shape.Items) + "|null"
	case shape.Ref == "" && shape.Items == nil && len(shape.Type) == 1 && shape.Minimum == nil && shape.Maximum == nil:
		return shape.Type[0]
	case shape.Ref == "" && shape.Items == nil && len(shape.Type) == 1 && shape.Minimum != nil && shape.Maximum != nil:
		return shape.Type[0] + " " + shape.Minimum.String() + ".." + shape.Maximum.String()
	}
	return string(raw)
}
