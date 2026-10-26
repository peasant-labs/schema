import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";

// Go and the published JSON Schema bound the teammate identity by Unicode code
// points; JavaScript string length counts UTF-16 units. The generated Zod
// validator must use the code-point measure too, or a 200-astral-character name
// is refused in the browser while the wire accepts it.
//
// The same corpus is read by session_teammate_bounds_test.go, which runs the Go
// predicate and the published JSON Schema over it, so all three layers are held
// to one file.

const fixture = parse(await readFile(new URL("../../testdata/typescript/session_teammate_bounds.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture).sort(), ["bound", "cases", "required_names"]);
assert.ok(Number.isSafeInteger(fixture.bound) && fixture.bound > 0, "bound must be a positive integer");
assert.ok(Array.isArray(fixture.cases.cases) && fixture.cases.cases.length > 0, "cases must be a non-empty sequence");
assert.equal(new Set(fixture.cases.cases.map((row) => row.name)).size, fixture.cases.cases.length, "case names must be unique");
assert.deepEqual(new Set(fixture.cases.cases.map((row) => row.name)), new Set(fixture.required_names), "every named case must exist");

// A corpus of short ASCII names would pass while the metric stayed wrong. Keep
// at least one value inside the code-point bound and over it in UTF-16 units.
const metricProbe = fixture.cases.cases.filter((row) => row.input.code_points <= fixture.bound && row.input.utf16_units > fixture.bound);
assert.ok(metricProbe.length > 0, "no case sits inside the code-point bound but over it in UTF-16 units; the astral-length regression cannot fail");

for (const row of fixture.cases.cases) {
  test(`teammate identity bound parity: ${row.name}`, () => {
    const input = row.input;
    assert.equal([...input.name].length, input.code_points, `${row.name}: fixture code_points disagrees with the name`);
    assert.equal(input.name.length, input.utf16_units, `${row.name}: fixture utf16_units disagrees with the name`);

    const result = schema.zSessionTeammate.safeParse({ team: input.team, name: input.name, evidence: input.evidence });
    assert.equal(result.success, row.expected.valid, `${row.name}: generated Zod accepted=${result.success}, want ${row.expected.valid}: ${result.error?.message}`);
  });
}
