package schema

// PublishRequest is the canonical wire type for CLI → Village upload.
// The CLI sends this; the village validates and persists it.
//
// Uses nested composites for logical grouping. The JSON tags on each composite
// struct preserve CLI wire compatibility — the top-level keys are identity, model,
// timestamp, source, git, project, stats, quality, subagents, and diagnostics.
//
// MIGRATION NOTE: The CLI's current flat UnifiedMetadata will need a coordinated
// update to send this nested structure. This is NOT backward-compatible with the
// current village handler — both CLI and village must update together.
//
// The Village operation-specific publish component requires model, so a body
// without it is rejected; this changes generated validation requiredness, not
// the canonical Go wire shape.
type PublishRequest struct {
	Identity    SessionIdentity `json:"identity"`
	Model       ModelInfo       `json:"model" required:"true"`
	Timestamp   TimestampInfo   `json:"timestamp"`
	Source      SourceInfo      `json:"source"`
	Git         GitContext      `json:"git"`
	Project     ProjectContext  `json:"project"`
	Stats       SessionStats    `json:"stats"`
	Quality     *QualityMetrics `json:"quality,omitempty"`
	Entries     []SessionEntry  `json:"entries,omitempty"`
	Subagents   []SubagentRef   `json:"subagents,omitempty"`
	Diagnostics DiagnosticsInfo `json:"diagnostics"`
	// License is the content license the contributor selected for this transcript
	// (CC0-1.0 / CC-BY-4.0 / CC-BY-SA-4.0). Optional — omitempty ⇒ a publish with no
	// license stores NULL (legacy/un-set). The village persists it to
	// transcripts.license_id; the vendored schema enum makes an invalid value a
	// documented schema-422. NOT `required` (keeps the Village publish body required set at [model]).
	License License `json:"license,omitempty"`
}

// PublishResponse is returned by the village after successful publish.
type PublishResponse struct {
	TranscriptID  string `json:"transcriptId"`  // Server-generated UUID
	BlobKey       string `json:"blobKey"`       // S3 storage path
	BlobSizeBytes int64  `json:"blobSizeBytes"` // Uploaded file size in bytes
	PublishedAt   int64  `json:"publishedAt"`   // Unix millis
	UpdatedAt     int64  `json:"updatedAt"`     // Unix millis
	Created       bool   `json:"created"`       // true if new, false if updated
}

// AnnotationPushItem is the wire type for a single annotation in a push request.
// ContentHash is SHA3-256 of the canonical JSON representation of all other fields,
// computed by ComputeContentHash(). Used for server-side deduplication.
type AnnotationPushItem struct {
	ContentHash   string                 `json:"contentHash"`
	TargetKind    TargetKind             `json:"targetKind"`
	SessionID     *string                `json:"sessionId,omitempty"`
	EntryTarget   *AnnotationEntryTarget `json:"entryTarget,omitempty"`
	AnnotationID  *string                `json:"annotationId,omitempty"`
	ProjectHash   *ProjectHash           `json:"projectHash,omitempty"`
	TypeID        string                 `json:"typeId"`
	Value         string                 `json:"value"`
	IsPrimary     bool                   `json:"isPrimary"`
	Confidence    *float64               `json:"confidence,omitempty"`
	Reason        *string                `json:"reason,omitempty"`
	AnnotatorName string                 `json:"annotatorName,omitempty"`
	Provenance    *Provenance            `json:"provenance,omitempty"`
}

// AnnotationEntryTarget identifies an entry-level annotation target.
// EndIndex is a half-open range [EntryIndex, EndIndex).
type AnnotationEntryTarget struct {
	SessionID  string `json:"sessionId"`
	EntryIndex int    `json:"entryIndex"`
	EndIndex   int    `json:"endIndex"` // half-open [start, end)
}

