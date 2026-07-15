# @peasant-labs/schema

Generated TypeScript bindings for the canonical `github.com/peasant-labs/schema`
wire contract. The package exposes named types from the current Village API,
Peasant local API, and shared type-catalog specs, plus schema-owned testcase and
quality-fixture helpers.

This package is not published yet. Its development version is
`0.0.0-development`; a published package version will follow the schema module
release tag, not an individual OpenAPI document version.

```ts
import type { SessionDetailPayload } from "@peasant-labs/schema";
import type { ReviewListPayload } from "@peasant-labs/schema/local-api";
import { loadCorpus } from "@peasant-labs/schema/testcase";
import { qualitySessions } from "@peasant-labs/schema/fixtures/quality";
```

The raw OpenAPI `components` and `paths` maps are generator internals. Consumers
should use the named exports.
