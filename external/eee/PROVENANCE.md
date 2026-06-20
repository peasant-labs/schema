# EEE Schema Provenance

| Field | Value |
|-------|-------|
| Source | https://raw.githubusercontent.com/evaleval/every_eval_ever/v0.2.1/eval.schema.json |
| Version | v0.2.1 |
| $schema | http://json-schema.org/draft-07/schema# |
| version field | 0.2.1 |
| Fetched | 2026-03-04 |
| Purpose | Validate EEE metric config output shape against external consumer contract |

## Notes

The authoritative `eval.schema.json` is the full EEE evaluation result document schema
from the [evaleval/every_eval_ever](https://github.com/evaleval/every_eval_ever) repository.
It defines the top-level structure for storing and validating LLM evaluation data.

Peasant's `EEEMetricConfig` maps to the `metric_config` sub-object within
`evaluation_results.items`, located at JSON pointer:

```
#/properties/evaluation_results/items/properties/metric_config
```

### Conformance constraint

The authoritative `metric_config` sub-schema (v0.2.1) requires `lower_is_better`.
It also enforces conditional requirements via `if/then/else`:

- `score_type: "levels"` → requires `level_names` and `has_unknown_level`
- `score_type: "continuous"` → requires `min_score` and `max_score`

**Note:** The `if` condition in the schema uses `properties` without a `required`
guard, which means the `then` branch triggers when `score_type` is absent. Peasant
always emits `score_type` to avoid this edge case and produce unambiguous output.

### Re-vendoring

To update to a new release:

1. Fetch the new `eval.schema.json` from the tagged release:
   ```
   https://raw.githubusercontent.com/evaleval/every_eval_ever/<TAG>/eval.schema.json
   ```
2. Replace this file with the new content exactly as fetched.
3. Update `eeeVendoredSchemaVersion` in `internal/export/eee_drift_test.go`.
4. Update this PROVENANCE.md with the new Source URL, Version, and Fetched date.
5. Run `go test -race ./internal/export/...` — update test fixtures if the
   `metric_config` sub-schema contract has changed.
