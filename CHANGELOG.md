# Changelog

All notable changes to the `github.com/peasant-labs/schema` contract module are
documented here. This project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Local API 0.7.0 and Types 0.10.0 add optional `TurnDetail.observedModel`
  evidence as the typed `ObservedModelID`. Present values are non-empty valid
  UTF-8 strings with no leading or trailing Unicode whitespace; accepted bytes,
  including Unicode, mixed case, slashes, and internal spaces, are preserved
  exactly.
  Generated OpenAPI, TypeScript, and Zod contracts enforce those shape and value
  constraints. Producers remain
  responsible for attaching observations only to assistant or subagent output
  because generated validators do not express a role-dependent condition.
- `SessionDetailPayload.model` is documented as the initial root-assistant seed
  for sticky model resolution: the earliest valid root observation in canonical
  order when available, otherwise legacy stored metadata. It is not the latest,
  most common, or final model.

## [v0.1.0] - 2026-08-02

First stable release of the `github.com/peasant-labs/schema` Go module and the
`@peasant-labs/schema` npm package. The contract artifacts are unchanged from
`v0.1.0-rc13`: Village API 0.12.0, PublishRequest 0.12.0,
AnnotationPushRequest 0.12.0, Local API 0.6.0, and Types 0.9.0.

### Changed

- The shared release guard now supports an exact, explicitly configured
  initial-final policy for repositories with no prior product release. Schema
  does not opt into that policy; this final follows the ordinary green,
  same-version ancestor release candidate path through `v0.1.0-rc13`.

## [v0.1.0-rc13] - 2026-08-01

### Fixed

- Village API 0.12.0 and Types 0.9.0 give authoritative publication identity
  its declared `parentSessionId` wire key. Go decoding, canonical operation
  fingerprints, generated OpenAPI, and generated Zod validators now agree while
  historical `SessionIdentity` surfaces retain their `parentUuid` key.

## [v0.1.0-rc12] - 2026-07-31

### Added

- Village API 0.11.0 and Types 0.8.0 declare the successor publication
  contract. Publish metadata carries the SHA3-256 digest of the exact transcript
  bytes and an optional final-visibility intent. The visibility intent never
  widens access during content replacement; widening remains a separate owner
  update.
- Successful publication returns a complete authoritative receipt for both
  creation (201) and replacement (200). The receipt identifies the transcript
  and canonical URL, records the actual visibility and content hash, echoes the
  canonical operation fingerprint, describes the applied license and
  associations, reports normalized values and blob facts, and distinguishes a
  newly created transcript from a replacement.
- Canonical publish operations have a static, domain-separated, length-framed
  fingerprint. The fingerprint covers replacement content, license intent,
  association intent, and the exact transcript content hash without depending
  on JSON property order or serializer behavior.
- The successor owner-update operation declares omission, clear, and replacement
  semantics independently and returns the complete editable transcript state on
  success.

### Changed

- Successor publication request, operation, receipt, and owner-update objects are
  closed recursively across Go decoding, generated OpenAPI, and built Zod
  validators. Unknown and duplicate properties are rejected instead of being
  ignored or stripped, while released Village 0.10.0 and Types 0.7.0 artifacts
  remain byte-frozen.
- `openapi.TranscriptPublishRequest` remains available as a deprecated alias to
  `openapi.AuthoritativeTranscriptPublishRequest`. The successor's nested
  publication projections are distinct Go types so their generated validators
  can be strict without closing legacy shared metadata components. Consumers
  that construct nested values explicitly must convert to the successor types;
  removing the compatibility alias is deferred until consumer migration is
  complete.
- The generated TypeScript test suite runs serially because its idempotence test
  intentionally regenerates a tracked contract artifact that sibling tests also
  inspect.

## [v0.1.0-rc11] - 2026-07-30

### Added

