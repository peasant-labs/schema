package schema

import (
	"time"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// VillageUUID is the canonical lowercase UUID string form the Village emits for
// users, collectives, repositories, and other server-owned rows.
type VillageUUID string

func (id VillageUUID) String() string { return string(id) }

// JSONSchema implements jsonschema.Exposer.
func (VillageUUID) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithFormat("uuid")
	s.WithTitle("Village UUID")
	s.WithDescription("Village-side canonical lowercase UUID identifier")
	s.WithPattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	s.WithExamples("123e4567-e89b-12d3-a456-426614174000")
	return s, nil
}

// VillageErrorResponse is the common JSON refusal envelope emitted by Village
// handlers through writeError.
type VillageErrorResponse struct {
	Error string `json:"error"`
}

// VillageGroupAcceptanceMode is the closed acceptance policy menu for a
// collective.
type VillageGroupAcceptanceMode string

const (
	VillageGroupAcceptanceOpen         VillageGroupAcceptanceMode = "open"
	VillageGroupAcceptanceVerifiedOnly VillageGroupAcceptanceMode = "verified_only"
	VillageGroupAcceptanceCurated      VillageGroupAcceptanceMode = "curated"
)

var AllVillageGroupAcceptanceModes = []VillageGroupAcceptanceMode{
	VillageGroupAcceptanceOpen,
	VillageGroupAcceptanceVerifiedOnly,
	VillageGroupAcceptanceCurated,
}

func (VillageGroupAcceptanceMode) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Group Acceptance Mode", "How a collective accepts new members and contributions", AllVillageGroupAcceptanceModes), nil
}

// VillageGroupDataAccess is the closed read policy menu for a collective's
// pooled transcript data.
type VillageGroupDataAccess string

const (
	VillageGroupDataAccessMembersOnly  VillageGroupDataAccess = "members_only"
	VillageGroupDataAccessContributors VillageGroupDataAccess = "contributors"
	VillageGroupDataAccessPublic       VillageGroupDataAccess = "public"
)

var AllVillageGroupDataAccessPolicies = []VillageGroupDataAccess{
	VillageGroupDataAccessMembersOnly,
	VillageGroupDataAccessContributors,
	VillageGroupDataAccessPublic,
}

func (VillageGroupDataAccess) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Group Data Access", "Who may read a collective's pooled transcript data", AllVillageGroupDataAccessPolicies), nil
}

// VillageTranscriptDeletionPolicy is the closed policy menu used when a member
// leaves a collective.
type VillageTranscriptDeletionPolicy string

const (
	VillageTranscriptDeletionUserChoice VillageTranscriptDeletionPolicy = "user_choice"
	VillageTranscriptDeletionMandatory  VillageTranscriptDeletionPolicy = "mandatory"
)

var AllVillageTranscriptDeletionPolicies = []VillageTranscriptDeletionPolicy{
	VillageTranscriptDeletionUserChoice,
	VillageTranscriptDeletionMandatory,
}

func (VillageTranscriptDeletionPolicy) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Transcript Deletion Policy", "Whether leaving a collective retracts contributed transcripts by default", AllVillageTranscriptDeletionPolicies), nil
}

// VillageGroupRole is the role stored for a user in one collective.
type VillageGroupRole string

const (
	VillageGroupRoleOwner       VillageGroupRole = "owner"
	VillageGroupRoleMember      VillageGroupRole = "member"
	VillageGroupRoleContributor VillageGroupRole = "contributor"
	VillageGroupRolePending     VillageGroupRole = "pending"
)

var AllVillageGroupRoles = []VillageGroupRole{
	VillageGroupRoleOwner,
	VillageGroupRoleMember,
	VillageGroupRoleContributor,
	VillageGroupRolePending,
}

func (VillageGroupRole) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Group Role", "A user's role in one collective", AllVillageGroupRoles), nil
}

// VillageAssignableGroupRole is the smaller role menu accepted by the member
// role update endpoint.
type VillageAssignableGroupRole string

const (
	VillageAssignableGroupRoleContributor VillageAssignableGroupRole = "contributor"
	VillageAssignableGroupRoleMember      VillageAssignableGroupRole = "member"
)

var AllVillageAssignableGroupRoles = []VillageAssignableGroupRole{
	VillageAssignableGroupRoleContributor,
	VillageAssignableGroupRoleMember,
}

func (VillageAssignableGroupRole) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Assignable Group Role", "Collective roles an owner may assign through the member role endpoint", AllVillageAssignableGroupRoles), nil
}

