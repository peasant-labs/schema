import assert from "node:assert/strict";
import test from "node:test";

import { zPublicRevisionRef, zSourceEntryRef, zSubmissionRef } from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixture = sessionGraphFixtures();
const validators = { revision: zPublicRevisionRef, source: zSourceEntryRef, submission: zSubmissionRef };

test("public reference runtime schemas enforce the canonical UTF-8 byte corpus", async (t) => {
  for (const row of fixture.refs.cases) {
    await t.test(row.name, () => {
      const value = Buffer.from(row.input.bytesBase64, "base64").toString("utf8");
      if (row.name.includes("invalid-utf8") || row.name.includes("invalid-bytes")) {
        assert.notDeepEqual(Buffer.from(value, "utf8"), Buffer.from(row.input.bytesBase64, "base64"), "JavaScript strings cannot preserve the fixture's invalid UTF-8 bytes");
        return;
      }
      const result = validators[row.input.alias].safeParse(value);
      assert.equal(result.success, row.classification === "must-pass");
      if (result.success) assert.equal(result.data, value);
    });
  }
});
