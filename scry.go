package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// ScryContractVersion is the version of the scry wire contract. Village and
// scry pin the same schema release, so both sides carry the same value.
// Village compares it with ScryHealthResponse.ContractVersion before it calls
// POST /v1/check, and the check request carries it so scry can refuse a
// request built for another contract.
const ScryContractVersion = "0.1.0"

// scrySemVerCorePattern is MAJOR.MINOR.PATCH without a prerelease, build
// metadata or a leading v. It is the single source for the struct tags below
// and for Validate.
const scrySemVerCorePattern = `^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`

var scrySemVerCoreRE = regexp.MustCompile(scrySemVerCorePattern)

const scryFlaggedTurnsDescription = "Zero-based positions in the top-level sessionDetail.turns array; never TurnDetail.index values and never positions in earlierHistory or child sessions. Flagged means verdict checked with at least one position."

// ScryVerdict is whether scry examined the transcript.
type ScryVerdict string

const (
	ScryVerdictChecked   ScryVerdict = "checked"
	ScryVerdictUnchecked ScryVerdict = "unchecked" // scry declined: unsupported content or harness
)

// AllScryVerdicts is the closed verdict set. A transcript is flagged when its
// verdict is checked and at least one turn position is listed; "flagged" is
// never a verdict.
var AllScryVerdicts = []ScryVerdict{ScryVerdictChecked, ScryVerdictUnchecked}

func (v ScryVerdict) IsValid() bool  { return v == ScryVerdictChecked || v == ScryVerdictUnchecked }
func (v ScryVerdict) String() string { return string(v) }
func (ScryVerdict) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Scry Verdict", "Whether scry examined the transcript; unchecked means scry declined unsupported content or harness", AllScryVerdicts), nil
}

// ScryCheckRequest is the body village posts to scry's POST /v1/check.
// Village and scry pin the same schema release. Scry stores and logs no
// content from it.
type ScryCheckRequest struct {
	ContractVersion string                       `json:"contractVersion" minLength:"1"`
	TranscriptID    TranscriptID                 `json:"transcriptId"`
	SessionDetail   SessionDetailPayload         `json:"sessionDetail"`
	PublishMetadata *AuthoritativePublishRequest `json:"publishMetadata,omitempty"`
	Record          ScryTranscriptRecord         `json:"record"`
}

// ScryTranscriptRecord holds only facts village itself holds. Everything the
// publisher reported (times, stats, model, commits, client) is read from
// PublishMetadata and never copied here.
type ScryTranscriptRecord struct {
	PublishedAtMs int64 `json:"publishedAtMs" minimum:"1" maximum:"9007199254740991"`
	UpdatedAtMs   int64 `json:"updatedAtMs" minimum:"1" maximum:"9007199254740991"`
}

// ScryCheckResponse is the verdict scry returns. Scry stores and logs no
// content, and the response never names a rule. Flagged is derived, never
// sent: Verdict == checked && len(FlaggedTurns) > 0.
type ScryCheckResponse struct {
	ScryVersion string      `json:"scryVersion" pattern:"^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)$"`
	Verdict     ScryVerdict `json:"verdict"`
	// FlaggedTurns holds zero-based positions in the top-level
	// sessionDetail.turns array; never TurnDetail.index values and never
	// positions in earlierHistory or child sessions. Flagged means verdict
	// checked with at least one position.
	FlaggedTurns []int `json:"flaggedTurns" nullable:"false" description:"Zero-based positions in the top-level sessionDetail.turns array; never TurnDetail.index values and never positions in earlierHistory or child sessions. Flagged means verdict checked with at least one position."`
}

// PrepareJSONSchema bounds every flaggedTurns item, which struct tags cannot
// reach.
func (ScryCheckResponse) PrepareJSONSchema(s *jsonschema.Schema) error {
	property, ok := s.Properties["flaggedTurns"]
	if !ok || property.TypeObject == nil {
		return fmt.Errorf("scry check response schema has no flaggedTurns property to bound")
	}
	item := jsonschema.Schema{}
	item.AddType(jsonschema.Integer)
	item.WithMinimum(0)
	item.WithMaximum(float64(maxSafeJSONInteger))
	property.TypeObject.WithItems(*(&jsonschema.Items{}).WithSchemaOrBool(item.ToSchemaOrBool()))
	s.Properties["flaggedTurns"] = property
	return nil
}

// ScryHealthResponse is the body of scry's GET /healthz. Village compares
// ContractVersion with ScryContractVersion before it calls POST /v1/check.
type ScryHealthResponse struct {
	ContractVersion string `json:"contractVersion" minLength:"1"`
	ScryVersion     string `json:"scryVersion" pattern:"^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)$"`
}

var (
	scryRecordFields   = fieldSet("publishedAtMs", "updatedAtMs")
	scryResponseFields = fieldSet("scryVersion", "verdict", "flaggedTurns")
	scryHealthFields   = fieldSet("contractVersion", "scryVersion")
)

func scryError(what, why, fix string) error { return publicationError(what, why, fix) }

