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
const inputFixtureSource = await readFile(new URL("../../testdata/typescript/project_hash_inputs.yaml", import.meta.url), "utf8");
const inputFixture = loadCorpus(inputFixtureSource, {
  decodeInput(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
    assert.deepEqual(Object.keys(value), ["kind"], `${path}: descriptor keys differ from the strict kind-only shape`);
    if (typeof value.kind !== "string") throw new TypeError(`${path}.kind: must be a string`);
    return { kind: value.kind };
  },
  decodeExpected(value, path) {
    if (typeof value !== "string") throw new TypeError(`${path}: must be a string`);
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

      const rendered = JSON.stringify(testCase.input);
      assertProjectHashFailure("newProjectHash", rendered, () => newProjectHash(testCase.input));
      assertProjectHashFailure("validateProjectHash", rendered, () => validateProjectHash(testCase.input));
    });
  }
});

test("ProjectHash owns actionable errors for strict non-string and hostile fixtures", async (t) => {
  assert.equal(inputFixture.cases.length, 9);
  assert.equal(checkMin(inputFixture, 9), undefined);
  assert.equal(new Set(inputFixture.cases.map((testCase) => testCase.name)).size, inputFixture.cases.length);
  assert.equal(new Set(inputFixture.cases.map((testCase) => testCase.input.kind)).size, inputFixture.cases.length);

  for (const testCase of inputFixture.cases) {
    await t.test(testCase.name, () => {
      const value = materializeInput(testCase.input);
      assert.equal(isProjectHash(value), false);
      assertProjectHashFailure("newProjectHash", testCase.expected, () => newProjectHash(value));
      assertProjectHashFailure("validateProjectHash", testCase.expected, () => validateProjectHash(value));
    });
  }
});

test("ProjectHash guard rejects a mutation of a valid fixture", () => {
  const valid = fixture.cases.find((testCase) => testCase.expected)?.input;
  assert.equal(typeof valid, "string");
  const mutated = valid.toUpperCase();
  assert.notEqual(mutated, valid);
  assert.equal(isProjectHash(mutated), false);
});

test("hostile descriptor mutation changes the owned rendering path", () => {
  const hostile = inputFixture.cases.find((testCase) => testCase.input.kind === "throwing-json-and-string");
  assert.notEqual(hostile, undefined);
  const mutated = materializeInput({ kind: "object" });
  const error = captureProjectHashFailure(() => validateProjectHash(mutated));
  assert.match(error.message, /failed for \{\}/);
  assert.doesNotMatch(error.message, new RegExp(escapeRegExp(hostile.expected)));
});

function assertProjectHashFailure(operation, rendered, invoke) {
  const error = captureProjectHashFailure(invoke);
  assert.match(error.message, /ProjectHash validation failed/);
  assert.match(error.message, new RegExp(escapeRegExp(rendered)));
  assert.match(error.message, /@peasant-labs\/schema ProjectHash/);
  assert.match(error.message, new RegExp(operation));
  assert.match(error.message, /64-character lowercase hexadecimal string/);
  assert.match(error.message, /cannot use it as a canonical project identity/);
  assert.match(error.message, /pass the lowercase SHA-256 hex digest/);
}

function captureProjectHashFailure(invoke) {
  try {
    invoke();
  } catch (error) {
    assert.equal(error instanceof TypeError, true);
    return error;
  }
  assert.fail("ProjectHash operation unexpectedly accepted an invalid value");
}

function materializeInput(descriptor) {
  switch (descriptor.kind) {
    case "undefined": return undefined;
    case "null": return null;
    case "number": return 42;
    case "object": return {};
    case "array": return [];
    case "bigint": return 42n;
    case "symbol": return Symbol("project-hash");
    case "throwing-json": return { toJSON() { throw new Error("hostile toJSON"); } };
    case "throwing-json-and-string": return {
      toJSON() { throw new Error("hostile toJSON"); },
      toString() { throw new Error("hostile toString"); },
    };
    default: throw new Error(`unknown ProjectHash input descriptor ${JSON.stringify(descriptor.kind)}`);
  }
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
