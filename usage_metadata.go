package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	jsonschema "github.com/swaggest/jsonschema-go"
	"io"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const maxSafeJSONInteger int64 = 9007199254740991

var jsonNumberPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

type UsageScope string

const (
	UsageScopeAssistant UsageScope = "assistant"
	UsageScopeTool      UsageScope = "tool"
	UsageScopeSummary   UsageScope = "summary"
)

var AllUsageScopes = []UsageScope{UsageScopeAssistant, UsageScopeTool, UsageScopeSummary}

func NewUsageScope(raw string) (UsageScope, error) {
	v := UsageScope(raw)
	if !v.IsValid() {
		return "", enumInputError("UsageScope", raw, stringValues(AllUsageScopes))
	}
	return v, nil
}

func (v UsageScope) IsValid() bool {
	return v == UsageScopeAssistant || v == UsageScopeTool || v == UsageScopeSummary
}
func (UsageScope) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Usage Scope", "Native owner scope for detailed token and cost evidence", AllUsageScopes), nil
}

type UsageCompleteness string

const (
	UsageComplete UsageCompleteness = "complete"
	UsagePartial  UsageCompleteness = "partial"
	UsageUnknown  UsageCompleteness = "unknown"
)

var AllUsageCompleteness = []UsageCompleteness{UsageComplete, UsagePartial, UsageUnknown}

func NewUsageCompleteness(raw string) (UsageCompleteness, error) {
	v := UsageCompleteness(raw)
	if !v.IsValid() {
		return "", enumInputError("UsageCompleteness", raw, stringValues(AllUsageCompleteness))
	}
	return v, nil
}

func (v UsageCompleteness) IsValid() bool {
	return v == UsageComplete || v == UsagePartial || v == UsageUnknown
}
func (UsageCompleteness) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Usage Completeness", "Completeness of the five base token fields", AllUsageCompleteness), nil
}

type UsageOwnerID string
type RecordedCostAmount string

const RecordedCostSourceHarnessEstimate = "recorded_harness_estimate"

type TokenUsageDetail struct {
	Input        *int64 `json:"input,omitempty"`
	Output       *int64 `json:"output,omitempty"`
	CacheRead    *int64 `json:"cacheRead,omitempty"`
	CacheWrite   *int64 `json:"cacheWrite,omitempty"`
	CacheWrite1h *int64 `json:"cacheWrite1h,omitempty"`
	Reasoning    *int64 `json:"reasoning,omitempty"`
	TotalTokens  *int64 `json:"totalTokens,omitempty"`
}
type RecordedCostDetail struct {
	Input      *RecordedCostAmount `json:"input,omitempty"`
	Output     *RecordedCostAmount `json:"output,omitempty"`
	CacheRead  *RecordedCostAmount `json:"cacheRead,omitempty"`
	CacheWrite *RecordedCostAmount `json:"cacheWrite,omitempty"`
	Total      *RecordedCostAmount `json:"total,omitempty"`
	Source     string              `json:"source"`
}
type UsageDetail struct {
	OwnerID        UsageOwnerID        `json:"ownerId"`
	SourceEntryRef SourceEntryRef      `json:"sourceEntryRef"`
	Scope          UsageScope          `json:"scope"`
	Completeness   UsageCompleteness   `json:"completeness"`
	Tokens         *TokenUsageDetail   `json:"tokens,omitempty"`
	Cost           *RecordedCostDetail `json:"cost,omitempty"`
}

type NativeMetadataKind string

const (
	NativeMetadataPiCustomData           NativeMetadataKind = "pi.custom.data"
	NativeMetadataPiCustomMessageDetails NativeMetadataKind = "pi.custommessage.details"
	NativeMetadataPiToolResultDetails    NativeMetadataKind = "pi.toolresult.details"
	NativeMetadataPiCompactionDetails    NativeMetadataKind = "pi.compaction.details"
	NativeMetadataPiBranchSummaryDetails NativeMetadataKind = "pi.branchsummary.details"
)

var AllNativeMetadataKinds = []NativeMetadataKind{NativeMetadataPiCustomData, NativeMetadataPiCustomMessageDetails, NativeMetadataPiToolResultDetails, NativeMetadataPiCompactionDetails, NativeMetadataPiBranchSummaryDetails}

func NewNativeMetadataKind(raw string) (NativeMetadataKind, error) {
	v := NativeMetadataKind(raw)
	if !v.IsValid() {
		return "", enumInputError("NativeMetadataKind", raw, stringValues(AllNativeMetadataKinds))
	}
	return v, nil
}

func (v NativeMetadataKind) IsValid() bool {
	for _, x := range AllNativeMetadataKinds {
		if v == x {
			return true
		}
	}
	return false
}
func (NativeMetadataKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Native Metadata Kind", "Kind of bounded non-conversational native metadata", AllNativeMetadataKinds), nil
}

type NativeMetadataSourceType string

const (
	NativeSourcePiCustom        NativeMetadataSourceType = "pi.custom"
	NativeSourcePiCustomMessage NativeMetadataSourceType = "pi.custom_message"
	NativeSourcePiMessage       NativeMetadataSourceType = "pi.message"
	NativeSourcePiCompaction    NativeMetadataSourceType = "pi.compaction"
	NativeSourcePiBranchSummary NativeMetadataSourceType = "pi.branch_summary"
)

