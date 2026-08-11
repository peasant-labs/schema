package schema_test

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// turnModelFixtureYAML is the strict session-detail corpus shared with the
// generated TypeScript contract tests.
//
//go:embed testdata/session-detail/transcripts/turn_models.yaml
var turnModelFixtureYAML []byte

type turnModelFixtureInput struct {
	Index     int         `yaml:"index"`
	Role      schema.Role `yaml:"role"`
	Content   string      `yaml:"content"`
	Timestamp time.Time   `yaml:"timestamp"`
	Depth     int         `yaml:"depth"`
	Model     *string     `yaml:"model,omitempty"`
}

type turnModelFixtureExpected struct {
	ModelPresent bool   `yaml:"model_present"`
	Model        string `yaml:"model,omitempty"`
}

func TestTurnDetailModelFixture(t *testing.T) {
	corpus, err := testcase.LoadCorpus[turnModelFixtureInput, turnModelFixtureExpected](turnModelFixtureYAML)
	if err != nil {
		t.Fatalf("load turn model corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 2)
	if len(corpus.Cases) != 2 {
		t.Fatalf("turn model corpus has %d cases, want exactly 2", len(corpus.Cases))
	}
	assert.RequireValid(t, corpus)

	present := 0
	absent := 0
	for _, fixtureCase := range corpus.Cases {
		t.Run(fixtureCase.Name, func(t *testing.T) {
			input := fixtureCase.Input
			if !input.Role.IsValid() {
				t.Fatalf("fixture role %q is not a known role", input.Role)
			}
			if (input.Model != nil) != fixtureCase.Expected.ModelPresent {
				t.Fatalf("fixture model presence=%v, expected presence=%v", input.Model != nil, fixtureCase.Expected.ModelPresent)
			}

			model := ""
			if input.Model != nil {
				model = *input.Model
			}
			turn := schema.TurnDetail{
				Index:     input.Index,
				Role:      input.Role,
				Content:   input.Content,
				Timestamp: input.Timestamp,
				Depth:     input.Depth,
				Model:     model,
			}
			wire, err := json.Marshal(turn)
			if err != nil {
				t.Fatalf("marshal TurnDetail: %v", err)
			}

			var object map[string]json.RawMessage
			if err := json.Unmarshal(wire, &object); err != nil {
				t.Fatalf("decode marshaled TurnDetail: %v", err)
			}
			rawModel, emitted := object["model"]
			if emitted != fixtureCase.Expected.ModelPresent {
				t.Fatalf("model emitted=%v, expected emitted=%v; wire=%s", emitted, fixtureCase.Expected.ModelPresent, wire)
			}
			if !emitted {
				absent++
				return
			}
			present++

			var got string
			if err := json.Unmarshal(rawModel, &got); err != nil {
				t.Fatalf("decode emitted model: %v", err)
			}
			if got != fixtureCase.Expected.Model {
				t.Fatalf("emitted model=%q, want %q", got, fixtureCase.Expected.Model)
			}
		})
	}
	if present != 1 || absent != 1 {
		t.Fatalf("fixture coverage present=%d absent=%d, want one of each", present, absent)
	}
}
