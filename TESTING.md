# Testing the peasant-labs/schema contract

This module is a contract-only leaf: types, closed enums, generated OpenAPI specs,
a publish-request validator, typed fixtures, and the codegen + release-guard
tooling. It has no runtime server. So the test suite is not about request/response
behaviour; it is about keeping the **contract honest**:

- the committed generated specs stay byte-identical to the Go source (no drift);
- a released spec version can never be silently mutated or deleted;
- a breaking change to the wire is caught before it ships;
- the module stays a dependency leaf (dev/CI tools never leak into `go.mod`);
- the release grammar and the publication gate behave as documented.

Everything is provisioned through the Nix flake dev shell. The whole suite is
runnable with a **bare `go` toolchain**: the three gate binaries (`oasdiff`,
`go-apidiff`, `vacuum`) are not `go.mod` requires, so the tests that drive them
`t.Skip` with an actionable message when the tool is absent, and run for real
inside `nix develop` (or after `direnv allow`).

## Documentation map

| Read this | For |
|---|---|
| [`README.md`](README.md) (Contract gates + Develop, build, regenerate) | The quickstart: the gate table and how to run `make check` / `make gates`. |
| **`TESTING.md`** (this file) | The comprehensive test strategy: what each layer protects and the real test behind it. |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | The regeneration + gates workflow, the leaf-audit rule, and the `go.work` flow for co-developing a consumer. |
| [`docs/release-runbook.md`](docs/release-runbook.md) | The release ceremony and how the gates gate a publish. |

## Running the tests

```bash
make check                          # the authoritative quality gate (bare-go runnable)
make test                           # go test -race ./...
make gates BASE_REF=origin/develop  # the breaking-change contract gates (needs the flake tools)
make freshness                      # git-diff backstop on the generated artifacts
pnpm --dir typescript run typecheck
pnpm --dir typescript test
pnpm --dir typescript run package:audit
pnpm --dir typescript run package:smoke
```

`make check` is `fmt` + `vet` + `freshness` + the release-workflow guard
(`release-guard check-workflow`) + `go test -race ./...`. The race suite is the
authoritative run: it includes the leaf audit, the vendor-hash guard, and the
`oasdiff` / `go-apidiff` synthetic-break tests (which skip when their binary is
absent).

Run one test or one package:

```bash
go test -race ./cmd/schema-gen/ -run TestRetiredSpecsImmutable -v
go test -race ./internal/contractgates/ -run TestOasdiffSyntheticBreak -v   # skips without oasdiff
```

### Bare `go` vs `nix develop`

- **Bare `go test ./...`** runs everything except the tool-gated synthetic-break
  tests, which `t.Skip` via `skipIfMissing` (see
  `internal/contractgates/synthetic_break_test.go`).
- **Inside `nix develop`** (or `direnv allow`) `oasdiff` / `go-apidiff` /
  `vacuum` are on `PATH`, so the synthetic-break tests and `make gates` run for
  real. CI runs `make gates` with `fetch-depth: 0` so the base ref is resolvable.

## Layer 1: the contract gates

These are the load-bearing invariants. Each has a real, named test (or gate
script) behind it.

