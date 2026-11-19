package openapi_test

import (
	"bytes"
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

type transcriptHeaderReadInput struct {
	Path        string `yaml:"path"`
	OperationID string `yaml:"operationId"`
}

type transcriptHeaderReadExpected struct {
	ResponseRef  string   `yaml:"responseRef"`
	Statuses     []int    `yaml:"statuses"`
	Parameters   []string `yaml:"parameters"`
	HeaderFields []string `yaml:"headerFields"`
}

type transcriptHeaderReadFixtures struct {
	RequiredNames []string                                                                 `yaml:"requiredNames"`
	Cases         []testcase.Case[transcriptHeaderReadInput, transcriptHeaderReadExpected] `yaml:"cases"`
}

func TestVillageTranscriptHeaderReadIsDeclared(t *testing.T) {
	data, err := os.ReadFile("testdata/transcript_header_read.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var f transcriptHeaderReadFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		t.Fatal(err)
	}
	corpus := testcase.Corpus[transcriptHeaderReadInput, transcriptHeaderReadExpected]{Cases: f.Cases}
	testassert.RequireValid(t, corpus)
	var names []string
	for _, c := range f.Cases {
		names = append(names, c.Name)
	}
	if !slices.Equal(names, f.RequiredNames) {
		t.Fatalf("header route cases = %v, want %v", names, f.RequiredNames)
	}
	document := builtSpecDocument(t, specpkg.BuildVillageAPISpec)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)["schemas"].(map[string]any)
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			path, ok := paths[c.Input.Path].(map[string]any)
			if !ok {
				t.Fatalf("missing header read at %s", c.Input.Path)
			}
			op, ok := path["get"].(map[string]any)
			if !ok || op["operationId"] != c.Input.OperationID {
				t.Fatalf("header operation = %v, want %s", path["get"], c.Input.OperationID)
			}
			params, _ := op["parameters"].([]any)
			var gotParams []string
			for _, raw := range params {
				p := raw.(map[string]any)
				gotParams = append(gotParams, p["in"].(string)+":"+p["name"].(string))
				if p["in"] == "path" && p["required"] != true {
					t.Fatal("header route identity must be required")
				}
			}
			slices.Sort(gotParams)
			wantParams := slices.Clone(c.Expected.Parameters)
			slices.Sort(wantParams)
			if !slices.Equal(gotParams, wantParams) {
				t.Fatalf("parameters = %v, want %v", gotParams, wantParams)
			}
			responses := op["responses"].(map[string]any)
			var statuses []int
			for code := range responses {
				status, err := strconv.Atoi(code)
				if err != nil {
					t.Fatal(err)
				}
				statuses = append(statuses, status)
			}
			slices.Sort(statuses)
			if !slices.Equal(statuses, c.Expected.Statuses) {
				t.Fatalf("statuses = %v, want %v", statuses, c.Expected.Statuses)
			}
			if got := declaredResponseRef(t, responses, 200); got != c.Expected.ResponseRef {
				t.Fatalf("response = %s, want %s", got, c.Expected.ResponseRef)
			}
			for _, status := range statuses {
				response := responses[strconv.Itoa(status)].(map[string]any)
				headers, _ := response["headers"].(map[string]any)
				if headers["Cache-Control"] == nil || headers["Vary"] == nil {
					t.Fatalf("status %d omits cache or vary headers", status)
				}
				if (status == 200 || status == 304) && headers["ETag"] == nil {
					t.Fatalf("status %d omits ETag", status)
				}
				if status == 503 && headers["Retry-After"] == nil {
					t.Fatal("503 omits Retry-After")
				}
				if status == 304 && response["content"] != nil {
					t.Fatal("304 must not declare a body")
				}
			}
			name := strings.TrimPrefix(c.Expected.ResponseRef, "#/components/schemas/")
			component, ok := components[name].(map[string]any)
			if !ok {
				t.Fatalf("missing response component %s", name)
			}
			properties := component["properties"].(map[string]any)
			for _, field := range c.Expected.HeaderFields {
				if properties[field] == nil {
					t.Fatalf("header component lacks %s", field)
				}
			}
		})
	}
}
