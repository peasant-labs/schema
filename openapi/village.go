package openapi

import (
	"fmt"
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// BuildVillageAPISpec builds an OpenAPI 3.1 specification for the Village API v1.1.
// It includes POST /api/v1/transcripts/publish with PublishRequest/PublishResponse,
// GET /api/v1/auth/cli/login and POST /api/v1/auth/cli/exchange for CLI authentication,
// and registers all content-layer types (SessionEntry, ToolCallKind, StopReason, Visibility)
// as reusable components.
func BuildVillageAPISpec() (*openapi31.Spec, error) {
	r := openapi31.NewReflector()
	registerHarnessSchema(r)
	r.Spec.Info.
		WithTitle("Transcript Publish API").
		// info.version is derived from VillageAPIVersion (the single source of
		// truth) so a doc-surface semver bump is a one-line edit in artifacts.go,
		// never a literal retyped here. See the package doc for the policy.
		WithVersion(VillageAPIVersion).
		WithDescription("API for publishing AI agent session transcripts to the data-leverage village. " +
			"v1.1 adds content-layer types (SessionEntry with ToolCallKind, StopReason, Visibility).")

	// POST /api/v1/transcripts/publish — PublishRequest in, PublishResponse out.
	oc, err := r.NewOperationContext(http.MethodPost, "/api/v1/transcripts/publish")
	if err != nil {
		return nil, fmt.Errorf("new publish operation: %w", err)
	}
	oc.AddReqStructure(new(schema.PublishRequest))
	oc.AddRespStructure(new(schema.PublishResponse))
	oc.SetDescription("Publish a transcript with session entries to the village.")
	oc.SetID("publishTranscript")
	oc.SetTags("transcripts")
	if err := r.AddOperation(oc); err != nil {
		return nil, fmt.Errorf("add publish operation: %w", err)
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

	// The reflector automatically registers component schemas for all types referenced
	// in PublishRequest, including SessionEntry, ToolCallKind, StopReason,
	// Provider, Role, EntryType, and all composite types.

	// Explicitly register content-layer types not yet referenced by PublishRequest
	// but part of the publish API domain (used in future visibility controls).
	if err := addComponentSchema(r, "Visibility", new(schema.Visibility)); err != nil {
		return nil, err
	}

	AddVillageExamples(r.Spec)

	// TODO(annotation): register annotation components for the village push format.
	//   - What: add AnnotationSummary, AnnotationTypeSummary, Provenance, ValueDomain
	//     as reusable OpenAPI components; extend PublishRequest with an optional
	//     Annotations []schema.AnnotationSummary field so buyers receive annotations
	//     alongside transcripts
	//   - Why: village consumers need annotation metadata to filter and evaluate datasets
	//   - Wiring: call addComponentSchema(r, "AnnotationSummary", new(schema.AnnotationSummary)) etc.
	//     after PublishRequest is registered; bump village spec version to "1.2"
	//   - When: the push-format annotation extension ships

	return r.Spec, nil
}
