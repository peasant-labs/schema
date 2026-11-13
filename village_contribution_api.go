package schema

import (
	jsonschema "github.com/swaggest/jsonschema-go"
)

// The paged contribution view bounds one page by the number of root sessions,
// never by the number of transcripts. A root is a session with no
// parent_session_id, or whose parent is not among the owner's published
// transcripts. A page carries each of its roots together with the whole family
// of helpers that descend from them, so a page can hold far more transcripts
// than its limit.
const (
	// VillageContributionPageLimitDefault is the number of roots a page holds
	// when the caller names no limit.
	VillageContributionPageLimitDefault = 500
	// VillageContributionPageLimitMinimum is the smallest page a caller may ask for.
	VillageContributionPageLimitMinimum = 1
	// VillageContributionPageLimitMaximum is the largest page a caller may ask
	// for, counted in roots.
	VillageContributionPageLimitMaximum = 500
)

// VillageContributableView is the closed view menu of the contributable
// operation. Omitting the parameter keeps the flat whole-library response, so
// the flat response has no token. The menu is distinct from the grouped-view
// menu that the pending, my-shares, transcript and detail reads share, which
// accepts only grouped.
type VillageContributableView string

const (
	// VillageContributableViewGrouped selects the grouped helper-list envelope.
	VillageContributableViewGrouped VillageContributableView = "grouped"
	// VillageContributableViewPaged selects one page of root sessions, each with
	// its family of helpers.
	VillageContributableViewPaged VillageContributableView = "paged"
	// VillageContributableViewSnapshot selects the compact record of every
	// session in the matching families of one project.
	VillageContributableViewSnapshot VillageContributableView = "snapshot"
)

var AllVillageContributableViews = []VillageContributableView{
	VillageContributableViewGrouped,
	VillageContributableViewPaged,
	VillageContributableViewSnapshot,
}

func (v VillageContributableView) IsValid() bool { return inSet(v, AllVillageContributableViews) }

func NewVillageContributableView(raw string) (VillageContributableView, error) {
	return newClosedValue(raw, "VillageContributableView", AllVillageContributableViews)
}

func (VillageContributableView) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Contributable View", "The view of the contributable operation. Omission keeps the flat response; grouped, paged and snapshot select the other envelopes. Any other value is refused with a 400 that names the accepted values.", AllVillageContributableViews), nil
}

// VillageContributionErrorCode is the closed machine-readable reason a
// contribution request was refused. It rides beside the human-readable error
// text; a client branches on the code and never parses the text.
type VillageContributionErrorCode string

const (
	// VillageContributionErrorMoved: a root now belongs to another project; the
	// detail names it in current_project_hash.
	VillageContributionErrorMoved VillageContributionErrorCode = "moved"
	// VillageContributionErrorMissing: the transcript is not among the caller's
	// published transcripts.
	VillageContributionErrorMissing VillageContributionErrorCode = "missing"
	// VillageContributionErrorLive: the transcript is already live in the collective.
	VillageContributionErrorLive VillageContributionErrorCode = "live"
	// VillageContributionErrorPrivateUnapproved: a private transcript was
	// selected without a matching private approval.
	VillageContributionErrorPrivateUnapproved VillageContributionErrorCode = "private_unapproved"
	// VillageContributionErrorConflict: the transcript changed in a way that
	// conflicts with the request.
	VillageContributionErrorConflict VillageContributionErrorCode = "conflict"
	// VillageContributionErrorPolicy: the collective's policy refuses the
	// request as a whole. Its detail omits transcript_id.
	VillageContributionErrorPolicy VillageContributionErrorCode = "policy"
	// VillageContributionErrorMembership: the caller may not contribute to the
	// collective.
	VillageContributionErrorMembership VillageContributionErrorCode = "membership"
	// VillageContributionErrorStaleSelection: the selection no longer matches
	// the library; read it again.
	VillageContributionErrorStaleSelection VillageContributionErrorCode = "stale_selection"
	// VillageContributionErrorSnapshotTimeout: the snapshot did not finish in time.
	VillageContributionErrorSnapshotTimeout VillageContributionErrorCode = "snapshot_timeout"
	// VillageContributionErrorSnapshotBusy: the server is already building the
	// maximum number of snapshots.
	VillageContributionErrorSnapshotBusy VillageContributionErrorCode = "snapshot_busy"
)

var AllVillageContributionErrorCodes = []VillageContributionErrorCode{
	VillageContributionErrorMoved,
	VillageContributionErrorMissing,
	VillageContributionErrorLive,
	VillageContributionErrorPrivateUnapproved,
	VillageContributionErrorConflict,
	VillageContributionErrorPolicy,
	VillageContributionErrorMembership,
	VillageContributionErrorStaleSelection,
	VillageContributionErrorSnapshotTimeout,
	VillageContributionErrorSnapshotBusy,
}

func (c VillageContributionErrorCode) IsValid() bool {
	return inSet(c, AllVillageContributionErrorCodes)
}

func NewVillageContributionErrorCode(raw string) (VillageContributionErrorCode, error) {
	return newClosedValue(raw, "VillageContributionErrorCode", AllVillageContributionErrorCodes)
}

func (VillageContributionErrorCode) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Contribution Error Code", "Machine-readable reason a contribution request was refused", AllVillageContributionErrorCodes), nil
}

