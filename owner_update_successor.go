package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// OwnerUpdateLicenseIntent preserves omission separately from explicit null.
// Omission preserves, null requests clear, and a canonical value replaces.
type OwnerUpdateLicenseIntent struct {
	set   bool
	value *License
}

func NewOwnerUpdateLicenseIntent(value *License) (OwnerUpdateLicenseIntent, error) {
	if value == nil {
		return OwnerUpdateLicenseIntent{set: true}, nil
	}
	if !value.IsValid() {
		return OwnerUpdateLicenseIntent{}, publicationError("owner update license", fmt.Sprintf("license %q is not canonical", *value), "use null to request clear or one canonical license")
	}
	copy := *value
	return OwnerUpdateLicenseIntent{set: true, value: &copy}, nil
}
func (v OwnerUpdateLicenseIntent) IsSet() bool { return v.set }
func (v OwnerUpdateLicenseIntent) Value() (License, bool) {
	if v.value == nil {
		return "", false
	}
	return *v.value, true
}
func (v OwnerUpdateLicenseIntent) MarshalJSON() ([]byte, error) {
	if !v.set {
		return nil, publicationError("owner update license", "an omitted intent was marshaled directly", "marshal OwnerTranscriptUpdateRequest so omission is preserved")
	}
	if v.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*v.value)
}
func (v *OwnerUpdateLicenseIntent) UnmarshalJSON(data []byte) error {
	v.set = true
	v.value = nil
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	var license License
	if err := json.Unmarshal(data, &license); err != nil {
		return publicationError("owner update license", err.Error(), "send null or a canonical license")
	}
	if !license.IsValid() {
		return publicationError("owner update license", fmt.Sprintf("license %q is not canonical", license), "send null or a canonical license")
	}
	v.value = &license
	return nil
}
func (OwnerUpdateLicenseIntent) JSONSchema() (jsonschema.Schema, error) {
	license, err := (License("")).JSONSchema()
	if err != nil {
		return jsonschema.Schema{}, err
	}
	null := jsonschema.Schema{}
	null.AddType(jsonschema.Null)
	s := jsonschema.Schema{}
	s.WithAnyOf(license.ToSchemaOrBool(), null.ToSchemaOrBool())
	s.WithDescription("Omitted preserves, null requests clear, and a canonical license replaces")
	return s, nil
}

// OwnerTranscriptUpdateRequest is the successor owner update request. Every
// field is optional; empty text and tags clear, while license null requests clear.
type OwnerTranscriptUpdateRequest struct {
	Title       *string                     `json:"title,omitempty" maxLength:"500" nullable:"false"`
	Description *string                     `json:"description,omitempty" nullable:"false"`
	Tags        *[]string                   `json:"tags,omitempty" nullable:"false"`
	License     OwnerUpdateLicenseIntent    `json:"license,omitempty"`
	Visibility  *TranscriptUpdateVisibility `json:"visibility,omitempty" nullable:"false"`
}

func DecodeOwnerTranscriptUpdateRequest(data []byte) (OwnerTranscriptUpdateRequest, error) {
	var request OwnerTranscriptUpdateRequest
	if err := decodeStrictDocument(data, &request); err != nil {
		return OwnerTranscriptUpdateRequest{}, err
	}
	return request, nil
}
func (request OwnerTranscriptUpdateRequest) MarshalJSON() ([]byte, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	type wire struct {
		Title       *string                     `json:"title,omitempty"`
		Description *string                     `json:"description,omitempty"`
		Tags        *[]string                   `json:"tags,omitempty"`
		License     json.RawMessage             `json:"license,omitempty"`
		Visibility  *TranscriptUpdateVisibility `json:"visibility,omitempty"`
	}
	out := wire{Title: request.Title, Description: request.Description, Visibility: request.Visibility}
	if request.Tags != nil {
		tags := *request.Tags
		if tags == nil {
			tags = []string{}
		}
		out.Tags = &tags
	}
	if request.License.IsSet() {
		raw, err := request.License.MarshalJSON()
		if err != nil {
			return nil, err
		}
		out.License = raw
	}
	return json.Marshal(out)
}
func (request *OwnerTranscriptUpdateRequest) UnmarshalJSON(data []byte) error {
	raw, err := decodeStrictObject(data, ownerUpdateRequestFields, "owner update request")
	if err != nil {
		return err
	}
	*request = OwnerTranscriptUpdateRequest{}
	if err = decodeOptionalNonNullField(raw, "title", &request.Title); err != nil {
		return err
	}
	if err = decodeOptionalNonNullField(raw, "description", &request.Description); err != nil {
		return err
	}
	if err = decodeOptionalNonNullField(raw, "tags", &request.Tags); err != nil {
		return err
	}
	if err = decodeOptionalNonNullField(raw, "visibility", &request.Visibility); err != nil {
		return err
	}
	if value, ok := raw["license"]; ok {
		if err = request.License.UnmarshalJSON(value); err != nil {
			return err
		}
	}
	return request.Validate()
}
func (request OwnerTranscriptUpdateRequest) Validate() error {
	if request.Title != nil && utf8.RuneCountInString(*request.Title) > TranscriptUpdateTitleMaxLength {
		return publicationError("owner update title", fmt.Sprintf("title has %d characters, exceeding %d", utf8.RuneCountInString(*request.Title), TranscriptUpdateTitleMaxLength), "shorten the title")
	}
	if request.Visibility != nil && !request.Visibility.IsValid() {
		return publicationError("owner update visibility", fmt.Sprintf("visibility %q is not accepted", *request.Visibility), "use private or public")
	}
	if value, ok := request.License.Value(); request.License.IsSet() && ok && !value.IsValid() {
		return publicationError("owner update license", "replacement license is invalid", "use null or a canonical license")
	}
	if request.Tags != nil {
		seen := map[string]struct{}{}
		for i, tag := range *request.Tags {
			if tag == "" || tag != strings.TrimSpace(tag) {
				return publicationError("owner update tags", fmt.Sprintf("tags[%d]=%q is empty or untrimmed", i, tag), "send distinct nonempty trimmed tags or [] to clear")
			}
			if _, ok := seen[tag]; ok {
				return publicationError("owner update tags", fmt.Sprintf("tag %q is duplicated", tag), "send each tag once")
			}
			seen[tag] = struct{}{}
		}
	}
	return nil
}

