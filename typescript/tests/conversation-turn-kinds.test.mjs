import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import { AllEntryTypes, AllRoles, isConversationTurn } from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/conversation_turn_kinds.yaml", import.meta.url), "utf8"));

test("the conversation turn kind corpus covers every role and entry type once", () => {
  assert.equal(AllRoles.length * AllEntryTypes.length, 32);
  assert.equal(fixture.kinds.cases.length, 32);
  assert.deepEqual([...fixture.required_names].sort(), fixture.kinds.cases.map((row) => row.name).sort());
  const seen = new Set();
  for (const row of fixture.kinds.cases) {
    const key = `${row.input.role}|${row.input.entryType}`;
    assert.equal(seen.has(key), false, `duplicate pair ${key}`);
    seen.add(key);
    assert.equal(isConversationTurn({ role: row.input.role, entryType: row.input.entryType, depth: row.input.depth }), row.expected.counts, row.name);
  }
  for (const role of AllRoles) {
    for (const entryType of AllEntryTypes) {
      assert.equal(seen.has(`${role}|${entryType}`), true, `missing pair ${role}|${entryType}`);
    }
  }
  const yes = new Set(fixture.kinds.cases.filter((row) => row.expected.counts).map((row) => `${row.input.role}|${row.input.entryType}`));
  assert.deepEqual([...yes].sort(), ["assistant|image", "assistant|text", "user|image", "user|text"]);
});

test("depth, missing and unknown kinds do not count", () => {
  for (const pair of [["user", "text"], ["user", "image"], ["assistant", "text"], ["assistant", "image"]]) {
    assert.equal(isConversationTurn({ role: pair[0], entryType: pair[1], depth: 1 }), false, `${pair[0]}|${pair[1]} at depth 1`);
  }
  assert.equal(isConversationTurn({ role: "user", entryType: undefined, depth: 0 }), false);
  assert.equal(isConversationTurn({ role: "future", entryType: "text", depth: 0 }), false);
  assert.equal(isConversationTurn({ role: "user", entryType: "future", depth: 0 }), false);
});
