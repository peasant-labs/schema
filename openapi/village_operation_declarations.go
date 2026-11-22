package openapi

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	openapicore "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// The media types a Village response may declare. application/json carries a
// reflected Go body; the two text media are declared as plain strings.
const (
	villageMediaJSON  = "application/json"
	villageMediaHTML  = "text/html"
	villageMediaPlain = "text/plain"
)

// villageAPIPrefix is the prefix of every route Village mounts inside its API
// group. A servers override is refused under it: those routes resolve against
// the document's own servers.
// unstring-allow: go-endpoint-literal — the API prefix itself, which no route member carries
const villageAPIPrefix = "/api/v1/"

// villageLimiterHeaders are the headers the per-client limiter sets on every
// response it lets through and on its own 429.
var villageLimiterHeaders = []villageHeader{
	{name: "X-RateLimit-Limit", required: true, description: "Requests one client may make in the current window.", valueType: jsonschema.Integer},
	{name: "X-RateLimit-Remaining", required: true, description: "Requests the client has left in the current window.", valueType: jsonschema.Integer},
	{name: "X-RateLimit-Reset", required: true, description: "Unix time, in seconds, at which the current window ends.", valueType: jsonschema.Integer},
}

const (
	villageRateLimitedDescription      = "Too Many Requests. The client made more requests in the current window than the limiter allows. The body is the plain text Too Many Requests."
	villageRateLimitRetryAfter         = "Seconds to wait before the window allows another request."
	villageRateLimitNoSniff            = "nosniff: the plain-text refusal must not be read as another media type. The limiter writes its refusal with Go's http.Error, which always sets it."
	villageSessionMissingDescription   = "Unauthorized. The request carries no valid session or API key."
	villageSessionLookupDescription    = "Service Unavailable. The account behind the session could not be read; retry after the Retry-After delay."
	villageSessionMergedDescription    = "Service Unavailable. Either the account behind the session could not be read, in which case Retry-After is set, or the operation's own dependency is unavailable."
	villageSessionRetryAfter           = "Seconds to wait before retrying the account lookup behind the session."
	villageSessionMergedRetryAfter     = "Seconds to wait before retrying. Present only when the account lookup behind the session failed."
	villageBrowserNegotiationSentence  = "Sent as an HTML page (text/html) when the first recognised media type in Accept is text/html or application/xhtml+xml; otherwise, including with no Accept or */*, as a VillageErrorResponse (application/json). The status is the same both ways."
	villageReflectorPlaceholderSummary = "No Content"
)

// villageBrowserPageHeaders are the headers Village sets only when it serves a
// browser refusal as its HTML page. OpenAPI declares headers per response, not
// per media type, so they are optional on the response and the description
// says which variant carries them. The policy is described by what it
// guarantees, not by its exact value, which Village owns.
var villageBrowserPageHeaders = []villageHeader{
	{name: "Content-Security-Policy", description: "Present only on the text/html variant: a restrictive policy that denies everything by default and allows only the page's styles and web fonts, so the page runs no script, cannot be framed and submits no form. Absent on the application/json variant."},
	{name: "X-Content-Type-Options", description: "Present only on the text/html variant, with the value nosniff. Absent on the application/json variant."},
}

