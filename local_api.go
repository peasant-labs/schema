package schema

import "time"

// --- Dashboard ---

// DashboardPayload is the data sent on the dashboard WebSocket channel.
type DashboardPayload struct {
	TotalSessions      int             `json:"totalSessions"`
	TotalTokens        int             `json:"totalTokens"`
	AvgDurationMins    float64         `json:"avgDurationMins"`
	HarnessBreakdown   map[Harness]int `json:"harnessBreakdown"`
	AvgTurnsPerSession float64         `json:"avgTurnsPerSession"`
	AcceptanceRate     float64         `json:"acceptanceRate"`
}

// --- Sessions ---

// SessionSummary is a session without turns, used in the sessions list.
type SessionSummary struct {
	ID            string    `json:"id"`
	Harness       Harness   `json:"harness"`
	StartTime     time.Time `json:"startTime"`
	DurationMins  float64   `json:"durationMins"`
	TotalTokens   int       `json:"totalTokens"`
	TurnCount     int       `json:"turnCount"`
	ToolCallCount int       `json:"toolCallCount"`
	Project       string    `json:"project,omitempty"`
	// ProjectHash is the opaque project identifier (projects.project_hash).
	// The frontend resolves display name → hash from this field for the
	// Map/Review REST endpoints (contract §9.1).
	ProjectHash     string  `json:"projectHash,omitempty"`
	Outcome         string  `json:"outcome,omitempty"`
	ParentSessionID *string `json:"parentSessionId,omitempty"`
	// Preview is the raw first user message of the session, sourced from the
	// already-redacted indexed transcript (session_entries.content_preview). It
	// is redaction-safe by construction. Empty when the session has no indexed
	// user entry. The web client formats it for display.
	Preview string `json:"preview,omitempty"`
}

// SessionsPayload is the data sent on the sessions WebSocket channel.
type SessionsPayload struct {
	Sessions []SessionSummary `json:"sessions"`
}

// --- Session Detail ---

