package openapi_test

import (
	"bytes"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

// Both APIs project the same response alternatives and selector parameters.
func TestLocalTranscriptPositionedReads(t *testing.T) {
	data, err := os.ReadFile("testdata/village_transcript_reads.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var fixture villageTranscriptReadFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	testassert.RequireMin(t, fixture.Routes, 3)
	testassert.RequireValid(t, fixture.Routes)
	names := make([]string, 0, len(fixture.Routes.Cases))
	local := builtSpecDocument(t, specpkg.BuildPeasantLocalAPISpec)
	paths := local["paths"].(map[string]any)
	components := local["components"].(map[string]any)["schemas"].(map[string]any)
	for _, c := range fixture.Routes.Cases {
		names = append(names, c.Name)
		t.Run(c.Name, func(t *testing.T) {
			path := strings.Replace(c.Input.Path, "/api/v1/transcripts/{id}", "/api/v1/sessions/{id}/transcript", 1)
			op := paths[path].(map[string]any)["get"].(map[string]any)
			parameters := op["parameters"].([]any)
			for _, expected := range c.Expected.Parameters {
				found := false
				for _, value := range parameters {
					param := value.(map[string]any)
					if param["name"] != expected.Name || param["in"] != expected.In {
						continue
					}
					found = true
					if (param["required"] == true) != expected.Required {
						t.Fatalf("%s required differs from the shared declaration", expected.Name)
					}
					if expected.Name == "v" {
						values := param["schema"].(map[string]any)["enum"].([]any)
						expectedGenerations := schema.SegmentedReadGenerations()
						if len(values) != len(expectedGenerations) {
							t.Fatalf("generation values: %v", values)
						}
						for i, generation := range expectedGenerations {
							if values[i] != float64(generation) {
								t.Fatalf("generation values: %v", values)
							}
						}
					}
				}
				if !found {
					t.Fatalf("missing %s parameter %s", expected.In, expected.Name)
				}
			}
			responses := op["responses"].(map[string]any)
			success := responses["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)["oneOf"].([]any)
			if len(success) != 2 || success[0].(map[string]any)["$ref"] != c.Expected.ResponseRef || success[1].(map[string]any)["$ref"] != c.Expected.PositionedResponseRef {
				t.Fatalf("unexpected response alternatives: %v", success)
			}
			for _, ref := range success {
				name := strings.TrimPrefix(ref.(map[string]any)["$ref"].(string), "#/components/schemas/")
				if components[name] == nil {
					t.Fatalf("response component %s is missing", name)
				}
			}
			if len(responses) != len(c.Expected.LocalStatuses) {
				t.Fatalf("unexpected statuses: %v", responses)
			}
			for _, code := range c.Expected.LocalStatuses {
				status := strconv.Itoa(code)
				if responses[status] == nil {
					t.Fatalf("missing status %s", status)
				}
			}
			headers := responses["503"].(map[string]any)["headers"].(map[string]any)
			if headers["Retry-After"] == nil {
				t.Fatal("503 omits Retry-After")
			}
		})
	}
	if !slices.Equal(names, fixture.RequiredNames) {
		t.Fatalf("cases %v do not match manifest %v", names, fixture.RequiredNames)
	}
}
