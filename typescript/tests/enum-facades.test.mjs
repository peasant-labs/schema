import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as root from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["enums"]);
assert.ok(Array.isArray(fixture.enums), "enums.yaml must provide an enums sequence for runtime facade verification");
assert.ok(fixture.enums.length > 0, "enums.yaml must provide at least one enum facade for runtime verification");

test("built root module exposes every catalog enum facade", async (t) => {
  for (const enumCase of fixture.enums) {
    await t.test(enumCase.name, () => {
      const guardName = `is${enumCase.name}`;
      assert.equal(Object.hasOwn(root, enumCase.name), true, `root module does not export the ${enumCase.name} enum facade`);
      assert.equal(Object.hasOwn(root, enumCase.all_name), true, `root module does not export the ${enumCase.all_name} inventory`);
      assert.equal(Object.hasOwn(root, guardName), true, `root module does not export the ${guardName} guard`);

      assert.equal(typeof root[enumCase.name], "object", `${enumCase.name} is not an enum value facade`);
      assert.equal(Array.isArray(root[enumCase.all_name]), true, `${enumCase.all_name} is not an enum inventory`);
      assert.equal(typeof root[guardName], "function", `${guardName} is not an enum guard`);
    });
  }
});