// VillageGroupViewerRole is the viewer's role in one collective. The empty
// value is the current wire value for a signed-out or non-member viewer on the
// group-detail response.
type VillageGroupViewerRole string

const (
	VillageGroupViewerRoleNone        VillageGroupViewerRole = ""
	VillageGroupViewerRoleOwner       VillageGroupViewerRole = "owner"
	VillageGroupViewerRoleMember      VillageGroupViewerRole = "member"
	VillageGroupViewerRoleContributor VillageGroupViewerRole = "contributor"
	VillageGroupViewerRolePending     VillageGroupViewerRole = "pending"
)

var AllVillageGroupViewerRoles = []VillageGroupViewerRole{
	VillageGroupViewerRoleNone,
	VillageGroupViewerRoleOwner,
	VillageGroupViewerRoleMember,
	VillageGroupViewerRoleContributor,
	VillageGroupViewerRolePending,
}

func (VillageGroupViewerRole) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Group Viewer Role", "The caller's role in a collective, or an empty string when the caller has none", AllVillageGroupViewerRoles), nil
}

// VillageProjectNameSource is the closed set of tiers that can produce a
// resolved project display name.
type VillageProjectNameSource string

const (
	VillageProjectNameSourceOverride  VillageProjectNameSource = "override"
	VillageProjectNameSourceConsented VillageProjectNameSource = "consented"
	VillageProjectNameSourceRemote    VillageProjectNameSource = "remote"
	VillageProjectNameSourcePath      VillageProjectNameSource = "path"
	VillageProjectNameSourcePrivacy   VillageProjectNameSource = "privacy"
)

var AllVillageProjectNameSources = []VillageProjectNameSource{
	VillageProjectNameSourceOverride,
	VillageProjectNameSourceConsented,
	VillageProjectNameSourceRemote,
	VillageProjectNameSourcePath,
	VillageProjectNameSourcePrivacy,
}

func (VillageProjectNameSource) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Project Name Source", "Which source tier produced a resolved project display name", AllVillageProjectNameSources), nil
}

// VillageTranscriptVisibility is the current Village transcript visibility menu.
// It deliberately uses the Village token shared, not the publish-contract token
// group used by Visibility.
type VillageTranscriptVisibility string

const (
	VillageTranscriptVisibilityPrivate VillageTranscriptVisibility = "private"
	VillageTranscriptVisibilityShared  VillageTranscriptVisibility = "shared"
	VillageTranscriptVisibilityPublic  VillageTranscriptVisibility = "public"
)

var AllVillageTranscriptVisibilities = []VillageTranscriptVisibility{
	VillageTranscriptVisibilityPrivate,
	VillageTranscriptVisibilityShared,
	VillageTranscriptVisibilityPublic,
}

func (VillageTranscriptVisibility) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Transcript Visibility", "Village transcript visibility as stored and served by web routes", AllVillageTranscriptVisibilities), nil
}

// VillageContributionStatus is the status a new contribution can open in.
type VillageContributionStatus string

const (
	VillageContributionStatusApproved VillageContributionStatus = "approved"
	VillageContributionStatusPending  VillageContributionStatus = "pending"
)

var AllVillageContributionStatuses = []VillageContributionStatus{
	VillageContributionStatusApproved,
	VillageContributionStatusPending,
}

func (VillageContributionStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Contribution Status", "Status assigned when a new collective contribution is opened", AllVillageContributionStatuses), nil
}

// VillageReviewDecision is the closed decision menu for collective review
// actions.
type VillageReviewDecision string

const (
	VillageReviewDecisionApproved VillageReviewDecision = "approved"
	VillageReviewDecisionRejected VillageReviewDecision = "rejected"
)

var AllVillageReviewDecisions = []VillageReviewDecision{
	VillageReviewDecisionApproved,
	VillageReviewDecisionRejected,
}

func (VillageReviewDecision) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Review Decision", "Decision applied by a collective owner to pending submissions", AllVillageReviewDecisions), nil
}

// VillageShareStatus is the full share-attempt ledger status menu. Pending,
// approved, and rejected can appear in current projections; retracted and
// revoked are terminal states kept for the historical event ledger.
type VillageShareStatus string

const (
	VillageShareStatusPending   VillageShareStatus = "pending"
	VillageShareStatusApproved  VillageShareStatus = "approved"
	VillageShareStatusRejected  VillageShareStatus = "rejected"
	VillageShareStatusRetracted VillageShareStatus = "retracted"
	VillageShareStatusRevoked   VillageShareStatus = "revoked"
)

