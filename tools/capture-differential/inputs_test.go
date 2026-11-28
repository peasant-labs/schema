package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/mutation_seeds_validation.yaml
var mutationSeedsValidationCases []byte

//go:embed testdata/mutation_seeds_validation_manifest.yaml
var mutationSeedsValidationManifest []byte

type mutationSeedsValidationExpected struct {
	ErrorContains []string `yaml:"errorContains"`
}

func TestLoadMutationsRejectsDuplicates(t *testing.T) {
	corpus, err := testcase.LoadCorpus[mutationSeeds, mutationSeedsValidationExpected](mutationSeedsValidationCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireMin(t, corpus, 1)
	caseassert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "mutation seeds validation", mutationSeedsValidationManifest, names)
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
