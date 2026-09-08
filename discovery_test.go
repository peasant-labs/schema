package schema_test

import (
	_ "embed"
	"encoding/json"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	schemaopenapi "github.com/peasant-labs/schema/openapi"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/contract/browser_discovery.yaml
var browserDiscoveryFixtureYAML []byte

type browserDiscoveryFacetCase struct {
	Name  string `yaml:"name"`
	Input string `yaml:"input"`
	Valid bool   `yaml:"valid"`
}

type browserDiscoveryTokenCase struct {
	Name      string  `yaml:"name"`
	TokensIn  *int64  `yaml:"tokens_in"`
	TokensOut *int64  `yaml:"tokens_out"`
	Legacy    *int64  `yaml:"legacy"`
	Expected  *string `yaml:"expected"`
}

type browserDiscoveryFixtures struct {
	RequiredNames struct {
		FacetCases []string `yaml:"facet_cases"`
		TokenCases []string `yaml:"token_cases"`
	} `yaml:"required_names"`
	FacetCases []browserDiscoveryFacetCase `yaml:"facet_cases"`
	TokenCases []browserDiscoveryTokenCase `yaml:"token_cases"`
}

func loadBrowserDiscoveryFixtures(t *testing.T) browserDiscoveryFixtures {
	t.Helper()
	var fixtures browserDiscoveryFixtures
	if err := yaml.Unmarshal(browserDiscoveryFixtureYAML, &fixtures); err != nil {
		t.Fatalf("load browser discovery fixtures: %v", err)
	}
	requireNames(t, "facet_cases", fixtures.RequiredNames.FacetCases, facetCaseNames(fixtures.FacetCases))
	requireNames(t, "token_cases", fixtures.RequiredNames.TokenCases, tokenCaseNames(fixtures.TokenCases))
	return fixtures
}

func requireNames(t *testing.T, arm string, required, actual []string) {
	t.Helper()
	sort.Strings(required)
	sort.Strings(actual)
	if !reflect.DeepEqual(required, actual) {
		t.Fatalf("browser discovery %s names = %v, required exactly %v", arm, actual, required)
	}
}

func facetCaseNames(cases []browserDiscoveryFacetCase) []string {
	names := make([]string, len(cases))
	for i := range cases {
		names[i] = cases[i].Name
	}
	return names
}
func tokenCaseNames(cases []browserDiscoveryTokenCase) []string {
	names := make([]string, len(cases))
	for i := range cases {
		names[i] = cases[i].Name
	}
	return names
}

func TestVillageHarnessFacetFixtures(t *testing.T) {
	fixtures := loadBrowserDiscoveryFixtures(t)
	spec, err := schemaopenapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(map[string]any{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"$ref":       "#/components/schemas/VillageDiscoveryResponse/properties/harness_facets",
		"components": spec.Components,
	})
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("browser-discovery.json", strings.NewReader(string(document))); err != nil {
		t.Fatal(err)
	}
	validator, err := compiler.Compile("browser-discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixtures.FacetCases {
		var raw any
		if err := json.Unmarshal([]byte(tc.Input), &raw); err != nil {
			t.Fatalf("%s: decode: %v", tc.Name, err)
		}
		valid := validator.Validate(raw) == nil
		if valid != tc.Valid {
			t.Errorf("%s: validity = %v, want %v", tc.Name, valid, tc.Valid)
		}
	}
}

func TestVillageDiscoveryTokenExamples(t *testing.T) {
	for _, tc := range loadBrowserDiscoveryFixtures(t).TokenCases {
		var got *big.Int
		if tc.TokensIn != nil || tc.TokensOut != nil {
			got = new(big.Int)
			if tc.TokensIn != nil {
				got.Add(got, big.NewInt(*tc.TokensIn))
			}
			if tc.TokensOut != nil {
				got.Add(got, big.NewInt(*tc.TokensOut))
			}
		} else if tc.Legacy != nil {
			got = big.NewInt(*tc.Legacy)
		}
		if tc.Expected == nil && got != nil {
			t.Errorf("%s: effective tokens = %s, want null", tc.Name, got)
		}
		if tc.Expected != nil && (got == nil || got.String() != *tc.Expected) {
			t.Errorf("%s: effective tokens = %v, want %s", tc.Name, got, *tc.Expected)
		}
	}
}

func TestPullListResponseHasNoHarnessFacets(t *testing.T) {
	typeOf := reflect.TypeFor[schema.PullListResponse]()
	if _, ok := typeOf.FieldByName("HarnessFacets"); ok {
		t.Fatal("PullListResponse must remain independent of browser discovery facets")
	}
	b, err := json.Marshal(schema.PullListResponse{Transcripts: []schema.PullTranscriptInfo{}, Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if json.Valid(b) && string(b) == "" {
		t.Fatal("unreachable")
	}
	var object map[string]any
	if err := json.Unmarshal(b, &object); err != nil {
		t.Fatal(err)
	}
	if _, ok := object["harness_facets"]; ok {
		t.Fatal("pull response serialized browser harness_facets")
	}
}

func TestVillageBrowserDiscoveryOperationIsRegistered(t *testing.T) {
	spec, err := schemaopenapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	path, ok := spec.Paths.MapOfPathItemValues["/api/v1/transcripts"]
	if !ok || path.Get == nil {
		t.Fatal("Village API must register GET /api/v1/transcripts independently from pull")
	}
	if path.Get.ID == nil || *path.Get.ID != "listTranscripts" {
		t.Fatalf("browser discovery operation id = %v, want listTranscripts", path.Get.ID)
	}
	description := ""
	if path.Get.Description != nil {
		description = *path.Get.Description
	}
	for _, anchor := range []string{"full viewer-visible default corpus", "user and unknown", "agent origin is excluded", "independently of all active filters, sorting, and pagination", "cannot overflow int64", "published_at descending", "transcript id descending"} {
		if !strings.Contains(description, anchor) {
			t.Errorf("browser discovery description missing %q", anchor)
		}
	}
	if _, pullOK := spec.Paths.MapOfPathItemValues["/api/v1/pull/transcripts"]; !pullOK {
		t.Fatal("browser discovery registration removed the independent pull listing")
	}
}
