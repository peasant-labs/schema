package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// MaxLocalKickstartPreviewNodeIDBytes bounds the decoded node ID. The request
// validator separately checks the encoded message, including JSON escaping.
// At worst 512 one-byte control characters take 3072 encoded bytes, leaving
// room for the bounded setup and request IDs and the surrounding envelope.
const MaxLocalKickstartPreviewNodeIDBytes = 512

type LocalKickstartPreviewOp string

const (
	LocalKickstartPreviewOpen  LocalKickstartPreviewOp = "open"
	LocalKickstartPreviewMore  LocalKickstartPreviewOp = "more"
	LocalKickstartPreviewClose LocalKickstartPreviewOp = "close"
)

var AllLocalKickstartPreviewOps = []LocalKickstartPreviewOp{LocalKickstartPreviewOpen, LocalKickstartPreviewMore, LocalKickstartPreviewClose}

type LocalKickstartPreviewContinuation string

const (
	LocalKickstartPreviewContinuationMore LocalKickstartPreviewContinuation = "more"
	LocalKickstartPreviewContinuationEnd  LocalKickstartPreviewContinuation = "end"
	LocalKickstartPreviewContinuationCut  LocalKickstartPreviewContinuation = "cut"
)

var AllLocalKickstartPreviewContinuations = []LocalKickstartPreviewContinuation{LocalKickstartPreviewContinuationMore, LocalKickstartPreviewContinuationEnd, LocalKickstartPreviewContinuationCut}

type LocalKickstartPreviewEventKind string

const (
	LocalKickstartPreviewOpened       LocalKickstartPreviewEventKind = "opened"
	LocalKickstartPreviewSegmentEvent LocalKickstartPreviewEventKind = "segment"
	LocalKickstartPreviewClosed       LocalKickstartPreviewEventKind = "closed"
)

var AllLocalKickstartPreviewEventKinds = []LocalKickstartPreviewEventKind{LocalKickstartPreviewOpened, LocalKickstartPreviewSegmentEvent, LocalKickstartPreviewClosed}

type LocalKickstartPreviewCutReason string

const (
	LocalKickstartPreviewCutSourceBudget       LocalKickstartPreviewCutReason = "source_budget"
	LocalKickstartPreviewCutSerializedBudget   LocalKickstartPreviewCutReason = "serialized_budget"
	LocalKickstartPreviewCutTurnBudget         LocalKickstartPreviewCutReason = "turn_budget"
	LocalKickstartPreviewCutOversizedRecord    LocalKickstartPreviewCutReason = "oversized_record"
	LocalKickstartPreviewCutOriginNotResumable LocalKickstartPreviewCutReason = "origin_not_resumable"
)

var AllLocalKickstartPreviewCutReasons = []LocalKickstartPreviewCutReason{LocalKickstartPreviewCutSourceBudget, LocalKickstartPreviewCutSerializedBudget, LocalKickstartPreviewCutTurnBudget, LocalKickstartPreviewCutOversizedRecord, LocalKickstartPreviewCutOriginNotResumable}

type LocalKickstartPreviewCloseReason string

