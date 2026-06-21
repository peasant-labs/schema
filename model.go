package schema

// ModelInfo holds harness and model version information for a session.
// JSON tags use camelCase to match CLI UnifiedMetadata wire format. The harness
// is keyed json:"harness" (SLICE-B1 emit-side flip; was json:"modelHarness").
//
// rc2 (#118): Harness and Model carry required:"true" so swaggest emits a
// SchemaModelInfo.required:["harness","model"] array — the village rejects a model
// object missing either key (the harness/model-within-model contract). Metadata
// only: these tags change the GENERATED schema's `required`, not the Go wire shape.
type ModelInfo struct {
	Harness        Harness  `json:"harness" required:"true"`
	Model          ModelID  `json:"model" required:"true"`
	HarnessVersion string   `json:"version,omitempty"`
	HostSlug       HostSlug `json:"hostSlug,omitempty"`
}
