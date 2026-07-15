package main

import (
	"bytes"
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

func TestGeneratedProjectHashLocationAssertionsMatchStrictFixture(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	locations, err := loadProjectHashLocations(root)
	if err != nil {
		t.Fatal(err)
	}
	want, err := renderProjectHashLocationAssertions(locations)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "typescript", "tests", "project-hash-locations.ts")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated ProjectHash location assertions drifted from strict fixture; regenerate %s with `go run ./cmd/schema-gen`", path)
	}
}

func TestProjectHashLocationExpectationMutationFailsClosed(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	locations, err := loadProjectHashLocations(root)
	if err != nil {
		t.Fatal(err)
	}
	mutated := locations
	mutated.Cases = append([]testcase.Case[projectHashLocation, string](nil), locations.Cases...)
	mutated.Cases[0].Expected = "string"
	if err := validateProjectHashLocations("mutated ProjectHash locations", mutated); err == nil {
		t.Fatal("ProjectHash location validator accepted a canonical location mutated to plain string")
	}
}

func TestProjectHashLocationLoaderMutationsFailStrictly(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "project_hash_locations.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	mutationData, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "project_hash_location_loader_mutations.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	mutations, err := testcase.LoadCorpus[string, bool](mutationData)
	if err != nil {
		t.Fatal(err)
	}
	if len(mutations.Cases) != 2 {
		t.Fatalf("ProjectHash location loader mutation fixture rows=%d, want exactly 2", len(mutations.Cases))
	}
	for _, testCase := range mutations.Cases {
		mutated := string(source)
		switch testCase.Input {
		case "add-unknown-field":
			mutated = strings.Replace(mutated, "    expected: ProjectHash\n", "    expected: ProjectHash\n    unexpected: true\n", 1)
		case "add-document":
			mutated += "\n---\ncases: []\n"
		default:
			t.Fatalf("ProjectHash location loader mutation fixture selected unknown kind %q", testCase.Input)
		}
		_, err := testcase.LoadCorpus[projectHashLocation, string]([]byte(mutated))
		accepted := err == nil
		if accepted != testCase.Expected {
			t.Fatalf("%s: accepted=%v, want %v", testCase.Name, accepted, testCase.Expected)
		}
	}
}

func TestCanonicalProjectHashBrandDoesNotRewriteSameSpellingProperties(t *testing.T) {
	path := filepath.Join(t.TempDir(), "types.ts")
	source := "/**\n * generated fixture\n */\n\n" +
		"export type components = { schemas: {\n" +
		"        ProjectHash: string;\n" +
		"        SameSpellingUnconstrained: { projectHash: string; targetProjectHash?: string };\n" +
		"} };\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	aliases := []schemaAlias{{Name: "ProjectHash", RawName: "ProjectHash", Surface: "types", Canonical: []byte(`{"type":"string"}`)}}
	if err := canonicalizeRootProjectHashIdentity(path, aliases); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "/**\n * generated fixture\n */\n\n" +
		"declare const projectHashBrand: unique symbol;\n" +
		"type CanonicalProjectHash = string & { readonly [projectHashBrand]: \"ProjectHash\" };\n\n" +
		"export type components = { schemas: {\n" +
		"        ProjectHash: CanonicalProjectHash;\n" +
		"        SameSpellingUnconstrained: { projectHash: string; targetProjectHash?: string };\n" +
		"} };\n"
	if string(got) != want {
		t.Fatalf("canonical ProjectHash branding changed an unrelated same-spelling property\ngot:\n%s\nwant:\n%s", got, want)
	}
}
