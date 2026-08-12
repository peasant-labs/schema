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
	// operation-specific publish request + SchemaModelInfo). Bumped to 0.4.0 when
	// PublishRequest.License + PullTranscriptInfo.License were added (additive
	// optional field = minor bump). Bumped to 0.5.0 when the pull skip-gate surface
	// (POST /api/v1/pull/transcripts/skip-gate) was added (additive = minor bump).
	// Bumped to 0.6.0 when shared operation components adopted exact canonical
	// Types schemas and the stricter publish HTTP body gained a distinct component
	// identity instead of shadowing the canonical PublishRequest.
	// Bumped to 0.7.0 when TargetKind gained the file_version and association
	// members and AnnotationSummary gained TargetFilePath/TargetContentHash and
	// TargetAssociationID;
	// harmonizeSharedTypeComponents propagates both into this spec's embedded
	// annotation components (additive = minor bump).
	// Bumped to 0.8.0 when the publish wire gained owner-scoped published
	// associations and annotation ingress gained targetAssociationId plus its
	// documented Village POST operation (additive = minor bump).
	// Bumped to 0.9.0 when Harness gained Strike and SessionID gained the two
	// observed Strike identifier forms (additive = minor bump).
	// Bumped to 0.10.0 when the owner transcript update operation
	// (PATCH /api/v1/transcripts/{id}) was declared: an operation the village
	// already served but the published contract never stated, together with its
	// tri-state license semantics, its narrowed visibility menu, and its full
	// reachable refusal set (additive = minor bump).
	// Each prior version's generated goldens are retained byte-frozen under the
	// retired-spec immutability guard.
	// Bumped to 0.11.0 when publication gained canonical operation currency,
	// authoritative success receipts, real multipart transport, and a complete
	// successor owner-update projection (additive = minor bump).
	// Bumped to 0.12.0 when authoritative publication identity adopted its
	// declared parentSessionId key while historical SessionIdentity retained
	// parentUuid (corrective additive version).
	VillageAPIVersion = "0.12.0"
	// PeasantLocalAPIVersion is the info.version of the local dashboard API spec.
	// Bumped to 0.2.0 when the Map/Review/Search surface was added (8 additive ops
	// + FrictionCluster = minor bump). Bumped to 0.3.0 when the project Git
	// timeline gained normalized session identities and authoritative many-to-many
	// commit bindings. Bumped to 0.4.0 when file-change statuses and diff-line
	// kinds became named closed sets while retaining their existing JSON tokens.
	// Bumped to 0.5.0 when the git+session timeline and insight-first code map
	// surface landed: durable atomic session-to-commit associations, ghost/rewrite
	// mapping, the mechanical insight envelope, node-grain
	// read-state/comprehension signals, and TaskSummary.ReadFiles (all
	// additive).
	// Bumped to 0.6.0 when Harness gained Strike and SessionID gained the two
	// observed Strike identifier forms (additive = minor bump).
	// Bumped to 0.7.0 when TurnDetail gained optional exact model observation
	// evidence for assistant-generated turns (additive = minor bump).
	// Prior versions stay byte-frozen.
	PeasantLocalAPIVersion = "0.7.0"
	// TypesVersion is the info.version of the types spec (the foundational shared
	// domain types catalog; formerly "shared-types").
	// Bumped to 0.2.0 when the catalog became the comprehensive canonical
	// cross-language contract surface instead of a small set of seed types.
	// Bumped to 0.3.0 when the map/review diff string fields became named closed
	// sets with generated runtime inventories and predicates.
	// Bumped to 0.4.0 when the Local API 0.5.0 surface's new catalog types
	// (SessionAssociation, RewrittenCommit, SessionInsight and their closed
	// sets) and the widened TargetKind landed (additive = minor bump).
	// Bumped to 0.5.0 when GitContext gained PublishedAssociation and annotation
	// push items gained TargetAssociationID (additive = minor bump).
	// Bumped to 0.6.0 when Harness gained Strike and SessionID gained the two
	// observed Strike identifier forms (additive = minor bump).
	// Bumped to 0.7.0 when the owner transcript update request, its narrowed
	// visibility menu, and its three-valued license entered the catalog so both
	// language bindings carry them (additive = minor bump). The operation's
	// refusal envelope is deliberately NOT catalogued: it is operation-scoped so
	// a release cannot freeze a transcript-update-specific name onto a generic
	// error shape.
	// Bumped to 0.8.0 for canonical publication operation, authoritative receipt,
	// publish visibility intent, and successor owner-update types.
	// Bumped to 0.9.0 when AuthoritativeSessionIdentity adopted parentSessionId
	// independently from the historical SessionIdentity parentUuid wire.
	// Bumped to 0.10.0 when TurnDetail gained optional exact model observation
	// evidence for assistant-generated turns (additive = minor bump).
	TypesVersion = "0.10.0"
)
