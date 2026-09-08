package schema

import (
	"fmt"
	jsonschema "github.com/swaggest/jsonschema-go"
)

type RelationshipNavigationStatus string

const (
	RelationshipNavigationResolved         RelationshipNavigationStatus = "resolved"
	RelationshipNavigationGeneralLinkOnly  RelationshipNavigationStatus = "general_link_only"
	RelationshipNavigationKnownUnavailable RelationshipNavigationStatus = "known_unavailable"
	RelationshipNavigationInaccessible     RelationshipNavigationStatus = "inaccessible"
	RelationshipNavigationUnknown          RelationshipNavigationStatus = "unknown"
	RelationshipNavigationConflicting      RelationshipNavigationStatus = "conflicting"
)

var AllRelationshipNavigationStatuses = []RelationshipNavigationStatus{RelationshipNavigationResolved, RelationshipNavigationGeneralLinkOnly, RelationshipNavigationKnownUnavailable, RelationshipNavigationInaccessible, RelationshipNavigationUnknown, RelationshipNavigationConflicting}

func (v RelationshipNavigationStatus) IsValid() bool {
	return inSet(v, AllRelationshipNavigationStatuses)
}
func NewRelationshipNavigationStatus(raw string) (RelationshipNavigationStatus, error) {
	return newClosedValue(raw, "RelationshipNavigationStatus", AllRelationshipNavigationStatuses)
}
func (RelationshipNavigationStatus) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Relationship Navigation Status", AllRelationshipNavigationStatuses)
}

type SessionRelationshipNavigation struct {
	Kind         SessionRelationshipKind      `json:"kind" yaml:"kind"`
	Status       RelationshipNavigationStatus `json:"status" yaml:"status"`
	LocalID      *SessionID                   `json:"localId,omitempty" yaml:"localId"`
	TranscriptID *TranscriptID                `json:"transcriptId,omitempty" yaml:"transcriptId"`
	Anchor       *PublicSourceAnchor          `json:"anchor,omitempty" yaml:"anchor"`
}

func (n SessionRelationshipNavigation) Validate() error {
	if !n.Kind.IsValid() || !n.Status.IsValid() {
		return fmt.Errorf("session relationship navigation validation failed at schema.SessionRelationshipNavigation.Validate: kind or status is outside its closed set; the host cannot render a safe target; use published values")
	}
	ids := 0
	if n.LocalID != nil {
		ids++
		if _, e := NewSessionID(string(*n.LocalID)); e != nil {
			return e
		}
	}
	if n.TranscriptID != nil {
		ids++
		if _, e := NewTranscriptID(string(*n.TranscriptID)); e != nil {
			return e
		}
	}
	resolved := n.Status == RelationshipNavigationResolved || n.Status == RelationshipNavigationGeneralLinkOnly
	if resolved && ids != 1 || !resolved && ids != 0 {
		return fmt.Errorf("session relationship navigation validation failed at schema.SessionRelationshipNavigation.Validate: status requires exactly one local or public target only when linkable; the host could leak or misroute a target; correct the status and ID presence")
	}
	if n.Anchor != nil {
		if !resolved || n.Kind != SessionRelationshipContextFrom {
			return fmt.Errorf("session relationship navigation validation failed at schema.SessionRelationshipNavigation.Validate: anchor is allowed only for linkable context_from navigation; omit the unsafe anchor")
		}
		if n.Status == RelationshipNavigationGeneralLinkOnly && n.Anchor.Kind != PublicSourceAnchorGeneral {
			return fmt.Errorf("session relationship navigation validation failed at schema.SessionRelationshipNavigation.Validate: general_link_only carries an exact redacted-entry anchor; the status promises only a general source link and consumers could follow an unverified boundary; omit the anchor or use general_source_session, or emit resolved after verifying the exact public revision")
		}
		return n.Anchor.Validate()
	}
	return nil
}

type SessionListItemKind string

