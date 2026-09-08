package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type SessionGraphFixtureExpected struct {
	ErrorContains string `yaml:"error_contains,omitempty"`
}
type SessionRelationshipsInput struct {
	Relationships []SessionRelationship `yaml:"relationships"`
}
type SessionGraphRawInput struct {
	RawJSON string `yaml:"rawJson"`
}
type SessionGraphRawExpected struct {
	RawJSON       string `yaml:"rawJson,omitempty"`
	ErrorContains string `yaml:"error_contains,omitempty"`
}
type SessionGraphRefInput struct {
	Alias       string `yaml:"alias"`
	BytesBase64 string `yaml:"bytesBase64"`
}
type SessionGraphRefExpected struct {
	ValueBase64   string `yaml:"valueBase64,omitempty"`
	ErrorContains string `yaml:"errorContains,omitempty"`
}
type SessionGraphEnumInput struct {
	Enum    string   `yaml:"enum"`
	Members []string `yaml:"members"`
	Empty   string   `yaml:"empty"`
	Unknown string   `yaml:"unknown"`
}
type SessionGraphFixtureCorpus struct {
	Durable          DurableGraphFixtures                                                        `yaml:"durable"`
	Refs             testcase.Corpus[SessionGraphRefInput, SessionGraphRefExpected]              `yaml:"refs"`
	Enums            testcase.Corpus[SessionGraphEnumInput, struct{}]                            `yaml:"enums"`
	Relationships    testcase.Corpus[SessionRelationship, SessionGraphFixtureExpected]           `yaml:"relationships"`
	RelationshipSets testcase.Corpus[SessionRelationshipsInput, SessionGraphFixtureExpected]     `yaml:"relationship_sets"`
	Provenance       testcase.Corpus[ContentProvenance, SessionGraphFixtureExpected]             `yaml:"provenance"`
	Navigation       testcase.Corpus[SessionRelationshipNavigation, SessionGraphFixtureExpected] `yaml:"navigation"`
	HelperGroups     testcase.Corpus[HelperGroupSummary, SessionGraphFixtureExpected]            `yaml:"helper_groups"`
	HelperContexts   testcase.Corpus[HelperContextSummary, SessionGraphFixtureExpected]          `yaml:"helper_contexts"`
	EarlierHistory   testcase.Corpus[EarlierHistorySection, SessionGraphFixtureExpected]         `yaml:"earlier_history"`
	RawRelationships testcase.Corpus[SessionGraphRawInput, SessionGraphRawExpected]              `yaml:"raw_relationships"`
	RawNavigation    testcase.Corpus[SessionGraphRawInput, SessionGraphRawExpected]              `yaml:"raw_navigation"`
}

func LoadSessionGraphFixtures() (SessionGraphFixtureCorpus, error) {
	var c SessionGraphFixtureCorpus
	decoder := yaml.NewDecoder(bytes.NewReader(SessionGraphProvenanceYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&c); err != nil {
		return c, fmt.Errorf("load session graph fixtures: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return c, fmt.Errorf("load session graph fixtures trailing document: %w", err)
		}
		return c, fmt.Errorf("load session graph fixtures: multiple YAML documents are not allowed")
	}
	if err := c.Refs.Validate(); err != nil {
		return c, fmt.Errorf("load session graph ref fixtures: %w", err)
	}
	if err := c.Enums.Validate(); err != nil {
		return c, fmt.Errorf("load session graph enum fixtures: %w", err)
	}
	if err := validatePrimitiveFixtureArms(c); err != nil {
		return c, err
	}
	if err := c.Durable.Validate(); err != nil {
		return c, err
	}
	return c, nil
}

func validatePrimitiveFixtureArms(c SessionGraphFixtureCorpus) error {
	arms := []struct {
		name     string
		validate func() error
	}{
		{"relationships", c.Relationships.Validate}, {"relationship_sets", c.RelationshipSets.Validate}, {"provenance", c.Provenance.Validate}, {"navigation", c.Navigation.Validate}, {"helper_groups", c.HelperGroups.Validate}, {"helper_contexts", c.HelperContexts.Validate}, {"earlier_history", c.EarlierHistory.Validate}, {"raw_relationships", c.RawRelationships.Validate}, {"raw_navigation", c.RawNavigation.Validate},
	}
	for _, arm := range arms {
		if err := arm.validate(); err != nil {
			return fmt.Errorf("load session graph %s fixtures: %w", arm.name, err)
		}
	}
	return nil
}

func ConstructSessionGraphRef(i SessionGraphRefInput) (string, error) {
	b, e := base64.StdEncoding.DecodeString(i.BytesBase64)
	if e != nil {
		return "", fmt.Errorf("decode reference fixture bytes: %w", e)
	}
	raw := string(b)
	switch i.Alias {
	case "source":
		v, e := NewSourceEntryRef(raw)
		return string(v), e
	case "submission":
		v, e := NewSubmissionRef(raw)
		return string(v), e
	case "revision":
		v, e := NewPublicRevisionRef(raw)
		return string(v), e
	default:
		return "", fmt.Errorf("unknown reference alias %q", i.Alias)
	}
}

func ValidateRawRelationshipFixture(i SessionGraphRawInput) (SessionRelationship, error) {
	var value SessionRelationship
	if err := json.Unmarshal([]byte(i.RawJSON), &value); err != nil {
		return value, fmt.Errorf("decode raw relationship primitive: %w", err)
	}
	return value, value.Validate()
}
func ValidateRawNavigationFixture(i SessionGraphRawInput) (SessionRelationshipNavigation, string, error) {
	var value SessionRelationshipNavigation
	if err := json.Unmarshal([]byte(i.RawJSON), &value); err != nil {
		return value, "", fmt.Errorf("decode raw navigation primitive: %w", err)
	}
	if err := value.Validate(); err != nil {
		return value, "", err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return value, "", fmt.Errorf("encode raw navigation primitive: %w", err)
	}
	return value, string(encoded), nil
}
