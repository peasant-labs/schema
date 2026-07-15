package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
)

const projectHashFixtureRows = 9

func loadProjectHashFixture(t *testing.T, root string) testcase.Corpus[string, bool] {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "project_hash.yaml"))
	if err != nil {
		t.Fatalf("read ProjectHash fixture: %v", err)
	}
	fixture, err := testcase.LoadCorpus[string, bool](data)
	if err != nil {
		t.Fatalf("strictly decode typed ProjectHash fixture: %v", err)
	}
	if len(fixture.Cases) != projectHashFixtureRows {
		t.Fatalf("ProjectHash fixture rows=%d, want exactly %d", len(fixture.Cases), projectHashFixtureRows)
	}
	if err := fixture.CheckMin(projectHashFixtureRows); err != nil {
		t.Fatal(err)
	}
	inputs := make(map[string]string, len(fixture.Cases))
	for _, testCase := range fixture.Cases {
		if prior, duplicate := inputs[testCase.Input]; duplicate {
			t.Fatalf("ProjectHash fixture cases %q and %q repeat the same input", prior, testCase.Name)
		}
		inputs[testCase.Input] = testCase.Name
	}
	return fixture
}

func TestProjectHashTypeScriptFixtureMatchesGoContract(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	fixture := loadProjectHashFixture(t, root)
	for _, testCase := range fixture.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			value, err := schema.NewProjectHash(testCase.Input)
			accepted := err == nil
			if accepted != testCase.Expected {
				t.Fatalf("NewProjectHash accepted=%v, want %v (error=%v)", accepted, testCase.Expected, err)
			}
			if accepted && value.String() != testCase.Input {
				t.Fatalf("ProjectHash.String()=%q, want exact input %q", value.String(), testCase.Input)
			}
		})
	}
}

func TestProjectHashFixtureMutationIsRejected(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	fixture := loadProjectHashFixture(t, root)
	valid := fixture.Cases[0].Input
	mutated := strings.ToUpper(valid)
	if mutated == valid {
		t.Fatal("ProjectHash mutation did not change the valid fixture")
	}
	if _, err := schema.NewProjectHash(mutated); err == nil {
		t.Fatalf("NewProjectHash accepted uppercase mutation %q", mutated)
	}
}

func TestGeneratedTypeScriptProjectHashIdentity(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	read := func(relative string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		return string(data)
	}

	rawTypes := read("typescript/src/internal/generated/types.ts")
	for _, required := range []string{
		`declare const projectHashBrand: unique symbol;`,
		`type CanonicalProjectHash = string & { readonly [projectHashBrand]: "ProjectHash" };`,
		`ProjectHash: CanonicalProjectHash;`,
		`projectHash?: (null | components["schemas"]["ProjectHash"]) | null;`,
		`targetProjectHash?: (null | components["schemas"]["ProjectHash"]) | null;`,
	} {
		if !strings.Contains(rawTypes, required) {
			t.Errorf("generated root Types catalog is missing branded identity fragment %q", required)
		}
	}
	if projectHashPropertyPattern.MatchString(rawTypes) {
		t.Error("generated root Types catalog retains a plain projectHash property")
	}

	rootFacade := read("typescript/src/index.ts")
	for _, required := range []string{
		`export type ProjectHash = TypesComponents["schemas"]["ProjectHash"];`,
		`export function isProjectHash(value: unknown): value is ProjectHash {`,
		`export function validateProjectHash(value: unknown): asserts value is ProjectHash {`,
		`export function newProjectHash(raw: string): ProjectHash {`,
	} {
		if !strings.Contains(rootFacade, required) {
			t.Errorf("generated package root is missing ProjectHash API fragment %q", required)
		}
	}

	for _, relative := range []string{
		"typescript/src/internal/generated/local-api.ts",
		"typescript/src/internal/generated/village-api.ts",
	} {
		source := read(relative)
		if projectHashPropertyPattern.MatchString(source) {
			t.Errorf("%s retains a plain projectHash operation field", relative)
		}
		if !strings.Contains(source, "Schema.ProjectHash") {
			t.Errorf("%s has no canonical root ProjectHash reference", relative)
		}
	}
}
