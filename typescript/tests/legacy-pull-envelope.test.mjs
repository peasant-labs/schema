import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import * as schema from "../dist/index.js";
import { checkMin, validateCorpus } from "../dist/testcase.js";

// The same legacy person corpus the Go Validate reads, against the generated
// 0.2.0 pull envelope validators.
const fixture = parse(await readFile(new URL("../../testdata/pull/legacy_envelope.yaml", import.meta.url), "utf8"));

test("generated legacy pull validators keep the anonymous sentinel", async (t) => {
  assert.equal(validateCorpus(fixture.persons), undefined);
  assert.equal(checkMin(fixture.persons, 9), undefined);
  assert.deepEqual(fixture.persons.cases.map((row) => row.name).sort(), [...fixture.required_names.persons].sort());
  for (const row of fixture.persons.cases) {
    await t.test(row.name, () => {
      const { holder, id, username, project_name: projectName } = row.input;
      const body = holder === "transcript"
        ? { transcriptId: "123e4567-e89b-12d3-a456-426614174003", ownerUserId: id, ownerUsername: username, visibility: "group", publishedAt: 1, updatedAt: 1, annotationCount: 0, ...(projectName ? { projectName } : {}) }
        : { authorUserId: id, authorUsername: username };
      const validator = holder === "transcript" ? schema.zLegacyPullTranscriptInfo : schema.zLegacyPullAnnotation;
      const parsed = validator.safeParse(body);
      assert.equal(parsed.success, row.expected.valid, parsed.success ? "accepted" : parsed.error.message);
      if (holder === "transcript") {
        assert.equal(schema.zLegacyPullListResponse.safeParse({ transcripts: [body], page: 1, limit: 1, total: 1 }).success, row.expected.valid);
      }
    });
  }
});
