package openapi_test

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"github.com/swaggest/openapi-go/openapi3"
	"gopkg.in/yaml.v3"
)

type localRouteInput struct {
	Path        string `yaml:"path"`
	OperationID string `yaml:"operationId"`
}
type localRouteParam struct {
	Name, In, Type string
	Required       bool
	Enum           []any
	Minimum        *int
}

type localRouteExpected struct {
	Parameters    []localRouteParam `yaml:"parameters"`
	Positioned    bool              `yaml:"positioned"`
	ResponseRefs  []string          `yaml:"responseRefs"`
	PrimaryRef    string            `yaml:"primaryRef"`
	LegacyBody    string            `yaml:"legacyBody"`
	GroupedBody   string            `yaml:"groupedBody"`
	RequiredRoot  []string          `yaml:"requiredRoot"`
	ForbiddenRoot []string          `yaml:"forbiddenRoot"`
	NonnullArrays []string          `yaml:"nonnullArrays"`
}

type positionedSingleDefInput struct {
	PositionedParameterNames []string `yaml:"positionedParameterNames"`
	PositionedRouteNames     []string `yaml:"positionedRouteNames"`
}

type positionedSingleDefExpected struct {
	PositionedResponseRefs []string `yaml:"positionedResponseRefs"`
}

type localRouteFixtures struct {
	RequiredNames              []string                                                               `yaml:"required_names"`
	PositionedParameters       []localRouteParam                                                      `yaml:"positioned_parameters"`
	PositionedSingleDefinition testcase.Corpus[positionedSingleDefInput, positionedSingleDefExpected] `yaml:"positioned_single_definition"`
	Routes                     testcase.Corpus[localRouteInput, localRouteExpected]                   `yaml:"routes"`
}

