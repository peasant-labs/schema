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
make test                           # go test ./... (RACE=1 adds -race)
make race-selective                 # go test -race over the validator schema caches
make gates BASE_REF=origin/develop  # the breaking-change contract gates (needs the flake tools)
make freshness                      # git-diff backstop on the generated artifacts
pnpm --dir typescript run typecheck
pnpm --dir typescript test
pnpm --dir typescript run package:audit
pnpm --dir typescript run package:smoke
```

`make check` is `fmt` + `vet` + `freshness` + TypeScript checks + the
unstring guard + `go test ./...`. That suite is the
authoritative run: it includes the leaf audit, the vendor-hash guard, and the
`oasdiff` / `go-apidiff` synthetic-break tests (which skip when their binary is
absent).

The default is no race detector, on every CI event and locally. The module's
only production concurrency is the lazy schema compilation behind
`ValidatePublishRequest`, `ValidateAnnotationPushRequest` (`sync.Once`) and
`ValidateDescribedValue` (`sync.Map`). `make race-selective` runs
`TestValidatorCachesConcurrently` (fixture
`testdata/validator_cache_concurrency.yaml`) alone under `-race`, so the first
compilation and every cache read contend; the merge-queue run requires it.
`make check RACE=1` (or `make test RACE=1`) is the opt-in full race pass.

`race-selective` names its tests in `RACE_SELECTIVE_TESTS` (Makefile) and
fails unless each one runs and passes, so a renamed test cannot leave it
passing with zero tests. New production concurrency must add a test that
exercises it from many goroutines and extend `RACE_SELECTIVE_TESTS`. The
`sync.Once` caches race if replaced by an unguarded nil check and the
`sync.Map` cache races if replaced by a plain map (both verified to report
`DATA RACE`); `sync.Map` is race-safe by construction, so for it the test
only guards against such a replacement.

Run one test or one package:

```bash
go test ./cmd/schema-gen/ -run TestRetiredSpecsImmutable -v
go test ./internal/contractgates/ -run TestOasdiffSyntheticBreak -v   # skips without oasdiff
```

### Bare `go` vs `nix develop`

- **Bare `go test ./...`** runs everything except the tool-gated synthetic-break
  tests, which `t.Skip` via `skipIfMissing` (see
  `internal/contractgates/synthetic_break_test.go`).
- **Inside `nix develop`** (or `direnv allow`) `oasdiff` / `go-apidiff` /
  `vacuum` are on `PATH`, so the synthetic-break tests and `make gates` run for
  real. CI runs `make gates` with `fetch-depth: 0` so the base ref is resolvable.

### Tests never read your git configuration

A test result must not depend on the machine. Every package whose tests start a
process (git, a shell script, `go`) has a `TestMain` that calls
`testutil.RunWithIsolatedGitConfig` (`internal/testutil/gitconfig.go`). Before
any test runs it sets `GIT_CONFIG_GLOBAL=/dev/null` and `GIT_CONFIG_NOSYSTEM=1`,
and unsets `GIT_CONFIG_COUNT`, `GIT_CONFIG_PARAMETERS` and the repository
location variables (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE`,
`GIT_COMMON_DIR`, `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`)
that a hook or `rebase -x` exports. A global `commit.gpgsign`, a
`url.*.insteadOf` rewrite or injected `GIT_CONFIG_*` values therefore never
reach a child, and a test's git never acts on an outer repository.
`GIT_CONFIG_GLOBAL` needs git 2.32 or newer; older git ignores it. Do not add
per-call `-c commit.gpgsign=false` or `GIT_CONFIG_GLOBAL` workarounds. Set the
commit identity per call (`GIT_AUTHOR_*`, `GIT_COMMITTER_*`), as the existing
helpers do.

