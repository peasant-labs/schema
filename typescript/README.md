# @peasant-labs/schema

Generated TypeScript bindings for the canonical `github.com/peasant-labs/schema`
wire contract. The package root mirrors the complete canonical Go Types catalog,
including runtime closed-set objects, frozen `All*` collections, and `is*`
guards. Endpoint path and operation maps remain available under `/local-api`
and `/village-api`; they reference canonical root payload types instead of
publishing a second domain namespace. The deprecated `/types` subpath is a pure
root re-export. Schema-owned testcase, quality-fixture, and
timeline-fixture helpers.

`ProjectHash` is a nominal string identity matching Go's validated newtype.
Construct values with `newProjectHash`, narrow unknown wire values with
`isProjectHash`, or assert a trust boundary with `validateProjectHash`. Root
payloads and Local/Village operation types use that same branded identity.

This package is not published yet. Its development version is
`0.0.0-development`; a published package version will follow the schema module
release tag, not an individual OpenAPI document version.

```ts
import { Role, isRole, newProjectHash, type SessionDetailPayload } from "@peasant-labs/schema";
import type { operations as LocalOperations } from "@peasant-labs/schema/local-api";
import { loadCorpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures, qualitySessions } from "@peasant-labs/schema/fixtures/quality";
import { loadTimelineFixtures } from "@peasant-labs/schema/fixtures/timeline";

const fixtures = loadQualityFixtures();
const timeline = loadTimelineFixtures();
qualitySessions(fixtures);
timeline.cases.length;
isRole(Role.User);
newProjectHash("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2");

type ReviewTimeline = LocalOperations["listReviewChanges"];
```

The raw OpenAPI component maps are generator internals. Consumers import domain
types from the package root and endpoint paths or operations from the matching
API subpath.
