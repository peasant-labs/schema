package schema

import "fmt"

// Validate checks that an annotation push item names one exclusive target arm.
// Association target validation is shared with AnnotationSummary so ingress and
// returned annotations cannot disagree about when targetAssociationId is valid.
func (item AnnotationPushItem) Validate() error {
	if !item.TargetKind.IsValid() {
		return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind is outside the closed target-kind set", "the annotation does not select a supported target discriminator", "use a member of schema.AllTargetKinds")
	}
	hasNonAssociationTarget := item.SessionID != nil || item.EntryTarget != nil || item.AnnotationID != nil || item.ProjectHash != nil
	if err := validateAssociationAnnotationTarget("schema.AnnotationPushItem.Validate", item.TargetKind, item.TargetAssociationID, hasNonAssociationTarget); err != nil {
		return err
	}

	switch item.TargetKind {
	case TargetAssociation:
		return nil
	case TargetSession:
		if item.SessionID == nil || *item.SessionID == "" || item.EntryTarget != nil || item.AnnotationID != nil || item.ProjectHash != nil {
			return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind session does not exclusively contain a non-empty sessionId", "the session arm requires exactly one session target", "set only sessionId for targetKind session")
		}
	case TargetEntry:
		if item.SessionID != nil || item.EntryTarget == nil || item.AnnotationID != nil || item.ProjectHash != nil {
			return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind entry does not exclusively contain entryTarget", "the entry arm requires exactly one entry target", "set only entryTarget for targetKind entry")
		}
		if err := item.EntryTarget.Validate(); err != nil {
			return fmt.Errorf("annotation push item validation failed at schema.AnnotationPushItem.Validate during wire-boundary validation: %w", err)
		}
	case TargetAnnotation:
		if item.SessionID != nil || item.EntryTarget != nil || item.AnnotationID == nil || *item.AnnotationID == "" || item.ProjectHash != nil {
			return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind annotation does not exclusively contain a non-empty annotationId", "the annotation arm requires exactly one annotation target", "set only annotationId for targetKind annotation")
		}
	case TargetProject:
		if item.SessionID != nil || item.EntryTarget != nil || item.AnnotationID != nil || item.ProjectHash == nil {
			return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind project does not exclusively contain projectHash", "the project arm requires exactly one project target", "set only projectHash for targetKind project")
		}
		if err := item.ProjectHash.Validate(); err != nil {
			return fmt.Errorf("annotation push item validation failed at schema.AnnotationPushItem.Validate during wire-boundary validation: %w", err)
		}
	case TargetFileVersion:
		return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind file_version has no push target representation", "the annotation push item has no file path and content hash fields for a file-version target", "use a supported push target kind or extend the wire before publishing file-version annotations")
	}
	return nil
}

// Validate checks the semantic invariants of an entry-level annotation target.
// JSON Schema can enforce the non-empty session identifier, while this runtime
// predicate owns the relational half-open range rule shared by every Go request
// and response path that validates an entry target.
func (target AnnotationEntryTarget) Validate() error {
	return validateAnnotationEntryTarget("schema.AnnotationEntryTarget.Validate", target.SessionID, target.EntryIndex, &target.EndIndex)
}

func validateAnnotationEntryTarget(location, sessionID string, entryIndex int, endIndex *int) error {
	if sessionID == "" {
		return annotationTargetValidationErrorAt(location, "entry target has an empty sessionId", "an entry target must identify the session containing the annotated entries", "set the entry target sessionId to a non-empty session ID")
	}
	if endIndex != nil && *endIndex <= entryIndex {
		return annotationTargetValidationErrorAt(location, "entry target has a non-positive half-open range", "an entry target range must contain at least one entry", "set the entry target end index greater than its entry index")
	}
	return nil
}

// Validate checks every annotation in a pushed batch through the public item
// validator. A caller can reject the batch before persistence when one item is
// malformed, preserving all-or-nothing handler semantics.
func (r AnnotationPushRequest) Validate() error {
	if r.Annotations == nil {
		return fmt.Errorf("annotation push request validation failed at schema.AnnotationPushRequest.Validate during wire-boundary validation: annotations is null; the request must carry an annotation array so the handler can validate the complete batch; send an empty array when there are no annotations")
	}
	for index, item := range r.Annotations {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("annotation push request validation failed at schema.AnnotationPushRequest.Validate during wire-boundary validation: annotations[%d]: %w", index, err)
		}
	}
	return nil
}
