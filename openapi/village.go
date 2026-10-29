package openapi

import (
	"fmt"
	"mime/multipart"
	"net/http"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	openapicore "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// AuthoritativeTranscriptPublishRequest is the Village 0.11 publish
// operation's HTTP body. Its distinct operation-only identity lets the
// successor OpenAPI contract be stricter than legacy shared metadata schemas.
type AuthoritativeTranscriptPublishRequest schema.AuthoritativePublishRequest

// TranscriptPublishRequest retains the previous operation-wrapper name while
// Go consumers migrate to AuthoritativeTranscriptPublishRequest.
// Deprecated: use AuthoritativeTranscriptPublishRequest. See schema issue #55.
type TranscriptPublishRequest = AuthoritativeTranscriptPublishRequest

type villageGroupedView string

func (villageGroupedView) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithEnum("grouped")
	return s, nil
}

type transcriptPublishMultipartRequest struct {
	Metadata       AuthoritativeTranscriptPublishRequest `formData:"metadata" required:"true" description:"PublishRequest JSON encoded with Content-Type application/json."`
	TranscriptFile *multipart.FileHeader                 `formData:"transcript_file" required:"true" description:"Exact transcript bytes whose SHA3-256 digest equals metadata.contentHash."`
}

// TranscriptUpdateErrorResponse is the body the owner update operation returns
// on every refusal. The village serves one uniform error envelope, so each
// declared non-success status carries this same shape and a client reads the
// reason from one field regardless of which refusal it hit.
//
// It lives here, operation-scoped, rather than in the shared type catalog. The
// shape is nothing but {error: string}, so promoting it to the canonical
// cross-language catalog would freeze a transcript-update-specific NAME onto a
// generic envelope at the next release tag, leaving whoever declares the second
// operation's refusals to reuse a misleading name, duplicate it, or take a
// breaking rename. Whether a shared envelope belongs in the catalog is a
// decision for the change that needs one, not a side effect of this one.
type TranscriptUpdateErrorResponse struct {
	// Error is the human-readable, actionable refusal reason. It is required
	// because the village emits it unconditionally: every declared refusal on
	// this operation, including the 401 raised by the authentication middleware
	// before the handler runs, is written by one helper that always sets this
	// field. Declaring it optional would understate what the server guarantees
	// and force a consumer to handle an absence that cannot occur.
	//
	// The tag is load-bearing rather than decorative. Go-tag requiredness is
	// applied to catalogued types by the Types generator; this type is
	// deliberately operation-scoped and outside that catalog, so the tag is the
	// only thing that emits the required array here.
	Error string `json:"error" required:"true"`
}

// pullTranscriptRequest is the request structure of the pull reads of one
// transcript: the transcript id and the optional pull contract header. Every
// such read reuses this one value so the header text exists once.
type pullTranscriptRequest struct {
	ID           string `path:"id"`
	PullContract string `header:"Peasant-Pull-Contract" description:"Optional. The newest pull contract this client reads, as MAJOR.MINOR.PATCH digits (for example 0.3.0). Village serves the highest version it supports that is no newer than this and no older than its minimum. Omitted or malformed (0.3, v0.3.0, any other text), it is treated as absent and Village serves its default (pullContractVersion), so a client with a buggy header still pulls. Below Village's minimum, the read is refused with 406 and a VillageErrorResponse."`
}