const (
	LocalKickstartPreviewClosedByClient LocalKickstartPreviewCloseReason = "closed_by_client"
	LocalKickstartPreviewReplaced       LocalKickstartPreviewCloseReason = "replaced"
	LocalKickstartPreviewSessionClosed  LocalKickstartPreviewCloseReason = "session_closed"
	LocalKickstartPreviewRescanned      LocalKickstartPreviewCloseReason = "rescanned"
	LocalKickstartPreviewIdle           LocalKickstartPreviewCloseReason = "idle"
	LocalKickstartPreviewEvicted        LocalKickstartPreviewCloseReason = "evicted"
	LocalKickstartPreviewShutdown       LocalKickstartPreviewCloseReason = "shutdown"
	LocalKickstartPreviewSourceChanged  LocalKickstartPreviewCloseReason = "source_changed"
	LocalKickstartPreviewLateResult     LocalKickstartPreviewCloseReason = "late_result"
	LocalKickstartPreviewUnreadable     LocalKickstartPreviewCloseReason = "unreadable"
	LocalKickstartPreviewOutOfOrder     LocalKickstartPreviewCloseReason = "out_of_order"
	LocalKickstartPreviewSlowClient     LocalKickstartPreviewCloseReason = "slow_client"
	LocalKickstartPreviewRowNotFound    LocalKickstartPreviewCloseReason = "row_not_found"
	LocalKickstartPreviewNotSessionRow  LocalKickstartPreviewCloseReason = "not_a_session_row"
	LocalKickstartPreviewUnavailable    LocalKickstartPreviewCloseReason = "unavailable"
	LocalKickstartPreviewNoTurns        LocalKickstartPreviewCloseReason = "no_turns"
	LocalKickstartPreviewCapacity       LocalKickstartPreviewCloseReason = "capacity"
)

var AllLocalKickstartPreviewCloseReasons = []LocalKickstartPreviewCloseReason{LocalKickstartPreviewClosedByClient, LocalKickstartPreviewReplaced, LocalKickstartPreviewSessionClosed, LocalKickstartPreviewRescanned, LocalKickstartPreviewIdle, LocalKickstartPreviewEvicted, LocalKickstartPreviewShutdown, LocalKickstartPreviewSourceChanged, LocalKickstartPreviewLateResult, LocalKickstartPreviewUnreadable, LocalKickstartPreviewOutOfOrder, LocalKickstartPreviewSlowClient, LocalKickstartPreviewRowNotFound, LocalKickstartPreviewNotSessionRow, LocalKickstartPreviewUnavailable, LocalKickstartPreviewNoTurns, LocalKickstartPreviewCapacity}

func NewLocalKickstartPreviewOp(raw string) (LocalKickstartPreviewOp, error) {
	return newKickstartToken(raw, "LocalKickstartPreviewOp", AllLocalKickstartPreviewOps)
}
func (v LocalKickstartPreviewOp) IsValid() bool  { return inSet(v, AllLocalKickstartPreviewOps) }
func (v LocalKickstartPreviewOp) String() string { return string(v) }
func (v *LocalKickstartPreviewOp) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartPreviewOp, "schema.LocalKickstartPreviewOp.UnmarshalJSON")
}
func NewLocalKickstartPreviewContinuation(raw string) (LocalKickstartPreviewContinuation, error) {
	return newKickstartToken(raw, "LocalKickstartPreviewContinuation", AllLocalKickstartPreviewContinuations)
}
func (v LocalKickstartPreviewContinuation) IsValid() bool {
	return inSet(v, AllLocalKickstartPreviewContinuations)
}
func (v LocalKickstartPreviewContinuation) String() string { return string(v) }
func (v *LocalKickstartPreviewContinuation) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartPreviewContinuation, "schema.LocalKickstartPreviewContinuation.UnmarshalJSON")
}
func NewLocalKickstartPreviewEventKind(raw string) (LocalKickstartPreviewEventKind, error) {
	return newKickstartToken(raw, "LocalKickstartPreviewEventKind", AllLocalKickstartPreviewEventKinds)
}
func (v LocalKickstartPreviewEventKind) IsValid() bool {
	return inSet(v, AllLocalKickstartPreviewEventKinds)
}
func (v LocalKickstartPreviewEventKind) String() string { return string(v) }
func (v *LocalKickstartPreviewEventKind) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartPreviewEventKind, "schema.LocalKickstartPreviewEventKind.UnmarshalJSON")
}
func NewLocalKickstartPreviewCutReason(raw string) (LocalKickstartPreviewCutReason, error) {
	return newKickstartToken(raw, "LocalKickstartPreviewCutReason", AllLocalKickstartPreviewCutReasons)
}
func (v LocalKickstartPreviewCutReason) IsValid() bool {
	return inSet(v, AllLocalKickstartPreviewCutReasons)
}
func (v LocalKickstartPreviewCutReason) String() string { return string(v) }
func (v *LocalKickstartPreviewCutReason) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartPreviewCutReason, "schema.LocalKickstartPreviewCutReason.UnmarshalJSON")
}
func NewLocalKickstartPreviewCloseReason(raw string) (LocalKickstartPreviewCloseReason, error) {
	return newKickstartToken(raw, "LocalKickstartPreviewCloseReason", AllLocalKickstartPreviewCloseReasons)
}
func (v LocalKickstartPreviewCloseReason) IsValid() bool {
	return inSet(v, AllLocalKickstartPreviewCloseReasons)
}
func (v LocalKickstartPreviewCloseReason) String() string { return string(v) }
func (v *LocalKickstartPreviewCloseReason) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartPreviewCloseReason, "schema.LocalKickstartPreviewCloseReason.UnmarshalJSON")
}

