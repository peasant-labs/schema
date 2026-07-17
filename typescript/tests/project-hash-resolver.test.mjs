import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { PROJECT_HASH_PATTERN, isCanonicalProjectHashSchema, shouldBrandProjectHash } from "../scripts/lib/project-hash-resolver.mjs";
import { loadCorpus } from "../dist/testcase.js";

// shouldBrandProjectHash (typescript/scripts/lib/project-hash-resolver.mjs) is
// the generation-time decision behind the Hey API Zod plugin's ProjectHash
// resolver: it stops a future unrelated 64-lowercase-hex field (for example a
// SHA-256 content hash) from silently receiving the ProjectHash brand merely
// for sharing its shape. Today's Types catalog has exactly one schema with
// this pattern, so the resolver never observably rejects anything in real
// generation output -- this suite is the committed hostile/negative-control
// proof that it actually would, driving the same pure function the generator
// calls (typescript/openapi-ts.config.mjs's $resolvers.string), not a
// reimplementation of it.

const fixtureSource = await readFile(new URL("../../testdata/typescript/project_hash_resolver_cases.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
    const allowedKeys = ["path", "pattern"];
    for (const key of Object.keys(value)) {
      if (!allowedKeys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
    }
    if (value.path !== null && !(Array.isArray(value.path) && value.path.every((segment) => typeof segment === "string"))) {
      throw new TypeError(`${path}.path: must be null or a sequence of strings`);
    }
    if (typeof value.pattern !== "string") throw new TypeError(`${path}.pattern: must be a string`);
    return { path: value.path, pattern: value.pattern };
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

// contextPath renders a fixture's plain segment array (or null) into the same
// shape the real Hey API SchemaVisitorContext.path Ref carries: a bracket-only
// "~ref" property (see @hey-api/codegen-core's Ref<T> = { '~ref': T }).
function contextPath(segments) {
  return segments === null ? undefined : { "~ref": segments };
}

test("ProjectHash resolver branding matches the strict shared fixture", async (t) => {
  assert.equal(fixture.cases.length, 6);
  assert.equal(new Set(fixture.cases.map((testCase) => testCase.name)).size, fixture.cases.length);

  const brandedCases = fixture.cases.filter((testCase) => testCase.expected);
  const rejectedCases = fixture.cases.filter((testCase) => !testCase.expected);
  assert.equal(brandedCases.length, 1, "expected exactly 1 positive (branded) case");
  assert.equal(rejectedCases.length, 5, "expected exactly 5 negative-control cases proving the guard is not vacuous");

  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      const schema = { pattern: testCase.input.pattern };
      const path = contextPath(testCase.input.path);
      assert.equal(shouldBrandProjectHash(schema, path), testCase.expected);
    });
  }
});

test("shouldBrandProjectHash rejects a matching path when the pattern does not match", () => {
  assert.equal(shouldBrandProjectHash({ pattern: "^[0-9a-f]{64}$" }, contextPath(["components", "schemas", "ProjectHash"])), true);
  assert.equal(shouldBrandProjectHash({ pattern: "different" }, contextPath(["components", "schemas", "ProjectHash"])), false);
  assert.equal(shouldBrandProjectHash({}, contextPath(["components", "schemas", "ProjectHash"])), false);
});

test("isCanonicalProjectHashSchema is the exact anchor shouldBrandProjectHash delegates to", () => {
  const canonical = contextPath(["components", "schemas", "ProjectHash"]);
  assert.equal(isCanonicalProjectHashSchema(canonical), true);
  assert.equal(isCanonicalProjectHashSchema(contextPath(["components", "schemas", "ContentHash"])), false);
  assert.equal(isCanonicalProjectHashSchema(undefined), false);
});

test("PROJECT_HASH_PATTERN is the canonical Go ProjectHash newtype's exact regex source", () => {
  assert.equal(PROJECT_HASH_PATTERN, "^[0-9a-f]{64}$");
});
