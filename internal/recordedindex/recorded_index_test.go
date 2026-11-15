package recordedindex_test

import (
	_ "embed"
	"strconv"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/architecture.yaml
var architectureYAML []byte

//go:embed testdata/architecture_manifest.yaml
var architectureManifestYAML []byte

type architectureExpected struct {
	State32 schema.TranscriptRecordedIndexState `yaml:"state32"`
	State64 schema.TranscriptRecordedIndexState `yaml:"state64"`
	Index32 *int64                              `yaml:"index32"`
	Index64 *int64                              `yaml:"index64"`
}

func TestRecordedIndexNativeArchitecture(t *testing.T) {
	corpus, err := testcase.LoadCorpus[string, architectureExpected](architectureYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "recorded index architecture", architectureManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turn, err := schema.NewTranscriptV2Turn([]byte(c.Input))
			if err != nil {
				t.Fatal(err)
			}
			state, index := schema.DecodeTranscriptRecordedIndex(turn)
			wantState, wantIndex := c.Expected.State64, c.Expected.Index64
			if strconv.IntSize == 32 {
				wantState, wantIndex = c.Expected.State32, c.Expected.Index32
			}
			if state != wantState || (index == nil) != (wantIndex == nil) ||
				(index != nil && int64(*index) != *wantIndex) {
				t.Fatalf("int%d: state=%s index=%v, want state=%s index=%v", strconv.IntSize, state, index, wantState, wantIndex)
			}
		})
	}
}