func (LocalKickstartPreviewOp) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Preview operation", "Operation on a connection-owned capture", AllLocalKickstartPreviewOps), nil
}
func (LocalKickstartPreviewContinuation) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Preview continuation", "Whether the capture can continue", AllLocalKickstartPreviewContinuations), nil
}
func (LocalKickstartPreviewEventKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Preview event kind", "Kind of capture event", AllLocalKickstartPreviewEventKinds), nil
}
func (LocalKickstartPreviewCutReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Preview cut reason", "Why the capture ended early", AllLocalKickstartPreviewCutReasons), nil
}
func (LocalKickstartPreviewCloseReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Preview close reason", "Why the capture closed", AllLocalKickstartPreviewCloseReasons), nil
}

// LocalKickstartPreviewRequest addresses a capture on the current connection.
type LocalKickstartPreviewRequest struct {
	Op        LocalKickstartPreviewOp `json:"op"`
	RequestID string                  `json:"requestId" minLength:"1" maxLength:"32" pattern:"^[A-Za-z0-9_-]+$"`
	SessionID string                  `json:"sessionId" minLength:"1" maxLength:"128" pattern:"^[A-Za-z0-9_-]+$"`
	NodeID    string                  `json:"nodeId,omitempty" maxLength:"512"`
	CaptureID string                  `json:"captureId,omitempty" minLength:"22" maxLength:"22" pattern:"^[A-Za-z0-9_-]{21}[AQgw]$"`
	After     uint32                  `json:"after,omitempty" minimum:"0" maximum:"4294967295"`
}

// UnmarshalJSON rejects fields that cannot belong to a preview request.
func (r *LocalKickstartPreviewRequest) UnmarshalJSON(data []byte) error {
	type request LocalKickstartPreviewRequest
	var value request
	if err := unmarshalStrict(data, &value, "schema.LocalKickstartPreviewRequest.UnmarshalJSON"); err != nil || !previewHasRequiredMembers(data, "op", "requestId", "sessionId") {
		return fmt.Errorf("kickstart preview request decode failed: invalid request shape; omit undeclared fields and send one request")
	}
	*r = LocalKickstartPreviewRequest(value)
	return nil
}

func previewHasRequiredMembers(data []byte, names ...string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return false
	}
	for _, name := range names {
		raw, found := fields[name]
		if !found || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return false
		}
	}
	return true
}

var previewIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func validPreviewCaptureID(raw string) bool {
	id, err := base64.RawURLEncoding.DecodeString(raw)
	return err == nil && len(id) == 16 && base64.RawURLEncoding.EncodeToString(id) == raw
}

