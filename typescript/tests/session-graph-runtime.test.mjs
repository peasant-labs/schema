import assert from "node:assert/strict";
import test from "node:test";

import { parseSessionDetailPayloadText } from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixtures = sessionGraphFixtures();

test("generated durable graph schema consumes the canonical count and recursive corpus", async (t) => {
  const arms = [fixtures.durable.round_trip, fixtures.durable.counts, fixtures.durable.invalid_counts, fixtures.recursive, fixtures.raw_durable];
  for (const arm of arms) for (const row of arm.cases) await t.test(row.name, () => {
    const raw = row.input.json ?? row.input.rawJSON;
    let result;
    try { result = parseSessionDetailPayloadText(raw); }
    catch (error) { assert.equal(row.classification, "must-fail", String(error)); return; }
    assert.equal(row.classification, "must-pass");
    if (Object.hasOwn(JSON.parse(raw), "inputSubmissionCount")) {
      assert.equal(Object.hasOwn(result, "inputSubmissionCount"), true);
      assert.equal(result.inputSubmissionCount, JSON.parse(raw).inputSubmissionCount);
    }
  });
});
