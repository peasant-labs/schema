package schema_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase/assert"
)

func requireDurableFixtures(t *testing.T) schema.DurableGraphFixtures {
	t.Helper()
	corpus, err := schema.LoadSessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	fx := corpus.Durable
	assert.RequireMin(t, fx.RoundTrip, 2)
	assert.RequireValid(t, fx.RoundTrip)
	assert.RequireMin(t, fx.Counts, 4)
	assert.RequireValid(t, fx.Counts)
	assert.RequireMin(t, fx.InvalidCounts, 8)
	assert.RequireValid(t, fx.InvalidCounts)
	assert.RequireMin(t, fx.Mirrors, 11)
	assert.RequireValid(t, fx.Mirrors)
	assert.RequireMin(t, fx.Digest, 6)
	assert.RequireValid(t, fx.Digest)
	want := make(map[string]bool, len(fx.RequiredNames))
	for _, name := range fx.RequiredNames {
		want[name] = true
	}
	for _, row := range fx.RoundTrip.Cases {
		delete(want, row.Name)
	}
	for _, row := range fx.Counts.Cases {
		delete(want, row.Name)
	}
	for _, row := range fx.InvalidCounts.Cases {
		delete(want, row.Name)
	}
	for _, row := range fx.Mirrors.Cases {
		delete(want, row.Name)
	}
	for _, row := range fx.Digest.Cases {
		delete(want, row.Name)
	}
	if len(want) != 0 {
		t.Fatalf("durable graph fixtures missing required names: %v", want)
	}
	return fx
}

func TestAuthoritativeGraphFieldsRejectReadStateBeforeDecode(t *testing.T) {
	fx, err := schema.LoadSessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	assert.RequireMin(t, fx.AuthoritativeRaw, 3)
	assert.RequireValid(t, fx.AuthoritativeRaw)
	for _, c := range fx.AuthoritativeRaw.Cases {
		t.Run(c.Name, func(t *testing.T) {
			request := validPublishRequest()
			request.Entries = []schema.AuthoritativeSessionEntry{{SessionID: "ses_digest", EntryIndex: 0, Harness: schema.HarnessClaudeCode, EntryType: schema.EntryTypeText, Role: schema.RoleAssistant}}
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var root map[string]json.RawMessage
			if json.Unmarshal(raw, &root) != nil {
				t.Fatal("decode request")
			}
			var entries []map[string]json.RawMessage
			if json.Unmarshal(root["entries"], &entries) != nil || len(entries) == 0 {
				t.Fatal("valid request has no entry")
			}
			var mutation map[string]json.RawMessage
			if json.Unmarshal([]byte(c.Input.RawJSON), &mutation) != nil {
				t.Fatal("decode fixture mutation")
			}
			for key, value := range mutation {
				entries[0][key] = value
			}
			root["entries"], _ = json.Marshal(entries)
			raw, _ = json.Marshal(root)
			_, err = schema.DecodeAuthoritativePublishMetadataRaw(raw)
			if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
				t.Fatalf("error=%v, want containing %q", err, c.Expected.ErrorContains)
			}
		})
	}
}

