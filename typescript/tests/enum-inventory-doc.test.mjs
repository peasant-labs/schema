import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import { renderEnumInventoryDoc } from "../scripts/lib/enum-inventory-doc.mjs";

const fixture = parse(await readFile(new URL("../../testdata/typescript/enum_inventory_docs.yaml", import.meta.url), "utf8"));
const manifest = parse(await readFile(new URL("../../testdata/typescript/enum_inventory_docs_manifest.yaml", import.meta.url), "utf8"));
const catalog = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));

function uniqueNames(names, label) {
  const seen = new Set();
  for (const name of names) {
    assert.ok(typeof name === "string" && name !== "" && name.trim() === name, `${label}: invalid case name ${JSON.stringify(name)}`);
    assert.ok(!seen.has(name), `${label}: repeated case name ${JSON.stringify(name)}`);
    seen.add(name);
  }
  return seen;
}

test("enum inventory documentation follows its fixture", async (t) => {
  const names = fixture.cases.map((row) => row.name);
  assert.notDeepEqual(names, [], "enum inventory documentation cases must not be empty");
  assert.ok(Array.isArray(manifest.requiredNames), "enum inventory documentation manifest must list requiredNames");
  assert.deepEqual(uniqueNames(names, "cases"), uniqueNames(manifest.requiredNames, "manifest"), "case names must match the manifest");
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
