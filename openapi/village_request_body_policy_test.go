package openapi

import (
	_ "embed"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"github.com/swaggest/openapi-go/openapi31"
)

//go:embed testdata/village_request_body_policy.yaml
var villageRequestBodyPolicyYAML []byte

type requestBodyPolicyInput struct {
	OperationID string `yaml:"operation_id"`
	Method      string `yaml:"method"`
	Path        string `yaml:"path"`
}

type requestBodyPolicyExpected struct {
	Class  string `yaml:"class"`
	Reason string `yaml:"reason"`
}

func TestVillageRequestBodyPolicy(t *testing.T) {
	corpus, err := testcase.LoadCorpus[requestBodyPolicyInput, requestBodyPolicyExpected](villageRequestBodyPolicyYAML)
	if err != nil {
		t.Fatalf("load village_request_body_policy.yaml: %v", err)
	}
	assert.RequireMin(t, corpus, 1)
	assert.RequireValid(t, corpus)

	spec, err := BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	rawWorking, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal working spec: %v", err)
	}
	var workingDoc map[string]interface{}
	if err := json.Unmarshal(rawWorking, &workingDoc); err != nil {
		t.Fatalf("unmarshal working spec: %v", err)
	}

	rawFrozen, err := os.ReadFile("../generated/village-api-0.44.0.json")
	if err != nil {
		t.Fatalf("read frozen spec: %v", err)
	}
	var frozenDoc map[string]interface{}
	if err := json.Unmarshal(rawFrozen, &frozenDoc); err != nil {
		t.Fatalf("unmarshal frozen spec: %v", err)
	}

	actualOps := collectJSONRequestOperations(spec)
	fixtureOps := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		fixtureOps = append(fixtureOps, c.Input.OperationID)
	}
	assertSameNames(t, "Village JSON request body operations", fixtureOps, actualOps)

	respRefs := collectResponseSchemaRefs(workingDoc)
	workingSchemas, _ := workingDoc["components"].(map[string]interface{})["schemas"].(map[string]interface{})
	frozenSchemas, _ := frozenDoc["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			reqSchema := findOperationJSONRequestSchema(t, workingDoc, c.Input.Path, strings.ToLower(c.Input.Method), c.Input.OperationID)
			switch c.Expected.Class {
			case "closed":
				visited := map[string]bool{}
				assertAllObjectsClosed(t, c.Input.OperationID, reqSchema, workingSchemas, visited)
				if ref, ok := reqSchema["$ref"].(string); ok {
					compName := strings.TrimPrefix(ref, "#/components/schemas/")
					if respRefs[compName] {
						t.Fatalf("%s request component %s is also used in a response; split it so responses stay open", c.Input.OperationID, compName)
					}
				}
			case "versioned-payload":
				if strings.TrimSpace(c.Expected.Reason) == "" {
					t.Fatalf("%s versioned-payload entry must provide a one-line reason", c.Input.OperationID)
				}
				frozenReq := findOperationJSONRequestSchema(t, frozenDoc, c.Input.Path, strings.ToLower(c.Input.Method), c.Input.OperationID)
				if !reflect.DeepEqual(reqSchema, frozenReq) {
					t.Fatalf("%s request schema changed from frozen baseline: got %#v, want %#v", c.Input.OperationID, reqSchema, frozenReq)
				}
				if ref, ok := reqSchema["$ref"].(string); ok {
					compName := strings.TrimPrefix(ref, "#/components/schemas/")
					if !reflect.DeepEqual(workingSchemas[compName], frozenSchemas[compName]) {
						t.Fatalf("%s versioned component %s changed from frozen baseline", c.Input.OperationID, compName)
					}
				}
			default:
				t.Fatalf("unknown policy class %q for %s", c.Expected.Class, c.Input.OperationID)
			}
		})
	}
}

