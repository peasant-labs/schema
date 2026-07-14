package schema_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/pull/skip_gate_cases.yaml
var skipGateCasesYAML []byte

// skipGateItemRow / skipGateResultRow mirror the YAML rows for skip-gate request
// items and response results.
type skipGateItemRow struct {
	TranscriptID     string   `yaml:"transcript_id"`
	ContentHash      string   `yaml:"content_hash"`
	AnnotationHashes []string `yaml:"annotation_hashes"`
}

type skipGateResultRow struct {
	TranscriptID       string `yaml:"transcript_id"`
	ContentCurrent     bool   `yaml:"content_current"`
	AnnotationsCurrent bool   `yaml:"annotations_current"`
}

// The per-arm input and expected types. Each arm's Corpus.Input is a COLLECTION
// (the request items or the response results) and its Expected is the global
// property the harness checks over that collection; the harness, not the Corpus,
// owns the comparison.
type itemsInput struct {
	Items []skipGateItemRow `yaml:"items"`
}

type resultsInput struct {
	Results []skipGateResultRow `yaml:"results"`
}

// canonicalExpected is the canonical form a request's items must take: the id
// order and, per id, the sorted + de-duplicated annotation set.
type canonicalExpected struct {
	Order          []string            `yaml:"order"`
	AnnotationSets map[string][]string `yaml:"annotation_sets"`
}

// orderExpected is the id order a response's results must take.
type orderExpected struct {
	Order []string `yaml:"order"`
}

// withheldExpected splits the asked ids into the ones the response must answer
// (present) and the ones it must withhold (absent from the results AND the bytes).
type withheldExpected struct {
	Present []string `yaml:"present"`
	Absent  []string `yaml:"absent"`
}

// skipGateFixtures is the segmented multi-axis fixture for the skip-gate wire
// types: a typed struct of named testcase.Corpus arms, one per behaviour, each
// with its own static (input, expected) types. It is the worked reference example
// of the segmented-fixture convention documented in TESTING.md.
type skipGateFixtures struct {
	RoundTrip          testcase.Corpus[itemsInput, struct{}]           `yaml:"round_trip"`
	Canonical          testcase.Corpus[itemsInput, canonicalExpected]  `yaml:"canonical"`
	OrdersByTranscript testcase.Corpus[resultsInput, orderExpected]    `yaml:"orders_by_transcript_id"`
	Withheld           testcase.Corpus[resultsInput, withheldExpected] `yaml:"withheld_by_omission"`
}

// LoadSkipGateFixtures decodes the embedded case corpus and guards every arm. The
// per-arm RequireMin enforces a case floor and RequireValid enforces per-case
// non-vacuity (in-set classification + provenance, non-empty ref + mutation), so a
// truncated, mis-keyed, or under-populated corpus fails loudly instead of letting
// an iterating test pass vacuously. All four arms are wired to BOTH guards.
func LoadSkipGateFixtures(t *testing.T) skipGateFixtures {
	t.Helper()
	var fx skipGateFixtures
	if err := yaml.Unmarshal(skipGateCasesYAML, &fx); err != nil {
		t.Fatalf("load skip-gate fixtures (testdata/pull/skip_gate_cases.yaml): %v", err)
	}
	assert.RequireMin(t, fx.RoundTrip, 2)
	assert.RequireValid(t, fx.RoundTrip)
	assert.RequireMin(t, fx.Canonical, 2)
	assert.RequireValid(t, fx.Canonical)
	assert.RequireMin(t, fx.OrdersByTranscript, 2)
	assert.RequireValid(t, fx.OrdersByTranscript)
	assert.RequireMin(t, fx.Withheld, 2)
	assert.RequireValid(t, fx.Withheld)
	return fx
}

