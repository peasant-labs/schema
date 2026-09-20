package schema

// ExchangeCodeRequest is the JSON body sent to the village CLI auth exchange endpoint.
// POST /api/v1/auth/cli/exchange
type ExchangeCodeRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// ExchangeCodeResponse is the JSON response from the village CLI auth exchange endpoint.
type ExchangeCodeResponse struct {
	APIKey   string `json:"api_key"`
	KeyID    string `json:"key_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// CLILoginQuery represents the query parameters for the CLI login initiation endpoint.
// GET /api/v1/auth/cli/login?port={port}&state={state}[&switch=true]
type CLILoginQuery struct {
	Port  int    `json:"port" query:"port" description:"Local callback server port"`
	State string `json:"state" query:"state" description:"OAuth state parameter for CSRF protection"`
	// Switch forces the browser session to be revoked and a fresh sign-in to be
	// required, so a user who picks a different account cannot silently receive a
	// key for the previously signed-in account. It is optional and defaults to
	// false; the default URL path is unchanged, and the parameter is stripped
	// from the sign-in redirect that follows a forced switch.
	Switch bool `json:"switch,omitempty" query:"switch" description:"Revoke the presented browser session and require a fresh sign-in before an exchange code is minted"`
}
