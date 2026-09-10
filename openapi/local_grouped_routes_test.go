package openapi_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

type localRouteInput struct {
	Path        string `yaml:"path"`
	OperationID string `yaml:"operationId"`
}
type localRouteExpected struct {
	Parameters []struct {
		Name, In, Type string
		Required       bool
		Enum           []string
		Minimum        *int
	} `yaml:"parameters"`
	ResponseRefs  []string `yaml:"responseRefs"`
	LegacyBody    string   `yaml:"legacyBody"`
	GroupedBody   string   `yaml:"groupedBody"`
	RequiredRoot  []string `yaml:"requiredRoot"`
	ForbiddenRoot []string `yaml:"forbiddenRoot"`
	NonnullArrays []string `yaml:"nonnullArrays"`
}
type localRouteFixtures struct {
	RequiredNames []string                                             `yaml:"required_names"`
	Routes        testcase.Corpus[localRouteInput, localRouteExpected] `yaml:"routes"`
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
	if len(fixtures.RequiredNames) != len(corpus.Cases) {
		t.Fatalf("route required-name manifest has %d names for %d cases", len(fixtures.RequiredNames), len(corpus.Cases))
	}
	for i, name := range fixtures.RequiredNames {
		if corpus.Cases[i].Name != name {
			t.Fatalf("route case[%d]=%q, want required name %q", i, corpus.Cases[i].Name, name)
		}
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
			if len(params) != len(c.Expected.Parameters) {
				t.Fatalf("parameters=%d, want %d", len(params), len(c.Expected.Parameters))
			}
			for index, expected := range c.Expected.Parameters {
				p := params[index].(map[string]any)
				if p["name"] != expected.Name || p["in"] != expected.In {
					t.Fatalf("parameters[%d]=%v/%v, want %q/%q", index, p["name"], p["in"], expected.Name, expected.In)
				}
				if (p["required"] == true) != expected.Required {
					t.Fatalf("parameter %q required=%v", expected.Name, p["required"])
				}
				shape := p["schema"].(map[string]any)
				if shape["type"] != expected.Type {
					t.Fatalf("parameter %q type=%v, want %q", expected.Name, shape["type"], expected.Type)
				}
				if expected.Minimum != nil && shape["minimum"] != float64(*expected.Minimum) {
					t.Fatalf("parameter %q minimum=%v", expected.Name, shape["minimum"])
				}
				if len(expected.Enum) != 0 {
					got := []string{}
					for _, v := range shape["enum"].([]any) {
						got = append(got, v.(string))
					}
					if !reflect.DeepEqual(got, expected.Enum) {
						t.Fatalf("parameter %q enum=%v", expected.Name, got)
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
			primary := components[strings.TrimPrefix(refs[len(refs)-1], "#/components/schemas/")].(map[string]any)
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
				if legacy.Sessions == nil || len(legacy.Sessions) != 1 || legacy.Sessions[0].SyncStatus != "new" {
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
