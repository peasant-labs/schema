package schema

import (
	"fmt"
	"strings"
)

// Validate checks that a published association can be addressed durably by a
// consumer. ObservedCommitHash remains producer data rather than a re-parsed
// Git object, but it must be present so it can participate in canonical
// owner/transcript/observed-hash identity.
func (a PublishedAssociation) Validate() error {
	if err := a.ID.Validate(); err != nil {
		return fmt.Errorf("published association validation failed at schema.PublishedAssociation.Validate during publish-request validation: %w", err)
	}
	if strings.TrimSpace(a.ObservedCommitHash) == "" {
		return fmt.Errorf("published association validation failed at schema.PublishedAssociation.Validate during publish-request validation: observedCommitHash is empty; consumers need the observed commit hash to preserve the canonical durable relationship; provide the commit hash observed for this session association")
	}
	return nil
}

// Validate checks request-local published association invariants. Associations
// are optional for publishers that predate this contract. When supplied, IDs
// and observed hashes are each unique so a request cannot introduce either an
// ambiguous durable identity or a same-relationship alias.
func (g GitContext) Validate() error {
	seenIDs := make(map[AssociationID]int, len(g.Associations))
	seenHashes := make(map[string]int, len(g.Associations))
	for index, association := range g.Associations {
		if err := association.Validate(); err != nil {
			return fmt.Errorf("git context validation failed at schema.GitContext.Validate during publish-request validation: associations[%d]: %w", index, err)
		}
		if prior, exists := seenIDs[association.ID]; exists {
			return fmt.Errorf("git context validation failed at schema.GitContext.Validate during publish-request validation: associations[%d] repeats ID %q from associations[%d]; request-local association IDs must be unique so consumers can retain one durable owner-scoped association; assign a distinct producer-owned ID", index, association.ID, prior)
		}
		if prior, exists := seenHashes[association.ObservedCommitHash]; exists {
			return fmt.Errorf("git context validation failed at schema.GitContext.Validate during publish-request validation: associations[%d] repeats observedCommitHash %q from associations[%d]; one owner transcript observed-commit relationship has one durable ID and aliases are forbidden; reuse the original association ID instead of adding a second association", index, association.ObservedCommitHash, prior)
		}
		seenIDs[association.ID] = index
		seenHashes[association.ObservedCommitHash] = index
	}
	return nil
}

// Validate checks the request-local associations carried in the publish wire.
// Other publish-body validation remains the responsibility of the generated
// PublishRequest JSON Schema at the HTTP boundary.
func (r PublishRequest) Validate() error {
	if err := r.Git.Validate(); err != nil {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: %w", err)
	}
	return nil
}

// ValidatePublicationRequest validates the complete successor publication
// request. PublishRequest.Validate retains the released rc11 association-only
// behavior for source compatibility.
func ValidatePublicationRequest(r AuthoritativePublishRequest) error {
	if err := r.ContentHash.Validate(); err != nil {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: %w", err)
	}
	if !r.VisibilityIntent.IsValid() {
		return fmt.Errorf("publish request validation failed at schema.ValidatePublicationRequest during publish-request validation: visibilityIntent %q is not omitted, private, or public; publish itself always keeps replacement private and never widens access; omit the field for a legacy caller or send one canonical intent", r.VisibilityIntent)
	}
	if _, err := NewSessionID(r.Identity.SessionID.String()); err != nil {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: identity.sessionId: %w", err)
	}
	if r.Identity.SchemaVersion <= 0 {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: identity.schemaVersion must be positive; the server cannot classify the producer schema; send the positive source schema version")
	}
	if !r.Model.Harness.IsKnown() || r.Model.Model == "" {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: model harness or model id is invalid; the server cannot classify the transcript producer; send canonical nonempty model identity")
	}
	if !r.Source.Format.IsValid() || r.Timestamp.Start <= 0 || r.Timestamp.End < r.Timestamp.Start {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: source format or timestamp range is invalid; the server cannot classify exact transcript metadata; send json/jsonl and positive ordered timestamps")
	}
	if err := r.Project.Hash.Validate(); err != nil || strings.TrimSpace(r.Project.Name) == "" {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: project identity is invalid; the server cannot bind the publication to a project; send a canonical hash and nonempty name")
	}
	if r.License != "" && !r.License.IsValid() {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: license %q is not canonical; the server cannot apply the requested grant; omit it or send a menu license", r.License)
	}
	legacyGit := GitContext{Branch: r.Git.Branch, Remote: r.Git.Remote, Worktree: r.Git.Worktree, Tracking: r.Git.Tracking, Associations: r.Git.Associations}
	for _, commit := range r.Git.Commits {
		legacyGit.Commits = append(legacyGit.Commits, CommitInfo(commit))
	}
	if err := legacyGit.Validate(); err != nil {
		return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: %w", err)
	}
	for i, entry := range r.Entries {
		if entry.SessionID != r.Identity.SessionID || !entry.Harness.IsKnown() || !entry.EntryType.IsValid() || !entry.Role.IsValid() || entry.EntryIndex < 0 || entry.Depth < 0 {
			return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: entries[%d] has inconsistent identity or classification; the server cannot canonicalize the complete replacement; use the root session id and canonical nonnegative entry fields", i)
		}
	}
	for i, subagent := range r.Subagents {
		if _, err := NewSessionID(subagent.SessionID.String()); err != nil || subagent.ParentUUID != r.Identity.SessionID {
			return fmt.Errorf("publish request validation failed at schema.PublishRequest.Validate during publish-request validation: subagents[%d] has invalid identity or parent binding; the server cannot canonicalize the replacement; use canonical session ids bound to the root session", i)
		}
	}
	return nil
}
