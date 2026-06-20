package schema

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// This file is the agentfilter-style loader for the pull transcript-reference
// parse fixtures (testdata/pull/transcript_refs.yaml). It mirrors the
// LoadAnnotationFixtures naming/conventions in testdata_annotations.go: a raw
// //go:embed YAML byte slice (fixtures.go) parsed into typed structs, with
// exported generators that materialise concrete test cases.
//
// The fixture declares NAMED AXES (id_casings, ref_forms) and attaches the
// expectation to the GENERATOR (valid_generation) and to the CATEGORY
// (invalid_refs / negative_lookalikes), rather than repeating a want per row.
// The loader expands those axes into flat RefCase values that
// internal/pull/types_test.go ranges over against the real ParseTranscriptRef.

// --- Raw YAML shapes ---

// PullRefFixtures is the top-level parsed transcript_refs.yaml structure.
type PullRefFixtures struct {
	Values            PullRefValues        `yaml:"values"`
	IDCasings         []IDCasing           `yaml:"id_casings"`
	RefForms          []RefForm            `yaml:"ref_forms"`
	ValidGeneration   ValidGeneration      `yaml:"valid_generation"`
	InvalidRefs       []InvalidRefCategory `yaml:"invalid_refs"`
	NegativeLookalike []NegativeLookalike  `yaml:"negative_lookalikes"`
}

// PullRefValues are the self-contained canonical strings the fixture owns. They
// are ADOPTED VERBATIM from internal/testutil (which re-exports them), so the
// fixture is the single source of truth with zero behaviour change.
type PullRefValues struct {
	UUIDLower   string `yaml:"uuid_lower"`
	VillageHost string `yaml:"village_host"`
}

// IDCasing is one entry of the id_casings axis: a named transform of
// values.uuid_lower (lower / upper / mixed).
type IDCasing struct {
	Name      string `yaml:"name"`
	Transform string `yaml:"transform"`
}

// RefForm is one entry of the ref_forms axis: a templated reference shape plus
// the FromURL the parser is expected to report for it.
type RefForm struct {
	Name     string `yaml:"name"`
	Template string `yaml:"template"`
	FromURL  bool   `yaml:"from_url"`
}

// ValidGeneration declares the valid cross-product and its attached expectation.
type ValidGeneration struct {
	Axes   []string            `yaml:"axes"`
	Expect ValidGenExpectation `yaml:"expect"`
}

// ValidGenExpectation is the expectation attached to the valid generator.
type ValidGenExpectation struct {
	Parses         bool   `yaml:"parses"`
	WantID         string `yaml:"want_id"`           // casing name whose value is the expected normalized ID
	FromURLPerForm bool   `yaml:"from_url_per_form"` // take FromURL from each ref_form
}

// InvalidRefCategory is a group of explicit reject cases sharing an expected
// error substring (attached to the category; a case may override it).
type InvalidRefCategory struct {
	Category    string              `yaml:"category"`
	ErrContains string              `yaml:"err_contains"`
	Cases       []InvalidRefSubcase `yaml:"cases"`
}

// InvalidRefSubcase is one reject case. Exactly one of Input/Template is set.
type InvalidRefSubcase struct {
	Name        string `yaml:"name"`
	Input       string `yaml:"input,omitempty"`
	Template    string `yaml:"template,omitempty"`
	ErrContains string `yaml:"err_contains,omitempty"` // overrides category default
}

// NegativeLookalike is a first-class reject case that LOOKS valid (e.g. a URL
// whose last path segment is a UUID but whose path is not the canonical
// /transcripts/<uuid> shape). ParseTranscriptRef rejects all of these, so each
// asserts LIVE in the InvalidCases loop.
type NegativeLookalike struct {
	Name        string `yaml:"name"`
	Template    string `yaml:"template"`
	ErrContains string `yaml:"err_contains"`
}

// --- Materialised case ---

