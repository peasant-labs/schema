package openapi

import (
	"net/http"

	"github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// addVillagePullRequestOperations declares the pull request prompt attachment
// surface: the GitHub webhook receiver, the attachment read, confirm, and
// detach routes, the caller's waiting requests, and the caller's settings.
func addVillagePullRequestOperations(r *openapi31.Reflector) error {
	pullPath := new(struct {
		Owner  string `path:"owner" description:"Repository owner login"`
		Name   string `path:"name" description:"Repository name"`
		Number int    `path:"number" minimum:"1" description:"Pull request number"`
	})
	webhookHeaders := new(struct {
		Event     string `header:"X-GitHub-Event" required:"true" description:"GitHub event name"`
		Delivery  string `header:"X-GitHub-Delivery" required:"true" description:"Unique delivery identifier; Village records it so a redelivery is idempotent"`
		Signature string `header:"X-Hub-Signature-256" required:"true" description:"HMAC SHA-256 of the raw body under the App webhook secret"`
	})

	operations := []villageOperationSpec{
		{
			method:        http.MethodPost,
			path:          "/api/v1/integrations/github/webhook",
			id:            "receiveGitHubWebhook",
			tag:           "integrations",
			description:   "Receive a GitHub App webhook delivery. The body is GitHub's payload, authenticated by the X-Hub-Signature-256 HMAC over the raw body; X-GitHub-Delivery makes redelivery idempotent. Handles installation, pull_request, check_run, and issue_comment events and acknowledges every other event without acting. Returns 501 when the App or its webhook secret is not configured.",
			requests:      []interface{}{webhookHeaders, new(schema.VillageGitHubWebhookPayload)},
			response:      new(schema.VillageStatusResponse),
			successStatus: http.StatusAccepted,
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusNotImplemented,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/pulls/{owner}/{name}/{number}",
			id:          "getPullRequestAttachment",
			tag:         "pulls",
			description: "Read one pull request's prompt attachment. The digest is null until the attachment reaches preview or attached; viewer_is_author is true when the caller is the pull request author. Readers who may not see the attached transcripts receive 403.",
			requests:    []interface{}{pullPath},
			response:    new(schema.VillagePullRequestAttachmentResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/pulls/{owner}/{name}/{number}/confirm",
			id:          "confirmPullRequestAttachment",
			tag:         "pulls",
			description: "Confirm a preview and post it. Only the pull request author may confirm; the attachment must be in preview, otherwise 409. Posting to GitHub failing returns 502 and leaves the attachment in preview.",
			requests:    []interface{}{pullPath},
			response:    new(schema.VillagePullRequestAttachmentResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusConflict,
				http.StatusInternalServerError,
				http.StatusBadGateway,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/pulls/{owner}/{name}/{number}",
			id:          "detachPullRequestAttachment",
			tag:         "pulls",
			description: "Detach the prompts from a pull request. Only the pull request author may detach. Deletes the comment, resets the check, and restores each transcript's previous visibility; an attachment already detached returns 409.",
			requests:    []interface{}{pullPath},
			response:    new(schema.VillagePullRequestAttachmentResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusConflict,
				http.StatusInternalServerError,
				http.StatusBadGateway,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/users/me/prompt-requests",
			id:          "listMyPromptRequests",
			tag:         "pulls",
			description: "List the caller's attachments that are waiting for a transcript from the caller's machine, with each repository's normalized remote so a client can match the repository it is about to push.",
			response:    new(schema.VillagePromptRequestsResponse),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/users/me/settings",
			id:          "getMySettings",
			tag:         "users",
			description: "Read the caller's settings, including preview_before_attach.",
			response:    new(schema.VillageUserSettings),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPatch,
			path:        "/api/v1/users/me/settings",
			id:          "updateMySettings",
			tag:         "users",
			description: "Patch the caller's settings. A field that is omitted is left unchanged.",
			requests:    []interface{}{new(schema.VillageUpdateUserSettingsRequest)},
			response:    new(schema.VillageUserSettings),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
	}

	if err := addVillageOperations(r, operations); err != nil {
		return err
	}
	if err := requireRequestBody(r.Spec, http.MethodPost, "/api/v1/integrations/github/webhook"); err != nil {
		return err
	}
	return requireRequestBody(r.Spec, http.MethodPatch, "/api/v1/users/me/settings")
}
