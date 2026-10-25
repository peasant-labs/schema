package schema_test

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_email_surface.yaml
var villageEmailSurfaceYAML []byte

type emailSurfaceFixture struct {
	AddressComponent string `yaml:"address_component"`
	Specs            map[string]struct {
		SelfOperations   []string          `yaml:"self_operations"`
		ExemptProperties map[string]string `yaml:"exempt_properties"`
	} `yaml:"specs"`
}

// TestNoResponseButOwnSettingsCarriesAnEmail walks every operation's
// responses in the served Village and Local API specs. Only the caller's own
// settings operations may reach the notification address, and no other
// response may declare a property named like an email unless the fixture
// exempts it with a reason. The village spec's self operations must reach the
// address, so the check cannot pass vacuously.
func TestNoResponseButOwnSettingsCarriesAnEmail(t *testing.T) {
	var f emailSurfaceFixture
	decoder := yaml.NewDecoder(strings.NewReader(string(villageEmailSurfaceYAML)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		t.Fatalf("decode email surface fixture: %v", err)
	}
	if f.AddressComponent == "" {
		t.Fatal("address_component is empty; the guard would have nothing to protect")
	}
	builders := map[string]func() (any, error){
		"village": func() (any, error) { return openapi.BuildVillageAPISpec() },
		"local":   func() (any, error) { return openapi.BuildPeasantLocalAPISpec() },
	}
	assertSameStrings(t, "specs the email guard walks", mapKeysAny(builders), mapKeysAny(f.Specs))

	for api, build := range builders {
		surface := f.Specs[api]
		t.Run(api, func(t *testing.T) {
			spec, err := build()
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(spec)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			components := doc["components"].(map[string]any)["schemas"].(map[string]any)
			self := map[string]bool{}
			for _, id := range surface.SelfOperations {
				self[id] = true
			}
			seenSelf := map[string]bool{}
			selfReachesAddress := false
			usedExemptions := map[string]bool{}
			for path, item := range doc["paths"].(map[string]any) {
				for method, rawOperation := range item.(map[string]any) {
					operation, ok := rawOperation.(map[string]any)
					if !ok {
						continue
					}
					id, _ := operation["operationId"].(string)
					reached := reachableComponents(components, schemaRefs(operation["responses"]))
					if self[id] {
						seenSelf[id] = true
						selfReachesAddress = selfReachesAddress || reached[f.AddressComponent]
						continue
					}
					label := strings.ToUpper(method) + " " + path + " (" + id + ")"
					if reached[f.AddressComponent] {
						t.Errorf("%s reaches %s; only the caller's own settings may carry an email address", label, f.AddressComponent)
					}
					for name := range reached {
						for property := range declaredPropertyNames(components[name]) {
							if !strings.Contains(strings.ToLower(property), "email") {
								continue
							}
							key := name + "." + property
							if _, exempt := surface.ExemptProperties[key]; exempt {
								usedExemptions[key] = true
								continue
							}
							t.Errorf("%s reaches %s, a property named like an email outside the caller's own settings", label, key)
						}
					}
				}
			}
			for _, id := range surface.SelfOperations {
				if !seenSelf[id] {
					t.Errorf("self operation %s is not declared; the exemption is stale", id)
				}
			}
			if len(surface.SelfOperations) > 0 && !selfReachesAddress {
				t.Errorf("no self operation reaches %s; the guard is vacuous", f.AddressComponent)
			}
			var unused []string
			for key := range surface.ExemptProperties {
				if !usedExemptions[key] {
					unused = append(unused, key)
				}
			}
			sort.Strings(unused)
			if len(unused) > 0 {
				t.Errorf("exempt properties no response reaches any more; remove them: %v", unused)
			}
		})
	}
}
