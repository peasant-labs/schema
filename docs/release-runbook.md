# peasant-labs/schema Release Runbook

Operator guide for cutting releases of the **contract library** `github.com/peasant-labs/schema`.

This pipeline **mirrors peasant's** PR-title-driven release ceremony but is adapted
for a contract library that **ships no binary**. Everything peasant does to build,
package, and smoke-test CLI binaries is **SUBTRACTED**; the contract gates
(`oasdiff` / `go-apidiff` / `vacuum`) are **ADDED**.

> **Status (rc1):** the pipeline is built and validated **locally** (actionlint,
> `nix build`, `make check`, `make gates` all green). It is **not yet proven on a
> live GitHub run** — the live `v0.1.0-rc1` release-PR is the MVP proof gate and
> must be cut by a maintainer (it needs a real push/PR + the GitHub App
> installation; an agent must not push a public tag). See **§5**.

---

## 1. The release model

A release is cut by **merging a `release(vX.Y.Z[-rcN]): <summary>`-titled PR into
`develop`** — *never* by pushing a tag by hand. The title/tag grammar lives once in
`internal/release` (exercised by `cmd/release-guard`); the workflows call that CLI.

```
release PR (title: "release(vX.Y.Z[-rcN]): …")  →  develop
   │  merged by a maintainer (approval gate deferred to the public flip — §2/§6)
   ▼
release-pr.yml (merge trigger)
   │  make nix-vendor-hash (commit a flake.nix fix if needed)
   │  → push annotated tag vX.Y.Z[-rcN] via the releaser App token
   ▼ (tag push triggers)
release.yml
   guard → nix-vendor-hash (freshness) → contract-gates ─┬─→ release      → GitHub Release (prerelease for -rcN)
                                                          │                  with the generated/ OpenAPI specs as ASSETS
                                                          └─→ npm-publish  → @peasant-labs/schema on npm
                                                                             (dist-tag next/latest, via OIDC trusted publishing)
```

`release` and `npm-publish` are independent siblings behind the same three gates: neither publish can block the other.

### What is SUBTRACTED vs peasant

No goreleaser binary builds, no deb/rpm/AUR/Homebrew/nix packaging, no
`release-validate` distro matrix, no `release-e2e`, no full-stack `e2e`, no
`smoke`. The publisher is a single **GitHub Release** carrying the OpenAPI specs.

### What is ADDED — the contract gates (the repo's CI value-add)

| Gate | Tool | What it does |
|------|------|--------------|
| OpenAPI breaking diff | `oasdiff breaking --fail-on ERR` | Fails on an ERR-level breaking change to a released spec vs the prior golden. |
| Exported-Go-API diff | `go-apidiff` | Fails if the module's exported Go API changes incompatibly vs the base ref. |
| OpenAPI lint | `vacuum lint -r .vacuum.yaml` | Fails on `error`-severity lint findings in the generated specs. |

All three are provisioned from the **flake dev shell** (`vacuum-go` from nixpkgs;
`oasdiff` + `go-apidiff` are source-built `buildGoModule` derivations) and are
**never** `go.mod` requires (the leaf-audit test enforces this). Each breaking
gate ships a Go **synthetic-break test** (`internal/contractgates/`) that proves
it fires; those run in `make check`.

### Generated request-validation assets

Each Village API version emits two JSON-only request schemas alongside the OpenAPI
documents: `publish-request-<version>.schema.json` and
`annotation-push-request-<version>.schema.json`. The root accessors
`PublishRequestSchemaJSON()` and `AnnotationPushRequestSchemaJSON()` return those
embedded bytes. Their canonical validators first enforce the extracted operation
schema, then decode the Go request type; `ValidateAnnotationPushRequest` additionally
enforces `entryTarget.endIndex > entryTarget.entryIndex`, a relational rule standard
JSON Schema cannot express.

