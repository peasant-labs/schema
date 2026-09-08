package schema_test

import (
	"embed"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_grouped_contract.yaml
var villageGroupedFixtureFS embed.FS

type villageRowInput struct {
	Variant      string `yaml:"variant"`
	SessionCount *int64 `yaml:"session_count"`
	VariantCount *int64 `yaml:"variant_count"`
	MismatchID   bool   `yaml:"mismatch_id"`
}
type villageItemInput struct {
	Kind       schema.SessionListItemKind `yaml:"kind"`
	Transcript bool                       `yaml:"transcript"`
	Context    bool                       `yaml:"context"`
}
type villageRouteInput struct {
	Path        string `yaml:"path"`
	OperationID string `yaml:"operation_id"`
	View        bool   `yaml:"view"`
	Scope       bool   `yaml:"scope"`
	Legacy      string `yaml:"legacy"`
	Grouped     string `yaml:"grouped"`
}
type villageExpected struct {
	Valid bool `yaml:"valid"`
}
type villageGroupedFixtures struct {
	Rows          testcase.Corpus[villageRowInput, villageExpected]   `yaml:"rows"`
	Items         testcase.Corpus[villageItemInput, villageExpected]  `yaml:"items"`
	Routes        testcase.Corpus[villageRouteInput, villageExpected] `yaml:"routes"`
	RequiredNames struct {
		Rows   []string `yaml:"rows"`
		Items  []string `yaml:"items"`
		Routes []string `yaml:"routes"`
	} `yaml:"required_names"`
}

func loadVillageGroupedFixtures(t *testing.T) villageGroupedFixtures {
	t.Helper()
	b, err := villageGroupedFixtureFS.ReadFile("testdata/village_grouped_contract.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var f villageGroupedFixtures
	if err := yaml.Unmarshal(b, &f); err != nil {
		t.Fatalf("load Village grouped contract fixtures: %v", err)
	}
	assertcase.RequireMin(t, f.Rows, 5)
	assertcase.RequireValid(t, f.Rows)
	assertcase.RequireMin(t, f.Items, 3)
	assertcase.RequireValid(t, f.Items)
	assertcase.RequireMin(t, f.Routes, 8)
	assertcase.RequireValid(t, f.Routes)
	requireNames(t, "rows", f.RequiredNames.Rows, namesOf(f.Rows))
	requireNames(t, "items", f.RequiredNames.Items, namesOf(f.Items))
	requireNames(t, "routes", f.RequiredNames.Routes, namesOf(f.Routes))
	return f
}

func namesOf[I, E any](c testcase.Corpus[I, E]) map[string]bool {
	m := map[string]bool{}
	for _, row := range c.Cases {
		m[row.Name] = true
	}
	return m
}
func requireNames(t *testing.T, arm string, required []string, actual map[string]bool) {
	t.Helper()
	for _, name := range required {
		if !actual[name] {
			t.Errorf("Village grouped %s corpus is missing required case %q", arm, name)
		}
	}
	if len(actual) != len(required) {
		t.Errorf("Village grouped %s exact case set differs: got %d names, required %d", arm, len(actual), len(required))
	}
}

func TestVillageSessionRowsPreserveRouteIdentityAndCounts(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	for _, c := range f.Rows.Cases {
		t.Run(c.Name, func(t *testing.T) {
			row := villageRow(c.Input)
			err := row.Validate()
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
		})
	}
}

func villageRow(in villageRowInput) schema.VillageSessionRow {
	id := schema.TranscriptID("123e4567-e89b-12d3-a456-426614174000")
	other := schema.TranscriptID("223e4567-e89b-12d3-a456-426614174000")
	owner := schema.VillageUUID("123e4567-e89b-12d3-a456-426614174001")
	local := schema.SessionID("123e4567-e89b-12d3-a456-426614174002")
	r := schema.VillageSessionRow{Session: schema.VillageTranscript{ID: id, OwnerID: owner, LocalID: local, InputSubmissionCount: in.SessionCount}}
	variantID := id
	if in.MismatchID {
		variantID = other
	}
	switch in.Variant {
	case "collective":
		r.Collective = &schema.VillageGroupTranscript{ID: variantID, OwnerID: owner, LocalID: local, InputSubmissionCount: in.VariantCount}
	case "pending":
		r.Pending = &schema.VillagePendingShare{TranscriptID: variantID, OwnerID: owner, LocalID: local, InputSubmissionCount: in.VariantCount}
	case "myShare":
		r.MyShare = &schema.VillageUserGroupShare{ID: variantID, OwnerID: owner, LocalID: local, InputSubmissionCount: in.VariantCount}
	case "contributable":
		r.Contributable = &schema.VillageContributableTranscript{ID: variantID, LocalID: local, InputSubmissionCount: in.VariantCount}
	}
	return r
}

func TestVillageSessionListItemDiscriminator(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	for _, c := range f.Items.Cases {
		t.Run(c.Name, func(t *testing.T) {
			item := schema.VillageSessionListItem{Kind: c.Input.Kind}
			if c.Input.Transcript {
				row := villageRow(villageRowInput{})
				item.Transcript = &row
			}
			if c.Input.Context {
				item.Context = &schema.HelperContextSummary{GroupID: "hg_fixture", OwnerStatus: schema.RelationshipNavigationUnknown}
			}
			err := item.Validate()
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
		})
	}
}

func TestVillageGroupedRoutesPreserveLegacyAndSelectedShapes(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(spec)
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]any)
	for _, c := range f.Routes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			op := paths[c.Input.Path].(map[string]any)["get"].(map[string]any)
			if op["operationId"] != c.Input.OperationID {
				t.Fatalf("operationId=%v", op["operationId"])
			}
			raw, _ := json.Marshal(op)
			text := string(raw)
			if c.Input.View && !contains(text, "\"name\":\"view\"") {
				t.Error("missing view query")
			}
			if c.Input.Scope && !contains(text, "\"name\":\"scope\"") {
				t.Error("missing scope query")
			}
			if c.Input.Legacy != "" && !contains(text, c.Input.Legacy) {
				t.Errorf("missing legacy shape %s", c.Input.Legacy)
			}
			if !contains(text, c.Input.Grouped) {
				t.Errorf("missing selected shape %s", c.Input.Grouped)
			}
			if c.Name == "durable-content" && contains(text, "RelationshipNavigation") {
				t.Error("durable content route exposes read navigation")
			}
		})
	}
}

func contains(s, needle string) bool { return strings.Contains(s, needle) }
