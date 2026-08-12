package schema

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/dayvidpham/bestiary"
	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- SessionID ---

// SessionID is a validated session identifier from a source provider.
// Accepted formats:
//   - UUID: "99d59925-36bc-424c-a789-8be54d9702ba"
//   - Claude subagent: "agent-a3aee4f"
//   - ACP session: "sess_3cd91f52effeXd3QAJ54jOyzv5" (ACP sess_ prefix)
//   - OpenCode session: "ses_3cd91f52effeXd3QAJ54jOyzv5"
//   - OpenCode message: "msg_001abc"
//   - Strike timestamped: "20260728T123456.123456789Z-ABCDEFGHIJKLMNOPQRST234567"
//   - Strike bare: "ABCDEFGHIJKLMNOPQRST234567"
type SessionID string

var (
	sessionIDUUID              = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	sessionIDSubagent          = regexp.MustCompile(`^agent-[a-f0-9]+$`)
	sessionIDOpenCode          = regexp.MustCompile(`^ses_[a-zA-Z0-9]+$`)
	sessionIDACP               = regexp.MustCompile(`^sess_[a-zA-Z0-9]+$`) // ACP sess_ prefix (double-s)
	sessionIDMessage           = regexp.MustCompile(`^msg_[a-zA-Z0-9]+$`)
	sessionIDStrikeTimestamped = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}\.[0-9]{9}Z-[A-Z2-7]{26}$`)
	sessionIDStrikeBare        = regexp.MustCompile(`^[A-Z2-7]{26}$`)
)

// NewSessionID validates and constructs a SessionID.
func NewSessionID(raw string) (SessionID, error) {
	if raw == "" {
		return "", fmt.Errorf("invalid session ID: empty string")
	}
	if strings.ContainsAny(raw, "/\\") || strings.Contains(raw, "..") {
		return "", fmt.Errorf("invalid session ID %q: contains path separator or traversal", raw)
	}
	if !sessionIDUUID.MatchString(raw) &&
		!sessionIDSubagent.MatchString(raw) &&
		!sessionIDOpenCode.MatchString(raw) &&
		!sessionIDACP.MatchString(raw) &&
		!sessionIDMessage.MatchString(raw) &&
		!sessionIDStrikeTimestamped.MatchString(raw) &&
		!sessionIDStrikeBare.MatchString(raw) {
		return "", fmt.Errorf("invalid session ID %q: must be UUID, agent-{hex}, ses_{id}, sess_{id} (ACP), msg_{id}, a Strike timestamped ID, or a 26-character uppercase RFC4648 base32 Strike ID", raw)
	}
	return SessionID(raw), nil
}

func (s SessionID) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (SessionID) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithFormat("session-id")
	s.WithTitle("Session ID")
	s.WithDescription("Unique session identifier (UUID, agent-{hex}, ses_{id}, sess_{id} (ACP), msg_{id}, or a Strike session ID)")
	s.WithPattern(`^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|agent-[a-f0-9]+|sess?_[a-zA-Z0-9]+|msg_[a-zA-Z0-9]+|[0-9]{8}T[0-9]{6}\.[0-9]{9}Z-[A-Z2-7]{26}|[A-Z2-7]{26})$`)
	s.WithExamples("99d59925-36bc-424c-a789-8be54d9702ba", "agent-a3aee4f", "ses_3cd91f52effeXd3QAJ54jOyzv5", "sess_3cd91f52effeXd3QAJ54jOyzv5", "20260728T123456.123456789Z-ABCDEFGHIJKLMNOPQRST234567", "ABCDEFGHIJKLMNOPQRST234567")
	return s, nil
}

// --- ModelID ---

// ModelID identifies a specific model version (e.g. "claude-opus-4-6").
type ModelID string

// NewModelID validates and constructs a ModelID.
func NewModelID(raw string) (ModelID, error) {
	if raw == "" {
		return "", fmt.Errorf("invalid model ID: empty string")
	}
	return ModelID(raw), nil
}

func (m ModelID) String() string { return string(m) }

// JSONSchema implements jsonschema.Exposer.
func (ModelID) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Model ID")
	s.WithDescription("Model identifier (e.g. 'claude-opus-4-6', 'gemini-2.0-flash')")
	s.WithMinLength(1)
	s.WithExamples("claude-opus-4-6", "gemini-2.0-flash", "codex-mini-latest")
	return s, nil
}

// --- ObservedModelID ---

// ObservedModelID is a model identifier observed on one assistant-generated
// turn. Producers preserve exact observations, including repeats and omissions;
// consumers carry observations forward only when deriving effective sticky state.
// Carry-forward never populates an omitted ObservedModel on the wire.
// Producers enforce that observations belong only to assistant or subagent output
// because generated shape validators cannot express that role condition.
type ObservedModelID ModelID

// NewObservedModelID validates and constructs an ObservedModelID. Accepted bytes
// are preserved exactly. Values must be valid UTF-8, non-empty, and may not have
// a Unicode White_Space code point at either edge. Internal whitespace and all
// other valid Unicode are retained without normalization or canonicalization.
func NewObservedModelID(raw string) (ObservedModelID, error) {
	if raw == "" {
		return "", fmt.Errorf("observed model validation failed at schema.NewObservedModelID while constructing assistant-turn source evidence: the value is empty, so no source observation can be represented and the caller cannot emit observedModel; omit the field when no observation exists or supply the exact non-empty identifier")
	}
	if !utf8.ValidString(raw) {
		return "", fmt.Errorf("observed model validation failed at schema.NewObservedModelID while constructing assistant-turn source evidence: value %q is not valid UTF-8, so the caller cannot emit observedModel without changing source bytes; provide the original identifier as valid UTF-8 and retry", raw)
	}
	first, _ := utf8.DecodeRuneInString(raw)
	last, _ := utf8.DecodeLastRuneInString(raw)
	if observedModelEdgeWhitespace(first) || observedModelEdgeWhitespace(last) {
		return "", fmt.Errorf("observed model validation failed at schema.NewObservedModelID while constructing assistant-turn source evidence: value %q has Unicode whitespace at an edge, but the wire preserves exact unpadded source bytes and the caller cannot emit observedModel; remove only the edge whitespace at the producing boundary and retry", raw)
	}
	return ObservedModelID(raw), nil
}

// String returns the exact source-observed identifier bytes.
func (m ObservedModelID) String() string { return string(m) }

// JSONSchema implements jsonschema.Exposer.
func (ObservedModelID) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Observed Model ID")
	s.WithDescription("Exact UTF-8 model identifier observed on an assistant-generated turn; producer-enforced as assistant or subagent evidence. Values are non-empty and may not have a Unicode White_Space code point at either edge; all accepted bytes, including Unicode, mixed case, slashes, and internal spaces, are preserved.")
	s.WithMinLength(1)
	s.WithPattern(observedModelPattern)
	s.WithExamples("anthropic/Claude-Opus-4-8", "provider/Model Family")
	return s, nil
}

// observedModelEdgeWhitespace is the Unicode White_Space property used by the
// OpenAPI pattern and generated ECMAScript validator. It is listed explicitly
// because Go's unicode.IsSpace also includes U+001C-U+001F, while ECMAScript's
// \s includes U+FEFF and omits U+0085. The contract uses the Unicode property,
// not either runtime's broader or narrower whitespace predicate.
func observedModelEdgeWhitespace(r rune) bool {
	switch {
	case r >= '\u0009' && r <= '\u000D':
		return true
	case r == '\u0020' || r == '\u0085' || r == '\u00A0' || r == '\u1680':
		return true
	case r >= '\u2000' && r <= '\u200A':
		return true
	case r == '\u2028' || r == '\u2029' || r == '\u202F' || r == '\u205F' || r == '\u3000':
		return true
	default:
		return false
	}
}

// observedModelPattern requires a complete string whose first and final code
// points are not Unicode White_Space. ECMAScript \s differs from that set only
// at U+0085 and U+FEFF, so the edge expression adds U+0085 and explicitly
// permits U+FEFF. The negative lookahead at the end is an absolute-end
// assertion: ECMAScript '$' may match before a final terminator. U+001C-U+001F
// are intentionally absent because they are not Unicode White_Space, even
// though Go's unicode.IsSpace classifies them as spaces.
const observedModelPattern = `^(?:\uFEFF|[^\s\x85])(?:[\s\S]*(?:\uFEFF|[^\s\x85]))?(?![\s\S])`

// --- ProjectHash ---

// ProjectHash is a SHA-256 hex digest of the project's origin URL or local path.
type ProjectHash string

var projectHashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// NewProjectHash validates and constructs a ProjectHash.
func NewProjectHash(raw string) (ProjectHash, error) {
	if !projectHashPattern.MatchString(raw) {
		return "", fmt.Errorf("invalid project hash %q: must be 64-character lowercase hex string", raw)
	}
	return ProjectHash(raw), nil
}

// Validate reports whether p is the canonical 64-character lowercase hex
// project identity accepted at wire boundaries.
func (p ProjectHash) Validate() error {
	if !projectHashPattern.MatchString(string(p)) {
		return fmt.Errorf("invalid project hash %q: must be a 64-character lowercase hex string", p)
	}
	return nil
}

func (p ProjectHash) String() string { return string(p) }

// JSONSchema implements jsonschema.Exposer.
func (ProjectHash) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Project Hash")
	s.WithDescription("SHA-256 hex digest of the project's origin URL or local path")
	s.WithPattern(`^[0-9a-f]{64}$`)
	s.WithMinLength(64)
	s.WithMaxLength(64)
	s.WithExamples("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2")
	return s, nil
}

// --- HostSlug ---

// HostSlug is a sanitized, filesystem-safe identifier derived from git remote.
// Contains [a-zA-Z0-9._<>-] characters. The <> characters support redaction
// placeholders like <USER> and <PATH> in redacted slugs.
type HostSlug string

var hostSlugPattern = regexp.MustCompile(`^[a-zA-Z0-9._<>-]+$`)

// NewHostSlug validates and constructs a HostSlug.
func NewHostSlug(raw string) (HostSlug, error) {
	if raw == "" {
		return "", fmt.Errorf("invalid host slug: empty string")
	}
	if strings.Contains(raw, "..") {
		return "", fmt.Errorf("invalid host slug %q: must not contain '..'", raw)
	}
	if !hostSlugPattern.MatchString(raw) {
		return "", fmt.Errorf("invalid host slug %q: must contain only [a-zA-Z0-9._<>-]", raw)
	}
	return HostSlug(raw), nil
}

func (h HostSlug) String() string { return string(h) }

// JSONSchema implements jsonschema.Exposer.
func (HostSlug) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Host Slug")
	s.WithDescription("Sanitized, filesystem-safe identifier derived from git remote; contains only [a-zA-Z0-9._<>-]")
	s.WithPattern(`^[a-zA-Z0-9._<>-]+$`)
	s.WithExamples("github.com--user--repo", "local--home-user-projects-myapp")
	return s, nil
}

// --- SourceFormat ---

// SourceFormat identifies the transcript file format.
type SourceFormat string

const (
	SourceFormatJSONL SourceFormat = "jsonl" // Claude Code JSONL transcripts
	SourceFormatJSON  SourceFormat = "json"  // OpenCode JSON transcripts
)

// AllSourceFormats is the canonical list of transcript source formats.
var AllSourceFormats = []SourceFormat{
	SourceFormatJSONL,
	SourceFormatJSON,
}

// IsValid returns true if the SourceFormat is one of the known variants.
func (f SourceFormat) IsValid() bool {
	switch f {
	case SourceFormatJSONL, SourceFormatJSON:
		return true
	}
	return false
}

func (f SourceFormat) String() string { return string(f) }

// JSONSchema implements jsonschema.Exposer.
func (SourceFormat) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Source Format")
	s.WithDescription("Transcript file format")
	s.WithEnum("jsonl", "json")
	s.WithExamples("jsonl", "json")
	return s, nil
}

// --- SessionOutcome ---

// SessionOutcome represents the resolution status of a session.
type SessionOutcome string

const (
	OutcomeResolved SessionOutcome = "resolved"
	OutcomePartial  SessionOutcome = "partial"
	OutcomeFailed   SessionOutcome = "failed"
)

// IsValid returns true if the outcome is one of the known variants.
func (o SessionOutcome) IsValid() bool {
	switch o {
	case OutcomeResolved, OutcomePartial, OutcomeFailed:
		return true
	}
	return false
}

func (o SessionOutcome) String() string { return string(o) }

// AllOutcomes is the canonical list of all known session outcomes.
var AllOutcomes = []SessionOutcome{
	OutcomeResolved, OutcomePartial, OutcomeFailed,
}

// JSONSchema implements jsonschema.Exposer.
func (SessionOutcome) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Session Outcome")
	s.WithDescription("Resolution status of the session")
	s.WithEnum("resolved", "partial", "failed")
	s.WithExamples("resolved", "partial", "failed")
	return s, nil
}

// --- Harness ---

// Harness identifies the coding tool or AI-assisted development environment
// that is driving the model interaction. Re-exported from bestiary.
type Harness = bestiary.Harness

const (
	HarnessClaudeCode  = bestiary.HarnessClaudeCode
	HarnessGeminiCLI   = bestiary.HarnessGeminiCLI
	HarnessCodex       = bestiary.HarnessCodex
	HarnessOpenCode    = bestiary.HarnessOpenCode
	HarnessCursor      = bestiary.HarnessCursor
	HarnessAntigravity = bestiary.HarnessAntigravity
	HarnessStrike      = bestiary.HarnessStrike
)

// AllHarnesses is the canonical list of harnesses that peasant supports for ingestion.
var AllHarnesses = []Harness{
	HarnessClaudeCode, HarnessGeminiCLI, HarnessCodex, HarnessOpenCode, HarnessCursor,
	HarnessStrike,
}

// Harnesses returns every harness identifier known to bestiary — the full set,
// a superset of AllHarnesses (which is only the ingestion-supported subset).
// Re-exported so callers can enumerate the canonical known set without importing
// bestiary directly.
func Harnesses() []Harness {
	return bestiary.Harnesses()
}

// HarnessDisplayName returns the human-readable name for a harness.
func HarnessDisplayName(h Harness) string {
	switch h {
	case HarnessClaudeCode:
		return "Claude Code"
	case HarnessGeminiCLI:
		return "Gemini CLI"
	case HarnessCodex:
		return "Codex"
	case HarnessOpenCode:
		return "OpenCode"
	case HarnessCursor:
		return "Cursor"
	case HarnessAntigravity:
		return "Antigravity"
	case HarnessStrike:
		return "Strike"
	}
	return string(h)
}

// HarnessJSONSchema returns the JSON Schema for the Harness type.
// This is a standalone function because Harness is a type alias for
// bestiary.Harness, and Go does not allow methods on imported types.
func HarnessJSONSchema() jsonschema.Schema {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Harness")
	s.WithDescription("AI coding tool or development environment")
	// Derive the enum from bestiary's canonical list rather than hardcoding it,
	// so a new harness added to bestiary flows through automatically (avoids DRY drift).
	all := Harnesses()
	enum := make([]any, len(all))
	for i, h := range all {
		enum[i] = string(h)
	}
	s.WithEnum(enum...)
	s.WithExamples("claude-code", "opencode")
	return s
}

// --- Visibility ---

// Visibility controls who can access a published transcript.
type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityGroup   Visibility = "group"
	VisibilityPublic  Visibility = "public"
)

// IsValid returns true if the visibility is one of the known variants.
func (v Visibility) IsValid() bool {
	switch v {
	case VisibilityPrivate, VisibilityGroup, VisibilityPublic:
		return true
	}
	return false
}

func (v Visibility) String() string { return string(v) }

// AllVisibilities is the canonical list of all known visibility levels.
var AllVisibilities = []Visibility{
	VisibilityPrivate, VisibilityGroup, VisibilityPublic,
}

// JSONSchema implements jsonschema.Exposer.
func (Visibility) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Visibility")
	s.WithDescription("Access control level for a published transcript")
	s.WithEnum("private", "group", "public")
	s.WithExamples("public", "private")
	return s, nil
}

// --- License ---

// License is the content license a contributor selects for a published
// transcript. Closed menu: the village `licenses` table (its migration 026)
// carries each license's obligations, and the peasant local store mirrors the set
// in a CHECK (migration v37). This is the single source of truth for the menu on
// the publish/pull wire.
type License string

const (
	LicenseCC0    License = "CC0-1.0"
	LicenseCCBY   License = "CC-BY-4.0"
	LicenseCCBYSA License = "CC-BY-SA-4.0"
)

// IsValid reports whether the license is one of the known menu entries.
func (l License) IsValid() bool {
	switch l {
	case LicenseCC0, LicenseCCBY, LicenseCCBYSA:
		return true
	}
	return false
}

func (l License) String() string { return string(l) }

// AllLicenses is the canonical menu of known licenses.
var AllLicenses = []License{
	LicenseCC0, LicenseCCBY, LicenseCCBYSA,
}

// LicenseMenu returns the known license IDs as a comma-separated string, derived
// from AllLicenses so help text and error messages can't drift from the canonical
// menu when a license is added.
func LicenseMenu() string {
	ids := make([]string, len(AllLicenses))
	for i, l := range AllLicenses {
		ids[i] = l.String()
	}
	return strings.Join(ids, ", ")
}

// JSONSchema implements jsonschema.Exposer.
func (License) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("License")
	s.WithDescription("Content license for a published transcript")
	// Derive the enum from AllLicenses rather than hardcoding it, so a new license
	// flows through automatically (matches the Harness pattern; avoids DRY drift).
	enum := make([]any, len(AllLicenses))
	for i, l := range AllLicenses {
		enum[i] = l.String()
	}
	s.WithEnum(enum...)
	s.WithExamples("CC0-1.0")
	return s, nil
}

// --- ToolCallKind ---

// ToolCallKind classifies a tool call in an agent session.
// Aligns with ACP's ToolCallUpdate.kind — enables ExplorationRatio without heuristic parsing.
type ToolCallKind string

const (
	ToolCallKindRead    ToolCallKind = "read"
	ToolCallKindEdit    ToolCallKind = "edit"
	ToolCallKindDelete  ToolCallKind = "delete"
	ToolCallKindMove    ToolCallKind = "move"
	ToolCallKindSearch  ToolCallKind = "search"
	ToolCallKindExecute ToolCallKind = "execute"
	ToolCallKindThink   ToolCallKind = "think"
	ToolCallKindFetch   ToolCallKind = "fetch"
	ToolCallKindOther   ToolCallKind = "other"
)

// IsValid returns true if the tool call kind is one of the known variants.
func (k ToolCallKind) IsValid() bool {
	switch k {
	case ToolCallKindRead, ToolCallKindEdit, ToolCallKindDelete, ToolCallKindMove,
		ToolCallKindSearch, ToolCallKindExecute, ToolCallKindThink, ToolCallKindFetch, ToolCallKindOther:
		return true
	}
	return false
}

func (k ToolCallKind) String() string { return string(k) }

// AllToolCallKinds is the canonical list of all known tool call kinds.
var AllToolCallKinds = []ToolCallKind{
	ToolCallKindRead, ToolCallKindEdit, ToolCallKindDelete, ToolCallKindMove,
	ToolCallKindSearch, ToolCallKindExecute, ToolCallKindThink, ToolCallKindFetch, ToolCallKindOther,
}

// JSONSchema implements jsonschema.Exposer.
func (ToolCallKind) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Tool Call Kind")
	s.WithDescription("Classification of a tool call, aligned with ACP ToolCallUpdate.kind")
	s.WithEnum("read", "edit", "delete", "move", "search", "execute", "think", "fetch", "other")
	s.WithExamples("read", "edit", "execute")
	return s, nil
}

// --- StopReason ---

// StopReason represents why a session or turn ended.
// ACP per-turn stop reasons; content-layer type for future use.
type StopReason string

const (
	StopReasonEndTurn         StopReason = "end_turn"
	StopReasonCancelled       StopReason = "cancelled"
	StopReasonMaxTokens       StopReason = "max_tokens"
	StopReasonMaxTurnRequests StopReason = "max_turn_requests"
	StopReasonRefusal         StopReason = "refusal"
)

// IsValid returns true if the stop reason is one of the known variants.
func (r StopReason) IsValid() bool {
	switch r {
	case StopReasonEndTurn, StopReasonCancelled, StopReasonMaxTokens,
		StopReasonMaxTurnRequests, StopReasonRefusal:
		return true
	}
	return false
}

func (r StopReason) String() string { return string(r) }

// AllStopReasons is the canonical list of all known stop reasons.
var AllStopReasons = []StopReason{
	StopReasonEndTurn, StopReasonCancelled, StopReasonMaxTokens,
	StopReasonMaxTurnRequests, StopReasonRefusal,
}

// JSONSchema implements jsonschema.Exposer.
func (StopReason) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Stop Reason")
	s.WithDescription("Reason why a session or turn ended (ACP-aligned)")
	s.WithEnum("end_turn", "cancelled", "max_tokens", "max_turn_requests", "refusal")
	s.WithExamples("end_turn", "max_tokens")
	return s, nil
}

// --- Role ---

// Role represents the sender of a message turn.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleSystem    Role = "system"
)

// IsValid returns true if the role is one of the known variants.
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleTool, RoleSystem:
		return true
	}
	return false
}

func (r Role) String() string { return string(r) }

// AllRoles is the canonical list of all known roles.
var AllRoles = []Role{
	RoleUser, RoleAssistant, RoleTool, RoleSystem,
}

// JSONSchema implements jsonschema.Exposer.
func (Role) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Role")
	s.WithDescription("Sender role of a message turn")
	s.WithEnum("user", "assistant", "tool", "system")
	s.WithExamples("user", "assistant")
	return s, nil
}

// --- EntryType ---

// EntryType classifies a single entry within an agent session transcript.
type EntryType string

const (
	EntryTypeText       EntryType = "text"
	EntryTypeToolUse    EntryType = "tool_use"
	EntryTypeToolResult EntryType = "tool_result"
	EntryTypeThinking   EntryType = "thinking"
	EntryTypeSystem     EntryType = "system"
	EntryTypeError      EntryType = "error"
	EntryTypeResult     EntryType = "result"
)

// IsValid returns true if the entry type is one of the known variants.
func (e EntryType) IsValid() bool {
	switch e {
	case EntryTypeText, EntryTypeToolUse, EntryTypeToolResult,
		EntryTypeThinking, EntryTypeSystem, EntryTypeError, EntryTypeResult:
		return true
	}
	return false
}

func (e EntryType) String() string { return string(e) }

// AllEntryTypes is the canonical list of all known entry types.
var AllEntryTypes = []EntryType{
	EntryTypeText, EntryTypeToolUse, EntryTypeToolResult,
	EntryTypeThinking, EntryTypeSystem, EntryTypeError, EntryTypeResult,
}

// JSONSchema implements jsonschema.Exposer.
func (EntryType) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Entry Type")
	s.WithDescription("Classification of a single entry within an agent session transcript")
	s.WithEnum("text", "tool_use", "tool_result", "thinking", "system", "error", "result")
	s.WithExamples("text", "tool_use", "tool_result")
	return s, nil
}
