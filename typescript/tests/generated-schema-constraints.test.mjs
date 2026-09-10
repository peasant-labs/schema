import assert from "node:assert/strict";
import test from "node:test";

import * as contract from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixtures = sessionGraphFixtures();

test("every canonical count carrier enforces the shared count corpus", async (t) => {
  for (const component of fixtures.typescriptConstraints.countCarriers) await t.test(component, () => {
    const schema = contract[`z${component}`];
    const field = component.startsWith("Village") ? "input_submission_count" : "inputSubmissionCount";
    assert.ok(schema?.shape?.[field], `${component} must expose its input submission count`);
    for (const row of fixtures.durable.counts.cases) {
      const raw = JSON.parse(row.input.json);
      const value = Object.hasOwn(raw, "inputSubmissionCount") ? raw.inputSubmissionCount : undefined;
      assert.equal(schema.shape[field].safeParse(value).success, row.classification === "must-pass", `${component}: ${row.name}`);
    }
  });
});

test("every narrow graph component rejects unknown properties", async (t) => {
  for (const [component, canonical] of Object.entries(fixtures.typescriptConstraints.strictObjects)) await t.test(component, () => {
    const schema = contract[`z${component}`];
    assert.equal(schema.safeParse(canonical).success, true, `${component} canonical shape`);
    assert.equal(schema.safeParse({...canonical, explanation: "must not be stripped"}).success, false, `${component} unknown property`);
  });
});