// toSkipGateItems / toSkipGateResults convert fixture rows to the production wire
// types.
func toSkipGateItems(rows []skipGateItemRow) []schema.PullSkipGateItem {
	out := make([]schema.PullSkipGateItem, len(rows))
	for i, r := range rows {
		out[i] = schema.PullSkipGateItem{
			TranscriptID:     schema.TranscriptID(r.TranscriptID),
			ContentHash:      r.ContentHash,
			AnnotationHashes: r.AnnotationHashes,
		}
	}
	return out
}

func toSkipGateResults(rows []skipGateResultRow) []schema.PullSkipGateResult {
	out := make([]schema.PullSkipGateResult, len(rows))
	for i, r := range rows {
		out[i] = schema.PullSkipGateResult{
			TranscriptID:       schema.TranscriptID(r.TranscriptID),
			ContentCurrent:     r.ContentCurrent,
			AnnotationsCurrent: r.AnnotationsCurrent,
		}
	}
	return out
}

// assertResponseShapeExact fails unless the marshaled response has EXACTLY the
// expected key-set at the top level and on each result, so an unexpected extra
// field (for example a denial or marker key) or a missing field reddens. This is
// the structural half of the leak-free contract: the response SHAPE carries no
// field that could echo a withheld id. The behavioral omission of non-pullable ids
// is enforced by village's handler test, not by this schema-only test.
func assertResponseShapeExact(t *testing.T, name string, respBytes []byte) {
	t.Helper()
	var top map[string]json.RawMessage
	if err := json.Unmarshal(respBytes, &top); err != nil {
		t.Fatalf("%s: unmarshal response: %v", name, err)
	}
	assertExactKeys(t, name+" response", top, "results")
	var results []map[string]json.RawMessage
	if err := json.Unmarshal(top["results"], &results); err != nil {
		t.Fatalf("%s: unmarshal results: %v", name, err)
	}
	for _, r := range results {
		assertExactKeys(t, name+" result", r, "transcriptId", "contentCurrent", "annotationsCurrent")
	}
}

// assertExactKeys fails unless obj's key-set equals want exactly: a missing key
// and an unexpected extra key each redden.
func assertExactKeys(t *testing.T, what string, obj map[string]json.RawMessage, want ...string) {
	t.Helper()
	wantSet := make(map[string]bool, len(want))
	for _, k := range want {
		wantSet[k] = true
		if _, ok := obj[k]; !ok {
			t.Errorf("%s: missing required key %q", what, k)
		}
	}
	for k := range obj {
		if !wantSet[k] {
			t.Errorf("%s: unexpected extra key %q on the wire", what, k)
		}
	}
}

// requireCoverage asserts that at least one case in the arm matches the property
// predicate, covering a required scenario by PROPERTY, never by case name or index.
// It is the per-arm predicate-based coverage assertion: dropping the covered case
// reddens here even if a valid filler keeps the case count at the floor, so a
// count-preserving swap that would leave the test passing vacuously is caught
// instead of silently deleting a behavior.
func requireCoverage[I any, E any](t *testing.T, cases []testcase.Case[I, E], pred func(testcase.Case[I, E]) bool, desc string) {
	t.Helper()
	for _, c := range cases {
		if pred(c) {
			return
		}
	}
	t.Errorf("no case covers the required scenario: %s", desc)
}

// strictlyAscending reports whether s is strictly ascending, so it is both sorted
// and free of duplicates.
func strictlyAscending(s []string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] >= s[i] {
			return false
		}
	}
	return true
}

// itemsAlreadyCanonical reports whether the items are already in the constructor's
// canonical form (ordered by transcriptId with each annotation set sorted and
// de-duplicated), which is the idempotence scenario, detected here by data property
// alone so it survives a count-preserving swap.
func itemsAlreadyCanonical(items []skipGateItemRow) bool {
	if len(items) < 2 {
		return false
	}
	ids := make([]string, len(items))
	for i, it := range items {
		ids[i] = it.TranscriptID
		if !strictlyAscending(it.AnnotationHashes) {
			return false
		}
	}
	return strictlyAscending(ids)
}

