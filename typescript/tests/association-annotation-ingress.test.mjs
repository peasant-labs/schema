import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parseAllDocuments, stringify } from "yaml";

import { Classification, loadCorpus } from "../dist/testcase.js";
import * as schema from "../dist/index.js";

const source = await readFile(new URL("../../testdata/publish/association_annotation_ingress.yaml", import.meta.url), "utf8");
const fixture = loadIngressFixture(source);

test("built root Zod schemas enforce published association and annotation ingress invariants", async (t) => {
  assert.ok(fixture.cases.cases.length >= 9, "association annotation ingress corpus must retain its validation floor");
  for (const testCase of fixture.cases.cases) {
    await t.test(testCase.name, () => {
      const item = {
        ...testCase.input.annotation,
        contentHash: "fixture-content-hash",
        typeId: "fixture.annotation",
        value: "fixture-value",
        isPrimary: false,
      };
      const accepted = schema.zGitContext.safeParse({ associations: testCase.input.associations }).success
        && schema.zAnnotationPushItem.safeParse(item).success;
      assert.equal(accepted, testCase.expected, `${testCase.name}: root Zod ingress verdict must match the shared corpus`);
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted verdict`);
    });
  }
});

test("association annotation ingress fixture strict decoding rejects unknown keys, duplicates, and trailing documents", async (t) => {
  assert.ok(fixture.strictDecoding.cases.length >= 4, "association annotation ingress strict corpus must retain its validation floor");
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
  const root = requireExactRecord(value, "association annotation ingress fixture", ["cases", "strict_decoding"]);
  return {
    cases: loadCorpus(stringify({ cases: root.cases }), {
      decodeInput: decodeIngressInput,
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
  requireRecord(input.annotation, `${path}.annotation`, ["targetKind", "targetAssociationId", "sessionId", "annotationId", "projectHash"]);
  if (input.hashComparison !== undefined) {
    requireExactRecord(input.hashComparison, `${path}.hashComparison`, ["alternateTargetAssociationId", "distinct"]);
  }
  return input;
}

function decodeExpectedBoolean(value, path) {
  if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
  return value;
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