When a Village API version is superseded, register the retiring Village JSON/YAML
pair and both JSON-only request schemas in `cmd/schema-gen/testdata/retired_specs.yaml`
with their frozen hashes. Current artifacts remain under freshness; only superseded
artifacts move into the retired registry.

### Workflow ownership

| Workflow | Trigger | Role |
|----------|---------|------|
| `tests.yml` | PR / push / `workflow_call` | `make check` + the `contract-gates` job (oasdiff/go-apidiff/vacuum vs base). |
| `release-pr.yml` | release PR open / merge | Validate title + maintainer; on merge: nix-vendor-hash + mint the annotated tag (App token). |
| `release.yml` | tag push `v*` | guard → nix-vendor-hash → contract-gates → publish the GitHub Release **and**, independently, publish the `@peasant-labs/schema` npm package. |
| `nix-vendor-hash.yml` | push to develop (dep files) | Recompute + commit the flake `vendorHash` after a dependency bump. |

The `release.yml` job graph (the `release` GitHub-Release job and the
`npm-publish` job each sit behind guard + nix-vendor-hash + contract-gates, as
independent siblings) is asserted by `release-guard check-workflow`, run in
`make check`.

---

## 2. Secrets & GitHub App (one-time setup)

| Secret | Used by | Purpose |
|--------|---------|---------|
| `PEASANT_RELEASER_APP_ID` | release-pr.yml, nix-vendor-hash.yml | The `peasant-labs-releaser` GitHub App id. |
| `PEASANT_RELEASER_APP_PRIVATE_KEY` | release-pr.yml, nix-vendor-hash.yml | That App's private key. |

**Required App installation (NOT testable pre-merge):** the `peasant-labs-releaser`
GitHub App must be **installed on `peasant-labs/schema`** with **Contents: write**
(it pushes the annotated tag + the `chore: update nix vendor hash` commit; the
default `GITHUB_TOKEN` cannot push a tag that re-triggers `release.yml`). This is
the same App peasant uses, additionally installed on the schema repo. The live
`v0.1.0-rc1` run is the first time this is exercised end-to-end.

No other secrets are needed (no registry tokens, no AUR key, no Homebrew tap — all
subtracted). npm publication deliberately holds **no secret at all**: it
authenticates via npm **Trusted Publishing** (OpenID Connect), so there is no
token to provision, rotate, or leak.

### npm publication setup (one-time, no secret)

`release.yml`'s `npm-publish` job authenticates to the npm registry via **npm
Trusted Publishing (OIDC)**, not a stored token: the job requests
`permissions: { id-token: write }`, GitHub mints it a short-lived OIDC token
scoped to the `npm-publish` GitHub Actions environment, and `pnpm publish`
(native OIDC trusted-publish support since pnpm's v11 publish
reimplementation; the flake pins pnpm 11.24.0) exchanges that token with
npmjs.com for a one-time publish credential - no `NPM_TOKEN` secret exists or
is needed. Provenance attestation is automatic on a trusted-publishing publish.

Two one-time, maintainer-side registrations make this work (both are already
unblocked: the package exists - `@peasant-labs/schema@0.1.0-rc6` was published
manually on 2026-07-17):

1. **GitHub Actions environment (create this FIRST).** Create an environment
   named `npm-publish` on `peasant-labs/schema` (Settings → Environments → New
   environment), with **no protection rules for now**. The `npm-publish` job is
   bound to it (`environment: npm-publish` in `release.yml`, asserted by
   `release-guard check-workflow`). Optional protection rules (a required
   reviewer, a tag-pattern restriction) are additional defense-in-depth that can
   be added later without any workflow change - deferred to the public-flip
   checklist (§6).
