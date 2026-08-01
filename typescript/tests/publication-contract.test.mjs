import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import { zAuthoritativePublishRequest, zAuthoritativePublishResponse, zCanonicalPublishOperation, zOwnerTranscriptUpdateRequest, zOwnerTranscriptUpdateResponse } from "../dist/index.js";
import { applyPublicationZodRefinements } from "../scripts/lib/publication-zod-refinements.mjs";

const corpus = parse(await readFile(new URL("../../testdata/publication/contract.yaml", import.meta.url), "utf8"));
const generatedZodSource = await readFile(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url), "utf8");

for (const arm of ["metadata", "owner_updates", "owner_update_responses", "operations", "responses", "visibility_intents"]) {
  const rows = corpus[arm];
  for (const row of rows.cases) {
    assert.equal(typeof row.expected.go_valid, "boolean", `${arm}/${row.name} must explicitly declare go_valid`);
    assert.equal(typeof row.expected.zod_valid, "boolean", `${arm}/${row.name} must explicitly declare zod_valid`);
  }
}

test("built publication metadata preserves optional non-null visibility intent and required content hash", () => {
  assert.equal(zAuthoritativePublishRequest.shape.visibilityIntent.isOptional(), true, "visibilityIntent omission is the legacy-compatible form");
  assert.equal(zAuthoritativePublishRequest.shape.visibilityIntent.isNullable(), false, "explicit null remains rejected");
  assert.equal(zAuthoritativePublishRequest.shape.contentHash.isOptional(), false, "contentHash remains required");
});

test("publication Zod refinement refuses a marked artifact with regressed visibility requiredness", () => {
  assert.equal(applyPublicationZodRefinements(generatedZodSource), generatedZodSource, "the generated artifact must already carry the checked optional property");
  const regressed = replaceExactly(generatedZodSource, "    visibilityIntent: zVisibilityIntent.optional()", "    visibilityIntent: zVisibilityIntent");
  assert.throws(
    () => applyPublicationZodRefinements(regressed),
    /visibilityIntent must remain optional for legacy compatibility/,
    "a marked artifact must not hide a requiredness regression from the postprocessor",
  );
});

for (const row of corpus.metadata.cases) {
  test(`publication metadata parity: ${row.name}`, () => {
    assert.equal(zAuthoritativePublishRequest.safeParse(JSON.parse(row.input)).success, row.expected.zod_valid);
  });
}
for (const row of corpus.owner_updates.cases) {
  test(`successor owner update parity: ${row.name}`, () => {
    assert.equal(zOwnerTranscriptUpdateRequest.safeParse(JSON.parse(row.input)).success, row.expected.zod_valid);
  });
}
for (const row of corpus.owner_update_responses.cases) {
  test(`successor owner update response parity: ${row.name}`, () => {
    assert.equal(zOwnerTranscriptUpdateResponse.safeParse(JSON.parse(row.input)).success, row.expected.zod_valid);
  });
}

for (const row of corpus.operations.cases) {
  test(`publication operation parity: ${row.name}`, () => {
    assert.equal(zCanonicalPublishOperation.safeParse(JSON.parse(row.input)).success, row.expected.zod_valid);
  });
}
for (const row of corpus.responses.cases) {
  test(`publication response parity: ${row.name}`, () => {
    assert.equal(zAuthoritativePublishResponse.safeParse(JSON.parse(row.input)).success, row.expected.zod_valid);
  });
}
for (const row of corpus.nested_unknowns.cases) {
  test(`successor nested unknown parity: ${row.name}`, () => {
    const base = row.arm === "metadata" ? corpus.nested_unknowns.metadata_base : corpus.nested_unknowns.operation_base;
    const value = row.value === undefined ? true : JSON.parse(row.value);
    const input = mutateJSON(base, row.path, value);
    const validator = row.arm === "metadata" ? zAuthoritativePublishRequest : zCanonicalPublishOperation;
    assert.equal(validator.safeParse(input).success, row.zod_valid);
  });
}

function replaceExactly(source, before, after) {
  assert.equal(source.split(before).length - 1, 1, `mutation target ${JSON.stringify(before)} must occur exactly once`);
  return source.replace(before, after);
}

function mutateJSON(base, path, value) {
  const document = JSON.parse(base);
  const parts = path.split(".");
  let cursor = document;
  for (const part of parts.slice(0, -1)) cursor = cursor[part];
  cursor[parts.at(-1)] = value;
  return document;
}
