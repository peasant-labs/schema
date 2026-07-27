import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { applyAssociationZodRefinements } from "../scripts/lib/association-zod-refinements.mjs";

const trackedGeneratedSource = await readFile(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url), "utf8");
const rawEvidence = "    evidence: z.array(zAssociationEvidenceObservation),";
const refinedEvidence = "    evidence: z.array(zAssociationEvidenceObservation).min(1),";
const rawSessionID = "    sessionId: zSessionID";
const refinedSessionID = "    sessionId: z.string().min(1)";

test("association Zod refinement postprocessor transforms the complete raw generated state", () => {
  assert.equal(applyAssociationZodRefinements(restoreRawState(trackedGeneratedSource)), trackedGeneratedSource);
});

test("association Zod refinement postprocessor preserves the complete refined generated state", () => {
  assert.equal(applyAssociationZodRefinements(trackedGeneratedSource), trackedGeneratedSource);
});

test("association Zod refinement postprocessor rejects a deliberately partial generated state", () => {
  const partialState = replaceExactly(trackedGeneratedSource, refinedEvidence, rawEvidence);
  assertActionableDriftError(partialState, /zSessionAssociation\.evidence=raw/);
});

test("association Zod refinement postprocessor rejects a duplicated generated signal", () => {
  const duplicateState = replaceExactly(trackedGeneratedSource, refinedEvidence, `${refinedEvidence}\n${refinedEvidence}`);
  assertActionableDriftError(duplicateState, /zSessionAssociation\.evidence=invalid\(raw=0, refined=2\)/);
});

test("association Zod refinement postprocessor rejects a missing generated signal", () => {
  const missingState = replaceExactly(trackedGeneratedSource, refinedEvidence, "");
  assertActionableDriftError(missingState, /zSessionAssociation\.evidence=invalid\(raw=0, refined=0\)/);
});

function restoreRawState(source) {
  let raw = replaceExactly(source, refinedEvidence, rawEvidence);
  raw = replaceExactlyInDeclaration(raw, "zSessionAssociation", refinedSessionID, rawSessionID);
  raw = removeRefinement(raw, "zAssociationEvidenceObservation");
  raw = removeRefinement(raw, "zSessionAssociation");
  raw = removeRefinement(raw, "zGitContext");
  raw = removeRefinement(raw, "zAnnotationEntryTarget");
  raw = removeRefinement(raw, "zAnnotationPushItem");
  raw = removeRefinement(raw, "zAnnotationSummary");
  return raw;
}

function removeRefinement(source, schemaName) {
  const marker = `export const ${schemaName} =`;
  const start = source.indexOf(marker);
  assert.notEqual(start, -1, `${schemaName}: generated source is missing the object declaration`);
  assert.equal(source.indexOf(marker, start + marker.length), -1, `${schemaName}: generated source must contain one object declaration`);
  const typeMarker = `\n\nexport type ${schemaName.slice(1)} =`;
  const end = source.indexOf(typeMarker, start);
  assert.notEqual(end, -1, `${schemaName}: generated source is missing the following type declaration`);
  const declaration = source.slice(start, end);
  const refinementStart = declaration.indexOf(".superRefine(");
  assert.notEqual(refinementStart, -1, `${schemaName}: generated source is missing the expected refinement`);
  assert.equal(declaration.indexOf(".superRefine(", refinementStart + 1), -1, `${schemaName}: generated source must contain one refinement`);
  const rawDeclaration = `${declaration.slice(0, refinementStart)};`;
  return `${source.slice(0, start)}${rawDeclaration}${source.slice(end)}`;
}

function assertActionableDriftError(source, expectedSignal) {
  assert.throws(
    () => applyAssociationZodRefinements(source),
    (error) => {
      assert.ok(error instanceof Error);
      assert.match(error.message, /TypeScript root Zod generation\/refinement operation in typescript\/scripts\/lib\/association-zod-refinements\.mjs/);
      assert.match(error.message, expectedSignal);
      assert.match(error.message, /pinned Hey API output or generated zod\.gen\.ts file drifted/);
      assert.match(error.message, /callers have no trustworthy root Zod output/);
      assert.match(error.message, /inspect the generated source, update this postprocessor for the pinned Hey API shape, then regenerate/);
      return true;
    },
  );
}

function replaceExactly(source, before, after) {
  assert.equal(source.split(before).length - 1, 1, `mutation target ${JSON.stringify(before)} must occur exactly once`);
  return source.replace(before, after);
}

function replaceExactlyInDeclaration(source, schemaName, before, after) {
  const marker = `export const ${schemaName} =`;
  const start = source.indexOf(marker);
  assert.notEqual(start, -1, `${schemaName}: generated source is missing the declaration`);
  const typeMarker = `\n\nexport type ${schemaName.slice(1)} =`;
  const end = source.indexOf(typeMarker, start);
  assert.notEqual(end, -1, `${schemaName}: generated source is missing the following type declaration`);
  const declaration = source.slice(start, end);
  assert.equal(declaration.split(before).length - 1, 1, `${schemaName}: mutation target ${JSON.stringify(before)} must occur exactly once`);
  return `${source.slice(0, start)}${declaration.replace(before, after)}${source.slice(end)}`;
}
