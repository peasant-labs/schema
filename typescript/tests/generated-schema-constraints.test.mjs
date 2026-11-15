import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import * as contract from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";
import { Classification, loadCorpus } from "../dist/testcase.js";

const fixtures = sessionGraphFixtures();
const integerFixture = "../../testdata/typescript/wire_integer_compatibility";
const integerSource = await readFile(new URL(`${integerFixture}.yaml`, import.meta.url), "utf8");
const integerManifest = parse(await readFile(new URL(`${integerFixture}_manifest.yaml`, import.meta.url), "utf8"), { uniqueKeys: true });
const integerCases = loadCorpus(integerSource, {
  decodeInput: (value) => value,
  decodeExpected: (value) => value,
}).cases;
assert.deepEqual(Object.keys(integerManifest), ["requiredNames"]);
assert.ok(integerManifest.requiredNames.length > 0);
assert.equal(new Set(integerManifest.requiredNames).size, integerManifest.requiredNames.length);
assert.deepEqual(integerCases.map((row) => row.name).sort(), [...integerManifest.requiredNames].sort());
const headerRows = integerCases.filter((row) => row.input.target === "header");
assert.deepEqual(
  headerRows.map((row) => row.name).sort(),
  integerManifest.requiredNames.filter((name) => name.startsWith("header-")).sort(),
);
for (const row of headerRows) {
  assert.equal(row.expected.valid, row.classification === Classification.MustPass, `${row.name}: verdict and classification must agree`);
}

// A counts row states its value under one of the two session count keys; the
// value, not the key, is what every carrier must accept or refuse.
const countValue = (row) => {
  const raw = JSON.parse(row.input.json);
  const key = ["inputSubmissionCount", "conversationTurnCount"].find((name) => Object.hasOwn(raw, name));
  return key === undefined ? undefined : raw[key];
};

test("every listed count carrier enforces its declared count corpus", async (t) => {
  const canonicalRows = [...fixtures.durable.counts.cases, ...fixtures.durable.invalid_counts.cases];
  for (const rows of [canonicalRows, headerRows]) {
    assert.ok(rows.some((row) => row.classification === Classification.MustPass) && rows.some((row) => row.classification === Classification.MustFail), "each count corpus must hold accepted and refused rows");
  }
  for (const { component, field, nullable = false, domain = "canonical" } of fixtures.typescriptConstraints.countCarriers) await t.test(`${component}.${field}`, () => {
    const schema = contract[`z${component}`];
    assert.ok(schema?.shape?.[field], `${component} must expose ${field}`);
    assert.ok(domain === "canonical" || domain === "served_int64", `${component}.${field}: unknown count domain ${domain}`);
    if (domain === "served_int64") assert.equal(nullable, false, `${component}.${field}: served header counts omit null`);
    const rows = domain === "canonical" ? canonicalRows : headerRows;
    for (const row of rows) {
      const value = countValue(row);
      let accepted = row.classification === Classification.MustPass;
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
