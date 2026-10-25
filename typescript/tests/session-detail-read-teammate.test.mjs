import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";

// A session_detail read carries the optional teammate identity beside the
// durable detail. The hand-written read normalization must preserve it, and it
// must preserve the three distinct absences exactly: an omitted key stays
// omitted, an own null stays null, and an invalid value is refused. Testing only
// a populated read would let a truthiness-based copy pass while silently
// collapsing the omitted and null cases.
//
// The raw WebSocket text lives in the fixture so the probe runs the production
// entry point a browser uses (parseServerMessageRaw) rather than a
// re-implementation of the normalization.

const fixture = parse(await readFile(new URL("../../testdata/typescript/session_detail_read_runtime.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture).sort(), ["cases", "required_names", "session_id"]);
assert.equal(typeof fixture.session_id, "string");
assert.ok(Array.isArray(fixture.cases) && fixture.cases.length > 0, "cases must be a non-empty sequence");
assert.equal(new Set(fixture.cases.map((row) => row.name)).size, fixture.cases.length, "case names must be unique");
assert.deepEqual(new Set(fixture.cases.map((row) => row.name)), new Set(fixture.required_names), "every named case must exist");

for (const row of fixture.cases) {
  test(`session detail read carries the teammate identity: ${row.name}`, () => {
    assert.equal(typeof row.message, "string", `${row.name} must carry the raw session_detail text`);
    const expected = row.expected;
    assert.equal(typeof expected.valid, "boolean", `${row.name} must declare its verdict`);

    const read = () => {
      const message = schema.parseServerMessageRaw(row.message);
      assert.equal(message.type, "session_detail");
      return message.data;
    };

    if (!expected.valid) {
      assert.equal(typeof expected.error_contains, "string", `${row.name} must name the refusal fragment`);
      assert.throws(read, (error) => error.message.includes(expected.error_contains), `${row.name} must be refused by the read parser`);
      return;
    }

    const data = read();
    assert.equal(data.id, fixture.session_id, `${row.name} lost the durable detail identity`);
    if (expected.teammate_absent) {
      assert.equal(Object.hasOwn(data, "teammate"), false, `${row.name} fabricated a teammate the read did not carry`);
      return;
    }
    if (expected.teammate_null) {
      assert.equal(Object.hasOwn(data, "teammate"), true, `${row.name} dropped an own null teammate`);
      assert.equal(data.teammate, null, `${row.name} changed an own null teammate`);
      return;
    }
    assert.deepEqual(data.teammate, expected.teammate, `${row.name} lost or changed the teammate identity`);
  });
}
