import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import {
  zSessionDetailPayload,
  zTranscriptSegmentDirectoryV2,
  zTranscriptSegmentRow,
  zTranscriptSegmentV2,
  zTranscriptSessionHeader,
  zTranscriptWindowV2,
  zTurnDetail,
} from "../dist/index.js";
import { Classification, loadCorpus } from "../dist/testcase.js";

const fixture = async (name) => readFile(new URL(`../../testdata/${name}`, import.meta.url), "utf8");
const headerDocument = parse(await fixture("transcript_content_header.yaml"), { uniqueKeys: true });
const headerCases = headerDocument.projection.cases.filter((row) => row.classification === Classification.MustPass);
assert.ok(headerCases.length > 0, "the served header projection corpus must not be empty");

test("the generated header accepts every successful Go projection without changing it", async (t) => {
  for (const row of headerCases) {
    await t.test(row.name, () => {
      for (const json of [row.expected.json, row.expected.other]) {
        if (json === undefined) continue;
        const value = JSON.parse(json);
        const result = zTranscriptSessionHeader.safeParse(value);
        assert.equal(result.success, true, result.success ? "" : JSON.stringify(result.error.issues));
        assert.deepEqual(result.data, value);
      }
    });
  }
});

async function corpus(name, keys) {
  const source = await fixture(`typescript/${name}.yaml`);
  const manifest = parse(await fixture(`typescript/${name}_manifest.yaml`), { uniqueKeys: true });
  const loaded = loadCorpus(source, {
    decodeInput(value, path) {
      if (typeof value !== "object" || value === null || Array.isArray(value) ||
          Object.keys(value).length !== keys.length || keys.some((key) => typeof value[key] !== "string")) {
        throw new TypeError(`${path}: expected string members ${keys.join(", ")}`);
      }
      return value;
    },
    decodeExpected(value, path) {
      if (typeof value !== "boolean") throw new TypeError(`${path}: expected a boolean verdict`);
      return value;
    },
  });
  assert.deepEqual(Object.keys(manifest), ["requiredNames"]);
  assert.ok(Array.isArray(manifest.requiredNames) && manifest.requiredNames.length > 0);
  assert.equal(new Set(manifest.requiredNames).size, manifest.requiredNames.length);
  assert.deepEqual(loaded.cases.map((row) => row.name).sort(), [...manifest.requiredNames].sort());
  return loaded.cases;
}

const shapes = await corpus("transcript_v2_generated_shapes", ["schema", "json"]);
const schemas = Object.freeze({
  directory: zTranscriptSegmentDirectoryV2,
  row: zTranscriptSegmentRow,
  segment: zTranscriptSegmentV2,
  window: zTranscriptWindowV2,
});

test("generated v2 shapes preserve required arrays and optional identity", async (t) => {
  for (const row of shapes) {
    await t.test(row.name, () => {
      const schema = schemas[row.input.schema];
      assert.ok(schema, `unknown generated schema ${row.input.schema}`);
      assert.equal(row.classification, row.expected ? Classification.MustPass : Classification.MustFail);
      assert.equal(schema.safeParse(JSON.parse(row.input.json)).success, row.expected);
    });
  }
});

const offsets = await corpus("transcript_generated_offsets", ["turn", "detail"]);
test("generated durable validators accept numeric offsets and retain the known limit", async (t) => {
  for (const row of offsets) {
    await t.test(row.name, () => {
      assert.equal(row.classification, row.expected ? Classification.MustPass : Classification.MustFail);
      for (const [schema, json] of [[zTurnDetail, row.input.turn], [zSessionDetailPayload, row.input.detail]]) {
        const value = JSON.parse(json);
        const result = schema.safeParse(value);
        assert.equal(result.success, row.expected, result.success ? "" : JSON.stringify(result.error.issues));
        if (result.success) assert.deepEqual(result.data, value);
      }
    });
  }
});
