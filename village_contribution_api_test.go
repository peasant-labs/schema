package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/route"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_contribution_api.yaml
var villageContributionYAML []byte

type contributionRawInput struct {
	Raw string `yaml:"raw"`
}
type contributionBoundInput struct {
	Constant string `yaml:"constant"`
}
type contributionShapeInput struct {
	Target string         `yaml:"target"`
	JSON   map[string]any `yaml:"json"`
}
type contributionExpected struct {
	Valid         bool     `yaml:"valid"`
	Value         int      `yaml:"value"`
	Present       []string `yaml:"present"`
	Absent        []string `yaml:"absent"`
	NestedPresent []string `yaml:"nested_present"`
	NestedAbsent  []string `yaml:"nested_absent"`
}
type contributionFixtures struct {
	RequiredNames struct {
		Views      []string `yaml:"views"`
		ErrorCodes []string `yaml:"error_codes"`
		Bounds     []string `yaml:"bounds"`
		Shapes     []string `yaml:"shapes"`
	} `yaml:"required_names"`
	Views      testcase.Corpus[contributionRawInput, contributionExpected]   `yaml:"views"`
	ErrorCodes testcase.Corpus[contributionRawInput, contributionExpected]   `yaml:"error_codes"`
	Bounds     testcase.Corpus[contributionBoundInput, contributionExpected] `yaml:"bounds"`
	Shapes     testcase.Corpus[contributionShapeInput, contributionExpected] `yaml:"shapes"`
}

func loadContributionFixtures(t *testing.T) contributionFixtures {
	t.Helper()
	var f contributionFixtures
	if err := yaml.Unmarshal(villageContributionYAML, &f); err != nil {
		t.Fatalf("load contribution fixtures: %v", err)
	}
	assertcase.RequireMin(t, f.Views, 8)
	assertcase.RequireValid(t, f.Views)
	assertcase.RequireMin(t, f.ErrorCodes, 14)
	assertcase.RequireValid(t, f.ErrorCodes)
	assertcase.RequireMin(t, f.Bounds, 5)
	assertcase.RequireValid(t, f.Bounds)
	assertcase.RequireMin(t, f.Shapes, 17)
	assertcase.RequireValid(t, f.Shapes)
	requireNames(t, "contributable view", f.RequiredNames.Views, namesOf(f.Views))
	requireNames(t, "contribution error code", f.RequiredNames.ErrorCodes, namesOf(f.ErrorCodes))
	requireNames(t, "contribution page bound", f.RequiredNames.Bounds, namesOf(f.Bounds))
	requireNames(t, "contribution shape", f.RequiredNames.Shapes, namesOf(f.Shapes))
	return f
}

func TestVillageContributableViewMenu(t *testing.T) {
	f := loadContributionFixtures(t)
	members := map[string]bool{}
	for _, c := range f.Views.Cases {
		if c.Expected.Valid {
			members[c.Input.Raw] = true
		}
	}
	if len(schema.AllVillageContributableViews) != len(members) {
		t.Errorf("menu has %d members, the corpus names %d valid tokens", len(schema.AllVillageContributableViews), len(members))
	}
	for _, view := range schema.AllVillageContributableViews {
		if !members[string(view)] {
			t.Errorf("menu member %q has no valid corpus case", view)
		}
	}
	for _, c := range f.Views.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := schema.NewVillageContributableView(c.Input.Raw)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
			if c.Expected.Valid && (!got.IsValid() || string(got) != c.Input.Raw) {
				t.Fatalf("constructor returned %q for %q", got, c.Input.Raw)
			}
			if schema.VillageContributableView(c.Input.Raw).IsValid() != c.Expected.Valid {
				t.Fatalf("IsValid disagrees with the constructor for %q", c.Input.Raw)
			}
		})
	}
}