- Village API 0.10.0 and Types 0.7.0 declare the owner transcript update
  operation, `PATCH /api/v1/transcripts/{id}`. The village has served this route
  since transcript governance landed, but the published contract never stated
  it, so a client could only call it from knowledge of the handler. Declaring it
  closes that drift.

  `TranscriptUpdateRequest` carries `title`, `description`, `visibility`, and
  `license`, all optional, where an omitted field means leave unchanged rather
  than reset. `license` is three-valued: omitted preserves the stored license,
  the empty string requests a clear, and a canonical menu license replaces.
  Those three states stay distinguishable on the wire because that distinction
  is what makes the server's rule expressible that a granted Creative Commons
  license can never be cleared, only replaced.

  `TranscriptUpdateVisibility` declares `private` and `public` only. It is
  deliberately narrower than `Visibility`, whose third member covers
  organization-scoped access: that capability is deferred, and declaring it on
  this operation would advertise a value the server refuses. `Visibility` itself
  is unchanged.

  The operation declares its reachable refusals, not only success, because two
  of them are contract rules rather than transport accidents: only the owner may
  call it, and anyone else receives 403 while neither the transcript nor its
  governance audit changes; and clearing a granted license is refused with 400.
  All refusals share one `TranscriptUpdateErrorResponse` envelope, whose `error`
  field is required: the village writes every refusal through a single helper
  that always sets it, including the 401 raised before the handler runs, so an
  optional declaration would understate the guarantee. 400 covers
  five distinct refusals: an unparseable transcript id, an undecodable body, a
  visibility outside the accepted set, a license outside the canonical menu, and
  the attempt to clear a granted license. 401 comes from the authentication
  boundary that runs before the handler and is distinct in remedy from 403:
  401 means re-authenticate, 403 means you are not the owner. 404 covers a
  failed lookup as well as a genuinely absent transcript, so it is not proof of
  absence. No prior operation in this module
  declared a 4xx or 5xx response, so this establishes that pattern for one
  operation only; it is not a shared error framework and nothing else is
  expected to adopt it without its own consumer driving the change.

  The 200 status is declared with **no body schema**, and that is deliberate:
  it does not mean the operation returns no body. The village does return one,
  but it currently serves an untyped object wrapping the stored row's internal
  columns (`owner_id`, `blob_key`, `project_hash`, `source_file_path` and
  others) at `backend/internal/handler/transcripts.go:723-727`. Those columns
  also serialize through pgx `pgtype` wrappers, so a consumer would receive
  `{"String":"x","Valid":true}` where it expects a string, which makes the
  served shape undecodable as a typed contract rather than merely leaky.
  Declaring a projection the village does not actually serve would break the
  property that the served contract and the declared contract cannot drift, so
  nothing is declared until the handler serves a shape worth declaring. Nothing
  consumes it today either: applied state is read back through
  `GET /api/v1/pull/transcripts/{id}`. Tracked, together with the separate
  defect that `tags` is decoded and silently dropped, at
  https://github.com/peasant-labs/village/issues/55. Adding the response schema
  once that lands is additive.

  Two behaviours the village has and this operation deliberately does not
  declare, recorded so they are not later mistaken for drift. The village
  accepts and stores a legacy `shared` visibility, which is not a member of this
  contract's `Visibility` enum at all (its third member is `group`, which the
  village refuses); declaring it would mean inventing an enum member to expose
  the deferred organization-ACL capability. And `uuid.Parse` accepts four
  identifier spellings the declared pattern rejects - uppercase, brace-wrapped,
  `urn:uuid`-prefixed, and 32 undashed hex digits - which stay undeclared
  because the village only ever emits the canonical lowercase form and this
  module already rejects the others at its own boundary.

  The path parameter is the canonical `TranscriptID` rather than a bare string.
  The village parses it with `uuid.Parse` and refuses anything else with a 400,
  so an unconstrained string described ids the server never accepts. The
  canonical type carries `format: uuid` and the lowercase-hex pattern this
  module already treats as canonical for transcript identifiers (the same shape
  `SessionID` accepts on its UUID branch, and the form `NewTranscriptID`
  enforces at the Go boundary), so this adopts the module's existing position
  rather than inventing a restriction.

  `title` declares `maxLength: 500`, matching the storage column. Nothing
  validated it on the way in, so an over-long title reached the database and
  surfaced as an opaque server error rather than a refusal naming the limit; a
  client can now catch it before sending. The two bounds count differently and
  the difference is deliberate rather than overlooked: the column counts code
  points while a JavaScript validator generated from this bound counts UTF-16
  code units, so nothing the contract accepts can be rejected by the column,
  at the cost of refusing a title of more than 250 astral characters that the
  column would have taken.

  The request body is declared **required**. Reflection marks a body optional by
  default, which would have described a request that always fails: the handler
  decodes unconditionally, so an absent or empty body is a guaranteed 400. An
  empty JSON object is the correct no-op and is accepted.

  `TranscriptUpdateErrorResponse` is deliberately operation-scoped rather than a
  member of the shared type catalog. Its shape is nothing but `{error: string}`,
  so cataloguing it would freeze a transcript-update-specific name onto a
  generic envelope at release time, leaving whoever declares the next
  operation's refusals to reuse a misleading name, duplicate it, or take a
  breaking rename. Whether a shared refusal envelope belongs in the catalog is a
  decision for the change that needs one.

  The request body is a closed object: unknown properties are refused rather
  than accepted and silently discarded, and no property admits JSON null.
  Refusing null is deliberate. The village decodes a null into the same nil
  pointer an omitted field produces, so null would mean *preserve* while a
  caller sending it almost certainly means *clear* - the opposite of the
  obvious reading, on exactly the fields where clearing is wanted. Each intent
  therefore gets one unambiguous spelling: omit to leave unchanged, send the
  empty string to clear. Both refusals are enforced in Go at the decode
  boundary and declared in the generated schemas, so the two cannot drift.

  This also retires the placeholder note where `Visibility` was registered as a
  component for "future visibility controls": this operation is that future,
  and it is now a real declared surface rather than an anticipated one.

  Village API 0.9.0, both 0.9.0 request schemas, and Types 0.6.0 are retired and
  byte-frozen. Local API stays 0.6.0.

