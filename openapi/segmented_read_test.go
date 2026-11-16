package openapi_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/route"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/segmented_read_spec.yaml
var segmentedReadSpecYAML []byte

//go:embed testdata/segmented_read_spec_manifest.yaml
var segmentedReadSpecManifestYAML []byte

//go:embed testdata/segmented_read_rename_golden.yaml
var segmentedReadRenameGoldenYAML []byte

//go:embed testdata/segmented_read_rename_golden_manifest.yaml
var segmentedReadRenameGoldenManifestYAML []byte

type segmentedReadSpecInput struct {
	Version string `yaml:"version"`
}

type segmentedReadSpecExpected struct {
	Roots         []string `yaml:"roots"`
	Absent        []string `yaml:"absent"`
	ErrorContains string   `yaml:"errorContains"`
}

func TestSegmentedReadSpec(t *testing.T) {
	corpus, err := testcase.LoadCorpus[segmentedReadSpecInput, segmentedReadSpecExpected](segmentedReadSpecYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read spec", segmentedReadSpecManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			spec, err := openapi.BuildSegmentedReadSpec(c.Input.Version)
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("error = %v, want %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if spec.Info.Version != c.Input.Version || spec.Paths != nil {
				t.Fatalf("unexpected version or paths: version=%q paths=%v", spec.Info.Version, spec.Paths)
			}
			for _, name := range c.Expected.Roots {
				if _, ok := spec.Components.Schemas[name]; !ok {
					t.Errorf("required component %q missing", name)
				}
			}
			for _, name := range c.Expected.Absent {
				if _, ok := spec.Components.Schemas[name]; ok {
					t.Errorf("other protocol's component %q present", name)
				}
			}
		})
	}
}

func TestSegmentedReadSpec_GenerationTableMatches(t *testing.T) {
	spec, err := openapi.BuildSegmentedReadSpec(schema.SegmentedReadAPIVersion)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Info.Description == nil {
		t.Fatal("generation mapping and host routes missing from description")
	}
	for _, generation := range schema.SegmentedReadGenerations() {
		version, ok := schema.SegmentedReadProtocolVersion(generation)
		if !ok || !strings.Contains(*spec.Info.Description, fmt.Sprintf("v=%d -> %s", generation, version)) {
			t.Errorf("generation %d mapping missing or inconsistent", generation)
		}
	}
	for _, member := range []route.Route{route.VillageGetTranscriptSegments, route.VillageGetTranscriptSegment, route.VillageGetTranscriptWindow} {
		if !strings.Contains(*spec.Info.Description, fmt.Sprintf("%s %s (%s)", member.Method(), member.Pattern(), member.OperationID())) {
			t.Errorf("host route %s missing", member.OperationID())
		}
	}
}

// TestSegmentedReadRenamePreservesProtocol compares the entire document to
// the pre-rename capture after its component names and references were
// normalized. The later required protocol echo is separately tested and is
// removed here so this test continues to isolate the mechanical rename.
func TestSegmentedReadRenamePreservesProtocol(t *testing.T) {
	type expected struct {
		SHA256 string `yaml:"sha256"`
	}
	corpus, err := testcase.LoadCorpus[segmentedReadSpecInput, expected](segmentedReadRenameGoldenYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read rename", segmentedReadRenameGoldenManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			spec, err := openapi.BuildSegmentedReadSpec(c.Input.Version)
			if err != nil {
				t.Fatal(err)
			}
			wire, err := spec.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]interface{}
			if err := json.Unmarshal(wire, &doc); err != nil {
				t.Fatal(err)
			}
			context := doc["components"].(map[string]interface{})["schemas"].(map[string]interface{})["TranscriptReadContext"].(map[string]interface{})
			delete(context["properties"].(map[string]interface{}), "protocolVersion")
			if required, ok := context["required"].([]interface{}); ok {
				filtered := required[:0]
				for _, field := range required {
					if field != "protocolVersion" {
						filtered = append(filtered, field)
					}
				}
				context["required"] = filtered
			}
			canonical, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(canonical)); got != c.Expected.SHA256 {
				t.Fatalf("rename changed more than component names or references: canonical document sha256 %s, want %s", got, c.Expected.SHA256)
			}
		})
	}
}