func TestVillageContributionErrorCodeMenu(t *testing.T) {
	f := loadContributionFixtures(t)
	members := map[string]bool{}
	for _, c := range f.ErrorCodes.Cases {
		if c.Expected.Valid {
			members[c.Input.Raw] = true
		}
	}
	if len(schema.AllVillageContributionErrorCodes) != len(members) {
		t.Errorf("menu has %d members, the corpus names %d valid tokens", len(schema.AllVillageContributionErrorCodes), len(members))
	}
	for _, code := range schema.AllVillageContributionErrorCodes {
		if !members[string(code)] {
			t.Errorf("menu member %q has no valid corpus case", code)
		}
	}
	for _, c := range f.ErrorCodes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := schema.NewVillageContributionErrorCode(c.Input.Raw)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
			if c.Expected.Valid && (!got.IsValid() || string(got) != c.Input.Raw) {
				t.Fatalf("constructor returned %q for %q", got, c.Input.Raw)
			}
		})
	}
}

// contributionBoundConstants maps the constant names the bounds corpus uses to
// the exported Go constants.
func contributionBoundConstants() map[string]int {
	return map[string]int{
		"default":                   schema.VillageContributionPageLimitDefault,
		"minimum":                   schema.VillageContributionPageLimitMinimum,
		"maximum":                   schema.VillageContributionPageLimitMaximum,
		"snapshot_max_bytes":        schema.VillageContributionSnapshotMaxBytes,
		"selection_max_transcripts": schema.VillageContributionSelectionMaxTranscripts,
	}
}

