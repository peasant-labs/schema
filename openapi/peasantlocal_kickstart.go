package openapi

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// kickstartErrorCodes names, for one route, the closed error codes each
// non-2xx status can carry. The statuses are the keys. The cross-origin 403
// that every non-GET route declares and the commit 409 body are not here: they
// keep their own bodies.
type kickstartErrorCodes = map[int][]schema.LocalKickstartErrorCode

// withKickstartErrors declares a route's error statuses from its code table.
// Every status answers with the closed LocalKickstartErrorResponse, so the
// statuses and the codes cannot be declared apart.
func withKickstartErrors(op localOperationSpec, codes kickstartErrorCodes) localOperationSpec {
	statuses := make([]int, 0, len(codes))
	for status := range codes {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	op.errorStatuses = statuses
	op.errorCodes = codes
	op.errorResponse = new(schema.LocalKickstartErrorResponse)
	return op
}

// kickstartUnavailable is the 503 every handler writes before it reads
// anything, when the server was started without the setup service.
func kickstartUnavailable() []schema.LocalKickstartErrorCode {
	return []schema.LocalKickstartErrorCode{schema.LocalKickstartErrorUnavailable}
}

// addKickstartOperations declares the local setup routes. Every method other
// than GET declares the shared cross-origin refusal through addLocalOperations.
// The tree and the preview are the whole answer for one session; neither route
// has a page, a cursor, or a limit. Every route can answer 500 and 503, so each
// declares both with the closed error body.
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
		}, kickstartErrorCodes{
			http.StatusInternalServerError: {schema.LocalKickstartErrorStatusUnreadable, schema.LocalKickstartErrorContractBreach},
			http.StatusServiceUnavailable:  kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions", id: "openKickstartSession", tag: "kickstart",
			description:  "Open or join the live session for scope. full is the wizard and selection is the selection page. The same scope returns the same id and revision; the other scope returns a different session and leaves the first untouched. The response is the reconnect snapshot, so the page needs no second read. A scope other than full or selection, and a body that is missing or not the declared shape, is 400 and changes nothing. 500 when the configuration file cannot be read or the first draft cannot be built, and 503 when the setup service is unavailable.",
			requests:     []interface{}{new(schema.LocalKickstartOpenRequest)},
			response:     new(schema.LocalKickstartSession),
			requiredBody: true,
		}, kickstartErrorCodes{
			http.StatusBadRequest: {schema.LocalKickstartErrorBodyMissing, schema.LocalKickstartErrorBodyInvalid, schema.LocalKickstartErrorScopeInvalid},
			http.StatusInternalServerError: {
				schema.LocalKickstartErrorConfigUnreadable, schema.LocalKickstartErrorSessionUnbuildable,
				schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			},
			http.StatusServiceUnavailable: kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}", id: "getKickstartSession", tag: "kickstart",
			description: "Read one session snapshot for reconnect. The forest is not in this body; load it from the tree route. 404 when no session has this id, 500 when the snapshot cannot be built, and 503 when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartSession),
		}, kickstartErrorCodes{
			http.StatusNotFound:            {schema.LocalKickstartErrorSessionNotFound},
			http.StatusInternalServerError: {schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach},
			http.StatusServiceUnavailable:  kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPatch, path: "/api/v1/kickstart/sessions/{id}", id: "patchKickstartSession", tag: "kickstart",
			description:  "Apply one stop's answers, the tree's selected ids, conflict resolutions, or the connect offer. revision is the session revision the caller last rendered. A stale write is applied and reported in the body; it is not a separate status. A stale result may carry no overwritten entries when no answered field changed after that revision. 400 when the body is not valid, when the session is being saved, or when an answer is refused: an unknown field, a field outside this scope, a field hidden in the current view, or a value the field does not accept. Nothing is changed on a 400. 404 when no session has this id, 500 when the answer cannot be applied because of a server defect, and 503 when the setup service is unavailable.",
			requests:     []interface{}{sessionPath, new(schema.LocalKickstartAnswers)},
			response:     new(schema.LocalKickstartPatchResult),
			requiredBody: true,
		}, kickstartErrorCodes{
			http.StatusBadRequest: {
				schema.LocalKickstartErrorBodyMissing, schema.LocalKickstartErrorBodyInvalid, schema.LocalKickstartErrorAnswersInvalid,
				schema.LocalKickstartErrorSessionCommitting, schema.LocalKickstartErrorRevisionAhead, schema.LocalKickstartErrorAnswerRefused,
				schema.LocalKickstartErrorScanNotDone, schema.LocalKickstartErrorFieldUnknown, schema.LocalKickstartErrorFieldOutOfScope,
				schema.LocalKickstartErrorFieldHidden, schema.LocalKickstartErrorTreeAnswerShape, schema.LocalKickstartErrorConnectOutOfScope,
				schema.LocalKickstartErrorConnectNotConnected, schema.LocalKickstartErrorConnectUnknown,
			},
			http.StatusNotFound: {schema.LocalKickstartErrorSessionNotFound},
			http.StatusInternalServerError: {
				schema.LocalKickstartErrorAnswerUnencodable, schema.LocalKickstartErrorSessionUnbuildable,
				schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			},
			http.StatusServiceUnavailable: kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}/tree", id: "getKickstartTree", tag: "kickstart",
			description: "Read the whole selection forest for one session. Each node carries the server-computed state. locked is a boolean, not a fourth state. The response has no page, cursor, or limit. 404 when no session has this id, 500 when the forest cannot be built, and 503 when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartTree),
		}, kickstartErrorCodes{
			http.StatusNotFound:            {schema.LocalKickstartErrorSessionNotFound},
			http.StatusInternalServerError: {schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach},
			http.StatusServiceUnavailable:  kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodGet, path: "/api/v1/kickstart/sessions/{id}/tree/{nodeId}/preview", id: "getKickstartPreview", tag: "kickstart",
			description: "Read the bounded typed preview lines for one forest row. 404 when the session or the row does not exist, 500 when the session's transcript cannot be read, and 503 when the setup service is unavailable.",
			requests:    []interface{}{previewPath},
			response:    new(schema.LocalKickstartPreview),
		}, kickstartErrorCodes{
			http.StatusNotFound:            {schema.LocalKickstartErrorSessionNotFound, schema.LocalKickstartErrorPreviewRowNotFound},
			http.StatusInternalServerError: {schema.LocalKickstartErrorPreviewUnreadable, schema.LocalKickstartErrorContractBreach},
			http.StatusServiceUnavailable:  kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions/{id}/scan", id: "scanKickstartSession", tag: "kickstart",
			description: "Rescan the filesystem for one session. Progress is pushed on the kickstart topic; this response is the scan state at accept. The forest is not in the response. 400 when the session is being saved, 404 when no session has this id, 500 when the scan state cannot be built, and 503 when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartScan),
		}, kickstartErrorCodes{
			http.StatusBadRequest:          {schema.LocalKickstartErrorSessionCommitting},
			http.StatusNotFound:            {schema.LocalKickstartErrorSessionNotFound},
			http.StatusInternalServerError: {schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach},
			http.StatusServiceUnavailable:  kickstartUnavailable(),
		}),
		withKickstartErrors(localOperationSpec{
			method: http.MethodPost, path: "/api/v1/kickstart/sessions/{id}/commit", id: "commitKickstartSession", tag: "kickstart",
			description: "Recompute, merge, swap the configuration, write retention, and start the import. 409 returns the mine/theirs rows and writes nothing. The success body carries the finish actions and the ingest state; it does not install a hook and has no reset control. 400 when the save is refused, for example because an answer is not acceptable or the session is already being saved; nothing is written. 404 when no session has this id. 500 when the configuration file cannot be read or written, and the previous file is unchanged. 503 when another change holds the configuration file, when it kept changing, or when the setup service is unavailable.",
			requests:    []interface{}{sessionPath},
			response:    new(schema.LocalKickstartCommitted),
			alternateResponses: map[int]interface{}{
				http.StatusConflict: new(schema.LocalKickstartCommitRefusal),
			},
		}, kickstartErrorCodes{
			http.StatusBadRequest: {schema.LocalKickstartErrorSessionCommitting, schema.LocalKickstartErrorCommitRefused},
			http.StatusNotFound:   {schema.LocalKickstartErrorSessionNotFound},
			http.StatusInternalServerError: {
				schema.LocalKickstartErrorConfigUnreadable, schema.LocalKickstartErrorConfigUnwritable,
				schema.LocalKickstartErrorSessionUnbuildable, schema.LocalKickstartErrorFailed, schema.LocalKickstartErrorContractBreach,
			},
			http.StatusServiceUnavailable: {
				schema.LocalKickstartErrorUnavailable, schema.LocalKickstartErrorConfigBusy, schema.LocalKickstartErrorConfigChanging,
			},
		}),
	})
}

