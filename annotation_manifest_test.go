package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// TestAnnotationManifestResponse_JSONRoundTrip verifies that marshaling and
// unmarshaling a manifest preserves both the hash set and the digest.
func TestAnnotationManifestResponse_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	orig := schema.NewAnnotationManifestResponse([]string{"aaa", "bbb", "ccc"})

	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got schema.AnnotationManifestResponse
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Hashes) != len(orig.Hashes) {
		t.Fatalf("hash count: got %d, want %d", len(got.Hashes), len(orig.Hashes))
	}
	for i := range orig.Hashes {
		if got.Hashes[i] != orig.Hashes[i] {
			t.Errorf("hash[%d]: got %q, want %q", i, got.Hashes[i], orig.Hashes[i])
		}
	}
	if got.Digest != orig.Digest {
		t.Errorf("digest: got %q, want %q", got.Digest, orig.Digest)
	}
}

// TestAnnotationManifestResponse_JSONKeys verifies the wire shape uses the exact
// JSON keys the village/client contract depends on.
func TestAnnotationManifestResponse_JSONKeys(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(schema.NewAnnotationManifestResponse([]string{"x"}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	js := string(b)
	for _, key := range []string{`"hashes"`, `"digest"`} {
		if !strings.Contains(js, key) {
			t.Errorf("manifest JSON missing key %s: %s", key, js)
		}
	}
}

// TestComputeManifestDigest_OrderIndependent is the core determinism guarantee:
// the digest depends only on the SET of hashes, not their input order.
func TestComputeManifestDigest_OrderIndependent(t *testing.T) {
	t.Parallel()

	ascending := []string{"aaa", "bbb", "ccc", "ddd"}
	shuffled := []string{"ccc", "aaa", "ddd", "bbb"}

	if got, want := schema.ComputeManifestDigest(shuffled), schema.ComputeManifestDigest(ascending); got != want {
		t.Errorf("digest is order-dependent: %q != %q", got, want)
	}
}

// TestComputeManifestDigest_DuplicateIndependent verifies the digest has SET
// semantics: duplicate multiplicity does not change the result.
func TestComputeManifestDigest_DuplicateIndependent(t *testing.T) {
	t.Parallel()

	withDupes := []string{"aaa", "aaa", "bbb", "bbb", "bbb"}
	unique := []string{"aaa", "bbb"}

	if got, want := schema.ComputeManifestDigest(withDupes), schema.ComputeManifestDigest(unique); got != want {
		t.Errorf("digest is multiplicity-dependent: %q != %q", got, want)
	}
}

// TestComputeManifestDigest_DistinctSetsDiffer verifies different sets produce
// different digests (the short-circuit must not collide on real divergence).
func TestComputeManifestDigest_DistinctSetsDiffer(t *testing.T) {
	t.Parallel()

	a := schema.ComputeManifestDigest([]string{"aaa", "bbb"})
	b := schema.ComputeManifestDigest([]string{"aaa", "ccc"})
	if a == b {
		t.Errorf("distinct hash sets collided on digest %q", a)
	}
}

// TestComputeManifestDigest_EmptySet verifies the empty set yields a stable,
// well-defined digest so two empty manifests compare equal (no-op short-circuit).
func TestComputeManifestDigest_EmptySet(t *testing.T) {
	t.Parallel()

	if got, want := schema.ComputeManifestDigest(nil), schema.ComputeManifestDigest([]string{}); got != want {
		t.Errorf("empty-set digest unstable: nil=%q, empty=%q", got, want)
	}
	if got := schema.ComputeManifestDigest(nil); got == "" {
		t.Error("empty-set digest must be a defined hash, got empty string")
	}
	if got := schema.ComputeManifestDigest([]string{"aaa"}); got == schema.ComputeManifestDigest(nil) {
		t.Error("non-empty set collided with empty-set digest")
	}
}

// TestNewAnnotationManifestResponse_Normalizes verifies the constructor emits a
// canonical (sorted, de-duplicated) hash set with a digest consistent with it.
func TestNewAnnotationManifestResponse_Normalizes(t *testing.T) {
	t.Parallel()

	r := schema.NewAnnotationManifestResponse([]string{"ccc", "aaa", "bbb", "aaa"})

	want := []string{"aaa", "bbb", "ccc"}
	if len(r.Hashes) != len(want) {
		t.Fatalf("normalized hash count: got %d (%v), want %d", len(r.Hashes), r.Hashes, len(want))
	}
	for i := range want {
		if r.Hashes[i] != want[i] {
			t.Errorf("normalized hash[%d]: got %q, want %q", i, r.Hashes[i], want[i])
		}
	}
	if r.Digest != r.ComputeDigest() {
		t.Errorf("constructor digest %q inconsistent with ComputeDigest %q", r.Digest, r.ComputeDigest())
	}
	if r.Digest != schema.ComputeManifestDigest(want) {
		t.Errorf("constructor digest %q != digest of normalized set %q", r.Digest, schema.ComputeManifestDigest(want))
	}
}

// TestAnnotationPushRequest_RetractionsOmitEmpty verifies the additive retraction
// field is omitted from the wire when empty — preserving the prior contract for
// clients that send no retractions.
func TestAnnotationPushRequest_RetractionsOmitEmpty(t *testing.T) {
	t.Parallel()

	req := schema.AnnotationPushRequest{
		Annotations: []schema.AnnotationPushItem{{ContentHash: "abc", TypeID: "t", Value: "v"}},
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "retractions") {
		t.Errorf("empty Retractions must be omitted from wire: %s", b)
	}
}

// TestAnnotationPushRequest_RetractionsRoundTrip verifies retractions survive a
// marshal/unmarshal cycle when present.
func TestAnnotationPushRequest_RetractionsRoundTrip(t *testing.T) {
	t.Parallel()

	req := schema.AnnotationPushRequest{
		Annotations: []schema.AnnotationPushItem{{ContentHash: "keep", TypeID: "t", Value: "v"}},
		Retractions: []string{"dead1", "dead2"},
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"retractions"`) {
		t.Errorf("present Retractions must appear on wire: %s", b)
	}

	var got schema.AnnotationPushRequest
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Retractions) != 2 || got.Retractions[0] != "dead1" || got.Retractions[1] != "dead2" {
		t.Errorf("retractions round-trip mismatch: got %v", got.Retractions)
	}
	if len(got.Annotations) != 1 || got.Annotations[0].ContentHash != "keep" {
		t.Errorf("annotations corrupted by additive field: got %v", got.Annotations)
	}
}
