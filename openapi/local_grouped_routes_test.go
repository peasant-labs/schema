package openapi_test

import (
	"encoding/json"
	"os"
	"testing"

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
	Parameters         []string `yaml:"parameters"`
	RequiredParameters []string `yaml:"requiredParameters"`
	ResponseRefs       []string `yaml:"responseRefs"`
}

func TestLocalGroupedRouteContract(t *testing.T) {
	data, err := os.ReadFile("testdata/local_grouped_routes.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var corpus testcase.Corpus[localRouteInput, localRouteExpected]
	if err := yaml.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	testassert.RequireMin(t, corpus, 5)
	testassert.RequireValid(t, corpus)
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
			required := map[string]bool{}
			for _, name := range c.Expected.RequiredParameters {
				required[name] = true
			}
			for index, name := range c.Expected.Parameters {
				p := params[index].(map[string]any)
				if p["name"] != name {
					t.Fatalf("parameters[%d]=%v, want %q", index, p["name"], name)
				}
				if (p["required"] == true) != required[name] {
					t.Fatalf("parameter %q required=%v", name, p["required"])
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
		})
	}
}
