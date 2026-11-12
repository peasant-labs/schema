package schema

import (
	"fmt"
	"math"
	"time"
	"unicode"
	"unicode/utf8"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- Dashboard ---

// DashboardPayload is the data sent on the dashboard WebSocket channel.
type DashboardPayload struct {
	TotalSessions    int             `json:"totalSessions"`
	TotalTokens      int             `json:"totalTokens"`
	AvgDurationMins  float64         `json:"avgDurationMins"`
	HarnessBreakdown map[Harness]int `json:"harnessBreakdown"`
	// AvgTurnsPerSession is the mean conversationTurnCount per session.
	AvgTurnsPerSession float64 `json:"avgTurnsPerSession"`
	AcceptanceRate     float64 `json:"acceptanceRate"`
}

// --- Sessions ---

// SessionOrigin is who drove a recorded session, as declared by the producer
// that recorded it.
//
// CONTRACT RULE, binding on every consumer:
//
//   - "user" and "agent" are decisions. A consumer stores the declared value and
//     does not re-derive it.
//   - "unknown" is also a decision: the producer looked and could not tell. A
//     consumer MUST fall back to whatever rule it would have applied had the
//     field been absent. It MUST NOT store "unknown" in place of a judgement it
//     is itself capable of making.
//   - An ABSENT field means the producer expressed no opinion, typically a build
//     older than this field. A consumer falls back the same way.
//
// So "unknown" and absent produce identical behaviour. They stay distinct
// because they are different facts: one producer examined the session, the other
// never knew the question existed. Only the first is evidence; both instruct the
// consumer to decide for itself.
type SessionOrigin string

const (
	// SessionOriginUser: a person drove the session.
	SessionOriginUser SessionOrigin = "user"
	// SessionOriginAgent: a program started the session.
	SessionOriginAgent SessionOrigin = "agent"
	// SessionOriginUnknown: the producer examined the session and could not
	// establish who drove it. It instructs the consumer to apply its own rule.
	SessionOriginUnknown SessionOrigin = "unknown"
)

// AllSessionOrigins is the canonical menu of declared session origins.
var AllSessionOrigins = []SessionOrigin{
	SessionOriginUser,
	SessionOriginAgent,
	SessionOriginUnknown,
}

// IsValid reports whether o is a value this contract accepts on the wire. It
// accepts the empty string: an absent declaration is valid and means the
// producer expressed no opinion.
func (o SessionOrigin) IsValid() bool {
	switch o {
	case "", SessionOriginUser, SessionOriginAgent, SessionOriginUnknown:
		return true
	}
	return false
}

func (o SessionOrigin) String() string { return string(o) }

// SessionTeammateMaxIdentityRunes is the bound on each SessionTeammate
// identity field, counted in Unicode code points. It matches the published
// maxLength, which counts code points too.
const SessionTeammateMaxIdentityRunes = 256

// SessionTeammateEvidence is the kind of record a teammate identity rests on.
// It is a closed set of its own rather than a member of EvidenceKind: the
// session-graph evidence menu is embedded in the published Village wire, so a
// kind that only a local read carries must not widen it for every consumer.
type SessionTeammateEvidence string

const (
	// SessionTeammateEvidenceNativeDynamicWorkflow: the stored parent
	// transcript records spawning this exact team and name.
	SessionTeammateEvidenceNativeDynamicWorkflow SessionTeammateEvidence = "native_dynamic_workflow"
)

// AllSessionTeammateEvidences is the canonical teammate evidence menu.
var AllSessionTeammateEvidences = []SessionTeammateEvidence{SessionTeammateEvidenceNativeDynamicWorkflow}

// IsValid reports whether e is a member of the teammate evidence menu.
func (e SessionTeammateEvidence) IsValid() bool { return inSet(e, AllSessionTeammateEvidences) }

// String returns the wire representation of the teammate evidence kind.
func (e SessionTeammateEvidence) String() string { return string(e) }

// JSONSchema implements jsonschema.Exposer.
func (SessionTeammateEvidence) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Session Teammate Evidence",
		"Record a local teammate identity rests on; only a native dynamic workflow spawn is recorded",
		AllSessionTeammateEvidences,
	), nil
}

// SessionTeammate is the Claude Code teammate identity of a session whose
// stored parent transcript records spawning that identity. It is local read
// evidence, never published transcript content: a session with only ordinary
// subagent parentage, another harness's parentage, or no recorded spawn edge
// carries no teammate object, so a consumer can tell the two apart.
type SessionTeammate struct {
	// Team is the recorded team name the spawn edge names. Non-empty, valid
	// UTF-8, at most 256 characters (Unicode code points), and no control
	// characters. The published minLength, maxLength, and pattern carry the
	// same three rules; Validate applies them directly.
	Team string `json:"team" minLength:"1" maxLength:"256" pattern:"^[^\\x00-\\x1F\\x7F-\\x9F]*$" description:"Recorded Claude Code team name of the spawn edge. Non-empty, valid UTF-8, at most 256 characters (Unicode code points), and no control characters."`
	// Name is the recorded teammate name the spawn edge names, under the same
	// bound as Team.
	Name string `json:"name" minLength:"1" maxLength:"256" pattern:"^[^\\x00-\\x1F\\x7F-\\x9F]*$" description:"Recorded Claude Code teammate name of the spawn edge. Non-empty, valid UTF-8, at most 256 characters (Unicode code points), and no control characters."`
	// Evidence is the kind of record the spawn edge rests on. Only
	// native_dynamic_workflow is accepted.
	Evidence SessionTeammateEvidence `json:"evidence"`
}

// Validate checks that the teammate identity is present, exact, and carries
// the one evidence kind a recorded spawn edge produces. The identity fields
// carry the published 256-character bound plus the no-control-character rule.
func (t *SessionTeammate) Validate() error {
	if t == nil {
		return nil
	}
	const where = "teammate validation failed at schema.SessionTeammate.Validate while projecting local session evidence: "
	if err := validateTeammateIdentityField(where, "team", t.Team); err != nil {
		return err
	}
	if err := validateTeammateIdentityField(where, "name", t.Name); err != nil {
		return err
	}
	if !t.Evidence.IsValid() {
		return fmt.Errorf("%sevidence %q is outside the closed set; a teammate object only exists for a spawn edge the parent transcript records, so another kind would claim a link the store cannot support; omit teammate or emit a member of schema.AllSessionTeammateEvidences", where, t.Evidence)
	}
	return nil
}

// validateTeammateIdentityField applies the identity field rule: non-empty
// valid UTF-8, at most SessionTeammateMaxIdentityRunes Unicode code points, and
// no control characters.
func validateTeammateIdentityField(where, field, value string) error {
	if value == "" {
		return fmt.Errorf("%s%s is empty, so the identity cannot be matched to the parent's recorded spawn; emit the harness-recorded %s or omit teammate", where, field, field)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s%s %q is not valid UTF-8, so it cannot be emitted on the wire without changing the harness bytes; emit the recorded %s as valid UTF-8", where, field, value, field)
	}
	if utf8.RuneCountInString(value) > SessionTeammateMaxIdentityRunes {
		return fmt.Errorf("%s%s is %d characters, over the %d-character bound, and truncation would fabricate an identity; emit the harness-recorded %s or omit teammate", where, field, utf8.RuneCountInString(value), SessionTeammateMaxIdentityRunes, field)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s%s %q contains the control character U+%04X, which a rendered identity line cannot carry; emit the harness-recorded %s without control characters or omit teammate", where, field, value, r, field)
		}
	}
	return nil
}

// equalOptionalTeammate reports whether two optional teammate identities carry
// the same recorded identity.
func equalOptionalTeammate(a, b *SessionTeammate) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// JSONSchema implements jsonschema.Exposer. The published enum is the declared
// menu only; the empty string is the absence of the field, which omitempty
// already expresses, so it is not a menu member.
func (SessionOrigin) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Session Origin",
		"Who drove a recorded session, as declared by the producer that recorded it",
		AllSessionOrigins,
	), nil
}

// SessionSummary is a session without turns, used in the sessions list.
type SessionSummary struct {
	ID                   string    `json:"id"`
	Harness              Harness   `json:"harness"`
	StartTime            time.Time `json:"startTime"`
	DurationMins         float64   `json:"durationMins"`
	TotalTokens          int       `json:"totalTokens"`
	TurnCount            int       `json:"turnCount" description:"Turns: the number of folded turns of the session's main partition, system and tool rows included. Equal to the segment directory's totalTurns at the same revision. Not the count a reader is shown; see conversationTurnCount."`
	InputSubmissionCount *int64    `json:"inputSubmissionCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	// ConversationTurnCount is the number of conversation turns in the session's
	// main partition (IsConversationTurn): the count a reader is shown. Absent
	// means unmeasured: a producer or publication older than this field. It
	// equals the segment directory's totalConversationTurns at the same revision.
	ConversationTurnCount *int64 `json:"conversationTurnCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	ToolCallCount         int    `json:"toolCallCount"`
	Project               string `json:"project,omitempty"`
	// ProjectHash is the opaque project identifier (projects.project_hash).
	// The frontend resolves display name → hash from this field for the
	// Map/Review REST endpoints (contract §9.1).
	ProjectHash ProjectHash `json:"projectHash,omitempty"`
	Outcome     string      `json:"outcome,omitempty"`
	// ParentSessionID is the durable logical parent of a spawned child. It may
	// name a session this store does not hold. Absent for a root or a fork.
	ParentSessionID *string `json:"parentSessionId,omitempty"`
	// ParentUnresolved marks a known spawned child whose parent cannot be
	// named. It is never set together with ParentSessionID. A client keeps
	// such a row out of the top-level list and does not invent a parent.
	ParentUnresolved *bool                 `json:"parentUnresolved,omitempty" nullable:"false"`
	RootSessionID    *SessionID            `json:"rootSessionId,omitempty"`
	Purpose          SessionPurpose        `json:"purpose,omitempty"`
	Relationships    []SessionRelationship `json:"relationships,omitempty"`
	// SessionOrigin is the producer's declaration of who drove this session.
	// Absent when the producer expressed no opinion; see SessionOrigin for the
	// rule a consumer applies to an absent or "unknown" value.
	SessionOrigin SessionOrigin `json:"sessionOrigin,omitempty"`
	// Preview is the raw first user message of the session, sourced from the
	// already-redacted indexed transcript (session_entries.content_preview). It
	// is redaction-safe by construction. Empty when the session has no indexed
	// user entry. The web client formats it for display.
	Preview string `json:"preview,omitempty"`
	// Teammate is the Claude Code teammate identity this session was spawned
	// under, present only when the stored parent transcript records spawning
	// that exact identity. Ordinary subagent parentage carries no teammate
	// object. Local read evidence, never published transcript content.
	Teammate *SessionTeammate `json:"teammate,omitempty"`
}

