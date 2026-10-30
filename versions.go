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
	// Bumped to 0.13.0 when schema negotiation gained optional enriched-content
	// capability advertisements (additive = minor bump).
	// Bumped to 0.14.0 when the Village API spec gained the schema-owned
	// collectives, contribution, review, and linked-repository web routes.
	// Bumped to 0.15.0 when the canonical harness inventory gained Pi.
	// Bumped to 0.16.0 when the pull request prompt attachment surface was
	// declared: the GitHub webhook receiver, the attachment read, confirm, and
	// detach routes, the prompt-request and user-settings routes, and two
	// collective settings (additive = minor bump).
	// Bumped to 0.17.0 for grouped transcript listings, helper-member reads,
	// durable provenance, and authorized relationship navigation.
	// Bumped to 0.18.0 when the prompt digest's commit items gained the
	// optional per-commit change counts, which harmonizeSharedTypeComponents
	// propagates into this spec's embedded digest item component (additive =
	// minor bump).
	// Bumped to 0.19.0 when a prompt digest skill item gained the optional args
	// typed after the command, propagated the same way (additive = minor bump).
	// Bumped to 0.20.0 when the group repository surface gained the available-
	// repositories read that feeds the repository picker (additive operation =
	// minor bump).
	// Bumped to 0.21.0 when the prompt-request list gained the pull request's
	// head remote beside its base, so a push from a clone of a fork matches the
	// request waiting on that fork's author (additive field = minor bump).
	// Bumped for retained unknown source evidence and preservation negotiation.
	// Bumped to 0.23.0 for the transcript pull request read, the caller's
	// totals, the automatic pull request linking setting, pull request titles,
	// head branches, and list summaries, the collective and repository counts,
	// and a create request that accepts a null organization (additive = minor
	// bump).
	// Bumped to 0.25.0 for the optional thinking level and its raw native
	// spelling on session metadata, session detail, and assistant turns, the
	// in-session thinking-level history, plus the thinking_level_v1 content
	// capability (additive = minor bump).
	// Bumped to 0.26.0 when the transcript pull list gained the optional
	// collectives filter, which narrows the list to the approved shares of the
	// listed collectives the caller belongs to (additive parameter = minor bump).
	// Canonical collective routes retain deprecated groups for a full release window.
	// Bumped to 0.28.0 when the grouped transcript metadata stopped declaring
	// diagnostics_warnings as bare strings: publication carries structured
	// DiagnosticEntry objects, so the declared read shape could never describe a
	// stored row (corrective type change = minor bump on a 0.x contract).
	// Bumped to 0.29.0 when every response that shows a person to another
	// viewer adopted VillagePublicOwner (a profile or anonymous, resolved by
	// the server per viewer) in place of the whole VillageUser, owner ids,
	// usernames, avatars, and discoverable flags; the pull envelope moved to
	// pull contract 0.3.0 with it, attesters, pull request authors and
	// requesters, collective creators, and repository linkers adopted it too,
	// and transcript rows gained the server-resolved parent_transcript_id
	// and made their repository fields nullable, withheld beside an anonymous
	// owner (breaking change on a 0.x contract = minor bump).
	// Bumped to 0.29.1 when the pull operations accepted the optional
	// Peasant-Pull-Contract request header, reported the served envelope in the
	// same response header, declared the frozen 0.2.0 Legacy* envelope beside
	// the 0.3.0 one, and the version response gained pullContractMaxVersion
	// (additive = patch bump; every 0.29.0 client reads it unchanged).
	// Bumped to 0.30.0 when collective invitations replaced the owner's add by
	// username: the add-member operation was removed (breaking change on a 0.x
	// contract = minor bump), and the invitation operations, the caller's
	// notification email operations, the one-click unsubscribe operation, and
	// the notification email and invitation-email preference on the caller's
	// settings were added.
	// Bumped to 0.31.0 for the optional unverified pull request read projection.
	// Bumped to 0.32.0 for the owner integrity-check operation, the optional
	// integrity field on the transcript metadata read, the Village-Content-Hash
	// header on the content read, the stored publish metadata pull read, and the
	// optional publishing client on publish metadata (additive = minor bump).
	// Bumped to 0.33.0 when the owner integrity-check operation was removed and
	// integrity became servable on a transcript of any visibility (an operation
	// removed; minor bump before 1.0).
	// Bumped to 0.34.0 when AuthoritativeModelInfo keyed its harness version
	// harnessVersion, the key the publish decoder accepts, instead of version,
	// which the decoder refused (minor bump before 1.0).
	VillageAPIVersion = "0.34.0"
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
	// Bumped to 0.8.0 when process-local UI capability discovery was added
	// (additive = minor bump).
	// Bumped to 0.9.0 when the session payloads gained the producer's declared
	// session origin and the by-id session-summaries operation was declared
	// (additive = minor bump).
	// Prior versions stay byte-frozen.
	// Bumped to 0.10.0 for detailed usage and native metadata evidence.
	// Bumped to 0.11.0 when TurnDetail gained the optional CommandInvocation
	// for skill and user slash-command turns (additive = minor bump).
	// Bumped to 0.12.0 for independent optional tool namespace evidence.
	// Bumped to 0.13.0 for grouped session listings, helper-member reads, and
	// the flat detail read projection with relationship navigation.
	// Bumped for retained unknown source evidence and durable diagnostics.
	// Bumped to 0.15.0 for publishing from the transcript: the publication
	// read, the typed collectives push, the collectives proxy, the declared
	// sign-in and redaction preview routes with sign-out, the settings routes,
	// the typed sync chooser status, and the declared cross-origin refusal on
	// write routes (additive = minor bump).
	// The Local API embeds Village catalog rows: VillageUserGroup in the
	// collectives proxy, VillagePromptRequest in push results, and
	// VillageShareStatus in the publication audience, with their closed sets.
	// A change to any of them also bumps this version;
	// testdata/local-api/village_components.yaml pins them per Local version.
	// Bumped to 0.16.0 for rule-required hooks that stop publishing when their
	// last active binding is removed, retaining the existing response fields.
	// Bumped to 0.17.0 for an optional stored-session target on rule creation.
	// Bumped to 0.18.0 when the session summary gained the optional
	// parentUnresolved marker and the sync row gained the optional
	// parentSessionId and parentUnresolved fields (additive = minor bump).
	// Bumped to 0.19.0 for the optional thinking level and raw native spelling
	// on session detail and assistant turns, and the in-session thinking-level
	// history (additive = minor bump).
	// Bumped to 0.20.0 for forward sync pages and metadata-only project options.
	// Bumped to 0.21.0 for independent account-scoped collective observations
	// and bounded contribution evidence, retaining the forward sync views.
	// Bumped to 0.22.0 when the embedded VillageUserCollective replaced its
	// created_by account id with the creator as a VillagePublicOwner.
	// Bumped to 0.23.0 when the session reads gained the recorded commits and
	// the pull requests detected in the session's own gh pr create calls
	// (additive operations = minor bump).
	// Bumped to 0.24.0 when the local session summaries and the flat detail
	// read gained the optional Claude Code teammate identity, with its own
	// teammate evidence set, a local read addition that never reaches the
	// publish wire (additive = minor bump).

	// Bumped to 0.25.0 when the stored-browsing reads stopped being scoped by
	// the saved ingestion selection: the home page, its child reads and its
	// project options describe the stored library, a page row may carry
	// outsideSelection while remaining visible, and a by-id or publication
	// read states the same rule. A saved selection scopes future ingest and
	// hook publication only, never a stored read.
	// Bumped to 0.26.0 for the optional unverified pull request read projection.
	PeasantLocalAPIVersion = "0.26.0"
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
	// Bumped to 0.11.0 for the content-capability contract catalog.
	// Bumped to 0.12.0 for the open process-local UI capabilities envelope.
	// Bumped to 0.13.0 when the declared session-origin closed set entered the
	// catalog and both session payloads gained its optional field (additive =
	// minor bump).
	// Bumped to 0.14.0 when the Village collectives, contribution, review, and
	// linked-repository DTOs entered the catalog.
	// Bumped to 0.15.0 for optional local metadata adapter/parser provenance.
	// Bumped to 0.16.0 for detailed usage and native metadata evidence while
	// retaining the adapter/parser provenance contract.
	// Bumped to 0.17.0 when CommandInvocation, the PromptDigest projection with
	// its DigestItemKind closed set, and the Village pull request attachment
	// DTOs entered the catalog.
	// Bumped to 0.18.0 for independent optional tool namespace evidence.
	// Bumped to 0.19.0 for durable session graph provenance, independent input
	// submission counts, flat read navigation, and grouped list DTOs.
	// Bumped to 0.20.0 when PromptDigestItem gained the optional additions,
	// deletions, and filesChanged counts a commit anchor carries (additive =
	// minor bump).
	// Bumped to 0.21.0 when a prompt digest skill item gained the optional args
	// typed after the command (additive = minor bump).
	// Bumped to 0.22.0 when the group available-repositories read added its
	// response types to the catalog (additive = minor bump).
	// Bumped to 0.23.0 when the catalog's prompt request gained the pull
	// request's head remote beside its base (additive = minor bump).
	// Bumped for retained unknown source evidence and durable diagnostics.
	// Bumped to 0.25.0 when the local publishing, sign-in, and settings types
	// and the Village pull request read, summary, and totals types entered the
	// catalog (additive = minor bump).
	// Bumped to 0.27.0 to describe rule-required hook consent in the local
	// publication indicator, without changing any field or enum.
	// Bumped to 0.28.0 for the mutually exclusive session-target rule request.
	// Bumped to 0.29.0 when SessionSummary gained the optional parentUnresolved
	// marker and LocalSyncSummary gained the optional parentSessionId and
	// parentUnresolved fields (additive = minor bump).
	// Bumped to 0.30.0 for the ThinkingLevel closed set, the bounded
	// ThinkingLevelRaw native spelling, and the ThinkingLevelChange history
	// entry (additive = minor bump).
	// Bumped to 0.31.0 for concrete sync pages, tokens, and project options.
	// Bumped to 0.32.0 for scoped collective observations, retained contribution
	// evidence and deprecated collective envelopes alongside shared sync pages.
	// Bumped to 0.33.0 when the catalog's Village transcript metadata adopted
	// structured DiagnosticEntry diagnostics warnings instead of bare strings,
	// matching what publication carries and Village stores.
	// Bumped to 0.34.0 when VillagePublicOwner and its kind entered the
	// catalog and the catalog's Village rows and pull envelope adopted it in
	// place of owner ids, usernames, avatars, and the whole VillageUser, and
	// the catalog's Village transcript rows gained parent_transcript_id.
	// Bumped to 0.34.1 when the frozen 0.2.0 pull envelope types
	// (LegacyPullTranscriptInfo, LegacyPullListResponse, LegacyPullAnnotation)
	// entered the catalog and SchemaVersionResponse gained
	// pullContractMaxVersion (additive = patch bump).
	// Bumped to 0.35.0 when the catalog gained the collective invitation and
	// notification email types and VillageUserSettings gained
	// notification_email and email_invitations. The add-member username
	// request stays in the catalog, deprecated, for the alias window.
	// Bumped to 0.36.0 when the session commits read and the detected session
	// pull request rows entered the catalog (additive = minor bump).
	// Bumped to 0.37.0 when SessionTeammate and its own SessionTeammateEvidence
	// closed set entered the catalog (additive = minor bump). The session-graph
	// EvidenceKind set is unchanged: teammate evidence is local read evidence
	// and must not widen the shared published menu.

	// Bumped to 0.38.0 to describe the local publication's selection fields as
	// ingest and push membership rather than browsing permission, without
	// changing any field or enum.
	// Bumped to 0.39.0 for unverified pull request evidence and exact source locators.
	// Bumped to 0.40.0 for the scry check request, response and health body, the
	// transcript integrity types, and AuthoritativeClientInfo (additive = minor
	// bump).
	// Bumped to 0.41.0 when VillageIntegrityCheckResult and
	// VillageIntegrityCheckResponse were removed with the owner integrity check
	// (minor bump before 1.0).
	// Bumped to 0.42.0 when AuthoritativeModelInfo keyed its harness version
	// harnessVersion instead of version (minor bump before 1.0).
	TypesVersion = "0.42.0"
)
