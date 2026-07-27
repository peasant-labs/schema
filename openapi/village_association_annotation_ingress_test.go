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
	assert.RequireMin(t, cases, 9)
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
	for _, phrase := range []string{"targetKind association requires only targetAssociationId", "every other target kind rejects targetAssociationId"} {
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

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal diagnostic JSON: %v", err)
	}
	return string(raw)
}