var AllVillageShareStatuses = []VillageShareStatus{
	VillageShareStatusPending,
	VillageShareStatusApproved,
	VillageShareStatusRejected,
	VillageShareStatusRetracted,
	VillageShareStatusRevoked,
}

func (VillageShareStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Village Share Status",
		"Status of one collective share-attempt event. Pending, approved, and rejected can appear in current projections; retracted and revoked are terminal ledger states.",
		AllVillageShareStatuses,
	), nil
}

// VillageShareEventActor is the actor class attached to one share event. It is
// an actor class, not a user identity.
type VillageShareEventActor string

const (
	VillageShareEventActorNone       VillageShareEventActor = ""
	VillageShareEventActorOwner      VillageShareEventActor = "owner"
	VillageShareEventActorCollective VillageShareEventActor = "collective"
	VillageShareEventActorModerator  VillageShareEventActor = "moderator"
)

var AllVillageShareEventActors = []VillageShareEventActor{
	VillageShareEventActorNone,
	VillageShareEventActorOwner,
	VillageShareEventActorCollective,
	VillageShareEventActorModerator,
}

func (VillageShareEventActor) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Share Event Actor", "Actor class that decided one collective share event", AllVillageShareEventActors), nil
}

// VillageGroup is the full collective row returned by create, update, and detail
// routes.
type VillageGroup struct {
	ID                       VillageUUID                     `json:"id"`
	Name                     string                          `json:"name"`
	Description              *string                         `json:"description"`
	CreatedBy                VillageUUID                     `json:"created_by"`
	CreatedAt                time.Time                       `json:"created_at"`
	UpdatedAt                time.Time                       `json:"updated_at"`
	AcceptanceMode           VillageGroupAcceptanceMode      `json:"acceptance_mode"`
	DataAccess               VillageGroupDataAccess          `json:"data_access"`
	LinkedGithubOrg          *string                         `json:"linked_github_org"`
	DisplayMembers           bool                            `json:"display_members"`
	TranscriptDeletionPolicy VillageTranscriptDeletionPolicy `json:"transcript_deletion_policy"`
}

// VillagePublicGroup is the compact row returned by GET /api/v1/groups/public.
type VillagePublicGroup struct {
	ID              VillageUUID                `json:"id"`
	Name            string                     `json:"name"`
	Description     *string                    `json:"description"`
	AcceptanceMode  VillageGroupAcceptanceMode `json:"acceptance_mode"`
	DataAccess      VillageGroupDataAccess     `json:"data_access"`
	LinkedGithubOrg *string                    `json:"linked_github_org"`
	CreatedAt       time.Time                  `json:"created_at"`
	MemberCount     int32                      `json:"member_count"`
	TranscriptCount int32                      `json:"transcript_count"`
}

// VillageUserGroup is one collective the signed-in caller belongs to.
type VillageUserGroup struct {
	ID                       VillageUUID                     `json:"id"`
	Name                     string                          `json:"name"`
	Description              *string                         `json:"description"`
	CreatedBy                VillageUUID                     `json:"created_by"`
	CreatedAt                time.Time                       `json:"created_at"`
	UpdatedAt                time.Time                       `json:"updated_at"`
	AcceptanceMode           VillageGroupAcceptanceMode      `json:"acceptance_mode"`
	DataAccess               VillageGroupDataAccess          `json:"data_access"`
	LinkedGithubOrg          *string                         `json:"linked_github_org"`
	DisplayMembers           bool                            `json:"display_members"`
	TranscriptDeletionPolicy VillageTranscriptDeletionPolicy `json:"transcript_deletion_policy"`
	Role                     VillageGroupRole                `json:"role"`
	MemberSince              time.Time                       `json:"member_since"`
	MemberCount              int32                           `json:"member_count"`
	TranscriptCount          int32                           `json:"transcript_count"`
}

