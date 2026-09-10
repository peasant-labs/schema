package schema

import (
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

type SourceEntryRef string
type SubmissionRef string
type PublicRevisionRef string

func newPublicGraphRef(raw, name string) (string, error) {
	if !validPublicRef(raw) {
		return "", fmt.Errorf("session graph validation failed at schema.New%s while constructing public evidence: value must contain 1..96 valid UTF-8 bytes and is preserved without trimming; the caller cannot safely attach graph evidence; provide the exact bounded opaque reference", name)
	}
	return raw, nil
}
func NewSourceEntryRef(raw string) (SourceEntryRef, error) {
	v, e := newPublicGraphRef(raw, "SourceEntryRef")
	return SourceEntryRef(v), e
}
func NewSubmissionRef(raw string) (SubmissionRef, error) {
	v, e := newPublicGraphRef(raw, "SubmissionRef")
	return SubmissionRef(v), e
}
func NewPublicRevisionRef(raw string) (PublicRevisionRef, error) {
	v, e := newPublicGraphRef(raw, "PublicRevisionRef")
	return PublicRevisionRef(v), e
}
func (v SourceEntryRef) Validate() error    { _, e := NewSourceEntryRef(string(v)); return e }
func (v SubmissionRef) Validate() error     { _, e := NewSubmissionRef(string(v)); return e }
func (v PublicRevisionRef) Validate() error { _, e := NewPublicRevisionRef(string(v)); return e }
func publicRefSchema(title string) jsonschema.Schema {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle(title)
	s.WithMinLength(1)
	s.WithMaxLength(96)
	s.WithFormat(PublicRefUTF8ByteFormat)
	return s
}

// PublicRefUTF8ByteFormat identifies the format assertion shared by generated
// validators for opaque public references. Unlike maxLength, which counts
// Unicode code points, this assertion counts the encoded UTF-8 bytes.
const PublicRefUTF8ByteFormat = "public-ref-utf8-96-bytes"

// ValidatePublicRefJSONSchemaFormat implements PublicRefUTF8ByteFormat for
// JSON Schema runtimes. Register it as a format assertion before compiling a
// schema that exposes one of the public reference types.
func ValidatePublicRefJSONSchemaFormat(value any) bool {
	raw, ok := value.(string)
	return ok && validPublicRef(raw)
}
func (SourceEntryRef) JSONSchema() (jsonschema.Schema, error) {
	return publicRefSchema("Source Entry Reference"), nil
}
func (SubmissionRef) JSONSchema() (jsonschema.Schema, error) {
	return publicRefSchema("Submission Reference"), nil
}
func (PublicRevisionRef) JSONSchema() (jsonschema.Schema, error) {
	return publicRefSchema("Public Revision Reference"), nil
}

type SessionRelationshipKind string

const (
	SessionRelationshipStartedBy   SessionRelationshipKind = "started_by"
	SessionRelationshipContextFrom SessionRelationshipKind = "context_from"
)

var AllSessionRelationshipKinds = []SessionRelationshipKind{SessionRelationshipStartedBy, SessionRelationshipContextFrom}

type RelationshipTargetState string

const (
	RelationshipTargetKnown                            RelationshipTargetState = "target_known"
	RelationshipTargetKnownRetained                    RelationshipTargetState = "target_known_retained"
	RelationshipTargetExplicitNone                     RelationshipTargetState = "explicit_none"
	RelationshipTargetUnknown                          RelationshipTargetState = "unknown"
	RelationshipTargetConflictingCurrentNativeEvidence RelationshipTargetState = "conflicting_current_native_evidence"
)

var AllRelationshipTargetStates = []RelationshipTargetState{RelationshipTargetKnown, RelationshipTargetKnownRetained, RelationshipTargetExplicitNone, RelationshipTargetUnknown, RelationshipTargetConflictingCurrentNativeEvidence}

type EvidenceKind string

const (
	EvidenceNativeTyped      EvidenceKind = "native_typed"
	EvidenceLifecycleTyped   EvidenceKind = "lifecycle_typed"
	EvidenceExistingAdapter  EvidenceKind = "existing_adapter"
	EvidenceRetainedLastGood EvidenceKind = "retained_last_good"
	EvidenceUnknown          EvidenceKind = "unknown"
	EvidenceConflict         EvidenceKind = "conflict"
)

var AllEvidenceKinds = []EvidenceKind{EvidenceNativeTyped, EvidenceLifecycleTyped, EvidenceExistingAdapter, EvidenceRetainedLastGood, EvidenceUnknown, EvidenceConflict}

type SessionPurpose string

const (
	SessionPurposeInteraction   SessionPurpose = "interaction"
	SessionPurposeDelegatedWork SessionPurpose = "delegated_work"
	SessionPurposeHelperReview  SessionPurpose = "helper_review"
	SessionPurposeUnknown       SessionPurpose = "unknown"
)

var AllSessionPurposes = []SessionPurpose{SessionPurposeInteraction, SessionPurposeDelegatedWork, SessionPurposeHelperReview, SessionPurposeUnknown}

type ContentOrigin string

const (
	ContentOriginSubmittedInput     ContentOrigin = "submitted_input"
	ContentOriginHarnessContext     ContentOrigin = "harness_context"
	ContentOriginAgentOutput        ContentOrigin = "agent_output"
	ContentOriginAgentCommunication ContentOrigin = "agent_communication"
	ContentOriginToolActivity       ContentOrigin = "tool_activity"
	ContentOriginSystemControl      ContentOrigin = "system_control"
	ContentOriginGeneratedSummary   ContentOrigin = "generated_summary"
	ContentOriginUnknown            ContentOrigin = "unknown"
)

var AllContentOrigins = []ContentOrigin{ContentOriginSubmittedInput, ContentOriginHarnessContext, ContentOriginAgentOutput, ContentOriginAgentCommunication, ContentOriginToolActivity, ContentOriginSystemControl, ContentOriginGeneratedSummary, ContentOriginUnknown}

type ActorOrigin string

const (
	ActorOriginOperator      ActorOrigin = "operator"
	ActorOriginAgentDelegate ActorOrigin = "agent_delegate"
	ActorOriginHarness       ActorOrigin = "harness"
	ActorOriginUnknown       ActorOrigin = "unknown"
)

var AllActorOrigins = []ActorOrigin{ActorOriginOperator, ActorOriginAgentDelegate, ActorOriginHarness, ActorOriginUnknown}

type DeliveryOrigin string

const (
	DeliveryOriginSessionAdmission DeliveryOrigin = "session_admission"
	DeliveryOriginGuardianReview   DeliveryOrigin = "guardian_review"
	DeliveryOriginSubagentDelivery DeliveryOrigin = "subagent_delivery"
	DeliveryOriginInheritedContext DeliveryOrigin = "inherited_context"
	DeliveryOriginToolDelivery     DeliveryOrigin = "tool_delivery"
	DeliveryOriginSystemLifecycle  DeliveryOrigin = "system_lifecycle"
	DeliveryOriginUnknown          DeliveryOrigin = "unknown"
)

var AllDeliveryOrigins = []DeliveryOrigin{DeliveryOriginSessionAdmission, DeliveryOriginGuardianReview, DeliveryOriginSubagentDelivery, DeliveryOriginInheritedContext, DeliveryOriginToolDelivery, DeliveryOriginSystemLifecycle, DeliveryOriginUnknown}

type ContentOwnership string

const (
	ContentOwnershipLocal     ContentOwnership = "local"
	ContentOwnershipInherited ContentOwnership = "inherited"
	ContentOwnershipUncertain ContentOwnership = "uncertain"
)

var AllContentOwnerships = []ContentOwnership{ContentOwnershipLocal, ContentOwnershipInherited, ContentOwnershipUncertain}

type InputModality string

const (
	InputModalityNone       InputModality = "none"
	InputModalityText       InputModality = "text"
	InputModalityMedia      InputModality = "media"
	InputModalityUserAction InputModality = "user_action"
	InputModalityMixed      InputModality = "mixed"
	InputModalityUnknown    InputModality = "unknown"
)

var AllInputModalities = []InputModality{InputModalityNone, InputModalityText, InputModalityMedia, InputModalityUserAction, InputModalityMixed, InputModalityUnknown}

type PublicSourceAnchorKind string

const (
	PublicSourceAnchorGeneral PublicSourceAnchorKind = "general_source_session"
	PublicSourceAnchorBefore  PublicSourceAnchorKind = "before_redacted_entry"
	PublicSourceAnchorThrough PublicSourceAnchorKind = "through_redacted_entry"
)

var AllPublicSourceAnchorKinds = []PublicSourceAnchorKind{PublicSourceAnchorGeneral, PublicSourceAnchorBefore, PublicSourceAnchorThrough}

type EarlierHistoryState string

const (
	EarlierHistoryUncertainMigrated   EarlierHistoryState = "uncertain_migrated"
	EarlierHistoryUncertainUnresolved EarlierHistoryState = "uncertain_unresolved"
)

var AllEarlierHistoryStates = []EarlierHistoryState{EarlierHistoryUncertainMigrated, EarlierHistoryUncertainUnresolved}

func inSet[T comparable](v T, all []T) bool {
	for _, x := range all {
		if v == x {
			return true
		}
	}
	return false
}
func newClosedValue[T ~string](raw, name string, all []T) (T, error) {
	v := T(raw)
	if !inSet(v, all) {
		return "", fmt.Errorf("session graph validation failed at schema.New%s while constructing evidence: value %q is outside the closed set; callers cannot classify the wire value; use one of the published values", name, raw)
	}
	return v, nil
}
func NewSessionRelationshipKind(raw string) (SessionRelationshipKind, error) {
	return newClosedValue(raw, "SessionRelationshipKind", AllSessionRelationshipKinds)
}
func NewRelationshipTargetState(raw string) (RelationshipTargetState, error) {
	return newClosedValue(raw, "RelationshipTargetState", AllRelationshipTargetStates)
}
func NewEvidenceKind(raw string) (EvidenceKind, error) {
	return newClosedValue(raw, "EvidenceKind", AllEvidenceKinds)
}
func NewSessionPurpose(raw string) (SessionPurpose, error) {
	return newClosedValue(raw, "SessionPurpose", AllSessionPurposes)
}
func NewContentOrigin(raw string) (ContentOrigin, error) {
	return newClosedValue(raw, "ContentOrigin", AllContentOrigins)
}
func NewActorOrigin(raw string) (ActorOrigin, error) {
	return newClosedValue(raw, "ActorOrigin", AllActorOrigins)
}
func NewDeliveryOrigin(raw string) (DeliveryOrigin, error) {
	return newClosedValue(raw, "DeliveryOrigin", AllDeliveryOrigins)
}
func NewContentOwnership(raw string) (ContentOwnership, error) {
	return newClosedValue(raw, "ContentOwnership", AllContentOwnerships)
}
func NewInputModality(raw string) (InputModality, error) {
	return newClosedValue(raw, "InputModality", AllInputModalities)
}
func NewPublicSourceAnchorKind(raw string) (PublicSourceAnchorKind, error) {
	return newClosedValue(raw, "PublicSourceAnchorKind", AllPublicSourceAnchorKinds)
}
func NewEarlierHistoryState(raw string) (EarlierHistoryState, error) {
	return newClosedValue(raw, "EarlierHistoryState", AllEarlierHistoryStates)
}
func (v SessionRelationshipKind) IsValid() bool { return inSet(v, AllSessionRelationshipKinds) }
func (v RelationshipTargetState) IsValid() bool { return inSet(v, AllRelationshipTargetStates) }
func (v EvidenceKind) IsValid() bool            { return inSet(v, AllEvidenceKinds) }
func (v SessionPurpose) IsValid() bool          { return v == "" || inSet(v, AllSessionPurposes) }
func (v ContentOrigin) IsValid() bool           { return inSet(v, AllContentOrigins) }
func (v ActorOrigin) IsValid() bool             { return inSet(v, AllActorOrigins) }
func (v DeliveryOrigin) IsValid() bool          { return inSet(v, AllDeliveryOrigins) }
func (v ContentOwnership) IsValid() bool        { return inSet(v, AllContentOwnerships) }
func (v InputModality) IsValid() bool           { return inSet(v, AllInputModalities) }
func (v PublicSourceAnchorKind) IsValid() bool  { return inSet(v, AllPublicSourceAnchorKinds) }
func (v EarlierHistoryState) IsValid() bool     { return inSet(v, AllEarlierHistoryStates) }

func enumSchema[T ~string](title string, all []T) (jsonschema.Schema, error) {
	return closedStringEnumSchema(title, "Closed session graph value", all), nil
}
func (SessionRelationshipKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Session Relationship Kind", AllSessionRelationshipKinds)
}
func (RelationshipTargetState) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Relationship Target State", AllRelationshipTargetStates)
}
func (EvidenceKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Evidence Kind", AllEvidenceKinds)
}
func (SessionPurpose) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Session Purpose", AllSessionPurposes)
}
func (ContentOrigin) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Content Origin", AllContentOrigins)
}
func (ActorOrigin) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Actor Origin", AllActorOrigins)
}
func (DeliveryOrigin) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Delivery Origin", AllDeliveryOrigins)
}
func (ContentOwnership) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Content Ownership", AllContentOwnerships)
}
func (InputModality) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Input Modality", AllInputModalities)
}
func (PublicSourceAnchorKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Public Source Anchor Kind", AllPublicSourceAnchorKinds)
}
func (EarlierHistoryState) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Earlier History State", AllEarlierHistoryStates)
}