func TestDurableGraphRoundTrip(t *testing.T) {
	fx := requireDurableFixtures(t)
	for _, row := range fx.RoundTrip.Cases {
		t.Run(row.Name, func(t *testing.T) {
			value, err := schema.DecodeSessionDetailPayloadRaw(durableDetailJSON(t, row.Input.JSON))
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			back, err := schema.DecodeSessionDetailPayloadRaw(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if back.TurnCount != row.Expected.TurnCount || (back.InputSubmissionCount != nil) != row.Expected.InputPresent {
				t.Fatalf("counts after round trip = turn %d input %v", back.TurnCount, back.InputSubmissionCount)
			}
			if back.InputSubmissionCount != nil && *back.InputSubmissionCount != row.Expected.InputCount {
				t.Fatalf("input count=%d", *back.InputSubmissionCount)
			}
			if len(back.ChildSessions) != row.Expected.HelperCount {
				t.Fatalf("saved helper count=%d want %d", len(back.ChildSessions), row.Expected.HelperCount)
			}
			var expectedJSON, actualJSON any
			if err := json.Unmarshal(durableDetailJSON(t, row.Input.JSON), &expectedJSON); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(encoded, &actualJSON); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(expectedJSON, actualJSON) {
				t.Fatalf("semantic round trip changed durable payload\nactual: %s\nexpected: %s", encoded, durableDetailJSON(t, row.Input.JSON))
			}
			if len(back.Turns) > 0 {
				entry := schema.SessionEntry{SessionID: "ses_graph", EntryIndex: back.Turns[0].Index, Harness: back.Harness, EntryType: back.Turns[0].EntryType, Role: back.Turns[0].Role, SourceEntryRef: back.Turns[0].SourceEntryRef, Provenance: back.Turns[0].Provenance}
				rawEntry, err := json.Marshal(entry)
				if err != nil {
					t.Fatal(err)
				}
				var decodedEntry schema.SessionEntry
				if err := json.Unmarshal(rawEntry, &decodedEntry); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(entry, decodedEntry) {
					t.Fatalf("session entry round trip changed evidence: %+v", decodedEntry)
				}
			}
			if row.Expected.HelperCount > 0 {
				metadata := schema.NewUnifiedMetadata()
				metadata.SessionID = schema.SessionID(back.ID)
				metadata.ModelHarness = back.Harness
				metadata.ParentUUID = back.ParentSessionID
				metadata.Purpose = back.Purpose
				metadata.Relationships = append([]schema.SessionRelationship(nil), back.Relationships...)
				metadata.Stats.TurnCount = back.TurnCount
				count := *back.InputSubmissionCount
				metadata.Stats.InputSubmissionCount = &count
				root := *back.RootSessionID
				metadata.RootSessionID = &root
				identity, stats, helpers, err := schema.BuildAuthoritativePublicationProjections(back, metadata, 11, true)
				if err != nil {
					t.Fatal(err)
				}
				if identity.RootSessionID == nil || *identity.RootSessionID != root || identity.Purpose != back.Purpose || stats.TurnCount != 5 || stats.InputSubmissionCount == nil || *stats.InputSubmissionCount != 1 || len(helpers) != 2 {
					t.Fatalf("authoritative 1/5/2 projection differs: identity=%+v stats=%+v helpers=%+v", identity, stats, helpers)
				}
			}
			gotMain := durableRefs(back.Turns)
			gotEarlier := []string{}
			for _, section := range back.EarlierHistory {
				gotEarlier = append(gotEarlier, durableRefs(section.Turns)...)
			}
			if strings.Join(gotMain, ",") != strings.Join(row.Expected.MainRefs, ",") || strings.Join(gotEarlier, ",") != strings.Join(row.Expected.EarlierRefs, ",") {
				t.Fatalf("refs main=%v earlier=%v", gotMain, gotEarlier)
			}
		})
	}
}

func durableRefs(turns []schema.TurnDetail) []string {
	out := []string{}
	for _, turn := range turns {
		if turn.SourceEntryRef != "" {
			out = append(out, string(turn.SourceEntryRef))
		}
		for _, tool := range turn.ToolCalls {
			if tool.CallEntryRef != "" {
				out = append(out, string(tool.CallEntryRef))
			}
			if tool.ResultEntryRef != "" {
				out = append(out, string(tool.ResultEntryRef))
			}
		}
	}
	return out
}

func TestDurableInputCounts(t *testing.T) {
	fx := requireDurableFixtures(t)
	for _, row := range fx.Counts.Cases {
		t.Run(row.Name, func(t *testing.T) {
			value, err := schema.DecodeSessionDetailPayloadRaw(durableDetailJSON(t, row.Input.JSON))
			if err != nil {
				t.Fatal(err)
			}
			if (value.InputSubmissionCount != nil) != row.Expected.Present {
				t.Fatalf("presence=%v", value.InputSubmissionCount != nil)
			}
			if value.InputSubmissionCount != nil && *value.InputSubmissionCount != row.Expected.Value {
				t.Fatalf("value=%d", *value.InputSubmissionCount)
			}
		})
	}
	for _, row := range fx.InvalidCounts.Cases {
		t.Run(row.Name, func(t *testing.T) {
			_, err := schema.DecodeSessionDetailPayloadRaw(durableDetailJSON(t, row.Input.JSON))
			if err == nil || !strings.Contains(err.Error(), row.Expected.ErrorContains) {
				t.Fatalf("error=%v want %q", err, row.Expected.ErrorContains)
			}
			_, publishErr := schema.DecodeAuthoritativePublishMetadataRaw(authoritativeCountJSON(t, row.Input.JSON))
			if publishErr == nil || !strings.Contains(publishErr.Error(), "inputSubmissionCount") {
				t.Fatalf("authoritative count error=%v", publishErr)
			}
			var metadata schema.UnifiedMetadata
			if metadataErr := json.Unmarshal(metadataCountJSON(t, row.Input.JSON), &metadata); metadataErr == nil || !strings.Contains(metadataErr.Error(), "inputSubmissionCount") {
				t.Fatalf("metadata count error=%v", metadataErr)
			}
		})
	}
}

func metadataCountJSON(t *testing.T, detailRaw string) []byte {
	t.Helper()
	var detail map[string]json.RawMessage
	if err := json.Unmarshal([]byte(detailRaw), &detail); err != nil {
		t.Fatal(err)
	}
	value := map[string]any{"schemaVersion": 10, "sessionId": "ses_bad", "parentUuid": nil, "harness": "claude-code", "model": "model", "version": "1", "timestamp": map[string]any{"start": 1, "end": 2}, "source": map[string]any{"format": "json"}, "git": map[string]any{}, "project": map[string]any{"hash": strings.Repeat("a", 64), "name": "project"}, "hostSlug": "host", "stats": map[string]json.RawMessage{"inputSubmissionCount": detail["inputSubmissionCount"]}, "subagents": []any{}, "diagnostics": map[string]any{"warnings": []any{}}, "contentHash": "", "metadataHash": "", "redaction": map[string]any{}}
	out, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func authoritativeCountJSON(t *testing.T, detailRaw string) []byte {
	t.Helper()
	var detail map[string]json.RawMessage
	if err := json.Unmarshal([]byte(detailRaw), &detail); err != nil {
		t.Fatal(err)
	}
	base, err := json.Marshal(validPublishRequest())
	if err != nil {
		t.Fatal(err)
	}
	var request map[string]json.RawMessage
	if err := json.Unmarshal(base, &request); err != nil {
		t.Fatal(err)
	}
	var stats map[string]json.RawMessage
	if err := json.Unmarshal(request["stats"], &stats); err != nil {
		t.Fatal(err)
	}
	stats["inputSubmissionCount"] = detail["inputSubmissionCount"]
	request["stats"], err = json.Marshal(stats)
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func durableDetailJSON(t *testing.T, raw string) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	defaults := map[string]any{"startTime": "2026-09-08T00:00:00Z", "endTime": "2026-09-08T00:00:01Z", "durationMins": 0.0, "totalTokens": 0, "tokensIn": 0, "tokensOut": 0, "toolCallCount": 0}
	for key, item := range defaults {
		if _, exists := value[key]; !exists {
			value[key] = item
		}
	}
	out, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestAuthoritativeGraphMirrors(t *testing.T) {
	fx := requireDurableFixtures(t)
	for _, row := range fx.Mirrors.Cases {
		t.Run(row.Name, func(t *testing.T) {
			detail, meta := projectionPair(t, row.Input)
			identity, stats, children, err := schema.BuildAuthoritativePublicationProjections(detail, meta, 11, row.Input.ProjectedMainCount)
			if (err == nil) != row.Expected.Accept || err != nil && !strings.Contains(err.Error(), row.Expected.ErrorContains) {
				t.Fatalf("error=%v accept=%v", err, row.Expected.Accept)
			}
			if err == nil && (identity.ParentSessionID == nil || *identity.ParentSessionID != *detail.Relationships[0].TargetLocalID || !optionalInt64Equal(stats.InputSubmissionCount, row.Input.DetailCount) || len(children) != 0) {
				t.Fatalf("derived identity/stats/children=%+v %+v %+v", identity, stats, children)
			}
		})
	}
}

func projectionPair(t *testing.T, input schema.DurableMirrorInput) (schema.SessionDetailPayload, schema.UnifiedMetadata) {
	t.Helper()
	parent, err := schema.NewSessionID(input.Parent)
	if err != nil {
		t.Fatal(err)
	}
	graphParent, err := schema.NewSessionID(input.GraphParent)
	if err != nil {
		t.Fatal(err)
	}
	relation := schema.SessionRelationship{Kind: schema.SessionRelationshipStartedBy, TargetState: schema.RelationshipTargetKnown, TargetLocalID: &graphParent, Evidence: schema.EvidenceNativeTyped}
	detail := schema.SessionDetailPayload{ID: "ses_child", Harness: schema.HarnessClaudeCode, TurnCount: input.DetailTurnCount, InputSubmissionCount: input.DetailCount, ParentSessionID: &parent, Relationships: []schema.SessionRelationship{relation}, Turns: []schema.TurnDetail{}}
	meta := schema.NewUnifiedMetadata()
	meta.SessionID = "ses_child"
	meta.ModelHarness = schema.HarnessClaudeCode
	meta.ParentUUID = &graphParent
	meta.Relationships = []schema.SessionRelationship{relation}
	meta.Stats.TurnCount = input.MetadataTurnCount
	meta.Stats.InputSubmissionCount = input.MetadataCount
	if input.DetailRoot != "" {
		root, err := schema.NewSessionID(input.DetailRoot)
		if err != nil {
			t.Fatal(err)
		}
		detail.RootSessionID = &root
	}
	if input.MetadataRoot != "" {
		root, err := schema.NewSessionID(input.MetadataRoot)
		if err != nil {
			t.Fatal(err)
		}
		meta.RootSessionID = &root
	}
	return detail, meta
}

func optionalInt64Equal(a, b *int64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func TestAuthoritativeDigestCountPresence(t *testing.T) {
	fx := requireDurableFixtures(t)
	for _, row := range fx.Digest.Cases {
		t.Run(row.Name, func(t *testing.T) {
			left := validPublishRequest()
			right := validPublishRequest()
			if row.Input.LeftPresent {
				left.Stats.InputSubmissionCount = &row.Input.LeftValue
			}
			if row.Input.RightPresent {
				right.Stats.InputSubmissionCount = &row.Input.RightValue
			}
			applyDigestVariant(t, &left, row.Input.LeftVariant)
			applyDigestVariant(t, &right, row.Input.RightVariant)
			lop, err := schema.CanonicalizePublishRequest(left)
			if err != nil {
				t.Fatal(err)
			}
			rop, err := schema.CanonicalizePublishRequest(right)
			if err != nil {
				t.Fatal(err)
			}
			ld, err := schema.FingerprintPublishOperation(lop)
			if err != nil {
				t.Fatal(err)
			}
			rd, err := schema.FingerprintPublishOperation(rop)
			if err != nil {
				t.Fatal(err)
			}
			if (ld != rd) != row.Expected.Different {
				t.Fatalf("digests %s %s", ld, rd)
			}
		})
	}
}

func applyDigestVariant(t *testing.T, request *schema.AuthoritativePublishRequest, variant string) {
	t.Helper()
	switch variant {
	case "":
	case "root":
		root, err := schema.NewSessionID("ses_root")
		if err != nil {
			t.Fatal(err)
		}
		request.Identity.RootSessionID = &root
	case "purpose":
		request.Identity.Purpose = schema.SessionPurposeDelegatedWork
	case "relationship":
		parent, err := schema.NewSessionID("ses_parent")
		if err != nil {
			t.Fatal(err)
		}
		request.Identity.ParentSessionID = &parent
		request.Identity.Relationships = []schema.SessionRelationship{{Kind: schema.SessionRelationshipStartedBy, TargetState: schema.RelationshipTargetKnown, TargetLocalID: &parent, Evidence: schema.EvidenceNativeTyped}}
	case "entry_provenance":
		request.Entries = []schema.AuthoritativeSessionEntry{{SessionID: "ses_digest", EntryIndex: 0, Harness: schema.HarnessClaudeCode, EntryType: schema.EntryTypeText, Role: schema.RoleUser, SourceEntryRef: "e_u1", Provenance: &schema.ContentProvenance{Origin: schema.ContentOriginSubmittedInput, Actor: schema.ActorOriginUnknown, Delivery: schema.DeliveryOriginSessionAdmission, Ownership: schema.ContentOwnershipLocal, Evidence: schema.EvidenceNativeTyped, InputModality: schema.InputModalityText, SubmissionRef: "s_u1"}}}
	default:
		t.Fatalf("unknown digest variant %q", variant)
	}
}

func validPublishRequest() schema.AuthoritativePublishRequest {
	return schema.AuthoritativePublishRequest{Identity: schema.AuthoritativeSessionIdentity{SessionID: "ses_digest", SchemaVersion: 11}, Model: schema.AuthoritativeModelInfo{Harness: schema.HarnessClaudeCode, Model: "model"}, Timestamp: schema.AuthoritativeTimestampInfo{Start: 1, End: 2}, Source: schema.AuthoritativeSourceInfo{Format: schema.SourceFormatJSON}, Project: schema.AuthoritativeProjectContext{Hash: schema.ProjectHash(strings.Repeat("a", 64)), Name: "project"}, ContentHash: schema.TranscriptContentHash(strings.Repeat("b", 64))}
}
