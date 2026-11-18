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

// goneDocument is the part of an OpenAPI document this test reads: each
// operation's id and the statuses it declares, with the schema of each body.
type goneDocument struct {
	Paths map[string]goneEntry `json:"paths"`
}

type goneEntry struct {
	Get    *goneOperation `json:"get"`
	Put    *goneOperation `json:"put"`
	Post   *goneOperation `json:"post"`
	Patch  *goneOperation `json:"patch"`
	Delete *goneOperation `json:"delete"`
}

type goneOperation struct {
	OperationID string                  `json:"operationId"`
	Responses   map[string]goneResponse `json:"responses"`
}

type goneResponse struct {
	Description string `json:"description"`
	Content     map[string]struct {
		Schema struct {
			Ref string `json:"$ref"`
		} `json:"schema"`
	} `json:"content"`
}

func (e goneEntry) operations() []*goneOperation {
	var operations []*goneOperation
	for _, operation := range []*goneOperation{e.Get, e.Put, e.Post, e.Patch, e.Delete} {
		if operation != nil {
			operations = append(operations, operation)
		}
	}
	return operations
}

func decodeGoneDocument(t *testing.T, raw []byte) goneDocument {
	t.Helper()
	var document goneDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode the Village API document: %v", err)
	}
	return document
}

func builtGoneDocument(t *testing.T) goneDocument {
	t.Helper()
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	return decodeGoneDocument(t, raw)
}

// operationsDeclaring410 maps the operation id of every operation that
// declares a 410 response to its path.
func (d goneDocument) operationsDeclaring410() map[string]string {
	found := map[string]string{}
	for path, entry := range d.Paths {
		for _, operation := range entry.operations() {
			if _, declares := operation.Responses["410"]; declares {
				found[operation.OperationID] = path
			}
		}
	}
	return found
}

func (d goneDocument) operation(t *testing.T, in villageAbsenteeGoneInput) *goneOperation {
	t.Helper()
	entry, ok := d.Paths[in.Path]
	if !ok {
		t.Fatalf("the Village API declares no path %s", in.Path)
	}
	for _, operation := range entry.operations() {
		if operation.OperationID == in.OperationID {
			return operation
		}
	}
	t.Fatalf("path %s declares no operation %s", in.Path, in.OperationID)
	return nil
}

// TestVillageAbsenteeGoneIsDeclared holds the six transcript reads to a 410
// that carries the shared error envelope and its copy, holds the pull
// operations to no 410, and holds the whole document to exactly those six
// operations. The check runs against the document the generator builds and
// against the document the module serves, so a stale embedded copy fails too.
func TestVillageAbsenteeGoneIsDeclared(t *testing.T) {
	fixtures := loadVillageAbsenteeGoneFixtures(t)
	documents := map[string]goneDocument{
		"built":  builtGoneDocument(t),
		"served": decodeGoneDocument(t, schema.VillageAPISpecJSON()),
	}
	for source, document := range documents {
		t.Run(source, func(t *testing.T) {
			for _, c := range fixtures.Declared.Cases {
				t.Run(c.Name, func(t *testing.T) {
					operation := document.operation(t, c.Input)
					response, ok := operation.Responses["410"]
					if !ok {
						t.Fatalf("%s declares no 410; statuses are %v", c.Input.OperationID, responseStatuses(operation.Responses))
					}
					if got := response.Content["application/json"].Schema.Ref; got != "#/components/schemas/SchemaVillageErrorResponse" {
						t.Fatalf("%s 410 body is %q, want the shared Village error envelope", c.Input.OperationID, got)
					}
					for _, phrase := range c.Expected.Contains {
						if !strings.Contains(response.Description, phrase) {
							t.Fatalf("%s 410 description %q does not contain %q", c.Input.OperationID, response.Description, phrase)
						}
					}
				})
			}
			for _, c := range fixtures.Absent.Cases {
				t.Run(c.Name, func(t *testing.T) {
					if _, declared := document.operation(t, c.Input).Responses["410"]; declared {
						t.Fatalf("%s declares 410; this operation must keep its 404", c.Input.OperationID)
					}
				})
			}
			t.Run("exact-set-and-no-pull-route", func(t *testing.T) {
				var want []string
				for _, c := range fixtures.Declared.Cases {
					want = append(want, c.Input.OperationID)
				}
				var got []string
				for id, path := range document.operationsDeclaring410() {
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

func responseStatuses(responses map[string]goneResponse) []string {
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		statuses = append(statuses, status)
	}
	slices.Sort(statuses)
	return statuses
}