type PublicSourceAnchor struct {
	Kind              PublicSourceAnchorKind `json:"kind" yaml:"kind"`
	SourceEntryRef    SourceEntryRef         `json:"sourceEntryRef,omitempty" yaml:"sourceEntryRef"`
	SourceRevisionRef PublicRevisionRef      `json:"sourceRevisionRef,omitempty" yaml:"sourceRevisionRef"`
}

func (a PublicSourceAnchor) Validate() error {
	if !a.Kind.IsValid() {
		return fmt.Errorf("session graph validation failed at schema.PublicSourceAnchor.Validate: kind %q is outside the closed set; navigation cannot interpret the anchor; use a published anchor kind", a.Kind)
	}
	entry, rev := a.SourceEntryRef != "", a.SourceRevisionRef != ""
	if a.Kind == PublicSourceAnchorGeneral {
		if entry || rev {
			return fmt.Errorf("session graph validation failed at schema.PublicSourceAnchor.Validate: general_source_session carries an entry or revision; the link could imply an unverified boundary; omit both references")
		}
		return nil
	}
	if !entry || !rev {
		return fmt.Errorf("session graph validation failed at schema.PublicSourceAnchor.Validate: a redacted-entry anchor requires both sourceEntryRef and sourceRevisionRef; an exact public boundary cannot be verified; provide both bounded references or use general_source_session")
	}
	if err := a.SourceEntryRef.Validate(); err != nil {
		return err
	}
	return a.SourceRevisionRef.Validate()
}