// OwnerTranscriptUpdateResponse is the complete authoritative editable state.
type OwnerTranscriptUpdateResponse struct {
	TranscriptID  TranscriptID               `json:"transcriptId"`
	TranscriptURL string                     `json:"transcriptUrl"`
	Title         *string                    `json:"title"`
	Description   *string                    `json:"description"`
	Tags          []string                   `json:"tags" nullable:"false"`
	License       *License                   `json:"license"`
	Visibility    TranscriptUpdateVisibility `json:"visibility"`
	UpdatedAt     int64                      `json:"updatedAt"`
}

func (response OwnerTranscriptUpdateResponse) Validate() error {
	if _, err := NewTranscriptID(response.TranscriptID.String()); err != nil {
		return err
	}
	if err := validatePublicationURL(response.TranscriptURL, response.TranscriptID); err != nil {
		return err
	}
	if response.Tags == nil {
		return publicationError("owner update response tags", "tags is null", "return a complete non-null array")
	}
	tags := response.Tags
	request := OwnerTranscriptUpdateRequest{Tags: &tags}
	if err := request.Validate(); err != nil {
		return err
	}
	if response.License != nil && !response.License.IsValid() {
		return publicationError("owner update response license", "license is invalid", "return null or a canonical license")
	}
	if !response.Visibility.IsValid() {
		return publicationError("owner update response visibility", "visibility is invalid", "return private or public")
	}
	if response.UpdatedAt <= 0 {
		return publicationError("owner update response timestamp", "updatedAt is not positive", "return authoritative positive Unix milliseconds")
	}
	return nil
}
func (response *OwnerTranscriptUpdateResponse) UnmarshalJSON(data []byte) error {
	raw, err := decodeStrictObject(data, ownerUpdateResponseFields, "owner update response")
	if err != nil {
		return err
	}
	for _, field := range []string{"transcriptId", "transcriptUrl", "title", "description", "tags", "license", "visibility", "updatedAt"} {
		value, ok := raw[field]
		if !ok {
			return publicationError("owner update response", fmt.Sprintf("required field %q is missing", field), "return the complete authoritative response")
		}
		if field != "title" && field != "description" && field != "license" && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return publicationError("owner update response", fmt.Sprintf("required field %q is null", field), "return a non-null value")
		}
	}
	type wire OwnerTranscriptUpdateResponse
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*response = OwnerTranscriptUpdateResponse(decoded)
	return response.Validate()
}

var ownerUpdateRequestFields = map[string]struct{}{"title": {}, "description": {}, "tags": {}, "license": {}, "visibility": {}}
var ownerUpdateResponseFields = map[string]struct{}{"transcriptId": {}, "transcriptUrl": {}, "title": {}, "description": {}, "tags": {}, "license": {}, "visibility": {}, "updatedAt": {}}

func decodeStrictDocument(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return publicationError("JSON document", "more than one JSON value or invalid trailing data", "send exactly one object")
	}
	return nil
}
func decodeStrictObject(data []byte, allowed map[string]struct{}, what string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return nil, publicationError(what, err.Error(), "send exactly one JSON object")
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return nil, publicationError(what, "value is not an object", "send a JSON object")
	}
	result := map[string]json.RawMessage{}
	for decoder.More() {
		nameToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name := nameToken.(string)
		if _, ok := allowed[name]; !ok {
			return nil, publicationError(what, fmt.Sprintf("field %q is unknown", name), "remove the unknown field")
		}
		if _, ok := result[name]; ok {
			return nil, publicationError(what, fmt.Sprintf("field %q appears more than once", name), "send each field once")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		result[name] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, publicationError(what, "more than one JSON value", "send exactly one object")
	}
	return result, nil
}
func decodeOptionalNonNullField[T any](raw map[string]json.RawMessage, name string, destination **T) error {
	value, ok := raw[name]
	if !ok {
		return nil
	}
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return publicationError("owner update request", fmt.Sprintf("field %q is null", name), "omit to preserve or send an explicit clearing value")
	}
	var decoded T
	if err := json.Unmarshal(value, &decoded); err != nil {
		return err
	}
	*destination = &decoded
	return nil
}
