import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { Classification, loadCorpus } from "../dist/testcase.js";
import * as schema from "../dist/index.js";

const fixtureSource = await readFile(new URL("../../openapi/testdata/typescript_requiredness.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    const input = requireRecord(value, path, ["component", "required", "optional", "nullable", "nonnullable"]);
    return {
      component: requireName(input.component, `${path}.component`),
      required: requireNameList(input.required, `${path}.required`),
      optional: requireNameList(input.optional, `${path}.optional`),
      nullable: requireNameList(input.nullable, `${path}.nullable`),
      nonnullable: requireNameList(input.nonnullable, `${path}.nonnullable`),
    };
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

assert.equal(fixture.cases.length, 8, "requiredness fixture must retain its eight representative structures");

test("built root Zod schemas preserve the shared listed-property requiredness corpus", async (t) => {
  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, Classification.MustPass, `${testCase.name}: requiredness rows must be must-pass cases`);
      assert.equal(testCase.expected, true, `${testCase.name}: requiredness rows must expect runtime parity`);

      const zComponent = schema[`z${testCase.input.component}`];
      assert.ok(zComponent !== undefined, `root facade does not export z${testCase.input.component}`);
      assert.equal(typeof zComponent, "object", `z${testCase.input.component} is not a Zod object schema`);
      assert.ok(zComponent !== null && Object.hasOwn(zComponent, "shape"), `z${testCase.input.component} does not expose an object shape`);

      const required = new Set(testCase.input.required);
      const optional = new Set(testCase.input.optional);
      const nullable = new Set(testCase.input.nullable);
      const nonnullable = new Set(testCase.input.nonnullable);
      const fields = new Set([...required, ...optional, ...nullable, ...nonnullable]);

      for (const field of fields) {
        assert.equal(Number(required.has(field)) + Number(optional.has(field)), 1, `${testCase.input.component}.${field} must be classified exactly once as required or optional`);
        assert.equal(Number(nullable.has(field)) + Number(nonnullable.has(field)), 1, `${testCase.input.component}.${field} must be classified exactly once as nullable or nonnullable`);

        const zField = zComponent.shape[field];
        assert.ok(zField !== undefined, `z${testCase.input.component}.shape does not contain ${field}`);
        assert.equal(typeof zField.isOptional, "function", `z${testCase.input.component}.shape.${field} does not report optionality`);
        assert.equal(typeof zField.isNullable, "function", `z${testCase.input.component}.shape.${field} does not report nullability`);
        assert.equal(zField.isOptional(), optional.has(field), `${testCase.input.component}.${field} optionality differs from the shared corpus`);
        assert.equal(zField.isNullable(), nullable.has(field), `${testCase.input.component}.${field} nullability differs from the shared corpus`);
      }
    });
  }
});

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

function requireNameList(value, path) {
  if (!Array.isArray(value)) throw new TypeError(`${path}: must be a sequence`);
  const fields = value.map((candidate, index) => requireName(candidate, `${path}[${index}]`));
  if (new Set(fields).size !== fields.length) throw new TypeError(`${path}: must not repeat a field`);
  return fields;
}

function requireName(value, path) {
  if (typeof value !== "string" || value.trim() === "") throw new TypeError(`${path}: must be a non-empty string`);
  if (value !== value.trim()) throw new TypeError(`${path}: must not have surrounding whitespace`);
  return value;
}