// VillageVisibleGroup is one collective the signed-in caller may see. Role and
// member_since are null when the caller can see the collective but is not a
// member.
type VillageVisibleGroup struct {
	ID                       VillageUUID                     `json:"id"`
	Name                     string                          `json:"name"`
	Description              *string                         `json:"description"`
	CreatedBy                VillageUUID                     `json:"created_by"`
	CreatedAt                time.Time                       `json:"created_at"`
	UpdatedAt                time.Time                       `json:"updated_at"`
	AcceptanceMode           VillageGroupAcceptanceMode      `json:"acceptance_mode"`
	DataAccess               VillageGroupDataAccess          `json:"data_access"`
	LinkedGithubOrg          *string                         `json:"linked_github_org"`
	DisplayMembers           bool                            `json:"display_members"`
	TranscriptDeletionPolicy VillageTranscriptDeletionPolicy `json:"transcript_deletion_policy"`
	Role                     *VillageGroupRole               `json:"role"`
	MemberSince              *time.Time                      `json:"member_since"`
	MemberCount              int32                           `json:"member_count"`
	TranscriptCount          int32                           `json:"transcript_count"`
}

// VillageCollectiveSearchResult is one collective returned by search surfaces.
type VillageCollectiveSearchResult struct {
	ID              VillageUUID `json:"id"`
	Name            string      `json:"name"`
	Description     *string     `json:"description"`
	LinkedGithubOrg *string     `json:"linked_github_org"`
	MemberCount     int32       `json:"member_count"`
	TranscriptCount int32       `json:"transcript_count"`
}

type VillageCollectiveSearchResponse struct {
	Collectives []VillageCollectiveSearchResult `json:"collectives" nullable:"false"`
}

type VillageCreateGroupRequest struct {
	Name            string                     `json:"name"`
	Description     string                     `json:"description,omitempty"`
	AcceptanceMode  VillageGroupAcceptanceMode `json:"acceptance_mode,omitempty"`
	DataAccess      VillageGroupDataAccess     `json:"data_access,omitempty"`
	LinkedGithubOrg string                     `json:"linked_github_org,omitempty"`
}

// VillageUpdateGroupRequest mirrors the current update handler. Name and
// description are decoded as strings and are written even when omitted, so a
// caller that wants to preserve them must send their current values.
type VillageUpdateGroupRequest struct {
	Name                     string                          `json:"name,omitempty"`
	Description              string                          `json:"description,omitempty"`
	DataAccess               VillageGroupDataAccess          `json:"data_access,omitempty"`
	AcceptanceMode           VillageGroupAcceptanceMode      `json:"acceptance_mode,omitempty"`
	LinkedGithubOrg          *string                         `json:"linked_github_org,omitempty"`
	DisplayMembers           *bool                           `json:"display_members,omitempty"`
	TranscriptDeletionPolicy VillageTranscriptDeletionPolicy `json:"transcript_deletion_policy,omitempty"`
}

type VillageStatusResponse struct {
	Status string `json:"status"`
}

type VillageGroupStatusRoleResponse struct {
	Status string           `json:"status"`
	Role   VillageGroupRole `json:"role"`
}

type VillageRemoveGroupMemberResponse struct {
	Status    string `json:"status"`
	Retracted bool   `json:"retracted"`
}

type VillageGroupMemberUsernameRequest struct {
	Username string `json:"username"`
}

type VillageGroupMemberRoleRequest struct {
	Role VillageAssignableGroupRole `json:"role"`
}

// VillageGroupMember is one member row in a collective roster.
type VillageGroupMember struct {
	Role           VillageGroupRole `json:"role"`
	JoinedAt       time.Time        `json:"joined_at"`
	ID             VillageUUID      `json:"id"`
	GithubUsername string           `json:"github_username"`
	DisplayName    *string          `json:"display_name"`
	AvatarURL      *string          `json:"avatar_url"`
	GithubOrgs     []string         `json:"github_orgs" nullable:"false"`
}

type VillageGroupTranscriptStats struct {
	TotalTranscripts int32 `json:"total_transcripts"`
	ContributorCount int32 `json:"contributor_count"`
	TotalTurns       int64 `json:"total_turns"`
	TotalDurationMs  int64 `json:"total_duration_ms"`
	TotalTokens      int64 `json:"total_tokens"`
}

type VillageGroupModelBreakdown struct {
	ModelProvider   string `json:"model_provider"`
	TranscriptCount int32  `json:"transcript_count"`
}

type VillageGroupContributor struct {
	ID              VillageUUID `json:"id"`
	GithubUsername  string      `json:"github_username"`
	AvatarURL       *string     `json:"avatar_url"`
	TranscriptCount int32       `json:"transcript_count"`
}

