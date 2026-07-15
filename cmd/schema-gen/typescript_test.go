package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type apiComponentIdentityMutation struct {
	Surface     string `yaml:"surface"`
	Component   string `yaml:"component"`
	Kind        string `yaml:"kind"`
	Target      string `yaml:"target"`
	Replacement string `yaml:"replacement"`
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

func TestCanonicalizeAPIOperationReferencesRejectsUnequalCanonicalCopies(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "api_component_identity.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	matrix, err := testcase.LoadCorpus[apiComponentIdentityMutation, bool](data)
	if err != nil {
		t.Fatal(err)
	}
	if len(matrix.Cases) != 2 {
		t.Fatalf("API component identity fixture has %d rows, want exactly 2", len(matrix.Cases))
	}
	surfaces := typeScriptSurfaces(root, filepath.Join(root, "typescript"))
	rootAliases, err := loadSchemaAliases(surfaces[0].name, surfaces[0].spec)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range matrix.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			surface, ok := typeScriptSurfaceByName(surfaces, tc.Input.Surface)
			if !ok {
				t.Fatalf("unknown TypeScript surface %q; use a surface from typeScriptSurfaces", tc.Input.Surface)
			}
			aliases, err := loadSchemaAliases(surface.name, surface.spec)
			if err != nil {
				t.Fatal(err)
			}
			if err := mutateAPIComponentAlias(aliases, tc.Input); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(surface.rawOutput)
			if err != nil {
				t.Fatal(err)
			}
			surface.rawOutput = filepath.Join(t.TempDir(), filepath.Base(surface.rawOutput))
			if err := os.WriteFile(surface.rawOutput, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			err = canonicalizeAPIOperationReferences(surface, aliases, rootAliases)
			accepted := err == nil
			if accepted != tc.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, tc.Expected)
			}
			if err != nil && (!strings.Contains(err.Error(), tc.Input.Component) || !strings.Contains(err.Error(), "schema is unequal")) {
				t.Fatalf("production canonicalization error is not actionable for %s: %v", tc.Input.Component, err)
			}
		})
	}
}

func typeScriptSurfaceByName(surfaces []typeScriptSurface, name string) (typeScriptSurface, bool) {
	for _, surface := range surfaces {
		if surface.name == name {
			return surface, true
		}
	}
	return typeScriptSurface{}, false
}

func mutateAPIComponentAlias(aliases []schemaAlias, mutation apiComponentIdentityMutation) error {
	for index := range aliases {
		if aliases[index].RawName != mutation.Component {
			continue
		}
		var document map[string]any
		if err := json.Unmarshal(aliases[index].Canonical, &document); err != nil {
			return fmt.Errorf("decode canonical API component %s: %w", mutation.Component, err)
		}
		switch mutation.Kind {
		case "change-property-type":
			properties, ok := document["properties"].(map[string]any)
			if !ok {
				return fmt.Errorf("API component %s has no properties object", mutation.Component)
			}
			property, ok := properties[mutation.Target].(map[string]any)
			if !ok {
				return fmt.Errorf("API component %s has no property %s", mutation.Component, mutation.Target)
			}
			property["type"] = mutation.Replacement
		case "drop-required":
			required, ok := document["required"].([]any)
			if !ok {
				return fmt.Errorf("API component %s has no required array", mutation.Component)
			}
			filtered := make([]any, 0, len(required))
			for _, field := range required {
				if field != mutation.Target {
					filtered = append(filtered, field)
				}
			}
			document["required"] = filtered
		default:
			return fmt.Errorf("unknown API component mutation %q", mutation.Kind)
		}
		canonical, err := json.Marshal(document)
		if err != nil {
			return fmt.Errorf("encode mutated API component %s: %w", mutation.Component, err)
		}
		aliases[index].Canonical = canonical
		return nil
	}
	return fmt.Errorf("API component mutation target %s does not exist", mutation.Component)
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

func TestRenderTimelineFixturesDeterministicAndComplete(t *testing.T) {
	fixtures, err := schema.LoadTimelineFixtures()
	if err != nil {
		t.Fatal(err)
	}
	first, err := renderTimelineFixtures(fixtures)
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	second, err := renderTimelineFixtures(fixtures)
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("timeline TypeScript rendering is nondeterministic")
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(root, "typescript", "src", "fixtures", "timeline.ts"))
	if err != nil {
		t.Fatalf("read generated timeline fixture module: %v", err)
	}
	if !bytes.Equal(first, generated) {
		t.Fatal("generated timeline fixture module differs from an exact production render")
	}
}

func TestRenderTestcaseModelDeterministicAndComplete(t *testing.T) {
	first, err := renderTestcaseModel()
	if err != nil {
		t.Fatal(err)
	}
	second, err := renderTestcaseModel()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("testcase TypeScript rendering is nondeterministic")
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(root, "typescript", "src", "internal", "generated", "testcase.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, generated) {
		t.Fatal("generated testcase model differs from exact Go-derived production render")
	}
}
