# @peasant-labs/schema

Generated TypeScript bindings for the canonical `github.com/peasant-labs/schema`
wire contract. The package root mirrors the complete canonical Go Types catalog,
including runtime closed-set objects, frozen `All*` collections, and `is*`
guards. Operation-policy views remain available under `/local-api` and
`/village-api`, alongside schema-owned testcase, quality-fixture, and
timeline-fixture helpers.

This package is not published yet. Its development version is
`0.0.0-development`; a published package version will follow the schema module
release tag, not an individual OpenAPI document version.

```ts
import { Role, isRole, type SessionDetailPayload } from "@peasant-labs/schema";
import type { ReviewListPayload } from "@peasant-labs/schema/local-api";
import { loadCorpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures, qualitySessions } from "@peasant-labs/schema/fixtures/quality";
import { loadTimelineFixtures } from "@peasant-labs/schema/fixtures/timeline";

const fixtures = loadQualityFixtures();
const timeline = loadTimelineFixtures();
qualitySessions(fixtures);
timeline.cases.length;
isRole(Role.User);
```

The raw OpenAPI `components` and `paths` maps are generator internals. Consumers
should use the named exports.
