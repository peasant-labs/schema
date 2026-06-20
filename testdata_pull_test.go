package schema_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
)

// These are the schema-module META-tests over the pull fixtures: they exercise
// the loader + generators (fixture STRUCTURE, coverage floors, disjointness) and
// pin the golden manifest's JSON shape. The behavioural assertions that drive the
// real pull.ParseTranscriptRef live in internal/pull/types_test.go.

func loadPullRefs(t *testing.T) *schema.PullRefFixtures {
	t.Helper()
	f, err := schema.LoadPullRefFixtures()
	if err != nil {
		t.Fatalf("LoadPullRefFixtures: %v", err)
	}
	return f
}

func TestPullRefFixtures_Structure(t *testing.T) {
	f := loadPullRefs(t)

	if f.Values.UUIDLower == "" {
		t.Error("values.uuid_lower is empty")
	}
	if f.Values.VillageHost == "" {
		t.Error("values.village_host is empty")
	}
	if len(f.IDCasings) == 0 {
		t.Error("id_casings axis is empty")
	}
	if len(f.RefForms) == 0 {
		t.Error("ref_forms axis is empty")
	}
	if len(f.ValidGeneration.Axes) != 2 {
		t.Errorf("valid_generation.axes = %v, want exactly [id_casings ref_forms]", f.ValidGeneration.Axes)
	}

	// Every invalid category must carry an expectation: either a category-level
	// err_contains, or one on every case.
	for _, cat := range f.InvalidRefs {
		if len(cat.Cases) == 0 {
			t.Errorf("invalid_refs category %q has no cases", cat.Category)
		}
		for _, sub := range cat.Cases {
			eff := cat.ErrContains
			if sub.ErrContains != "" {
				eff = sub.ErrContains
			}
			if eff == "" {
				t.Errorf("invalid_refs %q/%q has no err_contains (category or case)", cat.Category, sub.Name)
			}
		}
	}
	// Every lookalike must carry an expectation.
	for _, look := range f.NegativeLookalike {
		if look.ErrContains == "" {
			t.Errorf("negative_lookalike %q has no err_contains", look.Name)
		}
	}

	// Templates may only use the known {id}/{host} placeholders.
	known := map[string]bool{"{id}": true, "{host}": true}
	checkPlaceholders := func(ctx, tmpl string) {
		for _, ph := range extractPlaceholders(tmpl) {
			if !known[ph] {
				t.Errorf("%s template %q uses unknown placeholder %q", ctx, tmpl, ph)
			}
		}
	}
	for _, form := range f.RefForms {
		checkPlaceholders("ref_form "+form.Name, form.Template)
	}
	for _, cat := range f.InvalidRefs {
		for _, sub := range cat.Cases {
			if sub.Template != "" {
				checkPlaceholders("invalid "+cat.Category+"/"+sub.Name, sub.Template)
			}
		}
	}
	for _, look := range f.NegativeLookalike {
		checkPlaceholders("lookalike "+look.Name, look.Template)
	}
}

// extractPlaceholders returns all {...} tokens in a template.
func extractPlaceholders(tmpl string) []string {
	var out []string
	for {
		i := strings.IndexByte(tmpl, '{')
		if i < 0 {
			return out
		}
		j := strings.IndexByte(tmpl[i:], '}')
		if j < 0 {
			return out
		}
		out = append(out, tmpl[i:i+j+1])
		tmpl = tmpl[i+j+1:]
	}
}

func TestPullRefFixtures_CoverageFloors(t *testing.T) {
	f := loadPullRefs(t)

	valid, err := f.ValidCases()
	if err != nil {
		t.Fatalf("ValidCases: %v", err)
	}
	if len(valid) < 15 {
		t.Errorf("ValidCases = %d, want >= 15 (3 casings x 5 forms)", len(valid))
	}

	invalid, err := f.InvalidCases()
	if err != nil {
		t.Fatalf("InvalidCases: %v", err)
	}
	if len(invalid) < 9 {
		t.Errorf("InvalidCases = %d, want >= 9", len(invalid))
	}

	// Every valid case must declare a normalized lowercase want_id and a non-empty
	// input.
	wantID := f.Values.UUIDLower
	for _, c := range valid {
		if c.WantErr {
			t.Errorf("valid case %q marked WantErr", c.Name)
		}
		if c.WantID != wantID {
			t.Errorf("valid case %q WantID = %q, want normalized %q", c.Name, c.WantID, wantID)
		}
		if c.Input == "" {
			t.Errorf("valid case %q has empty input", c.Name)
		}
	}
	// Every invalid case must expect an error with a substring.
	for _, c := range invalid {
		if !c.WantErr {
			t.Errorf("invalid case %q not marked WantErr", c.Name)
		}
		if c.ErrContains == "" {
			t.Errorf("invalid case %q has empty ErrContains", c.Name)
		}
	}
}

