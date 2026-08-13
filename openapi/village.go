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

// BuildVillageAPISpec builds the current OpenAPI 3.1 specification for the
// Village API. It describes transcript publishing, CLI authentication,
// annotation registry and manifest synchronization, schema negotiation, and
// group-scoped transcript discovery and pull operations.
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
			"and manifest synchronization, schema negotiation, and group-scoped transcript discovery, " +
			"content, annotations, and currency checks.")

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
	versionOC.SetDescription("Advertise the annotation schema version plus the push and pull " +
		"contract acceptance/negotiation windows ([Min, Current]) so the CLI can preflight " +
		"and version-negotiate before publishing or pulling.")
	versionOC.SetID("getSchemaVersion")
	versionOC.SetTags("schema")
	if err := r.AddOperation(versionOC); err != nil {
		return nil, fmt.Errorf("add schema version operation: %w", err)
	}

	// --- Pull surface: AuthRequired; thin wrappers gated by
	// canPullTranscript (404-not-403). Registers the pull envelope wire types so
	// the village and the relocated client share one contract.

	// GET /api/v1/pull/transcripts — list pullable transcripts (own + group-shared;
	// public excluded). Offset pagination via page/limit query params.
	pullListOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts")
	if err != nil {
		return nil, fmt.Errorf("new pull list operation: %w", err)
	}
	pullListOC.AddReqStructure(new(struct {
		Page  int `query:"page"`
		Limit int `query:"limit"`
	}))
	pullListOC.AddRespStructure(new(schema.PullListResponse))
	pullListOC.SetDescription("List transcripts the authenticated user may pull (own + group-shared; " +
		"public excluded by the canPullTranscript policy). Offset pagination.")
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
		ID string `path:"id"`
	}))
	pullMetaOC.AddRespStructure(new(schema.PullTranscriptInfo))
	pullMetaOC.SetDescription("Get a pullable transcript's metadata (owner, visibility, " +
		"server-computed content_hash, push contract version). 404 (not 403) when not pullable.")
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
	pullContentOC.AddReqStructure(new(struct {
		ID string `path:"id"`
	}))
	pullContentOC.SetDescription("Stream the transcript blob as served. Sends ETag: \"<content_hash>\"; " +
		"a matching If-None-Match request yields 304 Not Modified.")
	pullContentOC.SetID("getPullTranscriptContent")
	pullContentOC.SetTags("pull")
	if err := r.AddOperation(pullContentOC); err != nil {
		return nil, fmt.Errorf("add pull content operation: %w", err)
	}

	// GET /api/v1/pull/transcripts/{id}/annotations — PullAnnotation rows with
	// author identity (users join on annotations.owner_id).
	pullAnnotOC, err := r.NewOperationContext(http.MethodGet, "/api/v1/pull/transcripts/{id}/annotations")
	if err != nil {
		return nil, fmt.Errorf("new pull annotations operation: %w", err)
	}
	pullAnnotOC.AddReqStructure(new(struct {
		ID string `path:"id"`
	}))
	pullAnnotOC.AddRespStructure(new([]schema.PullAnnotation))
	pullAnnotOC.SetDescription("List a transcript's annotations as PullAnnotation rows (AnnotationSummary " +
		"plus author identity), so pulled annotations can be foreign-marked and own-authored rows excluded.")
	pullAnnotOC.SetID("getPullTranscriptAnnotations")
	pullAnnotOC.SetTags("pull")
	if err := r.AddOperation(pullAnnotOC); err != nil {
		return nil, fmt.Errorf("add pull annotations operation: %w", err)
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

	return r.Spec, nil
}

// requirePatchRequestBody marks the PATCH request body at one path as required
// after reflection. It fails closed rather than silently doing nothing, because
// a missing path or operation would leave the body advertised as optional while
// every gate stayed green.
func requirePatchRequestBody(spec *openapi31.Spec, path string) error {
	if spec == nil || spec.Paths == nil {
		return fmt.Errorf("require PATCH request body for %s: the specification has no paths, so the body would remain advertised as optional", path)
	}
	item, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok {
		return fmt.Errorf("require PATCH request body for %s: the path is absent from the specification; either the operation moved or this call names a stale path, and the body would remain advertised as optional", path)
	}
	if item.Patch == nil {
		return fmt.Errorf("require PATCH request body for %s: the path declares no PATCH operation; the body would remain advertised as optional", path)
	}
	if item.Patch.RequestBody == nil || item.Patch.RequestBody.RequestBody == nil {
		return fmt.Errorf("require PATCH request body for %s: the operation declares no request body to mark required; the server decodes a body unconditionally, so a contract without one would describe a request that always fails", path)
	}
	required := true
	item.Patch.RequestBody.RequestBody.Required = &required
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
