package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// addKickstartOperations declares the local setup routes. Every method other
// than GET declares the shared cross-origin refusal through addLocalOperations.
// The tree and the preview are the whole answer for one session; neither route
// has a page, a cursor, or a limit.
func addKickstartOperations(r *openapi31.Reflector) error {
	sessionPath := new(struct {
		ID string `path:"id" description:"Kickstart session id"`
	})
	previewPath := new(struct {
		ID     string `path:"id" description:"Kickstart session id"`
		NodeID string `path:"nodeId" description:"Forest row id"`
	})
	return addLocalOperations(r, []localOperationSpec{
		{
			method: http.MethodGet, path: "/api/v1/kickstart/status", id: "getKickstartStatus", tag: "kickstart",
			description: "Read the first-run redirect hook. configured is the only status flag. activeSessionId, when present, is the one full-scope session id; it is omitted when that session is not open. The response is not a session directory and has no per-scope map.",
			response:    new(schema.LocalKickstartStatus),
		},
		{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions", id: "openKickstartSession", tag: "kickstart",
			description:   "Open or join the live session for scope. full is the wizard and selection is the selection page. The same scope returns the same id and revision; the other scope returns a different session and leaves the first untouched. The response is the reconnect snapshot, so the page needs no second read. A scope other than full or selection is 400 and changes nothing.",
			requests:      []interface{}{new(schema.LocalKickstartOpenRequest)},
			response:      new(schema.LocalKickstartSession),
			requiredBody:  true,
			errorStatuses: []int{http.StatusBadRequest},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}", id: "getKickstartSession", tag: "kickstart",
			description:   "Read one session snapshot for reconnect. The forest is not in this body; load it from the tree route. 404 when no session has this id.",
			requests:      []interface{}{sessionPath},
			response:      new(schema.LocalKickstartSession),
			errorStatuses: []int{http.StatusNotFound},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodPatch, path: "/api/v1/kickstart/sessions/{id}", id: "patchKickstartSession", tag: "kickstart",
			description:   "Apply one stop's answers, the tree's selected ids, conflict resolutions, or the connect offer. revision is the session revision the caller last rendered. A stale write is applied and reported in the body; it is not a separate status. A stale result may carry no overwritten entries when no answered field changed after that revision. 400 when the body is not valid, and 404 when no session has this id.",
			requests:      []interface{}{sessionPath, new(schema.LocalKickstartAnswers)},
			response:      new(schema.LocalKickstartPatchResult),
			requiredBody:  true,
			errorStatuses: []int{http.StatusBadRequest, http.StatusNotFound},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}/tree", id: "getKickstartTree", tag: "kickstart",
			description:   "Read the whole selection forest for one session. Each node carries the server-computed state. locked is a boolean, not a fourth state. The response has no page, cursor, or limit. 404 when no session has this id.",
			requests:      []interface{}{sessionPath},
			response:      new(schema.LocalKickstartTree),
			errorStatuses: []int{http.StatusNotFound},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}/tree/{nodeId}/preview", id: "getKickstartPreview", tag: "kickstart",
			description:   "Read the bounded typed preview lines for one forest row. 404 when the session or the row does not exist.",
			requests:      []interface{}{previewPath},
			response:      new(schema.LocalKickstartPreview),
			errorStatuses: []int{http.StatusNotFound},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions/{id}/scan", id: "scanKickstartSession", tag: "kickstart",
			description:   "Rescan the filesystem for one session. Progress is pushed on the kickstart topic; this response is the scan state at accept. The forest is not in the response. 404 when no session has this id.",
			requests:      []interface{}{sessionPath},
			response:      new(schema.LocalKickstartScan),
			errorStatuses: []int{http.StatusNotFound},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions/{id}/commit", id: "commitKickstartSession", tag: "kickstart",
			description:   "Recompute, merge, swap the configuration, write retention, and start the import. 409 returns the mine/theirs rows and writes nothing. The success body carries the finish actions and the ingest state; it does not install a hook and has no reset control. 404 when no session has this id.",
			requests:      []interface{}{sessionPath},
			response:      new(schema.LocalKickstartCommitted),
			errorStatuses: []int{http.StatusNotFound},
			errorResponse: new(LocalErrorResponse),
			alternateResponses: map[int]interface{}{
				http.StatusConflict: new(schema.LocalKickstartCommitRefusal),
			},
		},
	})
}
