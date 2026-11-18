# Agent guide for peasant-labs/schema

`github.com/peasant-labs/schema` is the canonical Go **wire contract** for the
peasant-labs system. This file is the fast orientation for an agent working in
this repo; the deep references it points to (`README.md`, `TESTING.md`,
`CONTRIBUTING.md`, `docs/release-runbook.md`) are authoritative where they go
further.

## What this repo is

The module defines the data every backend produces and every client consumes:
`SessionDetailPayload`, `TurnDetail`, `ToolCallDetail`, `CommitInfo`, the closed
enums, the annotation types, the push envelopes, and the `License` surface.
peasant and village both produce this wire from their backends and render it in
their clients.

It is a **contract-only leaf**: types, closed enums, generated OpenAPI specs, the
publish-request JSON Schema validator, typed fixtures, and the codegen plus
release-guard tooling. Its `typescript/` package generates the same contract
definitions and runtime schemas for TypeScript consumers and ships inside the
module's tagged releases (first in `v0.1.0-rc6`); npm publication is automated
in the release ceremony (`release.yml`'s `npm-publish` job, behind the same
gates as the GitHub Release) and `@peasant-labs/types` is deprecated. The TypeScript package retains type-only
`paths` and `operations` contracts at `/local-api` and `/village-api`, while
shared payloads resolve to the canonical package root. Those subpaths and the Go
`route` package also carry generated route members, one per operation, which
`go run ./cmd/schema-gen` projects from the built documents. Neither language package
provides an HTTP or WebSocket transport client SDK. It has
**no runtime server**: no HTTP handler and no
WebSocket hub (the WebSocket types live here, but the hub stays in peasant). So
the test suite is about keeping the contract honest, not about request/response
behaviour (see `TESTING.md`).

**Recorded exception:** `RemoteLabel` (`remote_label.go`) is a small
presentation rule, not a serialized contract type or an OpenAPI surface. It
lives here anyway because this module is the only Go dependency both peasant
and village share, and the rule must render byte-identical display labels in
both; duplicating it per-repo would let the two renderings drift. Any further
presentation logic added to this module needs its own recorded exception
here, not a silent precedent from this one.

**Recorded exception:** the TypeScript redaction category bindings project the
released redaction engine's presentation labels for shared browser consumers.
`tools/redaction-categories/` is an isolated, checksum-pinned generation module;
it calls the engine's public inventory, validation and label APIs. The root Go
module remains a dependency leaf. This adds no wire enum, OpenAPI surface or
standalone Go validator policy. Regenerate through `make schema`; never copy a
category-to-label mapping into a consumer.

**Recorded exception:** `commandpr` uses `mvdan.cc/sh/v3` solely to parse recorded shell commands for shared unverified pull request evidence. It never executes commands, fetches remote state, or changes access; sharing the parser prevents consumer detection rules from drifting.

**Recorded exception:** `transcript_cut.go` holds producer rules, not wire
types: `TranscriptTurnByteLength` and `TranscriptJSONStringByteLength` (the byte
measure), `CutTranscriptSegments` (the cut, with each segment's conversation
turn span), `ResolveTranscriptWindow` (the anchor window) and the constants
`TranscriptSegmentBudgetBytes` and `TranscriptWindowSpan{Default,Minimum,Maximum}`.
They are pure functions over wire types and live here because every server that
serves the segmented transcript reads must cut the same turns into the same
segments; a copy per backend would let boundaries drift, and a directory from
one server would then not match a segment from another. The budget is a
contract value: changing it moves every boundary on every server. The moved
literals in `testdata/transcript_cut.yaml` and `testdata/transcript_windows.yaml`
are the proof that a boundary did not move; never re-record them to make a
change pass.

These functions have **no TypeScript counterpart**, by decision. The cut needs
every turn of a session, which no browser holds, and both callers are Go
servers. A browser reads the directory, the segments and the window a server
resolved; it never cuts or resolves. The only span bound on the TypeScript side
is the generated Zod schema of `TranscriptWindow` (`span` 1..10), which comes
from the struct tags on `TranscriptWindow.Span`; no handwritten TypeScript
validator restates it. `TestTranscriptWindowSpanConstantsMatchTheDeclaredRange`
holds those tags, and the `span` query parameter of the Local and Village
window reads, equal to the Go constants. Do not add a TypeScript cut, measure or
window resolver; if a browser ever needs one, that is a contract change with its
own recorded decision.

**Recorded exception:** `transcript_numbering.go` holds the cooking and
numbering rule for the same reason as the cut: every server that serves the
partitioned reads must decide membership and assign the same display number.
Its JavaScript counterpart is the browser rule that already exists, because a
whole-document render or Markdown export has no server-numbered row. The
counterpart is pinned by one shared fixture, is never consulted for a
server-numbered row, and fails that fixture on divergence. The rule against a
TypeScript cut, measure, or window resolver above is unchanged.

**Recorded exception:** `transcript_entry_locate.go` holds the exact-member
scan over served rows. Both servers use the same rule to locate an existing
relationship entry; copying it per host would make link targets drift.

