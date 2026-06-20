package schema

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// ErrInvalidValue is returned when a value is not permissible for an annotation domain.
// Callers should wrap this error with fmt.Errorf("%w: ...", ErrInvalidValue) for context.
var ErrInvalidValue = errors.New("invalid annotation value")

// schemaCache caches compiled jsonschema.Schema objects keyed by constraint spec string.
// Compiling a JSON schema is expensive; caching amortises the cost across repeated
// annotation validations for the same annotation type.
var schemaCache sync.Map // map[string]*jsonschema.Schema

// ValidateAnnotationValue checks that value is permissible for the given ValueDomain.
//
// Rules:
//   - value must not be empty (regardless of domain kind).
//   - For enumerated domains: value must be in PermissibleValues.
//   - For described domains: delegates to ValidateDescribedValue (datatype coercion + JSON schema).
//
// Returns nil if valid. Returns an error wrapping ErrInvalidValue otherwise.
func ValidateAnnotationValue(domain ValueDomain, value string) error {
	if value == "" {
		return fmt.Errorf("%w: value must not be empty", ErrInvalidValue)
	}

	switch domain.Kind {
	case DomainEnumerated:
		return validateEnumeratedAnnotationValue(domain.PermissibleValues, value)
	case DomainDescribed:
		return ValidateDescribedValue(domain, value)
	default:
		return fmt.Errorf("%w: unknown value domain kind %q", ErrInvalidValue, domain.Kind)
	}
}

// validateEnumeratedAnnotationValue checks that value is one of the allowed strings.
func validateEnumeratedAnnotationValue(permissibleValues []string, value string) error {
	for _, allowed := range permissibleValues {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("%w: %q not in permissible values %v", ErrInvalidValue, value, permissibleValues)
}

// ValidateDescribedValue validates that value is permissible for the given described ValueDomain.
// It performs two checks in order:
//
//  1. Datatype coercion: the string value must parse as the declared Datatype before schema
//     validation. This catches "3.7" for an integer field before jsonschema sees it.
//
//  2. JSON-schema validation: the ConstraintSpec is compiled (with caching) and the
//     coerced Go value is validated against it.
//
// Returns nil if valid, or an error wrapping ErrInvalidValue when rejected.
// Returns a non-ErrInvalidValue error when the ConstraintSpec is malformed JSON.
//
// Callers:
//   - ValidateAnnotationValue (delegated for described domains)
//
// Where: annotation_validate.go ValidateDescribedValue.
func ValidateDescribedValue(domain ValueDomain, value string) error {
	if value == "" {
		return fmt.Errorf(
			"%w: value must not be empty (type: described/%s, location: annotation_validate.go ValidateDescribedValue)",
			ErrInvalidValue, domain.Datatype,
		)
	}

	// Step 1: Datatype coercion — parse value as the declared datatype.
	coerced, err := coerceDatatype(domain.Datatype, value)
	if err != nil {
		return fmt.Errorf(
			"%w: %v (datatype=%s, value=%q, location: annotation_validate.go coerceDatatype)",
			ErrInvalidValue, err, domain.Datatype, value,
		)
	}

	// Step 2: JSON-schema validation — compile ConstraintSpec and validate coerced value.
	// Empty constraint spec: skip JSON schema validation (no constraints to enforce).
	if domain.ConstraintSpec == "" || domain.ConstraintSpec == "{}" {
		return nil
	}

	compiled, err := compiledSchema(domain.ConstraintSpec)
	if err != nil {
		// Malformed ConstraintSpec is a configuration error, not a value error.
		// Return as-is (does NOT wrap ErrInvalidValue).
		return fmt.Errorf(
			"annotation_validate: malformed ConstraintSpec %q: %w (location: annotation_validate.go ValidateDescribedValue)",
			domain.ConstraintSpec, err,
		)
	}

	if err := compiled.Validate(coerced); err != nil {
		return fmt.Errorf(
			"%w: %v (constraintSpec=%q, value=%q, location: annotation_validate.go ValidateDescribedValue)",
			ErrInvalidValue, err, domain.ConstraintSpec, value,
		)
	}
	return nil
}

// coerceDatatype parses value as the declared datatype and returns the Go value
// suitable for JSON schema validation:
//
//   - integer → int64 (via strconv.ParseInt base 10)
//   - real    → float64 (via strconv.ParseFloat)
//   - boolean → bool (via strconv.ParseBool: "true"/"false"/"1"/"0" etc.)
//   - text    → string (no conversion; JSON schema validates the raw string)
//
// Where: annotation_validate.go coerceDatatype.
func coerceDatatype(dt AnnotationDatatype, value string) (any, error) {
	switch dt {
	case DatatypeInteger:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q as integer: %w", value, err)
		}
		return v, nil

	case DatatypeReal:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q as real (float64): %w", value, err)
		}
		return v, nil

	case DatatypeBoolean:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q as boolean (accept: true/false/1/0): %w", value, err)
		}
		return v, nil

	case DatatypeText:
		return value, nil

	default:
		return nil, fmt.Errorf("unknown datatype %q", dt)
	}
}

// compiledSchema returns a cached compiled *jsonschema.Schema for the given constraint spec.
// The spec is expected to be a JSON object (e.g. {"minimum":0,"maximum":1}).
//
// Cache is keyed by the exact constraint spec string. Concurrent access is safe
// via sync.Map's atomic LoadOrStore semantics.
//
// Where: annotation_validate.go compiledSchema.
func compiledSchema(constraintSpec string) (*jsonschema.Schema, error) {
	// Fast path: cache hit.
	if v, ok := schemaCache.Load(constraintSpec); ok {
		return v.(*jsonschema.Schema), nil
	}

	// Compile the schema. Use a unique URI to avoid compiler state conflicts.
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7

	const schemaURI = "peasant:///constraint"
	if err := compiler.AddResource(schemaURI, bytes.NewReader([]byte(constraintSpec))); err != nil {
		return nil, fmt.Errorf("add resource: %w", err)
	}

	compiled, err := compiler.Compile(schemaURI)
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	// Store in cache (another goroutine may have stored first; use the winner).
	actual, _ := schemaCache.LoadOrStore(constraintSpec, compiled)
	return actual.(*jsonschema.Schema), nil
}
