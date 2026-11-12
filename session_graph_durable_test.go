package schema_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase/assert"
	"github.com/santhosh-tekuri/jsonschema/v5"
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

// durableCountField names the count a counts row carries: the corpus holds
// rows for both optional session counts, told apart by the row name.
type durableCountField struct {
	wire   string
	detail func(schema.SessionDetailPayload) *int64
	stats  func(schema.AuthoritativeSessionStats) *int64
}

func countFieldOf(rowName string) durableCountField {
	if strings.HasPrefix(rowName, "conversation-count-") {
		return durableCountField{wire: "conversationTurnCount", detail: func(v schema.SessionDetailPayload) *int64 { return v.ConversationTurnCount }, stats: func(v schema.AuthoritativeSessionStats) *int64 { return v.ConversationTurnCount }}
	}
	return durableCountField{wire: "inputSubmissionCount", detail: func(v schema.SessionDetailPayload) *int64 { return v.InputSubmissionCount }, stats: func(v schema.AuthoritativeSessionStats) *int64 { return v.InputSubmissionCount }}
}

func requireCount(t *testing.T, where string, got *int64, want struct {
	Present bool
	Value   int64
}) {
	t.Helper()
	if (got != nil) != want.Present {
		t.Fatalf("%s presence=%v, want %v", where, got != nil, want.Present)
	}
	if got != nil && *got != want.Value {
		t.Fatalf("%s value=%d, want %d", where, *got, want.Value)
	}
}

// publishStatsContract compiles the stats object of the served publish-request
// JSON Schema, the bytes village enforces.
func publishStatsContract(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := schema.NewJSONSchemaCompiler()
	if err := compiler.AddResource("publish-request.json", bytes.NewReader(schema.PublishRequestSchemaJSON())); err != nil {
		t.Fatalf("add publish-request schema resource: %v", err)
	}
	contract, err := compiler.Compile("publish-request.json#/properties/stats")
	if err != nil {
		t.Fatalf("compile publish-request stats schema: %v", err)
	}
	return contract
}

// publishStatsSchemaError validates the stats object of a publish request
// against the served schema.
func publishStatsSchemaError(t *testing.T, contract *jsonschema.Schema, request []byte) error {
	t.Helper()
	var value struct {
		Stats any `json:"stats"`
	}
	if err := json.Unmarshal(request, &value); err != nil {
		t.Fatal(err)
	}
	if value.Stats == nil {
		t.Fatal("publish request fixture carries no stats object")
	}
	return contract.Validate(value.Stats)
}