type VillageGroupDetailResponse struct {
	Group          VillageGroup                 `json:"group"`
	Members        []VillageGroupMember         `json:"members" nullable:"false"`
	Stats          VillageGroupTranscriptStats  `json:"stats"`
	Models         []VillageGroupModelBreakdown `json:"models" nullable:"false"`
	Contributors   []VillageGroupContributor    `json:"contributors" nullable:"false"`
	CanRead        bool                         `json:"can_read"`
	YourRole       VillageGroupViewerRole       `json:"your_role"`
	Transcripts    []VillageGroupTranscript     `json:"transcripts" nullable:"false"`
	PendingMembers []VillageGroupMember         `json:"pending_members,omitempty" nullable:"false"`
}

// VillageTranscript is the current transcript row projection used by Village web
// list and collective routes. Storage object keys and source-machine paths are
// not part of this public projection.
type VillageTranscript struct {
	ID                      TranscriptID                `json:"id"`
	OwnerID                 VillageUUID                 `json:"owner_id"`
	LocalID                 SessionID                   `json:"local_id"`
	Title                   *string                     `json:"title"`
	Description             *string                     `json:"description"`
	Visibility              VillageTranscriptVisibility `json:"visibility"`
	ModelProvider           string                      `json:"model_provider"`
	ModelName               *string                     `json:"model_name"`
	HarnessVersion          *string                     `json:"harness_version"`
	SessionStart            *time.Time                  `json:"session_start"`
	SessionEnd              *time.Time                  `json:"session_end"`
	TurnCount               *int32                      `json:"turn_count"`
	TokenCount              *int32                      `json:"token_count"`
	BlobSizeBytes           *int64                      `json:"blob_size_bytes"`
	SchemaVersion           string                      `json:"schema_version"`
	PublishedAt             time.Time                   `json:"published_at"`
	UpdatedAt               time.Time                   `json:"updated_at"`
	ParentSessionID         *SessionID                  `json:"parent_session_id"`
	IngestedAt              *time.Time                  `json:"ingested_at"`
	SourceFormat            *SourceFormat               `json:"source_format"`
	GitBranch               *string                     `json:"git_branch"`
	GitRemote               *string                     `json:"git_remote"`
	ProjectHash             ProjectHash                 `json:"project_hash"`
	ProjectName             *string                     `json:"project_name"`
	ProjectDisplayName      string                      `json:"project_display_name"`
	ProjectNameSource       VillageProjectNameSource    `json:"project_name_source"`
	ProjectRemoteLabel      string                      `json:"project_remote_label"`
	ToolCallCount           *int32                      `json:"tool_call_count"`
	SubagentCount           *int32                      `json:"subagent_count"`
	DurationMs              *int64                      `json:"duration_ms"`
	Subagents               *[]map[string]interface{}   `json:"subagents"`
	DiagnosticsWarnings     *[]string                   `json:"diagnostics_warnings"`
	DiagnosticsPartial      *bool                       `json:"diagnostics_partial"`
	TokensIn                *int64                      `json:"tokens_in"`
	TokensOut               *int64                      `json:"tokens_out"`
	TitleGenerated          *string                     `json:"title_generated"`
	Outcome                 *SessionOutcome             `json:"outcome"`
	FilesTouched            *int32                      `json:"files_touched"`
	LinesChanged            *int32                      `json:"lines_changed"`
	RetryLoops              *int32                      `json:"retry_loops"`
	RetryTokensWasted       *int32                      `json:"retry_tokens_wasted"`
	WithinSessionReverts    *int32                      `json:"within_session_reverts"`
	SignalDensity           *float32                    `json:"signal_density"`
	SpecQualityScore        *float32                    `json:"spec_quality_score"`
	ExplorationRatio        *float32                    `json:"exploration_ratio"`
	ScopeBreadth            *int32                      `json:"scope_breadth"`
	DiscoveryTurns          *int32                      `json:"discovery_turns"`
	M2TokenOutcomeRatio     *float32                    `json:"m2_token_outcome_ratio"`
	M3UniqueToolCount       *int32                      `json:"m3_unique_tool_count"`
	M4ErrorRecoveryCount    *int32                      `json:"m4_error_recovery_count"`
	M4ConsecutiveErrorMax   *int32                      `json:"m4_consecutive_error_max"`
	M5ContextUtilizationPct *float32                    `json:"m5_context_utilization_pct"`
	M5PeakContextTokens     *int32                      `json:"m5_peak_context_tokens"`
	M5AvgMessageTokens      *int32                      `json:"m5_avg_message_tokens"`
	M6OutputSurvivalPct     *float32                    `json:"m6_output_survival_pct"`
	M6LinesSurvived         *int32                      `json:"m6_lines_survived"`
	M6LinesTotal            *int32                      `json:"m6_lines_total"`
	M7SpecWordCount         *int32                      `json:"m7_spec_word_count"`
	M7SpecHasExamples       *bool                       `json:"m7_spec_has_examples"`
	M7SpecHasConstraints    *bool                       `json:"m7_spec_has_constraints"`
	ComputedAt              *time.Time                  `json:"computed_at"`
	ComputeVersion          *int32                      `json:"compute_version"`
	ContentHash             *TranscriptContentHash      `json:"content_hash"`
	LicenseID               *License                    `json:"license_id"`
	SessionOrigin           SessionOrigin               `json:"session_origin"`
}

