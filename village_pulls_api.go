package schema

import (
	"fmt"
	"strings"
	"time"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// VillagePullRequestAttachmentState is the closed lifecycle of one pull
// request's prompt attachment. The design's state machine names each
// transition; the wire only carries the current state.
type VillagePullRequestAttachmentState string

const (
	// VillagePullRequestAttachmentRequested: a non-author asked; nothing is exposed.
	VillagePullRequestAttachmentRequested VillagePullRequestAttachmentState = "requested"
	// VillagePullRequestAttachmentWaiting: the author consented; no matching transcript exists yet.
	VillagePullRequestAttachmentWaiting VillagePullRequestAttachmentState = "waiting"
	// VillagePullRequestAttachmentPreview: the digest is computed and awaits the author's confirm on Village.
	VillagePullRequestAttachmentPreview VillagePullRequestAttachmentState = "preview"
	// VillagePullRequestAttachmentAttached: the comment is posted and the transcripts are shared.
	VillagePullRequestAttachmentAttached VillagePullRequestAttachmentState = "attached"
	// VillagePullRequestAttachmentDetached: the comment is deleted and prior visibility restored.
	VillagePullRequestAttachmentDetached VillagePullRequestAttachmentState = "detached"
)

var AllVillagePullRequestAttachmentStates = []VillagePullRequestAttachmentState{
	VillagePullRequestAttachmentRequested,
	VillagePullRequestAttachmentWaiting,
	VillagePullRequestAttachmentPreview,
	VillagePullRequestAttachmentAttached,
	VillagePullRequestAttachmentDetached,
}

func (s VillagePullRequestAttachmentState) IsValid() bool {
	for _, known := range AllVillagePullRequestAttachmentStates {
		if s == known {
			return true
		}
	}
	return false
}

func (s VillagePullRequestAttachmentState) String() string { return string(s) }

func (VillagePullRequestAttachmentState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Pull Request Attachment State", "Current lifecycle state of a pull request's prompt attachment", AllVillagePullRequestAttachmentStates), nil
}

// VillagePromptsCheckMode decides the check-run conclusion a collective's
// linked repositories receive when no prompts are attached.
type VillagePromptsCheckMode string

const (
	// VillagePromptsCheckInformational: neutral when nothing is attached. The default.
	VillagePromptsCheckInformational VillagePromptsCheckMode = "informational"
	// VillagePromptsCheckRequired: failure when nothing is attached, so branch protection can require prompts.
	VillagePromptsCheckRequired VillagePromptsCheckMode = "required"
)

var AllVillagePromptsCheckModes = []VillagePromptsCheckMode{VillagePromptsCheckInformational, VillagePromptsCheckRequired}

func (m VillagePromptsCheckMode) IsValid() bool {
	for _, known := range AllVillagePromptsCheckModes {
		if m == known {
			return true
		}
	}
	return false
}

func (m VillagePromptsCheckMode) String() string { return string(m) }

func (VillagePromptsCheckMode) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Prompts Check Mode", "Check-run conclusion policy for a collective's linked repositories when no prompts are attached", AllVillagePromptsCheckModes), nil
}

// VillagePullRequestAttachment is one pull request's attachment row. Title and
// HeadRef are the pull request's title and head branch; they are null when
// Village does not know them.
type VillagePullRequestAttachment struct {
	ID                  VillageUUID                       `json:"id"`
	Owner               string                            `json:"owner"`
	Name                string                            `json:"name"`
	Number              int                               `json:"number"`
	Title               *string                           `json:"title"`
	HeadRef             *string                           `json:"head_ref"`
	HeadSHA             string                            `json:"head_sha"`
	IsPrivateRepository bool                              `json:"is_private_repository"`
	State               VillagePullRequestAttachmentState `json:"state"`
	AuthorUserID        *VillageUUID                      `json:"author_user_id"`
	RequestedByGithubID *int64                            `json:"requested_by_github_id"`
	CommentID           *int64                            `json:"comment_id"`
	CheckRunID          *int64                            `json:"check_run_id"`
	CreatedAt           time.Time                         `json:"created_at"`
	UpdatedAt           time.Time                         `json:"updated_at"`
	ConfirmedAt         *time.Time                        `json:"confirmed_at"`
	DetachedAt          *time.Time                        `json:"detached_at"`
}

// VillagePullRequestAttachedTranscript is one transcript in an attachment, in
// chain order, with the visibility to restore on detach.
type VillagePullRequestAttachedTranscript struct {
	TranscriptID       TranscriptID                `json:"transcript_id"`
	Position           int                         `json:"position"`
	PreviousVisibility VillageTranscriptVisibility `json:"previous_visibility"`
	Title              *string                     `json:"title"`
	SessionStart       *time.Time                  `json:"session_start"`
}

