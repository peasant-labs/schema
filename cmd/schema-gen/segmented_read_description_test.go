package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/segmented_read_version_description.yaml
var segmentedReadVersionDescriptionYAML []byte

//go:embed testdata/segmented_read_version_description_manifest.yaml
var segmentedReadVersionDescriptionManifestYAML []byte

func TestGeneratedSegmentedReadVersionDescriptions(t *testing.T) {
	type input struct {
		Role   string `yaml:"role"`
		Format string `yaml:"format"`
	}
	type expected struct {
		EchoDescription   bool `yaml:"echoDescription"`
		FrozenWithoutEcho bool `yaml:"frozenWithoutEcho"`
	}
	corpus, err := testcase.LoadCorpus[input, expected](segmentedReadVersionDescriptionYAML)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "segmented read generated version descriptions", segmentedReadVersionDescriptionManifestYAML, names)

	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	frozen := make(map[string]retiredSpec)
	for _, spec := range loadServedFrozenSpecRegistry(t, root) {
		frozen[spec.Name] = spec
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var version string
			switch c.Input.Role {
			case "current":
				version = schema.SegmentedReadAPIVersion
			case "minimum":
				version = schema.SegmentedReadMinVersion
			default:
				t.Fatalf("unknown protocol role %q", c.Input.Role)
			}
			if c.Input.Format != "json" && c.Input.Format != "yaml" {
				t.Fatalf("unknown artifact format %q", c.Input.Format)
			}
			if (c.Input.Role == "current") != c.Expected.EchoDescription ||
				(c.Input.Role == "minimum") != c.Expected.FrozenWithoutEcho {
				t.Fatal("fixture expectations do not match the protocol role")
			}
			stem := "segmented-read-" + version
			filename := stem + "." + c.Input.Format
			generated, ok := artifacts[filename]
			if !ok {
				t.Fatalf("generator did not emit %q", filename)
			}
			var doc map[string]any
			if err := yaml.Unmarshal(generated, &doc); err != nil {
				t.Fatalf("decode generated document %q: %v", filename, err)
			}
			if c.Expected.FrozenWithoutEcho {
				pin, ok := frozen[stem]
				if !ok {
					t.Fatalf("served frozen protocol %q is not pinned", stem)
				}
				want := pin.JSONSHA256
				if c.Input.Format == "yaml" {
					want = pin.YAMLSHA256
				}
				if got := fmt.Sprintf("%x", sha256.Sum256(generated)); got != want {
					t.Fatalf("%q sha256 = %s, want frozen hash %s", filename, got, want)
				}
				if hasProtocolVersionMember(doc) {
					t.Fatalf("%q unexpectedly declares a protocolVersion member", filename)
				}
			}
			if c.Expected.EchoDescription {
				components, ok := doc["components"].(map[string]any)
				if !ok {
					t.Fatal("protocol components are missing")
				}
				schemas, ok := components["schemas"].(map[string]any)
				if !ok {
					t.Fatal("protocol schemas are missing")
				}
				context, ok := schemas["TranscriptReadContext"].(map[string]any)
				if !ok {
					t.Fatal("positioned read context is missing")
				}
				properties, ok := context["properties"].(map[string]any)
				if !ok {
					t.Fatal("positioned read context properties are missing")
				}
				echo, ok := properties["protocolVersion"].(map[string]any)
				if !ok {
					t.Fatal("positioned read protocol echo is missing")
				}
				if echo["type"] != "string" {
					t.Errorf("positioned read protocol echo type = %v, want string", echo["type"])
				}
				description, _ := echo["description"].(string)
				if !strings.Contains(description, schema.SegmentedReadAPIVersion) {
					t.Fatalf("protocol echo description %q does not name current version %q", description, schema.SegmentedReadAPIVersion)
				}
				if _, ok := echo["const"]; ok {
					t.Error("protocol echo must not freeze its version as a schema const")
				}
				if _, ok := echo["enum"]; ok {
					t.Error("protocol echo must remain an open string, not a schema enum")
				}
			}
			committed, err := os.ReadFile(filepath.Join(root, "generated", filename))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(generated, committed) {
				t.Fatalf("generated %q differs from the committed document; run make schema", filename)
			}
		})
	}
}

func hasProtocolVersionMember(value any) bool {
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			if key == "protocolVersion" || hasProtocolVersionMember(child) {
				return true
			}
		}
	case []any:
		for _, child := range node {
			if hasProtocolVersionMember(child) {
				return true
			}
		}
	}
	return false
}