`TestEveryPackageThatStartsProcessesIsolatesGitConfig` scans test files that
import `os/exec`. Their package must have an unconstrained `TestMain` ending in
`os.Exit(testutil.RunWithIsolatedGitConfig(...))`, with no other `m.Run`,
`goleak.VerifyTestMain`, `os.Exit`, or isolation-helper call. Setup before the
exit is allowed. The helper's own package instead calls unexported
`isolateGitConfig` in a top-level `if err := isolateGitConfig(); err != nil`
that exits nonzero, then ends with `os.Exit(m.Run())`, without a competing
runner, exit, or isolation call. It runs independently so a broken helper
cannot skip its own tests. The fixture corpus is
`internal/testutil/testdata/git_isolation_scanner.yaml`.
`TestIsolatedChildGitReadsNoHostileSource` runs a real child git under every
hostile source and checks that it reads none of them once isolated.

The guard cannot see:

- a test that starts git only through non-test code (it checks `_test.go`
  imports only);
- a process started without `os/exec` (`os.StartProcess`, `syscall.Exec`);
- git run by a dependency outside the module;
- a `TestMain` that calls the helper on only some paths;
- a child whose environment a test rebuilds from scratch without the isolation
  variables.

## Layer 1: the contract gates

These are the load-bearing invariants. Each has a real, named test (or gate
script) behind it.