var AllNativeMetadataSourceTypes = []NativeMetadataSourceType{NativeSourcePiCustom, NativeSourcePiCustomMessage, NativeSourcePiMessage, NativeSourcePiCompaction, NativeSourcePiBranchSummary}

func NewNativeMetadataSourceType(raw string) (NativeMetadataSourceType, error) {
	v := NativeMetadataSourceType(raw)
	if !v.IsValid() {
		return "", enumInputError("NativeMetadataSourceType", raw, stringValues(AllNativeMetadataSourceTypes))
	}
	return v, nil
}

func (v NativeMetadataSourceType) IsValid() bool {
	for _, x := range AllNativeMetadataSourceTypes {
		if v == x {
			return true
		}
	}
	return false
}
func (NativeMetadataSourceType) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Native Metadata Source Type", "Native source category for public metadata", AllNativeMetadataSourceTypes), nil
}

type NativePiMessageRole string

const NativePiMessageRoleToolResult NativePiMessageRole = "toolResult"

var AllNativePiMessageRoles = []NativePiMessageRole{NativePiMessageRoleToolResult}

func NewNativePiMessageRole(raw string) (NativePiMessageRole, error) {
	v := NativePiMessageRole(raw)
	if !v.IsValid() {
		return "", enumInputError("NativePiMessageRole", raw, stringValues(AllNativePiMessageRoles))
	}
	return v, nil
}

func stringValues[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}
	return out
}

func enumInputError(name, raw string, allowed []string) error {
	return fmt.Errorf("%s construction failed at schema.New%s: value %q is outside the closed set [%s]; callers cannot safely classify the wire value; use one of the published values", name, name, raw, strings.Join(allowed, ", "))
}

func (v NativePiMessageRole) IsValid() bool { return v == NativePiMessageRoleToolResult }
func (NativePiMessageRole) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Native Pi Message Role", "Pi message role needed for public metadata validation", AllNativePiMessageRoles), nil
}

type NativeSourceRef struct {
	EntryRef    SourceEntryRef           `json:"entryRef"`
	SourceType  NativeMetadataSourceType `json:"sourceType"`
	MessageRole NativePiMessageRole      `json:"messageRole,omitempty"`
}
type NativeAttachmentRef struct {
	TurnIndex  *int   `json:"turnIndex,omitempty"`
	ToolCallID string `json:"toolCallId,omitempty"`
}
type NativeMetadataRecord struct {
	ID         string               `json:"id"`
	Kind       NativeMetadataKind   `json:"kind"`
	Source     NativeSourceRef      `json:"source"`
	Attachment *NativeAttachmentRef `json:"attachment,omitempty"`
	CustomType string               `json:"customType,omitempty"`
	Data       json.RawMessage      `json:"data"`
}

func ValidateUsageDetail(v UsageDetail) error {
	if !validPublicRef(string(v.OwnerID)) || !validPublicRef(string(v.SourceEntryRef)) {
		return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: ownerId and sourceEntryRef must each contain 1..96 valid UTF-8 bytes; attribution cannot be checked; emit bounded nonreversible public references")
	}
	if !v.Scope.IsValid() || !v.Completeness.IsValid() {
		return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: scope %q or completeness %q is outside its closed set; consumers cannot classify the owner; use the published enum values", v.Scope, v.Completeness)
	}
	present := 0
	if v.Tokens != nil {
		fields := []*int64{v.Tokens.Input, v.Tokens.Output, v.Tokens.CacheRead, v.Tokens.CacheWrite, v.Tokens.TotalTokens}
		for _, p := range append(fields, v.Tokens.CacheWrite1h, v.Tokens.Reasoning) {
			if p != nil && (*p < 0 || *p > maxSafeJSONInteger) {
				return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: a token field is outside the JS-safe nonnegative integer range; cross-language values would be unsafe; emit 0..9007199254740991")
			}
		}
		for _, p := range fields {
			if p != nil {
				present++
			}
		}
		if v.Tokens.Reasoning != nil && v.Tokens.Output != nil && *v.Tokens.Reasoning > *v.Tokens.Output {
			return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: reasoning exceeds output; subset evidence is inconsistent; correct the recorded fields")
		}
		if v.Tokens.CacheWrite1h != nil && v.Tokens.CacheWrite != nil && *v.Tokens.CacheWrite1h > *v.Tokens.CacheWrite {
			return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: cacheWrite1h exceeds cacheWrite; subset evidence is inconsistent; correct the recorded fields")
		}
		if present == 5 {
			sum := *v.Tokens.Input + *v.Tokens.Output
			if sum > maxSafeJSONInteger-*v.Tokens.CacheRead || sum+*v.Tokens.CacheRead > maxSafeJSONInteger-*v.Tokens.CacheWrite || sum+*v.Tokens.CacheRead+*v.Tokens.CacheWrite != *v.Tokens.TotalTokens {
				return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: totalTokens does not equal the checked sum of base token fields; totals are unreliable; correct the recorded total")
			}
		}
	}
	expected := UsageUnknown
	if present == 5 {
		expected = UsageComplete
	} else if present > 0 {
		expected = UsagePartial
	}
	if v.Completeness != expected {
		return fmt.Errorf("detailed usage validation failed at schema.ValidateUsageDetail: completeness %q disagrees with base token presence; consumers would misstate unknown or partial evidence; use %q", v.Completeness, expected)
	}
	if v.Cost != nil {
		if v.Cost.Source != RecordedCostSourceHarnessEstimate {
			return fmt.Errorf("recorded cost validation failed at schema.ValidateUsageDetail: source %q is unsupported; consumers cannot label the estimate; use recorded_harness_estimate", v.Cost.Source)
		}
		for _, p := range []*RecordedCostAmount{v.Cost.Input, v.Cost.Output, v.Cost.CacheRead, v.Cost.CacheWrite, v.Cost.Total} {
			if p != nil {
				if !jsonNumberPattern.MatchString(string(*p)) {
					return fmt.Errorf("recorded cost validation failed at schema.ValidateUsageDetail: amount %q is not JSON number syntax; consumers cannot preserve it as recorded numeric evidence; emit a finite nonnegative JSON-number string", *p)
				}
				n, e := strconv.ParseFloat(string(*p), 64)
				if e != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
					return fmt.Errorf("recorded cost validation failed at schema.ValidateUsageDetail: amount %q is not a finite nonnegative JS number; preserve a valid recorded numeric string without repricing", *p)
				}
			}
		}
	}
	return nil
}

