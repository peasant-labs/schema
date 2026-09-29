package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// The publishing corpora share one boundary harness. Each case names a Types
// component and a JSON payload. The harness runs the payload through two
// independent layers: the published component schema (the shape every
// generated binding enforces) and the Go decoder plus Validate (the relational
// rules only Go states). A case says which layer rejects it, so a shape case
// cannot pass because Go happens to refuse it, and a semantic case cannot pass
// because the schema happens to refuse it. The TypeScript suite replays the
// same files against the generated Zod schemas and must agree with the shape
// layer.

//go:embed testdata/local-api/sync_sessions.yaml
var syncSessionsBoundaryYAML []byte

//go:embed testdata/local-api/publications.yaml
var publicationsBoundaryYAML []byte

//go:embed testdata/local-api/sync_push.yaml
var syncPushBoundaryYAML []byte

//go:embed testdata/local-api/village_collectives.yaml
var villageCollectivesBoundaryYAML []byte

//go:embed testdata/local-api/sync_auth.yaml
var syncAuthBoundaryYAML []byte

//go:embed testdata/local-api/settings.yaml
var settingsBoundaryYAML []byte

//go:embed testdata/pulls/transcript_pull_requests.yaml
var transcriptPullRequestsBoundaryYAML []byte

//go:embed testdata/pulls/personal_stats.yaml
var personalStatsBoundaryYAML []byte

type boundaryFixture struct {
	RequiredNames []string                                         `yaml:"required_names"`
	Operations    []boundaryOperation                              `yaml:"operations"`
	Cases         []testcase.Case[boundaryInput, boundaryExpected] `yaml:"cases"`
}

// boundaryOperation binds an operation to the Types components it carries.
type boundaryOperation struct {
	API        string              `yaml:"api"`
	Method     string              `yaml:"method"`
	Path       string              `yaml:"path"`
	Request    string              `yaml:"request,omitempty"`
	Response   string              `yaml:"response"`
	Statuses   []string            `yaml:"statuses"`
	Parameters []boundaryParameter `yaml:"parameters,omitempty"`
}

type boundaryParameter struct {
	Name     string   `yaml:"name"`
	In       string   `yaml:"in"`
	Required bool     `yaml:"required"`
	Enum     []string `yaml:"enum,omitempty"`
}

type boundaryInput struct {
	Schema  string `yaml:"schema"`
	Payload any    `yaml:"payload"`
}

type boundaryExpected struct {
	Valid bool `yaml:"valid"`
	// RejectedBy names the layer that refuses an invalid case: shape for the
	// published component schema, semantics for the Go validator.
	RejectedBy    string `yaml:"rejected_by,omitempty"`
	ErrorContains string `yaml:"error_contains,omitempty"`
}

// boundaryDecoders is the exact set of components the publishing corpora may
// name. Each decodes the payload into its Go type and applies its validator.
var boundaryDecoders = map[string]func([]byte) (any, error){
	"LocalSyncSessionsPayload":              decodeValidated[schema.LocalSyncSessionsPayload],
	"LocalSessionRow":                       decodeValidated[schema.LocalSessionRow],
	"LocalPublicationsResponse":             decodeValidated[schema.LocalPublicationsResponse],
	"SyncPushRequest":                       decodeValidated[schema.SyncPushRequest],
	"SyncPushResponse":                      decodeValidated[schema.SyncPushResponse],
	"LocalVillageCollectivesResponse":       decodeValidated[schema.LocalVillageCollectivesResponse],
	"SyncAuthResponse":                      decodeValidated[schema.SyncAuthResponse],
	"SyncLoginResponse":                     decodeValidated[schema.SyncLoginResponse],
	"SyncLogoutResponse":                    decodeValidated[schema.SyncLogoutResponse],
	"SyncRedactionsResponse":                decodeValidated[schema.SyncRedactionsResponse],
	"LocalSettingsResponse":                 decodeValidated[schema.LocalSettingsResponse],
	"LocalSetting":                          decodeValidated[schema.LocalSetting],
	"LocalSettingUpdateRequest":             decodeValidated[schema.LocalSettingUpdateRequest],
	"AutoPublishBindingRequest":             decodeValidated[schema.AutoPublishBindingRequest],
	"AutoPublishBinding":                    decodeValidated[schema.AutoPublishBinding],
	"AutoPublishRemovalResponse":            decodeValidated[schema.AutoPublishRemovalResponse],
	"VillageUserSettings":                   decodeOnly[schema.VillageUserSettings],
	"VillageUpdateUserSettingsRequest":      decodeOnly[schema.VillageUpdateUserSettingsRequest],
	"VillageTranscriptPullRequestsResponse": decodeValidated[schema.VillageTranscriptPullRequestsResponse],
	"VillagePullRequestsSummary":            decodeValidated[schema.VillagePullRequestsSummary],
	"VillageUserStats":                      decodeValidated[schema.VillageUserStats],
	"VillageGroupTranscriptStats":           decodeOnly[schema.VillageGroupTranscriptStats],
	"VillageAvailableRepositoriesResponse":  decodeOnly[schema.VillageAvailableRepositoriesResponse],
}

