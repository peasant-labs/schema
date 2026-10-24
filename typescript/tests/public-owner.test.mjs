import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import * as schema from "../dist/index.js";
import { checkMin, validateCorpus } from "../dist/testcase.js";

// The same person corpus the Go decoder reads. The generated validator must
// read every case the way the Go reader does: the declared shapes exactly,
// and any kind a later version adds as anonymous, keeping no identity from it.
const fixture = parse(await readFile(new URL("../../testdata/village_public_owner.yaml", import.meta.url), "utf8"));

test("generated VillagePublicOwner reads every kind fail closed", async (t) => {
  assert.equal(validateCorpus(fixture.decode), undefined);
  assert.equal(checkMin(fixture.decode, 12), undefined);
  assert.deepEqual(fixture.decode.cases.map((row) => row.name).sort(), [...fixture.required_names.decode].sort());
  for (const row of fixture.decode.cases) {
    await t.test(row.name, () => {
      const parsed = schema.zVillagePublicOwner.safeParse(JSON.parse(row.input.json));
      assert.equal(parsed.success, row.expected.valid, parsed.success ? "accepted" : parsed.error.message);
      if (!parsed.success) return;
      const owner = parsed.data;
      assert.equal(owner.kind === "profile", row.expected.profile);
      assert.equal(owner.kind === "profile" && owner.id === row.input.viewer, row.expected.is_account);
      if (owner.kind !== "profile") assert.deepEqual(owner, { kind: "anonymous" });
    });
  }
});

test("generated VillagePublicOwner accepts the exact producer shapes", async (t) => {
  for (const row of fixture.encode.cases.filter((candidate) => candidate.expected.valid)) {
    await t.test(row.name, () => {
      const wire = JSON.parse(row.expected.json);
      const parsed = schema.zVillagePublicOwner.safeParse(wire);
      assert.ok(parsed.success);
      assert.deepEqual(parsed.data, wire);
    });
  }
});