// VillageContributionErrorDetail is one per-transcript reason inside a refusal.
// A request-level reason such as policy omits transcript_id.
type VillageContributionErrorDetail struct {
	TranscriptID *TranscriptID                `json:"transcript_id,omitempty"`
	Code         VillageContributionErrorCode `json:"code"`
	// CurrentProjectHash names the project a moved transcript now belongs to.
	CurrentProjectHash *ProjectHash `json:"current_project_hash,omitempty"`
}

// VillageContributionFamilyMember is one session of a contribution family: the
// contributable row, plus whether it matched the request's q search. A family
// appears when its root or any helper matches, so a member can be present only
// because another member matched.
type VillageContributionFamilyMember struct {
	// Transcript is the member's own row. Its project_hash and
	// project_display_name are the member's own project, which may differ from
	// the project of the family's root. already_shared on the row marks a
	// member that is already live; a live member stays locked in a selection.
	Transcript VillageContributableTranscript `json:"transcript"`
	// Matched is true when this member itself matched q. It is absent when the
	// request carried no q.
	Matched *bool `json:"matched,omitempty"`
}

// VillageContributionPageItem is one family: a root session and every helper
// that descends from it. The family belongs to its root's project, whatever
// project each helper is in. Selecting the root selects its helpers by default.
type VillageContributionPageItem struct {
	Root    VillageContributionFamilyMember   `json:"root"`
	Helpers []VillageContributionFamilyMember `json:"helpers" nullable:"false"`
}

// VillageContributionPageResponse is one page of families. Its size is counted
// in root sessions. The page carries no global counts; the project picker is
// the count authority.
type VillageContributionPageResponse struct {
	GroupID VillageUUID                   `json:"group_id"`
	Items   []VillageContributionPageItem `json:"items" nullable:"false"`
	// NextCursor is the opaque cursor of the following page. It is absent on
	// the last page.
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

// VillageContributionSnapshotRecord is the compact record of one session in a
// snapshot: its identity, visibility, latest contribution status, own project
// and parent link, and nothing else.
type VillageContributionSnapshotRecord struct {
	ID         TranscriptID                `json:"id"`
	Visibility VillageTranscriptVisibility `json:"visibility"`
	// LatestStatus is the status of the latest contribution attempt; absence
	// means no attempt.
	LatestStatus *VillageShareStatus `json:"latest_status,omitempty"`
	// ProjectHash is the session's own project, which may differ from the
	// project of the root of its family.
	ProjectHash        ProjectHash   `json:"project_hash"`
	ParentTranscriptID *TranscriptID `json:"parent_transcript_id,omitempty"`
}

// VillageContributionSnapshotResponse is the exact selection set of one
// project: every session in every matching family whose root is in that
// project. It has no record cap and no byte cap; the server may refuse with
// snapshot_timeout or snapshot_busy. Count equals len(Records) and counts
// sessions, not roots.
type VillageContributionSnapshotResponse struct {
	GroupID     VillageUUID                         `json:"group_id"`
	ProjectHash ProjectHash                         `json:"project_hash"`
	Records     []VillageContributionSnapshotRecord `json:"records" nullable:"false"`
	Count       int                                 `json:"count"`
}

// VillageContributionProjectSummary is one project in the picker. Its counts
// are in sessions: every transcript in the matching families whose root is in
// the project.
type VillageContributionProjectSummary struct {
	ProjectHash        ProjectHash              `json:"project_hash"`
	ProjectDisplayName string                   `json:"project_display_name"`
	ProjectNameSource  VillageProjectNameSource `json:"project_name_source"`
	RepositoryState    *string                  `json:"repository_state,omitempty"`
	// MatchingCount is the sessions in the project's matching families.
	MatchingCount int `json:"matching_count"`
	// SelectableCount is the sessions among them that are not already live.
	SelectableCount int `json:"selectable_count"`
}

// VillageContributionProviderFacet counts the matching sessions of one model
// provider.
type VillageContributionProviderFacet struct {
	Provider string `json:"provider"`
	Count    int    `json:"count"`
}

// VillageContributionProjectPage is one page of the project picker. Its page
// size is counted in projects; MatchingTotal, SelectableTotal and the provider
// counts are in sessions across every page.
type VillageContributionProjectPage struct {
	Projects   []VillageContributionProjectSummary `json:"projects" nullable:"false"`
	NextCursor *string                             `json:"next_cursor,omitempty"`
	HasMore    bool                                `json:"has_more"`
	// MatchingTotal is the sessions in all matching families.
	MatchingTotal int `json:"matching_total"`
	// SelectableTotal is the sessions among them that are not already live.
	SelectableTotal int                                `json:"selectable_total"`
	Providers       []VillageContributionProviderFacet `json:"providers" nullable:"false"`
}

// VillageBatchShareSelectionRequest writes one exact selection. TranscriptIDs
// lists every session to offer, roots and helpers alike; the server does not
// expand a root into its helpers. PrivateApprovals lists the private
// transcripts among them whose owner approved their visibility.
type VillageBatchShareSelectionRequest struct {
	ProjectHash      ProjectHash    `json:"project_hash"`
	TranscriptIDs    []TranscriptID `json:"transcript_ids" nullable:"false" minItems:"1"`
	PrivateApprovals []TranscriptID `json:"private_approvals" nullable:"false"`
}