- Village API 0.9.0, Local API 0.6.0, and Types 0.6.0 add the `strike`
  harness and both observed Strike session ID forms: a timestamped prefix plus
  26 uppercase RFC4648 base32 characters, or the 26-character identifier by
  itself. Existing session ID forms and path-safety checks remain unchanged.
  Village API 0.8.0, both 0.8.0 request schemas, Local API 0.5.0, and Types
  0.5.0 are retired and byte-frozen.

- Village API 0.8.0 and Types 0.5.0 add published association ownership and
  annotation ingress. `GitContext.Associations` carries optional
  `PublishedAssociation` records with an opaque producer-owned `AssociationID`
  and observed commit hash. IDs and observed hashes are unique within a publish
  request. Consumers retain one durable ID per owner, transcript, and observed
  hash: exact replay is idempotent, while changed bindings and aliases are
  rejected. `AnnotationPushItem.TargetAssociationID` selects an association
  target exclusively, contributes to its content hash, and is documented by the
  Village `POST /api/v1/annotations` operation using the canonical push request
  and response types. Village API 0.7.0, its publish schema, and Types 0.4.0 are
  retired and byte-frozen.

- Local API 0.5.0 (`PeasantLocalAPIVersion`), Village API 0.7.0, and Types 0.4.0:
  the git+session timeline and insight-first code map wire surface, all additive.
   - `CommitRef` and `RewrittenCommit` carry `Associations
     []SessionAssociation`, each mirroring `SessionIDs` one-for-one in the same
     rank order. A `SessionAssociation` has a durable opaque `AssociationID`, a
     closed `AssociationConclusion`, `Confidence`, and a non-empty canonical
     sequence of atomic `AssociationEvidenceObservation` values. The atomic
     kinds are `recorded_commit`, `touched_file`, `branch_membership`, and
     `time_window`; a confirmed association may use one authoritative
     `recorded_commit` observation. `HasSession`/`SessionIDs` keep their
     existing compatibility-mirror semantics unchanged.
  - `ReviewListPayload` and `MapNodeDetailPayload` gain `RewrittenCommits
    []RewrittenCommit`: the session-era commit resolution ledger
    (`RewriteResolution` x `RewriteMethod` x `Confidence`). `live` remains a
    valid ledger row and only non-live rows render as ghosts; frozen field and
    enum names remain unchanged.
  - `MapNodeDetailPayload` and `ChangeDetailPayload` gain `Insights
    []SessionInsight`, a (`InsightKind` x `InsightProvenance` x `Confidence`)
    envelope with evidence and subjects, alongside the existing
    `Unusual`/`Frictions` signals (retained unchanged). `Classification` is
    declared on the wire but MUST be nil until a future revision
    populates it with no shape change.
  - `MapNode` gains node-grain comprehension signals: `AgentEditedCount`,
    `ReadCount`, `ReadAttribution` (`ReadAttributionState`), the composed
    `ReadState` (`ReadStateGrade`, ordinal `none < viewed < reviewed <
    reviewed_in_detail`), and the per-node region-coverage counts
    `ChangedRegionCount`/`AttributedRegionCount`/`ReviewedRegionCount`.
  - `TaskSummary` gains `ReadFiles []string`, the per-file derivation of
    `ReadCount`, mirroring `EditedFiles`'s sorted-distinct-non-nil invariants.
   - `TargetKind` gains a `file_version` member (a whole-file, content-hash
     keyed read-state receipt target) and an `association` member. An
     association annotation uses only `AnnotationSummary.TargetAssociationID`;
     it never embeds an association copy. `AnnotationSummary` retains
     `TargetFilePath`/`TargetContentHash` as the file-version discriminator pair.
  - Ten new enum-exhaustion corpora, a segmented insight fixture (mechanical /
    mined / classification-must-be-nil / rejections), and an extended project
    timeline corpus back every new closed set and cross-reference invariant.
  - `types-0.3.0`, `peasantlocal-api-0.4.0`, `village-api-0.6.0`, and
    `publish-request-0.6.0.schema` are retired and byte-frozen.

