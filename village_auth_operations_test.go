package schema_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_auth_operations.yaml
var villageAuthOperationsYAML []byte

type authRouteInput struct {
	Method        string `yaml:"method"`
	Path          string `yaml:"path"`
	OperationID   string `yaml:"operation_id"`
	SuccessStatus int    `yaml:"success_status"`
	Statuses      []int  `yaml:"statuses"`
}

type authOperationRef struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type authStatusSetInput struct {
	Status     int                `yaml:"status"`
	Operations []authOperationRef `yaml:"operations"`
}

type authStrictInput struct {
	Component    string `yaml:"component"`
	UnknownField string `yaml:"unknown_field"`
}

type authRoundTripInput struct {
	Component string `yaml:"component"`
	JSON      string `yaml:"json"`
}

type authNeutralHandleInput struct {
	Component       string `yaml:"component"`
	DeprecatedField string `yaml:"deprecated_field"`
	NeutralField    string `yaml:"neutral_field"`
}

type authCLILoginInput struct {
	Path      string `yaml:"path"`
	Parameter string `yaml:"parameter"`
	Location  string `yaml:"location"`
	Required  bool   `yaml:"required"`
	Type      string `yaml:"type"`
}

type authExpected struct {
	Valid bool `yaml:"valid"`
}

type villageAuthOperationFixtures struct {
	Routes         testcase.Corpus[authRouteInput, authExpected]         `yaml:"routes"`
	StatusSets     testcase.Corpus[authStatusSetInput, authExpected]     `yaml:"status_sets"`
	StrictBodies   testcase.Corpus[authStrictInput, authExpected]        `yaml:"strict_bodies"`
	RoundTrip      testcase.Corpus[authRoundTripInput, authExpected]     `yaml:"round_trip"`
	NeutralHandles testcase.Corpus[authNeutralHandleInput, authExpected] `yaml:"neutral_handles"`
	CLILogin       testcase.Corpus[authCLILoginInput, authExpected]      `yaml:"cli_login"`
	RequiredNames  struct {
		Routes         []string `yaml:"routes"`
		StatusSets     []string `yaml:"status_sets"`
		StrictBodies   []string `yaml:"strict_bodies"`
		RoundTrip      []string `yaml:"round_trip"`
		NeutralHandles []string `yaml:"neutral_handles"`
		CLILogin       []string `yaml:"cli_login"`
	} `yaml:"required_names"`
	ExactPaths struct {
		RateLimited  []string `yaml:"rate_limited"`
		CSRF         []string `yaml:"csrf"`
		ForcedChange []string `yaml:"forced_change"`
		Admin401     []string `yaml:"admin_401"`
		Admin403     []string `yaml:"admin_403"`
	} `yaml:"status_sets_exact_paths"`
}

func loadVillageAuthOperationFixtures(t *testing.T) villageAuthOperationFixtures {
	t.Helper()
	var f villageAuthOperationFixtures
	if err := yaml.Unmarshal(villageAuthOperationsYAML, &f); err != nil {
		t.Fatalf("load Village auth operation fixtures (testdata/village_auth_operations.yaml): %v", err)
	}
	assertcase.RequireMin(t, f.Routes, 17)
	assertcase.RequireValid(t, f.Routes)
	assertcase.RequireMin(t, f.StatusSets, 5)
	assertcase.RequireValid(t, f.StatusSets)
	assertcase.RequireMin(t, f.StrictBodies, 8)
	assertcase.RequireValid(t, f.StrictBodies)
	assertcase.RequireMin(t, f.RoundTrip, 21)
	assertcase.RequireValid(t, f.RoundTrip)
	assertcase.RequireMin(t, f.NeutralHandles, 3)
	assertcase.RequireValid(t, f.NeutralHandles)
	assertcase.RequireMin(t, f.CLILogin, 1)
	assertcase.RequireValid(t, f.CLILogin)
	requireNames(t, "routes", f.RequiredNames.Routes, namesOf(f.Routes))
	requireNames(t, "status_sets", f.RequiredNames.StatusSets, namesOf(f.StatusSets))
	requireNames(t, "strict_bodies", f.RequiredNames.StrictBodies, namesOf(f.StrictBodies))
	requireNames(t, "round_trip", f.RequiredNames.RoundTrip, namesOf(f.RoundTrip))
	requireNames(t, "neutral_handles", f.RequiredNames.NeutralHandles, namesOf(f.NeutralHandles))
	requireNames(t, "cli_login", f.RequiredNames.CLILogin, namesOf(f.CLILogin))
	return f
}

