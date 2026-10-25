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

// --- Requestable pull envelope versions ---

// PullContractHeader names the request and response header that selects and
// reports the pull envelope version on the four pull operations: the list,
// the metadata read, the annotations read and the content read.
//
// A client sends the newest pull contract it reads. Village serves the highest
// version it supports that is no newer than the requested one and no older than
// its minimum; with no header it serves its default (the advertised
// pullContractVersion); a requested version below its minimum is the existing
// incompatible-contract refusal. Every pull response carries the header with
// the version actually served, so a client knows exactly which envelope it got.
//
// Compatibility: a patch bump within a minor (0.3.x) is additive, so a client
// that reads 0.3.0 accepts any 0.3.z and ignores fields it does not know. Only
// a new minor needs a new client, and Village keeps serving older minors to
// older clients for a deprecation period.
const PullContractHeader = "Peasant-Pull-Contract"

const (
	// LegacyPullContractVersion is the pull envelope every released CLI before
	// the public owner reads: LegacyPullTranscriptInfo, LegacyPullListResponse
	// and LegacyPullAnnotation. Village serves it by default, with people
	// anonymous to the caller blanked.
	LegacyPullContractVersion ContractVersion = "0.2.0"
	// PublicOwnerPullContractVersion is the envelope that carries people as
	// VillagePublicOwner: PullTranscriptInfo, PullListResponse and
	// PullAnnotation. Village serves it to a client that asks for it.
	PublicOwnerPullContractVersion ContractVersion = "0.3.0"
)

// LegacyPullAnonymousUsername is the username the 0.2.0 envelope carries for
// a person anonymous to the caller, beside an empty id. The empty id plus
// "anon" is the one value every released CLI tolerates: it decodes, it never
// equals the caller's own id, so the transcript or annotation is foreign, and
// it never names the hidden person.
const LegacyPullAnonymousUsername = "anon"

// LegacyPullTranscriptInfo is pull envelope 0.2.0's metadata view of one
// pullable transcript, frozen byte-for-byte from the shape released before the
// public owner (schema v0.32.0). Village serves it to clients that do not ask
// for a newer envelope, and must build it with LegacyPullTranscriptInfoFrom
// rather than its own structs.
//
// Blanking rule: a person anonymous to the caller is OwnerUserID "" with
// OwnerUsername "anon", and ProjectName is omitted beside them. A profile
// carries the real id and username. Validate refuses any other pairing.
type LegacyPullTranscriptInfo struct {
	TranscriptID    TranscriptID    `json:"transcriptId"`
	LocalID         string          `json:"localId,omitempty"`
	OwnerUserID     string          `json:"ownerUserId"`
	OwnerUsername   string          `json:"ownerUsername"`
	Title           string          `json:"title,omitempty"`
	Harness         Harness         `json:"harness,omitempty"`
	ProjectName     string          `json:"projectName,omitempty"`
	Visibility      Visibility      `json:"visibility"`
	License         License         `json:"license,omitempty"`
	ContentHash     string          `json:"contentHash,omitempty"`
	ContractVersion ContractVersion `json:"contractVersion,omitempty"`
	PublishedAt     int64           `json:"publishedAt"`
	UpdatedAt       int64           `json:"updatedAt"`
	AnnotationCount int             `json:"annotationCount"`
}

// LegacyPullListResponse is pull envelope 0.2.0's paginated listing.
type LegacyPullListResponse struct {
	Transcripts []LegacyPullTranscriptInfo `json:"transcripts"`
	Page        int                        `json:"page"`
	Limit       int                        `json:"limit"`
	Total       int                        `json:"total"`
}

// LegacyPullAnnotation is pull envelope 0.2.0's annotation row. An author
// anonymous to the caller is AuthorUserID "" with AuthorUsername "anon". The
// embedded summary's targetProjectHash and targetFilePath are left absent when
// the annotated transcript's owner is anonymous to the caller, exactly as on
// PullAnnotation; the response carries no owner, so Village enforces that.
type LegacyPullAnnotation struct {
	AnnotationSummary
	AuthorUserID   string `json:"authorUserId"`
	AuthorUsername string `json:"authorUsername"`
}

