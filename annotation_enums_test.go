package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/peasant-labs/schema"
)

// --- ScaleKind ---

func TestScaleKind_IsValid_AllKnown(t *testing.T) {
	t.Parallel()
	for _, k := range schema.AllScaleKinds {
		if !k.IsValid() {
			t.Errorf("AllScaleKinds[%s] should be valid", k)
		}
	}
}

func TestScaleKind_IsValid_Unknown(t *testing.T) {
	t.Parallel()
	if schema.ScaleKind("unknown").IsValid() {
		t.Error("unknown scale kind should be invalid")
	}
	if schema.ScaleKind("").IsValid() {
		t.Error("empty scale kind should be invalid")
	}
}

func TestScaleKind_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		kind schema.ScaleKind
		want string
	}{
		{schema.ScaleNominal, "nominal"},
		{schema.ScaleOrdinal, "ordinal"},
		{schema.ScaleContinuous, "continuous"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("ScaleKind(%q).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

func TestScaleKind_AllThreeValues(t *testing.T) {
	t.Parallel()
	if len(schema.AllScaleKinds) != 3 {
		t.Errorf("AllScaleKinds: expected 3 values, got %d", len(schema.AllScaleKinds))
	}

	// Verify all three specific values exist.
	found := map[schema.ScaleKind]bool{}
	for _, k := range schema.AllScaleKinds {
		found[k] = true
	}
	for _, expected := range []schema.ScaleKind{schema.ScaleNominal, schema.ScaleOrdinal, schema.ScaleContinuous} {
		if !found[expected] {
			t.Errorf("AllScaleKinds: missing %q", expected)
		}
	}
}

func TestScaleKind_JSONSchema_HasEnum(t *testing.T) {
	t.Parallel()
	s, err := schema.ScaleKind("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("ScaleKind JSONSchema: expected enum values")
	}
	// Verify all three enum values are present.
	enumMap := make(map[string]bool)
	for _, e := range s.Enum {
		if str, ok := e.(string); ok {
			enumMap[str] = true
		}
	}
	for _, expected := range []string{"nominal", "ordinal", "continuous"} {
		if !enumMap[expected] {
			t.Errorf("ScaleKind JSONSchema: missing enum value %q", expected)
		}
	}
}

// --- ValidateScaleDomainCombo ---

func TestValidateScaleDomainCombo_ValidCombinations(t *testing.T) {
	t.Parallel()
	validCases := []struct {
		scale  schema.ScaleKind
		domain schema.ValueDomainKind
		name   string
	}{
		{schema.ScaleNominal, schema.DomainEnumerated, "enumerated+nominal"},
		{schema.ScaleOrdinal, schema.DomainEnumerated, "enumerated+ordinal"},
		{schema.ScaleContinuous, schema.DomainDescribed, "described+continuous"},
		{schema.ScaleNominal, schema.DomainDescribed, "described+nominal"},
	}
	for _, tc := range validCases {
		if err := schema.ValidateScaleDomainCombo(tc.scale, tc.domain); err != nil {
			t.Errorf("%s: expected valid combination, got error: %v", tc.name, err)
		}
	}
}

func TestValidateScaleDomainCombo_DescribedOrdinalRejected(t *testing.T) {
	t.Parallel()
	// described+ordinal is invalid: ordinal requires explicit ordering via permissible values.
	err := schema.ValidateScaleDomainCombo(schema.ScaleOrdinal, schema.DomainDescribed)
	if err == nil {
		t.Error("described+ordinal should be rejected by ValidateScaleDomainCombo")
	}
}

func TestValidateScaleDomainCombo_EnumeratedContinuousRejected(t *testing.T) {
	t.Parallel()
	// enumerated+continuous is invalid: continuous ranges require a described constraint spec.
	err := schema.ValidateScaleDomainCombo(schema.ScaleContinuous, schema.DomainEnumerated)
	if err == nil {
		t.Error("enumerated+continuous should be rejected by ValidateScaleDomainCombo")
	}
}

// --- ScaleKind on AnnotationTypeSummary JSON serialization ---

func TestAnnotationTypeSummary_ScaleKind_OmitEmptyWhenZero(t *testing.T) {
	t.Parallel()
	ats := schema.AnnotationTypeSummary{
		TypeID:      "test.type",
		Version:     1,
		DisplayName: "Test",
		Family:      "test",
		Class:       "test",
		// ScaleKind not set — should be omitted from JSON (omitempty)
		ValueDomain: schema.ValueDomain{
			Kind:              schema.DomainEnumerated,
			Datatype:          schema.DatatypeText,
			PermissibleValues: []string{"a", "b"},
		},
	}
	data, err := json.Marshal(ats)
	if err != nil {
		t.Fatalf("marshal AnnotationTypeSummary: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if _, ok := m["scaleKind"]; ok {
		t.Error("scaleKind should be omitted from JSON when zero-value (omitempty)")
	}
}

func TestAnnotationTypeSummary_ScaleKind_PresentWhenSet(t *testing.T) {
	t.Parallel()
	ats := schema.AnnotationTypeSummary{
		TypeID:      "test.type",
		Version:     1,
		DisplayName: "Test",
		Family:      "test",
		Class:       "test",
		ScaleKind:   schema.ScaleOrdinal,
		ValueDomain: schema.ValueDomain{
			Kind:              schema.DomainEnumerated,
			Datatype:          schema.DatatypeText,
			PermissibleValues: []string{"poor", "fair", "good", "excellent"},
		},
	}
	data, err := json.Marshal(ats)
	if err != nil {
		t.Fatalf("marshal AnnotationTypeSummary: %v", err)
	}
	var got schema.AnnotationTypeSummary
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal AnnotationTypeSummary: %v", err)
	}
	if got.ScaleKind != schema.ScaleOrdinal {
		t.Errorf("ScaleKind: got %q, want %q", got.ScaleKind, schema.ScaleOrdinal)
	}
}

// TestValueDomain_OrdinalEnumerated_RoundTrip tests the BDD criterion:
// Given PermissibleValues=["poor","fair","good","excellent"] on an enumerated ValueDomain,
// when marshalled and unmarshalled, then PermissibleValues round-trips correctly.
//
// Note: ScaleKind is now a top-level field on AnnotationTypeSummary, not on ValueDomain.
// Rank validation is in the annotations.validateValue function, not in ValueDomain itself.
func TestValueDomain_OrdinalEnumerated_RoundTrip(t *testing.T) {
	t.Parallel()
	vd := schema.ValueDomain{
		Kind:              schema.DomainEnumerated,
		Datatype:          schema.DatatypeText,
		PermissibleValues: []string{"poor", "fair", "good", "excellent"},
	}
	data, err := json.Marshal(vd)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got schema.ValueDomain
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.PermissibleValues) != 4 {
		t.Errorf("PermissibleValues: want 4, got %d", len(got.PermissibleValues))
	}
}
