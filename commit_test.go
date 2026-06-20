package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// --- SLICE-1-L2: CommitInfo JSON round-trip ---

func TestCommitInfoMarshalUnmarshal(t *testing.T) {
	ci := schema.CommitInfo{
		Hash:        "abc123def456abc123def456abc123def456abc1",
		Message:     "feat: add session-to-commit linking",
		AuthorName:  "Alice Dev",
		AuthorEmail: "alice@example.com",
		CommitTime:  1708700060000,
		AuthorTime:  1708700050000,
	}

	b, err := json.Marshal(ci)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	s := string(b)
	for _, want := range []string{"hash", "message", "authorName", "authorEmail", "commitTime", "authorTime"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected JSON key %q in output, got: %s", want, s)
		}
	}

	var decoded schema.CommitInfo
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Hash != ci.Hash {
		t.Errorf("Hash: got %q, want %q", decoded.Hash, ci.Hash)
	}
	if decoded.Message != ci.Message {
		t.Errorf("Message: got %q, want %q", decoded.Message, ci.Message)
	}
	if decoded.AuthorName != ci.AuthorName {
		t.Errorf("AuthorName: got %q, want %q", decoded.AuthorName, ci.AuthorName)
	}
	if decoded.AuthorEmail != ci.AuthorEmail {
		t.Errorf("AuthorEmail: got %q, want %q", decoded.AuthorEmail, ci.AuthorEmail)
	}
	if decoded.CommitTime != ci.CommitTime {
		t.Errorf("CommitTime: got %d, want %d", decoded.CommitTime, ci.CommitTime)
	}
	if decoded.AuthorTime != ci.AuthorTime {
		t.Errorf("AuthorTime: got %d, want %d", decoded.AuthorTime, ci.AuthorTime)
	}
}

// --- SLICE-1-L2: v3 metadata reads without error under v4 code ---

// TestMetadataSchemaV3V4Compatibility verifies that a v3 metadata JSON blob
// (which has no "commits" field) can be unmarshalled by v4 code without error,
// and that Git.Commits is nil (safe zero value).
func TestMetadataSchemaV3V4Compatibility(t *testing.T) {
	// Minimal v3 metadata JSON — no "commits" field in "git".
	v3JSON := `{
		"schemaVersion": 3,
		"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba",
		"modelHarness": "claude-code",
		"model": "claude-opus-4-6",
		"version": "2.1.47",
		"timestamp": {"start": 1708700000000, "end": 1708700060000},
		"source": {"format": "jsonl"},
		"git": {"branch": "main"},
		"project": {"hash": "a3aee4f000000000000000000000000000000000000000000000000000000000", "name": "peasant"},
		"hostSlug": "github.com",
		"stats": {"turnCount": 0, "toolCallCount": 0, "subagentCount": 0, "durationMs": 0, "tokensIn": 0, "tokensOut": 0},
		"subagents": [],
		"diagnostics": {"warnings": []}
	}`

	var meta schema.UnifiedMetadata
	if err := json.Unmarshal([]byte(v3JSON), &meta); err != nil {
		t.Fatalf("v3 metadata unmarshal failed under v4 code: %v", err)
	}

	// Commits field must be nil (absent in v3) — safe zero value.
	if meta.Git.Commits != nil {
		t.Errorf("expected Git.Commits to be nil for v3 metadata, got: %v", meta.Git.Commits)
	}
}

// --- SLICE-1-L2: GitContext.Commits is present and optional (omitempty) ---

func TestGitContextCommitsField_OmittedWhenNil(t *testing.T) {
	gc := schema.GitContext{} // no commits
	b, _ := json.Marshal(gc)
	s := string(b)
	if strings.Contains(s, "commits") {
		t.Errorf("'commits' should be omitted when nil, got: %s", s)
	}
}

func TestGitContextCommitsField_PresentWhenSet(t *testing.T) {
	gc := schema.GitContext{
		Commits: []schema.CommitInfo{
			{
				Hash:        "abc1234",
				Message:     "fix: patch",
				AuthorName:  "Bob",
				AuthorEmail: "bob@example.com",
				CommitTime:  1708700060000,
				AuthorTime:  1708700050000,
			},
		},
	}
	b, _ := json.Marshal(gc)
	s := string(b)
	if !strings.Contains(s, "commits") {
		t.Errorf("'commits' should be present when non-nil, got: %s", s)
	}
	if !strings.Contains(s, "abc1234") {
		t.Errorf("commit hash should appear in JSON, got: %s", s)
	}
}

func TestGitContextCommitsField_EmptySliceOmitted(t *testing.T) {
	// An empty (non-nil) slice with omitempty is omitted by encoding/json.
	gc := schema.GitContext{Commits: []schema.CommitInfo{}}
	b, _ := json.Marshal(gc)
	s := string(b)
	// encoding/json omitempty omits both nil and empty slices.
	if strings.Contains(s, "commits") {
		t.Errorf("'commits' should be omitted for empty slice with omitempty, got: %s", s)
	}
}

// --- MetadataSchemaVersion is 9 (v9: harness key unification — self-heal stale modelHarness) ---

func TestMetadataSchemaVersion_IsV9(t *testing.T) {
	if schema.MetadataSchemaVersion != 9 {
		t.Errorf("MetadataSchemaVersion: got %d, want 9", schema.MetadataSchemaVersion)
	}
}

// --- NewUnifiedMetadata sets v9 schema version ---

func TestNewUnifiedMetadata_SchemaVersionV9(t *testing.T) {
	meta := schema.NewUnifiedMetadata()
	if meta.SchemaVersion != 9 {
		t.Errorf("NewUnifiedMetadata().SchemaVersion: got %d, want 9", meta.SchemaVersion)
	}
}

// --- Back-compat read: pre-v9 on-disk metadata keyed the harness as "modelHarness" ---

func TestUnifiedMetadata_LegacyModelHarnessKey_BackCompat(t *testing.T) {
	// A pre-v9 (v8) on-disk blob: harness under the legacy "modelHarness" key.
	v8JSON := `{
		"schemaVersion": 8,
		"sessionId": "99d59925-36bc-424c-a789-8be54d9702ba",
		"modelHarness": "claude-code",
		"model": "claude-opus-4-6"
	}`
	var meta schema.UnifiedMetadata
	if err := json.Unmarshal([]byte(v8JSON), &meta); err != nil {
		t.Fatalf("legacy v8 metadata unmarshal failed: %v", err)
	}
	if meta.ModelHarness != schema.HarnessClaudeCode {
		t.Errorf("legacy modelHarness not adopted: got %q, want %q", meta.ModelHarness, schema.HarnessClaudeCode)
	}

	// The canonical v9 "harness" key wins when both are present.
	v9JSON := `{"schemaVersion": 9, "harness": "opencode", "modelHarness": "claude-code"}`
	var meta9 schema.UnifiedMetadata
	if err := json.Unmarshal([]byte(v9JSON), &meta9); err != nil {
		t.Fatalf("v9 metadata unmarshal failed: %v", err)
	}
	if meta9.ModelHarness != schema.HarnessOpenCode {
		t.Errorf("canonical harness key must win: got %q, want %q", meta9.ModelHarness, schema.HarnessOpenCode)
	}
}
