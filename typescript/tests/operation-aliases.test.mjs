import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { canonicalOperationAliases } from "../scripts/lib/operation-aliases.mjs";
import { loadCorpus } from "../dist/testcase.js";

// canonicalOperationAliases is the generation-time guard that stops a
// Local/Village API component from silently masquerading as a canonical root
// Types component when their schemas have drifted (see
// typescript/scripts/lib/operation-aliases.mjs). It happens to never throw on
// today's already-consistent specs, which proves nothing about whether it
// would fire on a real mismatch; this suite is the "gates have their own
// tests" synthetic-break proof TESTING.md's philosophy requires (mirroring
// internal/contractgates/synthetic_break_test.go's TestOasdiffSyntheticBreak
// pattern), plus the fixture-backed collision-safety corpus previously proved
// against the deleted bespoke generator.

const fixtureSource = await readFile(new URL("../../testdata/typescript/collision_cases.yaml", import.meta.url), "utf8");
const fixture = loadCorpus(fixtureSource, {
  decodeInput(value, path) {
    if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
    const allowedKeys = ["name", "first_schema", "second_schema"];
    for (const key of Object.keys(value)) {
      if (!allowedKeys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
    }
    if (typeof value.name !== "string") throw new TypeError(`${path}.name: must be a string`);
    if (typeof value.first_schema !== "string") throw new TypeError(`${path}.first_schema: must be a string`);
    if (typeof value.second_schema !== "string") throw new TypeError(`${path}.second_schema: must be a string`);
    return { name: value.name, firstSchema: JSON.parse(value.first_schema), secondSchema: JSON.parse(value.second_schema) };
  },
  decodeExpected(value, path) {
    if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
    return value;
  },
});

test("canonicalOperationAliases collision safety matches the strict shared fixture", async (t) => {
  assert.equal(fixture.cases.length, 4);
  assert.equal(new Set(fixture.cases.map((testCase) => testCase.name)).size, fixture.cases.length);

  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      const rootSpec = { components: { schemas: { [testCase.input.name]: testCase.input.firstSchema } } };
      const apiSpec = { components: { schemas: { [testCase.input.name]: testCase.input.secondSchema } } };
      if (testCase.expected) {
        const aliases = canonicalOperationAliases("test", apiSpec, rootSpec);
        assert.equal(aliases.get(testCase.input.name), testCase.input.name);
        return;
      }
      assert.throws(
        () => canonicalOperationAliases("test", apiSpec, rootSpec),
        /normalizing to canonical root type .* with an unequal schema; aliasing it would misrepresent the HTTP contract/,
      );
    });
  }
});

test("canonicalOperationAliases synthetic-break proof: a deliberately mismatched schema fires the throw", () => {
  const rootSpec = { components: { schemas: { Widget: { type: "object", properties: { count: { type: "integer" } }, required: ["count"] } } } };
  const mismatchedApiSpec = { components: { schemas: { Widget: { type: "object", properties: { count: { type: "string" } }, required: ["count"] } } } };
  assert.throws(
    () => canonicalOperationAliases("synthetic", mismatchedApiSpec, rootSpec),
    (error) => {
      assert.match(error.message, /TypeScript synthetic operation generation found API component Widget normalizing to canonical root type Widget with an unequal schema/);
      assert.match(error.message, /rename the operation-specific component or make it exactly equal to the canonical Go\/OpenAPI definition/);
      return true;
    },
  );

  const matchedApiSpec = { components: { schemas: { Widget: { type: "object", properties: { count: { type: "integer" } }, required: ["count"] } } } };
  const aliases = canonicalOperationAliases("synthetic", matchedApiSpec, rootSpec);
  assert.equal(aliases.get("Widget"), "Widget");
});

test("canonicalOperationAliases rejects a document missing components.schemas", () => {
  assert.throws(
    () => canonicalOperationAliases("synthetic", {}, { components: { schemas: {} } }),
    /could not compare components\.schemas with the canonical Types catalog/,
  );
  assert.throws(
    () => canonicalOperationAliases("synthetic", { components: { schemas: {} } }, {}),
    /could not compare components\.schemas with the canonical Types catalog/,
  );
});