func TestLocalGroupedRouteContract(t *testing.T) {
	data, err := os.ReadFile("testdata/local_grouped_routes.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures localRouteFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	corpus := fixtures.Routes
	testassert.RequireMin(t, corpus, 5)
	testassert.RequireValid(t, corpus)
	testassert.RequireMin(t, fixtures.PositionedSingleDefinition, 1)
	testassert.RequireValid(t, fixtures.PositionedSingleDefinition)
	if len(fixtures.RequiredNames) != len(corpus.Cases) {
		t.Fatalf("route required-name manifest has %d names for %d cases", len(fixtures.RequiredNames), len(corpus.Cases))
	}
	for i, name := range fixtures.RequiredNames {
		if corpus.Cases[i].Name != name {
			t.Fatalf("route case[%d]=%q, want required name %q", i, corpus.Cases[i].Name, name)
		}
	}
	for _, singleDef := range fixtures.PositionedSingleDefinition.Cases {
		t.Run(singleDef.Name, func(t *testing.T) {
			paramNames := make([]string, len(fixtures.PositionedParameters))
			for i, p := range fixtures.PositionedParameters {
				paramNames[i] = p.Name
			}
			if !slices.Equal(paramNames, singleDef.Input.PositionedParameterNames) {
				t.Fatalf("positioned_parameters=%v, want %v", paramNames, singleDef.Input.PositionedParameterNames)
			}
			var positionedRoutes []string
			seenRefs := map[string]int{}
			for _, route := range corpus.Cases {
				for _, p := range route.Expected.Parameters {
					if slices.Contains(singleDef.Input.PositionedParameterNames, p.Name) {
						t.Fatalf("route %s duplicates positioned parameter %q inline; define it only in positioned_parameters", route.Name, p.Name)
					}
				}
				if route.Expected.Positioned {
					positionedRoutes = append(positionedRoutes, route.Name)
				}
				for _, ref := range route.Expected.ResponseRefs {
					if slices.Contains(singleDef.Expected.PositionedResponseRefs, ref) {
						seenRefs[ref]++
					}
				}
			}
			if !slices.Equal(positionedRoutes, singleDef.Input.PositionedRouteNames) {
				t.Fatalf("positioned routes=%v, want %v", positionedRoutes, singleDef.Input.PositionedRouteNames)
			}
			for _, ref := range singleDef.Expected.PositionedResponseRefs {
				if seenRefs[ref] != 1 {
					t.Fatalf("positioned responseRef %s appeared %d times across routes, want 1", ref, seenRefs[ref])
				}
			}
		})
	}
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(spec)
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)["schemas"].(map[string]any)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			op := paths[c.Input.Path].(map[string]any)["get"].(map[string]any)
			if op["operationId"] != c.Input.OperationID {
				t.Fatalf("operationId=%v", op["operationId"])
			}
			params := op["parameters"].([]any)
			expectedParams := c.Expected.Parameters
			if c.Expected.Positioned {
				expectedParams = mergePositionedParameters(c.Expected.Parameters, fixtures.PositionedParameters)
			}
			if len(params) != len(expectedParams) {
				t.Fatalf("parameters=%d, want %d", len(params), len(expectedParams))
			}
			for index, expected := range expectedParams {
				p := params[index].(map[string]any)
				if p["name"] != expected.Name || p["in"] != expected.In {
					t.Fatalf("parameters[%d]=%v/%v, want %q/%q", index, p["name"], p["in"], expected.Name, expected.In)
				}
				if (p["required"] == true) != expected.Required {
					t.Fatalf("parameter %q required=%v", expected.Name, p["required"])
				}
				shape := p["schema"].(map[string]any)
				if ref, ok := shape["$ref"].(string); ok {
					shape = components[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
				}
				if shape["type"] != expected.Type {
					t.Fatalf("parameter %q type=%v, want %q", expected.Name, shape["type"], expected.Type)
				}
				if expected.Minimum != nil && shape["minimum"] != float64(*expected.Minimum) {
					t.Fatalf("parameter %q minimum=%v", expected.Name, shape["minimum"])
				}
				if len(expected.Enum) != 0 {
					got := shape["enum"].([]any)
					if len(got) != len(expected.Enum) {
						t.Fatalf("parameter %q enum=%v, want %v", expected.Name, got, expected.Enum)
					}
					for i, want := range expected.Enum {
						if localRouteJSONValue(t, expected.Name, got[i]) != localRouteJSONValue(t, expected.Name, want) {
							t.Fatalf("parameter %q enum[%d]=%v (%s), want %v (%s); the wire type differs", expected.Name, i, got[i], localRouteJSONValue(t, expected.Name, got[i]), want, localRouteJSONValue(t, expected.Name, want))
						}
					}
				}
			}
			schemaValue := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
			refs := []string{}
			if oneOf, ok := schemaValue["oneOf"].([]any); ok {
				for _, arm := range oneOf {
					refs = append(refs, arm.(map[string]any)["$ref"].(string))
				}
			} else {
				refs = append(refs, schemaValue["$ref"].(string))
			}
			if len(refs) != len(c.Expected.ResponseRefs) {
				t.Fatalf("response refs=%v", refs)
			}
			for i := range refs {
				if refs[i] != c.Expected.ResponseRefs[i] {
					t.Fatalf("response refs=%v, want %v", refs, c.Expected.ResponseRefs)
				}
			}
			for _, ref := range refs {
				if _, ok := components[strings.TrimPrefix(ref, "#/components/schemas/")]; !ok {
					t.Fatalf("response ref %q does not resolve", ref)
				}
			}
			primaryRef := refs[len(refs)-1]
			if c.Expected.PrimaryRef != "" {
				if !slices.Contains(refs, c.Expected.PrimaryRef) {
					t.Fatalf("primary response %s is not in %v", c.Expected.PrimaryRef, refs)
				}
				primaryRef = c.Expected.PrimaryRef
			}
			primary := components[strings.TrimPrefix(primaryRef, "#/components/schemas/")].(map[string]any)
			properties, _ := primary["properties"].(map[string]any)
			for _, name := range c.Expected.RequiredRoot {
				if _, ok := properties[name]; !ok {
					t.Fatalf("response component missing %q", name)
				}
			}
			for _, name := range c.Expected.ForbiddenRoot {
				if _, ok := properties[name]; ok {
					t.Fatalf("response component unexpectedly contains wrapper %q", name)
				}
			}
			for _, name := range c.Expected.NonnullArrays {
				if properties[name].(map[string]any)["type"] != "array" {
					t.Fatalf("%s is not a non-null array", name)
				}
			}
			if c.Expected.LegacyBody != "" {
				var legacy schema.LocalSyncSessionsPayload
				if err := json.Unmarshal([]byte(c.Expected.LegacyBody), &legacy); err != nil {
					t.Fatal(err)
				}
				if legacy.Sessions == nil || len(legacy.Sessions) != 1 || legacy.Sessions[0].SyncStatus != schema.SyncStatusNew {
					t.Fatalf("legacy sync body lost concrete row: %+v", legacy)
				}
			}
			if c.Expected.GroupedBody != "" {
				var grouped schema.LocalSessionListPayload
				if err := json.Unmarshal([]byte(c.Expected.GroupedBody), &grouped); err != nil {
					t.Fatal(err)
				}
				if err := grouped.Validate(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	ws := components["SessionDetailReadPayload"].(map[string]any)["properties"].(map[string]any)
	if _, ok := ws["id"]; !ok {
		t.Fatal("WS session_detail read component is not flat at id")
	}
	legacyWS := components["SchemaSessionsPayload"].(map[string]any)["properties"].(map[string]any)["sessions"].(map[string]any)
	if types, ok := legacyWS["type"].([]any); !ok || len(types) == 0 || types[0] != "array" {
		t.Fatal("legacy WS sessions no longer includes its existing array shape")
	}
}

// localRouteJSONValue renders one parameter enum member through encoding/json so
// the comparison is type-exact: json.Marshal(25) is "25" while json.Marshal("25")
// is "\"25\"", and fmt.Sprint would render both as 25.
func localRouteJSONValue(t *testing.T, name string, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("parameter %q enum value %v cannot be encoded: %v", name, value, err)
	}
	return string(encoded)
}

func mergePositionedParameters(routeParams, positioned []localRouteParam) []localRouteParam {
	insertAt := len(routeParams)
	for i, p := range routeParams {
		if p.In == string(openapi3.ParameterInPath) {
			insertAt = i
			break
		}
	}
	merged := make([]localRouteParam, 0, len(routeParams)+len(positioned))
	merged = append(merged, routeParams[:insertAt]...)
	merged = append(merged, positioned...)
	merged = append(merged, routeParams[insertAt:]...)
	return merged
}
