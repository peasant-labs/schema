import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import * as schema from "../dist/index.js";
const f = parse(await readFile(new URL("../../testdata/collective_lossless_boundaries.yaml", import.meta.url), "utf8"), { merge: true, maxAliasCount: 10000 });
for (const arm of ["timestamps", "scopes", "statuses"]) {
  assert.deepEqual(new Set(f[arm].cases.map(c => c.name)), new Set(f.required_names[arm]));
  assert.equal(new Set(f[arm].cases.map(c => c.name)).size, f[arm].cases.length);
}
function raw(input, extra = {}) {
  const value = structuredClone(input.payload);
  for (const path of input.omit ?? []) {
    const parts = path.split("."); let node = value;
    for (const part of parts.slice(0, -1)) node = node[part];
    delete node[parts.at(-1)];
  }
  for (const [path, item] of Object.entries({ ...input.set, ...extra })) {
    const parts = path.split("."); let node = value;
    for (const part of parts.slice(0, -1)) node = node[part];
    node[parts.at(-1)] = structuredClone(item);
  }
  return value;
}
for (const c of f.timestamps.cases) test(`public lossless timestamp: ${c.name}`, () => {
  const saved = raw(c.input), parsed = schema[`z${c.input.target}`].safeParse(saved);
  assert.equal(parsed.success, c.expected.valid, JSON.stringify(parsed.error?.issues));
  if (c.input.target !== "LocalRemoteContributionAttempt") return;
  if (!c.expected.valid) {
    assert.throws(() => schema.reconcileVillageContributionAttempt(undefined, saved));
    return;
  }
  assert.deepEqual(parsed.data, saved, "raw evidence strings must not be rounded or rewritten");
  if (c.input.incomingSet === undefined) return;
  const incoming = raw(c.input, c.input.incomingSet);
  if (c.expected.conflict) {
    assert.throws(() => schema.reconcileVillageContributionAttempt(saved, incoming));
    return;
  }
  assert.equal(schema.reconcileVillageContributionAttempt(saved, incoming).disposition, c.expected.disposition);
});
for (const c of f.statuses.cases) test(`public cancellation precedence: ${c.name}`, () => {
  const value = raw(c.input);
  assert.equal(schema.collectiveFetchHTTPStatus(value), c.expected.status);
  const parsed = schema.zCollectiveFetchReport.parse(value);
  assert.deepEqual(parsed, value, "status selection must not discard committed effects or failure evidence");
});
for (const c of f.scopes.cases) test(`public raw authority: ${c.name}`, () => {
  const i = f.identity;
  if (!c.expected.valid) {
    assert.throws(() => schema.normalizeCollectiveObservationScope(c.input.url, i.userId));
    const framed = btoa(JSON.stringify(["1", c.input.url, i.userId, i.incarnation, i.kind, i.generation])).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
    assert.throws(() => schema.parseContributionTraversalRevision(framed));
    return;
  }
  const scope = schema.normalizeCollectiveObservationScope(c.input.url, i.userId);
  assert.equal(scope.villageUrl, c.expected.normalized);
  const identity = { scope, incarnation: i.incarnation, kind: i.kind, generation: i.generation };
  assert.equal(schema.encodeContributionTraversalRevision(identity), c.expected.token);
  assert.deepEqual(schema.parseContributionTraversalRevision(c.expected.token), identity);
});