// TestVillageAuthRoutesDeclareExactStatusSets pins each local-account operation
// to its exact response-code set and asserts every non-success status is carried
// by the shared error envelope, so an operation cannot gain or lose a declared
// refusal silently.
func TestVillageAuthRoutesDeclareExactStatusSets(t *testing.T) {
	f := loadVillageAuthOperationFixtures(t)
	document := villageSpecDocument(t)
	paths := document["paths"].(map[string]any)
	for _, c := range f.Routes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			op := villagePathOperation(t, paths, c.Input.Method, c.Input.Path)
			if op["operationId"] != c.Input.OperationID {
				t.Fatalf("operationId=%v, want %v", op["operationId"], c.Input.OperationID)
			}
			responses := op["responses"].(map[string]any)
			got := map[int]bool{}
			for status := range responses {
				got[statusCode(t, status)] = true
			}
			want := map[int]bool{}
			for _, status := range c.Input.Statuses {
				want[status] = true
			}
			for status := range want {
				if !got[status] {
					t.Errorf("declared status %d is absent from %s %s", status, c.Input.Method, c.Input.Path)
				}
			}
			for status := range got {
				if !want[status] {
					t.Errorf("undeclared status %d appears on %s %s", status, c.Input.Method, c.Input.Path)
				}
			}
			for _, status := range c.Input.Statuses {
				if status == c.Input.SuccessStatus {
					continue
				}
				if !responseHasRef(responses, status, "SchemaVillageErrorResponse") {
					t.Errorf("status %d on %s %s does not carry the Village error envelope", status, c.Input.Method, c.Input.Path)
				}
			}
			if !responseHasRef(responses, c.Input.SuccessStatus, "") {
				t.Errorf("success status %d on %s %s declares no response schema", c.Input.SuccessStatus, c.Input.Method, c.Input.Path)
			}
		})
	}
}

// TestVillageAuthStatusSetsCoverMiddlewareOutcomes proves each shared status is
// produced by the documented middleware rule, and that the fixture's per-group
// route manifest exactly names the routes each status set covers. A status that
// quietly appears on a new route, or a route that drops a required status, fails
// here.
func TestVillageAuthStatusSetsCoverMiddlewareOutcomes(t *testing.T) {
	f := loadVillageAuthOperationFixtures(t)
	document := villageSpecDocument(t)
	paths := document["paths"].(map[string]any)
	routeName := map[string]string{}
	for _, c := range f.Routes.Cases {
		routeName[fmt.Sprintf("%s %s", c.Input.Method, c.Input.Path)] = c.Name
	}
	exact := map[string][]string{
		"rate-limited-credential-entry-points":   f.ExactPaths.RateLimited,
		"csrf-refused-cookie-mutations":          f.ExactPaths.CSRF,
		"forced-change-gate-protected-mutations": f.ExactPaths.ForcedChange,
		"admin-boundary-anonymous":               f.ExactPaths.Admin401,
		"admin-boundary-member":                  f.ExactPaths.Admin403,
	}
	for _, c := range f.StatusSets.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var covered []string
			for _, ref := range c.Input.Operations {
				op := villagePathOperation(t, paths, ref.Method, ref.Path)
				if !responseHasStatus(op["responses"].(map[string]any), c.Input.Status) {
					t.Fatalf("%s %s does not declare status %d", ref.Method, ref.Path, c.Input.Status)
				}
				covered = append(covered, routeName[fmt.Sprintf("%s %s", ref.Method, ref.Path)])
			}
			want, ok := exact[c.Name]
			if !ok {
				t.Fatalf("fixture case %q has no status_sets_exact_paths manifest; every status set must record the routes it covers", c.Name)
			}
			if !sameStringSet(covered, want) {
				t.Fatalf("status set %q covers routes %v, want %v", c.Name, covered, want)
			}
		})
	}
}

