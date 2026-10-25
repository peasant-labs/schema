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
	// NameTerms are the substrings that mark a property name as contact
	// data; Formats are the string formats that do.
	NameTerms []string `yaml:"name_terms"`
	Formats   []string `yaml:"formats"`
	Specs     map[string]struct {
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
	if len(f.NameTerms) == 0 || len(f.Formats) == 0 {
		t.Fatal("name_terms and formats must both be non-empty; the guard would match nothing")
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
						for property, propertySchema := range declaredProperties(components[name]) {
							if !looksLikeContactAddress(property, propertySchema, f.NameTerms, f.Formats) {
								continue
							}
							key := name + "." + property
							if _, exempt := surface.ExemptProperties[key]; exempt {
								usedExemptions[key] = true
								continue
							}
							t.Errorf("%s reaches %s, a property named or formatted like an email address outside the caller's own settings", label, key)
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

// declaredProperties maps every property a component declares, including
// those inside allOf, oneOf, anyOf, items, and inline objects, to its schema.
// A name declared twice keeps both schemas reachable through a slice.
func declaredProperties(value any) map[string][]any {
	properties := map[string][]any{}
	var walk func(any)
	walk = func(v any) {
		switch typed := v.(type) {
		case map[string]any:
			for key, child := range typed {
				if declared, ok := child.(map[string]any); ok && key == "properties" {
					for name, property := range declared {
						properties[name] = append(properties[name], property)
						walk(property)
					}
					continue
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return properties
}

// looksLikeContactAddress reports whether a property's name contains one of
// the contact terms, or any of its schemas (including nested anyOf and oneOf
// arms, but not referenced components, which are walked on their own) carries
// one of the contact formats.
func looksLikeContactAddress(name string, schemas []any, terms, formats []string) bool {
	lower := strings.ToLower(name)
	for _, term := range terms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	found := false
	var walk func(any)
	walk = func(v any) {
		switch typed := v.(type) {
		case map[string]any:
			if format, ok := typed["format"].(string); ok {
				for _, want := range formats {
					if format == want {
						found = true
					}
				}
			}
			for key, child := range typed {
				if key == "properties" {
					continue
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	for _, schema := range schemas {
		walk(schema)
	}
	return found
}
