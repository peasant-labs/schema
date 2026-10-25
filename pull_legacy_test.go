package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/pull/legacy_envelope.yaml
var legacyPullFixtureYAML []byte

type legacyPersonInput struct {
	Holder      string `yaml:"holder"`
	ID          string `yaml:"id"`
	Username    string `yaml:"username"`
	ProjectName string `yaml:"project_name"`
}

type legacyConversionInput struct {
	Anonymous bool `yaml:"anonymous"`
	// Malformed builds a profile Validate would refuse: no_username or
	// noncanonical_id. The converter must still leak nothing.
	Malformed   string  `yaml:"malformed"`
	ProjectName *string `yaml:"project_name"`
}

type legacyServedBodyInput struct {
	Path string `yaml:"path"`
	Body string `yaml:"body"`
}

type legacyConversionExpected struct {
	OwnerUserID   string `yaml:"owner_user_id"`
	OwnerUsername string `yaml:"owner_username"`
	ProjectName   string `yaml:"project_name"`
}

type legacyPullFixtures struct {
	FrozenAgainst struct {
		Document   string            `yaml:"document"`
		Components map[string]string `yaml:"components"`
	} `yaml:"frozen_against"`
	ExactFields map[string]struct {
		Required []string `yaml:"required"`
		Optional []string `yaml:"optional"`
	} `yaml:"exact_fields"`
	AnonymousUsername string `yaml:"anonymous_username"`
	Header            string `yaml:"header"`
	Versions          struct {
		Legacy      string `yaml:"legacy"`
		PublicOwner string `yaml:"public_owner"`
	} `yaml:"versions"`
	Persons       testcase.Corpus[legacyPersonInput, villageExpected]              `yaml:"persons"`
	Conversions   testcase.Corpus[legacyConversionInput, legacyConversionExpected] `yaml:"conversions"`
	ServedBodies  testcase.Corpus[legacyServedBodyInput, villageExpected]          `yaml:"served_bodies"`
	RequiredNames struct {
		Persons      []string `yaml:"persons"`
		Conversions  []string `yaml:"conversions"`
		ServedBodies []string `yaml:"served_bodies"`
	} `yaml:"required_names"`
}

func loadLegacyPullFixtures(t *testing.T) legacyPullFixtures {
	t.Helper()
	var f legacyPullFixtures
	if err := yaml.Unmarshal(legacyPullFixtureYAML, &f); err != nil {
		t.Fatalf("load legacy pull fixtures: %v", err)
	}
	assertcase.RequireMin(t, f.Persons, 9)
	assertcase.RequireValid(t, f.Persons)
	assertcase.RequireMin(t, f.Conversions, 5)
	assertcase.RequireMin(t, f.ServedBodies, 7)
	assertcase.RequireValid(t, f.ServedBodies)
	assertSameStrings(t, "legacy served body case names", f.RequiredNames.ServedBodies, mapKeys(namesOf(f.ServedBodies)))
	assertcase.RequireValid(t, f.Conversions)
	assertSameStrings(t, "legacy person case names", f.RequiredNames.Persons, mapKeys(namesOf(f.Persons)))
	assertSameStrings(t, "legacy conversion case names", f.RequiredNames.Conversions, mapKeys(namesOf(f.Conversions)))
	return f
}

func TestLegacyPullConstantsMatchTheFixture(t *testing.T) {
	f := loadLegacyPullFixtures(t)
	if schema.LegacyPullAnonymousUsername != f.AnonymousUsername || schema.PullContractHeader != f.Header ||
		string(schema.LegacyPullContractVersion) != f.Versions.Legacy || string(schema.PublicOwnerPullContractVersion) != f.Versions.PublicOwner {
		t.Fatalf("constants %q %q %q %q differ from the fixture %+v", schema.LegacyPullAnonymousUsername, schema.PullContractHeader, schema.LegacyPullContractVersion, schema.PublicOwnerPullContractVersion, f)
	}
}

// TestLegacyPullEnvelopeIsFrozen requires each Legacy* component to equal the
// released 0.2.0 component it replaces, with only component names changed.
func TestLegacyPullEnvelopeIsFrozen(t *testing.T) {
	f := loadLegacyPullFixtures(t)
	raw, err := os.ReadFile(filepath.FromSlash(f.FrozenAgainst.Document))
	if err != nil {
		t.Fatal(err)
	}
	var released map[string]any
	if err := json.Unmarshal(raw, &released); err != nil {
		t.Fatal(err)
	}
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	var current map[string]any
	if err := json.Unmarshal(mustJSON(t, spec), &current); err != nil {
		t.Fatal(err)
	}
	releasedComponents := released["components"].(map[string]any)["schemas"].(map[string]any)
	currentComponents := current["components"].(map[string]any)["schemas"].(map[string]any)
	rename := map[string]string{}
	for legacy, original := range f.FrozenAgainst.Components {
		rename["#/components/schemas/"+original] = "#/components/schemas/" + legacy
	}
	for legacy, original := range f.FrozenAgainst.Components {
		want := renameRefs(releasedComponents[original], rename)
		if !reflect.DeepEqual(currentComponents[legacy], want) {
			t.Errorf("%s drifted from the released %s:\n got %s\nwant %s", legacy, original, mustJSON(t, currentComponents[legacy]), mustJSON(t, want))
		}
	}
}

