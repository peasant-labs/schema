package openapi_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
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

//go:embed testdata/segmented_read_discovery.yaml
var segmentedReadDiscoveryYAML []byte

//go:embed testdata/segmented_read_discovery_manifest.yaml
var segmentedReadDiscoveryManifestYAML []byte

type segmentedReadSpecInput struct {
	Version string `yaml:"version"`
}

type segmentedReadSpecExpected struct {
	Roots                      []string `yaml:"roots"`
	Absent                     []string `yaml:"absent"`
	ErrorContains              string   `yaml:"errorContains"`
	ProtocolVersionDescription string   `yaml:"protocolVersionDescription"`
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
			wire, err := spec.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]interface{}
			if err := json.Unmarshal(wire, &doc); err != nil {
				t.Fatal(err)
			}
			components := doc["components"].(map[string]interface{})["schemas"].(map[string]interface{})
			if c.Expected.ProtocolVersionDescription != "" {
				context, ok := components["TranscriptReadContext"].(map[string]interface{})
				if !ok {
					t.Fatal("positioned read context component is missing")
				}
				properties, ok := context["properties"].(map[string]interface{})
				if !ok {
					t.Fatal("positioned read context properties are missing")
				}
				echo, ok := properties["protocolVersion"].(map[string]interface{})
				if !ok {
					t.Fatal("positioned read protocol echo is missing")
				}
				description, _ := echo["description"].(string)
				if !strings.Contains(description, c.Expected.ProtocolVersionDescription) {
					t.Errorf("protocol echo description %q does not contain %q", description, c.Expected.ProtocolVersionDescription)
				}
			}
			var visit func(interface{})
			visit = func(value interface{}) {
				switch node := value.(type) {
				case map[string]interface{}:
					for key, child := range node {
						if key == "$ref" {
							ref, ok := child.(string)
							const prefix = "#/components/schemas/"
							if !ok || !strings.HasPrefix(ref, prefix) {
								t.Errorf("non-local component reference %v", child)
							} else if _, ok := components[strings.TrimPrefix(ref, prefix)]; !ok {
								t.Errorf("component reference %q is unresolved", ref)
							}
						} else {
							visit(child)
						}
					}
				case []interface{}:
					for _, child := range node {
						visit(child)
					}
				}
			}
			visit(components)
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

func TestSegmentedReadHostGenerationParametersMatch(t *testing.T) {
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	generations := schema.SegmentedReadGenerations()
	want := make([]interface{}, len(generations))
	for i, generation := range generations {
		want[i] = generation
	}
	for _, member := range []route.Route{route.VillageGetTranscriptSegments, route.VillageGetTranscriptSegment, route.VillageGetTranscriptWindow} {
		t.Run(member.OperationID(), func(t *testing.T) {
			item, ok := spec.Paths.MapOfPathItemValues[member.Pattern()]
			if !ok || item.Get == nil || item.Get.Description == nil {
				t.Fatal("host read or its description is missing")
			}
			if _, ok := item.Get.Responses.MapOfResponseOrReferenceValues["400"]; !ok {
				t.Fatal("the route does not declare a bad-generation refusal")
			}
			var found bool
			for _, ref := range item.Get.Parameters {
				param := ref.Parameter
				if param == nil || param.Name != "v" || string(param.In) != "query" {
					continue
				}
				found = true
				if !reflect.DeepEqual(param.Schema["enum"], want) || param.Description == nil {
					t.Fatalf("generation parameter = %#v, want enum %v and description", param, want)
				}
				for _, generation := range generations {
					version, ok := schema.SegmentedReadProtocolVersion(generation)
					mapping := fmt.Sprintf("v=%d -> %s", generation, version)
					if !ok || !strings.Contains(*param.Description, mapping) || !strings.Contains(*item.Get.Description, mapping) {
						t.Errorf("generation mapping %q is missing from parameter or operation", mapping)
					}
				}
			}
			if !found {
				t.Fatal("the host route has no generation selector")
			}
		})
	}
}

func TestSegmentedReadDiscoveryOperation(t *testing.T) {
	type input struct {
		Path string `yaml:"path"`
	}
	type expected struct {
		Contains []string `yaml:"contains"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedReadDiscoveryYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read discovery operation", segmentedReadDiscoveryManifestYAML, names)
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			item, ok := spec.Paths.MapOfPathItemValues[c.Input.Path]
			if !ok || item.Get == nil || item.Get.Description == nil {
				t.Fatalf("GET %s or its description is missing", c.Input.Path)
			}
			for _, text := range c.Expected.Contains {
				if !strings.Contains(*item.Get.Description, text) {
					t.Errorf("GET %s description does not contain %q", c.Input.Path, text)
				}
			}
		})
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
