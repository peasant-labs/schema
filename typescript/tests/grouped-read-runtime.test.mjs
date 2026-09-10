import assert from "node:assert/strict";
import test from "node:test";
import * as schema from "../dist/index.js";
import { sessionGraphFixtures } from "../dist/fixtures/session-graph.js";

const fixtures = sessionGraphFixtures().grouped_read;

test("grouped read corpus has exact named coverage", () => {
  assert.ok(fixtures.cases.cases.length > 0);
  assert.deepEqual(new Set(fixtures.cases.cases.map(row => row.name)), new Set(fixtures.requiredNames));
  assert.equal(new Set(fixtures.requiredNames).size, fixtures.requiredNames.length);
});

for (const row of fixtures.cases.cases) {
  test(`public grouped read boundary: ${row.name}`, () => {
    const validator = schema[`z${row.input.schema}`];
    assert.equal(typeof validator?.safeParse, "function", `missing public validator for ${row.input.schema}`);
    const original = structuredClone(row.input.payload);
    const result = validator.safeParse(row.input.payload);
    assert.equal(result.success, row.expected.valid, result.error?.message);
    assert.deepEqual(row.input.payload, original, "validation mutated caller data");
    if (result.success) {
      // Existing int64 API fields are represented as bigint by the package;
      // these fixtures use exactly representable wire values for round trips.
      const wire = JSON.parse(JSON.stringify(result.data, (_, value) => typeof value === "bigint" ? Number(value) : value));
      assert.deepEqual(wire, original, "read projection stripped or fabricated a field");
    }
  });
}
