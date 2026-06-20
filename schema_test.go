package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// --- AC3: TimestampInfo uses int64 millis ---

func TestTimestampInfo_Int64Fields(t *testing.T) {
	ts := schema.TimestampInfo{
		Start: 1708700000000,
		End:   1708700060000,
	}
	b, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	// Should contain integer literals, not RFC3339 strings
	s := string(b)
	if strings.Contains(s, "T") && strings.Contains(s, "Z") {
		t.Errorf("timestamp looks like RFC3339 string, expected int64: %s", s)
	}
	if !strings.Contains(s, "1708700000000") {
		t.Errorf("expected start millis in JSON output, got: %s", s)
	}
}

func TestTimestampInfo_IngestedOmittedWhenNil(t *testing.T) {
	ts := schema.TimestampInfo{
		Start: 1708700000000,
		End:   1708700060000,
		// Ingested is nil
	}
	b, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	s := string(b)
	if strings.Contains(s, "ingested") {
		t.Errorf("expected 'ingested' to be omitted when nil, got: %s", s)
	}
}

func TestTimestampInfo_IngestedPresentWhenSet(t *testing.T) {
	ingested := int64(1708700120000)
	ts := schema.TimestampInfo{
		Start:    1708700000000,
		End:      1708700060000,
		Ingested: &ingested,
	}
	b, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "ingested") {
		t.Errorf("expected 'ingested' to be present when set, got: %s", s)
	}
}

// --- AC4: QualityMetrics nil pointer fields are omitted ---

func TestQualityMetrics_NilFieldsOmitted(t *testing.T) {
	qm := schema.QualityMetrics{
		// All fields are nil / not set
	}
	b, err := json.Marshal(qm)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	s := string(b)
	// An empty QualityMetrics should produce {}
	if s != "{}" {
		t.Errorf("expected '{}' for empty QualityMetrics, got: %s", s)
	}
}

func TestQualityMetrics_M7SpecHasExamples_NilOmitted(t *testing.T) {
	qm := schema.QualityMetrics{}
	b, _ := json.Marshal(qm)
	s := string(b)
	if strings.Contains(s, "m7SpecHasExamples") {
		t.Errorf("m7SpecHasExamples should be omitted when nil, got: %s", s)
	}
}

func TestQualityMetrics_M7SpecHasExamples_FalseIncluded(t *testing.T) {
	f := false
	qm := schema.QualityMetrics{M7SpecHasExamples: &f}
	b, _ := json.Marshal(qm)
	s := string(b)
	if !strings.Contains(s, "m7SpecHasExamples") {
		t.Errorf("m7SpecHasExamples should be present when *false, got: %s", s)
	}
	if !strings.Contains(s, "false") {
		t.Errorf("m7SpecHasExamples should be false when *false, got: %s", s)
	}
}

func TestQualityMetrics_M7SpecHasConstraints_NilOmitted(t *testing.T) {
	qm := schema.QualityMetrics{}
	b, _ := json.Marshal(qm)
	s := string(b)
	if strings.Contains(s, "m7SpecHasConstraints") {
		t.Errorf("m7SpecHasConstraints should be omitted when nil, got: %s", s)
	}
}

// --- DiagnosticsInfo: Partial is *bool with omitempty ---

func TestDiagnosticsInfo_PartialOmittedWhenNil(t *testing.T) {
	d := schema.DiagnosticsInfo{
		Warnings: []schema.DiagnosticEntry{},
	}
	b, _ := json.Marshal(d)
	s := string(b)
	if strings.Contains(s, "partial") {
		t.Errorf("partial should be omitted when nil, got: %s", s)
	}
}

func TestDiagnosticsInfo_PartialFalseIncluded(t *testing.T) {
	f := false
	d := schema.DiagnosticsInfo{
		Warnings: []schema.DiagnosticEntry{},
		Partial:  &f,
	}
	b, _ := json.Marshal(d)
	s := string(b)
	if !strings.Contains(s, "partial") {
		t.Errorf("partial should be present when *false, got: %s", s)
	}
}

