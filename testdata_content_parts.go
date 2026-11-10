package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/content_parts.yaml
var contentPartsYAML []byte

//go:embed testdata/content_parts_manifest.yaml
var contentPartsManifestYAML []byte

//go:embed testdata/content_parts_detail.yaml
var contentPartsDetailYAML []byte

//go:embed testdata/content_parts_detail_manifest.yaml
var contentPartsDetailManifestYAML []byte

// ContentPartFixtureInput is one ordered part plus the validation mode under
// test. Part is a JSON object so a fixture case exercises the real decoder as
// well as the validator.
type ContentPartFixtureInput struct {
	Mode string `yaml:"mode"`
	Part string `yaml:"part"`
}

// ContentPartFixtureExpected is the accepted outcome and, for a rejection, a
// substring naming the rule that must refuse.
type ContentPartFixtureExpected struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"errorContains,omitempty"`
}

// ContentPartsFixtures is the canonical local/published image-part matrix with
// its exact required-name manifest.
type ContentPartsFixtures struct {
	RequiredNames []string                                                             `yaml:"requiredCaseNames"`
	Cases         testcase.Corpus[ContentPartFixtureInput, ContentPartFixtureExpected] `yaml:"cases"`
}

// ContentPartDetailFixtureInput is one whole detail payload plus the payload
// mode under test. Detail is a JSON object so a fixture case exercises the raw
// decoder as well as the payload validator.
type ContentPartDetailFixtureInput struct {
	Mode   string `yaml:"mode"`
	Detail string `yaml:"detail"`
}

// ContentPartDetailFixtureExpected is the accepted outcome and, for a
// rejection, a substring naming the rule that must refuse.
type ContentPartDetailFixtureExpected struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"errorContains,omitempty"`
}

// ContentPartDetailFixtures is the detail-level local/published matrix with its
// exact required-name manifest.
type ContentPartDetailFixtures struct {
	RequiredNames []string                                                                         `yaml:"requiredCaseNames"`
	Cases         testcase.Corpus[ContentPartDetailFixtureInput, ContentPartDetailFixtureExpected] `yaml:"cases"`
}

// LoadContentPartsFixtures loads the canonical ordered content-part matrix and
// proves its exact case-name inventory. The manifest is independent of the
// corpus so a count-preserving rename still fails.
func LoadContentPartsFixtures() (ContentPartsFixtures, error) {
	corpus, err := testcase.LoadCorpus[ContentPartFixtureInput, ContentPartFixtureExpected](contentPartsYAML)
	if err != nil {
		return ContentPartsFixtures{}, fmt.Errorf("load content parts corpus: %w", err)
	}
	names, err := loadRequiredCaseNames(contentPartsManifestYAML, "content parts manifest")
	if err != nil {
		return ContentPartsFixtures{}, err
	}
	if err := requireFixtureNames(corpus.Cases, names); err != nil {
		return ContentPartsFixtures{}, fmt.Errorf("content parts fixtures: %w", err)
	}
	return ContentPartsFixtures{RequiredNames: names, Cases: corpus}, nil
}

// LoadContentPartDetailFixtures loads the canonical detail-level
// local/published matrix and proves its exact case-name inventory.
func LoadContentPartDetailFixtures() (ContentPartDetailFixtures, error) {
	corpus, err := testcase.LoadCorpus[ContentPartDetailFixtureInput, ContentPartDetailFixtureExpected](contentPartsDetailYAML)
	if err != nil {
		return ContentPartDetailFixtures{}, fmt.Errorf("load content part detail corpus: %w", err)
	}
	names, err := loadRequiredCaseNames(contentPartsDetailManifestYAML, "content part detail manifest")
	if err != nil {
		return ContentPartDetailFixtures{}, err
	}
	if err := requireFixtureNames(corpus.Cases, names); err != nil {
		return ContentPartDetailFixtures{}, fmt.Errorf("content part detail fixtures: %w", err)
	}
	return ContentPartDetailFixtures{RequiredNames: names, Cases: corpus}, nil
}

func loadRequiredCaseNames(data []byte, label string) ([]string, error) {
	var manifest struct {
		RequiredCaseNames []string `yaml:"requiredCaseNames"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("load %s: %w", label, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("load %s: expected exactly one YAML document: %v", label, err)
	}
	return manifest.RequiredCaseNames, nil
}
