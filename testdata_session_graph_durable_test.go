package schema_test

import (
	_ "embed"
	"fmt"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/session_graph_provenance.yaml
var durableGraphFixtureYAML []byte

type durableJSONInput struct {
	JSON string `yaml:"json"`
}
type durableRoundTripExpected struct {
	TurnCount    int      `yaml:"turn_count"`
	InputPresent bool     `yaml:"input_present"`
	InputCount   int64    `yaml:"input_count"`
	MainRefs     []string `yaml:"main_refs"`
	EarlierRefs  []string `yaml:"earlier_refs"`
}
type durableCountExpected struct {
	Present bool  `yaml:"present"`
	Value   int64 `yaml:"value"`
}
type durableErrorExpected struct {
	ErrorContains string `yaml:"error_contains"`
}
type durableMirrorInput struct {
	DetailCount   int64  `yaml:"detail_count"`
	MetadataCount int64  `yaml:"metadata_count"`
	Parent        string `yaml:"parent"`
	GraphParent   string `yaml:"graph_parent"`
}
type durableMirrorExpected struct {
	Accept        bool   `yaml:"accept"`
	ErrorContains string `yaml:"error_contains"`
}
type durableDigestInput struct {
	LeftPresent  bool  `yaml:"left_present"`
	LeftValue    int64 `yaml:"left_value"`
	RightPresent bool  `yaml:"right_present"`
	RightValue   int64 `yaml:"right_value"`
}
type durableDigestExpected struct {
	Different bool `yaml:"different"`
}

type durableGraphFixtures struct {
	RoundTrip     testcase.Corpus[durableJSONInput, durableRoundTripExpected] `yaml:"round_trip"`
	Counts        testcase.Corpus[durableJSONInput, durableCountExpected]     `yaml:"counts"`
	InvalidCounts testcase.Corpus[durableJSONInput, durableErrorExpected]     `yaml:"invalid_counts"`
	Mirrors       testcase.Corpus[durableMirrorInput, durableMirrorExpected]  `yaml:"mirrors"`
	Digest        testcase.Corpus[durableDigestInput, durableDigestExpected]  `yaml:"digest"`
	RequiredNames []string                                                    `yaml:"required_names"`
}

func loadDurableGraphFixtures() (durableGraphFixtures, error) {
	var envelope struct {
		Durable durableGraphFixtures `yaml:"durable"`
	}
	if err := yaml.Unmarshal(durableGraphFixtureYAML, &envelope); err != nil {
		return durableGraphFixtures{}, fmt.Errorf("load durable graph fixtures: %w", err)
	}
	return envelope.Durable, nil
}
