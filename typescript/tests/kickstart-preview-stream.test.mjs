import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import { mergeLocalKickstartPreview, parseLocalKickstartPreviewRequest, parseLocalKickstartPreviewValue } from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/local-api/kickstart_preview_stream.yaml", import.meta.url), "utf8"), { merge: true });
const manifest = parse(await readFile(new URL("../../testdata/local-api/kickstart_preview_stream_manifest.yaml", import.meta.url), "utf8"));

test("preview stream fixture retains every required case", () => {
  assert.deepEqual(fixture.cases.map((c) => c.name).sort(), [...manifest.requiredNames].sort());
});

for (const c of fixture.cases) {
  test(c.name, () => {
    let state;
    const run = () => {
      for (const [i, name] of c.frames.entries()) {
        const frame = structuredClone(fixture.frames[name]);
        if (c.mutate?.frame === i) {
          const path = c.mutate.path.split(".");
          let target = frame;
          for (const part of path.slice(0, -1)) target = target[part];
          target[path.at(-1)] = c.mutate.value;
        }
        const oldSegments = structuredClone(state?.segments ?? []);
        state = mergeLocalKickstartPreview(state, frame);
        assert.deepEqual(state.segments.slice(0, oldSegments.length), oldSegments);
      }
    };
    if (c.producerOnly !== undefined) assert.ok(c.producerOnly.length > 0);
    if (c.tsValid ?? c.valid) {
      assert.doesNotThrow(run);
      if (c.expectSegments !== undefined) assert.equal(state.segments.length, c.expectSegments);
      if (c.expectUnknown !== undefined) assert.equal(state.segments.flatMap((s) => s.retainedUnknown ?? []).length, c.expectUnknown);
      if (c.expectMetadata !== undefined) assert.equal(state.segments.flatMap((s) => s.nativeMetadata ?? []).length, c.expectMetadata);
      if (c.expectHasNewer !== undefined) assert.equal(state.segments.at(-1).hasNewer, c.expectHasNewer);
      if (c.expectSourceBytesLoaded !== undefined) assert.equal(state.directory.sourceBytesLoaded, c.expectSourceBytesLoaded);
      if (c.expectSourceBytesTotal !== undefined) assert.equal(state.directory.sourceBytesTotal, c.expectSourceBytesTotal);
    } else assert.throws(run, (error) => !error.message.includes("ses_test") && !error.message.includes("/private/"));
  });
}

const requestManifest = parse(await readFile(new URL("../../testdata/local-api/kickstart_preview_requests_manifest.yaml", import.meta.url), "utf8"));
test("preview request fixture retains every required case", () => {
  assert.deepEqual(fixture.requestCases.map((c) => c.name).sort(), [...requestManifest.requiredNames].sort());
});
for (const c of fixture.requestCases) {
  test(c.name, () => {
    const request = structuredClone(c.request);
    if (c.repeatedNodeBytes) request.nodeId = "\u0001".repeat(c.repeatedNodeBytes);
    if (c.maxEnvelope) {
      request.requestId = "r".repeat(32);
      request.sessionId = "s".repeat(128);
    }
    if (c.valid) assert.doesNotThrow(() => parseLocalKickstartPreviewRequest(request));
    else assert.throws(() => parseLocalKickstartPreviewRequest(request));
  });
}

const httpManifest = parse(await readFile(new URL("../../testdata/local-api/kickstart_preview_http_manifest.yaml", import.meta.url), "utf8"));
test("HTTP preview fixture retains every required case", () => {
  assert.deepEqual(fixture.previewCases.map((c) => c.name).sort(), [...httpManifest.requiredNames].sort());
});
for (const c of fixture.previewCases) {
  test(c.name, () => {
    if (c.valid) assert.doesNotThrow(() => parseLocalKickstartPreviewValue(c.preview));
    else assert.throws(() => parseLocalKickstartPreviewValue(c.preview));
  });
}