- Automated npm publication of `@peasant-labs/schema` in the release ceremony.
  `release.yml` gains an `npm-publish` job, gated behind the same guard →
  nix-vendor-hash → contract-gates chain as the GitHub Release publish job and
  independent of it: it installs the locked TypeScript toolchain, re-runs the
  package gates (typecheck, test, package:audit, package:smoke) on the release
  commit, stamps the package version from the tag (stripping the leading `v`;
  the committed manifest stays `0.0.0-development` + `private: true` as the
  local safety) and drops the `private` flag in the CI working copy only, then
  publishes with `--access public` under a dist-tag derived from the tag
  grammar (`-rcN` prerelease to `next`, final `vX.Y.Z` to `latest`).
  Authentication is npm Trusted Publishing (GitHub Actions OIDC) - the job
  requests `permissions.id-token: write`, runs under the `npm-publish` GitHub
  environment, and holds no npm token secret. `release-guard check-workflow`'s
  per-repo policy (`.github/release-guard.policy.yml`) and its `WorkflowPolicy`
  grammar (`internal/release`) now require the `npm-publish` job with the same
  needs-edges plus its `id-token: write` permission and `npm-publish`
  environment binding, and the fixture corpus gained synthetic-break cases
  proving each assertion fires when the job, its needs, its permission, or its
  environment is missing or wrong.

## [v0.1.0-rc6] - 2026-07-17

### Added

