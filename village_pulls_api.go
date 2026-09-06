package schema

import (
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

// VillagePullRequestAttachment is one pull request's attachment row.
type VillagePullRequestAttachment struct {
	ID                  VillageUUID                       `json:"id"`
	Owner               string                            `json:"owner"`
	Name                string                            `json:"name"`
	Number              int                               `json:"number"`
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
// Remote is the normalized repository remote so the CLI can match the
// repository it is pushing without re-deriving the rule.
type VillagePromptRequest struct {
	Owner       string                            `json:"owner"`
	Name        string                            `json:"name"`
	Number      int                               `json:"number"`
	State       VillagePullRequestAttachmentState `json:"state"`
	Remote      string                            `json:"remote"`
	RequestedAt time.Time                         `json:"requested_at"`
}

type VillagePromptRequestsResponse struct {
	Requests []VillagePromptRequest `json:"requests" nullable:"false"`
}

// VillageUserSettings is the caller's own settings surface.
type VillageUserSettings struct {
	PreviewBeforeAttach bool `json:"preview_before_attach"`
}

// VillageUpdateUserSettingsRequest patches the caller's settings. An omitted
// field is left unchanged.
type VillageUpdateUserSettingsRequest struct {
	PreviewBeforeAttach *bool `json:"preview_before_attach,omitempty"`
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