// TestVillageContributionPageLimitMatchesDeclaredRange holds the Go constants,
// the frozen product values and the spec's limit parameters equal, the way the
// transcript window span constants are held.
func TestVillageContributionPageLimitMatchesDeclaredRange(t *testing.T) {
	f := loadContributionFixtures(t)
	constants := contributionBoundConstants()
	for _, c := range f.Bounds.Cases {
		got, ok := constants[c.Input.Constant]
		if !ok {
			t.Fatalf("case %s names unknown constant %q", c.Name, c.Input.Constant)
		}
		if got != c.Expected.Value {
			t.Errorf("%s: constant = %d, want %d", c.Name, got, c.Expected.Value)
		}
	}
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	// A path item also carries members that are not operations, such as
	// parameters, so only each read's get operation is decoded.
	var doc struct {
		Paths map[string]struct {
			Get struct {
				Parameters []struct {
					Name   string `json:"name"`
					Schema struct {
						Minimum *float64 `json:"minimum"`
						Maximum *float64 `json:"maximum"`
						Default *float64 `json:"default"`
					} `json:"schema"`
				} `json:"parameters"`
			} `json:"get"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, read := range []route.Route{
		route.VillageListContributableTranscripts,
		route.VillageListContributableTranscriptsDeprecated,
		route.VillageListContributionProjects,
		route.VillageListContributionProjectsDeprecated,
	} {
		found := false
		for _, parameter := range doc.Paths[read.Pattern()].Get.Parameters {
			if parameter.Name != contributionLimitParameter {
				continue
			}
			found = true
			for _, bound := range []struct {
				name string
				got  *float64
				want int
			}{
				{"minimum", parameter.Schema.Minimum, constants["minimum"]},
				{"maximum", parameter.Schema.Maximum, constants["maximum"]},
				{"default", parameter.Schema.Default, constants["default"]},
			} {
				if bound.got == nil || int(*bound.got) != bound.want {
					t.Errorf("%s limit %s = %v, want %d", read.OperationID(), bound.name, bound.got, bound.want)
				}
			}
		}
		if !found {
			t.Fatalf("%s declares no limit parameter", read.OperationID())
		}
	}
}

// contributionLimitParameter is the page-size query parameter of the paged
// contribution reads.
const contributionLimitParameter = "limit"

// contributionTargets maps the Go type name a shape case names to a new value
// of that type.
var contributionTargets = map[string]func() any{
	"VillageContributionPageResponse":     func() any { return &schema.VillageContributionPageResponse{} },
	"VillageContributionSnapshotResponse": func() any { return &schema.VillageContributionSnapshotResponse{} },
	"VillageContributionSnapshotRecord":   func() any { return &schema.VillageContributionSnapshotRecord{} },
	"VillageContributionProjectPage":      func() any { return &schema.VillageContributionProjectPage{} },
	"VillageContributionErrorResponse":    func() any { return &schema.VillageContributionErrorResponse{} },
	"VillageBatchShareSelectionRequest":   func() any { return &schema.VillageBatchShareSelectionRequest{} },
}

func contributionHas(node any, path string) bool {
	for _, part := range strings.Split(path, ".") {
		switch v := node.(type) {
		case map[string]any:
			next, ok := v[part]
			if !ok {
				return false
			}
			node = next
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(v) {
				return false
			}
			node = v[i]
		default:
			return false
		}
	}
	return true
}

func TestVillageContributionWireShapes(t *testing.T) {
	f := loadContributionFixtures(t)
	for _, c := range f.Shapes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			newTarget, ok := contributionTargets[c.Input.Target]
			if !ok {
				t.Fatalf("unknown target %q", c.Input.Target)
			}
			target := newTarget()
			raw, err := json.Marshal(c.Input.JSON)
			if err != nil {
				t.Fatal(err)
			}
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.DisallowUnknownFields()
			err = decoder.Decode(target)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
			if err != nil {
				return
			}
			out, err := json.Marshal(target)
			if err != nil {
				t.Fatal(err)
			}
			var emitted map[string]any
			if err := json.Unmarshal(out, &emitted); err != nil {
				t.Fatal(err)
			}
			for _, path := range append(append([]string{}, c.Expected.Present...), c.Expected.NestedPresent...) {
				if !contributionHas(emitted, path) {
					t.Errorf("emitted JSON lacks %s: %s", path, out)
				}
			}
			for _, path := range append(append([]string{}, c.Expected.Absent...), c.Expected.NestedAbsent...) {
				if contributionHas(emitted, path) {
					t.Errorf("emitted JSON carries %s: %s", path, out)
				}
			}
		})
	}
}

// TestVillageContributionSelectionBoundMatchesDeclaredRange holds the selection
// bound constant equal to the maxItems of both arrays of the selection request,
// and the snapshot size constant equal to the limit the operation and the
// snapshot response state in prose.
func TestVillageContributionSelectionBoundMatchesDeclaredRange(t *testing.T) {
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	type described struct {
		MaxItems    *float64 `json:"maxItems"`
		Description string   `json:"description"`
	}
	var doc struct {
		Components struct {
			Schemas struct {
				SelectionRequest struct {
					Properties map[string]described `json:"properties"`
				} `json:"SchemaVillageBatchShareSelectionRequest"`
				Snapshot described `json:"SchemaVillageContributionSnapshotResponse"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"transcript_ids", "private_approvals"} {
		property, ok := doc.Components.Schemas.SelectionRequest.Properties[name]
		if !ok {
			t.Fatalf("selection request declares no %s", name)
		}
		if property.MaxItems == nil || int(*property.MaxItems) != schema.VillageContributionSelectionMaxTranscripts {
			t.Errorf("%s maxItems = %v, want %d", name, property.MaxItems, schema.VillageContributionSelectionMaxTranscripts)
		}
		if !strings.Contains(property.Description, strconv.Itoa(schema.VillageContributionSelectionMaxTranscripts)) {
			t.Errorf("%s description does not state the bound: %q", name, property.Description)
		}
	}
	if schema.VillageContributionSnapshotMaxBytes != 8*1024*1024 {
		t.Errorf("snapshot limit = %d bytes, want 8 MiB", schema.VillageContributionSnapshotMaxBytes)
	}
	statement := "limited to 8 MiB (" + strconv.Itoa(schema.VillageContributionSnapshotMaxBytes) + " bytes) of response body"
	if description := doc.Components.Schemas.Snapshot.Description; !strings.Contains(description, statement) || !strings.Contains(description, string(schema.VillageContributionErrorSnapshotTooLarge)) {
		t.Errorf("snapshot response description does not state the limit and its code: %q", description)
	}
}
