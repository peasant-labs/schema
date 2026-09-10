package schema

import (
	"encoding/json"
	"reflect"
)

// UnmarshalJSON preserves required member-page presence before Go zero values
// can turn a missing or null total into a measured zero.
func (p *LocalHelperMembersPayload) UnmarshalJSON(raw []byte) error {
	type wire LocalHelperMembersPayload
	if err := validateWireShape(raw, reflect.TypeFor[wire](), "LocalHelperMembersPayload"); err != nil {
		return err
	}
	var value wire
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	decoded := LocalHelperMembersPayload(value)
	if err := decoded.Validate(); err != nil {
		return err
	}
	*p = decoded
	return nil
}

func (p *VillageHelperMembersPayload) UnmarshalJSON(raw []byte) error {
	type wire VillageHelperMembersPayload
	if err := validateWireShape(raw, reflect.TypeFor[wire](), "VillageHelperMembersPayload"); err != nil {
		return err
	}
	var value wire
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	decoded := VillageHelperMembersPayload(value)
	if err := decoded.Validate(); err != nil {
		return err
	}
	*p = decoded
	return nil
}
