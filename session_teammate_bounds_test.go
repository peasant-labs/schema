package schema_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// TestSessionTeammateBoundsParity runs the shared identity corpus through the
// Go predicate and the published JSON Schema. The TypeScript suite runs the same
// file through the generated Zod validator, so one corpus holds all three
// layers to the same code-point measure of length.
func TestSessionTeammateBoundsParity(t *testing.T) {
	fixtures, err := schema.LoadSessionTeammateBoundsFixtures()
	if err != nil {
		t.Fatalf("LoadSessionTeammateBoundsFixtures: %v", err)
	}

	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("BuildTypesSpec: %v", err)
	}
	document, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Types document: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("types.json", bytes.NewReader(document)); err != nil {
		t.Fatalf("add Types document resource: %v", err)
	}
	compiled, err := compiler.Compile("types.json#/components/schemas/SessionTeammate")
	if err != nil {
		t.Fatalf("compile SessionTeammate from the Types document: %v", err)
	}

	// A corpus of short names would pass while the metric stayed wrong, so the
	// corpus must carry at least one value inside the code-point bound and over
	// the same bound counted in UTF-16 units.
	metricProbe := 0
	for _, c := range fixtures.Cases.Cases {
		if c.Input.CodePoints <= fixtures.Bound && c.Input.UTF16Units > fixtures.Bound {
			metricProbe++
		}
	}
	if metricProbe == 0 {
		t.Fatalf("the corpus carries no case inside the %d-code-point bound but over it in UTF-16 units; without one the astral-length regression this corpus guards cannot fail", fixtures.Bound)
	}

	for _, c := range fixtures.Cases.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := utf8.RuneCountInString(c.Input.Name); got != c.Input.CodePoints {
				t.Fatalf("fixture declares %d code points but the name carries %d; the metric assertion would compare the wrong quantity", c.Input.CodePoints, got)
			}
			if got := len(utf16.Encode([]rune(c.Input.Name))); got != c.Input.UTF16Units {
				t.Fatalf("fixture declares %d UTF-16 units but the name carries %d", c.Input.UTF16Units, got)
			}

			teammate := schema.SessionTeammate{
				Team:     c.Input.Team,
				Name:     c.Input.Name,
				Evidence: schema.SessionTeammateEvidence(c.Input.Evidence),
			}
			if got := teammate.Validate() == nil; got != c.Expected.Valid {
				t.Fatalf("Go Validate accepted=%t, want %t", got, c.Expected.Valid)
			}

			raw, err := json.Marshal(map[string]any{
				"team":     c.Input.Team,
				"name":     c.Input.Name,
				"evidence": c.Input.Evidence,
			})
			if err != nil {
				t.Fatalf("marshal identity: %v", err)
			}
			var instance any
			if err := json.Unmarshal(raw, &instance); err != nil {
				t.Fatalf("unmarshal identity: %v", err)
			}
			if got := compiled.Validate(instance) == nil; got != c.Expected.Valid {
				t.Fatalf("published JSON Schema accepted=%t, want %t", got, c.Expected.Valid)
			}
		})
	}
}