// describeErrorCodes states the closed codes of each declared error status in
// that status's response description. It fails when errorCodes and
// errorStatuses disagree, because a status with no codes, or codes for an
// undeclared status, would leave the document describing a different set than
// the one it declares.
func describeErrorCodes(spec *openapi31.Spec, op localOperationSpec) error {
	if op.errorCodes == nil {
		return nil
	}
	declared := make(map[int]struct{}, len(op.errorStatuses))
	for _, status := range op.errorStatuses {
		declared[status] = struct{}{}
		if len(op.errorCodes[status]) == 0 {
			return fmt.Errorf("local API operation %s %s at openapi.describeErrorCodes: error status %d has no codes, so a client could not tell which tokens it carries; list the codes for %d in errorCodes", op.method, op.path, status, status)
		}
	}
	for status := range op.errorCodes {
		if _, ok := declared[status]; !ok {
			return fmt.Errorf("local API operation %s %s at openapi.describeErrorCodes: errorCodes names status %d, which errorStatuses does not declare; declare the status or drop its codes", op.method, op.path, status)
		}
	}
	item, ok := spec.Paths.MapOfPathItemValues[op.path]
	if !ok {
		return fmt.Errorf("local API operation %s %s at openapi.describeErrorCodes: the path is not in the document, so its error statuses cannot be described; add the operation before describing it", op.method, op.path)
	}
	operation := pathOperation(item, op.method)
	if operation == nil {
		return fmt.Errorf("local API operation %s %s at openapi.describeErrorCodes: the path has no %s operation, so its error statuses cannot be described", op.method, op.path, op.method)
	}
	for _, status := range op.errorStatuses {
		key := statusText(status)
		response, ok := operation.Responses.MapOfResponseOrReferenceValues[key]
		if !ok || response.Response == nil {
			return fmt.Errorf("local API operation %s %s at openapi.describeErrorCodes: status %s is not declared, so its codes cannot be described; declare the status before listing its codes", op.method, op.path, key)
		}
		tokens := make([]string, len(op.errorCodes[status]))
		for i, code := range op.errorCodes[status] {
			tokens[i] = string(code)
		}
		response.Response.Description = kickstartStatusSummary(status) + " The JSON error code is one of: " + strings.Join(tokens, ", ") + "."
		operation.Responses.MapOfResponseOrReferenceValues[key] = response
	}
	return nil
}

func pathOperation(item openapi31.PathItem, method string) *openapi31.Operation {
	switch method {
	case http.MethodGet:
		return item.Get
	case http.MethodPost:
		return item.Post
	case http.MethodPatch:
		return item.Patch
	case http.MethodPut:
		return item.Put
	case http.MethodDelete:
		return item.Delete
	default:
		return nil
	}
}

func kickstartStatusSummary(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "The request was refused and nothing was changed."
	case http.StatusNotFound:
		return "No session or forest row matches."
	case http.StatusInternalServerError:
		return "The server could not finish the request."
	case http.StatusServiceUnavailable:
		return "The setup service cannot take the request now."
	default:
		return http.StatusText(status) + "."
	}
}
