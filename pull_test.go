package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// File-level test constants. This schema module is a separate (leaf) module and
// cannot import peasant's internal/testutil, so the CLAUDE.md rule for that case applies: define
// file-level constants instead of repeating inline literals (a change to the
// canonical example UUID is then one edit, not five).
const (
	testValidTranscriptUUID = "99d59925-36bc-424c-a789-8be54d9702ba"
	testOwnerID             = "user-1"
	testOwnerName           = "alice"
)

// --- NewTranscriptID validation ---
//
// Bare-string casts via the constructor under test are the intended subject here
// (constructor validation tests are the one place raw casts are expected — see
// CLAUDE.md type-safety rules).

func TestNewTranscriptID_ValidUUID(t *testing.T) {
	const raw = testValidTranscriptUUID
	tid, err := schema.NewTranscriptID(raw)
	if err != nil {
		t.Fatalf("expected valid UUID to succeed: %v", err)
	}
	if tid.String() != raw {
		t.Errorf("String() = %q, want %q", tid.String(), raw)
	}
}

func TestNewTranscriptID_EmptyRejected(t *testing.T) {
	if _, err := schema.NewTranscriptID(""); err == nil {
		t.Fatal("expected empty string to be rejected")
	}
}

func TestNewTranscriptID_NonUUIDRejected(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"arbitrary", "not-a-uuid"},
		{"agent-form", "agent-a3aee4f"},                               // SessionID accepts this; TranscriptID must not
		{"uppercase-hex", "99D59925-36BC-424C-A789-8BE54D9702BA"},     // canonical form is lowercase
		{"too-short", "99d59925-36bc-424c-a789"},                      // truncated
		{"path-traversal", "../99d59925-36bc-424c-a789-8be54d9702ba"}, // never reaches FS layout
		{"url-slug", "https://village/transcripts/99d59925"},
		{"trailing-junk", "99d59925-36bc-424c-a789-8be54d9702ba/x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := schema.NewTranscriptID(tc.raw); err == nil {
				t.Errorf("expected %q to be rejected", tc.raw)
			}
		})
	}
}

func TestNewTranscriptID_ErrorIsActionable(t *testing.T) {
	_, err := schema.NewTranscriptID("garbage")
	if err == nil {
		t.Fatal("expected error")
	}
	// The error must name the offending value and the expected form.
	msg := err.Error()
	if !strings.Contains(msg, "garbage") || !strings.Contains(msg, "UUID") {
		t.Errorf("error not actionable: %q", msg)
	}
}

// --- PullTranscriptInfo wire shape ---

func TestPullTranscriptInfo_JSONRoundTrip(t *testing.T) {
	info := schema.PullTranscriptInfo{
		TranscriptID:    schema.TranscriptID(testValidTranscriptUUID),
		LocalID:         "ses_local123",
		OwnerUserID:     testOwnerID,
		OwnerUsername:   testOwnerName,
		Title:           "Refactor pipeline",
		Harness:         schema.HarnessClaudeCode,
		ProjectName:     "peasant",
		Visibility:      schema.VisibilityGroup,
		ContentHash:     "deadbeef",
		ContractVersion: "0.1.0",
		PublishedAt:     1700000000000,
		UpdatedAt:       1700000001000,
		AnnotationCount: 3,
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got schema.PullTranscriptInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != info {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, info)
	}
	// Verify the wire keys are the camelCase JSON tags consumers depend on.
	for _, key := range []string{
		`"transcriptId"`, `"ownerUserId"`, `"ownerUsername"`,
		`"visibility"`, `"publishedAt"`, `"updatedAt"`, `"annotationCount"`,
	} {
		if !strings.Contains(string(data), key) {
			t.Errorf("missing wire key %s in %s", key, data)
		}
	}
}