// TestVillageAuthStrictRequestBodiesRejectUnknownFields proves the request bodies
// that must reject unknown fields are closed in the emitted Types catalog and
// that the Village operation references the same closed component. A field the
// caller could send is therefore an unambiguously refused request rather than a
// silently dropped one.
func TestVillageAuthStrictRequestBodiesRejectUnknownFields(t *testing.T) {
	f := loadVillageAuthOperationFixtures(t)
	typesComponents := typesSpecComponents(t)
	villageComponents := villageSpecDocument(t)["components"].(map[string]any)["schemas"].(map[string]any)
	for _, c := range f.StrictBodies.Cases {
		t.Run(c.Name, func(t *testing.T) {
			canonical, ok := typesComponents[c.Input.Component]
			if !ok {
				t.Fatalf("strict body %s is absent from the emitted Types catalog", c.Input.Component)
			}
			component := canonical.(map[string]any)
			if component["additionalProperties"] != false {
				t.Errorf("strict body %s is not closed: additionalProperties=%v", c.Input.Component, component["additionalProperties"])
			}
			properties, ok := component["properties"].(map[string]any)
			if !ok {
				t.Fatalf("strict body %s declares no properties map", c.Input.Component)
			}
			if _, present := properties[c.Input.UnknownField]; present {
				t.Errorf("strict body %s unexpectedly declares the unknown field %q", c.Input.Component, c.Input.UnknownField)
			}
			embedded, ok := villageComponents["Schema"+c.Input.Component].(map[string]any)
			if !ok {
				t.Fatalf("strict body %s is absent from the Village API components", c.Input.Component)
			}
			if embedded["additionalProperties"] != false {
				t.Errorf("Village API body %s is not closed: additionalProperties=%v", c.Input.Component, embedded["additionalProperties"])
			}
		})
	}
}

// TestVillageAuthCatalogTypesRoundTrip proves every new catalog type survives a
// JSON decode and re-encode unchanged, so a consumer binding cannot silently
// lose a field the contract declares.
func TestVillageAuthCatalogTypesRoundTrip(t *testing.T) {
	f := loadVillageAuthOperationFixtures(t)
	catalog := map[string]reflect.Type{}
	for _, entry := range openapi.TypeCatalogEntries() {
		valueType := reflect.TypeOf(entry.Value)
		for valueType.Kind() == reflect.Pointer {
			valueType = valueType.Elem()
		}
		catalog[entry.Name] = valueType
	}
	for _, c := range f.RoundTrip.Cases {
		t.Run(c.Name, func(t *testing.T) {
			valueType, ok := catalog[c.Input.Component]
			if !ok {
				t.Fatalf("round-trip component %s is not in the Types catalog", c.Input.Component)
			}
			value := reflect.New(valueType)
			if err := json.Unmarshal([]byte(c.Input.JSON), value.Interface()); err != nil {
				t.Fatalf("decode canonical %s: %v", c.Input.Component, err)
			}
			encoded, err := json.Marshal(value.Interface())
			if err != nil {
				t.Fatalf("encode %s: %v", c.Input.Component, err)
			}
			var want, got any
			if err := json.Unmarshal([]byte(c.Input.JSON), &want); err != nil {
				t.Fatalf("decode canonical generic %s: %v", c.Input.Component, err)
			}
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatalf("decode emitted generic %s: %v", c.Input.Component, err)
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("round trip changed %s:\n got: %s\nwant: %s", c.Input.Component, encoded, c.Input.JSON)
			}
		})
	}
}

// TestVillageNeutralHandleIsDeprecatedOnlyOnTheForgeField proves the generated
// spec marks github_username deprecated and leaves username unmarked on the user
// projection and the collective rosters.
func TestVillageNeutralHandleIsDeprecatedOnlyOnTheForgeField(t *testing.T) {
	f := loadVillageAuthOperationFixtures(t)
	components := villageSpecDocument(t)["components"].(map[string]any)["schemas"].(map[string]any)
	for _, c := range f.NeutralHandles.Cases {
		t.Run(c.Name, func(t *testing.T) {
			component, ok := components[c.Input.Component].(map[string]any)
			if !ok {
				t.Fatalf("component %s is absent from the Village API spec", c.Input.Component)
			}
			properties := component["properties"].(map[string]any)
			deprecated, ok := properties[c.Input.DeprecatedField].(map[string]any)
			if !ok {
				t.Fatalf("%s does not declare %s", c.Input.Component, c.Input.DeprecatedField)
			}
			if deprecated["deprecated"] != true {
				t.Errorf("%s.%s must be marked deprecated", c.Input.Component, c.Input.DeprecatedField)
			}
			neutral, ok := properties[c.Input.NeutralField].(map[string]any)
			if !ok {
				t.Fatalf("%s does not declare the neutral %s", c.Input.Component, c.Input.NeutralField)
			}
			if _, marked := neutral["deprecated"]; marked {
				t.Errorf("%s.%s must not be marked deprecated", c.Input.Component, c.Input.NeutralField)
			}
			required := component["required"].([]any)
			if !containsJSONString(required, c.Input.NeutralField) {
				t.Errorf("%s.%s must be present on every projection", c.Input.Component, c.Input.NeutralField)
			}
		})
	}
}