// checkVillageOperation refuses a contradictory declaration before anything is
// reflected. It is pure: it reads only the declaration.
func checkVillageOperation(op villageOperationSpec) error {
	where := fmt.Sprintf("Village operation %s %s at openapi.checkVillageOperation", op.method, op.path)

	for _, response := range op.responses {
		if err := checkVillageResponse(where, op, response); err != nil {
			return err
		}
	}

	if op.successStatus != 0 && op.response == nil && len(op.responseAlternatives) == 0 {
		return fmt.Errorf("%s: successStatus %d is set but no response is declared, so the status would be dropped silently; set response, or declare the status in responses", where, op.successStatus)
	}

	required := op.session == villageSessionRequired
	for _, status := range op.allListedStatuses() {
		if op.rateLimited && status == http.StatusTooManyRequests {
			return fmt.Errorf("%s: status 429 is listed by hand on a rate-limited operation; rateLimited already declares the limiter's 429 with Retry-After, so a second declaration would contradict it; remove 429 from the lists", where)
		}
		if required && status == http.StatusUnauthorized {
			return fmt.Errorf("%s: status 401 is listed by hand on an operation behind the required session; session already declares the middleware's 401, so a second declaration would contradict it; remove 401 from the lists", where)
		}
	}

	owners := map[int]string{}
	claim := func(status int, owner string) error {
		if previous, ok := owners[status]; ok {
			return fmt.Errorf("%s: status %d is declared by both %s and %s, so the document could describe only one of them; declare each status in exactly one place", where, status, previous, owner)
		}
		owners[status] = owner
		return nil
	}
	if op.response != nil || len(op.responseAlternatives) > 0 {
		status := op.successStatus
		if status == 0 {
			status = http.StatusOK
		}
		if err := claim(status, "the success response"); err != nil {
			return err
		}
	}
	for _, status := range op.errorStatuses {
		if err := claim(status, "errorStatuses"); err != nil {
			return err
		}
	}
	for _, response := range op.responses {
		if err := claim(response.status, "responses"); err != nil {
			return err
		}
	}
	for _, status := range op.browserErrorStatuses {
		if err := claim(status, "browserErrorStatuses"); err != nil {
			return err
		}
	}
	if required {
		// The handler's own 503 merges with the session's: both are
		// VillageErrorResponse JSON, so one response describes them.
		if owner, ok := owners[http.StatusServiceUnavailable]; ok && owner != "errorStatuses" {
			return fmt.Errorf("%s: status 503 is declared by both the required session and %s; only errorStatuses may share the session's JSON 503; move the status to errorStatuses or drop it", where, owner)
		}
	}

	hasSuccess := op.response != nil || len(op.responseAlternatives) > 0
	for _, response := range op.responses {
		if response.status >= 200 && response.status < 400 {
			hasSuccess = true
		}
	}
	switch {
	case op.declaredStub && hasSuccess:
		return fmt.Errorf("%s: declaredStub is set but the operation declares a success response; a declared stub answers only refusals until its behaviour exists; drop the success or drop declaredStub", where)
	case op.declaredStub && !op.declaresStatus(http.StatusNotImplemented):
		return fmt.Errorf("%s: declaredStub is set but status 501 is not declared; a declared stub must state the 501 it answers; add 501 to errorStatuses", where)
	case !op.declaredStub && !hasSuccess:
		return fmt.Errorf("%s: the operation declares no success response (response, responseAlternatives, or a 2xx or 3xx in responses), so the reflector would invent a 204 the server never sends; declare the success, or set declaredStub for a route that is not implemented yet", where)
	}

	if len(op.servers) > 0 && strings.HasPrefix(op.path, villageAPIPrefix) {
		return fmt.Errorf("%s: servers overrides the document's servers on a path under %s; routes under the API prefix resolve against the document's servers; drop servers, or mount the route outside the API prefix", where, villageAPIPrefix)
	}
	return nil
}

// checkVillageResponse refuses a response entry that cannot be declared as
// written.
func checkVillageResponse(where string, op villageOperationSpec, response villageResponse) error {
	if response.status < 100 || response.status > 599 {
		return fmt.Errorf("%s: responses lists status %d, which is not an HTTP status; use a status from 100 to 599", where, response.status)
	}
	if strings.TrimSpace(response.description) == "" {
		return fmt.Errorf("%s: response %d has no description; OpenAPI requires one and a reader needs to know when it is sent; describe the response", where, response.status)
	}
	media := map[string]bool{}
	for _, medium := range response.media {
		switch medium.mediaType {
		case villageMediaJSON:
			if medium.body == nil {
				return fmt.Errorf("%s: response %d declares application/json without a Go body, so the document would carry no schema; set the body to the value the handler encodes", where, response.status)
			}
		case villageMediaHTML, villageMediaPlain:
			if medium.body != nil {
				return fmt.Errorf("%s: response %d declares %s with a Go body; a text medium is a plain string and reflects no structure; drop the body", where, response.status, medium.mediaType)
			}
		default:
			return fmt.Errorf("%s: response %d declares media type %q, which the Village generator does not declare; use application/json, text/html or text/plain", where, response.status, medium.mediaType)
		}
		if media[medium.mediaType] {
			return fmt.Errorf("%s: response %d declares media type %s twice; one status carries one schema per media type; merge the entries", where, response.status, medium.mediaType)
		}
		media[medium.mediaType] = true
	}

	reserved := map[string]string{}
	if op.rateLimited {
		for _, header := range villageLimiterHeaders {
			reserved[strings.ToLower(header.name)] = "rateLimited"
		}
	}
	seen := map[string]bool{}
	location := false
	for _, header := range response.headers {
		key := strings.ToLower(header.name)
		if key == "" {
			return fmt.Errorf("%s: response %d declares a header with no name; name the header", where, response.status)
		}
		if seen[key] {
			return fmt.Errorf("%s: response %d declares header %s twice; a header is declared once per response; merge the entries", where, response.status, header.name)
		}
		if owner, ok := reserved[key]; ok {
			return fmt.Errorf("%s: response %d declares header %s, which %s already declares on every response; drop it from the response", where, response.status, header.name, owner)
		}
		seen[key] = true
		if key == "location" && header.required {
			location = true
		}
	}
	if response.status >= 300 && response.status < 400 && !location {
		return fmt.Errorf("%s: redirect %d has no required Location header, so a client could not learn where it is sent; declare Location with required set", where, response.status)
	}
	return nil
}