// VillagePullRequestAttachmentResponse is the read, confirm, and detach
// response. Digest is null until the attachment reaches preview or attached.
type VillagePullRequestAttachmentResponse struct {
	Attachment     VillagePullRequestAttachment           `json:"attachment"`
	Digest         *PromptDigest                          `json:"digest"`
	Transcripts    []VillagePullRequestAttachedTranscript `json:"transcripts" nullable:"false"`
	ViewerIsAuthor bool                                   `json:"viewer_is_author"`
}

// VillagePromptRequest is one attachment waiting on the caller's machine.
// Remote is the full name of the repository the pull request was opened against,
// and HeadRemote the full name of the repository its head came from: a fork's
// repository, or the base itself for a same-repository pull request. Together
// they let the CLI match the repository it is pushing, normalizing its own git
// remote first, without re-deriving either rule, so a push from a clone of a
// fork matches the request waiting on that fork's author.
type VillagePromptRequest struct {
	Owner       string                            `json:"owner"`
	Name        string                            `json:"name"`
	Number      int                               `json:"number"`
	State       VillagePullRequestAttachmentState `json:"state"`
	Remote      string                            `json:"remote"`
	HeadRemote  string                            `json:"head_remote"`
	RequestedAt time.Time                         `json:"requested_at"`
}

type VillagePromptRequestsResponse struct {
	Requests []VillagePromptRequest `json:"requests" nullable:"false"`
}

// VillageUserSettings is the caller's own settings surface.
// AutoAttachPullRequests is the caller's choice to link their own transcripts
// to a pull request automatically when it opens in a repository one of their
// collectives links. It is off by default and never changes who can read a
// transcript.
type VillageUserSettings struct {
	PreviewBeforeAttach    bool `json:"preview_before_attach"`
	AutoAttachPullRequests bool `json:"auto_attach_pull_requests"`
}

// VillageUpdateUserSettingsRequest patches the caller's settings. An omitted
// field is left unchanged.
type VillageUpdateUserSettingsRequest struct {
	PreviewBeforeAttach    *bool `json:"preview_before_attach,omitempty"`
	AutoAttachPullRequests *bool `json:"auto_attach_pull_requests,omitempty"`
}

// VillageTranscriptPullRequest is one pull request that includes a transcript,
// for the transcript page's list. It names the pull request and carries its
// title and head branch when Village knows them. It is narrower than the
// attachment row on purpose: a reader of a transcript needs to recognize the
// pull request, not the attachment's internal bookkeeping.
type VillageTranscriptPullRequest struct {
	Owner   string                            `json:"owner" minLength:"1"`
	Name    string                            `json:"name" minLength:"1"`
	Number  int                               `json:"number" minimum:"1"`
	Title   *string                           `json:"title"`
	HeadRef *string                           `json:"head_ref"`
	State   VillagePullRequestAttachmentState `json:"state"`
}

// Validate checks the reference and the state. The read lists only attached and
// detached pull requests, so a pending request never reaches a reader.
func (p VillageTranscriptPullRequest) Validate() error {
	ref := VillagePullRequestRef{Owner: p.Owner, Name: p.Name, Number: p.Number}
	if err := ref.Validate(); err != nil {
		return err
	}
	if p.State != VillagePullRequestAttachmentAttached && p.State != VillagePullRequestAttachmentDetached {
		return fmt.Errorf("transcript pull request validation failed for %s/%s#%d at schema.VillageTranscriptPullRequest.Validate: state %q; the read lists only attached and detached pull requests, so a pending request never reaches a reader", p.Owner, p.Name, p.Number, p.State)
	}
	return nil
}

// VillageTranscriptPullRequestsResponse is the response of GET
// /api/v1/transcripts/{id}/pulls: the pull requests that include the transcript,
// attached or detached, that the caller may read. This is the visibility rule
// the pull request counts refer to: an attachment on a private repository is
// listed only to the pull request author, a member of the collective the
// attachment belongs to, or, for an attached attachment, a reader of the
// repository, and one whose visibility check cannot complete is omitted.
type VillageTranscriptPullRequestsResponse struct {
	PullRequests []VillageTranscriptPullRequest `json:"pull_requests" nullable:"false"`
}

