import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

import { parse } from "yaml";
import { z } from "zod";

import {
  parseSessionDetailPayloadText,
  parseTranscriptContentText,
  parseTranscriptDownloadText,
  transcriptContentIsStrict,
} from "../dist/index.js";

const input = z.strictObject({
  subject: z.enum(["download-read", "source-entry", "source-detail-entry"]),
  replay: z.array(z.enum(["go", "ts"])).min(1),
  raw: z.string().optional(),
  depth: z.int().nonnegative().optional(),
  keyBytes: z.int().nonnegative().optional(),
  bytes: z.int().nonnegative().optional(),
  syntheticBytes: z.boolean().optional(),
});
const row = z.strictObject({
  name: z.string().min(1),
  classification: z.enum(["must-pass", "must-fail"]),
  provenance: z.strictObject({source: z.string(), ref: z.string().min(1)}),
  mutation: z.strictObject({description: z.string().min(1)}),
  input,
  expected: z.strictObject({
    outcome: z.enum(["strict", "legacy", "records", "refused"]),
    failure: z.string().optional(),
    contains: z.string().optional(),
  }),
});
const fixtures = z.strictObject({cases: z.array(row).min(1)}).parse(parse(
  readFileSync(new URL("../../testdata/transcript_download_text.yaml", import.meta.url), "utf8"), {uniqueKeys: true},
));
const manifest = z.strictObject({requiredNames: z.array(z.string()).min(1)}).parse(parse(
  readFileSync(new URL("../../testdata/transcript_download_text_manifest.yaml", import.meta.url), "utf8"), {uniqueKeys: true},
));
assert.deepEqual(fixtures.cases.map((item) => item.name).sort(), [...manifest.requiredNames].sort());
assert.equal(new Set(fixtures.cases.map((item) => item.name)).size, fixtures.cases.length);

function expand(item) {
  if (item.syntheticBytes) {
    const prefix = '{"padding":"';
    const suffix = '","sessionDetail":{"id":"legacy","turns":[]},"kind":"session_detail","contractVersion":"0.1.1"}';
    const text = prefix + "x".repeat(item.bytes - prefix.length - suffix.length) + suffix;
    assert.equal(text.length, item.bytes);
    return text;
  }
  let text = (item.raw ?? "").replaceAll("{{key}}", "k".repeat(item.keyBytes ?? 0));
  if (item.depth) text = "[".repeat(item.depth) + text + "]".repeat(item.depth);
  return text;
}

for (const item of fixtures.cases) {
  if (!item.input.replay.includes("ts")) continue;
  test(item.name, () => {
    const text = expand(item.input);
    const run = () => {
      switch (item.input.subject) {
        case "download-read": return parseTranscriptDownloadText(text);
        case "source-entry": return parseTranscriptContentText(text);
        case "source-detail-entry": return parseSessionDetailPayloadText(text);
        default: assert.fail(`unknown subject ${item.input.subject}`);
      }
    };
    if (item.expected.outcome === "refused") {
      assert.throws(run, (error) => {
        assert.ok(error instanceof Error);
        if (item.input.subject === "download-read") {
          assert.ok(error.message.startsWith(`${item.expected.failure}:`), error.message);
        } else {
          assert.ok(error.message.includes(item.expected.contains), error.message);
        }
        return true;
      });
      return;
    }
    const result = run();
    assert.equal(result.kind, item.expected.outcome);
    if (result.kind === "records") {
      assert.deepEqual(result.records, text.split("\n").filter((line) => line.trim() !== "").map((line) => JSON.parse(line)));
      assert.equal("detail" in result, false);
      return;
    }
    const served = JSON.parse(text);
    const detail = served !== null && typeof served === "object" && Object.hasOwn(served, "sessionDetail")
      ? served.sessionDetail : served;
    assert.deepEqual(result.detail, detail);
    assert.equal("records" in result, false);
    assert.equal(transcriptContentIsStrict(detail), item.expected.outcome === "strict");
  });
}