func ValidateSessionDetailPayload(value SessionDetailPayload) error {
	knownHarness := false
	for _, harness := range Harnesses() {
		if value.Harness == harness {
			knownHarness = true
			break
		}
	}
	if !knownHarness || (value.Outcome != "" && !value.Outcome.IsValid()) || (value.SessionOrigin != "" && !value.SessionOrigin.IsValid()) {
		return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: harness, outcome, or sessionOrigin is outside its closed set; consumers cannot classify the session; use published enum values")
	}
	if err := ValidateInputSubmissionCount(value.InputSubmissionCount, "sessionDetail.inputSubmissionCount"); err != nil {
		return err
	}
	if value.RootSessionID != nil {
		if _, err := NewSessionID(string(*value.RootSessionID)); err != nil {
			return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: rootSessionId is malformed; graph identity cannot be preserved; provide a canonical session identifier: %w", err)
		}
	}
	if !value.Purpose.IsValid() {
		return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: purpose %q is outside its closed set; consumers cannot classify the session; use a published purpose or omit it", value.Purpose)
	}
	if err := ValidateSessionRelationships(value.Relationships); err != nil {
		return err
	}
	parent, err := durableStartedByTarget(value.Relationships)
	if err != nil {
		return err
	}
	if len(value.Relationships) > 0 && value.ParentSessionID != nil && (parent == nil || *value.ParentSessionID != *parent) {
		return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: parentSessionId disagrees with durable started_by relationship; consumers could navigate to the wrong parent; derive the legacy field from the durable relationship")
	}
	if err := validateTurnEvidence(value.Turns); err != nil {
		return err
	}
	return ValidateNativeMetadata(value)
}

