import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parseAllDocuments, stringify } from "yaml";

import { Classification, loadCorpus } from "../dist/testcase.js";
import * as schema from "../dist/index.js";

const source = await readFile(new URL("../../testdata/publish/association_annotation_ingress.yaml", import.meta.url), "utf8");
const fixture = loadIngressFixture(source);

test("built root Zod schemas enforce published association and annotation ingress invariants", async (t) => {
  assert.ok(fixture.cases.cases.length >= 11, "association annotation ingress corpus must retain its validation floor");
  for (const testCase of fixture.cases.cases) {
    await t.test(testCase.name, () => {
      const item = {
        ...testCase.input.annotation,
        contentHash: "fixture-content-hash",
        typeId: "fixture.annotation",
        value: "fixture-value",
        isPrimary: false,
      };
      const publishAccepted = schema.zGitContext.safeParse({ associations: testCase.input.associations }).success;
      const annotationAccepted = schema.zAnnotationPushItem.safeParse(item).success;
      assert.equal(publishAccepted, testCase.expected.publishRequestValid, `${testCase.name}: root Zod published-association verdict must match the shared corpus`);
      assert.equal(annotationAccepted, testCase.expected.annotationRequestValid, `${testCase.name}: root Zod annotation target verdict must match the shared corpus`);
      assert.equal(testCase.classification, publishAccepted && annotationAccepted ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the combined asserted verdict`);
    });
  }
});

test("built root Zod annotation request schema rejects null and accepts an empty batch", async (t) => {
  assert.ok(fixture.annotationRequestShapes.cases.length >= 2, "annotation request shape corpus must retain its validation floor");
  for (const testCase of fixture.annotationRequestShapes.cases) {
    await t.test(testCase.name, () => {
      const accepted = schema.zAnnotationPushRequest.safeParse(testCase.input).success;
      assert.equal(accepted, testCase.expected, `${testCase.name}: root Zod annotation request verdict must match the shared corpus`);
      assert.equal(testCase.classification, accepted ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted verdict`);
    });
  }
});

test("built root Zod entry target validator matches the shared malformed-entry rows", async (t) => {
  const directEntryCases = fixture.cases.cases.filter((testCase) => testCase.expected.annotationEntryTargetValid !== undefined);
  assert.deepEqual(directEntryCases.map((testCase) => testCase.name).sort(), [
    "entry annotation is valid",
    "entry annotation rejects empty session ID",
    "entry annotation rejects equal range at the runtime boundary",
    "entry annotation rejects reversed range at the runtime boundary",
  ].sort(), "direct entry target coverage must retain the shared valid and malformed rows");
  for (const testCase of directEntryCases) {
    await t.test(testCase.name, () => {
      const accepted = schema.zAnnotationEntryTarget.safeParse(testCase.input.annotation.entryTarget).success;
      assert.equal(accepted, testCase.expected.annotationEntryTargetValid, `${testCase.name}: exported root Zod entry target verdict must match the shared corpus`);
    });
  }
});

test("association annotation ingress corpus covers every target-kind arm", () => {
  const expectedKinds = new Set(schema.AllTargetKinds);
  const coverage = new Map();
  for (const testCase of fixture.cases.cases) {
    const kind = testCase.input.annotation.targetKind;
    assert.ok(expectedKinds.has(kind), `${testCase.name}: fixture has an unexpected target kind ${JSON.stringify(kind)}`);
    const observed = coverage.get(kind) ?? { valid: 0, invalid: 0 };
    if (testCase.expected.annotationRequestValid) observed.valid += 1;
    else observed.invalid += 1;
    if (["targetAssociationId", "sessionId", "entryTarget", "annotationId", "projectHash"].some((field) => Object.hasOwn(testCase.input.annotation, field) && testCase.input.annotation[field] === null)) {
      observed.explicitNull = (observed.explicitNull ?? 0) + 1;
    }
    coverage.set(kind, observed);
  }
  assert.deepEqual([...coverage.keys()].sort(), [...expectedKinds].sort(), "fixture target kinds must exactly match the public TargetKind catalog");
  for (const kind of schema.AllTargetKinds) {
    const observed = coverage.get(kind);
    assert.ok(observed, `fixture target kinds are missing ${kind}`);
    if (kind === schema.TargetKind.FileVersion) {
      assert.equal(observed.valid, 0, "file_version must have no valid AnnotationPushItem representation");
      assert.ok(observed.invalid > 0, "file_version must retain a rejection case");
      continue;
    }
    assert.ok(observed.valid > 0, `${kind} must retain a valid AnnotationPushItem case`);
    assert.ok(observed.invalid > 0, `${kind} must retain an invalid AnnotationPushItem case`);
    assert.ok((observed.explicitNull ?? 0) > 0, `${kind} must retain an explicit-null inactive-arm case`);
  }
});

