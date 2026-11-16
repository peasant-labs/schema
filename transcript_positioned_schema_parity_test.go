package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed testdata/typescript/transcript_positioned_generated_shapes.yaml
var generatedPositionedShapesYAML []byte

//go:embed testdata/typescript/transcript_positioned_generated_shapes_manifest.yaml
var generatedPositionedShapesManifestYAML []byte

type generatedPositionedShapeInput struct {
	Schema string `yaml:"schema"`
	JSON   string `yaml:"json"`
}

// TestPositionedPublishedShapeParity exercises the actual Types document on
// the same cases the generated browser schemas replay.
func TestPositionedPublishedShapeParity(t *testing.T) {
	fixture := decodeStrictYAML[testcase.Corpus[generatedPositionedShapeInput, bool]](t, generatedPositionedShapesYAML, "positioned generated shapes")
	caseassert.RequireValid(t, fixture)
	names := make([]string, 0, len(fixture.Cases))
	for _, c := range fixture.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "positioned generated shapes", generatedPositionedShapesManifestYAML, names)
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("types.json", bytes.NewReader(document)); err != nil {
		t.Fatal(err)
	}
	components := map[string]string{
		"directory": "PositionedDirectory",
		"segment":   "PositionedSegment",
		"row":       "TranscriptSegmentRow",
		"window":    "PositionedWindow",
	}
	compiled := make(map[string]*jsonschema.Schema, len(components))
	for key, component := range components {
		compiled[key] = compilePublishedComponent(t, compiler, "types.json", component)
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			validator := compiled[c.Input.Schema]
			if validator == nil {
				t.Fatalf("unknown schema %q", c.Input.Schema)
			}
			var value any
			if err := json.Unmarshal([]byte(c.Input.JSON), &value); err != nil {
				t.Fatal(err)
			}
			err := validator.Validate(value)
			if (err == nil) != c.Expected || (c.Classification == testcase.MustPass) != c.Expected {
				t.Fatalf("published schema validation = %v, expected accepted=%t", err, c.Expected)
			}
		})
	}
}
