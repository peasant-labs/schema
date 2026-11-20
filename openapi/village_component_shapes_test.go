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
			var normalized map[string]interface{}
			if err = json.Unmarshal(raw, &normalized); err != nil {
				t.Fatal(err)
			}
			if c.MutationRemoveRequired != "" {
				required := normalized["required"].([]interface{})
				var changed []interface{}
				for _, n := range required {
					if n != c.MutationRemoveRequired {
						changed = append(changed, n)
					}
				}
				normalized["required"] = changed
			}
			matches := villageShapeMatches(normalized, c)
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

func villageShapeMatches(got map[string]interface{}, c villageShapeCase) bool {
	var actual []string
	required, _ := got["required"].([]interface{})
	for _, v := range required {
		actual = append(actual, v.(string))
	}
	expected := append([]string(nil), c.Required...)
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		return false
	}
	props, ok := got["properties"].(map[string]interface{})
	if !ok {
		return false
	}
	projected := make(map[string]interface{}, len(props))
	for k, v := range props {
		projected[k] = villageShapeProjection(v)
	}
	// Normalize YAML numbers and maps through JSON before comparing.
	b, _ := json.Marshal(c.Properties)
	var expectedProps map[string]interface{}
	_ = json.Unmarshal(b, &expectedProps)
	return reflect.DeepEqual(projected, expectedProps)
}

func villageShapeProjection(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for _, k := range []string{"type", "format", "$ref", "anyOf", "items"} {
			if val, ok := x[k]; ok {
				out[k] = villageShapeProjection(val)
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, v := range x {
			out[i] = villageShapeProjection(v)
		}
		return out
	default:
		return x
	}
}
