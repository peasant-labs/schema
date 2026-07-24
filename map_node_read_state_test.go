package schema_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/map_node_read_state.yaml
var mapNodeReadStateCasesYAML []byte

type mapNodeReadStateFixtureInput struct {
	ReadAttribution string `yaml:"readAttribution"`
	ReadState       string `yaml:"readState"`
}

// TestMapNode_ReadStateFixtureContract drives every read-state case in
// testdata/local-api/map_node_read_state.yaml against the real
// MapNode.Validate.
func TestMapNode_ReadStateFixtureContract(t *testing.T) {
	corpus, err := testcase.LoadCorpus[mapNodeReadStateFixtureInput, bool](mapNodeReadStateCasesYAML)
	if err != nil {
		t.Fatalf("load map node read state corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 6)
	assert.RequireValid(t, corpus)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			node := schema.MapNode{
				ID:              "internal/api",
				ReadAttribution: schema.ReadAttributionState(c.Input.ReadAttribution),
				ReadState:       schema.ReadStateGrade(c.Input.ReadState),
			}
			err := node.Validate()
			valid := err == nil
			if valid != c.Expected {
				t.Fatalf("Validate() error=%v (valid=%v), want valid=%v", err, valid, c.Expected)
			}
			if c.Classification == testcase.MustFail {
				for _, required := range []string{" at schema.", "during wire-boundary validation"} {
					if !strings.Contains(err.Error(), required) {
						t.Errorf("validation error %q is not actionable; missing %q", err, required)
					}
				}
			}
		})
	}
}
