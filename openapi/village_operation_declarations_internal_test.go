package openapi

import (
	_ "embed"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"github.com/swaggest/openapi-go/openapi31"
)

//go:embed testdata/village_operation_declarations.yaml
var villageOperationDeclarationsYAML []byte

// requiredVillageDeclarationCases names every case of the Village declaration
// corpus. A case added or removed without an entry here fails the test, so a
// rule cannot lose its proof silently.
var requiredVillageDeclarationCases = []string{
	"redirect-without-location",
	"status-declared-twice",
	"session-503-outside-error-statuses",
	"json-media-without-body",
	"text-media-with-body",
	"unsupported-media-type",
	"media-declared-twice",
	"no-success-without-declared-stub",
	"declared-stub-without-501",
	"declared-stub-with-success",
	"success-status-without-response",
	"rate-limited-lists-429",
	"session-required-lists-401",
	"servers-on-api-path",
	"duplicate-header",
	"header-repeats-limiter-header",
	"response-without-description",
	"status-out-of-range",
	"placeholder-204-remains",
	"declared-status-missing",
	"complete-redirect-accepted",
	"session-required-declares-refusals",
	"session-503-merges-with-handler-503",
	"declared-stub-accepted",
	"json-success-with-headers-at-root",
	"browser-error-summary-keeps-negotiation",
}

// villageDeclarationInput is one probe declaration and the step it is run
// through.
type villageDeclarationInput struct {
	// Call is add_operations (check, reflect, declare and describe), check
	// (checkVillageOperation alone) or verify_statuses (reflect alone, then
	// verifyVillageStatuses, which sees what the reflector left).
	Call      string       `yaml:"call"`
	Operation villageProbe `yaml:"operation"`
}

// villageDeclarationExpected is what the step must produce: the error text of
// a refused declaration, or the document entry of an accepted one.
type villageDeclarationExpected struct {
	ErrorContains string                       `yaml:"error_contains"`
	Document      *villageOperationExpectation `yaml:"document"`
}

// villageProbe is the part of a villageOperationSpec a declaration can get
// wrong. Go values cannot be written in YAML, so bodies and requests are named
// and resolved by probeBody and probeRequest.
type villageProbe struct {
	Method               string                 `yaml:"method"`
	Path                 string                 `yaml:"path"`
	ID                   string                 `yaml:"id"`
	Description          string                 `yaml:"description"`
	Session              string                 `yaml:"session"`
	RateLimited          bool                   `yaml:"rate_limited"`
	Request              string                 `yaml:"request"`
	Response             string                 `yaml:"response"`
	SuccessStatus        int                    `yaml:"success_status"`
	ErrorStatuses        []int                  `yaml:"error_statuses"`
	ErrorSummaries       map[int]string         `yaml:"error_summaries"`
	BrowserErrorStatuses []int                  `yaml:"browser_error_statuses"`
	DeclaredStub         bool                   `yaml:"declared_stub"`
	Servers              []string               `yaml:"servers"`
	Responses            []villageProbeResponse `yaml:"responses"`
}

type villageProbeResponse struct {
	Status      int                  `yaml:"status"`
	Description string               `yaml:"description"`
	Headers     []villageProbeHeader `yaml:"headers"`
	Media       []villageProbeMedia  `yaml:"media"`
}

type villageProbeHeader struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required"`
}

type villageProbeMedia struct {
	Type string `yaml:"type"`
	Body string `yaml:"body"`
}

// villageProbeSignInRequest carries a query parameter and a cookie parameter,
// as a sign-in callback does.
type villageProbeSignInRequest struct {
	State      string `query:"state" description:"State echoed by the provider."`
	OAuthState string `cookie:"oauth_state" description:"State cookie set when sign-in started."`
}

// probeBodies, probeRequests and probeSessions resolve the names a probe
// uses. Each name is the Go identifier of the value it stands for.
var (
	probeBodies = map[string]func() interface{}{
		"VillageStatusResponse": func() interface{} { return new(schema.VillageStatusResponse) },
		"VillageErrorResponse":  func() interface{} { return new(schema.VillageErrorResponse) },
	}
	probeRequests = map[string]func() interface{}{
		"villageProbeSignInRequest": func() interface{} { return new(villageProbeSignInRequest) },
	}
	probeSessions = map[string]villageSession{
		"villageSessionNone":     villageSessionNone,
		"villageSessionOptional": villageSessionOptional,
		"villageSessionRequired": villageSessionRequired,
	}
)

func probeBody(t *testing.T, name string) interface{} {
	t.Helper()
	if name == "" {
		return nil
	}
	body, ok := probeBodies[name]
	if !ok {
		t.Fatalf("probe names the unknown body %q; name a type listed in probeBodies", name)
	}
	return body()
}

