package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// jsonKeys returns the top-level object keys of marshaled JSON for key-presence
// assertions. Fails the test if v does not marshal to a JSON object.
func jsonKeys(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal to object failed: %v (json=%s)", err, b)
	}
	return m
}

// --- D2: TranscriptContent envelope round-trip ---

// TestTranscriptContent_RoundTrip verifies the structured push wire body
// serializes with contractVersion + kind + a nested sessionDetail carrying its
// own embedded schemaVersion, and round-trips back to an equal value.
func TestTranscriptContent_RoundTrip(t *testing.T) {
	env := schema.TranscriptContent{
		ContractVersion: schema.PushContractVersion("0.1.0"),
		Kind:            schema.ContentKindSessionDetail,
		SessionDetail: &schema.SessionDetailPayload{
			SchemaVersion: schema.PushContractVersion("0.1.0"),
			ID:            "99d59925-36bc-424c-a789-8be54d9702ba",
			Harness:       schema.HarnessClaudeCode,
			Turns:         []schema.TurnDetail{},
		},
	}

	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope failed: %v", err)
	}
	s := string(b)

	for _, want := range []string{`"contractVersion":"0.1.0"`, `"kind":"session_detail"`, `"sessionDetail"`} {
		if !strings.Contains(s, want) {
			t.Errorf("envelope JSON missing %s; got: %s", want, s)
		}
	}
	// The nested payload must carry its embedded schemaVersion AND the unified
	// harness key (NOT the legacy provider key).
	if !strings.Contains(s, `"schemaVersion":"0.1.0"`) {
		t.Errorf("nested sessionDetail missing embedded schemaVersion; got: %s", s)
	}
	if !strings.Contains(s, `"harness":"claude-code"`) {
		t.Errorf("nested sessionDetail missing unified harness key; got: %s", s)
	}

	var decoded schema.TranscriptContent
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal envelope failed: %v", err)
	}
	if decoded.ContractVersion != env.ContractVersion {
		t.Errorf("ContractVersion: got %q, want %q", decoded.ContractVersion, env.ContractVersion)
	}
	if decoded.Kind != schema.ContentKindSessionDetail {
		t.Errorf("Kind: got %q, want %q", decoded.Kind, schema.ContentKindSessionDetail)
	}
	if decoded.SessionDetail == nil {
		t.Fatal("SessionDetail nil after round-trip")
	}
	if decoded.SessionDetail.SchemaVersion != env.SessionDetail.SchemaVersion {
		t.Errorf("embedded SchemaVersion: got %q, want %q",
			decoded.SessionDetail.SchemaVersion, env.SessionDetail.SchemaVersion)
	}
	if decoded.SessionDetail.Harness != schema.HarnessClaudeCode {
		t.Errorf("embedded Harness: got %q, want %q",
			decoded.SessionDetail.Harness, schema.HarnessClaudeCode)
	}
}

// TestContentKind_SessionDetailValue pins the wire value of the only emitted kind.
func TestContentKind_SessionDetailValue(t *testing.T) {
	if got := schema.ContentKindSessionDetail.String(); got != "session_detail" {
		t.Errorf("ContentKindSessionDetail: got %q, want %q", got, "session_detail")
	}
}

// --- Δ1: embedded SchemaVersion presence + omitempty behavior ---

// TestSessionDetailPayload_EmbeddedSchemaVersion verifies the embedded
// schemaVersion is emitted when set, and omitted when empty (so the local
// WebSocket session_detail channel does not gain a spurious empty field).
func TestSessionDetailPayload_EmbeddedSchemaVersion(t *testing.T) {
	withVersion := jsonKeys(t, schema.SessionDetailPayload{
		SchemaVersion: schema.PushContractVersion("0.1.0"),
		ID:            "s1",
		Harness:       schema.HarnessClaudeCode,
	})
	if _, ok := withVersion["schemaVersion"]; !ok {
		t.Error("expected schemaVersion key present when set")
	}

	withoutVersion := jsonKeys(t, schema.SessionDetailPayload{ID: "s1", Harness: schema.HarnessClaudeCode})
	if _, ok := withoutVersion["schemaVersion"]; ok {
		t.Error("expected schemaVersion key OMITTED when empty (omitempty)")
	}
}

// --- Emit-side provider->harness KEY flip (FAILS until L3) ---

// TestSessionEntry_UnifiedHarnessKey asserts the content-wire SessionEntry emits
// the harness under json:"harness" (the key the village reads + the envelope
// migrate-on-read sniff key), NOT the legacy json:"provider".
func TestSessionEntry_UnifiedHarnessKey(t *testing.T) {
	keys := jsonKeys(t, schema.SessionEntry{
		SessionID: "99d59925-36bc-424c-a789-8be54d9702ba",
		Harness:   schema.HarnessClaudeCode,
	})
	if _, ok := keys["harness"]; !ok {
		t.Error("SessionEntry must emit json:\"harness\" (unified key)")
	}
	if _, ok := keys["provider"]; ok {
		t.Error("SessionEntry must NOT emit legacy json:\"provider\" (flip incomplete)")
	}
}

// TestModelInfo_UnifiedHarnessKey asserts ModelInfo emits json:"harness", NOT
// the legacy json:"modelHarness".
func TestModelInfo_UnifiedHarnessKey(t *testing.T) {
	keys := jsonKeys(t, schema.ModelInfo{Harness: schema.HarnessClaudeCode, Model: "claude-opus-4-6"})
	if _, ok := keys["harness"]; !ok {
		t.Error("ModelInfo must emit json:\"harness\" (unified key)")
	}
	if _, ok := keys["modelHarness"]; ok {
		t.Error("ModelInfo must NOT emit legacy json:\"modelHarness\" (flip incomplete)")
	}
}

// TestUnifiedMetadata_UnifiedHarnessKey asserts UnifiedMetadata emits
// json:"harness", NOT the legacy json:"modelHarness".
func TestUnifiedMetadata_UnifiedHarnessKey(t *testing.T) {
	meta := schema.NewUnifiedMetadata()
	meta.ModelHarness = schema.HarnessClaudeCode
	keys := jsonKeys(t, meta)
	if _, ok := keys["harness"]; !ok {
		t.Error("UnifiedMetadata must emit json:\"harness\" (unified key)")
	}
	if _, ok := keys["modelHarness"]; ok {
		t.Error("UnifiedMetadata must NOT emit legacy json:\"modelHarness\" (flip incomplete)")
	}
}

// --- TRAP-KEY PRESERVATION (regression guard; must ALWAYS pass) ---

// TestAnnotatorSummary_ProviderKeyPreserved guards the model-VENDOR credential
// key: AnnotatorSummary.ProviderKey is json:"providerKey" and MUST NOT be flipped
// to harness — it is NOT the coding-tool harness (it is the model provider
// credential). This is the SLICE-B1 trap (annotation.go:77).
func TestAnnotatorSummary_ProviderKeyPreserved(t *testing.T) {
	pk := "anthropic"
	keys := jsonKeys(t, schema.AnnotatorSummary{ProviderKey: &pk})
	if _, ok := keys["providerKey"]; !ok {
		t.Error("AnnotatorSummary MUST preserve json:\"providerKey\" (model-vendor credential TRAP)")
	}
	if _, ok := keys["harness"]; ok {
		t.Error("AnnotatorSummary.ProviderKey MUST NOT be flipped to json:\"harness\" (TRAP violated)")
	}
}