// resultsAscending reports whether the results are already ordered by transcriptId.
func resultsAscending(results []skipGateResultRow) bool {
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.TranscriptID
	}
	return strictlyAscending(ids)
}

// hasOutOfOrderIDs reports whether the items' ids are NOT already ascending, so the
// case exercises the transcriptId ordering step.
func hasOutOfOrderIDs(items []skipGateItemRow) bool {
	if len(items) < 2 {
		return false
	}
	ids := make([]string, len(items))
	for i, it := range items {
		ids[i] = it.TranscriptID
	}
	return !strictlyAscending(ids)
}

// hasUnsortedSet reports whether some item's annotation set is out of ascending
// order, so the case exercises the per-set sort step.
func hasUnsortedSet(items []skipGateItemRow) bool {
	for _, it := range items {
		for i := 1; i < len(it.AnnotationHashes); i++ {
			if it.AnnotationHashes[i-1] > it.AnnotationHashes[i] {
				return true
			}
		}
	}
	return false
}

// hasDuplicateHash reports whether some item's annotation set contains a duplicate,
// so the case exercises the per-set de-duplication step.
func hasDuplicateHash(items []skipGateItemRow) bool {
	for _, it := range items {
		seen := make(map[string]bool, len(it.AnnotationHashes))
		for _, h := range it.AnnotationHashes {
			if seen[h] {
				return true
			}
			seen[h] = true
		}
	}
	return false
}

// canonicalIsNoOp reports whether a canonical case is a genuine idempotence case:
// its input is already in canonical form AND its expected is a no-op of that input
// (expected order equals the input ids in order, and each expected annotation set
// equals the input set for that id). Computed from the loaded case alone, so it
// asserts coverage by structure, not by re-hardcoding fixture rows.
func canonicalIsNoOp(c testcase.Case[itemsInput, canonicalExpected]) bool {
	if !itemsAlreadyCanonical(c.Input.Items) {
		return false
	}
	if len(c.Expected.Order) != len(c.Input.Items) {
		return false
	}
	for i, it := range c.Input.Items {
		if c.Expected.Order[i] != it.TranscriptID {
			return false
		}
		if strings.Join(c.Expected.AnnotationSets[it.TranscriptID], ",") != strings.Join(it.AnnotationHashes, ",") {
			return false
		}
	}
	return true
}

