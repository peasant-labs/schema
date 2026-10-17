package openapi_test

import (
	_ "embed"
	"encoding/json"
	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
	"testing"
)

//go:embed testdata/collective_observation_routes.yaml
var observationRoutesYAML []byte

type observationRoute struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Method   string   `yaml:"method"`
	ID       string   `yaml:"id"`
	Statuses []string `yaml:"statuses"`
	Body     string   `yaml:"body"`
}
type observationRoutes struct {
	RequiredNames   []string           `yaml:"required_names"`
	Local           []observationRoute `yaml:"local"`
	AliasOperations []string           `yaml:"alias_operations"`
}

func loadObservationRoutes(t *testing.T) observationRoutes {
	t.Helper()
	var f observationRoutes
	if err := yaml.Unmarshal(observationRoutesYAML, &f); err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, r := range f.Local {
		names = append(names, r.Name)
	}
	assertSameStringSet(t, "observation route manifest", names, f.RequiredNames)
	assertSameStringSet(t, "required observation routes", f.RequiredNames, strings.Fields("offline-catalog bounded-evidence metadata-only-fetch explicit-track offline-untrack"))
	assertSameStringSet(t, "required compatibility operations", f.AliasOperations, strings.Fields("searchCollectives listPublicGroups createGroup listGroups listVisibleGroups getGroup updateGroup deleteGroup joinGroup addGroupMember updateGroupMemberRole removeGroupMember listContributableTranscripts batchShareProject listPendingShares listMyGroupShares batchReviewShares reviewShare removeGroupTranscript listGroupRepositories linkGroupRepository listAvailableGroupRepositories unlinkGroupRepository listGroupRepositoryCommits"))
	return f
}
func TestCollectiveObservationRoutes(t *testing.T) {
	f := loadObservationRoutes(t)
	spec, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]interface{}
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]interface{})
	if _, exists := paths["/api/v1/village/collective-observations/refresh"]; exists {
		t.Fatal("unshipped refresh route must not exist")
	}
	for _, r := range f.Local {
		t.Run(r.Name, func(t *testing.T) {
			item, ok := paths[r.Path].(map[string]interface{})
			if !ok {
				t.Fatalf("missing path %s", r.Path)
			}
			op, ok := item[r.Method].(map[string]interface{})
			if !ok || op["operationId"] != r.ID {
				t.Fatalf("missing/mismatched operation %s", r.ID)
			}
			if _, body := op["requestBody"]; body {
				t.Fatal("metadata operation must have no body")
			}
			responses := op["responses"].(map[string]interface{})
			statuses := []string{}
			for s := range responses {
				statuses = append(statuses, s)
			}
			assertSameStringSet(t, r.ID, statuses, r.Statuses)
			for status, raw := range responses {
				response := raw.(map[string]interface{})
				media := response["content"].(map[string]interface{})["application/json"].(map[string]interface{})
				body := media["schema"].(map[string]interface{})
				if r.Name == "metadata-only-fetch" && status == "401" {
					alternatives, ok := body["oneOf"].([]interface{})
					if !ok {
						t.Fatal("auth refusal must retain typed scoped reports and truthful no-scope errors")
					}
					refs := []string{}
					for _, raw := range alternatives {
						refs = append(refs, raw.(map[string]interface{})["$ref"].(string))
					}
					assertSameStringSet(t, "initial or mid-invocation auth refusal", refs, []string{"#/components/schemas/SchemaCollectiveFetchReport", "#/components/schemas/OpenapiLocalErrorResponse"})
					continue
				}
				want := "OpenapiLocalErrorResponse"
				if status == "200" || (r.Name == "metadata-only-fetch" && status != "403") {
					want = "Schema" + r.Body
				}
				if r.Name == "bounded-evidence" && status == "409" {
					want = "SchemaLocalContributionTraversalConflict"
				}
				if body["$ref"] != "#/components/schemas/"+want {
					t.Fatalf("%s status %s body=%v want %s", r.Name, status, body, want)
				}
			}
		})
	}
}
func TestCollectiveAliasesPreserveErrorAndQueryParity(t *testing.T) {
	f := loadObservationRoutes(t)
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]interface{}
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]interface{})
	names := []string{}
	for path, raw := range paths {
		if path != "/api/v1/collectives" && !strings.HasPrefix(path, "/api/v1/collectives/") {
			continue
		}
		canonical := raw.(map[string]interface{})
		legacy, ok := paths[strings.Replace(path, "/collectives", "/groups", 1)].(map[string]interface{})
		if !ok {
			t.Fatalf("alias missing for %s", path)
		}
		for method, rawOp := range canonical {
			if method == "parameters" {
				continue
			}
			op, ok := rawOp.(map[string]interface{})
			if !ok || op["operationId"] == nil {
				continue
			}
			id := op["operationId"].(string)
			names = append(names, id)
			old, ok := legacy[method].(map[string]interface{})
			if !ok || old["operationId"] != id+"Deprecated" || old["deprecated"] != true {
				t.Fatalf("deprecated operation mismatch %s", id)
			}
			delete(old, "operationId")
			delete(old, "deprecated")
			if id == "listContributableTranscripts" {
				// Only the grouped response spelling differs during the alias window.
				grouped := func(operation map[string]interface{}) map[string]interface{} {
					return operation["responses"].(map[string]interface{})["200"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})["oneOf"].([]interface{})[1].(map[string]interface{})
				}
				legacyGrouped, canonicalGrouped := grouped(old), grouped(op)
				if legacyGrouped["$ref"] != "#/components/schemas/SchemaVillageDeprecatedGroupedContributableResponse" || canonicalGrouped["$ref"] != "#/components/schemas/SchemaVillageGroupedContributableResponse" {
					t.Fatal("grouped contribution route families lost their distinct declared envelopes")
				}
				legacyGrouped["$ref"] = canonicalGrouped["$ref"]
			}
			copy := map[string]interface{}{}
			for k, v := range op {
				if k != "operationId" {
					copy[k] = v
				}
			}
			if !reflect.DeepEqual(old, copy) {
				t.Fatalf("canonical/alias request, response or error drift at %s", id)
			}
		}
	}
	assertSameStringSet(t, "complete alias operation inventory", names, f.AliasOperations)
	for name := range spec.Components.Schemas {
		if strings.HasPrefix(name, "SchemaVillageGroup") && name != "SchemaVillageGroupedContributableResponse" && name != "SchemaVillageGroupedCollectiveDetailResponse" {
			t.Fatalf("historical component leaked: %s", name)
		}
	}
}
func TestGroupedCollectiveIdentifierUsesCanonicalKey(t *testing.T) {
	data, err := json.Marshal(schema.VillageGroupedContributableResponse{})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if _, old := fields["groupId"]; old {
		t.Fatal("grouped contributable identifier retained incorrect key")
	}
	if _, ok := fields["group_id"]; !ok {
		t.Fatal("group_id is missing")
	}
}
