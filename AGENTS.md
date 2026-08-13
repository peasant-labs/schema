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
shared payloads resolve to the canonical package root. Neither language package
provides an HTTP or WebSocket transport client SDK. It has
**no runtime server**: no HTTP handler and no
WebSocket hub (the WebSocket types live here, but the hub stays in peasant). So
the test suite is about keeping the contract honest, not about request/response
behaviour (see `TESTING.md`).

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

Releases are PR-title driven. Open a PR to `develop` titled
`release(vX.Y.Z[-rcN]): <summary>` (for example
`release(v0.1.0-rc1): first release candidate`). The title grammar is the single
source enforced by `cmd/release-guard parse-title`, and the release workflow gate
keys on the title, so it runs regardless of which files the PR touches (a
content-carrying release PR is supported). On merge,
`.github/workflows/release-pr.yml` mints the annotated tag using the
`peasant-labs-releaser` GitHub App token (the default CI token cannot push a tag
that re-triggers `release.yml`), and the tag drives `release.yml` to publish a
GitHub Release with the OpenAPI specs as assets (a prerelease for `-rcN`). A final
`vX.Y.Z` requires a same-version `-rcN` that is green and an ancestor of the final
commit (`release-guard check-final`).

**Minting a public tag is a reserved maintainer action.** An agent must not push
a public release tag or perform the cut. Prepare and queue the release PR; the
maintainer merges it and CI mints the tag. Tags are append-only: never move or
delete a release tag, and the old nested `pkg/schema/v*` tags (from when this
lived inside peasant) are retained forever.

When reading an `rc` number, disambiguate three things that look alike: a peasant
product release version, the schema module's own release tag, and an internal
codename that merely looks version-like. Only a real tag on this module gates a
release here.

## Spec versioning: bump, do not mutate

A released spec version is immutable. To change a surface, bump its version stamp
in `versions.go` (the current versions are stamped there as `VillageAPIVersion`,
`PeasantLocalAPIVersion`, and `TypesVersion`) and regenerate; do not edit a
version already released. Retired versions are byte-frozen and guarded by
`cmd/schema-gen/retired_specs_test.go`, which fails on both an in-place edit and a
deletion. An intentional stamp bump is exempt from the exported-Go-API advisory
gate; every other incompatible change is reported.

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
