package schema

import (
	"fmt"
	"slices"

	"github.com/swaggest/jsonschema-go"
)

// ContentCapability identifies an optional transcript-content behavior that a
// client must negotiate before emitting evidence that older servers may lose.
type ContentCapability string

const (
	// ContentCapabilityObservedModel guarantees exact observedModel survival
	// through publish, typed migration, canonical rewrite, and re-emission. Its
	// meaning is immutable; an incompatible contract requires a new token.
	ContentCapabilityObservedModelV1 ContentCapability = "observed_model_v1"
)

// AllContentCapabilities is the canonical closed capability inventory.
var AllContentCapabilities = []ContentCapability{ContentCapabilityObservedModelV1}

// IsValid reports whether c belongs to the closed capability inventory.
func (c ContentCapability) IsValid() bool { return c == ContentCapabilityObservedModelV1 }

// Validate rejects capability identifiers outside the canonical inventory.
func (c ContentCapability) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("content capability validation failed at schema.ContentCapability.Validate while negotiating enriched transcript content: capability %q is outside the closed contract set, so the caller cannot determine preservation semantics; use one of %v from AllContentCapabilities", c, AllContentCapabilities)
	}
	return nil
}

// JSONSchema deliberately leaves discovery tokens forward-open. Go producers
// still use Validate and AllContentCapabilities as their closed inventory.
func (ContentCapability) JSONSchema() (jsonschema.Schema, error) {
	var s jsonschema.Schema
	s.AddType(jsonschema.String)
	s.WithTitle("Content Capability")
	s.WithDescription("Opaque, forward-open revision token in the deployment-specific content-capability set. Clients use exact set membership, never parse _v1 as Semantic Versioning or infer ranges, ignore unknown tokens, and tolerate duplicate or unordered input. Servers emit only their pinned known inventory, reject duplicates, and serialize lexicographically; token meanings are immutable. observed_model_v1 is required iff any root or nested assistant turn has observedModel, not for session model alone, and guarantees assistant-only value validation before persistence with no invalid DB/blob side effects plus byte-exact accepted observedModel strings through storage, typed migration, rewrite, serving, and pull; JSON whitespace and key order are excluded.")
	return s, nil
}

// KnownContentCapabilities filters unknown discovery tokens, removes duplicates,
// and returns known tokens in canonical lexicographic order.
func KnownContentCapabilities(advertised []ContentCapability) []ContentCapability {
	known := make(map[ContentCapability]struct{}, len(advertised))
	for _, capability := range advertised {
		if capability.IsValid() {
			known[capability] = struct{}{}
		}
	}
	result := make([]ContentCapability, 0, len(known))
	for capability := range known {
		result = append(result, capability)
	}
	slices.Sort(result)
	return result
}

// ValidateContentCapabilityAdvertisements validates a server-produced list.
// Discovery readers stay forward-open, but producers must emit only known,
// unique tokens in canonical lexicographic order.
func ValidateContentCapabilityAdvertisements(advertised []ContentCapability) error {
	for i, capability := range advertised {
		if err := capability.Validate(); err != nil {
			return fmt.Errorf("content capability advertisement validation failed at schema.ValidateContentCapabilityAdvertisements while validating item %d: the server emitted an unknown token and clients cannot rely on its semantics; update the schema inventory before advertising it: %w", i, err)
		}
		if i > 0 && advertised[i-1] >= capability {
			return fmt.Errorf("content capability advertisement validation failed at schema.ValidateContentCapabilityAdvertisements while checking item %d: advertisements are duplicated or not in canonical lexicographic order, so server output is nondeterministic; deduplicate and sort with KnownContentCapabilities", i)
		}
	}
	return nil
}

// MissingContentCapabilities returns required tokens absent from discovery.
// Unknown discovery tokens are ignored and omitted and empty lists are equal.
func MissingContentCapabilities(advertised, required []ContentCapability) []ContentCapability {
	available := KnownContentCapabilities(advertised)
	missing := make([]ContentCapability, 0, len(required))
	for _, capability := range KnownContentCapabilities(required) {
		if !slices.Contains(available, capability) {
			missing = append(missing, capability)
		}
	}
	return missing
}

// RequiredContentCapabilities scans root and nested turns for optional evidence.
// The session-level Model seed is legacy metadata and does not require a
// capability. An observedModel on any assistant turn, including a nested
// subagent turn, requires observed_model_v1.
func RequiredContentCapabilities(payload SessionDetailPayload) []ContentCapability {
	for _, turn := range payload.Turns {
		if turn.ObservedModel != "" {
			return []ContentCapability{ContentCapabilityObservedModelV1}
		}
	}
	return nil
}

// ValidateObservedModelEvidence enforces the enriched-publish attribution rule.
// Subagent output uses the assistant role at a nonzero depth, so RoleAssistant
// covers both root assistants and subagents. Omitted evidence is valid. Servers
// advertising observed_model_v1 guarantee this validation completes before any
// persistence or other side effect and that accepted observedModel value bytes
// survive publish, storage, migration, rewrite, and serving exactly. JSON
// whitespace, object-key ordering, and other formatting are not guaranteed.
func ValidateObservedModelEvidence(role Role, observed ObservedModelID) error {
	if observed == "" {
		return nil
	}
	if _, err := NewObservedModelID(observed.String()); err != nil {
		return fmt.Errorf("observed-model evidence validation failed at schema.ValidateObservedModelEvidence during enriched publish: the observedModel value is invalid, so the server must reject the publish rather than persist ambiguous evidence; fix or omit the source observation: %w", err)
	}
	if role != RoleAssistant {
		return fmt.Errorf("observed-model evidence validation failed at schema.ValidateObservedModelEvidence during enriched publish: role %q is not assistant, so observedModel cannot describe assistant or subagent output and the server must reject the publish; omit observedModel from user, system, and tool turns", role)
	}
	return nil
}