// AnnotationPushRequest is the body sent to POST /api/v1/annotations.
//
// Retractions is an ADDITIVE, backwards-compatible field: a set
// of content-hashes the client wants the village to DROP (tombstone) for this
// owner. It carries propagated deletions/supersessions inline on the existing
// push request rather than via a separate DELETE endpoint (avoids a second
// auth/owner-scoping path and round-trip). The hashes are the SAME hex SHA3
// content-hashes used for dedup; only hashes the owner authored AND locally
// retired are sent, so a foreign machine's annotation can never be retracted.
// omitempty preserves the prior wire shape for clients that send no retractions.
type AnnotationPushRequest struct {
	Annotations []AnnotationPushItem `json:"annotations"`
	Retractions []string             `json:"retractions,omitempty"`
}

// AnnotationPushResponse is returned by the village after processing an annotation push.
type AnnotationPushResponse struct {
	Created int                    `json:"created"`
	Updated int                    `json:"updated"`
	Skipped int                    `json:"skipped"`
	Errors  int                    `json:"errors"`
	Results []AnnotationPushResult `json:"results,omitempty"`
}

// AnnotationPushStatus is the per-item outcome status from the village.
type AnnotationPushStatus string

const (
	PushStatusCreated AnnotationPushStatus = "created"
	PushStatusUpdated AnnotationPushStatus = "updated"
	PushStatusSkipped AnnotationPushStatus = "skipped"
	PushStatusError   AnnotationPushStatus = "error"
)

// String returns the string representation of the push status.
func (s AnnotationPushStatus) String() string { return string(s) }

// AnnotationPushResult is the per-item result within an AnnotationPushResponse.
type AnnotationPushResult struct {
	ContentHash string               `json:"contentHash,omitempty"`
	Status      AnnotationPushStatus `json:"status"`
	Error       string               `json:"error,omitempty"`
}

// SchemaVersionResponse is returned by GET /api/v1/schema/version.
// It communicates which annotation schema AND which push CONTENT contract the
// village currently supports, so the CLI can preflight and version-negotiate.
//
// PushContractVersion / MinPushContractVersion advertise the village's accept
// WINDOW [Min, Current] for the TranscriptContent push wire:
//   - PushContractVersion is the CURRENT contract the village emits/prefers.
//   - MinPushContractVersion is the PUSH-ACCEPTANCE FLOOR: the oldest contract
//     the village will still accept on the publish path. A CLI ahead of Current
//     downgrade-emits toward this window (never "upgrade the village").
//
// TWO DISTINCT FLOORS: MinPushContractVersion is the push-acceptance
// floor (gates INCOMING uploads). It is deliberately SEPARATE from the village's
// display MIGRATE-ON-READ floor (how far back stored blobs can be normalized for
// rendering), which may reach FURTHER back than MinPushContractVersion — the
// village can still render a legacy stored blob it would no longer ACCEPT as a
// fresh push. The migrate-on-read floor lives village-side and is not
// advertised here; only the push-acceptance window is negotiated over the wire.
//
// PullContractVersion / MinPullContractVersion advertise the village's PULL
// envelope WINDOW [Min, Current] (PullTranscriptInfo/PullListResponse/
// PullAnnotation shapes, the /api/v1/pull/* endpoint semantics, ETag behaviour).
// This window is DISTINCT from the push window: it does NOT version the blob
// (the stored blob carries its own publish-time push contract version). Both
// fields are `omitempty` so an OLDER village that predates the pull surface
// emits the prior wire shape; the CLI treats an ABSENT advertisement as
// "village too old for pull" (actionable error), not as compatible.
type SchemaVersionResponse struct {
	AnnotationSchemaVersion string              `json:"annotationSchemaVersion"`
	SupportedTargetKinds    []string            `json:"supportedTargetKinds"`
	SupportedTypeIDs        []string            `json:"supportedTypeIds"`
	PushContractVersion     PushContractVersion `json:"pushContractVersion"`
	MinPushContractVersion  PushContractVersion `json:"minPushContractVersion"`
	PullContractVersion     PushContractVersion `json:"pullContractVersion,omitempty"`
	MinPullContractVersion  PushContractVersion `json:"minPullContractVersion,omitempty"`
}
