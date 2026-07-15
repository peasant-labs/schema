package openapi_test

import (
	"embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/typescript_catalog.yaml testdata/typescript_requiredness.yaml
var typeCatalogFixtures embed.FS

type catalogFixture struct {
	Types []catalogFixtureRow `yaml:"types"`
}

type catalogFixtureRow struct {
	GoName      string `yaml:"go_name"`
	Disposition string `yaml:"disposition"`
	Component   string `yaml:"component"`
	Reason      string `yaml:"reason"`
}

func TestTypesCatalogAccountsForEveryExportedRootType(t *testing.T) {
	var fixture catalogFixture
	decodeStrictFixture(t, "testdata/typescript_catalog.yaml", &fixture)

	declared := exportedRootTypes(t)
	classified := map[string]catalogFixtureRow{}
	for _, row := range fixture.Types {
		if strings.TrimSpace(row.GoName) == "" || strings.TrimSpace(row.Reason) == "" {
			t.Fatalf("catalog classification has blank name or reason: %+v", row)
		}
		switch row.Disposition {
		case "catalog", "alias", "runtime-only", "fixture-only":
		default:
			t.Fatalf("type %s has unknown disposition %q", row.GoName, row.Disposition)
		}
		if _, duplicate := classified[row.GoName]; duplicate {
			t.Fatalf("type %s is classified more than once", row.GoName)
		}
		classified[row.GoName] = row
	}
	assertSameStringSet(t, "exported Go types and classifications", declared, keys(classified))

	descriptors := map[string]string{}
	for _, descriptor := range specpkg.TypeCatalogEntries() {
		if _, duplicate := descriptors[descriptor.Name]; duplicate {
			t.Fatalf("Types catalog descriptor %s is duplicated", descriptor.Name)
		}
		descriptors[descriptor.Name] = descriptor.Name
	}
	var catalogNames []string
	for _, row := range fixture.Types {
		if row.Disposition == "catalog" {
			if row.Component != row.GoName {
				t.Fatalf("catalog type %s must preserve its Go name, got component %q", row.GoName, row.Component)
			}
			catalogNames = append(catalogNames, row.Component)
		}
	}
	assertSameStringSet(t, "catalog classifications and production descriptors", catalogNames, keys(descriptors))

	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	componentNames := make([]string, 0, len(spec.Components.Schemas))
	for name := range spec.Components.Schemas {
		componentNames = append(componentNames, name)
	}
	assertSameStringSet(t, "production descriptors and emitted components", keys(descriptors), componentNames)
	for _, forbidden := range []string{"Provider", "BestiaryHarness"} {
		if _, leaked := spec.Components.Schemas[forbidden]; leaked {
			t.Fatalf("Types spec leaked historical Harness name %q", forbidden)
		}
	}
}

type requirednessInput struct {
	Component   string   `yaml:"component"`
	Required    []string `yaml:"required"`
	Optional    []string `yaml:"optional"`
	Nullable    []string `yaml:"nullable"`
	Nonnullable []string `yaml:"nonnullable"`
}

func TestTypesCatalogPreservesGoJSONRequiredness(t *testing.T) {
	var fixture testcase.Corpus[requirednessInput, bool]
	decodeStrictFixture(t, "testdata/typescript_requiredness.yaml", &fixture)
	if err := fixture.Validate(); err != nil {
		t.Fatalf("validate requiredness fixture: %v", err)
	}
	if len(fixture.Cases) != 5 {
		t.Fatalf("requiredness fixture has %d rows, want exactly 5 representative structures", len(fixture.Cases))
	}
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	for _, tc := range fixture.Cases {
		if !tc.Expected {
			t.Fatalf("requiredness fixture %q must expect parity", tc.Name)
		}
		input := tc.Input
		schemaMap := spec.Components.Schemas[input.Component]
		required := map[string]struct{}{}
		switch raw := schemaMap["required"].(type) {
		case []interface{}:
			for _, value := range raw {
				required[value.(string)] = struct{}{}
			}
		case []string:
			for _, value := range raw {
				required[value] = struct{}{}
			}
		default:
			t.Fatalf("component %s has no required array", input.Component)
		}
		for _, name := range input.Required {
			if _, ok := required[name]; !ok {
				t.Errorf("component %s field %s is optional in Types but required by Go JSON", input.Component, name)
			}
		}
		for _, name := range input.Optional {
			if _, ok := required[name]; ok {
				t.Errorf("component %s field %s is required in Types but uses omitempty in Go JSON", input.Component, name)
			}
		}
		properties := schemaMap["properties"].(map[string]interface{})
		for _, name := range input.Nullable {
			if !schemaAllowsNull(properties[name].(map[string]interface{})) {
				t.Errorf("component %s field %s is a Go pointer but Types rejects null", input.Component, name)
			}
		}
		for _, name := range input.Nonnullable {
			if schemaAllowsNull(properties[name].(map[string]interface{})) {
				t.Errorf("component %s field %s is not a Go pointer but Types allows null", input.Component, name)
			}
		}
	}
}

func schemaAllowsNull(schemaMap map[string]interface{}) bool {
	if types, ok := schemaMap["type"].([]interface{}); ok {
		for _, value := range types {
			if value == "null" {
				return true
			}
		}
	}
	if alternatives, ok := schemaMap["anyOf"].([]interface{}); ok {
		for _, value := range alternatives {
			if candidate, ok := value.(map[string]interface{}); ok && candidate["type"] == "null" {
				return true
			}
		}
	}
	return false
}

func exportedRootTypes(t *testing.T) []string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(file))
	packages, err := parser.ParseDir(token.NewFileSet(), root, func(info fs.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse root Go package: %v", err)
	}
	var names []string
	for _, file := range packages["schema"].Files {
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				if typeSpec.Name.IsExported() {
					names = append(names, typeSpec.Name.Name)
				}
			}
		}
	}
	return names
}

func decodeStrictFixture(t *testing.T, path string, target interface{}) {
	t.Helper()
	data, err := typeCatalogFixtures.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func keys[V any](values map[string]V) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	return out
}

func assertSameStringSet(t *testing.T, label string, left, right []string) {
	t.Helper()
	l, r := map[string]struct{}{}, map[string]struct{}{}
	for _, value := range left {
		l[value] = struct{}{}
	}
	for _, value := range right {
		r[value] = struct{}{}
	}
	for value := range l {
		if _, ok := r[value]; !ok {
			t.Errorf("%s: %q only on left", label, value)
		}
	}
	for value := range r {
		if _, ok := l[value]; !ok {
			t.Errorf("%s: %q only on right", label, value)
		}
	}
}
