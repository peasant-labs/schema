package schema

import "context"

// AnnotationTypeReader provides read-only access to the annotation type registry.
// Consumed by classifiers, the ingest pipeline, and the REST API handler.
// Returns AnnotationTypeSummary (wire type) — the internal domain object stays in internal/annotations.
type AnnotationTypeReader interface {
	// GetType returns the annotation type summary for the given type_id string.
	// Returns an error wrapping ErrTypeNotFound if no matching type exists.
	GetType(ctx context.Context, typeID string) (*AnnotationTypeSummary, error)

	// ListTypes returns all annotation types matching the given filter.
	// An empty TypeFilter returns all non-deprecated/retired types.
	ListTypes(ctx context.Context, f TypeFilter) ([]AnnotationTypeSummary, error)

	// ValidateValue validates that value is permissible for the annotation type identified by typeID.
	// Returns nil if valid. Returns an error wrapping ErrTypeNotFound if the type does not exist,
	// or an error wrapping ErrInvalidValue if the value is not permissible.
	ValidateValue(ctx context.Context, typeID string, value string) error
}

// AnnotationRegistry extends AnnotationTypeReader with mutation operations.
// Used by admin commands and the registry management layer.
type AnnotationRegistry interface {
	AnnotationTypeReader

	// Register inserts a new annotation type from the given definition.
	// Returns the newly created AnnotationTypeSummary (status=proposed by default).
	Register(ctx context.Context, def TypeDefinition) (*AnnotationTypeSummary, error)

	// Activate transitions a type from proposed -> active.
	Activate(ctx context.Context, typeID string) error

	// Deprecate transitions a type from active -> deprecated, recording the superseding typeID.
	Deprecate(ctx context.Context, typeID string, supersededBy string) error

	// AddDependency records that typeID depends on dependsOn (V14: cycle detection enforced).
	// required=true means the dependency must be satisfied before typeID can produce a value.
	// rationale documents the reason for the dependency.
	AddDependency(ctx context.Context, typeID, dependsOn string, required bool, rationale string) error

	// GetDependencies returns the dependency entries for typeID.
	GetDependencies(ctx context.Context, typeID string) ([]TypeDependency, error)
}

// TypeFilter constrains which annotation types ListTypes returns.
// Zero values mean "no filter" (include all non-deprecated types).
type TypeFilter struct {
	// Status filters to types with this status (e.g., StatusActive).
	// Zero value includes all statuses except deprecated/retired (unless IncludeDeprecated is set).
	Status AnnotationStatus

	// FamilyID filters to types in this family by UUID FK. "" means all families.
	FamilyID string

	// Origin filters to types from this origin. Zero value means all origins.
	Origin TypeOrigin

	// IncludeDeprecated includes deprecated types when true.
	// Has no effect when Status is set (Status takes precedence).
	IncludeDeprecated bool
}

// TypeDefinition is the input for registering a new annotation type.
type TypeDefinition struct {
	// TypeID is the dot-notation identifier (e.g., "quality.my_signal").
	// Must match the CHECK (type_id LIKE '%.%') constraint.
	TypeID string

	// DisplayName is the human-readable name.
	DisplayName string

	// Description is optional free-text documentation.
	Description string

	// FamilyID is the UUID FK to annotation_families.id.
	FamilyID string

	// ValueDomain defines permissible values for this type.
	ValueDomain ValueDomain

	// LowerIsBetter is nil for non-ordinal types, false for higher-is-better, true for lower-is-better.
	LowerIsBetter *bool

	// Origin is who created this type (system, user, group).
	Origin TypeOrigin

	// AllowedTargetKinds specifies which target kinds this type allows (V16).
	// Nil or empty means all target kinds are allowed.
	AllowedTargetKinds []TargetKind
}

// TypeDependency represents a dependency edge in the annotation_type_deps table.
type TypeDependency struct {
	// TypeID is the dependent annotation type.
	TypeID string

	// DependsOn is the type that must be computed first.
	DependsOn string

	// Required: true means this dependency must be satisfied before TypeID can produce a value.
	Required bool

	// Rationale documents why this dependency exists.
	Rationale string
}
