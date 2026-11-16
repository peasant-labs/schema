import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import { readTranscriptServedTurn } from "../dist/index.js";
import { Classification, loadCorpus } from "../dist/testcase.js";

const source = await readFile(new URL("../../testdata/typescript/transcript_served_turn.yaml", import.meta.url), "utf8");
const manifest = parse(await readFile(new URL("../../testdata/typescript/transcript_served_turn_manifest.yaml", import.meta.url), "utf8"), { uniqueKeys: true });
const corpus = loadCorpus(source, {
  decodeInput(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value) || typeof value.json !== "string" || Object.keys(value).some((key) => key !== "json")) throw new TypeError(`${path}: expected only a JSON text string`);
    return value;
  },
  decodeExpected(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value) || typeof value.accepted !== "boolean" || Object.keys(value).some((key) => key !== "accepted" && key !== "json") || (value.accepted && typeof value.json !== "string")) throw new TypeError(`${path}: expected an acceptance verdict and accepted JSON value`);
    return value;
  },
});

assert.deepEqual(Object.keys(manifest), ["requiredNames"]);
assert.ok(Array.isArray(manifest.requiredNames) && manifest.requiredNames.length > 0);
assert.equal(new Set(manifest.requiredNames).size, manifest.requiredNames.length);
assert.deepEqual(corpus.cases.map((row) => row.name).sort(), [...manifest.requiredNames].sort());

test("positioned turn reader retains served values without asserting the legacy shape", async (t) => {
  for (const row of corpus.cases) {
    await t.test(row.name, () => {
      const input = JSON.parse(row.input.json);
      const result = readTranscriptServedTurn(input);
      assert.equal(row.classification, row.expected.accepted ? Classification.MustPass : Classification.MustFail);
      if (!row.expected.accepted) {
        assert.equal(result, undefined);
        return;
      }
      assert.equal(result, input, "the boundary must retain the received object, not rewrite it");
      assert.deepEqual(result, JSON.parse(row.expected.json));
      assert.equal(Object.hasOwn(result, "index"), Object.hasOwn(JSON.parse(row.expected.json), "index"));
    });
  }
});
