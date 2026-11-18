import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import { applyClosedSetZodTypes } from "../scripts/lib/closed-set-zod-types.mjs";

const catalog = { enums: [{ name: "Example" }] };
const fixture = parse(await readFile(new URL("./fixtures/closed-set-zod-types.yaml", import.meta.url), "utf8"));
const requiredNames = [
  "a-bare-reference-is-accepted",
  "an-optional-call-is-accepted",
  "a-nullable-call-is-accepted",
  "a-nullish-call-is-accepted",
  "an-optional-chained-wrapper-call-is-accepted",
  "a-wrapper-call-split-across-lines-is-accepted",
  "an-undeclared-method-call-fails",
  "an-optional-chained-undeclared-call-fails",
  "spaces-around-the-dot-do-not-hide-an-undeclared-call",
  "a-newline-before-the-dot-does-not-hide-an-undeclared-call",
  "an-undeclared-method-passed-by-reference-fails",
  "a-bracket-access-fails-as-unverifiable",
  "an-optional-chained-bracket-access-fails-as-unverifiable",
];
assert.deepEqual(new Set(fixture.cases.map((row) => row.name)), new Set(requiredNames));
assert.equal(fixture.cases.length, new Set(requiredNames).size, "fixture names must be unique");

function generated(field) {
  return [
    "import * as z from 'zod';",
    "",
    "export const zExample = z.enum(['first', 'second']);",
    "",
    "export type Example = z.infer<typeof zExample>;",
    "",
    `export const zHolder = z.object({ example: ${field} });`,
    "",
  ].join("\n");
}

// A re-declared wrapper keeps the named set. Any other member resolves Zod's
// `this` against the unwrapped schema and would drop the named set from the
// field, and a computed member cannot be checked at all, so generation must fail
// rather than emit either.
for (const row of fixture.cases) {
  test(row.name, () => {
    switch (row.outcome) {
      case "accepted": {
        const rendered = applyClosedSetZodTypes(generated(row.field), catalog);
        assert.match(rendered, /^export const zExample = closedSet<ClosedSets\.Example>\(\)\(z\.enum\(\['first', 'second'\]\)\);$/m);
        assert.ok(rendered.includes(`example: ${row.field} `), `the field expression must render unchanged: ${row.field}`);
        break;
      }
      case "undeclared-member":
      case "computed-member":
        assert.throws(() => applyClosedSetZodTypes(generated(row.field), catalog), new RegExp(row.error, "s"));
        break;
      default:
        assert.fail(`unknown outcome ${row.outcome} in fixtures/closed-set-zod-types.yaml case ${row.name}; use accepted, undeclared-member or computed-member`);
    }
  });
}
