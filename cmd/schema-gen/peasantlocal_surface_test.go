package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type localAPISurfaceFixture struct {
	Operations []localAPIOperationIdentity                    `yaml:"operations"`
	Mutations  testcase.Corpus[localAPISurfaceMutation, bool] `yaml:"mutations"`
}

type localAPIOperationIdentity struct {
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	OperationID string `yaml:"operation_id"`
}

type localAPISurfaceMutation struct {
	Kind        string `yaml:"kind"`
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	OperationID string `yaml:"operation_id"`
}

func TestPeasantLocalGeneratedSurfaceHasExactIdentity(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	fixture := loadLocalAPISurfaceFixture(t, root)
	specPath := filepath.Join(root, "generated", "peasantlocal-api-"+schema.PeasantLocalAPIVersion+".json")
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := localAPIOperations(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateLocalAPIOperations(actual, fixture.Operations); err != nil {
		t.Fatal(err)
	}
}

func TestPeasantLocalGeneratedSurfaceMutationProof(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	fixture := loadLocalAPISurfaceFixture(t, root)
	canonical := make(map[string]string, len(fixture.Operations))
	for _, operation := range fixture.Operations {
		canonical[operation.Path+"#"+operation.Method] = operation.OperationID
	}
	for _, mutation := range fixture.Mutations.Cases {
		t.Run(mutation.Name, func(t *testing.T) {
			mutated := make(map[string]string, len(canonical)+1)
			for key, value := range canonical {
				mutated[key] = value
			}
			key := mutation.Input.Path + "#" + mutation.Input.Method
			switch mutation.Input.Kind {
			case "remove":
				delete(mutated, key)
			case "add", "redirect":
				mutated[key] = mutation.Input.OperationID
			default:
				t.Fatalf("unknown local API surface mutation %q", mutation.Input.Kind)
			}
			accepted := validateLocalAPIOperations(mutated, fixture.Operations) == nil
			if accepted != mutation.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, mutation.Expected)
			}
		})
	}
}

func loadLocalAPISurfaceFixture(t *testing.T, root string) localAPISurfaceFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "local_api_surface.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture localAPISurfaceFixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Mutations.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Mutations.CheckMin(3); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func localAPIOperations(spec []byte) (map[string]string, error) {
	var document struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(spec, &document); err != nil {
		return nil, fmt.Errorf("decode local API spec: %w", err)
	}
	operations := make(map[string]string)
	for path, methods := range document.Paths {
		for method, operation := range methods {
			if operation.OperationID == "" {
				continue
			}
			operations[path+"#"+strings.ToLower(method)] = operation.OperationID
		}
	}
	return operations, nil
}

func validateLocalAPIOperations(actual map[string]string, expected []localAPIOperationIdentity) error {
	want := make(map[string]string, len(expected))
	for _, operation := range expected {
		key := operation.Path + "#" + operation.Method
		if _, duplicate := want[key]; duplicate {
			return fmt.Errorf("local API fixture repeats operation %s", key)
		}
		want[key] = operation.OperationID
	}
	keys := func(values map[string]string) []string {
		out := make([]string, 0, len(values))
		for key := range values {
			out = append(out, key)
		}
		sort.Strings(out)
		return out
	}
	for _, key := range keys(want) {
		if actual[key] != want[key] {
			return fmt.Errorf("local API operation %s=%q, want %q; restore the fixture-accounted path and operation identity", key, actual[key], want[key])
		}
	}
	for _, key := range keys(actual) {
		if _, expected := want[key]; !expected {
			return fmt.Errorf("local API exposes unaccounted operation %s=%q; add an intentional fixture row or remove the stray operation", key, actual[key])
		}
	}
	return nil
}