func collectJSONRequestOperations(spec *openapi31.Spec) []string {
	var out []string
	if spec == nil || spec.Paths == nil {
		return out
	}
	for _, item := range spec.Paths.MapOfPathItemValues {
		for _, op := range []*openapi31.Operation{
			item.Get, item.Post, item.Put, item.Patch, item.Delete,
		} {
			if op == nil || op.RequestBody == nil || op.RequestBody.RequestBody == nil {
				continue
			}
			if _, ok := op.RequestBody.RequestBody.Content[villageMediaJSON]; !ok {
				continue
			}
			if op.ID != nil {
				out = append(out, *op.ID)
			}
		}
	}
	return out
}

func findOperationJSONRequestSchema(t *testing.T, doc map[string]interface{}, path, method, opID string) map[string]interface{} {
	t.Helper()
	paths, _ := doc["paths"].(map[string]interface{})
	item, _ := paths[path].(map[string]interface{})
	op, _ := item[method].(map[string]interface{})
	if op == nil || op["operationId"] != opID {
		t.Fatalf("operation %s (%s %s) not found", opID, method, path)
	}
	rb, _ := op["requestBody"].(map[string]interface{})
	content, _ := rb["content"].(map[string]interface{})
	media, _ := content[villageMediaJSON].(map[string]interface{})
	schemaMap, _ := media["schema"].(map[string]interface{})
	if schemaMap == nil {
		t.Fatalf("operation %s has no application/json request schema", opID)
	}
	return schemaMap
}

func assertAllObjectsClosed(t *testing.T, where string, node map[string]interface{}, schemas map[string]interface{}, visited map[string]bool) {
	t.Helper()
	if node == nil {
		return
	}
	if ref, ok := node["$ref"].(string); ok {
		name := strings.TrimPrefix(ref, "#/components/schemas/")
		if !visited[name] {
			visited[name] = true
			target, ok := schemas[name].(map[string]interface{})
			if !ok {
				t.Fatalf("%s references missing schema %s", where, name)
			}
			assertAllObjectsClosed(t, where+"->"+name, target, schemas, visited)
		}
	}
	if isObjectSchema(node) {
		ap, ok := node["additionalProperties"].(bool)
		if !ok || ap {
			t.Fatalf("%s object schema must have additionalProperties: false, got %#v", where, node["additionalProperties"])
		}
	}
	if props, ok := node["properties"].(map[string]interface{}); ok {
		for k, raw := range props {
			if propMap, ok := raw.(map[string]interface{}); ok {
				assertAllObjectsClosed(t, where+"."+k, propMap, schemas, visited)
			}
		}
	}
	if items, ok := node["items"].(map[string]interface{}); ok {
		assertAllObjectsClosed(t, where+"[]", items, schemas, visited)
	}
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		if arms, ok := node[key].([]interface{}); ok {
			for _, rawArm := range arms {
				if arm, ok := rawArm.(map[string]interface{}); ok {
					assertAllObjectsClosed(t, where+"."+key, arm, schemas, visited)
				}
			}
		}
	}
}

func collectResponseSchemaRefs(doc map[string]interface{}) map[string]bool {
	out := map[string]bool{}
	paths, _ := doc["paths"].(map[string]interface{})
	for _, rawItem := range paths {
		item, _ := rawItem.(map[string]interface{})
		for _, method := range []string{"get", "post", "put", "patch", "delete"} {
			op, _ := item[method].(map[string]interface{})
			if op == nil {
				continue
			}
			collectRefsRecursive(op["responses"], out)
		}
	}
	return out
}

func collectRefsRecursive(v interface{}, out map[string]bool) {
	switch val := v.(type) {
	case map[string]interface{}:
		if ref, ok := val["$ref"].(string); ok {
			out[strings.TrimPrefix(ref, "#/components/schemas/")] = true
		}
		for _, child := range val {
			collectRefsRecursive(child, out)
		}
	case []interface{}:
		for _, child := range val {
			collectRefsRecursive(child, out)
		}
	}
}