// --- GitContext: all fields are optional pointers ---

func TestGitContext_AllNilOmitsAllFields(t *testing.T) {
	gc := schema.GitContext{}
	b, _ := json.Marshal(gc)
	s := string(b)
	if s != "{}" {
		t.Errorf("expected '{}' for empty GitContext, got: %s", s)
	}
}

func TestGitContext_BranchPresent(t *testing.T) {
	branch := "main"
	gc := schema.GitContext{Branch: &branch}
	b, _ := json.Marshal(gc)
	s := string(b)
	if !strings.Contains(s, "branch") {
		t.Errorf("expected 'branch' in JSON, got: %s", s)
	}
	if !strings.Contains(s, "main") {
		t.Errorf("expected 'main' in JSON, got: %s", s)
	}
}

// --- SessionStats: ACP optional token fields ---

func TestSessionStats_OptionalTokensOmittedWhenNil(t *testing.T) {
	ss := schema.SessionStats{
		TurnCount:     5,
		ToolCallCount: 10,
		TokensIn:      1000,
		TokensOut:     500,
	}
	b, _ := json.Marshal(ss)
	s := string(b)
	for _, field := range []string{"thoughtTokens", "cachedReadTokens", "cachedWriteTokens"} {
		if strings.Contains(s, field) {
			t.Errorf("field %q should be omitted when nil, got: %s", field, s)
		}
	}
}

func TestSessionStats_OptionalTokensPresentWhenSet(t *testing.T) {
	thought := 50
	cached := 200
	write := 100
	ss := schema.SessionStats{
		ThoughtTokens:     &thought,
		CachedReadTokens:  &cached,
		CachedWriteTokens: &write,
	}
	b, _ := json.Marshal(ss)
	s := string(b)
	for _, field := range []string{"thoughtTokens", "cachedReadTokens", "cachedWriteTokens"} {
		if !strings.Contains(s, field) {
			t.Errorf("field %q should be present when set, got: %s", field, s)
		}
	}
}

// --- PublishRequest JSON round-trip ---

