package schema

// ValueDomain defines the permissible values for an annotation type (ISO 11179).
// For enumerated domains, PermissibleValues holds the finite set of allowed strings.
// For described domains, ConstraintSpec holds a JSON-encoded range or pattern.
type ValueDomain struct {
	Kind              ValueDomainKind    `json:"kind"`
	Datatype          AnnotationDatatype `json:"datatype"`
	PermissibleValues []string           `json:"permissibleValues,omitempty"` // for enumerated
	ConstraintSpec    string             `json:"constraintSpec,omitempty"`    // for described; JSON
}

// Provenance records how a non-human annotator derived its annotation value.
// Only populated for rule-based and agent annotators; nil for human annotations.
type Provenance struct {
	Method   string            `json:"method"` // "heuristic", "regex", "llm_judge", "manual"
	Function string            `json:"function,omitempty"`
	Version  string            `json:"version,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

// AnnotationTypeSummary is the wire format for annotation types in API responses.
// Class is derived via the annotation_families → annotation_classes join (BCNF: no redundant column).
type AnnotationTypeSummary struct {
	ID               string           `json:"id,omitempty"` // UUID PK — populated from store for CLI use; omitted in REST responses
	TypeID           string           `json:"typeId"`
	Version          int              `json:"version"`
	DisplayName      string           `json:"displayName"`
	Description      string           `json:"description,omitempty"`
	Family           string           `json:"family"`
	Class            string           `json:"class"`
	ScaleKind        ScaleKind        `json:"scaleKind,omitempty"` // measurement level (ISO 11179 Part 5)
	ValueDomain      ValueDomain      `json:"valueDomain"`
	LowerIsBetter    *bool            `json:"lowerIsBetter,omitempty"`
	Status           AnnotationStatus `json:"status"`
	Origin           TypeOrigin       `json:"origin"`
	PriorityOverride *int             `json:"priorityOverride,omitempty"`
	// AllowedTargetKinds lists the target kinds this type may annotate (V16).
	// Empty/omitted means the type places no restriction (all kinds allowed).
	// Clients filter entry-level pickers to types whose list includes "entry".
	AllowedTargetKinds []TargetKind `json:"allowedTargetKinds,omitempty"`
}

// AnnotationSummary is the wire format for annotations in API responses.
// TargetKind is derived via TPT child table JOINs (annotations_with_target view).
type AnnotationSummary struct {
	ID                  string       `json:"id"`
	TargetKind          TargetKind   `json:"targetKind" yaml:"targetKind"`
	TargetSessionID     *string      `json:"targetSessionId,omitempty" yaml:"targetSessionId,omitempty"`
	TargetEntryIndex    *int         `json:"targetEntryIndex,omitempty" yaml:"targetEntryIndex,omitempty"`
	TargetEntryEndIndex *int         `json:"targetEntryEndIndex,omitempty" yaml:"targetEntryEndIndex,omitempty"` // V16: half-open [start, end)
	TargetAnnotID       *string      `json:"targetAnnotationId,omitempty" yaml:"targetAnnotationId,omitempty"`
	TargetProjectHash   *ProjectHash `json:"targetProjectHash,omitempty" yaml:"targetProjectHash,omitempty"`
	// TargetAssociationID identifies a durable session-to-commit association.
	// It is an ID target, never an embedded association copy.
	TargetAssociationID *AssociationID `json:"targetAssociationId,omitempty" yaml:"targetAssociationId,omitempty"`
	// TargetFilePath and TargetContentHash discriminate a TargetFileVersion
	// annotation (the 5th TPT arm): a whole-file
	// read-state receipt keyed to a specific content hash of a specific
	// repo-relative path, so an agent edit that changes the content hash
	// invalidates the receipt without deleting it.
	TargetFilePath    *string       `json:"targetFilePath,omitempty" yaml:"targetFilePath,omitempty"`
	TargetContentHash *string       `json:"targetContentHash,omitempty" yaml:"targetContentHash,omitempty"`
	IsPrimary         bool          `json:"isPrimary"`
	AnnotatorKind     AnnotatorKind `json:"annotatorKind"`
	AnnotatorName     string        `json:"annotatorName"`
	TypeID            string        `json:"typeId"`
	TypeName          string        `json:"typeName"`
	Value             string        `json:"value"`
	Confidence        *float64      `json:"confidence,omitempty"`
	Reason            *string       `json:"reason,omitempty"`
	Provenance        *Provenance   `json:"provenance,omitempty"`
	ContentHash       *string       `json:"contentHash,omitempty"` // V16: push dedup
	CreatedAt         int64         `json:"createdAt"`
	SupersededBy      *string       `json:"supersededBy,omitempty"`
}

// AnnotatorSummary is the wire format for annotators in API responses.
// ModelID and ProviderKey are only populated for agent annotators.
type AnnotatorSummary struct {
	ID          string        `json:"id"`
	Kind        AnnotatorKind `json:"kind"`
	Name        string        `json:"name"`
	DisplayName string        `json:"displayName"`
	Description string        `json:"description,omitempty"`
	ModelID     *string       `json:"modelId,omitempty"`
	// DO NOT TOUCH (TRAP): ProviderKey is the model-VENDOR credential
	// (e.g. "anthropic"), a DIFFERENT axis from the coding-tool Harness. The
	// harness-key changeover left this as json:"providerKey" on purpose.
	// Never flip it to json:"harness" — enforced by ast-grep/no-trap-harness-flip.yml.
	ProviderKey *string `json:"providerKey,omitempty"`
	Status      string  `json:"status"`
}

// TaxonomyNode represents a class node in the class > family > type taxonomy tree.
type TaxonomyNode struct {
	Class    string               `json:"class"`
	Families []TaxonomyFamilyNode `json:"families"`
}

// TaxonomyFamilyNode represents a family node with its associated annotation types.
type TaxonomyFamilyNode struct {
	Family string                  `json:"family"`
	Types  []AnnotationTypeSummary `json:"types"`
}

// --- Batch Annotation API ---

// BatchCreateAnnotationsRequest is the JSON body for POST /api/v1/annotations/batch.
// All annotations are committed as a single all-or-nothing SQLite transaction.
type BatchCreateAnnotationsRequest struct {
	Annotations []CreateAnnotationRequest `json:"annotations"`
}

// BatchCreateAnnotationsResponse is the JSON response for POST /api/v1/annotations/batch (201 Created).
// IDs are returned in the same order as the request Annotations slice.
type BatchCreateAnnotationsResponse struct {
	IDs []string `json:"ids"`
}

// BatchCreateAnnotationsErrorResponse is the JSON error response for POST /api/v1/annotations/batch (400).
// FailingIndex is the zero-based index of the first annotation that failed validation.
type BatchCreateAnnotationsErrorResponse struct {
	Error        string `json:"error"`
	FailingIndex int    `json:"failingIndex"`
}
