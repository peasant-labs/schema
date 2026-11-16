import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import { parseSegmentedReadVersionAdvertisement, parseSchemaVersionAdvertisement } from "../dist/index.js";
import { Classification, loadCorpus } from "../dist/testcase.js";

const fixture = await readFile(new URL("../../testdata/typescript/segmented_read_version_advertisement.yaml", import.meta.url), "utf8");
const manifest = parse(await readFile(new URL("../../testdata/typescript/segmented_read_version_advertisement_manifest.yaml", import.meta.url), "utf8"), { uniqueKeys: true });
const corpus = loadCorpus(fixture, {
  decodeInput(value) { return value; },
  decodeExpected(value) { return value; },
});
assert.deepEqual(corpus.cases.map((row) => row.name).sort(), [...manifest.requiredNames].sort());

test("segmented read advertisement parses its independent opaque window", async (t) => {
  for (const row of corpus.cases) {
    await t.test(row.name, () => {
      if (row.classification === Classification.MustFail) {
        assert.throws(() => parseSegmentedReadVersionAdvertisement(row.input.json), (error) => {
          assert.ok(error instanceof TypeError);
          assert.ok(error.message.includes(row.expected.errorContains), error.message);
          return true;
        });
      } else {
        const actual = parseSegmentedReadVersionAdvertisement(row.input.json);
        assert.deepEqual(actual, row.expected.outcome === "absent" ? null : {current: row.expected.current, min: row.expected.min});
        if (row.expected.capabilities !== undefined) {
          const capabilities = parseSchemaVersionAdvertisement(row.input.json);
          assert.deepEqual(capabilities, row.expected.capabilities);
          if (row.input.capabilityBaseline !== undefined) {
            assert.deepEqual(capabilities, parseSchemaVersionAdvertisement(row.input.capabilityBaseline));
          }
        }
      }
    });
  }
});
