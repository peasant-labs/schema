package schema_test

import (
	"bytes"
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/dayvidpham/bestiary"
	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/harness_lifecycle.yaml
var harnessLifecycleYAML []byte

//go:embed testdata/harness_lifecycle_manifest.yaml
var harnessLifecycleManifestYAML []byte

func TestHarnessInventoriesMatchCatalog(t *testing.T) {
	for _, enum := range loadClosedEnumFixture(t).Enums {
		if enum.Name != "Harness" {
			continue
		}
		var known, declared []string
		for _, member := range enum.Members {
			value, ok := member.Value.(string)
			if !ok {
				t.Fatalf("Harness member %s has non-string value %v", member.Name, member.Value)
			}
			known = append(known, value)
		}
		for _, value := range enum.AllValues {
			name, ok := value.(string)
			if !ok {
				t.Fatalf("AllHarnesses has non-string fixture value %v", value)
			}
			declared = append(declared, name)
		}
		assertUniqueFixtureNames(t, "known harness catalog", known)
		assertUniqueFixtureNames(t, "declared harness catalog", declared)
		var gotKnown, gotDeclared []string
		for _, harness := range schema.Harnesses() {
			if !harness.IsKnown() {
				t.Errorf("Harnesses returned unrecognised harness %q", harness)
			}
			gotKnown = append(gotKnown, string(harness))
		}
		for _, harness := range schema.AllHarnesses {
			gotDeclared = append(gotDeclared, string(harness))
		}
		if !reflect.DeepEqual(gotKnown, known) {
			t.Errorf("Harnesses() = %v, want %v", gotKnown, known)
		}
		if !reflect.DeepEqual(gotDeclared, declared) {
			t.Errorf("AllHarnesses = %v, want %v", gotDeclared, declared)
		}
		return
	}
	t.Fatal("Harness is missing from the enum catalog")
}

func TestAntigravityAliasesMatchBestiary(t *testing.T) {
	if schema.HarnessAntigravityCLI != bestiary.HarnessAntigravityCLI {
		t.Error("HarnessAntigravityCLI must re-export the bestiary value")
	}
	if schema.HarnessAntigravityIDE != bestiary.HarnessAntigravityIDE {
		t.Error("HarnessAntigravityIDE must re-export the bestiary value")
	}
}

func TestHarnessEnumAcrossAPIComponents(t *testing.T) {
	var expected []string
	for _, enum := range loadClosedEnumFixture(t).Enums {
		if enum.Name == "Harness" {
			for _, member := range enum.Members {
				expected = append(expected, mustJSONValue(t, enum.Name, member.Value))
			}
			break
		}
	}
	if len(expected) == 0 {
		t.Fatal("Harness is missing from the enum catalog")
	}
	check := func(name string, component map[string]any) {
		t.Helper()
		if component == nil {
			t.Errorf("%s spec has no Harness component", name)
			return
		}
		actual := schemaEnumJSONValues(t, name, component["enum"])
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%s spec Harness enum = %v, want %v", name, actual, expected)
		}
	}
	types, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	check("Types", types.Components.Schemas["Harness"])
	village, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("build Village API spec: %v", err)
	}
	check("Village", village.Components.Schemas["BestiaryHarness"])
	local, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("build Local API spec: %v", err)
	}
	check("Local", local.Components.Schemas["BestiaryHarness"])
}

func TestHarnessLifecycleDocumentation(t *testing.T) {
	type lifecycleInput struct {
		Symbol string `yaml:"symbol"`
	}
	type lifecycleExpected struct {
		Contains []string `yaml:"contains"`
		Excludes []string `yaml:"excludes"`
	}
	var fixture struct {
		Cases []testcase.Case[lifecycleInput, lifecycleExpected] `yaml:"cases"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(harnessLifecycleYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode harness lifecycle fixture: %v", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("harness lifecycle fixture must contain exactly one YAML document: %v", err)
	}
	if err := (testcase.Corpus[lifecycleInput, lifecycleExpected]{Cases: fixture.Cases}).Validate(); err != nil {
		t.Fatalf("validate harness lifecycle corpus: %v", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "types.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse types.go documentation: %v", err)
	}
	docs := make(map[string]string)
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			doc := value.Doc
			if doc == nil && len(gen.Specs) == 1 {
				doc = gen.Doc
			}
			if doc != nil {
				for _, name := range value.Names {
					docs[name.Name] = doc.Text()
				}
			}
		}
	}
	var names []string
	caseSymbols := make(map[string]bool)
	for _, row := range fixture.Cases {
		names = append(names, row.Name)
		caseSymbols[row.Input.Symbol] = true
		t.Run(row.Name, func(t *testing.T) {
			if row.Classification != testcase.MustPass || row.Input.Symbol == "" || len(row.Expected.Contains) == 0 {
				t.Fatal("lifecycle case must name a symbol and required documentation")
			}
			doc, ok := docs[row.Input.Symbol]
			if !ok {
				t.Fatalf("types.go has no documentation for %s", row.Input.Symbol)
			}
			for _, want := range row.Expected.Contains {
				if !strings.Contains(doc, want) {
					t.Errorf("%s documentation must contain %q; got %q", row.Input.Symbol, want, doc)
				}
			}
			for _, forbidden := range row.Expected.Excludes {
				if strings.Contains(doc, forbidden) {
					t.Errorf("%s documentation must not contain %q; got %q", row.Input.Symbol, forbidden, doc)
				}
			}
		})
	}
	assertUniqueFixtureNames(t, "lifecycle cases", names)
	requireManifestNames(t, "harness lifecycle", harnessLifecycleManifestYAML, names)

	for symbol, doc := range docs {
		if (!strings.HasPrefix(symbol, "Harness") && symbol != "AllHarnesses") || caseSymbols[symbol] {
			continue
		}
		for _, line := range strings.Split(doc, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Deprecated:") {
				t.Errorf("%s must not have a Deprecated line", symbol)
			}
		}
	}

	markers, paragraphs := 0, 0
	paragraphStart := true
	for _, line := range strings.Split(docs["HarnessAntigravity"], "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			paragraphStart = true
			continue
		}
		if strings.HasPrefix(line, "Deprecated:") {
			markers++
			if paragraphStart {
				paragraphs++
			}
		}
		paragraphStart = false
	}
	if markers != 1 || paragraphs != 1 {
		t.Errorf("HarnessAntigravity must have exactly one Deprecated paragraph; got %d lines in %d paragraphs", markers, paragraphs)
	}
}
