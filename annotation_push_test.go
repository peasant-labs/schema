package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// --- JSON round-trip tests ---

// TestAnnotationPushItem_JSONRoundTrip verifies that marshaling and unmarshaling
// an AnnotationPushItem preserves all fields.
func TestAnnotationPushItem_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	sessionID := "test-session-id"
	confidence := 0.95
	reason := "high overlap"
	provenance := &schema.Provenance{
		Method:   "heuristic",
		Function: "computeRetryLoops",
		Version:  "1.0",
	}

	item := schema.AnnotationPushItem{
		ContentHash:   "abc123",
		TargetKind:    schema.TargetSession,
		SessionID:     &sessionID,
		TypeID:        "retry_loops",
		Value:         "3",
		IsPrimary:     true,
		Confidence:    &confidence,
		Reason:        &reason,
		AnnotatorName: "system",
		Provenance:    provenance,
	}

	b, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal AnnotationPushItem: %v", err)
	}

	var got schema.AnnotationPushItem
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal AnnotationPushItem: %v", err)
	}

	if got.ContentHash != item.ContentHash {
		t.Errorf("ContentHash: want %q, got %q", item.ContentHash, got.ContentHash)
	}
	if got.TargetKind != item.TargetKind {
		t.Errorf("TargetKind: want %q, got %q", item.TargetKind, got.TargetKind)
	}
	if got.SessionID == nil || *got.SessionID != sessionID {
		t.Errorf("SessionID: want %q, got %v", sessionID, got.SessionID)
	}
	if got.TypeID != item.TypeID {
		t.Errorf("TypeID: want %q, got %q", item.TypeID, got.TypeID)
	}
	if got.Value != item.Value {
		t.Errorf("Value: want %q, got %q", item.Value, got.Value)
	}
	if got.IsPrimary != item.IsPrimary {
		t.Errorf("IsPrimary: want %v, got %v", item.IsPrimary, got.IsPrimary)
	}
	if got.Confidence == nil || *got.Confidence != confidence {
		t.Errorf("Confidence: want %v, got %v", confidence, got.Confidence)
	}
	if got.Reason == nil || *got.Reason != reason {
		t.Errorf("Reason: want %q, got %v", reason, got.Reason)
	}
	if got.AnnotatorName != item.AnnotatorName {
		t.Errorf("AnnotatorName: want %q, got %q", item.AnnotatorName, got.AnnotatorName)
	}
	if got.Provenance == nil || got.Provenance.Method != provenance.Method {
		t.Errorf("Provenance.Method: want %q, got %v", provenance.Method, got.Provenance)
	}
}

// TestAnnotationPushRequest_JSONRoundTrip verifies that marshaling and unmarshaling
// an AnnotationPushRequest preserves the annotations slice.
func TestAnnotationPushRequest_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	sessionID := "session-abc"
	req := schema.AnnotationPushRequest{
		Annotations: []schema.AnnotationPushItem{
			{
				ContentHash: "hash1",
				TargetKind:  schema.TargetSession,
				SessionID:   &sessionID,
				TypeID:      "retry_loops",
				Value:       "2",
				IsPrimary:   true,
			},
			{
				ContentHash: "hash2",
				TargetKind:  schema.TargetSession,
				SessionID:   &sessionID,
				TypeID:      "outcome",
				Value:       "success",
				IsPrimary:   true,
			},
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal AnnotationPushRequest: %v", err)
	}

	var got schema.AnnotationPushRequest
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal AnnotationPushRequest: %v", err)
	}

	if len(got.Annotations) != len(req.Annotations) {
		t.Fatalf("Annotations len: want %d, got %d", len(req.Annotations), len(got.Annotations))
	}
	for i, want := range req.Annotations {
		if got.Annotations[i].ContentHash != want.ContentHash {
			t.Errorf("[%d] ContentHash: want %q, got %q", i, want.ContentHash, got.Annotations[i].ContentHash)
		}
		if got.Annotations[i].TypeID != want.TypeID {
			t.Errorf("[%d] TypeID: want %q, got %q", i, want.TypeID, got.Annotations[i].TypeID)
		}
	}
}

// TestAnnotationPushResponse_JSONRoundTrip verifies that marshaling and unmarshaling
// an AnnotationPushResponse preserves all aggregate fields and per-item results.
func TestAnnotationPushResponse_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	resp := schema.AnnotationPushResponse{
		Created: 3,
		Updated: 1,
		Skipped: 2,
		Errors:  0,
		Results: []schema.AnnotationPushResult{
			{ContentHash: "hash1", Status: schema.PushStatusCreated},
			{ContentHash: "hash2", Status: schema.PushStatusUpdated},
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal AnnotationPushResponse: %v", err)
	}

	var got schema.AnnotationPushResponse
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal AnnotationPushResponse: %v", err)
	}

	if got.Created != resp.Created {
		t.Errorf("Created: want %d, got %d", resp.Created, got.Created)
	}
	if got.Updated != resp.Updated {
		t.Errorf("Updated: want %d, got %d", resp.Updated, got.Updated)
	}
	if got.Skipped != resp.Skipped {
		t.Errorf("Skipped: want %d, got %d", resp.Skipped, got.Skipped)
	}
	if got.Errors != resp.Errors {
		t.Errorf("Errors: want %d, got %d", resp.Errors, got.Errors)
	}
	if len(got.Results) != len(resp.Results) {
		t.Fatalf("Results len: want %d, got %d", len(resp.Results), len(got.Results))
	}
	if got.Results[0].ContentHash != "hash1" || got.Results[0].Status != schema.PushStatusCreated {
		t.Errorf("Results[0]: want {hash1, %s}, got %+v", schema.PushStatusCreated, got.Results[0])
	}
}