// legacyPullPerson renders a public owner in the 0.2.0 shape.
func legacyPullPerson(person VillagePublicOwner) (id, username string, shown bool) {
	if person.IsProfile() {
		return string(person.ID), person.Username, true
	}
	return "", LegacyPullAnonymousUsername, false
}

// validateLegacyPullPerson refuses a 0.2.0 person that pairs an anonymous
// marker with identity: an empty id must come with "anon", and a non-empty id
// with a non-empty username.
func validateLegacyPullPerson(id, username string, at string) error {
	if id == "" && username != LegacyPullAnonymousUsername {
		return fmt.Errorf("legacy pull envelope validation failed at %s: an empty id carries username %q; a hidden person would be named; send username %q beside an empty id", at, username, LegacyPullAnonymousUsername)
	}
	if id != "" && username == "" {
		return fmt.Errorf("legacy pull envelope validation failed at %s: an id carries no username; released CLIs cannot display the person; send the profile's username, or an empty id with %q", at, LegacyPullAnonymousUsername)
	}
	return nil
}

// LegacyPullTranscriptInfoFrom renders a 0.3.0 transcript in the 0.2.0 shape,
// applying the blanking rule. It is the only way Village should build one.
func LegacyPullTranscriptInfoFrom(info PullTranscriptInfo) LegacyPullTranscriptInfo {
	id, username, shown := legacyPullPerson(info.Owner)
	legacy := LegacyPullTranscriptInfo{
		TranscriptID: info.TranscriptID, LocalID: info.LocalID, OwnerUserID: id, OwnerUsername: username,
		Title: info.Title, Harness: info.Harness, Visibility: info.Visibility, License: info.License,
		ContentHash: info.ContentHash, ContractVersion: info.ContractVersion, PublishedAt: info.PublishedAt,
		UpdatedAt: info.UpdatedAt, AnnotationCount: info.AnnotationCount,
	}
	if shown && info.ProjectName != nil {
		legacy.ProjectName = *info.ProjectName
	}
	return legacy
}

// LegacyPullListResponseFrom renders a 0.3.0 listing in the 0.2.0 shape.
func LegacyPullListResponseFrom(list PullListResponse) LegacyPullListResponse {
	legacy := LegacyPullListResponse{Page: list.Page, Limit: list.Limit, Total: list.Total}
	if list.Transcripts != nil {
		legacy.Transcripts = make([]LegacyPullTranscriptInfo, 0, len(list.Transcripts))
	}
	for _, info := range list.Transcripts {
		legacy.Transcripts = append(legacy.Transcripts, LegacyPullTranscriptInfoFrom(info))
	}
	return legacy
}

// LegacyPullAnnotationFrom renders a 0.3.0 annotation row in the 0.2.0 shape.
func LegacyPullAnnotationFrom(annotation PullAnnotation) LegacyPullAnnotation {
	id, username, _ := legacyPullPerson(annotation.Author)
	return LegacyPullAnnotation{AnnotationSummary: annotation.AnnotationSummary, AuthorUserID: id, AuthorUsername: username}
}

// Validate checks the blanking rule: an empty owner id comes with "anon" and
// no project name, and an owner id comes with a username.
func (p LegacyPullTranscriptInfo) Validate() error {
	if err := validateLegacyPullPerson(p.OwnerUserID, p.OwnerUsername, "schema.LegacyPullTranscriptInfo.Validate"); err != nil {
		return err
	}
	if p.OwnerUserID == "" && p.ProjectName != "" {
		return fmt.Errorf("legacy pull envelope validation failed at schema.LegacyPullTranscriptInfo.Validate: projectName sits beside an anonymous owner; it identifies the owner's repository; omit it for an anonymous owner")
	}
	return nil
}

// Validate checks every listed transcript.
func (r LegacyPullListResponse) Validate() error {
	for _, info := range r.Transcripts {
		if err := info.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks the blanking rule on the author.
func (a LegacyPullAnnotation) Validate() error {
	return validateLegacyPullPerson(a.AuthorUserID, a.AuthorUsername, "schema.LegacyPullAnnotation.Validate")
}