// TestVillageCLILoginDeclaresOptionalSwitchParameter proves the CLI login query
// gains exactly one optional boolean switch parameter, so the default URL path is
// unchanged and the switch appears only for forced re-authentication.
func TestVillageCLILoginDeclaresOptionalSwitchParameter(t *testing.T) {
	f := loadVillageAuthOperationFixtures(t)
	document := villageSpecDocument(t)
	paths := document["paths"].(map[string]any)
	for _, c := range f.CLILogin.Cases {
		t.Run(c.Name, func(t *testing.T) {
			op := villagePathOperation(t, paths, "GET", c.Input.Path)
			var found map[string]any
			for _, raw := range op["parameters"].([]any) {
				parameter := raw.(map[string]any)
				if parameter["name"] == c.Input.Parameter {
					found = parameter
					break
				}
			}
			if found == nil {
				t.Fatalf("CLI login does not declare parameter %s", c.Input.Parameter)
			}
			if found["in"] != c.Input.Location {
				t.Errorf("parameter %s in=%v, want %v", c.Input.Parameter, found["in"], c.Input.Location)
			}
			if found["required"] == true {
				t.Errorf("parameter %s must be optional", c.Input.Parameter)
			}
			schemaMap := found["schema"].(map[string]any)
			if schemaMap["type"] != c.Input.Type {
				t.Errorf("parameter %s type=%v, want %v", c.Input.Parameter, schemaMap["type"], c.Input.Type)
			}
		})
	}
}

func villageSpecDocument(t *testing.T) map[string]any {
	t.Helper()
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("build Village API spec: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Village API spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode Village API spec: %v", err)
	}
	return document
}

func typesSpecComponents(t *testing.T) map[string]any {
	t.Helper()
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Types spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode Types spec: %v", err)
	}
	return document["components"].(map[string]any)["schemas"].(map[string]any)
}

func villagePathOperation(t *testing.T, paths map[string]any, method, path string) map[string]any {
	t.Helper()
	item, ok := paths[path].(map[string]any)
	if !ok {
		t.Fatalf("path %s is absent from the Village API spec", path)
	}
	op, ok := item[strings.ToLower(method)].(map[string]any)
	if !ok {
		t.Fatalf("method %s is absent from path %s", method, path)
	}
	return op
}

func statusCode(t *testing.T, key string) int {
	t.Helper()
	var code int
	if _, err := fmt.Sscanf(key, "%d", &code); err != nil {
		t.Fatalf("response key %q is not a status code: %v", key, err)
	}
	return code
}

func responseHasStatus(responses map[string]any, status int) bool {
	_, ok := responses[fmt.Sprintf("%d", status)]
	return ok
}

// responseHasRef reports whether the response for status carries a JSON schema.
// When ref is non-empty the schema must reference exactly that component.
func responseHasRef(responses map[string]any, status int, ref string) bool {
	response, ok := responses[fmt.Sprintf("%d", status)].(map[string]any)
	if !ok {
		return false
	}
	content, ok := response["content"].(map[string]any)
	if !ok {
		return false
	}
	media, ok := content["application/json"].(map[string]any)
	if !ok {
		return false
	}
	schemaMap, ok := media["schema"].(map[string]any)
	if !ok {
		return false
	}
	if ref == "" {
		return true
	}
	return schemaMap["$ref"] == "#/components/schemas/"+ref
}

func containsJSONString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := map[string]int{}
	for _, value := range left {
		seen[value]++
	}
	for _, value := range right {
		seen[value]--
	}
	for _, count := range seen {
		if count != 0 {
			return false
		}
	}
	return true
}
