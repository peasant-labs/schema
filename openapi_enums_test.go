package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
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
	Name      string                    `yaml:"name"`
	AllName   string                    `yaml:"all_name"`
	GoAllVar  string                    `yaml:"go_all_var"`
	Members   []closedEnumFixtureMember `yaml:"members"`
	AllValues []any                     `yaml:"all_values"`
}

type closedEnumFixtureMember struct {
	Name  string `yaml:"name"`
	Value any    `yaml:"value"`
}

func TestTypesOpenAPIExposesEveryClosedEnum(t *testing.T) {
	fixture := loadClosedEnumFixture(t)
	if len(fixture.Enums) == 0 {
		t.Fatal("closed enum fixture has no enum definitions")
	}

	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types OpenAPI spec: %v", err)
	}
	fixtureNames := make([]string, 0, len(fixture.Enums))
	for _, enum := range fixture.Enums {
		fixtureNames = append(fixtureNames, enum.Name)
	}
	openAPIEnumNames := make([]string, 0)
	for name, component := range spec.Components.Schemas {
		if _, ok := component["enum"]; ok {
			openAPIEnumNames = append(openAPIEnumNames, name)
		}
	}
	assertSameStringSet(t, "Types OpenAPI enum components and TypeScript enum catalog", openAPIEnumNames, fixtureNames)

	for _, enum := range fixture.Enums {
		component, ok := spec.Components.Schemas[enum.Name]
		if !ok {
			t.Errorf("closed enum %s is absent from the Types OpenAPI catalog", enum.Name)
			continue
		}
		actual := schemaEnumJSONValues(t, enum.Name, component["enum"])
		expected := make([]string, len(enum.Members))
		for i, member := range enum.Members {
			expected[i] = mustJSONValue(t, enum.Name, member.Value)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("closed enum %s OpenAPI values=%v, want canonical Go values %v", enum.Name, actual, expected)
		}
	}
}

func loadClosedEnumFixture(t *testing.T) closedEnumFixture {
	t.Helper()
	var fixture closedEnumFixture
	decoder := yaml.NewDecoder(bytes.NewReader(closedEnumFixtureYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode closed enum fixture: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			t.Fatalf("decode trailing closed enum fixture document: %v", err)
		}
		t.Fatal("decode closed enum fixture: multiple YAML documents are not allowed")
	}
	return fixture
}

func assertSameStringSet(t *testing.T, label string, left, right []string) {
	t.Helper()
	leftSet, rightSet := make(map[string]struct{}, len(left)), make(map[string]struct{}, len(right))
	for _, name := range left {
		leftSet[name] = struct{}{}
	}
	for _, name := range right {
		rightSet[name] = struct{}{}
	}
	for name := range leftSet {
		if _, ok := rightSet[name]; !ok {
			t.Errorf("%s: %q only in Types OpenAPI components", label, name)
		}
	}
	for name := range rightSet {
		if _, ok := leftSet[name]; !ok {
			t.Errorf("%s: %q only in TypeScript enum catalog", label, name)
		}
	}
}

// schemaEnumJSONValues renders each OpenAPI enum value through encoding/json.
// Comparing the JSON text keeps the wire type exact: json.Marshal(25) is "25"
// while json.Marshal("25") is "\"25\"", so a numeric member can never satisfy a
// string expectation. fmt.Sprint is refused because it renders both as "25" and
// would pass while the wire type is wrong.
func schemaEnumJSONValues(t *testing.T, name string, raw any) []string {
	t.Helper()
	switch values := raw.(type) {
	case []any:
		out := make([]string, len(values))
		for i, value := range values {
			out[i] = mustJSONValue(t, name, value)
		}
		return out
	default:
		t.Fatalf("closed enum %s has OpenAPI enum type %T, want a value list", name, raw)
		return nil
	}
}

func mustJSONValue(t *testing.T, name string, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("closed enum %s value %v cannot be encoded: %v", name, value, err)
	}
	return string(encoded)
}
