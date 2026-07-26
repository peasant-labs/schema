package schema

import "fmt"

// Validate checks that TargetKind selects exactly one AnnotationSummary target
// arm. It validates the response contract at its shared boundary so producers
// cannot emit mixed target identities.
func (a AnnotationSummary) Validate() error {
	if !a.TargetKind.IsValid() {
		return annotationTargetValidationError("targetKind is outside the closed target-kind set", "the annotation does not select a supported target discriminator", "use a member of schema.AllTargetKinds")
	}
	if a.TargetAssociationID != nil {
		if err := a.TargetAssociationID.Validate(); err != nil {
			return fmt.Errorf("annotation summary validation failed at schema.AnnotationSummary.Validate during wire-boundary validation: %w", err)
		}
	}
	hasNonAssociationTarget := a.TargetSessionID != nil || a.TargetEntryIndex != nil || a.TargetEntryEndIndex != nil || a.TargetAnnotID != nil || a.TargetProjectHash != nil || a.TargetFilePath != nil || a.TargetContentHash != nil
	switch a.TargetKind {
	case TargetAssociation:
		if a.TargetAssociationID == nil {
			return annotationTargetValidationError("targetKind association has no targetAssociationId", "the association arm requires the durable association identifier", "set targetAssociationId and clear every other target field")
		}
		if hasNonAssociationTarget {
			return annotationTargetValidationError("targetKind association mixes targetAssociationId with another target arm", "an annotation may name exactly one target identity", "keep only targetAssociationId for targetKind association")
		}
		return nil
	case TargetSession:
		if a.TargetAssociationID != nil || a.TargetSessionID == nil || *a.TargetSessionID == "" || a.TargetEntryIndex != nil || a.TargetEntryEndIndex != nil || a.TargetAnnotID != nil || a.TargetProjectHash != nil || a.TargetFilePath != nil || a.TargetContentHash != nil {
			return annotationTargetValidationError("targetKind session does not exclusively contain a non-empty targetSessionId", "the session arm requires exactly one session target", "set only targetSessionId for targetKind session")
		}
	case TargetEntry:
		if a.TargetAssociationID != nil || a.TargetSessionID == nil || *a.TargetSessionID == "" || a.TargetEntryIndex == nil || a.TargetAnnotID != nil || a.TargetProjectHash != nil || a.TargetFilePath != nil || a.TargetContentHash != nil {
			return annotationTargetValidationError("targetKind entry does not exclusively contain targetSessionId and targetEntryIndex", "the entry arm requires one session entry target", "set targetSessionId and targetEntryIndex only, with optional targetEntryEndIndex, for targetKind entry")
		}
		if a.TargetEntryEndIndex != nil && *a.TargetEntryEndIndex <= *a.TargetEntryIndex {
			return annotationTargetValidationError("targetKind entry has targetEntryEndIndex not greater than targetEntryIndex", "an entry range is half-open and must contain at least one entry", "set targetEntryEndIndex greater than targetEntryIndex or omit it for a single entry")
		}
	case TargetAnnotation:
		if a.TargetAssociationID != nil || a.TargetSessionID != nil || a.TargetEntryIndex != nil || a.TargetEntryEndIndex != nil || a.TargetAnnotID == nil || *a.TargetAnnotID == "" || a.TargetProjectHash != nil || a.TargetFilePath != nil || a.TargetContentHash != nil {
			return annotationTargetValidationError("targetKind annotation does not exclusively contain a non-empty targetAnnotationId", "the annotation arm requires exactly one annotation target", "set only targetAnnotationId for targetKind annotation")
		}
	case TargetProject:
		if a.TargetAssociationID != nil || a.TargetSessionID != nil || a.TargetEntryIndex != nil || a.TargetEntryEndIndex != nil || a.TargetAnnotID != nil || a.TargetProjectHash == nil || a.TargetFilePath != nil || a.TargetContentHash != nil {
			return annotationTargetValidationError("targetKind project does not exclusively contain targetProjectHash", "the project arm requires exactly one project target", "set only targetProjectHash for targetKind project")
		}
		if err := a.TargetProjectHash.Validate(); err != nil {
			return fmt.Errorf("annotation summary validation failed at schema.AnnotationSummary.Validate during wire-boundary validation: %w", err)
		}
	case TargetFileVersion:
		if a.TargetAssociationID != nil || a.TargetSessionID != nil || a.TargetEntryIndex != nil || a.TargetEntryEndIndex != nil || a.TargetAnnotID != nil || a.TargetProjectHash != nil || a.TargetFilePath == nil || *a.TargetFilePath == "" || a.TargetContentHash == nil || *a.TargetContentHash == "" {
			return annotationTargetValidationError("targetKind file_version does not exclusively contain targetFilePath and targetContentHash", "the file-version arm requires one path and one content hash", "set only non-empty targetFilePath and targetContentHash for targetKind file_version")
		}
	}
	return nil
}

func annotationTargetValidationError(what, why, remediation string) error {
	return fmt.Errorf("annotation summary validation failed at schema.AnnotationSummary.Validate during wire-boundary validation: %s; %s; %s", what, why, remediation)
}