func decodeValidated[T interface{ Validate() error }](raw []byte) (any, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, value.Validate()
}

func decodeOnly[T any](raw []byte) (any, error) {
	var value T
	err := json.Unmarshal(raw, &value)
	return value, err
}

func TestPublishingContractBoundaries(t *testing.T) {
	harness := newBoundaryHarness(t)
	for _, corpus := range []struct {
		name string
		data []byte
	}{
		{"testdata/local-api/sync_sessions.yaml", syncSessionsBoundaryYAML},
		{"testdata/local-api/publications.yaml", publicationsBoundaryYAML},
		{"testdata/local-api/sync_push.yaml", syncPushBoundaryYAML},
		{"testdata/local-api/village_collectives.yaml", villageCollectivesBoundaryYAML},
		{"testdata/local-api/sync_auth.yaml", syncAuthBoundaryYAML},
		{"testdata/local-api/settings.yaml", settingsBoundaryYAML},
		{"testdata/pulls/transcript_pull_requests.yaml", transcriptPullRequestsBoundaryYAML},
		{"testdata/pulls/personal_stats.yaml", personalStatsBoundaryYAML},
	} {
		t.Run(corpus.name, func(t *testing.T) {
			fixture := loadBoundaryFixture(t, corpus.data)
			for _, operation := range fixture.Operations {
				harness.requireOperation(t, operation)
			}
			for _, c := range fixture.Cases {
				t.Run(c.Name, func(t *testing.T) { harness.run(t, c) })
			}
		})
	}
}

func loadBoundaryFixture(t *testing.T, data []byte) boundaryFixture {
	t.Helper()
	var fixture boundaryFixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode boundary fixture: %v", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("boundary fixture must hold one YAML document: %v", err)
	}
	corpus := testcase.Corpus[boundaryInput, boundaryExpected]{Cases: fixture.Cases}
	assertcase.RequireMin(t, corpus, 1)
	assertcase.RequireValid(t, corpus)
	names := make([]string, 0, len(fixture.Cases))
	for _, c := range fixture.Cases {
		names = append(names, c.Name)
		switch {
		case c.Expected.Valid && c.Expected.RejectedBy != "":
			t.Fatalf("case %s is valid but names a rejecting layer", c.Name)
		case !c.Expected.Valid && c.Expected.RejectedBy != "shape" && c.Expected.RejectedBy != "semantics":
			t.Fatalf("case %s is invalid and must name rejected_by shape or semantics, got %q", c.Name, c.Expected.RejectedBy)
		case c.Expected.Valid != (c.Classification == testcase.MustPass):
			t.Fatalf("case %s classification %s disagrees with valid=%v", c.Name, c.Classification, c.Expected.Valid)
		}
		if _, ok := boundaryDecoders[c.Input.Schema]; !ok {
			t.Fatalf("case %s names component %q, which has no Go decoder in the boundary harness", c.Name, c.Input.Schema)
		}
	}
	requireExactNames(t, fixture.RequiredNames, names)
	return fixture
}

// requireExactNames is the required-name manifest check: every required case
// exists and no case exists outside the manifest.
func requireExactNames(t *testing.T, required, actual []string) {
	t.Helper()
	want := append([]string(nil), required...)
	got := append([]string(nil), actual...)
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("case names differ from required_names:\nrequired %v\nactual   %v", want, got)
	}
	for i := 1; i < len(got); i++ {
		if got[i] == got[i-1] {
			t.Fatalf("case name %q repeats", got[i])
		}
	}
}

