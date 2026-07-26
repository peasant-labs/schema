import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { Classification, checkMin, loadCorpus } from "../dist/testcase.js";
import * as schema from "../dist/index.js";

const fixtureSource = await readFile(new URL("../../testdata/typescript/root_zod_association_validation.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    const input = requireRecord(value, path, ["schema", "value"]);
    if (input.schema !== "session-association" && input.schema !== "annotation-summary" && input.schema !== "session-id") {
      throw new TypeError(`${path}.schema: must be session-association, annotation-summary, or session-id`);
    }
    if (input.schema === "session-id" && typeof input.value !== "string") {
      throw new TypeError(`${path}.value: must be a string for session-id`);
    }
    if (input.schema !== "session-id" && (typeof input.value !== "object" || input.value === null || Array.isArray(input.value))) {
      throw new TypeError(`${path}.value: must be a mapping`);
    }
    return input;
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});
const sharedFixtureSource = await readFile(new URL("../../testdata/local-api/association_parity.yaml", import.meta.url), "utf8");
const sharedFixture = loadCorpus(sharedFixtureSource, {
  decodeInput(value, path) {
    return requireRecord(value, path, ["id", "sessionId", "conclusion", "confidence", "evidence"]);
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

assert.equal(fixture.cases.length, 31, "root Zod association validation fixture must retain all representative structural cases");
assert.equal(checkMin(fixture, 31), undefined);
assert.equal(new Set(fixture.cases.map((testCase) => testCase.name)).size, fixture.cases.length, "root Zod association validation fixture case names must be unique");
assertSharedParityFixtureInventory(sharedFixture);

test("built root Zod schemas enforce association and annotation structural invariants", async (t) => {
  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted parse result`);
      assert.equal(rootSchema(testCase.input.schema).safeParse(testCase.input.value).success, testCase.expected);
    });
  }
});

test("built root Zod session associations share the Go parity corpus", async (t) => {
  for (const testCase of sharedFixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted parse result`);
      assert.equal(schema.zSessionAssociation.safeParse(testCase.input).success, testCase.expected);
    });
  }
});

function rootSchema(name) {
  switch (name) {
    case "session-association": return schema.zSessionAssociation;
    case "annotation-summary": return schema.zAnnotationSummary;
    case "session-id": return schema.zSessionID;
    default: throw new TypeError(`root Zod association validation fixture named unsupported schema ${JSON.stringify(name)}`);
  }
}

function assertSharedParityFixtureInventory(parityFixture) {
  const requiredNames = new Set([
    "go_valid_non_empty_session_id_is_accepted",
    "distinct_same_kind_recorded_commit_observations_in_canonical_order_are_accepted",
    "unicode_e000_then_10000_utf8_order_is_accepted",
    "unicode_10000_then_e000_utf8_order_is_rejected",
  ]);
  assert.equal(parityFixture.cases.length, requiredNames.size, "shared association parity fixture must retain its complete representative corpus");
  assert.equal(checkMin(parityFixture, requiredNames.size), undefined);
  const names = new Set(parityFixture.cases.map((testCase) => testCase.name));
  assert.equal(names.size, parityFixture.cases.length, "shared association parity fixture case names must be unique");
  for (const name of requiredNames) {
    assert.ok(names.has(name), `shared association parity fixture is missing ${name}`);
  }
}

function requireRecord(value, path, keys) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
  const record = Object.fromEntries(Object.entries(value));
  for (const key of Object.keys(record)) {
    if (!keys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
  }
  for (const key of keys) {
    if (!Object.hasOwn(record, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  }
  return record;
}
