import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { parse } from "yaml";
import { z } from "zod";
import {
  scanRawJsonText, parseSessionDetailPayloadText, parseSessionDetailPayloadValue, parseTranscriptContentText,
  parseServerMessageRaw, parseNativeMetadataRecordsText, parseNativeMetadataRecordsValue,
} from "../dist/index.js";

const generation = z.strictObject({
  unit: z.string(), repeat: z.int().nonnegative(), count: z.int().nonnegative().optional(),
  depth: z.int().nonnegative().optional(), whitespace: z.int().nonnegative().optional(), copies: z.int().nonnegative().optional(),
});
const row = z.strictObject({
  name: z.string().min(1), operation: z.enum(["scanner", "metadata", "detail", "turn", "metadata-records", "metadata-value", "owner-unicode"]),
  raw: z.string(), accepted: z.boolean(), errorCategory: z.enum(["lexical", "validation"]).optional(),
  errorContains: z.string().optional(), targets: z.string().optional(), pointers: z.array(z.string()).optional(),
  goOwnerBytesHex: z.string().regex(/^(?:[0-9a-f]{2})+$/).optional(),
  maxDepth: z.int().positive().optional(), generate: generation.optional(),
}).refine(value => value.accepted || value.errorCategory !== undefined, "rejections need an error category");
const fixture = z.strictObject({requiredCaseNames: z.array(z.string().min(1)), cases: z.array(row)}).parse(
  parse(readFileSync(new URL("../../testdata/contract/raw_json_boundaries.yaml", import.meta.url), "utf8"), {uniqueKeys: true}),
);
const names = new Set();
for (const item of fixture.cases) {
  assert(!names.has(item.name), `duplicate fixture name ${item.name}`);
  names.add(item.name);
}
assert.equal(new Set(fixture.requiredCaseNames).size, fixture.requiredCaseNames.length, "duplicate required name");
assert.deepEqual(names, new Set(fixture.requiredCaseNames), "required manifest must name every case exactly once");

function expand(item) {
  let raw = item.raw;
  const g = item.generate;
  if (g) {
    let value = g.unit.repeat(g.repeat);
    if (g.count) value = `[${Array(g.count).fill(`"${value}"`).join("," + " ".repeat(g.whitespace ?? 0))}]`;
    for (let i = 0; i < (g.depth ?? 0); i++) value = `{"x":${value}}`;
    raw = raw.replaceAll("{{value}}", value);
    if (g.copies) raw = `[${Array.from({length: g.copies}, (_, i) => raw.replaceAll("{{index}}", String(i))).join(",")}]`;
  }
  return raw;
}

function detail(turns, metadata = "", harness = "claude-code") {
  return `{"id":"s","harness":"${harness}","startTime":"2020-01-01T00:00:00Z","endTime":"2020-01-01T00:00:00Z","durationMins":0,"totalTokens":0,"tokensIn":0,"tokensOut":0,"turnCount":0,"toolCallCount":0,"turns":${turns}${metadata ? `,"nativeMetadata":${metadata}` : ""}}`;
}

function outcome(item, run, expected) {
  if (item.accepted) {
    const actual = run();
    if (actual !== undefined) assert.deepEqual(actual, JSON.parse(expected), "accepted output must preserve costs, refs and metadata");
    return;
  }
  assert.throws(run, error => {
    assert.equal(error.message.startsWith("Raw JSON validation failed") ? "lexical" : "validation", item.errorCategory, error.message);
    assert(error.message.toLowerCase().includes((item.errorContains ?? "").toLowerCase()), error.message);
    return true;
  });
}

async function detailExits(t, item, raw) {
  await t.test("detail", () => outcome(item, () => parseSessionDetailPayloadText(raw), raw));
  await t.test("transcript", () => {
    const envelope = `{"kind":"session_detail","contractVersion":"1.0.0","sessionDetail":${raw}}`;
    outcome(item, () => parseTranscriptContentText(envelope), envelope);
  });
  await t.test("websocket", () => {
    const envelope = `{"data":${raw},"type":"session_detail"}`;
    outcome(item, () => parseServerMessageRaw(envelope), envelope);
  });
  if (item.errorCategory !== "lexical") await t.test("value", () => outcome(item, () => parseSessionDetailPayloadValue(JSON.parse(raw)), raw));
}

for (const item of fixture.cases) test(item.name, async t => {
  const raw = expand(item);
  switch (item.operation) {
    case "scanner": outcome(item, () => scanRawJsonText(raw, {maxDocumentBytes: 8 << 20, maxDocumentDepth: item.maxDepth ?? 64, opaqueMetadataPointers: item.pointers}), raw); break;
    case "metadata": outcome(item, () => { scanRawJsonText(raw, {maxDocumentBytes: 65536, maxDocumentDepth: 32, opaqueMetadataPointers: [""]}); return JSON.parse(raw); }, raw); break;
    case "turn": await detailExits(t, item, detail(`[${raw}]`)); break;
    case "detail": await detailExits(t, item, raw); break;
    case "owner-unicode": outcome(item, () => parseSessionDetailPayloadValue(JSON.parse(detail(`[${raw}]`))), raw); break;
    case "metadata-value": outcome(item, () => parseNativeMetadataRecordsValue(JSON.parse(raw), JSON.parse(item.targets)), raw); break;
    case "metadata-records": {
      const targets = JSON.parse(item.targets);
      outcome(item, () => parseNativeMetadataRecordsText(raw, targets), raw);
      if (item.errorCategory !== "lexical") outcome(item, () => parseNativeMetadataRecordsValue(JSON.parse(raw), targets), raw);
      if (item.errorCategory === "lexical") await t.test("raw-public-bounds", async t => detailExits(t, item, detail(item.targets, raw)));
      await t.test("pi-public-roots", async t => {
        await detailExits(t, item, detail(item.targets, raw, "pi"));
      });
      break;
    }
    default: assert.fail(`unsupported fixture operation ${item.operation}`);
  }
});
