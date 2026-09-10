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
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/contract/browser_discovery.yaml
var browserDiscoveryFixtureYAML []byte

type browserDiscoveryTokenInput struct {
	TokensIn  *int64 `yaml:"tokens_in"`
	TokensOut *int64 `yaml:"tokens_out"`
	Legacy    *int64 `yaml:"legacy"`
}

type browserDiscoveryFixtures struct {
	FacetCases         testcase.Corpus[string, bool]                        `yaml:"facet_cases"`
	EnvelopeCases      testcase.Corpus[string, bool]                        `yaml:"envelope_cases"`
	TokenExamples      testcase.Corpus[browserDiscoveryTokenInput, *string] `yaml:"token_examples"`
	DescriptionAnchors []string                                             `yaml:"description_anchors"`
	RequiredNames      struct {
		FacetCases    []string `yaml:"facet_cases"`
		EnvelopeCases []string `yaml:"envelope_cases"`
		TokenExamples []string `yaml:"token_examples"`
	} `yaml:"required_names"`
}

func loadBrowserDiscoveryFixtures(t *testing.T) browserDiscoveryFixtures {
	t.Helper()
	var fixtures browserDiscoveryFixtures
	if err := yaml.Unmarshal(browserDiscoveryFixtureYAML, &fixtures); err != nil {
		t.Fatalf("load browser discovery fixtures: %v", err)
	}
	caseassert.RequireMin(t, fixtures.FacetCases, 6)
	caseassert.RequireValid(t, fixtures.FacetCases)
	caseassert.RequireMin(t, fixtures.EnvelopeCases, 5)
	caseassert.RequireValid(t, fixtures.EnvelopeCases)
	caseassert.RequireMin(t, fixtures.TokenExamples, 7)
	caseassert.RequireValid(t, fixtures.TokenExamples)
	requireCorpusNames(t, "facet_cases", fixtures.RequiredNames.FacetCases, fixtures.FacetCases)
	requireCorpusNames(t, "envelope_cases", fixtures.RequiredNames.EnvelopeCases, fixtures.EnvelopeCases)
	requireCorpusNames(t, "token_examples", fixtures.RequiredNames.TokenExamples, fixtures.TokenExamples)
	if len(fixtures.DescriptionAnchors) < 13 {
		t.Fatalf("browser discovery description anchors = %d, want at least 13", len(fixtures.DescriptionAnchors))
	}
	return fixtures
}

func requireCorpusNames[I any, E any](t *testing.T, arm string, required []string, corpus testcase.Corpus[I, E]) {
	t.Helper()
	actual := make([]string, len(corpus.Cases))
	for i := range corpus.Cases {
		actual[i] = corpus.Cases[i].Name
	}
	sort.Strings(actual)
	sort.Strings(required)
	if !reflect.DeepEqual(actual, required) {
		t.Fatalf("browser discovery %s names = %v, require exactly %v", arm, actual, required)
	}
}