// TurnDetail is a turn with full content for the detail view.
type TurnDetail struct {
	Index       int              `json:"index"`
	Role        Role             `json:"role"`
	Content     string           `json:"content"`
	ToolCalls   []ToolCallDetail `json:"toolCalls,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
	Depth       int              `json:"depth"`
	ParentIndex *int             `json:"parentIndex,omitempty"`
	AgentName   string           `json:"agentName,omitempty"`

	// Enrichment fields — propagated from session_entries.
	EntryType   EntryType   `json:"entryType,omitempty"`
	HasThinking bool        `json:"hasThinking,omitempty"`
	StopReason  *StopReason `json:"stopReason,omitempty"`
	TokensIn    *int        `json:"tokensIn,omitempty"`
	TokensOut   *int        `json:"tokensOut,omitempty"`
}

// ToolCallDetail is a tool call in the detail view.
type ToolCallDetail struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Arguments  string       `json:"arguments"`
	Result     string       `json:"result"`
	DurationMs *int         `json:"durationMs,omitempty"`
	ExitCode   *int         `json:"exitCode,omitempty"`
	FilePath   string       `json:"filePath,omitempty"`
	IsError    bool         `json:"isError,omitempty"`
	ToolKind   ToolCallKind `json:"toolKind,omitempty"`
}

// SessionDetailPayload is the data sent on the session_detail WebSocket channel
// AND promoted to the versioned `peasant push` wire body inside a
// TranscriptContent envelope.
//
// SchemaVersion is the EMBEDDED push-contract version. It is advisory: when
// this payload travels inside a TranscriptContent envelope the envelope's
// ContractVersion is authoritative (see TranscriptContent's conflict-winner
// note). The embedded copy keeps the payload self-describing when stored on its
// own. It is omitempty so the local WebSocket session_detail channel (which
// never sets it) does not gain a spurious "schemaVersion":"" field; the push
// builder sets it explicitly.
type SessionDetailPayload struct {
	SchemaVersion PushContractVersion `json:"schemaVersion,omitempty"`
	ID            string              `json:"id"`
	Harness       Harness             `json:"harness"`
	StartTime     time.Time           `json:"startTime"`
	EndTime       time.Time           `json:"endTime"`
	DurationMins  float64             `json:"durationMins"`
	TotalTokens   int                 `json:"totalTokens"`
	TokensIn      int                 `json:"tokensIn"`
	TokensOut     int                 `json:"tokensOut"`
	TurnCount     int                 `json:"turnCount"`
	ToolCallCount int                 `json:"toolCallCount"`
	Turns         []TurnDetail        `json:"turns"`
	// Optional fields — populated when backend has the data.
	Source           string            `json:"source,omitempty"`
	Status           string            `json:"status,omitempty"`
	Project          string            `json:"project,omitempty"`
	Model            string            `json:"model,omitempty"`
	WorkingDirectory string            `json:"workingDirectory,omitempty"`
	GitBranch        string            `json:"gitBranch,omitempty"`
	GitRemote        string            `json:"gitRemote,omitempty"`
	ChildSessions    []ChildSessionRef `json:"childSessions,omitempty"`
	// Outcome is the heuristic resolution status of the session
	// (resolved/partial/failed), sourced from session_metrics.outcome. Empty
	// when the session has no computed outcome. The metadata columns expose no
	// per-signal reason, so no reason field accompanies it yet.
	Outcome SessionOutcome `json:"outcome,omitempty"`
	// Scorecard carries the per-session quality signals used by the "How this
	// session went" self-assessment card. Nil when the session has no computed
	// metrics. Sourced from the same session_metrics row that backs QualitySession.
	Scorecard *SessionScorecard `json:"scorecard,omitempty"`
}

// SessionScorecard holds the deterministic per-session quality signals needed
// by the Highlights self-assessment card. It is a flat projection of the
// session_metrics fields the three axis cards (token efficiency, prompt
// quality, loop efficiency) consume. Fields are pointers so the client can
// distinguish "not computed" (nil) from a real zero value when applying
// threshold bands.
type SessionScorecard struct {
	// Token efficiency inputs.
	M2TokenOutcomeRatio     *float64 `json:"m2TokenOutcomeRatio,omitempty"`
	M5ContextUtilizationPct *float64 `json:"m5ContextUtilizationPct,omitempty"`
	M6OutputSurvivalPct     *float64 `json:"m6OutputSurvivalPct,omitempty"`
	RetryTokensWasted       *int     `json:"retryTokensWasted,omitempty"`
	TotalTokens             *int     `json:"totalTokens,omitempty"`
	CostTotalUSD            *float64 `json:"costTotalUsd,omitempty"`
	// Prompt quality inputs.
	SpecQualityScore     *float64 `json:"specQualityScore,omitempty"`
	SignalDensity        *float64 `json:"signalDensity,omitempty"`
	M7SpecHasExamples    *bool    `json:"m7SpecHasExamples,omitempty"`
	M7SpecHasConstraints *bool    `json:"m7SpecHasConstraints,omitempty"`
	// Loop efficiency inputs.
	M4ConsecutiveErrorMax *int `json:"m4ConsecutiveErrorMax,omitempty"`
	WithinSessionReverts  *int `json:"withinSessionReverts,omitempty"`
	// Outcome echoes the session outcome so the card can apply the
	// "failed outcome with above-median cost" token-efficiency trigger.
	Outcome SessionOutcome `json:"outcome,omitempty"`
}

// ChildSessionRef is a lightweight reference to a child (subagent) session.
type ChildSessionRef struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"startTime"`
	Project   string    `json:"project,omitempty"`
}

// --- Trends ---

// DayStats holds aggregated data for a single day in trends.
type DayStats struct {
	Date     string `json:"date"` // "2006-01-02" format
	Tokens   int    `json:"tokens"`
	Sessions int    `json:"sessions"`
}

// TrendsPayload is the data sent on the trends WebSocket channel.
type TrendsPayload struct {
	Days          []DayStats `json:"days"`
	TotalTokens   int        `json:"totalTokens"`
	TotalSessions int        `json:"totalSessions"`
}

// --- Quality ---

// QualityPayload is the data sent on the quality WebSocket channel.
type QualityPayload struct {
	Sessions []QualitySession `json:"sessions"`
}

