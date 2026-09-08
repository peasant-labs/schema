package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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
//
// v10: optional AdapterVersion records the successful Peasant adapter/parser.
//
//	The v9-to-v10 change is local metadata bookkeeping, not a reason to read
//	native sources again. Consumers can losslessly adopt v9 metadata without
//	inventing the missing producer revision. Native refresh is a separate decision.
const MetadataSchemaVersion = 10

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
	SchemaVersion int                   `json:"schemaVersion"`
	SessionID     SessionID             `json:"sessionId"`
	ParentUUID    *SessionID            `json:"parentUuid"` // nil for root sessions, pointer for nullable JSON
	ModelHarness  Harness               `json:"harness"`
	Model         ModelID               `json:"model"`
	Version       string                `json:"version"` // provider tool version (e.g. "2.1.47")
	Timestamp     TimestampInfo         `json:"timestamp"`
	Source        SourceInfo            `json:"source"`
	Git           GitContext            `json:"git"`
	Project       ProjectContext        `json:"project"`
	HostSlug      HostSlug              `json:"hostSlug"`
	Stats         SessionStats          `json:"stats"`
	Subagents     []SubagentRef         `json:"subagents"`
	RootSessionID *SessionID            `json:"rootSessionId,omitempty"`
	Purpose       SessionPurpose        `json:"purpose,omitempty"`
	Relationships []SessionRelationship `json:"relationships,omitempty"`
	CWD           string                `json:"cwd,omitempty"`       // Real project working directory (v7+)
	DerivedAt     *int64                `json:"derivedAt,omitempty"` // Unix ms when metadata.json was derived from DB (v8+); nil if written before DB insert
	Diagnostics   DiagnosticsInfo       `json:"diagnostics"`
	ContentHash   string                `json:"contentHash"`  // SHA3-256 of transcript bytes
	MetadataHash  string                `json:"metadataHash"` // SHA3-256 of metadata (excluding hashes + redaction)
	Redaction     RedactionInfo         `json:"redaction"`
	// AdapterVersion identifies the Peasant adapter/parser that successfully
	// produced this artifact, not the native harness release in Version.
	// Omission means unknown historical provenance; a present value must be positive.
	AdapterVersion *int `json:"adapterVersion,omitempty" minimum:"1" nullable:"false" description:"Successful Peasant adapter/parser revision for this local artifact, not the native harness release. Omit when unknown; a present revision must be positive."`
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

// PublishedAssociation is the producer-owned durable identity for one observed
// session-to-commit relationship published with a transcript. Its ID is opaque
// to consumers. The enclosing PublishRequest.Identity.SessionID supplies the
// session identity, so it is deliberately not repeated here.
//
// Within a publish request, both ID and ObservedCommitHash must be unique.
// Consumers retain one durable ID for each owner, transcript, and observed
// commit hash: an exact replay is idempotent, while a changed binding or a
// second ID for the same relationship must be rejected rather than aliased.
type PublishedAssociation struct {
	ID                 AssociationID `json:"id" yaml:"id"`
	ObservedCommitHash string        `json:"observedCommitHash" yaml:"observedCommitHash"`
}

// GitContext holds git repository state at the time of the session.
type GitContext struct {
	Branch       *string                `json:"branch,omitempty"`       // Current git branch
	Remote       *string                `json:"remote,omitempty"`       // Git remote URL
	Worktree     *string                `json:"worktree,omitempty"`     // Worktree path (if applicable)
	Tracking     *string                `json:"tracking,omitempty"`     // Upstream tracking branch (e.g. "origin/main")
	Commits      []CommitInfo           `json:"commits,omitempty"`      // Commits produced during this session (v4+)
	Associations []PublishedAssociation `json:"associations,omitempty"` // Durable observed session-to-commit relationships
}

// ProjectContext identifies the project associated with a session.
type ProjectContext struct {
	Hash     ProjectHash `json:"hash"`               // SHA-256 of project origin URL or path
	FilePath string      `json:"filePath,omitempty"` // Local repo path
	Name     string      `json:"name"`               // Repo basename
}

