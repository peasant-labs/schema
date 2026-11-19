import assert from "node:assert/strict";
import test from "node:test";

import { applyClosedSetZodTypes } from "../scripts/lib/closed-set-zod-types.mjs";

const catalog = { enums: [{ name: "Example" }] };

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

test("closed-set schemas used through a re-declared wrapper are accepted", () => {
  for (const field of ["zExample", "zExample.optional()", "zExample.nullable()", "zExample.nullish()"]) {
    const rendered = applyClosedSetZodTypes(generated(field), catalog);
    assert.match(rendered, /^export const zExample = closedSet<ClosedSets\.Example>\(\)\(z\.enum\(\['first', 'second'\]\)\);$/m);
    assert.match(rendered, new RegExp(`example: ${field.replace(/[.()]/g, "\\$&")} `));
  }
});

// Any other method resolves Zod's `this` against the unwrapped schema and would
// drop the named set from the field, so generation must fail rather than emit it.
test("a closed-set schema used through an undeclared wrapper fails generation", () => {
  assert.throws(
    () => applyClosedSetZodTypes(generated("zExample.default('first')"), catalog),
    /found "zExample\.default\(".*re-declares only \.optional\(\), \.nullable\(\), \.nullish\(\).*lose the named Example set/,
  );
});
