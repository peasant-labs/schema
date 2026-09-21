import assert from "node:assert/strict";
import {readFile} from "node:fs/promises";
import test from "node:test";
import {parse} from "yaml";
import {AllHarnesses, AllContentCapabilities, parseSessionDetailPayloadText, parseSessionDetailPayloadValue, parseTranscriptContentText, requiredContentCapabilities, missingContentCapabilities, zSessionDetailPayload} from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/retained_unknown.yaml", import.meta.url), "utf8"), {merge: true});
assert.deepEqual(new Set(fixture.cases.map(row => row.name)), new Set(fixture.required_names));
assert.deepEqual(new Set(fixture.required_harnesses), new Set(AllHarnesses));

function input(recipe) {
  const detail = JSON.parse(fixture.base_detail), record = {...JSON.parse(fixture.base_record), ...JSON.parse(recipe.record_patch ?? "{}")};
  if (recipe.drop_record_field) delete record[recipe.drop_record_field];
  if (recipe.payload_bytes) record.payload = JSON.stringify((recipe.payload_unit ?? "x").repeat(recipe.payload_bytes));
  if (recipe.payload_depth) record.payload = (`{${JSON.stringify(recipe.payload_depth_key)}:`).repeat(recipe.payload_depth) + "0" + "}".repeat(recipe.payload_depth);
  detail.retainedUnknown = [record];
  if (recipe.sibling_blocks) detail.retainedUnknown = Array.from({length: recipe.sibling_blocks}, (_, i) => ({...JSON.parse(fixture.base_record), pointer: `/content/${i}`, position: i + 4}));
  if (recipe.second_record_patch !== undefined) detail.retainedUnknown.push({...JSON.parse(fixture.base_record), ...JSON.parse(recipe.second_record_patch)});
  if (recipe.harness) detail.harness = recipe.harness;
  if (recipe.drop_retention) delete detail.retainedUnknown;
  if (recipe.drop_diagnostics) delete detail.diagnostics;
  return Object.assign(detail, JSON.parse(recipe.detail_patch ?? "{}"));
}

test("retained evidence survives public TypeScript boundaries without numeric loss", async t => {
  for (const row of fixture.cases) await t.test(row.name, () => {
    const detail = input(row.input);
    let text = JSON.stringify(detail);
    if (row.input.transport_bytes) text += " ".repeat(row.input.transport_bytes - Buffer.byteLength(text));
    const envelopeText = `{"contractVersion":"1.0.0","kind":"session_detail","sessionDetail":${text}}`;
    if (row.expected.error_excludes) {
      const safeRejection = (parse) => {
        assert.throws(parse, error => {
          assert.ok(String(error).includes(row.expected.error_contains));
          for (const sentinel of row.expected.error_excludes) assert.ok(!String(error).includes(sentinel), "diagnostic disclosed untrusted retained evidence");
          assert.equal(error.cause, undefined, "underlying scanner error must not escape through cause");
          return true;
        });
      };
      const original = structuredClone(detail);
      safeRejection(() => parseSessionDetailPayloadValue(detail));
      safeRejection(() => parseSessionDetailPayloadText(text));
      safeRejection(() => parseTranscriptContentText(envelopeText));
      assert.deepEqual(detail, original, "rejected evidence must not be sanitized or mutated");
    }
    assert.equal(zSessionDetailPayload.safeParse(detail).success, row.expected.shape_valid, "generated shape");
    if (row.expected.value_valid ?? row.expected.valid) assert.doesNotThrow(() => parseSessionDetailPayloadValue(detail));
    else assert.throws(() => parseSessionDetailPayloadValue(detail));
    if (row.expected.envelope_valid ?? row.expected.valid) assert.doesNotThrow(() => parseTranscriptContentText(envelopeText));
    else assert.throws(() => parseTranscriptContentText(envelopeText));
    if (!row.expected.valid) {
      assert.throws(() => parseSessionDetailPayloadText(text));
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

const boundaries = parse(await readFile(new URL("../../testdata/retained_unknown_boundaries.yaml", import.meta.url), "utf8"), {merge: true});
assert.deepEqual(new Set(boundaries.aliases.cases.map(row => row.name)), new Set(boundaries.required_alias_names));
assert.deepEqual(new Set(boundaries.preservation_mutations.cases.map(row => row.name)), new Set(boundaries.required_mutation_names));

function injectAlias(value, recipe, path = recipe.path ?? []) {
  if (!path.length) return recipe.first
    ? {[recipe.key]: JSON.parse(recipe.value), ...value}
    : {...value, [recipe.key]: JSON.parse(recipe.value)};
  const [key, ...tail] = path;
  const clone = structuredClone(value);
  clone[key] = injectAlias(clone[key], recipe, tail);
  return clone;
}

test("canonical wire aliases refuse before evidence decoding", async t => {
  const detail = input({detail_patch: boundaries.base_patch});
  assert.deepEqual(requiredContentCapabilities(parseSessionDetailPayloadValue(detail)), AllContentCapabilities);
  for (const row of boundaries.aliases.cases) await t.test(row.name, () => {
    const envelope = {contractVersion: "1.0.0", kind: "session_detail", sessionDetail: detail};
    const mutated = injectAlias(row.input.envelope ? envelope : detail, row.input);
    if (!row.input.envelope) {
      if (row.expected) {
        assert.deepEqual(requiredContentCapabilities(parseSessionDetailPayloadText(JSON.stringify(mutated))), AllContentCapabilities);
        assert.deepEqual(requiredContentCapabilities(parseSessionDetailPayloadValue(mutated)), AllContentCapabilities);
      } else {
        assert.throws(() => parseSessionDetailPayloadText(JSON.stringify(mutated)), /alias/);
        assert.throws(() => parseSessionDetailPayloadValue(mutated), /alias/);
      }
      envelope.sessionDetail = mutated;
    }
    const rawEnvelope = JSON.stringify(row.input.envelope ? mutated : envelope);
    if (row.expected) assert.doesNotThrow(() => parseTranscriptContentText(rawEnvelope));
    else assert.throws(() => parseTranscriptContentText(rawEnvelope), /alias/);
  });
});

test("original evidence oracle detects truncated or normalized payloads", async t => {
  for (const row of boundaries.preservation_mutations.cases) await t.test(row.name, () => {
    const source = fixture.cases.find(item => item.name === row.input.source_case);
    assert.ok(source);
    const original = input(source.input), actual = parseSessionDetailPayloadText(JSON.stringify(original));
    assert.deepEqual(actual.retainedUnknown, original.retainedUnknown);
    actual.retainedUnknown[0].payload = row.input.replacement;
    assert.equal(JSON.stringify(actual.retainedUnknown) === JSON.stringify(original.retainedUnknown), row.expected);
  });
});

test("escaped-equivalent and plain duplicate recipes remain distinct", () => {
  const escaped = input(fixture.cases.find(row => row.name === "duplicate_payload_keys").input).retainedUnknown[0].payload;
  const plain = input(fixture.cases.find(row => row.name === "plain_duplicate_payload_keys").input).retainedUnknown[0].payload;
  assert.ok(escaped.includes("\\u0061"));
  assert.ok(!plain.includes("\\u0061"));
  assert.notEqual(escaped, plain);
});
