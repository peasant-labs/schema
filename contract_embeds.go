package schema

import "embed"

// EvalSchemaJSON is the vendored EEE (eval) JSON Schema
// (external/eee/eval.schema.json) — the contract that peasant's `export`
// surface validates its evaluation payloads against. Exported so consumers
// (peasant internal/export conformance + drift tests) read the schema through
// the contract leaf instead of a cross-module filesystem path. Provenance and
// re-vendoring policy: external/eee/PROVENANCE.md.
//
//go:embed external/eee/eval.schema.json
var EvalSchemaJSON []byte

// ContractCorpusFS is the push/pull back-compat golden corpus tree
// (testdata/contract/**): the current contract shape plus the retained legacy
// shapes (legacy-metadata-field / legacy-provider-keyed / legacy-raw-jsonl).
// Exported as an embed.FS so consumers (peasant internal/push back-compat
// round-trip tests) enumerate and read the corpus through the contract leaf
// rather than a cross-module filesystem path. Walk it from the "testdata/contract"
// root (e.g. fs.ReadDir(schema.ContractCorpusFS, "testdata/contract")).
//
//go:embed testdata/contract
var ContractCorpusFS embed.FS
