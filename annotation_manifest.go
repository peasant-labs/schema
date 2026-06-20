package schema

import (
	"encoding/hex"
	"sort"

	"golang.org/x/crypto/sha3"
)

// AnnotationManifestResponse is the body returned by GET /api/v1/annotations/manifest
// (PROPOSAL-4 C3). It advertises the SET of annotation content-hashes the village
// currently holds for the authenticated owner, so a push client can SKIP any local
// annotation whose hash already appears here (server-authoritative skip-gate).
//
// Hashes are hex-encoded SHA3-256 content-hashes — the SAME identity the client
// already computes via content_hash.go (ComputeContentHash / ComputeAnnotationHash).
// No annotation CONTENT is carried, only its hash, so the manifest is privacy-safe.
//
// Digest is a deterministic, ORDER-INDEPENDENT digest over the hash set (see
// ComputeManifestDigest). A no-op re-push can short-circuit: if the client computes
// the same digest over its own local hash set, nothing has diverged.
//
// This type is ADDITIVE — it introduces a new endpoint's response shape and does
// not alter any existing publish/annotation-push wire contract.
type AnnotationManifestResponse struct {
	Hashes []string `json:"hashes"`
	Digest string   `json:"digest"`
}

// NewAnnotationManifestResponse builds a manifest from a set of content-hashes,
// normalizing the set (sorted + de-duplicated) and computing the matching digest.
//
// Normalizing on construction means the wire payload is canonical: two villages
// holding the same logical set emit byte-identical manifests, and the Digest is
// always consistent with the Hashes it accompanies.
func NewAnnotationManifestResponse(hashes []string) AnnotationManifestResponse {
	normalized := sortedUniqueHashes(hashes)
	return AnnotationManifestResponse{
		Hashes: normalized,
		Digest: digestOfSorted(normalized),
	}
}

// ComputeDigest recomputes the order-independent digest over this manifest's
// Hashes. It is provided so a client can verify the server's Digest field or
// compare against its own locally-derived digest for the no-op short-circuit.
func (r AnnotationManifestResponse) ComputeDigest() string {
	return ComputeManifestDigest(r.Hashes)
}

// ComputeManifestDigest computes a deterministic SHA3-256 digest over a SET of
// content-hashes. The result depends only on the underlying set, not on input
// order or duplicate multiplicity: the hashes are sorted and de-duplicated before
// hashing. Two callers observing the same logical set always produce the same
// digest, which is what makes the no-op short-circuit correct across machines.
//
// The empty set hashes to a fixed, well-defined value (the digest of no input),
// so two empty manifests compare equal.
func ComputeManifestDigest(hashes []string) string {
	return digestOfSorted(sortedUniqueHashes(hashes))
}

// digestOfSorted hashes an ALREADY-sorted, already-deduplicated slice. Each hash
// is delimited by a newline so that concatenation is unambiguous (no two distinct
// sets can serialize to the same byte stream).
func digestOfSorted(sorted []string) string {
	h := sha3.New256()
	for _, hash := range sorted {
		_, _ = h.Write([]byte(hash))
		_, _ = h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// sortedUniqueHashes returns a new slice containing the input hashes sorted
// ascending with adjacent duplicates removed. The input is not mutated. A nil or
// empty input yields an empty (non-nil) slice so callers get stable set semantics.
func sortedUniqueHashes(hashes []string) []string {
	out := make([]string, len(hashes))
	copy(out, hashes)
	sort.Strings(out)

	deduped := out[:0]
	var prev string
	for i, hash := range out {
		if i == 0 || hash != prev {
			deduped = append(deduped, hash)
			prev = hash
		}
	}
	return deduped
}
