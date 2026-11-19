package assert

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
)

//go:embed testdata/manifest_cases.yaml
var manifestCasesYAML []byte

//go:embed testdata/manifest_cases_manifest.yaml
var manifestCasesManifestYAML []byte

type manifestInput struct {
	Manifest  string   `yaml:"manifest"`
	CaseNames []string `yaml:"caseNames"`
}

type manifestExpected struct {
	ErrorContains string `yaml:"errorContains"`
}

func TestManifestNames(t *testing.T) {
	corpus, err := testcase.LoadCorpus[manifestInput, manifestExpected](manifestCasesYAML)
	if err != nil {
		t.Fatalf("load manifest validation cases: %v", err)
	}
	var names []string
	for _, row := range corpus.Cases {
		names = append(names, row.Name)
	}
	RequireManifestNames(t, "manifest validation cases", manifestCasesManifestYAML, names)
	for _, row := range corpus.Cases {
		t.Run(row.Name, func(t *testing.T) {
			err := checkManifestNames([]byte(row.Input.Manifest), row.Input.CaseNames)
			if row.Classification == testcase.MustPass {
				if err != nil {
					t.Fatalf("valid manifest and cases rejected: %v", err)
				}
				return
			}
			if row.Expected.ErrorContains == "" || err == nil || !strings.Contains(err.Error(), row.Expected.ErrorContains) {
				t.Fatalf("expected an error containing %q, got %v", row.Expected.ErrorContains, err)
			}
		})
	}
}
