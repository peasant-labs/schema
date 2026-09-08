import assert from "node:assert/strict";
import test from "node:test";

import { parseServerMessageRaw, parseSessionDetailPayloadText, parseTranscriptContentText } from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixtures = sessionGraphFixtures();

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
    const content = JSON.stringify({kind: "session_detail", contractVersion: "1", sessionDetail: detail});
    const websocket = JSON.stringify({type: "session_detail", data: detail});
    if (row.classification === "must-fail") assert.throws(() => parseTranscriptContentText(content));
    else assert.doesNotThrow(() => parseTranscriptContentText(content));
    if (row.name === "raw-forbidden-relationship-navigation") assert.doesNotThrow(() => parseServerMessageRaw(websocket));
    else if (row.classification === "must-fail") assert.throws(() => parseServerMessageRaw(websocket));
    else assert.doesNotThrow(() => parseServerMessageRaw(websocket));
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
