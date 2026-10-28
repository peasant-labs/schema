package schema_test

import (
	"bytes"
	_ "embed"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"gopkg.in/yaml.v3"
)

func TestVillageIntegrityClosedSets(t *testing.T) {
	var want []string
	for _, enum := range loadClosedEnumFixture(t).Enums {
		if enum.Name == "VillageIntegrityCheckResult" {
			want = enum.AllValues
		}
	}
	if len(want) == 0 {
		t.Fatal("enums.yaml does not define VillageIntegrityCheckResult")
	}
	got := make([]string, 0, len(schema.AllVillageIntegrityCheckResults))
	for _, result := range schema.AllVillageIntegrityCheckResults {
		if !result.IsValid() {
			t.Fatalf("%q is listed but not valid", result)
		}
		got = append(got, result.String())
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AllVillageIntegrityCheckResults = %v, want %v", got, want)
	}
	if schema.VillageIntegrityCheckResult("clean").IsValid() || schema.VillageIntegrityCheckResult("").IsValid() {
		t.Fatal("a result outside the closed set is valid")
	}
}

//go:embed testdata/integrity/integrity_surfaces.yaml
var integritySurfacesYAML []byte

type integritySurfaces struct {
	IntegrityCarriers []string            `yaml:"integrity_carriers"`
	ExactProperties   map[string][]string `yaml:"exact_properties"`
	ContentHashHeader struct {
		Name string `yaml:"name"`
	} `yaml:"content_hash_header"`
	DescriptionAnchors map[string][]string `yaml:"description_anchors"`
}

func loadIntegritySurfaces(t *testing.T) integritySurfaces {
	t.Helper()
	var f integritySurfaces
	d := yaml.NewDecoder(bytes.NewReader(integritySurfacesYAML))
	d.KnownFields(true)
	if err := d.Decode(&f); err != nil {
		t.Fatal(err)
	}
	return f
}

func integritySchemas(h *boundaryHarness, api string) map[string]map[string]any {
	out := map[string]map[string]any{}
	raw, _ := h.documents[api]["components"].(map[string]any)["schemas"].(map[string]any)
	for name, value := range raw {
		if m, ok := value.(map[string]any); ok {
			out[strings.TrimPrefix(name, "Schema")] = m
		}
	}
	return out
}

func integrityOperation(t *testing.T, h *boundaryHarness, method, path string) map[string]any {
	t.Helper()
	paths, _ := h.documents["village"]["paths"].(map[string]any)
	item, _ := paths[path].(map[string]any)
	operation, ok := item[method].(map[string]any)
	if !ok {
		t.Fatalf("village %s %s is not declared", strings.ToUpper(method), path)
	}
	return operation
}

func requireAnchors(t *testing.T, label, text string, anchors []string) {
	t.Helper()
	if len(anchors) == 0 {
		t.Fatalf("%s has no anchors in the fixture", label)
	}
	for _, anchor := range anchors {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(anchor)) {
			t.Errorf("%s does not contain %q: %s", label, anchor, text)
		}
	}
}

func TestIntegritySurfacesAreExact(t *testing.T) {
	f := loadIntegritySurfaces(t)
	h := newBoundaryHarness(t)
	types := integritySchemas(h, "types")

	carriers := map[string]bool{}
	for _, name := range f.IntegrityCarriers {
		carriers[name] = true
	}
	for _, api := range []string{"types", "village"} {
		for name, component := range integritySchemas(h, api) {
			props, _ := component["properties"].(map[string]any)
			if _, ok := props["integrity"]; ok && !carriers[name] {
				t.Errorf("%s component %s has an integrity property but is not an integrity carrier", api, name)
			}
		}
	}

	for name, want := range f.ExactProperties {
		component, ok := types[name]
		if !ok {
			t.Errorf("Types has no component %s", name)
			continue
		}
		props, _ := component["properties"].(map[string]any)
		got := make([]string, 0, len(props))
		for prop := range props {
			got = append(got, prop)
		}
		requireSameSet(t, name+" properties", want, got)
		if closed, ok := component["additionalProperties"].(bool); !ok || closed {
			t.Errorf("%s must declare additionalProperties false, got %v", name, component["additionalProperties"])
		}
	}

	content := integrityOperation(t, h, "get", "/api/v1/transcripts/{id}/content")
	response, _ := content["responses"].(map[string]any)["200"].(map[string]any)
	headers, _ := response["headers"].(map[string]any)
	header, ok := headers[f.ContentHashHeader.Name].(map[string]any)
	if !ok {
		t.Fatalf("the content 200 response does not declare the %s header", f.ContentHashHeader.Name)
	}
	if required, _ := header["required"].(bool); required {
		t.Errorf("%s must be optional: village sends no header while it holds no hash for the served copy", f.ContentHashHeader.Name)
	}
	ref, _ := header["schema"].(map[string]any)["$ref"].(string)
	hashName := strings.TrimPrefix(strings.TrimPrefix(ref, "#/components/schemas/"), "Schema")
	if hashName != "TranscriptContentHash" {
		t.Errorf("%s schema references %q, want the TranscriptContentHash component", f.ContentHashHeader.Name, ref)
	}
	village := integritySchemas(h, "village")
	if pattern, _ := village[hashName]["pattern"].(string); pattern == "" {
		t.Errorf("TranscriptContentHash carries no pattern")
	}
	description, _ := header["description"].(string)
	requireAnchors(t, "content hash header description", description, []string{"bare", "not an ETag", "stored copy", "absent"})

	pull := integrityOperation(t, h, "get", "/api/v1/pull/transcripts/{id}/publish-metadata")
	pullResponse, _ := pull["responses"].(map[string]any)["200"].(map[string]any)
	pullHeaders, _ := pullResponse["headers"].(map[string]any)
	if _, ok := pullHeaders[schema.PullContractHeader]; !ok {
		t.Errorf("the publish metadata 200 response does not declare the %s response header", schema.PullContractHeader)
	}
	pullDescription, _ := pull["description"].(string)
	requireAnchors(t, "publish metadata description", pullDescription, f.DescriptionAnchors["publish_metadata_operation"])

	check := integrityOperation(t, h, "post", "/api/v1/transcripts/{id}/integrity-check")
	checkDescription, _ := check["description"].(string)
	requireAnchors(t, "integrity check description", checkDescription, f.DescriptionAnchors["integrity_check_operation"])

	metadata := integrityOperation(t, h, "get", "/api/v1/transcripts/{id}")
	metadataDescription, _ := metadata["description"].(string)
	requireAnchors(t, "transcript metadata description", metadataDescription, f.DescriptionAnchors["transcript_metadata_operation"])

	for _, ref := range [][2]string{{"ScryCheckResponse", "flaggedTurns"}, {"VillageTranscriptIntegrity", "flagged_turns"}} {
		props, _ := types[ref[0]]["properties"].(map[string]any)
		prop, _ := props[ref[1]].(map[string]any)
		text, _ := prop["description"].(string)
		requireAnchors(t, ref[0]+"."+ref[1]+" description", text, f.DescriptionAnchors["flagged_turn_properties"])
	}
}