| Gate | Where it lives | Consequence |
|---|---|---|
| Codegen freshness | `cmd/schema-gen/freshness_test.go`, `cmd/schema-gen/redactions_freshness_test.go`, `cmd/schema-gen/routes_test.go`, `cmd/schema-gen/generated_accounted_test.go`; `make freshness` | **Hard.** A committed spec that drifts from the Go source fails the build. |
| Retired-version immutability | `cmd/schema-gen/retired_specs_test.go` | **Hard.** A frozen spec that is edited or deleted fails. |
| OpenAPI breaking diff (`oasdiff`) | `scripts/contract-gates.sh`, `cmd/schema-gen/report.go`; `internal/contractgates/synthetic_break_test.go` | An ERR-level break is reported against the base surface version and must be declared in the title. |
| OpenAPI lint (`vacuum`) | `scripts/contract-gates.sh` + `.vacuum.yaml` | **Hard.** An `error`-severity finding in a generated spec fails. |
| Exported-Go-API diff (`go-apidiff`) | `scripts/contract-gates.sh`; `internal/contractgates/synthetic_break_test.go` | Incompatible changes must be declared in the title; an unparseable diff fails closed. |
| Leaf-purity audit | `leaf_audit_test.go` | **Hard.** A `go.mod` direct require outside the allowed set fails. |
| Vendor-hash stability | `vendorhash_test.go` | **Hard.** A local-path `replace` in `go.mod` fails. |
| Release grammar + guard | pinned `release-intent`, `internal/release/*_test.go`, `cmd/release-guard/*_test.go`, `cmd/schema-gen/release_adoption_workflow_test.go` | **Hard.** Malformed intent titles or tags are rejected; CI checks workflow policy and the publication job wiring. |
| License menu exhaustive coverage | `licensecorpus/licensecorpus_test.go` (`TestLicenseCorpus_ExhaustiveCoverage`) | **Hard.** Widening `schema.AllLicenses` without regenerating the corpus fails (a menu member with no case). |
| License corpus regen-freshness | `licensecorpus/licensecorpus_test.go` (`TestLicenseCorpus_Freshness`) | **Hard.** A committed `license_corpus.yaml` that drifts from a fresh `RenderCorpus` (hand-edit or stale) fails. |
| Local API 0.5.0 closed-set exhaustive coverage + regen-freshness | `enumcorpus/enumcorpus_test.go` (`TestEnumCorpora_ExhaustiveCoverageAndFreshness`, one subtest pair per enum) | **Hard.** `enumcorpus.BuildCorpus`/`RenderCorpus` generalize `licensecorpus`'s pattern over any closed string enum; `cmd/gen-enum-corpora` regenerates all ten committed corpora (`AssociationConclusion`, `AssociationEvidenceKind`, `Confidence`, `RewriteResolution`, `RewriteMethod`, `InsightKind`, `InsightProvenance`, `ReadAttributionState`, `ReadStateGrade`, `TargetKind`) in one `go generate`. |
| `ReadStateGrade` registry-seed cross-check | `enumcorpus/enumcorpus_test.go` (`TestReadStateGradeRegistrySeedCrossCheck`) | **Hard.** `schema.ReadStateGradeRegistrySeedPermissibleValues` must byte-equal `AllReadStateGrades` minus `none`; the peasant-side read-state registry seed pins the other half against this exported value. |
| Declared session origin | `session_origin_test.go`, `testdata_session_origin.go`; `testdata/local-api/session_origin.yaml` | **Hard.** The corpus derives its menu from `AllSessionOrigins` and its behaviour coverage from `AllSessionOriginBehaviours`, both by set membership, so widening the enum or deleting the rows that carry a behaviour fails at load. Every menu member round-trips on BOTH `SessionDetailPayload` and `SessionSummary`, and the published `SessionOrigin` component must refuse an off-menu token with a needle naming the whole menu. |
| Local publishing and Village pull request read boundaries | `publishing_contract_boundaries_test.go`; `testdata/publishing_corpora.yaml` (the one corpus list both languages read), `testdata/local-api/{sync_sessions,publications,sync_push,village_collectives,sync_auth,settings}.yaml`, `testdata/pulls/{transcript_pull_requests,personal_stats}.yaml`; `typescript/tests/publishing-contract-zod.test.mjs` | **Hard.** One harness runs every case through two independent layers: the published Types component (JSON Schema) and the Go decoder plus `Validate`. Each invalid case names the layer that refuses it (`rejected_by: shape` or `semantics`), so a relational rule cannot pass because the schema happens to refuse the payload, and a shape rule cannot pass because Go happens to; a shape case also states `go_rejects`, and `go_error_contains` names the Go rule when a closed-set guard must be the one that refuses, so an unrelated relational rule cannot mask it. Valid cases round-trip through Go and the emitted JSON must still satisfy the component. Each file's `operations` bind the component to its Local or Village operation with its parameters and, where the corpus owns the operation's exact status set, the component of every declared refusal body, and `required_names` is the exact case manifest. A boundary corpus in those fixture directories that the shared list omits fails the run. The TypeScript suite replays the same files against the generated Zod schemas, which must accept exactly the cases the shape layer accepts. |
| Local write routes declare the cross-origin refusal | `cross_origin_writes_test.go`; `testdata/local-api/cross_origin_writes.yaml` | **Hard.** The manifest is the exact set of declared state-changing Local API operations. Each must declare 403 with the shared JSON error body and state the refusal in its description, and a declared write operation outside the manifest fails. |
| Working and frozen specs | `cmd/schema-gen/freeze_status_test.go`, `freeze_handoff_test.go`, `freeze_writer_test.go`; `cmd/schema-gen/testdata/freeze_*.yaml` | **Hard.** Real oasdiff classifies the fixture documents, the handoff rejects stale or incoherent plans, and the real writer's recorded writes must be authorized by the release config. The writer test uses an isolated temporary source copy and recompiles the generator; ordinary temporary-directory cleanup removes it on completion, while an interrupted process may leave it under the test temporary directory. |
| Segmented document and protocol versions | `openapi/segmented_read_test.go`, `segmented_read_versions_test.go`, `cmd/schema-gen/freeze_writer_test.go` | **Hard.** A document bump changes its title and version only; the real freeze retains the minimum document, Village working specs, protocol selectors and read-context validation. |
| Release asset selection | `cmd/schema-gen/release_assets_test.go`; `cmd/schema-gen/testdata/release_assets*.yaml` | **Hard.** The shipped selection script copies numbered documents, including request schemas, and rejects a distribution already containing working files. |
| Village request bodies against the served document | `openapi/village_request_bodies_test.go`; `request_bodies` in `openapi/testdata/village_collectives_operations.yaml` | **Hard.** Each body is validated against the request schema its operation declares, as Village does when it enforces the served document, and `request_body_names` is the exact case manifest; the create-group request accepts a null `linked_github_org` exactly as update does. |
| TypeScript closed-set completeness | `openapi_enums_test.go`; `testdata/typescript/enums.yaml` | **Hard.** Every canonical Go closed set is an OpenAPI enum before TypeScript generation can run. |
| Route members against the documents | `route/route_test.go`; `cmd/schema-gen/routes_test.go`; `typescript/tests/route-members.ts`, `route-members.test.mjs`; `cmd/schema-gen/testdata/route_members.yaml`, `testdata/route_paths.yaml`, `testdata/typescript/route_path_refusals.yaml` | **Hard.** `route.AllVillage` and `route.AllLocal` equal the built documents' operations (method, pattern, operationId), the TypeScript members equal them at run time and `keyof paths` at compile time, every member mounts on one `ServeMux`, the generator refuses a document it cannot project, a missing path parameter is a compile error, and both languages refuse an empty, `.` or `..` path parameter value with the same message from one shared fixture. |
| TypeScript generated-file freshness | `make freshness` | **Hard.** Hey API/Zod root output, `openapi-typescript` operation contracts, and YAML-derived fixture data are byte-stable after regeneration. |
| TypeScript typecheck + fixture tests | `typescript/tsconfig*.json`, `typescript/tests/` | **Hard.** Public types compile and both languages accept/reject the same strict YAML matrix. |
| TypeScript operation-alias collision safety | `typescript/scripts/lib/operation-aliases.mjs`; `typescript/tests/operation-aliases.test.mjs`; `testdata/typescript/collision_cases.yaml` | **Hard.** A Local/Village API component whose schema drifts from the canonical root type it would alias to stops generation with an actionable error; a synthetic-break test proves the throw fires on a real mismatch, not only on today's already-consistent specs. |
| TypeScript ProjectHash brand-resolver safety | `typescript/scripts/lib/project-hash-resolver.mjs`; `typescript/tests/project-hash-resolver.test.mjs`; `testdata/typescript/project_hash_resolver_cases.yaml` | **Hard.** The generation-time decision that brands a string schema as `ProjectHash` requires both the canonical pattern AND the canonical `#/components/schemas/ProjectHash` `$ref` location; a 6-case fixture (1 positive, 5 negative-control) drives the extracted decision function directly, proving a same-pattern differently-named or differently-located schema is never branded. |
| TypeScript closed-set member and field types | `typescript/tests/enum-member-types.ts` (compiled by `pnpm typecheck` and the packed-consumer smoke); `typescript/tests/enum-member-aliases.test.mjs` | **Hard.** Each member is its own literal type, the named set type derives from the member table, and every multi-member generated Zod schema reports that named type, so `TurnDetail["role"]` is `Role` with a const `Role` table beside it. A wrong literal fails compilation, and the schema/table pair is checked at compile time. |
| TypeScript public contract completeness | `typescript/scripts/generate-contract-support.mjs` (`renderPublicContract`) | **Hard.** Every Types OpenAPI catalog component must produce exactly one matching Hey API Zod export; a missing, extra, or renamed export fails generation instead of only asserting the facade is non-empty. |
| TypeScript facade export identity | `typescript/tests/public-exports.test.mjs`; `testdata/typescript/public_exports.yaml`, `public_export_mutations.yaml` | **Hard.** The hand-maintained root/`local-api`/`village-api` facade exports (aliases, version constants, ProjectHash functions, forbidden names) match a fixture; a dedicated add/remove/duplicate/redirect mutation corpus proves the check is not vacuous. |
| TypeScript ProjectHash wire-location coverage | `typescript/tests/project-hash-locations.ts`, `project-hash-locations.test.mjs`; `testdata/typescript/project_hash_locations.yaml` | **Hard.** Compile-time proof that the ProjectHash brand reaches all 5 named wire locations plus a same-spelling negative control; a runtime coupling test fails if a location is dropped from either the fixture or the compile-time file (removing a passing assertion is not by itself a type error). |
| Published package manifest, content + tarball imports | `typescript/scripts/package-*.mjs`; `typescript/tests/package-audit.test.mjs` | **Hard.** The real packed manifest retains exact npm provenance repository metadata, only audited files ship, and every public subpath imports from a disposable packed install. |

