package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type publicExportFixture struct {
	Aliases   []publicAliasFixture    `yaml:"aliases"`
	Constants []publicConstantFixture `yaml:"constants"`
	Forbidden []string                `yaml:"forbidden"`
}

type publicAliasFixture struct {
	Name     string   `yaml:"name"`
	Target   string   `yaml:"target"`
	Surfaces []string `yaml:"surfaces"`
}

type publicConstantFixture struct {
	Name     string   `yaml:"name"`
	Value    string   `yaml:"value"`
	Surfaces []string `yaml:"surfaces"`
}

type publicCatalogFixture struct {
	ForbiddenComponents []string                 `yaml:"forbidden_components"`
	Types               []publicCatalogTypeEntry `yaml:"types"`
}

type publicCatalogTypeEntry struct {
	GoName      string `yaml:"go_name"`
	Disposition string `yaml:"disposition"`
	Component   string `yaml:"component"`
	Reason      string `yaml:"reason"`
}

type publicExportMutationKind string

const (
	publicExportRemove    publicExportMutationKind = "remove"
	publicExportAdd       publicExportMutationKind = "add"
	publicExportDuplicate publicExportMutationKind = "duplicate"
	publicExportRedirect  publicExportMutationKind = "redirect"
)

type publicExportMutationInput struct {
	Kind      publicExportMutationKind `yaml:"kind"`
	Namespace string                   `yaml:"namespace"`
	Name      string                   `yaml:"name"`
	Target    string                   `yaml:"target"`
}

var (
	schemaTypeExportPattern  = regexp.MustCompile(`^export type ([A-Za-z][A-Za-z0-9_]*) = (?:TypesComponents|components)\["schemas"\]\["([^"]+)"\];$`)
	runtimeTypeExportPattern = regexp.MustCompile(`^export type ([A-Za-z][A-Za-z0-9_]*) = \(typeof ([A-Za-z][A-Za-z0-9_]*)\)\[keyof typeof ([A-Za-z][A-Za-z0-9_]*)\];$`)
	aliasTypeExportPattern   = regexp.MustCompile(`^export type ([A-Za-z][A-Za-z0-9_]*) = ([A-Za-z][A-Za-z0-9_]*);$`)
	runtimeValuePattern      = regexp.MustCompile(`^export const ([A-Za-z][A-Za-z0-9_]*) = Object\.freeze\(\{$`)
	runtimeValuesPattern     = regexp.MustCompile(`^export const ([A-Za-z][A-Za-z0-9_]*) = Object\.freeze\(\[.* satisfies readonly ([A-Za-z][A-Za-z0-9_]*)\[\]\);$`)
	literalValuePattern      = regexp.MustCompile(`^export const ([A-Za-z][A-Za-z0-9_]*) = (.+) as const;$`)
	runtimeGuardPattern      = regexp.MustCompile(`^export function (is[A-Za-z][A-Za-z0-9_]*)\(value: unknown\): value is ([A-Za-z][A-Za-z0-9_]*) \{$`)
)

func TestGeneratedPublicExportsHaveExactIdentity(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	public, catalog, enums := loadPublicExportFixtures(t, root)

	rootSource := readPublicFacade(t, filepath.Join(root, "typescript", "src", "index.ts"))
	if err := validatePublicExports(rootSource, expectedPublicExports(t, "root", public, catalog, enums), public.Forbidden); err != nil {
		t.Fatalf("generated package root exports: %v", err)
	}

	typesSource := readPublicFacade(t, filepath.Join(root, "typescript", "src", "types.ts"))
	if err := validatePublicExports(typesSource, expectedPublicExports(t, "types", public, catalog, enums), public.Forbidden); err != nil {
		t.Fatalf("generated /types exports: %v", err)
	}
}

func TestGeneratedPublicExportIdentityGateMutationProof(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	public, catalog, enums := loadPublicExportFixtures(t, root)
	expected := expectedPublicExports(t, "root", public, catalog, enums)
	source := readPublicFacade(t, filepath.Join(root, "typescript", "src", "index.ts"))

	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "public_export_mutations.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	matrix, err := testcase.LoadCorpus[publicExportMutationInput, bool](data)
	if err != nil {
		t.Fatal(err)
	}
	if len(matrix.Cases) != 6 {
		t.Fatalf("public export mutation fixture has %d rows, want exactly 6", len(matrix.Cases))
	}
	if err := matrix.CheckMin(6); err != nil {
		t.Fatal(err)
	}
	for _, tc := range matrix.Cases {
		mutated := mutatePublicExports(t, source, tc.Input)
		accepted := validatePublicExports(mutated, expected, public.Forbidden) == nil
		if accepted != tc.Expected {
			t.Fatalf("%s: accepted=%v, want %v", tc.Name, accepted, tc.Expected)
		}
	}
}

