package schema_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
)

//go:embed testdata/session-detail/transcripts/turn_commands.yaml
var turnCommandFixtureYAML []byte

//go:embed testdata/session-detail/transcripts/turn_commands_manifest.yaml
var turnCommandManifestYAML []byte

type turnCommandFixtureInput struct {
	Name string `yaml:"name"`
	Args string `yaml:"args,omitempty"`
}

type turnCommandFixtureExpected struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"errorContains,omitempty"`
	WireHasArgs   bool   `yaml:"wireHasArgs,omitempty"`
}

var turnCommandTimestamp = time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

func TestCommandInvocationFixture(t *testing.T) {
	corpus := loadTurnCommandFixtures(t)
	for _, fixtureCase := range corpus.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			invocation, err := schema.NewCommandInvocation(fixtureCase.Input.Name, fixtureCase.Input.Args)
			if (err == nil) != fixtureCase.Expected.Accepted {
				t.Fatalf("NewCommandInvocation(%q, %q) err=%v, want accepted=%v", fixtureCase.Input.Name, fixtureCase.Input.Args, err, fixtureCase.Expected.Accepted)
			}
			if !fixtureCase.Expected.Accepted {
				if fixtureCase.Expected.ErrorContains == "" {
					t.Fatal("must-fail case declares no errorContains needle")
				}
				if !strings.Contains(err.Error(), fixtureCase.Expected.ErrorContains) {
					t.Fatalf("error %q does not contain %q", err, fixtureCase.Expected.ErrorContains)
				}
				return
			}
			if invocation.Name != fixtureCase.Input.Name || invocation.Args != fixtureCase.Input.Args {
				t.Fatalf("constructor changed accepted bytes: got %+v", invocation)
			}
			if err := invocation.Validate(); err != nil {
				t.Fatalf("Validate rejected a constructed invocation: %v", err)
			}

			turn := schema.TurnDetail{
				Index:     0,
				Role:      schema.RoleUser,
				Content:   strings.TrimSpace(fixtureCase.Input.Name + " " + fixtureCase.Input.Args),
				Timestamp: turnCommandTimestamp,
				Command:   &invocation,
			}
			wire, err := json.Marshal(turn)
			if err != nil {
				t.Fatalf("marshal TurnDetail: %v", err)
			}
			var object map[string]json.RawMessage
			if err := json.Unmarshal(wire, &object); err != nil {
				t.Fatalf("decode marshaled TurnDetail: %v", err)
			}
			rawCommand, present := object["command"]
			if !present {
				t.Fatalf("command omitted from wire: %s", wire)
			}
			var command map[string]json.RawMessage
			if err := json.Unmarshal(rawCommand, &command); err != nil {
				t.Fatalf("decode command object: %v", err)
			}
			var gotName string
			if err := json.Unmarshal(command["name"], &gotName); err != nil || gotName != fixtureCase.Input.Name {
				t.Fatalf("command.name=%q err=%v, want %q", gotName, err, fixtureCase.Input.Name)
			}
			if _, hasArgs := command["args"]; hasArgs != fixtureCase.Expected.WireHasArgs {
				t.Fatalf("command.args present=%v, want %v; wire=%s", hasArgs, fixtureCase.Expected.WireHasArgs, wire)
			}
		})
	}
}

func TestTurnDetailOmitsAbsentCommand(t *testing.T) {
	wire, err := json.Marshal(schema.TurnDetail{Index: 0, Role: schema.RoleUser, Content: "hello", Timestamp: turnCommandTimestamp})
	if err != nil {
		t.Fatalf("marshal TurnDetail: %v", err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(wire, &object); err != nil {
		t.Fatalf("decode marshaled TurnDetail: %v", err)
	}
	if _, present := object["command"]; present {
		t.Fatalf("a turn without an invocation must not emit command: %s", wire)
	}
}

func TestNewCommandInvocationRejectsInvalidUTF8(t *testing.T) {
	_, err := schema.NewCommandInvocation("/aura\xff", "")
	if err == nil {
		t.Fatal("a name with an invalid UTF-8 byte must be rejected")
	}
	if !strings.Contains(err.Error(), "not valid UTF-8") {
		t.Fatalf("error %q does not name the UTF-8 defect", err)
	}
}

func TestTurnCommandFixtureInventoryRejectsCountPreservingRename(t *testing.T) {
	manifest, err := decodeTurnModelFixtureManifest(turnCommandManifestYAML)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	old := "name: " + manifest.RequiredCaseNames[0]
	mutated := strings.Replace(string(turnCommandFixtureYAML), old, "name: unregistered-command-case", 1)
	if mutated == string(turnCommandFixtureYAML) {
		t.Fatalf("mutation did not apply: %q not found", old)
	}
	corpus, err := testcase.LoadCorpus[turnCommandFixtureInput, turnCommandFixtureExpected]([]byte(mutated))
	if err != nil {
		t.Fatalf("load mutated corpus: %v", err)
	}
	if err := validateCorpusInventory("turn command", corpus, manifest); err == nil {
		t.Fatal("count-preserving rename passed the required-name inventory")
	}
}

func loadTurnCommandFixtures(t *testing.T) testcase.Corpus[turnCommandFixtureInput, turnCommandFixtureExpected] {
	t.Helper()
	corpus, err := testcase.LoadCorpus[turnCommandFixtureInput, turnCommandFixtureExpected](turnCommandFixtureYAML)
	if err != nil {
		t.Fatalf("load turn command corpus: %v", err)
	}
	manifest, err := decodeTurnModelFixtureManifest(turnCommandManifestYAML)
	if err != nil {
		t.Fatalf("load turn command manifest: %v", err)
	}
	if err := validateCorpusInventory("turn command", corpus, manifest); err != nil {
		t.Fatalf("validate turn command fixture inventory: %v", err)
	}
	return corpus
}

// validateCorpusInventory is the generic required-name inventory check: the
// corpus must hold exactly the manifest's cases, no more and no fewer, with no
// duplicate names. It is shared by every corpus in this package that carries a
// manifest of the same shape.
func validateCorpusInventory[I, E any](label string, corpus testcase.Corpus[I, E], manifest turnModelFixtureManifest) error {
	if len(corpus.Cases) != manifest.ExpectedCaseCount {
		return fmt.Errorf("%s corpus has %d cases, want exactly %d", label, len(corpus.Cases), manifest.ExpectedCaseCount)
	}
	required := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		required[name] = struct{}{}
	}
	actual := make(map[string]struct{}, len(corpus.Cases))
	for _, fixtureCase := range corpus.Cases {
		if _, duplicate := actual[fixtureCase.Name]; duplicate {
			return fmt.Errorf("%s corpus repeats case name %q", label, fixtureCase.Name)
		}
		actual[fixtureCase.Name] = struct{}{}
		if _, registered := required[fixtureCase.Name]; !registered {
			return fmt.Errorf("%s corpus contains unregistered case %q", label, fixtureCase.Name)
		}
	}
	for name := range required {
		if _, present := actual[name]; !present {
			return fmt.Errorf("%s corpus is missing required case %q", label, name)
		}
	}
	return nil
}
