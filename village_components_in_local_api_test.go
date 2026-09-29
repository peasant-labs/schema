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

// TestLocalAPIPinsEmbeddedVillageComponents notices a Village change that
// rewrites the current Local API document. The Local API embeds Village
// catalog rows, and harmonization copies each one from the Types catalog, so a
// Village-only change would otherwise mutate a released Local spec with every
// other gate green. It cannot tell a released version from an unreleased one;
// its message says which fix applies. A gate that compares each current spec
// with its bytes at the last release tag would replace it.
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
	var changed []string
	for name, sum := range actual {
		if pinned[name] != sum {
			changed = append(changed, name)
		}
	}
	for name := range pinned {
		if _, ok := actual[name]; !ok {
			changed = append(changed, name)
		}
	}
	sort.Strings(changed)
	if len(changed) > 0 || len(pinned) != len(fixture.Components) {
		// The new sums are printed without a local_version line on purpose:
		// under an unchanged version, a pasted block would rewrite a released
		// Local spec in place, so the reader decides which case applies.
		var sums []string
		for _, name := range changed {
			sums = append(sums, fmt.Sprintf("%s: %s", name, actual[name]))
		}
		t.Fatalf("Village rows embedded in Local API %s changed: %v. If Local API %s is released, bump PeasantLocalAPIVersion first; the test then prints the pins for the new version. Only a version that is not yet released may be re-pinned in place, with these sums:\n%s", schema.PeasantLocalAPIVersion, changed, schema.PeasantLocalAPIVersion, strings.Join(sums, "\n"))
	}
}