// VillageGroupTranscript deliberately lists the transcript fields instead of
// embedding VillageTranscript. Village flattens its embedded response on the
// JSON wire, and this shape preserves required fields in OpenAPI and TypeScript
// generation.
type VillageGroupTranscript struct {
	ID                      TranscriptID                `json:"id"`
	OwnerID                 VillageUUID                 `json:"owner_id"`
	LocalID                 SessionID                   `json:"local_id"`
	Title                   *string                     `json:"title"`
	Description             *string                     `json:"description"`
	Visibility              VillageTranscriptVisibility `json:"visibility"`
	ModelProvider           string                      `json:"model_provider"`
	ModelName               *string                     `json:"model_name"`
	HarnessVersion          *string                     `json:"harness_version"`
	SessionStart            *time.Time                  `json:"session_start"`
	SessionEnd              *time.Time                  `json:"session_end"`
	TurnCount               *int32                      `json:"turn_count"`
	TokenCount              *int32                      `json:"token_count"`
	BlobSizeBytes           *int64                      `json:"blob_size_bytes"`
	SchemaVersion           string                      `json:"schema_version"`
	PublishedAt             time.Time                   `json:"published_at"`
	UpdatedAt               time.Time                   `json:"updated_at"`
	ParentSessionID         *SessionID                  `json:"parent_session_id"`
	IngestedAt              *time.Time                  `json:"ingested_at"`
	SourceFormat            *SourceFormat               `json:"source_format"`
	GitBranch               *string                     `json:"git_branch"`
	GitRemote               *string                     `json:"git_remote"`
	ProjectHash             ProjectHash                 `json:"project_hash"`
	ProjectName             *string                     `json:"project_name"`
	ProjectDisplayName      string                      `json:"project_display_name"`
	ProjectNameSource       VillageProjectNameSource    `json:"project_name_source"`
	ProjectRemoteLabel      string                      `json:"project_remote_label"`
	ToolCallCount           *int32                      `json:"tool_call_count"`
	SubagentCount           *int32                      `json:"subagent_count"`
	DurationMs              *int64                      `json:"duration_ms"`
	Subagents               *[]map[string]interface{}   `json:"subagents"`
	DiagnosticsWarnings     *[]string                   `json:"diagnostics_warnings"`
	DiagnosticsPartial      *bool                       `json:"diagnostics_partial"`
	TokensIn                *int64                      `json:"tokens_in"`
	TokensOut               *int64                      `json:"tokens_out"`
	TitleGenerated          *string                     `json:"title_generated"`
	Outcome                 *SessionOutcome             `json:"outcome"`
	FilesTouched            *int32                      `json:"files_touched"`
	LinesChanged            *int32                      `json:"lines_changed"`
	RetryLoops              *int32                      `json:"retry_loops"`
	RetryTokensWasted       *int32                      `json:"retry_tokens_wasted"`
	WithinSessionReverts    *int32                      `json:"within_session_reverts"`
	SignalDensity           *float32                    `json:"signal_density"`
	SpecQualityScore        *float32                    `json:"spec_quality_score"`
	ExplorationRatio        *float32                    `json:"exploration_ratio"`
	ScopeBreadth            *int32                      `json:"scope_breadth"`
	DiscoveryTurns          *int32                      `json:"discovery_turns"`
	M2TokenOutcomeRatio     *float32                    `json:"m2_token_outcome_ratio"`
	M3UniqueToolCount       *int32                      `json:"m3_unique_tool_count"`
	M4ErrorRecoveryCount    *int32                      `json:"m4_error_recovery_count"`
	M4ConsecutiveErrorMax   *int32                      `json:"m4_consecutive_error_max"`
	M5ContextUtilizationPct *float32                    `json:"m5_context_utilization_pct"`
	M5PeakContextTokens     *int32                      `json:"m5_peak_context_tokens"`
	M5AvgMessageTokens      *int32                      `json:"m5_avg_message_tokens"`
	M6OutputSurvivalPct     *float32                    `json:"m6_output_survival_pct"`
	M6LinesSurvived         *int32                      `json:"m6_lines_survived"`
	M6LinesTotal            *int32                      `json:"m6_lines_total"`
	M7SpecWordCount         *int32                      `json:"m7_spec_word_count"`
	M7SpecHasExamples       *bool                       `json:"m7_spec_has_examples"`
	M7SpecHasConstraints    *bool                       `json:"m7_spec_has_constraints"`
	ComputedAt              *time.Time                  `json:"computed_at"`
	ComputeVersion          *int32                      `json:"compute_version"`
	ContentHash             *TranscriptContentHash      `json:"content_hash"`
	LicenseID               *License                    `json:"license_id"`
	SessionOrigin           SessionOrigin               `json:"session_origin"`
	OwnerUsername           string                      `json:"owner_username"`
	OwnerAvatarURL          *string                     `json:"owner_avatar_url"`
	OwnerIsDiscoverable     bool                        `json:"owner_is_discoverable"`
}

