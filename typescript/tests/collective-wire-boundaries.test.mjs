import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import * as schema from "../dist/index.js";

const f = parse(await readFile(new URL("../../testdata/collective_wire_boundaries.yaml", import.meta.url), "utf8"), { merge: true, maxAliasCount: 10000 });
for (const arm of ["scopes", "payloads", "statuses"]) {
  assert.deepEqual(new Set(f[arm].cases.map(c => c.name)), new Set(f.required_names[arm]));
  assert.equal(new Set(f[arm].cases.map(c => c.name)).size, f[arm].cases.length);
}
function materialize(input) {
  const payload = structuredClone(input.payload);
  for (const path of input.omit ?? []) { const parts = path.split("."); let node = payload; for (const part of parts.slice(0, -1)) node = node[part]; delete node[parts.at(-1)]; }
  return payload;
}
for (const c of f.payloads.cases) test(`public raw collective wire: ${c.name}`, () => {
  const raw=materialize(c.input),result = schema[`z${c.input.target}`].safeParse(raw);
  assert.equal(result.success, c.expected.valid, JSON.stringify(result.error?.issues));
  if(result.success&&(c.input.target==="VillageShareEvent"||c.input.target==="LocalContributionHistoryResponse"))assert.deepEqual(result.data,raw,"attempt presence and nanosecond timestamps must survive parsing");
});
for (const c of f.statuses.cases) test(`public status helper validates raw report: ${c.name}`, () => {
  const raw = materialize(c.input);
  assert.equal(schema.collectiveFetchHTTPStatus(raw), c.expected.status);
  assert.equal(schema.zCollectiveFetchReport.safeParse(raw).success, c.expected.valid);
});
for (const c of f.scopes.cases) test(`public exact scope and token bytes: ${c.name}`, () => {
  const i = f.identity;
  if (!c.expected.valid) {
    assert.throws(() => schema.normalizeCollectiveObservationScope(c.input.url, i.userId));
    const raw = JSON.stringify(["1", c.input.url, i.userId, i.incarnation, i.kind, i.generation]);
    const framed = btoa(String.fromCharCode(...new TextEncoder().encode(raw))).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
    assert.throws(() => schema.parseContributionTraversalRevision(framed));
    return;
  }
  const scope = schema.normalizeCollectiveObservationScope(c.input.url, i.userId);
  assert.equal(scope.villageUrl, c.expected.normalized);
  assert.equal(typeof c.expected.token, "string");
  const identity = { scope, incarnation: i.incarnation, kind: i.kind, generation: i.generation };
  assert.equal(schema.encodeContributionTraversalRevision(identity), c.expected.token);
  assert.deepEqual(schema.parseContributionTraversalRevision(c.expected.token), identity);
});
