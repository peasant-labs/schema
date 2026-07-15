import assert from "node:assert/strict";
import test from "node:test";

import {
  allQualityFixtures,
  qualityFixture,
  qualityFixtureSet,
  qualitySessions,
  qualityVariations,
} from "../dist/fixtures/quality.js";

test("quality fixtures expose the complete Go-validated catalog", () => {
  assert.equal(qualitySessions.length, 5);
  assert.equal(qualityFixture("resolved_typical").id, "sess-000");
  assert.deepEqual(qualityFixtureSet("project_mix").map((row) => row.id), ["sess-000", "sess-002"]);
  assert.equal(qualityVariations.tokenRatios.length, 5);
  assert.equal(qualityVariations.metrics.specQualityScore.length, 7);
});

test("quality accessors do not expose mutable canonical rows", () => {
  const first = qualityFixture("resolved_typical");
  first.project = "changed by consumer";
  assert.equal(qualityFixture("resolved_typical").project, "fortuna");

  const all = allQualityFixtures();
  all[0].title = "changed by consumer";
  assert.equal(qualityFixture("resolved_typical").title, "Fix authentication middleware");
});
