# Changelog

All notable changes to the `github.com/peasant-labs/schema` contract module are
documented here. This project adheres to [Semantic Versioning](https://semver.org/).

## [v0.1.0-rc2] — 2026-07-06

Second release candidate. Adds the **transcript-licensing contract** and regenerates
the OpenAPI specs at `0.4.0`, so the consumers (`peasant`, `village`) can pin a single
public contract that carries the license surface. Published as a GitHub **prerelease**.

### Added

- **License surface** — a `License` type with the canonical menu `CC0-1.0` /
  `CC-BY-4.0` / `CC-BY-SA-4.0` (`AllLicenses`, `IsValid()`, `LicenseMenu()`), and an
  optional `license` field on `PublishRequest` and `PullTranscriptInfo`. The
  publish-request JSON-Schema `license` enum is derived from `AllLicenses`, never a
  literal.
- **`0.4.0` OpenAPI specs** — `VillageAPIVersion` bumped `0.3.0 → 0.4.0`; regenerated
  `generated/village-api-0.4.0.{json,yaml}` + `generated/publish-request-0.4.0.schema.json`
  carrying the `license` property. The `0.3.0 → 0.4.0` delta is **additive** (oasdiff).
- **Retired-spec completeness guard** — a codegen test asserting every committed spec
  under `generated/` is either current-generated or registered in the retired-spec
  immutability registry. The retiring `0.3.0` specs are now registered and byte-frozen.
- **Quality-fixtures loader** for the contract fixtures.
- **`LICENSE`** — the module is now licensed **Apache-2.0** (`meta.license` set to
  `asl20`; the prior placeholder is removed).

### Notes

- The `VillageAPIVersion` constant (and its sibling spec-version stamps) is an
  intentional version marker consumers pin against; its value change on each spec
  release is exempted from the exported-Go-API breaking gate, with drift detected on
  the consumer side. The retained `pkg/schema/v*` tags remain out of scope as release
  references of this module — unchanged.

## [v0.1.0-rc1] — 2026-06-20

First release candidate of the extracted schema contract: `pkg/schema` lifted out
of `peasant-labs/peasant` into this standalone public module
([peasant-labs/peasant#114](https://github.com/peasant-labs/peasant/issues/114)).
Published as a GitHub **prerelease**; the consumers (`peasant`, `village`) pin
`@v0.1.0-rc1` to validate the pipeline end-to-end before a final `v0.1.0`.

### Added

- **Standalone `github.com/peasant-labs/schema` leaf module** — the generated
  OpenAPI specs, typed contract fixtures, the redaction-example single source of
  truth, and the `validate` / `openapi` / `generated` packages — guarded by a
  leaf-purity audit that pins the dependency set to the allowed contract deps.
- **Release pipeline** — PR-title-driven `release-pr.yml` / `release.yml` +
  `release-guard`, with the contract gates (oasdiff / go-apidiff / vacuum)
  provisioned through the Nix flake devShell rather than `go.mod`.
- **Fixture accessors** `EvalSchemaJSON` + `ContractCorpusFS`, and
  `PublishRequestSchemaJSON()` — validation now compiles the generated
  publish-request schema as the single byte-source.

### Notes

- The retained `pkg/schema/v*` tags (from when the contract lived nested inside
  `peasant`, up to `v1.3.0`) are **not** release references of this module and are
  never moved or deleted.
