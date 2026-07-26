package schema

import (
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// Map / Review wire contract.
//
// These payloads back the five REST endpoints of the Map and Review surfaces
// (contract §3) and are mirrored byte-for-byte by web/src/types/map.ts. All
// JSON field names are camelCase, all timestamps are Unix-millisecond int64
// fields suffixed "Ms" (matching the session_entries convention), and all
// slices serialize as [] — never null. Use the New*Payload constructors (or
// otherwise guarantee non-nil slices) before marshaling.

// --- Map graph ---

// MapNodeKind classifies a map node within the path-derived tree.
type MapNodeKind string

// MapNodeKind values. Top-level directories are modules, nested directories
// are packages, and files are leaves (mirrors codegraph's node kinds).
const (
	MapNodeKindModule  MapNodeKind = "module"
	MapNodeKindPackage MapNodeKind = "package"
	MapNodeKindFile    MapNodeKind = "file"
)

// AllMapNodeKinds is the canonical list of map node classifications.
var AllMapNodeKinds = []MapNodeKind{
	MapNodeKindModule,
	MapNodeKindPackage,
	MapNodeKindFile,
}

// String returns the wire representation of the node kind.
func (k MapNodeKind) String() string { return string(k) }

// IsValid reports whether k is one of the defined node kinds.
func (k MapNodeKind) IsValid() bool {
	switch k {
	case MapNodeKindModule, MapNodeKindPackage, MapNodeKindFile:
		return true
	}
	return false
}

// JSONSchema implements jsonschema.Exposer.
func (MapNodeKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Map Node Kind",
		"Path-derived map node classification",
		AllMapNodeKinds,
	), nil
}

// MapGraphPayload is the full map graph for one project, served by
// GET /api/v1/map/{projectHash}.
type MapGraphPayload struct {
	ProjectHash     ProjectHash     `json:"projectHash"`
	RepoFound       bool            `json:"repoFound"` // canonical_cwd resolved to a git repo
	RepoPath        string          `json:"repoPath,omitempty"`
	ParsedLanguages []string        `json:"parsedLanguages"` // e.g. ["go","typescript"]; empty => activity-only
	Nodes           []MapNode       `json:"nodes"`           // all zoom levels; parent links form the tree
	StructureEdges  []MapEdge       `json:"structureEdges"`  // parsed imports, aggregated per node pair
	ActivityEdges   []ActivityEdge  `json:"activityEdges"`   // co-EDIT observations
	Violations      []EdgeViolation `json:"violations"`      // cycles + wrong-way edges
	GeneratedAtMs   int64           `json:"generatedAtMs"`
	AtCommit        string          `json:"atCommit,omitempty"` // set when ?commit= was used
}

// NewMapGraphPayload returns a MapGraphPayload with all slices initialized
// to empty (never-nil marshal guarantee).
func NewMapGraphPayload(projectHash ProjectHash) *MapGraphPayload {
	return &MapGraphPayload{
		ProjectHash:     projectHash,
		ParsedLanguages: []string{},
		Nodes:           []MapNode{},
		StructureEdges:  []MapEdge{},
		ActivityEdges:   []ActivityEdge{},
		Violations:      []EdgeViolation{},
	}
}

// MapNode is one square card on the map: a module, package, or file.
type MapNode struct {
	ID            string      `json:"id"`               // repo-relative path ("internal/ingest", "web/src/lib/api.ts")
	Parent        string      `json:"parent,omitempty"` // ID of parent node ("" for top-level modules)
	Kind          MapNodeKind `json:"kind"`
	Name          string      `json:"name"` // display leaf ("ingest")
	Language      string      `json:"language,omitempty"`
	Layer         int         `json:"layer"`         // 0 = top row; deterministic
	Order         int         `json:"order"`         // stable sort within layer
	Loc           int         `json:"loc"`           // size metric (lines)
	FileCount     int         `json:"fileCount"`     // 1 for files
	RecordedFiles int         `json:"recordedFiles"` // files whose last edit is attributable to a recorded session
	TotalFiles    int         `json:"totalFiles"`
	TouchCount    int         `json:"touchCount"`    // recorded edits in window (activity size metric)
	EffortDensity float64     `json:"effortDensity"` // 0..1 per-file re-edit/error density rollup (0 when unknown)

	// AgentEditedCount / ReadCount / ReadAttribution are the node-grain
	// comprehension signals behind the ranked entry list's tri-state debt
	// tag. ReadAttribution is the honesty field:
	// a zero ReadCount is "unavailable" (no recoverable per-file read
	// attribution for any editing session), "partial" (some do), or
	// "complete" (all do) - never silently indistinguishable from unread.
	AgentEditedCount int                  `json:"agentEditedCount"`
	ReadCount        int                  `json:"readCount"`
	ReadAttribution  ReadAttributionState `json:"readAttribution" required:"true"`

	// ReadState is the composed effective read-state grade for the node's
	// current content version. ChangedRegionCount
	// / AttributedRegionCount / ReviewedRegionCount are the minimal per-node
	// region-coverage counts over that same current version: total changed
	// hunks, hunks the server could attribute to a producing turn, and
	// attributed hunks whose producing turn carries a reviewed+ grade. They
	// supply the client's hunk-linked hover ("N of M changed regions
	// reviewed"). All are server-computed; the client-side debt
	// derivation stays a pure function over these MapNode scalars.
	ReadState             ReadStateGrade `json:"readState" required:"true"`
	ChangedRegionCount    int            `json:"changedRegionCount"`
	AttributedRegionCount int            `json:"attributedRegionCount"`
	ReviewedRegionCount   int            `json:"reviewedRegionCount"`
}

// Validate checks a MapNode's closed-set fields fail closed. It does not
// check cross-region-count consistency (e.g. ReviewedRegionCount <=
// AttributedRegionCount <= ChangedRegionCount); those are producer-side
// invariants, not wire shape rules.
func (n MapNode) Validate() error {
	if err := n.ReadAttribution.Validate(); err != nil {
		return fmt.Errorf("map node validation failed for %q at schema.MapNode.Validate during wire-boundary validation: %w", n.ID, err)
	}
	if err := n.ReadState.Validate(); err != nil {
		return fmt.Errorf("map node validation failed for %q at schema.MapNode.Validate during wire-boundary validation: %w", n.ID, err)
	}
	return nil
}

// ReadAttributionState reports whether per-file read attribution is
// recoverable for a node's editing sessions. It is the
// honesty axis that keeps a zero ReadCount from being misread as "known
// unread" when it is really "never recorded".
type ReadAttributionState string

// ReadAttributionState values.
const (
	ReadAttributionComplete    ReadAttributionState = "complete"
	ReadAttributionPartial     ReadAttributionState = "partial"
	ReadAttributionUnavailable ReadAttributionState = "unavailable"
)

// AllReadAttributionStates is the canonical list of read attribution states.
var AllReadAttributionStates = []ReadAttributionState{
	ReadAttributionComplete,
	ReadAttributionPartial,
	ReadAttributionUnavailable,
}

// String returns the wire representation of the read attribution state.
func (s ReadAttributionState) String() string { return string(s) }