| Gate | Where it lives | Consequence |
|---|---|---|
| Codegen freshness | `cmd/schema-gen/freshness_test.go`, `cmd/schema-gen/redactions_freshness_test.go`, `cmd/schema-gen/generated_accounted_test.go`; `make freshness` | **Hard.** A committed spec that drifts from the Go source fails the build. |
| Retired-version immutability | `cmd/schema-gen/retired_specs_test.go` | **Hard.** A frozen spec that is edited or deleted fails. |
| OpenAPI breaking diff (`oasdiff`) | `scripts/contract-gates.sh` (`make gates`); `internal/contractgates/synthetic_break_test.go` | **Hard.** An ERR-level break vs the prior golden fails. |
| OpenAPI lint (`vacuum`) | `scripts/contract-gates.sh` + `.vacuum.yaml` | **Hard.** An `error`-severity finding in a generated spec fails. |
| Exported-Go-API diff (`go-apidiff`) | `scripts/contract-gates.sh`; `internal/contractgates/synthetic_break_test.go` | **Advisory (pre-1.0).** Incompatible changes warn; an unparseable diff fails closed. |
| Leaf-purity audit | `leaf_audit_test.go` | **Hard.** A `go.mod` direct require outside the allowed set fails. |
| Vendor-hash stability | `vendorhash_test.go` | **Hard.** A local-path `replace` in `go.mod` fails. |
| Release grammar + guard | `internal/release/*_test.go`, `cmd/release-guard/*_test.go` | **Hard.** A malformed release title/tag, or a publish (GitHub Release or npm) behind an un-gated workflow, is rejected. |
| License menu exhaustive coverage | `licensecorpus/licensecorpus_test.go` (`TestLicenseCorpus_ExhaustiveCoverage`) | **Hard.** Widening `schema.AllLicenses` without regenerating the corpus fails (a menu member with no case). |
| License corpus regen-freshness | `licensecorpus/licensecorpus_test.go` (`TestLicenseCorpus_Freshness`) | **Hard.** A committed `license_corpus.yaml` that drifts from a fresh `RenderCorpus` (hand-edit or stale) fails. |
| Local API 0.5.0 closed-set exhaustive coverage + regen-freshness | `enumcorpus/enumcorpus_test.go` (`TestEnumCorpora_ExhaustiveCoverageAndFreshness`, one subtest pair per enum) | **Hard.** `enumcorpus.BuildCorpus`/`RenderCorpus` generalize `licensecorpus`'s pattern over any closed string enum; `cmd/gen-enum-corpora` regenerates all ten committed corpora (`AssociationConclusion`, `AssociationEvidenceKind`, `Confidence`, `RewriteResolution`, `RewriteMethod`, `InsightKind`, `InsightProvenance`, `ReadAttributionState`, `ReadStateGrade`, `TargetKind`) in one `go generate`. |
| `ReadStateGrade` registry-seed cross-check | `enumcorpus/enumcorpus_test.go` (`TestReadStateGradeRegistrySeedCrossCheck`) | **Hard.** `schema.ReadStateGradeRegistrySeedPermissibleValues` must byte-equal `AllReadStateGrades` minus `none`; the peasant-side read-state registry seed pins the other half against this exported value. |
| TypeScript closed-set completeness | `openapi_enums_test.go`; `testdata/typescript/enums.yaml` | **Hard.** Every canonical Go closed set is an OpenAPI enum before TypeScript generation can run. |
| TypeScript generated-file freshness | `make freshness` | **Hard.** Hey API/Zod root output, `openapi-typescript` operation contracts, and YAML-derived fixture data are byte-stable after regeneration. |
| TypeScript typecheck + fixture tests | `typescript/tsconfig*.json`, `typescript/tests/` | **Hard.** Public types compile and both languages accept/reject the same strict YAML matrix. |
| TypeScript operation-alias collision safety | `typescript/scripts/lib/operation-aliases.mjs`; `typescript/tests/operation-aliases.test.mjs`; `testdata/typescript/collision_cases.yaml` | **Hard.** A Local/Village API component whose schema drifts from the canonical root type it would alias to stops generation with an actionable error; a synthetic-break test proves the throw fires on a real mismatch, not only on today's already-consistent specs. |
| TypeScript ProjectHash brand-resolver safety | `typescript/scripts/lib/project-hash-resolver.mjs`; `typescript/tests/project-hash-resolver.test.mjs`; `testdata/typescript/project_hash_resolver_cases.yaml` | **Hard.** The generation-time decision that brands a string schema as `ProjectHash` requires both the canonical pattern AND the canonical `#/components/schemas/ProjectHash` `$ref` location; a 6-case fixture (1 positive, 5 negative-control) drives the extracted decision function directly, proving a same-pattern differently-named or differently-located schema is never branded. |
| TypeScript public contract completeness | `typescript/scripts/generate-contract-support.mjs` (`renderPublicContract`) | **Hard.** Every Types OpenAPI catalog component must produce exactly one matching Hey API Zod export; a missing, extra, or renamed export fails generation instead of only asserting the facade is non-empty. |
| TypeScript facade export identity | `typescript/tests/public-exports.test.mjs`; `testdata/typescript/public_exports.yaml`, `public_export_mutations.yaml` | **Hard.** The hand-maintained root/`local-api`/`village-api` facade exports (aliases, version constants, ProjectHash functions, forbidden names) match a fixture; a dedicated add/remove/duplicate/redirect mutation corpus proves the check is not vacuous. |
| TypeScript ProjectHash wire-location coverage | `typescript/tests/project-hash-locations.ts`, `project-hash-locations.test.mjs`; `testdata/typescript/project_hash_locations.yaml` | **Hard.** Compile-time proof that the ProjectHash brand reaches all 5 named wire locations plus a same-spelling negative control; a runtime coupling test fails if a location is dropped from either the fixture or the compile-time file (removing a passing assertion is not by itself a type error). |
| Published package content + tarball imports | `typescript/scripts/package-*.mjs` | **Hard.** Only audited files ship, and every public subpath imports from a disposable packed install. |

