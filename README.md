# peasant-labs/schema

The **canonical data / wire contract** for [Peasant](https://github.com/peasant-labs/peasant)
and the village transcript commons — one contract-only Go **leaf module** that every
backend produces and every client consumes.

```
module github.com/peasant-labs/schema
```

It holds the domain + wire types, the closed-set enums, the publish/pull/push
envelopes, the versioned OpenAPI specs, the publish-request validator, the typed
fixtures, and the codegen + release-guard tooling. It has **no runtime**: no HTTP
server, no WebSocket hub, no CLI product. Everything here is types, constants,
generated specs, and the gates that keep them honest.

---

## Why this module exists

The peasant backend, the village backend, and their frontends all speak the same
`session_detail` / publish / pull wire. If each side kept its own copy of those
types, the copies would drift — and a silent wire mismatch between a producer and a
consumer is exactly the class of bug that is expensive to find.

This module is the **single source of truth** for that wire. It was extracted from
peasant's old nested `pkg/schema` into a standalone, public, versioned Go module so
that:

- **One definition, many consumers.** The Go types here ARE the contract. Peasant
  emits `SessionDetailPayload` / `PublishRequest`; the village validates and stores
  the same shapes; the frontends render them. Nobody redefines the wire.
- **Served ≡ enforced.** The village serves its OpenAPI doc from `VillageAPISpecJSON()`
  and validates inbound publishes against `PublishRequestSchemaJSON()` — the *same*
  embedded bytes. The documented spec and the enforced schema are one artifact, so
  they can never diverge.
- **It stays a leaf.** The module depends on nothing first-party except `bestiary`
  (a sibling), imports no runtime machinery, and a leaf-audit gate keeps its
  dependency set pinned. That is what lets consumers pin it cheaply (see
  [Pinning](#how-pinning-works)).

---

## Coordinates & status

| | |
|---|---|
| Module path | `github.com/peasant-labs/schema` |
| Default branch | `develop` (releases are cut from it; `main`/tags carry the releases) |
| Latest tag | `v0.1.0-rc3` (GitHub **prerelease**) — `v0.1.0-rc1` and `v0.1.0-rc2` are also published prereleases |
| License | Apache-2.0 |
| Spec versions | Village API `0.4.0` · Peasant Local API `0.2.0` · Types `0.1.0` (see [`versions.go`](versions.go)) |

### Consumers

| Consumer | Language | How it pins / uses the contract |
|---|---|---|
| **peasant** (Go backend + web) | Go | `go.mod` requires `github.com/peasant-labs/schema@v0.1.0-rc3`; produces `SessionDetailPayload`, imports the enums, mirrors `AllLicenses` in its SQLite CHECKs. |
| **village** (Go backend) | Go | `backend/go.mod` requires `@v0.1.0-rc2`; serves `GET /openapi.json` from `VillageAPISpecJSON()` and **enforces** inbound publishes via `ValidatePublishRequest` (both read the embedded spec, so served ≡ enforced). |
| **transcript-browser** (`@peasant-labs/types`) | TypeScript | A hand-maintained TS port that has **drifted** from the Go — treat the Go here as authoritative when they disagree. The durable fix is OpenAPI→TS codegen off `generated/` (an open follow-up), which retires the hand port. |

> Village currently pins one tag behind peasant (`rc2` vs `rc3`). Because the module
> is normal `go get`-pinned, each consumer moves independently; a re-pin is a
> one-line `go.mod` change plus `go mod tidy`.

---

## What's in the contract

The Go source (all `package schema` in the repo root) is grouped by surface. Every
identifier and enum below is a wire type — a JSON-tagged struct or a closed-set
string newtype with `IsValid()` / `String()` / a `JSONSchema()` exposer and an
`All…` canonical list.

| Surface | File(s) | Representative types |
|---|---|---|
| **Local dashboard / session-detail wire** | `local_api.go` | `SessionDetailPayload`, `TurnDetail`, `ToolCallDetail`, `SessionSummary`, `SessionScorecard`, `ChildSessionRef`, `DashboardPayload`, `TrendsPayload`, `QualityPayload`, project-familiarity payloads, the WebSocket envelope (`ClientMessage` / `ServerMessage`, `MessageType`, `ChannelTopic`, `ChannelSubscription`), `CreateAnnotationRequest`/`Response` |
| **Session metadata & git** | `metadata.go` | `UnifiedMetadata`, `RedactionInfo`, `TimestampInfo`, `SourceInfo`, `CommitInfo`, `GitContext`, `ProjectContext`, `SessionStats`, `SubagentRef`, `DiagnosticsInfo`; the `MetadataSchemaVersion` constant |
| **Content layer** | `content.go`, `identity.go` | `SessionEntry` (1:1 with an indexed transcript entry), `SessionIdentity` |
| **Publish envelope** | `publish.go`, `publish_validate.go` | `PublishRequest`, `PublishResponse`, `ModelInfo`; the annotation push wire (`AnnotationPushItem` / `Request` / `Response`), `SchemaVersionResponse`; `ValidatePublishRequest` |
| **Pull envelope** | `pull.go` | `TranscriptID`, `PullTranscriptInfo`, `PullListResponse`, `PullAnnotation` |
| **Push content envelope** | `push_content.go` | `TranscriptContent` (self-describing, versioned blob body), `ContractVersion` / `PushContractVersion`, `ContentKind` |
| **Map / Review REST** | `map_api.go` | `MapGraphPayload`, `MapNode`, `MapEdge`, `ActivityEdge`, `EdgeViolation`, `MapNodeDetailPayload`, `TaskSummary`, `CommitRef`, `ProjectTasksPayload`, `ProjectSummary`, `ReviewListPayload`, `ChangeSummary`, `ChangeDetailPayload`, `ChangeDiffPayload`, `FrictionCluster`, `FileChange`, `DiffHunk` |
| **Search REST** | `search_api.go` | `SearchPayload`, `SearchResult` |
| **Quality metrics** | `quality.go` | `QualityMetrics`, `QualitySession` (session stats + derived quality/cost signals) |
| **Annotations** | `annotation.go`, `annotation_enums.go`, `annotation_manifest.go`, `annotation_registry.go`, `annotation_validate.go` | `AnnotationSummary`, `AnnotationTypeSummary`, `AnnotatorSummary`, `ValueDomain`, `Provenance`, `TaxonomyNode`, batch request/response types; the enums `AnnotatorKind`, `AnnotationStatus`, `TargetKind`, `ScaleKind`, `ValueDomainKind`, `AnnotationDatatype`, `TypeOrigin` |
| **Identifiers & enums** | `types.go` | Validated newtypes `SessionID`, `ModelID`, `ProjectHash`, `HostSlug` (+ `TranscriptID` in `pull.go`); closed enums `Role`, `EntryType`, `ToolCallKind`, `StopReason`, `SessionOutcome`, `SourceFormat`, `Visibility`, `License`, and `Harness` (re-exported from `bestiary`) |
| **Auth & commands** | `auth.go`, `command.go` | `ExchangeCodeRequest`/`Response`, `CLILoginQuery`; `BuiltinCommand` + `IsClaudeBuiltinCommand` |
| **Redaction fixtures** | `redactions.go` | `RedactionInfo` staleness helpers, `RedactionFixtureLevel`, the `RedactionExamples` corpus, `LoadRedactionExamples` (the redaction *engine* stays in peasant; this module carries only the metadata type and the fixture corpus) |
| **Embedded specs & fixtures** | `specs.go`, `contract_embeds.go`, `fixtures.go` | `VillageAPISpecJSON()`, `PublishRequestSchemaJSON()`, `EvalSchemaJSON`, `ContractCorpusFS`, and the embedded YAML/JSON fixture corpora |

### The session-detail payload (the central object)

`SessionDetailPayload` is the object every producer builds and every renderer
consumes. It travels two ways:

- On peasant's local **`session_detail` WebSocket channel**, and
- Wrapped in a **`TranscriptContent`** envelope as the versioned `peasant push` wire
  body (replacing raw provider JSONL with a normalized, self-describing blob).

It carries the session header (`Harness`, timing, token totals), the ordered
`[]TurnDetail` (each with its `[]ToolCallDetail`, ACP-aligned enrichment, and
optional thinking / stop-reason / token fields), an optional `SessionScorecard`,
and the resolved `SessionOutcome`.

### The License surface

The content-license menu is owned by this contract (`types.go`), so producers,
enforcers, and UIs all read one closed set:

| Symbol | Value / behaviour |
|---|---|
| `License` | Closed string newtype |
| `LicenseCC0` / `LicenseCCBY` / `LicenseCCBYSA` | `CC0-1.0` · `CC-BY-4.0` · `CC-BY-SA-4.0` |
| `AllLicenses` | The canonical ordered menu |
| `License.IsValid()` | Reports membership in the menu |
| `LicenseMenu()` | The menu as a comma-separated string for help text / errors (derived from `AllLicenses`, never a literal) |

The publish/pull wire carries it as an optional `license` field (`PublishRequest`,
`PullTranscriptInfo`); the generated JSON-Schema's `license` enum is derived from
`AllLicenses`, so widening the menu flows through the schema automatically. Village
enforces it; peasant mirrors it in two SQLite CHECK constraints.

### Generated OpenAPI specs (`generated/`)

`go run ./cmd/schema-gen` renders the Go source into OpenAPI 3.1 specs, committed as
byte-frozen goldens (JSON + YAML) plus the standalone PublishRequest JSON-Schema
(draft 2020-12). Three spec families are emitted, from the builders in `openapi/`:

| Spec family | Builder | Covers | Current version |
|---|---|---|---|
| **Village API** | `BuildVillageAPISpec` | publish / pull / annotations / auth / schema-version | `0.4.0` |
| **Peasant Local API** | `BuildPeasantLocalAPISpec` | the local dashboard REST + Map / Review / Search surface | `0.2.0` |
| **Types** | `BuildTypesSpec` | the foundational shared domain-type catalog | `0.1.0` |

The current specs are read back into the binary via `//go:embed generated`. Two
version-aware accessors expose the bytes so consumers follow the `go.mod` pin
without vendoring their own copy:

- **`VillageAPISpecJSON()`** — the current Village API spec (village serves it as
  `GET /openapi.json`).
- **`PublishRequestSchemaJSON()`** — the current PublishRequest JSON-Schema (the
  single byte-source `ValidatePublishRequest` compiles and the village enforces
  through).

Both derive their filename from the version constant in `versions.go`, so a version
bump re-points them in lockstep — consumers need no change.

---

## Versioned specs & frozen goldens

Released spec versions are **immutable**. The rule is **bump, don't mutate**: you
never add or change surface on an already-shipped spec in place. You bump the
relevant version constant in [`versions.go`](versions.go) (minor for additive
surface, major for a breaking change), regenerate, and commit — the new version is
emitted while every prior version stays byte-frozen exactly as shipped.

Two gates enforce this (both run in `make check` via `cmd/schema-gen`):

- **Codegen-freshness gate** — the committed artifacts under `generated/` (and the
  generated redaction fixture) must be **byte-identical** to what the generator
  emits from the Go source. A stale or hand-edited current spec fails the build;
  the fix is `go run ./cmd/schema-gen` + commit. (`make freshness` is the git-diff
  backstop for the same invariant.)
- **Retired-versions immutability guard** (`cmd/schema-gen/retired_specs_test.go`) —
  every retired version is content-hash-pinned (sha256 of its committed bytes) in a
  registry and asserted present-and-frozen. It fails loudly on **both** an in-place
  edit (hash mismatch) *and* a deletion (missing file) — the exact gap the freshness
  gate can't see, because freshness only diffs versions the generator *still* emits.
  A version is moved into the registry **at the moment it is frozen** (the same
  change that bumps the live constant past it), so there is never a window where a
  retired spec is mutable-and-unguarded. A permanent negative-control self-test
  proves the guard actually fires.

Currently frozen (retired) goldens: Village API `0.1.0` / `0.2.0` / `0.3.0`,
PublishRequest schema `0.2.0` / `0.3.0`, and Peasant Local API `0.1.0`. The
still-generated current versions (Village API `0.4.0`, PublishRequest `0.4.0`,
Peasant Local API `0.2.0`, Types `0.1.0`) live under the freshness gate instead.

The versioning procedure itself is codified in the `versions.go` doc comments, the
"Regeneration & gates" section of [`CONTRIBUTING.md`](CONTRIBUTING.md), and the
release ceremony in [`docs/release-runbook.md`](docs/release-runbook.md); there is
no separate versioning document.

---

## How pinning works

Because the contract is a normal published, tagged Go module, a consumer pins **one
version**:

```bash
go get github.com/peasant-labs/schema@v0.1.0-rc3
```

That single pin replaces the old model of vendoring an in-tree copy of the types. It
also structurally retires the cross-repo cost that model carried: there is no
`vendorHash` / private-module-auth tax to fold first-party schema source into a
consumer's build. The Nix `buildGoModule` `vendorHash` covers only the third-party
module graph, and a **no-local-`replace` rule** (proven by
`TestVendorHashStableOnFirstPartyEdit`) forbids the local-path `replace` directive
that would fold first-party source into the vendor graph — so a schema edit can
never drift a consumer's hash.

**The ceremony (in effect):** a contract change lands as **its own schema-repo PR +
tag first**, and only *then* do the consumer PRs that re-pin to the new tag land.
The wire is defined once, published once, and adopted deliberately — never edited
in two repos at once. Because the village both serves and enforces from the same
pinned module bytes, its served spec and its enforcement can't drift from each other
or from the tag.

---

## Contract gates

| Gate | Tool | Consequence |
|---|---|---|
| OpenAPI breaking diff | `oasdiff breaking --fail-on ERR` | **Hard** — fails on an ERR-level breaking change vs the prior golden. |
| OpenAPI lint | `vacuum lint -r .vacuum.yaml` | **Hard** — fails on `error`-severity findings in the generated specs. |
| Exported-Go-API diff | `go-apidiff` | **Advisory while pre-1.0** — an incompatible exported-Go-API change surfaces a workflow warning + a single sticky PR comment listing the changed/removed symbols (and any additive changes), then exits success. Detection is unchanged; only the consequence is. An unrecognizable diff-tool output still fails closed. |

The advisory posting runs in a write-scoped companion workflow
(`go-apidiff-comment.yml`) so the read-only Tests workflow can be reused as a gate.
Alongside these, `make check` runs the codegen-freshness gate, the retired-versions
immutability guard, the leaf-audit, the `vendorHash`-stability proof, and the
synthetic-break tests that prove the `oasdiff` / `go-apidiff` gates actually fire.

---

## Develop, build, regenerate

Everything is provisioned through **Nix** — the flake `devShell` *is* the
dev-dependency manifest (Go has no dev/prod dependency split). You do not install Go
or any tool globally.

Prerequisites: [Nix](https://nixos.org/download.html) with flakes enabled, and
(recommended) [`direnv`](https://direnv.net/) with `nix-direnv`.

```bash
# Option A — direnv (recommended)
cd <this-worktree>
direnv allow          # loads the flake devShell on cd; the committed .envrc is `use flake .`

# Option B — nix develop
nix develop           # drops you into the Go 1.26 dev shell
```

Inside the shell (`go`, `gopls`, `golangci-lint`, `oasdiff`, `go-apidiff`, `vacuum`,
`actionlint`, … on `PATH`):

```bash
make check                            # quality gate: gofmt + vet + release-workflow guard + `go test -race ./...`
                                      #   (incl. leaf-audit, freshness, immutability, and the synthetic-break tests)
make gates BASE_REF=origin/develop    # breaking-change gates vs a base ref (oasdiff + go-apidiff + vacuum)
go run ./cmd/schema-gen               # regenerate generated/ specs (+ Redoc HTML); commit the result
nix build                             # hermetic buildGoModule (cmd/schema-gen + cmd/release-guard)
```

`make check` is fully runnable with a bare `go` toolchain — the gate-binary tests
`t.Skip` with an actionable message when a tool is absent and run for real inside
`nix develop`. After any change to the Go schema source, regenerate and commit
`generated/` (the freshness gate enforces byte-identity).

### The leaf rule

The module's `go.mod` direct `require` set must stay a **subset** of the allowed
contract dependencies — enforced by `TestLeafAudit_GoModRequiresAreAllowed`:

```
github.com/dayvidpham/bestiary
github.com/google/go-github/v88          # runtime seam for cmd/release-guard's GitHub API, not a dev tool
github.com/santhosh-tekuri/jsonschema/v5
github.com/swaggest/jsonschema-go
github.com/swaggest/openapi-go
golang.org/x/crypto
gopkg.in/yaml.v3
```

Dev/CI tools (`oasdiff`, `go-apidiff`, `vacuum`, linters) live in the flake
`devTools`, **never** in `go.mod`. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the
leaf-audit and no-local-`replace` (`vendorHash`-stability) rules and the `go.work`
workflow for co-developing consumers against a local worktree before a version is
cut.

---

## Releasing

Releases are cut by **merging a PR titled `release(vX.Y.Z[-rcN]): <summary>` into
`develop`** — never by hand-tagging. The title/tag grammar and final-release guard
live once in `internal/release`, exposed through **`cmd/release-guard`**, the single
canonical release-gating tool (built on `go-github`; it also serves peasant via a
per-repo `.github/release-guard.policy.yml`). On merge, the release workflow mints
the annotated tag and publishes a GitHub Release (prerelease for `-rcN`) with the
`generated/` OpenAPI specs attached as assets.

The maintainer-approval assertion in the release PR is **deferred to the public
flip** (a single active maintainer plus GitHub's no-self-approval rule make it
unsatisfiable today); the `release-guard check-approval` guard code stays live and
tested, re-enabled alongside branch protection at the flip. The full operator guide —
secrets, the GitHub App, the ceremony, and the public-flip checklist — is in
[`docs/release-runbook.md`](docs/release-runbook.md).

---

## Repository layout

| Path | What |
|---|---|
| `*.go` (root, `package schema`) | The contract: domain + wire types, enums, content-hashing, metadata, publish/pull/push envelopes, annotations, the embedded-spec accessors |
| `versions.go` | Single source of truth for the versioned-spec semvers |
| `openapi/` | OpenAPI 3.1 spec builders (`BuildVillageAPISpec`, `BuildPeasantLocalAPISpec`, `BuildTypesSpec`) |
| `generated/` | Committed OpenAPI spec goldens (JSON + YAML) + the PublishRequest JSON-Schema — gate-checked; regenerate with `go run ./cmd/schema-gen` |
| `external/`, `testdata/`, `fixtures.go` | Vendored external schemas + typed test fixtures |
| `cmd/schema-gen/` | Regenerates `generated/`; hosts the freshness / immutability / surface gate tests |
| `cmd/release-guard/` + `internal/release/` | The PR-title/tag grammar + release-gating CLI |
| `docs/release-runbook.md` · `CONTRIBUTING.md` · `CHANGELOG.md` | Operator + contributor references |

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
