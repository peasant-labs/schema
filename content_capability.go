package schema

import (
	"fmt"

	"github.com/swaggest/jsonschema-go"
)

// ContentCapability identifies an optional transcript-content behavior that a
// client must negotiate before emitting evidence that older servers may lose.
type ContentCapability string

const (
	// ContentCapabilityObservedModel guarantees exact observedModel survival
	// through publish, typed migration, canonical rewrite, and re-emission.
	ContentCapabilityObservedModel ContentCapability = "observed_model"
)

// AllContentCapabilities is the canonical closed capability inventory.
var AllContentCapabilities = []ContentCapability{ContentCapabilityObservedModel}

// IsValid reports whether c belongs to the closed capability inventory.
func (c ContentCapability) IsValid() bool { return c == ContentCapabilityObservedModel }

// Validate rejects capability identifiers outside the canonical inventory.
func (c ContentCapability) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("content capability validation failed at schema.ContentCapability.Validate while negotiating enriched transcript content: capability %q is outside the closed contract set, so the caller cannot determine preservation semantics; use one of %v from AllContentCapabilities", c, AllContentCapabilities)
	}
	return nil
}

// JSONSchema declares the closed capability inventory to generated clients.
func (ContentCapability) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Content Capability", "Optional enriched transcript-content behavior that a client must negotiate before emission", AllContentCapabilities), nil
}

// ContentCapabilityVersion identifies one semantic version of a capability.
// It is independent from the broader push-envelope contract version.
type ContentCapabilityVersion string

const (
	// ContentCapabilityVersionObservedModelV1 is the first observed-model
	// preservation and publish-acceptance contract.
	ContentCapabilityVersionObservedModelV1 ContentCapabilityVersion = "1.0.0"
)

// AllContentCapabilityVersions is the canonical closed capability-version inventory.
var AllContentCapabilityVersions = []ContentCapabilityVersion{ContentCapabilityVersionObservedModelV1}

// IsValid reports whether v belongs to the closed capability-version inventory.
func (v ContentCapabilityVersion) IsValid() bool {
	return v == ContentCapabilityVersionObservedModelV1
}

// Validate rejects capability versions outside the canonical inventory.
func (v ContentCapabilityVersion) Validate() error {
	if !v.IsValid() {
		return fmt.Errorf("content capability version validation failed at schema.ContentCapabilityVersion.Validate while negotiating enriched transcript content: version %q is outside the closed contract set, so the caller cannot determine preservation semantics; use one of %v from AllContentCapabilityVersions", v, AllContentCapabilityVersions)
	}
	return nil
}

// JSONSchema declares the closed capability-version inventory to generated clients.
func (ContentCapabilityVersion) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Content Capability Version", "Semantic version of one optional enriched transcript-content behavior", AllContentCapabilityVersions), nil
}

// ContentCapabilityAdvertisement declares one server-supported enriched-content contract.
type ContentCapabilityAdvertisement struct {
	Capability ContentCapability        `json:"capability"`
	Version    ContentCapabilityVersion `json:"version"`
}

// Validate rejects unknown values and capability/version combinations.
func (a ContentCapabilityAdvertisement) Validate() error {
	if err := a.Capability.Validate(); err != nil {
		return fmt.Errorf("content capability advertisement validation failed at schema.ContentCapabilityAdvertisement.Validate while validating capability: %w", err)
	}
	if err := a.Version.Validate(); err != nil {
		return fmt.Errorf("content capability advertisement validation failed at schema.ContentCapabilityAdvertisement.Validate while validating version: %w", err)
	}
	if a.Capability != ContentCapabilityObservedModel || a.Version != ContentCapabilityVersionObservedModelV1 {
		return fmt.Errorf("content capability advertisement validation failed at schema.ContentCapabilityAdvertisement.Validate while matching capability to version: capability %q does not define version %q, so a client cannot rely on the advertised semantics; advertise observed_model with version 1.0.0", a.Capability, a.Version)
	}
	return nil
}

// ValidateObservedModelEvidence enforces the enriched-publish attribution rule.
// Subagent output uses the assistant role at a nonzero depth, so RoleAssistant
// covers both root assistants and subagents. Omitted evidence is valid.
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