// TestAnnotationPushItem_OmitEmpty verifies that optional fields (Confidence,
// Reason, SessionID, EntryTarget, AnnotationID, ProjectHash, Provenance) are
// absent from the JSON output when they are nil.
func TestAnnotationPushItem_OmitEmpty(t *testing.T) {
	t.Parallel()
	item := schema.AnnotationPushItem{
		ContentHash: "abc",
		TargetKind:  schema.TargetSession,
		TypeID:      "outcome",
		Value:       "success",
		IsPrimary:   false,
	}

	b, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal AnnotationPushItem: %v", err)
	}
	s := string(b)

	for _, field := range []string{"sessionId", "entryTarget", "annotationId", "projectHash", "confidence", "reason", "provenance"} {
		if strings.Contains(s, `"`+field+`"`) {
			t.Errorf("JSON should not contain %q when nil; got: %s", field, s)
		}
	}
}

// --- Content hash tests ---

// TestComputeContentHash_Deterministic verifies that the same AnnotationPushItem
// always produces the same content hash.
func TestComputeContentHash_Deterministic(t *testing.T) {
	t.Parallel()
	sessionID := "session-abc"
	item := schema.AnnotationPushItem{
		TargetKind:    schema.TargetSession,
		SessionID:     &sessionID,
		TypeID:        "retry_loops",
		Value:         "3",
		IsPrimary:     true,
		AnnotatorName: "system",
	}

	hash1 := item.ComputeContentHash()
	hash2 := item.ComputeContentHash()
	hash3 := item.ComputeContentHash()

	if hash1 == "" {
		t.Fatal("ComputeContentHash returned empty string")
	}
	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("ComputeContentHash not deterministic: %q, %q, %q", hash1, hash2, hash3)
	}
}

// TestComputeContentHash_FieldChange verifies that changing any significant field
// in AnnotationPushItem produces a different content hash.
func TestComputeContentHash_FieldChange(t *testing.T) {
	t.Parallel()
	sessionID := "session-abc"
	base := schema.AnnotationPushItem{
		TargetKind:    schema.TargetSession,
		SessionID:     &sessionID,
		TypeID:        "retry_loops",
		Value:         "3",
		IsPrimary:     true,
		AnnotatorName: "system",
	}
	baseHash := base.ComputeContentHash()

	cases := []struct {
		name    string
		mutator func(schema.AnnotationPushItem) schema.AnnotationPushItem
	}{
		{
			name: "TypeID changed",
			mutator: func(item schema.AnnotationPushItem) schema.AnnotationPushItem {
				item.TypeID = "outcome"
				return item
			},
		},
		{
			name: "Value changed",
			mutator: func(item schema.AnnotationPushItem) schema.AnnotationPushItem {
				item.Value = "5"
				return item
			},
		},
		{
			name: "IsPrimary changed",
			mutator: func(item schema.AnnotationPushItem) schema.AnnotationPushItem {
				item.IsPrimary = false
				return item
			},
		},
		{
			name: "AnnotatorName changed",
			mutator: func(item schema.AnnotationPushItem) schema.AnnotationPushItem {
				item.AnnotatorName = "user"
				return item
			},
		},
		{
			name: "TargetKind changed",
			mutator: func(item schema.AnnotationPushItem) schema.AnnotationPushItem {
				item.TargetKind = schema.TargetProject
				item.SessionID = nil
				return item
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			modified := tc.mutator(base)
			modifiedHash := modified.ComputeContentHash()
			if modifiedHash == baseHash {
				t.Errorf("hash unchanged after mutation %q: both are %q", tc.name, baseHash)
			}
		})
	}
}

// TestComputeContentHash_JSONOrdering verifies that the struct-field ordering
// guarantee of encoding/json produces deterministic hashes, and that nil vs
// non-nil optional fields produce distinct hashes.
func TestComputeContentHash_JSONOrdering(t *testing.T) {
	t.Parallel()
	// Two items identical except SessionID is nil vs non-nil.
	// TargetKind is the same to isolate the nil-field effect.
	sessionA := "session-a"
	withSession := schema.AnnotationPushItem{
		TargetKind: schema.TargetSession,
		SessionID:  &sessionA,
		TypeID:     "retry_loops",
		Value:      "1",
	}
	withoutSession := schema.AnnotationPushItem{
		TargetKind: schema.TargetSession,
		SessionID:  nil,
		TypeID:     "retry_loops",
		Value:      "1",
	}

	if withSession.ComputeContentHash() == withoutSession.ComputeContentHash() {
		t.Error("nil vs non-nil SessionID should produce different hashes")
	}

	// Same call twice must produce same result (documents struct-ordering guarantee).
	h1 := withSession.ComputeContentHash()
	h2 := withSession.ComputeContentHash()
	if h1 != h2 {
		t.Errorf("repeated calls not deterministic: %q vs %q", h1, h2)
	}
}

// TestComputeContentHash_NotAffectedByContentHashField verifies that setting
// ContentHash to a non-empty value does NOT change the computed hash, because
// ComputeContentHash zeros it before hashing.
func TestComputeContentHash_NotAffectedByContentHashField(t *testing.T) {
	t.Parallel()
	sessionID := "session-abc"
	item := schema.AnnotationPushItem{
		TargetKind:    schema.TargetSession,
		SessionID:     &sessionID,
		TypeID:        "retry_loops",
		Value:         "3",
		IsPrimary:     true,
		AnnotatorName: "system",
	}

	// Hash with ContentHash empty.
	hash1 := item.ComputeContentHash()

	// Set ContentHash to some previous value (as if this item had been hashed before).
	item.ContentHash = hash1
	hash2 := item.ComputeContentHash()

	if hash1 != hash2 {
		t.Errorf("ContentHash field should not affect computed hash: %q != %q", hash1, hash2)
	}
}
