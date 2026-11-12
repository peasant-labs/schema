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
	HelperCount  int      `yaml:"helper_count"`
}
type durableCountExpected struct {
	Present bool  `yaml:"present"`
	Value   int64 `yaml:"value"`
}
type durableErrorExpected struct {
	ErrorContains string `yaml:"error_contains"`
}

// DurableMirrorInput describes two copies of one session's facts. Surface
// selects the pair under test: empty is the publication projection (session
// detail beside metadata); local_row, village_collective and village_my_share
// are grouped rows, where "detail" is the session copy and "metadata" is the
// route copy.
type DurableMirrorInput struct {
	Surface                   string `yaml:"surface"`
	DetailCount               *int64 `yaml:"detail_count"`
	MetadataCount             *int64 `yaml:"metadata_count"`
	DetailConversationCount   *int64 `yaml:"detail_conversation_count"`
	MetadataConversationCount *int64 `yaml:"metadata_conversation_count"`
	DetailDisplayedCount      *int64 `yaml:"detail_displayed_count"`
	MetadataDisplayedCount    *int64 `yaml:"metadata_displayed_count"`
	Parent                    string `yaml:"parent"`
	GraphParent               string `yaml:"graph_parent"`
	DetailRoot                string `yaml:"detail_root"`
	MetadataRoot              string `yaml:"metadata_root"`
	ProjectedMainCount        bool   `yaml:"projected_main_count"`
	DetailTurnCount           int    `yaml:"detail_turn_count"`
	MetadataTurnCount         int    `yaml:"metadata_turn_count"`
}
type durableMirrorExpected struct {
	Accept        bool   `yaml:"accept"`
	ErrorContains string `yaml:"error_contains"`
}
type durableDigestInput struct {
	LeftPresent       bool   `yaml:"left_present"`
	LeftValue         int64  `yaml:"left_value"`
	RightPresent      bool   `yaml:"right_present"`
	RightValue        int64  `yaml:"right_value"`
	LeftConversation  *int64 `yaml:"left_conversation"`
	RightConversation *int64 `yaml:"right_conversation"`
	LeftVariant       string `yaml:"left_variant"`
	RightVariant      string `yaml:"right_variant"`
}
type durableDigestExpected struct {
	Different bool `yaml:"different"`
	// RightDigest, when set, is the literal fingerprint of the right request.
	// It pins the frame encoding, which is frozen once a release carries it.
	RightDigest string `yaml:"right_digest"`
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
