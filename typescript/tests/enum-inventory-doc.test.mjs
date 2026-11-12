import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import { renderEnumInventoryDoc } from "../scripts/lib/enum-inventory-doc.mjs";

const fixture = parse(await readFile(new URL("../../testdata/typescript/enum_inventory_docs.yaml", import.meta.url), "utf8"));
const catalog = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));

test("enum inventory documentation follows its fixture", async (t) => {
  const names = fixture.cases.map((row) => row.name);
  assert.equal(new Set(names).size, names.length, "case names must be unique");
  assert.deepEqual(new Set(names), new Set(fixture.required_names));
  assert.deepEqual(new Set(fixture.required_names), new Set([
    "harness-catalog-doc-remains-unchanged",
    "another-subset-uses-generic-doc",
    "complete-inventory-has-no-subset-doc",
    "subset-doc-cannot-close-comment",
  ]));
  for (const row of fixture.cases) {
    await t.test(row.name, () => {
      assert.ok(row.provenance.source && row.provenance.ref && row.mutation.description);
      const enumCase = row.input.catalog_enum === undefined
        ? row.input.enum
        : catalog.enums.find((entry) => entry.name === row.input.catalog_enum);
      assert.ok(enumCase, `enum fixture for ${row.name} is missing`);
      if (row.classification === "must-fail") {
        assert.throws(() => renderEnumInventoryDoc(enumCase), (error) => error instanceof Error && error.message.includes(row.expected));
      } else {
        assert.equal(row.classification, "must-pass");
        assert.equal(renderEnumInventoryDoc(enumCase), row.expected);
      }
    });
  }
});