// IsValid reports whether s is one of the defined read attribution states.
func (s ReadAttributionState) IsValid() bool {
	switch s {
	case ReadAttributionComplete, ReadAttributionPartial, ReadAttributionUnavailable:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the read-attribution wire
// boundary.
func (s ReadAttributionState) Validate() error {
	if s.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"read attribution state validation failed for %q at schema.ReadAttributionState.Validate during wire-boundary validation: the value is not one of complete, partial, or unavailable; callers cannot distinguish unread from unrecorded; use a member of schema.AllReadAttributionStates",
		s,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (ReadAttributionState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Read Attribution State",
		"Whether per-file read attribution is recoverable for a node's editing sessions: complete, partial, or unavailable",
		AllReadAttributionStates,
	), nil
}

// ReadStateGrade is the ordinal closed set of explicit read-state acts
// (the read-state grade design): none < viewed < reviewed <
// reviewed_in_detail. The ordering is registry data (the peasant-side
// system-origin TypeDefinition seed); this Go closed set is the sole
// typed copy on the wire, kept identical to the registry seed by
// ReadStateGradeRegistrySeedPermissibleValues.
type ReadStateGrade string

// ReadStateGrade values, in ascending ordinal order.
const (
	ReadStateGradeNone             ReadStateGrade = "none"
	ReadStateGradeViewed           ReadStateGrade = "viewed"
	ReadStateGradeReviewed         ReadStateGrade = "reviewed"
	ReadStateGradeReviewedInDetail ReadStateGrade = "reviewed_in_detail"
)

// AllReadStateGrades is the canonical ordered list of read-state grades,
// ascending: none < viewed < reviewed < reviewed_in_detail.
var AllReadStateGrades = []ReadStateGrade{
	ReadStateGradeNone,
	ReadStateGradeViewed,
	ReadStateGradeReviewed,
	ReadStateGradeReviewedInDetail,
}

// ReadStateGradeRegistrySeedPermissibleValues is AllReadStateGrades with the
// unstated zero grade "none" removed, in the same ascending order: the
// registered set the peasant-side read-state registry seed's
// PermissibleValues must byte-equal. TestReadStateGradeRegistrySeedCrossCheck pins this
// module's half of the cross-check; the peasant-side registry seed test pins
// the other half against this exported value.
var ReadStateGradeRegistrySeedPermissibleValues = []string{
	string(ReadStateGradeViewed),
	string(ReadStateGradeReviewed),
	string(ReadStateGradeReviewedInDetail),
}

// String returns the wire representation of the read-state grade.
func (g ReadStateGrade) String() string { return string(g) }

// IsValid reports whether g is one of the defined read-state grades.
func (g ReadStateGrade) IsValid() bool {
	switch g {
	case ReadStateGradeNone, ReadStateGradeViewed, ReadStateGradeReviewed, ReadStateGradeReviewedInDetail:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the read-state-grade wire
// boundary.
func (g ReadStateGrade) Validate() error {
	if g.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"read state grade validation failed for %q at schema.ReadStateGrade.Validate during wire-boundary validation: the value is not one of none, viewed, reviewed, or reviewed_in_detail; callers cannot render the comprehension-debt tag's clear/partial state; use a member of schema.AllReadStateGrades before serving the payload",
		g,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (ReadStateGrade) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Read State Grade",
		"Ordinal explicit read-state act: none, viewed, reviewed, or reviewed_in_detail",
		AllReadStateGrades,
	), nil
}

// MapEdge is a structure (import) dependency between two nodes.
type MapEdge struct {
	From  string `json:"from"` // node ID
	To    string `json:"to"`
	Count int    `json:"count"` // underlying import count (aggregated)
}

// ActivityEdge is a co-edit observation: two nodes repeatedly edited by the
// same tasks.
type ActivityEdge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	TaskCount int    `json:"taskCount"` // distinct tasks that edited both
}

// EdgeViolationKind classifies a structural violation on the map.
type EdgeViolationKind string

// EdgeViolationKind values (mirrors codegraph's violation kinds).
const (
	EdgeViolationCycle    EdgeViolationKind = "cycle"
	EdgeViolationWrongWay EdgeViolationKind = "wrong_way"
)

// AllEdgeViolationKinds is the canonical list of map structure violations.
var AllEdgeViolationKinds = []EdgeViolationKind{
	EdgeViolationCycle,
	EdgeViolationWrongWay,
}

// String returns the wire representation of the violation kind.
func (k EdgeViolationKind) String() string { return string(k) }

// IsValid reports whether k is one of the defined violation kinds.
func (k EdgeViolationKind) IsValid() bool {
	switch k {
	case EdgeViolationCycle, EdgeViolationWrongWay:
		return true
	}
	return false
}

// JSONSchema implements jsonschema.Exposer.
func (EdgeViolationKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Edge Violation Kind",
		"Structural violation detected on a map edge",
		AllEdgeViolationKinds,
	), nil
}

// EdgeViolation flags an edge that breaks the layering discipline.
type EdgeViolation struct {
	Kind EdgeViolationKind `json:"kind"`
	From string            `json:"from"`
	To   string            `json:"to"`
}

// --- Node detail (rail panel) ---

// MapNodeDetailPayload backs the node rail panel, served by
// GET /api/v1/map/{projectHash}/node?path=<id>.
type MapNodeDetailPayload struct {
	Path          string      `json:"path"`
	Kind          MapNodeKind `json:"kind"`
	Language      string      `json:"language,omitempty"` // e.g. "go", "typescript"; "" for activity-only nodes
	Loc           int         `json:"loc"`
	RecordedFiles int         `json:"recordedFiles"`
	TotalFiles    int         `json:"totalFiles"`
	SessionCount  int         `json:"sessionCount"`
	TaskCount     int         `json:"taskCount"`
	LastTouchMs   *int64      `json:"lastTouchMs,omitempty"`
	// DependsOn / UsedBy are the node's structural role, derived deterministically
	// from the parsed import graph (what this area does): the node IDs
	// this node imports, and those that import it. Most-connected first, capped.
	// Empty when there is no parsed graph (activity-only) or no edges.
	DependsOn     []string      `json:"dependsOn"`
	UsedBy        []string      `json:"usedBy"`
	ShapedBy      []TaskSummary `json:"shapedBy"`          // most recent first, cap 20
	RecentCommits []CommitRef   `json:"recentCommits"`     // touching this node, cap 10
	RetryLoops    int           `json:"retryLoops"`        // summed over touching sessions
	ReEdits       int           `json:"reEdits"`           // re-edited files within this node
	CostUsd       *float64      `json:"costUsd,omitempty"` // nil when unknown
	// RewrittenCommits lists ghost commits touching this node. It is empty when
	// the resolver found no ghosts here.
	RewrittenCommits []RewrittenCommit `json:"rewrittenCommits" required:"true" nullable:"false"`
	// Insights carries mechanical and mined insight envelopes for this node.
	// It is additive alongside the per-change Unusual/Frictions signals, never
	// a replacement for them.
	Insights []SessionInsight `json:"insights" required:"true" nullable:"false"`
}

// NewMapNodeDetailPayload returns a MapNodeDetailPayload with all slices
// initialized to empty (never-nil marshal guarantee).
func NewMapNodeDetailPayload(path string) *MapNodeDetailPayload {
	return &MapNodeDetailPayload{
		Path:             path,
		DependsOn:        []string{},
		UsedBy:           []string{},
		ShapedBy:         []TaskSummary{},
		RecentCommits:    []CommitRef{},
		RewrittenCommits: []RewrittenCommit{},
		Insights:         []SessionInsight{},
	}
}

// Validate checks the additive rewrite and insight invariants: slices are non-nil,
// every RecentCommit and RewrittenCommit is well-formed, every RewrittenCommit's
// SuccessorHash (when set) is present in RecentCommits, and every SessionInsight
// is well-formed (including the Classification-must-be-nil rule). Unlike
// ReviewListPayload, a node detail payload carries no independent session table,
// so RewrittenCommits.SessionIDs are checked for well-formedness only (not
// cross-referenced against a session list this payload does not have).
func (p MapNodeDetailPayload) Validate() error {
	if p.DependsOn == nil || p.UsedBy == nil || p.ShapedBy == nil || p.RecentCommits == nil || p.RewrittenCommits == nil || p.Insights == nil {
		return fmt.Errorf("map node detail validation: dependsOn, usedBy, shapedBy, recentCommits, rewrittenCommits, and insights must be arrays; initialize the payload with NewMapNodeDetailPayload before serving it")
	}
	for index, commit := range p.RecentCommits {
		if err := validateCommitRefShape(commit, "map node detail"); err != nil {
			return fmt.Errorf("map node detail validation: recentCommits[%d]: %w", index, err)
		}
	}
	for index, ghost := range p.RewrittenCommits {
		if err := ghost.Validate(); err != nil {
			return fmt.Errorf("map node detail validation: rewrittenCommits[%d]: %w", index, err)
		}
		if ghost.SuccessorHash != nil {
			found := false
			for _, commit := range p.RecentCommits {
				if commit.Hash == *ghost.SuccessorHash {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("map node detail validation: rewrittenCommits[%d] (ghost %q) has successorHash %q that is not present in recentCommits; include the successor commit or leave successorHash nil", index, ghost.GhostHash, *ghost.SuccessorHash)
			}
		}
	}
	if err := validateSessionInsights(p.Insights); err != nil {
		return fmt.Errorf("map node detail validation: %w", err)
	}
	return nil
}

// TaskSummary is one task: a depth-0 user turn and everything until the next
// user turn (spec §2 "Task", v1 grain).
type TaskSummary struct {
	SessionID   string   `json:"sessionId"`
	EntryIndex  int      `json:"entryIndex"` // depth-0 user-turn entry index (task identity)
	Title       string   `json:"title"`      // first words of the user turn, <=80 chars
	StartMs     *int64   `json:"startMs,omitempty"`
	Outcome     string   `json:"outcome,omitempty"` // session-level outcome
	EditedFiles []string `json:"editedFiles"`
	ReadCount   int      `json:"readCount"`
	// ReadFiles is the per-file derivation of ReadCount: repo-relative paths,
	// sorted, distinct, non-nil, mirroring
	// EditedFiles' invariants. Retroactively recoverable for any already-
	// ingested session with at least one depth-1 tool_use entry carrying
	// non-NULL tool_input (see MapNode.ReadAttribution for the honest
	// residual-gap signal when it is not).
	ReadFiles []string `json:"readFiles" required:"true" nullable:"false"`
	RetryLoop bool     `json:"retryLoop"` // an error streak >=2 occurs inside this task's range
	Labels    []string `json:"labels"`    // effective auto/manual annotation values, plain strings
}

// NewTaskSummary returns a TaskSummary with all slices initialized to empty
// (never-nil marshal guarantee).
func NewTaskSummary(sessionID string, entryIndex int) TaskSummary {
	return TaskSummary{
		SessionID:   sessionID,
		EntryIndex:  entryIndex,
		EditedFiles: []string{},
		ReadFiles:   []string{},
		Labels:      []string{},
	}
}

// Validate checks TaskSummary's ReadFiles invariant: non-nil, sorted
// ascending, and free of duplicates (mirroring the invariants EditedFiles is
// already expected to carry).
func (t TaskSummary) Validate() error {
	if t.ReadFiles == nil {
		return fmt.Errorf("task summary validation failed for session %q entry %d at schema.TaskSummary.Validate during wire-boundary validation: readFiles is null; initialize the array (even empty) with NewTaskSummary before serving it; create the TaskSummary with NewTaskSummary before filling readFiles", t.SessionID, t.EntryIndex)
	}
	for i := 1; i < len(t.ReadFiles); i++ {
		switch {
		case t.ReadFiles[i] == t.ReadFiles[i-1]:
			return fmt.Errorf("task summary validation failed for session %q entry %d at schema.TaskSummary.Validate during wire-boundary validation: readFiles has a duplicate entry %q; the derivation must deduplicate by repo-relative path before serving the payload; deduplicate readFiles before serving the payload", t.SessionID, t.EntryIndex, t.ReadFiles[i])
		case t.ReadFiles[i] < t.ReadFiles[i-1]:
			return fmt.Errorf("task summary validation failed for session %q entry %d at schema.TaskSummary.Validate during wire-boundary validation: readFiles is not sorted ascending (%q precedes %q); sort repo-relative paths before serving the payload; sort readFiles in ascending repo-relative order before serving the payload", t.SessionID, t.EntryIndex, t.ReadFiles[i-1], t.ReadFiles[i])
		}
	}
	return nil
}

// CommitRef is lightweight commit metadata for time strips and rail panels.
type CommitRef struct {
	Hash       string `json:"hash" yaml:"hash"`
	Subject    string `json:"subject" yaml:"subject"`
	TimeMs     *int64 `json:"timeMs,omitempty" yaml:"timeMs,omitempty"`
	HasSession bool   `json:"hasSession" yaml:"hasSession"` // compatibility mirror of len(SessionIDs) > 0
	// SessionIDs names authoritative session_commits bindings in the same
	// strictly increasing rank order as ReviewListPayload.Sessions.
	SessionIDs []SessionID `json:"sessionIds" yaml:"sessionIds" required:"true" nullable:"false"`
	// Associations decomposes each SessionIDs binding into its kind,
	// confidence, and evidence instead of one flat
	// badge. It mirrors SessionIDs one-for-one in the same rank order:
	// Associations[i].SessionID == SessionIDs[i] for every i.
	Associations []SessionAssociation `json:"associations" yaml:"associations" required:"true" nullable:"false"`
}

// NewCommitRef returns commit metadata with non-nil session ID and
// association arrays.
func NewCommitRef(hash, subject string) CommitRef {
	return CommitRef{Hash: hash, Subject: subject, SessionIDs: []SessionID{}, Associations: []SessionAssociation{}}
}

// AssociationKind classifies the strength of evidence linking a recorded
// session to a commit. An association is never a flat badge.
type AssociationKind string

// AssociationKind values. bound = commit-in-branch AND touch overlap;
// candidate = one arm matched; temporal = time-window proximity only.
const (
	AssociationKindBound     AssociationKind = "bound"
	AssociationKindCandidate AssociationKind = "candidate"
	AssociationKindTemporal  AssociationKind = "temporal"
)

// AllAssociationKinds is the canonical list of session<->commit association
// kinds.
var AllAssociationKinds = []AssociationKind{
	AssociationKindBound,
	AssociationKindCandidate,
	AssociationKindTemporal,
}

// String returns the wire representation of the association kind.
func (k AssociationKind) String() string { return string(k) }

// IsValid reports whether k is one of the defined association kinds.
func (k AssociationKind) IsValid() bool {
	switch k {
	case AssociationKindBound, AssociationKindCandidate, AssociationKindTemporal:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the association-kind wire
// boundary.
func (k AssociationKind) Validate() error {
	if k.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"association kind validation failed for %q at schema.AssociationKind.Validate during wire-boundary validation: the value is not one of bound, candidate, or temporal; callers cannot classify the session<->commit relationship; use a member of schema.AllAssociationKinds",
		k,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (AssociationKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Association Kind",
		"Strength of evidence linking a recorded session to a commit: bound, candidate, or temporal",
		AllAssociationKinds,
	), nil
}

// Confidence classifies how strongly the evidence behind a derived
// relationship (a session<->commit association, or a ghost-commit rewrite
// resolution) supports its conclusion.
type Confidence string

// Confidence values.
const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// AllConfidences is the canonical list of confidence levels.
var AllConfidences = []Confidence{
	ConfidenceHigh,
	ConfidenceMedium,
	ConfidenceLow,
}

// String returns the wire representation of the confidence level.
func (c Confidence) String() string { return string(c) }

// IsValid reports whether c is one of the defined confidence levels.
func (c Confidence) IsValid() bool {
	switch c {
	case ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the confidence wire boundary.
func (c Confidence) Validate() error {
	if c.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"confidence validation failed for %q at schema.Confidence.Validate during wire-boundary validation: the value is not one of high, medium, or low; callers cannot weigh the derived relationship; use a member of schema.AllConfidences",
		c,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (Confidence) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Confidence",
		"Strength of evidence behind a derived relationship: high, medium, or low",
		AllConfidences,
	), nil
}

// AssociationEvidence names which signal(s) produced a session<->commit
// association. Producers preserve every matching arm instead of collapsing
// them.
type AssociationEvidence string

// AssociationEvidence values.
const (
	AssociationEvidenceCommitAndTouch AssociationEvidence = "commit_and_touch"
	AssociationEvidenceCommitOnly     AssociationEvidence = "commit_only"
	AssociationEvidenceTouchOnly      AssociationEvidence = "touch_only"
	AssociationEvidenceBranchOnly     AssociationEvidence = "branch_only"
	AssociationEvidenceTimeWindow     AssociationEvidence = "time_window"
)

// AllAssociationEvidences is the canonical list of association evidence
// signals.
var AllAssociationEvidences = []AssociationEvidence{
	AssociationEvidenceCommitAndTouch,
	AssociationEvidenceCommitOnly,
	AssociationEvidenceTouchOnly,
	AssociationEvidenceBranchOnly,
	AssociationEvidenceTimeWindow,
}

// String returns the wire representation of the association evidence.
func (e AssociationEvidence) String() string { return string(e) }

// IsValid reports whether e is one of the defined association evidence
// signals.
func (e AssociationEvidence) IsValid() bool {
	switch e {
	case AssociationEvidenceCommitAndTouch, AssociationEvidenceCommitOnly, AssociationEvidenceTouchOnly, AssociationEvidenceBranchOnly, AssociationEvidenceTimeWindow:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the association-evidence wire
// boundary.
func (e AssociationEvidence) Validate() error {
	if e.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"association evidence validation failed for %q at schema.AssociationEvidence.Validate during wire-boundary validation: the value is not one of commit_and_touch, commit_only, touch_only, branch_only, or time_window; callers cannot explain why the association was made; use a member of schema.AllAssociationEvidences",
		e,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (AssociationEvidence) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Association Evidence",
		"Which signal(s) produced a session<->commit association",
		AllAssociationEvidences,
	), nil
}

// SessionAssociation decomposes a session<->commit relationship into three
// independent axes instead of one flat badge.
type SessionAssociation struct {
	SessionID  SessionID           `json:"sessionId" yaml:"sessionId" required:"true"`
	Kind       AssociationKind     `json:"kind" yaml:"kind" required:"true"`
	Confidence Confidence          `json:"confidence" yaml:"confidence" required:"true"`
	Evidence   AssociationEvidence `json:"evidence" yaml:"evidence" required:"true"`
}

// Validate checks a single association's own fields are well-formed. It does
// not check cross-references to a payload's session table or the surrounding
// CommitRef shape; callers with that context use validateCommitRefShape before
// checking payload-level session membership.
func (a SessionAssociation) Validate() error {
	if a.SessionID == "" {
		return fmt.Errorf("session association validation failed at schema.SessionAssociation.Validate during wire-boundary validation: sessionId is empty; every association must name the session it links, callers cannot resolve an anonymous association; set SessionID before validating or serving the association")
	}
	if err := a.Kind.Validate(); err != nil {
		return fmt.Errorf("session association validation failed for session %q at schema.SessionAssociation.Validate during wire-boundary validation: %w", a.SessionID, err)
	}
	if err := a.Confidence.Validate(); err != nil {
		return fmt.Errorf("session association validation failed for session %q at schema.SessionAssociation.Validate during wire-boundary validation: %w", a.SessionID, err)
	}
	if err := a.Evidence.Validate(); err != nil {
		return fmt.Errorf("session association validation failed for session %q at schema.SessionAssociation.Validate during wire-boundary validation: %w", a.SessionID, err)
	}
	return nil
}

// validateCommitRefShape checks one CommitRef's own shape: SessionIDs and
// Associations are non-nil, HasSession mirrors SessionIDs, SessionIDs contain
// no empty or duplicate values, Associations mirrors SessionIDs one-for-one in
// the same rank order, and every association is individually valid.
func validateCommitRefShape(commit CommitRef, label string) error {
	if commit.SessionIDs == nil {
		return fmt.Errorf("%s validation: commit %q has null sessionIds; initialize every CommitRef with NewCommitRef, including commits with no sessions", label, commit.Hash)
	}
	if commit.HasSession != (len(commit.SessionIDs) > 0) {
		return fmt.Errorf("%s validation: commit %q has hasSession=%t but %d sessionIds; hasSession must mirror whether the authoritative binding list is non-empty", label, commit.Hash, commit.HasSession, len(commit.SessionIDs))
	}
	seen := make(map[SessionID]struct{}, len(commit.SessionIDs))
	for index, sessionID := range commit.SessionIDs {
		if sessionID == "" {
			return fmt.Errorf("%s validation: commit %q has sessionIds[%d] empty; every binding must name a recorded session identity", label, commit.Hash, index)
		}
		if _, exists := seen[sessionID]; exists {
			return fmt.Errorf("%s validation: commit %q repeats sessionId %q at sessionIds[%d]; deduplicate bindings before serving the payload", label, commit.Hash, sessionID, index)
		}
		seen[sessionID] = struct{}{}
	}
	if commit.Associations == nil {
		return fmt.Errorf("%s validation: commit %q has null associations; initialize every CommitRef with NewCommitRef, including commits with no associations", label, commit.Hash)
	}
	if len(commit.Associations) != len(commit.SessionIDs) {
		return fmt.Errorf("%s validation: commit %q has %d associations but %d sessionIds; Associations must mirror SessionIDs one-for-one in the same rank order", label, commit.Hash, len(commit.Associations), len(commit.SessionIDs))
	}
	for index, association := range commit.Associations {
		if err := association.Validate(); err != nil {
			return fmt.Errorf("%s validation: commit %q at association index %d: %w", label, commit.Hash, index, err)
		}
		if association.SessionID != commit.SessionIDs[index] {
			return fmt.Errorf("%s validation: commit %q has associations[%d].sessionId %q but sessionIds[%d] is %q; Associations must equal SessionIDs in the same rank order", label, commit.Hash, index, association.SessionID, index, commit.SessionIDs[index])
		}
	}
	return nil
}

// validateCommitRefAssociations checks the association cross-reference
// invariants for one commit's Associations against the payload's known session
// table: every association SessionID is present in the payload's session table.
func validateCommitRefAssociations(commit CommitRef, knownSessions map[SessionID]struct{}) error {
	for _, association := range commit.Associations {
		if _, exists := knownSessions[association.SessionID]; !exists {
			return fmt.Errorf("review list validation: commit %q references association sessionId %q that is not present in sessions; include it once in ReviewListPayload.Sessions or remove the stale association", commit.Hash, association.SessionID)
		}
	}
	return nil
}

// RewriteResolution classifies whether a ledger-observed commit hash is
// still live on the default branch, was rewritten to a resolvable successor,
// or could not be resolved.
type RewriteResolution string

// RewriteResolution values.
const (
	RewriteResolutionLive       RewriteResolution = "live"
	RewriteResolutionRewritten  RewriteResolution = "rewritten"
	RewriteResolutionUnresolved RewriteResolution = "unresolved"
)

// AllRewriteResolutions is the canonical list of rewrite resolutions.
var AllRewriteResolutions = []RewriteResolution{
	RewriteResolutionLive,
	RewriteResolutionRewritten,
	RewriteResolutionUnresolved,
}

// String returns the wire representation of the rewrite resolution.
func (r RewriteResolution) String() string { return string(r) }

// IsValid reports whether r is one of the defined rewrite resolutions.
func (r RewriteResolution) IsValid() bool {
	switch r {
	case RewriteResolutionLive, RewriteResolutionRewritten, RewriteResolutionUnresolved:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the rewrite-resolution wire
// boundary.
func (r RewriteResolution) Validate() error {
	if r.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"rewrite resolution validation failed for %q at schema.RewriteResolution.Validate during wire-boundary validation: the value is not one of live, rewritten, or unresolved; callers cannot render the ghost commit's resolution state; use a member of schema.AllRewriteResolutions",
		r,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (RewriteResolution) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Rewrite Resolution",
		"Whether a ledger-observed commit hash is live, was rewritten, or could not be resolved",
		AllRewriteResolutions,
	), nil
}

// RewriteMethod names the mechanism the resolver used to map a ghost commit
// to its successor, in the order the resolver attempts them.
type RewriteMethod string

// RewriteMethod values.
const (
	RewriteMethodHash            RewriteMethod = "hash"
	RewriteMethodPatchID         RewriteMethod = "patch_id"
	RewriteMethodAuthorIdentity  RewriteMethod = "author_identity"
	RewriteMethodMessageEmbedded RewriteMethod = "message_embedded"
	RewriteMethodTemporal        RewriteMethod = "temporal"
	RewriteMethodNone            RewriteMethod = "none"
)

// AllRewriteMethods is the canonical list of rewrite resolution methods.
var AllRewriteMethods = []RewriteMethod{
	RewriteMethodHash,
	RewriteMethodPatchID,
	RewriteMethodAuthorIdentity,
	RewriteMethodMessageEmbedded,
	RewriteMethodTemporal,
	RewriteMethodNone,
}

// String returns the wire representation of the rewrite method.
func (m RewriteMethod) String() string { return string(m) }

// IsValid reports whether m is one of the defined rewrite methods.
func (m RewriteMethod) IsValid() bool {
	switch m {
	case RewriteMethodHash, RewriteMethodPatchID, RewriteMethodAuthorIdentity, RewriteMethodMessageEmbedded, RewriteMethodTemporal, RewriteMethodNone:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the rewrite-method wire
// boundary.
func (m RewriteMethod) Validate() error {
	if m.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"rewrite method validation failed for %q at schema.RewriteMethod.Validate during wire-boundary validation: the value is not one of hash, patch_id, author_identity, message_embedded, temporal, or none; callers cannot explain how the resolution was reached; use a member of schema.AllRewriteMethods",
		m,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (RewriteMethod) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Rewrite Method",
		"Mechanism the resolver used to map a ghost commit to its successor",
		AllRewriteMethods,
	), nil
}

// RewrittenCommit is one session-era commit resolution ledger row. The row may
// record a live commit, a rewritten ghost, or an unresolved ghost; only
// non-live rows render as ghosts in timeline views.
type RewrittenCommit struct {
	GhostHash string `json:"ghostHash" yaml:"ghostHash" required:"true"`
	// Subject and AuthorTimeMs are "" / nil when the ledger row never
	// recorded that metadata (degraded providers).
	Subject      string      `json:"subject" yaml:"subject"`
	AuthorTimeMs *int64      `json:"authorTimeMs,omitempty" yaml:"authorTimeMs,omitempty"`
	SessionIDs   []SessionID `json:"sessionIds" yaml:"sessionIds" required:"true" nullable:"false"`
	// SuccessorHash is nil when Resolution is unresolved; non-nil when
	// rewritten. It is never populated for Resolution=live (a live ledger
	// hash is not a ghost).
	SuccessorHash *string           `json:"successorHash,omitempty" yaml:"successorHash,omitempty"`
	Resolution    RewriteResolution `json:"resolution" yaml:"resolution" required:"true"`
	Method        RewriteMethod     `json:"method" yaml:"method" required:"true"`
	Confidence    Confidence        `json:"confidence" yaml:"confidence" required:"true"`
}

// Validate checks a single RewrittenCommit's own fields: the resolution and
// method enums are in-set, SessionIDs are present and non-empty, SuccessorHash
// presence matches Resolution, and Method==none iff Resolution==unresolved. It
// does not check cross-references to a payload's session table or commit set;
// callers with that context use validateRewrittenCommits.
func (r RewrittenCommit) Validate() error {
	if r.GhostHash == "" {
		return fmt.Errorf("rewritten commit validation failed at schema.RewrittenCommit.Validate during wire-boundary validation: ghostHash is empty; every rewritten commit must name the ghost hash it maps, callers cannot resolve an anonymous ghost; populate GhostHash before validating or serving the rewrite")
	}
	if r.SessionIDs == nil {
		return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: sessionIds is null; a rewritten commit without an originating session cannot be placed on a timeline lane; initialize sessionIds with at least one originating session ID before serving the payload", r.GhostHash)
	}
	if len(r.SessionIDs) == 0 {
		return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: sessionIds is empty; a rewritten commit without an originating session cannot be placed on a timeline lane; include at least one originating session ID before serving the payload", r.GhostHash)
	}
	for index, sessionID := range r.SessionIDs {
		if sessionID == "" {
			return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: sessionIds[%d] is empty; a rewritten commit without a recorded session identity cannot be placed on a timeline lane; provide a non-empty recorded session ID at sessionIds[%d] before serving the payload", r.GhostHash, index, index)
		}
	}
	if err := r.Resolution.Validate(); err != nil {
		return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: %w", r.GhostHash, err)
	}
	if err := r.Method.Validate(); err != nil {
		return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: %w", r.GhostHash, err)
	}
	if err := r.Confidence.Validate(); err != nil {
		return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: %w", r.GhostHash, err)
	}
	switch r.Resolution {
	case RewriteResolutionLive:
		if r.SuccessorHash != nil {
			return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: resolution is %q but successorHash is %q; a live commit is already its own reachable identity and cannot map to a successor; clear successorHash, or use resolution %q when the hash names a replacement commit", r.GhostHash, r.Resolution, *r.SuccessorHash, RewriteResolutionRewritten)
		}
	case RewriteResolutionRewritten:
		if r.SuccessorHash == nil {
			return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: resolution is %q but successorHash is nil; a rewritten ghost must name its successor commit, or the resolution should be unresolved; set successorHash when resolution is rewritten", r.GhostHash, r.Resolution)
		}
	case RewriteResolutionUnresolved:
		if r.SuccessorHash != nil {
			return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: resolution is %q but successorHash is %q; an unresolved ghost must not name a successor, or the resolution should be rewritten; clear successorHash when resolution is unresolved", r.GhostHash, r.Resolution, *r.SuccessorHash)
		}
	}
	if (r.Method == RewriteMethodNone) != (r.Resolution == RewriteResolutionUnresolved) {
		return fmt.Errorf("rewritten commit validation failed for ghost %q at schema.RewrittenCommit.Validate during wire-boundary validation: method %q and resolution %q are inconsistent; method must be none if and only if resolution is unresolved; pair method none with unresolved resolution, or use a non-none method with a rewritten resolution", r.GhostHash, r.Method, r.Resolution)
	}
	return nil
}

// validateRewrittenCommits checks the rewrite cross-reference invariants for a
// payload's session-era commit resolution ledger against its known session
// table and known commit hash set: every entry's own fields are well-formed,
// every SessionID is present in the payload's session table, and SuccessorHash
// (when set) is present in the payload's commit set. label identifies the
// owning payload in error messages.
func validateRewrittenCommits(commits []RewrittenCommit, knownSessions map[SessionID]struct{}, knownCommitHashes map[string]struct{}, label string) error {
	if commits == nil {
		return fmt.Errorf("%s validation: rewrittenCommits is null; initialize the array (even empty) before serving the payload; initialize the payload with NewReviewListPayload before serving it", label)
	}
	for index, ghost := range commits {
		if err := ghost.Validate(); err != nil {
			return fmt.Errorf("%s validation: rewrittenCommits[%d]: %w", label, index, err)
		}
		for _, sessionID := range ghost.SessionIDs {
			if _, exists := knownSessions[sessionID]; !exists {
				return fmt.Errorf("%s validation: rewrittenCommits[%d] (ghost %q) references sessionId %q that is not present in the payload's known sessions; include it in the session table or remove the stale reference", label, index, ghost.GhostHash, sessionID)
			}
		}
		if ghost.SuccessorHash != nil {
			if _, exists := knownCommitHashes[*ghost.SuccessorHash]; !exists {
				return fmt.Errorf("%s validation: rewrittenCommits[%d] (ghost %q) has successorHash %q that is not present in the payload's commit set; include the successor commit or leave successorHash nil", label, index, ghost.GhostHash, *ghost.SuccessorHash)
			}
		}
	}
	return nil
}

// InsightKind classifies a SessionInsight.
type InsightKind string

// InsightKind values.
const (
	InsightKindDecision  InsightKind = "decision"
	InsightKindFriction  InsightKind = "friction"
	InsightKindUnusual   InsightKind = "unusual"
	InsightKindRetryLoop InsightKind = "retry_loop"
)

// AllInsightKinds is the canonical list of insight kinds.
var AllInsightKinds = []InsightKind{
	InsightKindDecision,
	InsightKindFriction,
	InsightKindUnusual,
	InsightKindRetryLoop,
}

// String returns the wire representation of the insight kind.
func (k InsightKind) String() string { return string(k) }

// IsValid reports whether k is one of the defined insight kinds.
func (k InsightKind) IsValid() bool {
	switch k {
	case InsightKindDecision, InsightKindFriction, InsightKindUnusual, InsightKindRetryLoop:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the insight-kind wire boundary.
func (k InsightKind) Validate() error {
	if k.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"insight kind validation failed for %q at schema.InsightKind.Validate during wire-boundary validation: the value is not one of decision, friction, unusual, or retry_loop; callers cannot classify the insight; use a member of schema.AllInsightKinds",
		k,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (InsightKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Insight Kind",
		"What a SessionInsight observed: a decision, friction, an unusual rate elevation, or a retry loop",
		AllInsightKinds,
	), nil
}

// InsightProvenance classifies how a SessionInsight was produced: mechanical
// (rule-derived) or mined.
type InsightProvenance string

// InsightProvenance values.
const (
	InsightProvenanceMechanical InsightProvenance = "mechanical"
	InsightProvenanceMined      InsightProvenance = "mined"
)

// AllInsightProvenances is the canonical list of insight provenances.
var AllInsightProvenances = []InsightProvenance{
	InsightProvenanceMechanical,
	InsightProvenanceMined,
}

// String returns the wire representation of the insight provenance.
func (p InsightProvenance) String() string { return string(p) }

// IsValid reports whether p is one of the defined insight provenances.
func (p InsightProvenance) IsValid() bool {
	switch p {
	case InsightProvenanceMechanical, InsightProvenanceMined:
		return true
	}
	return false
}

// Validate rejects values that cannot cross the insight-provenance wire
// boundary.
func (p InsightProvenance) Validate() error {
	if p.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"insight provenance validation failed for %q at schema.InsightProvenance.Validate during wire-boundary validation: the value is not one of mechanical or mined; callers cannot tell whether the insight was rule-derived or mined; use a member of schema.AllInsightProvenances",
		p,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (InsightProvenance) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Insight Provenance",
		"How a SessionInsight was produced: mechanical (rule-derived) or mined",
		AllInsightProvenances,
	), nil
}

// InsightEvidence is one traceability pointer for a SessionInsight: the
// recorded session (and, when known, the turn / file / commit) that grounds
// it. Every mechanical insight carries at least one.
type InsightEvidence struct {
	SessionID  SessionID `json:"sessionId" yaml:"sessionId" required:"true"`
	EntryIndex *int      `json:"entryIndex,omitempty" yaml:"entryIndex,omitempty"`
	File       string    `json:"file,omitempty" yaml:"file,omitempty"`
	CommitHash string    `json:"commitHash,omitempty" yaml:"commitHash,omitempty"`
}

// Validate checks that the evidence item carries a traceable session identity.
func (e InsightEvidence) Validate() error {
	if e.SessionID == "" {
		return fmt.Errorf("insight evidence validation failed at schema.InsightEvidence.Validate during wire-boundary validation: sessionId is empty; an evidence item without a recorded session identity cannot anchor the traceability spine; provide a non-empty recorded session ID before serving the insight")
	}
	return nil
}

// InsightClassification is the reserved taxonomy tuple (category x cause x
// severity(scope, locus) x resolution). Its bare string fields keep the wire
// shape stable for future closed sets. The current contract requires
// SessionInsight.Classification to remain nil; see SessionInsight.Validate.
type InsightClassification struct {
	Category      string `json:"category" yaml:"category"`
	Cause         string `json:"cause" yaml:"cause"`
	SeverityScope string `json:"severityScope" yaml:"severityScope"`
	SeverityLocus string `json:"severityLocus" yaml:"severityLocus"`
	Resolution    string `json:"resolution" yaml:"resolution"`
}

// SessionInsight is one insight: a (kind x provenance x confidence) envelope
// with evidence and subjects, additive alongside ChangeDetailPayload's
// existing Unusual/Frictions signals. Current mechanical producers leave
// Classification nil. A future contract may define closed classification
// sets without changing this envelope's shape.
type SessionInsight struct {
	Kind       InsightKind       `json:"kind" yaml:"kind" required:"true"`
	Provenance InsightProvenance `json:"provenance" yaml:"provenance" required:"true"`
	Confidence Confidence        `json:"confidence" yaml:"confidence" required:"true"`
	Title      string            `json:"title" yaml:"title" required:"true"`
	Summary    string            `json:"summary,omitempty" yaml:"summary,omitempty"`
	// Subjects names the node ids / file paths the insight is about.
	Subjects []string `json:"subjects" yaml:"subjects" required:"true" nullable:"false"`
	// Evidence is the traceability spine: every mechanical insight carries
	// at least one item.
	Evidence []InsightEvidence `json:"evidence" yaml:"evidence" required:"true" nullable:"false"`
	// Classification MUST be nil under the current contract (see Validate).
	// The field is reserved until its per-field taxonomy sets are defined.
	Classification *InsightClassification `json:"classification,omitempty" yaml:"classification,omitempty"`
}

// Validate checks the insight invariants: Kind/Provenance/Confidence are in-set,
// Subjects and Evidence are non-nil, every evidence item carries a non-empty
// session identity, every mechanical insight carries at least one evidence
// item, and Classification is nil. A future contract may replace the
// must-be-nil rule with per-field closed sets without changing the shape.
func (i SessionInsight) Validate() error {
	if err := i.Kind.Validate(); err != nil {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: %w", i.Title, err)
	}
	if err := i.Provenance.Validate(); err != nil {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: %w", i.Title, err)
	}
	if err := i.Confidence.Validate(); err != nil {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: %w", i.Title, err)
	}
	if i.Subjects == nil {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: subjects is null; initialize the array (even empty) before serving the payload; initialize Subjects before serving the insight", i.Title)
	}
	if i.Evidence == nil {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: evidence is null; initialize the array (even empty) before serving the payload; initialize Evidence before serving the insight", i.Title)
	}
	if i.Provenance == InsightProvenanceMechanical && len(i.Evidence) == 0 {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: provenance is mechanical but evidence is empty; every mechanical insight must cite at least one evidence item so the traceability spine (signal -> file change -> session) holds; include at least one evidence item before serving a mechanical insight", i.Title)
	}
	for index, evidence := range i.Evidence {
		if err := evidence.Validate(); err != nil {
			return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: evidence[%d]: %w", i.Title, index, err)
		}
	}
	if i.Classification != nil {
		return fmt.Errorf("session insight validation failed for %q at schema.SessionInsight.Validate during wire-boundary validation: classification is non-nil, but the current contract has no closed classification sets and consumers cannot validate those values; leave classification nil until the contract defines its category, cause, severity, and resolution sets; clear Classification before serving the insight", i.Title)
	}
	return nil
}

// validateSessionInsights runs SessionInsight.Validate over every insight,
// returning the first failure with its index for locality.
func validateSessionInsights(insights []SessionInsight) error {
	for index, insight := range insights {
		if err := insight.Validate(); err != nil {
			return fmt.Errorf("insights[%d]: %w", index, err)
		}
	}
	return nil
}

// --- Tasks (Tasks lens + file filter) ---

// ProjectTasksPayload backs the Tasks lens, served by
// GET /api/v1/map/{projectHash}/tasks?file=<path>.
type ProjectTasksPayload struct {
	ProjectHash ProjectHash   `json:"projectHash"`
	Tasks       []TaskSummary `json:"tasks"` // reverse-chronological, cap 500
	FileFilter  string        `json:"fileFilter,omitempty"`
}

// NewProjectTasksPayload returns a ProjectTasksPayload with all slices
// initialized to empty (never-nil marshal guarantee).
func NewProjectTasksPayload(projectHash ProjectHash) *ProjectTasksPayload {
	return &ProjectTasksPayload{
		ProjectHash: projectHash,
		Tasks:       []TaskSummary{},
	}
}

// --- Project summaries (home picker) ---

// ProjectSummariesPayload backs the home project picker, served by
// GET /api/v1/projects/summary.
type ProjectSummariesPayload struct {
	Projects []ProjectSummary `json:"projects"`
}

// NewProjectSummariesPayload returns a ProjectSummariesPayload with all
// slices initialized to empty (never-nil marshal guarantee).
func NewProjectSummariesPayload() *ProjectSummariesPayload {
	return &ProjectSummariesPayload{Projects: []ProjectSummary{}}
}

// ProjectSummary is one row of the home picker: a project with its recorded
// stats (sessions · recorded coverage · last work · open changes).
type ProjectSummary struct {
	ProjectHash   ProjectHash `json:"projectHash"`
	Project       string      `json:"project"`       // display name (canonical cwd, else the hash)
	Sessions      int         `json:"sessions"`      // recorded session count
	RecordedFiles int         `json:"recordedFiles"` // coverage numerator (same rule as MapNode)
	TotalFiles    int         `json:"totalFiles"`    // coverage denominator
	LastWorkMs    *int64      `json:"lastWorkMs,omitempty"`
	OpenChanges   int         `json:"openChanges"` // local non-default branches not merged (0 when no repo)
}

// ProjectResolutionPayload resolves one explicitly requested project display
// identity to its opaque hash without enumerating sibling projects. It exists
// for stable deep links when discovery lists are narrowed by user selection.
type ProjectResolutionPayload struct {
	Project     string      `json:"project" required:"true"`
	ProjectHash ProjectHash `json:"projectHash" required:"true"`
}

// --- Review ---

// ReviewListPayload lists a project's changes, served by
// GET /api/v1/review/{projectHash}.
type ReviewListPayload struct {
	ProjectHash   ProjectHash          `json:"projectHash"`
	RepoFound     bool                 `json:"repoFound"`
	DefaultBranch string               `json:"defaultBranch,omitempty"`
	Changes       []ChangeSummary      `json:"changes" required:"true" nullable:"false"`       // open first, then merged
	RecentCommits []CommitRef          `json:"recentCommits" required:"true" nullable:"false"` // default-branch, cap 200 (time strip)
	Sessions      []TimelineSessionRef `json:"sessions" required:"true" nullable:"false"`      // complete visible project timeline identities, including sessions not linked to displayed commits
	// RewrittenCommits lists the project's full session-era commit resolution
	// ledger. Only non-live rows render as ghosts; live rows remain valid history
	// entries. It is empty when the resolver found no relevant session-era rows.
	RewrittenCommits []RewrittenCommit `json:"rewrittenCommits" required:"true" nullable:"false"`
}

// TimelineSessionRef is identity and display metadata for a recorded session
// available to the project timeline. Producers order these references by
// known startMs descending, then sessionId ascending; missing startMs follows
// every known timestamp and is likewise ordered by sessionId. HasCommitBinding
// is computed from the complete authoritative session_commits relation, not
// merely the bounded default-branch commit window returned alongside it.
// CommitRef.SessionIDs names bindings that are visible inside that window.
type TimelineSessionRef struct {
	SessionID        SessionID `json:"sessionId" yaml:"sessionId"`
	Title            string    `json:"title" yaml:"title"`
	Harness          Harness   `json:"harness" yaml:"harness"`
	StartMs          *int64    `json:"startMs,omitempty" yaml:"startMs,omitempty"`
	HasCommitBinding bool      `json:"hasCommitBinding" yaml:"hasCommitBinding"`
}

// NewReviewListPayload returns a ReviewListPayload with all slices
// initialized to empty (never-nil marshal guarantee).
func NewReviewListPayload(projectHash ProjectHash) *ReviewListPayload {
	return &ReviewListPayload{
		ProjectHash:      projectHash,
		Changes:          []ChangeSummary{},
		RecentCommits:    []CommitRef{},
		Sessions:         []TimelineSessionRef{},
		RewrittenCommits: []RewrittenCommit{},
	}
}

// Validate checks the normalized timeline relationship and compatibility
// invariants. It does not infer candidate or temporal associations, and it
// validates each CommitRef's shape before checking timeline membership and
// rank-order rules.
func (p ReviewListPayload) Validate() error {
	if err := p.ProjectHash.Validate(); err != nil {
		return fmt.Errorf("review list validation: projectHash is invalid: %w; resolve the canonical project identity before serving the payload", err)
	}
	if p.Changes == nil || p.RecentCommits == nil || p.Sessions == nil || p.RewrittenCommits == nil {
		return fmt.Errorf("review list validation: changes, recentCommits, sessions, and rewrittenCommits must be arrays; initialize the payload with NewReviewListPayload before serving it")
	}
	knownSessions := make(map[SessionID]TimelineSessionRef, len(p.Sessions))
	sessionRanks := make(map[SessionID]int, len(p.Sessions))
	knownHarnesses := make(map[Harness]struct{})
	for _, harness := range Harnesses() {
		knownHarnesses[harness] = struct{}{}
	}
	for index, session := range p.Sessions {
		if session.SessionID == "" {
			return fmt.Errorf("review list validation: timeline session has an empty sessionId; producers must emit a stable session identity")
		}
		if _, exists := knownSessions[session.SessionID]; exists {
			return fmt.Errorf("review list validation: duplicate timeline session %q; normalize sessions by sessionId before serving the payload", session.SessionID)
		}
		if _, known := knownHarnesses[session.Harness]; !known {
			return fmt.Errorf("review list validation: timeline session %q at index %d has unknown harness %q; use one of schema.Harnesses() before serving the payload", session.SessionID, index, session.Harness)
		}
		knownSessions[session.SessionID] = session
		sessionRanks[session.SessionID] = index
		if index > 0 {
			previous := p.Sessions[index-1]
			outOfOrder := false
			switch {
			case previous.StartMs == nil && session.StartMs != nil:
				outOfOrder = true
			case previous.StartMs != nil && session.StartMs != nil && *previous.StartMs < *session.StartMs:
				outOfOrder = true
			case (previous.StartMs == nil && session.StartMs == nil) || (previous.StartMs != nil && session.StartMs != nil && *previous.StartMs == *session.StartMs):
				outOfOrder = previous.SessionID > session.SessionID
			}
			if outOfOrder {
				return fmt.Errorf("review list validation: timeline sessions %q and %q violate canonical ordering at index %d; producers must sort known startMs descending, break equal timestamps by sessionId ascending, and place missing startMs last", previous.SessionID, session.SessionID, index)
			}
		}
	}
	knownSessionSet := make(map[SessionID]struct{}, len(knownSessions))
	for sessionID := range knownSessions {
		knownSessionSet[sessionID] = struct{}{}
	}
	for commitIndex, commit := range p.RecentCommits {
		if err := validateCommitRefShape(commit, "review list"); err != nil {
			return err
		}
		previousRank := -1
		for bindingIndex, sessionID := range commit.SessionIDs {
			session, exists := knownSessions[sessionID]
			if !exists {
				return fmt.Errorf("review list validation: commit %q references unknown sessionId %q; include it once in sessions or remove the stale binding", commit.Hash, sessionID)
			}
			if !session.HasCommitBinding {
				return fmt.Errorf("review list validation: commit %q references sessionId %q but that session has hasCommitBinding=false; set hasCommitBinding=true because a visible commit reference proves an authoritative binding", commit.Hash, sessionID)
			}
			rank := sessionRanks[sessionID]
			if rank <= previousRank {
				return fmt.Errorf("review list validation: commit %q at index %d has noncanonical sessionIds order at binding %d; order every binding by the strictly increasing rank of ReviewListPayload.Sessions", commit.Hash, commitIndex, bindingIndex)
			}
			previousRank = rank
		}
		if err := validateCommitRefAssociations(commit, knownSessionSet); err != nil {
			return err
		}
	}
	if err := validateRewrittenCommits(p.RewrittenCommits, knownSessionSet, recentCommitHashes(p.RecentCommits), "review list"); err != nil {
		return err
	}
	return nil
}

// recentCommitHashes returns the set of commit hashes visible in commits, for
// cross-referencing a RewrittenCommit.SuccessorHash against a payload's
// known commit set.
func recentCommitHashes(commits []CommitRef) map[string]struct{} {
	hashes := make(map[string]struct{}, len(commits))
	for _, commit := range commits {
		hashes[commit.Hash] = struct{}{}
	}
	return hashes
}

// ChangeSummary is one row of the Review list: a local branch measured
// against the default branch.
type ChangeSummary struct {
	Branch       string `json:"branch"`
	AheadCount   int    `json:"aheadCount"`
	BehindCount  int    `json:"behindCount"`
	FilesChanged int    `json:"filesChanged"`
	SessionCount int    `json:"sessionCount"`
	TaskCount    int    `json:"taskCount"`
	NewEdges     int    `json:"newEdges"`
	RemovedEdges int    `json:"removedEdges"`
	Violations   int    `json:"violations"`
	LastWorkMs   *int64 `json:"lastWorkMs,omitempty"`
	Merged       bool   `json:"merged"`
	MergedAtMs   *int64 `json:"mergedAtMs,omitempty"`
	// Reverted is true when this change was merged and later undone by a
	// `git revert` on the default branch (git-native signal only).
	Reverted bool `json:"reverted,omitempty"`

	// Graph anchors (Changes graph): how this row attaches to lane 0
	// (ReviewListPayload.RecentCommits). Open branches fork at BaseHash and
	// sit at TipCommitMs; merged rows rejoin at MergeCommitHash.
	BaseHash        string `json:"baseHash,omitempty"`        // merge-base commit hash (fork anchor; open branches)
	TipCommitMs     *int64 `json:"tipCommitMs,omitempty"`     // branch tip committer time (row position; open branches)
	MergeCommitHash string `json:"mergeCommitHash,omitempty"` // merge commit hash (join anchor; merged rows)
}

// ChangeDetailPayload backs the Review change-detail surface, served by
// GET /api/v1/review/{projectHash}/change?branch=<name>.
type ChangeDetailPayload struct {
	Branch            string          `json:"branch"`
	BaseRef           string          `json:"baseRef"` // merge-base hash
	DefaultBranch     string          `json:"defaultBranch"`
	Files             []FileChange    `json:"files"`
	Slice             MapSlice        `json:"slice"` // touched nodes + 1-hop
	NewEdges          []MapEdge       `json:"newEdges"`
	RemovedEdges      []MapEdge       `json:"removedEdges"`
	NewNodes          []string        `json:"newNodes"` // node IDs
	RemovedNodes      []string        `json:"removedNodes"`
	Violations        []EdgeViolation `json:"violations"` // NEW violations introduced by this change
	Work              []ChangeSession `json:"work"`
	UnrecordedCommits []CommitRef     `json:"unrecordedCommits"`
	// Unusual holds NEUTRAL rate-elevation observations vs the project baseline
	// (e.g. more retry loops per conversation than usual) — facts, never a
	// verdict or grade.
	Unusual []UnusualSignal `json:"unusual"`
	// Frictions holds NEUTRAL recurring-friction counts keyed by (kind, file):
	// "this kind of friction touched this file N times across M conversations"
	// Facts are for orientation, never a verdict.
	Frictions []FrictionCluster `json:"frictions"`
	// Insights carries mechanical and mined insight envelopes for this change.
	// It is additive alongside Unusual/Frictions above, never a replacement
	// for them.
	Insights     []SessionInsight `json:"insights" required:"true" nullable:"false"`
	LinesAdded   int              `json:"linesAdded"`
	LinesRemoved int              `json:"linesRemoved"`
	OutputTokens int64            `json:"outputTokens"` // SUM of output_tokens over bound sessions
	CostUsd      *float64         `json:"costUsd,omitempty"`
}

// NewChangeDetailPayload returns a ChangeDetailPayload with all slices
// (including the nested MapSlice's) initialized to empty (never-nil marshal
// guarantee).
func NewChangeDetailPayload(branch string) *ChangeDetailPayload {
	return &ChangeDetailPayload{
		Branch:            branch,
		Files:             []FileChange{},
		Slice:             NewMapSlice(),
		NewEdges:          []MapEdge{},
		RemovedEdges:      []MapEdge{},
		NewNodes:          []string{},
		RemovedNodes:      []string{},
		Violations:        []EdgeViolation{},
		Work:              []ChangeSession{},
		UnrecordedCommits: []CommitRef{},
		Unusual:           []UnusualSignal{},
		Frictions:         []FrictionCluster{},
		Insights:          []SessionInsight{},
	}
}

// Validate checks that Insights is non-nil and every entry is well-formed,
// including the Classification-must-be-nil rule. It also validates each
// UnrecordedCommit's shape. Other ChangeDetailPayload fields do not currently
// define validation rules here.
func (p ChangeDetailPayload) Validate() error {
	if p.Insights == nil {
		return fmt.Errorf("change detail validation: insights must be an array; initialize the payload with NewChangeDetailPayload before serving it")
	}
	for index, commit := range p.UnrecordedCommits {
		if err := validateCommitRefShape(commit, "change detail"); err != nil {
			return fmt.Errorf("change detail validation: unrecordedCommits[%d]: %w", index, err)
		}
	}
	if err := validateSessionInsights(p.Insights); err != nil {
		return fmt.Errorf("change detail validation: %w", err)
	}
	return nil
}

// UnusualSignal is one neutral rate-elevation: a per-conversation rate for this
// change that runs notably above the project baseline. Facts for orientation —
// the surface shows, it does not grade.
type UnusualSignal struct {
	Kind       string  `json:"kind"`       // e.g. "retryLoops"
	Label      string  `json:"label"`      // plain, neutral
	PerChange  float64 `json:"perChange"`  // this change's per-conversation rate
	PerProject float64 `json:"perProject"` // the project's per-conversation baseline
}

// FrictionCluster is a NEUTRAL count of a recurring friction signal keyed to a
// file: "this kind of friction touched this file N times across M
// conversations". A fact for orientation — the surface shows, it does not grade
// Kind is a stable slug ("retryLoop") so more kinds can be added
// without a breaking change.
type FrictionCluster struct {
	Kind     string `json:"kind"`     // signal slug, e.g. "retryLoop"
	Label    string `json:"label"`    // plain, neutral (e.g. "retry loops")
	File     string `json:"file"`     // repo-relative path
	Count    int    `json:"count"`    // occurrences (retry-loop tasks touching this file)
	Sessions int    `json:"sessions"` // distinct conversations those occurrences span
}

// MapSlice is a scoped sub-map: the touched nodes plus their one-hop
// neighborhood, with layer/order preserved from the full map.
type MapSlice struct {
	Nodes          []MapNode      `json:"nodes"` // layer/order preserved from full map
	StructureEdges []MapEdge      `json:"structureEdges"`
	ActivityEdges  []ActivityEdge `json:"activityEdges"`
}

// NewMapSlice returns a MapSlice with all slices initialized to empty
// (never-nil marshal guarantee).
func NewMapSlice() MapSlice {
	return MapSlice{
		Nodes:          []MapNode{},
		StructureEdges: []MapEdge{},
		ActivityEdges:  []ActivityEdge{},
	}
}

// FileChangeStatus classifies a file-level delta using Git's canonical
// one-letter status tokens.
type FileChangeStatus string

const (
	FileChangeStatusModified FileChangeStatus = "M"
	FileChangeStatusAdded    FileChangeStatus = "A"
	FileChangeStatusDeleted  FileChangeStatus = "D"
	FileChangeStatusRenamed  FileChangeStatus = "R"
)

// AllFileChangeStatuses is the canonical inventory of file change statuses.
var AllFileChangeStatuses = []FileChangeStatus{
	FileChangeStatusModified,
	FileChangeStatusAdded,
	FileChangeStatusDeleted,
	FileChangeStatusRenamed,
}

// String returns the wire representation of the file change status.
func (s FileChangeStatus) String() string { return string(s) }

// IsValid reports whether s is a canonical file change status.
func (s FileChangeStatus) IsValid() bool {
	switch s {
	case FileChangeStatusModified, FileChangeStatusAdded, FileChangeStatusDeleted, FileChangeStatusRenamed:
		return true
	}
	return false
}

// Validate rejects values that cannot cross a file-change wire boundary.
func (s FileChangeStatus) Validate() error {
	if s.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"file change status validation failed for %q at schema.FileChangeStatus.Validate during wire-boundary validation: the value is not one of M, A, D, or R; callers cannot classify the file delta; use a member of schema.AllFileChangeStatuses",
		s,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (FileChangeStatus) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("File Change Status")
	s.WithDescription("Git file delta status: modified, added, deleted, or renamed")
	s.WithEnum("M", "A", "D", "R")
	s.WithExamples("M", "A", "D", "R")
	return s, nil
}

// FileChange is one file-level delta of a change (branch vs merge-base).
// LinesAdded/LinesRemoved are the per-file numstat churn (0 for binary files or
// when numstat is unavailable) — the change-weight treemap's sizing input
// Always present; 0 is meaningful, so no omitempty.
type FileChange struct {
	Path         string           `json:"path"`
	Status       FileChangeStatus `json:"status"`
	OldPath      *string          `json:"oldPath,omitempty"`
	LinesAdded   int              `json:"linesAdded"`
	LinesRemoved int              `json:"linesRemoved"`
}

// ChangeDiffPayload is the rendered unified diff of ONE changed file of a change
// (branch vs its merge-base with the default branch) — the lazy per-file
// companion to ChangeDetailPayload (contract §3, GET /review/{projectHash}/diff
// ?branch=&file=). Binary files come back Binary=true with no hunks; files
// exceeding the size cap come back Truncated.
type ChangeDiffPayload struct {
	Branch    string           `json:"branch"`
	File      string           `json:"file"` // the new path
	OldPath   *string          `json:"oldPath,omitempty"`
	Status    FileChangeStatus `json:"status"`
	Binary    bool             `json:"binary"`
	Truncated bool             `json:"truncated"`
	Hunks     []DiffHunk       `json:"hunks"`
}

// NewChangeDiffPayload returns a ChangeDiffPayload with Hunks initialized to
// empty (never-nil marshal guarantee).
func NewChangeDiffPayload(branch, file string) *ChangeDiffPayload {
	return &ChangeDiffPayload{Branch: branch, File: file, Hunks: []DiffHunk{}}
}

// DiffHunk is one "@@ -oldStart,oldLines +newStart,newLines @@" section.
// Line numbers in the gutter are derivable from OldStart/NewStart plus position.
type DiffHunk struct {
	OldStart int        `json:"oldStart"`
	OldLines int        `json:"oldLines"`
	NewStart int        `json:"newStart"`
	NewLines int        `json:"newLines"`
	Header   string     `json:"header,omitempty"`
	Lines    []DiffLine `json:"lines"`
	// Attribution (the mission climax): the recorded conversation that wrote
	// most of this hunk's added lines, resolved via git blame → commit →
	// session. Empty when the hunk's new lines trace to no recorded session
	// (hand-written, or authored outside this tool).
	SessionID    string `json:"sessionId,omitempty"`
	SessionTitle string `json:"sessionTitle,omitempty"`
}

// DiffLineKind classifies a unified-diff line.
type DiffLineKind string

const (
	DiffLineKindContext DiffLineKind = "context"
	DiffLineKindAdd     DiffLineKind = "add"
	DiffLineKindDelete  DiffLineKind = "del"
)

// AllDiffLineKinds is the canonical inventory of unified-diff line kinds.
var AllDiffLineKinds = []DiffLineKind{
	DiffLineKindContext,
	DiffLineKindAdd,
	DiffLineKindDelete,
}

// String returns the wire representation of the diff line kind.
func (k DiffLineKind) String() string { return string(k) }

// IsValid reports whether k is a canonical diff line kind.
func (k DiffLineKind) IsValid() bool {
	switch k {
	case DiffLineKindContext, DiffLineKindAdd, DiffLineKindDelete:
		return true
	}
	return false
}

// Validate rejects values that cannot cross a diff-line wire boundary.
func (k DiffLineKind) Validate() error {
	if k.IsValid() {
		return nil
	}
	return fmt.Errorf(
		"diff line kind validation failed for %q at schema.DiffLineKind.Validate during wire-boundary validation: the value is not one of context, add, or del; callers cannot render the line with correct diff semantics; use a member of schema.AllDiffLineKinds",
		k,
	)
}

// JSONSchema implements jsonschema.Exposer.
func (DiffLineKind) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Diff Line Kind")
	s.WithDescription("Unified-diff line kind: context, addition, or deletion")
	s.WithEnum("context", "add", "del")
	s.WithExamples("context", "add", "del")
	return s, nil
}

// DiffLine is one line within a hunk. Kind is "context" | "add" | "del";
// Text excludes the leading +/-/space marker.
type DiffLine struct {
	Kind DiffLineKind `json:"kind"`
	Text string       `json:"text"`
}

// ChangeBinding states how strongly a session is tied to a change
// (contract §2 binding rule, spec §6.2).
type ChangeBinding string

// ChangeBinding values. A session is bound when at least one of its linked
// commits is contained in the branch AND its recorded edits overlap the
// branch's changed files; one-arm matches (or git_branch equality alone)
// are candidates. Candidates are never silently dropped.
const (
	ChangeBindingBound     ChangeBinding = "bound"
	ChangeBindingCandidate ChangeBinding = "candidate"
)

// AllChangeBindings is the canonical list of session-to-change bindings.
var AllChangeBindings = []ChangeBinding{
	ChangeBindingBound,
	ChangeBindingCandidate,
}

// String returns the wire representation of the binding.
func (b ChangeBinding) String() string { return string(b) }

// IsValid reports whether b is one of the defined bindings.
func (b ChangeBinding) IsValid() bool {
	switch b {
	case ChangeBindingBound, ChangeBindingCandidate:
		return true
	}
	return false
}

// JSONSchema implements jsonschema.Exposer.
func (ChangeBinding) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Change Binding",
		"Strength of the evidence connecting a recorded session to a code change",
		AllChangeBindings,
	), nil
}

// ChangeSession is one recorded session behind a change, with its tasks.
type ChangeSession struct {
	SessionID string        `json:"sessionId"`
	Title     string        `json:"title"`
	Harness   string        `json:"harness"`
	StartMs   *int64        `json:"startMs,omitempty"`
	Binding   ChangeBinding `json:"binding"` // bound = commit-in-branch AND touch overlap; candidate = one arm only
	Tasks     []TaskSummary `json:"tasks"`
}

// NewChangeSession returns a ChangeSession with all slices initialized to
// empty (never-nil marshal guarantee).
func NewChangeSession(sessionID string, binding ChangeBinding) ChangeSession {
	return ChangeSession{
		SessionID: sessionID,
		Binding:   binding,
		Tasks:     []TaskSummary{},
	}
}