func TestPullRefFixtures_Disjointness(t *testing.T) {
	f := loadPullRefs(t)
	valid, err := f.ValidCases()
	if err != nil {
		t.Fatalf("ValidCases: %v", err)
	}
	invalid, err := f.InvalidCases()
	if err != nil {
		t.Fatalf("InvalidCases: %v", err)
	}

	invalidInputs := make(map[string]string, len(invalid))
	for _, c := range invalid {
		invalidInputs[c.Input] = c.Name
	}
	for _, vc := range valid {
		if name, clash := invalidInputs[vc.Input]; clash {
			t.Errorf("generated valid input %q (case %q) also appears as reject case %q", vc.Input, vc.Name, name)
		}
	}
}

func TestPullStatusFixtures_Load(t *testing.T) {
	f, err := schema.LoadPullStatusFixtures()
	if err != nil {
		t.Fatalf("LoadPullStatusFixtures: %v", err)
	}
	if len(f.Statuses) != 6 {
		t.Errorf("statuses = %d, want 6", len(f.Statuses))
	}
	seenWire := map[string]bool{}
	for _, m := range f.Statuses {
		if m.ConstName == "" || m.Wire == "" {
			t.Errorf("status %q has empty const_name or wire", m.Name)
		}
		if seenWire[m.Wire] {
			t.Errorf("duplicate wire string %q", m.Wire)
		}
		seenWire[m.Wire] = true
	}
}

// TestPullManifestExample_ShapePin unmarshals the golden manifest into the real
// PullManifest, re-marshals it, normalizes formatting via json.Indent, and
// byte-compares against the committed file. This pins the documented IP7 field
// shape + the servedETag(quoted)/servedBlobHash(raw) split.
func TestPullManifestExample_ShapePin(t *testing.T) {
	raw := schema.PullManifestExampleJSON

	var m manifestShape
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal golden manifest: %v", err)
	}

	// Spot-check the load-bearing invariants the golden example demonstrates.
	if m.ManifestVersion != 1 {
		t.Errorf("manifestVersion = %d, want 1", m.ManifestVersion)
	}
	if !strings.HasPrefix(m.ServedETag, `"`) || !strings.HasSuffix(m.ServedETag, `"`) {
		t.Errorf("servedETag %q must be QUOTED (verbatim transport token)", m.ServedETag)
	}
	if strings.HasPrefix(m.ServedBlobHash, `"`) {
		t.Errorf("servedBlobHash %q must be RAW (unquoted content identity)", m.ServedBlobHash)
	}
	if got := strings.Trim(m.ServedETag, `"`); got != m.ServedBlobHash {
		t.Errorf("servedETag inner %q must equal raw servedBlobHash %q", got, m.ServedBlobHash)
	}
	if len(m.Annotations) == 0 {
		t.Error("golden manifest should demonstrate >=1 annotation provenance entry")
	}

	// Round-trip: re-marshal -> normalize -> byte-compare against committed bytes.
	remarshaled, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	var got, want bytes.Buffer
	if err := json.Indent(&got, remarshaled, "", "  "); err != nil {
		t.Fatalf("indent re-marshaled: %v", err)
	}
	if err := json.Indent(&want, bytes.TrimSpace(raw), "", "  "); err != nil {
		t.Fatalf("indent committed: %v", err)
	}
	if !bytes.Equal(got.Bytes(), want.Bytes()) {
		t.Errorf("golden manifest drift after round-trip:\n--- got ---\n%s\n--- want ---\n%s",
			got.String(), want.String())
	}
}

// manifestShape mirrors internal/pull.PullManifest's JSON tags. This schema
// module cannot import internal/pull (it is the lower layer), so the shape is
// pinned here against the same JSON tags. internal/pull/manifest_golden_test.go
// asserts the REAL PullManifest unmarshals the same bytes (cross-module pin).
type manifestShape struct {
	ManifestVersion     int                `json:"manifestVersion"`
	VillageURL          string             `json:"villageURL"`
	VillageHost         string             `json:"villageHost"`
	TranscriptID        string             `json:"transcriptId"`
	LocalSessionID      string             `json:"localSessionId,omitempty"`
	OwnerUserID         string             `json:"ownerUserId"`
	OwnerUsername       string             `json:"ownerUsername"`
	ServedETag          string             `json:"servedETag,omitempty"`
	ServedBlobHash      string             `json:"servedBlobHash,omitempty"`
	BlobContractVersion string             `json:"blobContractVersion,omitempty"`
	PullEnvelopeVersion string             `json:"pullEnvelopeVersion"`
	PulledAt            int64              `json:"pulledAt"`
	Annotations         []manifestAnnShape `json:"annotations"`
}

type manifestAnnShape struct {
	ContentHash    string `json:"contentHash"`
	AuthorUserID   string `json:"authorUserId"`
	AuthorUsername string `json:"authorUsername"`
}
