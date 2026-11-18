package openapi_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_absentee_gone.yaml
var villageAbsenteeGoneYAML []byte

//go:embed testdata/village_absentee_gone_manifest.yaml
var villageAbsenteeGoneManifestYAML []byte

type villageAbsenteeGoneInput struct {
	OperationID string `yaml:"operationId"`
	Path        string `yaml:"path"`
}

type villageAbsenteeGoneExpected struct {
	Contains []string `yaml:"contains"`
}

type villageAbsenteeGoneFixtures struct {
	Declared testcase.Corpus[villageAbsenteeGoneInput, villageAbsenteeGoneExpected] `yaml:"declared"`
	Absent   testcase.Corpus[villageAbsenteeGoneInput, villageAbsenteeGoneExpected] `yaml:"absent"`
}

func loadVillageAbsenteeGoneFixtures(t *testing.T) villageAbsenteeGoneFixtures {
	t.Helper()
	var fixtures villageAbsenteeGoneFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(villageAbsenteeGoneYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	testassert.RequireMin(t, fixtures.Declared, 6)
	testassert.RequireValid(t, fixtures.Declared)
	testassert.RequireMin(t, fixtures.Absent, 1)
	testassert.RequireValid(t, fixtures.Absent)
	var names []string
	for _, c := range fixtures.Declared.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.Absent.Cases {
		names = append(names, c.Name)
	}
	testassert.RequireManifestNames(t, "village absentee 410 declarations", villageAbsenteeGoneManifestYAML, names)
	return fixtures
}

// villageOperationsDeclaring410 returns the operation ids of every operation in
// the document that declares a 410 response, keyed by operation id with the
// path as the value. It reads the document as served JSON, so it holds what a
// client of the document sees.
func villageOperationsDeclaring410(t *testing.T, document map[string]any) map[string]string {
	t.Helper()
	found := map[string]string{}
	paths, _ := document["paths"].(map[string]any)
	for path, rawItem := range paths {
		item, _ := rawItem.(map[string]any)
		for _, rawOperation := range item {
			operation, ok := rawOperation.(map[string]any)
			if !ok {
				continue
			}
			responses, _ := operation["responses"].(map[string]any)
			if _, declares := responses["410"]; declares {
				id, _ := operation["operationId"].(string)
				found[id] = path
			}
		}
	}
	return found
}

func decodeServedVillageDocument(t *testing.T) map[string]any {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(schema.VillageAPISpecJSON(), &document); err != nil {
		t.Fatalf("decode the served Village API document: %v", err)
	}
	return document
}

// TestVillageAbsenteeGoneIsDeclared holds the six transcript reads to a 410
// that carries the shared error envelope and its copy, holds the pull
// operations to no 410, and holds the whole document to exactly those six
// operations. The check runs against the document the generator builds and
// against the document the module serves, so a stale embedded copy fails too.
func TestVillageAbsenteeGoneIsDeclared(t *testing.T) {
	fixtures := loadVillageAbsenteeGoneFixtures(t)
	documents := map[string]map[string]any{
		"built":  builtSpecDocument(t, specpkg.BuildVillageAPISpec),
		"served": decodeServedVillageDocument(t),
	}
	for source, document := range documents {
		t.Run(source, func(t *testing.T) {
			paths := document["paths"].(map[string]any)
			operationFor := func(t *testing.T, in villageAbsenteeGoneInput) map[string]any {
				t.Helper()
				item, ok := paths[in.Path].(map[string]any)
				if !ok {
					t.Fatalf("the Village API declares no path %s", in.Path)
				}
				for _, raw := range item {
					if operation, ok := raw.(map[string]any); ok && operation["operationId"] == in.OperationID {
						return operation
					}
				}
				t.Fatalf("path %s declares no operation %s", in.Path, in.OperationID)
				return nil
			}
			for _, c := range fixtures.Declared.Cases {
				t.Run(c.Name, func(t *testing.T) {
					responses := operationFor(t, c.Input)["responses"].(map[string]any)
					response, ok := responses["410"].(map[string]any)
					if !ok {
						t.Fatalf("%s declares no 410; statuses are %v", c.Input.OperationID, responseStatuses(responses))
					}
					if got := declaredResponseRef(t, responses, 410); got != "#/components/schemas/SchemaVillageErrorResponse" {
						t.Fatalf("%s 410 body is %s, want the shared Village error envelope", c.Input.OperationID, got)
					}
					description, _ := response["description"].(string)
					for _, phrase := range c.Expected.Contains {
						if !strings.Contains(description, phrase) {
							t.Fatalf("%s 410 description %q does not contain %q", c.Input.OperationID, description, phrase)
						}
					}
				})
			}
			for _, c := range fixtures.Absent.Cases {
				t.Run(c.Name, func(t *testing.T) {
					responses := operationFor(t, c.Input)["responses"].(map[string]any)
					if _, declared := responses["410"]; declared {
						t.Fatalf("%s declares 410; this operation must keep its 404", c.Input.OperationID)
					}
				})
			}
			t.Run("exact-set-and-no-pull-route", func(t *testing.T) {
				var want []string
				for _, c := range fixtures.Declared.Cases {
					want = append(want, c.Input.OperationID)
				}
				found := villageOperationsDeclaring410(t, document)
				var got []string
				for id, path := range found {
					got = append(got, id)
					if strings.Contains(path, "/pulls") || strings.Contains(path, "/pull/") || strings.Contains(id, "Pull") {
						t.Errorf("%s at %s declares 410; no pull operation may", id, path)
					}
				}
				slices.Sort(got)
				slices.Sort(want)
				if !slices.Equal(got, want) {
					t.Fatalf("operations declaring 410 are %v, want exactly %v", got, want)
				}
			})
		})
	}
}

func responseStatuses(responses map[string]any) []string {
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		statuses = append(statuses, status)
	}
	slices.Sort(statuses)
	return statuses
}
