package schema

import (
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// VillageTranscriptIntegrity is served only for a public transcript that scry
// flagged; its presence is the flag. FlaggedTurns are zero-based positions in
// the top-level SessionDetailPayload.turns array of the content whose bare hash
// is the same response's transcript.content_hash (never TurnDetail.index
// values, never positions in earlierHistory or child sessions); empty when the
// flag is sticky and the current copy has no flagged turn.
type VillageTranscriptIntegrity struct {
	FlaggedTurns []int `json:"flagged_turns" nullable:"false" description:"Zero-based positions in the top-level sessionDetail.turns array of the content whose bare hash is transcript.content_hash; never TurnDetail.index values and never positions in earlierHistory or child sessions. Presence of integrity is the flag."`
}

// PrepareJSONSchema bounds every flagged_turns item, which struct tags cannot
// reach.
func (VillageTranscriptIntegrity) PrepareJSONSchema(s *jsonschema.Schema) error {
	property, ok := s.Properties["flagged_turns"]
	if !ok || property.TypeObject == nil {
		return fmt.Errorf("village transcript integrity schema has no flagged_turns property to bound")
	}
	item := jsonschema.Schema{}
	item.AddType(jsonschema.Integer)
	item.WithMinimum(0)
	item.WithMaximum(float64(maxSafeJSONInteger))
	property.TypeObject.WithItems(*(&jsonschema.Items{}).WithSchemaOrBool(item.ToSchemaOrBool()))
	s.Properties["flagged_turns"] = property
	return nil
}

var (
	villageTranscriptIntegrityFields = fieldSet("flagged_turns")
	villageIntegrityCheckFields      = fieldSet("result", "remaining_today")
)

// Validate refuses a null, unsorted, repeated or out-of-range position. The
// turn count is not on this body; village holds the content and bounds the
// positions against it.
func (i VillageTranscriptIntegrity) Validate() error {
	return validateTurnIndexes("flagged_turns", i.FlaggedTurns)
}

// UnmarshalJSON refuses unknown members and a null flagged_turns.
func (i *VillageTranscriptIntegrity) UnmarshalJSON(data []byte) error {
	type wire VillageTranscriptIntegrity
	var decoded wire
	if err := decodeRequiredStrict(data, villageTranscriptIntegrityFields, nil, &decoded, "village transcript integrity"); err != nil {
		return err
	}
	*i = VillageTranscriptIntegrity(decoded)
	return nil
}

// VillageIntegrityCheckResult is what an owner's check request found.
type VillageIntegrityCheckResult string

const (
	VillageIntegrityCheckResultFlagged    VillageIntegrityCheckResult = "flagged"
	VillageIntegrityCheckResultNotFlagged VillageIntegrityCheckResult = "not_flagged"
	VillageIntegrityCheckResultUnchecked  VillageIntegrityCheckResult = "unchecked" // scry declined; the allowance is not spent
)

// AllVillageIntegrityCheckResults is the closed result set.
var AllVillageIntegrityCheckResults = []VillageIntegrityCheckResult{
	VillageIntegrityCheckResultFlagged,
	VillageIntegrityCheckResultNotFlagged,
	VillageIntegrityCheckResultUnchecked,
}

func (r VillageIntegrityCheckResult) IsValid() bool {
	for _, known := range AllVillageIntegrityCheckResults {
		if r == known {
			return true
		}
	}
	return false
}
func (r VillageIntegrityCheckResult) String() string { return string(r) }
func (VillageIntegrityCheckResult) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Integrity Check Result", "What an owner's integrity check found; unchecked means scry declined and the daily allowance is not spent", AllVillageIntegrityCheckResults), nil
}

// VillageIntegrityCheckResponse is the answer to an owner's check request. It
// never lists positions: they ride on the transcript metadata read.
type VillageIntegrityCheckResponse struct {
	Result         VillageIntegrityCheckResult `json:"result"`
	RemainingToday int                         `json:"remaining_today" minimum:"0" maximum:"9007199254740991"`
}

// Validate refuses a result outside the closed set and a negative or unsafe
// remaining count.
func (r VillageIntegrityCheckResponse) Validate() error {
	const what = "village integrity check response"
	if !r.Result.IsValid() {
		return scryError(what, fmt.Sprintf("result %q is outside the closed set %v", r.Result, AllVillageIntegrityCheckResults), "send flagged, not_flagged or unchecked")
	}
	if r.RemainingToday < 0 || int64(r.RemainingToday) > maxSafeJSONInteger {
		return scryError(what, fmt.Sprintf("remaining_today %d is outside 0..%d", r.RemainingToday, maxSafeJSONInteger), "send the checks left today")
	}
	return nil
}

// UnmarshalJSON refuses unknown and missing members.
func (r *VillageIntegrityCheckResponse) UnmarshalJSON(data []byte) error {
	type wire VillageIntegrityCheckResponse
	var decoded wire
	if err := decodeRequiredStrict(data, villageIntegrityCheckFields, nil, &decoded, "village integrity check response"); err != nil {
		return err
	}
	*r = VillageIntegrityCheckResponse(decoded)
	return nil
}
