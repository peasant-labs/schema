package schema

import (
	"fmt"

	"github.com/peasant-labs/schema/testcase"
)

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
type DurableMirrorInput struct {
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

type DurableGraphFixtures struct {
	RoundTrip     testcase.Corpus[durableJSONInput, durableRoundTripExpected] `yaml:"round_trip"`
	Counts        testcase.Corpus[durableJSONInput, durableCountExpected]     `yaml:"counts"`
	InvalidCounts testcase.Corpus[durableJSONInput, durableErrorExpected]     `yaml:"invalid_counts"`
	Mirrors       testcase.Corpus[DurableMirrorInput, durableMirrorExpected]  `yaml:"mirrors"`
	Digest        testcase.Corpus[durableDigestInput, durableDigestExpected]  `yaml:"digest"`
	RequiredNames []string                                                    `yaml:"required_names"`
}

func (c DurableGraphFixtures) Validate() error {
	if err := c.RoundTrip.Validate(); err != nil {
		return fmt.Errorf("load durable graph round-trip fixtures: %w", err)
	}
	if err := c.Counts.Validate(); err != nil {
		return fmt.Errorf("load durable graph count fixtures: %w", err)
	}
	if err := c.InvalidCounts.Validate(); err != nil {
		return fmt.Errorf("load durable graph invalid-count fixtures: %w", err)
	}
	if err := c.Mirrors.Validate(); err != nil {
		return fmt.Errorf("load durable graph mirror fixtures: %w", err)
	}
	if err := c.Digest.Validate(); err != nil {
		return fmt.Errorf("load durable graph digest fixtures: %w", err)
	}
	return nil
}