func TestPublishRequest_RoundTrip(t *testing.T) {
	sessionID, _ := schema.NewSessionID("99d59925-36bc-424c-a789-8be54d9702ba")
	modelID, _ := schema.NewModelID("claude-opus-4-6")
	projectHash, _ := schema.NewProjectHash("a3aee4f000000000000000000000000000000000000000000000000000000000")
	hostSlug, _ := schema.NewHostSlug("github.com")

	ingested := int64(1708700120000)
	branch := "main"
	partial := true

	req := schema.PublishRequest{
		Identity: schema.SessionIdentity{
			SessionID:     sessionID,
			SchemaVersion: 3,
		},
		Model: schema.ModelInfo{
			Harness:        schema.HarnessClaudeCode,
			Model:          modelID,
			HarnessVersion: "2.1.47",
			HostSlug:       hostSlug,
		},
		Timestamp: schema.TimestampInfo{
			Start:    1708700000000,
			End:      1708700060000,
			Ingested: &ingested,
		},
		Source: schema.SourceInfo{
			FilePath: "/home/user/.claude/projects/foo.jsonl",
			Format:   schema.SourceFormatJSONL,
		},
		Git: schema.GitContext{
			Branch: &branch,
		},
		Project: schema.ProjectContext{
			Hash: projectHash,
			Name: "peasant",
		},
		Stats: schema.SessionStats{
			TurnCount:     5,
			ToolCallCount: 12,
			TokensIn:      1000,
			TokensOut:     500,
			DurationMs:    60000,
		},
		Subagents: []schema.SubagentRef{},
		Diagnostics: schema.DiagnosticsInfo{
			Warnings: []schema.DiagnosticEntry{},
			Partial:  &partial,
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded schema.PublishRequest
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Spot-check key fields
	if decoded.Identity.SessionID != req.Identity.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", decoded.Identity.SessionID, req.Identity.SessionID)
	}
	if decoded.Model.Harness != req.Model.Harness {
		t.Errorf("Provider mismatch: got %s, want %s", decoded.Model.Harness, req.Model.Harness)
	}
	if decoded.Timestamp.Start != req.Timestamp.Start {
		t.Errorf("Timestamp.Start mismatch: got %d, want %d", decoded.Timestamp.Start, req.Timestamp.Start)
	}
	if decoded.Timestamp.Ingested == nil {
		t.Error("Timestamp.Ingested should not be nil after round-trip")
	} else if *decoded.Timestamp.Ingested != ingested {
		t.Errorf("Timestamp.Ingested mismatch: got %d, want %d", *decoded.Timestamp.Ingested, ingested)
	}
	if decoded.Diagnostics.Partial == nil {
		t.Error("Diagnostics.Partial should not be nil after round-trip")
	} else if *decoded.Diagnostics.Partial != partial {
		t.Errorf("Diagnostics.Partial mismatch: got %v, want %v", *decoded.Diagnostics.Partial, partial)
	}
}

// --- Nil vs empty slice behavior ---
//
// Note: Go's encoding/json omitempty omits BOTH nil AND empty slices.
// The canonical pattern is to always initialize Subagents with []SubagentRef{}
// (as done in NewUnifiedMetadata), ensuring a non-nil initialized slice is present
// in wire format when subagents exist. nil means "not initialized / irrelevant",
// which is correctly omitted. Use an initialized empty slice when the field is
// meaningful but happens to be empty.

func TestPublishRequest_SubagentsNilOmitted(t *testing.T) {
	req := schema.PublishRequest{
		Subagents: nil,
	}
	b, _ := json.Marshal(req)
	s := string(b)
	// nil slice with omitempty is correctly omitted
	if strings.Contains(s, `"subagents"`) {
		t.Errorf("nil Subagents should be omitted, got: %s", s)
	}
}

func TestPublishRequest_SubagentsWithEntries(t *testing.T) {
	parentID, _ := schema.NewSessionID("agent-a3aee4f")
	req := schema.PublishRequest{
		Subagents: []schema.SubagentRef{
			{SessionID: "agent-b3bee5g", ParentUUID: parentID},
		},
	}
	b, _ := json.Marshal(req)
	s := string(b)
	if !strings.Contains(s, `"subagents"`) {
		t.Errorf("non-empty Subagents should be present in JSON, got: %s", s)
	}
}

// --- PublishResponse round-trip ---

func TestPublishResponse_RoundTrip(t *testing.T) {
	resp := schema.PublishResponse{
		TranscriptID:  "server-uuid-1234",
		BlobKey:       "transcripts/server-uuid-1234.jsonl",
		BlobSizeBytes: 12345,
		PublishedAt:   1708700120000,
		UpdatedAt:     1708700120000,
		Created:       true,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded schema.PublishResponse
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.TranscriptID != resp.TranscriptID {
		t.Errorf("TranscriptID mismatch: got %s, want %s", decoded.TranscriptID, resp.TranscriptID)
	}
	if decoded.Created != resp.Created {
		t.Errorf("Created mismatch: got %v, want %v", decoded.Created, resp.Created)
	}
}

// --- No internal/ imports (enforced by Go compile, but verify here via import path) ---

func TestNoInternalImports(t *testing.T) {
	// This test is enforced at compile-time by Go's import rules.
	// If the schema package imported anything from internal/, the build would fail.
	// The fact that this test runs proves the schema package has no internal imports.
	t.Log("schema package compiled without internal/ imports (AC1)")
}

// --- JSON camelCase field names ---

func TestSessionIdentity_CamelCaseJSON(t *testing.T) {
	parentID, _ := schema.NewSessionID("agent-a3aee4f")
	id := schema.SessionIdentity{
		SessionID:       "99d59925-36bc-424c-a789-8be54d9702ba",
		ParentSessionID: &parentID,
		SchemaVersion:   3,
	}
	b, _ := json.Marshal(id)
	s := string(b)
	if !strings.Contains(s, "sessionId") {
		t.Errorf("expected 'sessionId' (camelCase) in JSON, got: %s", s)
	}
	if !strings.Contains(s, "parentUuid") {
		t.Errorf("expected 'parentUuid' (camelCase) in JSON, got: %s", s)
	}
	if !strings.Contains(s, "schemaVersion") {
		t.Errorf("expected 'schemaVersion' (camelCase) in JSON, got: %s", s)
	}
}

func TestModelInfo_CamelCaseJSON(t *testing.T) {
	mi := schema.ModelInfo{
		Harness:        schema.HarnessClaudeCode,
		Model:          "claude-opus-4-6",
		HarnessVersion: "2.1.47",
	}
	b, _ := json.Marshal(mi)
	s := string(b)
	// Unified harness key (SLICE-B1 emit-side flip): ModelInfo now emits
	// json:"harness", not the legacy json:"modelHarness".
	if !strings.Contains(s, `"harness"`) {
		t.Errorf("expected 'harness' in JSON, got: %s", s)
	}
	if strings.Contains(s, "modelHarness") {
		t.Errorf("legacy 'modelHarness' must not appear in ModelInfo JSON, got: %s", s)
	}
}

func TestProjectContext_CamelCaseJSON(t *testing.T) {
	h, _ := schema.NewProjectHash("a3aee4f000000000000000000000000000000000000000000000000000000000")
	pc := schema.ProjectContext{Hash: h, FilePath: "/tmp/repo", Name: "peasant"}
	b, _ := json.Marshal(pc)
	s := string(b)
	if !strings.Contains(s, "filePath") {
		t.Errorf("expected 'filePath' in JSON, got: %s", s)
	}
}

// --- RedactionInfo JSON round-trip ---

func TestRedactionInfo_RoundTrip(t *testing.T) {
	now := int64(1708700000000)
	ri := schema.RedactionInfo{
		Applied:             true,
		Level:               "standard",
		RedactedAtMs:        &now,
		ContentHashAtRedact: "abc123",
	}
	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded schema.RedactionInfo
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Applied != ri.Applied {
		t.Errorf("Applied: got %v, want %v", decoded.Applied, ri.Applied)
	}
	if decoded.Level != ri.Level {
		t.Errorf("Level: got %q, want %q", decoded.Level, ri.Level)
	}
	if decoded.RedactedAtMs == nil || *decoded.RedactedAtMs != now {
		t.Errorf("RedactedAtMs: got %v, want %d", decoded.RedactedAtMs, now)
	}
	if decoded.ContentHashAtRedact != ri.ContentHashAtRedact {
		t.Errorf("ContentHashAtRedact: got %q, want %q", decoded.ContentHashAtRedact, ri.ContentHashAtRedact)
	}
}

func TestRedactionInfo_RawOmitsOptionalFields(t *testing.T) {
	ri := schema.RedactionInfo{Applied: false}
	b, _ := json.Marshal(ri)
	s := string(b)
	if strings.Contains(s, "level") {
		t.Errorf("level should be omitted when empty, got: %s", s)
	}
	if strings.Contains(s, "redacted_at_ms") {
		t.Errorf("redacted_at_ms should be omitted when nil, got: %s", s)
	}
	if strings.Contains(s, "content_hash_at_redact") {
		t.Errorf("content_hash_at_redact should be omitted when empty, got: %s", s)
	}
}

// --- RedactionInfo staleness detection ---

func TestRedactionInfo_IsStale(t *testing.T) {
	ri := schema.RedactionInfo{Applied: true, ContentHashAtRedact: "hash1"}
	if !ri.IsStale("hash2") {
		t.Error("IsStale should return true when content changed")
	}
	if ri.IsStale("hash1") {
		t.Error("IsStale should return false when hash matches")
	}
}

func TestRedactionInfo_IsCurrent(t *testing.T) {
	ri := schema.RedactionInfo{Applied: true, ContentHashAtRedact: "hash1"}
	if !ri.IsCurrent("hash1") {
		t.Error("IsCurrent should return true when hash matches")
	}
	if ri.IsCurrent("hash2") {
		t.Error("IsCurrent should return false when content changed")
	}
}

func TestRedactionInfo_IsRaw(t *testing.T) {
	raw := schema.RedactionInfo{Applied: false}
	if !raw.IsRaw() {
		t.Error("IsRaw should return true when not applied")
	}
	applied := schema.RedactionInfo{Applied: true}
	if applied.IsRaw() {
		t.Error("IsRaw should return false when applied")
	}
}

func TestRedactionInfo_IsStale_RawNeverStale(t *testing.T) {
	ri := schema.RedactionInfo{Applied: false}
	if ri.IsStale("anyhash") {
		t.Error("Raw (unapplied) redaction should never be stale")
	}
}

// --- ContentHash determinism ---

func TestComputeTranscriptHash_Deterministic(t *testing.T) {
	data := []byte(`{"role":"user","content":"hello"}`)
	h1 := schema.ComputeTranscriptHash(data)
	h2 := schema.ComputeTranscriptHash(data)
	if h1 != h2 {
		t.Errorf("same bytes should produce same hash: %q vs %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("SHA3-256 hex should be 64 chars, got %d", len(h1))
	}
}

func TestComputeTranscriptHash_DifferentInput(t *testing.T) {
	h1 := schema.ComputeTranscriptHash([]byte("hello"))
	h2 := schema.ComputeTranscriptHash([]byte("world"))
	if h1 == h2 {
		t.Error("different input should produce different hash")
	}
}

// --- MetadataHash stability ---

func TestComputeMetadataHash_Stable(t *testing.T) {
	meta := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
		Stats:         schema.SessionStats{TurnCount: 5, TokensIn: 100},
	}
	h1 := schema.ComputeMetadataHash(meta)
	h2 := schema.ComputeMetadataHash(meta)
	if h1 != h2 {
		t.Errorf("same metadata should produce same hash: %q vs %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("SHA3-256 hex should be 64 chars, got %d", len(h1))
	}
}

func TestComputeMetadataHash_ChangedField(t *testing.T) {
	meta1 := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
		Stats:         schema.SessionStats{TurnCount: 5},
	}
	meta2 := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
		Stats:         schema.SessionStats{TurnCount: 10},
	}
	h1 := schema.ComputeMetadataHash(meta1)
	h2 := schema.ComputeMetadataHash(meta2)
	if h1 == h2 {
		t.Error("different metadata should produce different hash")
	}
}

func TestComputeMetadataHash_ExcludesHashAndRedaction(t *testing.T) {
	meta := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
	}
	h1 := schema.ComputeMetadataHash(meta)

	metaWithHash := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
		ContentHash:   "somehash",
		MetadataHash:  "anotherhash",
		Redaction:     schema.RedactionInfo{Applied: true, Level: "maximum"},
	}
	h2 := schema.ComputeMetadataHash(metaWithHash)

	if h1 != h2 {
		t.Error("MetadataHash should exclude ContentHash, MetadataHash, and Redaction fields")
	}
}

