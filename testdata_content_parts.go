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

// LoadContentPartsFixtures loads the canonical ordered content-part matrix and
// proves its exact case-name inventory. The manifest is independent of the
// corpus so a count-preserving rename still fails.
func LoadContentPartsFixtures() (ContentPartsFixtures, error) {
	corpus, err := testcase.LoadCorpus[ContentPartFixtureInput, ContentPartFixtureExpected](contentPartsYAML)
	if err != nil {
		return ContentPartsFixtures{}, fmt.Errorf("load content parts corpus: %w", err)
	}
	var manifest struct {
		RequiredCaseNames []string `yaml:"requiredCaseNames"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contentPartsManifestYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return ContentPartsFixtures{}, fmt.Errorf("load content parts manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ContentPartsFixtures{}, fmt.Errorf("load content parts manifest: expected exactly one YAML document: %v", err)
	}
	if err := requireFixtureNames(corpus.Cases, manifest.RequiredCaseNames); err != nil {
		return ContentPartsFixtures{}, fmt.Errorf("content parts fixtures: %w", err)
	}
	return ContentPartsFixtures{RequiredNames: manifest.RequiredCaseNames, Cases: corpus}, nil
}