### Codegen freshness

The isolated `tools/redaction-categories/` module calls the checksum-pinned,
released engine to generate the browser category bindings. It has no root Go
dependency. `make check` runs its vet and race tests. The TypeScript redaction
corpus checks exact inventories, both conversion directions, invalid JavaScript
inputs and invalid projection mutations. Its freshness test runs the real tool
twice and compares the result with the committed browser output. Packed-package
smoke checks runtime conversion and positive and negative declaration use from
the package root. No scan wire or standalone Go validator policy changes.

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

`scripts/contract-gates.sh` runs the contract checks against a base ref
(`make gates BASE_REF=origin/develop`). The report compares committed `HEAD`
with the base; it does not include uncommitted changes:

- **`vacuum`** lints generated OpenAPI specs, including segmented reads, against
  `.vacuum.yaml`, failing on `error` severity. Request JSON Schemas are not
  OpenAPI documents.
- Numbered files already present on the base are immutable. For the report,
  **`oasdiff`** compares each surface's base and head versions, including a new
  numbered file, and records ERR-level breaks.
- **`go-apidiff`** reports incompatible exported Go changes and compatible
  additions. Intentional spec-version stamp value changes are exempt; an
  unparseable result fails closed. The reviewed `workflow_run` PR comment
  carries its tri-state result, including for fork PRs.