func loadPublicExportFixtures(t *testing.T, root string) (publicExportFixture, publicCatalogFixture, enumCatalogFixture) {
	t.Helper()
	var public publicExportFixture
	decodeStrictYAMLFile(t, filepath.Join(root, "testdata", "typescript", "public_exports.yaml"), &public)
	var catalog publicCatalogFixture
	decodeStrictYAMLFile(t, filepath.Join(root, "openapi", "testdata", "typescript_catalog.yaml"), &catalog)
	var enums enumCatalogFixture
	decodeStrictYAMLFile(t, filepath.Join(root, "testdata", "typescript", "enums.yaml"), &enums)
	return public, catalog, enums
}

func decodeStrictYAMLFile(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func expectedPublicExports(t *testing.T, surface string, public publicExportFixture, catalog publicCatalogFixture, enums enumCatalogFixture) map[string]string {
	t.Helper()
	expected := map[string]string{}
	runtimeNames := map[string]struct{}{}
	for _, enum := range enums.Enums {
		runtimeNames[enum.Name] = struct{}{}
		addExpectedExport(t, expected, "value", enum.Name, "runtime-enum:"+enum.Name)
		addExpectedExport(t, expected, "type", enum.Name, "runtime-enum:"+enum.Name)
		addExpectedExport(t, expected, "value", enum.AllName, "runtime-values:"+enum.Name)
		addExpectedExport(t, expected, "function", "is"+enum.Name, "runtime-guard:"+enum.Name)
	}
	for _, entry := range catalog.Types {
		if entry.Disposition != "catalog" {
			continue
		}
		if _, runtime := runtimeNames[entry.Component]; runtime {
			continue
		}
		addExpectedExport(t, expected, "type", entry.GoName, "schema:"+entry.Component)
	}
	for _, alias := range public.Aliases {
		validatePublicFixtureSurfaces(t, alias.Name, alias.Surfaces)
		if includesSurface(alias.Surfaces, surface) {
			addExpectedExport(t, expected, "type", alias.Name, "alias:"+alias.Target)
		}
	}
	for _, constant := range public.Constants {
		validatePublicFixtureSurfaces(t, constant.Name, constant.Surfaces)
		if includesSurface(constant.Surfaces, surface) {
			addExpectedExport(t, expected, "value", constant.Name, "literal:"+constant.Value)
		}
	}
	return expected
}

func validatePublicFixtureSurfaces(t *testing.T, name string, surfaces []string) {
	t.Helper()
	if len(surfaces) == 0 {
		t.Fatalf("public export fixture %s has no surfaces", name)
	}
	seen := map[string]struct{}{}
	for _, surface := range surfaces {
		switch surface {
		case "root", "types":
		default:
			t.Fatalf("public export fixture %s has unknown surface %q", name, surface)
		}
		if _, duplicate := seen[surface]; duplicate {
			t.Fatalf("public export fixture %s repeats surface %q", name, surface)
		}
		seen[surface] = struct{}{}
	}
}

func includesSurface(surfaces []string, wanted string) bool {
	for _, surface := range surfaces {
		if surface == wanted {
			return true
		}
	}
	return false
}

func addExpectedExport(t *testing.T, expected map[string]string, namespace, name, target string) {
	t.Helper()
	key := namespace + ":" + name
	if prior, duplicate := expected[key]; duplicate {
		t.Fatalf("independent fixtures define public export %s twice (%s and %s)", key, prior, target)
	}
	expected[key] = target
}

func readPublicFacade(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func validatePublicExports(source []byte, expected map[string]string, forbidden []string) error {
	actual, err := parsePublicExports(source)
	if err != nil {
		return err
	}
	for _, name := range forbidden {
		if _, leaked := actual["type:"+name]; leaked {
			return fmt.Errorf("forbidden historical public name %s is exported in the type namespace", name)
		}
		if _, leaked := actual["value:"+name]; leaked {
			return fmt.Errorf("forbidden historical public name %s is exported in the value namespace", name)
		}
		if _, leaked := actual["function:"+name]; leaked {
			return fmt.Errorf("forbidden historical public name %s is exported in the function namespace", name)
		}
	}
	for _, key := range sortedExportKeys(expected) {
		want := expected[key]
		got, ok := actual[key]
		if !ok {
			return fmt.Errorf("missing public export %s targeting %s", key, want)
		}
		if got != want {
			return fmt.Errorf("public export %s targets %s, want %s", key, got, want)
		}
	}
	for _, key := range sortedExportKeys(actual) {
		if _, ok := expected[key]; !ok {
			return fmt.Errorf("unexpected public export %s targeting %s", key, actual[key])
		}
	}
	return nil
}

func parsePublicExports(source []byte) (map[string]string, error) {
	exports := map[string]string{}
	for _, line := range strings.Split(string(source), "\n") {
		if !strings.HasPrefix(line, "export ") {
			continue
		}
		var namespace, name, target string
		switch {
		case schemaTypeExportPattern.MatchString(line):
			matches := schemaTypeExportPattern.FindStringSubmatch(line)
			namespace, name, target = "type", matches[1], "schema:"+matches[2]
		case runtimeTypeExportPattern.MatchString(line):
			matches := runtimeTypeExportPattern.FindStringSubmatch(line)
			if matches[1] != matches[2] || matches[1] != matches[3] {
				return nil, fmt.Errorf("runtime type export has inconsistent targets: %s", line)
			}
			namespace, name, target = "type", matches[1], "runtime-enum:"+matches[1]
		case aliasTypeExportPattern.MatchString(line):
			matches := aliasTypeExportPattern.FindStringSubmatch(line)
			namespace, name, target = "type", matches[1], "alias:"+matches[2]
		case runtimeValuePattern.MatchString(line):
			matches := runtimeValuePattern.FindStringSubmatch(line)
			namespace, name, target = "value", matches[1], "runtime-enum:"+matches[1]
		case runtimeValuesPattern.MatchString(line):
			matches := runtimeValuesPattern.FindStringSubmatch(line)
			namespace, name, target = "value", matches[1], "runtime-values:"+matches[2]
		case literalValuePattern.MatchString(line):
			matches := literalValuePattern.FindStringSubmatch(line)
			literal, err := normalizePublicLiteral(matches[2])
			if err != nil {
				return nil, fmt.Errorf("parse public constant %s: %w", matches[1], err)
			}
			namespace, name, target = "value", matches[1], "literal:"+literal
		case runtimeGuardPattern.MatchString(line):
			matches := runtimeGuardPattern.FindStringSubmatch(line)
			namespace, name, target = "function", matches[1], "runtime-guard:"+matches[2]
		default:
			return nil, fmt.Errorf("unrecognized generated public export: %s", line)
		}
		key := namespace + ":" + name
		if prior, duplicate := exports[key]; duplicate {
			return nil, fmt.Errorf("duplicate generated public export %s (%s and %s)", key, prior, target)
		}
		exports[key] = target
	}
	return exports, nil
}

func normalizePublicLiteral(source string) (string, error) {
	if unquoted, err := strconv.Unquote(source); err == nil {
		return unquoted, nil
	}
	if _, err := strconv.Atoi(source); err == nil {
		return source, nil
	}
	return "", fmt.Errorf("unsupported literal %q", source)
}

func sortedExportKeys(exports map[string]string) []string {
	keys := make([]string, 0, len(exports))
	for key := range exports {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mutatePublicExports(t *testing.T, source []byte, mutation publicExportMutationInput) []byte {
	t.Helper()
	line := ""
	if mutation.Kind != publicExportAdd {
		line = findPublicExportLine(t, source, mutation.Namespace, mutation.Name)
	}
	mutated := string(source)
	switch mutation.Kind {
	case publicExportRemove:
		mutated = strings.Replace(mutated, line+"\n", "", 1)
	case publicExportDuplicate:
		mutated += line + "\n"
	case publicExportAdd:
		mutated += renderMutatedExportLine(t, mutation) + "\n"
	case publicExportRedirect:
		mutated = strings.Replace(mutated, line, renderMutatedExportLine(t, mutation), 1)
	default:
		t.Fatalf("public export mutation fixture selected unknown kind %q", mutation.Kind)
	}
	return []byte(mutated)
}

func findPublicExportLine(t *testing.T, source []byte, namespace, name string) string {
	t.Helper()
	prefix := "export " + namespace + " " + name
	if namespace == "value" {
		prefix = "export const " + name
	}
	if namespace == "function" {
		prefix = "export function " + name
	}
	for _, line := range strings.Split(string(source), "\n") {
		if strings.HasPrefix(line, prefix+" ") || strings.HasPrefix(line, prefix+"(") {
			return line
		}
	}
	t.Fatalf("mutation fixture could not find %s export %s", namespace, name)
	return ""
}

func renderMutatedExportLine(t *testing.T, mutation publicExportMutationInput) string {
	t.Helper()
	if strings.TrimSpace(mutation.Target) == "" {
		t.Fatalf("%s mutation for %s requires a target", mutation.Kind, mutation.Name)
	}
	switch mutation.Namespace {
	case "type":
		return fmt.Sprintf("export type %s = TypesComponents[\"schemas\"][%q];", mutation.Name, mutation.Target)
	case "value":
		return fmt.Sprintf("export const %s = %s as const;", mutation.Name, mutation.Target)
	default:
		t.Fatalf("%s mutation does not support namespace %q", mutation.Kind, mutation.Namespace)
	}
	return ""
}
