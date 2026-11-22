package recordedindex_test

import (
	"bytes"
	_ "embed"
	"errors"
	"strconv"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/architecture.yaml
var architectureYAML []byte

//go:embed testdata/architecture_manifest.yaml
var architectureManifestYAML []byte

type architectureExpected struct {
	Refusal32 schema.TranscriptCaptureRefusal     `yaml:"refusal32"`
	Refusal64 schema.TranscriptCaptureRefusal     `yaml:"refusal64"`
	State32   schema.TranscriptRecordedIndexState `yaml:"state32"`
	State64   schema.TranscriptRecordedIndexState `yaml:"state64"`
	Index32   *int64                              `yaml:"index32"`
	Index64   *int64                              `yaml:"index64"`
}

type architectureCorpus = testcase.Corpus[string, architectureExpected]

// The served arm reads a recorded index from one served turn; the captured arm
// captures a session detail first, where the typed decode uses the native int.
type architectureFixture struct {
	ServedTurn   architectureCorpus `yaml:"servedTurn"`
	CapturedTurn architectureCorpus `yaml:"capturedTurn"`
}

func loadArchitectureFixture(t *testing.T) architectureFixture {
	t.Helper()
	var f architectureFixture
	decoder := yaml.NewDecoder(bytes.NewReader(architectureYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, f.ServedTurn)
	caseassert.RequireValid(t, f.CapturedTurn)
	names := make([]string, 0, len(f.ServedTurn.Cases)+len(f.CapturedTurn.Cases))
	for _, c := range append(append([]testcase.Case[string, architectureExpected]{}, f.ServedTurn.Cases...), f.CapturedTurn.Cases...) {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "recorded index architecture", architectureManifestYAML, names)
	return f
}

func (e architectureExpected) native() (schema.TranscriptCaptureRefusal, schema.TranscriptRecordedIndexState, *int64) {
	if strconv.IntSize == 32 {
		return e.Refusal32, e.State32, e.Index32
	}
	return e.Refusal64, e.State64, e.Index64
}

func requireIndex(t *testing.T, state schema.TranscriptRecordedIndexState, index *int, wantState schema.TranscriptRecordedIndexState, wantIndex *int64) {
	t.Helper()
	if state != wantState || (index == nil) != (wantIndex == nil) ||
		(index != nil && int64(*index) != *wantIndex) {
		t.Fatalf("int%d: state=%s index=%v, want state=%s index=%v", strconv.IntSize, state, index, wantState, wantIndex)
	}
}

func TestRecordedIndexNativeArchitecture(t *testing.T) {
	f := loadArchitectureFixture(t)
	for _, c := range f.ServedTurn.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turn, err := schema.NewTranscriptServedTurn([]byte(c.Input))
			if err != nil {
				t.Fatal(err)
			}
			state, index := schema.DecodeTranscriptRecordedIndex(turn)
			_, wantState, wantIndex := c.Expected.native()
			requireIndex(t, state, index, wantState, wantIndex)
		})
	}
	for _, c := range f.CapturedTurn.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turns, err := schema.NewTranscriptCookTurns([]byte(c.Input))
			wantRefusal, wantState, wantIndex := c.Expected.native()
			if wantRefusal != "" {
				var refusal *schema.TranscriptCaptureError
				if !errors.As(err, &refusal) || refusal.Refusal != wantRefusal {
					t.Fatalf("int%d: capture error = %v, want refusal %s", strconv.IntSize, err, wantRefusal)
				}
				return
			}
			if err != nil || len(turns) != 1 {
				t.Fatalf("int%d: capture = %d turns, %v", strconv.IntSize, len(turns), err)
			}
			state, index := turns[0].RecordedIndex()
			requireIndex(t, state, index, wantState, wantIndex)
		})
	}
}