// BuildVillageAPISpec builds the current OpenAPI 3.1 specification for the
// Village API. It describes transcript publishing, CLI authentication,
// annotation registry and manifest synchronization, schema negotiation, group-
// scoped transcript discovery and pull operations, collectives, contributions,
// review actions, and linked repositories.
func BuildVillageAPISpec() (*openapi31.Spec, error) {
	r := openapi31.NewReflector()
	registerHarnessSchema(r)
	r.Spec.Info.
		WithTitle("Transcript Publish API").
		// info.version is derived from VillageAPIVersion (the single source of
		// truth) so a doc-surface semver bump is a one-line edit in artifacts.go,
		// never a literal retyped here. See the package doc for the policy.
		WithVersion(VillageAPIVersion).
		WithDescription("Village API for transcript publishing, CLI authentication, annotation registry " +
			"and manifest synchronization, schema negotiation, group-scoped transcript discovery, " +
			"content, annotations, currency checks, collectives, shares, contributions, review, and linked repositories.")

	// POST /api/v1/transcripts/publish — PublishRequest in, PublishResponse out.
	oc, err := r.NewOperationContext(http.MethodPost, "/api/v1/transcripts/publish")
	if err != nil {
		return nil, fmt.Errorf("new publish operation: %w", err)
	}
	oc.AddReqStructure(new(transcriptPublishMultipartRequest))
	oc.AddRespStructure(new(schema.AuthoritativePublishResponse), openapicore.WithHTTPStatus(http.StatusCreated))
	oc.AddRespStructure(new(schema.AuthoritativePublishResponse), openapicore.WithHTTPStatus(http.StatusOK))
	oc.SetDescription("Publish exact transcript bytes with typed JSON metadata. Capability negotiation is deployment-specific and uses exact string membership, never SemVer ordering. A client with required capabilities missing from the destination MUST refuse before remote upload and MUST NOT strip evidence; dry-run negotiation is local only. A server advertising observed_model_v1 MUST validate assistant-only attribution and observedModel values before persistence or any side effect, and MUST preserve accepted observedModel string bytes through storage, migration, rewrite, and serving; JSON formatting is excluded. Assistant includes nested subagent output represented by role assistant. Creation returns 201 and replacement returns 200; both carry the complete authoritative receipt.")
	oc.SetID("publishTranscript")
	oc.SetTags("transcripts")
	if err := r.AddOperation(oc); err != nil {
		return nil, fmt.Errorf("add publish operation: %w", err)
	}
	if err := registerPublishMetadataComponent(r); err != nil {
		return nil, err
	}
	if err := setMultipartMetadataEncoding(r.Spec, "/api/v1/transcripts/publish"); err != nil {
		return nil, err
	}

	// PATCH /api/v1/transcripts/{id} — owner-only partial metadata/governance
	// update of an already-published transcript. The village has served this
	// since the governance work landed; declaring it here closes the drift where
	// a handler enforced rules the published contract never stated.
	//
	// Every refusal is declared, not just the happy path, because two of them
	// are contract rules rather than transport accidents: the ownership boundary
	// (403, leaving state and the governance audit untouched) and the
	// irrevocability of a granted license (400). A client reading only a success
	// shape would not learn either.
	updateOC, err := r.NewOperationContext(http.MethodPatch, "/api/v1/transcripts/{id}")
	if err != nil {
		return nil, fmt.Errorf("new transcript update operation: %w", err)
	}
	// The path parameter is the canonical TranscriptID, not a bare string. The
	// village parses it with uuid.Parse and refuses anything else with a 400, so
	// declaring an unconstrained string described a request the server always
	// rejects. TranscriptID also carries the lowercase-hex form this module
	// already treats as canonical for transcript identifiers (see the pattern on
	// TranscriptID itself and the matching SessionID UUID branch), so this is the
	// module's existing position rather than a restriction invented here.
	//
	// This narrowing is deliberate and was ratified after measurement, so do not
	// widen it on the observation alone. uuid.Parse ACCEPTS four further
	// spellings the declared pattern rejects: uppercase hex, brace-wrapped,
	// urn:uuid-prefixed, and 32 raw hex digits with no dashes; village acts on
	// all of them. They stay undeclared because village emits identifiers via
	// uuid.UUID.String(), which is always canonical lowercase, so no client can
	// hold another form unless it manufactures one; because this module already
	// rejects the other forms at its own boundary in NewTranscriptID; and because
	// accepting five spellings for one identity is itself a defect surface.
	// Declaring one canonical form is better contract design, not merely
	// narrower.
	updateOC.AddReqStructure(new(struct {
		ID schema.TranscriptID `path:"id" description:"Transcript identifier"`
	}))
	updateOC.AddReqStructure(new(schema.OwnerTranscriptUpdateRequest))
	// The successor contract returns the complete authoritative editable state.
	// Village must implement this projection before re-pinning the schema module.
	updateOC.AddRespStructure(new(schema.OwnerTranscriptUpdateResponse), openapicore.WithHTTPStatus(http.StatusOK))
	for _, status := range []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	} {
		updateOC.AddRespStructure(new(TranscriptUpdateErrorResponse), openapicore.WithHTTPStatus(status))
	}
	updateOC.SetDescription("Update an owned transcript's metadata and governance axes. Every field is optional: omission preserves stored state; empty title, description, or tags clear those fields; license null requests clear; and a canonical license replaces. Explicit null for title, description, tags, or visibility is rejected, as are unknown fields and invalid or duplicate tags. Clearing an already granted Creative Commons license remains subject to the server's irrevocability rule. Only the owner may call this operation. A successful update returns the complete typed authoritative editable state, including the canonical transcript URL and positive update timestamp. Visibility accepts private and public; organization-scoped visibility remains deferred.")
	updateOC.SetID("updateTranscript")
	updateOC.SetTags("transcripts")
	if err := r.AddOperation(updateOC); err != nil {
		return nil, fmt.Errorf("add transcript update operation: %w", err)
	}
	// Reflection marks a request body optional by default. Here that would be a
	// false statement: the handler decodes the body unconditionally, so omitting
	// it (or sending an empty one) is a guaranteed 400. An empty JSON object is
	// the correct way to send a no-op, and it is accepted. Mark the body required
	// so the contract describes a request that can actually succeed.
	if err := requirePatchRequestBody(r.Spec, "/api/v1/transcripts/{id}"); err != nil {
		return nil, err
	}

	// GET /api/v1/auth/cli/login — browser OAuth initiation.
	// The CLI opens this URL in the browser; the village handles the OAuth flow
	// and redirects back to the CLI's local callback server.
	loginOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/auth/cli/login")
	if err != nil {
		return nil, fmt.Errorf("new login operation: %w", err)
	}
	loginOC.AddReqStructure(new(schema.CLILoginQuery))
	loginOC.SetDescription("Initiate CLI login via browser OAuth flow. " +
		"The CLI opens this URL in the user's browser with a local callback port and CSRF state. " +
		"After authentication, the village redirects the browser to the CLI's local callback server.")
	loginOC.SetID("cliLogin")
	loginOC.SetTags("auth")
	if err := r.AddOperation(loginOC); err != nil {
		return nil, fmt.Errorf("add login operation: %w", err)
	}

	// POST /api/v1/auth/cli/exchange — exchange OAuth code for credentials.
	exchangeOC, err := r.NewOperationContext(http.MethodPost, "/api/v1/auth/cli/exchange")
	if err != nil {
		return nil, fmt.Errorf("new exchange operation: %w", err)
	}
	exchangeOC.AddReqStructure(new(schema.ExchangeCodeRequest))
	exchangeOC.AddRespStructure(new(schema.ExchangeCodeResponse))
	exchangeOC.SetDescription("Exchange an OAuth authorization code and state for API credentials. " +
		"Called by the CLI's local callback server after receiving the browser redirect.")
	exchangeOC.SetID("cliExchangeCode")
	exchangeOC.SetTags("auth")
	if err := r.AddOperation(exchangeOC); err != nil {
		return nil, fmt.Errorf("add exchange operation: %w", err)
	}

	// POST /api/v1/annotations accepts owner-authenticated annotation pushes.
	// The handler validates the shared discriminated target contract before
	// persistence: targetKind association requires only targetAssociationId, and
	// every other target kind rejects that field. Association ownership is
	// resolved by the consumer with the authenticated owner plus the opaque ID.
	annotationPushOC, err := r.NewOperationContext(http.MethodPost, "/api/v1/annotations")
	if err != nil {
		return nil, fmt.Errorf("new annotation push operation: %w", err)
	}
	annotationPushOC.AddReqStructure(new(schema.AnnotationPushRequest))
	annotationPushOC.AddRespStructure(new(schema.AnnotationPushResponse), openapicore.WithHTTPStatus(http.StatusOK))
	annotationPushOC.SetDescription("Push annotations for the authenticated owner. Every item selects exactly one target arm: " +
		"targetKind association requires only targetAssociationId, while every other target kind rejects targetAssociationId. " +
		"Entry targets require a non-empty sessionId; endIndex must be greater than entryIndex and is enforced by the request validation boundary. " +
		"The server resolves association targets owner-scoped before writing the all-or-nothing batch.")
	annotationPushOC.SetID("pushAnnotations")
	annotationPushOC.SetTags("annotations")
	if err := r.AddOperation(annotationPushOC); err != nil {
		return nil, fmt.Errorf("add annotation push operation: %w", err)
	}

	// GET /api/v1/annotations/manifest — server-authoritative annotation skip-gate
	// (GH #69). Returns the set of annotation content-hashes the village holds
	// for the authenticated owner, plus a deterministic order-independent digest
	// for a no-op short-circuit. Emitted here so the vendored village-api spec
	// carries the route + AnnotationManifestResponse schema (a re-vendor must not
	// clobber it).
	manifestOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/annotations/manifest")
	if err != nil {
		return nil, fmt.Errorf("new annotation manifest operation: %w", err)
	}
	manifestOC.AddRespStructure(new(schema.AnnotationManifestResponse))
	manifestOC.SetDescription("Return the set of annotation content-hashes the village holds for the " +
		"authenticated owner (the server-authoritative skip-gate manifest), plus a deterministic " +
		"digest over the sorted hash set for a no-op short-circuit.")
	manifestOC.SetID("getAnnotationManifest")
	manifestOC.SetTags("annotations")
	if err := r.AddOperation(manifestOC); err != nil {
		return nil, fmt.Errorf("add annotation manifest operation: %w", err)
	}

	// GET /api/v1/schema/version — version-negotiation handshake. Advertises the
	// annotation schema version, the supported target kinds/type IDs, AND BOTH the
	// push and pull contract WINDOWS ([Min, Current]). The CLI preflights this
	// before push and before pull; an ABSENT pull window (older village) is treated
	// as "village too old for pull" (actionable error), not as compatible.
	versionOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/schema/version")
	if err != nil {
		return nil, fmt.Errorf("new schema version operation: %w", err)
	}
	versionOC.AddRespStructure(new(schema.SchemaVersionResponse))
	versionOC.SetDescription("Advertise the annotation schema version plus push and pull contract windows and the deployment-specific content-capability set. contentCapabilities omitted or empty means none and null is invalid. Its items are opaque, forward-open revision strings: clients match exactly without SemVer or suffix parsing, ignore unknown tokens, and tolerate/deduplicate duplicates; server output contains only pinned known tokens, rejects duplicates, has no semantic order, and is serialized lexicographically. observed_model_v1 is required exactly when any root or nested assistant turn carries observedModel, not for session model alone. It guarantees assistant-only value validation before persistence with no invalid database or blob side effects and byte-exact accepted observedModel strings through storage, typed migration, rewrite, serving, and pull; JSON whitespace and key order are not guaranteed.")
	versionOC.SetID("getSchemaVersion")
	versionOC.SetTags("schema")
	if err := r.AddOperation(versionOC); err != nil {
		return nil, fmt.Errorf("add schema version operation: %w", err)
	}

	// --- Pull surface: AuthRequired; thin wrappers gated by
	// canPullTranscript (404-not-403). Registers the pull envelope wire types so
	// the village and the relocated client share one contract.

	// GET /api/v1/pull/transcripts — list pullable transcripts (own + group-shared;
	// public excluded). Offset pagination via page/limit query params, plus an
	// optional collectives filter that narrows the list to the approved shares of
	// the listed collectives the caller belongs to.
	pullListOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts")
	if err != nil {
		return nil, fmt.Errorf("new pull list operation: %w", err)
	}
	pullListOC.AddReqStructure(new(struct {
		PullContract string `header:"Peasant-Pull-Contract" description:"Optional. The newest pull contract this client reads, as MAJOR.MINOR.PATCH digits (for example 0.3.0). Village serves the highest version it supports that is no newer than this and no older than its minimum. Omitted or malformed (0.3, v0.3.0, any other text), it is treated as absent and Village serves its default (pullContractVersion), so a client with a buggy header still pulls. Below Village's minimum, the read is refused with 406 and a VillageErrorResponse."`
		Page         int    `query:"page"`
		Limit        int    `query:"limit"`
		Collectives  string `query:"collectives" description:"Optional comma-separated collective identifiers. Narrows the list to transcripts with an approved share into any listed collective the caller belongs to. Omission, or only empty elements, lists every pullable transcript."`
	}))
	pullListOC.AddRespStructure(new(schema.PullListResponse))
	pullListOC.AddRespStructure(new(schema.LegacyPullListResponse))
	pullListOC.SetDescription("List transcripts the authenticated user may pull (own + group-shared; " +
		"public excluded by the canPullTranscript policy). Offset pagination. The optional " +
		"collectives filter narrows the list to the approved shares of the listed collectives the " +
		"caller belongs to; a collective the caller is not in contributes nothing.")
	pullListOC.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(http.StatusNotAcceptable))
	pullListOC.SetID("listPullableTranscripts")
	pullListOC.SetTags("pull")
	if err := r.AddOperation(pullListOC); err != nil {
		return nil, fmt.Errorf("add pull list operation: %w", err)
	}

	// GET /api/v1/pull/transcripts/{id} — metadata + owner + visibility + content_hash.
	pullMetaOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts/{id}")
	if err != nil {
		return nil, fmt.Errorf("new pull metadata operation: %w", err)
	}
	pullMetaOC.AddReqStructure(new(struct {
		ID           string `path:"id"`
		PullContract string `header:"Peasant-Pull-Contract" description:"Optional. The newest pull contract this client reads, as MAJOR.MINOR.PATCH digits (for example 0.3.0). Village serves the highest version it supports that is no newer than this and no older than its minimum. Omitted or malformed (0.3, v0.3.0, any other text), it is treated as absent and Village serves its default (pullContractVersion), so a client with a buggy header still pulls. Below Village's minimum, the read is refused with 406 and a VillageErrorResponse."`
	}))
	pullMetaOC.AddRespStructure(new(schema.PullTranscriptInfo))
	pullMetaOC.AddRespStructure(new(schema.LegacyPullTranscriptInfo))
	pullMetaOC.SetDescription("Get a pullable transcript's metadata (the owner as the caller may see them, " +
		"visibility, server-computed content_hash, push contract version). The server resolves the owner per " +
		"viewer; a caller always sees its own profile on its own transcripts. 404 (not 403) when not pullable.")
	pullMetaOC.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(http.StatusNotAcceptable))
	pullMetaOC.SetID("getPullTranscript")
	pullMetaOC.SetTags("pull")
	if err := r.AddOperation(pullMetaOC); err != nil {
		return nil, fmt.Errorf("add pull metadata operation: %w", err)
	}

	// GET /api/v1/pull/transcripts/{id}/content — streams the blob; ETag is the
	// server-computed content_hash; If-None-Match ⇒ 304. (Binary stream body; not a
	// typed response struct.)
	pullContentOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts/{id}/content")
	if err != nil {
		return nil, fmt.Errorf("new pull content operation: %w", err)
	}
	pullContentOC.AddReqStructure(new(pullTranscriptRequest))
	pullContentOC.SetDescription("Stream the transcript blob as served. Sends ETag: \"<content_hash>\"; " +
		"a matching If-None-Match request yields 304 Not Modified.")
	pullContentOC.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(http.StatusNotAcceptable))
	pullContentOC.SetID("getPullTranscriptContent")
	pullContentOC.SetTags("pull")
	if err := r.AddOperation(pullContentOC); err != nil {
		return nil, fmt.Errorf("add pull content operation: %w", err)
	}

	// GET /api/v1/pull/transcripts/{id}/publish-metadata: the publish metadata
	// village stored for the transcript, for its owner and the owners of the
	// collectives it was submitted to.
	pullPublishMetadataOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts/{id}/publish-metadata")
	if err != nil {
		return nil, fmt.Errorf("new pull publish metadata operation: %w", err)
	}
	pullPublishMetadataOC.AddReqStructure(new(pullTranscriptRequest))
	pullPublishMetadataOC.AddRespStructure(new(schema.AuthoritativePublishRequest))
	pullPublishMetadataOC.SetDescription("Get the publish metadata village stored for the transcript: the validated publish request, re-encoded, not a reconstruction. " +
		"Only the transcript's owner, or an owner of a collective the transcript was submitted to (pending or approved), may read it. " +
		"Every other caller gets 404, including one who can read the transcript, so the route never reveals more than the transcript does. " +
		"It also answers 404 when the transcript's owner is anonymous to the caller (the public-owner rule decides), so it never exposes a hidden owner's repository, paths or commit authors. " +
		"404 also means village holds no stored metadata for the transcript (transcripts published before metadata was stored). " +
		"406 means the requested pull contract is below village's minimum.")
	for _, status := range []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusNotAcceptable, http.StatusInternalServerError} {
		pullPublishMetadataOC.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(status))
	}
	pullPublishMetadataOC.SetID("getPullTranscriptPublishMetadata")
	pullPublishMetadataOC.SetTags("pull")
	if err := r.AddOperation(pullPublishMetadataOC); err != nil {
		return nil, fmt.Errorf("add pull publish metadata operation: %w", err)
	}

	// GET /api/v1/pull/transcripts/{id}/annotations — PullAnnotation rows with
	// the author as the caller may see them.
	pullAnnotOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts/{id}/annotations")
	if err != nil {
		return nil, fmt.Errorf("new pull annotations operation: %w", err)
	}
	pullAnnotOC.AddReqStructure(new(struct {
		ID           string `path:"id"`
		PullContract string `header:"Peasant-Pull-Contract" description:"Optional. The newest pull contract this client reads, as MAJOR.MINOR.PATCH digits (for example 0.3.0). Village serves the highest version it supports that is no newer than this and no older than its minimum. Omitted or malformed (0.3, v0.3.0, any other text), it is treated as absent and Village serves its default (pullContractVersion), so a client with a buggy header still pulls. Below Village's minimum, the read is refused with 406 and a VillageErrorResponse."`
	}))
	pullAnnotOC.AddRespStructure(new([]schema.PullAnnotation))
	pullAnnotOC.AddRespStructure(new([]schema.LegacyPullAnnotation))
	pullAnnotOC.SetDescription("List a transcript's annotations as PullAnnotation rows (AnnotationSummary " +
		"plus the author as the caller may see them, resolved by the server per viewer). The caller's own rows " +
		"carry the caller's profile, so own-authored rows are excluded by comparing the author's profile id " +
		"with the caller's account; an anonymous author is never the caller and is foreign-marked.")
	pullAnnotOC.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(http.StatusNotAcceptable))
	pullAnnotOC.SetID("getPullTranscriptAnnotations")
	pullAnnotOC.SetTags("pull")
	if err := r.AddOperation(pullAnnotOC); err != nil {
		return nil, fmt.Errorf("add pull annotations operation: %w", err)
	}

	if err := configurePullContractResponses(r.Spec); err != nil {
		return nil, err
	}

	// POST /api/v1/pull/transcripts/skip-gate. Batch currency check: the client
	// sends, per transcript it holds, the id + the content-hash it holds + its own
	// annotation-hash set, and receives per PULLABLE id {contentCurrent,
	// annotationsCurrent}, so it can skip re-pulling unchanged transcripts. A
	// non-pullable id is WITHHELD by omission from results (404-not-403 spirit) so
	// the batch cannot become a currency oracle over ids the caller cannot pull.
	skipGateOC, err := r.NewOperationContext(http.MethodPost, "/api/v1/pull/transcripts/skip-gate")
	if err != nil {
		return nil, fmt.Errorf("new pull skip-gate operation: %w", err)
	}
	skipGateOC.AddReqStructure(new(schema.PullSkipGateRequest))
	skipGateOC.AddRespStructure(new(schema.PullSkipGateResponse))
	skipGateOC.SetDescription("Batch currency check for a pulling client: per transcript it holds, send the " +
		"id, the content-hash it holds, and its own annotation-hash set; receive per pullable id " +
		"{contentCurrent, annotationsCurrent}. Non-pullable ids are withheld by omission from results.")
	skipGateOC.SetID("pullSkipGate")
	skipGateOC.SetTags("pull")
	if err := r.AddOperation(skipGateOC); err != nil {
		return nil, fmt.Errorf("add pull skip-gate operation: %w", err)
	}

	if err := addVillageCollectiveOperations(r); err != nil {
		return nil, err
	}
	if err := addVillagePullRequestOperations(r); err != nil {
		return nil, err
	}
	if err := addVillageInvitationOperations(r); err != nil {
		return nil, err
	}

	// The reflector automatically registers component schemas for all types referenced
	// in PublishRequest, including SessionEntry, ToolCallKind, StopReason,
	// Provider, Role, EntryType, and all composite types.

	// Explicitly register content-layer types not referenced by PublishRequest but
	// part of the publish API domain. Visibility is no longer merely anticipated:
	// the owner update operation above declares a real visibility surface, using
	// its own narrowed menu rather than this general enum.
	if err := addComponentSchema(r, "Visibility", new(schema.Visibility)); err != nil {
		return nil, err
	}

	AddVillageExamples(r.Spec)
	if err := harmonizeSharedTypeComponents(r.Spec); err != nil {
		return nil, fmt.Errorf("harmonize Village API shared components: %w", err)
	}
	if err := addVillageCollectiveAliases(r.Spec); err != nil {
		return nil, err
	}
	if err := setTranscriptContentHashHeader(r.Spec); err != nil {
		return nil, err
	}

	return r.Spec, nil
}

