package schema

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// CollectiveTimestampPattern is the lossless UTC RFC3339 wire grammar for
// collective evidence: seconds are required, precision is at most nanoseconds.
const CollectiveTimestampPattern = `^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?Z$`

var collectiveTimestampGrammar = regexp.MustCompile(CollectiveTimestampPattern)

// Inspect raw values before time.Time can silently discard excess precision.
// Traverse only declared fields on this metadata boundary, not additive extras
// or unrelated legacy payloads. Nullable timestamps retain explicit null.
func validateCollectiveRawTimes(data []byte, typ reflect.Type, field string) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeOf(time.Time{}) {
		var text string
		if json.Unmarshal(data, &text) != nil || !collectiveTimestampGrammar.MatchString(text) {
			return observationError("timestamp " + field + " requires UTC RFC3339 with explicit seconds and zero to nine fractional digits; unsupported precision would lose evidence, so supply a lossless nanosecond timestamp ending in Z before retrying")
		}
		value, err := time.Parse(time.RFC3339Nano, text)
		if err != nil || value.IsZero() {
			return observationError("timestamp " + field + " must be a real nonzero UTC calendar instant; correct the source timestamp before retrying")
		}
		return nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if json.Unmarshal(data, &fields) != nil {
			return nil
		}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if f.IsExported() && name != "-" {
				if raw, ok := fields[name]; ok {
					if err := validateCollectiveRawTimes(raw, f.Type, field+"."+name); err != nil {
						return err
					}
				}
			}
		}
	case reflect.Slice, reflect.Array:
		var rows []json.RawMessage
		if json.Unmarshal(data, &rows) != nil {
			return nil
		}
		for _, row := range rows {
			if err := validateCollectiveRawTimes(row, typ.Elem(), field+"[]"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *LocalRemoteContributionAttempt) UnmarshalJSON(data []byte) error {
	type plain LocalRemoteContributionAttempt
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := LocalRemoteContributionAttempt(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}

func (v *VillageShareEvent) UnmarshalJSON(data []byte) error {
	type plain VillageShareEvent
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := VillageShareEvent(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *VillageContributedCollective) UnmarshalJSON(data []byte) error {
	type plain VillageContributedCollective
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := VillageContributedCollective(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *VillageContributedCollectivesResponse) UnmarshalJSON(data []byte) error {
	type plain VillageContributedCollectivesResponse
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := VillageContributedCollectivesResponse(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}

// Preserve required zero/false values without interpreting absence or JSON null
// as those values. Additive object fields remain open except the history union.
// Named alias destinations prevent recursion while nested public types still
// execute their own decoders. No untrusted payload bytes enter error messages.
func decodeCollectiveObject(data []byte, target interface{}) error {
	if err := requireCollectiveFields(data, target); err != nil {
		return err
	}
	if json.Unmarshal(data, target) != nil {
		return observationError("metadata payload has an invalid field type or timestamp")
	}
	return nil
}
func requireCollectiveFields(data []byte, target interface{}) error {
	if err := validateCollectiveRawTimes(data, reflect.TypeOf(target).Elem(), "metadata"); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return observationError("metadata payload must be a JSON object")
	}
	t := reflect.TypeOf(target).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := strings.Split(f.Tag.Get("json"), ",")
		if !f.IsExported() || tag[0] == "-" {
			continue
		}
		optional := false
		for _, part := range tag[1:] {
			if part == "omitempty" {
				optional = true
			}
		}
		if optional {
			continue
		}
		value, exists := fields[tag[0]]
		if !exists {
			return observationError("metadata payload is missing a required " + tag[0] + " field")
		}
		if f.Type.Kind() != reflect.Pointer && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return observationError("required metadata field " + tag[0] + " cannot be null")
		}
	}
	return nil
}

func (v *CollectiveObservationScope) UnmarshalJSON(data []byte) error {
	type plain CollectiveObservationScope
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveObservationScope(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *OperationIssue) UnmarshalJSON(data []byte) error {
	type plain OperationIssue
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := OperationIssue(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *CollectiveObservation) UnmarshalJSON(data []byte) error {
	type plain CollectiveObservation
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveObservation(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *LocalConfirmedContributionReceipt) UnmarshalJSON(data []byte) error {
	type plain LocalConfirmedContributionReceipt
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := LocalConfirmedContributionReceipt(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *LocalOperationDiagnostic) UnmarshalJSON(data []byte) error {
	type plain LocalOperationDiagnostic
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := LocalOperationDiagnostic(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *PullAssociationMetadataResult) UnmarshalJSON(data []byte) error {
	type plain PullAssociationMetadataResult
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := PullAssociationMetadataResult(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *CollectiveFetchSourceResult) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchSourceResult
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	*v = CollectiveFetchSourceResult(value)
	return nil
}
func (v *CollectiveFetchAssociationScopeResult) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchAssociationScopeResult
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	*v = CollectiveFetchAssociationScopeResult(value)
	return nil
}
func (v *CollectiveFetchAssociationResults) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchAssociationResults
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	*v = CollectiveFetchAssociationResults(value)
	return nil
}
func (v *CollectiveFetchDomainResult) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchDomainResult
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveFetchDomainResult(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *CollectiveFetchReport) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchReport
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveFetchReport(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *CollectiveFetchDomainSummary) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchDomainSummary
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveFetchDomainSummary(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *CollectiveFetchSummary) UnmarshalJSON(data []byte) error {
	type plain CollectiveFetchSummary
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveFetchSummary(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *CollectiveObservationResponse) UnmarshalJSON(data []byte) error {
	type plain CollectiveObservationResponse
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := CollectiveObservationResponse(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
func (v *VillageOrganizationCollective) UnmarshalJSON(data []byte) error {
	type plain VillageOrganizationCollective
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	if !transcriptIDUUID.MatchString(string(value.ID)) || value.MemberCount < 0 || value.TranscriptCount < 0 {
		return observationError("organization collective has invalid identity or negative counts")
	}
	*v = VillageOrganizationCollective(value)
	return nil
}
func (v *VillageOrganizationCollectivesResponse) UnmarshalJSON(data []byte) error {
	type plain VillageOrganizationCollectivesResponse
	var value plain
	if err := decodeCollectiveObject(data, &value); err != nil {
		return err
	}
	result := VillageOrganizationCollectivesResponse(value)
	if err := result.Validate(); err != nil {
		return err
	}
	*v = result
	return nil
}
