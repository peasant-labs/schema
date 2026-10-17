package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

func addLocalHomeOperations(r *openapi31.Reflector) error {
	return addLocalOperations(r, []localOperationSpec{
		{
			method: http.MethodGet, path: "/api/v1/home/sessions", id: "listHomeSessions", tag: "home",
			description:   "Read a bounded page of selection-visible roots, or immediate children of parentSessionId. Roots and children are ordered by startTime descending then local session ID ascending. Apply root project/publication filters before paging. Child requests reject projectHash and publication (400), and unknown or unselected parents return 404 without child identities. limit defaults to 25 (1..100), offset to 0, publication to all. Only page IDs receive title, redaction-safe preview, sync readiness and canonical publication/hook hydration; audience is absent. Auto filtering may inspect candidate repositories before paging; ordinary filters inspect hooks only for returned rows. nextOffset exists exactly when more matching metadata exists and equals offset+limit. Empty rows is []. scopeRevision hashes effective selection and signed-in origin/owner, not a data snapshot; offset pages may shift after inserts/deletes. Invalid bounds, enum, parent or hash return 400. Selection discovery failure returns 500 with code selection_visibility and recovery guidance; unavailable readers return 503 and store failures 500, never empty success.",
			requests:      []interface{}{new(schema.LocalHomeSessionsQuery)},
			response:      new(schema.LocalHomeSessionsPage),
			errorStatuses: []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/home/stats", id: "getHomeStats", tag: "home",
			description:   "Read independent metadata-only Home aggregates after the first page renders. roots counts the global selected root cohort using current signed-in origin/owner receipts across historical project hashes; rootFacets alone is narrowed by projectHash. all = published + notPublished. projects is the complete selected root picker sorted by projectName then projectHash. Activity uses the existing selection-only activity cohort, including children and nonpushable sessions, UTC Monday weeks, the existing streak and rounded nonnegative-duration median; an empty median is null and zero is real. weeks contains eight consecutive weeks in ascending order. No transcript, entries, quality serialization, annotations, Git, hook checks or Village HTTP are read. There is no auto count. scopeRevision hashes effective selection and signed-in origin/owner, not a snapshot. Invalid projectHash returns 400; selection discovery failure returns 500 with code selection_visibility and recovery guidance; unavailable readers return 503 and store failures 500, never partial zero success.",
			requests:      []interface{}{new(schema.LocalHomeStatsQuery)},
			response:      new(schema.LocalHomeStats),
			errorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusServiceUnavailable},
			errorResponse: new(LocalErrorResponse),
		},
	})
}
