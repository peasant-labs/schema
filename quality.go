package schema

// QualityMetrics holds session statistics and derived quality signals for a session.
// The type name is intentionally broad post-unification: it covers two distinct groups.
//
// Group A — Basic session stats (TurnCount, SubagentCount, TotalTokens, InputTokens,
// OutputTokens, ToolCalls, TitleGenerated, Outcome, DurationMinutes): populated for all
// fully-ingested sessions; a nil value here indicates the session has not been fully ingested.
//
// Group B — Derived quality signals (FilesTouched, LinesChanged, RetryLoops, M-series
// fields, SignalDensity, SpecQualityScore, ExplorationRatio, etc.): populated only when
// quality analysis has run. Nil = "not computed" — analysis may not have been performed yet.
//
// All fields are pointers with omitempty so that serialisation omits fields that have not
// been populated, regardless of which group they belong to.
// JSON tags use camelCase to match the backend data contract (QualitySession) field names.
type QualityMetrics struct {
	// Basic session counts
	TurnCount            *int            `json:"turnCount,omitempty"`
	SubagentCount        *int            `json:"subagentCount,omitempty"`
	TotalTokens          *int            `json:"totalTokens,omitempty"`
	InputTokens          *int            `json:"inputTokens,omitempty"`
	OutputTokens         *int            `json:"outputTokens,omitempty"`
	ToolCalls            *int            `json:"toolCalls,omitempty"`
	TitleGenerated       *string         `json:"titleGenerated,omitempty"`
	Outcome              *SessionOutcome `json:"outcome,omitempty"`
	FilesTouched         *int            `json:"filesTouched,omitempty"`
	LinesChanged         *int            `json:"linesChanged,omitempty"`
	RetryLoops           *int            `json:"retryLoops,omitempty"`
	RetryTokensWasted    *int            `json:"retryTokensWasted,omitempty"`
	WithinSessionReverts *int            `json:"withinSessionReverts,omitempty"`
	SignalDensity        *float64        `json:"signalDensity,omitempty"`
	SpecQualityScore     *float64        `json:"specQualityScore,omitempty"`
	ExplorationRatio     *float64        `json:"explorationRatio,omitempty"`
	ScopeBreadth         *int            `json:"scopeBreadth,omitempty"`
	DiscoveryTurns       *int            `json:"discoveryTurns,omitempty"`
	DurationMinutes      *float64        `json:"durationMinutes,omitempty"`
	// M-series advanced metrics (v2 computed)
	M2TokenOutcomeRatio     *float64 `json:"m2TokenOutcomeRatio,omitempty"`
	M3UniqueToolCount       *int     `json:"m3UniqueToolCount,omitempty"`
	M4ErrorRecoveryCount    *int     `json:"m4ErrorRecoveryCount,omitempty"`
	M4ConsecutiveErrorMax   *int     `json:"m4ConsecutiveErrorMax,omitempty"`
	M5ContextUtilizationPct *float64 `json:"m5ContextUtilizationPct,omitempty"`
	M5PeakContextTokens     *int     `json:"m5PeakContextTokens,omitempty"`
	M5AvgMessageTokens      *int     `json:"m5AvgMessageTokens,omitempty"`
	M6OutputSurvivalPct     *float64 `json:"m6OutputSurvivalPct,omitempty"`
	M6LinesSurvived         *int     `json:"m6LinesSurvived,omitempty"`
	M6LinesTotal            *int     `json:"m6LinesTotal,omitempty"`
	M7SpecWordCount         *int     `json:"m7SpecWordCount,omitempty"`
	M7SpecHasExamples       *bool    `json:"m7SpecHasExamples,omitempty"`
	M7SpecHasConstraints    *bool    `json:"m7SpecHasConstraints,omitempty"`
	// v3 cost analytics
	CostInputUSD      *float64 `json:"costInputUsd,omitempty"`
	CostOutputUSD     *float64 `json:"costOutputUsd,omitempty"`
	CostReasoningUSD  *float64 `json:"costReasoningUsd,omitempty"`
	CostCacheReadUSD  *float64 `json:"costCacheReadUsd,omitempty"`
	CostCacheWriteUSD *float64 `json:"costCacheWriteUsd,omitempty"`
	CostTotalUSD      *float64 `json:"costTotalUsd,omitempty"`
	CostModelID       *string  `json:"costModelId,omitempty"`
	// v3 scope classification
	Scope *string `json:"scope,omitempty"`
	// Compute metadata
	ComputedAt     *int64 `json:"computedAt,omitempty"`     // Unix millis
	ComputeVersion *int   `json:"computeVersion,omitempty"` // 0=v1 migrated, 1+=v2 computed
}
