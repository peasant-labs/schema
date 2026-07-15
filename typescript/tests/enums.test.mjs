import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["enums"]);

test("runtime closed sets match the exact Go-derived catalog", async (t) => {
  for (const enumCase of fixture.enums) {
    await t.test(enumCase.name, () => {
      const expectedMembers = Object.fromEntries(enumCase.members.map((member) => [member.name, member.value]));
      assert.deepEqual(schema[enumCase.name], expectedMembers);
      assert.deepEqual(schema[enumCase.all_name], enumCase.all_values);
      const guard = schema[`is${enumCase.name}`];
      for (const member of enumCase.members) assert.equal(guard(member.value), true);
      assert.equal(guard("unknown-enum-sentinel"), false);
      assert.equal(guard(undefined), false);
    });
  }
});
