package schema

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// PublishVerdictFixtures is the parsed testdata/publish/verdicts.yaml corpus.
type PublishVerdictFixtures struct {
	Cases []PublishVerdictCase `yaml:"cases"`
}

// PublishVerdictCase is one concrete publish body and its schema/HTTP verdict.
type PublishVerdictCase struct {
	Name   string                    `yaml:"name"`
	Body   string                    `yaml:"body"`
	Expect PublishVerdictExpectation `yaml:"expect"`
}

// PublishVerdictExpectation attaches expected outcomes to a publish verdict row.
type PublishVerdictExpectation struct {
	SchemaAccepts bool   `yaml:"schema_accepts"`
	LegacyAccepts *bool  `yaml:"legacy_accepts,omitempty"`
	HTTPStatus    int    `yaml:"http_status,omitempty"`
	ErrorCategory string `yaml:"error_category,omitempty"`
	ErrorContains string `yaml:"error_contains,omitempty"`
}

// LoadPublishVerdictFixtures parses PublishVerdictsYAML into structured fixtures.
func LoadPublishVerdictFixtures() (*PublishVerdictFixtures, error) {
	var f PublishVerdictFixtures
	if err := yaml.Unmarshal(PublishVerdictsYAML, &f); err != nil {
		return nil, fmt.Errorf("load publish verdict fixtures: %w", err)
	}
	return &f, nil
}

// CaseByName returns the verdict row with the given stable name.
func (f *PublishVerdictFixtures) CaseByName(name string) (PublishVerdictCase, bool) {
	for _, c := range f.Cases {
		if c.Name == name {
			return c, true
		}
	}
	return PublishVerdictCase{}, false
}

// ModelHarness returns body.model.harness when the case body contains one.
func (c PublishVerdictCase) ModelHarness() (string, bool, error) {
	var doc struct {
		Model struct {
			Harness string `json:"harness"`
		} `json:"model"`
	}
	if err := json.Unmarshal([]byte(c.Body), &doc); err != nil {
		return "", false, fmt.Errorf("parse publish verdict body %q: %w", c.Name, err)
	}
	if doc.Model.Harness == "" {
		return "", false, nil
	}
	return doc.Model.Harness, true, nil
}
