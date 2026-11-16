package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
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

//go:embed testdata/segmented_read_protocol_echo.yaml
var segmentedReadProtocolEchoYAML []byte

//go:embed testdata/segmented_read_protocol_echo_manifest.yaml
var segmentedReadProtocolEchoManifestYAML []byte

//go:embed testdata/segmented_read_spec_accessors.yaml
var segmentedReadSpecAccessorsYAML []byte

//go:embed testdata/segmented_read_spec_accessors_manifest.yaml
var segmentedReadSpecAccessorsManifestYAML []byte

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

func TestTranscriptReadContextProtocolVersion(t *testing.T) {
	type input struct {
		Version string `yaml:"version"`
	}
	type expected struct {
		ErrorContains string `yaml:"errorContains"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedReadProtocolEchoYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read protocol echo", segmentedReadProtocolEchoManifestYAML, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			id := schema.SessionID("ses_test")
			context := schema.TranscriptReadContext{
				ProtocolVersion: c.Input.Version,
				Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &id},
				Revision:        "1", Representation: "r1.test",
				Validation: schema.TranscriptReadServedLegacy, NativePresence: schema.TranscriptNativeAbsent,
			}
			err := context.Validate()
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("context error = %v, want %q", err, c.Expected.ErrorContains)
				}
			} else if err != nil {
				t.Fatalf("declared protocol refused: %v", err)
			}
		})
	}
}

func TestSegmentedReadSpecAccessors(t *testing.T) {
	type input struct {
		Role string `yaml:"role"`
	}
	type expected struct {
		Artifact string `yaml:"artifact"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedReadSpecAccessorsYAML)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read spec accessors", segmentedReadSpecAccessorsManifestYAML, names)
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var got []byte
			switch c.Input.Role {
			case "current":
				got = schema.SegmentedReadSpecJSON()
			case "minimum":
				got = schema.SegmentedReadMinSpecJSON()
			default:
				t.Fatalf("unknown accessor role %q", c.Input.Role)
			}
			want, ok := artifacts[c.Expected.Artifact]
			if !ok || !bytes.Equal(got, want) {
				t.Fatalf("embedded document does not match generated artifact %q", c.Expected.Artifact)
			}
		})
	}
}
