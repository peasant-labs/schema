package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

func TestRequiredContentCapabilitiesSessionGraph(t *testing.T) {
	fixtures, err := schema.LoadContentCapabilitySessionGraphFixtures()
	if err != nil {
		t.Fatalf("load graph capability fixtures: %v", err)
	}
	for _, c := range fixtures.Derivation.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var p schema.SessionDetailPayload
			if err := json.Unmarshal([]byte(c.Input.DetailJSON), &p); err != nil {
				t.Fatalf("decode durable detail: %v", err)
			}
			if err := schema.ValidateSessionDetailPayload(p); err != nil {
				t.Fatalf("validate durable detail: %v", err)
			}
			if got := schema.RequiredContentCapabilities(p); !slices.Equal(got, c.Expected.Capabilities) {
				t.Fatalf("capabilities=%v, want %v", got, c.Expected.Capabilities)
			}
			before := c.Input.DetailJSON
			if c.Input.ReadJSON != "" {
				var read schema.SessionDetailReadPayload
				if err := json.Unmarshal([]byte(c.Input.ReadJSON), &read); err != nil {
					t.Fatalf("decode read DTO: %v", err)
				}
				if _, err := schema.DecodeSessionDetailPayloadRaw([]byte(c.Input.RejectedDurableJSON)); err == nil {
					t.Fatal("durable input accepted read-only navigation key")
				}
			}
			if c.Input.DetailJSON != before {
				t.Fatal("capability derivation mutated fixture input")
			}
		})
	}
	for _, c := range fixtures.Reader.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var r schema.SchemaVersionResponse
			err := json.Unmarshal([]byte(c.Input.JSON), &r)
			if c.Expected.ErrorContains != "" {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("error=%v, want containing %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			known := schema.KnownContentCapabilities(r.ContentCapabilities)
			if !slices.Equal(known, c.Expected.Known) && len(c.Expected.Known) > 0 {
				t.Fatalf("known=%v want=%v", known, c.Expected.Known)
			}
			missing := schema.MissingContentCapabilities(r.ContentCapabilities, []schema.ContentCapability{schema.ContentCapabilitySessionGraphProvenanceV1})
			if !slices.Equal(missing, c.Expected.Missing) {
				t.Fatalf("missing=%v want=%v", missing, c.Expected.Missing)
			}
		})
	}
	for _, c := range fixtures.Producer.Cases {
		t.Run(c.Name, func(t *testing.T) {
			err := schema.ValidateContentCapabilityAdvertisements(c.Input.Tokens)
			if c.Expected.Accepted != (err == nil) {
				t.Fatalf("accepted=%v want=%v error=%v", err == nil, c.Expected.Accepted, err)
			}
			if err != nil && !strings.Contains(err.Error(), c.Expected.ErrorContains) {
				t.Fatalf("error=%v want containing %q", err, c.Expected.ErrorContains)
			}
		})
	}
}

func TestContentCapabilityInventoryIsCanonical(t *testing.T) {
	want := []schema.ContentCapability{schema.ContentCapabilityDetailedUsageV1, schema.ContentCapabilityNativeMetadataV1, schema.ContentCapabilityObservedModelV1, schema.ContentCapabilitySessionGraphProvenanceV1}
	if !slices.Equal(schema.AllContentCapabilities, want) {
		t.Fatalf("AllContentCapabilities=%v, want exact canonical inventory %v", schema.AllContentCapabilities, want)
	}
	if err := schema.ValidateContentCapabilityAdvertisements(want); err != nil {
		t.Fatalf("canonical producer advertisement rejected: %v", err)
	}
}

func TestContentCapabilitySessionGraphFixtureStrictness(t *testing.T) {
	fixtures, err := schema.LoadContentCapabilitySessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	data, err := yaml.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	unknown := bytes.Replace(data, []byte("detail_json:"), []byte("detail_jsno:"), 1)
	if _, err := schema.DecodeContentCapabilitySessionGraphFixtures(unknown); err == nil {
		t.Fatal("unknown fixture input key was accepted")
	}
	renamed := bytes.Replace(data, []byte("name: legacy-empty"), []byte("name: renamed-legacy-empty"), 1)
	if _, err := schema.DecodeContentCapabilitySessionGraphFixtures(renamed); err == nil {
		t.Fatal("count-preserving required-name rename was accepted")
	}
}

//go:embed testdata/contract/content_capabilities.yaml
var contentCapabilityFixtureYAML []byte

//go:embed testdata/contract/content_capabilities_manifest.yaml
var contentCapabilityManifestYAML []byte

type contentCapabilityInput struct {
	Operation     string                     `yaml:"operation"`
	Advertised    []schema.ContentCapability `yaml:"advertised,omitempty"`
	Required      []schema.ContentCapability `yaml:"required,omitempty"`
	Turns         []contentCapabilityTurn    `yaml:"turns,omitempty"`
	SessionModel  string                     `yaml:"sessionModel,omitempty"`
	Role          schema.Role                `yaml:"role,omitempty"`
	ObservedModel string                     `yaml:"observedModel,omitempty"`
}