// QualitySession is the analytics payload for the quality dashboard.
type QualitySession struct {
	ID              string  `json:"id"`
	Date            string  `json:"date"` // "2006-01-02" format
	Project         string  `json:"project"`
	TotalTokens     int     `json:"totalTokens"`
	InputTokens     int     `json:"inputTokens"`
	OutputTokens    int     `json:"outputTokens"`
	TurnCount       int     `json:"turnCount"`
	ToolCalls       int     `json:"toolCalls"`
	DurationMinutes float64 `json:"durationMinutes"`
	// v2 fields -- zero-valued until implemented.
	Scope                string  `json:"scope"`
	Title                string  `json:"title"`
	Outcome              string  `json:"outcome"`
	FilesTouched         int     `json:"filesTouched"`
	LinesChanged         int     `json:"linesChanged"`
	RetryLoops           int     `json:"retryLoops"`
	RetryTokensWasted    int     `json:"retryTokensWasted"`
	WithinSessionReverts int     `json:"withinSessionReverts"`
	SignalDensity        float64 `json:"signalDensity"`
	SpecQualityScore     float64 `json:"specQualityScore"`
	ExplorationRatio     float64 `json:"explorationRatio"`
	ScopeBreadth         int     `json:"scopeBreadth"`
	DiscoveryTurns       int     `json:"discoveryTurns"`
	// EffectiveAnnotations are the non-superseded annotations for this session.
	// Clients derive display labels (e.g. humanLabel, agentLabel) by filtering
	// on typeId and annotatorKind — no pre-computed string fields needed.
	EffectiveAnnotations []AnnotationSummary `json:"effectiveAnnotations,omitempty"`
}

// --- Project Familiarity ---

// InteractionType classifies how deeply a session engaged with a file.
type InteractionType string

const (
	InteractionMentioned  InteractionType = "mentioned"
	InteractionRead       InteractionType = "read"
	InteractionDiscussed  InteractionType = "discussed"
	InteractionQuestioned InteractionType = "questioned"
)

func (i InteractionType) String() string { return string(i) }

// DecayLevel classifies how recently a file was engaged.
type DecayLevel string

const (
	DecayFresh      DecayLevel = "fresh"
	DecayFading     DecayLevel = "fading"
	DecayStale      DecayLevel = "stale"
	DecayUnexplored DecayLevel = "unexplored"
)

func (d DecayLevel) String() string { return string(d) }

// FileFamiliarity represents familiarity data for a single file.
type FileFamiliarity struct {
	Path          string     `json:"path"`
	Depth         int        `json:"depth"` // 0=unexplored, 1=touched, 2=engaged, 3=deep
	SessionCount  int        `json:"sessionCount"`
	TotalTurns    int        `json:"totalTurns"`
	HumanTurns    int        `json:"humanTurns"`
	LastEngagedAt *string    `json:"lastEngagedAt"` // ISO 8601, nil if never
	DaysSince     *int       `json:"daysSince"`     // nil if never engaged
	DecayLevel    DecayLevel `json:"decayLevel"`
	IsSourceFile  bool       `json:"isSourceFile"` // true for relevant source files, false for config/generated
}

// WalkthroughStep is a single file visit in a session trail.
type WalkthroughStep struct {
	File    string `json:"file"`
	Line    *int   `json:"line,omitempty"`
	Excerpt string `json:"excerpt"`
}

// WalkthroughTrail is the file path extracted from a single session.
type WalkthroughTrail struct {
	SessionID  string            `json:"sessionId"`
	Title      string            `json:"title"`
	Date       string            `json:"date"`
	TurnCount  int               `json:"turnCount"`
	Steps      []WalkthroughStep `json:"steps"`
	IsCoherent bool              `json:"isCoherent"` // true if linear path, false if scattered
}

// ReviewSuggestion nudges the user to revisit a fading file.
type ReviewSuggestion struct {
	Path            string `json:"path"`
	LastEngaged     string `json:"lastEngaged"`
	DaysSince       int    `json:"daysSince"`
	SuggestedPrompt string `json:"suggestedPrompt"`
}

// FamiliarityPayload is the data sent on the project_familiarity WebSocket channel.
type FamiliarityPayload struct {
	ProjectHash     string             `json:"projectHash"`
	FamiliarityPct  float64            `json:"familiarityPct"`  // % of source files engaged
	UnexploredCount int                `json:"unexploredCount"` // source files with 0 engagement
	FreshnessDays   *int               `json:"freshnessDays"`   // days since last learning session
	Files           []FileFamiliarity  `json:"files"`
	Trails          []WalkthroughTrail `json:"trails"`
	Suggestions     []ReviewSuggestion `json:"suggestions"`
}

// --- Annotations ---

// AnnotationsPayload is the data sent on the annotations WebSocket channel.
// Axis and ID echo the subscription parameters so clients can correlate updates.
type AnnotationsPayload struct {
	Axis        AnnotationAxis      `json:"axis"`
	ID          string              `json:"id"`
	Annotations []AnnotationSummary `json:"annotations"`
}

