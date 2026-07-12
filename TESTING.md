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
| Release grammar + guard | `internal/release/*_test.go`, `cmd/release-guard/*_test.go` | **Hard.** A malformed release title/tag, or a publish behind un-gated workflow, is rejected. |
| License menu exhaustive coverage | `licensecorpus/licensecorpus_test.go` (`TestLicenseCorpus_ExhaustiveCoverage`) | **Hard.** Widening `schema.AllLicenses` without regenerating the corpus fails (a menu member with no case). |
| License corpus regen-freshness | `licensecorpus/licensecorpus_test.go` (`TestLicenseCorpus_Freshness`) | **Hard.** A committed `license_corpus.yaml` that drifts from a fresh `RenderCorpus` (hand-edit or stale) fails. |

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
the required gates.

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
- **Pure loader + pure validators.** `LoadCorpus[I, E]` unmarshals the YAML and
  returns an error rather than failing a test. `Case.Validate` / `Corpus.Validate`
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
static `(I_arm, E_arm)` at its harness with no downcast. The shape, illustrated
for a four-arm feature:

```go
type segmentedFixtures struct {
    RoundTrip          testcase.Corpus[itemsInput, struct{}]
    Canonical          testcase.Corpus[itemsInput, canonicalExpected]
    OrdersByTranscript testcase.Corpus[resultsInput, orderExpected]
    Withheld           testcase.Corpus[resultsInput, withheldExpected]
}
```

Classify each arm before modeling it. The discriminator: does the arm have two or
more genuinely-distinct, independent `input -> expected` examples, each a real,
non-filler behavior? If yes it is a **case-list** and reuses `Corpus[I, E]`, and
`I` may be a collection and `E` a global property of that collection, because the
per-arm harness (not `Corpus`) owns the comparison; `Corpus` only holds the data.
If an arm is inherently one assertion over one fixed collection with no plural
examples, it is a **global-property** arm, and the convention reserves a plain
typed struct with arm-level provenance for it (no current arm needs this; it is
documented for the future).

Guard every arm at the top of the test, pairing the size floor with the
non-vacuity check. `assert.RequireValid` is the loud `*testing.T` wrapper around
`Corpus.Validate`, symmetric to `RequireMin` around `CheckMin`; each call names
the arm it guards by its call site:

```go
assert.RequireMin(t, fx.RoundTrip, 2);          assert.RequireValid(t, fx.RoundTrip)
assert.RequireMin(t, fx.Canonical, 2);          assert.RequireValid(t, fx.Canonical)
assert.RequireMin(t, fx.OrdersByTranscript, 2); assert.RequireValid(t, fx.OrdersByTranscript)
assert.RequireMin(t, fx.Withheld, 2);           assert.RequireValid(t, fx.Withheld)
```

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
to copy, and a reviewer confirms every arm is guarded.
