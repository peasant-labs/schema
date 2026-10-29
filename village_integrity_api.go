package schema

import (
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// VillageTranscriptIntegrity is served when the transcript has an open flag, to
// every caller allowed to read it; presence is the flag. FlaggedTurns are
// zero-based positions in the top-level SessionDetailPayload.turns array of the
// content whose bare hash is the same response's transcript.content_hash (never TurnDetail.index
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
