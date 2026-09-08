package schema_test

import (
	"encoding/json"
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
	assert.RequireMin(t, fx.InvalidCounts, 5)
	assert.RequireValid(t, fx.InvalidCounts)
	assert.RequireMin(t, fx.Mirrors, 3)
	assert.RequireValid(t, fx.Mirrors)
	assert.RequireMin(t, fx.Digest, 2)
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
			identity, stats, children, err := schema.BuildAuthoritativePublicationProjections(detail, meta, 11)
			if (err == nil) != row.Expected.Accept || err != nil && !strings.Contains(err.Error(), row.Expected.ErrorContains) {
				t.Fatalf("error=%v accept=%v", err, row.Expected.Accept)
			}
			if err == nil && (identity.ParentSessionID == nil || *identity.ParentSessionID != *detail.Relationships[0].TargetLocalID || stats.InputSubmissionCount == nil || *stats.InputSubmissionCount != row.Input.DetailCount || len(children) != 0) {
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
	detail := schema.SessionDetailPayload{ID: "ses_child", Harness: schema.HarnessClaudeCode, TurnCount: 5, InputSubmissionCount: &input.DetailCount, ParentSessionID: &parent, Relationships: []schema.SessionRelationship{relation}, Turns: []schema.TurnDetail{}}
	meta := schema.NewUnifiedMetadata()
	meta.SessionID = "ses_child"
	meta.ModelHarness = schema.HarnessClaudeCode
	meta.ParentUUID = &graphParent
	meta.Relationships = []schema.SessionRelationship{relation}
	meta.Stats.TurnCount = 5
	meta.Stats.InputSubmissionCount = &input.MetadataCount
	return detail, meta
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

func validPublishRequest() schema.AuthoritativePublishRequest {
	return schema.AuthoritativePublishRequest{Identity: schema.AuthoritativeSessionIdentity{SessionID: "ses_digest", SchemaVersion: 11}, Model: schema.AuthoritativeModelInfo{Harness: schema.HarnessClaudeCode, Model: "model"}, Timestamp: schema.AuthoritativeTimestampInfo{Start: 1, End: 2}, Source: schema.AuthoritativeSourceInfo{Format: schema.SourceFormatJSON}, Project: schema.AuthoritativeProjectContext{Hash: schema.ProjectHash(strings.Repeat("a", 64)), Name: "project"}, ContentHash: schema.TranscriptContentHash(strings.Repeat("b", 64))}
}
