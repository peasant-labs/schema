import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import { assertPackageRepositoryMetadata } from "../scripts/lib/package-repository-metadata.mjs";
import { loadCorpus } from "../dist/testcase.js";

const packageFixture = parse(await readFile(new URL("./fixtures/package-files.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(packageFixture), ["files", "repository"]);

const mutationFixture = loadCorpus(await readFile(new URL("./fixtures/package-repository-mutations.yaml", import.meta.url), "utf8"), {
  decodeInput(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
    for (const key of Object.keys(value)) {
      if (key !== "repository") throw new TypeError(`${path}.${key}: unknown key`);
    }
    if (value.repository === undefined) return {};
    if (typeof value.repository !== "object" || value.repository === null || Array.isArray(value.repository)) {
      throw new TypeError(`${path}.repository: must be a mapping when present`);
    }
    if (typeof value.repository.type !== "string" || typeof value.repository.url !== "string") {
      throw new TypeError(`${path}.repository: must declare string type and url values`);
    }
    return { repository: { type: value.repository.type, url: value.repository.url } };
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

test("package repository audit rejects fixture-backed metadata mutations", async (t) => {
  assert.equal(mutationFixture.cases.length, 3);
  assert.equal(new Set(mutationFixture.cases.map((testCase) => testCase.name)).size, mutationFixture.cases.length);

  const accepted = mutationFixture.cases.filter((testCase) => testCase.expected);
  const rejected = mutationFixture.cases.filter((testCase) => !testCase.expected);
  assert.equal(accepted.length, 1, "expected exactly one canonical repository metadata case");
  assert.equal(rejected.length, 2, "expected missing and wrong repository metadata mutation cases");
  assert.ok(rejected.some((testCase) => !Object.hasOwn(testCase.input, "repository")), "missing repository metadata mutation is required");
  assert.ok(rejected.some((testCase) => testCase.input.repository?.url !== packageFixture.repository.url), "wrong repository metadata mutation is required");

  for (const testCase of mutationFixture.cases) {
    await t.test(testCase.name, () => {
      if (testCase.expected) {
        assert.doesNotThrow(() => assertPackageRepositoryMetadata(testCase.input, packageFixture.repository));
        return;
      }
      assert.throws(
        () => assertPackageRepositoryMetadata(testCase.input, packageFixture.repository),
        /package\/package\.json must declare the canonical peasant-labs\/schema git repository metadata for npm provenance/,
      );
    });
  }
});