- An unpublished `@peasant-labs/schema` TypeScript package whose root is generated
  from the comprehensive Types 0.3 Go catalog. It mirrors the Go module with
  named wire types, Zod runtime schemas, closed sets, frozen registries and
  guards, plus `/testcase`, `/fixtures`, `/fixtures/quality`, and
  `/fixtures/timeline` subpaths. The package remains at
  `0.0.0-development`; no publication workflow is enabled.
- Cross-language testcase corpus helpers and Go-shaped typed quality-fixture accessors.
  The Go and TypeScript loaders share strict YAML cases, and consumers read
  generated quality data rather than parsing repository-relative YAML.
- Normalized timeline session identities, authoritative many-to-many commit
  bindings, and explicit project resolution in Peasant Local API 0.3.0. The
  canonical 16-family relationship corpus is available to Go and TypeScript
  consumers through schema-owned typed fixture loaders. Its exact-identity
  mutation oracle remains schema-repo test infrastructure and is not published.

### Changed

- Peasant Local API `0.4.0` and Types `0.3.0` promote file-change statuses and
  unified-diff line kinds from unconstrained strings to the named closed sets
  `FileChangeStatus` (`M`, `A`, `D`, `R`) and `DiffLineKind` (`context`, `add`,
  `del`). The JSON tokens are unchanged. Go exposes canonical inventories,
  predicates, and actionable validators; generated TypeScript exposes matching
  runtime objects, inventories, types, and predicates. Strict fixture and
  independent identity-manifest mutations guard both sets.
- Village API `0.6.0` gives the stricter publish HTTP body its own
  `OpenapiTranscriptPublishRequest` component identity and uses exact canonical
  Types schemas for every same-name shared operation component. Released `0.5.0`
  artifacts remain byte-frozen.
- Publish request schema compilation is lazy and thread-safe so a newly bumped
  Village API version can generate its embedded schema before that artifact
  exists. Runtime validation still compiles the exact embedded bytes on first use.
- `testcase.LoadCorpus` now rejects unknown fields, duplicate case names, and
  trailing YAML documents. The quality fixture loader now models the complete
  variation catalog and rejects unknown fields, duplicate names, and trailing
  documents.
- `@peasant-labs/types` is documented as deprecated. New TypeScript consumers
  use the schema-owned generated package instead of extending the handwritten
  port.
- The package root is the canonical shared TypeScript contract namespace. Hey
  API's Zod plugin generates definitions and runtime validators from the Types
  OpenAPI document. The existing `/local-api` and `/village-api` subpaths retain
  type-only `paths` and `operations` namespaces, generated with
  `openapi-typescript`; shared operation payloads resolve to root identities.
  The deprecated `/types` root re-export is retained. No HTTP or WebSocket
  transport client SDK is generated.
- TypeScript `ProjectHash` now mirrors the Go validated newtype with one
  generated nominal identity plus `newProjectHash`, `isProjectHash`, and
  `validateProjectHash`. Root payloads and Local/Village operations carry the
  brand, so plain strings cannot silently cross a project-identity boundary.
- The TypeScript package docs now spell out the validation boundary: generated
  Zod schemas cover structural shape only; Go `Validate` owns cross-field
  semantics at the trust boundary, and consumer adapters remain responsible for
  any post-parse semantic handling.
- The bespoke Go-to-TypeScript emitter is removed. TypeScript `testcase`
  `Classification` and `ProvenanceSource` closed sets are generated from
  `testcase.go`'s `AllClassifications`/`AllProvenanceSources`; only the
  YAML-decoding mechanics are handwritten. Generated fixture data comes
  directly from the schema-owned YAML corpora.

## [v0.1.0-rc5] - 2026-07-14

Fixture-contract correction with no OpenAPI wire-shape change.

### Changed

- The canonical git remote redaction example now records `maximum` as the rule's minimum firing
  level while retaining the semantic `project` category. Consumers can distinguish activation
  policy from category without introducing a second project-identity vocabulary.

