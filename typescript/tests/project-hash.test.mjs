import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { isProjectHash, newProjectHash, validateProjectHash } from "../dist/index.js";
import { checkMin, loadCorpus } from "../dist/testcase.js";

const fixtureSource = await readFile(new URL("../../testdata/typescript/project_hash.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    if (typeof value !== "string") throw new TypeError(`${path}: must be a string`);
    return value;
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

test("ProjectHash runtime helpers match the strict shared fixture", async (t) => {
  assert.equal(fixture.cases.length, 9);
  assert.equal(checkMin(fixture, 9), undefined);
  assert.equal(new Set(fixture.cases.map((testCase) => testCase.name)).size, fixture.cases.length);
  assert.equal(new Set(fixture.cases.map((testCase) => testCase.input)).size, fixture.cases.length);

  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(isProjectHash(testCase.input), testCase.expected);
      if (testCase.expected) {
        assert.equal(newProjectHash(testCase.input), testCase.input);
        let narrowed = /** @type {unknown} */ (testCase.input);
        validateProjectHash(narrowed);
        assert.equal(narrowed, testCase.input);
        return;
      }

      for (const [operation, invoke] of [
        ["newProjectHash", () => newProjectHash(testCase.input)],
        ["validateProjectHash", () => validateProjectHash(testCase.input)],
      ]) {
        assert.throws(invoke, (error) => {
          assert.equal(error instanceof TypeError, true);
          assert.match(error.message, /ProjectHash validation failed/);
          assert.match(error.message, new RegExp(escapeRegExp(JSON.stringify(testCase.input))));
          assert.match(error.message, /@peasant-labs\/schema ProjectHash/);
          assert.match(error.message, new RegExp(operation));
          assert.match(error.message, /64-character lowercase hexadecimal string/);
          assert.match(error.message, /cannot use it as a canonical project identity/);
          assert.match(error.message, /pass the lowercase SHA-256 hex digest/);
          return true;
        });
      }
    });
  }
});

test("ProjectHash runtime boundary rejects non-string values", () => {
  for (const value of [undefined, null, 42, {}, []]) {
    assert.equal(isProjectHash(value), false);
    assert.throws(() => validateProjectHash(value), TypeError);
  }
});

test("ProjectHash guard rejects a mutation of a valid fixture", () => {
  const valid = fixture.cases.find((testCase) => testCase.expected)?.input;
  assert.equal(typeof valid, "string");
  const mutated = valid.toUpperCase();
  assert.notEqual(mutated, valid);
  assert.equal(isProjectHash(mutated), false);
});

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
