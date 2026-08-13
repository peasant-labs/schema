package openapi

import (
	"fmt"
	"net/http"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	openapicore "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// BuildPeasantLocalAPISpec builds an OpenAPI 3.1 specification for the local web dashboard API v1.0.
// It includes REST routes (/health, /sessions, /config/mock, /shutdown) and
// WebSocket channel message schemas (DashboardPayload, SessionsPayload,
// SessionDetailPayload, TrendsPayload, QualityPayload) as JSON Schema components.
func BuildPeasantLocalAPISpec() (*openapi31.Spec, error) {
	r := openapi31.NewReflector()
	registerHarnessSchema(r)
	r.Spec.Info.
		WithTitle("Peasant Local Dashboard API").
		// Derived from PeasantLocalAPIVersion (single source — see package doc).
		WithVersion(PeasantLocalAPIVersion).
		WithDescription("Local web dashboard API for viewing AI agent session analytics. " +
			"Includes REST endpoints and WebSocket channel message schemas.")

	// --- REST routes ---

	// GET /api/v1/health
	if err := addRESTOp(r, http.MethodGet, "/api/v1/health",
		"getHealth", "Health check endpoint.", []string{"health"},
		nil, new(schema.HealthResponse)); err != nil {
		return nil, err
	}

	// GET /api/v1/sessions
	if err := addRESTOp(r, http.MethodGet, "/api/v1/sessions",
		"listSessions", "List session summaries.", []string{"sessions"},
		nil, new(schema.SessionsPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/sessions/{id}
	// The reflector requires path parameters to be declared via a request struct
	// with `path:"name"` tags. Use an anonymous struct here so the "id" path
	// parameter is registered correctly.
	if err := addRESTOp(r, http.MethodGet, "/api/v1/sessions/{id}",
		"getSession", "Get session detail by ID.", []string{"sessions"},
		new(struct {
			ID string `path:"id"`
		}),
		new(schema.SessionDetailPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/config/mock
	if err := addRESTOp(r, http.MethodGet, "/api/v1/config/mock",
		"getMockConfig", "Get mock data configuration.", []string{"config"},
		nil, new(schema.MockConfigResponse)); err != nil {
		return nil, err
	}

	// GET /api/v1/config/capabilities
	if err := addRESTOp(r, http.MethodGet, "/api/v1/config/capabilities",
		"getUICapabilities", "Discover optional UI behavior enabled for this server process.", []string{"config"},
		nil, new(schema.UICapabilitiesResponse)); err != nil {
		return nil, err
	}

	// POST /api/v1/shutdown
	if err := addRESTOp(r, http.MethodPost, "/api/v1/shutdown",
		"postShutdown", "Gracefully shutdown the server (localhost only).", []string{"lifecycle"},
		nil, new(schema.ShutdownResponse)); err != nil {
		return nil, err
	}

	// --- WebSocket channel message schemas (as components, not operations) ---
	// These are JSON Schema components describing the payload shapes
	// sent over WebSocket channels.
	wsPayloads := []struct {
		name string
		val  interface{}
	}{
		{"DashboardPayload", new(schema.DashboardPayload)},
		{"SessionDetailPayload", new(schema.SessionDetailPayload)},
		{"TrendsPayload", new(schema.TrendsPayload)},
		{"QualityPayload", new(schema.QualityPayload)},
		{"ClientMessage", new(schema.ClientMessage)},
		{"ServerMessage", new(schema.ServerMessage)},
	}

	jsr := r.JSONSchemaReflector()
	for _, p := range wsPayloads {
		s, err := jsr.Reflect(p.val, jsonschema.CollectDefinitions(
			func(name string, defSchema jsonschema.Schema) {
				sm, smErr := defSchema.ToSchemaOrBool().ToSimpleMap()
				if smErr != nil {
					return
				}
				r.SpecEns().ComponentsEns().WithSchemasItem(name, sm)
			},
		))
		if err != nil {
			return nil, fmt.Errorf("reflect WS schema %s: %w", p.name, err)
		}
		sm, err := s.ToSchemaOrBool().ToSimpleMap()
		if err != nil {
			return nil, fmt.Errorf("marshal WS schema %s: %w", p.name, err)
		}
		r.SpecEns().ComponentsEns().WithSchemasItem(p.name, sm)
	}

	// Fix $ref values: JSON Schema Draft 4 CollectDefinitions uses "#/definitions/"
	// but OAS 3.1 components live under "#/components/schemas/". Walk all component
	// schemas and rewrite any "#/definitions/" prefix to "#/components/schemas/".
	if comps := r.SpecEns().Components; comps != nil {
		for _, schemaMap := range comps.Schemas {
			fixDefinitionRefs(schemaMap)
		}
	}

	AddPeasantLocalExamples(r.Spec)

	// --- Annotation REST endpoints ---

	// GET /api/v1/annotations?session_id={id}
	if err := addRESTOp(r, http.MethodGet, "/api/v1/annotations",
		"listAnnotations", "List annotations for a session.", []string{"annotations"},
		new(struct {
			SessionID string `query:"session_id" required:"true" description:"Session ID to filter annotations"`
		}),
		new([]schema.AnnotationSummary)); err != nil {
		return nil, err
	}

	// POST /api/v1/annotations (201 Created)
	{
		oc, err := r.NewOperationContext(http.MethodPost, "/api/v1/annotations")
		if err != nil {
			return nil, fmt.Errorf("new operation POST /api/v1/annotations: %w", err)
		}
		oc.AddReqStructure(new(schema.CreateAnnotationRequest))
		oc.AddRespStructure(new(schema.CreateAnnotationResponse), openapicore.WithHTTPStatus(http.StatusCreated))
		oc.SetDescription("Create a new annotation.")
		oc.SetID("createAnnotation")
		oc.SetTags("annotations")
		if err := r.AddOperation(oc); err != nil {
			return nil, fmt.Errorf("add operation POST /api/v1/annotations: %w", err)
		}
	}

	// GET /api/v1/annotation-types
	if err := addRESTOp(r, http.MethodGet, "/api/v1/annotation-types",
		"listAnnotationTypes", "List all registered annotation types.", []string{"annotations"},
		new(struct {
			Status string `query:"status" description:"Filter by lifecycle status (active, deprecated, etc.)"`
			Origin string `query:"origin" description:"Filter by origin (system, user, group)"`
		}),
		new([]schema.AnnotationTypeSummary)); err != nil {
		return nil, err
	}

	// --- Map / Review REST endpoints (impl contract §3) ---

	// GET /api/v1/map/{projectHash}?commit=<sha>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/map/{projectHash}",
		"getMapGraph", "Get the full map graph for a project (optionally at a commit).", []string{"map"},
		new(struct {
			ProjectHash schema.ProjectHash `path:"projectHash" description:"Opaque project hash"`
			Commit      string             `query:"commit" description:"Optional commit SHA to build the graph at (default HEAD)"`
		}),
		new(schema.MapGraphPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/map/{projectHash}/node?path=<id>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/map/{projectHash}/node",
		"getMapNodeDetail", "Get the rail panel detail for one map node.", []string{"map"},
		new(struct {
			ProjectHash schema.ProjectHash `path:"projectHash" description:"Opaque project hash"`
			Path        string             `query:"path" required:"true" description:"Repo-relative node ID"`
		}),
		new(schema.MapNodeDetailPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/map/{projectHash}/tasks?file=<path>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/map/{projectHash}/tasks",
		"listProjectTasks", "List a project's tasks (reverse-chronological, cap 500).", []string{"map"},
		new(struct {
			ProjectHash schema.ProjectHash `path:"projectHash" description:"Opaque project hash"`
			File        string             `query:"file" description:"Optional file or directory filter"`
		}),
		new(schema.ProjectTasksPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/projects/summary
	if err := addRESTOp(r, http.MethodGet, "/api/v1/projects/summary",
		"listProjectSummaries", "List per-project summary rows for the home picker (sessions, recorded coverage, last work, open changes).", []string{"map"},
		nil,
		new(schema.ProjectSummariesPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/projects/resolve?name=<display-identity>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/projects/resolve",
		"resolveProject", "Resolve one explicit project display identity without enumerating sibling projects.", []string{"map"},
		new(struct {
			Name string `query:"name" required:"true" description:"Exact project display identity from a saved route"`
		}),
		new(schema.ProjectResolutionPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/review/{projectHash}
	if err := addRESTOp(r, http.MethodGet, "/api/v1/review/{projectHash}",
		"listReviewChanges", "List a project's changes (open branches, then merged).", []string{"review"},
		new(struct {
			ProjectHash schema.ProjectHash `path:"projectHash" description:"Opaque project hash"`
		}),
		new(schema.ReviewListPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/review/{projectHash}/change?branch=<name>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/review/{projectHash}/change",
		"getChangeDetail", "Get the Review detail payload for one branch.", []string{"review"},
		new(struct {
			ProjectHash schema.ProjectHash `path:"projectHash" description:"Opaque project hash"`
			Branch      string             `query:"branch" required:"true" description:"Branch name (may contain slashes)"`
		}),
		new(schema.ChangeDetailPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/review/{projectHash}/diff?branch=<name>&file=<path>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/review/{projectHash}/diff",
		"getChangeDiff", "Get the rendered per-file unified diff for one changed file of a branch.", []string{"review"},
		new(struct {
			ProjectHash schema.ProjectHash `path:"projectHash" description:"Opaque project hash"`
			Branch      string             `query:"branch" required:"true" description:"Branch name (may contain slashes)"`
			File        string             `query:"file" required:"true" description:"Repo-relative file path"`
		}),
		new(schema.ChangeDiffPayload)); err != nil {
		return nil, err
	}

	// GET /api/v1/search?q=<query>&limit=<n>
	if err := addRESTOp(r, http.MethodGet, "/api/v1/search",
		"searchMessages", "Full-text search across recorded (redacted) message entries; matches the first ~2000 chars of each turn plus truncated tool input/output.", []string{"search"},
		new(struct {
			Q     string `query:"q" required:"true" description:"Search query (min 2 chars; whitespace tokens ANDed)"`
			Limit int    `query:"limit" description:"Max results (default 20, capped at 50)"`
		}),
		new(schema.SearchPayload)); err != nil {
		return nil, err
	}

	// --- Annotation WebSocket channel payload schema ---
	// Register AnnotationsPayload as a component so the spec documents
	// the shape of annotations WebSocket channel messages.
	if err := addComponentSchema(r, "AnnotationsPayload", new(schema.AnnotationsPayload)); err != nil {
		return nil, err
	}

	// Fix annotation component $ref values (same rewrite as above).
	if comps := r.SpecEns().Components; comps != nil {
		for _, schemaMap := range comps.Schemas {
			fixDefinitionRefs(schemaMap)
		}
	}
	if err := harmonizeSharedTypeComponents(r.Spec); err != nil {
		return nil, fmt.Errorf("harmonize Peasant Local API shared components: %w", err)
	}
	// Keep the public response component name identical across the Local API and
	// Types catalogs instead of exposing the reflector's package-prefixed alias.
	components := r.SpecEns().ComponentsEns().Schemas
	if reflected, ok := components["SchemaUICapabilitiesResponse"]; ok {
		components["UICapabilitiesResponse"] = reflected
		delete(components, "SchemaUICapabilitiesResponse")
		if path := r.Spec.Paths.MapOfPathItemValues["/api/v1/config/capabilities"]; path.Get != nil {
			response := path.Get.Responses.MapOfResponseOrReferenceValues["200"]
			media := response.Response.Content["application/json"]
			media.Schema["$ref"] = "#/components/schemas/UICapabilitiesResponse"
			response.Response.Content["application/json"] = media
			path.Get.Responses.MapOfResponseOrReferenceValues["200"] = response
			r.Spec.Paths.MapOfPathItemValues["/api/v1/config/capabilities"] = path
		}
	}

	return r.Spec, nil
}