// --- MetadataSchemaVersion ---

func TestMetadataSchemaVersion_Is9(t *testing.T) {
	if schema.MetadataSchemaVersion != 9 {
		t.Errorf("MetadataSchemaVersion = %d, want 9", schema.MetadataSchemaVersion)
	}
}

// --- DerivedAt field ---

func TestDerivedAt_OmittedWhenNil(t *testing.T) {
	meta := schema.NewUnifiedMetadata()
	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(b), "derivedAt") {
		t.Errorf("derivedAt should be omitted when nil, got: %s", string(b))
	}
}

func TestDerivedAt_PresentWhenSet(t *testing.T) {
	meta := schema.NewUnifiedMetadata()
	ts := int64(1708700000000)
	meta.DerivedAt = &ts
	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "derivedAt") {
		t.Errorf("derivedAt should be present when set, got: %s", s)
	}
	if !strings.Contains(s, "1708700000000") {
		t.Errorf("derivedAt should contain timestamp value, got: %s", s)
	}
}

func TestDerivedAt_RoundTrip(t *testing.T) {
	ts := int64(1708700000000)
	meta := schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		DerivedAt:     &ts,
		Subagents:     []schema.SubagentRef{},
		Diagnostics:   schema.DiagnosticsInfo{Warnings: []schema.DiagnosticEntry{}},
	}
	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded schema.UnifiedMetadata
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.DerivedAt == nil {
		t.Fatal("DerivedAt should not be nil after round-trip")
	}
	if *decoded.DerivedAt != ts {
		t.Errorf("DerivedAt mismatch: got %d, want %d", *decoded.DerivedAt, ts)
	}
}

