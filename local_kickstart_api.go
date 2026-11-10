package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// Kickstart is the local setup session. The contract owns the routes, the
// live-topic message, and the closed tokens. Peasant owns which stops exist,
// which fields are visible, and what a commit writes. Nothing here is paged:
// the tree is the whole forest, and no payload carries a page, a cursor, or a
// limit.

func newKickstartToken[T ~string](raw, name string, all []T) (T, error) {
	value := T(raw)
	if !inSet(value, all) {
		return "", enumInputError(name, raw, stringValues(all))
	}
	return value, nil
}

func unmarshalKickstartToken[T ~string](data []byte, dest *T, parse func(string) (T, error), where string) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("kickstart decode failed at %s: the token is null; send one published string, or omit the field when it is optional", where)
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("kickstart decode failed at %s: %v; send a JSON string from the closed set", where, err)
	}
	parsed, err := parse(raw)
	if err != nil {
		return err
	}
	*dest = parsed
	return nil
}

func unmarshalStrict(data []byte, dest any, where string) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("kickstart decode failed at %s: %v; remove fields the contract does not declare", where, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("kickstart decode failed at %s: trailing JSON after the value; send one JSON value", where)
	}
	return nil
}

func requireNonEmpty(value, where, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("kickstart validation failed at %s: %s is empty; the page cannot address the value; emit a non-empty string", where, field)
	}
	return nil
}

func requireList[T any](items []T, where, field string) error {
	if items == nil {
		return fmt.Errorf("kickstart validation failed at %s: %s is null; emit [] when there are none", where, field)
	}
	return nil
}

// LocalKickstartScope is the answer set a session collects. full is the whole
// wizard. selection is the standalone selection page. The informal label
// "kickstart" is not a token.
type LocalKickstartScope string

const (
	LocalKickstartScopeFull      LocalKickstartScope = "full"
	LocalKickstartScopeSelection LocalKickstartScope = "selection"
)

// AllLocalKickstartScopes is the canonical scope menu.
var AllLocalKickstartScopes = []LocalKickstartScope{LocalKickstartScopeFull, LocalKickstartScopeSelection}

// NewLocalKickstartScope refuses a token outside the menu.
func NewLocalKickstartScope(raw string) (LocalKickstartScope, error) {
	return newKickstartToken(raw, "LocalKickstartScope", AllLocalKickstartScopes)
}

func (v LocalKickstartScope) IsValid() bool  { return inSet(v, AllLocalKickstartScopes) }
func (v LocalKickstartScope) String() string { return string(v) }
func (LocalKickstartScope) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Scope", "Answer set a kickstart session collects: the whole wizard, or the selection page", AllLocalKickstartScopes), nil
}
func (v *LocalKickstartScope) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartScope, "schema.LocalKickstartScope.UnmarshalJSON")
}

// LocalKickstartStopKind is the layout of one stop. The flow key is the
// identity; this token only chooses the layout.
type LocalKickstartStopKind string

const (
	LocalKickstartStopConnection LocalKickstartStopKind = "connection"
	LocalKickstartStopConfig     LocalKickstartStopKind = "config"
	LocalKickstartStopReview     LocalKickstartStopKind = "review"
	LocalKickstartStopIngest     LocalKickstartStopKind = "ingest"
)

// AllLocalKickstartStopKinds is the canonical stop-layout menu.
var AllLocalKickstartStopKinds = []LocalKickstartStopKind{LocalKickstartStopConnection, LocalKickstartStopConfig, LocalKickstartStopReview, LocalKickstartStopIngest}

// NewLocalKickstartStopKind refuses a token outside the menu.
func NewLocalKickstartStopKind(raw string) (LocalKickstartStopKind, error) {
	return newKickstartToken(raw, "LocalKickstartStopKind", AllLocalKickstartStopKinds)
}

func (v LocalKickstartStopKind) IsValid() bool  { return inSet(v, AllLocalKickstartStopKinds) }
func (v LocalKickstartStopKind) String() string { return string(v) }
func (LocalKickstartStopKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Stop Kind", "Layout of one kickstart stop", AllLocalKickstartStopKinds), nil
}
func (v *LocalKickstartStopKind) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartStopKind, "schema.LocalKickstartStopKind.UnmarshalJSON")
}

// LocalKickstartFieldKind is the control shape of one field. The server owns
// visibility, options, and validity; the client switches on this token only
// to choose a control.
type LocalKickstartFieldKind string

const (
	LocalKickstartFieldToggle LocalKickstartFieldKind = "toggle"
	LocalKickstartFieldChoice LocalKickstartFieldKind = "choice"
	LocalKickstartFieldAction LocalKickstartFieldKind = "action"
	LocalKickstartFieldTree   LocalKickstartFieldKind = "tree"
	LocalKickstartFieldNotice LocalKickstartFieldKind = "notice"
)

// AllLocalKickstartFieldKinds is the canonical field-control menu.
var AllLocalKickstartFieldKinds = []LocalKickstartFieldKind{LocalKickstartFieldToggle, LocalKickstartFieldChoice, LocalKickstartFieldAction, LocalKickstartFieldTree, LocalKickstartFieldNotice}

// NewLocalKickstartFieldKind refuses a token outside the menu.
func NewLocalKickstartFieldKind(raw string) (LocalKickstartFieldKind, error) {
	return newKickstartToken(raw, "LocalKickstartFieldKind", AllLocalKickstartFieldKinds)
}

func (v LocalKickstartFieldKind) IsValid() bool  { return inSet(v, AllLocalKickstartFieldKinds) }
func (v LocalKickstartFieldKind) String() string { return string(v) }
func (LocalKickstartFieldKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Field Kind", "Control shape of one kickstart field", AllLocalKickstartFieldKinds), nil
}
func (v *LocalKickstartFieldKind) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartFieldKind, "schema.LocalKickstartFieldKind.UnmarshalJSON")
}

// LocalKickstartLineKind is one typed preview or example line. A future guide
// example reuses this set; it does not invent a second line type.
type LocalKickstartLineKind string

const (
	LocalKickstartLineText   LocalKickstartLineKind = "text"
	LocalKickstartLineLabel  LocalKickstartLineKind = "label"
	LocalKickstartLineBefore LocalKickstartLineKind = "before"
	LocalKickstartLineAfter  LocalKickstartLineKind = "after"
	LocalKickstartLineSpacer LocalKickstartLineKind = "spacer"
)

// AllLocalKickstartLineKinds is the canonical line menu.
var AllLocalKickstartLineKinds = []LocalKickstartLineKind{LocalKickstartLineText, LocalKickstartLineLabel, LocalKickstartLineBefore, LocalKickstartLineAfter, LocalKickstartLineSpacer}

// NewLocalKickstartLineKind refuses a token outside the menu.
func NewLocalKickstartLineKind(raw string) (LocalKickstartLineKind, error) {
	return newKickstartToken(raw, "LocalKickstartLineKind", AllLocalKickstartLineKinds)
}

func (v LocalKickstartLineKind) IsValid() bool  { return inSet(v, AllLocalKickstartLineKinds) }
func (v LocalKickstartLineKind) String() string { return string(v) }
func (LocalKickstartLineKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Line Kind", "Typed preview or example line", AllLocalKickstartLineKinds), nil
}
func (v *LocalKickstartLineKind) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartLineKind, "schema.LocalKickstartLineKind.UnmarshalJSON")
}

// LocalKickstartNodeKind is one row in the selection forest.
type LocalKickstartNodeKind string

