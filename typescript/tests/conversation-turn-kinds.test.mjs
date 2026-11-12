import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import { AllEntryTypes, AllRoles, isConversationTurn } from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/conversation_turn_kinds.yaml", import.meta.url), "utf8"));

// A row that omits a key replays as a turn that lacks the field.
const turnOf = (row) => ({ role: row.input.role, entryType: row.input.entryType, depth: row.input.depth });

test("the conversation turn kind corpus covers every role and entry type once", () => {
  assert.deepEqual([...fixture.required_names].sort(), fixture.kinds.cases.map((row) => row.name).sort());
  const seen = new Set();
  for (const row of fixture.kinds.cases) {
    assert.equal(row.input.depth, 0, `${row.name}: the kind corpus decides top-level pairs only`);
    const key = `${row.input.role}|${row.input.entryType}`;
    assert.equal(seen.has(key), false, `duplicate pair ${key}`);
    seen.add(key);
    assert.equal(isConversationTurn(turnOf(row)), row.expected.counts, row.name);
  }
  const listed = AllRoles.flatMap((role) => AllEntryTypes.map((entryType) => `${role}|${entryType}`));
  assert.deepEqual([...seen].sort(), [...listed].sort());
  const yes = new Set(fixture.kinds.cases.filter((row) => row.expected.counts).map((row) => `${row.input.role}|${row.input.entryType}`));
  assert.deepEqual([...yes].sort(), ["assistant|image", "assistant|text", "user|image", "user|text"]);
});

test("nested, missing and unknown kinds do not count", async (t) => {
  assert.deepEqual([...fixture.outside_required_names].sort(), fixture.outside.cases.map((row) => row.name).sort());
  for (const row of fixture.outside.cases) await t.test(row.name, () => {
    assert.equal(isConversationTurn(turnOf(row)), row.expected.counts);
  });
});