func validateTurnEvidence(turns []TurnDetail) error {
	owners := map[UsageOwnerID]bool{}
	sources := map[string]bool{}
	turnIndexes := map[int]bool{}
	toolIDs := map[string]bool{}
	for i := range turns {
		t := &turns[i]
		if err := ValidateObservedModelEvidence(t.Role, t.ObservedModel); err != nil {
			return err
		}
		if !t.Role.IsValid() || (t.EntryType != "" && !t.EntryType.IsValid()) || (t.StopReason != nil && !t.StopReason.IsValid()) {
			return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turn role, entryType, or stopReason is outside its closed set; consumers cannot classify the turn; use published enum values")
		}
		if turnIndexes[t.Index] {
			return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: duplicate turn index %d makes metadata attachment ambiguous; emit unique turn indices", t.Index)
		}
		turnIndexes[t.Index] = true
		if t.SourceEntryRef != "" && !validOptionalPublicRef(string(t.SourceEntryRef)) {
			return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turn sourceEntryRef is invalid UTF-8 or exceeds 96 bytes; attribution cannot be checked; emit a bounded public reference or omit it")
		}
		if t.Provenance != nil {
			if err := t.Provenance.Validate(); err != nil {
				return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turns[%d].provenance: %w", i, err)
			}
		}
		if t.Usage != nil {
			if !((t.Role == RoleAssistant && t.Usage.Scope == UsageScopeAssistant) || (t.Role == RoleSystem && t.Usage.Scope == UsageScopeSummary)) {
				return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turn %d usage scope does not match its role; attribution is ambiguous; use assistant for assistant turns or summary for visible system summaries", t.Index)
			}
			if t.SourceEntryRef == "" || t.Usage.SourceEntryRef != t.SourceEntryRef {
				return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turn usage sourceEntryRef does not match its turn; attribution is ambiguous; emit matching bounded source references")
			}
			if err := validateUniqueUsage(*t.Usage, owners, sources); err != nil {
				return err
			}
		}
		for j := range t.ToolCalls {
			tool := &t.ToolCalls[j]
			if tool.ToolKind != "" && !tool.ToolKind.IsValid() {
				return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: toolKind is outside its closed set; the tool cannot be classified; use a published tool kind")
			}
			if tool.CallProvenance != nil {
				if err := tool.CallProvenance.Validate(); err != nil {
					return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turns[%d].toolCalls[%d].callProvenance: %w", i, j, err)
				}
			}
			if tool.ResultProvenance != nil {
				if err := tool.ResultProvenance.Validate(); err != nil {
					return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: turns[%d].toolCalls[%d].resultProvenance: %w", i, j, err)
				}
			}
			if tool.ID != "" {
				if toolIDs[tool.ID] {
					return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: duplicate tool call id %q makes metadata attachment ambiguous; emit unique tool ids", tool.ID)
				}
				toolIDs[tool.ID] = true
			}
			for _, ref := range []string{tool.ID, string(tool.CallEntryRef), string(tool.ResultEntryRef)} {
				if ref != "" && !validOptionalPublicRef(ref) {
					return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: a tool source reference is invalid UTF-8 or exceeds 96 bytes; attribution cannot be checked; emit bounded public references or omit them")
				}
			}
			u := t.ToolCalls[j].Usage
			if u != nil {
				if u.Scope != UsageScopeTool {
					return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: tool usage has non-tool scope; attribution is ambiguous; use scope tool")
				}
				if t.ToolCalls[j].ResultEntryRef == "" || u.SourceEntryRef != t.ToolCalls[j].ResultEntryRef {
					return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: tool usage sourceEntryRef does not match resultEntryRef; attribution is ambiguous; emit matching bounded result references")
				}
				if err := validateUniqueUsage(*u, owners, sources); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func validateUniqueUsage(v UsageDetail, owners map[UsageOwnerID]bool, sources map[string]bool) error {
	if err := ValidateUsageDetail(v); err != nil {
		return err
	}
	if owners[v.OwnerID] || sources[string(v.SourceEntryRef)] {
		return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload: duplicate usage ownerId or sourceEntryRef; native owners would be conflated; emit one distinct public reference per native owner")
	}
	owners[v.OwnerID] = true
	sources[string(v.SourceEntryRef)] = true
	return nil
}
func ValidateTranscriptContent(v TranscriptContent) error {
	if !v.Kind.IsValid() || v.SessionDetail == nil {
		return fmt.Errorf("transcript content validation failed at schema.ValidateTranscriptContent: kind is unsupported or sessionDetail is absent; the envelope cannot be consumed; send kind session_detail with its payload")
	}
	return ValidateSessionDetailPayload(*v.SessionDetail)
}

func ValidateNativeMetadata(v SessionDetailPayload) error {
	if len(v.NativeMetadata) > 0 && v.Harness != HarnessPi {
		return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadata: pi metadata is attached to harness %q; consumers would misattribute evidence; emit it only for harness pi", v.Harness)
	}
	return ValidateNativeMetadataRecords(v.NativeMetadata, v.Turns)
}

// ValidateNativeMetadataRecords validates metadata and its public targets before
// a producer attaches them to a harness-specific session envelope.
func ValidateNativeMetadataRecords(records []NativeMetadataRecord, targets []TurnDetail) error {
	if err := validateTurnEvidence(targets); err != nil {
		return err
	}
	if len(records) > 256 {
		return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadata: record count exceeds 256; payload cannot be bounded; reduce metadata")
	}
	ids := map[string]bool{}
	turns := map[int]*TurnDetail{}
	for i := range targets {
		if _, exists := turns[targets[i].Index]; exists {
			return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadataRecords: duplicate turn index makes attachment ambiguous; provide unique target indices")
		}
		turns[targets[i].Index] = &targets[i]
	}
	total := 0
	for _, m := range records {
		if len(m.Data) == 0 || isNullRaw(m.Data) {
			return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadataRecords: required data is missing or null; evidence cannot be checked; provide a non-null JSON value")
		}
		if !m.Kind.IsValid() || !m.Source.SourceType.IsValid() || !validPublicRef(m.ID) || !validPublicRef(string(m.Source.EntryRef)) || ids[m.ID] {
			return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadata: id, kind, source type, or source ref is invalid or duplicated; attribution cannot be trusted; emit unique bounded refs and published enum values")
		}
		if m.CustomType != "" && (!utf8.ValidString(m.CustomType) || len(m.CustomType) > 128) {
			return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadata: customType is invalid UTF-8 or exceeds 128 bytes; consumers cannot identify the extension; emit a shorter valid value")
		}
		if m.Attachment != nil && !validOptionalPublicRef(m.Attachment.ToolCallID) {
			return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadataRecords: attachment tool reference is invalid UTF-8 or exceeds 96 bytes; attribution is unsafe; use a bounded public reference")
		}
		ids[m.ID] = true
		total += len(m.Data)
		if len(m.Data) > 65536 || total > 1048576 {
			return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadata: data exceeds per-record or aggregate byte budget; outward metadata is unsafe; reduce it below 64 KiB per record and 1 MiB total")
		}
		if _, err := DecodeNativeMetadataDataRaw(m.Data); err != nil {
			return err
		}
		if err := validateMetadataMatrix(m, turns); err != nil {
			return err
		}
	}
	return nil
}

// DecodeNativeMetadataRecordsRaw preserves presence and lexical evidence before
// validating the same record matrix used by detail and transcript consumers.
func DecodeNativeMetadataRecordsRaw(raw []byte, targets []TurnDetail) ([]NativeMetadataRecord, error) {
	if err := ScanRawJSONDocument(raw, RawJSONPathPolicy{MaxDocumentBytes: 8 << 20, MaxDocumentDepth: 64, OpaqueMetadataPointers: []string{"/*/data"}}); err != nil {
		return nil, err
	}
	if isNullRaw(raw) {
		return nil, fmt.Errorf("native metadata validation failed at schema.DecodeNativeMetadataRecordsRaw: records root is null; no record array can be consumed; provide an array, including an empty array when needed")
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	for _, item := range items {
		if err := validateMetadataRawPresence(item); err != nil {
			return nil, err
		}
	}
	var records []NativeMetadataRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, err
	}
	return records, ValidateNativeMetadataRecords(records, targets)
}
func validPublicRef(v string) bool         { return v != "" && utf8.ValidString(v) && len(v) <= 96 }
func validOptionalPublicRef(v string) bool { return utf8.ValidString(v) && len(v) <= 96 }

func validateMetadataMatrix(m NativeMetadataRecord, turns map[int]*TurnDetail) error {
	attach := m.Attachment
	fail := func() error {
		return fmt.Errorf("native metadata validation failed at schema.ValidateNativeMetadata: kind %q disagrees with source, role, customType, or attachment target; public correspondence is not checkable; follow the published kind/source/attachment matrix", m.Kind)
	}
	switch m.Kind {
	case NativeMetadataPiCustomData:
		if m.Source.SourceType != NativeSourcePiCustom || m.Source.MessageRole != "" || attach != nil || m.CustomType == "" {
			return fail()
		}
	case NativeMetadataPiCustomMessageDetails:
		if m.Source.SourceType != NativeSourcePiCustomMessage || m.Source.MessageRole != "" || attach == nil || attach.TurnIndex == nil || attach.ToolCallID != "" || m.CustomType == "" {
			return fail()
		}
		t := turns[*attach.TurnIndex]
		if t == nil || t.SourceEntryRef != m.Source.EntryRef || t.Role != RoleSystem {
			return fail()
		}
	case NativeMetadataPiToolResultDetails:
		if m.Source.SourceType != NativeSourcePiMessage || m.Source.MessageRole != NativePiMessageRoleToolResult || attach == nil || attach.TurnIndex == nil || attach.ToolCallID == "" || m.CustomType != "" {
			return fail()
		}
		t := turns[*attach.TurnIndex]
		if t == nil {
			return fail()
		}
		ok := false
		for _, x := range t.ToolCalls {
			if x.ID == attach.ToolCallID && x.ResultEntryRef == m.Source.EntryRef && (x.Usage == nil || x.Usage.SourceEntryRef == m.Source.EntryRef) {
				ok = true
			}
		}
		if !ok {
			return fail()
		}
	case NativeMetadataPiCompactionDetails, NativeMetadataPiBranchSummaryDetails:
		source := NativeSourcePiCompaction
		if m.Kind == NativeMetadataPiBranchSummaryDetails {
			source = NativeSourcePiBranchSummary
		}
		if m.Source.SourceType != source || m.Source.MessageRole != "" || attach == nil || attach.TurnIndex == nil || attach.ToolCallID != "" || m.CustomType != "" {
			return fail()
		}
		t := turns[*attach.TurnIndex]
		if t == nil || t.Role != RoleSystem || t.SourceEntryRef != m.Source.EntryRef || t.Usage == nil || t.Usage.Scope != UsageScopeSummary || t.Usage.SourceEntryRef != m.Source.EntryRef {
			return fail()
		}
	}
	return nil
}

func DecodeSessionDetailPayloadRaw(raw []byte) (SessionDetailPayload, error) {
	if err := ScanRawJSONDocument(raw, sessionDetailRawPolicy()); err != nil {
		return SessionDetailPayload{}, err
	}
	if err := validateSessionDetailRawShape(raw); err != nil {
		return SessionDetailPayload{}, err
	}
	var v SessionDetailPayload
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, err
	}
	return v, ValidateSessionDetailPayload(v)
}
func DecodeTranscriptContentRaw(raw []byte) (TranscriptContent, error) {
	if err := ScanRawJSONDocument(raw, transcriptRawPolicy()); err != nil {
		return TranscriptContent{}, err
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return TranscriptContent{}, err
	}
	detail, ok := envelope["sessionDetail"]
	if !ok || isNullRaw(detail) {
		return TranscriptContent{}, fmt.Errorf("transcript content validation failed at schema.DecodeTranscriptContentRaw: sessionDetail is missing or null; the envelope cannot be consumed; send a non-null detail object")
	}
	if err := validateSessionDetailRawShape(detail); err != nil {
		return TranscriptContent{}, err
	}
	var v TranscriptContent
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, err
	}
	return v, ValidateTranscriptContent(v)
}
func DecodeAuthoritativePublishMetadataRaw(raw []byte) (AuthoritativePublishRequest, error) {
	if err := ScanRawJSONDocument(raw, RawJSONPathPolicy{MaxDocumentBytes: 4 << 20, MaxDocumentDepth: 64}); err != nil {
		return AuthoritativePublishRequest{}, err
	}
	return DecodeAuthoritativePublishRequest(raw)
}
func DecodeNativeMetadataDataRaw(raw []byte) (json.RawMessage, error) {
	p := RawJSONPathPolicy{MaxDocumentBytes: 65536, MaxDocumentDepth: 32, OpaqueMetadataPointers: []string{""}}
	if err := scanRawJSONDocument(raw, p, true); err != nil {
		return nil, err
	}
	return append(json.RawMessage(nil), raw...), nil
}

type RawJSONPathPolicy struct {
	MaxDocumentBytes       int
	MaxDocumentDepth       int
	OpaqueMetadataPointers []string
	CostPointers           []string
}

func sessionDetailRawPolicy() RawJSONPathPolicy {
	return RawJSONPathPolicy{MaxDocumentBytes: 8 << 20, MaxDocumentDepth: 64, OpaqueMetadataPointers: []string{"/nativeMetadata/*/data"}}
}
func transcriptRawPolicy() RawJSONPathPolicy {
	return RawJSONPathPolicy{MaxDocumentBytes: 8 << 20, MaxDocumentDepth: 64, OpaqueMetadataPointers: []string{"/sessionDetail/nativeMetadata/*/data"}}
}
func ScanRawJSONDocument(raw []byte, p RawJSONPathPolicy) error {
	return scanRawJSONDocument(raw, p, false)
}
func scanRawJSONDocument(raw []byte, p RawJSONPathPolicy, rootOpaque bool) error {
	if p.MaxDocumentBytes > 0 && len(raw) > p.MaxDocumentBytes {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: document exceeds %d bytes; decoding would exceed the caller boundary; reduce the document", p.MaxDocumentBytes)
	}
	if !utf8.Valid(raw) {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: document contains invalid UTF-8; decoding would replace source bytes; encode all JSON text as valid UTF-8")
	}
	if err := validateRawUnicodeEscapes(raw); err != nil {
		return err
	}
	rootOpaque = rootOpaque || matchPointer("", p.OpaqueMetadataPointers)
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	scanner := rawJSONScanner{raw: raw, decoder: d, policy: p}
	metadataDepth := 0
	if rootOpaque {
		metadataDepth = 1
	}
	if err := scanner.value("", 1, metadataDepth); err != nil {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument during syntax scanning: %w; decoding cannot safely continue; send complete valid JSON within the selected limits", err)
	}
	var x any
	if err := d.Decode(&x); err == nil {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: trailing JSON value remains; root decoding is ambiguous; send exactly one JSON value")
	} else if err != io.EOF {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: malformed trailing input remains: %w; remove trailing data", err)
	}
	return nil
}

type rawJSONScanner struct {
	raw           []byte
	decoder       *json.Decoder
	policy        RawJSONPathPolicy
	metadataBytes int
}

func (s *rawJSONScanner) value(path string, depth, metadataDepth int) (result error) {
	d, p := s.decoder, s.policy
	if metadataDepth == 0 && matchPointer(path, p.OpaqueMetadataPointers) {
		metadataDepth = 1
	}
	opaque := metadataDepth > 0
	if metadataDepth > 32 {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: metadata depth exceeds 32 at %s; nested metadata is unsafe; flatten the selected data", path)
	}
	if metadataDepth == 1 {
		start := int(d.InputOffset())
		for start < len(s.raw) && strings.ContainsRune(" \t\r\n:,", rune(s.raw[start])) {
			start++
		}
		defer func() {
			if result != nil {
				return
			}
			size := int(d.InputOffset()) - start
			s.metadataBytes += size
			if size > 65536 || s.metadataBytes > 1048576 {
				result = fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: metadata data exceeds its byte budget at %s; selected raw evidence is unsafe; reduce data below 64 KiB per value and 1 MiB aggregate", path)
			}
		}()
	}
	childMetadataDepth := 0
	if opaque {
		childMetadataDepth = metadataDepth + 1
	}
	if p.MaxDocumentDepth > 0 && depth > p.MaxDocumentDepth {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: depth exceeds %d at %s; nested input is unsafe; flatten metadata", p.MaxDocumentDepth, path)
	}
	tok, err := d.Token()
	if err != nil {
		return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: syntax failed at %s: %w; decoding cannot safely continue; send complete valid JSON", path, err)
	}
	switch v := tok.(type) {
	case json.Delim:
		if v == '{' {
			seen := map[string]bool{}
			members := 0
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				key := k.(string)
				if !utf8.ValidString(key) || (opaque && len(key) > 512) {
					return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: object key is invalid or exceeds 512 bytes at %s; metadata is unsafe; shorten the key", path)
				}
				if seen[key] {
					return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: duplicate object key %q at %s; map decoding would silently discard evidence; remove the duplicate", key, path)
				}
				seen[key] = true
				members++
				child := path + "/" + strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
				if opaque && members > 256 {
					return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: metadata object exceeds 256 members at %s; metadata is unsafe; reduce members", child)
				}
				if e := s.value(child, depth+1, childMetadataDepth); e != nil {
					return e
				}
			}
			_, err = d.Token()
			return err
		}
		if v == '[' {
			n := 0
			for d.More() {
				n++
				if opaque && n > 4096 {
					return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: metadata array exceeds 4096 items at %s; metadata is unsafe; reduce items", path)
				}
				if e := s.value(path+"/"+strconv.Itoa(n-1), depth+1, childMetadataDepth); e != nil {
					return e
				}
			}
			_, err = d.Token()
			return err
		}
	case string:
		if opaque && len(v) > 16384 {
			return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: metadata string exceeds 16 KiB at %s; metadata is unsafe; shorten it", path)
		}
	case json.Number:
		if opaque {
			if err := validateMetadataNumber(string(v)); err != nil {
				return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: metadata number at %s is unsafe: %w", path, err)
			}
		}
	}
	return nil
}
func matchPointer(path string, patterns []string) bool {
	for _, p := range patterns {
		if p == "" && path == "" {
			return true
		}
		pp := strings.Split(p, "/")
		xp := strings.Split(path, "/")
		if len(pp) != len(xp) {
			continue
		}
		ok := true
		for i := range pp {
			if pp[i] != "*" && pp[i] != xp[i] {
				ok = false
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func validateMetadataNumber(s string) error {
	if !strings.ContainsAny(s, ".eE") {
		n, e := strconv.ParseInt(s, 10, 64)
		if e != nil || n < -maxSafeJSONInteger || n > maxSafeJSONInteger {
			return fmt.Errorf("integer %q is outside JS-safe bounds; use -9007199254740991..9007199254740991", s)
		}
		return nil
	}
	n, e := strconv.ParseFloat(s, 64)
	if e != nil || math.IsInf(n, 0) || math.IsNaN(n) {
		return fmt.Errorf("number %q is not finite binary64", s)
	}
	mantissa := s
	if i := strings.IndexAny(mantissa, "eE"); i >= 0 {
		mantissa = mantissa[:i]
	}
	if n == 0 && strings.Trim(mantissa, "-0.") != "" {
		return fmt.Errorf("number %q underflows to zero", s)
	}
	if math.Trunc(n) == n && (n < -float64(maxSafeJSONInteger) || n > float64(maxSafeJSONInteger)) {
		return fmt.Errorf("integer-valued number %q is outside JS-safe bounds; use -9007199254740991..9007199254740991", s)
	}
	return nil
}

func validateSessionDetailRawShape(raw []byte) error {
	if err := validateWireShape(raw, reflect.TypeFor[SessionDetailPayload](), "sessionDetail"); err != nil {
		return err
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("session detail raw validation failed at schema.DecodeSessionDetailPayloadRaw: root must be a JSON object: %w; send the required detail object", err)
	}
	if count, present := root["inputSubmissionCount"]; present {
		if err := validateRawInputSubmissionCount(count, "sessionDetail.inputSubmissionCount"); err != nil {
			return err
		}
	}
	var turns []json.RawMessage
	if err := json.Unmarshal(root["turns"], &turns); err != nil {
		return fmt.Errorf("session detail raw validation failed at schema.DecodeSessionDetailPayloadRaw: turns must be an array; provide a turn array")
	}
	for _, turnRaw := range turns {
		var turn map[string]json.RawMessage
		if json.Unmarshal(turnRaw, &turn) != nil {
			continue
		}
		if usage, exists := turn["usage"]; exists && !isNullRaw(usage) {
			if err := validateUsageRawShape(usage); err != nil {
				return err
			}
		}
		var tools []json.RawMessage
		if rawTools, exists := turn["toolCalls"]; exists && json.Unmarshal(rawTools, &tools) == nil {
			for _, toolRaw := range tools {
				var tool map[string]json.RawMessage
				if json.Unmarshal(toolRaw, &tool) == nil {
					if usage, exists := tool["usage"]; exists && !isNullRaw(usage) {
						if err := validateUsageRawShape(usage); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	var metadata []json.RawMessage
	if rawMetadata, exists := root["nativeMetadata"]; exists && !isNullRaw(rawMetadata) && json.Unmarshal(rawMetadata, &metadata) == nil {
		for _, item := range metadata {
			if err := validateMetadataRawPresence(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRawInputSubmissionCount(raw json.RawMessage, path string) error {
	if isNullRaw(raw) {
		return fmt.Errorf("input submission count raw validation failed at schema decoder for %s: explicit null cannot mean unknown or measured zero; omit the field for unknown or send an integer from 0 through 9007199254740991", path)
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err != nil {
		return fmt.Errorf("input submission count raw validation failed at schema decoder for %s: value is not a JSON integer; callers cannot preserve it exactly; send 0..9007199254740991: %w", path, err)
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil {
		return fmt.Errorf("input submission count raw validation failed at schema decoder for %s: value %q is not an integer in 0..9007199254740991; callers cannot preserve it exactly; send a canonical integer", path, number.String())
	}
	return ValidateInputSubmissionCount(&value, path)
}

func validateUsageRawShape(raw json.RawMessage) error {
	var usage map[string]json.RawMessage
	if json.Unmarshal(raw, &usage) != nil {
		return nil
	}
	for _, name := range []string{"ownerId", "sourceEntryRef", "scope", "completeness"} {
		value, ok := usage[name]
		if !ok || isNullRaw(value) {
			return fmt.Errorf("detailed usage raw validation failed at schema.DecodeSessionDetailPayloadRaw: required field %q is missing or null; attribution would be erased; provide a non-null value", name)
		}
	}
	if tokensRaw, ok := usage["tokens"]; ok && !isNullRaw(tokensRaw) {
		var tokens map[string]json.RawMessage
		if json.Unmarshal(tokensRaw, &tokens) == nil {
			for name, value := range tokens {
				if isNullRaw(value) {
					return fmt.Errorf("detailed usage raw validation failed at schema.DecodeSessionDetailPayloadRaw: token field %q is explicitly null; null cannot count as zero or presence; omit unknown fields", name)
				}
			}
		}
	}
	return nil
}

func validateMetadataRawPresence(raw json.RawMessage) error {
	if err := validateWireShape(raw, reflect.TypeFor[NativeMetadataRecord](), "nativeMetadata"); err != nil {
		return err
	}
	var metadata map[string]json.RawMessage
	if json.Unmarshal(raw, &metadata) != nil {
		return nil
	}
	for _, name := range []string{"id", "kind", "source", "data"} {
		value, ok := metadata[name]
		if !ok || isNullRaw(value) {
			return fmt.Errorf("native metadata raw validation failed at schema.DecodeSessionDetailPayloadRaw: required field %q is missing or null; evidence cannot be checked; provide it", name)
		}
	}
	var kind NativeMetadataKind
	_ = json.Unmarshal(metadata["kind"], &kind)
	var source map[string]json.RawMessage
	_ = json.Unmarshal(metadata["source"], &source)
	if attachment, exists := metadata["attachment"]; exists {
		if isNullRaw(attachment) {
			return fmt.Errorf("native metadata raw validation failed at schema detail decoder: attachment is explicitly null; attribution would be erased; omit forbidden attachments or provide the required target")
		}
		var target map[string]json.RawMessage
		_ = json.Unmarshal(attachment, &target)
		if index, exists := target["turnIndex"]; exists && isNullRaw(index) {
			return fmt.Errorf("native metadata raw validation failed at schema detail decoder: attachment.turnIndex is explicitly null; target is ambiguous; provide its index")
		}
		if _, exists := target["toolCallId"]; exists && kind != NativeMetadataPiToolResultDetails {
			return fmt.Errorf("native metadata raw validation failed at schema detail decoder: attachment.toolCallId is forbidden for this kind; target would be misattributed; omit the field")
		}
	}
	forbidden := func(name string) error {
		if _, ok := metadata[name]; ok {
			return fmt.Errorf("native metadata raw validation failed at schema.DecodeSessionDetailPayloadRaw: field %q is present for kind %q where it is forbidden; typed decoding would erase presence evidence; omit the field", name, kind)
		}
		return nil
	}
	sourceForbidden := func(name string) error {
		if _, ok := source[name]; ok {
			return fmt.Errorf("native metadata raw validation failed at schema.DecodeSessionDetailPayloadRaw: source.%s is present for kind %q where it is forbidden; omit the field", name, kind)
		}
		return nil
	}
	switch kind {
	case NativeMetadataPiCustomData:
		if err := forbidden("attachment"); err != nil {
			return err
		}
		return sourceForbidden("messageRole")
	case NativeMetadataPiCustomMessageDetails:
		return sourceForbidden("messageRole")
	case NativeMetadataPiCompactionDetails, NativeMetadataPiBranchSummaryDetails:
		if err := forbidden("customType"); err != nil {
			return err
		}
		return sourceForbidden("messageRole")
	case NativeMetadataPiToolResultDetails:
		return forbidden("customType")
	}
	return nil
}

func isNullRaw(raw json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }

func validateRawUnicodeEscapes(raw []byte) error {
	inString := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) || raw[i] != 'u' || i+4 >= len(raw) {
			continue
		}
		value, err := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if err != nil {
			continue
		}
		r := rune(value)
		if utf16.IsSurrogate(r) {
			if r < 0xD800 || r > 0xDBFF || i+10 >= len(raw) || raw[i+5] != '\\' || raw[i+6] != 'u' {
				return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: unpaired Unicode surrogate escape near byte %d; decoding would replace evidence; use a valid surrogate pair", i-1)
			}
			low, err := strconv.ParseUint(string(raw[i+7:i+11]), 16, 16)
			if err != nil || low < 0xDC00 || low > 0xDFFF {
				return fmt.Errorf("raw JSON validation failed at schema.ScanRawJSONDocument: invalid Unicode surrogate pair near byte %d; decoding would replace evidence; use a valid pair", i-1)
			}
			i += 10
		} else {
			i += 4
		}
	}
	return nil
}