type boundaryHarness struct {
	compiler  *jsonschema.Compiler
	documents map[string]map[string]any
	contracts map[string]*jsonschema.Schema
}

func newBoundaryHarness(t *testing.T) *boundaryHarness {
	t.Helper()
	harness := &boundaryHarness{compiler: schema.NewJSONSchemaCompiler(), documents: map[string]map[string]any{}, contracts: map[string]*jsonschema.Schema{}}
	builders := map[string]func() (any, error){
		"types":   func() (any, error) { return openapi.BuildTypesSpec() },
		"local":   func() (any, error) { return openapi.BuildPeasantLocalAPISpec() },
		"village": func() (any, error) { return openapi.BuildVillageAPISpec() },
	}
	for name, build := range builders {
		spec, err := build()
		if err != nil {
			t.Fatalf("build %s spec: %v", name, err)
		}
		raw, err := json.Marshal(spec)
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		harness.documents[name] = document
		if err := harness.compiler.AddResource(name+".json", bytes.NewReader(raw)); err != nil {
			t.Fatal(err)
		}
	}
	return harness
}

func (h *boundaryHarness) contract(t *testing.T, document, component string) *jsonschema.Schema {
	t.Helper()
	key := document + "#" + component
	if compiled, ok := h.contracts[key]; ok {
		return compiled
	}
	compiled, err := h.compiler.Compile(document + ".json#/components/schemas/" + component)
	if err != nil {
		t.Fatalf("compile %s: %v", key, err)
	}
	h.contracts[key] = compiled
	return compiled
}

func (h *boundaryHarness) run(t *testing.T, c testcase.Case[boundaryInput, boundaryExpected]) {
	raw, err := json.Marshal(c.Input.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var instance any
	if err := json.Unmarshal(raw, &instance); err != nil {
		t.Fatal(err)
	}
	shapeErr := h.contract(t, "types", c.Input.Schema).Validate(instance)
	value, goErr := boundaryDecoders[c.Input.Schema](raw)
	switch {
	case c.Expected.Valid:
		if shapeErr != nil || goErr != nil {
			t.Fatalf("valid case rejected: shape=%v go=%v", shapeErr, goErr)
		}
		h.requireRoundTrip(t, c.Input.Schema, instance, value)
	case c.Expected.RejectedBy == "shape":
		if shapeErr == nil {
			t.Fatalf("the published component accepted a case the shape layer must refuse (go=%v)", goErr)
		}
		requireErrorContains(t, shapeErr, c.Expected.ErrorContains)
	default:
		if shapeErr != nil {
			t.Fatalf("the published component refused a case only the Go validator may refuse: %v", shapeErr)
		}
		if goErr == nil {
			t.Fatal("the Go validator accepted a case it must refuse")
		}
		requireErrorContains(t, goErr, c.Expected.ErrorContains)
	}
}

func requireErrorContains(t *testing.T, err error, needle string) {
	t.Helper()
	if needle != "" && !strings.Contains(fmt.Sprintf("%#v", err), needle) && !strings.Contains(err.Error(), needle) {
		t.Fatalf("error %q does not mention %q", err, needle)
	}
}

// requireRoundTrip proves the Go type loses no accepted field and emits JSON
// the published component accepts.
func (h *boundaryHarness) requireRoundTrip(t *testing.T, component string, instance, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var output any
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatal(err)
	}
	if err := requireJSONSubset(instance, output, "$"); err != nil {
		t.Fatalf("Go round trip changed the payload: %v; emitted %s", err, encoded)
	}
	if err := h.contract(t, "types", component).Validate(output); err != nil {
		t.Fatalf("Go emitted JSON the published component refuses: %v; emitted %s", err, encoded)
	}
}