const (
	LocalKickstartNodeProject LocalKickstartNodeKind = "project"
	LocalKickstartNodeGroup   LocalKickstartNodeKind = "group"
	LocalKickstartNodeBranch  LocalKickstartNodeKind = "branch"
	LocalKickstartNodeSession LocalKickstartNodeKind = "session"
)

// AllLocalKickstartNodeKinds is the canonical node-kind menu.
var AllLocalKickstartNodeKinds = []LocalKickstartNodeKind{LocalKickstartNodeProject, LocalKickstartNodeGroup, LocalKickstartNodeBranch, LocalKickstartNodeSession}

// NewLocalKickstartNodeKind refuses a token outside the menu.
func NewLocalKickstartNodeKind(raw string) (LocalKickstartNodeKind, error) {
	return newKickstartToken(raw, "LocalKickstartNodeKind", AllLocalKickstartNodeKinds)
}

func (v LocalKickstartNodeKind) IsValid() bool  { return inSet(v, AllLocalKickstartNodeKinds) }
func (v LocalKickstartNodeKind) String() string { return string(v) }
func (LocalKickstartNodeKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Node Kind", "Row kind in the selection forest", AllLocalKickstartNodeKinds), nil
}
func (v *LocalKickstartNodeKind) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartNodeKind, "schema.LocalKickstartNodeKind.UnmarshalJSON")
}

// LocalKickstartNodeState is the server-computed selection of one row and its
// descendants. locked is not a state; it is a separate boolean.
type LocalKickstartNodeState string

const (
	LocalKickstartNodeNone LocalKickstartNodeState = "none"
	LocalKickstartNodeSome LocalKickstartNodeState = "some"
	LocalKickstartNodeAll  LocalKickstartNodeState = "all"
)

// AllLocalKickstartNodeStates is the canonical node-state menu.
var AllLocalKickstartNodeStates = []LocalKickstartNodeState{LocalKickstartNodeNone, LocalKickstartNodeSome, LocalKickstartNodeAll}

// NewLocalKickstartNodeState refuses a token outside the menu.
func NewLocalKickstartNodeState(raw string) (LocalKickstartNodeState, error) {
	return newKickstartToken(raw, "LocalKickstartNodeState", AllLocalKickstartNodeStates)
}

func (v LocalKickstartNodeState) IsValid() bool  { return inSet(v, AllLocalKickstartNodeStates) }
func (v LocalKickstartNodeState) String() string { return string(v) }
func (LocalKickstartNodeState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Node State", "Server-computed selection of one forest row: none, some, or all. locked is a boolean, not a state", AllLocalKickstartNodeStates), nil
}
func (v *LocalKickstartNodeState) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartNodeState, "schema.LocalKickstartNodeState.UnmarshalJSON")
}

// LocalKickstartScanState is the discovery job. Progress is pushed; this token
// is the phase the push reports.
type LocalKickstartScanState string

const (
	LocalKickstartScanIdle    LocalKickstartScanState = "idle"
	LocalKickstartScanRunning LocalKickstartScanState = "running"
	LocalKickstartScanDone    LocalKickstartScanState = "done"
	LocalKickstartScanError   LocalKickstartScanState = "error"
)

// AllLocalKickstartScanStates is the canonical scan-state menu.
var AllLocalKickstartScanStates = []LocalKickstartScanState{LocalKickstartScanIdle, LocalKickstartScanRunning, LocalKickstartScanDone, LocalKickstartScanError}

// NewLocalKickstartScanState refuses a token outside the menu.
func NewLocalKickstartScanState(raw string) (LocalKickstartScanState, error) {
	return newKickstartToken(raw, "LocalKickstartScanState", AllLocalKickstartScanStates)
}

func (v LocalKickstartScanState) IsValid() bool  { return inSet(v, AllLocalKickstartScanStates) }
func (v LocalKickstartScanState) String() string { return string(v) }
func (LocalKickstartScanState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Scan State", "Discovery job state pushed on the kickstart topic", AllLocalKickstartScanStates), nil
}
func (v *LocalKickstartScanState) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartScanState, "schema.LocalKickstartScanState.UnmarshalJSON")
}

// LocalKickstartIngestState is the import started at the end of a commit.
// busy means an import is already running and the finish panel offers import
// now. The ingest status route is only the reconnect fallback.
type LocalKickstartIngestState string

const (
	LocalKickstartIngestIdle    LocalKickstartIngestState = "idle"
	LocalKickstartIngestRunning LocalKickstartIngestState = "running"
	LocalKickstartIngestBusy    LocalKickstartIngestState = "busy"
	LocalKickstartIngestFailed  LocalKickstartIngestState = "failed"
)

// AllLocalKickstartIngestStates is the canonical ingest-state menu.
var AllLocalKickstartIngestStates = []LocalKickstartIngestState{LocalKickstartIngestIdle, LocalKickstartIngestRunning, LocalKickstartIngestBusy, LocalKickstartIngestFailed}

// NewLocalKickstartIngestState refuses a token outside the menu.
func NewLocalKickstartIngestState(raw string) (LocalKickstartIngestState, error) {
	return newKickstartToken(raw, "LocalKickstartIngestState", AllLocalKickstartIngestStates)
}

func (v LocalKickstartIngestState) IsValid() bool  { return inSet(v, AllLocalKickstartIngestStates) }
func (v LocalKickstartIngestState) String() string { return string(v) }
func (LocalKickstartIngestState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Ingest State", "Import state pushed on the kickstart topic", AllLocalKickstartIngestStates), nil
}
func (v *LocalKickstartIngestState) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartIngestState, "schema.LocalKickstartIngestState.UnmarshalJSON")
}

// LocalKickstartResolution chooses which side of one conflict row to keep.
// mine is the draft. theirs is the file as it is now. The words keep-mine and
// take-theirs are page copy, not wire tokens.
type LocalKickstartResolution string

const (
	LocalKickstartResolutionMine   LocalKickstartResolution = "mine"
	LocalKickstartResolutionTheirs LocalKickstartResolution = "theirs"
)

// AllLocalKickstartResolutions is the canonical conflict-resolution menu.
var AllLocalKickstartResolutions = []LocalKickstartResolution{LocalKickstartResolutionMine, LocalKickstartResolutionTheirs}

// NewLocalKickstartResolution refuses a token outside the menu.
func NewLocalKickstartResolution(raw string) (LocalKickstartResolution, error) {
	return newKickstartToken(raw, "LocalKickstartResolution", AllLocalKickstartResolutions)
}

func (v LocalKickstartResolution) IsValid() bool  { return inSet(v, AllLocalKickstartResolutions) }
func (v LocalKickstartResolution) String() string { return string(v) }
func (LocalKickstartResolution) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Resolution", "Which side of one conflict row to keep: mine is the draft, theirs is the file", AllLocalKickstartResolutions), nil
}
func (v *LocalKickstartResolution) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartResolution, "schema.LocalKickstartResolution.UnmarshalJSON")
}

// LocalKickstartNextStep is one finish action. The web list is these five.
// The terminal list is the first four; that filter belongs to the terminal,
// not to a second wire set.
type LocalKickstartNextStep string

const (
	LocalKickstartNextStepConfig       LocalKickstartNextStep = "config"
	LocalKickstartNextStepWebStart     LocalKickstartNextStep = "web-start"
	LocalKickstartNextStepVillageLogin LocalKickstartNextStep = "village-login"
	LocalKickstartNextStepVillagePush  LocalKickstartNextStep = "village-push"
	LocalKickstartNextStepAutoPublish  LocalKickstartNextStep = "auto-publish-settings"
)

