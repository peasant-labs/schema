import assert from "node:assert/strict";
import test from "node:test";

import { Classification } from "../dist/testcase.js";
import {
  loadTimelineFixtureManifest,
  loadTimelineFixtures,
  validateTimelineFixtureIdentity,
} from "../dist/fixtures/timeline.js";

test("timeline fixtures expose the complete Go-validated corpus", () => {
  const fixtures = loadTimelineFixtures();
  assert.equal(fixtures.cases.length, 16);
  assert.equal(fixtures.cases.filter((fixture) => fixture.classification === Classification.MustPass).length, 5);
  assert.equal(fixtures.cases.filter((fixture) => fixture.classification === Classification.MustFail).length, 11);
  assert.equal(fixtures.cases[0]?.name, "many_to_many_bindings");
  assert.deepEqual(fixtures.cases[0]?.input.commits[0]?.sessionIds, ["session-a", "session-b"]);
});

test("timeline manifest rejects count-preserving YAML mutations", () => {
  const fixtures = loadTimelineFixtures();
  const manifest = loadTimelineFixtureManifest();
  assert.equal(validateTimelineFixtureIdentity(fixtures, manifest), undefined);
  for (const mutation of manifest.mutations.cases) {
    const mutated = structuredClone(fixtures);
    const target = mutated.cases.find((candidate) => candidate.name === mutation.input.target);
    assert.ok(target, `missing mutation target ${mutation.input.target}`);
    target.name = mutation.input.replacementName;
    if (mutation.input.kind === "replace") {
      target.classification = mutation.input.replacementClassification;
    }
    assert.equal(validateTimelineFixtureIdentity(mutated, manifest) === undefined, mutation.expected, mutation.name);
  }
});

test("timeline fixture loads return independent graphs", () => {
  const first = loadTimelineFixtures();
  const second = loadTimelineFixtures();
  first.cases[0].input.sessions[0].title = "mutated";
  assert.equal(second.cases[0]?.input.sessions[0]?.title, "first task");
});