type VillageTranscriptShare struct {
	GroupID   VillageUUID `json:"group_id"`
	GroupName string      `json:"group_name"`
	SharedAt  time.Time   `json:"shared_at"`
}

type VillageShareTranscriptRequest struct {
	GroupIDs []VillageUUID `json:"group_ids" nullable:"false"`
}

type VillageContributedCollective struct {
	ID                    VillageUUID `json:"id"`
	Name                  string      `json:"name"`
	Description           *string     `json:"description"`
	LinkedGithubOrg       *string     `json:"linked_github_org"`
	ApprovedCount         int32       `json:"approved_count"`
	PendingCount          int32       `json:"pending_count"`
	RejectedAttemptCount  int32       `json:"rejected_attempt_count"`
	WithdrawnAttemptCount int32       `json:"withdrawn_attempt_count"`
}

type VillageContributedCollectivesResponse struct {
	Collectives []VillageContributedCollective `json:"collectives" nullable:"false"`
}

type VillageTranscriptCollective struct {
	ID              VillageUUID `json:"id"`
	Name            string      `json:"name"`
	Description     *string     `json:"description"`
	LinkedGithubOrg *string     `json:"linked_github_org"`
	SharedAt        time.Time   `json:"shared_at"`
}

type VillageTranscriptCollectivesResponse struct {
	Collectives []VillageTranscriptCollective `json:"collectives" nullable:"false"`
}

type VillagePendingShare struct {
	TranscriptID        TranscriptID `json:"transcript_id"`
	Title               *string      `json:"title"`
	ModelProvider       string       `json:"model_provider"`
	OwnerID             VillageUUID  `json:"owner_id"`
	LocalID             SessionID    `json:"local_id"`
	ParentSessionID     *SessionID   `json:"parent_session_id"`
	ProjectHash         ProjectHash  `json:"project_hash"`
	ProjectName         *string      `json:"project_name"`
	Branch              *string      `json:"branch"`
	OwnerUsername       string       `json:"owner_username"`
	OwnerIsDiscoverable bool         `json:"owner_is_discoverable"`
	SharedAt            time.Time    `json:"shared_at"`
}

type VillageUserGroupShare struct {
	ID              TranscriptID                `json:"id"`
	Title           *string                     `json:"title"`
	ModelProvider   string                      `json:"model_provider"`
	ModelName       *string                     `json:"model_name"`
	Visibility      VillageTranscriptVisibility `json:"visibility"`
	PublishedAt     time.Time                   `json:"published_at"`
	TurnCount       *int32                      `json:"turn_count"`
	TokensIn        *int64                      `json:"tokens_in"`
	TokensOut       *int64                      `json:"tokens_out"`
	OwnerID         VillageUUID                 `json:"owner_id"`
	LocalID         SessionID                   `json:"local_id"`
	ParentSessionID *SessionID                  `json:"parent_session_id"`
	// Status uses the shared lifecycle enum. Current my-shares responses emit the
	// live subset today; terminal values are reserved for ledger history.
	Status   VillageShareStatus `json:"status"`
	SharedAt time.Time          `json:"shared_at"`
}

type VillageBatchShareRequest struct {
	ProjectHash         ProjectHash    `json:"project_hash"`
	TranscriptIDs       []TranscriptID `json:"transcript_ids,omitempty"`
	VisibilityConfirmed bool           `json:"visibility_confirmed"`
}