test("association annotation ingress fixture strict decoding rejects unknown keys, duplicates, and trailing documents", async (t) => {
  const expectedCaseNames = [
    "canonical ingress corpus shape is accepted",
    "unknown ingress corpus field is rejected",
    "duplicate ingress corpus field is rejected",
    "unknown annotation field is rejected",
    "duplicate annotation field is rejected",
    "unknown entry target field is rejected",
    "duplicate entry target field is rejected",
    "trailing ingress corpus document is rejected",
  ];
  assert.deepEqual(fixture.strictDecoding.cases.map((testCase) => testCase.name).sort(), expectedCaseNames.sort(), "association annotation ingress strict corpus must retain its exact inventory");
  for (const testCase of fixture.strictDecoding.cases) {
    await t.test(testCase.name, () => {
      let accepted = true;
      try {
        loadIngressFixture(testCase.input);
      } catch {
        accepted = false;
      }
      assert.equal(accepted, testCase.expected, `${testCase.name}: strict loader verdict must match the shared corpus`);
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted verdict`);
    });
  }
});

function loadIngressFixture(raw) {
  const documents = parseAllDocuments(raw, { strict: true, uniqueKeys: true });
  assert.equal(documents.length, 1, "association annotation ingress fixture must contain exactly one YAML document");
  const document = documents[0];
  assert.ok(document !== undefined, "association annotation ingress fixture parser returned no document");
  assert.equal(document.errors.length, 0, `association annotation ingress fixture YAML parsing failed: ${document.errors.map((error) => error.message).join("; ")}`);
  const value = document.toJS({ maxAliasCount: 0 });
  const root = requireExactRecord(value, "association annotation ingress fixture", ["cases", "annotation_request_shapes", "strict_decoding"]);
  return {
    cases: loadCorpus(stringify({ cases: root.cases }), {
      decodeInput: decodeIngressInput,
      decodeExpected: decodeIngressExpected,
    }),
    annotationRequestShapes: loadCorpus(stringify({ cases: root.annotation_request_shapes.cases }), {
      decodeInput(value, path) {
        return requireExactRecord(value, path, ["annotations"]);
      },
      decodeExpected: decodeExpectedBoolean,
    }),
    strictDecoding: loadCorpus(stringify({ cases: root.strict_decoding.cases }), {
      decodeInput(value, path) {
        if (typeof value !== "string") throw new TypeError(`${path}: must be a YAML string`);
        return value;
      },
      decodeExpected: decodeExpectedBoolean,
    }),
  };
}

function decodeIngressInput(value, path) {
  const input = requireRecord(value, path, ["associations", "annotation", "hashComparison"]);
  for (const key of ["associations", "annotation"]) {
    if (!Object.hasOwn(input, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  }
  if (!Array.isArray(input.associations)) throw new TypeError(`${path}.associations: must be a sequence`);
  for (const [index, association] of input.associations.entries()) {
    requireExactRecord(association, `${path}.associations[${index}]`, ["id", "observedCommitHash"]);
  }
  const annotation = requireRecord(input.annotation, `${path}.annotation`, ["targetKind", "targetAssociationId", "sessionId", "entryTarget", "annotationId", "projectHash"]);
  if (annotation.entryTarget !== undefined && annotation.entryTarget !== null) {
    requireExactRecord(annotation.entryTarget, `${path}.annotation.entryTarget`, ["sessionId", "entryIndex", "endIndex"]);
  }
  if (input.hashComparison !== undefined) {
    requireExactRecord(input.hashComparison, `${path}.hashComparison`, ["alternateTargetAssociationId", "distinct"]);
  }
  return input;
}

function decodeExpectedBoolean(value, path) {
  if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
  return value;
}

function decodeIngressExpected(value, path) {
  const expected = requireRecord(value, path, ["publishRequestValid", "annotationRequestValid", "annotationOperationSchemaValid", "annotationEntryTargetValid"]);
  if (typeof expected.publishRequestValid !== "boolean") throw new TypeError(`${path}.publishRequestValid: must be a boolean`);
  if (typeof expected.annotationRequestValid !== "boolean") throw new TypeError(`${path}.annotationRequestValid: must be a boolean`);
  if (expected.annotationOperationSchemaValid !== undefined && typeof expected.annotationOperationSchemaValid !== "boolean") throw new TypeError(`${path}.annotationOperationSchemaValid: must be a boolean when present`);
  if (expected.annotationEntryTargetValid !== undefined && typeof expected.annotationEntryTargetValid !== "boolean") throw new TypeError(`${path}.annotationEntryTargetValid: must be a boolean when present`);
  return expected;
}

function requireExactRecord(value, path, keys) {
  const record = requireRecord(value, path, keys);
  for (const key of keys) {
    if (!Object.hasOwn(record, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  }
  return record;
}

function requireRecord(value, path, keys) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
  const record = Object.fromEntries(Object.entries(value));
  for (const key of Object.keys(record)) {
    if (!keys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
  }
  return record;
}