// RefCase is a single concrete ParseTranscriptRef test case, materialised from
// the axes/categories. It is what internal/pull/types_test.go ranges over.
type RefCase struct {
	// Name is a stable, human-readable case identifier (used as the t.Run name).
	Name string
	// Input is the raw string fed to ParseTranscriptRef.
	Input string
	// WantID is the expected canonical (lowercase) TranscriptID string for a
	// PARSING case; empty for a reject case.
	WantID string
	// WantFromURL is the expected TranscriptRef.FromURL for a parsing case.
	WantFromURL bool
	// WantErr is true when ParseTranscriptRef is expected to return an error.
	WantErr bool
	// ErrContains is a substring the returned error must contain (reject cases).
	ErrContains string
}

// --- Loader ---

// LoadPullRefFixtures parses PullRefsYAML into structured fixtures. Mirrors
// LoadAnnotationFixtures.
func LoadPullRefFixtures() (*PullRefFixtures, error) {
	var f PullRefFixtures
	if err := yaml.Unmarshal(PullRefsYAML, &f); err != nil {
		return nil, fmt.Errorf("load pull ref fixtures: %w", err)
	}
	return &f, nil
}

// --- Exported canonical-value accessors ---

// UUIDLower returns the canonical lowercase UUID the fixture owns (== testutil
// TestTranscriptUUID).
func (f *PullRefFixtures) UUIDLower() string { return f.Values.UUIDLower }

// VillageHost returns the canonical village host the fixture owns (== testutil
// TestVillageHost).
func (f *PullRefFixtures) VillageHost() string { return f.Values.VillageHost }

// --- Template + transform helpers ---

// applyCasing applies a named id_casing transform to values.uuid_lower.
func (f *PullRefFixtures) applyCasing(c IDCasing) (string, error) {
	switch c.Transform {
	case "lower":
		return f.Values.UUIDLower, nil
	case "upper":
		return strings.ToUpper(f.Values.UUIDLower), nil
	case "mixed":
		return mixedCase(f.Values.UUIDLower), nil
	default:
		return "", fmt.Errorf("unknown id_casing transform %q (case %q)", c.Transform, c.Name)
	}
}

