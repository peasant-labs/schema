import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import {
  zAnalyticsSessionStats,
  zContributorBreakdown,
  zLongestStreak,
  zMedianP90,
  zOutcomeDistribution,
  zPersonalMedians,
  zProjectAnalytics,
  zQualityPage,
  zReturningContributorRate,
  zSessionToCommitRate,
  zWeekAvgDuration,
  zWeekContributors,
  zWeekCount,
  zWeekNewContributors,
} from "../dist/index.js";

const manifest = parse(await readFile(new URL("../../testdata/project_analytics.yaml", import.meta.url), "utf8"));

const schemas = {
  ProjectAnalytics: zProjectAnalytics,
  WeekCount: zWeekCount,
  WeekContributors: zWeekContributors,
  ReturningContributorRate: zReturningContributorRate,
  LongestStreak: zLongestStreak,
  WeekNewContributors: zWeekNewContributors,
  SessionToCommitRate: zSessionToCommitRate,
  WeekAvgDuration: zWeekAvgDuration,
  OutcomeDistribution: zOutcomeDistribution,
  MedianP90: zMedianP90,
  AnalyticsSessionStats: zAnalyticsSessionStats,
  ContributorBreakdown: zContributorBreakdown,
};

test("the generated analytics types carry exactly the manifest field set", () => {
  assert.equal(manifest.types.length, Object.keys(schemas).length);
  for (const declared of manifest.types) {
    const schema = schemas[declared.name];
    assert.notEqual(schema, undefined, `manifest names unknown analytics type ${declared.name}`);
    assert.deepEqual(Object.keys(schema.shape).sort(), [...declared.fields].sort(), declared.name);
  }
});

test("a personal median is absent rather than fabricated when its sample is empty", () => {
  assert.deepEqual(Object.keys(zPersonalMedians.shape).sort(), ["retryShare", "specQualityScore", "withinSessionReverts"]);
  assert.deepEqual(zPersonalMedians.parse({}), {});
  for (const [name, schema] of Object.entries(zPersonalMedians.shape)) {
    assert.equal(schema.safeParse(undefined).success, true, `${name} must stay optional`);
    assert.equal(schema.safeParse(null).success, true, `${name} may be an explicit null`);
  }
});

test("the quality page carries one offset page and the full analytics summary", () => {
  assert.deepEqual(Object.keys(zQualityPage.shape).sort(), ["items", "nextItems", "page", "perPage", "revision", "summary", "total"]);
  assert.deepEqual(Object.keys(zQualityPage.shape.summary.shape).sort(), Object.keys(zProjectAnalytics.shape).sort());
});
