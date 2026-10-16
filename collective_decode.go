package schema

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
)

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
