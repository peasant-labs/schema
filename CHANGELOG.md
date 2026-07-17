# Changelog

All notable changes to the `github.com/peasant-labs/schema` contract module are
documented here. This project adheres to [Semantic Versioning](https://semver.org/).

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
