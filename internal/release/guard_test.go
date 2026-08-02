package release_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

type finalCheckInput struct {
	Final                  release.Version              `yaml:"final"`
	InitialFinal           release.Version              `yaml:"initialFinal"`
	PriorReleases          []release.Version            `yaml:"priorReleases"`
	RCs                    []release.RCStatus           `yaml:"rcs"`
	ProductTagScan         release.ProductTagScanStatus `yaml:"productTagScan"`
	InitialFinalCompletion release.CompletionStatus     `yaml:"initialFinalCompletion"`
}

type finalCheckExpected struct {
	ErrorContains string `yaml:"errorContains"`
}

//go:embed testdata/guard/final_cases.yaml
var finalCasesYAML []byte

func TestCheckFinal(t *testing.T) {
	corpus, err := testcase.LoadCorpus[finalCheckInput, finalCheckExpected](finalCasesYAML)
	if err != nil {
		t.Fatalf("load final guard corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 10)
	assert.RequireValid(t, corpus)

	for _, c := range corpus.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			err := release.CheckFinal(c.Input.Final, release.FinalEvidence{
				RCs: c.Input.RCs, ProductTagScan: c.Input.ProductTagScan,
				PriorReleases: c.Input.PriorReleases, InitialFinalCompletion: c.Input.InitialFinalCompletion,
			}, release.FinalPolicy{InitialFinal: c.Input.InitialFinal})
			if c.Classification == testcase.MustFail {
				if err == nil {
					t.Fatal("CheckFinal returned nil, want rejection")
				}
				if !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("CheckFinal error %q does not contain %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckFinal returned unexpected error: %v", err)
			}
		})
	}
}
