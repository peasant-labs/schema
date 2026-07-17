# Contributing to peasant-labs/schema

This repo is the **public API contract** — a single, contract-only Go leaf
module. The rules below keep it a clean leaf and keep its Nix build reproducible.

## Worktree layout (per-branch worktrees under a shared parent)

We develop across the peasant client, the village server, and this schema
contract simultaneously, so all three live as sibling clones under one shared
parent, with **one worktree per branch** (never branch-switching in place):

```
~/dev/peasant-labs/
├── schema/
│   ├── __dummy__/                              # bare-ish anchor checkout (see below)
│   ├── peasant-114--breaking--extract-pkg-schema/   # one worktree per branch
│   └── <other-branch>/
├── peasant/        # github.com/peasant-labs/peasant clones, same per-branch layout
└── village/        # the marketplace server clones, same per-branch layout
```

### The `--no-checkout` / `__dummy__` anchor

Clone once with `--no-checkout` so the initial checkout doesn't litter a default
branch, then add a throwaway `__dummy__` worktree as the stable anchor every
real per-branch worktree is created from:

```bash
git clone --no-checkout git@github.com:peasant-labs/schema.git ~/dev/peasant-labs/schema/.git-anchor
# create the dummy anchor on an orphan/empty ref so no real branch is "checked out" twice
git -C ~/dev/peasant-labs/schema/.git-anchor worktree add --detach ../__dummy__
# then one worktree per branch you work on:
git -C ~/dev/peasant-labs/schema/.git-anchor worktree add ../<branch> <branch>
```

The point: a branch is only ever checked out in exactly one worktree, so parallel
agents/sessions never fight over the index, and you can have peasant + village +
schema all on coordinated branches at once.

## go.work for branch combinations

Because consumers and the contract evolve together, use a `go.work` at the shared
parent (`~/dev/peasant-labs/`) to point peasant/village at a LOCAL schema worktree
before the version is cut/published:

```bash
cd ~/dev/peasant-labs
go work init
go work use ./schema/<branch> ./peasant/<branch> ./village/<branch>/backend
```

To switch between branch combinations without editing the shared file, keep
per-combo work files and select with `GOWORK`:

```bash
GOWORK=~/dev/peasant-labs/go.work.extract-114 go test ./...
```

`go.work`/`go.work.sum` are gitignored — they are local dev wiring, never
committed. Once `schema@vX.Y.Z` is published, consumers drop the `go work use`
and pin the version in their `go.mod`.

## The leaf-audit rule

This module's `go.mod` direct `require` set MUST be a subset of:

```
github.com/dayvidpham/bestiary
github.com/santhosh-tekuri/jsonschema/v5
github.com/swaggest/jsonschema-go
github.com/swaggest/openapi-go
golang.org/x/crypto
gopkg.in/yaml.v3
```

- **No `tools.go`, no `go get -tool`.** Dev/CI tools (`oasdiff`, `go-apidiff`,
  `vacuum`, linters) live in the flake `devTools`, never in `go.mod`.
- **No `coder/websocket` / no runtime deps.** The WebSocket *types* live here; the
  Hub runtime stays in peasant forever.
- A leaf-audit test/gate fails if anything outside the allowed set enters
  `require`.

## The no-local-`replace` (vendorHash-stability) rule

`go.mod` MUST NOT contain a local-path `replace` directive (`=> ./…` or `=> ../…`).
The Nix `buildGoModule` `vendorHash` covers only the third-party module graph
(`go.mod` + `go.sum`), so a first-party edit can never drift it — UNLESS a local
`replace` folds first-party source into the vendor graph (the #119 pathology).
`TestVendorHashStableOnFirstPartyEdit` enforces this structurally.

When you DO bump a third-party dependency, recompute the `vendorHash`: set it to
`nixpkgs.lib.fakeHash` in `flake.nix`, run `nix build`, and copy the reported
`got:` hash back. (A first-party-only change never needs this.)

## Nix package lookups

When checking whether a package exists in nixpkgs, use **`nix-search <name>`** —
never `nix search nixpkgs` (it is slow and can hang).

## Regeneration & gates

```bash
pnpm --dir typescript install --frozen-lockfile --ignore-scripts
make check                 # gofmt + vet + release-workflow guard + go test -race ./...
make gates BASE_REF=origin/develop  # oasdiff + go-apidiff + vacuum vs a base ref
make schema                # regenerate OpenAPI, docs, and TypeScript contract outputs
nix build                  # hermetic build + go test ./... sanity gate
```

The pnpm install is development tooling only and is pinned by
`typescript/pnpm-lock.yaml`. It does not change `go.mod` or the Go module's
leaf dependency set. `make schema` first regenerates the canonical OpenAPI
documents from Go, then uses the pinned Hey API Zod plugin to derive the root
contract and runtime schemas, `openapi-typescript` to derive type-only Local and
Village operation contracts, and canonical YAML for fixture data. Never
hand-edit a generated TypeScript file.

Before reporting a TypeScript contract change ready, run:

```bash
pnpm --dir typescript run typecheck
pnpm --dir typescript test
pnpm --dir typescript run package:audit
pnpm --dir typescript run package:smoke
```

The last gate packs the artifact, installs it into a disposable directory, and
imports every public subpath. This package remains unpublished until a separate
release change explicitly enables package publication.

Gates that run in CI (and locally):

- **`make check`** (`.github/workflows/tests.yml` → `check` job): codegen-freshness
  (committed specs == generator output), retired-spec immutability (released specs
  byte-frozen by sha256), the peasantlocal/village surface spot-checks, the
  **leaf-audit** (`go.mod` direct requires ⊆ the allowed set), the
  `vendorHash`-stability proof, and the **synthetic-break tests** that prove the
  `oasdiff` / `go-apidiff` gates actually fire.
- **`make gates`** (`tests.yml` → `contract-gates` job, PR-only): the breaking-change
  gates diff the regenerated specs + exported Go API against the PR base —
  `oasdiff breaking --fail-on ERR`, `go-apidiff`, and `vacuum lint -r .vacuum.yaml`.
  Provisioned entirely from the flake dev shell (never `go.mod`); run inside
  `nix develop`.

## Release ceremony

Releases mirror peasant's PR-title-driven pipeline (the title/tag grammar and
final-release guard are in `internal/release`, exposed via `cmd/release-guard`):

1. Open a PR to `develop` titled `release(vX.Y.Z[-rcN]): <summary>`
   (e.g. `release(v0.1.0-rc1): first release candidate`). The grammar is enforced
   by `release-guard parse-title`.
2. A maintainer (admin/maintain) approves — `release-guard check-maintainer` /
   `check-approval` gate this.
3. On merge, the release workflow mints the annotated tag (via the releaser App
   token); the tag triggers the release run, which publishes a GitHub Release with
   the OpenAPI specs as downloadable assets (prerelease for `-rcN`).
4. A FINAL `vX.Y.Z` requires a same-version `-rcN` that is green AND an ancestor of
   the final commit — enforced by `release-guard check-final` / `release.CheckFinal`.
5. Old `pkg/schema/v*` tags (from when this lived nested in peasant) are **retained
   forever**; new releases use the bare `vX.Y.Z` path. Tags are append-only — never
   move or delete a release tag.

The CI workflows that wire these guards live in `.github/workflows/`
(`release-pr.yml` mints the tag on merge; `release.yml` runs guard →
nix-vendor-hash → contract-gates → publish). The full operator guide — secrets,
the GitHub App installation, the `v0.1.0-rc1` cut, and troubleshooting — is in
[`docs/release-runbook.md`](docs/release-runbook.md).
