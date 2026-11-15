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

//go:embed testdata/typescript/transcript_v2_generated_shapes.yaml
var generatedV2ShapesYAML []byte

//go:embed testdata/typescript/transcript_v2_generated_shapes_manifest.yaml
var generatedV2ShapesManifestYAML []byte

type generatedV2ShapeInput struct {
	Schema string `yaml:"schema"`
	JSON   string `yaml:"json"`
}

// TestTranscriptV2PublishedShapeParity exercises the actual Types document on
// the same cases the generated browser schemas replay.
func TestTranscriptV2PublishedShapeParity(t *testing.T) {
	fixture := decodeStrictYAML[testcase.Corpus[generatedV2ShapeInput, bool]](t, generatedV2ShapesYAML, "v2 generated shapes")
	caseassert.RequireValid(t, fixture)
	names := make([]string, 0, len(fixture.Cases))
	for _, c := range fixture.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "v2 generated shapes", generatedV2ShapesManifestYAML, names)
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
		"directory": "TranscriptSegmentDirectoryV2",
		"segment":   "TranscriptSegmentV2",
		"row":       "TranscriptSegmentRow",
		"window":    "TranscriptWindowV2",
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