type contentCapabilityTurn struct {
	Role          schema.Role `yaml:"role"`
	Depth         int         `yaml:"depth,omitempty"`
	ObservedModel string      `yaml:"observedModel,omitempty"`
}

type contentCapabilityExpected struct {
	Accepted      bool                       `yaml:"accepted"`
	ErrorContains string                     `yaml:"errorContains,omitempty"`
	Capabilities  []schema.ContentCapability `yaml:"capabilities,omitempty"`
}

type contentCapabilityManifest struct {
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

func TestContentCapabilityContractFixtures(t *testing.T) {
	corpus, _ := loadContentCapabilityFixtures(t)
	for _, fixtureCase := range corpus.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			var err error
			var capabilities []schema.ContentCapability
			switch fixtureCase.Input.Operation {
			case "filter":
				capabilities = schema.KnownContentCapabilities(fixtureCase.Input.Advertised)
			case "validate":
				err = schema.ValidateContentCapabilityAdvertisements(fixtureCase.Input.Advertised)
			case "missing":
				capabilities = schema.MissingContentCapabilities(fixtureCase.Input.Advertised, fixtureCase.Input.Required)
				if len(capabilities) != 0 {
					err = fmt.Errorf("required capabilities are missing")
				}
			case "scan":
				turns := make([]schema.TurnDetail, len(fixtureCase.Input.Turns))
				for i, turn := range fixtureCase.Input.Turns {
					turns[i] = schema.TurnDetail{Role: turn.Role, Depth: turn.Depth, ObservedModel: schema.ObservedModelID(turn.ObservedModel)}
				}
				capabilities = schema.RequiredContentCapabilities(schema.SessionDetailPayload{Model: fixtureCase.Input.SessionModel, Turns: turns})
			case "evidence":
				err = schema.ValidateObservedModelEvidence(fixtureCase.Input.Role, schema.ObservedModelID(fixtureCase.Input.ObservedModel))
			default:
				t.Fatalf("unknown fixture operation %q", fixtureCase.Input.Operation)
			}
			if (err == nil) != fixtureCase.Expected.Accepted {
				t.Fatalf("real contract validator accepted=%v, want %v; err=%v", err == nil, fixtureCase.Expected.Accepted, err)
			}
			if err != nil && !strings.Contains(err.Error(), fixtureCase.Expected.ErrorContains) {
				t.Fatalf("validator error %q does not contain %q", err, fixtureCase.Expected.ErrorContains)
			}
			if !slices.Equal(capabilities, fixtureCase.Expected.Capabilities) {
				t.Fatalf("capabilities=%v, want %v", capabilities, fixtureCase.Expected.Capabilities)
			}
		})
	}
}

func TestContentCapabilityInventoryRejectsCountPreservingRename(t *testing.T) {
	_, manifest := loadContentCapabilityFixtures(t)
	old := "name: " + manifest.RequiredCaseNames[0]
	mutated := strings.Replace(string(contentCapabilityFixtureYAML), old, "name: unregistered-capability-case", 1)
	corpus, err := testcase.LoadCorpus[contentCapabilityInput, contentCapabilityExpected]([]byte(mutated))
	if err != nil {
		t.Fatalf("load renamed fixture: %v", err)
	}
	if err := validateContentCapabilityInventory(corpus, manifest); err == nil {
		t.Fatal("count-preserving fixture rename passed the independent name inventory")
	}
}

func loadContentCapabilityFixtures(t *testing.T) (testcase.Corpus[contentCapabilityInput, contentCapabilityExpected], contentCapabilityManifest) {
	t.Helper()
	corpus, err := testcase.LoadCorpus[contentCapabilityInput, contentCapabilityExpected](contentCapabilityFixtureYAML)
	if err != nil {
		t.Fatalf("load capability corpus: %v", err)
	}
	var manifest contentCapabilityManifest
	decoder := yaml.NewDecoder(bytes.NewReader(contentCapabilityManifestYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode capability manifest: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("capability manifest must contain exactly one YAML document: %v", err)
	}
	if err := validateContentCapabilityInventory(corpus, manifest); err != nil {
		t.Fatalf("validate capability fixture inventory: %v", err)
	}
	return corpus, manifest
}

func validateContentCapabilityInventory(corpus testcase.Corpus[contentCapabilityInput, contentCapabilityExpected], manifest contentCapabilityManifest) error {
	if len(corpus.Cases) != manifest.ExpectedCaseCount || len(manifest.RequiredCaseNames) != manifest.ExpectedCaseCount {
		return fmt.Errorf("capability fixture inventory count mismatch: cases=%d names=%d expected=%d", len(corpus.Cases), len(manifest.RequiredCaseNames), manifest.ExpectedCaseCount)
	}
	required := make(map[string]bool, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" || required[name] {
			return fmt.Errorf("capability manifest name %q is empty or duplicated", name)
		}
		required[name] = true
	}
	for _, fixtureCase := range corpus.Cases {
		if !required[fixtureCase.Name] {
			return fmt.Errorf("capability corpus contains unregistered case %q", fixtureCase.Name)
		}
		delete(required, fixtureCase.Name)
	}
	if len(required) != 0 {
		return fmt.Errorf("capability corpus is missing required cases: %v", required)
	}
	return nil
}