// SessionDetailReadPayload is the local read projection. Anonymous embedding
// keeps every durable detail field at the JSON root while adding authorized
// navigation metadata that is never part of published transcript content.
type SessionDetailReadPayload struct {
	SessionDetailPayload
	// Title is the stored display title of the session, read from the stored
	// session record rather than reconstructed from the quality rows. It is
	// omitted when the producer has no stored title, and it never changes the
	// durable push payload.
	Title string `json:"title,omitempty"`
	// Medians is the library-wide personal-medians baseline, cached per
	// revision. Nil when the library holds no metric sample.
	Medians *PersonalMedians `json:"medians,omitempty"`
	// Revision is store_state.revision rendered as a decimal string. It is the
	// same revision the quality page carries, so a client can drop cached
	// medians when it changes. It is omitted when a durable payload is read
	// without the local revision projection.
	Revision               string                          `json:"revision,omitempty"`
	RelationshipNavigation []SessionRelationshipNavigation `json:"relationshipNavigation,omitempty"`
	// Teammate is the Claude Code teammate identity this session was spawned
	// under, present only when the stored parent transcript records spawning
	// that exact identity. Local read evidence, never published transcript
	// content: SessionDetailPayload itself never carries it.
	Teammate *SessionTeammate `json:"teammate,omitempty"`
}

// Validate checks the durable detail payload, the authorized navigation
// metadata, and the optional local teammate evidence.
func (p SessionDetailReadPayload) Validate() error {
	if err := ValidateSessionDetailPayload(p.SessionDetailPayload); err != nil {
		return err
	}
	for _, navigation := range p.RelationshipNavigation {
		if err := navigation.Validate(); err != nil {
			return err
		}
	}
	return p.Teammate.Validate()
}

// LocalSyncSummary preserves the existing sync-list row alongside its matching
// SessionSummary when grouped list mode is requested.
type LocalSyncSummary struct {
	ID                   string      `json:"id"`
	Harness              Harness     `json:"harness"`
	ProjectName          string      `json:"projectName"`
	ProjectHash          ProjectHash `json:"projectHash"`
	HostSlug             string      `json:"hostSlug"`
	StartTime            string      `json:"startTime"`
	DurationMs           int64       `json:"durationMs"`
	TotalTokens          int         `json:"totalTokens"`
	TurnCount            int         `json:"turnCount" description:"Turns: the number of folded turns of the session's main partition, system and tool rows included. Equal to the segment directory's totalTurns at the same revision. Not the count a reader is shown; see conversationTurnCount."`
	Model                string      `json:"model"`
	InputSubmissionCount *int64      `json:"inputSubmissionCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	// ConversationTurnCount is the number of conversation turns in the session's
	// main partition (IsConversationTurn): the count a reader is shown. Absent
	// means unmeasured: a producer or publication older than this field. It
	// equals the segment directory's totalConversationTurns at the same revision.
	ConversationTurnCount *int64     `json:"conversationTurnCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	SyncStatus            SyncStatus `json:"syncStatus"`
	// HoldReason says why a held row waits. Only a held row carries one.
	HoldReason SyncHoldReason `json:"holdReason,omitempty"`
	// PreviouslyPushed reports durable evidence that this session was
	// published before. It is independent of SyncStatus: a held row can have
	// been published, so a client reads this field instead of inferring it.
	PreviouslyPushed bool `json:"previouslyPushed"`
	// ParentSessionID is the durable logical parent of a spawned child, the
	// same value its SessionSummary carries. It may name a session this store
	// does not hold. Absent for a root or a fork.
	ParentSessionID *string `json:"parentSessionId,omitempty" nullable:"false"`
	// ParentUnresolved marks a known spawned child whose parent cannot be
	// named. At most one of ParentSessionID and ParentUnresolved=true is set.
	ParentUnresolved *bool `json:"parentUnresolved,omitempty" nullable:"false"`
	// Teammate is the Claude Code teammate identity this session was spawned
	// under, the same value its SessionSummary carries. Present only when the
	// stored parent transcript records spawning that exact identity.
	Teammate *SessionTeammate `json:"teammate,omitempty"`
}

// LocalSyncSessionsPayload is the exact omitted-view response produced by the
// existing local sync chooser. Every row carries its status; a session absent
// from the response has no status here, and a client must not read absence as
// new.
type LocalSyncSessionsPayload struct {
	Sessions []LocalSyncSummary `json:"sessions" nullable:"false"`
}

// LocalSessionRow is one concrete local route row. Sync and Matches retain the
// existing route-specific data instead of coercing every route into one shape.
// Title is the stored session title; the search variant omits it because its
// match rows already carry the matched content.
type LocalSessionRow struct {
	Session SessionSummary    `json:"session"`
	Sync    *LocalSyncSummary `json:"sync,omitempty"`
	Matches []SearchResult    `json:"matches,omitempty" nullable:"false"`
	Title   string            `json:"title,omitempty"`
}

// LocalSessionPage is the flat offset page of concrete route rows.
type LocalSessionPage = OffsetPage[LocalSessionRow]

// LocalSessionListItem is exactly one transcript row or owner context
// container, with zero or more collapsed helper groups.
type LocalSessionListItem struct {
	Kind         SessionListItemKind   `json:"kind"`
	Transcript   *LocalSessionRow      `json:"transcript,omitempty"`
	Context      *HelperContextSummary `json:"context,omitempty"`
	HelperGroups []HelperGroupSummary  `json:"helperGroups,omitempty" nullable:"false"`
}

// LocalSessionListPayload is the opt-in grouped REST list response. It carries
// the shared offset page fields plus the grouped-only counts.
type LocalSessionListPayload struct {
	Items                []LocalSessionListItem `json:"items" nullable:"false"`
	NextItems            []LocalSessionListItem `json:"nextItems" nullable:"false"`
	Page                 int                    `json:"page" minimum:"1" maximum:"9007199254740991"`
	PerPage              PageSize               `json:"perPage"`
	Total                int                    `json:"total" minimum:"0" maximum:"9007199254740991"`
	Revision             string                 `json:"revision" minLength:"1"`
	OrdinarySessionTotal int                    `json:"ordinarySessionTotal" minimum:"0" maximum:"9007199254740991"`
	HelperThreadTotal    int                    `json:"helperThreadTotal" minimum:"0" maximum:"9007199254740991"`
}

// LocalHelperMembersPayload is a page of saved helper sessions selected by an
// opaque scope returned from the originating grouped query. It carries the same
// offset fields as every other local list page.
type LocalHelperMembersPayload struct {
	Members   []LocalSessionListItem `json:"members" nullable:"false"`
	NextItems []LocalSessionListItem `json:"nextItems" nullable:"false"`
	Page      int                    `json:"page" minimum:"1" maximum:"9007199254740991"`
	PerPage   PageSize               `json:"perPage"`
	Total     int                    `json:"total" minimum:"0" maximum:"9007199254740991"`
	Revision  string                 `json:"revision" minLength:"1"`
}

// Validate checks route-variant identity and count mirrors without deriving
// one metric from another.
func (r LocalSessionRow) Validate() error {
	if r.Session.ID == "" {
		return fmt.Errorf("local session row validation failed at schema.LocalSessionRow.Validate during grouped response construction: session.id is empty; callers cannot open the saved transcript; provide its durable local session ID")
	}
	if err := ValidateInputSubmissionCount(r.Session.InputSubmissionCount, "LocalSessionRow.session.inputSubmissionCount"); err != nil {
		return err
	}
	if err := ValidateConversationTurnCount(r.Session.ConversationTurnCount, "LocalSessionRow.session.conversationTurnCount"); err != nil {
		return err
	}
	if err := r.Session.Teammate.Validate(); err != nil {
		return err
	}
	if r.Sync != nil && len(r.Matches) != 0 {
		return fmt.Errorf("local session row validation failed for %q at schema.LocalSessionRow.Validate during grouped route projection: sync and matches are both populated; the row cannot belong to two incompatible route variants; retain only the originating sync or search arm", r.Session.ID)
	}
	if r.Sync != nil {
		s := r.Sync
		if err := s.Validate(); err != nil {
			return err
		}
		if s.ID != r.Session.ID || s.Harness != r.Session.Harness || s.ProjectName != r.Session.Project || s.ProjectHash != r.Session.ProjectHash || s.TotalTokens != r.Session.TotalTokens || s.TurnCount != r.Session.TurnCount || !equalOptionalInt64(s.InputSubmissionCount, r.Session.InputSubmissionCount) || !equalOptionalInt64(s.ConversationTurnCount, r.Session.ConversationTurnCount) || !equalOptionalString(s.ParentSessionID, r.Session.ParentSessionID) || !equalOptionalBool(s.ParentUnresolved, r.Session.ParentUnresolved) {
			return fmt.Errorf("local session row validation failed for %q at schema.LocalSessionRow.Validate during grouped sync projection: sync identity, parent, or count mirrors disagree with session; callers could display or select the wrong saved transcript; derive both summaries from the same selected database row", r.Session.ID)
		}
		if !equalOptionalTeammate(s.Teammate, r.Session.Teammate) {
			return fmt.Errorf("local session row validation failed for %q at schema.LocalSessionRow.Validate during grouped sync projection: sync.teammate disagrees with session.teammate; the list and the sync chooser would show different teammate identities for one saved transcript; derive both summaries from the same evidence read", r.Session.ID)
		}
	}
	for i, match := range r.Matches {
		if match.SessionID != r.Session.ID {
			return fmt.Errorf("local session row validation failed for %q at schema.LocalSessionRow.Validate during grouped search projection: matches[%d].sessionId=%q; search evidence belongs to another transcript; retain only matches for this exact session", r.Session.ID, i, match.SessionID)
		}
	}
	return nil
}

