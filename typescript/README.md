# @peasant-labs/schema

TypeScript bindings for the canonical `github.com/peasant-labs/schema` wire
contract. The package is a contract-only leaf: it contains data definitions,
runtime Zod schemas, Go-shaped closed-set values and guards, version constants,
and schema-owned fixture helpers. It does not contain HTTP or WebSocket clients.
Consumers that need endpoint typing can import type-only OpenAPI `paths` and
`operations` namespaces from `@peasant-labs/schema/local-api` and
`@peasant-labs/schema/village-api`; these exports do not contain a transport.

The Go contract and its generated Types OpenAPI document are authoritative.
Hey API's Zod plugin derives both the TypeScript definitions and runtime schemas
from that document. `openapi-typescript` derives the Local and Village operation
contracts from their OpenAPI documents. A small generated facade exposes the
same closed-set shape as Go (`Role.User`, `AllRoles`, `isRole`) without
maintaining a handwritten second contract. The retired `@peasant-labs/types`
package must not receive new wire definitions.

`ProjectHash` is a nominal string identity matching Go's validated newtype.
Construct values with `newProjectHash`, narrow unknown wire values with
`isProjectHash`, or assert a trust boundary with `validateProjectHash`.

This package is not published yet. Its development version is
`0.0.0-development`; a published package version will follow the schema module
release tag, not an individual OpenAPI document version.

```ts
import {
  Role,
  isRole,
  newProjectHash,
  zSessionDetailPayload,
  type SessionDetailPayload,
} from "@peasant-labs/schema";
import { loadCorpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures, qualitySessions } from "@peasant-labs/schema/fixtures/quality";
import { loadTimelineFixtures } from "@peasant-labs/schema/fixtures/timeline";
import type { operations as LocalOperations } from "@peasant-labs/schema/local-api";

const detail: SessionDetailPayload = zSessionDetailPayload.parse(input);
type SessionDetailResponse = LocalOperations["getSession"]["responses"][200]["content"]["application/json"];
const fixtures = loadQualityFixtures();
const timeline = loadTimelineFixtures();
qualitySessions(fixtures);
timeline.cases.length;
isRole(Role.User);
newProjectHash("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2");
const response: SessionDetailResponse = detail;
void response;
```
