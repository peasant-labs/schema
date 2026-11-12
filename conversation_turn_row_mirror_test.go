package schema_test

import (
	"testing"
	"time"

	schema "github.com/peasant-labs/schema"
)

// TestLocalSessionRowConversationMirror pins that LocalSessionRow.Validate
// refuses a grouped row whose conversation counts disagree.
func TestLocalSessionRowConversationMirror(t *testing.T) {
	start, err := time.Parse(time.RFC3339, "2026-09-20T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	hash := schema.ProjectHash("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	one := int64(1)
	two := int64(2)
	session := schema.SessionSummary{ID: "s-conv", Harness: schema.HarnessOpenCode, StartTime: start, DurationMins: 1, TotalTokens: 10, TurnCount: 2, ToolCallCount: 0, Project: "worker", ProjectHash: hash, ConversationTurnCount: &one}
	sync := schema.LocalSyncSummary{ID: "s-conv", Harness: schema.HarnessOpenCode, ProjectName: "worker", ProjectHash: hash, HostSlug: "host", StartTime: "2026-09-20T10:00:00Z", DurationMs: 1000, TotalTokens: 10, TurnCount: 2, Model: "m", SyncStatus: schema.SyncStatusNew, PreviouslyPushed: false, ConversationTurnCount: &two}
	row := schema.LocalSessionRow{Session: session, Sync: &sync}
	if err := row.Validate(); err == nil {
		t.Fatalf("LocalSessionRow with disagreeing conversation counts is accepted")
	}
	// Matching counts remain valid.
	sync.ConversationTurnCount = &one
	if err := row.Validate(); err != nil {
		t.Fatalf("LocalSessionRow with agreeing conversation counts is rejected: %v", err)
	}
}

// TestVillageSessionRowConversationMirror pins that VillageSessionRow.Validate
// refuses copies whose conversation or displayed counts disagree.
func TestVillageSessionRowConversationMirror(t *testing.T) {
	one := int64(1)
	two := int64(2)
	for _, variant := range []string{"collective", "myShare"} {
		row := villageRow(villageRowInput{Variant: variant})
		row.Session.ConversationTurnCount = &one
		row.Session.DisplayedTurnCount = &one
		switch variant {
		case "collective":
			row.Collective.ConversationTurnCount = &two
			row.Collective.DisplayedTurnCount = &one
		case "myShare":
			row.MyShare.ConversationTurnCount = &two
			row.MyShare.DisplayedTurnCount = &one
		}
		if err := row.Validate(); err == nil {
			t.Fatalf("VillageSessionRow %s with disagreeing conversation counts is accepted", variant)
		}
		switch variant {
		case "collective":
			row.Collective.ConversationTurnCount = &one
		case "myShare":
			row.MyShare.ConversationTurnCount = &one
		}
		if err := row.Validate(); err != nil {
			t.Fatalf("VillageSessionRow %s with agreeing counts is rejected: %v", variant, err)
		}
	}
}
