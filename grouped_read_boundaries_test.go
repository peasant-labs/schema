package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/grouped_read_boundaries.yaml
var groupedReadYAML []byte

type groupedReadInput struct {
	Schema  string         `yaml:"schema"`
	Route   string         `yaml:"route"`
	Payload map[string]any `yaml:"payload"`
}
type groupedReadExpected struct {
	Valid bool `yaml:"valid"`
}
type groupedReadFixtures struct {
	RequiredNames []string                                               `yaml:"required_names"`
	Cases         testcase.Corpus[groupedReadInput, groupedReadExpected] `yaml:"cases"`
}

func loadGroupedReadFixtures(t *testing.T) groupedReadFixtures {
	t.Helper()
	var f groupedReadFixtures
	d := yaml.NewDecoder(bytes.NewReader(groupedReadYAML))
	d.KnownFields(true)
	if err := d.Decode(&f); err != nil {
		t.Fatal(err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		t.Fatalf("fixture must contain one document: %v", err)
	}
	assertcase.RequireMin(t, f.Cases, 1)
	assertcase.RequireValid(t, f.Cases)
	requireNames(t, "grouped read boundaries", f.RequiredNames, namesOf(f.Cases))
	return f
}

func TestGroupedReadBoundaries(t *testing.T) {
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}
	compiler := schema.NewJSONSchemaCompiler()
	if err := compiler.AddResource("types.json", bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
	village, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	villageBytes, err := json.Marshal(village)
	if err != nil {
		t.Fatal(err)
	}
	var villageRoot map[string]any
	if err := json.Unmarshal(villageBytes, &villageRoot); err != nil {
		t.Fatal(err)
	}
	if err := compiler.AddResource("village.json", bytes.NewReader(villageBytes)); err != nil {
		t.Fatal(err)
	}
	for _, c := range loadGroupedReadFixtures(t).Cases.Cases {
		t.Run(c.Name, func(t *testing.T) {
			raw, err := json.Marshal(c.Input.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var input any
			if err := json.Unmarshal(raw, &input); err != nil {
				t.Fatal(err)
			}
			contract, err := compiler.Compile("types.json#/components/schemas/" + c.Input.Schema)
			if err != nil {
				t.Fatal(err)
			}
			shapeErr := contract.Validate(input)
			if c.Input.Route != "" {
				op := villageRoot["paths"].(map[string]any)[c.Input.Route].(map[string]any)["get"].(map[string]any)
				response := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
				if !responseRefs(response)["Schema"+c.Input.Schema] {
					t.Fatalf("actual operation does not bind %s: %v", c.Input.Schema, response)
				}
				operationContract, err := compiler.Compile("village.json#/components/schemas/Schema" + c.Input.Schema)
				if err != nil {
					t.Fatal(err)
				}
				if (operationContract.Validate(input) == nil) != (shapeErr == nil) {
					t.Fatal("resolved operation response differs from canonical root schema")
				}
			}
			value, decodeErr := decodeGroupedRead(c.Input.Schema, raw)
			valid := shapeErr == nil && decodeErr == nil
			if valid != c.Expected.Valid {
				t.Fatalf("valid=%v want=%v schema=%v decode=%v", valid, c.Expected.Valid, shapeErr, decodeErr)
			}
			// The real Go member decoder must reject invalid semantics and raw
			// nulls independently of the JSON Schema test above.
			if c.Input.Schema == "LocalHelperMembersPayload" || c.Input.Schema == "VillageHelperMembersPayload" {
				if (decodeErr == nil) != c.Expected.Valid {
					t.Fatalf("Go member decoder parity: %v", decodeErr)
				}
			}
			if valid {
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var output any
				if err := json.Unmarshal(encoded, &output); err != nil {
					t.Fatal(err)
				}
				if err := requireJSONSubset(input, output, "$"); err != nil {
					t.Fatal(err)
				}
				if c.Input.Schema == "VillageGroupDetailRecord" {
					if !reflect.DeepEqual(input, output) {
						t.Fatalf("read settings fabricated or changed: %s", encoded)
					}
				}
			}
		})
	}
}

func decodeGroupedRead(name string, raw []byte) (any, error) {
	var value any
	switch name {
	case "LocalHelperMembersPayload":
		value = new(schema.LocalHelperMembersPayload)
	case "VillageHelperMembersPayload":
		value = new(schema.VillageHelperMembersPayload)
	case "VillageGroupDetailRecord":
		value = new(schema.VillageGroupDetailRecord)
	case "VillageGroup":
		value = new(schema.VillageGroup)
	case "TranscriptContent":
		value = new(schema.TranscriptContent)
	case "LocalSessionListPayload":
		value = new(schema.LocalSessionListPayload)
	case "VillageSessionListPayload":
		value = new(schema.VillageSessionListPayload)
	case "VillageGroupDetailResponse":
		value = new(schema.VillageGroupDetailResponse)
	case "VillageGroupedGroupDetailResponse":
		value = new(schema.VillageGroupedGroupDetailResponse)
	default:
		return nil, fmt.Errorf("unknown grouped fixture schema %q", name)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		return nil, err
	}
	if v, ok := value.(interface{ Validate() error }); ok {
		return value, v.Validate()
	}
	return value, nil
}