// Validate checks each pull request and that each appears once.
func (r VillageTranscriptPullRequestsResponse) Validate() error {
	if r.PullRequests == nil {
		return fmt.Errorf("transcript pull requests validation failed at schema.VillageTranscriptPullRequestsResponse.Validate: pull_requests is null; emit [] when no pull request includes the transcript")
	}
	seen := make(map[VillagePullRequestRef]struct{}, len(r.PullRequests))
	for _, pull := range r.PullRequests {
		if err := pull.Validate(); err != nil {
			return err
		}
		ref := VillagePullRequestRef{Owner: pull.Owner, Name: pull.Name, Number: pull.Number}
		if _, duplicate := seen[ref]; duplicate {
			return fmt.Errorf("transcript pull requests validation failed at schema.VillageTranscriptPullRequestsResponse.Validate: %s/%s#%d is listed twice; list each pull request once", ref.Owner, ref.Name, ref.Number)
		}
		seen[ref] = struct{}{}
	}
	return nil
}

// VillagePullRequestRef names one pull request.
type VillagePullRequestRef struct {
	Owner  string `json:"owner" minLength:"1"`
	Name   string `json:"name" minLength:"1"`
	Number int    `json:"number" minimum:"1"`
}

// Validate checks that the reference names a repository and a pull request.
func (r VillagePullRequestRef) Validate() error {
	if strings.TrimSpace(r.Owner) == "" || strings.TrimSpace(r.Name) == "" || r.Number < 1 {
		return fmt.Errorf("pull request reference validation failed at schema.VillagePullRequestRef.Validate: %q/%q#%d does not name a pull request; emit the repository owner, name, and a positive number", r.Owner, r.Name, r.Number)
	}
	return nil
}

// VillagePullRequestsSummary summarizes the pull requests a transcript is
// attached to, for a list row. Count is the number of pull requests whose
// attachment is in state attached and that the caller may read under the
// visibility rule of VillageTranscriptPullRequestsResponse, failing closed;
// detached pull requests are not counted. Recent holds the most recent of them, at most as
// many as the server shows, so a row reads "#42, #45 +2" without one read per
// row.
type VillagePullRequestsSummary struct {
	Count  int32                   `json:"count" minimum:"0"`
	Recent []VillagePullRequestRef `json:"recent" nullable:"false"`
}

// Validate checks that the recent list fits in the count and does not repeat.
func (s VillagePullRequestsSummary) Validate() error {
	if s.Count < 0 || s.Recent == nil || len(s.Recent) > int(s.Count) {
		return fmt.Errorf("pull request summary validation failed at schema.VillagePullRequestsSummary.Validate: count %d with %d recent pull requests; recent is an array no longer than a nonnegative count", s.Count, len(s.Recent))
	}
	seen := make(map[VillagePullRequestRef]struct{}, len(s.Recent))
	for _, ref := range s.Recent {
		if err := ref.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[ref]; duplicate {
			return fmt.Errorf("pull request summary validation failed at schema.VillagePullRequestsSummary.Validate: %s/%s#%d is listed twice; list each pull request once", ref.Owner, ref.Name, ref.Number)
		}
		seen[ref] = struct{}{}
	}
	return nil
}

// VillageUserStats is the response of GET /api/v1/users/me/stats: totals over
// every transcript the caller published. PullRequestCount counts the distinct
// pull requests whose attachment is in state attached, includes one of those
// transcripts, and that the caller may read under the visibility rule of
// VillageTranscriptPullRequestsResponse, failing closed; detached pull
// requests are not counted.
type VillageUserStats struct {
	TotalTranscripts int32 `json:"total_transcripts" minimum:"0"`
	TotalTurns       int64 `json:"total_turns" minimum:"0"`
	TotalDurationMs  int64 `json:"total_duration_ms" minimum:"0"`
	TotalTokens      int64 `json:"total_tokens" minimum:"0"`
	PullRequestCount int32 `json:"pull_request_count" minimum:"0"`
}

// Validate checks that every total is nonnegative.
func (s VillageUserStats) Validate() error {
	if s.TotalTranscripts < 0 || s.TotalTurns < 0 || s.TotalDurationMs < 0 || s.TotalTokens < 0 || s.PullRequestCount < 0 {
		return fmt.Errorf("personal stats validation failed at schema.VillageUserStats.Validate: a total is negative; emit nonnegative totals")
	}
	return nil
}

// VillageGitHubWebhookPayload is GitHub's event payload, forwarded verbatim.
// Village verifies the HMAC over the raw body and reads only the fields the
// event needs, so the contract leaves the object open.
type VillageGitHubWebhookPayload map[string]any

func (VillageGitHubWebhookPayload) JSONSchema() (jsonschema.Schema, error) {
	open := true
	s := jsonschema.Schema{}
	s.AddType(jsonschema.Object)
	s.WithTitle("GitHub Webhook Payload")
	s.WithDescription("GitHub event payload forwarded verbatim; validated by the X-Hub-Signature-256 HMAC over the raw body, never by shape")
	s.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: &open})
	return s, nil
}
