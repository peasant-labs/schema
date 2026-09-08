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

type SessionGraphFixtureInput struct {
	Operation    string                         `yaml:"operation"`
	Value        string                         `yaml:"value,omitempty"`
	Relationship *SessionRelationship           `yaml:"relationship,omitempty"`
	Relations    []SessionRelationship          `yaml:"relations,omitempty"`
	Provenance   *ContentProvenance             `yaml:"provenance,omitempty"`
	Navigation   *SessionRelationshipNavigation `yaml:"navigation,omitempty"`
	Group        *HelperGroupSummary            `yaml:"group,omitempty"`
	Context      *HelperContextSummary          `yaml:"context,omitempty"`
	Earlier      *EarlierHistorySection         `yaml:"earlier,omitempty"`
	RawJSON      string                         `yaml:"rawJson,omitempty"`
}
type SessionGraphFixtureExpected struct {
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
	Unknown string   `yaml:"unknown"`
}
type SessionGraphFixtureCorpus struct {
	Durable   DurableGraphFixtures                                                   `yaml:"durable"`
	Refs      testcase.Corpus[SessionGraphRefInput, SessionGraphRefExpected]         `yaml:"refs"`
	Enums     testcase.Corpus[SessionGraphEnumInput, struct{}]                       `yaml:"enums"`
	Semantics testcase.Corpus[SessionGraphFixtureInput, SessionGraphFixtureExpected] `yaml:"semantics"`
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
	if err := c.Semantics.Validate(); err != nil {
		return c, fmt.Errorf("load session graph semantic fixtures: %w", err)
	}
	if err := c.Durable.Validate(); err != nil {
		return c, err
	}
	return c, nil
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

func ValidateSessionGraphFixtureInput(i SessionGraphFixtureInput) error {
	switch i.Operation {
	case "source_ref":
		_, e := NewSourceEntryRef(i.Value)
		return e
	case "submission_ref":
		_, e := NewSubmissionRef(i.Value)
		return e
	case "revision_ref":
		_, e := NewPublicRevisionRef(i.Value)
		return e
	case "relationship":
		if i.Relationship == nil {
			return fmt.Errorf("missing relationship")
		}
		return i.Relationship.Validate()
	case "relations":
		return ValidateSessionRelationships(i.Relations)
	case "provenance":
		if i.Provenance == nil {
			return fmt.Errorf("missing provenance")
		}
		return i.Provenance.Validate()
	case "navigation":
		if i.Navigation == nil {
			return fmt.Errorf("missing navigation")
		}
		return i.Navigation.Validate()
	case "group":
		if i.Group == nil {
			return fmt.Errorf("missing group")
		}
		return i.Group.Validate()
	case "context":
		if i.Context == nil {
			return fmt.Errorf("missing context")
		}
		return i.Context.Validate()
	case "earlier":
		if i.Earlier == nil {
			return fmt.Errorf("missing earlier")
		}
		return i.Earlier.Validate()
	case "raw_relationship":
		var value SessionRelationship
		if err := json.Unmarshal([]byte(i.RawJSON), &value); err != nil {
			return fmt.Errorf("decode raw relationship primitive: %w", err)
		}
		return value.Validate()
	case "raw_navigation":
		var value SessionRelationshipNavigation
		if err := json.Unmarshal([]byte(i.RawJSON), &value); err != nil {
			return fmt.Errorf("decode raw navigation primitive: %w", err)
		}
		return value.Validate()
	default:
		return fmt.Errorf("unknown fixture operation %q", i.Operation)
	}
}
