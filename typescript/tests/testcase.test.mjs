import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
  Classification,
  AllClassifications,
  AllProvenanceSources,
  ProvenanceSource,
  checkMin,
  loadCorpus,
  validateCorpus,
} from "../dist/testcase.js";

const matrixSource = await readFile(new URL("../../testcase/testdata/load_cases.yaml", import.meta.url), "utf8");
const minimumMatrixSource = await readFile(new URL("../../testcase/testdata/check_min_cases.yaml", import.meta.url), "utf8");
const decoderCaseSource = await readFile(new URL("./fixtures/decoder-case.yaml", import.meta.url), "utf8");

const stringDecoder = (value, path) => {
  if (typeof value !== "string") throw new Error(`${path}: must be a string`);
  return value;
};
const booleanDecoder = (value, path) => {
  if (typeof value !== "boolean") throw new Error(`${path}: must be a boolean`);
  return value;
};

test("TypeScript and Go share the strict loader matrix", async (t) => {
  const matrix = loadCorpus(matrixSource, {
    decodeInput: stringDecoder,
    decodeExpected: booleanDecoder,
  });
  assert.equal(matrix.cases.length, 16);
  assert.equal(checkMin(matrix, 16), undefined);
  assert.equal(validateCorpus(matrix), undefined);

  for (const testCase of matrix.cases) {
    await t.test(testCase.name, () => {
      let accepted = true;
      try {
        loadCorpus(testCase.input, {
          decodeInput: stringDecoder,
          decodeExpected: booleanDecoder,
        });
      } catch {
        accepted = false;
      }
      assert.equal(accepted, testCase.expected);
    });
  }
});

test("TypeScript and Go share minimum-size boundaries", () => {
  const matrix = loadCorpus(minimumMatrixSource, {
    decodeInput(value, path) {
      if (typeof value !== "object" || value === null || Array.isArray(value)) throw new Error(`${path}: must be a mapping`);
      if (typeof value.size !== "number" || typeof value.minimum !== "number") throw new Error(`${path}: size and minimum must be numbers`);
      return { size: value.size, minimum: value.minimum };
    },
    decodeExpected: booleanDecoder,
  });
  assert.equal(matrix.cases.length, 4);
  for (const testCase of matrix.cases) {
    const corpus = { cases: Array.from({ length: testCase.input.size }, () => ({})) };
    assert.equal(checkMin(corpus, testCase.input.minimum) === undefined, testCase.expected, testCase.name);
  }
});

test("closed testcase values mirror Go", () => {
  assert.deepEqual(Object.values(Classification), ["must-pass", "must-fail"]);
  assert.deepEqual(AllClassifications, ["must-pass", "must-fail"]);
  assert.deepEqual(Object.values(ProvenanceSource), ["requirement", "bug", "enum", "boundary", "manual"]);
  assert.deepEqual(AllProvenanceSources, ["requirement", "bug", "enum", "boundary", "manual"]);
});

test("generic decoders own the input and expected boundary", () => {
  assert.throws(
    () => loadCorpus(decoderCaseSource, {
      decodeInput(value, path) {
        assert.equal(typeof value, "object");
        const keys = Object.keys(value);
        if (keys.some((key) => key !== "value")) throw new Error(`${path}: unknown input key`);
        return value.value;
      },
      decodeExpected: booleanDecoder,
    }),
    /unknown input key/,
  );
});