func TestDurableInputCounts(t *testing.T) {
	fx := requireDurableFixtures(t)
	contract := publishStatsContract(t)
	for _, row := range fx.Counts.Cases {
		t.Run(row.Name, func(t *testing.T) {
			field := countFieldOf(row.Name)
			want := struct {
				Present bool
				Value   int64
			}{row.Expected.Present, row.Expected.Value}
			value, err := schema.DecodeSessionDetailPayloadRaw(durableDetailJSON(t, row.Input.JSON))
			if err != nil {
				t.Fatal(err)
			}
			requireCount(t, "session detail "+field.wire, field.detail(value), want)
			publishRaw := authoritativeCountJSON(t, row.Input.JSON)
			request, err := schema.DecodeAuthoritativePublishMetadataRaw(publishRaw)
			if err != nil {
				t.Fatalf("publication decoder refuses a must-pass count: %v", err)
			}
			requireCount(t, "publication stats "+field.wire, field.stats(request.Stats), want)
			if err := publishStatsSchemaError(t, contract, publishRaw); err != nil {
				t.Fatalf("publish-request schema refuses a must-pass count: %v", err)
			}
		})
	}
	for _, row := range fx.InvalidCounts.Cases {
		t.Run(row.Name, func(t *testing.T) {
			_, err := schema.DecodeSessionDetailPayloadRaw(durableDetailJSON(t, row.Input.JSON))
			if err == nil || !strings.Contains(err.Error(), row.Expected.ErrorContains) {
				t.Fatalf("error=%v want %q", err, row.Expected.ErrorContains)
			}
			wantField := countFieldOf(row.Name).wire
			publishRaw := authoritativeCountJSON(t, row.Input.JSON)
			_, publishErr := schema.DecodeAuthoritativePublishMetadataRaw(publishRaw)
			if publishErr == nil || !strings.Contains(publishErr.Error(), wantField) {
				t.Fatalf("authoritative count error=%v", publishErr)
			}
			if schemaErr := publishStatsSchemaError(t, contract, publishRaw); schemaErr == nil || !strings.Contains(schemaErr.Error(), wantField) {
				t.Fatalf("publish-request schema error=%v, want one naming %s", schemaErr, wantField)
			}
			var metadata schema.UnifiedMetadata
			if metadataErr := json.Unmarshal(metadataCountJSON(t, row.Input.JSON), &metadata); metadataErr == nil || !strings.Contains(metadataErr.Error(), wantField) {
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
	stats := map[string]json.RawMessage{}
	if count, present := detail["inputSubmissionCount"]; present {
		stats["inputSubmissionCount"] = count
	}
	if count, present := detail["conversationTurnCount"]; present {
		stats["conversationTurnCount"] = count
	}
	value := map[string]any{"schemaVersion": 10, "sessionId": "ses_bad", "parentUuid": nil, "harness": "claude-code", "model": "model", "version": "1", "timestamp": map[string]any{"start": 1, "end": 2}, "source": map[string]any{"format": "json"}, "git": map[string]any{}, "project": map[string]any{"hash": strings.Repeat("a", 64), "name": "project"}, "hostSlug": "host", "stats": stats, "subagents": []any{}, "diagnostics": map[string]any{"warnings": []any{}}, "contentHash": "", "metadataHash": "", "redaction": map[string]any{}}
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
	if count, present := detail["inputSubmissionCount"]; present {
		stats["inputSubmissionCount"] = count
	} else {
		delete(stats, "inputSubmissionCount")
	}
	if count, present := detail["conversationTurnCount"]; present {
		stats["conversationTurnCount"] = count
	} else {
		delete(stats, "conversationTurnCount")
	}
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
			var err error
			switch row.Input.Surface {
			case "":
				detail, meta := projectionPair(t, row.Input)
				identity, stats, children, buildErr := schema.BuildAuthoritativePublicationProjections(detail, meta, 11, row.Input.ProjectedMainCount)
				err = buildErr
				if err == nil && (identity.ParentSessionID == nil || *identity.ParentSessionID != *detail.Relationships[0].TargetLocalID || !optionalInt64Equal(stats.InputSubmissionCount, row.Input.DetailCount) || !optionalInt64Equal(stats.ConversationTurnCount, row.Input.DetailConversationCount) || len(children) != 0) {
					t.Fatalf("derived identity/stats/children=%+v %+v %+v", identity, stats, children)
				}
			case "local_row":
				err = localMirrorRow(t, row.Input).Validate()
			case "village_collective", "village_my_share":
				err = villageMirrorRow(t, row.Input).Validate()
			default:
				t.Fatalf("unknown mirror surface %q", row.Input.Surface)
			}
			if (err == nil) != row.Expected.Accept || err != nil && !strings.Contains(err.Error(), row.Expected.ErrorContains) {
				t.Fatalf("error=%v accept=%v want containing %q", err, row.Expected.Accept, row.Expected.ErrorContains)
			}
			if !row.Expected.Accept && row.Expected.ErrorContains == "" {
				t.Fatalf("must-fail mirror row carries no error_contains; any refusal would pass")
			}
		})
	}
}

// localMirrorRow builds a grouped local row whose session and sync copies
// carry the row's counts.
func localMirrorRow(t *testing.T, input schema.DurableMirrorInput) schema.LocalSessionRow {
	t.Helper()
	start, err := time.Parse(time.RFC3339, "2026-09-20T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	hash := schema.ProjectHash(strings.Repeat("c", 64))
	session := schema.SessionSummary{ID: "s-conv", Harness: schema.HarnessOpenCode, StartTime: start, DurationMins: 1, TotalTokens: 10, TurnCount: 2, Project: "worker", ProjectHash: hash, InputSubmissionCount: input.DetailCount, ConversationTurnCount: input.DetailConversationCount}
	sync := schema.LocalSyncSummary{ID: "s-conv", Harness: schema.HarnessOpenCode, ProjectName: "worker", ProjectHash: hash, HostSlug: "host", StartTime: "2026-09-20T10:00:00Z", DurationMs: 1000, TotalTokens: 10, TurnCount: 2, Model: "m", SyncStatus: schema.SyncStatusNew, InputSubmissionCount: input.MetadataCount, ConversationTurnCount: input.MetadataConversationCount}
	return schema.LocalSessionRow{Session: session, Sync: &sync}
}

// villageMirrorRow builds a grouped village row whose session copy and route
// copy carry the row's counts.
func villageMirrorRow(t *testing.T, input schema.DurableMirrorInput) schema.VillageSessionRow {
	t.Helper()
	variant := map[string]string{"village_collective": "collective", "village_my_share": "myShare"}[input.Surface]
	row := villageRow(villageRowInput{Variant: variant, SessionCount: input.DetailCount, VariantCount: input.MetadataCount})
	row.Session.ConversationTurnCount = input.DetailConversationCount
	row.Session.DisplayedTurnCount = input.DetailDisplayedCount
	switch variant {
	case "collective":
		row.Collective.ConversationTurnCount = input.MetadataConversationCount
		row.Collective.DisplayedTurnCount = input.MetadataDisplayedCount
	case "myShare":
		row.MyShare.ConversationTurnCount = input.MetadataConversationCount
		row.MyShare.DisplayedTurnCount = input.MetadataDisplayedCount
	}
	return row
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
	detail := schema.SessionDetailPayload{ID: "ses_child", Harness: schema.HarnessClaudeCode, TurnCount: input.DetailTurnCount, InputSubmissionCount: input.DetailCount, ConversationTurnCount: input.DetailConversationCount, ParentSessionID: &parent, Relationships: []schema.SessionRelationship{relation}, Turns: []schema.TurnDetail{}}
	meta := schema.NewUnifiedMetadata()
	meta.SessionID = "ses_child"
	meta.ModelHarness = schema.HarnessClaudeCode
	meta.ParentUUID = &graphParent
	meta.Relationships = []schema.SessionRelationship{relation}
	meta.Stats.TurnCount = input.MetadataTurnCount
	meta.Stats.InputSubmissionCount = input.MetadataCount
	meta.Stats.ConversationTurnCount = input.MetadataConversationCount
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
			left.Stats.ConversationTurnCount = row.Input.LeftConversation
			right.Stats.ConversationTurnCount = row.Input.RightConversation
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
			if row.Expected.RightDigest != "" && string(rd) != row.Expected.RightDigest {
				t.Fatalf("right digest=%s, want the pinned %s; the fingerprint encoding is frozen by release, so a changed digest breaks replay detection for stored publications", rd, row.Expected.RightDigest)
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
