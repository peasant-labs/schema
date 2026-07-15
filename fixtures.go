package schema

import _ "embed"

// SessionsYAML contains the raw YAML content for mock sessions.
//
//go:embed testdata/sync/sessions.yaml
var SessionsYAML []byte

// QualitySessionsYAML contains the raw YAML content for mock quality sessions.
//
//go:embed testdata/quality/sessions.yaml
var QualitySessionsYAML []byte

// AnnotationsYAML contains the raw YAML content for annotation test fixtures.
//
//go:embed testdata/annotations/annotations.yaml
var AnnotationsYAML []byte

// PullRefsYAML contains the raw YAML content for transcript-reference parse
// fixtures (internal/pull.ParseTranscriptRef).
//
//go:embed testdata/pull/transcript_refs.yaml
var PullRefsYAML []byte

// PullStatusesYAML contains the raw YAML content for PullStatus<->wire mappings.
//
//go:embed testdata/pull/pull_statuses.yaml
var PullStatusesYAML []byte

// PullManifestExampleJSON is the golden one-complete-PullManifest example
// (round-trip shape-pin source).
//
//go:embed testdata/pull/manifest.example.json
var PullManifestExampleJSON []byte

// PublishVerdictsYAML contains the shared publish schema verdict corpus.
//
//go:embed testdata/publish/verdicts.yaml
var PublishVerdictsYAML []byte

// RedactionsYAML contains the GENERATED redaction example corpus
// (testdata/session-detail/redactions.yaml). It is produced by
// `go run ./cmd/schema-gen` from RedactionExamples (see redactions.go) and is
// the single source of truth consumed by peasant's behavioural conformance test
// (pkg/redact) and the web mock codegen. Regenerate + commit on any change to
// RedactionExamples; the leaf freshness gate enforces byte-identity.
//
//go:embed testdata/session-detail/redactions.yaml
var RedactionsYAML []byte

// TimelineYAML contains project timeline validation cases.
//
//go:embed testdata/local-api/timeline.yaml
var TimelineYAML []byte

// TimelineManifestYAML is the exact family/name/classification identity and
// mutation-proof manifest for the project timeline corpus.
//
//go:embed testdata/local-api/timeline_manifest.yaml
var TimelineManifestYAML []byte
