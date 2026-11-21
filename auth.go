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
//
// Both parameters are required: Village refuses a request missing either with
// 400, as it does a port that is not an integer from 1024 to 65535, the range
// the port declares.
type CLILoginQuery struct {
	Port  int    `json:"port" query:"port" required:"true" minimum:"1024" maximum:"65535" description:"Local callback server port"`
	State string `json:"state" query:"state" required:"true" description:"OAuth state parameter for CSRF protection"`
}
