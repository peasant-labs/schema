import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import * as contract from "../dist/index.js";

// This suite runs the local-account closed-body and round-trip corpus against the
// BUILT root exports, not the generated source. It exists because a plainly
// rendered `z.object` strips unknown keys and reports success, so a consumer of
// the npm package would receive a validator that accepts a body the Go decoder
// and the emitted OpenAPI component (additionalProperties: false) both refuse.
// The Go side asserts the spec is closed; this side asserts the runtime
// validator actually rejects what the spec closes, so the two bindings cannot
// drift apart behind a green generation gate.
const fixtureSource = await readFile(new URL("../../testdata/typescript/village_auth_zod.yaml", import.meta.url), "utf8");
const fixture = parse(fixtureSource);

const strictBodies = fixture?.strict_bodies;
const roundTrip = fixture?.round_trip;
assert.ok(Array.isArray(strictBodies) && strictBodies.length >= 8, "the local-account strict-body corpus is missing or too small; this suite would pass while checking nothing");
assert.ok(Array.isArray(roundTrip) && roundTrip.length >= 3, "the local-account round-trip corpus is missing or too small; this suite would pass while checking nothing");

test("every closed local-account request body rejects the unknown field it declares", async (t) => {
  for (const row of strictBodies) {
    await t.test(row.name, () => {
      const schema = contract[row.component];
      assert.ok(schema?.safeParse, `${row.component} is not exported as a runtime schema from the root facade`);
      assert.equal(schema.safeParse(row.canonical).success, true, `${row.component} rejected its own canonical body`);
      const withUnknown = { ...row.canonical, [row.unknown_field]: "must not be stripped" };
      assert.equal(schema.safeParse(withUnknown).success, false, `${row.component} accepted and stripped the unknown field ${row.unknown_field}`);
    });
  }
});

test("every local-account round-trip value survives a parse and re-encode unchanged", async (t) => {
  for (const row of roundTrip) {
    await t.test(row.name, () => {
      const schema = contract[row.component];
      assert.ok(schema?.safeParse, `${row.component} is not exported as a runtime schema from the root facade`);
      const parsed = schema.safeParse(row.value);
      assert.equal(parsed.success, true, `${row.component} rejected its canonical value`);
      assert.deepEqual(parsed.data, row.value, `${row.component} changed the value it round-tripped`);
    });
  }
});
