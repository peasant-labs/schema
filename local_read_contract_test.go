package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/local_read_contract.yaml
var localReadContractYAML []byte

type localReadInput struct {
	Kind string `yaml:"kind"`
	JSON string `yaml:"json"`
}
type localReadExpected struct {
	Valid         bool     `yaml:"valid"`
	RequiredKeys  []string `yaml:"requiredKeys"`
	ForbiddenKeys []string `yaml:"forbiddenKeys"`
}
type localReadFixtures struct {
	RequiredNames []string                                           `yaml:"required_names"`
	RoundTrip     testcase.Corpus[localReadInput, localReadExpected] `yaml:"round_trip"`
	Validation    testcase.Corpus[localReadInput, bool]              `yaml:"validation"`
}

func loadLocalReadFixtures(t *testing.T) localReadFixtures {
	t.Helper()
	var fixtures localReadFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(localReadContractYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("load local read contract: %v", err)
	}
	testassert.RequireMin(t, fixtures.RoundTrip, 7)
	testassert.RequireValid(t, fixtures.RoundTrip)
	testassert.RequireMin(t, fixtures.Validation, 3)
	testassert.RequireValid(t, fixtures.Validation)
	seen := map[string]bool{}
	for _, c := range fixtures.RoundTrip.Cases {
		seen[c.Name] = true
	}
	for _, c := range fixtures.Validation.Cases {
		seen[c.Name] = true
	}
	for _, name := range fixtures.RequiredNames {
		if !seen[name] {
			t.Fatalf("required local read case %q is missing", name)
		}
	}
	if len(seen) != len(fixtures.RequiredNames) {
		t.Fatalf("local read required-name manifest has %d names for %d cases", len(fixtures.RequiredNames), len(seen))
	}
	return fixtures
}

func TestLocalReadContractRoundTrip(t *testing.T) {
	for _, c := range loadLocalReadFixtures(t).RoundTrip.Cases {
		t.Run(c.Name, func(t *testing.T) {
			encoded, err := decodeValidateLocal(c.Input)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, want %v: %v", err == nil, c.Expected.Valid, err)
			}
			var root map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &root); err != nil {
				t.Fatal(err)
			}
			for _, key := range c.Expected.RequiredKeys {
				if key == "inputSubmissionCount" && c.Input.Kind == "row" {
					var row map[string]json.RawMessage
					_ = json.Unmarshal(root["session"], &row)
					if _, ok := row[key]; !ok {
						t.Fatalf("session missing %q", key)
					}
					continue
				}
				if _, ok := root[key]; !ok {
					t.Fatalf("root missing %q in %s", key, encoded)
				}
			}
			for _, key := range c.Expected.ForbiddenKeys {
				if _, ok := root[key]; ok {
					t.Fatalf("root unexpectedly contains %q", key)
				}
			}
		})
	}
}

func TestLocalReadContractValidation(t *testing.T) {
	for _, c := range loadLocalReadFixtures(t).Validation.Cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := decodeValidateLocal(c.Input)
			if (err == nil) != c.Expected {
				t.Fatalf("valid=%v, want %v", err == nil, c.Expected)
			}
		})
	}
}

func decodeValidateLocal(input localReadInput) ([]byte, error) {
	var value any
	switch input.Kind {
	case "detail":
		value = &schema.SessionDetailReadPayload{}
	case "row":
		value = &schema.LocalSessionRow{}
	case "item":
		value = &schema.LocalSessionListItem{}
	case "members":
		value = &schema.LocalHelperMembersPayload{}
	default:
		return nil, fmt.Errorf("unknown local read fixture kind %q", input.Kind)
	}
	if err := json.Unmarshal([]byte(input.JSON), value); err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case *schema.SessionDetailReadPayload:
		if err := schema.ValidateSessionDetailPayload(typed.SessionDetailPayload); err != nil {
			return nil, err
		}
		for _, navigation := range typed.RelationshipNavigation {
			if err := navigation.Validate(); err != nil {
				return nil, err
			}
		}
	case *schema.LocalSessionRow:
		if err := typed.Validate(); err != nil {
			return nil, err
		}
	case *schema.LocalSessionListItem:
		if err := typed.Validate(); err != nil {
			return nil, err
		}
	case *schema.LocalHelperMembersPayload:
		if err := typed.Validate(); err != nil {
			return nil, err
		}
	}
	return json.Marshal(value)
}
