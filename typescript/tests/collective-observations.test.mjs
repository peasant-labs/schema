import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import * as schema from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/collective_observations.yaml", import.meta.url), "utf8"), { merge: true, maxAliasCount: 10000 });
for (const [name, values] of Object.entries(fixture.menus)) test(`collective exact closed menu: ${name}`, () => {
  const validator = schema[`z${name}`];
  assert.deepEqual(new Set(validator.options), new Set(values));
  for (const value of values) assert.equal(validator.safeParse(value).success, true);
  for (const value of fixture.invalid_menu_values) assert.equal(validator.safeParse(value).success, false);
});
for (const arm of ["validation", "tokens", "history", "organizations"]) {
  assert.deepEqual(new Set(fixture[arm].cases.map(c => c.name)), new Set(fixture.required_names[arm]));
  assert.equal(fixture[arm].cases.length, new Set(fixture[arm].cases.map(c => c.name)).size, "duplicate fixture name");
}
for (const c of [...fixture.validation.cases, ...fixture.organizations.cases]) test(`collective contract shared validation: ${c.name}`, () => {
  const validator = schema[`z${c.input.target}`];
  assert.ok(validator, "public generated validator is missing");
  const before = structuredClone(c.input.payload), result = validator.safeParse(c.input.payload);
  assert.equal(result.success, c.expected.valid, JSON.stringify(result.error?.issues));
  if (result.success && c.expected.status !== undefined) assert.equal(schema.collectiveFetchHTTPStatus(result.data), c.expected.status);
  assert.deepEqual(c.input.payload, before, "validation mutated source evidence");
});
for (const c of fixture.tokens.cases) test(`collective contract shared token: ${c.name}`, () => {
  const i = c.input;
  let identity, current;
  try {
    identity = { scope: schema.normalizeCollectiveObservationScope(i.url, i.userId), incarnation: i.incarnation, kind: i.kind, generation: i.generation };
    current = schema.encodeContributionTraversalRevision(identity);
  } catch (error) { assert.equal(c.expected.valid, false, error.message); return; }
  assert.equal(c.expected.valid, true);
  assert.equal(identity.scope.villageUrl, c.expected.normalized);
  if (c.expected.token !== undefined) assert.equal(current, c.expected.token);
  assert.deepEqual(schema.parseContributionTraversalRevision(current), identity);
  const query = { kind: i.queryKind ?? i.kind, page: i.page, limit: i.limit, revision: current }, other = structuredClone(identity);
  switch (i.revision) {
    case "none": delete query.revision; break;
    case "malformed": query.revision = "not-a-token"; break;
    case "other-kind": other.kind = "receipts"; break;
    case "other-user": other.scope.userId = i.incarnation; break;
    case "other-incarnation": other.incarnation = i.userId; break;
    case "older-generation": other.generation = "0"; break;
    case "other-scope": other.scope.villageUrl += "/other"; break;
  }
  if (JSON.stringify(other) !== JSON.stringify(identity)) query.revision = schema.encodeContributionTraversalRevision(other);
  assert.equal(schema.validateContributionHistoryQuery(query, current), c.expected.status);
});
for (const c of fixture.history.cases) test(`collective contract shared strict history: ${c.name}`, () => {
  const payload = structuredClone(c.input.payload);
  payload.traversalRevision = schema.encodeContributionTraversalRevision({ scope: payload.scope, incarnation: fixture.tokens.cases[0].input.incarnation, kind: payload.traversalRevision === "$other-token" ? "receipts" : payload.kind, generation: "1" });
  const result = schema.zLocalContributionHistoryResponse.safeParse(payload);
  assert.equal(result.success, c.expected.valid, JSON.stringify(result.error?.issues));
  if (result.success) assert.deepEqual(result.data, payload, "selected empty array must survive validation");
});
