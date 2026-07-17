import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { checkMin, loadCorpus } from "../dist/testcase.js";

// project-hash-locations.ts's Same<> assertions typecheck fine even if one is
// silently deleted (removing a passing compile-time check is not a type
// error), so nothing in `pnpm typecheck` alone would catch a dropped wire
// location. This test is the coupling: it asserts every named location in the
// shared fixture still has a matching tagged assertion in the .ts file's
// source, so a location dropped from either side fails here.

const fixtureSource = await readFile(new URL("../../testdata/typescript/project_hash_locations.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    if (typeof value !== "string") throw new TypeError(`${path}: must be a string`);
    return value;
  },
  decodeExpected(value, path) {
    if (typeof value !== "string") throw new TypeError(`${path}: must be a string`);
    return value;
  },
});
const compileTimeSource = await readFile(new URL("./project-hash-locations.ts", import.meta.url), "utf8");

test("ProjectHash wire-location coverage matches the strict shared fixture and its compile-time proof", async (t) => {
  assert.equal(fixture.cases.length, 6);
  assert.equal(checkMin(fixture, 6), undefined);
  assert.equal(new Set(fixture.cases.map((testCase) => testCase.name)).size, fixture.cases.length);
  assert.equal(new Set(fixture.cases.map((testCase) => testCase.input)).size, fixture.cases.length);

  const brandedLocations = fixture.cases.filter((testCase) => testCase.expected === "ProjectHash");
  const negativeControls = fixture.cases.filter((testCase) => testCase.expected === "string");
  assert.equal(brandedLocations.length, 5, "expected exactly 5 wire locations asserted as the ProjectHash brand");
  assert.equal(negativeControls.length, 1, "expected exactly 1 same-spelling negative control so the check cannot pass vacuously");

  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      const tag = new RegExp(`//\\s*projectHashLocation:${escapeRegExp(testCase.input)}\\b`);
      assert.match(
        compileTimeSource,
        tag,
        `${testCase.name}: no "// projectHashLocation:${testCase.input}" compile-time Same<> assertion found in project-hash-locations.ts`,
      );
    });
  }
});

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
