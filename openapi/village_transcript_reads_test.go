package openapi_test

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

type villageTranscriptReadInput struct {
	Path        string `yaml:"path"`
	OperationID string `yaml:"operationId"`
}

type villageTranscriptReadExpected struct {
	Tag        string `yaml:"tag"`
	Parameters []struct {
		Name     string `yaml:"name"`
		In       string `yaml:"in"`
		Type     string `yaml:"type"`
		Required bool   `yaml:"required"`
		Minimum  *int   `yaml:"minimum"`
		Maximum  *int   `yaml:"maximum"`
	} `yaml:"parameters"`
	ResponseRef   string   `yaml:"responseRef"`
	RequiredRoot  []string `yaml:"requiredRoot"`
	ForbiddenRoot []string `yaml:"forbiddenRoot"`
	NonnullArrays []string `yaml:"nonnullArrays"`
	Statuses      []int    `yaml:"statuses"`
	ErrorRef      string   `yaml:"errorRef"`
}

type villageTranscriptReadFixtures struct {
	RequiredNames []string                                                                   `yaml:"required_names"`
	Routes        testcase.Corpus[villageTranscriptReadInput, villageTranscriptReadExpected] `yaml:"routes"`
}

// TestVillageTranscriptReadsAreDeclared holds the Village API's directory,
// segment and window reads to the fixture: path, operation id, tag, ordered
// parameters, the response component and the exact status set. It also holds
// each response component equal to the one the Local API declares, so the two
// servers cannot serve different shapes under one type name.
func TestVillageTranscriptReadsAreDeclared(t *testing.T) {
	data, err := os.ReadFile("testdata/village_transcript_reads.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures villageTranscriptReadFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	corpus := fixtures.Routes
	testassert.RequireMin(t, corpus, 3)
	testassert.RequireValid(t, corpus)
	names := make([]string, len(corpus.Cases))
	for i, c := range corpus.Cases {
		names[i] = c.Name
	}
	if !slices.Equal(names, fixtures.RequiredNames) {
		t.Fatalf("village transcript read cases are %v, want the required names %v", names, fixtures.RequiredNames)
	}

	village := builtSpecDocument(t, specpkg.BuildVillageAPISpec)
	local := builtSpecDocument(t, specpkg.BuildPeasantLocalAPISpec)
	paths := village["paths"].(map[string]any)
	components := village["components"].(map[string]any)["schemas"].(map[string]any)
	localComponents := local["components"].(map[string]any)["schemas"].(map[string]any)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			item, ok := paths[c.Input.Path].(map[string]any)
			if !ok {
				t.Fatalf("the Village API declares no path %s", c.Input.Path)
			}
			if len(item) != 1 {
				t.Fatalf("path %s declares %d methods, want only get", c.Input.Path, len(item))
			}
			op, ok := item["get"].(map[string]any)
			if !ok {
				t.Fatalf("path %s declares no get", c.Input.Path)
			}
			if op["operationId"] != c.Input.OperationID {
				t.Fatalf("operationId=%v, want %s", op["operationId"], c.Input.OperationID)
			}
			if tags, _ := op["tags"].([]any); len(tags) != 1 || tags[0] != c.Expected.Tag {
				t.Fatalf("tags=%v, want [%s]", op["tags"], c.Expected.Tag)
			}

			params, _ := op["parameters"].([]any)
			if len(params) != len(c.Expected.Parameters) {
				t.Fatalf("parameters=%d, want %d: %v", len(params), len(c.Expected.Parameters), params)
			}
			for index, expected := range c.Expected.Parameters {
				p := params[index].(map[string]any)
				if p["name"] != expected.Name || p["in"] != expected.In {
					t.Fatalf("parameters[%d]=%v/%v, want %q/%q", index, p["name"], p["in"], expected.Name, expected.In)
				}
				if (p["required"] == true) != expected.Required {
					t.Fatalf("parameter %q required=%v, want %v", expected.Name, p["required"], expected.Required)
				}
				shape := p["schema"].(map[string]any)
				if ref, ok := shape["$ref"].(string); ok {
					shape = components[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
				}
				if shape["type"] != expected.Type {
					t.Fatalf("parameter %q type=%v, want %q", expected.Name, shape["type"], expected.Type)
				}
				for bound, want := range map[string]*int{"minimum": expected.Minimum, "maximum": expected.Maximum} {
					got, declared := shape[bound]
					if want == nil && declared {
						t.Fatalf("parameter %q declares %s=%v, want none", expected.Name, bound, got)
					}
					if want != nil && got != float64(*want) {
						t.Fatalf("parameter %q %s=%v, want %d", expected.Name, bound, got, *want)
					}
				}
			}

			responses := op["responses"].(map[string]any)
			statuses := make([]int, 0, len(responses))
			for code := range responses {
				status, err := strconv.Atoi(code)
				if err != nil {
					t.Fatalf("response status %q is not numeric", code)
				}
				statuses = append(statuses, status)
			}
			slices.Sort(statuses)
			if !slices.Equal(statuses, c.Expected.Statuses) {
				t.Fatalf("statuses=%v, want %v", statuses, c.Expected.Statuses)
			}
			for _, status := range statuses {
				ref := declaredResponseRef(t, responses, status)
				want := c.Expected.ErrorRef
				if status == 200 {
					want = c.Expected.ResponseRef
				}
				if ref != want {
					t.Fatalf("status %d body=%s, want %s", status, ref, want)
				}
			}

			name := strings.TrimPrefix(c.Expected.ResponseRef, "#/components/schemas/")
			component, ok := components[name].(map[string]any)
			if !ok {
				t.Fatalf("response ref %q does not resolve", c.Expected.ResponseRef)
			}
			properties, _ := component["properties"].(map[string]any)
			for _, root := range c.Expected.RequiredRoot {
				if _, ok := properties[root]; !ok {
					t.Fatalf("response component is missing %q", root)
				}
			}
			for _, root := range c.Expected.ForbiddenRoot {
				if _, ok := properties[root]; ok {
					t.Fatalf("response component carries %q", root)
				}
			}
			for _, root := range c.Expected.NonnullArrays {
				if properties[root].(map[string]any)["type"] != "array" {
					t.Fatalf("%s is not a non-null array", root)
				}
			}
			// The two specs name a nested component with or without the Schema
			// prefix, so the prefix is dropped from references before comparing.
			unprefixed := func(value any) []byte {
				encoded, _ := json.Marshal(value)
				return bytes.ReplaceAll(encoded, []byte(`"#/components/schemas/Schema`), []byte(`"#/components/schemas/`))
			}
			villageJSON := unprefixed(component)
			localJSON := unprefixed(localComponents[name])
			if !bytes.Equal(villageJSON, localJSON) {
				t.Fatalf("component %s differs between the Village and Local specs:\nvillage %s\nlocal   %s", name, villageJSON, localJSON)
			}
		})
	}
}

func builtSpecDocument[T any](t *testing.T, build func() (T, error)) map[string]any {
	t.Helper()
	spec, err := build()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func declaredResponseRef(t *testing.T, responses map[string]any, status int) string {
	t.Helper()
	response := responses[strconv.Itoa(status)].(map[string]any)
	content, ok := response["content"].(map[string]any)
	if !ok {
		t.Fatalf("status %d declares no body", status)
	}
	ref, ok := content["application/json"].(map[string]any)["schema"].(map[string]any)["$ref"].(string)
	if !ok {
		t.Fatalf("status %d body is not a component reference", status)
	}
	return ref
}
