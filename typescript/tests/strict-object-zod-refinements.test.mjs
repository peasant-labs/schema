import assert from "node:assert/strict";
import test from "node:test";

import { readFile } from "node:fs/promises";

import { applyStrictObjectZodRefinements } from "../scripts/lib/strict-object-zod-refinements.mjs";

// The COMMITTED artifact, not a hand-written imitation. A two-declaration mock
// cannot show how the postprocessor behaves among the real file's other
// declarations, which is exactly where its defect lived: a span that ran past
// its own declaration into the next one. The sibling refinement suite this set
// mirrors reads the same file for the same reason.
const trackedGeneratedSource = await readFile(
  new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url),
  "utf8",
);

// This postprocessor closes the generated runtime validator for the owner-update
// body. It runs on generator output rather than on committed source, so the
// idempotence suite cannot reach it: that suite shells out to `pnpm run
// generate`, which recreates the contract from scratch, meaning the sanctioned
// path only ever feeds this function RAW input. A test that exercises only the
// sanctioned path cannot find a defect in the unsanctioned one, so these call the
// function directly.
//
// The state machine the function declares — raw, refined, or a mixed state it
// must refuse — is the thing under test. Its refined arm is the one nothing
// previously covered, and it was broken: locating the declaration's end by
// scanning for the raw terminator alone stopped matching once `.strict()` was
// appended, so the span ran into the FOLLOWING declaration and appended
// `.strict()` to a schema deliberately left open.

// A miniature generated contract: one declaration this postprocessor closes,
// followed by one it must never touch.
const RAW = `export const zTranscriptUpdateRequest = z.object({
    description: z.string().optional(),
    title: z.string().max(500).optional()
});

export const zTrendsPayload = z.object({
    days: z.array(zDayStats).nullable()
});
`;

const REFINED = RAW.replace(
  `    title: z.string().max(500).optional()\n});`,
  `    title: z.string().max(500).optional()\n}).strict();`,
);

function strictCount(source) {
  return (source.match(/\.strict\(\)/g) ?? []).length;
}

test("strict-object Zod refinement postprocessor transforms the complete raw generated state", () => {
  const result = applyStrictObjectZodRefinements(RAW);
  assert.equal(result, REFINED, "a raw contract must gain exactly the declared closure and nothing else");
  assert.equal(strictCount(result), 1, "exactly one declaration may be closed");
});

test("strict-object Zod refinement postprocessor preserves the complete refined generated state", () => {
  // The regression this file exists for. Feeding back an already-refined
  // contract must be a no-op; previously it appended `.strict()` to the NEXT
  // declaration, silently closing a payload the contract leaves open.
  const result = applyStrictObjectZodRefinements(REFINED);
  assert.equal(result, REFINED, "an already-refined contract must be returned unchanged");
  assert.equal(strictCount(result), 1, "re-running must not add a second closure");
  assert.ok(
    !/zTrendsPayload = z\.object\(\{[\s\S]*?\}\)\.strict\(\)/.test(result),
    "re-running must not close zTrendsPayload, which the contract deliberately leaves open",
  );
});

test("strict-object Zod refinement postprocessor is idempotent across repeated application", () => {
  const once = applyStrictObjectZodRefinements(RAW);
  const twice = applyStrictObjectZodRefinements(once);
  assert.equal(twice, once, "applying the refinement twice must equal applying it once");
});

test("strict-object Zod refinement postprocessor rejects a missing generated signal", () => {
  const withoutDeclaration = RAW.replace("export const zTranscriptUpdateRequest = z.object({", "export const zSomethingElse = z.object({");
  assert.throws(
    () => applyStrictObjectZodRefinements(withoutDeclaration),
    /expected exactly one .* declaration .* but found 0/s,
    "a vanished declaration must abort generation rather than silently ship an open validator",
  );
});

test("strict-object Zod refinement postprocessor rejects a duplicated generated signal", () => {
  const duplicated = `${RAW}\n${RAW}`;
  assert.throws(
    () => applyStrictObjectZodRefinements(duplicated),
    /expected exactly one .* declaration .* but found 2/s,
    "an ambiguous duplicate must abort rather than guess which declaration to close",
  );
});

test("strict-object Zod refinement postprocessor rejects an unterminated declaration", () => {
  const truncated = "export const zTranscriptUpdateRequest = z.object({\n    title: z.string()";
  assert.throws(
    () => applyStrictObjectZodRefinements(truncated),
    /no terminating/,
    "an unrecognizable declaration shape must abort rather than produce an arbitrary span",
  );
});

test("strict-object Zod refinement postprocessor is a no-op on the committed artifact", () => {
  // The committed file is already refined, so re-running must return it byte
  // for byte. This is the real-artifact form of the regression above: on the
  // tracked file the broken span ran into a neighbouring declaration.
  assert.equal(
    applyStrictObjectZodRefinements(trackedGeneratedSource),
    trackedGeneratedSource,
    "the committed contract is already refined and must survive another pass unchanged",
  );
});

test("strict-object Zod refinement postprocessor closes exactly one declaration in the committed artifact", () => {
  const declarations = trackedGeneratedSource.match(/^export const z\w+ = z\.object\(\{/gm) ?? [];
  assert.ok(declarations.length > 50, `expected the committed artifact to hold many object declarations, found ${declarations.length}`);
  const closed = trackedGeneratedSource.match(/\}\)\.strict\(\);/g) ?? [];
  assert.equal(closed.length, 1, `exactly one declaration may be closed in the committed artifact, found ${closed.length}`);
});

test("strict-object Zod refinement postprocessor closes only the declarations it names", () => {
  const result = applyStrictObjectZodRefinements(RAW);
  const trends = result.slice(result.indexOf("zTrendsPayload"));
  assert.ok(
    !trends.includes(".strict()"),
    "a schema outside the declared strict set must be left open; closing a payload the producer may extend would break consumers",
  );
});
