package schema

import "time"

// VillageDiscoveryQuery is the query surface of GET /api/v1/transcripts.
// Unknown sort values retain the server's legacy default ordering.
type VillageDiscoveryQuery struct {
	Page     int           `query:"page" minimum:"1" description:"One-based result page. Values below one use the server default."`
	Limit    int           `query:"limit" minimum:"1" maximum:"100" description:"Maximum rows per page. Values outside 1 through 100 use the server default."`
	Query    string        `query:"q" description:"Case-insensitive title or description search."`
	Provider string        `query:"provider" description:"Harness identifier filter retained under the historical provider query name."`
	Owner    string        `query:"owner" description:"Owner username filter."`
	Project  string        `query:"project" description:"Case-insensitive project-name filter."`
	Repo     string        `query:"repo" description:"Case-insensitive repository-remote filter."`
	Org      string        `query:"org" description:"Visible GitHub organization filter."`
	Tags     string        `query:"tags" description:"Comma-separated tag names."`
	Origin   SessionOrigin `query:"origin" description:"Optional exact session-origin scope. Omission uses the default non-agent discovery scope."`
	Sort     string        `query:"sort" description:"Ordering mode: recent, turns, tokens, or duration. Unknown values retain legacy recent ordering."`
}

// VillageHarnessFacet reports the number of distinct transcripts for one
// harness in the viewer-visible default discovery corpus.
type VillageHarnessFacet struct {
	Harness Harness `json:"harness" required:"true"`
	Count   int64   `json:"count" required:"true" minimum:"1"`
}

type VillageDiscoveryUser struct {
	ID               VillageUUID `json:"id"`
	GitHubID         int64       `json:"github_id"`
	GitHubUsername   string      `json:"github_username"`
	DisplayName      *string     `json:"display_name"`
	AvatarURL        *string     `json:"avatar_url"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	IsDiscoverable   bool        `json:"is_discoverable"`
	UsernameChosen   bool        `json:"username_chosen"`
	ProviderUsername *string     `json:"provider_username"`
	Provider         string      `json:"provider"`
	ProviderUserID   string      `json:"provider_user_id"`
}

type VillageDiscoveryTag struct {
	ID   VillageUUID `json:"id"`
	Name string      `json:"name"`
}

type VillageDiscoveryOwnerOrg struct {
	UserID    VillageUUID `json:"user_id"`
	OrgLogin  string      `json:"org_login"`
	AvatarURL *string     `json:"avatar_url"`
}

type VillageDiscoveryShare struct {
	TranscriptID   TranscriptID               `json:"transcript_id"`
	GroupID        VillageUUID                `json:"group_id"`
	GroupName      string                     `json:"group_name"`
	AcceptanceMode VillageGroupAcceptanceMode `json:"acceptance_mode"`
	Status         VillageShareStatus         `json:"status"`
	SharedAt       time.Time                  `json:"shared_at"`
}

type VillageDiscoveryAttestation struct {
	TranscriptID    TranscriptID `json:"transcript_id"`
	OrgLogin        string       `json:"org_login"`
	AttestationType string       `json:"attestation_type"`
	CreatedAt       time.Time    `json:"created_at"`
}

// VillageDiscoveryTranscriptProjection reuses the shared Village transcript
// projection while overriding the two database JSON byte columns with their
// actual browser serialization. The current handler stores them as []byte, so
// encoding/json emits null for nil and a base64 string for populated values.
// This browser-specific correction must not change other existing contracts
// that already use VillageTranscript.
type VillageDiscoveryTranscriptProjection struct {
	VillageTranscript   `json:",inline"`
	Subagents           []byte `json:"subagents"`
	DiagnosticsWarnings []byte `json:"diagnostics_warnings"`
}

// VillageDiscoveryTranscript wraps one browser result exactly as the Village
// list handler serves it. The wrapper is intentionally distinct from pull.
type VillageDiscoveryTranscript struct {
	Transcript   VillageDiscoveryTranscriptProjection `json:"transcript"`
	Tags         []VillageDiscoveryTag                `json:"tags" nullable:"false"`
	Owner        VillageDiscoveryUser                 `json:"owner"`
	OwnerOrgs    []VillageDiscoveryOwnerOrg           `json:"owner_orgs"`
	Shares       []VillageDiscoveryShare              `json:"shares"`
	Attestations []VillageDiscoveryAttestation        `json:"attestations"`
}

// VillageDiscoveryResponse is the browser discovery envelope returned by GET
// /api/v1/transcripts. HarnessFacets is independent of every active request
// filter, sorting, and pagination. It covers all distinct transcripts readable
// by the current viewer under the default non-agent origin scope. Authentication
// is optional: an anonymous viewer's corpus contains public transcripts only;
// an authenticated viewer additionally receives owned and readable shared
// transcripts. User and unknown origins are included and agent origin is excluded. Entries have
// positive counts and deterministic harness order; an empty corpus emits [].
type VillageDiscoveryResponse struct {
	Transcripts   []VillageDiscoveryTranscript `json:"transcripts" nullable:"false"`
	HarnessFacets []VillageHarnessFacet        `json:"harness_facets" nullable:"false"`
	Total         int64                        `json:"total"`
	AgentTotal    int64                        `json:"agent_total"`
	Page          int                          `json:"page"`
	Limit         int                          `json:"limit"`
}