// TestSkipGateFixtures_Coverage asserts coverage of each arm's required scenarios
// by property, so a count-preserving swap (dropping a covered case and adding a
// valid filler that keeps the count at the floor) reddens instead of silently
// losing the behavior. Each arm covers BOTH its boundary scenario(s) and the
// representative case its harness mutation-proof depends on: without the
// representative coverage assertion, swapping the messy/scrambled/mixed
// representative case for a second boundary case would leave the arm's harness
// mutation-proof vacuous (e.g. a canonical corpus of only already-canonical inputs
// no longer catches a broken sort). It complements the RequireMin floor and the
// RequireValid non-vacuity guard, the same coverage-assertion discipline the
// license and grammar corpora apply.
func TestSkipGateFixtures_Coverage(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	// round_trip: the nil/empty boundary case, and a multi-item-distinct representative case so fidelity is real.
	requireCoverage(t, fx.RoundTrip.Cases, func(c testcase.Case[itemsInput, struct{}]) bool {
		for _, it := range c.Input.Items {
			if len(it.AnnotationHashes) == 0 {
				return true
			}
		}
		return false
	}, "round_trip: a case with a nil or empty annotation set (the nil-to-non-null normalization path)")
	requireCoverage(t, fx.RoundTrip.Cases, func(c testcase.Case[itemsInput, struct{}]) bool {
		if len(c.Input.Items) < 2 {
			return false
		}
		ids := make(map[string]bool, len(c.Input.Items))
		hasSet := false
		for _, it := range c.Input.Items {
			ids[it.TranscriptID] = true
			if len(it.AnnotationHashes) > 0 {
				hasSet = true
			}
		}
		return len(ids) == len(c.Input.Items) && hasSet
	}, "round_trip: a multi-item representative case with distinct ids and a non-empty annotation set, so round-trip fidelity is non-vacuous")

	// canonical: the idempotence boundary case, and three representative cases so ordering,
	// per-set sort, and per-set dedup are each independently exercised (a single
	// "messy" predicate would let an order-only case mask a dropped-dedup bug).
	requireCoverage(t, fx.Canonical.Cases, canonicalIsNoOp,
		"canonical: an already-canonical (idempotence) case whose input is canonical and whose expected is a no-op of that input")
	requireCoverage(t, fx.Canonical.Cases, func(c testcase.Case[itemsInput, canonicalExpected]) bool {
		return hasOutOfOrderIDs(c.Input.Items)
	}, "canonical: a case with an out-of-order id pair, so the transcriptId ordering is exercised")
	requireCoverage(t, fx.Canonical.Cases, func(c testcase.Case[itemsInput, canonicalExpected]) bool {
		return hasUnsortedSet(c.Input.Items)
	}, "canonical: a case with an unsorted annotation set, so the per-set sort is exercised")
	requireCoverage(t, fx.Canonical.Cases, func(c testcase.Case[itemsInput, canonicalExpected]) bool {
		return hasDuplicateHash(c.Input.Items)
	}, "canonical: a case with a duplicated annotation hash, so the per-set de-duplication is exercised")

	// orders: the already-ordered and empty boundary cases, and a scrambled representative case that exercises the sort.
	requireCoverage(t, fx.OrdersByTranscript.Cases, func(c testcase.Case[resultsInput, orderExpected]) bool {
		return len(c.Input.Results) >= 2 && resultsAscending(c.Input.Results)
	}, "orders_by_transcript_id: an already-ordered case (results already ascending by transcriptId)")
	requireCoverage(t, fx.OrdersByTranscript.Cases, func(c testcase.Case[resultsInput, orderExpected]) bool {
		return len(c.Input.Results) == 0
	}, "orders_by_transcript_id: an empty case (no results, the empty-to-non-null normalization)")
	requireCoverage(t, fx.OrdersByTranscript.Cases, func(c testcase.Case[resultsInput, orderExpected]) bool {
		return len(c.Input.Results) >= 2 && !resultsAscending(c.Input.Results)
	}, "orders_by_transcript_id: a scrambled representative case (multi-result input not already ordered) that exercises the sort")

	// withheld: the all- and none-withheld boundary cases, and a mixed representative case (the primary leak-free assertion).
	requireCoverage(t, fx.Withheld.Cases, func(c testcase.Case[resultsInput, withheldExpected]) bool {
		return len(c.Expected.Present) == 0 && len(c.Expected.Absent) > 0
	}, "withheld_by_omission: an all-withheld case (no present ids, some absent)")
	requireCoverage(t, fx.Withheld.Cases, func(c testcase.Case[resultsInput, withheldExpected]) bool {
		return len(c.Expected.Present) >= 1 && len(c.Expected.Absent) == 0
	}, "withheld_by_omission: a none-withheld case (some present, no absent)")
	requireCoverage(t, fx.Withheld.Cases, func(c testcase.Case[resultsInput, withheldExpected]) bool {
		return len(c.Expected.Present) >= 1 && len(c.Expected.Absent) >= 1
	}, "withheld_by_omission: a mixed some-withheld representative case (present and absent both non-empty), the primary leak-free assertion")
}

