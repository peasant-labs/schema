package schema

import "sort"

// PullSkipGateRequest is the body a pulling client POSTs to the pull skip-gate
// endpoint. For each transcript it already holds, the client sends the id plus
// the content-hash it holds and its OWN annotation-hash set for that id, so the
// server can answer, per id, whether the stored transcript and the client's
// annotations are still current, letting the client skip re-downloading what has
// not diverged. Only the client's held HASHES travel, never content, so the
// request is privacy-safe.
//
// This type is ADDITIVE: it introduces a new endpoint's request shape and does
// not alter any existing pull/publish/annotation-push wire contract.
type PullSkipGateRequest struct {
	// The per-transcript items the client is asking a currency question about.
	// NewPullSkipGateRequest orders them by transcriptId and never returns nil, so the
	// wire is a canonical, (possibly empty) NON-null array.
	Items []PullSkipGateItem `json:"items" description:"The per-transcript items the client is asking a currency question about (ordered by transcriptId, de-duplicated per-item annotation set)." nullable:"false"`
}

// PullSkipGateItem is one transcript the client holds and wants a currency answer for:
// the transcript id, the content-hash the client currently has for it, and the set
// of annotation content-hashes the client itself holds for that transcript.
type PullSkipGateItem struct {
	// The id of a transcript the client holds and is asking about.
	TranscriptID TranscriptID `json:"transcriptId" description:"The id of a transcript the client holds and is asking about."`
	// Hex-encoded SHA3-256 content-hash the client currently holds for this
	// transcript's served blob; compared by VALUE against the server's stored hash
	// to decide contentCurrent.
	ContentHash string `json:"contentHash" description:"Hex-encoded SHA3-256 served-blob content-hash the client currently holds for this transcript; compared by value against the server's stored hash."`
	// The client's OWN annotation content-hashes for this transcript (the set it
	// already has locally). Compared as a SET against the owner-scoped server set to
	// decide annotationsCurrent. NewPullSkipGateRequest sorts + de-duplicates it, so it
	// is always a (possibly empty) NON-null array.
	AnnotationHashes []string `json:"annotationHashes" description:"The client's own annotation content-hashes for this transcript (sorted, de-duplicated); compared as a set against the owner-scoped server set." nullable:"false"`
}

// PullSkipGateResponse is the per-id currency answer for a PullSkipGateRequest.
//
// LEAK-FREE WITHHELD SEMANTICS: Results carries an entry ONLY for transcript ids
// the caller may PULL. A non-pullable id is OMITTED from Results entirely, never
// echoed with a "denied" or "unknown" marker, because any per-id echo would itself
// be an existence / currency oracle over arbitrary ids the caller cannot pull. So
// the caller sends N ids and receives <= N results; an ABSENT id means "unanswered
// / withheld", the 404-not-403 anti-enumeration spirit applied to a batch currency
// probe.
//
// This is enforced in two distinct places. The RESPONSE SHAPE carries no denial or
// marker field a withheld id could ride on: it is exactly {results:
// [{transcriptId, contentCurrent, annotationsCurrent}]}, checked by this module's
// exact-key-set response test, so adding and populating a wire-visible extra field
// reddens the assertion. An inert zero-valued `omitempty` field is not emitted and
// is outside this check. The actual OMISSION of non-pullable ids from Results is
// the village handler's pull-scoping test, since only the server knows which ids
// are pullable; this constructor merely canonicalizes the entries it is given and
// never invents one.
//
// This type is ADDITIVE: it introduces a new endpoint's response shape and does
// not alter any existing pull/publish/annotation-push wire contract.
type PullSkipGateResponse struct {
	// The per-id currency answers, present ONLY for pullable ids (non-pullable ids
	// are omitted; see the type doc). NewPullSkipGateResponse orders them by
	// transcriptId and never returns nil, so the wire is a canonical, (possibly
	// empty) NON-null array.
	Results []PullSkipGateResult `json:"results" description:"Per-id currency answers, present only for pullable ids (non-pullable ids are withheld by omission); ordered by transcriptId." nullable:"false"`
}

// PullSkipGateResult is the currency answer for one PULLABLE transcript id: whether the
// stored blob still matches the client's held content-hash, and whether the
// owner-scoped annotation set still matches the client's held annotation set.
type PullSkipGateResult struct {
	// The transcript id this answer is for (always one the caller may pull).
	TranscriptID TranscriptID `json:"transcriptId" description:"The transcript id this currency answer is for (always one the caller may pull)."`
	// True when the server's stored served-blob content-hash equals the hash the
	// client sent for this id; false when it has diverged (or the server holds none).
	ContentCurrent bool `json:"contentCurrent" description:"True when the server's stored content-hash equals the client's held hash for this id; false when it has diverged or the server holds none."`
	// True when the owner-scoped annotation set the server holds for this id equals
	// the client's held annotation set; false when it differs (missing or extra).
	AnnotationsCurrent bool `json:"annotationsCurrent" description:"True when the owner-scoped server annotation set for this id equals the client's held set; false when it differs (missing or extra)."`
}

// NewPullSkipGateRequest builds a CANONICAL skip-gate request: each item's
// annotation-hash set is sorted + de-duplicated (the server compares it as a SET,
// so order and multiplicity are irrelevant), and the items are ordered by
// transcriptId. Provided the request's transcript ids are UNIQUE, two clients
// asking about the same logical state emit byte-identical requests: the items are
// ordered with an unstable sort on transcriptId, so byte-identity holds only when
// that key has no ties (one item per transcript id, which a well-formed request
// satisfies). Every item's AnnotationHashes is a non-null array.
func NewPullSkipGateRequest(items []PullSkipGateItem) PullSkipGateRequest {
	normalized := make([]PullSkipGateItem, len(items))
	for i, item := range items {
		item.AnnotationHashes = sortedUniqueHashes(item.AnnotationHashes)
		normalized[i] = item
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].TranscriptID < normalized[j].TranscriptID
	})
	return PullSkipGateRequest{Items: normalized}
}

// NewPullSkipGateResponse builds a CANONICAL skip-gate response from the per-id answers
// the server computed for the PULLABLE ids ONLY. Non-pullable ids must already be
// omitted by the caller (see PullSkipGateResponse's leak-free contract); this
// constructor does not invent entries. Results are ordered by transcriptId so the
// wire payload is deterministic, PROVIDED the response's transcript ids are UNIQUE:
// the results are ordered with an unstable sort on transcriptId, so determinism
// holds only when that key has no ties (one result per pullable id, which the
// pull-scope answer satisfies). The slice is never nil.
func NewPullSkipGateResponse(results []PullSkipGateResult) PullSkipGateResponse {
	out := make([]PullSkipGateResult, len(results))
	copy(out, results)
	sort.Slice(out, func(i, j int) bool {
		return out[i].TranscriptID < out[j].TranscriptID
	})
	return PullSkipGateResponse{Results: out}
}
