import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
  loadMedianFixtures,
  loadQualityFixtures,
  qualitySessions,
  qualitySessionsForSet,
  sessionByName,
} from "../dist/fixtures/quality.js";
import { checkMin, loadCorpus, validateCorpus } from "../dist/testcase.js";

const mutationMatrixSource = await readFile(new URL("./fixtures/quality-mutation-cases.yaml", import.meta.url), "utf8");

const mutationMatrix = loadCorpus(mutationMatrixSource, {
  decodeInput(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value)) throw new Error(`${path}: must be a mapping`);
    const keys = Object.keys(value);
    if (keys.some((key) => key !== "kind" && key !== "replacement")) throw new Error(`${path}: unknown input key`);
    if (typeof value.kind !== "string") throw new Error(`${path}.kind: must be a string`);
    if (typeof value.replacement !== "string" && typeof value.replacement !== "number") throw new Error(`${path}.replacement: must be a string or number`);
    return { kind: value.kind, replacement: value.replacement };
  },
  decodeExpected(value, path) {
    if (typeof value !== "string" && typeof value !== "number") throw new Error(`${path}: must be a string or number`);
    return value;
  },
});

test("quality fixtures expose the complete Go-validated catalog", () => {
  const fixtures = loadQualityFixtures();
  assert.equal(qualitySessions(fixtures).length, 5);
  assert.equal(sessionByName(fixtures, "resolved_typical").id, "sess-000");
  assert.deepEqual(qualitySessionsForSet(fixtures, "project_mix").map((row) => row.id), ["sess-000", "sess-002"]);
  assert.equal(fixtures.variations.tokenRatios.length, 5);
  assert.equal(fixtures.variations.metrics.specQualityScore.length, 7);
});

test("the personal-medians corpus is exported with every exclusion case", () => {
  const fixtures = loadQualityFixtures();
  const cases = loadMedianFixtures();
  assert.deepEqual(cases.map((medianCase) => medianCase.name).sort(), [...fixtures.requiredMedianNames].sort());
  const empty = cases.find((medianCase) => medianCase.name === "empty_set_returns_no_medians");
  assert.deepEqual(empty.sessions, []);
  assert.deepEqual(empty.expected, {});
  const typical = cases.find((medianCase) => medianCase.name === "typical");
  assert.deepEqual(typical.expected, { retryShare: 0.05, specQualityScore: 80, withinSessionReverts: 1 });
  const nonPositive = cases.find((medianCase) => medianCase.name === "non_positive_spec_score_excluded");
  assert.deepEqual(nonPositive.expected, { retryShare: 0, specQualityScore: 60, withinSessionReverts: 0 });
});

test("quality accessors do not expose mutable canonical rows", async (t) => {
  assert.equal(mutationMatrix.cases.length, 4);
  assert.equal(checkMin(mutationMatrix, 4), undefined);
  assert.equal(validateCorpus(mutationMatrix), undefined);

  for (const testCase of mutationMatrix.cases) {
    await t.test(testCase.name, () => {
      const fixtures = loadQualityFixtures();
      let actual;
      switch (testCase.input.kind) {
        case "session-scalar":
          fixtures.sessions[0].project = testCase.input.replacement;
          actual = loadQualityFixtures().sessions[0].project;
          break;
        case "nested-set-case":
          fixtures.sets[0].cases[0] = testCase.input.replacement;
          actual = loadQualityFixtures().sets[0].cases[0];
          break;
        case "nested-metric-value":
          fixtures.variations.metrics.specQualityScore[0].value = testCase.input.replacement;
          actual = loadQualityFixtures().variations.metrics.specQualityScore[0].value;
          break;
        case "accessor-result": {
          const accessed = sessionByName(fixtures, "resolved_typical");
          accessed.project = testCase.input.replacement;
          actual = sessionByName(fixtures, "resolved_typical").project;
          break;
        }
        default:
          throw new Error(`quality mutation fixture selected unknown kind ${JSON.stringify(testCase.input.kind)}`);
      }
      assert.equal(actual, testCase.expected);
    });
  }
});