// AllLocalKickstartNextSteps is the canonical finish-action menu.
var AllLocalKickstartNextSteps = []LocalKickstartNextStep{
	LocalKickstartNextStepConfig,
	LocalKickstartNextStepWebStart,
	LocalKickstartNextStepVillageLogin,
	LocalKickstartNextStepVillagePush,
	LocalKickstartNextStepAutoPublish,
}

// NewLocalKickstartNextStep refuses a token outside the menu.
func NewLocalKickstartNextStep(raw string) (LocalKickstartNextStep, error) {
	return newKickstartToken(raw, "LocalKickstartNextStep", AllLocalKickstartNextSteps)
}

func (v LocalKickstartNextStep) IsValid() bool  { return inSet(v, AllLocalKickstartNextSteps) }
func (v LocalKickstartNextStep) String() string { return string(v) }
func (LocalKickstartNextStep) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Next Step", "One finish action. auto-publish-settings links to the settings section; it does not install a hook", AllLocalKickstartNextSteps), nil
}
func (v *LocalKickstartNextStep) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartNextStep, "schema.LocalKickstartNextStep.UnmarshalJSON")
}

// LocalKickstartDiffLineType is one line in the structured config diff.
// The tokens are the ones the review diff renders: add, del, and ctx.
type LocalKickstartDiffLineType string

const (
	LocalKickstartDiffAdd LocalKickstartDiffLineType = "add"
	LocalKickstartDiffDel LocalKickstartDiffLineType = "del"
	LocalKickstartDiffCtx LocalKickstartDiffLineType = "ctx"
)

// AllLocalKickstartDiffLineTypes is the canonical diff-line menu.
var AllLocalKickstartDiffLineTypes = []LocalKickstartDiffLineType{LocalKickstartDiffAdd, LocalKickstartDiffDel, LocalKickstartDiffCtx}

// NewLocalKickstartDiffLineType refuses a token outside the menu.
func NewLocalKickstartDiffLineType(raw string) (LocalKickstartDiffLineType, error) {
	return newKickstartToken(raw, "LocalKickstartDiffLineType", AllLocalKickstartDiffLineTypes)
}

func (v LocalKickstartDiffLineType) IsValid() bool  { return inSet(v, AllLocalKickstartDiffLineTypes) }
func (v LocalKickstartDiffLineType) String() string { return string(v) }
func (LocalKickstartDiffLineType) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Diff Line Type", "Structured config-diff line: add, del, or ctx", AllLocalKickstartDiffLineTypes), nil
}
func (v *LocalKickstartDiffLineType) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartDiffLineType, "schema.LocalKickstartDiffLineType.UnmarshalJSON")
}

// LocalKickstartEventKind is the live-topic discriminator. session, answers,
// and conflicts each carry the full view. They stay separate so a consumer
// can tell which mutation produced the snapshot. scan and ingest carry only
// their own object. closed ends the session.
type LocalKickstartEventKind string

const (
	LocalKickstartEventSession   LocalKickstartEventKind = "session"
	LocalKickstartEventAnswers   LocalKickstartEventKind = "answers"
	LocalKickstartEventConflicts LocalKickstartEventKind = "conflicts"
	LocalKickstartEventScan      LocalKickstartEventKind = "scan"
	LocalKickstartEventIngest    LocalKickstartEventKind = "ingest"
	LocalKickstartEventClosed    LocalKickstartEventKind = "closed"
)

// AllLocalKickstartEventKinds is the canonical event-kind menu.
var AllLocalKickstartEventKinds = []LocalKickstartEventKind{
	LocalKickstartEventSession,
	LocalKickstartEventAnswers,
	LocalKickstartEventConflicts,
	LocalKickstartEventScan,
	LocalKickstartEventIngest,
	LocalKickstartEventClosed,
}

// NewLocalKickstartEventKind refuses a token outside the menu.
func NewLocalKickstartEventKind(raw string) (LocalKickstartEventKind, error) {
	return newKickstartToken(raw, "LocalKickstartEventKind", AllLocalKickstartEventKinds)
}

func (v LocalKickstartEventKind) IsValid() bool  { return inSet(v, AllLocalKickstartEventKinds) }
func (v LocalKickstartEventKind) String() string { return string(v) }
func (LocalKickstartEventKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Event Kind", "Live kickstart topic discriminator. One event carries one session id", AllLocalKickstartEventKinds), nil
}
func (v *LocalKickstartEventKind) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartEventKind, "schema.LocalKickstartEventKind.UnmarshalJSON")
}

// LocalKickstartCloseReason is why a session ended. committed moves the page
// to the finish panel. expired offers to reopen. There is no replaced reason.
type LocalKickstartCloseReason string

const (
	LocalKickstartClosedCommitted LocalKickstartCloseReason = "committed"
	LocalKickstartClosedExpired   LocalKickstartCloseReason = "expired"
)

// AllLocalKickstartCloseReasons is the canonical close-reason menu.
var AllLocalKickstartCloseReasons = []LocalKickstartCloseReason{LocalKickstartClosedCommitted, LocalKickstartClosedExpired}

// NewLocalKickstartCloseReason refuses a token outside the menu.
func NewLocalKickstartCloseReason(raw string) (LocalKickstartCloseReason, error) {
	return newKickstartToken(raw, "LocalKickstartCloseReason", AllLocalKickstartCloseReasons)
}

func (v LocalKickstartCloseReason) IsValid() bool  { return inSet(v, AllLocalKickstartCloseReasons) }
func (v LocalKickstartCloseReason) String() string { return string(v) }
func (LocalKickstartCloseReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Close Reason", "Why a kickstart session ended: committed or expired", AllLocalKickstartCloseReasons), nil
}
func (v *LocalKickstartCloseReason) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartCloseReason, "schema.LocalKickstartCloseReason.UnmarshalJSON")
}

// LocalKickstartConnectAnswer is the boundary offer's accept or skip. It is
// not a stop field and it never enters the configuration file.
type LocalKickstartConnectAnswer string

const (
	LocalKickstartConnectAccept LocalKickstartConnectAnswer = "accept"
	LocalKickstartConnectSkip   LocalKickstartConnectAnswer = "skip"
)

// AllLocalKickstartConnectAnswers is the canonical connect-answer menu.
var AllLocalKickstartConnectAnswers = []LocalKickstartConnectAnswer{LocalKickstartConnectAccept, LocalKickstartConnectSkip}

// NewLocalKickstartConnectAnswer refuses a token outside the menu.
func NewLocalKickstartConnectAnswer(raw string) (LocalKickstartConnectAnswer, error) {
	return newKickstartToken(raw, "LocalKickstartConnectAnswer", AllLocalKickstartConnectAnswers)
}

func (v LocalKickstartConnectAnswer) IsValid() bool  { return inSet(v, AllLocalKickstartConnectAnswers) }
func (v LocalKickstartConnectAnswer) String() string { return string(v) }
func (LocalKickstartConnectAnswer) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Kickstart Connect Answer", "Boundary offer answer: accept or skip. accept requires village to be connected", AllLocalKickstartConnectAnswers), nil
}
func (v *LocalKickstartConnectAnswer) UnmarshalJSON(data []byte) error {
	return unmarshalKickstartToken(data, v, NewLocalKickstartConnectAnswer, "schema.LocalKickstartConnectAnswer.UnmarshalJSON")
}

