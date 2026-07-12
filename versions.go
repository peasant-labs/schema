package schema

// Versioned-spec info.version values. These are the SINGLE SOURCE OF TRUTH for
// each versioned OpenAPI spec's semantic version: they feed WithVersion(...) in
// the openapi.Build*Spec functions (re-exported there as openapi.VillageAPIVersion
// etc.), the generated artifact filenames, AND the version-aware accessors in this
// package (e.g. VillageAPISpecJSON). A version bump is a one-line edit here.
//
// They live in the contract root (not the openapi sub-package) so the embedded-spec
// accessors below can key off them without an import cycle (openapi imports this
// package, so this package cannot import openapi).
//
// Bump the relevant one when its API surface changes (minor for additive surface,
// major for breaking changes). Retired versions stay byte-frozen under the
// retired-spec immutability guard; new surface goes on a new version.
const (
	// VillageAPIVersion is the info.version of the Village API spec. Bumped to
	// 0.2.0 when the /api/v1/pull surface was added (additive = minor bump).
	// Bumped to 0.3.0 (rc2 #118) when model + model.harness/model.model became
	// required — a tightening of the publish contract (selective required arrays on
	// SchemaPublishRequest + SchemaModelInfo). Bumped to 0.4.0 when
	// PublishRequest.License + PullTranscriptInfo.License were added (additive
	// optional field = minor bump). Bumped to 0.5.0 when the pull skip-gate surface
	// (POST /api/v1/pull/transcripts/skip-gate) was added (additive = minor bump).
	// Each prior version's generated goldens are retained byte-frozen under the
	// retired-spec immutability guard.
	VillageAPIVersion = "0.5.0"
	// PeasantLocalAPIVersion is the info.version of the local dashboard API spec.
	// Bumped to 0.2.0 when the Map/Review/Search surface was added (8 additive ops
	// + FrictionCluster = minor bump). The retired 0.1.0 spec is retained byte-frozen
	// (no longer emitted) under the retired-versions immutability guard.
	PeasantLocalAPIVersion = "0.2.0"
	// TypesVersion is the info.version of the types spec (the foundational shared
	// domain types catalog; formerly "shared-types").
	TypesVersion = "0.1.0"
)
