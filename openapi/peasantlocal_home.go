package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

func addLocalHomeOperations(r *openapi31.Reflector) error {
	return addLocalOperations(r, []localOperationSpec{{
		method: http.MethodGet, path: "/api/v1/home/sessions", id: "listHomeSessions", tag: "home",
		description:   "Read a bounded page of selection-visible roots, or immediate children of parentSessionId. Roots and children are ordered by startTime descending then local session ID ascending. Apply root project/publication filters before paging. Child requests reject projectHash and publication (400), and unknown or unselected parents return 404 without child identities. limit defaults to 25 (1..100), offset to 0, publication to all. Only page IDs receive title, redaction-safe preview, sync readiness and canonical publication/hook hydration; audience is absent. Auto filtering may inspect candidate repositories before paging; ordinary filters inspect hooks only for returned rows. nextOffset exists exactly when more matching metadata exists and equals offset+limit. Empty rows is []. Offset pages are live, not a data snapshot, and may shift after inserts/deletes. Invalid bounds, enum, parent or hash return 400. Selection discovery failure returns 500 with code selection_visibility and recovery guidance; unavailable readers return 503 and store failures 500, never empty success.",
		requests:      []interface{}{new(schema.LocalHomeSessionsQuery)},
		response:      new(schema.LocalHomeSessionsPage),
		errorStatuses: []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable},
		errorResponse: new(LocalErrorResponse),
	}})
}