// LocalKickstartStatus is the first-run redirect hook. configured is the only
// status flag. activeSessionId, when present, is the one full-scope session
// id, not a directory and not a per-scope map.
type LocalKickstartStatus struct {
	Configured      bool   `json:"configured"`
	ActiveSessionID string `json:"activeSessionId,omitempty"`
}

// UnmarshalJSON refuses a field other than configured and activeSessionId.
func (s *LocalKickstartStatus) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartStatus
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartStatus.UnmarshalJSON"); err != nil {
		return err
	}
	*s = LocalKickstartStatus(decoded)
	return nil
}

// Validate checks the redirect hook. An empty active session id is absence,
// not a third status.
func (s LocalKickstartStatus) Validate() error {
	if strings.TrimSpace(s.ActiveSessionID) != s.ActiveSessionID {
		return fmt.Errorf("kickstart status validation failed at schema.LocalKickstartStatus.Validate: activeSessionId has surrounding space; emit the session id, or omit the field when no full-scope session is open")
	}
	return nil
}

// LocalKickstartOpenRequest opens or joins the session for one scope.
type LocalKickstartOpenRequest struct {
	Scope LocalKickstartScope `json:"scope"`
}

// UnmarshalJSON refuses an unknown field on the open body.
func (r *LocalKickstartOpenRequest) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartOpenRequest
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartOpenRequest.UnmarshalJSON"); err != nil {
		return err
	}
	*r = LocalKickstartOpenRequest(decoded)
	return nil
}

// Validate checks that the body names a published scope.
func (r LocalKickstartOpenRequest) Validate() error {
	if !r.Scope.IsValid() {
		return fmt.Errorf("kickstart open validation failed at schema.LocalKickstartOpenRequest.Validate: scope %q is outside the closed set; send full or selection", r.Scope)
	}
	return nil
}

// LocalKickstartGuide is the stop's intro and hints. An example reuses
// LocalKickstartLine; this struct does not grow an example field.
type LocalKickstartGuide struct {
	Intro string   `json:"intro,omitempty"`
	Hints []string `json:"hints,omitempty" nullable:"false"`
}

// UnmarshalJSON refuses an unknown guide field.
func (g *LocalKickstartGuide) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartGuide
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartGuide.UnmarshalJSON"); err != nil {
		return err
	}
	*g = LocalKickstartGuide(decoded)
	return nil
}

// Validate checks that a present hint list is not null.
func (g LocalKickstartGuide) Validate() error {
	if g.Hints == nil {
		return nil
	}
	for i, hint := range g.Hints {
		if strings.TrimSpace(hint) == "" {
			return fmt.Errorf("kickstart guide validation failed at schema.LocalKickstartGuide.Validate: hint %d is empty; omit the hint or emit its text", i)
		}
	}
	return nil
}

// LocalKickstartOption is one choice. Value is the stable wire token.
type LocalKickstartOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// UnmarshalJSON refuses an unknown option field.
func (o *LocalKickstartOption) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartOption
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartOption.UnmarshalJSON"); err != nil {
		return err
	}
	*o = LocalKickstartOption(decoded)
	return nil
}

// Validate checks the token and the label. An empty token is allowed when
// that is a meaningful choice; the label is what the page shows.
func (o LocalKickstartOption) Validate() error {
	return requireNonEmpty(o.Label, "schema.LocalKickstartOption.Validate", "label")
}

// LocalKickstartLine is one typed preview or example line.
type LocalKickstartLine struct {
	Kind LocalKickstartLineKind `json:"kind"`
	Text string                 `json:"text"`
}

// UnmarshalJSON refuses an unknown line field.
func (l *LocalKickstartLine) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartLine
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartLine.UnmarshalJSON"); err != nil {
		return err
	}
	*l = LocalKickstartLine(decoded)
	return nil
}

// Validate checks the line kind.
func (l LocalKickstartLine) Validate() error {
	if !l.Kind.IsValid() {
		return fmt.Errorf("kickstart line validation failed at schema.LocalKickstartLine.Validate: kind %q is outside the closed set; emit a member of schema.AllLocalKickstartLineKinds", l.Kind)
	}
	return nil
}

// LocalKickstartSummary is the review stop's derived explanation. The client
// renders these rows and composes no sentence.
type LocalKickstartSummary struct {
	Values  []string `json:"values" nullable:"false"`
	Effects []string `json:"effects" nullable:"false"`
}

// UnmarshalJSON refuses an unknown summary field.
func (s *LocalKickstartSummary) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartSummary
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartSummary.UnmarshalJSON"); err != nil {
		return err
	}
	*s = LocalKickstartSummary(decoded)
	return nil
}

// Validate checks both lists are present.
func (s LocalKickstartSummary) Validate() error {
	if err := requireList(s.Values, "schema.LocalKickstartSummary.Validate", "values"); err != nil {
		return err
	}
	return requireList(s.Effects, "schema.LocalKickstartSummary.Validate", "effects")
}

// LocalKickstartDiffLine is one structured diff line. oldNo and newNo are
// omitted on a line that has no number on that side.
type LocalKickstartDiffLine struct {
	Type  LocalKickstartDiffLineType `json:"type"`
	OldNo int                        `json:"oldNo,omitempty" minimum:"0"`
	NewNo int                        `json:"newNo,omitempty" minimum:"0"`
	Text  string                     `json:"text"`
}

// UnmarshalJSON refuses an unknown diff-line field.
func (l *LocalKickstartDiffLine) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartDiffLine
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartDiffLine.UnmarshalJSON"); err != nil {
		return err
	}
	*l = LocalKickstartDiffLine(decoded)
	return nil
}

// Validate checks the line type and refuses negative numbers; zero denotes an omitted side.
func (l LocalKickstartDiffLine) Validate() error {
	if !l.Type.IsValid() {
		return fmt.Errorf("kickstart diff line validation failed at schema.LocalKickstartDiffLine.Validate: type %q is outside the closed set; emit add, del, or ctx", l.Type)
	}
	if l.OldNo < 0 || l.NewNo < 0 {
		return fmt.Errorf("kickstart diff line validation failed at schema.LocalKickstartDiffLine.Validate: a line number is negative; omit a side that has no number, or emit a positive number")
	}
	return nil
}

// LocalKickstartHunk is one contiguous change. Header is the verbatim range
// text, for example "-18,7 +18,10".
type LocalKickstartHunk struct {
	Header string                   `json:"header" minLength:"1"`
	Lines  []LocalKickstartDiffLine `json:"lines" nullable:"false"`
}

// UnmarshalJSON refuses an unknown hunk field.
func (h *LocalKickstartHunk) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartHunk
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartHunk.UnmarshalJSON"); err != nil {
		return err
	}
	*h = LocalKickstartHunk(decoded)
	return nil
}

