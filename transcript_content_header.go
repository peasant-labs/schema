package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// TranscriptSessionHeaderPresenceMembers names the fields whose absence must
// remain distinct from an explicitly supplied zero value.
var TranscriptSessionHeaderPresenceMembers = []string{
	"id", "harness", "startTime", "endTime", "durationMins",
	"totalTokens", "tokensIn", "tokensOut", "turnCount", "toolCallCount",
}

// TranscriptSessionHeader is the bounded metadata projection of the served
// session detail. It excludes nested records and preserves the absence of
// source members which would otherwise acquire a typed zero value.
type TranscriptSessionHeader struct {
	SchemaVersion PushContractVersion `json:"schemaVersion,omitempty"`
	ID            *string             `json:"id,omitempty" nullable:"false"`
	Harness       *string             `json:"harness,omitempty" nullable:"false"`
	// Timestamps retain the served spelling, even if it cannot be re-encoded
	// by time.Time. They make no claim that the value is a usable instant.
	StartTime    *string  `json:"startTime,omitempty" nullable:"false"`
	EndTime      *string  `json:"endTime,omitempty" nullable:"false"`
	DurationMins *float64 `json:"durationMins,omitempty" nullable:"false"`
	// Integer counters retain served values. JavaScript numbers may round values
	// outside the safe-integer range, so Zod cannot check their exact original
	// int64 lexemes. Go's constructor and served bytes are authoritative.
	// Readers treat unsafe counts as unknown: never display them as exact or use
	// them for ordering or totals.
	TotalTokens           *int                  `json:"totalTokens,omitempty" nullable:"false"`
	TokensIn              *int                  `json:"tokensIn,omitempty" nullable:"false"`
	TokensOut             *int                  `json:"tokensOut,omitempty" nullable:"false"`
	TurnCount             *int                  `json:"turnCount,omitempty" nullable:"false" description:"Turns: the number of folded turns of the session's main partition, system and tool rows included. Equal to the segment directory's totalTurns at the same revision. Not the count a reader is shown; see conversationTurnCount."`
	InputSubmissionCount  *int64                `json:"inputSubmissionCount,omitempty" nullable:"false"`
	ConversationTurnCount *int64                `json:"conversationTurnCount,omitempty" nullable:"false"`
	ToolCallCount         *int                  `json:"toolCallCount,omitempty" nullable:"false"`
	Source                string                `json:"source,omitempty"`
	Status                string                `json:"status,omitempty"`
	Project               string                `json:"project,omitempty"`
	Model                 string                `json:"model,omitempty"`
	ThinkingLevel         ThinkingLevel         `json:"thinkingLevel,omitempty"`
	ThinkingLevelRaw      ThinkingLevelRaw      `json:"thinkingLevelRaw,omitempty"`
	ThinkingLevelHistory  []ThinkingLevelChange `json:"thinkingLevelHistory,omitempty"`
	WorkingDirectory      string                `json:"workingDirectory,omitempty"`
	GitBranch             string                `json:"gitBranch,omitempty"`
	GitRemote             string                `json:"gitRemote,omitempty"`
	ParentSessionID       *string               `json:"parentSessionId,omitempty" nullable:"false"`
	RootSessionID         *string               `json:"rootSessionId,omitempty" nullable:"false"`
	Purpose               string                `json:"purpose,omitempty"`
	// EarlierHistory contains the string states actually stated by sections, in
	// order. It is not positional and its length is not a section count.
	EarlierHistory []string `json:"earlierHistory,omitempty" nullable:"false"`
	Outcome        string   `json:"outcome,omitempty"`
	SessionOrigin  string   `json:"sessionOrigin,omitempty"`
}

// NewTranscriptSessionHeader projects the served sessionDetail object, using
// exact decoded member names. It must not be passed an envelope or a typed
// re-encoding of the served document.
func NewTranscriptSessionHeader(sessionDetail json.RawMessage) (TranscriptSessionHeader, error) {
	const where = "transcript header projection failed at schema.NewTranscriptSessionHeader: "
	var header TranscriptSessionHeader
	if !utf8.Valid(sessionDetail) || !json.Valid(sessionDetail) {
		return header, fmt.Errorf("%sthe served detail is not one valid UTF-8 JSON object; no header can be projected; pass the served sessionDetail object", where)
	}
	members, err := transcriptRawMembers(sessionDetail)
	if err != nil {
		return header, fmt.Errorf("%sthe served detail is not one object with unique members: %w; pass the served sessionDetail object", where, err)
	}
	if _, ok := members["sessionDetail"]; ok {
		return header, fmt.Errorf("%san envelope was supplied instead of its detail; select the served sessionDetail object", where)
	}
	value := reflect.ValueOf(&header).Elem()
	typeOf := value.Type()
	for i := range value.NumField() {
		name := strings.Split(typeOf.Field(i).Tag.Get("json"), ",")[0]
		raw, ok := members[name]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		if name == "earlierHistory" {
			states, err := transcriptHeaderHistory(raw)
			if err != nil {
				return TranscriptSessionHeader{}, fmt.Errorf("%searlierHistory cannot be projected: %w; pass the served section array", where, err)
			}
			header.EarlierHistory = states
			continue
		}
		if err := json.Unmarshal(raw, value.Field(i).Addr().Interface()); err != nil {
			return TranscriptSessionHeader{}, fmt.Errorf("%s%s cannot be decoded: %w; pass the served value in its declared type", where, name, err)
		}
	}
	return header, nil
}

