package schema_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/segmented_read_generations.yaml
var segmentedReadGenerationsYAML []byte

//go:embed testdata/segmented_read_generations_manifest.yaml
var segmentedReadGenerationsManifestYAML []byte

//go:embed testdata/segmented_read_advertisement.yaml
var segmentedReadAdvertisementYAML []byte

//go:embed testdata/segmented_read_advertisement_manifest.yaml
var segmentedReadAdvertisementManifestYAML []byte

func TestSegmentedReadProtocolVersion(t *testing.T) {
	type input struct {
		Generation int `yaml:"generation"`
	}
	type expected struct {
		Version   string `yaml:"version"`
		Supported bool   `yaml:"supported"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedReadGenerationsYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read generation", segmentedReadGenerationsManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			version, ok := schema.SegmentedReadProtocolVersion(c.Input.Generation)
			if version != c.Expected.Version || ok != c.Expected.Supported {
				t.Fatalf("generation %d = (%q, %t), want (%q, %t)", c.Input.Generation, version, ok, c.Expected.Version, c.Expected.Supported)
			}
		})
	}
}

func TestSegmentedReadAdvertisementProducer(t *testing.T) {
	type input struct {
		Current string `yaml:"current"`
		Min     string `yaml:"min"`
	}
	type expected struct {
		Outcome string `yaml:"outcome"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedReadAdvertisementYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read advertisement", segmentedReadAdvertisementManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, err := json.Marshal(schema.SchemaVersionResponse{
				SegmentedReadVersion: c.Input.Current, MinSegmentedReadVersion: c.Input.Min,
			})
			if c.Expected.Outcome == "refused" {
				if err == nil {
					t.Fatal("producer emitted an incomplete version window")
				}
				return
			}
			if err != nil {
				t.Fatalf("marshal advertisement: %v", err)
			}
			var object map[string]any
			if err := json.Unmarshal(wire, &object); err != nil {
				t.Fatal(err)
			}
			current, hasCurrent := object["segmentedReadVersion"]
			minimum, hasMinimum := object["minSegmentedReadVersion"]
			if c.Expected.Outcome == "absent" {
				if hasCurrent || hasMinimum {
					t.Fatalf("empty pair was emitted: %s", wire)
				}
			} else if !hasCurrent || !hasMinimum || current != c.Input.Current || minimum != c.Input.Min {
				t.Fatalf("pair not emitted exactly: %s", wire)
			}
		})
	}
}
