package schema_test

import (
	"bytes"
	_ "embed"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/project_analytics.yaml
var projectAnalyticsManifestYAML []byte

type projectAnalyticsManifest struct {
	Types []projectAnalyticsManifestType `yaml:"types"`
}

type projectAnalyticsManifestType struct {
	Name   string   `yaml:"name"`
	Fields []string `yaml:"fields"`
}

// projectAnalyticsTypes is the analytics bundle the manifest names. The Go
// reflection check and the generated OpenAPI component are both compared with
// the same fixture, so the struct and the generated TypeScript type cannot
// drift from the dashboard typedef.
var projectAnalyticsTypes = map[string]reflect.Type{
	"ProjectAnalytics":         reflect.TypeOf(schema.ProjectAnalytics{}),
	"WeekCount":                reflect.TypeOf(schema.WeekCount{}),
	"WeekContributors":         reflect.TypeOf(schema.WeekContributors{}),
	"ReturningContributorRate": reflect.TypeOf(schema.ReturningContributorRate{}),
	"LongestStreak":            reflect.TypeOf(schema.LongestStreak{}),
	"WeekNewContributors":      reflect.TypeOf(schema.WeekNewContributors{}),
	"SessionToCommitRate":      reflect.TypeOf(schema.SessionToCommitRate{}),
	"WeekAvgDuration":          reflect.TypeOf(schema.WeekAvgDuration{}),
	"OutcomeDistribution":      reflect.TypeOf(schema.OutcomeDistribution{}),
	"MedianP90":                reflect.TypeOf(schema.MedianP90{}),
	"AnalyticsSessionStats":    reflect.TypeOf(schema.AnalyticsSessionStats{}),
	"ContributorBreakdown":     reflect.TypeOf(schema.ContributorBreakdown{}),
}

func TestProjectAnalyticsFieldSetMatchesManifest(t *testing.T) {
	var manifest projectAnalyticsManifest
	decoder := yaml.NewDecoder(bytes.NewReader(projectAnalyticsManifestYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode testdata/project_analytics.yaml: %v", err)
	}
	if len(manifest.Types) != len(projectAnalyticsTypes) {
		t.Fatalf("manifest declares %d types, want %d", len(manifest.Types), len(projectAnalyticsTypes))
	}

	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}

	seen := make(map[string]struct{}, len(manifest.Types))
	for _, declared := range manifest.Types {
		if _, duplicate := seen[declared.Name]; duplicate {
			t.Fatalf("manifest declares %s more than once", declared.Name)
		}
		seen[declared.Name] = struct{}{}

		valueType, ok := projectAnalyticsTypes[declared.Name]
		if !ok {
			t.Fatalf("manifest type %s has no Go type in the analytics bundle", declared.Name)
		}
		assertSameFieldSet(t, "Go struct "+declared.Name, goJSONFields(valueType), declared.Fields)

		component, ok := spec.Components.Schemas[declared.Name]
		if !ok {
			t.Fatalf("Types spec has no component %s; add the type to openapi.TypeCatalogEntries", declared.Name)
		}
		properties, ok := component["properties"].(map[string]any)
		if !ok || len(properties) == 0 {
			t.Fatalf("Types component %s has no properties", declared.Name)
		}
		generated := make([]string, 0, len(properties))
		for name := range properties {
			generated = append(generated, name)
		}
		assertSameFieldSet(t, "generated TypeScript type "+declared.Name, generated, declared.Fields)
	}
	for name := range projectAnalyticsTypes {
		if _, ok := seen[name]; !ok {
			t.Fatalf("manifest is missing the analytics type %s", name)
		}
	}
}

// goJSONFields returns the JSON property names a Go struct emits, including
// promoted fields of an embedded struct.
func goJSONFields(valueType reflect.Type) []string {
	var fields []string
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if !field.IsExported() {
			continue
		}
		if field.Anonymous {
			fields = append(fields, goJSONFields(field.Type)...)
			continue
		}
		parts := strings.Split(field.Tag.Get("json"), ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		fields = append(fields, name)
	}
	return fields
}

func assertSameFieldSet(t *testing.T, label string, actual, expected []string) {
	t.Helper()
	actualSet := make(map[string]struct{}, len(actual))
	for _, name := range actual {
		if _, duplicate := actualSet[name]; duplicate {
			t.Fatalf("%s repeats field %q", label, name)
		}
		actualSet[name] = struct{}{}
	}
	expectedSet := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		expectedSet[name] = struct{}{}
	}
	if len(actualSet) != len(expectedSet) {
		t.Fatalf("%s has %d fields, want %d: %v", label, len(actualSet), len(expectedSet), expected)
	}
	for name := range expectedSet {
		if _, ok := actualSet[name]; !ok {
			t.Fatalf("%s is missing field %q", label, name)
		}
	}
	for name := range actualSet {
		if _, ok := expectedSet[name]; !ok {
			t.Fatalf("%s carries extra field %q", label, name)
		}
	}
}
