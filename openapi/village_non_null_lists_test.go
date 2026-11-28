package openapi

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"github.com/swaggest/openapi-go/openapi31"
)

//go:embed testdata/village_non_null_lists.yaml
var villageNonNullListsYAML []byte

var requiredVillageNonNullListCases = []string{
	"top-level-slice-non-null",
	"nested-nullable-array-preserved",
	"frozen-village-spec-untouched",
}

type nonNullListInput struct {
	Kind       string `yaml:"kind"`
	Path       string `yaml:"path"`
	FrozenFile string `yaml:"frozen_file"`
}

type nonNullListExpected struct {
	ResponseType   string   `yaml:"response_type"`
	NestedProperty string   `yaml:"nested_property"`
	NestedTypes    []string `yaml:"nested_types"`
	FrozenType     string   `yaml:"frozen_type"`
	FrozenTypes    []string `yaml:"frozen_types"`
}

type probeObjectWithNullableSlice struct {
	Items []string `json:"items"`
}

func TestVillageNonNullTopLevelLists(t *testing.T) {
	corpus, err := testcase.LoadCorpus[nonNullListInput, nonNullListExpected](villageNonNullListsYAML)
	if err != nil {
		t.Fatalf("load village_non_null_lists.yaml: %v", err)
	}
	assert.RequireMin(t, corpus, len(requiredVillageNonNullListCases))
	assert.RequireValid(t, corpus)

	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	assertSameNames(t, "Village non-null list cases", names, requiredVillageNonNullListCases)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			switch c.Input.Kind {
			case "top-level-slice":
				r := openapi31.NewReflector()
				err := addVillageOperations(r, []villageOperationSpec{{
					method:      http.MethodGet,
					path:        c.Input.Path,
					id:          "probeList",
					tag:         "probe",
					description: "Probe top-level slice response.",
					response:    new([]schema.VillageTagUsage),
				}})
				if err != nil {
					t.Fatalf("addVillageOperations: %v", err)
				}
				schemaMap := responseJSONSchema(t, r.Spec, c.Input.Path, http.MethodGet, "200")
				if got := schemaMap["type"]; got != c.Expected.ResponseType {
					t.Fatalf("top-level slice type = %#v, want %q", got, c.Expected.ResponseType)
				}

			case "nested-slice":
				r := openapi31.NewReflector()
				err := addVillageOperations(r, []villageOperationSpec{{
					method:      http.MethodGet,
					path:        c.Input.Path,
					id:          "probeObject",
					tag:         "probe",
					description: "Probe object with nested nullable slice.",
					response:    new(probeObjectWithNullableSlice),
				}})
				if err != nil {
					t.Fatalf("addVillageOperations: %v", err)
				}
				schemaMap := responseJSONSchema(t, r.Spec, c.Input.Path, http.MethodGet, "200")
				resolved := resolveSchemaRef(t, r.Spec, schemaMap)
				props, _ := resolved["properties"].(map[string]interface{})
				prop, _ := props[c.Expected.NestedProperty].(map[string]interface{})
				gotTypes := stringSlice(prop["type"])
				if !reflect.DeepEqual(gotTypes, c.Expected.NestedTypes) {
					t.Fatalf("nested %q type = %#v, want %#v", c.Expected.NestedProperty, prop["type"], c.Expected.NestedTypes)
				}

			case "frozen-check":
				spec, err := BuildVillageAPISpec()
				if err != nil {
					t.Fatalf("BuildVillageAPISpec: %v", err)
				}
				working := responseJSONSchema(t, spec, c.Input.Path, http.MethodGet, "200")
				if got := working["type"]; got != c.Expected.ResponseType {
					t.Fatalf("working %s type = %#v, want %q", c.Input.Path, got, c.Expected.ResponseType)
				}
				raw, err := os.ReadFile(c.Input.FrozenFile)
				if err != nil {
					t.Fatalf("read frozen spec %s: %v", c.Input.FrozenFile, err)
				}
				var frozenDoc map[string]interface{}
				if err := json.Unmarshal(raw, &frozenDoc); err != nil {
					t.Fatalf("unmarshal frozen spec: %v", err)
				}
				frozenSchema := rawDocResponseSchema(t, frozenDoc, c.Input.Path, "get", "200")
				gotTypes := stringSlice(frozenSchema["type"])
				if !reflect.DeepEqual(gotTypes, c.Expected.FrozenTypes) {
					t.Fatalf("frozen %s type = %#v, want %#v", c.Input.Path, frozenSchema["type"], c.Expected.FrozenTypes)
				}

			default:
				t.Fatalf("unknown case kind %q", c.Input.Kind)
			}
		})
	}
}

