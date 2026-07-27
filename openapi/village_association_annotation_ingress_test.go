package openapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/internal/testutil"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase/assert"
)

// TestBuildVillageAPISpec_AnnotationPushIngress verifies that the generated
// Village operation names the canonical request and response components and
// documents the same association-target exclusivity that the typed validators
// enforce. Its fixture loop keeps the documented field wired to every ingress
// case, rather than testing a hand-written surrogate schema.
func TestBuildVillageAPISpec_AnnotationPushIngress(t *testing.T) {
	fixtures, err := testutil.DecodeAssociationAnnotationIngressFixtures(schema.AssociationAnnotationIngressYAML)
	if err != nil {
		t.Fatalf("LoadAssociationAnnotationIngressFixtures: %v", err)
	}
	cases := fixtures.CaseCorpus()
	assert.RequireMin(t, cases, 11)
	assert.RequireValid(t, cases)

	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	raw, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal Village API spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode Village API spec: %v", err)
	}

	operation := villageAnnotationPushOperation(t, document)
	description, _ := operation["description"].(string)
	for _, phrase := range []string{"targetKind association requires only targetAssociationId", "every other target kind rejects targetAssociationId", "endIndex must be greater than entryIndex and is enforced by the request validation boundary"} {
		if !strings.Contains(description, phrase) {
			t.Errorf("annotation push operation description is missing validation rule %q: %q", phrase, description)
		}
	}
	if got := operation["operationId"]; got != "pushAnnotations" {
		t.Errorf("annotation push operationId=%v, want pushAnnotations", got)
	}
	if !operationReferencesComponent(operation, "AnnotationPushRequest") || !operationReferencesComponent(operation, "AnnotationPushResponse") {
		t.Fatalf("annotation push operation must use AnnotationPushRequest and AnnotationPushResponse: %s", raw)
	}

	itemSchema := villageComponent(t, document, "AnnotationPushItem")
	properties, ok := itemSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("AnnotationPushItem component has no properties")
	}
	targetAssociation, ok := properties["targetAssociationId"].(map[string]any)
	if !ok {
		t.Fatal("AnnotationPushItem component is missing targetAssociationId")
	}
	if !strings.Contains(mustJSON(t, targetAssociation), "AssociationID") {
		t.Errorf("targetAssociationId must retain the AssociationID schema, got %s", mustJSON(t, targetAssociation))
	}
	for _, fixture := range cases.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			if fixture.Input.Annotation.TargetKind == string(schema.TargetAssociation) || fixture.Input.Annotation.TargetAssociationID != nil {
				if _, exists := properties["targetAssociationId"]; !exists {
					t.Fatal("generated annotation push schema cannot express this fixture's association target")
				}
			}
		})
	}
}

// TestBuildVillageAPISpec_AnnotationPushOperationSchema runs the shared corpus
// through the actual generated operation request schema. It extracts the
// operation's request-body component and its generated component graph rather
// than recreating a test-only validator.
func TestBuildVillageAPISpec_AnnotationPushOperationSchema(t *testing.T) {
	fixtures, err := testutil.DecodeAssociationAnnotationIngressFixtures(schema.AssociationAnnotationIngressYAML)
	if err != nil {
		t.Fatalf("DecodeAssociationAnnotationIngressFixtures: %v", err)
	}
	cases := fixtures.CaseCorpus()
	assert.RequireMin(t, cases, 11)
	assert.RequireValid(t, cases)
	assert.RequireMin(t, fixtures.AnnotationRequestShapes, 2)
	assert.RequireValid(t, fixtures.AnnotationRequestShapes)

	spec, err := specpkg.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}
	raw, err := spec.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal Village API spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode Village API spec: %v", err)
	}
	operationSchema := compileSchema(t, "village-annotation-push.json", standaloneOperationRequestSchema(t, document))

	for _, fixture := range cases.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			item := testutil.AnnotationPushItemTargetJSON(fixture.Input.Annotation)
			item["contentHash"] = "fixture-content-hash"
			item["typeId"] = "fixture.annotation"
			item["value"] = "fixture-value"
			item["isPrimary"] = false
			requireExplicitNullTargetFields(t, fixture.Input.Annotation, item)
			body := mustJSONBytes(t, map[string]any{"annotations": []any{item}})
			if got := accepts(t, operationSchema, body); got != fixture.Expected.OperationSchemaValid() {
				t.Fatalf("generated Village annotation operation schema accepted=%t, want %t", got, fixture.Expected.OperationSchemaValid())
			}
		})
	}
	for _, fixture := range fixtures.AnnotationRequestShapes.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			body := mustJSONBytes(t, map[string]any{"annotations": fixture.Input.Annotations})
			if got := accepts(t, operationSchema, body); got != fixture.Expected {
				t.Fatalf("generated Village annotation operation schema accepted=%t, want %t", got, fixture.Expected)
			}
		})
	}
}