func compileContractLocation(t *testing.T, documentSource any, location string) *jsonschema.Schema {
	t.Helper()
	document, err := json.Marshal(documentSource)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(document, &root); err != nil {
		t.Fatal(err)
	}
	root["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	root["$ref"] = location
	document, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("village-operation.json", strings.NewReader(string(document))); err != nil {
		t.Fatal(err)
	}
	validator, err := compiler.Compile("village-operation.json")
	if err != nil {
		t.Fatal(err)
	}
	return validator
}

func compileVillageContractLocation(t *testing.T, location string) *jsonschema.Schema {
	t.Helper()
	spec, err := schemaopenapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	return compileContractLocation(t, spec, location)
}

func validateFixtureJSON(t *testing.T, validator *jsonschema.Schema, corpus testcase.Corpus[string, bool]) {
	t.Helper()
	for _, tc := range corpus.Cases {
		var raw any
		if err := json.Unmarshal([]byte(tc.Input), &raw); err != nil {
			t.Fatalf("%s: decode fixture JSON: %v", tc.Name, err)
		}
		validationErr := validator.Validate(raw)
		valid := validationErr == nil
		if valid != tc.Expected {
			t.Errorf("%s: contract validity = %v, want %v: %v", tc.Name, valid, tc.Expected, validationErr)
		}
	}
}

func TestVillageHarnessFacetFixtures(t *testing.T) {
	fixtures := loadBrowserDiscoveryFixtures(t)
	validateFixtureJSON(t, compileVillageContractLocation(t, "#/paths/~1api~1v1~1transcripts/get/responses/200/content/application~1json/schema"), wrapFacetArrays(fixtures.FacetCases))
}

func wrapFacetArrays(corpus testcase.Corpus[string, bool]) testcase.Corpus[string, bool] {
	for i := range corpus.Cases {
		corpus.Cases[i].Input = `{"transcripts":[],"harness_facets":` + corpus.Cases[i].Input + `,"total":0,"agent_total":0,"page":1,"limit":20}`
	}
	return corpus
}

func TestVillageDiscoveryEnvelopeFixtures(t *testing.T) {
	fixtures := loadBrowserDiscoveryFixtures(t)
	operationValidator := compileVillageContractLocation(t, "#/paths/~1api~1v1~1transcripts/get/responses/200/content/application~1json/schema")
	typesSpec, err := schemaopenapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	catalogValidator := compileContractLocation(t, typesSpec, "#/components/schemas/VillageDiscoveryResponse")
	validateFixtureJSON(t, operationValidator, fixtures.EnvelopeCases)
	validateFixtureJSON(t, catalogValidator, fixtures.EnvelopeCases)
	for _, tc := range fixtures.EnvelopeCases.Cases {
		if !tc.Expected {
			continue
		}
		var decoded schema.VillageDiscoveryResponse
		if err := json.Unmarshal([]byte(tc.Input), &decoded); err != nil {
			t.Fatalf("%s: typed decode: %v", tc.Name, err)
		}
		reencoded, err := json.Marshal(decoded)
		if err != nil {
			t.Fatalf("%s: typed re-encode: %v", tc.Name, err)
		}
		var originalJSON, reencodedJSON any
		if err := json.Unmarshal([]byte(tc.Input), &originalJSON); err != nil {
			t.Fatalf("%s: decode original semantic JSON: %v", tc.Name, err)
		}
		if err := json.Unmarshal(reencoded, &reencodedJSON); err != nil {
			t.Fatalf("%s: decode re-encoded semantic JSON: %v", tc.Name, err)
		}
		if !reflect.DeepEqual(originalJSON, reencodedJSON) {
			t.Errorf("%s: typed round trip changed semantic JSON\noriginal: %s\nre-encoded: %s", tc.Name, tc.Input, reencoded)
		}
		if err := operationValidator.Validate(reencodedJSON); err != nil {
			t.Errorf("%s: typed re-encoding violates registered operation: %v", tc.Name, err)
		}
		if err := catalogValidator.Validate(reencodedJSON); err != nil {
			t.Errorf("%s: typed re-encoding violates Types catalog: %v", tc.Name, err)
		}
		if len(decoded.Transcripts) != 1 || len(decoded.HarnessFacets) != 1 {
			t.Fatalf("%s: typed envelope lost wrapped transcript or facet", tc.Name)
		}
		row := decoded.Transcripts[0]
		if strings.Contains(tc.Name, "null-enrichments") && (row.OwnerOrgs != nil || row.Shares != nil || row.Attestations != nil || row.Transcript.Subagents != nil || row.Transcript.DiagnosticsWarnings != nil) {
			t.Fatalf("%s: null production values were not preserved", tc.Name)
		}
		if strings.Contains(tc.Name, "populated") && (string(row.Transcript.Subagents) != `[{"id":"child"}]` || string(row.Transcript.DiagnosticsWarnings) != `["warning"]` || len(row.OwnerOrgs) != 1 || len(row.Shares) != 1 || len(row.Attestations) != 1) {
			t.Fatalf("%s: populated encoded metadata or associations were not preserved", tc.Name)
		}
	}
}

func TestVillageDiscoveryTokenExamples(t *testing.T) {
	for _, tc := range loadBrowserDiscoveryFixtures(t).TokenExamples.Cases {
		var got *big.Int
		if tc.Input.TokensIn != nil || tc.Input.TokensOut != nil {
			got = new(big.Int)
			if tc.Input.TokensIn != nil {
				got.Add(got, big.NewInt(*tc.Input.TokensIn))
			}
			if tc.Input.TokensOut != nil {
				got.Add(got, big.NewInt(*tc.Input.TokensOut))
			}
		} else if tc.Input.Legacy != nil {
			got = big.NewInt(*tc.Input.Legacy)
		}
		if tc.Expected == nil && got != nil {
			t.Errorf("%s arithmetic example = %s, want null", tc.Name, got)
		}
		if tc.Expected != nil && (got == nil || got.String() != *tc.Expected) {
			t.Errorf("%s arithmetic example = %v, want %s", tc.Name, got, *tc.Expected)
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
	if strings.Contains(string(b), "harness_facets") {
		t.Fatal("pull response serialized browser harness_facets")
	}
}

func TestVillageBrowserDiscoveryOperationIsRegistered(t *testing.T) {
	fixtures := loadBrowserDiscoveryFixtures(t)
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
	for _, anchor := range fixtures.DescriptionAnchors {
		if !strings.Contains(description, anchor) {
			t.Errorf("browser discovery description missing %q", anchor)
		}
	}
	for _, status := range []string{"200", "400", "500"} {
		if _, ok := path.Get.Responses.MapOfResponseOrReferenceValues[status]; !ok {
			t.Errorf("browser discovery missing %s response", status)
		}
	}
	for _, status := range []string{"400", "500"} {
		validator := compileVillageContractLocation(t, "#/paths/~1api~1v1~1transcripts/get/responses/"+status+"/content/application~1json/schema")
		if err := validator.Validate(map[string]any{"error": "actionable refusal"}); err != nil {
			t.Errorf("browser discovery %s error envelope: %v", status, err)
		}
	}
	if _, ok := path.Get.Responses.MapOfResponseOrReferenceValues["401"]; ok {
		t.Error("optional-auth browser discovery must not promise a 401 response")
	}
	if _, pullOK := spec.Paths.MapOfPathItemValues["/api/v1/pull/transcripts"]; !pullOK {
		t.Fatal("browser discovery registration removed the independent pull listing")
	}
}
