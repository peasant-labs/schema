import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import * as root from "../dist/index.js";
import { collectiveCompatibilityAliases } from "../scripts/lib/collective-compatibility.mjs";

const fixture = parse(await readFile(new URL("../../testdata/typescript/collective_compatibility.yaml", import.meta.url), "utf8"));
const baseline = JSON.parse(await readFile(new URL("../../generated/types-0.30.0.json", import.meta.url), "utf8"));
const current = JSON.parse(await readFile(new URL(`../../generated/types-${root.TypesVersion}.json`, import.meta.url), "utf8"));
const enums = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));
const source = await readFile(new URL("../../collective_compatibility.go", import.meta.url), "utf8");
const byName = rows => [...rows].sort((a,b) => a.name.localeCompare(b.name, "en"));

test("retained npm aliases exactly cover renamed public names in the frozen prior contract", () => {
  assert.equal(fixture.baseline, "a5997e3c5f81da393640972aa00e0ba5d035f82b");
  const removed = Object.keys(baseline.components.schemas).filter(name => !Object.hasOwn(current.components.schemas, name));
  assert.deepEqual(new Set(removed), new Set(fixture.aliases.map(row => row.name)));
  assert.deepEqual(collectiveCompatibilityAliases(source, enums), byName(fixture.aliases));
  // Check the entire prior named schema/enum/predicate surface, not only names
  // included in today's updated export inventory. The prior spec is hash frozen.
  for (const [name, component] of Object.entries(baseline.components.schemas)) {
    assert.equal(typeof root[`z${name}`]?.safeParse, "function", `prior public validator z${name} disappeared`);
    if (component.enum !== undefined) {
      assert.equal(typeof root[name], "object", `prior enum ${name} disappeared`);
      assert.equal(typeof root[`is${name}`], "function", `prior predicate is${name} disappeared`);
    }
  }
});
for (const row of fixture.aliases) test(`deprecated npm alias preserves canonical identity: ${row.name}`, () => {
  assert.equal(root[`z${row.name}`], root[`z${row.canonical}`]);
  if (row.all_name !== undefined) {
    assert.equal(root[row.name], root[row.canonical]);
    assert.equal(root[row.all_name], root[row.canonical_all_name]);
    assert.equal(root[`is${row.name}`], root[`is${row.canonical}`]);
    assert.deepEqual(Object.values(root[row.name]), baseline.components.schemas[row.name].enum);
  }
});
