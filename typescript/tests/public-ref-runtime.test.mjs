import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import { zPublicRevisionRef, zSourceEntryRef, zSubmissionRef } from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/session_graph_provenance.yaml", import.meta.url), "utf8"));
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
