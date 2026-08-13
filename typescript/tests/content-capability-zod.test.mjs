import assert from "node:assert/strict";
import test from "node:test";

import * as schema from "../dist/index.js";

test("discovery Zod parsing is forward-open and preserves unknown and duplicate strings", () => {
  const omitted = schema.zSchemaVersionResponse.safeParse({
    annotationSchemaVersion: "1", supportedTargetKinds: [], supportedTypeIds: [],
    pushContractVersion: "1", minPushContractVersion: "1",
  });
  assert.equal(omitted.success, true);

  const tokens = ["future_v9", "observed_model_v1", "observed_model_v1"];
  const discovered = schema.zSchemaVersionResponse.safeParse({
    annotationSchemaVersion: "1", supportedTargetKinds: [], supportedTypeIds: [],
    pushContractVersion: "1", minPushContractVersion: "1", contentCapabilities: tokens,
  });
  assert.equal(discovered.success, true);
  assert.deepEqual(discovered.data.contentCapabilities, tokens);

  const empty = schema.zSchemaVersionResponse.safeParse({
    annotationSchemaVersion: "1", supportedTargetKinds: [], supportedTypeIds: [],
    pushContractVersion: "1", minPushContractVersion: "1", contentCapabilities: [],
  });
  assert.equal(empty.success, true);

  const withNull = schema.zSchemaVersionResponse.safeParse({
    annotationSchemaVersion: "1", supportedTargetKinds: [], supportedTypeIds: [],
    pushContractVersion: "1", minPushContractVersion: "1", contentCapabilities: null,
  });
  assert.equal(withNull.success, false);
});
