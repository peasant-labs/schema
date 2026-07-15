# @peasant-labs/schema

Generated TypeScript bindings for the canonical `github.com/peasant-labs/schema`
wire contract. The package root mirrors the complete canonical Go Types catalog,
including runtime closed-set objects, frozen `All*` collections, and `is*`
guards. Endpoint path and operation maps remain available under `/local-api`
and `/village-api`; they reference canonical root payload types instead of
publishing a second domain namespace. The deprecated `/types` subpath is a pure
root re-export. Schema-owned testcase, quality-fixture, and
timeline-fixture helpers.

This package is not published yet. Its development version is
`0.0.0-development`; a published package version will follow the schema module
release tag, not an individual OpenAPI document version.

```ts
import { Role, isRole, type SessionDetailPayload } from "@peasant-labs/schema";
import type { operations as LocalOperations } from "@peasant-labs/schema/local-api";
import { loadCorpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures, qualitySessions } from "@peasant-labs/schema/fixtures/quality";
import { loadTimelineFixtures } from "@peasant-labs/schema/fixtures/timeline";

const fixtures = loadQualityFixtures();
const timeline = loadTimelineFixtures();
qualitySessions(fixtures);
timeline.cases.length;
isRole(Role.User);

type ReviewTimeline = LocalOperations["listReviewChanges"];
```

The raw OpenAPI component maps are generator internals. Consumers import domain
types from the package root and endpoint paths or operations from the matching
API subpath.
