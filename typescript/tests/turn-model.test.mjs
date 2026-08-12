import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parseAllDocuments } from "yaml";

import { Classification, loadCorpus } from "../dist/testcase.js";
import { zObservedModelID, zTurnDetail } from "../dist/index.js";

const fixtureSource = await readFile(new URL("../../testdata/session-detail/transcripts/turn_models.yaml", import.meta.url), "utf8");
const manifestSource = await readFile(new URL("../../testdata/session-detail/transcripts/turn_models_manifest.yaml", import.meta.url), "utf8");
const fixture = loadTurnModelFixture(fixtureSource);
const manifest = loadTurnModelManifest(manifestSource);

assertTurnModelInventory(fixture, manifest, "turn model fixture");

test("generated TurnDetail and ObservedModelID schemas enforce the shared observed-model corpus", async (t) => {
  for (const fixtureCase of fixture.cases) {
    await t.test(fixtureCase.name, () => {
      const { observedModel, ...baseTurn } = fixtureCase.input;
      const input = fixtureCase.expected.observedModelPresent ? { ...baseTurn, observedModel } : baseTurn;
      const identifierResult = fixtureCase.expected.observedModelPresent
        ? zObservedModelID.safeParse(observedModel)
        : { success: true, data: undefined };
      const turnResult = zTurnDetail.safeParse(input);

      assert.equal(fixtureCase.classification, fixtureCase.expected.accepted ? Classification.MustPass : Classification.MustFail, `${fixtureCase.name}: classification must agree with expected acceptance`);
      assert.equal(identifierResult.success, fixtureCase.expected.accepted, `${fixtureCase.name}: generated ObservedModelID schema acceptance differs from the fixture`);
      assert.equal(turnResult.success, fixtureCase.expected.accepted, `${fixtureCase.name}: generated TurnDetail runtime validation differs from the fixture`);

      if (!fixtureCase.expected.accepted) {
        const issueText = turnResult.success ? "" : turnResult.error.issues.map((issue) => issue.message).join("; ");
        assert.match(issueText, new RegExp(fixtureCase.expected.errorContains, "i"), `${fixtureCase.name}: generated runtime diagnostic must identify the rejected boundary`);
        return;
      }
      assert.equal(turnResult.success, true);
      if (!turnResult.success) return;
      assert.equal(Object.hasOwn(turnResult.data, "observedModel"), fixtureCase.expected.observedModelPresent, `${fixtureCase.name}: generated schema changed observedModel presence`);
      if (fixtureCase.expected.observedModelPresent) {
        assert.equal(turnResult.data.observedModel, fixtureCase.expected.observedModel, `${fixtureCase.name}: generated TurnDetail validation changed observed model bytes`);
        assert.equal(identifierResult.data, fixtureCase.expected.observedModel, `${fixtureCase.name}: generated identifier validation changed observed model bytes`);
      }
    });
  }
});

test("turn model fixture inventory rejects a count-preserving rename", () => {
  const oldName = `name: ${manifest.requiredCaseNames[0]}`;
  const mutatedSource = replaceFixtureText(fixtureSource, oldName, "name: unregistered-observed-model-case");
  const mutatedFixture = loadTurnModelFixture(mutatedSource);
  assert.throws(
    () => assertTurnModelInventory(mutatedFixture, manifest, "mutated turn model fixture"),
    /unregistered-observed-model-case/,
    "count-preserving fixture name drift must fail the independent inventory",
  );
});