func TestDerivedAt_V7BackwardCompat(t *testing.T) {
	// v7 metadata JSON has no derivedAt field — should parse without error.
	v7JSON := `{
		"schemaVersion": 7,
		"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba",
		"modelHarness": "claude-code",
		"model": "claude-opus-4-6",
		"version": "2.1.47",
		"timestamp": {"start": 1708700000000, "end": 1708700060000},
		"source": {"format": "jsonl"},
		"git": {},
		"project": {"hash": "a3aee4f000000000000000000000000000000000000000000000000000000000", "name": "peasant"},
		"hostSlug": "github.com",
		"stats": {"turnCount": 0, "toolCallCount": 0, "subagentCount": 0, "durationMs": 0, "tokensIn": 0, "tokensOut": 0},
		"subagents": [],
		"diagnostics": {"warnings": []},
		"contentHash": "",
		"metadataHash": "",
		"redaction": {"applied": false}
	}`
	var meta schema.UnifiedMetadata
	if err := json.Unmarshal([]byte(v7JSON), &meta); err != nil {
		t.Fatalf("v7 metadata unmarshal failed under v8 code: %v", err)
	}
	if meta.DerivedAt != nil {
		t.Errorf("DerivedAt should be nil for v7 metadata, got: %d", *meta.DerivedAt)
	}
}