// allListedStatuses returns every status the declaration lists by hand.
func (op villageOperationSpec) allListedStatuses() []int {
	statuses := append([]int(nil), op.errorStatuses...)
	for _, response := range op.responses {
		statuses = append(statuses, response.status)
	}
	return append(statuses, op.browserErrorStatuses...)
}

// declaresStatus reports whether the declaration lists status by hand.
func (op villageOperationSpec) declaresStatus(status int) bool {
	for _, listed := range op.allListedStatuses() {
		if listed == status {
			return true
		}
	}
	return false
}

// declaredStatuses returns every status the operation's document entry must
// carry once it is declared: the success, the listed statuses, and the
// statuses the session and limiter add.
func (op villageOperationSpec) declaredStatuses() map[string]bool {
	statuses := map[string]bool{}
	if op.response != nil || len(op.responseAlternatives) > 0 {
		status := op.successStatus
		if status == 0 {
			status = http.StatusOK
		}
		statuses[statusText(status)] = true
	}
	for _, status := range op.allListedStatuses() {
		statuses[statusText(status)] = true
	}
	if op.session == villageSessionRequired {
		statuses[statusText(http.StatusUnauthorized)] = true
		statuses[statusText(http.StatusServiceUnavailable)] = true
	}
	if op.rateLimited {
		statuses[statusText(http.StatusTooManyRequests)] = true
	}
	return statuses
}

// reflectVillageOperation adds one operation to the reflector: its requests,
// its success, its JSON-bodied responses and its JSON refusals, so every Go
// body registers its components exactly as the reflector always has. Headers,
// text media and placeholders are left to declareVillageResponses.
func reflectVillageOperation(r *openapi31.Reflector, op villageOperationSpec) error {
	oc, err := r.NewOperationContext(op.method, op.path)
	if err != nil {
		return fmt.Errorf("new Village operation %s %s: %w", op.method, op.path, err)
	}
	for _, request := range op.requests {
		oc.AddReqStructure(request)
	}
	if len(op.responseAlternatives) > 0 {
		for _, response := range op.responseAlternatives {
			oc.AddRespStructure(response)
		}
	} else if op.response != nil {
		if op.successStatus != 0 && op.successStatus != http.StatusOK {
			oc.AddRespStructure(op.response, openapicore.WithHTTPStatus(op.successStatus))
		} else {
			oc.AddRespStructure(op.response)
		}
	}
	for _, status := range op.errorStatuses {
		oc.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(status))
	}
	for _, response := range op.responses {
		for _, medium := range response.media {
			if medium.mediaType == villageMediaJSON {
				oc.AddRespStructure(medium.body, openapicore.WithHTTPStatus(response.status))
			}
		}
	}
	for _, status := range op.browserErrorStatuses {
		oc.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(status))
	}
	if op.session == villageSessionRequired {
		oc.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(http.StatusUnauthorized))
		if !op.declaresStatus(http.StatusServiceUnavailable) {
			oc.AddRespStructure(new(schema.VillageErrorResponse), openapicore.WithHTTPStatus(http.StatusServiceUnavailable))
		}
	}
	oc.SetDescription(op.description)
	oc.SetID(op.id)
	oc.SetTags(op.tag)
	if err := r.AddOperation(oc); err != nil {
		return fmt.Errorf("add Village operation %s %s: %w", op.method, op.path, err)
	}
	return nil
}

