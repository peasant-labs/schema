package contractgates

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/segmented_doc_stamp.yaml
var segmentedDocStampCases []byte

//go:embed testdata/segmented_doc_stamp_manifest.yaml
var segmentedDocStampManifest []byte

func TestSegmentedDocumentStampExemption(t *testing.T) {
	skipIfMissing(t, "bash")
	type input struct {
		APIDiff string `yaml:"apidiff"`
	}
	type expected struct {
		PayloadContains string `yaml:"payload_contains"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedDocStampCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented document stamp", segmentedDocStampManifest, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			exit, payload := runEvaluateAPIDiff(t, moduleRoot(t), c.Input.APIDiff)
			if exit != 0 {
				t.Fatalf("evaluate-apidiff exit = %d, payload %q", exit, payload)
			}
			if c.Classification == testcase.MustPass && strings.TrimSpace(payload) != "" {
				t.Fatalf("document stamp was not exempt: %q", payload)
			}
			if c.Classification == testcase.MustFail && !strings.Contains(payload, c.Expected.PayloadContains) {
				t.Fatalf("real API change missing from payload: %q", payload)
			}
		})
	}
}
