package schema

import (
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- AnnotatorKind ---

// AnnotatorKind identifies the type of entity that produced an annotation.
// Priority for effective annotation resolution: human(3) > agent(2) > rule(1).
type AnnotatorKind string

const (
	AnnotatorHuman AnnotatorKind = "human"
	AnnotatorAgent AnnotatorKind = "agent"
	AnnotatorRule  AnnotatorKind = "rule"
)

// IsValid returns true if the annotator kind is one of the known variants.
func (k AnnotatorKind) IsValid() bool {
	switch k {
	case AnnotatorHuman, AnnotatorAgent, AnnotatorRule:
		return true
	}
	return false
}

func (k AnnotatorKind) String() string { return string(k) }

// Priority returns the resolution priority for effective annotation selection.
// Higher value wins: human(3) > agent(2) > rule(1). Returns 0 for unknown kinds.
func (k AnnotatorKind) Priority() int {
	switch k {
	case AnnotatorHuman:
		return 3
	case AnnotatorAgent:
		return 2
	case AnnotatorRule:
		return 1
	default:
		return 0
	}
}

// AllAnnotatorKinds is the canonical list of all known annotator kinds.
var AllAnnotatorKinds = []AnnotatorKind{
	AnnotatorHuman, AnnotatorAgent, AnnotatorRule,
}

// JSONSchema implements jsonschema.Exposer.
func (AnnotatorKind) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Annotator Kind")
	s.WithDescription("Type of entity that produced an annotation: human, agent (AI model), or rule (automated classifier)")
	s.WithEnum("human", "agent", "rule")
	s.WithExamples("human", "agent", "rule")
	return s, nil
}

// --- AnnotationStatus ---

// AnnotationStatus is the lifecycle state of an annotation type (ISO 11179 Part 6).
type AnnotationStatus string

const (
	StatusProposed   AnnotationStatus = "proposed"
	StatusActive     AnnotationStatus = "active"
	StatusDeprecated AnnotationStatus = "deprecated"
	StatusRetired    AnnotationStatus = "retired"
)

// IsValid returns true if the annotation status is one of the known variants.
func (s AnnotationStatus) IsValid() bool {
	switch s {
	case StatusProposed, StatusActive, StatusDeprecated, StatusRetired:
		return true
	}
	return false
}

func (s AnnotationStatus) String() string { return string(s) }

// AllAnnotationStatuses is the canonical list of all known annotation statuses.
var AllAnnotationStatuses = []AnnotationStatus{
	StatusProposed, StatusActive, StatusDeprecated, StatusRetired,
}

// JSONSchema implements jsonschema.Exposer.
func (AnnotationStatus) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Annotation Status")
	s.WithDescription("ISO 11179 lifecycle state of an annotation type")
	s.WithEnum("proposed", "active", "deprecated", "retired")
	s.WithExamples("active", "deprecated")
	return s, nil
}

// --- ValueDomainKind ---

// ValueDomainKind distinguishes enumerated (finite set of allowed values) from
// described (range or pattern constraint) annotation value domains (ISO 11179).
type ValueDomainKind string

const (
	DomainEnumerated ValueDomainKind = "enumerated"
	DomainDescribed  ValueDomainKind = "described"
)

// IsValid returns true if the value domain kind is one of the known variants.
func (k ValueDomainKind) IsValid() bool {
	switch k {
	case DomainEnumerated, DomainDescribed:
		return true
	}
	return false
}

func (k ValueDomainKind) String() string { return string(k) }

// AllValueDomainKinds is the canonical list of all known value domain kinds.
var AllValueDomainKinds = []ValueDomainKind{
	DomainEnumerated, DomainDescribed,
}

// JSONSchema implements jsonschema.Exposer.
func (ValueDomainKind) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Value Domain Kind")
	s.WithDescription("ISO 11179 value domain: enumerated (finite allowed set) or described (range/pattern constraint)")
	s.WithEnum("enumerated", "described")
	s.WithExamples("enumerated", "described")
	return s, nil
}

// --- AnnotationDatatype ---

// AnnotationDatatype is the storage type for annotation values.
type AnnotationDatatype string

const (
	DatatypeText    AnnotationDatatype = "text"
	DatatypeInteger AnnotationDatatype = "integer"
	DatatypeReal    AnnotationDatatype = "real"
	DatatypeBoolean AnnotationDatatype = "boolean"
)

// IsValid returns true if the annotation datatype is one of the known variants.
func (d AnnotationDatatype) IsValid() bool {
	switch d {
	case DatatypeText, DatatypeInteger, DatatypeReal, DatatypeBoolean:
		return true
	}
	return false
}

func (d AnnotationDatatype) String() string { return string(d) }

// AllAnnotationDatatypes is the canonical list of all known annotation datatypes.
var AllAnnotationDatatypes = []AnnotationDatatype{
	DatatypeText, DatatypeInteger, DatatypeReal, DatatypeBoolean,
}

// JSONSchema implements jsonschema.Exposer.
func (AnnotationDatatype) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Annotation Datatype")
	s.WithDescription("Storage type for annotation values (maps to SQLite STRICT column type)")
	s.WithEnum("text", "integer", "real", "boolean")
	s.WithExamples("text", "integer")
	return s, nil
}

// --- TypeOrigin ---

// TypeOrigin identifies who created an annotation type.
type TypeOrigin string