- **`release-intent check-contract`** rejects an undeclared break. Locally,
  `make gates` fails conservatively without a title; set `TITLE` and put the
  pinned tool on `PATH` to judge the same report as CI.

A gate that can never fail is worthless, so the gates have their own tests.
`TestContractGatesAllAndReleaseVerdict` runs the real `all` path, including
vacuum, go-apidiff, `go run ./cmd/schema-gen report`, and oasdiff, in a
temporary repository with two-commit ranges. It verifies the requested output
and head, ERR-level break, immutable-file refusal, and independently written
report. It builds `release-intent` from the workflow's pinned infra commit,
checks the config and full workflow policy, and runs tag-time
`check-contract --until ... --floor ... --report ...` on declared and
undeclared breaking subjects. The latter must report `under-declared-breaking`.
The first uncached run needs network access to download the pinned tool module.
The report tests skip without oasdiff locally, but fail if `CI` is set.
`TestPreTagPolicyResolvesChecks` also resolves policy step IDs through the
release job's needed jobs and asserts the exact declared pre-tag command.
`TestMergeGateVendorHash` runs the mounted merge-gate shell with fixture needs,
including failed and skipped vendor-hash results for both release and content.

`internal/contractgates/synthetic_break_test.go` proves they fire:
`TestOasdiffSyntheticBreak` removes an endpoint from a committed golden spec and
asserts `oasdiff breaking --fail-on ERR` exits non-zero, while
`TestOasdiffNoBreakOnIdenticalSpec` pins the no-false-positive side. The eleven
`TestGoAPIDiff*` regressions cover both the real tool and the runner's failure
boundaries. `TestGoAPIDiffSyntheticBreak`, `TestGoAPIDiffStampExemption`, and
`TestGoAPIDiffGateRunner` run the real binary against throwaway repositories;
the first proves a removed export is reported, the second anchors the
stamp-exemption format, and the third verifies warning, clean, and compatible
signals. `TestGoAPIDiffGateRunnerLinkedWorktree` runs the real binary from a
clean linked worktree and proves the runner uses a standalone clone instead of
letting `go-apidiff` inspect linked-worktree metadata.

`TestGoAPIDiffStampExemptionFilter` and
`TestGoAPIDiffCompatibleExtraction` drive the shipped decision seams with
canned real-format output, including the unparseable-section fail-closed path
and separate compatible/incompatible channels. The remaining five runner tests
use controlled fakes: `TestGoAPIDiffGateRunnerFailClosed` rejects an unparseable
report without writing either signal; `TestGoAPIDiffGateRunnerInvocationFailure`
rejects a non-zero tool invocation; `TestGoAPIDiffGateRunnerIsolatedCloneCleanup`
verifies immutable base/head hashes reach a standalone `--no-local` clone with
no alternates and that successful cleanup removes it;
`TestGoAPIDiffGateRunnerTimeoutFailsClosed` rejects a timed-out isolated run;
and `TestGoAPIDiffGateRunnerCleanupFailureFailsClosed` rejects failed cleanup
without publishing signals. The runner validates a whole-second timeout from 1
through 600 seconds and fails closed for isolated setup or cleanup errors.

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

