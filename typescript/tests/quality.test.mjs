import assert from "node:assert/strict";
import test from "node:test";

import {
  loadQualityFixtures,
  qualitySessions,
  qualitySessionsForSet,
  sessionByName,
} from "../dist/fixtures/quality.js";

test("quality fixtures expose the complete Go-validated catalog", () => {
  const fixtures = loadQualityFixtures();
  assert.equal(qualitySessions(fixtures).length, 5);
  assert.equal(sessionByName(fixtures, "resolved_typical").id, "sess-000");
  assert.deepEqual(qualitySessionsForSet(fixtures, "project_mix").map((row) => row.id), ["sess-000", "sess-002"]);
  assert.equal(fixtures.variations.tokenRatios.length, 5);
  assert.equal(fixtures.variations.metrics.specQualityScore.length, 7);
});

test("quality accessors do not expose mutable canonical rows", () => {
  const firstLoad = loadQualityFixtures();
  const first = sessionByName(firstLoad, "resolved_typical");
  first.project = "changed by consumer";
  firstLoad.variations.metrics.specQualityScore[0].name = "changed by consumer";

  const secondLoad = loadQualityFixtures();
  assert.equal(sessionByName(secondLoad, "resolved_typical").project, "fortuna");
  assert.notEqual(secondLoad.variations.metrics.specQualityScore[0].name, "changed by consumer");
});