const (
	OriginSystem TypeOrigin = "system"
	OriginUser   TypeOrigin = "user"
	OriginGroup  TypeOrigin = "group"
)

// IsValid returns true if the type origin is one of the known variants.
func (o TypeOrigin) IsValid() bool {
	switch o {
	case OriginSystem, OriginUser, OriginGroup:
		return true
	}
	return false
}

func (o TypeOrigin) String() string { return string(o) }

// AllTypeOrigins is the canonical list of all known type origins.
var AllTypeOrigins = []TypeOrigin{
	OriginSystem, OriginUser, OriginGroup,
}

// JSONSchema implements jsonschema.Exposer.
func (TypeOrigin) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Type Origin")
	s.WithDescription("Who created an annotation type: system (built-in), user (individual), or group (shared)")
	s.WithEnum("system", "user", "group")
	s.WithExamples("system", "user")
	return s, nil
}

// --- ScaleKind ---

// ScaleKind classifies the measurement level of an annotation value domain
// (Stevens 1946 levels of measurement, ISO 11179 Part 5).
//
//   - nominal: categories without order (e.g. session scope: feature/bug/docs)
//   - ordinal: ordered categories with no meaningful interval (e.g. approval: deny<approve)
//   - continuous: numeric range with meaningful intervals (e.g. confidence 0.0–1.0)
//
// Valid combinations with ValueDomainKind:
//   - enumerated + nominal: OK (categories without order)
//   - enumerated + ordinal: OK (ordered categories with permissible values list)
//   - described + continuous: OK (range with JSON schema constraint spec)
//   - described + nominal: OK (pattern-constrained categories)
//   - described + ordinal: REJECTED (ordinal requires explicit ordering via permissible values)
//   - enumerated + continuous: REJECTED (continuous ranges must be described, not enumerated)
type ScaleKind string

const (
	ScaleNominal    ScaleKind = "nominal"
	ScaleOrdinal    ScaleKind = "ordinal"
	ScaleContinuous ScaleKind = "continuous"
)

// IsValid returns true if the scale kind is one of the known variants.
func (k ScaleKind) IsValid() bool {
	switch k {
	case ScaleNominal, ScaleOrdinal, ScaleContinuous:
		return true
	}
	return false
}

func (k ScaleKind) String() string { return string(k) }

// AllScaleKinds is the canonical list of all known scale kinds.
var AllScaleKinds = []ScaleKind{
	ScaleNominal, ScaleOrdinal, ScaleContinuous,
}

// JSONSchema implements jsonschema.Exposer.
func (ScaleKind) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Scale Kind")
	s.WithDescription("Stevens measurement level: nominal (categories without order), ordinal (ordered categories), continuous (numeric range)")
	s.WithEnum("nominal", "ordinal", "continuous")
	s.WithExamples("nominal", "ordinal", "continuous")
	return s, nil
}

// ValidateScaleDomainCombo returns an error when the (scale, domain) combination
// is structurally incoherent:
//
//   - described + ordinal: rejected because ordinal requires explicit ordering
//     via permissible values (enumerated), not a range constraint.
//   - enumerated + continuous: rejected because continuous ranges must be described
//     with a JSON-schema constraint, not a finite list of permissible values.
func ValidateScaleDomainCombo(scale ScaleKind, domain ValueDomainKind) error {
	switch {
	case domain == DomainDescribed && scale == ScaleOrdinal:
		return fmt.Errorf("invalid scale+domain combination: described+ordinal is not allowed — ordinal domains require explicit ordering via enumerated permissible values (annotation_enums.go: ValidateScaleDomainCombo)")
	case domain == DomainEnumerated && scale == ScaleContinuous:
		return fmt.Errorf("invalid scale+domain combination: enumerated+continuous is not allowed — continuous domains require a described constraint spec (annotation_enums.go: ValidateScaleDomainCombo)")
	}
	return nil
}

// --- TargetKind ---

// TargetKind identifies which of an annotation's six target arms is populated:
// session, transcript entry, annotation, project, file_version, or association.
// The file_version arm identifies one repository-relative file at one content
// hash; the association arm identifies a durable association ID.
type TargetKind string

const (
	TargetSession     TargetKind = "session"
	TargetEntry       TargetKind = "entry"      // turn, tool call, tool result
	TargetAnnotation  TargetKind = "annotation" // meta-annotation
	TargetProject     TargetKind = "project"    // project-level annotation
	TargetFileVersion TargetKind = "file_version"
	TargetAssociation TargetKind = "association"
)

// IsValid returns true if the target kind is one of the known variants.
func (k TargetKind) IsValid() bool {
	switch k {
	case TargetSession, TargetEntry, TargetAnnotation, TargetProject, TargetFileVersion, TargetAssociation:
		return true
	}
	return false
}

func (k TargetKind) String() string { return string(k) }

// AllTargetKinds is the canonical list of all known target kinds.
var AllTargetKinds = []TargetKind{
	TargetSession, TargetEntry, TargetAnnotation, TargetProject, TargetFileVersion, TargetAssociation,
}

// JSONSchema implements jsonschema.Exposer.
func (TargetKind) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Target Kind")
	s.WithDescription("What is being annotated: session-level, entry-level (turn/tool call), meta-annotation, project-level, a specific file version (content-hash keyed read-state receipt), or a durable session-to-commit association")
	s.WithEnum("session", "entry", "annotation", "project", "file_version", "association")
	s.WithExamples("session", "entry", "file_version", "association")
	return s, nil
}
