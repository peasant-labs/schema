import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { Classification, loadCorpus } from "../dist/testcase.js";
import { zTurnDetail } from "../dist/index.js";

const fixtureSource = await readFile(new URL("../../testdata/session-detail/transcripts/turn_models.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    const input = requireRecord(value, path, ["index", "role", "content", "timestamp", "depth", "model"]);
    requireKeys(input, path, ["index", "role", "content", "timestamp", "depth"]);
    requireInteger(input.index, `${path}.index`);
    requireString(input.role, `${path}.role`);
    requireString(input.content, `${path}.content`);
    requireString(input.timestamp, `${path}.timestamp`);
    requireInteger(input.depth, `${path}.depth`);
    if (Object.hasOwn(input, "model")) requireString(input.model, `${path}.model`);
    return input;
  },
  decodeExpected(value, path) {
    const expected = requireRecord(value, path, ["model_present", "model"]);
    requireKeys(expected, path, ["model_present"]);
    if (typeof expected.model_present !== "boolean") throw new TypeError(`${path}.model_present: must be a boolean`);
    if (Object.hasOwn(expected, "model")) requireString(expected.model, `${path}.model`);
    return expected;
  },
});

assert.equal(fixture.cases.length, 2, "turn model fixture must retain one present and one absent row");

test("generated TurnDetail Zod schema accepts present and absent model rows", async (t) => {
  let present = 0;
  let absent = 0;
  for (const fixtureCase of fixture.cases) {
    await t.test(fixtureCase.name, () => {
      assert.equal(fixtureCase.classification, Classification.MustPass, `${fixtureCase.name}: fixture row must be accepted`);
      const result = zTurnDetail.safeParse(fixtureCase.input);
      assert.equal(result.success, true, `${fixtureCase.name}: generated TurnDetail schema rejected the fixture row`);
      if (!result.success) return;

      const hasModel = Object.hasOwn(fixtureCase.input, "model");
      assert.equal(hasModel, fixtureCase.expected.model_present, `${fixtureCase.name}: fixture expected model presence disagrees with input`);
      assert.equal(Object.hasOwn(result.data, "model"), hasModel, `${fixtureCase.name}: generated schema changed model presence`);
      if (hasModel) {
        present += 1;
        assert.equal(result.data.model, fixtureCase.expected.model, `${fixtureCase.name}: generated model differs from fixture`);
      } else {
        absent += 1;
      }
    });
  }
  assert.equal(present, 1, "fixture coverage must include one generated present model row");
  assert.equal(absent, 1, "fixture coverage must include one generated absent model row");
});

function requireRecord(value, path, allowedKeys) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
  const record = Object.fromEntries(Object.entries(value));
  for (const key of Object.keys(record)) {
    if (!allowedKeys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
  }
  return record;
}

function requireKeys(record, path, requiredKeys) {
  for (const key of requiredKeys) {
    if (!Object.hasOwn(record, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  }
}

function requireInteger(value, path) {
  if (!Number.isSafeInteger(value)) throw new TypeError(`${path}: must be a safe integer`);
}

function requireString(value, path) {
  if (typeof value !== "string" || value.trim() === "") throw new TypeError(`${path}: must be a non-empty string`);
  return value;
}
