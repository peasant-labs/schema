package schema_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_download_strictness.yaml
var transcriptDownloadStrictnessYAML []byte

//go:embed testdata/transcript_download_strictness_manifest.yaml
var transcriptDownloadStrictnessManifestYAML []byte

type downloadStrictnessInput struct {
	Detail string `yaml:"detail"`
}

type downloadStrictnessExpected struct {
	Strict bool `yaml:"strict"`
}

func TestTranscriptContentStrictMemberNames(t *testing.T) {
	corpus := decodeStrictYAML[testcase.Corpus[downloadStrictnessInput, downloadStrictnessExpected]](t, transcriptDownloadStrictnessYAML, "transcript strictness member fixture")
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
		t.Run(c.Name, func(t *testing.T) {
			if got := schema.TranscriptContentStrict(json.RawMessage(c.Input.Detail)); got != c.Expected.Strict {
				t.Fatalf("strictness = %v, want %v", got, c.Expected.Strict)
			}
		})
	}
	caseassert.RequireManifestNames(t, "transcript strictness member", transcriptDownloadStrictnessManifestYAML, names)
}
