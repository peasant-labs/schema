package openapi_test

import (
	"embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
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
	ForbiddenComponents []string            `yaml:"forbidden_components"`
	Types               []catalogFixtureRow `yaml:"types"`
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
	if len(fixture.ForbiddenComponents) == 0 {
		t.Fatal("catalog fixture must name historical components that may not be reintroduced")
	}
	seenForbidden := map[string]struct{}{}
	for _, forbidden := range fixture.ForbiddenComponents {
		if strings.TrimSpace(forbidden) == "" {
			t.Fatal("catalog fixture contains a blank forbidden component")
		}
		if _, duplicate := seenForbidden[forbidden]; duplicate {
			t.Fatalf("catalog fixture repeats forbidden component %q", forbidden)
		}
		seenForbidden[forbidden] = struct{}{}
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

// TestTypesCatalogPreservesListedPropertyRequiredness keeps the representative
// fixture honest: every property listed in a case must be classified exactly
// once on the presence axis (required or optional) and exactly once on the
// nullability axis (nullable or nonnullable). This is representative coverage,
// not an exhaustive catalog of every emitted property.
func TestTypesCatalogPreservesListedPropertyRequiredness(t *testing.T) {
	var fixture testcase.Corpus[requirednessInput, bool]
	decodeStrictFixture(t, "testdata/typescript_requiredness.yaml", &fixture)
	if err := fixture.Validate(); err != nil {
		t.Fatalf("validate requiredness fixture: %v", err)
	}
	if len(fixture.Cases) != 13 {
		t.Fatalf("requiredness fixture has %d rows, want exactly 13 representative structures", len(fixture.Cases))
	}
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	for _, tc := range fixture.Cases {
		if !tc.Expected {
			t.Fatalf("requiredness fixture %q must expect parity", tc.Name)
		}
		requireListedPropertyRequirednessClassification(t, tc.Input.Component, spec.Components.Schemas[tc.Input.Component], tc.Input)
	}
}

func requireListedPropertyRequirednessClassification(t *testing.T, component string, schemaMap map[string]interface{}, input requirednessInput) {
	t.Helper()

	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("component %s has no properties map", component)
	}
	requiredNames := schemaRequiredNames(t, component, schemaMap)

	required := collectUniqueNames(t, component, "required", input.Required)
	optional := collectUniqueNames(t, component, "optional", input.Optional)
	nullable := collectUniqueNames(t, component, "nullable", input.Nullable)
	nonnullable := collectUniqueNames(t, component, "nonnullable", input.Nonnullable)

	listed := map[string]struct{}{}
	for _, set := range []map[string]struct{}{required, optional, nullable, nonnullable} {
		for name := range set {
			listed[name] = struct{}{}
		}
	}

	for name := range listed {
		prop, ok := properties[name]
		if !ok {
			t.Fatalf("component %s fixture references property %q that the spec does not emit", component, name)
		}

		presenceCount := 0
		if _, ok := required[name]; ok {
			presenceCount++
		}
		if _, ok := optional[name]; ok {
			presenceCount++
		}
		if presenceCount != 1 {
			t.Fatalf("component %s property %q is classified %d times on the presence axis; want exactly once as required or optional", component, name, presenceCount)
		}

		nullabilityCount := 0
		if _, ok := nullable[name]; ok {
			nullabilityCount++
		}
		if _, ok := nonnullable[name]; ok {
			nullabilityCount++
		}
		if nullabilityCount != 1 {
			t.Fatalf("component %s property %q is classified %d times on the nullability axis; want exactly once as nullable or nonnullable", component, name, nullabilityCount)
		}

		if _, ok := required[name]; ok {
			if _, ok := requiredNames[name]; !ok {
				t.Fatalf("component %s property %q is required in the fixture but not required in the generated Types spec", component, name)
			}
		} else if _, ok := requiredNames[name]; ok {
			t.Fatalf("component %s property %q is optional in the fixture but required in the generated Types spec", component, name)
		}

		propertySchema, ok := prop.(map[string]interface{})
		if !ok {
			t.Fatalf("component %s property %q has unexpected schema type %T", component, name, prop)
		}
		allowsNull := schemaAllowsNull(propertySchema)
		if _, ok := nullable[name]; ok {
			if !allowsNull {
				t.Fatalf("component %s property %q is classified nullable in the fixture but the Types schema rejects null", component, name)
			}
		} else if allowsNull {
			t.Fatalf("component %s property %q is classified nonnullable in the fixture but the Types schema allows null", component, name)
		}
	}
}

func schemaRequiredNames(t *testing.T, component string, schemaMap map[string]interface{}) map[string]struct{} {
	t.Helper()

	required := map[string]struct{}{}
	switch raw := schemaMap["required"].(type) {
	case []interface{}:
		for _, value := range raw {
			name, ok := value.(string)
			if !ok {
				t.Fatalf("component %s has non-string required entry of type %T", component, value)
			}
			required[name] = struct{}{}
		}
	case []string:
		for _, name := range raw {
			required[name] = struct{}{}
		}
	case nil:
		return required
	default:
		t.Fatalf("component %s has unexpected required-list type %T", component, raw)
	}
	return required
}

func collectUniqueNames(t *testing.T, component, axis string, values []string) map[string]struct{} {
	t.Helper()

	set := make(map[string]struct{}, len(values))
	for _, raw := range values {
		name := strings.TrimSpace(raw)
		if name == "" {
			t.Fatalf("component %s has a blank %s property name", component, axis)
		}
		if name != raw {
			t.Fatalf("component %s property %q on the %s axis has surrounding whitespace", component, raw, axis)
		}
		if _, duplicate := set[name]; duplicate {
			t.Fatalf("component %s repeats %s property %q", component, axis, name)
		}
		set[name] = struct{}{}
	}
	return set
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
	if _, err := os.Stat(root); err != nil {
		workingDirectory, workingErr := os.Getwd()
		if workingErr != nil {
			t.Fatalf("locate root Go package after trimmed build path %q: %v", root, workingErr)
		}
		root = filepath.Dir(workingDirectory)
	}
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
