package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type enumCatalogFixture struct {
	Enums []struct {
		Name      string                 `yaml:"name"`
		AllName   string                 `yaml:"all_name"`
		Members   []typeScriptEnumMember `yaml:"members"`
		AllValues []string               `yaml:"all_values"`
	} `yaml:"enums"`
}

func TestTypeScriptRuntimeEnumsMatchExactFixture(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "enums.yaml"))
	if err != nil {
		t.Fatalf("read enum fixture: %v", err)
	}
	var fixture enumCatalogFixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode enum fixture: %v", err)
	}
	descriptors, err := typeScriptEnums()
	if err != nil {
		t.Fatalf("build production enum descriptors: %v", err)
	}
	if len(fixture.Enums) != len(descriptors) {
		t.Fatalf("enum fixture rows=%d, descriptors=%d", len(fixture.Enums), len(descriptors))
	}
	for i, descriptor := range descriptors {
		row := fixture.Enums[i]
		if row.Name != descriptor.Name || row.AllName != descriptor.AllName {
			t.Fatalf("enum row %d identity=(%s,%s), want (%s,%s)", i, row.Name, row.AllName, descriptor.Name, descriptor.AllName)
		}
		if !equalEnumMembers(row.Members, descriptor.Members) || strings.Join(row.AllValues, "\x00") != strings.Join(descriptor.All, "\x00") {
			t.Fatalf("enum %s mapping differs from exact YAML catalog", descriptor.Name)
		}
	}
}

func TestWorkflowCoversEveryTypeScriptFixtureFamily(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "workflow_paths.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		RequiredPaths []string `yaml:"required_paths"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.RequiredPaths) != 3 {
		t.Fatalf("workflow path fixture has %d rows, want 3", len(fixture.RequiredPaths))
	}
	workflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "tests.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range fixture.RequiredPaths {
		if got := bytes.Count(workflow, []byte("- \""+path+"\"")); got != 2 {
			t.Errorf("workflow path %q occurs %d times, want once in pull_request and once in push", path, got)
		}
	}
}

func equalEnumMembers(a, b []typeScriptEnumMember) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTypeScriptCatalogCompletenessAndCollisions(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	for _, surface := range typeScriptSurfaces(root, filepath.Join(root, "typescript")) {
		got, err := loadSchemaAliases(surface.name, surface.spec)
		if err != nil {
			t.Fatalf("load %s aliases: %v", surface.name, err)
		}
		if _, err := deduplicateAliases(got); err != nil {
			t.Fatalf("deduplicate %s aliases: %v", surface.name, err)
		}
		if surface.name == "types" && len(got) != len(openapi.TypeCatalogEntries()) {
			t.Fatalf("Types aliases = %d, want one per production descriptor (%d)", len(got), len(openapi.TypeCatalogEntries()))
		}
	}
}

func TestTypeScriptCollisionFailsClosedOnUnequalSchemas(t *testing.T) {
	type collisionInput struct {
		Name         string `yaml:"name"`
		FirstSchema  string `yaml:"first_schema"`
		SecondSchema string `yaml:"second_schema"`
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "collision_cases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	matrix, err := testcase.LoadCorpus[collisionInput, bool](data)
	if err != nil {
		t.Fatal(err)
	}
	if err := matrix.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(matrix.Cases) != 2 {
		t.Fatalf("collision corpus has %d rows, want 2", len(matrix.Cases))
	}
	for _, tc := range matrix.Cases {
		aliases := []schemaAlias{
			{Name: tc.Input.Name, RawName: "Schema" + tc.Input.Name, Surface: "village", Canonical: []byte(tc.Input.FirstSchema)},
			{Name: tc.Input.Name, RawName: tc.Input.Name, Surface: "types", Canonical: []byte(tc.Input.SecondSchema)},
		}
		_, err := deduplicateAliases(aliases)
		if got := err == nil; got != tc.Expected {
			t.Fatalf("%s: accepted=%v, want %v", tc.Name, got, tc.Expected)
		}
		if err != nil && (!strings.Contains(err.Error(), tc.Input.Name) || !strings.Contains(err.Error(), "unequal canonical schemas")) {
			t.Fatalf("%s: collision error is not actionable: %v", tc.Name, err)
		}
	}
}

func TestGeneratedTypeScriptFilesFullyAccounted(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	sourceRoot := filepath.Join(root, "typescript", "src")
	var expected []string
	for _, path := range typeScriptGeneratedFiles(root, filepath.Join(root, "typescript")) {
		relative, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, filepath.ToSlash(relative))
	}
	var actual []string
	err = filepath.WalkDir(sourceRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".ts" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(data, []byte(generatedTypeScriptHead)) || bytes.HasPrefix(data, []byte("/**\n * This file was auto-generated by openapi-typescript.")) {
			relative, err := filepath.Rel(sourceRoot, path)
			if err != nil {
				return err
			}
			actual = append(actual, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk TypeScript sources: %v", err)
	}
	sort.Strings(expected)
	sort.Strings(actual)
	if strings.Join(actual, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("generated TypeScript files are not fully accounted\nactual:\n%s\nexpected:\n%s", strings.Join(actual, "\n"), strings.Join(expected, "\n"))
	}
}

func TestRenderQualityFixturesDeterministicAndComplete(t *testing.T) {
	fixtures, err := loadQualityFixturesForGenerator()
	if err != nil {
		t.Fatal(err)
	}
	first, err := renderQualityFixtures(fixtures)
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	second, err := renderQualityFixtures(fixtures)
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("quality TypeScript rendering is nondeterministic")
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(root, "typescript", "src", "fixtures", "quality.ts"))
	if err != nil {
		t.Fatalf("read generated quality fixture module: %v", err)
	}
	if !bytes.Equal(first, generated) {
		t.Fatal("generated quality fixture module differs from an exact production render")
	}
}

func loadQualityFixturesForGenerator() (*schema.QualityFixtures, error) {
	return schema.LoadQualityFixtures()
}
