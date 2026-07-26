import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { Classification, checkMin, loadCorpus } from "../dist/testcase.js";
import * as schema from "../dist/index.js";

const fixtureSource = await readFile(new URL("../../testdata/typescript/root_zod_association_validation.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    const input = requireRecord(value, path, ["schema", "value"]);
    if (input.schema !== "session-association" && input.schema !== "annotation-summary") {
      throw new TypeError(`${path}.schema: must be session-association or annotation-summary`);
    }
    if (typeof input.value !== "object" || input.value === null || Array.isArray(input.value)) {
      throw new TypeError(`${path}.value: must be a mapping`);
    }
    return input;
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

assert.equal(fixture.cases.length, 30, "root Zod association validation fixture must retain all representative structural cases");
assert.equal(checkMin(fixture, 30), undefined);
assert.equal(new Set(fixture.cases.map((testCase) => testCase.name)).size, fixture.cases.length, "root Zod association validation fixture case names must be unique");

test("built root Zod schemas enforce association and annotation structural invariants", async (t) => {
  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted parse result`);
      assert.equal(rootSchema(testCase.input.schema).safeParse(testCase.input.value).success, testCase.expected);
    });
  }
});

function rootSchema(name) {
  switch (name) {
    case "session-association": return schema.zSessionAssociation;
    case "annotation-summary": return schema.zAnnotationSummary;
    default: throw new TypeError(`root Zod association validation fixture named unsupported schema ${JSON.stringify(name)}`);
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