`internal/release/*_test.go` cover tag grammar and classification, while
`cmd/release-guard/*_test.go` cover the CLI and its output. Final releases do
not require a prior rc. CI runs pinned `release-intent check-workflows` against
the workflow directory and the release queue policy.
`TestReleaseAdoptionPublicationWiring` separately checks that the release
workflow contains its required jobs, pins their dependency sets, and gives the
npm job its own `id-token: write` permission and scalar `npm-publish`
environment. Its required-name manifest prevents a case from silently
disappearing. These schema-specific checks run in `make check`.

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

`testdata/retained_unknown.yaml` drives the Go typed/raw detail and envelope
decoders, generated Types schema, TypeScript generated shape validators and
public runtime parsers. Strict loading guards required case names and the full
ingestion harness inventory (including all six native fallback producers).
Deletion mutations prove every required case is load-bearing. The corpus covers
legacy absence, open kind namespaces, required source positions and pointers,
duplicate/overlapping locations, invalid JSON and Unicode, exact payload-text
round trips with large number literals and whitespace, partial-state mirrors,
canonical publication fingerprint binding, and existing transport byte limits.
Shape checks are deliberately distinguished from semantic source-order and
embedded-JSON checks: receivers use the public raw decoders, not shape-only
validation before persistence. Capability accumulation and producer ordering
also run through `testdata/content_capability_session_graph.yaml` in both
languages.

`testdata/retained_unknown_boundaries.yaml` adds ordered canonical/alias
overwrites at standalone, envelope and nested evidence boundaries. The baseline
carries every preservation capability, so a case-fold overwrite cannot make
negotiation disappear. Independent fixture-only structs capture original source
evidence before production decoding; negative truncation and normalization
mutations prove that preservation assertions fail on byte loss. Payload-only
mutations must also change the real canonical publication fingerprint. The main
corpus covers 12,000 sibling pointers without changing their external order,
escape-heavy and Unicode input, and actual transport-byte boundaries independent
of optional HTML escaping during a different serialization.

The retained corpus also pins valid `~0`/`~1` pointer siblings, escaped-component
ancestry and whole-record root conflicts in both insertion orders. Private-key
sentinels in duplicate, malformed and depth-invalid JSON must remain absent from
typed and raw Go detail/envelope errors and public TypeScript errors. These
checks also forbid underlying scanner cause chains, require the safe canonical
field/index, and prove rejection does not mutate the evidence.

Typed test corpora live in `testdata/` trees, one family per surface:

- top level `testdata/`: `annotations`, `contract`, `local-api`, `publish`, `pull`,
  `quality`, `session-detail`, `sync`;
- `internal/release/testdata/`: `grammar`;
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

Another example is `openapi/publish_schema_parity_test.go`
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
- **Required case names.** `testcase/assert.RequireManifestNames` checks a
  separate YAML manifest against fixture case names by exact membership, rejects
  empty or repeated names, and requires a single document with known fields.
  Use it for fixtures whose cases must not disappear or be added silently.

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

The release tag and version-kind corpora in
`internal/release/grammar_corpus_test.go` use `testcase.LoadCorpus` and guard
required case names with manifests. The older version-validation corpus retains
its ad-hoc loader.

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

The declared-session-origin family is the second worked example, and it segments
for a different reason: `SessionOriginFixtures` in `testdata_session_origin.go`
carries a predicate arm over raw tokens, a round-trip arm over encoded payloads,
and a schema arm over the generated Types component. Alongside the arms it
declares two required-NAME manifests, `required_behaviours` and `required_menu`,
plus the by-id operation's `description_anchors`. The loader compares each to its
Go closed set by membership in both directions, never by a row count, so a
deleted row, a deleted anchor, or a widened enum fails at load rather than
quietly reducing coverage.

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
