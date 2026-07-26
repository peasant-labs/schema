import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { Classification } from "../dist/testcase.js";
import { loadTimelineFixtures } from "../dist/fixtures/timeline.js";

test("timeline fixtures expose the complete Go-validated corpus", () => {
  const fixtures = loadTimelineFixtures();
  assert.equal(fixtures.cases.length, 24);
  assert.equal(fixtures.cases.filter((fixture) => fixture.classification === Classification.MustPass).length, 7);
  assert.equal(fixtures.cases.filter((fixture) => fixture.classification === Classification.MustFail).length, 17);
  assert.equal(new Set(fixtures.cases.map((fixture) => fixture.family)).size, 24);
  assert.equal(fixtures.cases[0]?.name, "many_to_many_bindings");
  assert.deepEqual(fixtures.cases[0]?.input.commits[0]?.sessionIds, ["session-a", "session-b"]);
  assert.equal(fixtures.successorAssociationMirrorCases.length, 5);
  assert.deepEqual(
    fixtures.successorAssociationMirrorCases.map((fixture) => fixture.name),
    [
      "rewrite_successor_association_id_drift_is_rejected",
      "rewrite_successor_association_session_drift_is_rejected",
      "rewrite_successor_association_conclusion_drift_is_rejected",
      "rewrite_successor_association_confidence_drift_is_rejected",
      "rewrite_successor_association_evidence_drift_is_rejected",
    ],
  );
  for (const fixture of fixtures.successorAssociationMirrorCases) {
    assert.equal(fixture.classification, Classification.MustFail);
    assert.equal(fixture.expected.repair?.kind, "replace_successor_association");
    assert.ok(fixture.expected.repair?.ghostHash);
    assert.ok(fixture.expected.repair?.successorHash);
    assert.ok(fixture.expected.repair?.associationId);
    assert.equal(fixture.expected.repair?.postMutationValid, true);
  }
});

test("timeline fixture loads return independent graphs", () => {
  const first = loadTimelineFixtures();
  const second = loadTimelineFixtures();
  first.cases[0].input.sessions[0].title = "mutated";
  first.successorAssociationMirrorCases[0].input.commits[0].associations[0].conclusion = "mutated";
  assert.equal(second.cases[0]?.input.sessions[0]?.title, "first task");
  assert.equal(second.successorAssociationMirrorCases[0]?.input.commits[0]?.associations[0]?.conclusion, "confirmed");
});

test("timeline package declarations omit schema-repo identity scaffolding", async () => {
  const declarations = await readFile(new URL("../dist/fixtures/timeline.d.ts", import.meta.url), "utf8");
  assert.doesNotMatch(declarations, /TimelineFixtureManifest|TimelineManifestMutation|validateTimelineFixtureIdentity/);
});
