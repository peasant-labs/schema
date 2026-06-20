package schema

import (
	"encoding/hex"
	"encoding/json"

	"golang.org/x/crypto/sha3"
)

// ComputeTranscriptHash computes the SHA3-256 hash of raw transcript bytes.
// Used to detect content changes for incremental re-processing and stale redaction detection.
func ComputeTranscriptHash(transcriptBytes []byte) string {
	h := sha3.New256()
	_, _ = h.Write(transcriptBytes)
	return hex.EncodeToString(h.Sum(nil))
}

// ComputeMetadataHash computes the SHA3-256 hash of metadata fields,
// excluding ContentHash, MetadataHash, and Redaction to avoid circular dependency.
// Changing any content-bearing metadata field produces a different hash.
func ComputeMetadataHash(meta *UnifiedMetadata) string {
	cp := *meta
	cp.ContentHash = ""
	cp.MetadataHash = ""
	cp.Redaction = RedactionInfo{}
	cp.DerivedAt = nil

	data, err := json.Marshal(&cp)
	if err != nil {
		return ""
	}
	h := sha3.New256()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// ComputeContentHash computes a SHA3-256 content hash over the canonical JSON
// representation of the AnnotationPushItem, excluding the ContentHash field itself.
//
// The hash is deterministic: the same logical annotation always produces the same
// hex string regardless of struct field order. This is because encoding/json marshals
// struct fields in declaration order, producing stable output for the same values.
//
// Use case: server-side deduplication. The caller sets item.ContentHash to this
// value before adding it to an AnnotationPushRequest.
func (item *AnnotationPushItem) ComputeContentHash() string {
	// Create a copy with ContentHash zeroed so the hash is over content only.
	copy := *item
	copy.ContentHash = ""

	b, err := json.Marshal(&copy)
	if err != nil {
		// json.Marshal only fails for types with cyclic references or marshal errors.
		// AnnotationPushItem has no such types; treat as an empty payload.
		b = []byte("{}")
	}

	h := sha3.New256()
	_, _ = h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// annotationHashPayload is the canonical struct for computing content hashes
// for annotation deduplication (R9). Field order is fixed — changing it would
// alter all existing hashes. New fields must be appended at the end.
//
// This struct covers the content-bearing fields of an annotation. It is used
// by both session-level and entry-level annotations via ComputeAnnotationHash.
type annotationHashPayload struct {
	AnnotationTypeID string      `json:"annotationTypeId"`
	AnnotatorID      string      `json:"annotatorId"`
	Value            string      `json:"value"`
	SessionID        *string     `json:"sessionId,omitempty"`
	EntryIndex       *int        `json:"entryIndex,omitempty"`
	EndIndex         *int        `json:"endIndex,omitempty"`
	Confidence       *float64    `json:"confidence,omitempty"`
	Reason           *string     `json:"reason,omitempty"`
	Provenance       *Provenance `json:"provenance,omitempty"`
}

// ComputeAnnotationHash computes a SHA3-256 content hash over the core fields of
// an annotation for deduplication purposes. The hash is deterministic: the same
// annotation content always produces the same hex string.
//
// Parameters mirror the annotation's content-bearing fields. Target-specific
// fields (sessionID, entryIndex, endIndex) are optional and contribute to the
// hash only when set (ensuring session-level and entry-level annotations of
// the same type/value/annotator produce different hashes).
func ComputeAnnotationHash(
	annotationTypeID string,
	annotatorID string,
	value string,
	sessionID *string,
	entryIndex *int,
	endIndex *int,
	confidence *float64,
	reason *string,
	provenance *Provenance,
) string {
	payload := annotationHashPayload{
		AnnotationTypeID: annotationTypeID,
		AnnotatorID:      annotatorID,
		Value:            value,
		SessionID:        sessionID,
		EntryIndex:       entryIndex,
		EndIndex:         endIndex,
		Confidence:       confidence,
		Reason:           reason,
		Provenance:       provenance,
	}

	b, err := json.Marshal(&payload)
	if err != nil {
		b = []byte("{}")
	}

	h := sha3.New256()
	_, _ = h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
