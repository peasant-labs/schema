package schema

import (
	jsonschema "github.com/swaggest/jsonschema-go"
	"time"
)

// VillageAttestationType is the closed menu of usage attestations.
type VillageAttestationType string

const (
	VillageAttestationUsedInTraining VillageAttestationType = "used_in_training"
	VillageAttestationReferenced     VillageAttestationType = "referenced"
	VillageAttestationEvaluated      VillageAttestationType = "evaluated"
	VillageAttestationDeployed       VillageAttestationType = "deployed"
)

var AllVillageAttestationTypes = []VillageAttestationType{VillageAttestationUsedInTraining, VillageAttestationReferenced, VillageAttestationEvaluated, VillageAttestationDeployed}

func (t VillageAttestationType) String() string { return string(t) }
func (t VillageAttestationType) IsValid() bool {
	for _, v := range AllVillageAttestationTypes {
		if t == v {
			return true
		}
	}
	return false
}
func (VillageAttestationType) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Attestation Type", "Usage attestation type", AllVillageAttestationTypes), nil
}

// --- account
type VillageDiscoverabilityRequest struct {
	IsDiscoverable *bool `json:"is_discoverable" required:"true" nullable:"false"`
}
type VillageUsernameRequest struct {
	Username string `json:"username" description:"Trimmed and lowercased, then matched against ^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$"`
}
type VillageCreateAPIKeyRequest struct {
	Label string `json:"label,omitempty"`
}

// Members in alphabetical order: the handler emitted a map.
type VillageCreatedAPIKey struct {
	CreatedAt time.Time   `json:"created_at" description:"When the key was created. The database sets it as the key is stored, so it is always present here."`
	ID        VillageUUID `json:"id"`
	Key       string      `json:"key" description:"The plaintext key. Returned only here, once."`
	Label     string      `json:"label" description:"The label as sent; an empty label is stored as null and listed as null."`
	Prefix    string      `json:"prefix"`
}
type VillageAPIKey struct { // ListUserAPIKeysRow order
	ID        VillageUUID `json:"id"`
	KeyPrefix string      `json:"key_prefix"`
	Label     *string     `json:"label"`
	CreatedAt *time.Time  `json:"created_at"`
	LastUsed  *time.Time  `json:"last_used"`
	RevokedAt *time.Time  `json:"revoked_at"`
}
type VillageOrganizationVisibilityRequest struct {
	Visible bool `json:"visible,omitempty" description:"Omitted means false."`
}
type VillageOrganizationVisibilityResponse struct { // alphabetical (map)
	Status  string `json:"status"`
	Visible bool   `json:"visible"`
}

// --- organizations, tags
type VillageOrganizationSearchResult struct { // SearchOrgsRow order
	OrgLogin        string  `json:"org_login"`
	AvatarURL       *string `json:"avatar_url"`
	MemberCount     int32   `json:"member_count"`
	TranscriptCount int32   `json:"transcript_count"`
}
type VillageOrganizationSearchResponse struct {
	Organizations []VillageOrganizationSearchResult `json:"organizations" nullable:"false"`
}
type VillageTagUsage struct { // ListTagsRow == ListPopularTagsRow order
	ID         VillageUUID `json:"id"`
	Name       string      `json:"name"`
	UsageCount int64       `json:"usage_count"`
}

// --- projects
type VillageResolvedProject struct { // Village resolvedProject order
	ProjectHash        ProjectHash              `json:"project_hash"`
	ProjectDisplayName string                   `json:"project_display_name"`
	ProjectNameSource  VillageProjectNameSource `json:"project_name_source"`
	ProjectRemoteLabel string                   `json:"project_remote_label"`
}
type VillageProjectCollective struct { // ProjectCollectiveRollupEntry order
	ID              VillageUUID `json:"id"`
	Name            string      `json:"name"`
	Description     *string     `json:"description"`
	LinkedGithubOrg *string     `json:"linked_github_org"`
	TranscriptCount int32       `json:"transcript_count" description:"Approved transcripts of this project in the collective."`
}
type VillageUserProjectResponse struct { // alphabetical (map)
	Collectives []VillageProjectCollective `json:"collectives" nullable:"false"`
	Owner       VillagePublicOwner         `json:"owner"`
	Project     VillageResolvedProject     `json:"project"`
	Transcripts []VillageTranscript        `json:"transcripts" nullable:"false"`
}
type VillageProjectDisplayNameRequest struct {
	DisplayName string `json:"display_name" description:"Trimmed; must be non-empty and at most 255 bytes."`
}

// --- transcripts
type VillageCreateAttestationRequest struct {
	OrgLogin        string                 `json:"org_login" minLength:"1"`
	AttestationType VillageAttestationType `json:"attestation_type"`
	Note            string                 `json:"note,omitempty" description:"Empty or omitted stores no note (null)."`
}
type VillageManualAnnotationRequest struct {
	TypeID     string  `json:"typeId" minLength:"1"`
	Value      string  `json:"value" minLength:"1"`
	EntryIndex *int    `json:"entryIndex" required:"true" nullable:"false" minimum:"0"`
	EndIndex   *int    `json:"endIndex,omitempty" nullable:"false" description:"Exclusive end; defaults to entryIndex+1; must be greater than entryIndex."`
	IsPrimary  bool    `json:"isPrimary,omitempty"`
	Reason     *string `json:"reason,omitempty"`
}
type VillageTranscriptAnnotationsResponse struct {
	Annotations []AnnotationSummary `json:"annotations" nullable:"false"`
}
type VillageTranscriptCommit struct { // transcriptCommitResponse order
	Sha         string  `json:"sha"`
	Message     *string `json:"message"`
	AuthorName  *string `json:"authorName" description:"Null when the transcript's owner is anonymous to the caller, and when the commit records no author name. A caller who may see the owner receives the recorded value."`
	AuthorEmail *string `json:"authorEmail" description:"Null when the transcript's owner is anonymous to the caller, and when the commit records no author email. A caller who may see the owner receives the recorded value."`
	AuthoredAt  *int64  `json:"authoredAt" format:"int64" description:"Unix milliseconds."`
	CommittedAt *int64  `json:"committedAt" format:"int64" description:"Unix milliseconds."`
	Order       int32   `json:"order"`
}
type VillageTranscriptCommitsResponse struct {
	Commits []VillageTranscriptCommit `json:"commits" nullable:"false"`
}

// --- health
type VillageHealthResponse struct { // healthProbe order
	Status        string   `json:"status" description:"Always ok when the process answers."`
	BuildRevision string   `json:"buildRevision,omitempty" pattern:"^[0-9a-f]{7,12}$"`
	Migrations    *int64   `json:"migrations,omitempty" nullable:"false" description:"Highest applied migration version; absent when the registry could not be read within one second."`
	Serves        []string `json:"serves" nullable:"false" description:"Forward-open capability tokens this deployment advertises. Village owns the token set and may add tokens without a contract change. Match exactly; ignore unknown tokens."`
}