// CreateAnnotationRequest is the JSON body for POST /api/v1/annotations.
// Target selection rules (mutually exclusive):
//   - Session-level: SessionID set, no entry/annotation fields.
//   - Entry-level:   SessionID + TargetEntryIndex set; TargetEntryEndIndex optional (defaults to index+1).
//   - Meta-annotation: TargetAnnotationID set (annotation on annotation).
type CreateAnnotationRequest struct {
	SessionID     string   `json:"sessionId"`
	TypeID        string   `json:"typeId"`
	Value         string   `json:"value"`
	IsPrimary     bool     `json:"isPrimary"`
	Confidence    *float64 `json:"confidence,omitempty"`
	Reason        *string  `json:"reason,omitempty"`
	AnnotatorName string   `json:"annotatorName,omitempty"`
	// Entry-level targeting — if set, creates an annotation_target_entry row.
	// TargetEntryEndIndex defaults to TargetEntryIndex+1 (single-entry span) when omitted.
	TargetEntryIndex    *int `json:"targetEntryIndex,omitempty"`
	TargetEntryEndIndex *int `json:"targetEntryEndIndex,omitempty"`
	// Meta-annotation targeting — if set, creates an annotation_target_annotation row.
	TargetAnnotationID *string `json:"targetAnnotationId,omitempty"`
}

// CreateAnnotationResponse is the JSON response for POST /api/v1/annotations (201 Created).
type CreateAnnotationResponse struct {
	ID string `json:"id"`
}

// --- Health ---

// HealthResponse is the JSON response for GET /api/v1/health.
type HealthResponse struct {
	Status string `json:"status"`
}

// --- Mock Config ---

// MockConfigResponse is the JSON response for GET /api/v1/config/mock.
type MockConfigResponse struct {
	Enabled bool     `json:"enabled"`
	Web     []string `json:"web,omitempty"`
	TUI     []string `json:"tui,omitempty"`
	API     []string `json:"api,omitempty"`
}

// --- Shutdown ---

// ShutdownResponse is the JSON response for POST /api/v1/shutdown.
type ShutdownResponse struct {
	Status string `json:"status"`
}

// --- WebSocket Messages ---

// MessageType represents the type discriminator for WebSocket messages.
type MessageType string

const (
	MsgSubscribe          MessageType = "subscribe"
	MsgUnsubscribe        MessageType = "unsubscribe"
	MsgDashboard          MessageType = "dashboard"
	MsgSessions           MessageType = "sessions"
	MsgSessionDetail      MessageType = "session_detail"
	MsgTrends             MessageType = "trends"
	MsgQuality            MessageType = "quality"
	MsgAnnotations        MessageType = "annotations"
	MsgProjectFamiliarity MessageType = "project_familiarity"
	MsgConnected          MessageType = "connected"
	MsgError              MessageType = "error"
)

// ChannelTopic identifies a subscribable data stream.
type ChannelTopic string

const (
	TopicDashboard          ChannelTopic = "dashboard"
	TopicSessions           ChannelTopic = "sessions"
	TopicSessionDetail      ChannelTopic = "session_detail"
	TopicTrends             ChannelTopic = "trends"
	TopicQuality            ChannelTopic = "quality"
	TopicAnnotations        ChannelTopic = "annotations"
	TopicProjectFamiliarity ChannelTopic = "project_familiarity"
)

// AnnotationAxis is the subscription dimension for annotation channels.
type AnnotationAxis string

const (
	AxisType    AnnotationAxis = "type"
	AxisSession AnnotationAxis = "session"
	AxisProject AnnotationAxis = "project"
)

// ChannelSubscription describes a single channel subscription.
// Topic is always required. Additional fields are topic-specific:
//   - TopicSessionDetail: ID = session ID
//   - TopicAnnotations:   Axis + ID
type ChannelSubscription struct {
	Topic ChannelTopic   `json:"topic"`
	ID    string         `json:"id,omitempty"`
	Axis  AnnotationAxis `json:"axis,omitempty"`
}

// ClientMessage is sent from the browser to the server via WebSocket.
type ClientMessage struct {
	Type     MessageType           `json:"type"`
	Channels []ChannelSubscription `json:"channels,omitempty"`
}

// ServerMessage is a message sent from the server to the browser via WebSocket.
type ServerMessage struct {
	Type    MessageType `json:"type"`
	Data    any         `json:"data,omitempty"`
	Message string      `json:"message,omitempty"` // for error type
	Version string      `json:"version,omitempty"` // for connected type
}