// declareVillageResponses completes a reflected operation: it removes the
// reflector's placeholder 204, adds the body-less and text responses, the
// text/html alternative of each browser refusal, the session and limiter
// refusals and headers, and the path-item servers. It then requires the
// operation's statuses to be exactly the declared ones.
func declareVillageResponses(spec *openapi31.Spec, op villageOperationSpec) error {
	where := fmt.Sprintf("Village operation %s %s at openapi.declareVillageResponses", op.method, op.path)
	item, ok := declaredPathItem(spec, op.path)
	if !ok {
		return fmt.Errorf("%s: the path is not in the document, so its responses cannot be declared; reflect the operation first", where)
	}
	operation := pathOperation(item, op.method)
	if operation == nil || operation.Responses == nil {
		return fmt.Errorf("%s: the path has no %s operation, so its responses cannot be declared; reflect the operation first", where, op.method)
	}
	responses := operation.Responses.MapOfResponseOrReferenceValues
	declared := op.declaredStatuses()

	// The reflector invents a 204 for an operation it reflected no response
	// for. A redirect-only operation is one; its real statuses come next.
	noContent := statusText(http.StatusNoContent)
	if placeholder, ok := responses[noContent]; ok && !declared[noContent] && isReflectorPlaceholder(placeholder) {
		delete(responses, noContent)
	}

	for _, declaredResponse := range op.responses {
		key := statusText(declaredResponse.status)
		response := ensureInlineResponse(responses, key)
		response.Description = declaredResponse.description
		for _, medium := range declaredResponse.media {
			if medium.mediaType == villageMediaJSON {
				if _, ok := response.Content[villageMediaJSON]; !ok {
					return fmt.Errorf("%s: response %s has no reflected application/json body; the reflector did not register the declared Go body", where, key)
				}
				continue
			}
			if err := setTextMedia(response, medium.mediaType); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}
		for _, header := range declaredResponse.headers {
			if err := setVillageHeader(response, header); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}
		responses[key] = openapi31.ResponseOrReference{Response: response}
	}

	for _, status := range op.browserErrorStatuses {
		key := statusText(status)
		response := ensureInlineResponse(responses, key)
		summary, ok := op.errorSummaries[status]
		if !ok {
			summary = http.StatusText(status) + "."
		}
		response.Description = summary + " " + villageBrowserNegotiationSentence
		if err := setTextMedia(response, villageMediaHTML); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		for _, header := range villageBrowserPageHeaders {
			if err := setVillageHeader(response, header); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}
		responses[key] = openapi31.ResponseOrReference{Response: response}
	}

	if op.session == villageSessionRequired {
		unauthorized := ensureInlineResponse(responses, statusText(http.StatusUnauthorized))
		unauthorized.Description = villageSessionMissingDescription
		responses[statusText(http.StatusUnauthorized)] = openapi31.ResponseOrReference{Response: unauthorized}

		unavailable := ensureInlineResponse(responses, statusText(http.StatusServiceUnavailable))
		retryAfter := villageHeader{name: "Retry-After", required: true, description: villageSessionRetryAfter, valueType: jsonschema.Integer}
		unavailable.Description = villageSessionLookupDescription
		if op.declaresStatus(http.StatusServiceUnavailable) {
			// The handler answers 503 too, without Retry-After.
			retryAfter.required = false
			retryAfter.description = villageSessionMergedRetryAfter
			unavailable.Description = villageSessionMergedDescription
		}
		if err := setVillageHeader(unavailable, retryAfter); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		responses[statusText(http.StatusServiceUnavailable)] = openapi31.ResponseOrReference{Response: unavailable}
	}

	if op.rateLimited {
		limited := ensureInlineResponse(responses, statusText(http.StatusTooManyRequests))
		limited.Description = villageRateLimitedDescription
		if err := setTextMedia(limited, villageMediaPlain); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if err := setVillageHeader(limited, villageHeader{name: "Retry-After", required: true, description: villageRateLimitRetryAfter, valueType: jsonschema.Integer}); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if err := setVillageHeader(limited, villageHeader{name: "X-Content-Type-Options", required: true, description: villageRateLimitNoSniff}); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		responses[statusText(http.StatusTooManyRequests)] = openapi31.ResponseOrReference{Response: limited}
		for key, response := range responses {
			if response.Response == nil {
				return fmt.Errorf("%s: response %s is a reference, so the limiter headers cannot be attached; declare it inline", where, key)
			}
			for _, header := range villageLimiterHeaders {
				if err := setVillageHeader(response.Response, header); err != nil {
					return fmt.Errorf("%s: %w", where, err)
				}
			}
			responses[key] = response
		}
	}

	if len(op.servers) > 0 {
		servers := make([]openapi31.Server, len(op.servers))
		for i, url := range op.servers {
			servers[i] = openapi31.Server{URL: url}
		}
		item.Servers = servers
	}
	spec.Paths.MapOfPathItemValues[op.path] = item

	return verifyVillageStatuses(spec, op)
}