// Validate checks the discriminated item arm and every nested route row/group.
func (i LocalSessionListItem) Validate() error {
	if !i.Kind.IsValid() {
		return fmt.Errorf("local session list item validation failed at schema.LocalSessionListItem.Validate during grouped response construction: kind %q is outside the closed set; callers cannot select an item arm; use a member of schema.AllSessionListItemKinds", i.Kind)
	}
	if (i.Transcript != nil) == (i.Context != nil) || i.Kind == SessionListItemTranscript && i.Transcript == nil || i.Kind == SessionListItemContextContainer && i.Context == nil {
		return fmt.Errorf("local session list item validation failed at schema.LocalSessionListItem.Validate during grouped response construction: kind must select exactly one matching transcript or context arm; callers cannot safely render or select the item; populate only the arm named by kind")
	}
	if i.Transcript != nil {
		if err := i.Transcript.Validate(); err != nil {
			return err
		}
	}
	if i.Context != nil {
		if err := i.Context.Validate(); err != nil {
			return err
		}
	}
	return validateHelperGroups(i.HelperGroups)
}

// Validate checks non-null collection and pagination/count boundaries.
func (p LocalSessionListPayload) Validate() error {
	if err := validateOffsetPagination(p.Page, p.PerPage, p.Total, p.Revision); err != nil {
		return err
	}
	if p.Items == nil || p.NextItems == nil || p.OrdinarySessionTotal < 0 || p.HelperThreadTotal < 0 || int64(p.OrdinarySessionTotal) > maxSafeJSONInteger || int64(p.HelperThreadTotal) > maxSafeJSONInteger {
		return fmt.Errorf("local session list validation failed at schema.LocalSessionListPayload.Validate during grouped response construction: items and nextItems must be arrays and both grouped counts must be nonnegative safe integers; callers cannot page a coherent selected set; initialize both arrays and emit bounded counts")
	}
	for _, item := range p.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	for _, item := range p.NextItems {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks a typed, non-null member page and every row mirror.
func (p LocalHelperMembersPayload) Validate() error {
	if err := validateOffsetPagination(p.Page, p.PerPage, p.Total, p.Revision); err != nil {
		return err
	}
	if len(p.Members) > int(p.PerPage) || len(p.Members) > p.Total {
		return fmt.Errorf("local helper members validation failed at schema.LocalHelperMembersPayload.Validate during paging: member count exceeds perPage or direct total; clients cannot page this scope; count direct saved helpers before paging")
	}
	if p.Members == nil || p.NextItems == nil {
		return fmt.Errorf("local helper members validation failed at schema.LocalHelperMembersPayload.Validate during scoped member response construction: members and nextItems must be arrays; callers cannot page the authorized scope; initialize both arrays and emit bounded pagination values")
	}
	seen := make(map[string]bool)
	groups := make(map[string]bool)
	for _, row := range p.Members {
		if row.Kind != SessionListItemTranscript || row.Transcript == nil || row.Context != nil {
			return fmt.Errorf("local helper members validation failed at schema.LocalHelperMembersPayload.Validate during scoped response construction: member is not a transcript; containers are not saved helpers; emit only transcript items with their immediate helper groups")
		}
		if err := row.Validate(); err != nil {
			return err
		}
		id := row.Transcript.Session.ID
		if seen[id] {
			return fmt.Errorf("local helper members validation failed at schema.LocalHelperMembersPayload.Validate during scoped response construction: duplicate member identity; paging would count one helper twice; emit each saved transcript once")
		}
		seen[id] = true
		for _, group := range row.HelperGroups {
			if groups[group.GroupID] {
				return fmt.Errorf("local helper members validation failed at schema.LocalHelperMembersPayload.Validate during paging: a groupId appears under multiple members; clients cannot assign its owner; emit each immediate-owner group once")
			}
			groups[group.GroupID] = true
		}
	}
	for _, row := range p.NextItems {
		if err := row.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SessionsPayload is the data sent on the sessions WebSocket channel.
type SessionsPayload struct {
	Sessions []SessionSummary `json:"sessions"`
}

// --- Session Detail ---

// TurnDetail is a turn with full content for the detail view.
type TurnDetail struct {
	Index   int    `json:"index"`
	Role    Role   `json:"role"`
	Content string `json:"content"`
	// ContentParts preserves the ordered text and image blocks of a turn. It is
	// additional evidence, never a replacement: Content remains the concatenated
	// text fallback and ToolCalls remains tool evidence. Parts stay in source
	// order, and an image-only turn keeps its carrier turn with empty Content.
	ContentParts []TurnContentPart `json:"contentParts,omitempty" nullable:"false"`
	ToolCalls    []ToolCallDetail  `json:"toolCalls,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	Depth        int               `json:"depth"`
	ParentIndex  *int              `json:"parentIndex,omitempty"`
	AgentName    string            `json:"agentName,omitempty"`
	// ObservedModel is the exact model identifier observed while producing this
	// assistant or subagent turn. It is optional source evidence: when present it
	// uses valid UTF-8 with no leading or trailing Unicode whitespace, while accepted
	// bytes are preserved exactly. Producers enforce the assistant-only role condition; generated shape
	// validators do not infer it from Role.
	ObservedModel ObservedModelID `json:"observedModel,omitempty"`
	// ThinkingLevel/ThinkingLevelRaw are assistant-turn evidence: set only on
	// assistant-generated output, never user or system turns. Raw follows the
	// same "differs from canonical, including unmapped" rule as the session field.
	ThinkingLevel    ThinkingLevel      `json:"thinkingLevel,omitempty"`
	ThinkingLevelRaw ThinkingLevelRaw   `json:"thinkingLevelRaw,omitempty"`
	SourceEntryRef   SourceEntryRef     `json:"sourceEntryRef,omitempty"`
	Provenance       *ContentProvenance `json:"provenance,omitempty"`
	Usage            *UsageDetail       `json:"usage,omitempty"`

	// Command is present when this turn invoked a skill or a user-defined
	// slash command. It is optional and safely ignorable; the invocation name
	// is also in Content. Producers set it on the turn that carries the
	// invocation, whatever Role that turn has: a harness-injected command
	// wrapper is commonly a RoleSystem turn. It is never set for built-in
	// harness commands. Consumers must not infer Role from Command or Command
	// from Role.
	Command *CommandInvocation `json:"command,omitempty"`

	// ControlKind names the control record this turn was projected from. Its
	// values are "mode", "permission-mode", "attachment", "agent-name", and
	// "custom-title". It is producer-owned evidence: a consumer must never infer
	// it from Role, Content, or any other field, and must treat a turn without
	// controlKind as an ordinary turn. A compaction notice (Claude
	// "compact-boundary", Codex "compacted") is a system row and never carries
	// controlKind.
	ControlKind *string `json:"controlKind,omitempty"`

	// Enrichment fields — propagated from session_entries.
	EntryType   EntryType   `json:"entryType,omitempty"`
	HasThinking bool        `json:"hasThinking,omitempty"`
	StopReason  *StopReason `json:"stopReason,omitempty"`
	TokensIn    *int        `json:"tokensIn,omitempty"`
	TokensOut   *int        `json:"tokensOut,omitempty"`
}

// ToolCallDetail is a tool call in the detail view.
type ToolCallDetail struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Namespace is independent source evidence, including a present empty string.
	// Omission means not recorded; producers preserve the exact redacted value.
	Namespace        *string            `json:"namespace,omitempty" nullable:"false"`
	Arguments        string             `json:"arguments"`
	Result           string             `json:"result"`
	DurationMs       *int               `json:"durationMs,omitempty"`
	ExitCode         *int               `json:"exitCode,omitempty"`
	FilePath         string             `json:"filePath,omitempty"`
	IsError          bool               `json:"isError,omitempty"`
	ToolKind         ToolCallKind       `json:"toolKind,omitempty"`
	CallEntryRef     SourceEntryRef     `json:"callEntryRef,omitempty"`
	ResultEntryRef   SourceEntryRef     `json:"resultEntryRef,omitempty"`
	CallProvenance   *ContentProvenance `json:"callProvenance,omitempty"`
	ResultProvenance *ContentProvenance `json:"resultProvenance,omitempty"`
	Usage            *UsageDetail       `json:"usage,omitempty"`
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
	SchemaVersion        PushContractVersion `json:"schemaVersion,omitempty"`
	ID                   string              `json:"id"`
	Harness              Harness             `json:"harness"`
	StartTime            time.Time           `json:"startTime"`
	EndTime              time.Time           `json:"endTime"`
	DurationMins         float64             `json:"durationMins"`
	TotalTokens          int                 `json:"totalTokens"`
	TokensIn             int                 `json:"tokensIn"`
	TokensOut            int                 `json:"tokensOut"`
	TurnCount            int                 `json:"turnCount" description:"Turns: the number of folded turns of the session's main partition, system and tool rows included. Equal to the segment directory's totalTurns at the same revision. Not the count a reader is shown; see conversationTurnCount."`
	InputSubmissionCount *int64              `json:"inputSubmissionCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	// ConversationTurnCount is the number of conversation turns in the session's
	// main partition (IsConversationTurn): the count a reader is shown. Absent
	// means unmeasured: a producer or publication older than this field. It
	// equals the segment directory's totalConversationTurns at the same revision.
	ConversationTurnCount *int64       `json:"conversationTurnCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	ToolCallCount         int          `json:"toolCallCount"`
	Turns                 []TurnDetail `json:"turns"`
	// Optional fields — populated when backend has the data.
	Source  string `json:"source,omitempty"`
	Status  string `json:"status,omitempty"`
	Project string `json:"project,omitempty"`
	// Model seeds sticky model resolution for the root assistant before the first
	// valid turn observation. It is the earliest valid root-assistant model in
	// canonical order when one is available, otherwise legacy stored metadata. It
	// is never the latest, most common, or final session model.
	Model string `json:"model,omitempty"`
	// ThinkingLevel seeds sticky level resolution for the root assistant before
	// the first valid turn observation: the earliest valid root-assistant
	// observation in source order when available, otherwise the stored session
	// seed; never latest, most common, or final. Raw rides the same rule.
	ThinkingLevel    ThinkingLevel    `json:"thinkingLevel,omitempty"`
	ThinkingLevelRaw ThinkingLevelRaw `json:"thinkingLevelRaw,omitempty"`
	// ThinkingLevelHistory records the in-session level changes observed after
	// the seed, in source order, with consecutive duplicates collapsed. The
	// first value stays in ThinkingLevel/ThinkingLevelRaw above; each entry here
	// carries the later canonical level and/or its exact native spelling.
	ThinkingLevelHistory []ThinkingLevelChange   `json:"thinkingLevelHistory,omitempty"`
	WorkingDirectory     string                  `json:"workingDirectory,omitempty"`
	GitBranch            string                  `json:"gitBranch,omitempty"`
	GitRemote            string                  `json:"gitRemote,omitempty"`
	ChildSessions        []ChildSessionRef       `json:"childSessions,omitempty"`
	ParentSessionID      *SessionID              `json:"parentSessionId,omitempty"`
	RootSessionID        *SessionID              `json:"rootSessionId,omitempty"`
	Purpose              SessionPurpose          `json:"purpose,omitempty"`
	Relationships        []SessionRelationship   `json:"relationships,omitempty"`
	EarlierHistory       []EarlierHistorySection `json:"earlierHistory,omitempty"`
	// Outcome is the heuristic resolution status of the session
	// (resolved/partial/failed), sourced from session_metrics.outcome. Empty
	// when the session has no computed outcome. The metadata columns expose no
	// per-signal reason, so no reason field accompanies it yet.
	Outcome SessionOutcome `json:"outcome,omitempty"`
	// SessionOrigin is the producer's declaration of who drove this session.
	// Absent when the producer expressed no opinion; see SessionOrigin for the
	// rule a consumer applies to an absent or "unknown" value.
	SessionOrigin SessionOrigin `json:"sessionOrigin,omitempty"`
	// Scorecard carries the per-session quality signals used by the "How this
	// session went" self-assessment card. Nil when the session has no computed
	// metrics. Sourced from the same session_metrics row that backs QualitySession.
	Scorecard       *SessionScorecard          `json:"scorecard,omitempty"`
	NativeMetadata  []NativeMetadataRecord     `json:"nativeMetadata,omitempty"`
	RetainedUnknown []RetainedUnknownRecord    `json:"retainedUnknown,omitempty" nullable:"false"`
	Diagnostics     *InterpretationDiagnostics `json:"diagnostics,omitempty" nullable:"false"`
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

// --- Transcript reads ---

// MaxTranscriptDirectoryRefs bounds one TranscriptSegmentDirectory read. The
// directory is the transcript's only global index, so the read is capped: above
// this many segments the server serves one page of this size and sets
// directoryPage/directoryPerPage. 10,000 segments is roughly 10 GiB of
// transcript, so the cap only exists to keep the directory read from becoming
// the unbounded payload the segment model removes.
const MaxTranscriptDirectoryRefs = 10000

// TranscriptSegmentDirectory is the transcript's global index: one ref per
// segment, in segment order, with the cut's scalars and no turn content. It is
// the only global index on the transcript path, so a client never receives a
// per-turn row to find its place.
type TranscriptSegmentDirectory struct {
	SessionID SessionID `json:"sessionId"`
	// Revision is the store revision the directory was cut under, as a decimal
	// string. A segment from a different revision must not be placed beside it.
	Revision string `json:"revision" minLength:"1"`
	// TotalTurns is the number of turns the session's main partition holds. It is
	// a count, never an ordinal: no position is derived from it.
	TotalTurns int `json:"totalTurns" minimum:"0" maximum:"9007199254740991"`
	// TotalConversationTurns is the number of conversation turns in the session's
	// main partition (IsConversationTurn). It is the total a position is stated
	// against, and the whole-session value on every page of refs. A producer that
	// cannot vouch for its position data omits it with every ref's
	// conversationTurns. Absence is never zero.
	TotalConversationTurns *int `json:"totalConversationTurns,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	// TotalSegments is the true number of segments, even when this read serves one
	// page of them.
	TotalSegments int                    `json:"totalSegments" minimum:"0" maximum:"9007199254740991"`
	Segments      []TranscriptSegmentRef `json:"segments" nullable:"false"`
	// DirectoryPage and DirectoryPerPage are set only when TotalSegments exceeds
	// MaxTranscriptDirectoryRefs; the read then serves one page of refs.
	DirectoryPage    *int `json:"directoryPage,omitempty" minimum:"1" maximum:"9007199254740991" nullable:"false"`
	DirectoryPerPage *int `json:"directoryPerPage,omitempty" minimum:"1" maximum:"9007199254740991" nullable:"false"`
}

// TranscriptSegmentRef is one segment's extent in the directory: its own
// ordinal and the closed range of sparse turn identities it holds. The ordinal
// is a segment namespace; firstTurn and lastTurn are turn identities, and no
// arithmetic converts one into the other.
type TranscriptSegmentRef struct {
	Index      int `json:"index" minimum:"0" maximum:"9007199254740991"`
	FirstTurn  int `json:"firstTurn" minimum:"0" maximum:"9007199254740991"`
	LastTurn   int `json:"lastTurn" minimum:"0" maximum:"9007199254740991"`
	ByteLength int `json:"byteLength" minimum:"0" maximum:"9007199254740991"`
	// ConversationTurns is where this segment sits in the session's conversation
	// turns. No arithmetic converts firstTurn or lastTurn into it.
	ConversationTurns *ConversationTurnSpan `json:"conversationTurns,omitempty" nullable:"false"`
}

// ConversationTurnSpan is a pair of counts, never a turn identity. A reader's
// range for the segment is Before+1 through Before+Count.
type ConversationTurnSpan struct {
	// Before is the number of conversation turns in every earlier segment of
	// the session, including segments on earlier directory pages. It starts at 0.
	Before int `json:"before" minimum:"0" maximum:"9007199254740991"`
	// Count is the number of conversation turns in this segment. It may be 0.
	Count int `json:"count" minimum:"0" maximum:"9007199254740991"`
}

// Validate checks the directory's ordering, extent, and page cap. The refs must
// be contiguous in segment ordinal, ascending in turn identity, and
// non-overlapping, because a viewer merges adjacent segments by their declared
// ranges.
func (d TranscriptSegmentDirectory) Validate() error {
	const where = "transcript segment directory validation failed at schema.TranscriptSegmentDirectory.Validate: "
	if !isDecimalRevision(d.Revision) {
		return fmt.Errorf("%srevision %q is not a decimal string; a segment cut under one revision must never be placed beside one cut under another; carry the store revision as decimal digits", where, d.Revision)
	}
	if d.TotalTurns < 0 || d.TotalSegments < 0 {
		return fmt.Errorf("%stotalTurns or totalSegments is negative; the read cannot describe a negative corpus; emit non-negative counts", where)
	}
	if d.Segments == nil {
		return fmt.Errorf("%ssegments is null; an empty session must answer with an empty array; emit a non-null array", where)
	}
	if len(d.Segments) > MaxTranscriptDirectoryRefs {
		return fmt.Errorf("%ssegments holds %d refs, above the %d-ref directory cap; the directory read would become the unbounded payload; serve one page and set directoryPage/directoryPerPage", where, len(d.Segments), MaxTranscriptDirectoryRefs)
	}
	paged := d.DirectoryPage != nil || d.DirectoryPerPage != nil
	start := 0
	if paged {
		if d.DirectoryPage == nil || d.DirectoryPerPage == nil {
			return fmt.Errorf("%sdirectoryPage and directoryPerPage must be set together; a half-specified page cannot be placed in the directory sequence; set both or neither", where)
		}
		if *d.DirectoryPage < 1 {
			return fmt.Errorf("%sdirectoryPage is %d; pages are 1-based; emit a positive page number", where, *d.DirectoryPage)
		}
		if *d.DirectoryPerPage != MaxTranscriptDirectoryRefs {
			return fmt.Errorf("%sdirectoryPerPage is %d, not the %d-ref directory cap; the page size is not caller-chosen; use MaxTranscriptDirectoryRefs", where, *d.DirectoryPerPage, MaxTranscriptDirectoryRefs)
		}
		if d.TotalSegments <= MaxTranscriptDirectoryRefs {
			return fmt.Errorf("%sthe directory declares page %d of %d with totalSegments %d; the page fields exist only above the %d-ref cap; serve the whole directory below it or omit directoryPage/directoryPerPage", where, *d.DirectoryPage, *d.DirectoryPerPage, d.TotalSegments, MaxTranscriptDirectoryRefs)
		}
		if len(d.Segments) > *d.DirectoryPerPage {
			return fmt.Errorf("%sserves %d refs on a %d-ref page; the declared page size is exceeded; serve at most directoryPerPage refs", where, len(d.Segments), *d.DirectoryPerPage)
		}
		if d.TotalSegments < len(d.Segments) {
			return fmt.Errorf("%stotalSegments %d is below the %d served refs; the true total cannot be smaller than a page of it; emit the whole-session total", where, d.TotalSegments, len(d.Segments))
		}
		page := int64(*d.DirectoryPage)
		perPage := int64(*d.DirectoryPerPage)
		if page-1 > int64(maxSafeJSONInteger)/perPage {
			return fmt.Errorf("%sdirectoryPage %d overflows the page offset at %d refs per page; the read cannot address a page that far out; emit a reachable one-based page", where, *d.DirectoryPage, perPage)
		}
		start = int((page - 1) * perPage)
	} else if d.TotalSegments != len(d.Segments) {
		return fmt.Errorf("%stotalSegments %d differs from the %d served refs; an unpaged directory must serve every segment; serve the whole directory or set directoryPage/directoryPerPage", where, d.TotalSegments, len(d.Segments))
	}
	if d.TotalTurns < len(d.Segments) {
		return fmt.Errorf("%stotalTurns %d is below the %d segments; every segment holds at least one turn; emit the whole-session turn count", where, d.TotalTurns, len(d.Segments))
	}
	for i, ref := range d.Segments {
		if ref.Index != start+i {
			return fmt.Errorf("%ssegment ref %d declares index %d, want %d; refs must be contiguous from the page's first ordinal; a gap would hide a segment from the viewer", where, i, ref.Index, start+i)
		}
		if ref.FirstTurn < 0 || ref.LastTurn < ref.FirstTurn {
			return fmt.Errorf("%ssegment %d declares turn range [%d, %d]; a segment covers a closed ascending range; emit firstTurn <= lastTurn", where, ref.Index, ref.FirstTurn, ref.LastTurn)
		}
		if ref.Index >= d.TotalSegments {
			return fmt.Errorf("%ssegment ref index %d is not below totalSegments %d; a page cannot name an ordinal the session does not have; emit the whole-session total or correct the page", where, ref.Index, d.TotalSegments)
		}
		if ref.ByteLength < 0 {
			return fmt.Errorf("%ssegment %d declares a negative byteLength; served bytes cannot be negative; emit a non-negative count", where, ref.Index)
		}
		if i > 0 && ref.FirstTurn <= d.Segments[i-1].LastTurn {
			return fmt.Errorf("%ssegment %d starts at turn %d, not after segment %d ending at %d; segment ranges must ascend without overlap or a turn would be served twice", where, ref.Index, ref.FirstTurn, d.Segments[i-1].Index, d.Segments[i-1].LastTurn)
		}
	}
	if err := validateDirectoryConversationTurns(d, where); err != nil {
		return err
	}
	return nil
}

func validateDirectoryConversationTurns(d TranscriptSegmentDirectory, where string) error {
	if d.TotalConversationTurns != nil && *d.TotalConversationTurns < 0 {
		return fmt.Errorf("%stotalConversationTurns is %d; a reader would print a negative total beside every position; emit the count of conversation turns or omit the field", where, *d.TotalConversationTurns)
	}
	for _, ref := range d.Segments {
		if ref.ConversationTurns == nil {
			continue
		}
		span := ref.ConversationTurns
		if span.Before < 0 || span.Count < 0 {
			return fmt.Errorf("%ssegment %d declares conversationTurns {before: %d, count: %d}; both are counts and cannot be negative, or a reader would print an impossible range; emit non-negative counts or omit conversationTurns on this ref", where, ref.Index, span.Before, span.Count)
		}
	}
	if d.TotalConversationTurns != nil {
		total := int64(*d.TotalConversationTurns)
		for _, ref := range d.Segments {
			if ref.ConversationTurns == nil {
				continue
			}
			end := int64(ref.ConversationTurns.Before) + int64(ref.ConversationTurns.Count)
			if end > total {
				return fmt.Errorf("%ssegment %d ends at conversation turn %d, past totalConversationTurns %d; a reader would print a position larger than the total; emit before + count <= totalConversationTurns or omit the position fields", where, ref.Index, end, total)
			}
		}
	}
	for i := 1; i < len(d.Segments); i++ {
		prev := d.Segments[i-1].ConversationTurns
		curr := d.Segments[i].ConversationTurns
		if prev == nil || curr == nil {
			continue
		}
		want := int64(prev.Before) + int64(prev.Count)
		if int64(curr.Before) != want {
			return fmt.Errorf("%ssegment %d declares %d conversation turns before it, but segment %d ends at %d; a turn would be numbered twice or skipped; emit before equal to the previous segment's before + count, or omit the position fields", where, d.Segments[i].Index, curr.Before, d.Segments[i-1].Index, want)
		}
	}
	return nil
}

// TranscriptSegment is one segment's turn content and its own facts. It is the
// transport unit of the transcript: at most the segment budget, or one oversized
// entry alone, so no response on the transcript path carries more than one
// segment's turns.
type TranscriptSegment struct {
	// Index is the segment ordinal, 0..totalSegments-1. It is not a turn identity.
	Index    int    `json:"index" minimum:"0" maximum:"9007199254740991"`
	Revision string `json:"revision" minLength:"1"`
	// FirstTurn and LastTurn are the closed range of sparse turn identities this
	// segment covers.
	FirstTurn int `json:"firstTurn" minimum:"0" maximum:"9007199254740991"`
	LastTurn  int `json:"lastTurn" minimum:"0" maximum:"9007199254740991"`
	// ByteLength is the sum of the served turns' ByteLength values: the
	// pre-bound size of their content plus every toolCalls[].arguments and
	// toolCalls[].result. It is not the marshalled size of the bounded rows; the
	// served bytes are at most this total, and the per-turn value is what the
	// truncation notice names.
	ByteLength int `json:"byteLength" minimum:"0" maximum:"9007199254740991"`
	// HasOlder and HasNewer say whether a boundary is drawn on that side. HasOlder
	// is index > 0; HasNewer is index < totalSegments-1. They are redundant with
	// the directory by design: a segment is renderable alone. A segment cannot
	// know the total, so the producer checks HasNewer with the directory
	// (ValidateTranscriptSegmentBoundaries).
	HasOlder bool             `json:"hasOlder"`
	HasNewer bool             `json:"hasNewer"`
	Turns    []TranscriptTurn `json:"turns" nullable:"false"`
}

// Validate checks that the segment's declared range is the exact range its
// turns occupy and that its declared byte length is the sum of the turn rows'
// byte lengths. A segment that declares a range or size its turns do not carry
// would silently shift every later boundary, so it is rejected rather than
// repaired.
func (s TranscriptSegment) Validate() error {
	const where = "transcript segment validation failed at schema.TranscriptSegment.Validate: "
	if s.Index < 0 {
		return fmt.Errorf("%sindex is %d; segment ordinals start at zero; emit the segment's own ordinal", where, s.Index)
	}
	if !isDecimalRevision(s.Revision) {
		return fmt.Errorf("%srevision %q is not a decimal string; a segment cut under one revision must never be placed beside one cut under another; carry the store revision as decimal digits", where, s.Revision)
	}
	if s.FirstTurn < 0 || s.LastTurn < s.FirstTurn {
		return fmt.Errorf("%sdeclares turn range [%d, %d]; a segment covers a closed ascending range; emit firstTurn <= lastTurn", where, s.FirstTurn, s.LastTurn)
	}
	if s.ByteLength < 0 {
		return fmt.Errorf("%sbyteLength is negative; served bytes cannot be negative; emit a non-negative count", where)
	}
	if s.HasOlder != (s.Index > 0) {
		return fmt.Errorf("%shasOlder is %v for segment %d; hasOlder is exactly index > 0; a boundary on the wrong side misstates whether the document continues", where, s.HasOlder, s.Index)
	}
	if s.Turns == nil {
		return fmt.Errorf("%sturns is null; a segment always holds at least one whole turn; emit a non-null array", where)
	}
	if len(s.Turns) == 0 {
		return fmt.Errorf("%sturns is empty; a segment always holds at least one whole turn; omit the segment instead", where)
	}
	previous := s.FirstTurn - 1
	var turnBytes int64
	for i := range s.Turns {
		turn := s.Turns[i].Index
		if turn < s.FirstTurn || turn > s.LastTurn {
			return fmt.Errorf("%sturn %d lies outside the declared range [%d, %d]; a turn must belong to exactly one segment range; correct the range or move the turn", where, turn, s.FirstTurn, s.LastTurn)
		}
		if turn <= previous {
			return fmt.Errorf("%sturn %d does not follow %d; segment turns must be in ascending identity order; an out-of-order turn cannot be rendered as part of the document", where, turn, previous)
		}
		if s.Turns[i].ByteLength < 0 {
			return fmt.Errorf("%sturn %d declares a negative byteLength; a turn's pre-bound size cannot be negative; emit a non-negative count", where, turn)
		}
		turnBytes += int64(s.Turns[i].ByteLength)
		previous = turn
	}
	if s.Turns[0].Index != s.FirstTurn || s.Turns[len(s.Turns)-1].Index != s.LastTurn {
		return fmt.Errorf("%sdeclares range [%d, %d] but its turns span [%d, %d]; the declared range must be exactly the turns served; correct firstTurn/lastTurn", where, s.FirstTurn, s.LastTurn, s.Turns[0].Index, s.Turns[len(s.Turns)-1].Index)
	}
	if int64(s.ByteLength) != turnBytes {
		return fmt.Errorf("%sdeclares byteLength %d but its turns total %d; the segment's byte length is the sum of its turn rows' byteLength values, so a client sizing a boundary would be told the wrong cost; emit the summed value", where, s.ByteLength, turnBytes)
	}
	return nil
}

// ValidateTranscriptSegmentBoundaries checks a segment's edge flags against the
// directory's segment count: HasOlder is Index > 0 and HasNewer is
// Index < totalSegments-1. A segment alone cannot know the total, so the
// producer calls this with the directory it built.
func ValidateTranscriptSegmentBoundaries(segment TranscriptSegment, totalSegments int) error {
	const where = "transcript segment boundary validation failed at schema.ValidateTranscriptSegmentBoundaries: "
	if totalSegments < 1 {
		return fmt.Errorf("%stotalSegments is %d; a served segment belongs to at least one segment; pass the directory's segment count", where, totalSegments)
	}
	if segment.Index < 0 || segment.Index >= totalSegments {
		return fmt.Errorf("%ssegment index %d is outside 0..%d; the segment does not belong to this directory; pass the ordinal from the directory the segment was cut under", where, segment.Index, totalSegments-1)
	}
	if want := segment.Index > 0; segment.HasOlder != want {
		return fmt.Errorf("%shasOlder is %v for segment %d of %d; hasOlder is exactly index > 0; a boundary on the wrong side misstates whether the document continues", where, segment.HasOlder, segment.Index, totalSegments)
	}
	if want := segment.Index < totalSegments-1; segment.HasNewer != want {
		return fmt.Errorf("%shasNewer is %v for segment %d of %d; hasNewer is exactly index < totalSegments-1; a boundary on the wrong side misstates whether the document continues", where, segment.HasNewer, segment.Index, totalSegments)
	}
	return nil
}

// TranscriptTurn is one served turn: the single TurnDetail definition plus the
// read's own facts. Nothing is copied flat, so a segment and the single-turn
// continuation read share one projection and one truncation story.
type TranscriptTurn struct {
	TurnDetail
	// Truncated reports that this turn's own text was bounded by the served-text
	// bound. ByteLength is the pre-bound size of content plus every
	// toolCalls[].arguments and toolCalls[].result: the size the turn had before
	// the bound was applied, so the truncation notice can name what the full read
	// will return. A segment's ByteLength is the sum of these values; the served
	// bytes of a truncated turn are at most its ByteLength.
	Truncated  bool `json:"truncated"`
	ByteLength int  `json:"byteLength" minimum:"0" maximum:"9007199254740991"`
}

// TranscriptWindow is the descriptor of an anchor window: where the window is,
// never what is in it. It carries no turn content, so a jump can never be the
// oversized payload. A viewer loads the covering segments through the segment
// read.
type TranscriptWindow struct {
	SessionID SessionID `json:"sessionId"`
	Revision  string    `json:"revision" minLength:"1"`
	// RequestedAnchor is the turn identity the caller asked for. It differs from
	// AnchorTurn when the read snapped to a user turn, and the viewer says so.
	RequestedAnchor int `json:"requestedAnchor" minimum:"0" maximum:"9007199254740991"`
	// AnchorTurn is the resolved user turn the window is centred on. It always
	// lies inside [firstTurn, lastTurn].
	AnchorTurn int `json:"anchorTurn" minimum:"0" maximum:"9007199254740991"`
	// Span is the number of assistant turns collected on each side. It is
	// validated 1..10: the design value narrows a window and cannot widen it past
	// the design.
	Span      int `json:"span" minimum:"1" maximum:"10"`
	FirstTurn int `json:"firstTurn" minimum:"0" maximum:"9007199254740991"`
	LastTurn  int `json:"lastTurn" minimum:"0" maximum:"9007199254740991"`
	// Segments is the ordered, contiguous run of segment ordinals that covers
	// [firstTurn, lastTurn]. The viewer loads exactly these.
	Segments []int `json:"segments" nullable:"false"`
	HasOlder bool  `json:"hasOlder"`
	HasNewer bool  `json:"hasNewer"`
}

// Validate checks the descriptor's own invariants: a positive span within the
// design bound, an anchor inside the window, and an ascending contiguous segment
// run. Coverage of [firstTurn, lastTurn] is checked against the directory by
// ValidateTranscriptWindowCoverage, because the descriptor deliberately carries
// no refs.
func (w TranscriptWindow) Validate() error {
	const where = "transcript window validation failed at schema.TranscriptWindow.Validate: "
	if !isDecimalRevision(w.Revision) {
		return fmt.Errorf("%srevision %q is not a decimal string; a window must name the cut its covering segments belong to; carry the store revision as decimal digits", where, w.Revision)
	}
	if w.Span < 1 || w.Span > 10 {
		return fmt.Errorf("%sspan is %d; the accepted range is 1..10 assistant turns per side; emit a value in that range", where, w.Span)
	}
	if w.FirstTurn < 0 || w.LastTurn < w.FirstTurn {
		return fmt.Errorf("%sdeclares turn range [%d, %d]; a window covers a closed ascending range; emit firstTurn <= lastTurn", where, w.FirstTurn, w.LastTurn)
	}
	if w.AnchorTurn < w.FirstTurn || w.AnchorTurn > w.LastTurn {
		return fmt.Errorf("%sanchorTurn %d lies outside [%d, %d]; the window must contain the user turn it is centred on; correct the extent", where, w.AnchorTurn, w.FirstTurn, w.LastTurn)
	}
	if w.Segments == nil || len(w.Segments) == 0 {
		return fmt.Errorf("%ssegments is empty or null; a window always names at least the segment covering its anchor; emit the ordered covering segment run", where)
	}
	for i, segment := range w.Segments {
		if segment < 0 {
			return fmt.Errorf("%ssegments[%d] is %d; a segment ordinal is non-negative; emit directory ordinals", where, i, segment)
		}
		if i > 0 && segment != w.Segments[i-1]+1 {
			return fmt.Errorf("%ssegments[%d] is %d after %d; the covering run must be contiguous and ascending, or the rendered window would have a gap inside it; emit every segment between the two edges", where, i, segment, w.Segments[i-1])
		}
	}
	return nil
}

// ValidateTranscriptWindowCoverage checks that the window's segment run is
// exactly the directory's run from the segment containing firstTurn to the one
// containing lastTurn. The descriptor carries no refs, so this check needs the
// directory the window was resolved against.
func ValidateTranscriptWindowCoverage(window TranscriptWindow, refs []TranscriptSegmentRef) error {
	const where = "transcript window coverage validation failed at schema.ValidateTranscriptWindowCoverage: "
	if len(refs) == 0 {
		return fmt.Errorf("%sthe directory holds no segments; a window cannot cover a turn; build the directory before resolving a window", where)
	}
	first, last := -1, -1
	for _, ref := range refs {
		if ref.FirstTurn <= window.FirstTurn && window.FirstTurn <= ref.LastTurn {
			first = ref.Index
		}
		if ref.FirstTurn <= window.LastTurn && window.LastTurn <= ref.LastTurn {
			last = ref.Index
		}
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("%swindow [%d, %d] is not covered by the directory's segment ranges; a viewer would load no segment for one edge; resolve the window against this directory", where, window.FirstTurn, window.LastTurn)
	}
	want := make([]int, 0, last-first+1)
	for segment := first; segment <= last; segment++ {
		want = append(want, segment)
	}
	if len(want) != len(window.Segments) {
		return fmt.Errorf("%swindow names %d segments, but covering [%d, %d] takes %d; the viewer would leave part of the window unloaded", where, len(window.Segments), window.FirstTurn, window.LastTurn, len(want))
	}
	for i := range want {
		if window.Segments[i] != want[i] {
			return fmt.Errorf("%swindow segments[%d] is %d, want %d; the descriptor must name the contiguous run that covers [%d, %d]", where, i, window.Segments[i], want[i], window.FirstTurn, window.LastTurn)
		}
	}
	return nil
}

// SessionDetailHeader is the metadata-only projection of a session: everything
// the durable SessionDetailPayload states about the session except its turns.
// It is a distinct type, not a replacement for the durable payload, so no
// release window ships "turns: []" to a content-expecting viewer. Turn content
// lives in TranscriptSegment rows.
type SessionDetailHeader struct {
	SchemaVersion        PushContractVersion `json:"schemaVersion,omitempty"`
	ID                   string              `json:"id"`
	Harness              Harness             `json:"harness"`
	StartTime            time.Time           `json:"startTime"`
	EndTime              time.Time           `json:"endTime"`
	DurationMins         float64             `json:"durationMins"`
	TotalTokens          int                 `json:"totalTokens"`
	TokensIn             int                 `json:"tokensIn"`
	TokensOut            int                 `json:"tokensOut"`
	TurnCount            int                 `json:"turnCount" description:"Turns: the number of folded turns of the session's main partition, system and tool rows included. Equal to the segment directory's totalTurns at the same revision. Not the count a reader is shown; see conversationTurnCount."`
	InputSubmissionCount *int64              `json:"inputSubmissionCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	// ConversationTurnCount is the number of conversation turns in the session's
	// main partition (IsConversationTurn): the count a reader is shown. Absent
	// means unmeasured: a producer or publication older than this field. It
	// equals the segment directory's totalConversationTurns at the same revision.
	ConversationTurnCount *int64                `json:"conversationTurnCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	ToolCallCount         int                   `json:"toolCallCount"`
	Source                string                `json:"source,omitempty"`
	Status                string                `json:"status,omitempty"`
	Project               string                `json:"project,omitempty"`
	Model                 string                `json:"model,omitempty"`
	ThinkingLevel         ThinkingLevel         `json:"thinkingLevel,omitempty"`
	ThinkingLevelRaw      ThinkingLevelRaw      `json:"thinkingLevelRaw,omitempty"`
	ThinkingLevelHistory  []ThinkingLevelChange `json:"thinkingLevelHistory,omitempty"`
	WorkingDirectory      string                `json:"workingDirectory,omitempty"`
	GitBranch             string                `json:"gitBranch,omitempty"`
	GitRemote             string                `json:"gitRemote,omitempty"`
	ChildSessions         []ChildSessionRef     `json:"childSessions,omitempty"`
	ParentSessionID       *SessionID            `json:"parentSessionId,omitempty"`
	RootSessionID         *SessionID            `json:"rootSessionId,omitempty"`
	Purpose               SessionPurpose        `json:"purpose,omitempty"`
	Relationships         []SessionRelationship `json:"relationships,omitempty"`
	// EarlierHistory carries section state only. Turns and the native metadata
	// that annotates them travel together on the durable read, and neither belongs
	// on a metadata-only projection: the section's native metadata is validated
	// against the turns it annotates.
	EarlierHistory  []EarlierHistoryState      `json:"earlierHistory,omitempty" nullable:"false"`
	Outcome         SessionOutcome             `json:"outcome,omitempty"`
	SessionOrigin   SessionOrigin              `json:"sessionOrigin,omitempty"`
	Scorecard       *SessionScorecard          `json:"scorecard,omitempty"`
	RetainedUnknown []RetainedUnknownRecord    `json:"retainedUnknown,omitempty" nullable:"false"`
	Diagnostics     *InterpretationDiagnostics `json:"diagnostics,omitempty" nullable:"false"`
}

// Validate checks the header's metadata against the same shared rules the
// durable payload applies where those rules do not need turns.
func (h SessionDetailHeader) Validate() error {
	if err := ValidateSessionDetailPayload(h.metadataPayload()); err != nil {
		return err
	}
	for i, state := range h.EarlierHistory {
		if !state.IsValid() {
			return fmt.Errorf("session detail header validation failed at schema.SessionDetailHeader.Validate: earlierHistory[%d] state %q is outside the closed set (schema.EarlierHistoryState); consumers cannot explain retained history; use a published state", i, state)
		}
	}
	return nil
}

// metadataPayload projects the header onto the durable payload's metadata-only
// shape so one validation rule serves both. Turns, earlier-history turns, and
// native metadata stay empty: they are not part of the header.
func (h SessionDetailHeader) metadataPayload() SessionDetailPayload {
	return SessionDetailPayload{
		SchemaVersion:         h.SchemaVersion,
		ID:                    h.ID,
		Harness:               h.Harness,
		StartTime:             h.StartTime,
		EndTime:               h.EndTime,
		DurationMins:          h.DurationMins,
		TotalTokens:           h.TotalTokens,
		TokensIn:              h.TokensIn,
		TokensOut:             h.TokensOut,
		TurnCount:             h.TurnCount,
		InputSubmissionCount:  h.InputSubmissionCount,
		ConversationTurnCount: h.ConversationTurnCount,
		ToolCallCount:         h.ToolCallCount,
		Source:                h.Source,
		Status:                h.Status,
		Project:               h.Project,
		Model:                 h.Model,
		ThinkingLevel:         h.ThinkingLevel,
		ThinkingLevelRaw:      h.ThinkingLevelRaw,
		ThinkingLevelHistory:  h.ThinkingLevelHistory,
		WorkingDirectory:      h.WorkingDirectory,
		GitBranch:             h.GitBranch,
		GitRemote:             h.GitRemote,
		ChildSessions:         h.ChildSessions,
		ParentSessionID:       h.ParentSessionID,
		RootSessionID:         h.RootSessionID,
		Purpose:               h.Purpose,
		Relationships:         h.Relationships,
		Outcome:               h.Outcome,
		SessionOrigin:         h.SessionOrigin,
		Scorecard:             h.Scorecard,
		RetainedUnknown:       h.RetainedUnknown,
		Diagnostics:           h.Diagnostics,
	}
}

// SessionDetailHeaderPayload is the session_detail frame body: the metadata
// header plus the local navigation and teammate evidence the durable push never
// carries.
type SessionDetailHeaderPayload struct {
	SessionDetailHeader
	RelationshipNavigation []SessionRelationshipNavigation `json:"relationshipNavigation,omitempty"`
	Teammate               *SessionTeammate                `json:"teammate,omitempty"`
}

// Validate checks the header, the authorized navigation metadata, and the
// optional local teammate evidence.
func (p SessionDetailHeaderPayload) Validate() error {
	if err := p.SessionDetailHeader.Validate(); err != nil {
		return err
	}
	for _, navigation := range p.RelationshipNavigation {
		if err := navigation.Validate(); err != nil {
			return err
		}
	}
	return p.Teammate.Validate()
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
	ID           string `json:"id"`
	Date         string `json:"date"` // "2006-01-02" format
	Project      string `json:"project"`
	TotalTokens  int    `json:"totalTokens"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
	TurnCount    int    `json:"turnCount" description:"Turns: the number of folded turns of the session's main partition, system and tool rows included. Equal to the segment directory's totalTurns at the same revision. Not the count a reader is shown; see conversationTurnCount."`
	// ConversationTurnCount is the number of conversation turns in the session's
	// main partition (IsConversationTurn): the count a reader is shown. Absent
	// means unmeasured: a producer or publication older than this field. It
	// equals the segment directory's totalConversationTurns at the same revision.
	ConversationTurnCount *int64  `json:"conversationTurnCount,omitempty" minimum:"0" maximum:"9007199254740991" nullable:"false"`
	ToolCalls             int     `json:"toolCalls"`
	DurationMinutes       float64 `json:"durationMinutes"`
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

// PersonalMedians is the library-wide baseline the transcript scorecard
// compares one session against. Every field is optional: a field is absent when
// its sample yields no value, so the comparison line degrades instead of
// fabricating a baseline. The YAML tags let the canonical medians fixture carry
// this exact shape, so the Go aggregate and the browser baseline read one corpus.
type PersonalMedians struct {
	// RetryShare is the median retryTokensWasted/totalTokens ratio over the
	// sessions whose totalTokens is positive.
	RetryShare *float64 `json:"retryShare,omitempty" yaml:"retryShare,omitempty"`
	// SpecQualityScore is the median spec quality score over the sessions whose
	// score is positive.
	SpecQualityScore *float64 `json:"specQualityScore,omitempty" yaml:"specQualityScore,omitempty"`
	// WithinSessionReverts is the median revert count over the sessions whose
	// count is nonnegative.
	WithinSessionReverts *float64 `json:"withinSessionReverts,omitempty" yaml:"withinSessionReverts,omitempty"`
}

// ProjectAnalytics is the analytics bundle the project dashboard renders. It is
// a schema type, not a presentation typedef: the server aggregate, the generated
// TypeScript type, and the dashboard typedef are one contract with these exact
// field names.
type ProjectAnalytics struct {
	TotalSessions            int                      `json:"totalSessions" minimum:"0" maximum:"9007199254740991"`
	TotalContributors        int                      `json:"totalContributors" minimum:"0" maximum:"9007199254740991"`
	TotalProjects            int                      `json:"totalProjects" minimum:"0" maximum:"9007199254740991"`
	SessionsPerWeek          []WeekCount              `json:"sessionsPerWeek" nullable:"false"`
	WeeklyActiveContributors []WeekContributors       `json:"weeklyActiveContributors" nullable:"false"`
	ReturningContributorRate ReturningContributorRate `json:"returningContributorRate"`
	LongestStreak            LongestStreak            `json:"longestStreak"`
	NewContributorVelocity   []WeekNewContributors    `json:"newContributorVelocity" nullable:"false"`
	SessionToCommitRate      SessionToCommitRate      `json:"sessionToCommitRate"`
	AvgDurationPerActiveWeek []WeekAvgDuration        `json:"avgDurationPerActiveWeek" nullable:"false"`
	OutcomeDistribution      OutcomeDistribution      `json:"outcomeDistribution"`
	SessionStats             AnalyticsSessionStats    `json:"sessionStats"`
	PerContributorBreakdown  []ContributorBreakdown   `json:"perContributorBreakdown" nullable:"false"`
}

// WeekCount is one bucket of the per-week session-count series. Week is the
// Monday-of-week UTC date key (YYYY-MM-DD).
type WeekCount struct {
	Week  string `json:"week"`
	Count int    `json:"count" minimum:"0" maximum:"9007199254740991"`
}

// WeekContributors is one bucket of the per-week unique-contributor series.
type WeekContributors struct {
	Week         string `json:"week"`
	Contributors int    `json:"contributors" minimum:"0" maximum:"9007199254740991"`
}

// ReturningContributorRate is the returning-contributor result. A returning
// contributor is active in at least two distinct weeks; Rate is Returning/Total
// in [0, 1], and 0 when no contributor exists.
type ReturningContributorRate struct {
	Total     int     `json:"total" minimum:"0" maximum:"9007199254740991"`
	Returning int     `json:"returning" minimum:"0" maximum:"9007199254740991"`
	Rate      float64 `json:"rate" minimum:"0" maximum:"1"`
}

// LongestStreak is the longest run of consecutive active weeks, with its first
// and last Monday-of-week keys. Both keys are null when no week is active.
type LongestStreak struct {
	Weeks     int     `json:"weeks" minimum:"0" maximum:"9007199254740991"`
	StartWeek *string `json:"startWeek"`
	EndWeek   *string `json:"endWeek"`
}

// WeekNewContributors is one bucket of the new-contributor series: the
// contributors whose first-ever session falls in that week.
type WeekNewContributors struct {
	Week            string `json:"week"`
	NewContributors int    `json:"newContributors" minimum:"0" maximum:"9007199254740991"`
}

// SessionToCommitRate is the session-to-commit result. Rate is WithCommit/Total
// in [0, 1], and 0 when no session exists.
type SessionToCommitRate struct {
	Total      int     `json:"total" minimum:"0" maximum:"9007199254740991"`
	WithCommit int     `json:"withCommit" minimum:"0" maximum:"9007199254740991"`
	Rate       float64 `json:"rate" minimum:"0" maximum:"1"`
}

// WeekAvgDuration is one bucket of the per-active-week average-duration series.
type WeekAvgDuration struct {
	Week            string  `json:"week"`
	AvgDurationMins float64 `json:"avgDurationMins" minimum:"0"`
	Sessions        int     `json:"sessions" minimum:"0" maximum:"9007199254740991"`
}

// OutcomeDistribution is the session-outcome tally. Total counts every session,
// including the ones whose outcome is unknown.
type OutcomeDistribution struct {
	Resolved int `json:"resolved" minimum:"0" maximum:"9007199254740991"`
	Partial  int `json:"partial" minimum:"0" maximum:"9007199254740991"`
	Failed   int `json:"failed" minimum:"0" maximum:"9007199254740991"`
	Unknown  int `json:"unknown" minimum:"0" maximum:"9007199254740991"`
	Total    int `json:"total" minimum:"0" maximum:"9007199254740991"`
}

// MedianP90 is one typical-value pair: the sample count, its median, and its
// ninetieth percentile. Median and P90 are null when the sample is empty.
type MedianP90 struct {
	Count  int      `json:"count" minimum:"0" maximum:"9007199254740991"`
	Median *float64 `json:"median"`
	P90    *float64 `json:"p90"`
}

// AnalyticsSessionStats is the analytics aggregate over the library. It is not
// the ingest SessionStats: the name keeps the two distinct in one package.
// TurnCount is the median of conversationTurnCount over the library.
type AnalyticsSessionStats struct {
	DurationMins  MedianP90 `json:"durationMins"`
	TotalTokens   MedianP90 `json:"totalTokens"`
	TurnCount     MedianP90 `json:"turnCount" description:"Median conversationTurnCount over the library."`
	ToolCallCount MedianP90 `json:"toolCallCount"`
}

// ContributorBreakdown is one contributor's tally inside the analytics bundle.
// FirstSeen and LastSeen are day keys (YYYY-MM-DD), null when the contributor
// has no dated session.
type ContributorBreakdown struct {
	ContributorID      string              `json:"contributorId"`
	Sessions           int                 `json:"sessions" minimum:"0" maximum:"9007199254740991"`
	ActiveWeeks        int                 `json:"activeWeeks" minimum:"0" maximum:"9007199254740991"`
	TotalTokens        int                 `json:"totalTokens" minimum:"0" maximum:"9007199254740991"`
	TotalDurationMins  int                 `json:"totalDurationMins" minimum:"0" maximum:"9007199254740991"`
	TotalTurns         int                 `json:"totalTurns" minimum:"0" maximum:"9007199254740991"`
	TotalToolCalls     int                 `json:"totalToolCalls" minimum:"0" maximum:"9007199254740991"`
	SessionsWithCommit int                 `json:"sessionsWithCommit" minimum:"0" maximum:"9007199254740991"`
	TotalCommits       int                 `json:"totalCommits" minimum:"0" maximum:"9007199254740991"`
	Outcomes           OutcomeDistribution `json:"outcomes"`
	FirstSeen          *string             `json:"firstSeen"`
	LastSeen           *string             `json:"lastSeen"`
}

// Validate checks the aggregate's rates and counts. The Types schema bounds the
// same fields; this is the Go boundary for a producer that constructs the
// bundle directly, so a negative count or an out-of-range rate never reaches a
// client.
func (a ProjectAnalytics) Validate() error {
	const where = "project analytics validation failed at schema.ProjectAnalytics.Validate: "
	for _, rate := range []struct {
		name  string
		value float64
	}{
		{"returningContributorRate.rate", a.ReturningContributorRate.Rate},
		{"sessionToCommitRate.rate", a.SessionToCommitRate.Rate},
	} {
		if math.IsNaN(rate.value) || math.IsInf(rate.value, 0) || rate.value < 0 || rate.value > 1 {
			return fmt.Errorf("%s%s is %v; a rate is a finite share in [0, 1]; divide the counted part by the counted total", where, rate.name, rate.value)
		}
	}
	for _, count := range []struct {
		name  string
		value int
	}{
		{"totalSessions", a.TotalSessions},
		{"totalContributors", a.TotalContributors},
		{"totalProjects", a.TotalProjects},
		{"outcomeDistribution.resolved", a.OutcomeDistribution.Resolved},
		{"outcomeDistribution.partial", a.OutcomeDistribution.Partial},
		{"outcomeDistribution.failed", a.OutcomeDistribution.Failed},
		{"outcomeDistribution.unknown", a.OutcomeDistribution.Unknown},
		{"outcomeDistribution.total", a.OutcomeDistribution.Total},
		{"sessionStats.durationMins.count", a.SessionStats.DurationMins.Count},
		{"sessionStats.totalTokens.count", a.SessionStats.TotalTokens.Count},
		{"sessionStats.turnCount.count", a.SessionStats.TurnCount.Count},
		{"sessionStats.toolCallCount.count", a.SessionStats.ToolCallCount.Count},
	} {
		if count.value < 0 || int64(count.value) > maxSafeJSONInteger {
			return fmt.Errorf("%s%s is %d; an analytics count is a non-negative safe integer; emit the counted value", where, count.name, count.value)
		}
	}
	for i, contributor := range a.PerContributorBreakdown {
		if contributor.Sessions < 0 || contributor.ActiveWeeks < 0 || contributor.TotalTokens < 0 || contributor.TotalDurationMins < 0 || contributor.TotalTurns < 0 || contributor.TotalToolCalls < 0 || contributor.SessionsWithCommit < 0 || contributor.TotalCommits < 0 {
			return fmt.Errorf("%sperContributorBreakdown[%d] carries a negative count; contributors are counted, never subtracted; emit non-negative totals", where, i)
		}
	}
	return nil
}

// QualityPage is the quality read for one offset page of sessions. The summary
// is the full analytics bundle computed over the library at the page revision;
// the rows are the requested page, so no surface subscribes to the full library.
type QualityPage struct {
	OffsetPage[QualitySession]
	Summary ProjectAnalytics `json:"summary"`
}

// Validate checks the common offset-page contract for the quality rows and the
// analytics summary carried beside them.
func (p QualityPage) Validate() error {
	if err := p.OffsetPage.Validate(); err != nil {
		return err
	}
	return p.Summary.Validate()
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

// AllInteractionTypes is the canonical list of project familiarity interactions.
var AllInteractionTypes = []InteractionType{
	InteractionMentioned,
	InteractionRead,
	InteractionDiscussed,
	InteractionQuestioned,
}

// IsValid reports whether i is a defined interaction type.
func (i InteractionType) IsValid() bool {
	switch i {
	case InteractionMentioned, InteractionRead, InteractionDiscussed, InteractionQuestioned:
		return true
	}
	return false
}

func (i InteractionType) String() string { return string(i) }

// JSONSchema implements jsonschema.Exposer.
func (InteractionType) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Interaction Type",
		"How deeply a session engaged with a project file",
		AllInteractionTypes,
	), nil
}

// DecayLevel classifies how recently a file was engaged.
type DecayLevel string

const (
	DecayFresh      DecayLevel = "fresh"
	DecayFading     DecayLevel = "fading"
	DecayStale      DecayLevel = "stale"
	DecayUnexplored DecayLevel = "unexplored"
)

// AllDecayLevels is the canonical list of project familiarity decay levels.
var AllDecayLevels = []DecayLevel{
	DecayFresh,
	DecayFading,
	DecayStale,
	DecayUnexplored,
}

// IsValid reports whether d is a defined familiarity decay level.
func (d DecayLevel) IsValid() bool {
	switch d {
	case DecayFresh, DecayFading, DecayStale, DecayUnexplored:
		return true
	}
	return false
}

func (d DecayLevel) String() string { return string(d) }

// JSONSchema implements jsonschema.Exposer.
func (DecayLevel) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Decay Level",
		"Recency band for project file familiarity",
		AllDecayLevels,
	), nil
}

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
	TurnCount  int               `json:"turnCount" description:"Trail size as the producer computes it; not a session turn count."`
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
	ProjectHash     ProjectHash        `json:"projectHash"`
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

// UICapabilitiesResponse reports optional UI behavior enabled for the running
// local Peasant server. Capability token meanings and lifecycle are owned by
// Peasant, not by the schema module. Omitted and empty both mean no optional
// capabilities; consumers compare known tokens by exact membership and ignore
// unknown tokens. Producers emit tokens in sorted order without duplicates.
type UICapabilitiesResponse struct {
	UICapabilities []string `json:"uiCapabilities,omitempty" description:"Set of opaque UI capability tokens enabled for this running Peasant server process. Omitted and empty mean no optional capabilities. Ordering is insignificant. Consumers must ignore unknown tokens and compare known tokens by exact membership. Token meanings and lifecycle are owned by Peasant. Peasant producers emit tokens in sorted order without duplicates." nullable:"false"`
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
	MsgKickstart          MessageType = "kickstart"
)

// AllMessageTypes is the canonical list of WebSocket message discriminators.
var AllMessageTypes = []MessageType{
	MsgSubscribe,
	MsgUnsubscribe,
	MsgDashboard,
	MsgSessions,
	MsgSessionDetail,
	MsgTrends,
	MsgQuality,
	MsgAnnotations,
	MsgProjectFamiliarity,
	MsgConnected,
	MsgError,
	MsgKickstart,
}

// IsValid reports whether m is a defined WebSocket message discriminator.
func (m MessageType) IsValid() bool {
	for _, candidate := range AllMessageTypes {
		if m == candidate {
			return true
		}
	}
	return false
}

// String returns the wire representation of the message discriminator.
func (m MessageType) String() string { return string(m) }

// JSONSchema implements jsonschema.Exposer.
func (MessageType) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Message Type",
		"WebSocket message discriminator",
		AllMessageTypes,
	), nil
}

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
	TopicKickstart          ChannelTopic = "kickstart"
)

// AllChannelTopics is the canonical list of subscribable WebSocket topics.
var AllChannelTopics = []ChannelTopic{
	TopicDashboard,
	TopicSessions,
	TopicSessionDetail,
	TopicTrends,
	TopicQuality,
	TopicAnnotations,
	TopicProjectFamiliarity,
	TopicKickstart,
}

// IsValid reports whether t is a defined WebSocket channel topic.
func (t ChannelTopic) IsValid() bool {
	for _, candidate := range AllChannelTopics {
		if t == candidate {
			return true
		}
	}
	return false
}

// String returns the wire representation of the channel topic.
func (t ChannelTopic) String() string { return string(t) }

// JSONSchema implements jsonschema.Exposer.
func (ChannelTopic) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Channel Topic",
		"Subscribable WebSocket data stream",
		AllChannelTopics,
	), nil
}

// AnnotationAxis is the subscription dimension for annotation channels.
type AnnotationAxis string

const (
	AxisType    AnnotationAxis = "type"
	AxisSession AnnotationAxis = "session"
	AxisProject AnnotationAxis = "project"
)

// AllAnnotationAxes is the canonical list of annotation subscription axes.
var AllAnnotationAxes = []AnnotationAxis{
	AxisType,
	AxisSession,
	AxisProject,
}

// IsValid reports whether a is a defined annotation subscription axis.
func (a AnnotationAxis) IsValid() bool {
	switch a {
	case AxisType, AxisSession, AxisProject:
		return true
	}
	return false
}

// String returns the wire representation of the annotation axis.
func (a AnnotationAxis) String() string { return string(a) }

// JSONSchema implements jsonschema.Exposer.
func (AnnotationAxis) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Annotation Axis",
		"Subscription dimension for annotation channels",
		AllAnnotationAxes,
	), nil
}

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
	Type    MessageType    `json:"type"`
	Data    any            `json:"data,omitempty"`
	Message string         `json:"message,omitempty"` // for error type
	Version string         `json:"version,omitempty"` // for connected type
	Topic   ChannelTopic   `json:"topic,omitempty"`   // failed subscription topic for error type
	ID      string         `json:"id,omitempty"`      // failed topic identity, when applicable
	Axis    AnnotationAxis `json:"axis,omitempty"`    // failed annotation axis, when applicable
}
