package openapi_test

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase/assert"
)

const ownerUpdatePath = "/api/v1/transcripts/{id}"

// villageSpecAsMap renders the built Village document the way a consumer reads
// it. Asserting on the marshaled document rather than the reflector's in-memory
// structures means these probes see exactly the bytes that ship.
func villageSpecAsMap(t *testing.T) map[string]any {
	t.Helper()
	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("build village spec: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("encode village spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode village spec: %v", err)
	}
	return document
}

func ownerUpdateOperation(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		t.Fatal("village document has no paths object")
	}
	item, ok := paths[ownerUpdatePath].(map[string]any)
	if !ok {
		t.Fatalf("village document declares no path %q; the owner update operation must be reachable from the published contract", ownerUpdatePath)
	}
	operation, ok := item["patch"].(map[string]any)
	if !ok {
		t.Fatalf("path %q declares no patch operation", ownerUpdatePath)
	}
	return operation
}

// resolveRef follows a local component reference to its schema object.
func resolveRef(t *testing.T, document map[string]any, ref string) map[string]any {
	t.Helper()
	const prefix = "#/components/schemas/"
	name, found := trimPrefix(ref, prefix)
	if !found {
		t.Fatalf("reference %q is not a local component reference", ref)
	}
	components, ok := document["components"].(map[string]any)
	if !ok {
		t.Fatal("village document has no components object")
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		t.Fatal("village document has no component schemas")
	}
	resolved, ok := schemas[name].(map[string]any)
	if !ok {
		t.Fatalf("village document has no component schema %q", name)
	}
	return resolved
}

func trimPrefix(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return "", false
	}
	return s[len(prefix):], true
}

// enumOf reads a component's enum members as strings, following one level of
// reference so a property that points at a named enum resolves.
func enumOf(t *testing.T, document map[string]any, property map[string]any) []string {
	t.Helper()
	target := property
	if ref, ok := property["$ref"].(string); ok {
		target = resolveRef(t, document, ref)
	}
	raw, ok := target["enum"].([]any)
	if !ok {
		t.Fatalf("schema %v declares no enum", target)
	}
	members := make([]string, 0, len(raw))
	for _, value := range raw {
		members = append(members, fmt.Sprint(value))
	}
	return members
}

func bodySchema(t *testing.T, operation map[string]any) map[string]any {
	t.Helper()
	body, ok := operation["requestBody"].(map[string]any)
	if !ok {
		t.Fatal("owner update operation declares no request body")
	}
	content, ok := body["content"].(map[string]any)
	if !ok {
		t.Fatal("owner update request body declares no content")
	}
	media, ok := content["application/json"].(map[string]any)
	if !ok {
		t.Fatal("owner update request body declares no application/json content")
	}
	s, ok := media["schema"].(map[string]any)
	if !ok {
		t.Fatal("owner update request body declares no schema")
	}
	return s
}

