import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parseAllDocuments } from "yaml";

import { Classification, loadCorpus } from "../dist/testcase.js";
import * as schema from "../dist/index.js";

const fixtureSource = await readFile(new URL("../../testdata/typescript/root_zod_association_validation.yaml", import.meta.url), "utf8");
const fixtureManifestSource = await readFile(new URL("../../testdata/typescript/root_zod_association_validation_manifest.yaml", import.meta.url), "utf8");
const fixture = loadRootFixture(fixtureSource);
const fixtureManifest = loadFixtureManifest(fixtureManifestSource, "root Zod association validation fixture manifest");
const sharedFixtureSource = await readFile(new URL("../../testdata/local-api/association_parity.yaml", import.meta.url), "utf8");
const sharedFixtureManifestSource = await readFile(new URL("../../testdata/local-api/association_parity_manifest.yaml", import.meta.url), "utf8");
const sharedFixture = loadSharedFixture(sharedFixtureSource);
const sharedFixtureManifest = loadFixtureManifest(sharedFixtureManifestSource, "shared association parity fixture manifest");

assertFixtureInventory(fixture, fixtureManifest, "root Zod association validation fixture");
assertFixtureInventory(sharedFixture, sharedFixtureManifest, "shared association parity fixture");

test("built root Zod schemas enforce association and annotation structural invariants", async (t) => {
  for (const testCase of fixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted parse result`);
      assert.equal(rootSchema(testCase.input.schema).safeParse(testCase.input.value).success, testCase.expected);
    });
  }
});

test("built root Zod session associations share the Go parity corpus", async (t) => {
  for (const testCase of sharedFixture.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, testCase.expected ? Classification.MustPass : Classification.MustFail, `${testCase.name}: classification must agree with the asserted parse result`);
      assert.equal(schema.zSessionAssociation.safeParse(testCase.input).success, testCase.expected);
    });
  }
});

test("root Zod association fixture manifests are strict inventories", () => {
  assert.ok(sharedFixtureManifest.requiredCaseNames.length >= 2, "shared association parity manifest needs two names for duplicate-name mutation coverage");
  const firstLine = sharedFixtureManifestSource.split("\n", 1)[0];
  assert.ok(firstLine !== undefined && firstLine !== "", "shared association parity manifest must begin with expectedCaseCount");

  assertFixtureManifestRejected(`${sharedFixtureManifestSource}\nunknown: value\n`, "unknown manifest field");
  assertFixtureManifestRejected(`${sharedFixtureManifestSource}\n${firstLine}\n`, "duplicate YAML key");
  assertFixtureManifestRejected(`${sharedFixtureManifestSource}\n---\n${firstLine}\n`, "second YAML document");
  assertFixtureManifestRejected(replaceFixtureText(sharedFixtureManifestSource, `  - ${sharedFixtureManifest.requiredCaseNames[1]}`, `  - ${sharedFixtureManifest.requiredCaseNames[0]}`), "duplicate required case name");
  assertFixtureManifestRejected(replaceFixtureText(sharedFixtureManifestSource, `  - ${sharedFixtureManifest.requiredCaseNames[0]}`, "  - \"\""), "blank required case name");
  assertFixtureManifestRejected(`expectedCaseCount: ${sharedFixtureManifest.expectedCaseCount}\nrequiredCaseNames: [1]\n`, "malformed required case name");
  assertFixtureManifestRejected(replaceFixtureText(sharedFixtureManifestSource, `expectedCaseCount: ${sharedFixtureManifest.expectedCaseCount}`, `expectedCaseCount: ${sharedFixtureManifest.expectedCaseCount + 1}`), "count and name mismatch");

  const unregisteredName = "unregistered_association_parity_case";
  const wrongNameCorpus = replaceFixtureText(sharedFixtureSource, `name: ${sharedFixtureManifest.requiredCaseNames[0]}`, `name: ${unregisteredName}`);
  assert.throws(
    () => assertFixtureInventory(loadSharedFixture(wrongNameCorpus), sharedFixtureManifest, "mutated shared association parity fixture"),
    new RegExp(unregisteredName),
    "count-preserving corpus name drift must fail the manifest inventory check",
  );
});

function rootSchema(name) {
  switch (name) {
    case "session-association": return schema.zSessionAssociation;
    case "annotation-summary": return schema.zAnnotationSummary;
    case "session-id": return schema.zSessionID;
    default: throw new TypeError(`root Zod association validation fixture named unsupported schema ${JSON.stringify(name)}`);
  }
}

function loadRootFixture(source) {
  return loadCorpus(source, {
    decodeInput(value, path) {
      const input = requireRecord(value, path, ["schema", "value"]);
      if (input.schema !== "session-association" && input.schema !== "annotation-summary" && input.schema !== "session-id") {
        throw new TypeError(`${path}.schema: must be session-association, annotation-summary, or session-id`);
      }
      if (input.schema === "session-id" && typeof input.value !== "string") {
        throw new TypeError(`${path}.value: must be a string for session-id`);
      }
      if (input.schema !== "session-id" && (typeof input.value !== "object" || input.value === null || Array.isArray(input.value))) {
        throw new TypeError(`${path}.value: must be a mapping`);
      }
      return input;
    },
    decodeExpected(value, path) {
      if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
      return value;
    },
  });
}

function loadSharedFixture(source) {
  return loadCorpus(source, {
    decodeInput(value, path) {
      return requireRecord(value, path, ["id", "sessionId", "conclusion", "confidence", "evidence"]);
    },
    decodeExpected(value, path) {
      if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
      return value;
    },
  });
}

function loadFixtureManifest(source, label) {
  let documents;
  try {
    documents = parseAllDocuments(source, { strict: true, uniqueKeys: true });
  } catch (error) {
    throw new TypeError(`${label}: YAML parsing failed`, { cause: error });
  }
  if (documents.length !== 1) {
    throw new TypeError(`${label}: expected exactly one YAML document, got ${documents.length}`);
  }
  const document = documents[0];
  if (document === undefined) {
    throw new TypeError(`${label}: YAML parser returned no document`);
  }
  if (document.errors.length > 0) {
    throw new TypeError(`${label}: YAML parsing failed: ${document.errors.map((error) => error.message).join("; ")}`);
  }

  let value;
  try {
    value = document.toJS({ maxAliasCount: 0 });
  } catch (error) {
    throw new TypeError(`${label}: YAML conversion failed`, { cause: error });
  }
  const manifest = requireExactRecord(value, label, ["expectedCaseCount", "requiredCaseNames"]);
  if (!Number.isSafeInteger(manifest.expectedCaseCount) || manifest.expectedCaseCount <= 0) {
    throw new TypeError(`${label}.expectedCaseCount: must be a positive safe integer`);
  }
  if (!Array.isArray(manifest.requiredCaseNames)) {
    throw new TypeError(`${label}.requiredCaseNames: must be a sequence`);
  }
  if (manifest.requiredCaseNames.length !== manifest.expectedCaseCount) {
    throw new TypeError(`${label}.requiredCaseNames: has ${manifest.requiredCaseNames.length} names, want exactly ${manifest.expectedCaseCount}`);
  }
  const names = new Set();
  for (const [index, name] of manifest.requiredCaseNames.entries()) {
    if (typeof name !== "string") {
      throw new TypeError(`${label}.requiredCaseNames[${index}]: must be a string`);
    }
    if (!hasNonBlankFixtureName(name)) {
      throw new TypeError(`${label}.requiredCaseNames[${index}]: must not be blank`);
    }
    if (names.has(name)) {
      throw new TypeError(`${label}.requiredCaseNames[${index}]: repeats required case name ${JSON.stringify(name)}`);
    }
    names.add(name);
  }
  return { expectedCaseCount: manifest.expectedCaseCount, requiredCaseNames: manifest.requiredCaseNames };
}

function assertFixtureInventory(corpus, manifest, label) {
  assert.equal(corpus.cases.length, manifest.expectedCaseCount, `${label}: has ${corpus.cases.length} cases, want exactly ${manifest.expectedCaseCount} from its manifest`);
  const requiredNames = new Set(manifest.requiredCaseNames);
  const actualNames = new Set();
  for (const [index, testCase] of corpus.cases.entries()) {
    assert.ok(hasNonBlankFixtureName(testCase.name), `${label}: case ${index} has a blank name`);
    assert.ok(!actualNames.has(testCase.name), `${label}: repeats case name ${JSON.stringify(testCase.name)}`);
    actualNames.add(testCase.name);
    assert.ok(requiredNames.has(testCase.name), `${label}: contains unregistered case ${JSON.stringify(testCase.name)}`);
  }
  for (const name of requiredNames) {
    assert.ok(actualNames.has(name), `${label}: is missing required case ${JSON.stringify(name)}`);
  }
}

function assertFixtureManifestRejected(source, label) {
  assert.throws(() => loadFixtureManifest(source, `mutated shared association parity manifest (${label})`), TypeError, `${label} must be rejected`);
}

function replaceFixtureText(source, oldText, replacement) {
  assert.equal(source.split(oldText).length - 1, 1, `fixture mutation target ${JSON.stringify(oldText)} must occur exactly once`);
  return source.replace(oldText, replacement);
}

function hasNonBlankFixtureName(value) {
  for (const character of value) {
    const codePoint = character.codePointAt(0);
    if (codePoint !== undefined && !isGoWhitespace(codePoint)) return true;
  }
  return false;
}

function isGoWhitespace(codePoint) {
  return (codePoint >= 0x0009 && codePoint <= 0x000D)
    || codePoint === 0x0020
    || codePoint === 0x0085
    || codePoint === 0x00A0
    || codePoint === 0x1680
    || (codePoint >= 0x2000 && codePoint <= 0x200A)
    || codePoint === 0x2028
    || codePoint === 0x2029
    || codePoint === 0x202F
    || codePoint === 0x205F
    || codePoint === 0x3000;
}

function requireExactRecord(value, path, keys) {
  const record = requireRecord(value, path, keys);
  for (const key of keys) {
    if (!Object.hasOwn(record, key)) {
      throw new TypeError(`${path}.${key}: required key is missing`);
    }
  }
  return record;
}

function requireRecord(value, path, keys) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
  const record = Object.fromEntries(Object.entries(value));
  for (const key of Object.keys(record)) {
    if (!keys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
  }
  for (const key of keys) {
    if (!Object.hasOwn(record, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  }
  return record;
}