function loadTurnModelFixture(source) {
  return loadCorpus(source, {
    decodeInput(value, path) {
      const input = requireExactRecord(value, path, ["index", "role", "content", "timestamp", "depth"], ["observedModel"]);
      requireInteger(input.index, `${path}.index`);
      requireString(input.role, `${path}.role`);
      requireString(input.content, `${path}.content`);
      requireString(input.timestamp, `${path}.timestamp`);
      requireInteger(input.depth, `${path}.depth`);
      return input;
    },
    decodeExpected(value, path) {
      const expected = requireExactRecord(value, path, ["accepted", "observedModelPresent"], ["observedModel", "errorContains"]);
      if (typeof expected.accepted !== "boolean") throw new TypeError(`${path}.accepted: must be a boolean`);
      if (typeof expected.observedModelPresent !== "boolean") throw new TypeError(`${path}.observedModelPresent: must be a boolean`);
      if (Object.hasOwn(expected, "observedModel")) requireString(expected.observedModel, `${path}.observedModel`);
      if (Object.hasOwn(expected, "errorContains")) requireString(expected.errorContains, `${path}.errorContains`);
      if (expected.accepted && expected.errorContains !== undefined) throw new TypeError(`${path}.errorContains: must be omitted for accepted rows`);
      if (!expected.accepted && expected.errorContains === undefined) throw new TypeError(`${path}.errorContains: is required for rejected rows`);
      return expected;
    },
  });
}

function loadTurnModelManifest(source) {
  const documents = parseAllDocuments(source, { strict: true, uniqueKeys: true });
  if (documents.length !== 1 || documents[0] === undefined || documents[0].errors.length !== 0) {
    throw new TypeError("turn model fixture manifest must be exactly one valid YAML document");
  }
  const manifest = requireExactRecord(documents[0].toJS({ maxAliasCount: 0 }), "turn model fixture manifest", ["expectedCaseCount", "requiredCaseNames"], []);
  if (!Number.isSafeInteger(manifest.expectedCaseCount) || manifest.expectedCaseCount <= 0) throw new TypeError("turn model fixture manifest expectedCaseCount must be a positive safe integer");
  if (!Array.isArray(manifest.requiredCaseNames) || manifest.requiredCaseNames.length !== manifest.expectedCaseCount) throw new TypeError("turn model fixture manifest requiredCaseNames must match expectedCaseCount exactly");
  const names = manifest.requiredCaseNames.map((name, index) => requireName(name, `turn model fixture manifest.requiredCaseNames[${index}]`));
  if (new Set(names).size !== names.length) throw new TypeError("turn model fixture manifest requiredCaseNames must be unique");
  return { expectedCaseCount: manifest.expectedCaseCount, requiredCaseNames: names };
}

function assertTurnModelInventory(corpus, inventory, label) {
  assert.equal(corpus.cases.length, inventory.expectedCaseCount, `${label}: case count must match the independent manifest exactly`);
  const requiredNames = new Set(inventory.requiredCaseNames);
  const actualNames = new Set();
  for (const fixtureCase of corpus.cases) {
    assert.ok(!actualNames.has(fixtureCase.name), `${label}: repeats case name ${JSON.stringify(fixtureCase.name)}`);
    actualNames.add(fixtureCase.name);
    assert.ok(requiredNames.has(fixtureCase.name), `${label}: contains unregistered case ${JSON.stringify(fixtureCase.name)}`);
  }
  for (const name of requiredNames) assert.ok(actualNames.has(name), `${label}: is missing required case ${JSON.stringify(name)}`);
}

function requireExactRecord(value, path, requiredKeys, optionalKeys) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
  const record = Object.fromEntries(Object.entries(value));
  const allowed = new Set([...requiredKeys, ...optionalKeys]);
  for (const key of Object.keys(record)) if (!allowed.has(key)) throw new TypeError(`${path}.${key}: unknown key`);
  for (const key of requiredKeys) if (!Object.hasOwn(record, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  return record;
}

function requireInteger(value, path) {
  if (!Number.isSafeInteger(value)) throw new TypeError(`${path}: must be a safe integer`);
}

function requireString(value, path) {
  if (typeof value !== "string") throw new TypeError(`${path}: must be a string`);
  return value;
}

function requireName(value, path) {
  requireString(value, path);
  if (value.trim() === "" || value !== value.trim()) throw new TypeError(`${path}: must be non-empty without edge whitespace`);
  return value;
}

function replaceFixtureText(source, oldText, replacement) {
  assert.equal(source.split(oldText).length - 1, 1, `fixture mutation target ${JSON.stringify(oldText)} must occur exactly once`);
  return source.replace(oldText, replacement);
}