**Recorded exception:** Kickstart response and view components are closed as well as
request bodies. Adding a response property is a breaking change to pinned consumers;
this choice is specific to the kickstart contract and is not a general precedent.

## Branches and landing

- `develop` is the default and integration branch; `main` is **reserved for
  releases** and advances only on a release. The remote `HEAD` points at
  `develop`.
- Work in a per-branch worktree under the shared parent, never by switching
  branches in place (see `CONTRIBUTING.md` for the worktree and `go.work`
  layout).
- Land a change through a GitHub squash-merge PR into `develop`. The squash
  message reads subject, then a body summary, then the squashed-commit list.
- After any merge, sync the local `develop` worktree (`git -C schema/develop
  pull`) before branching new work, so a feature branch never starts behind
  origin.

## The cross-repo contract ceremony

A change to the wire is its **own** schema PR (and, when it needs a version, its
own tag) **before** the consumer PRs that re-pin it. peasant (`go.mod`) and
village (`backend/go.mod`) pin a published `schema` version, and village both
serves AND enforces the spec from the pinned module, so the served contract and
the enforced contract can never drift. Do not land a consumer re-pin against an
unpublished schema change.

Before adding or changing an optional compatibility-sensitive field or a
content-capability token, read
`docs/content-capability-negotiation.md`. It is authoritative for deciding when
negotiation is required and for the wire, deployment, evolution, and cross-repo
verification rules.

## Releases and tags (reserved)

Titles carry release intent as `type(scope)[level]: subject`; below 1.0 a
breaking change uses `![minor]`. The pinned `release-intent` checks the live
title against the Go and OpenAPI contract report. A functional spec change
requires a releasing title that names each changed surface. A releasing title
must not name an unchanged or docs-only surface unless it was already ahead on
the base. A non-releasing title such as `docs(local-api): clarify a field`
may mention a surface. `spec-change-needs-bump` and `surface-mismatch` name
the respective failures.

Content PRs regenerate working specs; they do not choose spec numbers or edit
the protected `versions.go`, numbered documents, freeze registries,
`CHANGELOG.md`, or the `typescript/package.json` version. The standing bot
release PR writes the frozen specs with the changelog and lands through the
merge queue.

`release-batch.yml` computes the module version and pushes an annotated tag on a
fresh release PR commit using the releaser App. The tag triggers `release.yml`,
which checks the contract, the Nix vendor hash and the npm version before
publishing the generated specs and npm package. The vendor-hash bot opens its
own queued PR; no workflow pushes `develop`.

Agents never push public tags or edit, stack on, or merge bot release PRs by
hand. Enqueueing a releasing content PR is the release decision. Tags are
append-only: never move or delete them, including the historical nested
`pkg/schema/v*` tags.

When reading an `rc` number, disambiguate three things that look alike: a peasant
product release version, the schema module's own release tag, and an internal
codename that merely looks version-like. Only a real tag on this module gates a
release here.

## Spec versioning: working and frozen documents

For Types, Village, Local, and segmented read, `generated/<stem>-next.*` is
the committed working document. `versions.go` names the latest frozen numbered document.
The two Village request schemas follow the Village version. The public
accessors and exported validators read the numbered frozen bytes; generators,
contract reports, and current-document tests read working bytes. No public
accessor exposes a working file.

`schema-gen status` computes `clean`, `docs-ahead`, or `ahead` per surface.
Docs-only changes need no bump and may wait for the next functional freeze.
The release job uses `release-intent freeze-plan`, then `schema-gen freeze`,
to raise constants, regenerate working version lines, copy numbered files and
update the frozen registry and version mirrors in the release PR. A release
tree must contain no `ahead` surface: the PR and merge-group check and the
candidate pre-tag check enforce this. A wrong stale freeze that bypasses the
queue burns one spec number; the next freeze advances again and retains the
wrong numbered file.

Segmented read separates the frozen document version
(`SegmentedReadSpecVersion`) from the selected wire protocol
(`SegmentedReadAPIVersion`, with `SegmentedReadMinVersion` as the oldest
served protocol). A document freeze moves only the document version. The
minimum protocol document and the generation selector table remain unchanged.
A breaking protocol change requiring a new `v=` generation is separate P1
work, not a document freeze.

When an open content PR predates this migration, merge the updated base,
drop its numbered-file and constant changes, regenerate the working files,
and list its functionally changed surfaces in the title. See the release
runbook for the full cutover and recovery flow.

## Leaf purity

The module stays a dependency leaf. Its `go.mod` direct `require` set must be a
subset of an explicit allowed set (`leaf_audit_test.go` is the authoritative list;
`CONTRIBUTING.md` explains the rationale). Dev and CI tools (`oasdiff`,
`go-apidiff`, `vacuum`, linters) live in the Nix flake dev shell, never in
`go.mod`: no `tools.go`, no `go get -tool`. And `go.mod` must carry no local-path
`replace` directive (`=> ./...` or `=> ../...`), which is the only way first-party
source can leak into the Nix `vendorHash`; `vendorhash_test.go` fails closed on
one. Co-develop a consumer against a local schema worktree with a gitignored
`go.work` (see `CONTRIBUTING.md`), never a committed `replace`.

