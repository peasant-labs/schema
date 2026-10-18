import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import * as schema from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/collective_canonical_fields.yaml", import.meta.url), "utf8"), { merge: true });
assert.deepEqual(new Set(fixture.cases.map(c => c.name)), new Set(fixture.required_names));
assert.equal(new Set(fixture.cases.map(c => c.name)).size, fixture.cases.length);
for (const c of fixture.cases) test(`public canonical JSON fields: ${c.name}`, () => {
  // Keep raw JSON until the public boundary, including order and duplicate keys.
  const raw = JSON.parse(c.input.raw);
  const result = schema[`z${c.input.target}`].safeParse(raw);
  assert.equal(result.success, c.expected.valid, JSON.stringify(result.error?.issues));
  if (!c.expected.valid) return;
  const expected = JSON.parse(c.expected.decoded);
  assert.deepEqual(result.data, expected, "decoded evidence must match exact canonical fields, not case variants");
  if (c.input.target === "CollectiveFetchReport") assert.equal(schema.collectiveFetchHTTPStatus(raw), c.expected.status);
  if (c.expected.disposition !== undefined) {
    const saved = c.input.target === "LocalContributionHistoryResponse" ? expected.remoteAttempts[0] : expected;
    const incoming = c.input.target === "LocalContributionHistoryResponse" ? result.data.remoteAttempts[0] : raw;
    const reconciled = schema.reconcileVillageContributionAttempt(saved, incoming);
    assert.equal(reconciled.disposition, c.expected.disposition);
    assert.deepEqual(reconciled.attempt, saved);
  }
});
