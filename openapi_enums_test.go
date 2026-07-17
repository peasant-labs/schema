package schema_test

import (
	_ "embed"
	"reflect"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/typescript/enums.yaml
var closedEnumFixtureYAML []byte

type closedEnumFixture struct {
	Enums []closedEnumFixtureRow `yaml:"enums"`
}

type closedEnumFixtureRow struct {
	Name    string                    `yaml:"name"`
	Members []closedEnumFixtureMember `yaml:"members"`
}

type closedEnumFixtureMember struct {
	Value string `yaml:"value"`
}

func TestTypesOpenAPIExposesEveryClosedEnum(t *testing.T) {
	var fixture closedEnumFixture
	if err := yaml.Unmarshal(closedEnumFixtureYAML, &fixture); err != nil {
		t.Fatalf("decode closed enum fixture: %v", err)
	}
	if len(fixture.Enums) == 0 {
		t.Fatal("closed enum fixture has no enum definitions")
	}

	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types OpenAPI spec: %v", err)
	}
	for _, enum := range fixture.Enums {
		component, ok := spec.Components.Schemas[enum.Name]
		if !ok {
			t.Errorf("closed enum %s is absent from the Types OpenAPI catalog", enum.Name)
			continue
		}
		actual := schemaEnumStrings(t, enum.Name, component["enum"])
		expected := make([]string, len(enum.Members))
		for i, member := range enum.Members {
			expected[i] = member.Value
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("closed enum %s OpenAPI values=%v, want canonical Go values %v", enum.Name, actual, expected)
		}
	}
}

func schemaEnumStrings(t *testing.T, name string, raw any) []string {
	t.Helper()
	switch values := raw.(type) {
	case []any:
		out := make([]string, len(values))
		for i, value := range values {
			text, ok := value.(string)
			if !ok {
				t.Fatalf("closed enum %s value %d has type %T, want string", name, i, value)
			}
			out[i] = text
		}
		return out
	case []string:
		return values
	default:
		t.Fatalf("closed enum %s has OpenAPI enum type %T, want a string list", name, raw)
		return nil
	}
}