func transcriptHeaderHistory(raw json.RawMessage) ([]string, error) {
	var sections []json.RawMessage
	if err := json.Unmarshal(raw, &sections); err != nil {
		return nil, err
	}
	var states []string
	for i, section := range sections {
		if bytes.Equal(bytes.TrimSpace(section), []byte("null")) {
			continue
		}
		members, err := transcriptRawMembers(section)
		if err != nil {
			return nil, fmt.Errorf("section %d is not an object with unique members: %w", i, err)
		}
		state, present := members["state"]
		if !present {
			continue
		}
		// Only a JSON string states a history value. In particular, null and
		// absent are not the empty string.
		state = bytes.TrimSpace(state)
		if len(state) == 0 || state[0] != '"' {
			continue
		}
		var text string
		if err := json.Unmarshal(state, &text); err != nil {
			return nil, fmt.Errorf("section %d state is not valid JSON: %w", i, err)
		}
		states = append(states, text)
	}
	return states, nil
}

// TranscriptSessionHeaderRepositoryMembers is the exact set of header fields
// removed by the repository-withholding projection.
var TranscriptSessionHeaderRepositoryMembers = []string{
	"project", "workingDirectory", "gitBranch", "gitRemote",
}

// WithholdSessionHeaderRepository omits repository identity from a header.
func WithholdSessionHeaderRepository(h TranscriptSessionHeader) TranscriptSessionHeader {
	h.Project = ""
	h.WorkingDirectory = ""
	h.GitBranch = ""
	h.GitRemote = ""
	return h
}

// TranscriptRetainedState describes evidence withheld from the bounded header.
type TranscriptRetainedState string

const (
	TranscriptRetainedAbsent   TranscriptRetainedState = "absent"
	TranscriptRetainedDeferred TranscriptRetainedState = "deferred"
)

var AllTranscriptRetainedStates = []TranscriptRetainedState{TranscriptRetainedAbsent, TranscriptRetainedDeferred}

func (v TranscriptRetainedState) IsValid() bool { return inSet(v, AllTranscriptRetainedStates) }
func (TranscriptRetainedState) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Retained State", AllTranscriptRetainedStates)
}

// TranscriptContentHeader is a bounded read of envelope and session metadata.
type TranscriptContentHeader struct {
	ContractVersion PushContractVersion     `json:"contractVersion"`
	Kind            ContentKind             `json:"kind"`
	Revision        string                  `json:"revision"`
	Representation  string                  `json:"representation"`
	ContentHash     string                  `json:"contentHash,omitempty"`
	SessionHeader   TranscriptSessionHeader `json:"sessionHeader"`
	Canonical       bool                    `json:"canonical"`
	OmittedImages   int                     `json:"omittedImages"`
	// FlaggedTurns are ascending, unique served positions. Turn is present if
	// and only if the associated row has a usable recorded index.
	FlaggedTurns     []TranscriptFlaggedTurn  `json:"flaggedTurns,omitempty"`
	RetainedEvidence *TranscriptRetainedState `json:"retainedEvidence,omitempty"`
}

// Validate checks the envelope facts and its canonical-harness claim.
func (h TranscriptContentHeader) Validate() error {
	const where = "transcript content header validation failed at schema.TranscriptContentHeader.Validate: "
	if !h.Kind.IsValid() || !isDecimalRevision(h.Revision) || h.Representation == "" || h.OmittedImages < 0 {
		return fmt.Errorf("%skind, revision, representation, or omittedImages is invalid; the header cannot be interpreted; emit a valid kind and decimal revision, a representation, and a nonnegative image count", where)
	}
	if h.RetainedEvidence != nil && !h.RetainedEvidence.IsValid() {
		return fmt.Errorf("%sretainedEvidence is outside its closed set; evidence availability cannot be interpreted; omit it or use a published state", where)
	}
	known := false
	if h.SessionHeader.Harness != nil {
		for _, harness := range Harnesses() {
			if *h.SessionHeader.Harness == string(harness) {
				known = true
				break
			}
		}
	}
	if h.Canonical != known {
		return fmt.Errorf("%scanonical and harness disagree; a document that passed the typed contract carries a canonical harness, while a legacy header does not; project the served detail again", where)
	}
	for i, mark := range h.FlaggedTurns {
		if mark.Position < 0 || (i > 0 && mark.Position <= h.FlaggedTurns[i-1].Position) {
			return fmt.Errorf("%sflaggedTurns positions are negative, repeated, or out of order; marks cannot be placed safely; emit unique positions in ascending order", where)
		}
		if mark.Turn != nil && (*mark.Turn < 0 || int64(*mark.Turn) > maxSafeJSONInteger) {
			return fmt.Errorf("%sflaggedTurns[%d].turn is outside the nonnegative safe-integer range; the mark cannot state a usable recorded identity; omit an absent or invalid index", where, i)
		}
	}
	return nil
}
