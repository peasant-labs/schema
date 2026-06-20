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
// GET /api/v1/auth/cli/login?port={port}&state={state}
type CLILoginQuery struct {
	Port  int    `json:"port" query:"port" description:"Local callback server port"`
	State string `json:"state" query:"state" description:"OAuth state parameter for CSRF protection"`
}
