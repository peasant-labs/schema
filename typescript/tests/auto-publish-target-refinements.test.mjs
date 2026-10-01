import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import { normalizeAutoPublishTargetZod, applyAutoPublishTargetZodRefinements } from "../scripts/lib/auto-publish-target-zod-refinements.mjs";

const source = await readFile(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url), "utf8");
const fixture = parse(await readFile(new URL("./fixtures/auto-publish-target-refinements.yaml", import.meta.url), "utf8"));
const requiredNames = [
  "normalization-preserves-the-real-refined-artifact",
  "target-refinement-preserves-the-real-refined-artifact",
  "a-missing-target-declaration-aborts-normalization",
  "a-duplicate-target-declaration-aborts-normalization",
  "an-unrecognized-intersection-aborts-normalization",
  "an-unclosed-target-object-aborts-refinement",
];
assert.deepEqual(new Set(fixture.cases.map(row => row.name)), new Set(requiredNames));
assert.equal(fixture.cases.length, new Set(requiredNames).size, "fixture names must be unique");
const declaration = "export const zAutoPublishRuleRequest = z.object({";
const start = source.indexOf(declaration);
const end = source.indexOf("\nexport type AutoPublishRuleRequest", start);
assert.ok(start >= 0 && end > start, "tests must mutate the actual generated target validator");
const block = source.slice(start, end);
for (const row of fixture.cases) {
  test(row.name, () => {
    let input = source;
    switch (row.mutation) {
      case "none": break;
      case "missing": input = source.replace(declaration, "export const zOtherRuleRequest = z.object({"); break;
      case "duplicate": input += block; break;
      case "intersection": input = source.replace(declaration, "export const zAutoPublishRuleRequest = z.intersection(z.string(), z.object({"); break;
      case "unclosed": input = source.slice(0, start) + block.slice(0, block.indexOf("\n}).strict()")) + "\n});\n" + source.slice(end); break;
      default: assert.fail(`unknown fixture mutation: ${row.mutation}`);
    }
    const apply = row.operation === "normalize" ? normalizeAutoPublishTargetZod : applyAutoPublishTargetZodRefinements;
    assert.ok(["normalize", "refine"].includes(row.operation), "fixture names a supported pass");
    if (row.error) assert.throws(() => apply(input), error => error.message.includes(row.error));
    else assert.equal(apply(input), source);
  });
}