type SessionRelationship struct {
	Kind          SessionRelationshipKind `json:"kind" yaml:"kind"`
	TargetState   RelationshipTargetState `json:"targetState" yaml:"targetState"`
	TargetLocalID *SessionID              `json:"targetLocalId,omitempty" yaml:"targetLocalId"`
	Evidence      EvidenceKind            `json:"evidence" yaml:"evidence"`
	Anchor        *PublicSourceAnchor     `json:"anchor,omitempty" yaml:"anchor"`
}

func (r SessionRelationship) Validate() error {
	if !r.Kind.IsValid() || !r.TargetState.IsValid() || !r.Evidence.IsValid() {
		return fmt.Errorf("session graph validation failed at schema.SessionRelationship.Validate: kind, targetState, or evidence is outside its closed set; consumers cannot interpret the relationship; use published values")
	}
	known := r.TargetState == RelationshipTargetKnown || r.TargetState == RelationshipTargetKnownRetained
	if known != (r.TargetLocalID != nil) {
		return fmt.Errorf("session graph validation failed at schema.SessionRelationship.Validate: target state and targetLocalId presence disagree; navigation could target the wrong session; known states require one ID and all other states forbid it")
	}
	if r.TargetLocalID != nil {
		if _, e := NewSessionID(string(*r.TargetLocalID)); e != nil {
			return fmt.Errorf("session graph validation failed at schema.SessionRelationship.Validate: targetLocalId is malformed; navigation cannot resolve it; provide a valid owner-local session ID: %w", e)
		}
	}
	if r.Anchor != nil {
		if r.Kind != SessionRelationshipContextFrom || !known {
			return fmt.Errorf("session graph validation failed at schema.SessionRelationship.Validate: anchor is allowed only on a known context_from relationship; consumers cannot bind it safely; omit the anchor or correct the relationship")
		}
		if e := r.Anchor.Validate(); e != nil {
			return e
		}
	}
	return nil
}
func ValidateSessionRelationships(rs []SessionRelationship) error {
	seen := map[SessionRelationshipKind]bool{}
	for _, r := range rs {
		if seen[r.Kind] {
			return fmt.Errorf("session graph validation failed at schema.ValidateSessionRelationships: relationship kind %q occurs more than once; navigation pairing is ambiguous; emit at most one relationship per kind", r.Kind)
		}
		seen[r.Kind] = true
		if e := r.Validate(); e != nil {
			return e
		}
	}
	return nil
}