### Codegen freshness

`TestCodegenFreshness_SpecsMatchSource` regenerates the OpenAPI specs from the Go
source via the **same** generator `cmd/schema-gen`'s `main()` uses
(`openapi.GenerateSpecArtifacts()`, not a mock) and asserts every committed
`generated/*.json` and `*.yaml` is byte-identical. It fails on a hand-edited spec
or on a Go schema change that was not regenerated and committed. The fix it names
is always `go run ./cmd/schema-gen` + commit. `TestCodegenFreshness_UnifiedHarnessKey`
additionally asserts the generated specs carry the unified `harness` wire key (not
the legacy `modelHarness`), so a source revert fails loudly with a clear cause.
`redactions_freshness_test.go` applies the same byte-identity rule to the generated
redaction fixture, and `TestGeneratedDirFullyAccounted` asserts every file in
`generated/` is one the generator emits (no orphan artifact).

`make freshness` is the git-diff backstop for the same invariant. It is largely
redundant with `TestCodegenFreshness` (both diff the generator's output), but it
uses `git add -N` so it also catches a future generator write-path that emits a
tracked artifact outside the shared artifact map, which the Go test (iterating that
map) could not see.

`make freshness` then regenerates the TypeScript contract with exact tool
versions pinned in `typescript/pnpm-lock.yaml`. Hey API's Zod plugin consumes the
canonical Types OpenAPI catalog without enabling an SDK/client plugin;
`openapi-typescript` emits type-only `paths` and `operations` contracts from the
Local and Village catalogs. The gate diffs those outputs, the Go-shaped enum
facade, version constants, and YAML-derived quality and timeline fixture data.

### Cross-language testcase and fixture gates

`testcase/testdata/load_cases.yaml` is one strict-loader matrix consumed by both
Go and TypeScript. It covers unknown envelope keys, duplicate YAML keys,
malformed input, trailing documents, invalid closed values, and duplicate case
names. The TypeScript generic loader requires `decodeInput` and
`decodeExpected` callbacks, because erased generic parameters cannot safely
decode `unknown`; the package never casts an unknown payload to caller-selected
`I` or `E`.

Quality fixture consumers do not parse repository-relative YAML. The TypeScript
generator reads the same canonical YAML as Go, maps its public field names, and
emits typed data behind clone-returning accessors. The Go loader remains the
strict source validation gate. TypeScript tests cover the five sessions, the
named set, and the full variation catalog.

The project timeline corpus uses the same schema-owned path. Each row in
`timeline.yaml` carries its stable family identity, and `LoadTimelineFixtures`
validates exactly 24 families with 7 accepted and 17 rejected relationship
cases. A separate schema-repo-only oracle and count-preserving rename and
replacement mutations prove the public corpus has the exact intended identities;
that review scaffolding is not generated or published. Generation emits only the
typed corpus and clone-returning `/fixtures/timeline` accessor, so TypeScript
consumers never reparse repository YAML or redefine the session-to-commit
relationship contract.

### TypeScript facade and operation-alias gates

A gate that can never fail is worthless here too. `canonicalOperationAliases`
(`typescript/scripts/lib/operation-aliases.mjs`) structurally compares every
Local/Village API OpenAPI component against the canonical root Types schema its
name normalizes to, and throws when they diverge; it happens to never throw
against today's already-consistent generated specs, which is not evidence it
would fire on a real mismatch. `typescript/tests/operation-aliases.test.mjs`
proves it: a synthetic-break case constructs a same-named root/API schema pair
that deliberately disagree and asserts generation throws with the expected
message, alongside `testdata/typescript/collision_cases.yaml`'s fixture-backed
equal/unequal and dropped-required-field/changed-property-type corpus.

The ProjectHash brand carries the same "never observed to fail" risk one layer
earlier, at generation time: `typescript/openapi-ts.config.mjs`'s Hey API
Zod-plugin resolver decides whether a string schema receives the nominal
`ProjectHash` brand. That decision (`shouldBrandProjectHash` /
`isCanonicalProjectHashSchema`, extracted to
`typescript/scripts/lib/project-hash-resolver.mjs` the same way
`canonicalOperationAliases` was) requires both the canonical
`^[0-9a-f]{64}$` pattern AND the canonical `#/components/schemas/ProjectHash`
`$ref` location; today's Types catalog has only one schema with that pattern,
so generation alone can never observe the guard reject anything.
`typescript/tests/project-hash-resolver.test.mjs` is the committed proof
instead: `testdata/typescript/project_hash_resolver_cases.yaml`'s 6 cases drive
the extracted function directly against synthetic path/pattern pairs, covering
the canonical case plus 5 ways it must fail closed (a same-pattern
differently-named component, the canonical path with the wrong pattern, a
nested non-top-level path, a path outside `components/schemas`, and a missing
path reference). This is a different layer from the ProjectHash
*wire-location* coverage described below: this gate proves the brand is
assigned correctly at the one place it originates; that one proves the
assigned brand then propagates to every wire location that should carry it.

The hand-maintained root, `/local-api`, and `/village-api` facade files
(`typescript/src/index.ts`, `local-api.ts`, `village-api.ts`) are five lines
each and carry no compiler-visible contract of their own: dropping one of their
re-export lines passes typecheck, `pnpm test`, `package:audit`, `package:smoke`,
and `make freshness` unless something specifically asserts the export exists.
`typescript/tests/public-exports.test.mjs` is that assertion: it checks runtime
constants and functions by value, type-only aliases by source-text wiring, and
runs `testdata/typescript/public_export_mutations.yaml`'s remove/add/duplicate/
redirect corpus through a small identity validator to prove the check is not
vacuous. `typescript/tests/project-hash-locations.ts` is the equivalent
compile-time proof for the ProjectHash brand: it is a `Same<>` assertion per
wire location, and because removing a passing assertion typechecks fine on its
own, `project-hash-locations.test.mjs` cross-checks the fixture's named
locations against the `.ts` file's source text so a dropped location fails
loudly instead of silently.

### Retired-version immutability

Released spec versions are immutable: the rule is bump, don't mutate. Once a
version is superseded, its generated goldens are frozen in place.
`TestRetiredSpecsImmutable` (`cmd/schema-gen/retired_specs_test.go`) pins each
retired artifact to the `sha256` of its committed bytes in a registry, so it fails
loudly on **both** an in-place edit (hash mismatch) and a deletion (missing file):
the exact gap the freshness gate cannot see, because freshness only diffs versions
the generator still emits. A version is moved into the registry at the moment it is
frozen (the same change that bumps the live version past it), so there is never a
window where a retired spec is mutable and unguarded. The still-generated current
versions are deliberately excluded here; they live under the freshness gate
instead (`versions.go` stays the single source of truth for which version is
current). `TestCheckFrozen_NegativeControl` is a permanent
negative control that proves the guard actually fires.

### The OpenAPI + Go-API gates (and their meta-gates)

`scripts/contract-gates.sh` runs three gates against a base ref (`make gates
BASE_REF=origin/develop`):

- **`vacuum`** lints every generated OpenAPI spec (the `*-api-*.yaml` surfaces and
  the `types-*.yaml` catalog; the publish-request JSON Schema is not an OpenAPI doc
  and is excluded) against `.vacuum.yaml`, failing on `error` severity. Hermetic:
  no git history needed.
- **`oasdiff`** breaking-diffs each generated spec against the same file on the base
  ref and fails on an ERR-level break. A version file that is new on the branch
  (absent at the base) has nothing to diff and is skipped.
- **`go-apidiff`** diffs the exported Go API vs the base ref under a pre-1.0
  **advisory** policy: an incompatible change surfaces a warning (and, in CI, a
  sticky PR comment) but does not fail the gate; an intentional spec-version stamp
  bump (`VillageAPIVersion` / `PeasantLocalAPIVersion` / `TypesVersion`) is exempt;
  and an `Incompatible changes` section that cannot be parsed **fails closed** (a
  blind gate must stop the line).

A gate that can never fail is worthless, so the gates have their own tests.
`internal/contractgates/synthetic_break_test.go` proves they fire:
`TestOasdiffSyntheticBreak` removes an endpoint from a committed golden spec and
asserts `oasdiff breaking --fail-on ERR` exits non-zero, while
`TestOasdiffNoBreakOnIdenticalSpec` pins the no-false-positive side. Three of the
six `TestGoAPIDiff*` tests run the real `go-apidiff` binary end-to-end:
`TestGoAPIDiffSyntheticBreak` (removes an exported func, asserts an incompatible
change is reported), `TestGoAPIDiffStampExemption`, and `TestGoAPIDiffGateRunner`.
The other three drive the decision seams (`evaluate-apidiff` /
`extract-compatible`) with canned real-format output:
`TestGoAPIDiffStampExemptionFilter` (the stamp-bump exemption cannot mask a real
break), `TestGoAPIDiffCompatibleExtraction`, and
`TestGoAPIDiffGateRunnerFailClosed` (the unparseable-section fail-closed path).

### Leaf-purity audit and vendor-hash stability

`TestLeafAudit_GoModRequiresAreAllowed` (`leaf_audit_test.go`) asserts the module's
`go.mod` direct `require` set is a subset of an explicit allowed set (`bestiary`,
`go-github/v88`, the two `jsonschema` libraries, `openapi-go`, `x/crypto`,
`yaml.v3`). Dev and CI tools live in the flake dev shell, never in `go.mod`, so
this audit is the only mechanical enforcement of the leaf boundary. Its
`parseDirectRequires` helper parses `go.mod` by hand on purpose: importing a
module-file parser would itself add a require and violate the invariant it guards.
`TestLeafAudit_NegativeFixtureFails` proves the audit rejects a disallowed require.

`TestVendorHashStableOnFirstPartyEdit` (`vendorhash_test.go`) guards the Nix
`buildGoModule` `vendorHash`. That hash covers only the third-party module graph
(`go.mod` + `go.sum`), so a first-party edit (a testdata YAML, a `.go` file) can
never drift it: the single way first-party source leaks into the vendor
computation is a local-path `replace` directive (`=> ./...` or `=> ../...`). The
test fails if `go.mod` ever grows one. Hermetic (no Nix needed); it is the durable
guard behind the flake's `vendorHash` comment.

### Release grammar and the publication guard

`internal/release/*_test.go` cover the release title/tag grammar and the
publication guard (`parse_test.go`, `guard_test.go`, `permission_test.go`,
`review_test.go`, `workflow_guard_test.go`, `workflow_run_test.go`), and
`cmd/release-guard/*_test.go` cover the CLI that exposes them (`main_test.go`, plus
the `go-github` and git seams in `githubclient_test.go` / `gitrunner_test.go`).
`release-guard check-workflow` runs inside `make check` to keep publication behind
the required gates: the per-repo policy (`.github/release-guard.policy.yml`)
requires both the `release` (GitHub Release) and `npm-publish` jobs to sit behind
`guard` + `nix-vendor-hash` + `contract-gates`, and additionally requires
`npm-publish` to carry `permissions.id-token: write` and run under the
`npm-publish` GitHub Actions environment (the two scopes its npm Trusted
Publishing / OIDC authentication needs - see `docs/release-runbook.md`).
`TestCheckReleaseWorkflow_Cases` (`workflow_guard_test.go`, fed by
`testdata/workflow/cases.yaml`) includes synthetic-break fixtures proving each
assertion fires for BOTH the "missing entirely" and the "present but wrong
value" mutation of the `id-token` permission and the environment binding -
`missing-npm-publish.yml` and `npm-publish-missing-needs.yml` cover the job and
its needs-edges; `npm-publish-missing-permissions.yml` /
`npm-publish-wrong-permissions.yml` cover the `id-token` permission absent vs.
present-but-wrong (`read` instead of `write`); `npm-publish-missing-environment.yml`
/ `npm-publish-wrong-environment.yml` cover the environment binding absent vs.
present-but-wrong (`production` instead of `npm-publish`) - so the `npm-publish`
job, one of its needs-edges, its `id-token` permission, and its environment
binding are each proven caught whether missing OR wrong, not just one mutation
shape per field.

The permission/environment checks only parse the exact forms this repo's
workflows use (the explicit `permissions:` map, a bare scalar `environment:`),
not GitHub's `permissions: write-all`/`read-all` shorthand or the
`environment: { name, url }` mapping form; both gaps fail CLOSED (a false
rejection, never a false accept). `npm-publish-permissions-shorthand.yml` and
`npm-publish-environment-mapping.yml` prove each unsupported-but-valid form is
rejected with a message naming the specific form and the fix, distinct from
the "missing entirely" message.

## Layer 2: unit tests alongside the sources

Every source file has its `*_test.go` sibling. These assert the concrete contract
shapes: the closed enums (`IsValid()` / `String()` / the canonical `All...` list),
the JSON tags and struct shapes, and the validator behaviour.

- **Domain and wire types:** `annotation_test.go`, `annotation_enums_test.go`,
  `annotation_validate_test.go`, `annotation_manifest_test.go`,
  `annotation_push_test.go`, `pull_test.go`, `push_content_test.go`,
  `commit_test.go`, `command_test.go`, `publish_validate_test.go`, `types_test.go`,
  `schema_test.go`, `map_api_test.go`, `search_api_test.go`, `specs_test.go`.
- **OpenAPI builders:** `openapi/*_test.go` covers the village, peasant-local, and
  types spec builders, the harness enum, the selective-required arrays, the
  publish-schema parity check, and the example objects.
- **Generator and release tooling:** `cmd/schema-gen/*_test.go`,
  `internal/release/*_test.go`, `cmd/release-guard/*_test.go`.

## Layer 3: fixture families and the fixture idiom

Typed test corpora live in `testdata/` trees, one family per surface:

- top level `testdata/`: `annotations`, `contract`, `publish`, `pull`, `quality`,
  `session-detail`, `sync`;
- `internal/release/testdata/`: `grammar`, `workflow`;
- `cmd/release-guard/testdata/`: `github`;
- `openapi/testdata/`.

### The idiom: typed struct + `//go:embed` + `yaml.Unmarshal`

A fixture family is a YAML corpus loaded through a typed struct. The loader lives in
a `_test.go` file so the embedded corpus compiles **only** into the test binary,
never into the production tool. The canonical shape (from
`internal/release/grammar_fixtures_test.go`):

```go
//go:embed testdata/grammar/versions.yaml
var grammarFixtureYAML []byte

type newVersionCase struct {
    Name    string          `yaml:"name"`
    Raw     string          `yaml:"raw"`
    Want    release.Version `yaml:"want"`
    WantErr bool            `yaml:"wantErr"`
}

func loadGrammarFixtures(t *testing.T) grammarFixtures {
    t.Helper()
    var f grammarFixtures
    if err := yaml.Unmarshal(grammarFixtureYAML, &f); err != nil {
        t.Fatalf("load grammar fixtures (testdata/grammar/versions.yaml): %v", err)
    }
    return f
}
```

Other real examples of the same idiom: `internal/release/workflow_fixtures_test.go`
(`//go:embed testdata/workflow`), `cmd/release-guard/githubfixtures_test.go`
(`//go:embed testdata/github`), and `openapi/publish_schema_parity_test.go`
(`//go:embed testdata/legacy-publish-request.schema.json`).

### Row-count guards

Fixture-backed tests assert a **minimum** corpus size so a silently gutted or
half-loaded family fails loudly instead of passing vacuously. For example
`testdata_publish_test.go` requires the publish verdict corpus to be non-empty and
at least 17 cases; `testdata_pull_test.go` requires at least 15 valid and 9 invalid
transcript-id cases across its two generation axes; `testdata_quality_test.go`
pins its session count. Add rows to the YAML corpus, not new inline literals, when
a contract case is added.

### The canonical corpus standard: `testcase`

The idiom above is ad hoc: each family hand-rolls its own row struct and loader.
`github.com/peasant-labs/schema/testcase` promotes it to a single canonical form
that new corpora adopt. It is a generic, pure-data corpus model whose closed-set
metadata keeps every case both traceable and non-vacuous.

- **Generic `Case[I, E]` / `Corpus[I, E]`.** A case is a named `Input` of type
  `I` with its `Expected` output of type `E`, plus a `Classification`, a
  `Provenance`, and a `Mutation`. The caller instantiates `I`/`E` at load time,
  so one model serves the `schema.License`-in/bool-out enum corpus and the
  `string`-in/struct-out grammar corpora alike (both real consumers below).
- **Closed-set `Classification`.** `must-pass` (an input the system under test
  must accept) or `must-fail` (one it must reject); a value outside the set fails
  validation.
- **`Provenance{source, ref}` + `Mutation{description}`.** Every case records WHY
  it exists and WHAT single change it embodies. `source` is a closed set
  (`requirement`, `bug`, `enum`, `boundary`, `manual`) and `ref` is a concrete
  pointer (a requirement id, a bug link, an enum name); `description` names the
  one change under test. For a must-fail case that change is the mutation that
  makes a valid input invalid, so a negative case is never vacuous.
- **Pure loader + pure validators.** `LoadCorpus[I, E]` unmarshals and validates
  the YAML, returning an error rather than failing a test. `Case.Validate` /
  `Corpus.Validate` remain available for programmatically assembled corpora and
  reject a vacuous case: an out-of-set classification or provenance source, an
  empty `ref`, or an empty mutation `description`. `CheckMin(n)` is the pure
  minimum-size floor (`len >= n`), so a corpus may grow without tripping it.
- **The leaf-purity split.** `testcase` carries no `testing` import: it is pure
  data, so it never drags `testing` into anything that consumes it. The loud
  `*testing.T` seams `RequireMin` and `RequireValid` live in the sibling
  `testcase/assert` subpackage and wrap the pure `CheckMin` and `Corpus.Validate`,
  keeping the size and validity logic in pure functions a negative-control test
  can drive without a `*testing.T`. This promotes the module's existing
  file-level testing-helper isolation to a package boundary.

The enum-exhaustion generator `licensecorpus` shows the payoff for a closed enum.
`BuildCorpus` enumerates `schema.AllLicenses` into one must-pass case per menu
member (each carrying `enum` provenance and a mutation naming what removing that
member would break), then appends the must-fail negatives (a non-menu id and the
empty license). `cmd/gen-license-corpus` renders it to the committed
`licensecorpus/testdata/license_corpus.yaml`. Two guards, both in the Layer 1
table, keep it honest: `TestLicenseCorpus_ExhaustiveCoverage` reddens if
`schema.AllLicenses` widens without the corpus regenerated (a menu member with no
case, backstopped by `assert.RequireMin` at `len(AllLicenses)` plus the two
negatives), and `TestLicenseCorpus_Freshness` byte-compares the committed file
against a fresh `RenderCorpus`, so a hand-edit or a stale artifact reddens.

The two worked migrations are the reference examples.
`internal/release/grammar_corpus_test.go` moves the `version_kind` and `parse_tag`
grammar corpora onto the standard (`TestVersionKindBaseIsRC`, `TestParseTag`),
loading through `testcase.LoadCorpus` and preserving every pre-migration
assertion. Each guards coverage two ways: an exact case-count control (a min-floor
would pass a silent drop that stays above the floor) plus a lean value-based
coverage assertion (an exact count cannot see a count-preserving swap that drops a
real case and adds a filler). The older `versions.yaml` corpus (`new_version` /
`parse_title`)
deliberately stays on its ad-hoc `loadGrammarFixtures` loader; `testcase` is the
generalization new corpora adopt, not a forced rewrite of every existing family.

### Segmented multi-axis fixtures

A uniform `Corpus[I, E]` covers one row shape and one harness. A feature's
fixtures are often segmented into named behavioral arms with heterogeneous
`(I, E)` and a distinct assertion per arm. The convention for that is a plain
typed struct of named `Corpus` fields, one per arm, so each arm keeps its own
static `(I_arm, E_arm)` at its harness with no downcast. The worked reference
example is the pull skip-gate: `skipGateFixtures` in `pull_skip_gate_test.go`, fed
by `testdata/pull/skip_gate_cases.yaml`, whose four arms are round-trip, canonical
form, ordering by transcript id, and withheld-by-omission:

```go
type skipGateFixtures struct {
    RoundTrip          testcase.Corpus[itemsInput, struct{}]
    Canonical          testcase.Corpus[itemsInput, canonicalExpected]
    OrdersByTranscript testcase.Corpus[resultsInput, orderExpected]
    Withheld           testcase.Corpus[resultsInput, withheldExpected]
}
```

Each arm's `Input` is a collection (the request items or the response results) and
its `Expected` is the global property the harness checks over that collection: the
canonical arm asserts id order plus each sorted, de-duplicated annotation set, and
the withheld arm asserts per case that a withheld id appears nowhere in the
marshaled response bytes (not merely absent from the slice). Each arm carries at
least two structurally-rich cases, the representative behavior plus an idempotence,
empty, or nil boundary case, so no arm can pass vacuously.

Classify each arm before modeling it. The discriminator: does the arm have two or
more genuinely-distinct, independent `input -> expected` examples, each a real,
non-filler behavior? If yes it is a **case-list** and reuses `Corpus[I, E]`, and
`I` may be a collection and `E` a global property of that collection, because the
per-arm harness (not `Corpus`) owns the comparison; `Corpus` only holds the data.
If an arm is inherently one assertion over one fixed collection with no plural
examples, it is a **global-property** arm, and the convention reserves a plain
typed struct with arm-level provenance for it (no current arm needs this; it is
documented for the future).

Guard every arm, pairing the size floor with the non-vacuity check; skip-gate does
this in `LoadSkipGateFixtures`, so every one of its tests guards all four arms.
`assert.RequireValid` is the loud `*testing.T` wrapper around `Corpus.Validate`,
symmetric to `RequireMin` around `CheckMin`; each call names the arm it guards by
its call site:

```go
assert.RequireMin(t, fx.RoundTrip, 2);          assert.RequireValid(t, fx.RoundTrip)
assert.RequireMin(t, fx.Canonical, 2);          assert.RequireValid(t, fx.Canonical)
assert.RequireMin(t, fx.OrdersByTranscript, 2); assert.RequireValid(t, fx.OrdersByTranscript)
assert.RequireMin(t, fx.Withheld, 2);           assert.RequireValid(t, fx.Withheld)
```

Beyond that floor-and-non-vacuity guard, each arm carries a predicate-based
coverage assertion (`requireCoverage` in `TestSkipGateFixtures_Coverage`): a
structural predicate over the loaded cases asserts that each required scenario is
still present, so a count-preserving swap that drops a load-bearing case reddens
even at a fixed count. It is the predicate-based counterpart to the grammar
corpora's value-based coverage assertion, which pins cases by expected value.

There is deliberately no generic `Suite` / `Arm` / `RequireSuite` container. A
single multi-axis consumer does not earn a reusable sweep abstraction, and a
static struct of named `Corpus` fields already delivers what one would: per-arm
distinct types, with a duplicate or empty arm made impossible by the compiler
(two same-named fields do not compile; a struct has no empty-suite state) instead
of guarded at runtime. Promoting a reusable sweep waits for a second
heterogeneous consumer to prove the pattern; until then the convention is the
struct plus the per-arm `RequireMin` + `RequireValid` guard. `RequireValid` has a
fixture-driven negative control symmetric to `RequireMin`'s: a deliberately
vacuous corpus fixture reddens it while a populated one passes. One boundary is
honest and shared by any struct-of-arms shape: nothing structurally forces every
declared arm to be wired to a guard call, so the guard block above is the pattern
to copy; skip-gate wires all four arms to both guards in its loader, and a reviewer
confirms every arm is guarded.
