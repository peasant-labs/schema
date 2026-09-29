package schema_test

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/village_components.yaml
var villageComponentsInLocalYAML []byte

type villageComponentsInLocal struct {
	LocalVersion string                    `yaml:"local_version"`
	Components   []villageComponentInLocal `yaml:"components"`
}

type villageComponentInLocal struct {
	Name   string `yaml:"name"`
	SHA256 string `yaml:"sha256"`
}

// TestLocalAPIPinsEmbeddedVillageComponents keeps a Village change from
// rewriting the current Local API document without a Local version bump. The
// Local API embeds Village catalog rows, and harmonization copies each one
// from the Types catalog, so a Village-only change would otherwise mutate a
// released Local spec with every other gate green.
func TestLocalAPIPinsEmbeddedVillageComponents(t *testing.T) {
	var fixture villageComponentsInLocal
	decoder := yaml.NewDecoder(bytes.NewReader(villageComponentsInLocalYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode Village component pins: %v", err)
	}
	spec, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	actual := map[string]string{}
	for name, component := range spec.Components.Schemas {
		canonical := strings.TrimPrefix(name, "Schema")
		if !strings.HasPrefix(canonical, "Village") {
			continue
		}
		encoded, err := json.Marshal(component)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(encoded)
		actual[canonical] = hex.EncodeToString(sum[:])
	}
	var current []string
	for name, sum := range actual {
		current = append(current, fmt.Sprintf("  - {name: %s, sha256: %s}", name, sum))
	}
	sort.Strings(current)
	repin := fmt.Sprintf("local_version: %s\ncomponents:\n%s", schema.PeasantLocalAPIVersion, strings.Join(current, "\n"))
	if fixture.LocalVersion != schema.PeasantLocalAPIVersion {
		t.Fatalf("the Village component pins are for Local API %s but the current version is %s; after a Local bump, re-pin testdata/local-api/village_components.yaml:\n%s", fixture.LocalVersion, schema.PeasantLocalAPIVersion, repin)
	}
	pinned := map[string]string{}
	for _, component := range fixture.Components {
		pinned[component.Name] = component.SHA256
	}
	if len(pinned) != len(fixture.Components) || len(pinned) != len(actual) {
		t.Fatalf("the Local API embeds %d Village components and the fixture pins %d; a Village row entered or left the Local API, so bump PeasantLocalAPIVersion and re-pin:\n%s", len(actual), len(fixture.Components), repin)
	}
	for name, sum := range actual {
		if pinned[name] != sum {
			t.Fatalf("Village component %s in the Local API changed under Local API %s; a released Local spec must not change in place, so bump PeasantLocalAPIVersion and re-pin:\n%s", name, schema.PeasantLocalAPIVersion, repin)
		}
	}
}
