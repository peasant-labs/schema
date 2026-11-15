package schema

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_ref_chain.yaml
var refChainYAML []byte

//go:embed testdata/transcript_ref_chain_manifest.yaml
var refChainManifestYAML []byte

type refChainInput struct {
	Context      string `yaml:"context"`
	FirstOrdinal int    `yaml:"firstOrdinal"`
	Total        int    `yaml:"total"`
	Refs         []struct {
		Index int `yaml:"index"`
		First int `yaml:"first"`
		Last  int `yaml:"last"`
	} `yaml:"refs"`
}

type refChainExpected struct {
	ErrorContains string `yaml:"errorContains"`
}

func TestTranscriptRefChainFirstOrdinal(t *testing.T) {
	corpus, err := testcase.LoadCorpus[refChainInput, refChainExpected](refChainYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "transcript ref chain", refChainManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var context transcriptRefChainContext
			switch c.Input.Context {
			case "window":
				context = transcriptRefChainWindow
			case "directory":
				context = transcriptRefChainDirectory
			default:
				t.Fatalf("fixture names unknown ref-chain context %q", c.Input.Context)
			}
			refs := make([]TranscriptSegmentRefV2, len(c.Input.Refs))
			for i, ref := range c.Input.Refs {
				refs[i] = TranscriptSegmentRefV2{Index: ref.Index, SourceFirstPosition: ref.First, SourceLastPosition: ref.Last}
			}
			err := validateTranscriptRefChain(refs, c.Input.FirstOrdinal, c.Input.Total, "ref-chain test: ", context, false)
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("ref-chain error = %v, want %q", err, c.Expected.ErrorContains)
				}
			} else if err != nil {
				t.Fatalf("valid ref chain was refused: %v", err)
			}
		})
	}
}
