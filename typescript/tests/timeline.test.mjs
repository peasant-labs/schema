import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { Classification } from "../dist/testcase.js";
import { loadTimelineFixtures } from "../dist/fixtures/timeline.js";

test("timeline fixtures expose the complete Go-validated corpus", () => {
  const fixtures = loadTimelineFixtures();
  assert.equal(fixtures.cases.length, 20);
  assert.equal(fixtures.cases.filter((fixture) => fixture.classification === Classification.MustPass).length, 5);
  assert.equal(fixtures.cases.filter((fixture) => fixture.classification === Classification.MustFail).length, 15);
  assert.equal(new Set(fixtures.cases.map((fixture) => fixture.family)).size, 20);
  assert.equal(fixtures.cases[0]?.name, "many_to_many_bindings");
  assert.deepEqual(fixtures.cases[0]?.input.commits[0]?.sessionIds, ["session-a", "session-b"]);
});

test("timeline fixture loads return independent graphs", () => {
  const first = loadTimelineFixtures();
  const second = loadTimelineFixtures();
  first.cases[0].input.sessions[0].title = "mutated";
  assert.equal(second.cases[0]?.input.sessions[0]?.title, "first task");
});

test("timeline package declarations omit schema-repo identity scaffolding", async () => {
  const declarations = await readFile(new URL("../dist/fixtures/timeline.d.ts", import.meta.url), "utf8");
  assert.doesNotMatch(declarations, /TimelineFixtureManifest|TimelineManifestMutation|validateTimelineFixtureIdentity/);
});
