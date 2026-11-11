package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/host_refusal.yaml
var hostRefusalYAML []byte

type hostRefusalFixture struct {
	Status             string                   `yaml:"status"`
	Component          string                   `yaml:"component"`
	Code               string                   `yaml:"code"`
	RequiredOperations []string                 `yaml:"required_operations"`
	ReadOperations     []writeOperationIdentity `yaml:"read_operations"`
}

// Every read that declares the host refusal states the same status, body and
// code, and the set of such reads is exactly the fixture's. The code token is
// pinned here, not derived from the generator, so changing it in the Go source
// fails this test before any regeneration.
func TestLocalReadsDeclareTheHostRefusal(t *testing.T) {
	var fixture hostRefusalFixture
	decoder := yaml.NewDecoder(bytes.NewReader(hostRefusalYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode host refusal fixture: %v", err)
	}
	if fixture.Status == "" || fixture.Component == "" || fixture.Code == "" || len(fixture.ReadOperations) == 0 {
		t.Fatal("host refusal fixture is missing the status, component, code, or read operations")
	}
	names := make([]string, 0, len(fixture.ReadOperations))
	pinned := make([]string, 0, len(fixture.ReadOperations))
	for _, read := range fixture.ReadOperations {
		names = append(names, read.OperationID)
		pinned = append(pinned, read.Method+" "+read.Path)
	}
	requireExactNames(t, fixture.RequiredOperations, names)

	document := buildLocalDocument(t)
	var declared []string
	for path, rawItem := range document["paths"].(map[string]any) {
		operation, ok := rawItem.(map[string]any)["get"].(map[string]any)
		if !ok {
			continue
		}
		if responses, _ := operation["responses"].(map[string]any); responses[fixture.Status] != nil {
			declared = append(declared, "get "+path)
		}
	}
	requireSameSet(t, "GET operations that declare the host refusal and the fixture's read_operations", declared, pinned)

	for _, read := range fixture.ReadOperations {
		t.Run(read.OperationID, func(t *testing.T) {
			item, _ := document["paths"].(map[string]any)[read.Path].(map[string]any)
			operation, ok := item[read.Method].(map[string]any)
			if !ok {
				t.Fatalf("%s %s is not declared", read.Method, read.Path)
			}
			if got := fmt.Sprint(operation["operationId"]); got != read.OperationID {
				t.Fatalf("operationId = %s, want %s", got, read.OperationID)
			}
			if description, _ := operation["description"].(string); !strings.Contains(description, fixture.Code) {
				t.Errorf("%s operation description does not name the code %q: %q", read.OperationID, fixture.Code, description)
			}
			responses, _ := operation["responses"].(map[string]any)
			response, _ := responses[fixture.Status].(map[string]any)
			body := mediaSchema(t, read.OperationID+" "+fixture.Status, response)
			ref, _ := body["$ref"].(string)
			name := strings.TrimPrefix(ref, "#/components/schemas/")
			if name != fixture.Component && name != "Schema"+fixture.Component {
				t.Errorf("%s status %s body references %q, want %s", read.OperationID, fixture.Status, ref, fixture.Component)
			}
			if description, _ := response["description"].(string); !strings.Contains(description, fixture.Code) {
				t.Errorf("%s status %s description does not name the code %q: %q", read.OperationID, fixture.Status, fixture.Code, description)
			}
		})
	}
}
