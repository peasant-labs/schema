import assert from "node:assert/strict";
import fs from "node:fs/promises";
import test from "node:test";
import YAML from "yaml";

import * as schema from "../dist/index.js";
import { validateCorpus } from "../dist/testcase.js";

// The Go boundary harness runs these corpora through the published component
// schemas and the Go validators. The generated Zod schemas carry the shape
// layer only, so each case must parse exactly when the shape layer accepts it:
// a valid case, or one only the Go validator refuses.
const manifest = YAML.parse(await fs.readFile(new URL("../../testdata/publishing_corpora.yaml", import.meta.url), "utf8"));
const corpora = manifest.corpora;
assert.ok(Array.isArray(corpora) && corpora.length > 0, "testdata/publishing_corpora.yaml lists no corpus");

for (const path of corpora) {
  const fixture = YAML.parse(await fs.readFile(new URL(`../../${path}`, import.meta.url), "utf8"));
  test(`${path} is a valid corpus`, () => {
    assert.equal(validateCorpus({ cases: fixture.cases }), undefined);
    assert.deepEqual(new Set(fixture.cases.map((row) => row.name)), new Set(fixture.required_names));
  });
  for (const row of fixture.cases) {
    test(`${path}: ${row.name}`, () => {
      const validator = schema[`z${row.input.schema}`];
      assert.equal(typeof validator?.safeParse, "function", `missing public validator for ${row.input.schema}`);
      const shapeAccepts = row.expected.valid || row.expected.rejected_by === "semantics";
      const result = validator.safeParse(row.input.payload);
      assert.equal(result.success, shapeAccepts, result.error?.message);
    });
  }
}
