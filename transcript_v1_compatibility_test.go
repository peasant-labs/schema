package schema_test

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_v1_compatibility.yaml
var transcriptV1CompatibilityYAML []byte

//go:embed testdata/transcript_v1_compatibility_manifest.yaml
var transcriptV1CompatibilityManifestYAML []byte

//go:embed testdata/transcript_segment_ref_v1_golden.json
var transcriptV1RefGolden []byte

//go:embed testdata/village_directory_v1_golden.json
var transcriptV1DirectoryGolden []byte

//go:embed testdata/local_api_document_v1_unchanged.json
var transcriptV1LocalDocument []byte

//go:embed generated/peasantlocal-api-0.38.0.json
var retiredTranscriptV1LocalDocument []byte

type v1CompatibilityInput struct {
	Kind string `yaml:"kind"`
	JSON string `yaml:"json"`
}

type v1CompatibilityExpected struct {
	Golden string `yaml:"golden"`
}

type v1CompatibilityFixture struct {
	LocalDocumentSHA256 string                                                         `yaml:"localDocumentSHA256"`
	RefMembers          []string                                                       `yaml:"refMembers"`
	TurnMembers         []string                                                       `yaml:"turnMembers"`
	Cases               []testcase.Case[v1CompatibilityInput, v1CompatibilityExpected] `yaml:"cases"`
}

func loadTranscriptV1Compatibility(t *testing.T) v1CompatibilityFixture {
	t.Helper()
	fixture := decodeStrictYAML[v1CompatibilityFixture](t, transcriptV1CompatibilityYAML, "v1 transcript compatibility fixture")
	corpus := testcase.Corpus[v1CompatibilityInput, v1CompatibilityExpected]{Cases: fixture.Cases}
	caseassert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "v1 transcript compatibility", transcriptV1CompatibilityManifestYAML, names)
	return fixture
}

func TestTranscriptV1WireGoldens(t *testing.T) {
	for _, c := range loadTranscriptV1Compatibility(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var value any
			var golden []byte
			switch c.Input.Kind {
			case "ref":
				value, golden = new(schema.TranscriptSegmentRef), transcriptV1RefGolden
				if c.Expected.Golden != "transcript_segment_ref_v1_golden.json" {
					t.Fatal("fixture names the wrong ref golden")
				}
			case "directory":
				value, golden = new(schema.TranscriptSegmentDirectory), transcriptV1DirectoryGolden
				if c.Expected.Golden != "village_directory_v1_golden.json" {
					t.Fatal("fixture names the wrong directory golden")
				}
			default:
				t.Fatalf("fixture names unknown wire kind %q", c.Input.Kind)
			}
			if err := json.Unmarshal([]byte(c.Input.JSON), value); err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			want := bytes.TrimSuffix(golden, []byte("\n"))
			if !bytes.Equal(got, want) {
				t.Fatalf("v1 wire changed:\ngot  %s\nwant %s", got, want)
			}
		})
	}
}

func TestTranscriptV1MemberSets(t *testing.T) {
	fixture := loadTranscriptV1Compatibility(t)
	assertMembers := func(typ reflect.Type, want []string) {
		t.Helper()
		got := make([]string, typ.NumField())
		for i := range got {
			got[i] = strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("%s JSON members = %v, want %v", typ.Name(), got, want)
		}
	}
	assertMembers(reflect.TypeFor[schema.TranscriptSegmentRef](), fixture.RefMembers)
	assertMembers(reflect.TypeFor[schema.TurnDetail](), fixture.TurnMembers)
}

func TestTranscriptV1LocalDocumentBytes(t *testing.T) {
	fixture := loadTranscriptV1Compatibility(t)
	if got := fmt.Sprintf("%x", sha256.Sum256(transcriptV1LocalDocument)); got != fixture.LocalDocumentSHA256 {
		t.Fatalf("stored v1 Local API document digest = %s, want %s", got, fixture.LocalDocumentSHA256)
	}
	if !bytes.Equal(retiredTranscriptV1LocalDocument, transcriptV1LocalDocument) {
		t.Fatal("the retired Local API document differs from its frozen v1 bytes")
	}
}
