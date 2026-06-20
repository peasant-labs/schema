package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/peasant-labs/schema"
)

func loadPublishVerdicts(t *testing.T) *schema.PublishVerdictFixtures {
	t.Helper()
	f, err := schema.LoadPublishVerdictFixtures()
	if err != nil {
		t.Fatalf("LoadPublishVerdictFixtures: %v", err)
	}
	return f
}

func TestPublishVerdictFixtures_Structure(t *testing.T) {
	f := loadPublishVerdicts(t)
	if len(f.Cases) == 0 {
		t.Fatal("publish verdict corpus is empty")
	}

	seen := map[string]bool{}
	for _, c := range f.Cases {
		if c.Name == "" {
			t.Error("publish verdict case has empty name")
		}
		if seen[c.Name] {
			t.Errorf("duplicate publish verdict case name %q", c.Name)
		}
		seen[c.Name] = true
		if c.Body == "" {
			t.Errorf("case %q has empty body", c.Name)
		}
		var doc any
		if err := json.Unmarshal([]byte(c.Body), &doc); err != nil {
			t.Errorf("case %q body is not valid JSON: %v", c.Name, err)
		}
		if c.Expect.ErrorCategory == "" {
			t.Errorf("case %q has no error_category", c.Name)
		}
		if !c.Expect.SchemaAccepts && c.Expect.HTTPStatus == 0 {
			t.Errorf("reject case %q has no http_status", c.Name)
		}
		if c.Expect.LegacyAccepts != nil && *c.Expect.LegacyAccepts == c.Expect.SchemaAccepts {
			t.Errorf("case %q sets legacy_accepts but it does not diverge from schema_accepts", c.Name)
		}
	}
}

func TestPublishVerdictFixtures_CoverageFloors(t *testing.T) {
	f := loadPublishVerdicts(t)
	if len(f.Cases) < 17 {
		t.Errorf("publish verdict cases = %d, want >= 17", len(f.Cases))
	}

	categories := map[string]int{}
	accepted := 0
	rejected := 0
	for _, c := range f.Cases {
		categories[c.Expect.ErrorCategory]++
		if c.Expect.SchemaAccepts {
			accepted++
		} else {
			rejected++
		}
	}
	if accepted < 3 {
		t.Errorf("accepted cases = %d, want >= 3", accepted)
	}
	if rejected < 14 {
		t.Errorf("rejected cases = %d, want >= 14", rejected)
	}
	for _, cat := range []string{"accepted", "enum", "type", "pattern"} {
		if categories[cat] == 0 {
			t.Errorf("publish verdict corpus has no %q cases", cat)
		}
	}
}

func TestPublishVerdictFixtures_UnknownHarnessRow(t *testing.T) {
	f := loadPublishVerdicts(t)
	c, ok := f.CaseByName("unknown-harness")
	if !ok {
		t.Fatal("publish verdict corpus missing unknown-harness row")
	}
	if c.Expect.SchemaAccepts {
		t.Error("unknown-harness must be a schema reject")
	}
	if c.Expect.HTTPStatus != 422 {
		t.Errorf("unknown-harness http_status = %d, want 422", c.Expect.HTTPStatus)
	}
	harness, ok, err := c.ModelHarness()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || harness == "" {
		t.Fatal("unknown-harness body must carry model.harness")
	}
	if schema.Harness(harness).IsKnown() {
		t.Errorf("unknown-harness body uses known harness %q", harness)
	}
}