// verifyVillageStatuses requires an operation's document entry to carry
// exactly the statuses its declaration states: no placeholder or stray status
// the server never sends, and none missing.
func verifyVillageStatuses(spec *openapi31.Spec, op villageOperationSpec) error {
	where := fmt.Sprintf("Village operation %s %s at openapi.verifyVillageStatuses", op.method, op.path)
	item, ok := declaredPathItem(spec, op.path)
	if !ok {
		return fmt.Errorf("%s: the path is not in the document; reflect the operation first", where)
	}
	operation := pathOperation(item, op.method)
	if operation == nil || operation.Responses == nil {
		return fmt.Errorf("%s: the path has no %s operation; reflect the operation first", where, op.method)
	}
	declared := op.declaredStatuses()
	var undeclared, missing []string
	for key := range operation.Responses.MapOfResponseOrReferenceValues {
		if !declared[key] {
			undeclared = append(undeclared, key)
		}
	}
	for key := range declared {
		if _, ok := operation.Responses.MapOfResponseOrReferenceValues[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(undeclared)
	sort.Strings(missing)
	if len(undeclared) > 0 {
		return fmt.Errorf("%s: the document carries status %s, which the declaration does not state (a 204 here is the reflector's placeholder for an operation with no reflected response); the contract would promise a response the server never sends; declare the status or remove it", where, strings.Join(undeclared, ", "))
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s: the declaration states status %s, which the document does not carry; the generator dropped a response; declare it through the operation's fields", where, strings.Join(missing, ", "))
	}
	return nil
}

func isReflectorPlaceholder(response openapi31.ResponseOrReference) bool {
	return response.Response != nil &&
		response.Response.Description == villageReflectorPlaceholderSummary &&
		len(response.Response.Content) == 0 &&
		len(response.Response.Headers) == 0
}

func ensureInlineResponse(responses map[string]openapi31.ResponseOrReference, key string) *openapi31.Response {
	if existing, ok := responses[key]; ok && existing.Response != nil {
		return existing.Response
	}
	return &openapi31.Response{}
}

func setTextMedia(response *openapi31.Response, mediaType string) error {
	schema, err := villageTypeSchema(jsonschema.String)
	if err != nil {
		return err
	}
	if response.Content == nil {
		response.Content = map[string]openapi31.MediaType{}
	}
	response.Content[mediaType] = openapi31.MediaType{Schema: schema}
	return nil
}

func setVillageHeader(response *openapi31.Response, header villageHeader) error {
	valueType := header.valueType
	if valueType == "" {
		valueType = jsonschema.String
	}
	schema, err := villageTypeSchema(valueType)
	if err != nil {
		return fmt.Errorf("declare header %s: %w", header.name, err)
	}
	if response.Headers == nil {
		response.Headers = map[string]openapi31.HeaderOrReference{}
	}
	value := openapi31.Header{Schema: schema}
	if header.description != "" {
		description := header.description
		value.Description = &description
	}
	if header.required {
		required := true
		value.Required = &required
	}
	response.Headers[header.name] = openapi31.HeaderOrReference{Header: &value}
	return nil
}

// villageTypeSchema is the JSON schema that admits one simple type, for a
// header value or a text body.
func villageTypeSchema(valueType jsonschema.SimpleType) (map[string]interface{}, error) {
	schema, err := valueType.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return nil, fmt.Errorf("build the %s schema: %w", valueType, err)
	}
	return schema, nil
}
