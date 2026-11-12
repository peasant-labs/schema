package openapi

import (
	"net/http"
	"sort"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// kickstartStatusSummaries is the first sentence of each kickstart error
// status's description. The codes follow it.
var kickstartStatusSummaries = map[int]string{
	http.StatusBadRequest:          "The request was refused and nothing was changed.",
	http.StatusNotFound:            "No session or forest row matches.",
	http.StatusInternalServerError: "The server could not finish the request.",
	http.StatusServiceUnavailable:  "The setup service cannot take the request now.",
}

// withKickstartErrors declares a route's error statuses from the codes it can
// answer. Each code carries its own status, schema.LocalKickstartErrorCode
// .HTTPStatus, so a route cannot place a code under a status the code does not
// answer with. Every status answers with the closed
// LocalKickstartErrorResponse. The codes of one status keep the order given.
func withKickstartErrors(op localOperationSpec, codes ...schema.LocalKickstartErrorCode) localOperationSpec {
	byStatus := map[int][]string{}
	for _, code := range codes {
		status := code.HTTPStatus()
		byStatus[status] = append(byStatus[status], string(code))
	}
	statuses := make([]int, 0, len(byStatus))
	for status := range byStatus {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	op.errorStatuses = statuses
	op.errorCodes = byStatus
	op.errorSummaries = kickstartStatusSummaries
	op.errorResponse = new(schema.LocalKickstartErrorResponse)
	return op
}

// addKickstartOperations declares the local setup routes. Every method other
// than GET declares the shared cross-origin refusal through addLocalOperations,
// and every GET declares the host refusal, which is the only refusal a read
// can meet. The tree and the preview are the whole answer for one session;
// neither route has a page, a cursor, or a limit. Every route can answer 500
// and 503, so each declares both with the closed error body.
func addKickstartOperations(r *openapi31.Reflector) error {
	sessionPath := new(struct {
		ID string `path:"id" description:"Kickstart session id"`
	})
	previewPath := new(struct {
		ID     string `path:"id" description:"Kickstart session id"`
		NodeID string `path:"nodeId" description:"Forest row id"`
	})
	return addLocalOperations(r, []localOperationSpec{
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/status", id: "getKickstartStatus", tag: "kickstart",
			description: "Read the first-run redirect hook. configured is the only status flag. activeSessionId, when present, is the one full-scope session id; it is omitted when that session is not open. The response is not a session directory and has no per-scope map. 500 when the configuration file cannot be inspected, and 503 when the setup service is unavailable.",
			response:    new(schema.LocalKickstartStatus),
			hostRefusal: true,
		},
			schema.LocalKickstartErrorStatusUnreadable, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions", id: "openKickstartSession", tag: "kickstart",
			description:  "Open or join the live session for scope. full is the wizard and selection is the selection page. The same scope returns the same id and revision; the other scope returns a different session and leaves the first untouched. The response is the reconnect snapshot, so the page needs no second read. A scope other than full or selection, and a body that is missing or not the declared shape, is 400 and changes nothing. 500 when the configuration file cannot be read or the first draft cannot be built, and 503 when the setup service is unavailable.",
			requests:     []interface{}{new(schema.LocalKickstartOpenRequest)},
			response:     new(schema.LocalKickstartSession),
			requiredBody: true,
		},
			schema.LocalKickstartErrorBodyMissing, schema.LocalKickstartErrorBodyInvalid, schema.LocalKickstartErrorScopeInvalid,
			schema.LocalKickstartErrorConfigUnreadable, schema.LocalKickstartErrorSessionUnbuildable,
			schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}", id: "getKickstartSession", tag: "kickstart",
			description: "Read one session snapshot for reconnect. The forest is not in this body; load it from the tree route. 404 when no session has this id, 500 when the snapshot cannot be built, and 503 when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartSession),
			hostRefusal: true,
		},
			schema.LocalKickstartErrorSessionNotFound,
			schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPatch, path: "/api/v1/kickstart/sessions/{id}", id: "patchKickstartSession", tag: "kickstart",
			description:  "Apply one stop's answers, the tree's selected ids, conflict resolutions, or the connect offer. revision is the session revision the caller last rendered. A stale write is applied and reported in the body; it is not a separate status. A stale result may carry no overwritten entries when no answered field changed after that revision. 400 when the body is not valid, when the session is being saved, or when an answer is refused: an unknown field, a field outside this scope, a field hidden in the current view, or a value the field does not accept. Nothing is changed on a 400. 404 when no session has this id, 500 when the answer cannot be applied because of a server defect, and 503 when the setup service is unavailable.",
			requests:     []interface{}{sessionPath, new(schema.LocalKickstartAnswers)},
			response:     new(schema.LocalKickstartPatchResult),
			requiredBody: true,
		},
			schema.LocalKickstartErrorBodyMissing, schema.LocalKickstartErrorBodyInvalid, schema.LocalKickstartErrorAnswersInvalid,
			schema.LocalKickstartErrorSessionCommitting, schema.LocalKickstartErrorRevisionAhead, schema.LocalKickstartErrorAnswerRefused,
			schema.LocalKickstartErrorScanNotDone, schema.LocalKickstartErrorFieldUnknown, schema.LocalKickstartErrorFieldOutOfScope,
			schema.LocalKickstartErrorFieldHidden, schema.LocalKickstartErrorTreeAnswerShape, schema.LocalKickstartErrorConnectOutOfScope,
			schema.LocalKickstartErrorConnectNotConnected, schema.LocalKickstartErrorConnectUnknown,
			schema.LocalKickstartErrorSessionNotFound,
			schema.LocalKickstartErrorAnswerUnencodable, schema.LocalKickstartErrorSessionUnbuildable,
			schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}/tree", id: "getKickstartTree", tag: "kickstart",
			description: "Read the whole selection forest for one session. Each node carries the server-computed state. locked is a boolean, not a fourth state. The response has no page, cursor, or limit. 404 when no session has this id, 500 when the forest cannot be built, and 503 when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartTree),
			hostRefusal: true,
		},
			schema.LocalKickstartErrorSessionNotFound,
			schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}/tree/{nodeId}/preview", id: "getKickstartPreview", tag: "kickstart",
			description: "Read the bounded typed preview lines for one forest row. 404 when the session or the row does not exist, 500 when the session's transcript cannot be read, and 503 when the setup service is unavailable.",
			requests:    []interface{}{previewPath},
			response:    new(schema.LocalKickstartPreview),
			hostRefusal: true,
		},
			schema.LocalKickstartErrorSessionNotFound, schema.LocalKickstartErrorPreviewRowNotFound,
			schema.LocalKickstartErrorPreviewUnreadable, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions/{id}/scan", id: "scanKickstartSession", tag: "kickstart",
			description: "Rescan the filesystem for one session. Progress is pushed on the kickstart topic; this response is the scan state at accept. The forest is not in the response. 400 when the session is being saved, 404 when no session has this id, 500 when the scan state cannot be built, and 503 when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartScan),
		},
			schema.LocalKickstartErrorSessionCommitting,
			schema.LocalKickstartErrorSessionNotFound,
			schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable,
		),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions/{id}/commit", id: "commitKickstartSession", tag: "kickstart",
			description: "Recompute, merge, swap the configuration, write retention, and start the import. 409 returns the mine/theirs rows and writes nothing. The success body carries the finish actions and the ingest state; it does not install a hook and has no reset control. 400 when the save is refused, for example because an answer is not acceptable or the session is already being saved; nothing is written. 404 when no session has this id. 500 when the configuration file cannot be read or written, and the previous file is unchanged. 503 when another change holds the configuration file, when it kept changing, or when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartCommitted),
			alternateResponses: map[int]interface{}{
				http.StatusConflict: new(schema.LocalKickstartCommitRefusal),
			},
		},
			schema.LocalKickstartErrorSessionCommitting, schema.LocalKickstartErrorCommitRefused,
			schema.LocalKickstartErrorSessionNotFound,
			schema.LocalKickstartErrorConfigUnreadable, schema.LocalKickstartErrorConfigUnwritable,
			schema.LocalKickstartErrorSessionUnbuildable, schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			schema.LocalKickstartErrorUnavailable, schema.LocalKickstartErrorConfigBusy, schema.LocalKickstartErrorConfigChanging,
		),
	})
}
