package schema

import (
	"fmt"
	"regexp"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- TranscriptID ---

// TranscriptID is a validated village-side transcript identifier. The village
// generates a UUID per published transcript (the same value carried in
// PublishResponse.TranscriptID); the pull surface keys off it. Newtype over
// string so the identifier flows typed across the pull pipeline, mirroring the
// SessionID/ProjectHash precedent.
type TranscriptID string

// transcriptIDUUID matches the canonical RFC-4122 lowercase-hex UUID form the
// village assigns to transcripts (the same shape SessionID accepts for its UUID
// branch).
var transcriptIDUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// NewTranscriptID validates and constructs a TranscriptID. The input must be a
// canonical lowercase-hex UUID; anything else is rejected so a pasted URL slug,
// truncated id, or path-bearing string never reaches the filesystem layout
// ({villageHost}/{transcriptId}/...).
func NewTranscriptID(raw string) (TranscriptID, error) {
	if raw == "" {
		return "", fmt.Errorf("invalid transcript ID: empty string")
	}
	if !transcriptIDUUID.MatchString(raw) {
		return "", fmt.Errorf("invalid transcript ID %q: must be a canonical lowercase-hex UUID (e.g. 99d59925-36bc-424c-a789-8be54d9702ba)", raw)
	}
	return TranscriptID(raw), nil
}

func (t TranscriptID) String() string { return string(t) }

// JSONSchema implements jsonschema.Exposer.
func (TranscriptID) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithFormat("uuid")
	s.WithTitle("Transcript ID")
	s.WithDescription("Village-side transcript identifier (canonical lowercase-hex UUID)")
	s.WithPattern(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	s.WithExamples("99d59925-36bc-424c-a789-8be54d9702ba")
	return s, nil
}

// --- Pull surface wire types ---

// PullTranscriptInfo is the village's metadata view of a single pullable
// transcript (GET /api/v1/pull/transcripts/{id}). ContentHash is the
// SERVER-COMPUTED hash of the served blob bytes (empty when the village has not
// yet computed one), and is distinct from UnifiedMetadata.ContentHash (the
// ingest-time, pre-push-redaction hash). ContractVersion records the push
// content contract the blob was published under (the blob carries its own
// publish-time version; the pull envelope does not version the blob).
//
// Owner is the transcript's owner as the caller may see them, resolved by the
// server per viewer. A caller pulling its own transcript always receives its
// own profile, so Owner.IsAccount(callerID) is exactly "this is mine"; an
// anonymous owner is never the caller. This envelope shape is pull contract
// 0.3.0, which replaced the ownerUserId and ownerUsername strings.
type PullTranscriptInfo struct {
	TranscriptID TranscriptID       `json:"transcriptId"`
	LocalID      string             `json:"localId,omitempty"` // peasant SessionID at publish — round-trip correlation
	Owner        VillagePublicOwner `json:"owner"`
	Title        string             `json:"title,omitempty"`
	Harness      Harness            `json:"harness,omitempty"` // existing typed form
	// ProjectName identifies the owner's repository, so it is omitted when
	// Owner is anonymous to the caller, as well as when it was not recorded. It
	// is a pointer so an explicit empty string is distinguishable from an
	// omitted field: withheld means absent, and an empty string beside an
	// anonymous owner is refused like any other value.
	ProjectName     *string         `json:"projectName,omitempty"`
	Visibility      Visibility      `json:"visibility"`                // existing typed form
	License         License         `json:"license,omitempty"`         // legal axis; omitempty ⇒ legacy/un-set
	ContentHash     string          `json:"contentHash,omitempty"`     // SERVED-BLOB hash; empty ⇒ server has none
	ContractVersion ContractVersion `json:"contractVersion,omitempty"` // push contract the blob was published under
	PublishedAt     int64           `json:"publishedAt"`
	UpdatedAt       int64           `json:"updatedAt"`
	AnnotationCount int             `json:"annotationCount"`
}

// Validate checks the owner and that an anonymous owner's project name is
// omitted.
func (p PullTranscriptInfo) Validate() error {
	return validateAnonymousRepository(p.Owner, presentFieldNames(p.repositoryFields()), "schema.PullTranscriptInfo.Validate")
}

func (p PullTranscriptInfo) repositoryFields() map[string]bool {
	return map[string]bool{"projectName": p.ProjectName != nil}
}

// Validate checks every listed transcript.
func (r PullListResponse) Validate() error {
	for _, info := range r.Transcripts {
		if err := info.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// PullListResponse is the village's paginated listing of pullable transcripts
// (GET /api/v1/pull/transcripts) — own + group-shared (public excluded by the
// canPullTranscript policy).
type PullListResponse struct {
	Transcripts []PullTranscriptInfo `json:"transcripts"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	Total       int                  `json:"total"`
}

// PullAnnotation is the authored-annotation row for the pull surface
// (GET /api/v1/pull/transcripts/{id}/annotations). It embeds the existing
// AnnotationSummary and adds the author as the caller may see them, resolved
// by the server per viewer. The caller's own rows always carry the caller's
// profile, so Author.IsAccount(callerID) identifies them for the
// annotation-refresh exclusion; an anonymous author is never the caller and is
// marked foreign. Pull contract 0.3.0 replaced the authorUserId and
// authorUsername strings with Author.
//
// The embedded summary's targetProjectHash and targetFilePath identify the
// owner's repository, so Village leaves both absent when the annotated
// transcript's owner is anonymous to the caller. AnnotationSummary keeps its
// Local API meaning; the rule applies to this Village envelope only. The
// annotations response carries no owner, so no validator here can pair the
// two; Village enforces and tests it on the server.
type PullAnnotation struct {
	AnnotationSummary                    // embedded existing summary (annotator entity, content hash, etc.)
	Author            VillagePublicOwner `json:"author"`
}