## [v0.1.0-rc4] - 2026-07-10

Additive contract change: the pull skip-gate surface. The Village API spec bumps to **0.5.0**
(additive = minor); the retired 0.4.0 specs are retained byte-frozen. No breaking change; no existing
wire shape is altered.

### Added

- **Pull skip-gate wire types + endpoint (`POST /api/v1/pull/transcripts/skip-gate`).** A pulling
  client sends, per transcript it holds, the id + the content-hash it holds + its own annotation-hash
  set (`PullSkipGateRequest` / `PullSkipGateItem`), and receives per pullable id
  `{contentCurrent, annotationsCurrent}` (`PullSkipGateResponse` / `PullSkipGateResult`), so it can
  skip re-pulling transcripts and annotations that have not diverged. Only held hashes travel, never
  content. A non-pullable id's currency is withheld by omission from the results (the caller receives
  at most the ids it sent; an absent id is unanswered), so the batch cannot become an
  existence/currency oracle over ids the caller cannot pull. Constructors canonicalize the request and
  response (per-item annotation set sorted + de-duplicated, entries ordered by transcriptId, non-null
  arrays), mirroring the annotation-manifest skip-gate precedent. The Village API spec bumps to 0.5.0;
  the 0.4.0 goldens are frozen under the retired-spec immutability guard.

### Changed

- **Skip-gate tests adopt the segmented multi-axis fixture convention.** The skip-gate corpus becomes a
  typed struct of four named `testcase.Corpus` arms (round-trip, canonical, ordering, withheld), each
  guarded by `assert.RequireMin` + the new `assert.RequireValid`, so every case carries provenance +
  mutation metadata and the withheld byte-non-leak is asserted per case. Test-only; no wire change.

## [v0.1.0-rc3] — 2026-07-08

Tooling + CI release candidate — **no wire-contract change from rc2** (the OpenAPI specs and the
version markers are unchanged). Bundles the release-guard consolidation, the advisory
exported-Go-API gate, release-CI hardening, and a small maintenance removal. Published as a GitHub
**prerelease**.

### Changed