type VillageBatchShareEntry struct {
	TranscriptID TranscriptID              `json:"transcript_id"`
	Status       VillageContributionStatus `json:"status"`
}

type VillageBatchShareResponse struct {
	ProjectHash   ProjectHash              `json:"project_hash"`
	Shared        []VillageBatchShareEntry `json:"shared" nullable:"false"`
	AlreadyShared []TranscriptID           `json:"already_shared" nullable:"false"`
}

type VillageBatchReviewRequest struct {
	TranscriptIDs []TranscriptID        `json:"transcript_ids" nullable:"false"`
	Status        VillageReviewDecision `json:"status"`
}

type VillageBatchReviewResponse struct {
	Decided        []TranscriptID `json:"decided" nullable:"false"`
	AlreadyDecided []TranscriptID `json:"already_decided" nullable:"false"`
}

type VillageReviewShareResponse struct {
	Status VillageReviewDecision `json:"status"`
}

type VillageReviewShareRequest struct {
	Status VillageReviewDecision `json:"status"`
}

type VillageContributableTranscript struct {
	ID                 TranscriptID                `json:"id"`
	LocalID            SessionID                   `json:"local_id"`
	Title              *string                     `json:"title"`
	Visibility         VillageTranscriptVisibility `json:"visibility"`
	ProjectHash        ProjectHash                 `json:"project_hash"`
	ProjectDisplayName string                      `json:"project_display_name"`
	ProjectNameSource  VillageProjectNameSource    `json:"project_name_source"`
	GitBranch          *string                     `json:"git_branch"`
	ParentSessionID    *SessionID                  `json:"parent_session_id"`
	SessionOrigin      SessionOrigin               `json:"session_origin"`
	ModelProvider      string                      `json:"model_provider"`
	PublishedAt        time.Time                   `json:"published_at"`
	AlreadyShared      bool                        `json:"already_shared"`
}

type VillageContributableResponse struct {
	GroupID     VillageUUID                      `json:"group_id"`
	Transcripts []VillageContributableTranscript `json:"transcripts" nullable:"false"`
}

type VillageShareEvent struct {
	EventNum       int32                  `json:"event_num"`
	Status         VillageShareStatus     `json:"status"`
	RecordedAt     time.Time              `json:"recorded_at"`
	DecidedAt      *time.Time             `json:"decided_at"`
	DecidedByActor VillageShareEventActor `json:"decided_by_actor"`
}

type VillageCollectiveSubmission struct {
	TranscriptID TranscriptID       `json:"transcript_id"`
	GroupID      VillageUUID        `json:"group_id"`
	Title        *string            `json:"title"`
	EventNum     int32              `json:"event_num"`
	Status       VillageShareStatus `json:"status"`
	RecordedAt   time.Time          `json:"recorded_at"`
}

type VillageLinkRepositoryRequest struct {
	Owner          string `json:"owner"`
	Name           string `json:"name"`
	InstallationID int64  `json:"installation_id"`
}

type VillageLinkedRepository struct {
	ID             VillageUUID `json:"id"`
	GroupID        VillageUUID `json:"group_id"`
	Owner          string      `json:"owner"`
	Name           string      `json:"name"`
	InstallationID int64       `json:"installation_id"`
	IsPrivate      bool        `json:"is_private"`
	LinkedBy       VillageUUID `json:"linked_by"`
	LastSyncedAt   *time.Time  `json:"last_synced_at"`
	CreatedAt      *time.Time  `json:"created_at"`
}

type VillageLinkedRepositoriesResponse struct {
	Repositories []VillageLinkedRepository `json:"repositories" nullable:"false"`
}

type VillageRepositoryCommit struct {
	SHA         string     `json:"sha"`
	Message     *string    `json:"message"`
	AuthorName  *string    `json:"author_name"`
	AuthorEmail *string    `json:"author_email"`
	AuthoredAt  *time.Time `json:"authored_at"`
	CommittedAt *time.Time `json:"committed_at"`
}

type VillageRepositoryCommitsResponse struct {
	Owner       string                    `json:"owner"`
	Name        string                    `json:"name"`
	Refreshed   bool                      `json:"refreshed"`
	LastSynced  *time.Time                `json:"last_synced"`
	CommitCount int                       `json:"commit_count"`
	Commits     []VillageRepositoryCommit `json:"commits" nullable:"false"`
}
