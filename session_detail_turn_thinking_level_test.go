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

//go:embed testdata/session-detail/transcripts/turn_thinking_levels.yaml
var turnThinkingLevelYAML []byte

//go:embed testdata/session-detail/transcripts/turn_thinking_levels_manifest.yaml
var turnThinkingLevelManifestYAML []byte

type turnThinkingLevelInput struct {
	// Turn is the JSON object members of the case turn; index, content,
	// timestamp, and depth default to a valid root turn when omitted.
	Turn string `yaml:"turn"`
	// Session is optional JSON object members merged onto the session detail.
	Session string `yaml:"session,omitempty"`
}

type turnThinkingLevelChangeExpected struct {
	Level string `yaml:"level,omitempty"`
	Raw   string `yaml:"raw,omitempty"`
}

type turnThinkingLevelExpected struct {
	Accepted                    bool                              `yaml:"accepted"`
	ThinkingLevel               string                            `yaml:"thinkingLevel,omitempty"`
	ThinkingLevelRaw            string                            `yaml:"thinkingLevelRaw,omitempty"`
	SessionThinkingLevel        string                            `yaml:"sessionThinkingLevel,omitempty"`
	SessionThinkingLevelHistory []turnThinkingLevelChangeExpected `yaml:"sessionThinkingLevelHistory,omitempty"`
	ErrorContains               string                            `yaml:"errorContains,omitempty"`
}

func loadTurnThinkingLevelFixtures(t *testing.T) testcase.Corpus[turnThinkingLevelInput, turnThinkingLevelExpected] {
	t.Helper()
	corpus, err := testcase.LoadCorpus[turnThinkingLevelInput, turnThinkingLevelExpected](turnThinkingLevelYAML)
	if err != nil {
		t.Fatalf("load turn thinking level corpus: %v", err)
	}
	assert.RequireValid(t, corpus)
	manifest, err := decodeTurnModelFixtureManifest(turnThinkingLevelManifestYAML)
	if err != nil {
		t.Fatalf("load turn thinking level manifest: %v", err)
	}
	if err := validateCorpusInventory("turn thinking level", corpus, manifest); err != nil {
		t.Fatalf("validate turn thinking level inventory: %v", err)
	}
	return corpus
}

// turnThinkingLevelWire builds a durable session detail whose turns are a plain
// root assistant turn followed by the fixture turn, so nested fixture turns
// have a partition-local parent.
func turnThinkingLevelWire(t *testing.T, in turnThinkingLevelInput) []byte {
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
	detail := map[string]any{
		"id": "fixture-session", "startTime": "2020-01-01T00:00:00Z", "endTime": "2020-01-01T00:00:00Z",
		"durationMins": 0, "totalTokens": 0, "tokensIn": 0, "tokensOut": 0, "turnCount": 2, "toolCallCount": 0,
		"harness": "codex", "outcome": "resolved", "sessionOrigin": "unknown", "turns": []any{root, turn},
	}
	if in.Session != "" {
		var session map[string]any
		if err := json.Unmarshal([]byte(in.Session), &session); err != nil {
			t.Fatalf("fixture session JSON: %v", err)
		}
		for k, v := range session {
			detail[k] = v
		}
	}
	b, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestTurnDetailThinkingLevelFixture(t *testing.T) {
	for _, c := range loadTurnThinkingLevelFixtures(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			payload, err := schema.DecodeSessionDetailPayloadRaw(turnThinkingLevelWire(t, c.Input))
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
			if string(got.ThinkingLevel) != c.Expected.ThinkingLevel || string(got.ThinkingLevelRaw) != c.Expected.ThinkingLevelRaw {
				t.Fatalf("turn level/raw=%q/%q, want %q/%q", got.ThinkingLevel, got.ThinkingLevelRaw, c.Expected.ThinkingLevel, c.Expected.ThinkingLevelRaw)
			}
			if string(payload.ThinkingLevel) != c.Expected.SessionThinkingLevel {
				t.Fatalf("session level=%q, want %q", payload.ThinkingLevel, c.Expected.SessionThinkingLevel)
			}
			if len(payload.ThinkingLevelHistory) != len(c.Expected.SessionThinkingLevelHistory) {
				t.Fatalf("session history=%v, want %v", payload.ThinkingLevelHistory, c.Expected.SessionThinkingLevelHistory)
			}
			for i, want := range c.Expected.SessionThinkingLevelHistory {
				got := payload.ThinkingLevelHistory[i]
				if string(got.Level) != want.Level || string(got.Raw) != want.Raw {
					t.Fatalf("session history[%d]=%q/%q, want %q/%q", i, got.Level, got.Raw, want.Level, want.Raw)
				}
			}
			if err := schema.ValidateSessionDetailPayload(payload); err != nil {
				t.Fatalf("accepted payload fails typed validation: %v", err)
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			again, err := schema.DecodeSessionDetailPayloadRaw(encoded)
			if err != nil {
				t.Fatalf("re-decode of re-encoded payload failed: %v", err)
			}
			if again.Turns[1].ThinkingLevel != got.ThinkingLevel || again.Turns[1].ThinkingLevelRaw != got.ThinkingLevelRaw || again.ThinkingLevel != payload.ThinkingLevel {
				t.Fatal("thinking level evidence changed across encode/decode")
			}
			if len(again.ThinkingLevelHistory) != len(payload.ThinkingLevelHistory) {
				t.Fatal("thinking level history changed across encode/decode")
			}
			for i := range payload.ThinkingLevelHistory {
				if again.ThinkingLevelHistory[i] != payload.ThinkingLevelHistory[i] {
					t.Fatalf("thinking level history[%d] changed across encode/decode", i)
				}
			}
		})
	}
}
