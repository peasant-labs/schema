package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"reflect"
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

// conversationTurnKindFixtures holds two corpora. Kinds decides every pair in
// AllRoles x AllEntryTypes at depth 0. Outside holds turns the pair table
// cannot express: a nested turn, a missing entry type, an unknown role or
// entry type.
type conversationTurnKindFixtures struct {
	RequiredNames        []string                                                                 `yaml:"required_names"`
	Kinds                testcase.Corpus[conversationTurnKindInput, conversationTurnKindExpected] `yaml:"kinds"`
	OutsideRequiredNames []string                                                                 `yaml:"outside_required_names"`
	Outside              testcase.Corpus[conversationTurnKindInput, conversationTurnKindExpected] `yaml:"outside"`
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
	assertcase.RequireMin(t, fixtures.Kinds, 1)
	assertcase.RequireValid(t, fixtures.Kinds)
	assertcase.RequireMin(t, fixtures.Outside, 1)
	assertcase.RequireValid(t, fixtures.Outside)
	return fixtures
}

// requireConversationKindNames fails unless the corpus names are exactly the
// manifest, each once.
func requireConversationKindNames(t *testing.T, arm string, required []string, corpus testcase.Corpus[conversationTurnKindInput, conversationTurnKindExpected]) {
	t.Helper()
	seen := map[string]bool{}
	for _, c := range corpus.Cases {
		if seen[c.Name] {
			t.Fatalf("conversation turn %s corpus repeats case %q", arm, c.Name)
		}
		seen[c.Name] = true
	}
	want := map[string]bool{}
	for _, name := range required {
		want[name] = true
		if !seen[name] {
			t.Errorf("required conversation turn %s case %q is missing", arm, name)
		}
	}
	for name := range seen {
		if !want[name] {
			t.Errorf("conversation turn %s case %q is not in the manifest", arm, name)
		}
	}
}

func conversationKindKey(role, entryType string, depth int) string {
	return fmt.Sprintf("%s|%s|%d", role, entryType, depth)
}

// TestConversationTurnKinds pins the contract-owned allowed list: one row per
// pair in AllRoles x AllEntryTypes at depth 0, failing on an unlisted or
// repeated pair, and the yes rows equal the allowed list.
func TestConversationTurnKinds(t *testing.T) {
	fixtures := loadConversationTurnKinds(t)
	requireConversationKindNames(t, "kind", fixtures.RequiredNames, fixtures.Kinds)
	seenPairs := map[string]bool{}
	yes := map[string]bool{}
	for _, c := range fixtures.Kinds.Cases {
		if c.Input.Depth != 0 {
			t.Fatalf("%s: the kind corpus decides top-level pairs only; move a depth-%d row to the outside corpus", c.Name, c.Input.Depth)
		}
		key := conversationKindKey(c.Input.Role, c.Input.EntryType, c.Input.Depth)
		if seenPairs[key] {
			t.Fatalf("conversation turn kinds repeats pair %q", key)
		}
		seenPairs[key] = true
		got := schema.IsConversationTurn(schema.Role(c.Input.Role), schema.EntryType(c.Input.EntryType), c.Input.Depth)
		if got != c.Expected.Counts {
			t.Errorf("%s: IsConversationTurn(%q, %q, %d) = %v, want %v", c.Name, c.Input.Role, c.Input.EntryType, c.Input.Depth, got, c.Expected.Counts)
		}
		if c.Expected.Counts {
			yes[c.Input.Role+"|"+c.Input.EntryType] = true
		}
	}
	listed := map[string]bool{}
	for _, role := range schema.AllRoles {
		for _, entryType := range schema.AllEntryTypes {
			key := conversationKindKey(string(role), string(entryType), 0)
			listed[key] = true
			if !seenPairs[key] {
				t.Errorf("conversation turn kinds misses pair (%q, %q)", string(role), string(entryType))
			}
		}
	}
	for key := range seenPairs {
		if !listed[key] {
			t.Errorf("conversation turn kinds row %q is outside AllRoles x AllEntryTypes; move it to the outside corpus", key)
		}
	}
	wantYes := map[string]bool{"user|text": true, "user|image": true, "assistant|text": true, "assistant|image": true}
	if !reflect.DeepEqual(yes, wantYes) {
		t.Errorf("conversation turn kinds yes rows = %v, want exactly %v", yes, wantYes)
	}
}

// TestConversationTurnOutsideKinds replays the turns the pair table cannot
// express. None of them counts.
func TestConversationTurnOutsideKinds(t *testing.T) {
	fixtures := loadConversationTurnKinds(t)
	requireConversationKindNames(t, "outside", fixtures.OutsideRequiredNames, fixtures.Outside)
	for _, c := range fixtures.Outside.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := schema.IsConversationTurn(schema.Role(c.Input.Role), schema.EntryType(c.Input.EntryType), c.Input.Depth)
			if got != c.Expected.Counts {
				t.Errorf("IsConversationTurn(%q, %q, %d) = %v, want %v", c.Input.Role, c.Input.EntryType, c.Input.Depth, got, c.Expected.Counts)
			}
		})
	}
}
