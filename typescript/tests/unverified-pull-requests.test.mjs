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

const rowFixture = parse(await readFile(new URL("../../testdata/unverified_pull_request_rows.yaml", import.meta.url), "utf8"));
test("public validators enforce canonical identity and source bounds", async (t) => {
  assert.equal(validateCorpus(rowFixture), undefined);
  for (const name of rowFixture.required_names) assert.ok(rowFixture.cases.some((c) => c.name === name), name);
  for (const row of rowFixture.cases) await t.test(row.name, () => {
    const value = JSON.parse(row.input.json);
    assert.equal(schema.zUnverifiedPullRequest.safeParse(value).success, row.expected.valid, "row");
    assert.equal(schema.zSessionPullRequestsResponse.safeParse({pullRequests: [], unverified_pull_requests: [value]}).success, row.expected.valid, "local");
    assert.equal(schema.zVillageTranscriptPullRequestsResponse.safeParse({pull_requests: [], unverified_pull_requests: [value]}).success, row.expected.valid, "village");
  });
});

const casingFixture = parse(await readFile(new URL("../../testdata/unverified_pull_request_casing.yaml", import.meta.url), "utf8"));
test("canonical property casing remains forward compatible", async (t) => {
  assert.equal(validateCorpus(casingFixture), undefined);
  for (const name of casingFixture.required_names) assert.ok(casingFixture.cases.some((c) => c.name === name), name);
  for (const row of casingFixture.cases) await t.test(row.name, () => {
    const value = JSON.parse(row.input.json);
    for (const result of [schema.zUnverifiedPullRequest.safeParse(value), schema.zSessionPullRequestsResponse.safeParse({pullRequests: [], unverified_pull_requests: [value]}), schema.zVillageTranscriptPullRequestsResponse.safeParse({pull_requests: [], unverified_pull_requests: [value]})]) assert.equal(result.success,row.expected.valid);
    const parsed = schema.zUnverifiedPullRequest.safeParse(value); if (parsed.success) assert.equal(parsed.data.owner,row.expected.owner);
  });
});
