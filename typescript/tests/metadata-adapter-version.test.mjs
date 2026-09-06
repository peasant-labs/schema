import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import { MetadataSchemaVersion, zUnifiedMetadata } from "../dist/index.js";
import { loadCorpus } from "../dist/testcase.js";

const corpus = loadCorpus(await readFile(new URL("../../testdata/metadata/adapter_versions.yaml", import.meta.url), "utf8"), {
  decodeInput(value) {
    assert.equal(typeof value, "string");
    return value;
  },
  decodeExpected(value) {
    assert.equal(typeof value.accepted, "boolean");
    assert.ok(Number.isInteger(value.schemaVersion));
    assert.ok(value.adapterVersion === null || Number.isInteger(value.adapterVersion));
    return value;
  },
});
const manifest = parse(await readFile(new URL("../../testdata/metadata/adapter_versions_manifest.yaml", import.meta.url), "utf8"));
assert.ok(manifest.requiredCaseNames.length > 0);
assert.deepEqual(new Set(corpus.cases.map((c) => c.name)), new Set(manifest.requiredCaseNames));

test("metadata adapter revisions match the Go fixture contract", async (t) => {
  assert.equal(MetadataSchemaVersion, 10);
  for (const c of corpus.cases) {
    await t.test(c.name, () => {
      const input = { ...JSON.parse(manifest.baseMetadata), ...JSON.parse(c.input) };
      const result = zUnifiedMetadata.safeParse(input);
      assert.equal(result.success, c.expected.accepted, JSON.stringify(result.error?.issues));
      if (result.success) {
        assert.equal(result.data.schemaVersion, c.expected.schemaVersion);
        assert.equal(result.data.adapterVersion ?? null, c.expected.adapterVersion);
        assert.equal(Object.hasOwn(result.data, "adapterVersion"), c.expected.adapterVersion !== null);
      } else {
        assert.ok(result.error.issues.some((issue) => issue.path[0] === "adapterVersion"));
      }
    });
  }
});
