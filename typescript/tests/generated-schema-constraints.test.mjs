import assert from "node:assert/strict";
import test from "node:test";

import * as contract from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixtures = sessionGraphFixtures();

// A counts row states its value under one of the two session count keys; the
// value, not the key, is what every carrier must accept or refuse.
const countValue = (row) => {
  const raw = JSON.parse(row.input.json);
  const key = ["inputSubmissionCount", "conversationTurnCount"].find((name) => Object.hasOwn(raw, name));
  return key === undefined ? undefined : raw[key];
};

test("every canonical count carrier enforces the shared count corpus", async (t) => {
  const rows = [...fixtures.durable.counts.cases, ...fixtures.durable.invalid_counts.cases];
  assert.ok(rows.some((row) => row.classification === "must-pass") && rows.some((row) => row.classification === "must-fail"), "the count corpus must hold accepted and refused rows");
  for (const { component, field, nullable = false } of fixtures.typescriptConstraints.countCarriers) await t.test(`${component}.${field}`, () => {
    const schema = contract[`z${component}`];
    assert.ok(schema?.shape?.[field], `${component} must expose ${field}`);
    for (const row of rows) {
      const value = countValue(row);
      let accepted = row.classification === "must-pass";
      // A nullable carrier is required and reads null as unmeasured.
      if (nullable && value === undefined) accepted = false;
      if (nullable && value === null) accepted = true;
      assert.equal(schema.shape[field].safeParse(value).success, accepted, `${component}.${field}: ${row.name}`);
    }
  });
});

test("every generated optional count field is a listed count carrier", () => {
  const countFields = new Set(fixtures.typescriptConstraints.countCarriers.map((carrier) => carrier.field));
  const listed = new Set(fixtures.typescriptConstraints.countCarriers.map(({ component, field }) => `${component}.${field}`));
  const found = [];
  for (const [name, schema] of Object.entries(contract)) {
    if (!name.startsWith("z") || schema?.shape === undefined) continue;
    for (const field of countFields) if (schema.shape[field] !== undefined) found.push(`${name.slice(1)}.${field}`);
  }
  assert.deepEqual(found.filter((carrier) => !listed.has(carrier)), [], "list every component that carries a count so its bounds are replayed");
});

test("every narrow graph component rejects unknown properties", async (t) => {
  for (const [component, canonical] of Object.entries(fixtures.typescriptConstraints.strictObjects)) await t.test(component, () => {
    const schema = contract[`z${component}`];
    assert.equal(schema.safeParse(canonical).success, true, `${component} canonical shape`);
    assert.equal(schema.safeParse({...canonical, explanation: "must not be stripped"}).success, false, `${component} unknown property`);
  });
});
