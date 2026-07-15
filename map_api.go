package schema

import "fmt"

// Map / Review wire contract (impl contract §2).
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
}

// NewMapNodeDetailPayload returns a MapNodeDetailPayload with all slices
// initialized to empty (never-nil marshal guarantee).
func NewMapNodeDetailPayload(path string) *MapNodeDetailPayload {
	return &MapNodeDetailPayload{
		Path:          path,
		DependsOn:     []string{},
		UsedBy:        []string{},
		ShapedBy:      []TaskSummary{},
		RecentCommits: []CommitRef{},
	}
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
	RetryLoop   bool     `json:"retryLoop"` // an error streak >=2 occurs inside this task's range
	Labels      []string `json:"labels"`    // effective auto/manual annotation values, plain strings
}

// NewTaskSummary returns a TaskSummary with all slices initialized to empty
// (never-nil marshal guarantee).
func NewTaskSummary(sessionID string, entryIndex int) TaskSummary {
	return TaskSummary{
		SessionID:   sessionID,
		EntryIndex:  entryIndex,
		EditedFiles: []string{},
		Labels:      []string{},
	}
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
}

// NewCommitRef returns commit metadata with a non-nil session ID array.
func NewCommitRef(hash, subject string) CommitRef {
	return CommitRef{Hash: hash, Subject: subject, SessionIDs: []SessionID{}}
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
		ProjectHash:   projectHash,
		Changes:       []ChangeSummary{},
		RecentCommits: []CommitRef{},
		Sessions:      []TimelineSessionRef{},
	}
}

// Validate checks the normalized timeline relationship and compatibility
// invariants. It does not infer candidate or temporal associations.
func (p ReviewListPayload) Validate() error {
	if err := p.ProjectHash.Validate(); err != nil {
		return fmt.Errorf("review list validation: projectHash is invalid: %w; resolve the canonical project identity before serving the payload", err)
	}
	if p.Changes == nil || p.RecentCommits == nil || p.Sessions == nil {
		return fmt.Errorf("review list validation: changes, recentCommits, and sessions must be arrays; initialize the payload with NewReviewListPayload before serving it")
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
	for commitIndex, commit := range p.RecentCommits {
		if commit.SessionIDs == nil {
			return fmt.Errorf("review list validation: commit %q has null sessionIds; initialize every CommitRef with NewCommitRef, including commits with no sessions", commit.Hash)
		}
		if commit.HasSession != (len(commit.SessionIDs) > 0) {
			return fmt.Errorf("review list validation: commit %q has hasSession=%t but %d sessionIds; hasSession must mirror whether the authoritative binding list is non-empty", commit.Hash, commit.HasSession, len(commit.SessionIDs))
		}
		seen := make(map[SessionID]bool, len(commit.SessionIDs))
		previousRank := -1
		for bindingIndex, sessionID := range commit.SessionIDs {
			if seen[sessionID] {
				return fmt.Errorf("review list validation: commit %q repeats sessionId %q; deduplicate bindings before serving the payload", commit.Hash, sessionID)
			}
			seen[sessionID] = true
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
	}
	return nil
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
	Frictions    []FrictionCluster `json:"frictions"`
	LinesAdded   int               `json:"linesAdded"`
	LinesRemoved int               `json:"linesRemoved"`
	OutputTokens int64             `json:"outputTokens"` // SUM of output_tokens over bound sessions
	CostUsd      *float64          `json:"costUsd,omitempty"`
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
	}
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

// FileChange is one file-level delta of a change (branch vs merge-base).
// LinesAdded/LinesRemoved are the per-file numstat churn (0 for binary files or
// when numstat is unavailable) — the change-weight treemap's sizing input
// Always present; 0 is meaningful, so no omitempty.
type FileChange struct {
	Path         string  `json:"path"`
	Status       string  `json:"status"` // "M" | "A" | "D" | "R"
	OldPath      *string `json:"oldPath,omitempty"`
	LinesAdded   int     `json:"linesAdded"`
	LinesRemoved int     `json:"linesRemoved"`
}

// ChangeDiffPayload is the rendered unified diff of ONE changed file of a change
// (branch vs its merge-base with the default branch) — the lazy per-file
// companion to ChangeDetailPayload (contract §3, GET /review/{projectHash}/diff
// ?branch=&file=). Binary files come back Binary=true with no hunks; files
// exceeding the size cap come back Truncated.
type ChangeDiffPayload struct {
	Branch    string     `json:"branch"`
	File      string     `json:"file"` // the new path
	OldPath   *string    `json:"oldPath,omitempty"`
	Status    string     `json:"status"` // "M" | "A" | "D" | "R"
	Binary    bool       `json:"binary"`
	Truncated bool       `json:"truncated"`
	Hunks     []DiffHunk `json:"hunks"`
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

// DiffLine is one line within a hunk. Kind is "context" | "add" | "del";
// Text excludes the leading +/-/space marker.
type DiffLine struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
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