type ContentProvenance struct {
	Origin        ContentOrigin    `json:"origin" yaml:"origin"`
	Actor         ActorOrigin      `json:"actor" yaml:"actor"`
	Delivery      DeliveryOrigin   `json:"delivery" yaml:"delivery"`
	Ownership     ContentOwnership `json:"ownership" yaml:"ownership"`
	Evidence      EvidenceKind     `json:"evidence" yaml:"evidence"`
	InputModality InputModality    `json:"inputModality" yaml:"inputModality"`
	SubmissionRef SubmissionRef    `json:"submissionRef,omitempty" yaml:"submissionRef"`
}

func (p ContentProvenance) Validate() error {
	if !p.Origin.IsValid() || !p.Actor.IsValid() || !p.Delivery.IsValid() || !p.Ownership.IsValid() || !p.Evidence.IsValid() || !p.InputModality.IsValid() {
		return fmt.Errorf("session graph validation failed at schema.ContentProvenance.Validate: one or more evidence dimensions are missing or outside their closed sets; content cannot be classified safely; provide every dimension using published values, including unknown")
	}
	if p.SubmissionRef != "" {
		return p.SubmissionRef.Validate()
	}
	return nil
}

type EarlierHistorySection struct {
	State          EarlierHistoryState    `json:"state" yaml:"state"`
	Turns          []TurnDetail           `json:"turns" yaml:"turns" nullable:"false"`
	NativeMetadata []NativeMetadataRecord `json:"nativeMetadata,omitempty" yaml:"nativeMetadata"`
}

func (s EarlierHistorySection) Validate() error {
	if !s.State.IsValid() {
		return fmt.Errorf("session graph validation failed at schema.EarlierHistorySection.Validate: state %q is outside the closed set; consumers cannot explain retained history; use a published state", s.State)
	}
	if s.Turns == nil {
		return fmt.Errorf("session graph validation failed at schema.EarlierHistorySection.Validate: turns is null; retained history ordering is unavailable; emit a non-null array")
	}
	return ValidateNativeMetadataRecords(s.NativeMetadata, s.Turns)
}
