package schema

import "encoding/json"

// MetadataSchemaVersion is the schema version written by this build of the ingest tool.
// v1: initial schema
// v2: SQLite persistence layer, store integration
// v3: transcript filename {sessionId}--transcript.{ext} (was {unixMillis}--),
//
//	timestamp extraction scans all lines (was first/last line only)
//
// v4: GitContext.Commits field added (session-to-commit linking)
// v5: Claude indexer fixes — system content extraction, tool_result tool_use_id linking
// v6: ContentHash, MetadataHash, RedactionInfo — deferred redaction model
// v7: CWD field — real working directory for context-aware slug redaction
// v8: DerivedAt field — Unix ms when metadata.json was derived from DB (DB as SOT)
// v9: harness key unification — UnifiedMetadata.ModelHarness re-keyed json:"modelHarness"
//
//	-> json:"harness" (emit-side flip). The bump forces the DIFF stage to
//	re-classify existing on-disk sessions as Updated so the stale "modelHarness" key
//	self-heals (re-extract + rewrite) on next ingest; UnmarshalJSON still accepts the
//	legacy key on pre-v9 files in the meantime.
const MetadataSchemaVersion = 9

// RedactionInfo tracks whether and when redaction was applied to a session's transcript.
// Level is stored as a string because this schema module is a public contract module that
// must not import internal packages. Callers should validate using redact.RedactionLevel(level).IsValid().
type RedactionInfo struct {
	Applied             bool   `json:"applied"`
	Level               string `json:"level,omitempty"`                  // redaction level used (validate with redact.RedactionLevel)
	RuleSetVersion      string `json:"rule_set_version,omitempty"`       // rule set version that produced this redaction (e.g. "1.1.0")
	RedactedAtMs        *int64 `json:"redacted_at_ms,omitempty"`         // unix ms when redacted
	ContentHashAtRedact string `json:"content_hash_at_redact,omitempty"` // content hash snapshot at redaction time
}

// IsStale returns true if the content has changed since redaction was applied.
func (r RedactionInfo) IsStale(currentContentHash string) bool {
	return r.Applied && r.ContentHashAtRedact != currentContentHash
}

// IsCurrent returns true if the redaction matches the current content.
func (r RedactionInfo) IsCurrent(currentContentHash string) bool {
	return r.Applied && r.ContentHashAtRedact == currentContentHash
}

// IsRaw returns true if no redaction has been applied.
func (r RedactionInfo) IsRaw() bool {
	return !r.Applied
}

// UnifiedMetadata is the on-disk JSON stored alongside each raw transcript.
// It drives the adapter layer for downstream consumers and incremental diff logic.
type UnifiedMetadata struct {
	SchemaVersion int             `json:"schemaVersion"`
	SessionID     SessionID       `json:"sessionId"`
	ParentUUID    *SessionID      `json:"parentUuid"` // nil for root sessions, pointer for nullable JSON
	ModelHarness  Harness         `json:"harness"`
	Model         ModelID         `json:"model"`
	Version       string          `json:"version"` // provider tool version (e.g. "2.1.47")
	Timestamp     TimestampInfo   `json:"timestamp"`
	Source        SourceInfo      `json:"source"`
	Git           GitContext      `json:"git"`
	Project       ProjectContext  `json:"project"`
	HostSlug      HostSlug        `json:"hostSlug"`
	Stats         SessionStats    `json:"stats"`
	Subagents     []SubagentRef   `json:"subagents"`
	CWD           string          `json:"cwd,omitempty"`       // Real project working directory (v7+)
	DerivedAt     *int64          `json:"derivedAt,omitempty"` // Unix ms when metadata.json was derived from DB (v8+); nil if written before DB insert
	Diagnostics   DiagnosticsInfo `json:"diagnostics"`
	ContentHash   string          `json:"contentHash"`  // SHA3-256 of transcript bytes
	MetadataHash  string          `json:"metadataHash"` // SHA3-256 of metadata (excluding hashes + redaction)
	Redaction     RedactionInfo   `json:"redaction"`
}

// TimestampInfo records session timing in Unix milliseconds.
type TimestampInfo struct {
	Start    int64  `json:"start"`              // Unix millis
	End      int64  `json:"end"`                // Unix millis
	Ingested *int64 `json:"ingested,omitempty"` // Unix millis; nil if not yet ingested
}

// SourceInfo identifies the original transcript file and its format.
type SourceInfo struct {
	FilePath string       `json:"filePath,omitempty"` // Original source transcript path
	Format   SourceFormat `json:"format"`             // "jsonl" or "json"
}

