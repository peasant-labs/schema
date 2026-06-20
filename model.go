package schema

// ModelInfo holds harness and model version information for a session.
// JSON tags use camelCase to match CLI UnifiedMetadata wire format. The harness
// is keyed json:"harness" (SLICE-B1 emit-side flip; was json:"modelHarness").
type ModelInfo struct {
	Harness        Harness  `json:"harness"`
	Model          ModelID  `json:"model"`
	HarnessVersion string   `json:"version,omitempty"`
	HostSlug       HostSlug `json:"hostSlug,omitempty"`
}
