package schema_test

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/conversation_turn_kinds.yaml
var conversationTurnKindsYAML []byte

type conversationTurnKindInput struct {
	Role      string `yaml:"role"`
	EntryType string `yaml:"entryType"`
	Depth     int    `yaml:"depth"`
}

type conversationTurnKindExpected struct {
	Counts bool `yaml:"counts"`
}

type conversationTurnKindFixtures struct {
	RequiredNames []string                                              `yaml:"required_names"`
	Kinds         testcase.Corpus[conversationTurnKindInput, conversationTurnKindExpected] `yaml:"kinds"`
}

func loadConversationTurnKinds(t *testing.T) conversationTurnKindFixtures {
	t.Helper()
	var fixtures conversationTurnKindFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(conversationTurnKindsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("load conversation turn kinds: %v", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("conversation turn kinds fixture must contain one document: %v", err)
	}
	assertcase.RequireMin(t, fixtures.Kinds, 32)
	assertcase.RequireValid(t, fixtures.Kinds)
	return fixtures
}

// TestConversationTurnKinds pins the contract-owned allowed list: one row per
// pair in AllRoles x AllEntryTypes, failing on an unlisted or repeated pair,
// and the yes rows equal the internal list behind IsConversationTurn.
func TestConversationTurnKinds(t *testing.T) {
	fixtures := loadConversationTurnKinds(t)
	if len(fixtures.RequiredNames) != 32 {
		t.Fatalf("conversation turn kinds manifest has %d names, want 32", len(fixtures.RequiredNames))
	}
	seenPairs := map[string]bool{}
	seenNames := map[string]bool{}
	for _, c := range fixtures.Kinds.Cases {
		if seenNames[c.Name] {
			t.Fatalf("conversation turn kinds repeats case %q", c.Name)
		}
		seenNames[c.Name] = true
		key := c.Input.Role + "|" + c.Input.EntryType + "|depth"
		if seenPairs[key] {
			t.Fatalf("conversation turn kinds repeats pair %q", key)
		}
		seenPairs[key] = true
		got := schema.IsConversationTurn(schema.Role(c.Input.Role), schema.EntryType(c.Input.EntryType), c.Input.Depth)
		if got != c.Expected.Counts {
			t.Errorf("%s: IsConversationTurn(%q, %q, %d) = %v, want %v", c.Name, c.Input.Role, c.Input.EntryType, c.Input.Depth, got, c.Expected.Counts)
		}
	}
	for _, name := range fixtures.RequiredNames {
		if !seenNames[name] {
			t.Errorf("required conversation turn kind %q is missing", name)
		}
	}
	if len(seenNames) != len(fixtures.RequiredNames) {
		t.Errorf("conversation turn kind manifest has %d names for %d cases", len(fixtures.RequiredNames), len(seenNames))
	}
	pairCount := 0
	for _, role := range schema.AllRoles {
		for _, entryType := range schema.AllEntryTypes {
			pairCount++
			key := string(role) + "|" + string(entryType) + "|depth"
			if !seenPairs[key] {
				t.Errorf("conversation turn kinds misses pair (%q, %q)", string(role), string(entryType))
			}
		}
	}
	if pairCount != 32 {
		t.Errorf("AllRoles x AllEntryTypes has %d pairs, want 32", pairCount)
	}
	yes := map[string]bool{}
	for _, c := range fixtures.Kinds.Cases {
		if c.Expected.Counts {
			yes[c.Input.Role+"|"+c.Input.EntryType] = true
		}
	}
	wantYes := map[string]bool{"user|text": true, "user|image": true, "assistant|text": true, "assistant|image": true}
	if len(yes) != len(wantYes) {
		t.Fatalf("conversation turn kinds has %d yes rows, want %d", len(yes), len(wantYes))
	}
	for key := range wantYes {
		if !yes[key] {
			t.Errorf("conversation turn kinds yes rows miss allowed pair %q", key)
		}
	}
	for key := range yes {
		if !wantYes[key] {
			t.Errorf("conversation turn kinds yes row %q is outside the allowed list", key)
		}
	}
}