const (
	SessionListItemTranscript       SessionListItemKind = "transcript"
	SessionListItemContextContainer SessionListItemKind = "context_container"
)

var AllSessionListItemKinds = []SessionListItemKind{SessionListItemTranscript, SessionListItemContextContainer}

func (v SessionListItemKind) IsValid() bool { return inSet(v, AllSessionListItemKinds) }
func NewSessionListItemKind(raw string) (SessionListItemKind, error) {
	return newClosedValue(raw, "SessionListItemKind", AllSessionListItemKinds)
}
func (SessionListItemKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Session List Item Kind", AllSessionListItemKinds)
}

type HelperGroupSummary struct {
	GroupID           string         `json:"groupId" yaml:"groupId"`
	Purpose           SessionPurpose `json:"purpose" yaml:"purpose"`
	HelperThreadCount int            `json:"helperThreadCount" yaml:"helperThreadCount" minimum:"0" maximum:"9007199254740991"`
	MemberScope       string         `json:"memberScope" yaml:"memberScope"`
}

func (s HelperGroupSummary) Validate() error {
	if s.GroupID == "" || s.MemberScope == "" {
		return fmt.Errorf("helper group validation failed at schema.HelperGroupSummary.Validate: groupId and memberScope must be non-empty; members cannot be fetched within the original query scope; refresh the grouped list")
	}
	if s.Purpose != SessionPurposeHelperReview {
		return fmt.Errorf("helper group validation failed at schema.HelperGroupSummary.Validate: purpose %q is not helper_review; this grouped surface cannot classify the members; emit only saved helper-review groups", s.Purpose)
	}
	if s.HelperThreadCount < 0 || int64(s.HelperThreadCount) > maxSafeJSONInteger {
		return fmt.Errorf("helper group validation failed at schema.HelperGroupSummary.Validate: helperThreadCount is negative; saved thread identity totals cannot be represented; emit a nonnegative count")
	}
	return nil
}

func validateHelperGroups(groups []HelperGroupSummary) error {
	seen := make(map[string]bool)
	for _, group := range groups {
		if err := group.Validate(); err != nil {
			return err
		}
		if seen[group.GroupID] {
			return fmt.Errorf("helper group validation failed at schema.validateHelperGroups during read projection: duplicate groupId; clients cannot maintain independent disclosure and paging state; emit each immediate-owner group once")
		}
		seen[group.GroupID] = true
	}
	return nil
}

func validateGroupedPagination(page, limit int, totals ...int) error {
	if page < 1 || limit < 1 || int64(page) > maxSafeJSONInteger || int64(limit) > maxSafeJSONInteger {
		return fmt.Errorf("grouped pagination validation failed at schema.validateGroupedPagination during read construction: page or limit is outside 1..9007199254740991; clients cannot page exactly; emit positive safe integers")
	}
	for _, total := range totals {
		if total < 0 || int64(total) > maxSafeJSONInteger {
			return fmt.Errorf("grouped pagination validation failed at schema.validateGroupedPagination during read construction: total is outside 0..9007199254740991; clients cannot preserve counts exactly; emit nonnegative safe integers")
		}
	}
	return nil
}

type HelperContextSummary struct {
	GroupID     string                       `json:"groupId" yaml:"groupId"`
	OwnerStatus RelationshipNavigationStatus `json:"ownerStatus" yaml:"ownerStatus"`
}

func (s HelperContextSummary) Validate() error {
	if s.GroupID == "" {
		return fmt.Errorf("helper context validation failed at schema.HelperContextSummary.Validate: groupId is empty; the context cannot be paired with its helper group; emit the stable group ID")
	}
	if !s.OwnerStatus.IsValid() {
		return fmt.Errorf("helper context validation failed at schema.HelperContextSummary.Validate: ownerStatus %q is outside the closed set; owner availability cannot be represented; use a published status", s.OwnerStatus)
	}
	return nil
}
