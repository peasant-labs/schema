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
