import assert from "node:assert/strict";
import {readFile} from "node:fs/promises";
import test from "node:test";
import {parse} from "yaml";
import {AllHarnesses, parseSessionDetailPayloadText, parseSessionDetailPayloadValue, parseTranscriptContentText, requiredContentCapabilities, missingContentCapabilities, zSessionDetailPayload} from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/retained_unknown.yaml", import.meta.url), "utf8"), {merge: true});
assert.deepEqual(new Set(fixture.cases.map(row => row.name)), new Set(fixture.required_names));
assert.deepEqual(new Set(fixture.required_harnesses), new Set(AllHarnesses));

function input(recipe) {
  const detail = JSON.parse(fixture.base_detail), record = {...JSON.parse(fixture.base_record), ...JSON.parse(recipe.record_patch ?? "{}")};
  if (recipe.drop_record_field) delete record[recipe.drop_record_field];
  if (recipe.payload_bytes) record.payload = JSON.stringify("x".repeat(recipe.payload_bytes));
  detail.retainedUnknown = [record];
  if (recipe.second_record_patch !== undefined) detail.retainedUnknown.push({...JSON.parse(fixture.base_record), ...JSON.parse(recipe.second_record_patch)});
  if (recipe.harness) detail.harness = recipe.harness;
  if (recipe.drop_retention) delete detail.retainedUnknown;
  if (recipe.drop_diagnostics) delete detail.diagnostics;
  return Object.assign(detail, JSON.parse(recipe.detail_patch ?? "{}"));
}

test("retained evidence survives public TypeScript boundaries without numeric loss", async t => {
  for (const row of fixture.cases) await t.test(row.name, () => {
    const detail = input(row.input), text = JSON.stringify(detail);
    assert.equal(zSessionDetailPayload.safeParse(detail).success, row.expected.shape_valid, "generated shape");
    if (!row.expected.valid) {
      assert.throws(() => parseSessionDetailPayloadText(text));
      assert.throws(() => parseSessionDetailPayloadValue(detail));
      assert.throws(() => parseTranscriptContentText(JSON.stringify({contractVersion: "1.0.0", kind: "session_detail", sessionDetail: detail})));
      return;
    }
    const parsed = parseSessionDetailPayloadText(text);
    assert.deepEqual(parsed.retainedUnknown, detail.retainedUnknown);
    assert.deepEqual(parsed.diagnostics, detail.diagnostics);
    const envelope = parseTranscriptContentText(JSON.stringify({contractVersion: "1.0.0", kind: "session_detail", sessionDetail: detail}));
    assert.deepEqual(envelope.sessionDetail, parsed);
    assert.deepEqual(parseSessionDetailPayloadValue(parsed), parsed);
    const required = requiredContentCapabilities(parsed);
    assert.equal(required.includes("retained_unknown_v1"), row.expected.capability);
    assert.deepEqual(missingContentCapabilities([], required), required);
    assert.deepEqual(missingContentCapabilities(["future_v9", "retained_unknown_v1"], required), []);
  });
});