// TestPullSkipGateRequest_JSONRoundTrip verifies that, for each round-trip case,
// marshaling then unmarshaling a request preserves every id, content-hash, and
// annotation-hash set, and that every annotation set normalizes to a non-null
// array (folding in the former NonNullAnnotationHashes check).
func TestPullSkipGateRequest_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	for _, c := range fx.RoundTrip.Cases {
		orig := schema.NewPullSkipGateRequest(toSkipGateItems(c.Input.Items))

		for i, item := range orig.Items {
			if item.AnnotationHashes == nil {
				t.Errorf("%s: item[%d] annotation set must normalize to a non-null array", c.Name, i)
			}
		}

		b, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("%s: marshal: %v", c.Name, err)
		}
		if strings.Contains(string(b), `"annotationHashes":null`) {
			t.Errorf("%s: annotationHashes serialized as null: %s", c.Name, b)
		}

		var got schema.PullSkipGateRequest
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("%s: unmarshal: %v", c.Name, err)
		}
		if len(got.Items) != len(orig.Items) {
			t.Fatalf("%s: item count: got %d, want %d", c.Name, len(got.Items), len(orig.Items))
		}
		for i := range orig.Items {
			if got.Items[i].TranscriptID != orig.Items[i].TranscriptID {
				t.Errorf("%s: item[%d] transcriptId: got %q, want %q", c.Name, i, got.Items[i].TranscriptID, orig.Items[i].TranscriptID)
			}
			if got.Items[i].ContentHash != orig.Items[i].ContentHash {
				t.Errorf("%s: item[%d] contentHash: got %q, want %q", c.Name, i, got.Items[i].ContentHash, orig.Items[i].ContentHash)
			}
			if strings.Join(got.Items[i].AnnotationHashes, ",") != strings.Join(orig.Items[i].AnnotationHashes, ",") {
				t.Errorf("%s: item[%d] annotationHashes: got %v, want %v", c.Name, i, got.Items[i].AnnotationHashes, orig.Items[i].AnnotationHashes)
			}
		}
	}
}

// TestPullSkipGateRequest_JSONKeys pins the exact request wire keys the village
// handler and the client contract depend on. The keys are a structural contract
// (not case data); the request bytes come from the corpus.
func TestPullSkipGateRequest_JSONKeys(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	b, err := json.Marshal(schema.NewPullSkipGateRequest(toSkipGateItems(fx.RoundTrip.Cases[0].Input.Items)))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	js := string(b)
	for _, key := range []string{`"items"`, `"transcriptId"`, `"contentHash"`, `"annotationHashes"`} {
		if !strings.Contains(js, key) {
			t.Errorf("request JSON missing key %s: %s", key, js)
		}
	}
}

// TestPullSkipGateResponse_JSONKeys pins the exact response wire keys (structural
// contract; bytes come from the corpus).
func TestPullSkipGateResponse_JSONKeys(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	b, err := json.Marshal(schema.NewPullSkipGateResponse(toSkipGateResults(fx.OrdersByTranscript.Cases[0].Input.Results)))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	js := string(b)
	for _, key := range []string{`"results"`, `"transcriptId"`, `"contentCurrent"`, `"annotationsCurrent"`} {
		if !strings.Contains(js, key) {
			t.Errorf("response JSON missing key %s: %s", key, js)
		}
	}
}

// TestNewPullSkipGateRequest_Canonical verifies the constructor produces a canonical
// request for each case: items ordered by transcriptId, and each item's annotation
// set sorted + de-duplicated (SET semantics: the server compares it as a set). The
// idempotence case pins that an already-canonical input is a no-op.
func TestNewPullSkipGateRequest_Canonical(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	for _, c := range fx.Canonical.Cases {
		req := schema.NewPullSkipGateRequest(toSkipGateItems(c.Input.Items))

		if len(req.Items) != len(c.Expected.Order) {
			t.Fatalf("%s: item count: got %d, want %d", c.Name, len(req.Items), len(c.Expected.Order))
		}
		for i, wantID := range c.Expected.Order {
			if string(req.Items[i].TranscriptID) != wantID {
				t.Errorf("%s: items not ordered by transcriptId at %d: got %q, want %q", c.Name, i, req.Items[i].TranscriptID, wantID)
			}
		}
		for _, item := range req.Items {
			want := strings.Join(c.Expected.AnnotationSets[string(item.TranscriptID)], ",")
			if got := strings.Join(item.AnnotationHashes, ","); got != want {
				t.Errorf("%s: annotation set for %s not sorted+deduped: got %q, want %q", c.Name, item.TranscriptID, got, want)
			}
		}
	}
}