func (r LocalKickstartPreviewRequest) Validate() error {
	fail := func() error {
		return fmt.Errorf("kickstart preview request validation failed: invalid operation or bounded identifier; send the fields required by the operation within the WebSocket read limit")
	}
	if !inSet(r.Op, AllLocalKickstartPreviewOps) || len(r.RequestID) == 0 || len(r.RequestID) > 32 || !previewIDPattern.MatchString(r.RequestID) || len(r.SessionID) == 0 || len(r.SessionID) > 128 || !previewIDPattern.MatchString(r.SessionID) {
		return fail()
	}
	switch r.Op {
	case LocalKickstartPreviewOpen:
		if len(r.NodeID) == 0 || len(r.NodeID) > MaxLocalKickstartPreviewNodeIDBytes || r.CaptureID != "" || r.After != 0 {
			return fail()
		}
	case LocalKickstartPreviewMore, LocalKickstartPreviewClose:
		if r.NodeID != "" || !validPreviewCaptureID(r.CaptureID) || (r.Op == LocalKickstartPreviewClose && r.After != 0) {
			return fail()
		}
	}
	encoded, err := json.Marshal(ClientMessage{Type: MsgKickstartPreview, Preview: &r})
	if err != nil || len(encoded) > 4096 {
		return fail()
	}
	return nil
}

