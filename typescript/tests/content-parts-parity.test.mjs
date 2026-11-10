import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import { parseSessionDetailPayloadValue, parseTranscriptContentText } from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/content_parts.yaml", import.meta.url), "utf8"));
const detailFixture = parse(await readFile(new URL("../../testdata/content_parts_detail.yaml", import.meta.url), "utf8"));

function detailWithPart(part) {
  return {
    id: "fixture-session",
    harness: "claude-code",
    startTime: "2020-01-01T00:00:00Z",
    endTime: "2020-01-01T00:00:00Z",
    durationMins: 0,
    totalTokens: 0,
    tokensIn: 0,
    tokensOut: 0,
    turnCount: 1,
    toolCallCount: 0,
    outcome: "resolved",
    sessionOrigin: "unknown",
    turns: [{index: 0, role: "user", content: "", depth: 0, timestamp: "2020-01-01T00:00:00Z", contentParts: [part]}],
  };
}

function parseInMode(detail, mode) {
  if (mode === "local") return parseSessionDetailPayloadValue(detail);
  if (mode === "published") return parseTranscriptContentText(JSON.stringify({contractVersion: "1.0.0", kind: "session_detail", sessionDetail: detail}));
  throw new TypeError(`unknown content-part mode ${JSON.stringify(mode)}`);
}

test("the shared content-part corpus keeps TypeScript and Go parity", async (t) => {
  for (const row of fixture.cases) {
    await t.test(row.name, () => {
      const detail = detailWithPart(JSON.parse(row.input.part));
      const run = () => parseInMode(detail, row.input.mode);
      if (row.expected.accepted) assert.doesNotThrow(run);
      else assert.throws(run);
    });
  }
});

test("the detail-level corpus keeps TypeScript and Go parity", async (t) => {
  for (const row of detailFixture.cases) {
    await t.test(row.name, () => {
      const run = () => parseInMode(JSON.parse(row.input.detail), row.input.mode);
      if (row.expected.accepted) assert.doesNotThrow(run);
      else assert.throws(run);
    });
  }
});

test("a published envelope refuses local image evidence", () => {
  const localInline = detailWithPart({
    type: "image",
    image: {
      ref: "a".repeat(64),
      mediaType: "image/png",
      byteSize: 2048,
    },
  });
  assert.doesNotThrow(() => parseSessionDetailPayloadValue(localInline), "a bare local detail keeps the local image shape");
  assert.throws(
    () => parseTranscriptContentText(JSON.stringify({contractVersion: "1.0.0", kind: "session_detail", sessionDetail: localInline})),
    /Session detail validation failed at @peasant-labs\/schema public parser/,
    "the publication envelope must refuse a local image part",
  );
});

test("a local detail refuses the published image shape", () => {
  const published = detailWithPart({type: "image", image: {omittedReason: "not_published"}});
  assert.doesNotThrow(() => parseTranscriptContentText(JSON.stringify({contractVersion: "1.0.0", kind: "session_detail", sessionDetail: published})));
  assert.throws(
    () => parseSessionDetailPayloadValue(published),
    /Session detail validation failed at @peasant-labs\/schema public parser/,
    "a durable local detail must refuse the not_published image reason",
  );
});