// TestNewPullSkipGateResponse_OrdersByTranscriptID verifies results are canonicalized
// by transcriptId regardless of input order, and that the slice is never nil (the
// empty case pins the non-null normalization).
func TestNewPullSkipGateResponse_OrdersByTranscriptID(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	for _, c := range fx.OrdersByTranscript.Cases {
		resp := schema.NewPullSkipGateResponse(toSkipGateResults(c.Input.Results))

		if resp.Results == nil {
			t.Errorf("%s: results must be a non-null array", c.Name)
		}
		if len(resp.Results) != len(c.Expected.Order) {
			t.Fatalf("%s: result count: got %d, want %d", c.Name, len(resp.Results), len(c.Expected.Order))
		}
		for i, wantID := range c.Expected.Order {
			if string(resp.Results[i].TranscriptID) != wantID {
				t.Errorf("%s: results not ordered by transcriptId at %d: got %q, want %q", c.Name, i, resp.Results[i].TranscriptID, wantID)
			}
		}
	}
}

// TestPullSkipGateResponse_WithheldByOmission checks the behavioral half of the
// leak-free contract per case at the schema level: a response built from only the
// PULLABLE ids answers exactly those ids (present answered, withheld absent from
// Results). That the response SHAPE carries no denial or marker field a withheld id
// could ride on is TestPullSkipGateResponse_ExactShape; whether the server actually
// omits non-pullable ids is village's handler test. A raw substring scan for a
// withheld id here would be vacuous, since withheld ids never enter the constructor.
func TestPullSkipGateResponse_WithheldByOmission(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	for _, c := range fx.Withheld.Cases {
		resp := schema.NewPullSkipGateResponse(toSkipGateResults(c.Input.Results))

		present := map[string]bool{}
		for _, r := range resp.Results {
			present[string(r.TranscriptID)] = true
		}
		for _, id := range c.Expected.Present {
			if !present[id] {
				t.Errorf("%s: pullable id %s must be answered: %v", c.Name, id, resp.Results)
			}
		}
		for _, id := range c.Expected.Absent {
			if present[id] {
				t.Errorf("%s: withheld id %s must be absent from Results, never echoed: %v", c.Name, id, resp.Results)
			}
		}
	}
}

// TestPullSkipGateResponse_ExactShape enforces the response SHAPE at the type level:
// the marshaled response has EXACTLY {"results"} at the top level and each result
// has EXACTLY {"transcriptId","contentCurrent","annotationsCurrent"}, so adding and
// populating a wire-visible extra field on PullSkipGateResponse or
// PullSkipGateResult reddens here, not merely the freshness gate. An inert
// zero-valued `omitempty` field is not emitted and is outside this assertion. This
// is the structural half of the leak-free contract: the wire carries no denial or
// marker field a withheld id could ride on. The behavioral omission of
// non-pullable ids is village's handler test.
func TestPullSkipGateResponse_ExactShape(t *testing.T) {
	t.Parallel()
	fx := LoadSkipGateFixtures(t)

	// Build a response with at least one result so the per-result key-set is checked.
	var results []skipGateResultRow
	for _, c := range fx.OrdersByTranscript.Cases {
		if len(c.Input.Results) > 0 {
			results = c.Input.Results
			break
		}
	}
	if len(results) == 0 {
		t.Fatal("no orders case has results to build a non-empty response from")
	}
	b, err := json.Marshal(schema.NewPullSkipGateResponse(toSkipGateResults(results)))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	assertResponseShapeExact(t, "skip-gate response", b)
}
