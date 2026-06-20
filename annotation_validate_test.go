package schema_test

import (
	"errors"
	"testing"

	"github.com/peasant-labs/schema"
)

// TestValidateAnnotationValue_Enumerated_Valid verifies that a value present in
// PermissibleValues passes validation.
func TestValidateAnnotationValue_Enumerated_Valid(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:              schema.DomainEnumerated,
		Datatype:          schema.DatatypeText,
		PermissibleValues: []string{"approve", "deny"},
	}

	for _, v := range []string{"approve", "deny"} {
		if err := schema.ValidateAnnotationValue(domain, v); err != nil {
			t.Errorf("ValidateAnnotationValue(%q): unexpected error: %v", v, err)
		}
	}
}

// TestValidateAnnotationValue_Enumerated_Invalid verifies that a value not in
// PermissibleValues returns an error wrapping ErrInvalidValue.
func TestValidateAnnotationValue_Enumerated_Invalid(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:              schema.DomainEnumerated,
		Datatype:          schema.DatatypeText,
		PermissibleValues: []string{"approve", "deny"},
	}

	err := schema.ValidateAnnotationValue(domain, "maybe")
	if err == nil {
		t.Fatal("ValidateAnnotationValue(maybe): expected error, got nil")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("ValidateAnnotationValue(maybe): error = %v, want wrapping ErrInvalidValue", err)
	}
}

// TestValidateAnnotationValue_Enumerated_Empty verifies that an empty value returns
// an error wrapping ErrInvalidValue for enumerated domains.
func TestValidateAnnotationValue_Enumerated_Empty(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:              schema.DomainEnumerated,
		Datatype:          schema.DatatypeText,
		PermissibleValues: []string{"approve", "deny"},
	}

	err := schema.ValidateAnnotationValue(domain, "")
	if err == nil {
		t.Fatal("ValidateAnnotationValue(''): expected error, got nil")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("ValidateAnnotationValue(''): error = %v, want wrapping ErrInvalidValue", err)
	}
}

// TestValidateAnnotationValue_Described_Valid verifies that a non-empty value passes
// validation for described domains (MVP: any non-empty string is accepted).
func TestValidateAnnotationValue_Described_Valid(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeText,
		ConstraintSpec: `{"min":0,"max":100}`,
	}

	for _, v := range []string{"42", "hello world", "some-free-text"} {
		if err := schema.ValidateAnnotationValue(domain, v); err != nil {
			t.Errorf("ValidateAnnotationValue(%q): unexpected error for described domain: %v", v, err)
		}
	}
}

// TestValidateAnnotationValue_Described_EmptyRejected verifies that an empty value is
// rejected even for described domains.
func TestValidateAnnotationValue_Described_EmptyRejected(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:     schema.DomainDescribed,
		Datatype: schema.DatatypeText,
	}

	err := schema.ValidateAnnotationValue(domain, "")
	if err == nil {
		t.Fatal("ValidateAnnotationValue(''): expected error for described domain, got nil")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("ValidateAnnotationValue(''): error = %v, want wrapping ErrInvalidValue", err)
	}
}

// TestValidateAnnotationValue_UnknownDomainKind verifies that an unknown domain kind
// returns an error wrapping ErrInvalidValue.
func TestValidateAnnotationValue_UnknownDomainKind(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:     schema.ValueDomainKind("unknown_kind"),
		Datatype: schema.DatatypeText,
	}

	err := schema.ValidateAnnotationValue(domain, "some-value")
	if err == nil {
		t.Fatal("ValidateAnnotationValue with unknown domain kind: expected error, got nil")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("ValidateAnnotationValue with unknown domain kind: error = %v, want wrapping ErrInvalidValue", err)
	}
}

// ---------------------------------------------------------------------------
// ValidateDescribedValue tests (V17: jsonschema + datatype coercion)
// ---------------------------------------------------------------------------

// TestValidateDescribedValue_ContinuousRange tests the BDD criterion:
// Given ScaleKind=continuous + ConstraintSpec={"minimum":0,"maximum":1},
// when value="0.75", then validates via jsonschema.
func TestValidateDescribedValue_ContinuousRange_Valid(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":1}`,
	}
	if err := schema.ValidateDescribedValue(domain, "0.75"); err != nil {
		t.Errorf("value=0.75 should be valid for [0,1] range: %v", err)
	}
}

// TestValidateDescribedValue_ContinuousRange_BoundaryMin tests that the minimum boundary is valid.
func TestValidateDescribedValue_ContinuousRange_BoundaryMin(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":1}`,
	}
	if err := schema.ValidateDescribedValue(domain, "0"); err != nil {
		t.Errorf("value=0 (minimum boundary) should be valid: %v", err)
	}
}

// TestValidateDescribedValue_ContinuousRange_BoundaryMax tests that the maximum boundary is valid.
func TestValidateDescribedValue_ContinuousRange_BoundaryMax(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":1}`,
	}
	if err := schema.ValidateDescribedValue(domain, "1"); err != nil {
		t.Errorf("value=1 (maximum boundary) should be valid: %v", err)
	}
}

// TestValidateDescribedValue_ContinuousRange_OutOfRange tests the BDD criterion:
// Given ScaleKind=continuous + ConstraintSpec={"minimum":0,"maximum":1},
// when value="1.5", then returns ErrInvalidValue.
func TestValidateDescribedValue_ContinuousRange_OutOfRange(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":1}`,
	}
	err := schema.ValidateDescribedValue(domain, "1.5")
	if err == nil {
		t.Error("value=1.5 should be rejected for [0,1] range")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("expected error to wrap ErrInvalidValue, got: %v", err)
	}
}

