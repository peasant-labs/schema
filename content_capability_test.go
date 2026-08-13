package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/contract/content_capabilities.yaml
var contentCapabilityFixtureYAML []byte

//go:embed testdata/contract/content_capabilities_manifest.yaml
var contentCapabilityManifestYAML []byte

type contentCapabilityInput struct {
	Capability    schema.ContentCapability        `yaml:"capability,omitempty"`
	Version       schema.ContentCapabilityVersion `yaml:"version,omitempty"`
	Role          schema.Role                     `yaml:"role,omitempty"`
	ObservedModel string                          `yaml:"observedModel,omitempty"`
}

type contentCapabilityExpected struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"errorContains,omitempty"`
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
			if fixtureCase.Input.Capability != "" || fixtureCase.Input.Version != "" {
				err = (schema.ContentCapabilityAdvertisement{Capability: fixtureCase.Input.Capability, Version: fixtureCase.Input.Version}).Validate()
			} else {
				err = schema.ValidateObservedModelEvidence(fixtureCase.Input.Role, schema.ObservedModelID(fixtureCase.Input.ObservedModel))
			}
			if (err == nil) != fixtureCase.Expected.Accepted {
				t.Fatalf("real contract validator accepted=%v, want %v; err=%v", err == nil, fixtureCase.Expected.Accepted, err)
			}
			if err != nil && !strings.Contains(err.Error(), fixtureCase.Expected.ErrorContains) {
				t.Fatalf("validator error %q does not contain %q", err, fixtureCase.Expected.ErrorContains)
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
