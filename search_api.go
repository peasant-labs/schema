package schema

// Search wire contract for full-text transcript search.
//
// SearchPayload backs GET /api/v1/search, the Cmd-K "Messages" group: a
// global, FTS5-backed full-text search over recorded (redacted) message
// entries. Each result deep-links to one task turn via (sessionId, entryIndex)
// — the same coordinates ProjectTasks / TaskSummary use. JSON field names are
// camelCase; Results serializes as [] — never null. Use NewSearchPayload (or
// otherwise guarantee a non-nil slice) before marshaling.
//
// Search matches only the recorded text: content_preview (capped + redacted,
// first ~2000 chars of each turn) plus truncated tool_input / tool_output.
// Full-content search is out of scope — full text exists only transiently in
// the export overlay.

// SearchResult is one FTS5 hit: a single message entry, ranked by relevance,
// with a snippet for display and the coordinates to deep-link to it.
type SearchResult struct {
	SessionID   string      `json:"sessionId"`
	Project     string      `json:"project"`               // raw canonical_cwd (else hash); web formats for display
	ProjectHash ProjectHash `json:"projectHash,omitempty"` // for round-trips
	EntryIndex  int         `json:"entryIndex"`            // depth-0 turn index — deep-link coordinate
	Role        string      `json:"role"`                  // user | assistant | ... (display facet)
	Snippet     string      `json:"snippet"`               // FTS5 snippet() with [match] markers
	Score       float64     `json:"score"`                 // negated bm25: higher = more relevant (result order is authoritative)
}

// SearchPayload is the result set for one query, served by GET /api/v1/search.
type SearchPayload struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

// NewSearchPayload returns a SearchPayload with Results initialized to empty
// (never-nil marshal guarantee).
func NewSearchPayload(query string) *SearchPayload {
	return &SearchPayload{Query: query, Results: []SearchResult{}}
}