// TestOwnerUpdateSpecExpectations drives every fixture probe against the built
// Village document, so the declaration can never drift from the source contract
// without a fixture row going red.
func TestOwnerUpdateSpecExpectations(t *testing.T) {
	fixtures, err := schema.LoadOwnerUpdateFixtures()
	if err != nil {
		t.Fatalf("load owner update fixtures: %v", err)
	}
	assert.RequireValid(t, fixtures.SpecExpectations)
	assert.RequireMin(t, fixtures.SpecExpectations, len(schema.AllOwnerUpdateSpecProbeKinds))

	document := villageSpecAsMap(t)
	operation := ownerUpdateOperation(t, document)

	covered := make(map[schema.OwnerUpdateSpecProbeKind]bool, len(schema.AllOwnerUpdateSpecProbeKinds))
	for _, c := range fixtures.SpecExpectations.Cases {
		covered[c.Input.Probe] = true
		t.Run(c.Name, func(t *testing.T) {
			var observed []string
			switch c.Input.Probe {
			case schema.OwnerUpdateProbeOperationDeclared:
				id, ok := operation["operationId"].(string)
				if !ok {
					t.Fatal("owner update operation declares no operationId")
				}
				observed = []string{id}

			case schema.OwnerUpdateProbePathParameter:
				parameters, ok := operation["parameters"].([]any)
				if !ok {
					t.Fatal("owner update operation declares no parameters")
				}
				for _, entry := range parameters {
					parameter, ok := entry.(map[string]any)
					if !ok {
						continue
					}
					if parameter["in"] != "path" {
						continue
					}
					if required, _ := parameter["required"].(bool); !required {
						t.Fatalf("path parameter %v must be required; the route addresses exactly one transcript", parameter["name"])
					}
					observed = append(observed, fmt.Sprint(parameter["name"]))
				}

			case schema.OwnerUpdateProbeBodyIsReference:
				ref, ok := bodySchema(t, operation)["$ref"].(string)
				if !ok {
					t.Fatal("owner update request body must be a component reference, not an inline object, so consumers generate a named type")
				}
				name, _ := trimPrefix(ref, "#/components/schemas/")
				observed = []string{name}

			case schema.OwnerUpdateProbeResponseStatuses:
				responses, ok := operation["responses"].(map[string]any)
				if !ok {
					t.Fatal("owner update operation declares no responses")
				}
				for status := range responses {
					observed = append(observed, status)
				}

			case schema.OwnerUpdateProbeSuccessHasNoBody:
				responses, ok := operation["responses"].(map[string]any)
				if !ok {
					t.Fatal("owner update operation declares no responses")
				}
				success, ok := responses["200"].(map[string]any)
				if !ok {
					t.Fatal("owner update operation must declare the 200 status even though its body is undeclared; omitting the status entirely would read as 'this never succeeds'")
				}
				if _, hasContent := success["content"]; hasContent {
					t.Fatal("the 200 body is deliberately undeclared because village serves internal storage columns through pgtype wrappers; declaring a success shape it does not serve would break served-equals-declared. See village issue 55 before changing this")
				}
				observed = []string{"no-content-declared"}

			case schema.OwnerUpdateProbeVisibilityEnum, schema.OwnerUpdateProbeLicenseEnum:
				body := resolveRef(t, document, bodySchema(t, operation)["$ref"].(string))
				properties, ok := body["properties"].(map[string]any)
				if !ok {
					t.Fatal("owner update body declares no properties")
				}
				field := "visibility"
				if c.Input.Probe == schema.OwnerUpdateProbeLicenseEnum {
					field = "license"
				}
				property, ok := properties[field].(map[string]any)
				if !ok {
					t.Fatalf("owner update body declares no %q property", field)
				}
				observed = enumOf(t, document, property)

			default:
				t.Fatalf("probe %q has no implementation; the probe set is closed and every member must be asserted", c.Input.Probe)
			}

			want := append([]string(nil), c.Expected.Strings...)
			got := append([]string(nil), observed...)
			sort.Strings(want)
			sort.Strings(got)
			if len(want) != len(got) {
				t.Fatalf("probe %q observed %v, want exactly %v", c.Input.Probe, got, want)
			}
			for i := range want {
				if want[i] != got[i] {
					t.Fatalf("probe %q observed %v, want exactly %v", c.Input.Probe, got, want)
				}
			}
		})
	}

	for _, kind := range schema.AllOwnerUpdateSpecProbeKinds {
		if !covered[kind] {
			t.Errorf("probe kind %q has no fixture row; every declared probe must be exercised or the closed set silently loses coverage", kind)
		}
	}
}

// TestOwnerUpdateRefusalsShareOneEnvelope pins that every declared refusal
// carries the same error body. A client reads the reason from one field
// regardless of which refusal it hit, and that is only true if the declaration
// says so.
func TestOwnerUpdateRefusalsShareOneEnvelope(t *testing.T) {
	document := villageSpecAsMap(t)
	operation := ownerUpdateOperation(t, document)
	responses, ok := operation["responses"].(map[string]any)
	if !ok {
		t.Fatal("owner update operation declares no responses")
	}
	if len(responses) == 0 {
		t.Fatal("owner update operation declares no refusals; the ownership boundary and the irrevocability rule must be readable from the contract")
	}
	refusals := 0
	for status, entry := range responses {
		// The success status carries no body by design (see the
		// success_has_no_body probe); this test is about the refusals sharing one
		// envelope, so skip it rather than weaken the envelope assertion.
		if status == "200" {
			continue
		}
		refusals++
		response, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("response %s is malformed", status)
		}
		content, ok := response["content"].(map[string]any)
		if !ok {
			t.Fatalf("response %s declares no content", status)
		}
		media, ok := content["application/json"].(map[string]any)
		if !ok {
			t.Fatalf("response %s declares no application/json content", status)
		}
		s, ok := media["schema"].(map[string]any)
		if !ok {
			t.Fatalf("response %s declares no schema", status)
		}
		ref, ok := s["$ref"].(string)
		if !ok {
			t.Fatalf("response %s must reference the shared refusal envelope", status)
		}
		resolved := resolveRef(t, document, ref)
		properties, ok := resolved["properties"].(map[string]any)
		if !ok {
			t.Fatalf("response %s envelope declares no properties", status)
		}
		if _, ok := properties["error"]; !ok {
			t.Fatalf("response %s envelope must carry the error field the village actually serves", status)
		}
	}
	if refusals == 0 {
		t.Fatal("no refusal responses were examined; this test would pass vacuously if the declaration lost every 4xx and 5xx")
	}
}
