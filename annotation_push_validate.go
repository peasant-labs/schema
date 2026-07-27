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
		if item.EntryTarget.SessionID == "" || item.EntryTarget.EndIndex <= item.EntryTarget.EntryIndex {
			return annotationTargetValidationErrorAt("schema.AnnotationPushItem.Validate", "targetKind entry has an empty sessionId or non-positive entry range", "an entry target requires a session and a half-open range containing at least one entry", "set entryTarget.sessionId and an endIndex greater than entryIndex")
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