func renameRefs(value any, rename map[string]string) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, child := range typed {
			if ref, ok := child.(string); ok && key == "$ref" {
				if renamed, ok := rename[ref]; ok {
					ref = renamed
				}
				out[key] = ref
				continue
			}
			out[key] = renameRefs(child, rename)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			out[i] = renameRefs(child, rename)
		}
		return out
	}
	return value
}

func TestLegacyPullEnvelopeExactFields(t *testing.T) {
	f := loadLegacyPullFixtures(t)
	project := "project"
	full := map[string]any{
		"LegacyPullTranscriptInfo": schema.LegacyPullTranscriptInfo{TranscriptID: "123e4567-e89b-12d3-a456-426614174003", LocalID: "local", OwnerUserID: "id", OwnerUsername: "alice", Title: "title", Harness: schema.HarnessCodex, ProjectName: project, Visibility: schema.VisibilityGroup, License: schema.LicenseCC0, ContentHash: "hash", ContractVersion: "0.1.0", PublishedAt: 1, UpdatedAt: 1, AnnotationCount: 1},
		"LegacyPullListResponse":   schema.LegacyPullListResponse{Transcripts: []schema.LegacyPullTranscriptInfo{}, Page: 1, Limit: 1},
		"LegacyPullAnnotation":     schema.LegacyPullAnnotation{AuthorUserID: "id", AuthorUsername: "bob"},
	}
	empty := map[string]any{
		"LegacyPullTranscriptInfo": schema.LegacyPullTranscriptInfo{},
		"LegacyPullListResponse":   schema.LegacyPullListResponse{},
		"LegacyPullAnnotation":     schema.LegacyPullAnnotation{},
	}
	assertSameStrings(t, "legacy holders", mapKeysAny(f.ExactFields), mapKeysAny(full))
	for holder, fields := range f.ExactFields {
		var emptyKeys, fullKeys map[string]any
		if err := json.Unmarshal(mustJSON(t, empty[holder]), &emptyKeys); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(mustJSON(t, full[holder]), &fullKeys); err != nil {
			t.Fatal(err)
		}
		if holder == "LegacyPullAnnotation" {
			// The embedded AnnotationSummary keeps its own pinned shape; only the
			// author fields are the legacy envelope's own.
			emptyKeys = pickKeys(emptyKeys, "authorUserId", "authorUsername")
			fullKeys = pickKeys(fullKeys, "authorUserId", "authorUsername")
		}
		assertSameStrings(t, holder+" fields always sent", fields.Required, mapKeysAny(emptyKeys))
		assertSameStrings(t, holder+" fields when set", append(append([]string{}, fields.Required...), fields.Optional...), mapKeysAny(fullKeys))
	}
}

func pickKeys(m map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		if value, ok := m[key]; ok {
			out[key] = value
		}
	}
	return out
}

func TestLegacyPullPersonsKeepTheSentinel(t *testing.T) {
	f := loadLegacyPullFixtures(t)
	for _, c := range f.Persons.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			switch c.Input.Holder {
			case "transcript":
				info := schema.LegacyPullTranscriptInfo{OwnerUserID: c.Input.ID, OwnerUsername: c.Input.Username, ProjectName: c.Input.ProjectName}
				err = info.Validate()
				if listErr := (schema.LegacyPullListResponse{Transcripts: []schema.LegacyPullTranscriptInfo{info}}).Validate(); (listErr == nil) != (err == nil) {
					t.Fatalf("list validation %v disagrees with row validation %v", listErr, err)
				}
			case "annotation":
				err = schema.LegacyPullAnnotation{AuthorUserID: c.Input.ID, AuthorUsername: c.Input.Username}.Validate()
			default:
				t.Fatalf("unknown holder %q", c.Input.Holder)
			}
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", err == nil, err)
			}
		})
	}
}

