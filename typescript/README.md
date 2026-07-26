# @peasant-labs/schema

Use this package when a TypeScript application needs typed, always-in-sync
access to the peasant-labs wire contract: the same domain types, closed-set
enums, and runtime validation Go services already produce and enforce,
optionally with type-only Local/Village operation contracts for API-aware
tooling.

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

The generated root Zod schemas enforce the wire's structural shape and selected
generator-owned local cross-field invariants, including association evidence and
annotation target rules. The Go `Validate` methods remain authoritative at the
trust boundary for payload-context and cross-object relationships. TypeScript
consumer adapters remain responsible for any post-parse semantic handling.

`ProjectHash` is a nominal string identity matching Go's validated newtype.
Construct values with `newProjectHash`, narrow unknown wire values with
`isProjectHash`, or assert a trust boundary with `validateProjectHash`.

This package is published to npm as `@peasant-labs/schema`. The committed
`package.json` in this directory stays `0.0.0-development` + `private: true` as
a local safety; the schema module's release pipeline stamps the real version
from the module release tag (not an individual OpenAPI document version) and
publishes at release time, so every published version corresponds to a tagged
schema module release. An `-rcN` tag publishes under the npm dist-tag `next`; a
final `vX.Y.Z` publishes under `latest`.

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