// CommitInfo records a single git commit linked to a session.
type CommitInfo struct {
	Hash        string `json:"hash"`        // commit SHA-1 (full or abbreviated)
	Message     string `json:"message"`     // commit message first line
	AuthorName  string `json:"authorName"`  // author display name
	AuthorEmail string `json:"authorEmail"` // author email (used for attribution filtering)
	CommitTime  int64  `json:"commitTime"`  // committer date, Unix millis
	AuthorTime  int64  `json:"authorTime"`  // author date, Unix millis
}

// GitContext holds git repository state at the time of the session.
type GitContext struct {
	Branch   *string      `json:"branch,omitempty"`   // Current git branch
	Remote   *string      `json:"remote,omitempty"`   // Git remote URL
	Worktree *string      `json:"worktree,omitempty"` // Worktree path (if applicable)
	Tracking *string      `json:"tracking,omitempty"` // Upstream tracking branch (e.g. "origin/main")
	Commits  []CommitInfo `json:"commits,omitempty"`  // Commits produced during this session (v4+)
}

// ProjectContext identifies the project associated with a session.
type ProjectContext struct {
	Hash     ProjectHash `json:"hash"`               // SHA-256 of project origin URL or path
	FilePath string      `json:"filePath,omitempty"` // Local repo path
	Name     string      `json:"name"`               // Repo basename
}

// SessionStats holds aggregate metrics extracted from transcript data.
type SessionStats struct {
	TurnCount     int   `json:"turnCount"`
	ToolCallCount int   `json:"toolCallCount"`
	SubagentCount int   `json:"subagentCount"`
	DurationMs    int64 `json:"durationMs"`
	TokensIn      int   `json:"tokensIn"`
	TokensOut     int   `json:"tokensOut"`
	// ACP-aligned token breakdown (optional — not all providers report these).
	ThoughtTokens     *int `json:"thoughtTokens,omitempty"`     // Reasoning/thinking tokens
	CachedReadTokens  *int `json:"cachedReadTokens,omitempty"`  // Prompt cache hits
	CachedWriteTokens *int `json:"cachedWriteTokens,omitempty"` // Prompt cache writes
}

// SubagentRef records a reference to a subagent session spawned during a parent session.
type SubagentRef struct {
	SessionID  SessionID `json:"sessionId"`
	ParentUUID SessionID `json:"parentUuid"`
}

// DiagnosticsInfo records issues encountered during ingestion.
type DiagnosticsInfo struct {
	Warnings []DiagnosticEntry `json:"warnings"`
	Partial  *bool             `json:"partial,omitempty"` // true if any file copy failed; nil if not determined
}

// DiagnosticEntry is a structured error object recorded during ingestion or processing.
type DiagnosticEntry struct {
	ErrorType   string `json:"errorType"`   // e.g. "parse_error", "permission_denied", "copy_failed"
	Location    string `json:"location"`    // e.g. "line 47", "debug/tool_output_3.json"
	Message     string `json:"message"`     // Human-readable description
	Remediation string `json:"remediation"` // Actionable fix suggestion
}

// UnmarshalJSON decodes UnifiedMetadata, accepting the legacy pre-v9 on-disk
// harness key. The canonical key is json:"harness" (v9+); pre-v9 files wrote
// json:"modelHarness". When the canonical key is absent but the legacy key is
// present, the legacy value is adopted so a pre-v9 file still reads its harness
// correctly in the window before the v9 DIFF-stage re-extract rewrites it.
// Marshalling is unaffected (struct tags emit only "harness").
func (m *UnifiedMetadata) UnmarshalJSON(data []byte) error {
	type alias UnifiedMetadata // avoid recursion into this method
	aux := &struct {
		LegacyModelHarness *Harness `json:"modelHarness"`
		*alias
	}{alias: (*alias)(m)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if m.ModelHarness == "" && aux.LegacyModelHarness != nil {
		m.ModelHarness = *aux.LegacyModelHarness
	}
	return nil
}

// NewUnifiedMetadata creates a UnifiedMetadata with SchemaVersion set to MetadataSchemaVersion
// and empty slices initialized (not nil) so JSON serialization produces [] instead of null.
func NewUnifiedMetadata() UnifiedMetadata {
	return UnifiedMetadata{
		SchemaVersion: MetadataSchemaVersion,
		Subagents:     []SubagentRef{},
		Diagnostics:   DiagnosticsInfo{Warnings: []DiagnosticEntry{}},
	}
}
