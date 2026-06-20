# peasant-labs/schema

The **public API contract** for [Peasant](https://github.com/peasant-labs/peasant)
and the [village](https://github.com/peasant-labs) marketplace — a single,
contract-only Go **leaf module**.

It holds the wire/domain types, the versioned OpenAPI specs, fixtures, the
request validators, the village SQL migrations, and the codegen + release-guard
tooling. It has **no runtime**: no HTTP server, no WebSocket hub, no CLI product.
Both consumers (the peasant client and the village server) depend on it; it
depends on nothing first-party except `bestiary` (a sibling dependency).

Module path: `github.com/peasant-labs/schema`

## What's inside

| Path | What |
|------|------|
| `*.go` (root, `package schema`) | Domain + wire types, content-hashing, metadata, push/pull contract types, annotations, the embedded-spec accessors (`VillageAPISpecJSON`). |
| `versions.go` | Single source of truth for the versioned-spec semvers (`VillageAPIVersion`, `PeasantLocalAPIVersion`, `TypesVersion`). |
| `openapi/` | OpenAPI 3.1 spec builders (`BuildVillageAPISpec`, `BuildPeasantLocalAPISpec`, `BuildTypesSpec`) + the artifact generator. |
| `generated/` | Committed OpenAPI spec goldens (JSON + YAML) + the standalone PublishRequest JSON Schema. **Gate-checked**; regenerate with `go run ./cmd/schema-gen`. |
| `validate/` | `ValidatePublishRequest` — the JSON-Schema validator the village enforces on publish. |
| `migrations/village/` | Embedded village SQL migrations (`embed.FS`). |
| `external/`, `testdata/`, `fixtures.go` | Vendored external schemas + test fixtures. |
| `cmd/schema-gen/` | Regenerates `generated/` specs + Redoc/HTML docs. The contract gates' freshness/immutability/surface tests live here. |
| `cmd/release-guard/` + `internal/release/` | The release pipeline's title/tag grammar + final-release guard CLI (see CONTRIBUTING). |

## Dev dependencies & provisioning

Everything is provisioned through **Nix** — the flake `devShell` IS the
dev-dependency manifest (Go has no dev/prod dependency split). You do **not**
install Go or any tool globally.

Prerequisites:

- [Nix](https://nixos.org/download.html) with flakes enabled (`experimental-features = nix-command flakes`).
- [`direnv`](https://direnv.net/) with the `nix-direnv` integration (recommended).

### Option A — direnv (recommended)

```bash
cd <this-worktree>
direnv allow      # loads the flake devShell automatically on cd
```

The committed `.envrc` is just `use flake .`. On entry you'll see
`Go go1.26.3 dev shell`, and `go`, `gopls`, `golangci-lint`, `ast-grep`,
`actionlint`, etc. are on `PATH`.

### Option B — nix develop

```bash
nix develop          # drops you into the dev shell
go test -race ./...
```

### Build / test / regenerate

```bash
make check                     # the quality gate: gofmt + vet + release-workflow
                               #   guard + `go test -race ./...` (incl. leaf-audit +
                               #   oasdiff/go-apidiff synthetic-break tests)
make gates BASE_REF=origin/develop  # breaking-change gates vs a base ref
                               #   (oasdiff spec diff + go-apidiff + vacuum lint)
go run ./cmd/schema-gen        # regenerate generated/ specs + docs/api HTML
nix build                      # hermetic buildGoModule (cmd/schema-gen + cmd/release-guard)
```

`make check` is fully runnable with a bare `go` toolchain — the synthetic-break
tests `t.Skip` (with an actionable message) when a gate binary is absent, and run
for real inside `nix develop`. The committed contract goldens in `generated/`
must stay byte-identical to what `go run ./cmd/schema-gen` emits (the
codegen-freshness gate), and retired spec versions are byte-frozen under the
immutability gate. Run the generator and commit `generated/` after any change to
the Go schema source.

## The leaf rule

This module's `go.mod` `require` set is a **subset** of:
`bestiary`, `santhosh-tekuri/jsonschema/v5`, `swaggest/jsonschema-go`,
`swaggest/openapi-go`, `golang.org/x/crypto`, `gopkg.in/yaml.v3`.

Contract-gate CLIs (`oasdiff`, `go-apidiff`, `vacuum`) are provisioned in the
flake `devTools`, **never** in `go.mod`. See CONTRIBUTING for the leaf-audit and
the no-local-`replace` (`vendorHash`-stability) rules.

## Open items

- **LICENSE** — this public repo has no committed `LICENSE` yet; the flake `meta`
  carries an honest placeholder (`LicenseRef-Peasant-Schema-Placeholder`,
  `free = false`) until the final license is chosen. This is a user decision.
- **CI + release pipeline** — DONE (SLICE-B): contract gates (`oasdiff` /
  `go-apidiff` / `vacuum`, provisioned via the flake), `tests.yml`,
  `release-pr.yml` / `release.yml`, and `nix-vendor-hash.yml`. The release
  ceremony and the `v0.1.0-rc1` cut are documented in
  [`docs/release-runbook.md`](docs/release-runbook.md).