2. **npm Trusted Publisher** (register AFTER step 1 - the environment must
   already exist for npm to bind to it). On npmjs.com, `@peasant-labs/schema` →
   Settings → Trusted Publisher → GitHub Actions, and register: organization or
   user `peasant-labs`, repository `schema`, workflow filename `release.yml`,
   environment name `npm-publish` (matching step 1 exactly - an npm Trusted
   Publisher registration is exact-match on all four fields, and a run whose
   job does not carry that exact `environment:` cannot mint a valid publish
   token). Scope the allowed action to **`npm publish` only** - not `npm stage
   publish` - since this ceremony publishes directly and least-privilege scoping
   costs nothing here.

Until both registrations exist, `npm-publish` fails at the `pnpm publish` step
with the registry's own rejection (see Troubleshooting below); that failure
does **not** block the GitHub Release published by the sibling `release` job.

### Maintainer gating

Authorization is GitHub collaborator **permission** (`admin`/`maintain`), asserted by
`release-guard check-maintainer` on the PR author — not a checked-in `MAINTAINERS`
file:

- On release-PR open/edit: the PR **author** must be `admin`/`maintain`, and the title
  must parse (`release-guard parse-title`).
- On merge: a pre-existing tag is a **hard fail**. The at-least-one-APPROVED-review
  assertion is **deferred until the public flip**: with a single active maintainer,
  GitHub's no-self-approval rule makes it unsatisfiable. `release-guard check-approval`
  remains implemented + tested (byte-identical to peasant's guard); its step in
  `release-pr.yml`'s tag job is commented out and is re-enabled alongside branch
  protection at the public flip (§6).

Branch protection (≥1 approval on `develop`) is defense-in-depth, configured at the
public flip.

---

## 3. Cutting an rc (the v0.1.0-rc1 ceremony)

1. Ensure `develop` is green (`make check`) and the specs are fresh
   (`go run ./cmd/schema-gen` produces no diff).
2. Open a PR to `develop` titled exactly:

   ```
   release(v0.1.0-rc1): first release candidate of the extracted schema contract
   ```

   `release-pr.yml` (open trigger) validates the title grammar + that **you are an
   admin/maintain collaborator**, and runs the full `tests.yml` gate (make check +
   contract-gates) on the PR.
3. **Merge** (the maintainer-approval assertion is deferred to the public flip — §2/§6;
   the guard code is live and tested, its workflow step commented out). On merge,
   `release-pr.yml`:
   - runs `make nix-vendor-hash` on the merge commit (commits a `chore: update nix
     vendor hash` fix to develop first if `flake.nix` changed),
   - pushes the annotated tag `v0.1.0-rc1` on the hash-current commit (App token).
4. The tag push triggers `release.yml`: guard → nix-vendor-hash freshness →
   contract-gates → **publish a prerelease GitHub Release** with the `generated/`
   OpenAPI specs attached as assets. Independently and behind the same three
   gates, the `npm-publish` job stamps `typescript/package.json`'s version from
   the tag (stripping the leading `v`; the committed manifest stays
   `0.0.0-development` + `private: true`) and publishes `@peasant-labs/schema` to
   npm - `next` for an `-rcN` tag, `latest` for a final `vX.Y.Z` (dist-tag chosen
   from the guard job's tag classification, not re-derived).

`rcN` tags publish **prereleases**. A final `vX.Y.Z` additionally requires a
same-version `-rcN` that is green **and** an ancestor of the final commit
(`release-guard check-final`).

The shared guard also supports an explicit bootstrap for a different repository
whose first product release must be one exact final version:

```
release-guard check-final --tag v1.0.0 --initial-final v1.0.0
```

This is a narrow policy, not a general bypass. The configured value must be an
exact final tag, the requested final must equal it, and a full `v*` tag scan must
prove that no prior valid product release tag exists. Because the current tag
already exists when the tag-triggered guard starts, the scan excludes that exact
tag and separately queries GitHub for a published Release with the configured
tag. During the first guard run, publication has not happened yet. A legitimate
retry after any pre-publication failure likewise sees no Release, so both may
proceed. Once the publisher creates the GitHub Release, that durable completion
record self-disables bootstrap on every later invocation, including a rerun whose
own Actions status is currently in progress. A malformed `v*` tag, tag-listing
failure, or repository-release lookup failure blocks the release. Once any prior
product release exists, the bootstrap no longer applies and the ordinary green same-version
ancestor-rc rule resumes even if `--initial-final` remains configured; later
finals do not depend on looking up the initial final's GitHub Release. A consumer
opting in must check out full history and tags and grant the workflow token read
access to repository releases before calling the command.

Schema does not opt into this policy. Its final `v0.1.0` continues through the
ordinary guard using the green ancestor `v0.1.0-rc13`.

### npm publication

Every tag from the one that lands this automation onward publishes
`@peasant-labs/schema` to npm automatically, alongside (never gating) the GitHub
Release: `v0.1.0-rc6` (the package's first tagged appearance) was published
**manually** under dist-tag `next` before this automation existed, from the same
staged-tag checkout this job now automates; the `npm-publish` job takes over
from the next tag. Authentication is npm Trusted Publishing (OIDC) - see §2 for
the one-time GitHub environment + npm Trusted Publisher registrations. Until
both exist, `npm-publish` fails at the registry (see Troubleshooting below) -
this does not block the GitHub Release, which the sibling `release` job still
publishes. Note the npm registry's read replicas can lag the primary by several
minutes after a publish; a `npm view`/website check run immediately after a
publish can read as "not there yet" even on a successful publish, so this repo's
tooling does not assert post-publish registry state.

Before its TypeScript package gates run, the `npm-publish` job stamps only the
working-copy version and removes `private`; neither change is committed. The
package audit then reads `package/package.json` from a real tarball made from
that staged manifest, including the exact repository metadata npm provenance
requires.

**Troubleshooting `npm-publish`:**

- **`pnpm publish` fails to authenticate / npm rejects the OIDC exchange
  ("no trusted publisher configured" or similar)** - the npm Trusted Publisher
  is not registered, or its four fields (org/user, repo, workflow filename,
  environment) don't exactly match this job. Register or correct it on
  npmjs.com (§2, step 2): organization `peasant-labs`, repository `schema`,
  workflow filename `release.yml`, environment `npm-publish`. Then re-run the
  failed job from the Actions run page; no new tag is needed.
- **The job errors before publish with a permissions/OIDC-token complaint** -
  the `npm-publish` job's `permissions: id-token: write` was removed or
  narrowed, or the `environment: npm-publish` binding was removed (both are
  asserted by `release-guard check-workflow` in `make check`, so this should
  fail on the PR before it ever reaches a tag; if it doesn't, the policy file
  drifted from the workflow - see `.github/release-guard.policy.yml`). Restore
  both and cut a new release PR.
- **npm provenance rejects the package manifest's repository metadata** -
  `typescript/package.json` must retain `repository.type: "git"` and
  `repository.url: "https://github.com/peasant-labs/schema"`. The
  `package:audit` gate checks this exact metadata before publication; restore it
  and cut a new release PR.
- **`E409` / "cannot publish over the previously published version"** - the tag's
  stripped version (`${GITHUB_REF_NAME#v}`) already exists on the npm registry.
  This means the tag was already published (check `npm view @peasant-labs/schema
  versions`, allowing for registry replica lag - see above) or was re-cut after
  a prior partial publish; npm tags are append-only just like git release tags,
  so cut a new version rather than retrying the same one.
- **The gate steps (typecheck/test/package:audit/package:smoke) fail** - treat
  identically to a `tests.yml` failure: the same gates already ran on the release
  PR, so a failure here on the tagged commit means something drifted between PR
  merge and tag (for example an out-of-band edit to `develop`); fix on `develop`
  and cut a new release rather than forcing this job to pass.

---

## 4. Tag namespaces & the retention guardrail

- New releases use **bare `vX.Y.Z[-rcN]`** tags on the new module path.
- The old **`pkg/schema/v*`** tags (from when the contract lived nested inside
  peasant, up to `v1.3.0`) are **retained forever** — `release-guard parse-tag`
  rejects them as release references, and `release.yml`'s `v*` filter does not
  match them. **Never move or delete a published tag.**

---

## 5. rc1 readiness — what is verified, and what needs a maintainer

**Verified locally in this worktree (all green):**

- `actionlint .github/workflows/*.yml` — clean.
- `make check` — gofmt + vet + `release-guard check-workflow` + `go test -race
  ./...` (incl. the leaf-audit test and the oasdiff/go-apidiff synthetic-break
  tests, which run for real in the dev shell).
- `make gates BASE_REF=<ref>` — oasdiff + go-apidiff + vacuum all green.
- `nix build` — the main module + the `oasdiff`/`go-apidiff` gate derivations
  build; `vendorHash` is current.

**Requires a maintainer / the user (an agent must not do these):**

1. **Push this branch + open the PR** into `develop` on
   `peasant-labs/schema` (GitHub).
2. **Install the `peasant-labs-releaser` GitHub App** on `peasant-labs/schema`
   with Contents: write, and add the `PEASANT_RELEASER_APP_ID` /
   `PEASANT_RELEASER_APP_PRIVATE_KEY` repo secrets (§2).
3. **Cut `v0.1.0-rc1`** via the §3 ceremony — the live run is the MVP proof gate.
   A maintainer must open the `release(v0.1.0-rc1): …` PR and merge it (the
   approval assertion is deferred to the public flip — §2/§6); the pipeline mints
   the tag and publishes the prerelease. **Do not push the `v0.1.0-rc1` tag
   manually** — the ceremony is the only sanctioned path, and the App-token tag
   push is what proves the installation works.

---

## 6. Public-flip checklist (one-time)

Run these, in order, when taking the schema repo public and enabling external
consumption of tagged releases. Mirrors peasant's public-flip checklist, minus the
binary/packaging items this contract library does not ship (no goreleaser, no
deb/rpm/AUR/Homebrew/nixpkgs — see "What is SUBTRACTED vs peasant" in §1).

- [ ] Make the repository **public**.
- [ ] Confirm the `peasant-labs-releaser` GitHub App has `Contents: write` on
      `peasant-labs/schema` (§2) — it pushes the annotated tag + the `chore: update
      nix vendor hash` commit.
- [ ] Configure branch protection on `develop` (≥1 approval) as defense-in-depth,
      and **re-enable the `check-approval` step** in `release-pr.yml`'s tag job
      (commented out pre-flip; the `release-guard check-approval` guard code is live
      and tested — byte-identical to peasant's). This closes the §2 deferral.
- [ ] **(Optional) Attach protection rules to the `npm-publish` GitHub
      environment** (§2) - a required reviewer and/or a tag-pattern restriction.
      This adds a human approval gate on every npm publish with **no workflow
      change**: the `npm-publish` job already targets `environment: npm-publish`,
      so GitHub enforces whatever rules the environment carries. Pairs with the
      `check-approval` re-enable above as the same class of defense-in-depth,
      deferred pre-flip for the same reason (single active maintainer).
- [ ] **Live verification (post-flip):**
      - A `release(vX.Y.Z-rcN): …` PR merges and `release-pr.yml` mints the annotated
        tag via the App token; `release.yml` runs guard → nix-vendor-hash freshness →
        contract-gates and publishes a prerelease GitHub Release with the `generated/`
        OpenAPI specs attached, and (independently) the `npm-publish` job publishes
        `@peasant-labs/schema@X.Y.Z-rcN` to npm under dist-tag `next` via OIDC trusted
        publishing.
      - `go get github.com/peasant-labs/schema@vX.Y.Z-rcN` resolves against the public
        module for a downstream consumer (peasant / village pin).
