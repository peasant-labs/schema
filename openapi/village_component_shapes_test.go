package openapi_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	specpkg "github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_component_shapes.yaml
var villageShapeFixture []byte

type villageShapeCase struct {
	Name                   string                 `yaml:"name"`
	Component              string                 `yaml:"component"`
	Required               []string               `yaml:"required"`
	Properties             map[string]interface{} `yaml:"properties"`
	MutationRemoveRequired string                 `yaml:"mutation_remove_required"`
}

var requiredVillageShapeNames = []string{
	"VillageDiscoverabilityRequest",
	"VillageUsernameRequest",
	"VillageCreateAPIKeyRequest",
	"VillageCreatedAPIKey",
	"VillageAPIKey",
	"VillageOrganizationVisibilityRequest",
	"VillageOrganizationVisibilityResponse",
	"VillageOrganizationSearchResult",
	"VillageOrganizationSearchResponse",
	"VillageTagUsage",
	"VillageResolvedProject",
	"VillageProjectCollective",
	"VillageUserProjectResponse",
	"VillageProjectDisplayNameRequest",
	"VillageCreateAttestationRequest",
	"VillageManualAnnotationRequest",
	"VillageTranscriptAnnotationsResponse",
	"VillageTranscriptCommit",
	"VillageTranscriptCommitsResponse",
	"VillageHealthResponse",
	"VillageUser",
	"mutation removes required commit timestamp"}

func TestVillageComponentShapes(t *testing.T) {
	var fx struct {
		Cases []villageShapeCase `yaml:"cases"`
	}
	if err := yaml.Unmarshal(villageShapeFixture, &fx); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(fx.Cases))
	for _, c := range fx.Cases {
		names = append(names, c.Name)
	}
	assertSameStringSet(t, "component shape fixture names", names, requiredVillageShapeNames)
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range fx.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, ok := spec.Components.Schemas[c.Component]
			if !ok {
				t.Fatalf("missing component %s", c.Component)
			}
			raw, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var component villageComponentShape
			if err = json.Unmarshal(raw, &component); err != nil {
				t.Fatal(err)
			}
			if c.MutationRemoveRequired != "" {
				var changed []string
				for _, n := range component.Required {
					if n != c.MutationRemoveRequired {
						changed = append(changed, n)
					}
				}
				component.Required = changed
			}
			matches := villageShapeMatches(t, component, c)
			if c.MutationRemoveRequired != "" {
				if matches {
					t.Fatal("required-field mutation went undetected")
				}
				return
			}
			if !matches {
				t.Fatalf("component %s shape differs from fixture", c.Component)
			}
		})
	}
}

// villageComponentShape is the part of a component schema the shape cases
// compare: its required members and its properties.
type villageComponentShape struct {
	Required   []string                   `json:"required"`
	Properties map[string]json.RawMessage `json:"properties"`
}

// villageShapeProjection keeps the members of a property schema that decide
// its shape: type, format, reference, alternatives and items.
type villageShapeProjection struct {
	Type   any                      `json:"type,omitempty"`
	Format string                   `json:"format,omitempty"`
	Ref    string                   `json:"$ref,omitempty"`
	AnyOf  []villageShapeProjection `json:"anyOf,omitempty"`
	Items  *villageShapeProjection  `json:"items,omitempty"`
}

func villageShapeMatches(t *testing.T, got villageComponentShape, c villageShapeCase) bool {
	t.Helper()
	actual := append([]string(nil), got.Required...)
	expected := append([]string(nil), c.Required...)
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		return false
	}
	if got.Properties == nil {
		return false
	}
	projected := make(map[string]villageShapeProjection, len(got.Properties))
	for name, raw := range got.Properties {
		var projection villageShapeProjection
		if err := json.Unmarshal(raw, &projection); err != nil {
			t.Fatalf("decode property %s of %s: %v", name, c.Component, err)
		}
		projected[name] = projection
	}
	// Normalize both sides through JSON before comparing: YAML numbers and
	// maps on one side, the projection's omitted members on the other.
	return normalizedJSON(t, projected) != nil && reflect.DeepEqual(normalizedJSON(t, projected), normalizedJSON(t, c.Properties))
}

func normalizedJSON(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal for comparison: %v", err)
	}
	var normalized any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		t.Fatalf("decode for comparison: %v", err)
	}
	return normalized
}
