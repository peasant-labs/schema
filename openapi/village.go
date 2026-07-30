package openapi

import (
	"fmt"
	"net/http"

	schema "github.com/peasant-labs/schema"
	openapicore "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// TranscriptPublishRequest is the Village publish operation's HTTP body.
// Its validation-requiredness is intentionally stricter than the canonical
// schema.PublishRequest Go wire shape, so it has a distinct operation-only
// OpenAPI identity and can never shadow the canonical language-binding type.
type TranscriptPublishRequest schema.PublishRequest

// TranscriptUpdateBody is the owner-update operation's HTTP body. It carries the
// canonical schema.TranscriptUpdateRequest shape under an operation-scoped name
// so the body is a referenced component rather than an inline object, matching
// every other operation in this spec. Component harmonization then replaces its
// contents with the canonical Types definition, so the name is operation-local
// while the schema stays canonical.
type TranscriptUpdateBody schema.TranscriptUpdateRequest

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
	oc.AddReqStructure(new(TranscriptPublishRequest))
	oc.AddRespStructure(new(schema.PublishResponse))
	oc.SetDescription("Publish a transcript with session entries to the village.")
	oc.SetID("publishTranscript")
	oc.SetTags("transcripts")
	if err := r.AddOperation(oc); err != nil {
		return nil, fmt.Errorf("add publish operation: %w", err)
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
	updateOC.AddReqStructure(new(struct {
		ID string `path:"id" description:"Transcript identifier"`
	}))
	updateOC.AddReqStructure(new(TranscriptUpdateBody))
	// The success status is declared with NO body schema. That is deliberate and
	// is not the same as "returns no body": the village does return one, but it
	// currently serves an untyped object wrapping the stored row's internal
	// columns (owner_id, blob_key, project_hash, source_file_path and more, at
	// village backend/internal/handler/transcripts.go:723-727), which must not
	// enter the public contract. Those columns also serialize through pgtype
	// wrappers, so a consumer would receive {"String":"x","Valid":true} where it
	// expects a string; the served shape is not merely leaky but undecodable as a
	// typed contract. Declaring a projection the village does not actually serve
	// would break the property that the served contract and the declared contract
	// cannot drift, so nothing is declared until the handler serves a shape worth
	// declaring. Adding a response schema later is additive.
	updateOC.AddRespStructure(nil, openapicore.WithHTTPStatus(http.StatusOK))
	for _, status := range []int{
		http.StatusBadRequest,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	} {
		updateOC.AddRespStructure(new(schema.TranscriptUpdateErrorResponse), openapicore.WithHTTPStatus(status))
	}
	updateOC.SetDescription("Update an owned transcript's metadata and governance axes. Every field is " +
		"optional and an omitted field is left unchanged, resolved against the locked stored row so a " +
		"concurrent edit is not reverted. License is three-valued: omit to preserve, send the empty " +
		"string to clear, send a menu license to replace. Clearing a license that was actually granted " +
		"is refused with 400 because a granted Creative Commons license is irrevocable. Only the owner " +
		"may call this; anyone else receives 403 and neither the transcript nor its governance audit " +
		"changes. Visibility accepts private and public; organization-scoped visibility is deferred. " +
		"400 covers five distinct refusals: an unparseable transcript id, an undecodable body, a " +
		"visibility outside the accepted set, a license outside the canonical menu, and the attempt to " +
		"clear a granted license. " +
		"The refusals are declared while the 200 body is NOT, and that asymmetry is deliberate rather " +
		"than an oversight. A client must distinguish 403 from 404 from each 400 to tell a user " +
		"anything useful, so those distinctions are exactly what this contract is for. The success " +
		"body has no such consumer: the applied state is read back through " +
		"GET /api/v1/pull/transcripts/{id}. The village does return a 200 body, but it currently " +
		"serves an untyped object wrapping the stored row's internal columns (owner_id, blob_key, " +
		"project_hash, source_file_path and others) at " +
		"backend/internal/handler/transcripts.go:723-727, and those columns serialize through pgtype " +
		"wrappers, so a consumer would receive {\"String\":\"x\",\"Valid\":true} where it expects a " +
		"string. Declaring a projection the village does not serve would break the property that the " +
		"served and declared contracts cannot drift, so nothing is declared until the handler serves a " +
		"shape worth declaring. Tracked at https://github.com/peasant-labs/village/issues/55; adding " +
		"the response schema later is additive. Do not 'harmonize' this by inventing a success body.")
	updateOC.SetID("updateTranscript")
	updateOC.SetTags("transcripts")
	if err := r.AddOperation(updateOC); err != nil {
		return nil, fmt.Errorf("add transcript update operation: %w", err)
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

	// Explicitly register content-layer types not yet referenced by PublishRequest
	// but part of the publish API domain (used in future visibility controls).
	if err := addComponentSchema(r, "Visibility", new(schema.Visibility)); err != nil {
		return nil, err
	}

	AddVillageExamples(r.Spec)
	if err := harmonizeSharedTypeComponents(r.Spec); err != nil {
		return nil, fmt.Errorf("harmonize Village API shared components: %w", err)
	}

	return r.Spec, nil
}