type villageOperationSpec struct {
	method               string
	path                 string
	id                   string
	tag                  string
	description          string
	requests             []interface{}
	response             interface{}
	responseAlternatives []interface{}
	successStatus        int
	errorStatuses        []int
}

func addVillageCollectiveOperations(r *openapi31.Reflector) error {
	groupPath := new(struct {
		ID schema.VillageUUID `path:"id" description:"Collective identifier"`
	})
	groupListQuery := new(struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	})
	groupSearchQuery := new(struct {
		Q     string `query:"q"`
		Limit int    `query:"limit"`
	})
	memberPath := new(struct {
		ID     schema.VillageUUID `path:"id" description:"Collective identifier"`
		UserID schema.VillageUUID `path:"userID" description:"Member user identifier"`
	})
	removeMemberRequest := new(struct {
		ID      schema.VillageUUID `path:"id" description:"Collective identifier"`
		UserID  schema.VillageUUID `path:"userID" description:"Member user identifier"`
		Retract bool               `query:"retract"`
	})
	groupTranscriptPath := new(struct {
		ID           schema.VillageUUID  `path:"id" description:"Collective identifier"`
		TranscriptID schema.TranscriptID `path:"transcriptID" description:"Transcript identifier"`
	})
	transcriptPath := new(struct {
		ID schema.TranscriptID `path:"id" description:"Transcript identifier"`
	})
	transcriptListQuery := new(struct {
		View     villageGroupedView   `query:"view" description:"Set to grouped to request helper-group list items; omission preserves the flat response."`
		Page     int                  `query:"page"`
		Limit    int                  `query:"limit"`
		Query    string               `query:"q" description:"Search text. Matches repository identity only on rows whose repository this viewer may see; the withheld repository fields of an anonymous owner's transcript are never searched."`
		Provider string               `query:"provider"`
		Owner    string               `query:"owner" description:"Username filter. Matches only owners this viewer may see as a profile; an owner shown as anonymous never matches."`
		Project  string               `query:"project" description:"Project filter. Matches only rows whose repository this viewer may see; a transcript whose owner is anonymous to this viewer never matches."`
		Repo     string               `query:"repo" description:"Repository filter. Matches only rows whose repository this viewer may see; a transcript whose owner is anonymous to this viewer never matches."`
		Org      string               `query:"org" description:"Organization filter. Matches only owners this viewer may see as a profile; an owner shown as anonymous never matches."`
		Tags     string               `query:"tags"`
		Origin   schema.SessionOrigin `query:"origin"`
		Sort     string               `query:"sort"`
	})
	groupedViewQuery := new(struct {
		View villageGroupedView `query:"view" description:"Set to grouped to request helper-group list items; omission preserves the legacy response."`
	})
	memberQuery := new(struct {
		GroupID string `path:"groupId" description:"Stable helper group identifier"`
		Scope   string `query:"scope" required:"true" description:"Opaque scope returned by the originating grouped list"`
		Page    int    `query:"page"`
		Limit   int    `query:"limit"`
	})
	transcriptGroupPath := new(struct {
		ID      schema.TranscriptID `path:"id" description:"Transcript identifier"`
		GroupID schema.VillageUUID  `path:"groupID" description:"Collective identifier"`
	})
	ownerCollectivePath := new(struct {
		GroupID schema.VillageUUID `path:"groupId" description:"Collective identifier"`
	})
	shareEventsPath := new(struct {
		GroupID      schema.VillageUUID  `path:"groupId" description:"Collective identifier"`
		TranscriptID schema.TranscriptID `path:"transcriptId" description:"Transcript identifier"`
	})
	repositoryPath := new(struct {
		ID    schema.VillageUUID `path:"id" description:"Collective identifier"`
		Owner string             `path:"owner" description:"Repository owner login"`
		Name  string             `path:"name" description:"Repository name"`
	})
	repositoryCommitsRequest := new(struct {
		ID      schema.VillageUUID `path:"id" description:"Collective identifier"`
		Owner   string             `path:"owner" description:"Repository owner login"`
		Name    string             `path:"name" description:"Repository name"`
		Refresh bool               `query:"refresh"`
	})

	operations := []villageOperationSpec{
		{method: http.MethodGet, path: "/api/v1/auth/me", id: "getAuthenticatedVillageUser", tag: "auth", description: "Read the stable authenticated user UUID before account-scoped metadata fetching. Usernames and credentials are not scope identity.", response: new(schema.VillageUser), errorStatuses: []int{401, 500}},
		{method: http.MethodGet, path: "/api/v1/auth/orgs", id: "listAuthenticatedVillageOrganizations", tag: "organizations", description: "Read the authenticated user organizations for scoped metadata discovery. Organization membership alone does not disclose linked collectives; use the viewer-authorized organization collective projection. Query failure is 500, never a successful empty set.", response: new([]schema.VillageMetadataUserOrganization), errorStatuses: []int{401, 500}},
		{method: http.MethodGet, path: "/api/v1/orgs/{login}", id: "getVillageOrganizationCollectives", tag: "organizations", description: "Read the existing organization response's viewer-permitted collectives projection. Additional organization statistics and members may be present but metadata consumers retain only the disclosed collective projection. Query errors, including linked-collective lookup failure, are 500 rather than fabricated empty success; unavailable organization is privacy-safe 404. Server disclosure remains authoritative.", requests: []interface{}{new(struct {
			Login string `path:"login" minLength:"1"`
		})}, response: new(schema.VillageOrganizationCollectivesResponse), errorStatuses: []int{400, 404, 500}},
		{
			method: http.MethodGet, path: "/api/v1/transcripts", id: "listTranscripts", tag: "transcripts",
			description: "List authorized transcripts using the existing global, recent, explore, profile, project, and search filters. view=grouped returns grouped list items; omitting view preserves the flat transcript array. The owner and org filters match only transcripts whose owner this viewer may see as a profile: a person shown to this viewer as anonymous never matches a username or an organization filter, so a filter cannot reveal who an anonymous owner is.",
			requests:    []interface{}{transcriptListQuery}, responseAlternatives: []interface{}{new(schema.VillageTranscriptListResponse), new(schema.VillageSessionListPayload)},
			errorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError},
		},
		{
			method: http.MethodGet, path: "/api/v1/transcripts/{id}", id: "getTranscriptMetadata", tag: "transcripts",
			description: "Get transcript metadata and viewer-authorized relationship navigation. Navigation is read-only and is never stored in transcript content. Nonowner collective associations use the same narrow viewer-visible/contributor-discoverable projection as the dedicated collective read; omitted associations do not disclose hidden existence. Query failures are 500, never successful empty arrays. integrity is present only when the transcript is flagged, for any reader allowed to read the transcript, whatever its visibility; its presence is the flag. Its flagged_turns are zero-based positions in the top-level sessionDetail.turns array (never TurnDetail.index values, never earlierHistory or child sessions) of the content whose bare hash is transcript.content_hash. A reader highlights turns only when the content read's Village-Content-Hash header equals transcript.content_hash. An empty list means the flag stands with no turn flagged in the current copy. A clean, pending or failed check adds nothing.",
			requests:    []interface{}{transcriptPath}, response: new(schema.VillageTranscriptMetadataResponse),
			errorStatuses: []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
		},
		{
			method: http.MethodGet, path: "/api/v1/transcripts/{id}/content", id: "getTranscriptContent", tag: "transcripts",
			description: "Get the durable transcript content envelope. This response never includes authorized relationship navigation. When the transcript's owner is anonymous to this viewer, sessionDetail is served without its repository identity: project, workingDirectory, gitBranch, gitRemote and every childSessions[].project are absent. Text inside the conversation is content and is not covered.",
			requests:    []interface{}{transcriptPath}, response: new(schema.TranscriptContent),
			errorStatuses: []int{http.StatusBadRequest, http.StatusNotFound},
		},
		{
			method: http.MethodGet, path: "/api/v1/transcript-groups/{groupId}/members", id: "listTranscriptGroupMembers", tag: "transcripts",
			description: "Page the authorized saved helper threads selected by an originating grouped list. The required opaque scope is replayed with current authorization and cannot widen the original route filters.",
			requests:    []interface{}{memberQuery}, response: new(schema.VillageHelperMembersPayload),
			errorStatuses: []int{http.StatusBadRequest, http.StatusForbidden, http.StatusConflict},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/public",
			id:          "listPublicGroups",
			tag:         "collectives",
			description: "List collectives for the public directory as compact rows with member and approved-transcript counts.",
			response:    new([]schema.VillagePublicGroup),
			errorStatuses: []int{
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/search",
			id:          "searchCollectives",
			tag:         "collectives",
			description: "Search visible collectives. An empty q returns an empty collectives array; limit defaults to 10 and is capped at 50.",
			requests:    []interface{}{groupSearchQuery},
			response:    new(schema.VillageCollectiveSearchResponse),
			errorStatuses: []int{
				http.StatusInternalServerError,
			},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/groups",
			id:            "createGroup",
			tag:           "collectives",
			description:   "Create a collective. Name is required; acceptance_mode defaults to open and data_access defaults to members_only when omitted. linked_github_org may be omitted or null for no organization, as on update.",
			requests:      []interface{}{new(schema.VillageCreateGroupRequest)},
			response:      new(schema.VillageGroup),
			successStatus: http.StatusCreated,
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups",
			id:          "listGroups",
			tag:         "collectives",
			description: "List collectives the authenticated caller belongs to, with the caller's role and membership time.",
			response:    new([]schema.VillageUserGroup),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/visible",
			id:          "listVisibleGroups",
			tag:         "collectives",
			description: "List collectives the authenticated caller may see. Role and member_since are null when the caller sees the collective through public or open visibility but is not a member.",
			response:    new([]schema.VillageVisibleGroup),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:               http.MethodGet,
			path:                 "/api/v1/groups/{id}",
			id:                   "getGroup",
			tag:                  "collectives",
			description:          "Get a collective with roster, stats, model breakdown, contributors, the caller's role, and a transcript page. pending_members is present only for owners. transcripts is an empty array when the caller may see the collective but may not read its data.",
			requests:             []interface{}{groupPath, groupListQuery, groupedViewQuery},
			response:             new(schema.VillageGroupDetailResponse),
			responseAlternatives: []interface{}{new(schema.VillageGroupDetailResponse), new(schema.VillageGroupedGroupDetailResponse)},
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusNotFound,
			},
		},
		{
			method:      http.MethodPatch,
			path:        "/api/v1/groups/{id}",
			id:          "updateGroup",
			tag:         "collectives",
			description: "Update an owned collective. Name and description are written from the request body's string values, so a caller that wants to preserve them must send their current values. linked_github_org omitted means preserve, empty string clears, and non-empty sets after visible-org validation.",
			requests:    []interface{}{groupPath, new(schema.VillageUpdateGroupRequest)},
			response:    new(schema.VillageGroup),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/groups/{id}",
			id:          "deleteGroup",
			tag:         "collectives",
			description: "Delete an owned collective and return a status receipt.",
			requests:    []interface{}{groupPath},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/groups/{id}/join",
			id:          "joinGroup",
			tag:         "collectives",
			description: "Join an open or verified collective as a contributor. Curated collectives require invitation and refuse this route.",
			requests:    []interface{}{groupPath},
			response:    new(schema.VillageGroupStatusRoleResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusConflict,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPatch,
			path:        "/api/v1/groups/{id}/members/{userID}/role",
			id:          "updateGroupMemberRole",
			tag:         "collectives",
			description: "Change a collective member's role to contributor or member. Owner and pending are not accepted by this operation.",
			requests:    []interface{}{memberPath, new(schema.VillageGroupMemberRoleRequest)},
			response:    new(schema.VillageGroupStatusRoleResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/groups/{id}/members/{userID}",
			id:          "removeGroupMember",
			tag:         "collectives",
			description: "Remove a member from a collective. The member may remove themselves; removing somebody else requires owner access. retract=true or a mandatory deletion policy retracts that member's live submissions.",
			requests:    []interface{}{removeMemberRequest},
			response:    new(schema.VillageRemoveGroupMemberResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:               http.MethodGet,
			path:                 "/api/v1/groups/{id}/contributable",
			id:                   "listContributableTranscripts",
			tag:                  "contributions",
			description:          "List every transcript the authenticated caller may offer to one collective, plus whether each transcript is already live in that collective. The response is deliberately unpaginated and bounded by the server row limit.",
			requests:             []interface{}{groupPath, groupedViewQuery},
			response:             new(schema.VillageContributableResponse),
			responseAlternatives: []interface{}{new(schema.VillageContributableResponse), new(schema.VillageGroupedContributableResponse), new(schema.VillageDeprecatedGroupedContributableResponse)},
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusRequestEntityTooLarge,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/groups/{id}/shares",
			id:          "batchShareProject",
			tag:         "contributions",
			description: "Offer selected transcripts from one owned project to one collective in one transaction. Omitted or empty transcript_ids means every owned transcript in the project. visibility_confirmed must be true when any selected private transcript would become visible to the collective.",
			requests:    []interface{}{groupPath, new(schema.VillageBatchShareRequest)},
			response:    new(schema.VillageBatchShareResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusConflict,
				http.StatusUnprocessableEntity,
				http.StatusInternalServerError,
			},
		},
		{
			method:               http.MethodGet,
			path:                 "/api/v1/groups/{id}/pending",
			id:                   "listPendingShares",
			tag:                  "contributions",
			description:          "List pending transcript submissions for an owned collective, oldest first.",
			requests:             []interface{}{groupPath, groupedViewQuery},
			responseAlternatives: []interface{}{new([]schema.VillagePendingShare), new(schema.VillageSessionListPayload)},
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:               http.MethodGet,
			path:                 "/api/v1/groups/{id}/my-shares",
			id:                   "listMyGroupShares",
			tag:                  "contributions",
			description:          "List the authenticated caller's own live submissions to one collective, including pending and approved rows.",
			requests:             []interface{}{groupPath, groupedViewQuery},
			responseAlternatives: []interface{}{new([]schema.VillageUserGroupShare), new(schema.VillageSessionListPayload)},
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPatch,
			path:        "/api/v1/groups/{id}/shares",
			id:          "batchReviewShares",
			tag:         "contributions",
			description: "Approve or reject many pending submissions in one owner-only action. The response separates ids that were decided by this action from ids that were already decided, never submitted to this collective, or stale.",
			requests:    []interface{}{groupPath, new(schema.VillageBatchReviewRequest)},
			response:    new(schema.VillageBatchReviewResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPatch,
			path:        "/api/v1/groups/{id}/shares/{transcriptID}",
			id:          "reviewShare",
			tag:         "contributions",
			description: "Approve or reject one pending submission to an owned collective. The decision changes the share-attempt status and does not mutate transcript visibility.",
			requests:    []interface{}{groupTranscriptPath, new(schema.VillageReviewShareRequest)},
			response:    new(schema.VillageReviewShareResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/groups/{id}/transcripts/{transcriptID}",
			id:          "removeGroupTranscript",
			tag:         "contributions",
			description: "Remove or revoke a transcript contribution from an owned collective without deleting the transcript from its owner's library.",
			requests:    []interface{}{groupTranscriptPath},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodPost,
			path:        "/api/v1/transcripts/{id}/share",
			id:          "shareTranscriptWithGroups",
			tag:         "contributions",
			description: "Offer one owned transcript to one or more collectives. The response is the transcript's share list, while duplicate or refused collectives are skipped unless every requested collective is already live.",
			requests:    []interface{}{transcriptPath, new(schema.VillageShareTranscriptRequest)},
			response:    new([]schema.VillageTranscriptShare),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusConflict,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/transcripts/{id}/share/{groupID}",
			id:          "unshareTranscriptFromGroup",
			tag:         "contributions",
			description: "Withdraw one owned transcript from one collective and return a status receipt.",
			requests:    []interface{}{transcriptGroupPath},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/users/me/collectives/contributions",
			id:          "listMyCollectiveContributions",
			tag:         "contributions",
			description: "List collectives the authenticated caller has offered transcripts to, with approved and pending transcript counts and rejected and withdrawn attempt counts.",
			response:    new(schema.VillageContributedCollectivesResponse),
			errorStatuses: []int{
				http.StatusUnauthorized,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/transcripts/{id}/collectives",
			id:          "listTranscriptCollectives",
			tag:         "contributions",
			description: "List visible collectives that hold a transcript. The server returns an empty collectives array, not a refusal, when memberships are hidden by collective visibility or contributor opt-in policy.",
			requests:    []interface{}{transcriptPath},
			response:    new(schema.VillageTranscriptCollectivesResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/users/me/collectives/{groupId}/submissions",
			id:          "listMyCollectiveSubmissions",
			tag:         "contributions",
			description: "List every transcript-collective pair the authenticated caller has offered to one collective, including pairs whose latest event was a withdrawal. A caller with no pair receives 404, matching no-such-collective.",
			requests:    []interface{}{ownerCollectivePath},
			response:    new([]schema.VillageCollectiveSubmission),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/users/me/collectives/{groupId}/transcripts/{transcriptId}/events",
			id:          "listShareEventHistory",
			tag:         "contributions",
			description: "List the full owner-only event history for one transcript and collective pair, oldest event first. The decided_by_actor field is an actor class, never a moderator user id. event_num identifies an attempt: pending may become approved/rejected/retracted/revoked at the same ordinal with immutable recorded_at and a newly recorded decided_at. An initially auto-approved attempt may have null decided_at. Terminal wire decision fields do not change; withdrawing/removing an approved contribution uses a new ordinal. Validate and reconcile with the canonical schema attempt helper, not a per-poll append log.",
			requests:    []interface{}{shareEventsPath},
			response:    new([]schema.VillageShareEvent),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/{id}/repositories",
			id:          "listGroupRepositories",
			tag:         "repositories",
			description: "List GitHub repositories linked to a collective. The caller must be an authenticated collective member.",
			requests:    []interface{}{groupPath},
			response:    new(schema.VillageLinkedRepositoriesResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/{id}/repositories/available",
			id:          "listAvailableGroupRepositories",
			tag:         "repositories",
			description: "List the repositories the GitHub App can offer this collective, from the installation for the collective's linked organization. The caller must be the collective owner; a collective with no linked organization returns an empty list.",
			requests:    []interface{}{groupPath},
			response:    new(schema.VillageAvailableRepositoriesResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotImplemented,
				http.StatusInternalServerError,
			},
		},
		{
			method:        http.MethodPost,
			path:          "/api/v1/groups/{id}/repositories",
			id:            "linkGroupRepository",
			tag:           "repositories",
			description:   "Link a GitHub repository to an owned collective after validating that the configured GitHub App installation can access it.",
			requests:      []interface{}{groupPath, new(schema.VillageLinkRepositoryRequest)},
			response:      new(schema.VillageLinkedRepository),
			successStatus: http.StatusCreated,
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotImplemented,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodDelete,
			path:        "/api/v1/groups/{id}/repositories/{owner}/{name}",
			id:          "unlinkGroupRepository",
			tag:         "repositories",
			description: "Unlink a GitHub repository from an owned collective.",
			requests:    []interface{}{repositoryPath},
			response:    new(schema.VillageStatusResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusInternalServerError,
			},
		},
		{
			method:      http.MethodGet,
			path:        "/api/v1/groups/{id}/repositories/{owner}/{name}/commits",
			id:          "listGroupRepositoryCommits",
			tag:         "repositories",
			description: "List cached commits for a linked repository. refresh=true fetches from GitHub and requires collective owner access; otherwise any collective member may read the cache.",
			requests:    []interface{}{repositoryCommitsRequest},
			response:    new(schema.VillageRepositoryCommitsResponse),
			errorStatuses: []int{
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusNotImplemented,
				http.StatusBadGateway,
				http.StatusInternalServerError,
			},
		},
	}

	if err := addVillageOperations(r, operations); err != nil {
		return err
	}
	for _, variant := range []struct {
		path    string
		legacy  map[string]interface{}
		grouped string
	}{
		{"/api/v1/transcripts", map[string]interface{}{"$ref": "#/components/schemas/SchemaVillageTranscriptListResponse"}, "SchemaVillageSessionListPayload"},
		{"/api/v1/groups/{id}", map[string]interface{}{"$ref": "#/components/schemas/SchemaVillageCollectiveDetailResponse"}, "SchemaVillageGroupedCollectiveDetailResponse"},
		{"/api/v1/groups/{id}/contributable", map[string]interface{}{"$ref": "#/components/schemas/SchemaVillageContributableResponse"}, "SchemaVillageGroupedContributableResponse"},
		{"/api/v1/groups/{id}/pending", map[string]interface{}{"type": "array", "items": map[string]interface{}{"$ref": "#/components/schemas/SchemaVillagePendingShare"}}, "SchemaVillageSessionListPayload"},
		{"/api/v1/groups/{id}/my-shares", map[string]interface{}{"type": "array", "items": map[string]interface{}{"$ref": "#/components/schemas/SchemaVillageUserCollectiveShare"}}, "SchemaVillageSessionListPayload"},
	} {
		if err := setVillageGroupedResponse(r.Spec, variant.path, variant.legacy, variant.grouped); err != nil {
			return err
		}
	}

	for _, body := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/groups"},
		{http.MethodPatch, "/api/v1/groups/{id}"},
		{http.MethodPatch, "/api/v1/groups/{id}/members/{userID}/role"},
		{http.MethodPost, "/api/v1/groups/{id}/shares"},
		{http.MethodPatch, "/api/v1/groups/{id}/shares"},
		{http.MethodPatch, "/api/v1/groups/{id}/shares/{transcriptID}"},
		{http.MethodPost, "/api/v1/transcripts/{id}/share"},
		{http.MethodPost, "/api/v1/groups/{id}/repositories"},
	} {
		if err := requireRequestBody(r.Spec, body.method, body.path); err != nil {
			return err
		}
	}

	return nil
}

func setVillageGroupedResponse(spec *openapi31.Spec, path string, legacy map[string]interface{}, grouped string) error {
	item, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok || item.Get == nil || item.Get.Responses == nil {
		return fmt.Errorf("configure grouped response for GET %s: operation is missing; clients could not distinguish grouped and omitted-view outputs; register the route before its response variants", path)
	}
	response, ok := item.Get.Responses.MapOfResponseOrReferenceValues["200"]
	if !ok || response.Response == nil {
		return fmt.Errorf("configure grouped response for GET %s: success response is missing; clients would have no legacy or grouped shape; register both response contracts", path)
	}
	media, ok := response.Response.Content["application/json"]
	if !ok {
		return fmt.Errorf("configure grouped response for GET %s: JSON media type is missing; clients cannot decode the selected view; register an application/json response", path)
	}
	media.Schema = map[string]interface{}{"oneOf": []interface{}{legacy, map[string]interface{}{"$ref": "#/components/schemas/" + grouped}}}
	response.Response.Content["application/json"] = media
	item.Get.Responses.MapOfResponseOrReferenceValues["200"] = response
	spec.Paths.MapOfPathItemValues[path] = item
	return nil
}

// addVillageOperations reflects one table of Village operations. Every request
// structure is added, the success response takes its declared status, and each
// error status shares the common VillageErrorResponse envelope.
func addVillageOperations(r *openapi31.Reflector, operations []villageOperationSpec) error {
	for _, op := range operations {
		oc, err := r.NewOperationContext(op.method, op.path)
		if err != nil {
			return fmt.Errorf("new Village operation %s %s: %w", op.method, op.path, err)
		}
		for _, request := range op.requests {
			oc.AddReqStructure(request)
		}
		if len(op.responseAlternatives) > 0 {
			for _, response := range op.responseAlternatives {
				oc.AddRespStructure(response)
			}
		} else if op.response != nil {
			if op.successStatus != 0 && op.successStatus != http.StatusOK {
				oc.AddRespStructure(op.response, openapicore.WithHTTPStatus(op.successStatus))
			} else {
				oc.AddRespStructure(op.response)
			}
		}
		for _, status := range op.errorStatuses {
			oc.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(status))
		}
		oc.SetDescription(op.description)
		oc.SetID(op.id)
		oc.SetTags(op.tag)
		if err := r.AddOperation(oc); err != nil {
			return fmt.Errorf("add Village operation %s %s: %w", op.method, op.path, err)
		}
	}
	return nil
}

// requirePatchRequestBody marks the PATCH request body at one path as required
// after reflection. It fails closed rather than silently doing nothing, because
// a missing path or operation would leave the body advertised as optional while
// every gate stayed green.
func requirePatchRequestBody(spec *openapi31.Spec, path string) error {
	return requireRequestBody(spec, http.MethodPatch, path)
}

func requireRequestBody(spec *openapi31.Spec, method, path string) error {
	if spec == nil || spec.Paths == nil {
		return fmt.Errorf("require %s request body for %s: the specification has no paths, so the body would remain advertised as optional", method, path)
	}
	item, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok {
		return fmt.Errorf("require %s request body for %s: the path is absent from the specification; either the operation moved or this call names a stale path, and the body would remain advertised as optional", method, path)
	}
	required := true
	switch method {
	case http.MethodPost:
		if item.Post == nil {
			return fmt.Errorf("require POST request body for %s: the path declares no POST operation; the body would remain advertised as optional", path)
		}
		if item.Post.RequestBody == nil || item.Post.RequestBody.RequestBody == nil {
			return fmt.Errorf("require POST request body for %s: the operation declares no request body to mark required; the server decodes a body unconditionally, so a contract without one would describe a request that always fails", path)
		}
		item.Post.RequestBody.RequestBody.Required = &required
	case http.MethodPatch:
		if item.Patch == nil {
			return fmt.Errorf("require PATCH request body for %s: the path declares no PATCH operation; the body would remain advertised as optional", path)
		}
		if item.Patch.RequestBody == nil || item.Patch.RequestBody.RequestBody == nil {
			return fmt.Errorf("require PATCH request body for %s: the operation declares no request body to mark required; the server decodes a body unconditionally, so a contract without one would describe a request that always fails", path)
		}
		item.Patch.RequestBody.RequestBody.Required = &required
	case http.MethodPut:
		if item.Put == nil {
			return fmt.Errorf("require PUT request body for %s: the path declares no PUT operation; the body would remain advertised as optional", path)
		}
		if item.Put.RequestBody == nil || item.Put.RequestBody.RequestBody == nil {
			return fmt.Errorf("require PUT request body for %s: the operation declares no request body to mark required; the server decodes a body unconditionally, so a contract without one would describe a request that always fails", path)
		}
		item.Put.RequestBody.RequestBody.Required = &required
	default:
		return fmt.Errorf("require %s request body for %s: unsupported method; extend requireRequestBody before adding this body gate", method, path)
	}
	spec.Paths.MapOfPathItemValues[path] = item
	return nil
}

func registerPublishMetadataComponent(r *openapi31.Reflector) error {
	var definitionErr error
	reflected, err := r.JSONSchemaReflector().Reflect(new(AuthoritativeTranscriptPublishRequest), jsonschema.CollectDefinitions(func(name string, definition jsonschema.Schema) {
		schemaMap, mapErr := definition.ToSchemaOrBool().ToSimpleMap()
		if mapErr != nil {
			definitionErr = fmt.Errorf("marshal multipart metadata dependency %s: %w", name, mapErr)
			return
		}
		fixDefinitionRefs(schemaMap)
		r.SpecEns().ComponentsEns().WithSchemasItem(name, schemaMap)
	}))
	if err != nil {
		return fmt.Errorf("reflect multipart publish metadata: %w", err)
	}
	if definitionErr != nil {
		return definitionErr
	}
	schemaMap, err := reflected.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return fmt.Errorf("marshal multipart publish metadata: %w", err)
	}
	fixDefinitionRefs(schemaMap)
	schemaMap["required"] = []interface{}{"contentHash", "model"}
	r.SpecEns().ComponentsEns().WithSchemasItem(publishRequestComponent, schemaMap)
	return nil
}

func setMultipartMetadataEncoding(spec *openapi31.Spec, operationPath string) error {
	item, ok := spec.Paths.MapOfPathItemValues[operationPath]
	if !ok || item.Post == nil || item.Post.RequestBody == nil || item.Post.RequestBody.RequestBody == nil {
		return fmt.Errorf("configure publish multipart encoding: POST %s has no reflected request body", operationPath)
	}
	body := item.Post.RequestBody.RequestBody
	body.WithRequired(true)
	media, ok := body.Content["multipart/form-data"]
	if !ok {
		return fmt.Errorf("configure publish multipart encoding: POST %s has no multipart/form-data media type", operationPath)
	}
	encoding := openapi31.Encoding{}
	encoding.WithContentType("application/json")
	media.WithEncodingItem("metadata", encoding)
	body.Content["multipart/form-data"] = media
	item.Post.RequestBody.RequestBody = body
	spec.Paths.MapOfPathItemValues[operationPath] = item
	return nil
}

// pullContractResponses names, per pull operation, the 0.3.0 and 0.2.0 JSON
// bodies a response may carry. The content read streams a blob and has none.
var pullContractResponses = []struct {
	path    string
	current map[string]interface{}
	legacy  map[string]interface{}
}{
	{"/api/v1/pull/transcripts", schemaRef("SchemaPullListResponse"), schemaRef("SchemaLegacyPullListResponse")},
	{"/api/v1/pull/transcripts/{id}", schemaRef("SchemaPullTranscriptInfo"), schemaRef("SchemaLegacyPullTranscriptInfo")},
	{"/api/v1/pull/transcripts/{id}/annotations", map[string]interface{}{"type": []interface{}{"null", "array"}, "items": schemaRef("SchemaPullAnnotation")}, map[string]interface{}{"type": []interface{}{"null", "array"}, "items": schemaRef("SchemaLegacyPullAnnotation")}},
	{"/api/v1/pull/transcripts/{id}/content", nil, nil},
	{"/api/v1/pull/transcripts/{id}/publish-metadata", nil, nil},
}

func schemaRef(name string) map[string]interface{} {
	return map[string]interface{}{"$ref": "#/components/schemas/" + name}
}

// configurePullContractResponses declares, on every pull operation, the
// Peasant-Pull-Contract response header that reports the served envelope and,
// where the response is JSON, the two bodies it selects between.
func configurePullContractResponses(spec *openapi31.Spec) error {
	required := true
	description := "The pull contract version this response was served under. 0.2.0 bodies are the Legacy* envelope; 0.3.0 and any later 0.3.z bodies are the public owner envelope."
	for _, op := range pullContractResponses {
		item, ok := spec.Paths.MapOfPathItemValues[op.path]
		if !ok || item.Get == nil || item.Get.Responses == nil {
			return fmt.Errorf("configure pull contract response for GET %s: operation is missing; clients could not learn which envelope they received; register the pull route first", op.path)
		}
		response, ok := item.Get.Responses.MapOfResponseOrReferenceValues["200"]
		if !ok || response.Response == nil {
			response = openapi31.ResponseOrReference{Response: &openapi31.Response{Description: "OK"}}
		}
		if response.Response.Headers == nil {
			response.Response.Headers = map[string]openapi31.HeaderOrReference{}
		}
		response.Response.Headers[schema.PullContractHeader] = openapi31.HeaderOrReference{Header: &openapi31.Header{
			Description: &description,
			Required:    &required,
			Schema:      map[string]interface{}{"type": "string", "pattern": `^[0-9]+\.[0-9]+\.[0-9]+$`},
		}}
		if op.current != nil {
			media, ok := response.Response.Content["application/json"]
			if !ok {
				return fmt.Errorf("configure pull contract response for GET %s: JSON media type is missing; clients cannot decode either envelope; register an application/json response", op.path)
			}
			// anyOf, not oneOf: neither envelope is a closed object, so an empty
			// page or an empty annotation list matches both. The response header
			// says which one was served.
			media.Schema = map[string]interface{}{"anyOf": []interface{}{op.current, op.legacy}}
			response.Response.Content["application/json"] = media
		}
		item.Get.Responses.MapOfResponseOrReferenceValues["200"] = response
		spec.Paths.MapOfPathItemValues[op.path] = item
	}
	return nil
}