// Validate checks the header and every line.
func (h LocalKickstartHunk) Validate() error {
	if err := requireNonEmpty(h.Header, "schema.LocalKickstartHunk.Validate", "header"); err != nil {
		return err
	}
	if err := requireList(h.Lines, "schema.LocalKickstartHunk.Validate", "lines"); err != nil {
		return err
	}
	for i := range h.Lines {
		if err := h.Lines[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartDiff is the structured before-to-after document diff. The
// client maps hunks onto the review diff and renders nothing itself.
// Truncated means the server omitted hunks past its bound; the per-field
// conflict rows are never truncated.
type LocalKickstartDiff struct {
	File      string               `json:"file" minLength:"1"`
	Hunks     []LocalKickstartHunk `json:"hunks" nullable:"false"`
	Truncated bool                 `json:"truncated,omitempty"`
}

// UnmarshalJSON refuses an unknown diff field.
func (d *LocalKickstartDiff) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartDiff
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartDiff.UnmarshalJSON"); err != nil {
		return err
	}
	*d = LocalKickstartDiff(decoded)
	return nil
}

// Validate checks the file label and every hunk.
func (d LocalKickstartDiff) Validate() error {
	if err := requireNonEmpty(d.File, "schema.LocalKickstartDiff.Validate", "file"); err != nil {
		return err
	}
	if err := requireList(d.Hunks, "schema.LocalKickstartDiff.Validate", "hunks"); err != nil {
		return err
	}
	for i := range d.Hunks {
		if err := d.Hunks[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartConnect is the boundary offer, already evaluated. The client
// renders it on the stop after afterStop and never recomputes the condition.
type LocalKickstartConnect struct {
	Offered   bool   `json:"offered"`
	AfterStop string `json:"afterStop,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
	Skip      string `json:"skip,omitempty"`
}

// UnmarshalJSON refuses an unknown connect field.
func (c *LocalKickstartConnect) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartConnect
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartConnect.UnmarshalJSON"); err != nil {
		return err
	}
	*c = LocalKickstartConnect(decoded)
	return nil
}

// Validate checks that an offered connect names the stop it follows.
func (c LocalKickstartConnect) Validate() error {
	if c.Offered && strings.TrimSpace(c.AfterStop) == "" {
		return fmt.Errorf("kickstart connect validation failed at schema.LocalKickstartConnect.Validate: offered is true and afterStop is empty; name the stop the offer follows, or set offered false")
	}
	return nil
}

// LocalKickstartField is one answerable control. Value reuses
// LocalSettingValue. Refusal reuses LocalSettingRefusal. Peasant owns the key
// set; an unknown key is a server refusal, not a second menu here.
type LocalKickstartField struct {
	Key         string                  `json:"key" minLength:"1"`
	Kind        LocalKickstartFieldKind `json:"kind"`
	Label       string                  `json:"label" minLength:"1"`
	Description string                  `json:"description,omitempty"`
	Value       LocalSettingValue       `json:"value"`
	Options     []LocalKickstartOption  `json:"options,omitempty" nullable:"false"`
	Example     []LocalKickstartLine    `json:"example,omitempty" nullable:"false"`
	Refusal     *LocalSettingRefusal    `json:"refusal,omitempty"`
}

// UnmarshalJSON refuses an unknown field property.
func (f *LocalKickstartField) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartField
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartField.UnmarshalJSON"); err != nil {
		return err
	}
	*f = LocalKickstartField(decoded)
	return nil
}

// Validate checks the control, its value, and any options or example lines.
func (f LocalKickstartField) Validate() error {
	if err := requireNonEmpty(f.Key, "schema.LocalKickstartField.Validate", "key"); err != nil {
		return err
	}
	if !f.Kind.IsValid() {
		return fmt.Errorf("kickstart field validation failed for %q at schema.LocalKickstartField.Validate: kind %q is outside the closed set; emit a member of schema.AllLocalKickstartFieldKinds", f.Key, f.Kind)
	}
	if err := requireNonEmpty(f.Label, "schema.LocalKickstartField.Validate", "label"); err != nil {
		return err
	}
	if len(bytes.TrimSpace(f.Value)) == 0 {
		return fmt.Errorf("kickstart field validation failed for %q at schema.LocalKickstartField.Validate: value is missing; emit the JSON value, or null when the key is unset", f.Key)
	}
	for i := range f.Options {
		if err := f.Options[i].Validate(); err != nil {
			return err
		}
	}
	for i := range f.Example {
		if err := f.Example[i].Validate(); err != nil {
			return err
		}
	}
	if f.Refusal != nil {
		if err := f.Refusal.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartStop is one step. HiddenBecause is a diagnostic token; the
// review copy is Summary. The connect offer is not a field of the stop.
type LocalKickstartStop struct {
	Key           string                 `json:"key" minLength:"1"`
	Kind          LocalKickstartStopKind `json:"kind"`
	Title         string                 `json:"title" minLength:"1"`
	Visible       bool                   `json:"visible"`
	HiddenBecause string                 `json:"hiddenBecause,omitempty"`
	Guide         LocalKickstartGuide    `json:"guide"`
	Fields        []LocalKickstartField  `json:"fields" nullable:"false"`
	Summary       *LocalKickstartSummary `json:"summary,omitempty"`
	Diff          *LocalKickstartDiff    `json:"diff,omitempty"`
}

// UnmarshalJSON refuses an unknown stop field.
func (s *LocalKickstartStop) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartStop
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartStop.UnmarshalJSON"); err != nil {
		return err
	}
	*s = LocalKickstartStop(decoded)
	return nil
}

// Validate checks the stop and its fields. A hidden stop may still be sent
// so the client does not decide visibility.
func (s LocalKickstartStop) Validate() error {
	if err := requireNonEmpty(s.Key, "schema.LocalKickstartStop.Validate", "key"); err != nil {
		return err
	}
	if !s.Kind.IsValid() {
		return fmt.Errorf("kickstart stop validation failed for %q at schema.LocalKickstartStop.Validate: kind %q is outside the closed set; emit a member of schema.AllLocalKickstartStopKinds", s.Key, s.Kind)
	}
	if err := requireNonEmpty(s.Title, "schema.LocalKickstartStop.Validate", "title"); err != nil {
		return err
	}
	if err := s.Guide.Validate(); err != nil {
		return err
	}
	if err := requireList(s.Fields, "schema.LocalKickstartStop.Validate", "fields"); err != nil {
		return err
	}
	for i := range s.Fields {
		if err := s.Fields[i].Validate(); err != nil {
			return err
		}
	}
	if s.Summary != nil {
		if err := s.Summary.Validate(); err != nil {
			return err
		}
	}
	if s.Diff != nil {
		if err := s.Diff.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartConflict is one answered field whose file value moved. mine
// is the draft. theirs is the file. The page offers those two sides and no
// third.
type LocalKickstartConflict struct {
	Field  string            `json:"field" minLength:"1"`
	Stop   string            `json:"stop" minLength:"1"`
	Label  string            `json:"label" minLength:"1"`
	Mine   LocalSettingValue `json:"mine"`
	Theirs LocalSettingValue `json:"theirs"`
}

// UnmarshalJSON refuses an unknown conflict field.
func (c *LocalKickstartConflict) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartConflict
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartConflict.UnmarshalJSON"); err != nil {
		return err
	}
	*c = LocalKickstartConflict(decoded)
	return nil
}

// Validate checks the row names both sides.
func (c LocalKickstartConflict) Validate() error {
	if err := requireNonEmpty(c.Field, "schema.LocalKickstartConflict.Validate", "field"); err != nil {
		return err
	}
	if err := requireNonEmpty(c.Stop, "schema.LocalKickstartConflict.Validate", "stop"); err != nil {
		return err
	}
	if err := requireNonEmpty(c.Label, "schema.LocalKickstartConflict.Validate", "label"); err != nil {
		return err
	}
	if len(bytes.TrimSpace(c.Mine)) == 0 || len(bytes.TrimSpace(c.Theirs)) == 0 {
		return fmt.Errorf("kickstart conflict validation failed for %q at schema.LocalKickstartConflict.Validate: mine or theirs is missing; emit both JSON values, using null only when that side is unset", c.Field)
	}
	return nil
}

// localKickstartDroppedLabelsMax is the maximum number of vanished choices named in a scan.
// Keep the maxItems tag on DroppedLabels in sync.
const localKickstartDroppedLabelsMax = 3

// LocalKickstartScan is the discovery job. DroppedLabels names at most three
// vanished choices; the count is DroppedSelections. Phase is server prose, not
// a closed set. The forest is not in this object.
type LocalKickstartScan struct {
	State             LocalKickstartScanState `json:"state"`
	Phase             string                  `json:"phase,omitempty"`
	Done              int                     `json:"done" minimum:"0"`
	Total             int                     `json:"total" minimum:"0"`
	Error             string                  `json:"error,omitempty"`
	LockedNodes       int                     `json:"lockedNodes,omitempty" minimum:"0"`
	DroppedSelections int                     `json:"droppedSelections,omitempty" minimum:"0"`
	DroppedLabels     []string                `json:"droppedLabels,omitempty" nullable:"false" maxItems:"3"`
}

// UnmarshalJSON refuses an unknown scan field.
func (s *LocalKickstartScan) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartScan
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartScan.UnmarshalJSON"); err != nil {
		return err
	}
	*s = LocalKickstartScan(decoded)
	return nil
}

// Validate checks the state and the vanished-choice bound.
func (s LocalKickstartScan) Validate() error {
	if !s.State.IsValid() {
		return fmt.Errorf("kickstart scan validation failed at schema.LocalKickstartScan.Validate: state %q is outside the closed set; emit a member of schema.AllLocalKickstartScanStates", s.State)
	}
	if s.Done < 0 || s.Total < 0 || s.LockedNodes < 0 || s.DroppedSelections < 0 {
		return fmt.Errorf("kickstart scan validation failed at schema.LocalKickstartScan.Validate: a count is negative; emit zero or a positive count")
	}
	if len(s.DroppedLabels) > localKickstartDroppedLabelsMax {
		return fmt.Errorf("kickstart scan validation failed at schema.LocalKickstartScan.Validate: droppedLabels has %d entries; name at most %d and put the rest in droppedSelections", len(s.DroppedLabels), localKickstartDroppedLabelsMax)
	}
	return nil
}

// LocalKickstartVillage is the sign-in fact the server already evaluated.
type LocalKickstartVillage struct {
	Connected bool   `json:"connected"`
	URL       string `json:"url,omitempty"`
	// UserID is an optional opaque account identifier, even when connected.
	UserID string `json:"userId,omitempty" description:"Optional opaque account identifier; may be absent even when connected"`
}

// UnmarshalJSON refuses an unknown village field.
func (v *LocalKickstartVillage) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartVillage
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartVillage.UnmarshalJSON"); err != nil {
		return err
	}
	*v = LocalKickstartVillage(decoded)
	return nil
}

// Validate accepts an optional account identifier, matching SyncAuthResponse.
func (v LocalKickstartVillage) Validate() error { return nil }

// LocalKickstartIngest is the import progress. New and Updated are counts of
// sessions the import has written, not a page.
type LocalKickstartIngest struct {
	State   LocalKickstartIngestState `json:"state"`
	Note    string                    `json:"note,omitempty"`
	Error   string                    `json:"error,omitempty"`
	New     int                       `json:"new" minimum:"0"`
	Updated int                       `json:"updated" minimum:"0"`
}

// UnmarshalJSON refuses an unknown ingest field.
func (n *LocalKickstartIngest) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartIngest
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartIngest.UnmarshalJSON"); err != nil {
		return err
	}
	*n = LocalKickstartIngest(decoded)
	return nil
}

// Validate checks the state and the counts.
func (n LocalKickstartIngest) Validate() error {
	if !n.State.IsValid() {
		return fmt.Errorf("kickstart ingest validation failed at schema.LocalKickstartIngest.Validate: state %q is outside the closed set; emit a member of schema.AllLocalKickstartIngestStates", n.State)
	}
	if n.New < 0 || n.Updated < 0 {
		return fmt.Errorf("kickstart ingest validation failed at schema.LocalKickstartIngest.Validate: a count is negative; emit zero or a positive count")
	}
	return nil
}

// LocalKickstartSession is the reconnect snapshot and the live view. The
// forest is not here; the page loads it from the tree route.
type LocalKickstartSession struct {
	ID        string                   `json:"id" minLength:"1"`
	Scope     LocalKickstartScope      `json:"scope"`
	Revision  uint64                   `json:"revision" minimum:"0"`
	OpenedAt  string                   `json:"openedAt" minLength:"1"`
	Stops     []LocalKickstartStop     `json:"stops" nullable:"false"`
	Conflicts []LocalKickstartConflict `json:"conflicts" nullable:"false"`
	Connect   LocalKickstartConnect    `json:"connect"`
	Scan      LocalKickstartScan       `json:"scan"`
	Village   LocalKickstartVillage    `json:"village"`
}

// UnmarshalJSON refuses an unknown session field, including a tree, a page,
// or a reset control.
func (s *LocalKickstartSession) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartSession
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartSession.UnmarshalJSON"); err != nil {
		return err
	}
	*s = LocalKickstartSession(decoded)
	return nil
}

// Validate checks the snapshot and that it does not carry the forest.
func (s LocalKickstartSession) Validate() error {
	if err := requireNonEmpty(s.ID, "schema.LocalKickstartSession.Validate", "id"); err != nil {
		return err
	}
	if !s.Scope.IsValid() {
		return fmt.Errorf("kickstart session validation failed for %q at schema.LocalKickstartSession.Validate: scope %q is outside the closed set; emit full or selection", s.ID, s.Scope)
	}
	if err := requireNonEmpty(s.OpenedAt, "schema.LocalKickstartSession.Validate", "openedAt"); err != nil {
		return err
	}
	if err := requireList(s.Stops, "schema.LocalKickstartSession.Validate", "stops"); err != nil {
		return err
	}
	if err := requireList(s.Conflicts, "schema.LocalKickstartSession.Validate", "conflicts"); err != nil {
		return err
	}
	for i := range s.Stops {
		if err := s.Stops[i].Validate(); err != nil {
			return err
		}
	}
	for i := range s.Conflicts {
		if err := s.Conflicts[i].Validate(); err != nil {
			return err
		}
	}
	if err := s.Connect.Validate(); err != nil {
		return err
	}
	if err := s.Scan.Validate(); err != nil {
		return err
	}
	return s.Village.Validate()
}

// LocalKickstartNode is one forest row. State is server-supplied. Locked is a
// boolean, not a fourth state: a locked row is disabled, and the page does
// not recompute either fact.
type LocalKickstartNode struct {
	ID       string                  `json:"id" minLength:"1"`
	ParentID string                  `json:"parentId,omitempty"`
	Kind     LocalKickstartNodeKind  `json:"kind"`
	State    LocalKickstartNodeState `json:"state"`
	Label    string                  `json:"label" minLength:"1"`
	// SessionID identifies a stored transcript, not the kickstart session carried by events.
	SessionID  string `json:"sessionId,omitempty"`
	Harness    string `json:"harness,omitempty"`
	ChildCount int    `json:"childCount,omitempty" minimum:"0"`
	Locked     bool   `json:"locked,omitempty"`
}

// UnmarshalJSON refuses an unknown node field.
func (n *LocalKickstartNode) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartNode
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartNode.UnmarshalJSON"); err != nil {
		return err
	}
	*n = LocalKickstartNode(decoded)
	return nil
}

// Validate checks the kind, the state, and that locked was not smuggled in
// as a state token. The state check is the closed set; locked is only a bool.
func (n LocalKickstartNode) Validate() error {
	if err := requireNonEmpty(n.ID, "schema.LocalKickstartNode.Validate", "id"); err != nil {
		return err
	}
	if !n.Kind.IsValid() {
		return fmt.Errorf("kickstart node validation failed for %q at schema.LocalKickstartNode.Validate: kind %q is outside the closed set; emit project, group, branch, or session", n.ID, n.Kind)
	}
	if !n.State.IsValid() {
		return fmt.Errorf("kickstart node validation failed for %q at schema.LocalKickstartNode.Validate: state %q is outside the closed set; emit none, some, or all. locked is a boolean, not a state", n.ID, n.State)
	}
	if err := requireNonEmpty(n.Label, "schema.LocalKickstartNode.Validate", "label"); err != nil {
		return err
	}
	if n.ChildCount < 0 {
		return fmt.Errorf("kickstart node validation failed for %q at schema.LocalKickstartNode.Validate: childCount is negative; emit zero or omit it", n.ID)
	}
	return nil
}

// LocalKickstartTree is the whole forest; its size is len(Nodes). There is no cursor.
type LocalKickstartTree struct {
	Nodes []LocalKickstartNode `json:"nodes" nullable:"false"`
}

// UnmarshalJSON refuses an unknown tree field, including a page or a cursor.
func (t *LocalKickstartTree) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartTree
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartTree.UnmarshalJSON"); err != nil {
		return err
	}
	*t = LocalKickstartTree(decoded)
	return nil
}

// Validate checks every node.
func (t LocalKickstartTree) Validate() error {
	if err := requireList(t.Nodes, "schema.LocalKickstartTree.Validate", "nodes"); err != nil {
		return err
	}
	for i := range t.Nodes {
		if err := t.Nodes[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartPreview is the bounded typed lines for one forest row.
type LocalKickstartPreview struct {
	Lines []LocalKickstartLine `json:"lines" nullable:"false"`
}

// UnmarshalJSON refuses an unknown preview field.
func (p *LocalKickstartPreview) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartPreview
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartPreview.UnmarshalJSON"); err != nil {
		return err
	}
	*p = LocalKickstartPreview(decoded)
	return nil
}

// Validate checks every line.
func (p LocalKickstartPreview) Validate() error {
	if err := requireList(p.Lines, "schema.LocalKickstartPreview.Validate", "lines"); err != nil {
		return err
	}
	for i := range p.Lines {
		if err := p.Lines[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartAnswers is the PATCH body: one stop's answers, the tree's
// selected ids, conflict resolutions, or the connect offer. revision is the
// session revision the caller last rendered.
type LocalKickstartAnswers struct {
	Revision      uint64                              `json:"revision" minimum:"0"`
	Values        map[string]LocalSettingValue        `json:"values,omitempty" nullable:"false"`
	SelectedIDs   []string                            `json:"selectedIds,omitempty" nullable:"false"`
	Resolutions   map[string]LocalKickstartResolution `json:"resolutions,omitempty" nullable:"false"`
	ConnectAnswer LocalKickstartConnectAnswer         `json:"connectAnswer,omitempty"`
}

// UnmarshalJSON refuses an unknown answer field.
func (a *LocalKickstartAnswers) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartAnswers
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartAnswers.UnmarshalJSON"); err != nil {
		return err
	}
	*a = LocalKickstartAnswers(decoded)
	return nil
}

// Validate checks resolutions and the connect token. Field keys stay open:
// Peasant owns that set and refuses an unknown key.
func (a LocalKickstartAnswers) Validate() error {
	if a.Values == nil && a.SelectedIDs == nil && a.Resolutions == nil && a.ConnectAnswer == "" {
		return fmt.Errorf("kickstart answers validation failed at schema.LocalKickstartAnswers.Validate: no mutation supplied; provide values, selectedIds, resolutions, or connectAnswer")
	}
	for key, resolution := range a.Resolutions {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("kickstart answers validation failed at schema.LocalKickstartAnswers.Validate: a resolution key is empty; name the conflict field")
		}
		if !resolution.IsValid() {
			return fmt.Errorf("kickstart answers validation failed for %q at schema.LocalKickstartAnswers.Validate: resolution %q is outside the closed set; emit mine or theirs", key, resolution)
		}
	}
	if a.ConnectAnswer != "" && !a.ConnectAnswer.IsValid() {
		return fmt.Errorf("kickstart answers validation failed at schema.LocalKickstartAnswers.Validate: connectAnswer %q is outside the closed set; emit accept or skip, or omit it", a.ConnectAnswer)
	}
	return nil
}

// LocalKickstartOverwrite names one field a stale write replaced, and the
// revision at which the replaced value was set.
type LocalKickstartOverwrite struct {
	Key      string            `json:"key" minLength:"1"`
	Previous LocalSettingValue `json:"previous"`
	Revision uint64            `json:"revision" minimum:"0"`
}

// UnmarshalJSON refuses an unknown overwrite field.
func (o *LocalKickstartOverwrite) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartOverwrite
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartOverwrite.UnmarshalJSON"); err != nil {
		return err
	}
	*o = LocalKickstartOverwrite(decoded)
	return nil
}

// Validate checks the field key and the previous JSON value.
func (o LocalKickstartOverwrite) Validate() error {
	if err := requireNonEmpty(o.Key, "schema.LocalKickstartOverwrite.Validate", "key"); err != nil {
		return err
	}
	if !json.Valid(o.Previous) {
		return fmt.Errorf("kickstart overwrite validation failed: previous must be a JSON value")
	}
	return nil
}

// LocalKickstartPatchResult is the PATCH response. A stale write is applied
// and reported; it is not a separate status. Overwritten names only the
// fields that changed after the caller's revision, so a stale result may
// legitimately carry none.
type LocalKickstartPatchResult struct {
	Revision    uint64                    `json:"revision" minimum:"0"`
	Stale       bool                      `json:"stale"`
	Overwritten []LocalKickstartOverwrite `json:"overwritten,omitempty" nullable:"false"`
	View        *LocalKickstartSession    `json:"view,omitempty"`
}

// MarshalJSON preserves a supplied empty overwritten list distinctly from an omitted list.
func (r LocalKickstartPatchResult) MarshalJSON() ([]byte, error) {
	type plain LocalKickstartPatchResult
	if r.Overwritten == nil {
		return json.Marshal(plain(r))
	}
	return json.Marshal(struct {
		plain
		Overwritten []LocalKickstartOverwrite `json:"overwritten"`
	}{plain: plain(r), Overwritten: r.Overwritten})
}

// UnmarshalJSON refuses an unknown patch-result field.
func (r *LocalKickstartPatchResult) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartPatchResult
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartPatchResult.UnmarshalJSON"); err != nil {
		return err
	}
	*r = LocalKickstartPatchResult(decoded)
	return nil
}

// Validate checks that a stale result names only fields that were replaced,
// and that a current result names nothing. A stale request may replace no
// field at all: another tab can advance the revision while this caller edits
// a field nobody touched, so stale is true with an absent or empty
// overwritten list. An entry appears only for a field changed after the
// caller's revision.
func (r LocalKickstartPatchResult) Validate() error {
	if !r.Stale && len(r.Overwritten) > 0 {
		return fmt.Errorf("kickstart patch validation failed at schema.LocalKickstartPatchResult.Validate: stale is false and overwritten is set; a current write replaced nothing; omit overwritten")
	}
	for i := range r.Overwritten {
		if err := r.Overwritten[i].Validate(); err != nil {
			return err
		}
	}
	if r.View != nil {
		if r.Revision != r.View.Revision {
			return fmt.Errorf("kickstart patch validation failed: revision differs from view revision")
		}
		if err := r.View.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartCommitted is the commit success body. NextSteps uses the
// closed finish menu. Diagnostics are the selection-conflict lines. The body
// has no reset control and no hook-install flag.
type LocalKickstartCommitted struct {
	// Saved is always true on a successful commit, including when ingest fails.
	Saved       bool                     `json:"saved"`
	ConfigPath  string                   `json:"configPath" minLength:"1"`
	Ingest      LocalKickstartIngest     `json:"ingest"`
	NextSteps   []LocalKickstartNextStep `json:"nextSteps" nullable:"false"`
	Diagnostics []string                 `json:"diagnostics,omitempty" nullable:"false"`
}

// UnmarshalJSON refuses an unknown commit field.
func (c *LocalKickstartCommitted) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartCommitted
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartCommitted.UnmarshalJSON"); err != nil {
		return err
	}
	*c = LocalKickstartCommitted(decoded)
	return nil
}

// Validate checks the path, the ingest, and every next step.
func (c LocalKickstartCommitted) Validate() error {
	if !c.Saved {
		return fmt.Errorf("kickstart commit validation failed: saved must be true on success")
	}
	if err := requireNonEmpty(c.ConfigPath, "schema.LocalKickstartCommitted.Validate", "configPath"); err != nil {
		return err
	}
	if err := c.Ingest.Validate(); err != nil {
		return err
	}
	if err := requireList(c.NextSteps, "schema.LocalKickstartCommitted.Validate", "nextSteps"); err != nil {
		return err
	}
	seen := map[LocalKickstartNextStep]struct{}{}
	for _, step := range c.NextSteps {
		if !step.IsValid() {
			return fmt.Errorf("kickstart commit validation failed at schema.LocalKickstartCommitted.Validate: next step %q is outside the closed set; emit a member of schema.AllLocalKickstartNextSteps", step)
		}
		if _, duplicate := seen[step]; duplicate {
			return fmt.Errorf("kickstart commit validation failed at schema.LocalKickstartCommitted.Validate: next step %q is repeated; list each action once", step)
		}
		seen[step] = struct{}{}
	}
	return nil
}

// LocalKickstartCommitRefusal is the 409 body when a commit finds conflict
// rows. Nothing was written. The rows are the same mine/theirs rows the
// review stop shows.
type LocalKickstartCommitRefusal struct {
	Error     string                   `json:"error" minLength:"1"`
	Conflicts []LocalKickstartConflict `json:"conflicts" nullable:"false"`
}

// UnmarshalJSON refuses an unknown refusal field.
func (r *LocalKickstartCommitRefusal) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartCommitRefusal
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartCommitRefusal.UnmarshalJSON"); err != nil {
		return err
	}
	*r = LocalKickstartCommitRefusal(decoded)
	return nil
}

// Validate checks the refusal names the reason and at least one row.
func (r LocalKickstartCommitRefusal) Validate() error {
	if err := requireNonEmpty(r.Error, "schema.LocalKickstartCommitRefusal.Validate", "error"); err != nil {
		return err
	}
	if err := requireList(r.Conflicts, "schema.LocalKickstartCommitRefusal.Validate", "conflicts"); err != nil {
		return err
	}
	if len(r.Conflicts) == 0 {
		return fmt.Errorf("kickstart commit refusal validation failed at schema.LocalKickstartCommitRefusal.Validate: conflicts is empty; a 409 names the rows that blocked the save, and an empty list is not a refusal")
	}
	for i := range r.Conflicts {
		if err := r.Conflicts[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalKickstartEvent is the server message data for the kickstart topic.
// SessionID identifies the kickstart session, not a stored transcript node.
// One event carries one session id. A view, when present, is that session's
// snapshot and no other. Subscribe yields one such snapshot per session id;
// the event itself is that one snapshot, not a map of sessions.
type LocalKickstartEvent struct {
	Kind      LocalKickstartEventKind   `json:"kind"`
	SessionID string                    `json:"sessionId" minLength:"1"`
	Revision  uint64                    `json:"revision" minimum:"0"`
	Reason    LocalKickstartCloseReason `json:"reason,omitempty"`
	View      *LocalKickstartSession    `json:"view,omitempty"`
	Scan      *LocalKickstartScan       `json:"scan,omitempty"`
	Ingest    *LocalKickstartIngest     `json:"ingest,omitempty"`
	Committed *LocalKickstartCommitted  `json:"committed,omitempty"`
}

// UnmarshalJSON refuses an unknown event field.
func (e *LocalKickstartEvent) UnmarshalJSON(data []byte) error {
	type plain LocalKickstartEvent
	var decoded plain
	if err := unmarshalStrict(data, &decoded, "schema.LocalKickstartEvent.UnmarshalJSON"); err != nil {
		return err
	}
	*e = LocalKickstartEvent(decoded)
	return nil
}

// Validate checks the kind, the one session id, and that the carried object
// matches the kind. A view whose id differs from sessionId is a second
// snapshot and is refused.
func (e LocalKickstartEvent) Validate() error {
	if !e.Kind.IsValid() {
		return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: kind %q is outside the closed set; emit a member of schema.AllLocalKickstartEventKinds", e.Kind)
	}
	if err := requireNonEmpty(e.SessionID, "schema.LocalKickstartEvent.Validate", "sessionId"); err != nil {
		return err
	}
	if e.View != nil && e.View.ID != e.SessionID {
		return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: view id %q does not match sessionId %q; one event carries one session snapshot", e.View.ID, e.SessionID)
	}
	if e.View != nil && e.View.Revision != e.Revision {
		return fmt.Errorf("kickstart event validation failed: revision differs from view revision")
	}
	if e.Kind != LocalKickstartEventClosed && e.Committed != nil {
		return fmt.Errorf("kickstart event validation failed: committed belongs only to a closed event")
	}
	switch e.Kind {
	case LocalKickstartEventSession, LocalKickstartEventAnswers, LocalKickstartEventConflicts:
		if e.Reason != "" || e.Scan != nil || e.Ingest != nil {
			return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: kind %s carries reason, scan, or ingest; those belong to closed, scan, and ingest events", e.Kind)
		}
		if e.View == nil {
			return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: kind %s has no view; emit the one session snapshot", e.Kind)
		}
		return e.View.Validate()
	case LocalKickstartEventScan:
		if e.Reason != "" || e.View != nil || e.Ingest != nil || e.Scan == nil {
			return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: a scan event carries only scan; the forest is loaded from the tree route")
		}
		return e.Scan.Validate()
	case LocalKickstartEventIngest:
		if e.Reason != "" || e.View != nil || e.Scan != nil || e.Ingest == nil {
			return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: an ingest event carries only ingest")
		}
		return e.Ingest.Validate()
	case LocalKickstartEventClosed:
		if !e.Reason.IsValid() {
			return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: closed reason %q is outside the closed set; emit committed or expired", e.Reason)
		}
		if e.Scan != nil || e.Ingest != nil {
			return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: a closed event does not carry scan or ingest")
		}
		if (e.Reason == LocalKickstartClosedCommitted) != (e.Committed != nil) {
			return fmt.Errorf("kickstart event validation failed: committed finish body is required only for committed closure")
		}
		if e.Committed != nil {
			if err := e.Committed.Validate(); err != nil {
				return err
			}
		}
		if e.View != nil {
			return e.View.Validate()
		}
		return nil
	default:
		return fmt.Errorf("kickstart event validation failed at schema.LocalKickstartEvent.Validate: kind %q is outside the closed set", e.Kind)
	}
}