// mixedCase returns an alternating upper/lower transform of s. Non-letter runes
// (digits, hyphens) pass through unchanged, so for an all-numeric UUID this is a
// no-op — the case still exercises the parser's normalize-to-lower path.
func mixedCase(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			if upper {
				b.WriteRune(r - 32)
			} else {
				b.WriteRune(r)
			}
			upper = !upper
		case r >= 'A' && r <= 'Z':
			if upper {
				b.WriteRune(r)
			} else {
				b.WriteRune(r + 32)
			}
			upper = !upper
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// subst substitutes {id} and {host} in a template. Unknown placeholders are
// caught by the structure meta-test, not here.
func (f *PullRefFixtures) subst(template, id string) string {
	r := strings.NewReplacer("{id}", id, "{host}", f.Values.VillageHost)
	return r.Replace(template)
}

// crossProduct is the generic agentfilter cross-product helper: it walks the
// id_casings x ref_forms grid and invokes fn for each (casing, form) pair. Kept
// generic (over the two named axes) so the generator and the disjointness/
// coverage meta-tests share ONE expansion.
func (f *PullRefFixtures) crossProduct(fn func(c IDCasing, form RefForm)) {
	for _, c := range f.IDCasings {
		for _, form := range f.RefForms {
			fn(c, form)
		}
	}
}

// --- Generators ---

// ValidCases materialises the valid cross-product (id_casings x ref_forms) with
// the generator-attached expectation: every case parses, normalizes its ID to
// values.uuid_lower, and reports FromURL per its ref_form.
func (f *PullRefFixtures) ValidCases() ([]RefCase, error) {
	wantID := f.Values.UUIDLower
	if f.ValidGeneration.Expect.WantID != "lower" {
		// The only supported normalization target is the lowercase value.
		if v, ok := f.casingValueByName(f.ValidGeneration.Expect.WantID); ok {
			wantID = v
		} else {
			return nil, fmt.Errorf(
				"valid_generation.expect.want_id %q is not a known id_casing name",
				f.ValidGeneration.Expect.WantID)
		}
	}

	var cases []RefCase
	var perr error
	f.crossProduct(func(c IDCasing, form RefForm) {
		if perr != nil {
			return
		}
		id, err := f.applyCasing(c)
		if err != nil {
			perr = err
			return
		}
		cases = append(cases, RefCase{
			Name:        "valid/" + form.Name + "/" + c.Name,
			Input:       f.subst(form.Template, id),
			WantID:      wantID,
			WantFromURL: form.FromURL,
			WantErr:     false,
		})
	})
	if perr != nil {
		return nil, perr
	}
	return cases, nil
}

// casingValueByName resolves a casing name to its transformed value.
func (f *PullRefFixtures) casingValueByName(name string) (string, bool) {
	for _, c := range f.IDCasings {
		if c.Name == name {
			v, err := f.applyCasing(c)
			if err != nil {
				return "", false
			}
			return v, true
		}
	}
	return "", false
}

// InvalidCases materialises the explicit invalid_refs plus the negative_lookalikes.
// Every case expects an error with the category/case-attached substring.
func (f *PullRefFixtures) InvalidCases() ([]RefCase, error) {
	var cases []RefCase
	for _, cat := range f.InvalidRefs {
		for _, sub := range cat.Cases {
			if sub.Input != "" && sub.Template != "" {
				return nil, fmt.Errorf(
					"invalid_refs case %q/%q sets BOTH input and template (exactly one allowed)",
					cat.Category, sub.Name)
			}
			input := sub.Input
			if sub.Template != "" {
				input = f.subst(sub.Template, f.Values.UUIDLower)
			}
			errContains := cat.ErrContains
			if sub.ErrContains != "" {
				errContains = sub.ErrContains
			}
			cases = append(cases, RefCase{
				Name:        "invalid/" + cat.Category + "/" + sub.Name,
				Input:       input,
				WantErr:     true,
				ErrContains: errContains,
			})
		}
	}
	for _, look := range f.NegativeLookalike {
		cases = append(cases, RefCase{
			Name:        "lookalike/" + look.Name,
			Input:       f.subst(look.Template, f.Values.UUIDLower),
			WantErr:     true,
			ErrContains: look.ErrContains,
		})
	}
	return cases, nil
}

// AllCases is ValidCases followed by InvalidCases — the full materialised suite
// internal/pull/types_test.go ranges over.
func (f *PullRefFixtures) AllCases() ([]RefCase, error) {
	valid, err := f.ValidCases()
	if err != nil {
		return nil, err
	}
	invalid, err := f.InvalidCases()
	if err != nil {
		return nil, err
	}
	return append(valid, invalid...), nil
}

// --- PullStatus fixtures ---

// PullStatusFixtures is the parsed pull_statuses.yaml structure.
type PullStatusFixtures struct {
	Statuses []PullStatusMapping `yaml:"statuses"`
}

// PullStatusMapping is one PullStatus<->wire row.
type PullStatusMapping struct {
	Name      string `yaml:"name"`
	ConstName string `yaml:"const_name"`
	Wire      string `yaml:"wire"`
}

// LoadPullStatusFixtures parses PullStatusesYAML into structured fixtures.
func LoadPullStatusFixtures() (*PullStatusFixtures, error) {
	var f PullStatusFixtures
	if err := yaml.Unmarshal(PullStatusesYAML, &f); err != nil {
		return nil, fmt.Errorf("load pull status fixtures: %w", err)
	}
	return &f, nil
}

// WireFor returns the expected wire string for a given const_name, or "".
func (f *PullStatusFixtures) WireFor(constName string) (string, bool) {
	for _, m := range f.Statuses {
		if m.ConstName == constName {
			return m.Wire, true
		}
	}
	return "", false
}
