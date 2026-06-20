# Changelog

All notable changes to the `github.com/peasant-labs/schema` contract module are
documented here. This project adheres to [Semantic Versioning](https://semver.org/).

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
