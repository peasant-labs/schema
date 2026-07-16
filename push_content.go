package schema

import jsonschema "github.com/swaggest/jsonschema-go"

// --- Push contract versioning + structured content envelope ---
//
// These types define the versioned wire contract for `peasant push`'s transcript
// body. Before this contract, peasant uploaded raw provider JSONL bytes as the
// transcript file; the village then had to re-parse provider-native formats. The
// TranscriptContent envelope replaces that with a self-describing, versioned blob
// carrying a normalized SessionDetailPayload — the same representation the local
// dashboard renders (one conversion path; see SessionToDetail).

// ContractVersion is the semantic version of a peasant/village WIRE contract. It
// is a newtype over string so that the version flows through the codebase typed
// rather than stringly. The same value-type carries BOTH axes (push and pull):
// they share one semver shape and one comparison/classification machinery, and
// the field/constant NAME disambiguates the axis (PushContractVersion vs
// PullContractVersion). See the PushContractVersion alias below.
//
// It tracks the push/village wire SEPARATELY from:
//   - MetadataSchemaVersion (the on-disk {sessionId}--metadata.json structure), and
//   - ingest.CurrentIndexVersion (the local indexer evolution),
//
// mirroring the CurrentIndexVersion precedent: a single canonical constant
// (defaults.PublishSchemaVersion) is the source of truth, bumped on a breaking
// change to the push wire format.
type ContractVersion string

// PushContractVersion is a zero-churn alias for ContractVersion kept for the
// push axis. Existing references (SchemaVersionResponse push fields, the
// defaults push constants) keep compiling unchanged; new code may use either
// name (they are the identical type).
type PushContractVersion = ContractVersion

// String returns the bare semver string for serialization boundaries (OpenAPI
// filenames, JSON tags handled by the marshaller, CLI output).
func (v ContractVersion) String() string { return string(v) }

// ContentKind discriminates the concrete payload carried inside a
// TranscriptContent envelope. It exists so the village can dispatch on the kind
// without sniffing the payload, and so future kinds (e.g. raw-jsonl passthrough)
// can be added without changing the envelope shape.
type ContentKind string

// ContentKindSessionDetail marks an envelope whose SessionDetail field carries a
// normalized SessionDetailPayload. It is the only kind peasant currently emits.
const ContentKindSessionDetail ContentKind = "session_detail"

// AllContentKinds is the canonical list of transcript content envelope kinds.
var AllContentKinds = []ContentKind{ContentKindSessionDetail}

// IsValid reports whether k is a defined content envelope kind.
func (k ContentKind) IsValid() bool { return k == ContentKindSessionDetail }

// String returns the bare kind string for serialization/log boundaries.
func (k ContentKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (ContentKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Content Kind",
		"Payload kind carried by a transcript content envelope",
		AllContentKinds,
	), nil
}

// TranscriptContent is the versioned, self-describing wire body that `peasant
// push` uploads in place of raw provider JSONL. The village decodes this
// envelope (migrate-on-read for legacy/older blobs) and renders SessionDetail.
//
// VERSION CONFLICT-WINNER: a TranscriptContent envelope carries TWO
// version markers — the envelope's ContractVersion and the embedded
// SessionDetailPayload.SchemaVersion. They are emitted IN LOCKSTROKE by peasant
// (both equal defaults.PublishSchemaVersion). If a decoder ever observes them
// to DISAGREE (e.g. a hand-edited or partially-migrated blob), the ENVELOPE's
// ContractVersion WINS and is authoritative for migrate-on-read dispatch; the
// embedded SchemaVersion is advisory and exists so a SessionDetailPayload
// extracted and stored on its OWN (outside any envelope) remains self-describing.
// The village ContentMigrator MUST honor this same rule.
type TranscriptContent struct {
	ContractVersion PushContractVersion   `json:"contractVersion"`
	Kind            ContentKind           `json:"kind"`
	SessionDetail   *SessionDetailPayload `json:"sessionDetail,omitempty"`
}