// Validate refuses a request built for another contract version, a transcript
// id that is not canonical, and a record time below 1.
func (r ScryCheckRequest) Validate() error {
	const what = "scry check request"
	if r.ContractVersion != ScryContractVersion {
		return scryError(what, fmt.Sprintf("contractVersion %q is not %q", r.ContractVersion, ScryContractVersion), "build the request for the pinned scry contract version")
	}
	if _, err := NewTranscriptID(string(r.TranscriptID)); err != nil {
		return scryError(what, fmt.Sprintf("transcriptId is invalid: %v", err), "send a canonical lowercase UUID")
	}
	for name, ms := range map[string]int64{"record.publishedAtMs": r.Record.PublishedAtMs, "record.updatedAtMs": r.Record.UpdatedAtMs} {
		if ms < 1 || ms > maxSafeJSONInteger {
			return scryError(what, fmt.Sprintf("%s %d is outside 1..%d", name, ms, maxSafeJSONInteger), "send the epoch milliseconds village holds")
		}
	}
	return nil
}

// UnmarshalJSON refuses unknown members, so a publisher field copied into the
// record fails.
func (r *ScryTranscriptRecord) UnmarshalJSON(data []byte) error {
	type wire ScryTranscriptRecord
	var decoded wire
	if err := decodeRequiredStrict(data, scryRecordFields, nil, &decoded, "scry transcript record"); err != nil {
		return err
	}
	*r = ScryTranscriptRecord(decoded)
	return nil
}

// Validate is the single public flagged-turns rule: closed verdict,
// flaggedTurns non-null, empty when unchecked, ascending, unique, each in
// 0..maxSafeJSONInteger and < turnCount (len(sessionDetail.turns)).
func (r ScryCheckResponse) Validate(turnCount int) error {
	const what = "scry check response"
	if !r.Verdict.IsValid() {
		return scryError(what, fmt.Sprintf("verdict %q is outside the closed set %v", r.Verdict, AllScryVerdicts), "send checked or unchecked")
	}
	if !scrySemVerCoreRE.MatchString(r.ScryVersion) {
		return scryError(what, fmt.Sprintf("scryVersion %q is not MAJOR.MINOR.PATCH", r.ScryVersion), "send a SemVer core version without a leading v")
	}
	if r.Verdict == ScryVerdictUnchecked && len(r.FlaggedTurns) > 0 {
		return scryError(what, "an unchecked verdict lists flagged turns", "send an empty flaggedTurns when verdict is unchecked")
	}
	if err := validateTurnIndexes("flaggedTurns", r.FlaggedTurns); err != nil {
		return err
	}
	if n := len(r.FlaggedTurns); n > 0 && r.FlaggedTurns[n-1] >= turnCount {
		return scryError(what, fmt.Sprintf("flaggedTurns position %d is not below the turn count %d", r.FlaggedTurns[n-1], turnCount), "flag only positions in the top-level sessionDetail.turns array")
	}
	return nil
}

// validateTurnIndexes is the one rule for a list of turn positions: non-null,
// strictly ascending (so unique), each in 0..maxSafeJSONInteger. A caller adds
// its own bounds and emptiness rules.
func validateTurnIndexes(at string, turns []int) error {
	if turns == nil {
		return scryError(at, "the list is null", "send an empty array when no turn is listed")
	}
	for i, turn := range turns {
		if turn < 0 || int64(turn) > maxSafeJSONInteger {
			return scryError(at, fmt.Sprintf("position %d at index %d is outside 0..%d", turn, i, maxSafeJSONInteger), "send zero-based positions")
		}
		if i > 0 && turn <= turns[i-1] {
			return scryError(at, fmt.Sprintf("position %d at index %d does not follow %d in ascending order without repeats", turn, i, turns[i-1]), "send positions in strictly ascending order, each once")
		}
	}
	return nil
}

// UnmarshalJSON refuses unknown members and a null flaggedTurns.
func (r *ScryCheckResponse) UnmarshalJSON(data []byte) error {
	raw, err := decodeStrictObject(data, scryResponseFields, "scry check response")
	if err != nil {
		return err
	}
	for field := range scryResponseFields {
		value, present := raw[field]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return scryError("scry check response", fmt.Sprintf("required field %q is missing or null", field), "send the complete response")
		}
	}
	type wire ScryCheckResponse
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return scryError("scry check response", err.Error(), "send values matching the declared field types")
	}
	*r = ScryCheckResponse(decoded)
	return nil
}

// Validate requires a non-empty contractVersion and a SemVer core scryVersion.
// It does not require the exact contract version: health reports what scry
// runs, and village compares the value itself.
func (r ScryHealthResponse) Validate() error {
	const what = "scry health response"
	if r.ContractVersion == "" {
		return scryError(what, "contractVersion is empty", "send the contract version scry runs")
	}
	if !scrySemVerCoreRE.MatchString(r.ScryVersion) {
		return scryError(what, fmt.Sprintf("scryVersion %q is not MAJOR.MINOR.PATCH", r.ScryVersion), "send a SemVer core version without a leading v")
	}
	return nil
}

// UnmarshalJSON refuses unknown members.
func (r *ScryHealthResponse) UnmarshalJSON(data []byte) error {
	type wire ScryHealthResponse
	var decoded wire
	if err := decodeRequiredStrict(data, scryHealthFields, nil, &decoded, "scry health response"); err != nil {
		return err
	}
	*r = ScryHealthResponse(decoded)
	return nil
}