## Testing and fixtures

Test cases live in `testdata/*.yaml` fixtures, never as inline case tables in
`_test.go`. The canonical form is the `testcase` corpus package: a generic
`Case[I, E]` / `Corpus[I, E]` with closed-set classification and provenance plus
mutation metadata, a pure loader, and pure non-vacuity validators, with the
`*testing.T` seams (`RequireMin`, `RequireValid`) split into `testcase/assert`.
New corpora adopt it; a feature whose fixtures split into heterogeneous behavioral
arms uses the segmented form (a typed struct of named per-arm corpora, each
guarded by `RequireMin` + `RequireValid`). `TESTING.md` documents the standard,
the segmented convention, and every gate behind it. A single contract change
should update one YAML corpus, not several inline tables.

## Codegen and freshness

`go run ./cmd/schema-gen` produces the OpenAPI specs and Redoc pages. `pnpm --dir
typescript run generate` uses the pinned Hey API Zod plugin for root contract
definitions and runtime schemas, `openapi-typescript` for the type-only Local and
Village operation contracts, and canonical YAML for fixture data.
`make schema` runs both stages. Never hand-edit a generated file. The committed output
must be byte-identical to a fresh generation: the freshness gate (`make freshness`
and `TestCodegenFreshness_SpecsMatchSource`) fails on drift, and the fix it names
is always regenerate then commit.

Install the locked TypeScript development toolchain with
`pnpm --dir typescript install --frozen-lockfile --ignore-scripts` before regenerating.

## Gates

`make check` is the authoritative full-repository gate: fmt, vet, the
release-workflow guard, TypeScript typecheck/tests/package smoke, and
`go test -race ./...` (which carries the freshness, immutability,
leaf-audit, vendor-hash, and synthetic-break checks). `make gates
BASE_REF=origin/develop` runs the breaking-change gates (`oasdiff`, `go-apidiff`,
`vacuum`) and needs the flake tools. `TESTING.md` has the full gate table and the
real test behind each one.

## Static guard

`make check` runs the unstring static guard. The pin lives in one place, the
`UNSTRING` variable in the `Makefile` (default
`go run github.com/peasant-labs/infra/unstring/cmd/unstring@v0.0.4`). The run
needs Go 1.26 or newer and ast-grep 0.51.2 exactly; the dev shell provides that
ast-grep, and any other version stops before a scan.

`.unstring-baseline.yaml` is generated. `init` writes it once. After a count
falls, refresh it with `make unstring-update`. Do not hand-edit or hand-merge
it.

A hit that is not a closed-set value may carry
`// unstring-allow: <rule-id> — <reason>` on the hit. The reason is at least
three words, after an em dash. `closed-set-literal` and `go-closed-set-literal`
accept no marker: the fix for a set value is the typed value.

`route/routes_gen.go` and `typescript/src/internal/generated/*.gen.ts` are
exempt as generated code. A `rule: "*"` exemption is accepted only when every
matched file carries a `Code generated … DO NOT EDIT` header, which those files
do. `local-api.ts`, `village-api.ts`, and `contract/zod.gen.ts` under
`typescript/src/internal/generated/` are generated too, but they lack that
header, so the tool will not exempt them; they stay in the scan and the
baseline.
`schema.module` is omitted: this repository is the contract, not a consumer of
its own package, so the catalog is this repository's typed constants and
TypeScript unions.

The tool contract is [`infra/unstring/README.md`](https://github.com/peasant-labs/infra/blob/unstring/v0.0.4/unstring/README.md).

## Hygiene

- Describe work by substance. Do not write internal task-tracking terminology
  into shipped code, comments, docs, or commit messages.
- No AI-slop punctuation in prose: use ASCII hyphens (not em-dashes or
  en-dashes), straight quotes (not smart quotes), and three dots rather than an
  ellipsis glyph. Semantic or diagram characters the docs already use (arrows,
  the identical-to sign, a middot separator) are fine.
- Never install a git hook (from `bd` or any other tool) without explicit
  permission. This is a hard rule.
- Never hand-merge a generated file on conflict: merge the source, rerun the
  generator, and commit the byte-identical result.

## Pointers

| Read | For |
|---|---|
| `README.md` | The contract overview: what is in the wire, the gate quickstart, and how pinning works. |
| `docs/content-capability-negotiation.md` | The authoritative content-capability wire semantics, field decision policy, deployment guarantees, and Schema/Village/Peasant test ownership. |
| `TESTING.md` | The full test strategy: every gate and the real test behind it, and the `testcase` standard. |
| `CONTRIBUTING.md` | The worktree and `go.work` layout, the leaf-audit and no-`replace` rules, regeneration, and the release ceremony. |
| `docs/release-runbook.md` | The release operator guide: secrets, the GitHub App, the tag cut, and troubleshooting. |
