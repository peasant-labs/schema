package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/deprecated_grouped_contributable.yaml
var deprecatedGroupedYAML []byte

type deprecatedGroupedInput struct {
	Schema  string                 `yaml:"schema"`
	Payload map[string]interface{} `yaml:"payload"`
	Keys    []string               `yaml:"keys"`
}

func TestDeprecatedGroupedContributableEnvelope(t *testing.T) {
	var fixture struct {
		Required  []string                                                 `yaml:"required_names"`
		Envelopes testcase.Corpus[deprecatedGroupedInput, villageExpected] `yaml:"envelopes"`
	}
	if err := yaml.Unmarshal(deprecatedGroupedYAML, &fixture); err != nil {
		t.Fatal(err)
	}
	assertcase.RequireValid(t, fixture.Envelopes)
	names := map[string]bool{}
	for _, row := range fixture.Envelopes.Cases {
		if names[row.Name] {
			t.Fatalf("duplicate case %s", row.Name)
		}
		names[row.Name] = true
		t.Run(row.Name, func(t *testing.T) {
			data, err := json.Marshal(row.Input.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var value interface{ Validate() error }
			switch row.Input.Schema {
			case "VillageGroupedContributableResponse":
				value = new(schema.VillageGroupedContributableResponse)
			case "VillageDeprecatedGroupedContributableResponse":
				value = new(schema.VillageDeprecatedGroupedContributableResponse)
			default:
				t.Fatalf("unknown schema %s", row.Input.Schema)
			}
			if err = json.Unmarshal(data, value); err == nil {
				err = value.Validate()
			}
			if (err == nil) != row.Expected.Valid {
				t.Fatalf("valid=%v want=%v: %v", err == nil, row.Expected.Valid, err)
			}
			if !row.Expected.Valid {
				return
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var roundTrip map[string]interface{}
			if err := json.Unmarshal(encoded, &roundTrip); err != nil {
				t.Fatal(err)
			}
			var original map[string]interface{}
			if err := json.Unmarshal(data, &original); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(roundTrip, original) {
				t.Fatalf("producer changed payload: %s", encoded)
			}
			keys := []string{}
			for key := range roundTrip {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			sort.Strings(row.Input.Keys)
			if !reflect.DeepEqual(keys, row.Input.Keys) {
				t.Fatalf("keys=%v want=%v", keys, row.Input.Keys)
			}
		})
	}
	requireNames(t, "deprecated grouped envelopes", fixture.Required, names)
}
