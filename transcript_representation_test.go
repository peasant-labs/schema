package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_representation.yaml
var transcriptRepresentationYAML []byte

//go:embed testdata/transcript_representation_manifest.yaml
var transcriptRepresentationManifestYAML []byte

type representationInput struct {
	Left       string `yaml:"left"`
	Right      string `yaml:"right"`
	Standalone bool   `yaml:"standalone"`
	Guard      string `yaml:"guard"`
}

type representationExpected struct {
	Golden   string `yaml:"golden"`
	Relation string `yaml:"relation"`
}

type representationFixtures struct {
	Baseline string                                                       `yaml:"baseline"`
	Fields   []string                                                     `yaml:"fields"`
	Cases    []testcase.Case[representationInput, representationExpected] `yaml:"cases"`
}

func TestTranscriptRepresentation(t *testing.T) {
	f := decodeStrictYAML[representationFixtures](t, transcriptRepresentationYAML, "transcript representation")
	corpus := testcase.Corpus[representationInput, representationExpected]{Cases: f.Cases}
	caseassert.RequireValid(t, corpus)
	names := make([]string, 0, len(f.Cases))
	for _, c := range f.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "transcript representation", transcriptRepresentationManifestYAML, names)
	makeInput := func(t *testing.T, patch string, standalone bool) schema.TranscriptRepresentationInputs {
		t.Helper()
		var input schema.TranscriptRepresentationInputs
		if !standalone {
			if err := json.Unmarshal([]byte(f.Baseline), &input); err != nil {
				t.Fatal(err)
			}
		}
		if patch != "" {
			if err := json.Unmarshal([]byte(patch), &input); err != nil {
				t.Fatal(err)
			}
		}
		return input
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			switch c.Input.Guard {
			case "fields":
				typ := reflect.TypeFor[schema.TranscriptRepresentationInputs]()
				fields := make([]string, typ.NumField())
				for i := range fields {
					fields[i] = typ.Field(i).Name
				}
				if !slices.Equal(fields, f.Fields) {
					t.Fatalf("representation fields = %v, want %v", fields, f.Fields)
				}
				return
			case "projection":
				want := []schema.TranscriptViewerProjection{schema.TranscriptViewerProjectionAnonymous, schema.TranscriptViewerProjectionAuthenticated}
				if !slices.Equal(schema.AllTranscriptViewerProjections, want) || schema.TranscriptViewerProjection("other").IsValid() {
					t.Fatalf("viewer projections = %v, want %v", schema.AllTranscriptViewerProjections, want)
				}
				return
			case "":
			default:
				t.Fatalf("unknown representation guard %q", c.Input.Guard)
			}
			left := schema.TranscriptRepresentation(makeInput(t, c.Input.Left, c.Input.Standalone))
			if c.Expected.Golden != "" && left != c.Expected.Golden {
				t.Fatalf("representation = %s, want %s", left, c.Expected.Golden)
			}
			if c.Expected.Relation != "" {
				right := schema.TranscriptRepresentation(makeInput(t, c.Input.Right, c.Input.Standalone))
				equal := left == right
				if equal != (c.Expected.Relation == "equal") {
					t.Fatalf("representation relation = equal:%t, want %s", equal, c.Expected.Relation)
				}
			}
		})
	}
}