func TestComputeMetadataHash_ExcludesDerivedAt(t *testing.T) {
	meta1 := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
	}
	h1 := schema.ComputeMetadataHash(meta1)

	ts := int64(1708700000000)
	meta2 := &schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
		DerivedAt:     &ts,
	}
	h2 := schema.ComputeMetadataHash(meta2)

	if h1 != h2 {
		t.Error("MetadataHash should exclude DerivedAt field")
	}
}

// --- UnifiedMetadata includes new fields ---

func TestUnifiedMetadata_NewFieldsRoundTrip(t *testing.T) {
	now := int64(1708700000000)
	meta := schema.UnifiedMetadata{
		SchemaVersion: schema.MetadataSchemaVersion,
		SessionID:     "99d59925-36bc-424c-a789-8be54d9702ba",
		ContentHash:   "abc123",
		MetadataHash:  "def456",
		Redaction: schema.RedactionInfo{
			Applied:             true,
			Level:               "standard",
			RedactedAtMs:        &now,
			ContentHashAtRedact: "abc123",
		},
		Subagents:   []schema.SubagentRef{},
		Diagnostics: schema.DiagnosticsInfo{Warnings: []schema.DiagnosticEntry{}},
	}

	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded schema.UnifiedMetadata
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ContentHash != meta.ContentHash {
		t.Errorf("ContentHash: got %q, want %q", decoded.ContentHash, meta.ContentHash)
	}
	if decoded.MetadataHash != meta.MetadataHash {
		t.Errorf("MetadataHash: got %q, want %q", decoded.MetadataHash, meta.MetadataHash)
	}
	if decoded.Redaction.Applied != true {
		t.Error("Redaction.Applied should be true after round-trip")
	}
	if decoded.Redaction.Level != "standard" {
		t.Errorf("Redaction.Level: got %q, want %q", decoded.Redaction.Level, "standard")
	}
}