func probeRequest(t *testing.T, name string) []interface{} {
	t.Helper()
	if name == "" {
		return nil
	}
	request, ok := probeRequests[name]
	if !ok {
		t.Fatalf("probe names the unknown request %q; name a type listed in probeRequests", name)
	}
	return []interface{}{request()}
}

func probeSession(t *testing.T, name string) villageSession {
	t.Helper()
	if name == "" {
		return villageSessionNone
	}
	session, ok := probeSessions[name]
	if !ok {
		t.Fatalf("probe names the unknown session %q; name a constant listed in probeSessions", name)
	}
	return session
}

func (p villageProbe) spec(t *testing.T) villageOperationSpec {
	t.Helper()
	op := villageOperationSpec{
		method: p.Method, path: p.Path, id: p.ID, tag: "probe", description: p.Description,
		session: probeSession(t, p.Session), rateLimited: p.RateLimited,
		requests: probeRequest(t, p.Request), response: probeBody(t, p.Response),
		successStatus: p.SuccessStatus, errorStatuses: p.ErrorStatuses,
		errorSummaries: p.ErrorSummaries, browserErrorStatuses: p.BrowserErrorStatuses,
		declaredStub: p.DeclaredStub, servers: p.Servers,
	}
	for _, response := range p.Responses {
		declared := villageResponse{status: response.Status, description: response.Description}
		for _, header := range response.Headers {
			declared.headers = append(declared.headers, villageHeader{name: header.Name, required: header.Required, description: "probe header"})
		}
		for _, medium := range response.Media {
			declared.media = append(declared.media, villageMedia{mediaType: medium.Type, body: probeBody(t, medium.Body)})
		}
		op.responses = append(op.responses, declared)
	}
	return op
}

// The generator's own declarations are valid, so the checks that refuse a
// contradictory one never fire in a build. Each must-fail case builds one
// contradictory declaration and requires the error its check must give; each
// must-pass case builds one accepted declaration and requires its exact
// document entry. A check that stops firing, or a capability that stops
// declaring what it promises, fails here.
func TestVillageOperationDeclarationsRefuseMistakes(t *testing.T) {
	corpus, err := testcase.LoadCorpus[villageDeclarationInput, villageDeclarationExpected](villageOperationDeclarationsYAML)
	if err != nil {
		t.Fatalf("load the Village declaration corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 1)
	assert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	assertSameNames(t, "Village declaration cases", names, requiredVillageDeclarationCases)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			switch c.Classification {
			case testcase.MustFail:
				if c.Expected.ErrorContains == "" || c.Expected.Document != nil {
					t.Fatal("a must-fail case names the error it must produce and no document")
				}
			case testcase.MustPass:
				if c.Expected.ErrorContains != "" || c.Expected.Document == nil {
					t.Fatal("a must-pass case names the document entry it must produce and no error")
				}
			}
			op := c.Input.Operation.spec(t)
			reflector := openapi31.NewReflector()
			switch c.Input.Call {
			case "add_operations":
				err = addVillageOperations(reflector, []villageOperationSpec{op})
			case "check":
				err = checkVillageOperation(op)
			case "verify_statuses":
				if reflectErr := reflectVillageOperation(reflector, op); reflectErr != nil {
					t.Fatalf("reflect the probe: %v", reflectErr)
				}
				err = verifyVillageStatuses(reflector.Spec, op)
			default:
				t.Fatalf("case names the unknown call %q; use add_operations, check or verify_statuses", c.Input.Call)
			}
			if c.Classification == testcase.MustFail {
				if err == nil {
					t.Fatalf("%s accepted a contradictory declaration; it must fail with %q", c.Input.Call, c.Expected.ErrorContains)
				}
				if !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("%s error = %q, want it to contain %q", c.Input.Call, err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s refused a valid declaration: %v", c.Input.Call, err)
			}
			for _, problem := range checkVillageOperationExpectation(reflector.Spec, op.method, op.path, *c.Expected.Document) {
				t.Error(problem)
			}
		})
	}
}

// assertSameNames requires two name lists to hold exactly the same members.
func assertSameNames(t *testing.T, label string, got, want []string) {
	t.Helper()
	gotSet := map[string]bool{}
	for _, name := range got {
		if gotSet[name] {
			t.Fatalf("%s: %q appears twice", label, name)
		}
		gotSet[name] = true
	}
	wantSet := map[string]bool{}
	for _, name := range want {
		wantSet[name] = true
	}
	var missing, extra []string
	for name := range wantSet {
		if !gotSet[name] {
			missing = append(missing, name)
		}
	}
	for name := range gotSet {
		if !wantSet[name] {
			extra = append(extra, name)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("%s: missing %v, unexpected %v; add or remove the entry together with its required name", label, missing, extra)
	}
}