func TestBuildVillageAPISpec_NoNullableTopLevelArraySuccessResponses(t *testing.T) {
	spec, err := BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	if spec.Paths == nil {
		t.Fatal("spec has no paths")
	}
	for path, item := range spec.Paths.MapOfPathItemValues {
		for method, op := range map[string]*openapi31.Operation{
			"GET": item.Get, "POST": item.Post, "PUT": item.Put, "PATCH": item.Patch, "DELETE": item.Delete,
		} {
			if op == nil || op.Responses == nil {
				continue
			}
			for status, respOrRef := range op.Responses.MapOfResponseOrReferenceValues {
				if !strings.HasPrefix(status, "2") || respOrRef.Response == nil {
					continue
				}
				media, ok := respOrRef.Response.Content[villageMediaJSON]
				if !ok {
					continue
				}
				if media.Schema == nil {
					continue
				}
				assertNoNullableArray(t, method+" "+path+" "+status, media.Schema)
			}
		}
	}
}

func assertNoNullableArray(t *testing.T, where string, schemaMap map[string]interface{}) {
	t.Helper()
	if types, ok := schemaMap["type"].([]interface{}); ok {
		for _, typ := range types {
			if typ == "array" {
				t.Fatalf("%s has nullable top-level array type %#v", where, types)
			}
		}
	}
	for _, key := range []string{"anyOf", "oneOf"} {
		arms, ok := schemaMap[key].([]interface{})
		if !ok {
			continue
		}
		for _, rawArm := range arms {
			if arm, ok := rawArm.(map[string]interface{}); ok {
				assertNoNullableArray(t, where+"."+key, arm)
			}
		}
	}
}

func responseJSONSchema(t *testing.T, spec *openapi31.Spec, path, method, status string) map[string]interface{} {
	t.Helper()
	item, ok := declaredPathItem(spec, path)
	if !ok {
		t.Fatalf("missing path %s", path)
	}
	op := pathOperation(item, method)
	if op == nil || op.Responses == nil {
		t.Fatalf("missing %s %s", method, path)
	}
	resp, ok := op.Responses.MapOfResponseOrReferenceValues[status]
	if !ok || resp.Response == nil {
		t.Fatalf("missing %s %s status %s", method, path, status)
	}
	media, ok := resp.Response.Content[villageMediaJSON]
	if !ok {
		t.Fatalf("missing %s %s status %s %s", method, path, status, villageMediaJSON)
	}
	if media.Schema == nil {
		t.Fatalf("missing schema on %s %s status %s %s", method, path, status, villageMediaJSON)
	}
	return media.Schema
}

func resolveSchemaRef(t *testing.T, spec *openapi31.Spec, schemaMap map[string]interface{}) map[string]interface{} {
	t.Helper()
	ref, ok := schemaMap["$ref"].(string)
	if !ok {
		return schemaMap
	}
	name := strings.TrimPrefix(ref, "#/components/schemas/")
	resolved, ok := spec.Components.Schemas[name]
	if !ok {
		t.Fatalf("missing component schema %s", name)
	}
	return resolved
}

func rawDocResponseSchema(t *testing.T, doc map[string]interface{}, path, method, status string) map[string]interface{} {
	t.Helper()
	paths, _ := doc["paths"].(map[string]interface{})
	item, _ := paths[path].(map[string]interface{})
	op, _ := item[method].(map[string]interface{})
	responses, _ := op["responses"].(map[string]interface{})
	resp, _ := responses[status].(map[string]interface{})
	content, _ := resp["content"].(map[string]interface{})
	media, _ := content[villageMediaJSON].(map[string]interface{})
	schemaMap, _ := media["schema"].(map[string]interface{})
	return schemaMap
}

func stringSlice(raw interface{}) []string {
	switch v := raw.(type) {
	case []interface{}:
		out := make([]string, len(v))
		for i, item := range v {
			out[i], _ = item.(string)
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}
