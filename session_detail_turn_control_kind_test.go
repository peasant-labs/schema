package schema_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/session-detail/transcripts/turn_control_kind.yaml
var turnControlKindYAML []byte

//go:embed testdata/session-detail/transcripts/turn_control_kind_manifest.yaml
var turnControlKindManifestYAML []byte

type turnControlKindInput struct {
	// Turn is the JSON object members of the case turn; index, content,
	// timestamp, and depth default to a valid root turn when omitted.
	Turn string `yaml:"turn"`
	// Harness defaults to claude when omitted.
	Harness string `yaml:"harness,omitempty"`
}

type turnControlKindExpected struct {
	Accepted bool `yaml:"accepted"`
	// ControlKind is the expected decoded kind; empty means the producer
	// classified nothing and the field must be absent.
	ControlKind string `yaml:"controlKind,omitempty"`
	// WirePresent asserts whether a re-encoded payload names the controlKind
	// member on the case turn at all.
	WirePresent bool `yaml:"wirePresent"`
	// ErrorContains is required for a must-fail case.
	ErrorContains string `yaml:"errorContains,omitempty"`
}

func loadTurnControlKindFixtures(t *testing.T) testcase.Corpus[turnControlKindInput, turnControlKindExpected] {
	t.Helper()
	corpus, err := testcase.LoadCorpus[turnControlKindInput, turnControlKindExpected](turnControlKindYAML)
	if err != nil {
		t.Fatalf("load turn control kind corpus: %v", err)
	}
	assert.RequireValid(t, corpus)
	manifest, err := decodeTurnModelFixtureManifest(turnControlKindManifestYAML)
	if err != nil {
		t.Fatalf("load turn control kind manifest: %v", err)
	}
	if err := validateCorpusInventory("turn control kind", corpus, manifest); err != nil {
		t.Fatalf("validate turn control kind inventory: %v", err)
	}
	return corpus
}

// turnControlKindWire builds a durable session detail whose turns are a plain
// root assistant turn followed by the fixture turn, so nested fixture turns
// have a partition-local parent.
func turnControlKindWire(t *testing.T, in turnControlKindInput) []byte {
	t.Helper()
	turn := map[string]any{"index": 1, "content": "", "timestamp": "2020-01-01T00:00:00Z", "depth": 0}
	var patch map[string]any
	if err := json.Unmarshal([]byte(in.Turn), &patch); err != nil {
		t.Fatalf("fixture turn JSON: %v", err)
	}
	for k, v := range patch {
		turn[k] = v
	}
	root := map[string]any{"index": 0, "role": "assistant", "content": "", "timestamp": "2020-01-01T00:00:00Z", "depth": 0}
	harness := in.Harness
	if harness == "" {
		harness = "claude-code"
	}
	detail := map[string]any{
		"id": "fixture-session", "startTime": "2020-01-01T00:00:00Z", "endTime": "2020-01-01T00:00:00Z",
		"durationMins": 0, "totalTokens": 0, "tokensIn": 0, "tokensOut": 0, "turnCount": 2, "toolCallCount": 0,
		"harness": harness, "outcome": "resolved", "sessionOrigin": "unknown", "turns": []any{root, turn},
	}
	b, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestTurnControlKindFixture(t *testing.T) {
	for _, c := range loadTurnControlKindFixtures(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			payload, err := schema.DecodeSessionDetailPayloadRaw(turnControlKindWire(t, c.Input))
			if (err == nil) != c.Expected.Accepted {
				t.Fatalf("DecodeSessionDetailPayloadRaw accepted=%v, want %v: %v", err == nil, c.Expected.Accepted, err)
			}
			if err != nil {
				if !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("error %q does not contain %q", err, c.Expected.ErrorContains)
				}
				return
			}
			got := payload.Turns[1]
			var decoded string
			if got.ControlKind != nil {
				decoded = *got.ControlKind
			}
			if decoded != c.Expected.ControlKind {
				t.Fatalf("turn controlKind=%q, want %q", decoded, c.Expected.ControlKind)
			}
			if err := schema.ValidateSessionDetailPayload(payload); err != nil {
				t.Fatalf("accepted payload fails typed validation: %v", err)
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				Turns []map[string]json.RawMessage `json:"turns"`
			}
			if err := json.Unmarshal(encoded, &wire); err != nil {
				t.Fatalf("re-decode encoded payload: %v", err)
			}
			_, present := wire.Turns[1]["controlKind"]
			if present != c.Expected.WirePresent {
				t.Fatalf("wire controlKind present=%v, want %v", present, c.Expected.WirePresent)
			}
			again, err := schema.DecodeSessionDetailPayloadRaw(encoded)
			if err != nil {
				t.Fatalf("re-decode of re-encoded payload failed: %v", err)
			}
			if (again.Turns[1].ControlKind == nil) != (got.ControlKind == nil) {
				t.Fatal("controlKind presence changed across encode/decode")
			}
			if got.ControlKind != nil && *again.Turns[1].ControlKind != *got.ControlKind {
				t.Fatalf("controlKind=%q changed across encode/decode, want %q", *again.Turns[1].ControlKind, *got.ControlKind)
			}
		})
	}
}
