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
type NativeLimitInput struct {
	MainRecords    int   `yaml:"main_records"`
	EarlierRecords []int `yaml:"earlier_records"`
	DataBytes      int   `yaml:"data_bytes"`
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
	Durable               DurableGraphFixtures                                                        `yaml:"durable"`
	Refs                  testcase.Corpus[SessionGraphRefInput, SessionGraphRefExpected]              `yaml:"refs"`
	Enums                 testcase.Corpus[SessionGraphEnumInput, struct{}]                            `yaml:"enums"`
	Relationships         testcase.Corpus[SessionRelationship, SessionGraphFixtureExpected]           `yaml:"relationships"`
	RelationshipSets      testcase.Corpus[SessionRelationshipsInput, SessionGraphFixtureExpected]     `yaml:"relationship_sets"`
	Provenance            testcase.Corpus[ContentProvenance, SessionGraphFixtureExpected]             `yaml:"provenance"`
	Navigation            testcase.Corpus[SessionRelationshipNavigation, SessionGraphFixtureExpected] `yaml:"navigation"`
	HelperGroups          testcase.Corpus[HelperGroupSummary, SessionGraphFixtureExpected]            `yaml:"helper_groups"`
	HelperContexts        testcase.Corpus[HelperContextSummary, SessionGraphFixtureExpected]          `yaml:"helper_contexts"`
	EarlierHistory        testcase.Corpus[EarlierHistorySection, SessionGraphFixtureExpected]         `yaml:"earlier_history"`
	RawRelationships      testcase.Corpus[SessionGraphRawInput, SessionGraphRawExpected]              `yaml:"raw_relationships"`
	RawNavigation         testcase.Corpus[SessionGraphRawInput, SessionGraphRawExpected]              `yaml:"raw_navigation"`
	Recursive             testcase.Corpus[SessionGraphRawInput, SessionGraphFixtureExpected]          `yaml:"recursive"`
	RawDurable            testcase.Corpus[SessionGraphRawInput, SessionGraphFixtureExpected]          `yaml:"raw_durable"`
	NativeLimits          testcase.Corpus[NativeLimitInput, SessionGraphFixtureExpected]              `yaml:"native_limits"`
	AuthoritativeRaw      testcase.Corpus[SessionGraphRawInput, SessionGraphFixtureExpected]          `yaml:"authoritative_raw"`
	RecursiveRequired     []string                                                                    `yaml:"recursive_required_names"`
	RawDurableRequired    []string                                                                    `yaml:"raw_durable_required_names"`
	NativeLimitRequired   []string                                                                    `yaml:"native_limit_required_names"`
	AuthoritativeRequired []string                                                                    `yaml:"authoritative_required_names"`
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
	if err := validateRequiredCaseNames("recursive", c.Recursive.Cases, c.RecursiveRequired); err != nil {
		return c, err
	}
	if err := validateRequiredCaseNames("raw_durable", c.RawDurable.Cases, c.RawDurableRequired); err != nil {
		return c, err
	}
	if err := validateRequiredCaseNames("native_limits", c.NativeLimits.Cases, c.NativeLimitRequired); err != nil {
		return c, err
	}
	if err := validateRequiredCaseNames("authoritative_raw", c.AuthoritativeRaw.Cases, c.AuthoritativeRequired); err != nil {
		return c, err
	}
	return c, nil
}

func validateRequiredCaseNames[I, E any](arm string, cases []testcase.Case[I, E], required []string) error {
	want, got := map[string]bool{}, map[string]bool{}
	for _, name := range required {
		if want[name] {
			return fmt.Errorf("load session graph %s fixtures: duplicate required name %q", arm, name)
		}
		want[name] = true
	}
	for _, c := range cases {
		got[c.Name] = true
	}
	for name := range want {
		if !got[name] {
			return fmt.Errorf("load session graph %s fixtures: required case %q is missing", arm, name)
		}
	}
	for name := range got {
		if !want[name] {
			return fmt.Errorf("load session graph %s fixtures: case %q is absent from required-name manifest", arm, name)
		}
	}
	return nil
}

func validatePrimitiveFixtureArms(c SessionGraphFixtureCorpus) error {
	arms := []struct {
		name     string
		validate func() error
	}{
		{"relationships", c.Relationships.Validate}, {"relationship_sets", c.RelationshipSets.Validate}, {"provenance", c.Provenance.Validate}, {"navigation", c.Navigation.Validate}, {"helper_groups", c.HelperGroups.Validate}, {"helper_contexts", c.HelperContexts.Validate}, {"earlier_history", c.EarlierHistory.Validate}, {"raw_relationships", c.RawRelationships.Validate}, {"raw_navigation", c.RawNavigation.Validate}, {"recursive", c.Recursive.Validate}, {"raw_durable", c.RawDurable.Validate}, {"native_limits", c.NativeLimits.Validate}, {"authoritative_raw", c.AuthoritativeRaw.Validate},
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
