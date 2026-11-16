package openapi_test

import (
	_ "embed"
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