// requireOperation checks that an operation exists with its exact statuses,
// binds the named components through the harmonized operation schemas, and
// declares its parameters.
func (h *boundaryHarness) requireOperation(t *testing.T, want boundaryOperation) {
	t.Helper()
	document, ok := h.documents[want.API]
	if !ok || want.API == "types" {
		t.Fatalf("operation %s %s names unknown api %q", want.Method, want.Path, want.API)
	}
	label := want.API + " " + strings.ToUpper(want.Method) + " " + want.Path
	paths, _ := document["paths"].(map[string]any)
	item, _ := paths[want.Path].(map[string]any)
	operation, ok := item[want.Method].(map[string]any)
	if !ok {
		t.Fatalf("%s is not declared", label)
	}
	responses, _ := operation["responses"].(map[string]any)
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		statuses = append(statuses, status)
	}
	requireSameSet(t, label+" statuses", want.Statuses, statuses)
	if want.Response != "" {
		requireBinding(t, h, label+" response", want.API, want.Response, mediaSchema(t, label, responses[successStatus(want.Statuses)]))
	}
	if want.Request != "" {
		body, _ := operation["requestBody"].(map[string]any)
		if required, _ := body["required"].(bool); !required {
			t.Fatalf("%s request body is not required", label)
		}
		requireBinding(t, h, label+" request", want.API, want.Request, mediaSchema(t, label, body))
	}
	declared := map[string]map[string]any{}
	parameters, _ := operation["parameters"].([]any)
	for _, raw := range parameters {
		parameter := raw.(map[string]any)
		declared[parameter["in"].(string)+":"+parameter["name"].(string)] = parameter
	}
	for _, parameter := range want.Parameters {
		got, ok := declared[parameter.In+":"+parameter.Name]
		if !ok {
			t.Fatalf("%s does not declare %s parameter %s", label, parameter.In, parameter.Name)
		}
		if required, _ := got["required"].(bool); required != parameter.Required {
			t.Fatalf("%s parameter %s required=%v, want %v", label, parameter.Name, required, parameter.Required)
		}
		if len(parameter.Enum) > 0 {
			schemaMap, _ := got["schema"].(map[string]any)
			rawEnum, _ := schemaMap["enum"].([]any)
			values := make([]string, 0, len(rawEnum))
			for _, value := range rawEnum {
				values = append(values, fmt.Sprint(value))
			}
			requireSameSet(t, label+" parameter "+parameter.Name+" enum", parameter.Enum, values)
		}
	}
	if len(declared) != len(want.Parameters) {
		t.Fatalf("%s declares %d parameters, the fixture names %d", label, len(declared), len(want.Parameters))
	}
}

func successStatus(statuses []string) string {
	for _, status := range statuses {
		if strings.HasPrefix(status, "2") {
			return status
		}
	}
	return "200"
}

func mediaSchema(t *testing.T, label string, container any) map[string]any {
	t.Helper()
	holder, _ := container.(map[string]any)
	content, _ := holder["content"].(map[string]any)
	media, _ := content["application/json"].(map[string]any)
	schemaMap, ok := media["schema"].(map[string]any)
	if !ok {
		t.Fatalf("%s declares no application/json schema", label)
	}
	return schemaMap
}

// requireBinding checks that an operation schema references the API
// document's copy of a Types component and that the copy is the Types schema
// itself, so a consumer reading the operation gets the catalogued shape.
func requireBinding(t *testing.T, h *boundaryHarness, label, api, component string, schemaMap map[string]any) {
	t.Helper()
	const prefix = "#/components/schemas/"
	refs := []any{schemaMap}
	if arms, ok := schemaMap["oneOf"].([]any); ok {
		refs = arms
	}
	name := ""
	for _, arm := range refs {
		ref, _ := arm.(map[string]any)["$ref"].(string)
		if candidate := strings.TrimPrefix(ref, prefix); candidate == "Schema"+component || candidate == component {
			name = candidate
		}
	}
	if name == "" {
		t.Fatalf("%s references %v, want the %s component", label, schemaMap, component)
	}
	components := h.documents[api]["components"].(map[string]any)["schemas"].(map[string]any)
	typesComponents := h.documents["types"]["components"].(map[string]any)["schemas"].(map[string]any)
	apiRequired := fmt.Sprint(components[name].(map[string]any)["required"])
	typesRequired := fmt.Sprint(typesComponents[component].(map[string]any)["required"])
	if apiRequired != typesRequired {
		t.Fatalf("%s binds %s whose required set %s differs from the Types component %s", label, name, apiRequired, typesRequired)
	}
}

func requireSameSet(t *testing.T, label string, want, got []string) {
	t.Helper()
	w := append([]string(nil), want...)
	g := append([]string(nil), got...)
	sort.Strings(w)
	sort.Strings(g)
	if strings.Join(w, ",") != strings.Join(g, ",") {
		t.Fatalf("%s = %v, want %v", label, g, w)
	}
}
