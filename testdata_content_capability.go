package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/content_capability_session_graph.yaml
var contentCapabilitySessionGraphYAML []byte

type ContentCapabilityDetailInput struct {
	DetailJSON          string `yaml:"detail_json"`
	ReadJSON            string `yaml:"read_json,omitempty"`
	RejectedDurableJSON string `yaml:"rejected_durable_json,omitempty"`
}
type ContentCapabilityExpected struct {
	Capabilities  []ContentCapability `yaml:"capabilities,omitempty"`
	Known         []ContentCapability `yaml:"known,omitempty"`
	Missing       []ContentCapability `yaml:"missing,omitempty"`
	Accepted      bool                `yaml:"accepted,omitempty"`
	ErrorContains string              `yaml:"error_contains,omitempty"`
}
type ContentCapabilityAdvertisementInput struct {
	JSON string `yaml:"json"`
}
type ContentCapabilityProducerInput struct {
	Tokens []ContentCapability `yaml:"tokens"`
}
type ContentCapabilityFixtureNames struct {
	Derivation []string `yaml:"derivation"`
	Reader     []string `yaml:"reader"`
	Producer   []string `yaml:"producer"`
}
type ContentCapabilitySessionGraphFixtures struct {
	RequiredNames ContentCapabilityFixtureNames                                                   `yaml:"required_names"`
	Derivation    testcase.Corpus[ContentCapabilityDetailInput, ContentCapabilityExpected]        `yaml:"derivation"`
	Reader        testcase.Corpus[ContentCapabilityAdvertisementInput, ContentCapabilityExpected] `yaml:"reader"`
	Producer      testcase.Corpus[ContentCapabilityProducerInput, ContentCapabilityExpected]      `yaml:"producer"`
}

// LoadContentCapabilitySessionGraphFixtures exposes the canonical source corpus
// used by Go validation and by generated TypeScript fixture consumers.
func LoadContentCapabilitySessionGraphFixtures() (ContentCapabilitySessionGraphFixtures, error) {
	return DecodeContentCapabilitySessionGraphFixtures(contentCapabilitySessionGraphYAML)
}

// DecodeContentCapabilitySessionGraphFixtures strictly decodes a supplied
// corpus. Generators use this seam to consume the same YAML bytes.
func DecodeContentCapabilitySessionGraphFixtures(data []byte) (ContentCapabilitySessionGraphFixtures, error) {
	var f ContentCapabilitySessionGraphFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		return f, fmt.Errorf("load content capability session graph fixtures: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return f, fmt.Errorf("load content capability session graph fixtures: expected exactly one YAML document: %v", err)
	}
	if err := f.Derivation.Validate(); err != nil {
		return f, err
	}
	if err := f.Reader.Validate(); err != nil {
		return f, err
	}
	if err := f.Producer.Validate(); err != nil {
		return f, err
	}
	if err := requireFixtureNames(f.Derivation.Cases, f.RequiredNames.Derivation); err != nil {
		return f, fmt.Errorf("derivation fixtures: %w", err)
	}
	if err := requireFixtureNames(f.Reader.Cases, f.RequiredNames.Reader); err != nil {
		return f, fmt.Errorf("reader fixtures: %w", err)
	}
	if err := requireFixtureNames(f.Producer.Cases, f.RequiredNames.Producer); err != nil {
		return f, fmt.Errorf("producer fixtures: %w", err)
	}
	return f, nil
}

func requireFixtureNames[I, E any](cases []testcase.Case[I, E], required []string) error {
	seen := make(map[string]bool, len(cases))
	for _, c := range cases {
		if seen[c.Name] {
			return fmt.Errorf("duplicate case %q", c.Name)
		}
		seen[c.Name] = true
	}
	requiredSeen := make(map[string]bool, len(required))
	for _, name := range required {
		if name == "" || requiredSeen[name] {
			return fmt.Errorf("required name %q is empty or duplicated", name)
		}
		requiredSeen[name] = true
	}
	if len(seen) != len(requiredSeen) {
		return fmt.Errorf("exact required-name membership differs: cases=%d required=%d", len(seen), len(required))
	}
	for name := range requiredSeen {
		if !seen[name] {
			return fmt.Errorf("required case %q is missing", name)
		}
	}
	return nil
}
