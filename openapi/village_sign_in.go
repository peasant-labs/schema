package openapi

import (
	"net/http"

	schema "github.com/peasant-labs/schema"
)

type villageSignInCallbackQuery struct {
	Code   string `query:"code" description:"Authorization code from the provider."`
	State  string `query:"state" description:"State echoed by the provider."`
	Error  string `query:"error" description:"Provider refusal, when present."`
	Cookie string `cookie:"oauth_state" description:"State cookie set at sign-in."`
}

type villageInstallQuery struct {
	GroupID schema.VillageUUID `query:"group_id" required:"true" description:"Collective to bind to the installation."`
}

type villageInstallCallbackQuery struct {
	InstallationID int64  `query:"installation_id" required:"true" minimum:"1" description:"GitHub App installation ID."`
	State          string `query:"state" description:"Optional signed installation state."`
}

func villageRedirect(status int, description string, cookie bool) villageResponse {
	headers := []villageHeader{{name: "Location", required: true, description: "Redirect destination."}}
	if cookie {
		headers = append(headers, villageHeader{name: "Set-Cookie", required: true, description: "oauth_state cookie with Path=/, Max-Age=300, HttpOnly and SameSite=Lax; Secure when the server base URL is HTTPS."})
	}
	return villageResponse{status: status, description: description, headers: headers, media: []villageMedia{{mediaType: villageMediaHTML}}}
}

func cliLoginResponses() villageOperationSpec {
	return villageOperationSpec{
		method: http.MethodGet, path: "/api/v1/auth/cli/login", rateLimited: true,
		responses:            []villageResponse{villageRedirect(http.StatusTemporaryRedirect, "Redirect to GitHub with an OAuth state cookie.", true)},
		browserErrorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError},
	}
}

func villageSignInOperations() []villageOperationSpec {
	providers := []struct{ path, name, label string }{
		{"github", "GitHub", "GitHub"},
		{"gitlab", "GitLab", "GitLab"},
		{"huggingface", "HuggingFace", "Hugging Face"},
		{"codeberg", "Codeberg", "Codeberg"},
		{"sourcehut", "SourceHut", "SourceHut"},
	}
	operations := make([]villageOperationSpec, 0, 12)
	for _, provider := range providers {
		path := "/api/v1/auth/" + provider.path
		entry := villageOperationSpec{
			method: http.MethodGet, path: path, id: "start" + provider.name + "SignIn", tag: "auth",
			description: "Start " + provider.label + " sign-in by redirecting to the provider. The oauth_state cookie expires after 300 seconds and is HttpOnly and SameSite=Lax; Secure is set for HTTPS.",
			rateLimited: true,
			responses:   []villageResponse{villageRedirect(http.StatusTemporaryRedirect, "Redirect to the provider authorization URL, including client_id, redirect_uri, response_type, scope and state.", true)},
		}
		callback := villageOperationSpec{
			method: http.MethodGet, path: path + "/callback", id: "complete" + provider.name + "SignIn", tag: "auth",
			description:          "Complete " + provider.label + " sign-in after verifying the oauth_state cookie. The web redirect carries a session token in its query.",
			requests:             []interface{}{new(villageSignInCallbackQuery)},
			responses:            []villageResponse{villageRedirect(http.StatusTemporaryRedirect, "Redirect to the sign-in destination. The web destination carries the session token; GitHub CLI sign-in uses a loopback URL with an exchange code and CLI state.", false)},
			browserErrorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError},
		}
		if provider.path != "github" {
			entry.description += " An unconfigured provider returns 503."
			entry.browserErrorStatuses = []int{http.StatusServiceUnavailable}
			callback.description += " An unconfigured provider returns 503."
			callback.browserErrorStatuses = append(callback.browserErrorStatuses, http.StatusServiceUnavailable)
		} else {
			callback.description += " A CLI sign-in instead redirects to its local callback with a one-time exchange code and CLI state."
		}
		operations = append(operations, entry, callback)
	}
	operations = append(operations,
		villageOperationSpec{
			method: http.MethodGet, path: "/api/v1/integrations/github/install", id: "startGitHubAppInstall", tag: "integrations",
			description: "Start a GitHub App installation for a collective owned by the caller.",
			session:     villageSessionRequired, requests: []interface{}{new(villageInstallQuery)},
			responses:            []villageResponse{villageRedirect(http.StatusFound, "Redirect to the GitHub App installation page.", false)},
			errorStatuses:        []int{http.StatusForbidden},
			browserErrorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusNotImplemented},
		},
		villageOperationSpec{
			method: http.MethodGet, path: "/api/v1/integrations/github/callback", id: "completeGitHubAppInstall", tag: "integrations",
			description: "Record a GitHub App installation and redirect to the bound collective settings or the collective list.",
			session:     villageSessionRequired, requests: []interface{}{new(villageInstallCallbackQuery)},
			responses:            []villageResponse{villageRedirect(http.StatusFound, "Redirect to the collective destination.", false)},
			browserErrorStatuses: []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusNotImplemented},
		},
	)
	return operations
}
