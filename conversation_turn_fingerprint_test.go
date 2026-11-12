package schema_test

import (
	"testing"

	"github.com/peasant-labs/schema"
)

// TestConversationTurnFingerprint pins that the publication fingerprint
// encodes conversationTurnCount only when present: absent leaves the
// fingerprint unchanged, present changes it.
func TestConversationTurnFingerprint(t *testing.T) {
	fingerprint := func(request schema.AuthoritativePublishRequest) string {
		t.Helper()
		operation, err := schema.CanonicalizePublishRequest(request)
		if err != nil {
			t.Fatal(err)
		}
		digest, err := schema.FingerprintPublishOperation(operation)
		if err != nil {
			t.Fatal(err)
		}
		return digest.String()
	}
	absentLeft := validPublishRequest()
	absentRight := validPublishRequest()
	if fingerprint(absentLeft) != fingerprint(absentRight) {
		t.Fatalf("absent conversationTurnCount changes the fingerprint")
	}
	present := validPublishRequest()
	zero := int64(0)
	present.Stats.ConversationTurnCount = &zero
	if fingerprint(absentLeft) == fingerprint(present) {
		t.Fatalf("present conversationTurnCount leaves the fingerprint unchanged")
	}
	other := validPublishRequest()
	other.Stats.ConversationTurnCount = &zero
	if fingerprint(present) != fingerprint(other) {
		t.Fatalf("equal conversationTurnCount values fingerprint differently")
	}
}
