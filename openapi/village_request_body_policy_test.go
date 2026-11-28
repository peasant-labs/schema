package openapi

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
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

type requestBodyPolicyClass string

const (
	requestBodyPolicyClosed           requestBodyPolicyClass = "closed"
	requestBodyPolicyVersionedPayload requestBodyPolicyClass = "versioned-payload"
)

type requestBodyPolicyExpected struct {
	Class  requestBodyPolicyClass `yaml:"class"`
	Reason string                 `yaml:"reason"`
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

	rawFrozen, err := os.ReadFile(filepath.Join("..", "generated", "village-api-0.44.0.json"))
	if err != nil {
		t.Fatalf("read frozen spec: %v", err)
	}
	var frozenSpec rawOpenAPIDoc
	if err := json.Unmarshal(rawFrozen, &frozenSpec); err != nil {
		t.Fatalf("unmarshal frozen spec: %v", err)
	}

	actualOps := collectJSONRequestOperations(spec)
	fixtureOps := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		fixtureOps = append(fixtureOps, c.Input.OperationID)
	}
	assertSameNames(t, "Village JSON request body operations", fixtureOps, actualOps)

	respRefs := collectResponseSchemaRefs(spec)
	workingSchemas := spec.Components.Schemas
	frozenSchemas := frozenSpec.Components.Schemas

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			reqSchema := findOperationJSONRequestSchema(t, spec, c.Input.Path, c.Input.Method, c.Input.OperationID)
			if strings.TrimSpace(c.Expected.Reason) == "" {
				// Plain closed request body.
				if c.Expected.Class != requestBodyPolicyClosed {
					t.Fatalf("unexpected policy class %q without reason for %s", c.Expected.Class, c.Input.OperationID)
				}
				visited := map[string]bool{}
				assertAllObjectsClosed(t, c.Input.OperationID, reqSchema, workingSchemas, visited)
				switch ref := reqSchema["$ref"].(type) {
				case string:
					compName := strings.TrimPrefix(ref, "#/components/schemas/")
					if respRefs[compName] {
						t.Fatalf("%s request component %s is also used in a response; split it so responses stay open", c.Input.OperationID, compName)
					}
				}
				return
			}

			if c.Expected.Class != requestBodyPolicyVersionedPayload {
				t.Fatalf("unexpected policy class %q with reason for %s", c.Expected.Class, c.Input.OperationID)
			}
			var frozenOp rawOpenAPIOperation
			if err := json.Unmarshal(frozenSpec.Paths[c.Input.Path][strings.ToLower(c.Input.Method)], &frozenOp); err != nil {
				t.Fatalf("unmarshal frozen operation %s: %v", c.Input.OperationID, err)
			}
			if frozenOp.OperationID != c.Input.OperationID || frozenOp.RequestBody == nil {
				t.Fatalf("frozen operation %s (%s %s) missing", c.Input.OperationID, c.Input.Method, c.Input.Path)
			}
			frozenReq := frozenOp.RequestBody.Content[villageMediaJSON].Schema
			if !reflect.DeepEqual(reqSchema, frozenReq) {
				t.Fatalf("%s request schema changed from frozen baseline: got %#v, want %#v", c.Input.OperationID, reqSchema, frozenReq)
			}
			switch ref := reqSchema["$ref"].(type) {
			case string:
				compName := strings.TrimPrefix(ref, "#/components/schemas/")
				if !reflect.DeepEqual(workingSchemas[compName], frozenSchemas[compName]) {
					t.Fatalf("%s versioned component %s changed from frozen baseline", c.Input.OperationID, compName)
				}
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

func findOperationJSONRequestSchema(t *testing.T, spec *openapi31.Spec, path, method, opID string) map[string]interface{} {
	t.Helper()
	item, ok := declaredPathItem(spec, path)
	if !ok {
		t.Fatalf("operation %s (%s %s) path not found", opID, method, path)
	}
	op := pathOperation(item, method)
	if op == nil || op.ID == nil || *op.ID != opID {
		t.Fatalf("operation %s (%s %s) not found", opID, method, path)
	}
	if op.RequestBody == nil || op.RequestBody.RequestBody == nil {
		t.Fatalf("operation %s has no requestBody", opID)
	}
	media, ok := op.RequestBody.RequestBody.Content[villageMediaJSON]
	if !ok || media.Schema == nil {
		t.Fatalf("operation %s has no application/json request schema", opID)
	}
	return media.Schema
}

func assertAllObjectsClosed(t *testing.T, where string, node map[string]interface{}, schemas map[string]map[string]interface{}, visited map[string]bool) {
	t.Helper()
	if node == nil {
		return
	}
	switch ref := node["$ref"].(type) {
	case string:
		name := strings.TrimPrefix(ref, "#/components/schemas/")
		if !visited[name] {
			visited[name] = true
			target, ok := schemas[name]
			if !ok {
				t.Fatalf("%s references missing schema %s", where, name)
			}
			assertAllObjectsClosed(t, where+"->"+name, target, schemas, visited)
		}
	}
	if isObjectSchema(node) {
		switch ap := node["additionalProperties"].(type) {
		case bool:
			if ap {
				t.Fatalf("%s object schema must have additionalProperties: false, got true", where)
			}
		default:
			t.Fatalf("%s object schema must have additionalProperties: false, got %#v", where, node["additionalProperties"])
		}
	}
	switch props := node["properties"].(type) {
	case map[string]interface{}:
		for k, raw := range props {
			if propMap, ok := raw.(map[string]interface{}); ok {
				assertAllObjectsClosed(t, where+"."+k, propMap, schemas, visited)
			}
		}
	}
	switch items := node["items"].(type) {
	case map[string]interface{}:
		assertAllObjectsClosed(t, where+"[]", items, schemas, visited)
	}
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		switch arms := node[key].(type) {
		case []interface{}:
			for _, rawArm := range arms {
				if arm, ok := rawArm.(map[string]interface{}); ok {
					assertAllObjectsClosed(t, where+"."+key, arm, schemas, visited)
				}
			}
		}
	}
}

func collectResponseSchemaRefs(spec *openapi31.Spec) map[string]bool {
	out := map[string]bool{}
	if spec == nil || spec.Paths == nil {
		return out
	}
	for _, item := range spec.Paths.MapOfPathItemValues {
		for _, op := range []*openapi31.Operation{
			item.Get, item.Post, item.Put, item.Patch, item.Delete,
		} {
			if op == nil || op.Responses == nil {
				continue
			}
			for _, resp := range op.Responses.MapOfResponseOrReferenceValues {
				if resp.Response == nil {
					continue
				}
				for _, media := range resp.Response.Content {
					collectRefsRecursive(media.Schema, out)
				}
			}
		}
	}
	return out
}

func collectRefsRecursive(v interface{}, out map[string]bool) {
	switch val := v.(type) {
	case map[string]interface{}:
		switch ref := val["$ref"].(type) {
		case string:
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