func TestPullTranscriptInfo_OmitEmpty(t *testing.T) {
	// A minimal info: empty optional fields must be omitted; required fields stay.
	info := schema.PullTranscriptInfo{
		TranscriptID:  schema.TranscriptID(testValidTranscriptUUID),
		OwnerUserID:   testOwnerID,
		OwnerUsername: testOwnerName,
		Visibility:    schema.VisibilityPrivate,
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, omitted := range []string{`"localId"`, `"title"`, `"harness"`, `"projectName"`, `"contentHash"`, `"contractVersion"`} {
		if strings.Contains(string(data), omitted) {
			t.Errorf("expected %s to be omitted, got %s", omitted, data)
		}
	}
	// annotationCount is NOT omitempty (a count of 0 is meaningful).
	if !strings.Contains(string(data), `"annotationCount"`) {
		t.Errorf("annotationCount should always be present, got %s", data)
	}
}

// --- PullListResponse ---

func TestPullListResponse_JSONRoundTrip(t *testing.T) {
	resp := schema.PullListResponse{
		Transcripts: []schema.PullTranscriptInfo{
			{
				TranscriptID:  schema.TranscriptID(testValidTranscriptUUID),
				OwnerUserID:   testOwnerID,
				OwnerUsername: testOwnerName,
				Visibility:    schema.VisibilityGroup,
			},
		},
		Page:  1,
		Limit: 50,
		Total: 1,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got schema.PullListResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Transcripts) != 1 || got.Total != 1 || got.Page != 1 || got.Limit != 50 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

// --- PullAnnotation: AnnotationSummary embed + author identity ---

func TestPullAnnotation_EmbedsSummaryWithAuthorIdentity(t *testing.T) {
	pa := schema.PullAnnotation{
		AnnotationSummary: schema.AnnotationSummary{
			ID:            "annot-1",
			TargetKind:    schema.TargetSession,
			IsPrimary:     true,
			AnnotatorKind: schema.AnnotatorHuman,
			AnnotatorName: "alice",
			TypeID:        "friction.retry-loop",
			TypeName:      "Retry Loop",
			Value:         "observed",
			CreatedAt:     1700000000000,
		},
		AuthorUserID:   "user-2",
		AuthorUsername: "bob",
	}
	data, err := json.Marshal(pa)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got schema.PullAnnotation
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Embedded summary fields survive the round-trip (flattened into the parent object).
	if got.ID != "annot-1" || got.TypeID != "friction.retry-loop" {
		t.Errorf("embedded AnnotationSummary not preserved: %+v", got)
	}
	// Author identity is the value-add over AnnotationSummary (needed to
	// foreign-mark and to exclude own-authored rows on refresh).
	if got.AuthorUserID != "user-2" || got.AuthorUsername != "bob" {
		t.Errorf("author identity not preserved: %+v", got)
	}
	for _, key := range []string{`"authorUserId"`, `"authorUsername"`, `"id"`, `"typeId"`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("missing wire key %s in %s", key, data)
		}
	}
}

// --- SchemaVersionResponse pull window (omitempty back-compat) ---

func TestSchemaVersionResponse_PullWindowOmittedWhenEmpty(t *testing.T) {
	// An older village that predates the pull surface emits no pull window —
	// the absent fields must NOT appear (omitempty), so the CLI can treat
	// "absent" as "village too old for pull".
	resp := schema.SchemaVersionResponse{
		AnnotationSchemaVersion: "3",
		PushContractVersion:     schema.PushContractVersion("0.1.0"),
		MinPushContractVersion:  schema.PushContractVersion("0.1.0"),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, omitted := range []string{`"pullContractVersion"`, `"minPullContractVersion"`} {
		if strings.Contains(string(data), omitted) {
			t.Errorf("expected %s omitted when empty, got %s", omitted, data)
		}
	}
}

func TestSchemaVersionResponse_PullWindowPresentWhenSet(t *testing.T) {
	resp := schema.SchemaVersionResponse{
		AnnotationSchemaVersion: "3",
		PushContractVersion:     schema.PushContractVersion("0.1.0"),
		MinPushContractVersion:  schema.PushContractVersion("0.1.0"),
		PullContractVersion:     schema.PushContractVersion("0.1.0"),
		MinPullContractVersion:  schema.PushContractVersion("0.1.0"),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got schema.SchemaVersionResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.PullContractVersion != "0.1.0" || got.MinPullContractVersion != "0.1.0" {
		t.Errorf("pull window not preserved: %+v", got)
	}
	for _, key := range []string{`"pullContractVersion"`, `"minPullContractVersion"`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("missing wire key %s in %s", key, data)
		}
	}
}
