package openapi

import (
	"fmt"
	"net/http"

	schema "github.com/peasant-labs/schema"
	openapicore "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// LocalErrorResponse is the JSON error body the local server writes when it
// refuses a request. Code is a stable machine-readable reason when the server
// names one.
//
// It is operation-scoped rather than catalogued, for the same reason as
// TranscriptUpdateErrorResponse: a generic {error, code} envelope does not
// belong in the cross-language catalog as a side effect of declaring one
// status.
type LocalErrorResponse struct {
	Error string `json:"error" required:"true"`
	Code  string `json:"code,omitempty"`
}

// localWriteRefusal states the cross-origin refusal every local write route
// shares. The server refuses a state-changing request that did not come from
// its own pages before the handler runs, so nothing changes.
const localWriteRefusal = "Returns 403 with a JSON error, and changes nothing, when the request did not come from this local server: its Host header does not name a loopback address, or a browser sent it from another origin."

// localOperationSpec is one Local API operation. Every operation whose method
// changes state declares the shared 403 refusal.
type localOperationSpec struct {
	method       string
	path         string
	id           string
	tag          string
	description  string
	requests     []interface{}
	response     interface{}
	requiredBody bool
}

func addLocalOperations(r *openapi31.Reflector, operations []localOperationSpec) error {
	for _, op := range operations {
		oc, err := r.NewOperationContext(op.method, op.path)
		if err != nil {
			return fmt.Errorf("new Local API operation %s %s: %w", op.method, op.path, err)
		}
		for _, request := range op.requests {
			oc.AddReqStructure(request)
		}
		if op.response != nil {
			oc.AddRespStructure(op.response)
		}
		description := op.description
		if op.method != http.MethodGet {
			oc.AddRespStructure(new(LocalErrorResponse), openapicore.WithHTTPStatus(http.StatusForbidden))
			description += " " + localWriteRefusal
		}
		oc.SetDescription(description)
		oc.SetID(op.id)
		oc.SetTags(op.tag)
		if err := r.AddOperation(oc); err != nil {
			return fmt.Errorf("add Local API operation %s %s: %w", op.method, op.path, err)
		}
		if op.requiredBody {
			if err := requireLocalRequestBody(r.Spec, op.method, op.path); err != nil {
				return err
			}
		}
	}
	return nil
}

// requireLocalRequestBody extends the Village body gate with PUT, which only
// the local settings routes use.
func requireLocalRequestBody(spec *openapi31.Spec, method, path string) error {
	if method != http.MethodPut {
		return requireRequestBody(spec, method, path)
	}
	item, ok := spec.Paths.MapOfPathItemValues[path]
	if !ok || item.Put == nil || item.Put.RequestBody == nil || item.Put.RequestBody.RequestBody == nil {
		return fmt.Errorf("require PUT request body for %s: the operation or its body is absent, so the body would remain advertised as optional", path)
	}
	required := true
	item.Put.RequestBody.RequestBody.Required = &required
	spec.Paths.MapOfPathItemValues[path] = item
	return nil
}

// addLocalPublishingOperations declares the local web's publishing surface:
// the publication read, the typed push, the collectives proxy, the Village
// sign-in routes, the redaction preview, and the settings routes.
func addLocalPublishingOperations(r *openapi31.Reflector) error {
	bindingPath := new(struct {
		ID string `path:"id" minLength:"1" description:"Auto-publish binding identifier"`
	})
	operations := []localOperationSpec{
		{
			method: http.MethodGet, path: "/api/v1/publications", id: "listPublications", tag: "publications",
			description: "Read the durable publication state of named local sessions. This read ignores the saved selection: a session the lists leave out is still returned, with outsideSelection true. An identifier that names no session on this computer is omitted. include=audience adds the collectives each published transcript is shared with. A client counts the turns recorded after publishedAt to say how many turns are new.",
			requests: []interface{}{new(struct {
				SessionIDs string `query:"sessionIds" required:"true" minLength:"1" description:"Comma-separated local session IDs"`
				Include    string `query:"include" enum:"audience" description:"Set to audience to add each published transcript's collectives"`
			})},
			response: new(schema.LocalPublicationsResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/sync/push", id: "pushSyncSessions", tag: "sync",
			description:  "Publish or update local sessions to Village and change who can read them. Publishing is collectives only: collectives.add shares each transcript with a collective, collectives.remove takes it back, and a collective named in neither keeps its access. The request carries no visibility and no license, and the server applies no default license to a publish that names collectives. Each session result lists its steps in the order they ran; a push stops at a failed step, so later steps are not_attempted and Village keeps what it had.",
			requests:     []interface{}{new(schema.SyncPushRequest)},
			response:     new(schema.SyncPushResponse),
			requiredBody: true,
		},
		{
			method: http.MethodGet, path: "/api/v1/village/collectives", id: "listVillageCollectives", tag: "village",
			description: "List the Village collectives the signed-in user belongs to. The local server reads them from Village with this computer's stored credential. With sessionId, the server suggests collectives by comparing schema.RemoteLabel of the session's git remote with each collective's linked repositories and linked GitHub organization; clients must not normalize remotes themselves.",
			requests: []interface{}{new(struct {
				SessionID string `query:"sessionId" description:"Local session to compute suggestions for"`
			})},
			response: new(schema.LocalVillageCollectivesResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/sync/auth", id: "getSyncAuth", tag: "sync",
			description: "Report whether this computer holds a valid Village credential. A signed-out computer returns authenticated false and names no account.",
			response:    new(schema.SyncAuthResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/sync/login", id: "startSyncLogin", tag: "sync",
			description: "Start the Village sign-in in the browser and return at once. pending means the sign-in started; poll getSyncAuth until it reports authenticated. already_authenticated means nothing started.",
			response:    new(schema.SyncLoginResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/sync/logout", id: "endSyncLogin", tag: "sync",
			description: "End this computer's Village sign-in by removing its stored Village credential. This route changes local state only. Logging out a computer that holds no credential returns already_logged_out.",
			response:    new(schema.SyncLogoutResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/sync/redactions", id: "getSyncRedactions", tag: "sync",
			description: "Preview what the redaction engine hides in one session before it leaves the machine, grouped by category and rule. Each item names its line and, when the match lies inside a transcript entry, its entryIndex in the same index space as TurnDetail.index. An omitted level uses the configured default; a level the server does not offer is refused.",
			requests: []interface{}{new(struct {
				SessionID string `query:"session_id" required:"true" minLength:"1" description:"Local session to scan"`
				Level     string `query:"level" description:"Requested redaction level; omit to use the configured default"`
			})},
			response: new(schema.SyncRedactionsResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/settings", id: "getSettings", tag: "settings",
			description: "Read every setting the local settings page edits, with each key's kind, value, and metadata, and every auto-publish binding with its hook state.",
			response:    new(schema.LocalSettingsResponse),
		},
		{
			method: http.MethodPatch, path: "/api/v1/settings", id: "updateSetting", tag: "settings",
			description:  "Change one setting. The body names one key and its new value, which must match the key's kind; the response is the setting as saved, with its metadata.",
			requests:     []interface{}{new(schema.LocalSettingUpdateRequest)},
			response:     new(schema.LocalSetting),
			requiredBody: true,
		},
		{
			method: http.MethodPut, path: "/api/v1/settings/auto-publish/{id}", id: "putAutoPublishBinding", tag: "settings",
			description:  "Create or replace one auto-publish binding and install or update its git hook. The response carries the hook status and, for a blocked hook, the remedy: Peasant never overwrites a hook it does not manage.",
			requests:     []interface{}{bindingPath, new(schema.AutoPublishBindingRequest)},
			response:     new(schema.AutoPublishBinding),
			requiredBody: true,
		},
		{
			method: http.MethodDelete, path: "/api/v1/settings/auto-publish/{id}", id: "deleteAutoPublishBinding", tag: "settings",
			description: "Remove one auto-publish binding and its git hook. The response carries the hook status after removal and, when a line must be removed from a hook Peasant does not manage, the remedy.",
			requests:    []interface{}{bindingPath},
			response:    new(schema.AutoPublishRemovalResponse),
		},
	}
	return addLocalOperations(r, operations)
}