- **`release-guard` adopts `google/go-github` and becomes the single canonical release-gating
  tool.** The gh-CLI shell-outs (collaborator permission / workflow-run status / PR reviews) are
  now typed `google/go-github` calls behind a mockable client; a server-side `HeadSHA` filter
  replaces the old 100-run client-side scan. `check-workflow` reads a per-repo
  `.github/release-guard.policy.yml`, so one tool serves schema, peasant, and later village; the
  tool reads `GH_TOKEN` (matching the release workflows). go-github is isolated to
  `cmd/release-guard`; the published contract surface is untouched. The `GitRunner` for the
  git-lineage checks is hardened (subcommand allowlist, leading-dash rejection, `--end-of-options`),
  and the grammar / policy / go-github seam tests are data-driven `testdata/*.yaml` fixtures.
  (peasant-labs/peasant#113)

- **The exported-Go-API breaking-change gate is now advisory while the module is pre-1.0, and now
  also surfaces non-breaking changes for review.** An incompatible exported-Go-API change no longer
  fails CI: it surfaces as a workflow warning annotation and a single sticky pull-request comment
  listing the changed or removed symbols, and the gate exits successfully. The comment additionally
  lists compatible (non-breaking / additive) API changes to aid reviewers; those never affect the
  gate outcome. The OpenAPI breaking-change diff and the OpenAPI linter remain hard gates. Detection
  is unchanged — the same incompatible changes are still found; only the consequence changed.
  Intentional spec-version stamp bumps stay exempt (no spurious warning), and an unrecognizable
  diff-tool output format still fails closed (it means the gate can no longer see whether a real
  break is hidden).

- **Release CI hardening.** The go-apidiff advisory comment is now posted from a dedicated
  `workflow_run` workflow so `tests.yml` stays read-only — a reusable-workflow permission conflict
  (the gate job requesting `pull-requests: write`) was previously blocking the release-PR pipeline
  from starting — and the comment now works on fork PRs. The release classifier decides rc-vs-final
  once via the tested `release-guard parse-tag`, and the vs-prior breaking-change diff is skipped
  for `-rc` prerelease tags (a prerelease has no prior same-line release to diff against).
  (peasant-labs/schema#15, #16, #21)

### Removed

- **The never-adopted `migrations/village` embed** — an embedded-SQL sharing mechanism that was
  never used by any consumer — is dropped; the module stays contract-only.
  (peasant-labs/schema#18)

## [v0.1.0-rc2] — 2026-07-06

Second release candidate. Adds the **transcript-licensing contract** and regenerates
the OpenAPI specs at `0.4.0`, so the consumers (`peasant`, `village`) can pin a single
public contract that carries the license surface. Published as a GitHub **prerelease**.

### Added

- **License surface** — a `License` type with the canonical menu `CC0-1.0` /
  `CC-BY-4.0` / `CC-BY-SA-4.0` (`AllLicenses`, `IsValid()`, `LicenseMenu()`), and an
  optional `license` field on `PublishRequest` and `PullTranscriptInfo`. The
  publish-request JSON-Schema `license` enum is derived from `AllLicenses`, never a
  literal.
- **`0.4.0` OpenAPI specs** — `VillageAPIVersion` bumped `0.3.0 → 0.4.0`; regenerated
  `generated/village-api-0.4.0.{json,yaml}` + `generated/publish-request-0.4.0.schema.json`
  carrying the `license` property. The `0.3.0 → 0.4.0` delta is **additive** (oasdiff).
- **Retired-spec completeness guard** — a codegen test asserting every committed spec
  under `generated/` is either current-generated or registered in the retired-spec
  immutability registry. The retiring `0.3.0` specs are now registered and byte-frozen.
- **Quality-fixtures loader** for the contract fixtures.
- **`LICENSE`** — the module is now licensed **Apache-2.0** (`meta.license` set to
  `asl20`; the prior placeholder is removed).

### Notes

- The `VillageAPIVersion` constant (and its sibling spec-version stamps) is an
  intentional version marker consumers pin against; its value change on each spec
  release is exempted from the exported-Go-API breaking gate, with drift detected on
  the consumer side. The retained `pkg/schema/v*` tags remain out of scope as release
  references of this module — unchanged.

## [v0.1.0-rc1] — 2026-06-20

First release candidate of the extracted schema contract: `pkg/schema` lifted out
of `peasant-labs/peasant` into this standalone public module
([peasant-labs/peasant#114](https://github.com/peasant-labs/peasant/issues/114)).
Published as a GitHub **prerelease**; the consumers (`peasant`, `village`) pin
`@v0.1.0-rc1` to validate the pipeline end-to-end before a final `v0.1.0`.

### Added

- **Standalone `github.com/peasant-labs/schema` leaf module** — the generated
  OpenAPI specs, typed contract fixtures, the redaction-example single source of
  truth, and the `validate` / `openapi` / `generated` packages — guarded by a
  leaf-purity audit that pins the dependency set to the allowed contract deps.
- **Release pipeline** — PR-title-driven `release-pr.yml` / `release.yml` +
  `release-guard`, with the contract gates (oasdiff / go-apidiff / vacuum)
  provisioned through the Nix flake devShell rather than `go.mod`.
- **Fixture accessors** `EvalSchemaJSON` + `ContractCorpusFS`, and
  `PublishRequestSchemaJSON()` — validation now compiles the generated
  publish-request schema as the single byte-source.

### Notes

- The retained `pkg/schema/v*` tags (from when the contract lived nested inside
  `peasant`, up to `v1.3.0`) are **not** release references of this module and are
  never moved or deleted.
