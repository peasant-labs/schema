package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/cross_origin_writes.yaml
var crossOriginWritesYAML []byte

type crossOriginWritesFixture struct {
	WriteOperations []writeOperationIdentity                    `yaml:"write_operations"`
	Refusal         crossOriginRefusal                          `yaml:"refusal"`
	Mutations       testcase.Corpus[writeSurfaceMutation, bool] `yaml:"mutations"`
}

type writeOperationIdentity struct {
	Method      string `yaml:"method"`
	Path        string `yaml:"path"`
	OperationID string `yaml:"operation_id"`
}

type crossOriginRefusal struct {
	Status            string   `yaml:"status"`
	Component         string   `yaml:"component"`
	Required          []string `yaml:"required"`
	Properties        []string `yaml:"properties"`
	DescriptionAnchor string   `yaml:"description_anchor"`
}

type writeSurfaceMutation struct {
	Kind   string `yaml:"kind"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

// writeSurface maps "method path" to the declared operation.
type writeSurface map[string]declaredWrite

type declaredWrite struct {
	operationID string
	refused     bool
}

func loadCrossOriginWrites(t *testing.T) crossOriginWritesFixture {
	t.Helper()
	var fixture crossOriginWritesFixture
	decoder := yaml.NewDecoder(bytes.NewReader(crossOriginWritesYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode cross-origin write fixture: %v", err)
	}
	assertcase.RequireMin(t, fixture.Mutations, 5)
	assertcase.RequireValid(t, fixture.Mutations)
	if len(fixture.WriteOperations) == 0 {
		t.Fatal("cross-origin write fixture names no write operation")
	}
	return fixture
}

// declaredWriteSurface reads every state-changing operation of the generated
// Local API document and whether it declares the shared refusal.
func declaredWriteSurface(t *testing.T, refusal crossOriginRefusal) (writeSurface, map[string]any) {
	t.Helper()
	spec, err := openapi.BuildPeasantLocalAPISpec()
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
	surface := writeSurface{}
	for path, rawItem := range document["paths"].(map[string]any) {
		for method, rawOperation := range rawItem.(map[string]any) {
			switch method {
			case "post", "put", "patch", "delete":
			default:
				continue
			}
			operation := rawOperation.(map[string]any)
			responses, _ := operation["responses"].(map[string]any)
			response, _ := responses[refusal.Status].(map[string]any)
			content, _ := response["content"].(map[string]any)
			media, _ := content["application/json"].(map[string]any)
			schemaMap, _ := media["schema"].(map[string]any)
			description, _ := operation["description"].(string)
			refused := schemaMap["$ref"] == "#/components/schemas/"+refusal.Component && strings.Contains(description, refusal.DescriptionAnchor)
			surface[method+" "+path] = declaredWrite{operationID: fmt.Sprint(operation["operationId"]), refused: refused}
		}
	}
	return surface, document
}

func validateWriteSurface(actual writeSurface, manifest []writeOperationIdentity) error {
	named := map[string]bool{}
	for _, want := range manifest {
		switch want.Method {
		case "post", "put", "patch", "delete":
		default:
			return fmt.Errorf("%s %s is not a write method; reads change nothing and are not refused", want.Method, want.Path)
		}
		key := want.Method + " " + want.Path
		named[key] = true
		got, ok := actual[key]
		if !ok {
			return fmt.Errorf("write operation %s is not declared", key)
		}
		if got.operationID != want.OperationID {
			return fmt.Errorf("write operation %s has operationId %s, want %s", key, got.operationID, want.OperationID)
		}
		if !got.refused {
			return fmt.Errorf("write operation %s does not declare the cross-origin refusal", key)
		}
	}
	var unaccounted []string
	for key := range actual {
		if !named[key] {
			unaccounted = append(unaccounted, key)
		}
	}
	sort.Strings(unaccounted)
	if len(unaccounted) > 0 {
		return fmt.Errorf("write operations outside the manifest: %v", unaccounted)
	}
	return nil
}

func TestLocalWriteRoutesDeclareCrossOriginRefusal(t *testing.T) {
	fixture := loadCrossOriginWrites(t)
	surface, document := declaredWriteSurface(t, fixture.Refusal)
	if err := validateWriteSurface(surface, fixture.WriteOperations); err != nil {
		t.Fatal(err)
	}
	component, ok := document["components"].(map[string]any)["schemas"].(map[string]any)[fixture.Refusal.Component].(map[string]any)
	if !ok {
		t.Fatalf("refusal component %s is not declared", fixture.Refusal.Component)
	}
	requireSameSet(t, "refusal required fields", fixture.Refusal.Required, anyStrings(component["required"]))
	properties := make([]string, 0)
	for name := range component["properties"].(map[string]any) {
		properties = append(properties, name)
	}
	requireSameSet(t, "refusal properties", fixture.Refusal.Properties, properties)
}

func TestLocalWriteRefusalMutationProof(t *testing.T) {
	fixture := loadCrossOriginWrites(t)
	canonical, _ := declaredWriteSurface(t, fixture.Refusal)
	for _, c := range fixture.Mutations.Cases {
		t.Run(c.Name, func(t *testing.T) {
			surface := writeSurface{}
			for key, value := range canonical {
				surface[key] = value
			}
			manifest := append([]writeOperationIdentity(nil), fixture.WriteOperations...)
			key := c.Input.Method + " " + c.Input.Path
			switch c.Input.Kind {
			case "none":
			case "drop_refusal":
				write := surface[key]
				write.refused = false
				surface[key] = write
			case "add_route":
				surface[key] = declaredWrite{operationID: "unexpected", refused: true}
			case "remove_route":
				delete(surface, key)
			case "read_route":
				manifest = append(manifest, writeOperationIdentity{Method: c.Input.Method, Path: c.Input.Path, OperationID: "read"})
				surface[key] = declaredWrite{operationID: "read", refused: true}
			default:
				t.Fatalf("unknown write surface mutation %q", c.Input.Kind)
			}
			accepted := validateWriteSurface(surface, manifest) == nil
			if accepted != c.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, c.Expected)
			}
		})
	}
}

func anyStrings(raw any) []string {
	values, _ := raw.([]any)
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, fmt.Sprint(value))
	}
	return out
}
