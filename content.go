package schema

// SessionEntry represents a single indexed entry within a session transcript.
// Defined for v2 content-layer use; NOT used in v1 PublishRequest wire format.
// Maps 1:1 to a row in the session_entries table.
// All optional fields use pointer types to distinguish absent from zero.
// JSON tags use camelCase aligned with ACP content model.
type SessionEntry struct {
	SessionID      SessionID     `json:"sessionId"`
	EntryIndex     int           `json:"entryIndex"`
	Harness        Harness       `json:"harness"`
	EntryType      EntryType     `json:"entryType"`
	Role           Role          `json:"role"`
	TimestampMs    *int64        `json:"timestampMs,omitempty"`
	ContentPreview *string       `json:"contentPreview,omitempty"` // max 500 chars
	TokensIn       *int          `json:"tokensIn,omitempty"`
	TokensOut      *int          `json:"tokensOut,omitempty"`
	HasToolUse     bool          `json:"hasToolUse"`
	ToolKind       *ToolCallKind `json:"toolKind,omitempty"` // ACP-aligned tool classification
	ToolNamesCSV   *string       `json:"toolNamesCsv,omitempty"`
	HasThinking    bool          `json:"hasThinking"`
	IsError        bool          `json:"isError"`
	StopReason     *StopReason   `json:"stopReason,omitempty"` // ACP per-turn stop reason
	RawByteLength  *int          `json:"rawByteLength,omitempty"`
	ToolCallID     *string       `json:"toolCallId,omitempty"`    // MCP correlation (was toolUseId)
	EntryID        *string       `json:"entryId,omitempty"`       // Provider-native ID
	ParentEntryID  *string       `json:"parentEntryId,omitempty"` // Parent entry link
	Depth          int           `json:"depth"`                   // 0 = message, 1 = content part
	ParentIndex    *int          `json:"parentIndex,omitempty"`   // entryIndex of parent (nil for depth=0)
	ToolInput      *string       `json:"toolInput,omitempty"`     // tool_use input JSON
	ToolOutput     *string       `json:"toolOutput,omitempty"`    // tool_result output JSON
	Extra          *string       `json:"extra,omitempty"`         // JSON overflow for provider-specific data
	PartType       *string       `json:"partType,omitempty"`      // provider's original part type label (nil for depth=0)
}