// SessionStats holds aggregate metrics extracted from transcript data.
type SessionStats struct {
	TurnCount            int    `json:"turnCount"`
	InputSubmissionCount *int64 `json:"inputSubmissionCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	ToolCallCount        int    `json:"toolCallCount"`
	SubagentCount        int    `json:"subagentCount"`
	DurationMs           int64  `json:"durationMs"`
	TokensIn             int    `json:"tokensIn"`
	TokensOut            int    `json:"tokensOut"`
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
// AdapterVersion is positive when present; omission represents unknown provenance.
func (m *UnifiedMetadata) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if statsRaw, present := raw["stats"]; present && !bytes.Equal(bytes.TrimSpace(statsRaw), []byte("null")) {
		var stats map[string]json.RawMessage
		if err := json.Unmarshal(statsRaw, &stats); err == nil {
			if count, exists := stats["inputSubmissionCount"]; exists {
				if err := validateRawInputSubmissionCount(count, "metadata.stats.inputSubmissionCount"); err != nil {
					return err
				}
			}
		}
	}
	type alias UnifiedMetadata // avoid recursion into this method
	next := *m
	next.AdapterVersion = nil
	aux := &struct {
		LegacyModelHarness *Harness        `json:"modelHarness"`
		AdapterVersion     json.RawMessage `json:"adapterVersion"`
		*alias
	}{alias: (*alias)(&next)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.AdapterVersion) > 0 {
		if bytes.Equal(bytes.TrimSpace(aux.AdapterVersion), []byte("null")) {
			return fmt.Errorf("decode metadata adapterVersion: null is not a producer revision; omit the field when unknown or provide a positive integer")
		}
		if err := json.Unmarshal(aux.AdapterVersion, &next.AdapterVersion); err != nil {
			return fmt.Errorf("decode metadata adapterVersion: provide a positive integer or omit unknown provenance: %w", err)
		}
	}
	if err := next.validateAdapterVersion(); err != nil {
		return err
	}
	if err := ValidateInputSubmissionCount(next.Stats.InputSubmissionCount, "metadata.stats.inputSubmissionCount"); err != nil {
		return err
	}
	if next.RootSessionID != nil {
		if _, err := NewSessionID(string(*next.RootSessionID)); err != nil {
			return fmt.Errorf("metadata graph validation failed at schema.UnifiedMetadata.UnmarshalJSON: rootSessionId is malformed; durable identity cannot be decoded; provide a canonical session identifier: %w", err)
		}
	}
	if !next.Purpose.IsValid() {
		return fmt.Errorf("metadata graph validation failed at schema.UnifiedMetadata.UnmarshalJSON: purpose %q is outside its closed set; consumers cannot classify the session; use a published purpose or omit it", next.Purpose)
	}
	if err := ValidateSessionRelationships(next.Relationships); err != nil {
		return err
	}
	if next.ModelHarness == "" && aux.LegacyModelHarness != nil {
		next.ModelHarness = *aux.LegacyModelHarness
	}
	*m = next
	return nil
}

// MarshalJSON preserves optional producer provenance and refuses invalid revisions
// even when the metadata was assembled directly rather than decoded from JSON.
func (m UnifiedMetadata) MarshalJSON() ([]byte, error) {
	if err := m.validateAdapterVersion(); err != nil {
		return nil, err
	}
	if err := ValidateInputSubmissionCount(m.Stats.InputSubmissionCount, "metadata.stats.inputSubmissionCount"); err != nil {
		return nil, err
	}
	if m.RootSessionID != nil {
		if _, err := NewSessionID(string(*m.RootSessionID)); err != nil {
			return nil, fmt.Errorf("metadata graph validation failed at schema.UnifiedMetadata.MarshalJSON: rootSessionId is malformed; durable identity cannot be serialized; provide a canonical session identifier: %w", err)
		}
	}
	if !m.Purpose.IsValid() {
		return nil, fmt.Errorf("metadata graph validation failed at schema.UnifiedMetadata.MarshalJSON: purpose %q is outside its closed set; consumers cannot classify the session; use a published purpose or omit it", m.Purpose)
	}
	if err := ValidateSessionRelationships(m.Relationships); err != nil {
		return nil, err
	}
	type alias UnifiedMetadata
	return json.Marshal(alias(m))
}

// ValidateInputSubmissionCount preserves unknown (nil) separately from measured
// zero and limits present values to the exact integer range shared with JS.
func ValidateInputSubmissionCount(value *int64, path string) error {
	if value != nil && (*value < 0 || *value > maxSafeJSONInteger) {
		return fmt.Errorf("input submission count validation failed at schema.ValidateInputSubmissionCount for %s: value %d is outside 0..9007199254740991; consumers cannot preserve the count exactly; omit an unknown count or provide a measured JS-safe integer", path, *value)
	}
	return nil
}

func (m UnifiedMetadata) validateAdapterVersion() error {
	if m.AdapterVersion != nil && *m.AdapterVersion <= 0 {
		return fmt.Errorf("validate metadata adapterVersion: got %d; producer revisions must be positive, so omit the field for unknown provenance instead", *m.AdapterVersion)
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