// TestValidateDescribedValue_DatatypeCoercion_IntegerRejectsFloat tests:
// BDD: Given Datatype=integer, when value="3.7", then returns ErrInvalidValue.
func TestValidateDescribedValue_DatatypeCoercion_IntegerRejectsFloat(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeInteger,
		ConstraintSpec: `{"minimum":0,"maximum":10}`,
	}
	err := schema.ValidateDescribedValue(domain, "3.7")
	if err == nil {
		t.Error("value=3.7 should be rejected for integer datatype")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("expected error to wrap ErrInvalidValue, got: %v", err)
	}
}

// TestValidateDescribedValue_DatatypeCoercion_IntegerAcceptsInt tests that valid integer values pass.
func TestValidateDescribedValue_DatatypeCoercion_IntegerAcceptsInt(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeInteger,
		ConstraintSpec: `{"minimum":0,"maximum":10}`,
	}
	if err := schema.ValidateDescribedValue(domain, "5"); err != nil {
		t.Errorf("value=5 should be valid for integer [0,10]: %v", err)
	}
}

// TestValidateDescribedValue_DatatypeCoercion_RealRejectsNonNumber tests:
// BDD: Given Datatype=real, when value="not_a_number", then returns ErrInvalidValue.
func TestValidateDescribedValue_DatatypeCoercion_RealRejectsNonNumber(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":1}`,
	}
	err := schema.ValidateDescribedValue(domain, "not_a_number")
	if err == nil {
		t.Error("value=not_a_number should be rejected for real datatype")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("expected error to wrap ErrInvalidValue, got: %v", err)
	}
}

// TestValidateDescribedValue_DatatypeCoercion_BooleanRejectsMaybe tests:
// BDD: Given Datatype=boolean, when value="maybe", then returns ErrInvalidValue.
func TestValidateDescribedValue_DatatypeCoercion_BooleanRejectsMaybe(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeBoolean,
		ConstraintSpec: `{}`,
	}
	err := schema.ValidateDescribedValue(domain, "maybe")
	if err == nil {
		t.Error("value=maybe should be rejected for boolean datatype")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("expected error to wrap ErrInvalidValue, got: %v", err)
	}
}

// TestValidateDescribedValue_DatatypeCoercion_BooleanAcceptsTrue tests valid boolean values.
func TestValidateDescribedValue_DatatypeCoercion_BooleanAcceptsTrue(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeBoolean,
		ConstraintSpec: `{}`,
	}
	for _, valid := range []string{"true", "false", "1", "0"} {
		if err := schema.ValidateDescribedValue(domain, valid); err != nil {
			t.Errorf("value=%q should be valid for boolean datatype: %v", valid, err)
		}
	}
}

// TestValidateDescribedValue_EmptyValue_Rejected tests that empty values are rejected.
func TestValidateDescribedValue_EmptyValue_Rejected(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeText,
		ConstraintSpec: `{}`,
	}
	err := schema.ValidateDescribedValue(domain, "")
	if err == nil {
		t.Error("empty value should be rejected")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("expected error to wrap ErrInvalidValue, got: %v", err)
	}
}

// TestValidateDescribedValue_SchemaCache_MultipleCalls tests that calling ValidateDescribedValue
// multiple times with the same ConstraintSpec does not error (exercises the cache).
func TestValidateDescribedValue_SchemaCache_MultipleCalls(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":100}`,
	}
	// Call multiple times with different valid values — all should succeed.
	for _, v := range []string{"0", "50", "99.9", "100"} {
		if err := schema.ValidateDescribedValue(domain, v); err != nil {
			t.Errorf("value=%q should be valid: %v", v, err)
		}
	}
}

// TestValidateDescribedValue_InvalidConstraintSpec_ReturnsError tests that a malformed JSON
// constraint spec returns an error (not a panic).
func TestValidateDescribedValue_InvalidConstraintSpec_ReturnsError(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `not valid json`,
	}
	err := schema.ValidateDescribedValue(domain, "0.5")
	if err == nil {
		t.Error("invalid JSON constraint spec should return an error")
	}
}

// TestValidateAnnotationValue_Described_DelegatesJSONSchema verifies that
// ValidateAnnotationValue with a described domain delegates to ValidateDescribedValue
// (which enforces datatype coercion + JSON schema) rather than accepting any non-empty string.
func TestValidateAnnotationValue_Described_DelegatesJSONSchema(t *testing.T) {
	t.Parallel()
	domain := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"minimum":0,"maximum":1}`,
	}

	// Value "1.5" is out of range — should fail (not be accepted as non-empty string).
	err := schema.ValidateAnnotationValue(domain, "1.5")
	if err == nil {
		t.Error("described domain: value=1.5 out of [0,1] range should be rejected by ValidateAnnotationValue")
	}
	if !errors.Is(err, schema.ErrInvalidValue) {
		t.Errorf("expected ErrInvalidValue, got: %v", err)
	}

	// Valid value should pass.
	if err := schema.ValidateAnnotationValue(domain, "0.5"); err != nil {
		t.Errorf("described domain: value=0.5 should be valid: %v", err)
	}
}