// LocalKickstartPreviewDirectory grows until the capture ends. Sent refs and
// segments are immutable; hasNewer records what was known at send time.
type LocalKickstartPreviewDirectory struct {
	SessionID               SessionID              `json:"sessionId"`
	Revision                string                 `json:"revision" minLength:"1"`
	Complete                bool                   `json:"complete"`
	LoadedSegments          int                    `json:"loadedSegments" minimum:"0" maximum:"9007199254740991"`
	LoadedTurns             int                    `json:"loadedTurns" minimum:"0" maximum:"9007199254740991"`
	LoadedConversationTurns *int                   `json:"loadedConversationTurns,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	TotalSegments           *int                   `json:"totalSegments,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	TotalTurns              *int                   `json:"totalTurns,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	TotalConversationTurns  *int                   `json:"totalConversationTurns,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	Segments                []TranscriptSegmentRef `json:"segments" nullable:"false"`
}

type LocalKickstartPreviewSegment struct {
	TranscriptSegment
	NativeMetadata  []NativeMetadataRecord  `json:"nativeMetadata,omitempty"`
	RetainedUnknown []RetainedUnknownRecord `json:"retainedUnknown,omitempty"`
}

// UnmarshalJSON preserves the canonical segment's required promoted fields.
func (s *LocalKickstartPreviewSegment) UnmarshalJSON(data []byte) error {
	type segment LocalKickstartPreviewSegment
	var value segment
	if err := unmarshalStrict(data, &value, "schema.LocalKickstartPreviewSegment.UnmarshalJSON"); err != nil {
		return fmt.Errorf("preview segment decode failed: invalid segment shape; discard the frame and reopen the preview")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return fmt.Errorf("preview segment decode failed: expected an object; discard the frame and reopen the preview")
	}
	segmentType := reflect.TypeFor[TranscriptSegment]()
	for i := 0; i < segmentType.NumField(); i++ {
		name := strings.Split(segmentType.Field(i).Tag.Get("json"), ",")[0]
		raw, found := fields[name]
		if !found || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("preview segment decode failed: a required segment field is missing or null; discard the frame and reopen the preview")
		}
	}
	*s = LocalKickstartPreviewSegment(value)
	return nil
}

// UnmarshalJSON rejects undeclared fields and null directory counters.
func (d *LocalKickstartPreviewDirectory) UnmarshalJSON(data []byte) error {
	type directory LocalKickstartPreviewDirectory
	var value directory
	if err := unmarshalStrict(data, &value, "schema.LocalKickstartPreviewDirectory.UnmarshalJSON"); err != nil {
		return fmt.Errorf("preview directory decode failed: invalid directory shape; discard the frame and reopen the preview")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return fmt.Errorf("preview directory decode failed: expected an object; discard the frame and reopen the preview")
	}
	typ := reflect.TypeFor[LocalKickstartPreviewDirectory]()
	for i := 0; i < typ.NumField(); i++ {
		parts := strings.Split(typ.Field(i).Tag.Get("json"), ",")
		raw, found := fields[parts[0]]
		if (len(parts) == 1 && !found) || (found && bytes.Equal(bytes.TrimSpace(raw), []byte("null"))) {
			return fmt.Errorf("preview directory decode failed: a directory field is missing or null; discard the frame and reopen the preview")
		}
	}
	*d = LocalKickstartPreviewDirectory(value)
	return nil
}

type LocalKickstartPreviewEvent struct {
	Kind         LocalKickstartPreviewEventKind    `json:"kind"`
	SessionID    string                            `json:"sessionId"`
	RequestID    string                            `json:"requestId,omitempty"`
	CaptureID    string                            `json:"captureId,omitempty"`
	Seq          uint32                            `json:"seq" minimum:"0" maximum:"4294967295"`
	Header       *SessionDetailPayload             `json:"header,omitempty"`
	Directory    *LocalKickstartPreviewDirectory   `json:"directory,omitempty"`
	Segment      *LocalKickstartPreviewSegment     `json:"segment,omitempty"`
	Continuation LocalKickstartPreviewContinuation `json:"continuation,omitempty"`
	CutReason    LocalKickstartPreviewCutReason    `json:"cutReason,omitempty"`
	Lines        []LocalKickstartLine              `json:"lines,omitempty"`
	Reason       LocalKickstartPreviewCloseReason  `json:"reason,omitempty"`
	Code         LocalKickstartErrorCode           `json:"code,omitempty"`
	// Message is a fixed actionable sentence, never source text, paths, or
	// identifiers the requester did not supply.
	Message string `json:"message,omitempty"`
}

// UnmarshalJSON rejects undeclared event fields without echoing source content.
func (frame *LocalKickstartPreviewEvent) UnmarshalJSON(data []byte) error {
	type event LocalKickstartPreviewEvent
	var value event
	if err := unmarshalStrict(data, &value, "schema.LocalKickstartPreviewEvent.UnmarshalJSON"); err != nil || !previewHasRequiredMembers(data, "kind", "sessionId", "seq") {
		return fmt.Errorf("kickstart preview event decode failed: invalid event shape; discard this capture and reopen the preview")
	}
	*frame = LocalKickstartPreviewEvent(value)
	return nil
}

type LocalKickstartPreviewState struct {
	SessionID    string
	CaptureID    string
	Seq          uint32
	Header       *SessionDetailPayload
	Directory    *LocalKickstartPreviewDirectory
	Segments     []LocalKickstartPreviewSegment
	Continuation LocalKickstartPreviewContinuation
	Closed       *LocalKickstartPreviewEvent
}

// Validate checks the event's discriminated shape without echoing private data.
func (frame LocalKickstartPreviewEvent) Validate() error {
	fail := func() error {
		return fmt.Errorf("kickstart preview event validation failed at schema.LocalKickstartPreviewEvent.Validate: the event has an invalid discriminator or field combination; the capture cannot be interpreted safely; reopen the preview")
	}
	if frame.SessionID == "" || !inSet(frame.Kind, AllLocalKickstartPreviewEventKinds) {
		return fail()
	}
	for _, line := range frame.Lines {
		if line.Validate() != nil {
			return fail()
		}
	}
	if frame.Kind == LocalKickstartPreviewClosed {
		if !inSet(frame.Reason, AllLocalKickstartPreviewCloseReasons) || frame.Header != nil || frame.Directory != nil || frame.Segment != nil || frame.Continuation != "" || frame.CutReason != "" || (frame.Code != "" && !frame.Code.IsValid()) {
			return fail()
		}
		return nil
	}
	if frame.Reason != "" || frame.Code != "" || frame.Message != "" || frame.Header == nil || frame.Directory == nil || !validPreviewCaptureID(frame.CaptureID) || !inSet(frame.Continuation, AllLocalKickstartPreviewContinuations) {
		return fail()
	}
	if frame.Kind == LocalKickstartPreviewOpened && (frame.Seq != 0 || len(frame.RequestID) == 0 || len(frame.RequestID) > 32 || !previewIDPattern.MatchString(frame.RequestID) || frame.Segment != nil || frame.Continuation != LocalKickstartPreviewContinuationMore) {
		return fail()
	}
	if frame.Kind == LocalKickstartPreviewSegmentEvent {
		if frame.RequestID != "" || (frame.Segment == nil && frame.Continuation == LocalKickstartPreviewContinuationMore) {
			return fail()
		}
	}
	if frame.Continuation == LocalKickstartPreviewContinuationCut {
		if !inSet(frame.CutReason, AllLocalKickstartPreviewCutReasons) {
			return fail()
		}
	} else if frame.CutReason != "" {
		return fail()
	}
	return nil
}

// MergeLocalKickstartPreview applies an ordered frame without mutating held data.
// As the producer-side merge, it verifies each turn's declared pre-bound byte
// length with TranscriptTurnByteLength: equality for untruncated turns and a
// lower bound for truncated turns. The TypeScript consumer checks declared
// lengths and their sum but does not duplicate the Go-only byte measure.
func MergeLocalKickstartPreview(held LocalKickstartPreviewState, frame LocalKickstartPreviewEvent) (LocalKickstartPreviewState, error) {
	fail := func() (LocalKickstartPreviewState, error) {
		return held, fmt.Errorf("kickstart preview merge failed: inconsistent frame or capture sequence; discard this capture and open the preview again")
	}
	if frame.Validate() != nil || held.Closed != nil {
		return fail()
	}
	if frame.Kind == LocalKickstartPreviewClosed {
		if held.Directory == nil || frame.SessionID != held.SessionID || frame.CaptureID != held.CaptureID || held.Seq == ^uint32(0) || frame.Seq != held.Seq+1 {
			return fail()
		}
		result := held
		result.Seq = frame.Seq
		result.Closed = &frame
		return result, nil
	}
	if frame.Directory.Segments == nil || len(frame.Header.Turns) != 0 || len(frame.Header.RetainedUnknown) != 0 || ValidateNativeMetadataRecords(frame.Header.NativeMetadata, nil) != nil {
		return fail()
	}
	d := frame.Directory
	if _, err := NewSessionID(string(d.SessionID)); err != nil {
		return fail()
	}
	if d.Complete != (frame.Continuation != LocalKickstartPreviewContinuationMore) || (frame.Continuation == LocalKickstartPreviewContinuationCut && !inSet(frame.CutReason, AllLocalKickstartPreviewCutReasons)) || (frame.Continuation != LocalKickstartPreviewContinuationCut && frame.CutReason != "") {
		return fail()
	}
	if d.LoadedSegments != len(d.Segments) || d.LoadedSegments > MaxTranscriptDirectoryRefs || d.LoadedTurns < 0 || !isDecimalRevision(d.Revision) || string(d.SessionID) != frame.Header.ID {
		return fail()
	}
	if d.Complete {
		if d.TotalSegments == nil || d.TotalTurns == nil || *d.TotalSegments != d.LoadedSegments || *d.TotalTurns != d.LoadedTurns || (d.LoadedConversationTurns == nil) != (d.TotalConversationTurns == nil) || (d.TotalConversationTurns != nil && *d.TotalConversationTurns != *d.LoadedConversationTurns) {
			return fail()
		}
	} else if d.TotalSegments != nil || d.TotalTurns != nil || d.TotalConversationTurns != nil {
		return fail()
	}
	conversationTurns := 0
	for i, ref := range d.Segments {
		if ref.Index != i || ref.FirstTurn < 0 || ref.LastTurn < ref.FirstTurn || (i > 0 && ref.FirstTurn <= d.Segments[i-1].LastTurn) || (ref.ConversationTurns == nil) != (d.LoadedConversationTurns == nil) {
			return fail()
		}
		if ref.ConversationTurns != nil {
			if ref.ConversationTurns.Before != conversationTurns || ref.ConversationTurns.Count < 0 {
				return fail()
			}
			conversationTurns += ref.ConversationTurns.Count
		}
	}
	if d.LoadedConversationTurns != nil && *d.LoadedConversationTurns != conversationTurns {
		return fail()
	}
	if held.Directory == nil {
		if frame.Kind != LocalKickstartPreviewOpened || frame.Seq != 0 || frame.Segment != nil || len(d.Segments) != 0 || d.Complete || frame.CaptureID == "" {
			return fail()
		}
	} else {
		if frame.Kind != LocalKickstartPreviewSegmentEvent || held.Directory.Complete || held.Seq == ^uint32(0) || frame.Seq != held.Seq+1 || frame.SessionID != held.SessionID || frame.CaptureID != held.CaptureID || d.Revision != held.Directory.Revision || d.SessionID != held.Directory.SessionID || frame.Header.ID != held.Header.ID {
			return fail()
		}
		if len(d.Segments) < len(held.Directory.Segments) || len(d.Segments) > len(held.Directory.Segments)+1 || !reflect.DeepEqual(d.Segments[:len(held.Directory.Segments)], held.Directory.Segments) {
			return fail()
		}
		if held.Header.Diagnostics != nil && held.Header.Diagnostics.Partial && (frame.Header.Diagnostics == nil || !frame.Header.Diagnostics.Partial) {
			return fail()
		}
	}
	segments := append([]LocalKickstartPreviewSegment(nil), held.Segments...)
	if frame.Segment != nil {
		s := frame.Segment
		if len(d.Segments) != len(segments)+1 || s.Index != len(segments) || s.Revision != d.Revision || s.Validate() != nil {
			return fail()
		}
		ref := d.Segments[s.Index]
		if ref.FirstTurn != s.FirstTurn || ref.LastTurn != s.LastTurn || ref.ByteLength != s.ByteLength {
			return fail()
		}
		if !d.Complete && !s.HasNewer {
			return fail()
		}
		if s.ByteLength > TranscriptSegmentBudgetBytes && (len(s.Turns) != 1 || !s.Turns[0].Truncated) {
			return fail()
		}
		targets := make([]TurnDetail, len(s.Turns))
		conversationCount := 0
		for i := range s.Turns {
			targets[i] = s.Turns[i].TurnDetail
			if IsConversationTurn(targets[i].Role, targets[i].EntryType, targets[i].Depth) {
				conversationCount++
			}
			measured := TranscriptTurnByteLength(targets[i])
			if measured > s.Turns[i].ByteLength || (!s.Turns[i].Truncated && measured != s.Turns[i].ByteLength) {
				return fail()
			}
		}
		if ref.ConversationTurns != nil && ref.ConversationTurns.Count != conversationCount {
			return fail()
		}
		if ValidateNativeMetadataRecords(s.NativeMetadata, targets) != nil {
			return fail()
		}
		segments = append(segments, *s)
	} else if len(d.Segments) != len(segments) || (held.Directory != nil && !d.Complete) {
		return fail()
	}
	count := 0
	detail := *frame.Header
	detail.Turns = nil
	detail.NativeMetadata = append([]NativeMetadataRecord(nil), frame.Header.NativeMetadata...)
	for _, s := range segments {
		count += len(s.Turns)
		for _, turn := range s.Turns {
			detail.Turns = append(detail.Turns, turn.TurnDetail)
		}
		detail.NativeMetadata = append(detail.NativeMetadata, s.NativeMetadata...)
		detail.RetainedUnknown = append(detail.RetainedUnknown, s.RetainedUnknown...)
	}
	if count != d.LoadedTurns || ValidateRetainedUnknown(detail) != nil || ValidateSessionDetailPayload(detail) != nil {
		return fail()
	}
	return LocalKickstartPreviewState{SessionID: frame.SessionID, CaptureID: frame.CaptureID, Seq: frame.Seq, Header: frame.Header, Directory: d, Segments: segments, Continuation: frame.Continuation}, nil
}
