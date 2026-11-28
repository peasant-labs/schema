package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

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

func TestLoadMutationsRejectsDuplicates(t *testing.T) {
	raw, err := os.ReadFile("testdata/mutation_seeds_validation.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []mutationSeedsValidationCase `yaml:"cases"`
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&corpus); err != nil {
		t.Fatalf("decode validation fixture: %v", err)
	}
	manifestRaw, err := os.ReadFile("testdata/mutation_seeds_validation_manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		RequiredNames []string `yaml:"requiredNames"`
	}
	manifestDec := yaml.NewDecoder(bytes.NewReader(manifestRaw))
	manifestDec.KnownFields(true)
	if err := manifestDec.Decode(&manifest); err != nil {
		t.Fatalf("decode validation manifest: %v", err)
	}
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	if !slices.Equal(names, manifest.RequiredNames) {
		t.Fatalf("validation cases %v do not match manifest %v", names, manifest.RequiredNames)
	}
	for _, c := range corpus.Cases {
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
