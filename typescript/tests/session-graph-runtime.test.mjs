import assert from "node:assert/strict";
import test from "node:test";

import { parseServerMessageRaw, parseSessionDetailPayloadText, parseTranscriptContentText } from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixtures = sessionGraphFixtures();

test("canonical aggregate native budgets pass through every public envelope", async (t) => {
  assert.deepEqual(new Set(fixtures.native_limits.cases.map(row => row.name)), new Set(fixtures.native_limit_required_names));
  for (const row of fixtures.native_limits.cases) await t.test(row.name, () => {
    const recipe = row.input;
    const detail = JSON.parse(recipe.detailJSON);
    let offset = 0;
    const records = (count) => Array.from({length: count}, () => {
      const record = JSON.parse(recipe.record_json);
      const suffix = String(offset++);
      record.id += suffix;
      record.source.entryRef += suffix;
      const chunkCount = Math.floor((recipe.data_bytes + 15002) / 15003);
      let remaining = recipe.data_bytes - 1 - 3 * chunkCount;
      record.data = [];
      while (remaining > 0) {
        const size = Math.min(remaining, 15000);
        record.data.push("x".repeat(size));
        remaining -= size;
      }
      assert.equal(new TextEncoder().encode(JSON.stringify(record.data)).length, recipe.data_bytes);
      return record;
    });
    detail.nativeMetadata = records(recipe.main_records);
    detail.earlierHistory = recipe.earlier_records.map(count => ({...JSON.parse(recipe.section_json), nativeMetadata: records(count)}));
    assertPublicExits(row, detail);
  });
});

test("generated durable graph schema consumes the canonical count and recursive corpus", async (t) => {
  const arms = [fixtures.durable.round_trip, fixtures.durable.counts, fixtures.durable.invalid_counts, fixtures.recursive, fixtures.raw_durable];
  for (const arm of arms) for (const row of arm.cases) await t.test(row.name, () => {
    const raw = row.input.json ?? row.input.rawJSON;
    let result;
    try { result = parseSessionDetailPayloadText(raw); }
    catch (error) { assert.equal(row.classification, "must-fail", String(error)); return; }
    assert.equal(row.classification, "must-pass");
    if (Object.hasOwn(JSON.parse(raw), "inputSubmissionCount")) {
      assert.equal(Object.hasOwn(result, "inputSubmissionCount"), true);
      assert.equal(result.inputSubmissionCount, JSON.parse(raw).inputSubmissionCount);
    }
  });
});

test("public envelope parsers apply durable graph semantics before generated parsing", async (t) => {
  const arms = [fixtures.durable.round_trip, fixtures.durable.counts, fixtures.durable.invalid_counts, fixtures.recursive, fixtures.raw_durable];
  for (const arm of arms) for (const row of arm.cases) await t.test(row.name, () => {
    const raw = row.input.json ?? row.input.rawJSON;
    const detail = JSON.parse(raw);
    if (row.name === "raw-forbidden-relationship-navigation") {
      assert.throws(() => parseTranscriptContentText(JSON.stringify({kind: "session_detail", contractVersion: "1", sessionDetail: detail})));
      assert.doesNotThrow(() => parseServerMessageRaw(JSON.stringify({type: "session_detail", data: detail})));
    } else assertPublicExits(row, detail);
  });
});

function assertPublicExits(row, detail) {
  const original = structuredClone(detail);
  const text = JSON.stringify(detail);
  const direct = () => parseSessionDetailPayloadText(text);
  const content = () => parseTranscriptContentText(JSON.stringify({kind: "session_detail", contractVersion: "1", sessionDetail: detail})).sessionDetail;
  const websocket = () => parseServerMessageRaw(JSON.stringify({type: "session_detail", data: detail})).data;
  if (row.classification === "must-fail") {
    assert.throws(direct); assert.throws(content); assert.throws(websocket);
  } else {
    const directResult = direct(), contentResult = content(), websocketResult = websocket();
    assert.deepEqual(contentResult, directResult);
    assert.deepEqual(websocketResult, directResult);
    if (text.includes('"submissionRef":""')) assert.equal(JSON.stringify(directResult).includes('"submissionRef":""'), false);
    if (row.name.startsWith("raw-empty-")) {
      assert.doesNotMatch(JSON.stringify(directResult), /"(?:sourceEntryRef|sourceRevisionRef|callEntryRef|resultEntryRef|submissionRef)":""/, "optional empty references must be omitted on every exit");
    }
  }
  assert.deepEqual(detail, original);
}

test("primitive relationship and provenance corpora pass through every public detail exit", async (t) => {
  const baseRow = fixtures.recursive.cases.find(row => row.classification === "must-pass");
  const base = JSON.parse(baseRow.input.json ?? baseRow.input.rawJSON);
  for (const row of [...fixtures.relationships.cases, ...fixtures.relationship_sets.cases]) await t.test(row.name, () => {
    const detail = {...base, relationships: row.input.relationships ?? [row.input]};
    assertPublicExits(row, detail);
  });
  for (const row of fixtures.provenance.cases) await t.test(row.name, () => {
    const detail = structuredClone(base);
    const turn = detail.turns[0] ?? detail.earlierHistory[0].turns[0];
    turn.provenance = row.input;
    assertPublicExits(row, detail);
    if (row.name === "provenance-invalid-submission" && row.classification === "must-pass") {
      assert.equal(turn.provenance.submissionRef, "");
    }
  });
});
