package main

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/mutation_seeds_validation.yaml
var mutationSeedsValidationCases []byte

//go:embed testdata/mutation_seeds_validation_manifest.yaml
var mutationSeedsValidationManifest []byte

type mutationSeedsValidationCase struct {
	Name     string        `yaml:"name"`
	Input    mutationSeeds `yaml:"input"`
	Expected struct {
		ErrorContains []string `yaml:"errorContains"`
	} `yaml:"expected"`
	Classification string `yaml:"classification"`
	Provenance     struct {
		Source string `yaml:"source"`
		Ref    string `yaml:"ref"`
	} `yaml:"provenance"`
	Mutation struct {
		Description string `yaml:"description"`
	} `yaml:"mutation"`
}

func loadMutationSeedsValidationFixtures(t *testing.T) ([]mutationSeedsValidationCase, []string) {
	t.Helper()
	var corpus struct {
		Cases []mutationSeedsValidationCase `yaml:"cases"`
	}
	dec := yaml.NewDecoder(bytes.NewReader(mutationSeedsValidationCases))
	dec.KnownFields(true)
	if err := dec.Decode(&corpus); err != nil {
		t.Fatalf("decode validation fixture: %v", err)
	}
	var manifest struct {
		RequiredNames []string `yaml:"requiredNames"`
	}
	manifestDec := yaml.NewDecoder(bytes.NewReader(mutationSeedsValidationManifest))
	manifestDec.KnownFields(true)
	if err := manifestDec.Decode(&manifest); err != nil {
		t.Fatalf("decode validation manifest: %v", err)
	}
	return corpus.Cases, manifest.RequiredNames
}

func TestLoadMutationsRejectsDuplicates(t *testing.T) {
	cases, requiredNames := loadMutationSeedsValidationFixtures(t)
	names := make([]string, len(cases))
	for i, c := range cases {
		names[i] = c.Name
	}
	if !slices.Equal(names, requiredNames) {
		t.Fatalf("validation cases %v do not match manifest %v", names, requiredNames)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if len(c.Expected.ErrorContains) == 0 {
				t.Fatal("validation case must name at least one expected error substring")
			}
			dir := t.TempDir()
			seedsPath := filepath.Join(dir, "mutation_seeds.yaml")
			manifestPath := filepath.Join(dir, "mutation_seeds_manifest.yaml")
			seedsBytes, err := yaml.Marshal(c.Input)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(seedsPath, seedsBytes, 0o644); err != nil {
				t.Fatal(err)
			}
			seedNames := make([]string, len(c.Input.Seeds))
			for i, s := range c.Input.Seeds {
				seedNames[i] = s.Name
			}
			mBytes, err := yaml.Marshal(struct {
				RequiredNames []string `yaml:"requiredNames"`
			}{RequiredNames: seedNames})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(manifestPath, mBytes, 0o644); err != nil {
				t.Fatal(err)
			}
			_, err = loadMutations(seedsPath)
			if err == nil {
				t.Fatalf("loadMutations(%s) succeeded, want error containing %v", seedsPath, c.Expected.ErrorContains)
			}
			for _, substr := range c.Expected.ErrorContains {
				if !strings.Contains(err.Error(), substr) {
					t.Fatalf("loadMutations error = %q, want substring %q", err.Error(), substr)
				}
			}
		})
	}
}
