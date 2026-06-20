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
   │  ≥1 maintainer approval; merged by a maintainer
   ▼
release-pr.yml (merge trigger)
   │  check-approval → make nix-vendor-hash (commit a flake.nix fix if needed)
   │  → push annotated tag vX.Y.Z[-rcN] via the releaser App token
   ▼ (tag push triggers)
release.yml
   guard → nix-vendor-hash (freshness) → contract-gates → publish
   ▼
GitHub Release (prerelease for -rcN) with the generated/ OpenAPI specs as ASSETS
```

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

### Workflow ownership

| Workflow | Trigger | Role |
|----------|---------|------|
| `tests.yml` | PR / push / `workflow_call` | `make check` + the `contract-gates` job (oasdiff/go-apidiff/vacuum vs base). |
| `release-pr.yml` | release PR open / merge | Validate title + maintainer; on merge: nix-vendor-hash + mint the annotated tag (App token). |
| `release.yml` | tag push `v*` | guard → nix-vendor-hash → contract-gates → publish the GitHub Release. |
| `nix-vendor-hash.yml` | push to develop (dep files) | Recompute + commit the flake `vendorHash` after a dependency bump. |

The `release.yml` job graph (publish behind guard + nix-vendor-hash +
contract-gates) is asserted by `release-guard check-workflow`, run in `make check`.

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
subtracted).

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
3. Get **one maintainer approval**, then **merge**. On merge, `release-pr.yml`:
   - asserts the approval (`release-guard check-approval`),
   - runs `make nix-vendor-hash` on the merge commit (commits a `chore: update nix
     vendor hash` fix to develop first if `flake.nix` changed),
   - pushes the annotated tag `v0.1.0-rc1` on the hash-current commit (App token).
4. The tag push triggers `release.yml`: guard → nix-vendor-hash freshness →
   contract-gates → **publish a prerelease GitHub Release** with the `generated/`
   OpenAPI specs attached as assets.

`rcN` tags publish **prereleases**. A final `vX.Y.Z` additionally requires a
same-version `-rcN` that is green **and** an ancestor of the final commit
(`release-guard check-final`).

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

1. **Push this SLICE-B branch + open the PR** into `develop` on
   `peasant-labs/schema` (GitHub).
2. **Install the `peasant-labs-releaser` GitHub App** on `peasant-labs/schema`
   with Contents: write, and add the `PEASANT_RELEASER_APP_ID` /
   `PEASANT_RELEASER_APP_PRIVATE_KEY` repo secrets (§2).
3. **Cut `v0.1.0-rc1`** via the §3 ceremony — the live run is the MVP proof gate.
   A maintainer must open the `release(v0.1.0-rc1): …` PR, approve, and merge; the
   pipeline mints the tag and publishes the prerelease. **Do not push the
   `v0.1.0-rc1` tag manually** — the ceremony is the only sanctioned path, and
   the App-token tag push is what proves the installation works.