func TestLegacyPullConvertersApplyTheBlankingRule(t *testing.T) {
	f := loadLegacyPullFixtures(t)
	for _, c := range f.Conversions.Cases {
		t.Run(c.Name, func(t *testing.T) {
			owner := schema.NewVillagePublicProfile("123e4567-e89b-12d3-a456-426614174001", "alice", nil, nil)
			author := schema.NewVillagePublicProfile("123e4567-e89b-12d3-a456-426614174001", "alice", nil, nil)
			if c.Input.Anonymous {
				owner, author = schema.AnonymousVillageOwner(), schema.AnonymousVillageOwner()
			}
			switch c.Input.Malformed {
			case "":
			case "no_username":
				owner = schema.VillagePublicOwner{Kind: schema.VillagePublicOwnerProfile, ID: "123e4567-e89b-12d3-a456-426614174001"}
			case "noncanonical_id":
				owner = schema.NewVillagePublicProfile("123E4567-E89B-12D3-A456-426614174001", "alice", nil, nil)
			default:
				t.Fatalf("unknown malformed profile %q", c.Input.Malformed)
			}
			if c.Input.Malformed != "" {
				author = owner
			}
			info := schema.LegacyPullTranscriptInfoFrom(schema.PullTranscriptInfo{Owner: owner, ProjectName: c.Input.ProjectName})
			if info.OwnerUserID != c.Expected.OwnerUserID || info.OwnerUsername != c.Expected.OwnerUsername || info.ProjectName != c.Expected.ProjectName {
				t.Fatalf("converted %+v, want %+v", info, c.Expected)
			}
			if err := info.Validate(); err != nil {
				t.Fatalf("a converted transcript fails its own rule: %v", err)
			}
			list := schema.LegacyPullListResponseFrom(schema.PullListResponse{Transcripts: []schema.PullTranscriptInfo{{Owner: owner, ProjectName: c.Input.ProjectName}}, Page: 1, Limit: 1, Total: 1})
			if len(list.Transcripts) != 1 || list.Transcripts[0] != info || list.Validate() != nil {
				t.Fatalf("list conversion disagrees with row conversion: %+v", list)
			}
			annotation := schema.LegacyPullAnnotationFrom(schema.PullAnnotation{Author: author})
			if annotation.AuthorUserID != c.Expected.OwnerUserID || annotation.AuthorUsername != c.Expected.OwnerUsername || annotation.Validate() != nil {
				t.Fatalf("annotation converted %+v, want id %q username %q", annotation, c.Expected.OwnerUserID, c.Expected.OwnerUsername)
			}
		})
	}
}

// TestPullOperationsDeclareTheContractHeader requires the request and
// response header on all four pull operations, and both envelopes on the JSON
// ones.
func TestPullOperationsDeclareTheContractHeader(t *testing.T) {
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(mustJSON(t, spec), &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]any)
	bodies := map[string][]string{
		"/api/v1/pull/transcripts":                  {"SchemaPullListResponse", "SchemaLegacyPullListResponse"},
		"/api/v1/pull/transcripts/{id}":             {"SchemaPullTranscriptInfo", "SchemaLegacyPullTranscriptInfo"},
		"/api/v1/pull/transcripts/{id}/annotations": {"SchemaPullAnnotation", "SchemaLegacyPullAnnotation"},
		"/api/v1/pull/transcripts/{id}/content":     nil,
	}
	for path, want := range bodies {
		get := paths[path].(map[string]any)["get"].(map[string]any)
		found := false
		for _, parameter := range get["parameters"].([]any) {
			p := parameter.(map[string]any)
			if p["in"] == "header" && p["name"] == schema.PullContractHeader {
				found = p["required"] != true
			}
		}
		if !found {
			t.Errorf("GET %s does not declare an optional %s request header", path, schema.PullContractHeader)
		}
		refusal, declared := get["responses"].(map[string]any)["406"].(map[string]any)
		if !declared || !schemaRefs(refusal["content"])["SchemaVillageErrorResponse"] {
			t.Errorf("GET %s does not declare the 406 VillageErrorResponse refusal for a version below the minimum", path)
		}
		ok := get["responses"].(map[string]any)["200"].(map[string]any)
		headers, _ := ok["headers"].(map[string]any)
		if _, present := headers[schema.PullContractHeader]; !present {
			t.Errorf("GET %s 200 does not declare the %s response header", path, schema.PullContractHeader)
		}
		if want == nil {
			continue
		}
		refs := schemaRefs(ok["content"])
		for _, name := range want {
			if !refs[name] {
				t.Errorf("GET %s 200 does not offer %s", path, name)
			}
		}
	}
	version := doc["components"].(map[string]any)["schemas"].(map[string]any)["SchemaSchemaVersionResponse"].(map[string]any)
	if _, ok := schemaProperties(version)["pullContractMaxVersion"]; !ok || strings.Contains(string(mustJSON(t, version["required"])), "pullContractMaxVersion") {
		t.Error("the version response must declare an optional pullContractMaxVersion")
	}
}

// TestPullResponsesAcceptBothEnvelopes validates bodies against the served
// document's 200 schema for each pull read, so a body that fits both
// envelopes (an empty page, no annotations) is accepted.
func TestPullResponsesAcceptBothEnvelopes(t *testing.T) {
	f := loadLegacyPullFixtures(t)
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	raw := mustJSON(t, spec)
	compiler := schema.NewJSONSchemaCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource("village.json", bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.ServedBodies.Cases {
		t.Run(c.Name, func(t *testing.T) {
			pointer := "village.json#/paths/" + strings.ReplaceAll(strings.ReplaceAll(c.Input.Path, "~", "~0"), "/", "~1") + "/get/responses/200/content/application~1json/schema"
			compiled, err := compiler.Compile(pointer)
			if err != nil {
				t.Fatalf("compile the served 200 schema for %s: %v", c.Input.Path, err)
			}
			var body any
			if err := json.Unmarshal([]byte(c.Input.Body), &body); err != nil {
				t.Fatal(err)
			}
			err = compiled.Validate(body)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", err == nil, err)
			}
		})
	}
}