func requireExplicitNullTargetFields(t *testing.T, fixture testutil.AssociationAnnotationIngressAnnotation, item map[string]any) {
	t.Helper()
	for _, field := range []string{"targetAssociationId", "sessionId", "entryTarget", "annotationId", "projectHash"} {
		if !fixture.HasExplicitNullTargetField(field) {
			continue
		}
		value, present := item[field]
		if !present || value != nil {
			t.Fatalf("fixture explicit-null target field %q must remain an explicit JSON null in the Village operation body, got present=%t value=%v", field, present, value)
		}
	}
}

func villageAnnotationPushOperation(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec has no paths")
	}
	path, ok := paths["/api/v1/annotations"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec is missing POST /api/v1/annotations")
	}
	operation, ok := path["post"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec is missing POST /api/v1/annotations operation")
	}
	return operation
}

func villageComponent(t *testing.T, document map[string]any, canonicalName string) map[string]any {
	t.Helper()
	components, ok := document["components"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec has no components")
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec has no component schemas")
	}
	for name, raw := range schemas {
		if name != canonicalName && name != "Schema"+canonicalName {
			continue
		}
		component, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("Village API component %s is not an object", name)
		}
		return component
	}
	t.Fatalf("Village API spec has no %s component", canonicalName)
	return nil
}

func operationReferencesComponent(operation map[string]any, component string) bool {
	raw, err := json.Marshal(operation)
	return err == nil && strings.Contains(string(raw), component)
}

func standaloneOperationRequestSchema(t *testing.T, document map[string]any) []byte {
	t.Helper()
	operation := villageAnnotationPushOperation(t, document)
	requestBody, ok := operation["requestBody"].(map[string]any)
	if !ok {
		t.Fatal("Village annotation push operation has no request body")
	}
	content, ok := requestBody["content"].(map[string]any)
	if !ok {
		t.Fatal("Village annotation push operation request body has no content")
	}
	jsonContent, ok := content["application/json"].(map[string]any)
	if !ok {
		t.Fatal("Village annotation push operation has no application/json request schema")
	}
	requestSchema, ok := jsonContent["schema"].(map[string]any)
	if !ok {
		t.Fatal("Village annotation push operation application/json content has no schema")
	}
	ref, ok := requestSchema["$ref"].(string)
	if !ok {
		t.Fatal("Village annotation push operation request schema is not a component reference")
	}
	components, ok := document["components"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec has no components")
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		t.Fatal("Village API spec has no component schemas")
	}
	standalone := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$ref":    rewriteOperationComponentRef(ref),
		"$defs":   rewriteOperationComponentRefs(schemas),
	}
	return mustJSONBytes(t, standalone)
}

func rewriteOperationComponentRefs(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		copy := make(map[string]any, len(typed))
		for key, child := range typed {
			if key == "$ref" {
				if ref, ok := child.(string); ok {
					copy[key] = rewriteOperationComponentRef(ref)
					continue
				}
			}
			copy[key] = rewriteOperationComponentRefs(child)
		}
		return copy
	case []any:
		copy := make([]any, len(typed))
		for index, child := range typed {
			copy[index] = rewriteOperationComponentRefs(child)
		}
		return copy
	default:
		return value
	}
}

func rewriteOperationComponentRef(ref string) string {
	const componentPrefix = "#/components/schemas/"
	if strings.HasPrefix(ref, componentPrefix) {
		return "#/$defs/" + strings.TrimPrefix(ref, componentPrefix)
	}
	return ref
}

func mustJSONBytes(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test JSON: %v", err)
	}
	return raw
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal diagnostic JSON: %v", err)
	}
	return string(raw)
}
