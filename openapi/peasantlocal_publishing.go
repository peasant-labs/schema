package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
	"github.com/swaggest/openapi-go/openapi31"
)

// addLocalPublishingOperations declares the local web's publishing surface:
// the publication read, the typed push, the collectives proxy, the Village
// sign-in routes, the redaction preview, and the settings routes.
func addLocalPublishingOperations(r *openapi31.Reflector) error {
	rulePath := new(struct {
		ID string `path:"id" minLength:"1" description:"Auto-publish rule identifier"`
	})
	operations := []localOperationSpec{
		{
			method: http.MethodGet, path: "/api/v1/publications", id: "listPublications", tag: "publications",
			description: "Read the durable publication state of named local sessions for the Village account this computer is signed in to; a signed-out computer reports every session unpublished. This read ignores the saved selection: a session the lists leave out is still returned, with outsideSelection true. An identifier that names no session on this computer is omitted. include=audience adds the collectives each published transcript is shared with, as [] when it is shared with none, and returns 502 when Village cannot be read. A client counts the turns recorded after publishedAt to say how many turns are new.",
			requests: []interface{}{new(struct {
				SessionIDs string `query:"sessionIds" required:"true" minLength:"1" description:"Comma-separated local session IDs"`
				Include    string `query:"include" enum:"audience" description:"Set to audience to add each published transcript's collectives"`
			})},
			response:      new(schema.LocalPublicationsResponse),
			errorStatuses: []int{http.StatusBadGateway},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/sync/push", id: "pushSyncSessions", tag: "sync",
			description:  "Publish or update local sessions to Village and change who can read them. Publishing is collectives only: collectives.add shares each transcript with a collective, collectives.remove takes it back, and a collective named in neither keeps its access. The request carries no visibility and no license, unknown fields are refused, and the server applies no default license to a publish that names collectives. Each session result lists its steps in the order they ran. A skipped step says why and is not a failure. A failed step keeps what Village had for it and makes the session an error, which the errors count counts; a later step the server did not run after a failure is not_attempted.",
			requests:     []interface{}{new(schema.SyncPushRequest)},
			response:     new(schema.SyncPushResponse),
			requiredBody: true,
		},
		{
			method: http.MethodGet, path: "/api/v1/village/collectives", id: "listVillageCollectives", tag: "village",
			description: "List the Village collectives the signed-in user belongs to. The local server reads them from Village with this computer's stored credential: 401 when this computer is not signed in, 502 when Village cannot be read. With sessionId, the server suggests collectives by comparing schema.RemoteLabel of the session's git remote with each collective's linked repositories and linked GitHub organization; clients must not normalize remotes themselves.",
			requests: []interface{}{new(struct {
				SessionID string `query:"sessionId" description:"Local session to compute suggestions for"`
			})},
			response:      new(schema.LocalVillageCollectivesResponse),
			errorStatuses: []int{http.StatusUnauthorized, http.StatusBadGateway},
			errorResponse: new(LocalErrorResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/sync/auth", id: "getSyncAuth", tag: "sync",
			description: "Report whether this computer holds a valid Village credential. A signed-out computer returns authenticated false and names no account.",
			response:    new(schema.SyncAuthResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/sync/login", id: "syncLogin", tag: "sync",
			description: "Start the Village sign-in in the browser and return at once. pending means the sign-in started; poll getSyncAuth until it reports authenticated. already_authenticated means nothing started.",
			response:    new(schema.SyncLoginResponse),
		},
		{
			method: http.MethodPost, path: "/api/v1/sync/logout", id: "syncLogout", tag: "sync",
			description: "End this computer's Village sign-in by removing its stored Village credential. This route changes local state only. Logging out a computer that holds no credential returns already_logged_out.",
			response:    new(schema.SyncLogoutResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/sync/redactions", id: "getSyncRedactions", tag: "sync",
			description: "Preview what the redaction engine hides in one session before it leaves the machine, grouped by category and rule. Each item names its line and, when the match lies inside a turn, the entryIndex of the turn that shows it (the TurnDetail.index space) and, for a match in a tool call, its toolCallId. One item stands for every occurrence of the same text under the same rule and names the first. An omitted level uses the configured default; a level the server does not offer is refused.",
			requests: []interface{}{new(struct {
				SessionID string `query:"session_id" required:"true" minLength:"1" description:"Local session to scan"`
				Level     string `query:"level" description:"Requested redaction level; omit to use the configured default"`
			})},
			response: new(schema.SyncRedactionsResponse),
		},
		{
			method: http.MethodGet, path: "/api/v1/settings", id: "getSettings", tag: "settings",
			description: "Read every setting the local settings page shows, with each key's kind, value (null when unset), and metadata, and every auto-publish rule with the recorded repositories it matches and their hook state per event.",
			response:    new(schema.LocalSettingsResponse),
		},
		{
			method: http.MethodPatch, path: "/api/v1/settings", id: "updateSetting", tag: "settings",
			description:   "Change one editable setting. The body names one key and its new value, which must match the key's kind; null unsets the key so the default applies. The response is the setting as saved, with its metadata. A refused update changes nothing and returns the key and the reason: 400 for an unknown or read-only key, a value of the wrong kind, or a configuration the value would make invalid, and 500 when the configuration file cannot be read or written.",
			requests:      []interface{}{new(schema.LocalSettingUpdateRequest)},
			response:      new(schema.LocalSetting),
			requiredBody:  true,
			errorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError},
			errorResponse: new(schema.LocalSettingRefusal),
		},
		{
			method: http.MethodPut, path: "/api/v1/settings/auto-publish/{id}", id: "putAutoPublishRule", tag: "settings",
			description:  "Create or replace one auto-publish rule. Saving a rule installs nothing: the response lists the recorded repositories the rule matches and each one's hook state per event, and installAutoPublishHooks installs in one repository at a time.",
			requests:     []interface{}{rulePath, new(schema.AutoPublishRuleRequest)},
			response:     new(schema.AutoPublishRule),
			requiredBody: true,
		},
		{
			method: http.MethodPost, path: "/api/v1/settings/auto-publish/{id}/install", id: "installAutoPublishHooks", tag: "settings",
			description:  "Install the rule's hooks in one recorded repository the rule matches. Events are independent: the response reports each event's hook, and a blocked hook carries the remedy, because Peasant never overwrites a hook it does not manage.",
			requests:     []interface{}{rulePath, new(schema.AutoPublishInstallRequest)},
			response:     new(schema.AutoPublishRepository),
			requiredBody: true,
		},
		{
			method: http.MethodDelete, path: "/api/v1/settings/auto-publish/{id}", id: "deleteAutoPublishRule", tag: "settings",
			description: "Remove one auto-publish rule and Peasant's hooks that no other rule needs. The response reports each matched repository's hooks after removal, with the remedy for a section that must be removed by hand from a hook Peasant does not manage.",
			requests:    []interface{}{rulePath},
			response:    new(schema.AutoPublishRemovalResponse),
		},
	}
	return addLocalOperations(r, operations)
}
