import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import * as schema from "../dist/index.js";
import { validateCorpus } from "../dist/testcase.js";

const fixture = parse(await readFile(new URL("../../testdata/unverified_pull_requests.yaml", import.meta.url), "utf8"));
test("unverified pull request read envelopes retain their optional nonnull evidence", async (t) => {
  assert.equal(validateCorpus(fixture), undefined);
  for (const name of fixture.required_names) assert.ok(fixture.cases.some((c) => c.name === name), name);
  for (const row of fixture.cases) await t.test(row.name, () => {
    const input = JSON.parse(row.input.json);
    const validator = row.input.kind === "local" ? schema.zSessionPullRequestsResponse : schema.zVillageTranscriptPullRequestsResponse;
    const result = validator.safeParse(input);
    assert.equal(result.success, row.expected.valid);
    if (result.success) assert.deepEqual(result.data, input);
  });
});
