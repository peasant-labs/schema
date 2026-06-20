package schema

// SessionIdentity holds the identification fields for a session.
// JSON tags use camelCase to match CLI UnifiedMetadata wire format.
type SessionIdentity struct {
	SessionID       SessionID  `json:"sessionId"`
	ParentSessionID *SessionID `json:"parentUuid,omitempty"`
	SchemaVersion   int        `json:"schemaVersion"`
}
