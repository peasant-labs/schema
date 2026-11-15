import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse, parseAllDocuments } from "yaml";

import { isTranscriptDirectoryPositionsValid } from "../dist/index.js";
import { Classification, loadCorpus } from "../dist/testcase.js";

// isTranscriptDirectoryPositionsValid is the browser mirror of Go's
// validateDirectoryConversationTurns. Two Go corpora hold the shared verdicts:
// testdata/transcript_segments.yaml (directory cases the Go validator test
// asserts) and testdata/transcript_cut.yaml (the cut's recorded spans). The Go
// suites run the same files, so a rule change that lands on only one side fails
// one of them; shared fixtures detect disagreement only for the cases they
// cover. The browser-only corpus holds the inputs the shared corpus cannot
// carry: explicit nulls, wrong types, unsafe integers, broken containers.
//
// A manifest names every case of each corpus and the suite asserts exact
// membership in both directions, so deleting or renaming the only test for a
// branch cannot leave every gate green.

const browserCorpusSource = await readFile(new URL("../../testdata/typescript/transcript_directory_positions.yaml", import.meta.url), "utf8");
const browserManifestSource = await readFile(new URL("../../testdata/typescript/transcript_directory_positions_manifest.yaml", import.meta.url), "utf8");
const browserCorpus = loadBrowserCorpus(browserCorpusSource);
const browserManifest = loadManifest(browserManifestSource, "browser directory position manifest");
assertInventory(browserCorpus.cases, browserManifest, "browser directory position corpus");

const sharedCorpusSource = await readFile(new URL("../../testdata/transcript_segments.yaml", import.meta.url), "utf8");
const sharedManifestSource = await readFile(new URL("../../testdata/typescript/transcript_directory_positions_shared_manifest.yaml", import.meta.url), "utf8");
const sharedDocument = parse(sharedCorpusSource, { uniqueKeys: true });
const sharedCases = sharedDocument?.shape_cases?.cases;
assert.ok(Array.isArray(sharedCases) && sharedCases.length > 0, "testdata/transcript_segments.yaml must hold at least one shape case");
const sharedManifest = loadManifest(sharedManifestSource, "shared directory position manifest");
const positionCases = sharedCases.filter(carriesDirectoryPositions);
assertInventory(positionCases, sharedManifest, "shared directory position selection");
assert.ok(
  sharedManifest.requiredCaseNames.includes("directory-total-conversation-turns-negative"),
  "the shared selection must name directory-total-conversation-turns-negative: it carries no span and is the case that proves the total is checked",
);

const cutCorpusSource = await readFile(new URL("../../testdata/transcript_cut.yaml", import.meta.url), "utf8");
const cutManifestSource = await readFile(new URL("../../testdata/transcript_cut_manifest.yaml", import.meta.url), "utf8");
const cutCorpus = parse(cutCorpusSource, { uniqueKeys: true });
const cutManifest = loadRequiredNamesManifest(cutManifestSource, "transcript cut manifest");
const cutCases = cutCorpus?.versions?.c1;
assert.ok(Array.isArray(cutCases) && cutCases.length > 0, "testdata/transcript_cut.yaml must hold the original c1 cases");

test("browser-only position corpus matches the helper verdict and leaves the input unchanged", async (t) => {
  for (const testCase of browserCorpus.cases) {
    await t.test(testCase.name, () => {
      const decoded = JSON.parse(testCase.input.json);
      const snapshot = structuredClone(decoded);
      assert.equal(
        testCase.classification,
        testCase.expected ? Classification.MustPass : Classification.MustFail,
        `${testCase.name}: classification must agree with the expected verdict`,
      );
      // The parsed number is asserted before the verdict: a lexeme JSON.parse
      // has already rounded cannot be told apart from a safe one afterwards.
      for (const claim of testCase.input.numbers) {
        assert.equal(
          Number.isSafeInteger(resolvePointer(decoded, claim.pointer)),
          claim.safeInteger,
          `${testCase.name}: ${claim.pointer} must parse to safeInteger=${claim.safeInteger}`,
        );
      }
      assert.equal(isTranscriptDirectoryPositionsValid(decoded), testCase.expected, `${testCase.name}: position verdict`);
      assert.deepEqual(decoded, snapshot, `${testCase.name}: the helper mutated the decoded directory`);
    });
  }
});

test("shared directory cases carry the same position verdict in the browser", async (t) => {
  for (const testCase of positionCases) {
    await t.test(testCase.name, () => {
      const verdict = isTranscriptDirectoryPositionsValid(structuredClone(testCase.input.payload));
      // expected.valid is the shared verdict: the Go test asserts it on the wire
      // validator and, for a rejected row, that the position-stripped payload
      // validates. A separate browser-only verdict column would let the two
      // sides drift while both suites stayed green.
      assert.equal(verdict, sharedVerdict(testCase.expected), `${testCase.name}: position verdict`);
    });
  }
});

test("the cut's recorded spans form a valid position chain", async () => {
  // TestCutTranscriptSegments verifies these spans in Go. This replay reads the
  // recorded segments as ordered directory refs with the total omitted and
  // asserts only the position verdict: no segment ordinal is read as a turn
  // identity, and this test does not run Go.
  const visited = new Set();
  let statedSpans = 0;
  for (const cutCase of cutCases) {
    const segments = cutCase.expected?.segments;
    assert.ok(Array.isArray(segments), `${cutCase.name}: a cut case must record its expected segments`);
    const refs = segments.map((segment) => {
      const ref = { index: segment.index, firstTurn: segment.firstTurn, lastTurn: segment.lastTurn, byteLength: segment.byteLength };
      if (segment.conversationTurns !== undefined) {
        statedSpans += 1;
        ref.conversationTurns = { before: segment.conversationTurns.before, count: segment.conversationTurns.count };
      }
      return ref;
    });
    assert.equal(isTranscriptDirectoryPositionsValid({ segments: refs }), true, `${cutCase.name}: the recorded spans must form a valid position chain`);
    visited.add(cutCase.name);
  }
  assert.ok(statedSpans > 0, "no cut case states a conversation span; the replay would prove nothing");
  assert.deepEqual([...visited].sort(), [...cutManifest.requiredNames].sort(), "the replay must visit every case named by its manifest");
});

test("the position manifests are strict inventories", () => {
  assert.ok(browserManifest.requiredCaseNames.some((name) => name === "directory-explicit-null-total"), "the browser manifest must name the explicit-null branch");
  assert.ok(sharedManifest.requiredCaseNames.length >= 2, "the shared manifest needs two names for duplicate-name mutation coverage");

  assertManifestRejected(loadManifest, `${browserManifestSource}\nunknown: value\n`, "browser unknown manifest field");
  assertManifestRejected(loadManifest, `${browserManifestSource}\nexpectedCaseCount: ${browserManifest.expectedCaseCount}\n`, "browser duplicate YAML key");
  assertManifestRejected(loadManifest, replaceFixtureText(browserManifestSource, `expectedCaseCount: ${browserManifest.expectedCaseCount}`, `expectedCaseCount: ${browserManifest.expectedCaseCount + 1}`), "browser count and name mismatch");
  assertManifestRejected(loadManifest, replaceFixtureText(browserManifestSource, `  - ${browserManifest.requiredCaseNames[1]}`, `  - ${browserManifest.requiredCaseNames[0]}`), "browser duplicate required case name");
  assertManifestRejected(loadManifest, replaceFixtureText(browserManifestSource, `  - ${browserManifest.requiredCaseNames[0]}`, '  - ""'), "browser blank required case name");
  assertManifestRejected(loadManifest, replaceFixtureText(sharedManifestSource, `expectedCaseCount: ${sharedManifest.expectedCaseCount}`, `expectedCaseCount: ${sharedManifest.expectedCaseCount - 1}`), "shared count and name mismatch");

  assertManifestRejected(loadRequiredNamesManifest, replaceFixtureText(cutManifestSource, `  - ${cutManifest.requiredNames[1]}`, `  - ${cutManifest.requiredNames[0]}`), "cut duplicate required case name");
  assertManifestRejected(loadRequiredNamesManifest, replaceFixtureText(cutManifestSource, `  - ${cutManifest.requiredNames[0]}`, '  - " "'), "cut blank required case name");
  // A NEL-only name is blank to Go's fixture loaders but not to String.trim;
  // this proves the loader carries the Go predicate rather than the trim shortcut.
  assertManifestRejected(loadRequiredNamesManifest, replaceFixtureText(cutManifestSource, `  - ${cutManifest.requiredNames[0]}`, '  - "\\u0085"'), "cut NEL-only required case name");
  assertManifestRejected(loadRequiredNamesManifest, "requiredNames: []\n", "cut empty required case list");
  assertManifestRejected(loadRequiredNamesManifest, `${cutManifestSource}\n---\nrequiredNames:\n  - ${cutManifest.requiredNames[0]}\n`, "cut second YAML document");
  assertManifestRejected(loadRequiredNamesManifest, `${cutManifestSource}\nunknown: value\n`, "cut unknown manifest field");
  assertManifestRejected(loadRequiredNamesManifest, "requiredNames: [1]\n", "cut malformed required case name");

  assertInventoryRejected(
    replaceFixtureText(browserCorpusSource, `name: ${browserManifest.requiredCaseNames[0]}`, "name: unregistered_directory_position_case"),
    "browser",
    "renamed browser case",
  );
  assertInventoryRejected(
    replaceFixtureText(sharedCorpusSource, `name: ${sharedManifest.requiredCaseNames[0]}`, "name: unregistered_shared_directory_case"),
    "shared",
    "renamed shared case",
  );
  assert.deepEqual(
    cutManifest.requiredNames.slice().sort(),
    cutCases.map((cutCase) => cutCase.name).sort(),
    "the cut replay manifest must name exactly the recorded cut cases",
  );
});

function loadBrowserCorpus(source) {
  return loadCorpus(source, {
    decodeInput(value, path) {
      const input = requireRecord(value, path, ["json", "numbers"]);
      if (typeof input.json !== "string") throw new TypeError(`${path}.json: must be a JSON document string`);
      const claims = input.numbers;
      if (claims !== undefined && !Array.isArray(claims)) throw new TypeError(`${path}.numbers: must be a sequence when present`);
      const numbers = (claims ?? []).map((claim, index) => {
        const row = requireRecord(claim, `${path}.numbers[${index}]`, ["pointer", "safeInteger"]);
        if (typeof row.pointer !== "string" || !row.pointer.startsWith("/")) throw new TypeError(`${path}.numbers[${index}].pointer: must be a JSON pointer starting with /`);
        if (typeof row.safeInteger !== "boolean") throw new TypeError(`${path}.numbers[${index}].safeInteger: must be a boolean`);
        return { pointer: row.pointer, safeInteger: row.safeInteger };
      });
      return { json: input.json, numbers };
    },
    decodeExpected(value, path) {
      if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean verdict`);
      return value;
    },
  });
}

function loadManifest(source, label) {
  const manifest = requireExactRecord(decodeSingleDocument(source, label), label, ["expectedCaseCount", "requiredCaseNames"]);
  if (!Number.isSafeInteger(manifest.expectedCaseCount) || manifest.expectedCaseCount <= 0) {
    throw new TypeError(`${label}.expectedCaseCount: must be a positive safe integer`);
  }
  const names = requireNameList(manifest.requiredCaseNames, label, "requiredCaseNames");
  if (names.length !== manifest.expectedCaseCount) {
    throw new TypeError(`${label}.requiredCaseNames: has ${names.length} names, want exactly ${manifest.expectedCaseCount}`);
  }
  return { expectedCaseCount: manifest.expectedCaseCount, requiredCaseNames: names };
}

function loadRequiredNamesManifest(source, label) {
  const manifest = requireExactRecord(decodeSingleDocument(source, label), label, ["requiredNames"]);
  return { requiredNames: requireNameList(manifest.requiredNames, label, "requiredNames") };
}

function decodeSingleDocument(source, label) {
  let documents;
  try {
    documents = parseAllDocuments(source, { strict: true, uniqueKeys: true });
  } catch (error) {
    throw new TypeError(`${label}: YAML parsing failed`, { cause: error });
  }
  if (documents.length !== 1) throw new TypeError(`${label}: expected exactly one YAML document, got ${documents.length}`);
  const document = documents[0];
  if (document === undefined) throw new TypeError(`${label}: YAML parser returned no document`);
  if (document.errors.length > 0) throw new TypeError(`${label}: YAML parsing failed: ${document.errors.map((error) => error.message).join("; ")}`);
  try {
    return document.toJS({ maxAliasCount: 0 });
  } catch (error) {
    throw new TypeError(`${label}: YAML conversion failed`, { cause: error });
  }
}

// requireNameList is the one name-list check both manifest loaders call: a
// non-empty list of unique, non-blank names. Blankness uses Go's unicode.IsSpace
// set, matching the sibling suite; JavaScript trim() would accept a name made of
// U+0085, which Go's fixture loaders refuse. The sibling suite keeps its own
// copy because extracting a support module shared with an unrelated corpus is
// more churn than this change carries, so the two copies must stay in step.
function requireNameList(value, label, field) {
  if (!Array.isArray(value)) throw new TypeError(`${label}.${field}: must be a sequence`);
  if (value.length === 0) throw new TypeError(`${label}.${field}: must name at least one case`);
  const seen = new Set();
  for (const [index, name] of value.entries()) {
    if (typeof name !== "string") throw new TypeError(`${label}.${field}[${index}]: must be a string`);
    if (!hasNonBlankFixtureName(name)) throw new TypeError(`${label}.${field}[${index}]: must not be blank`);
    if (seen.has(name)) throw new TypeError(`${label}.${field}[${index}]: repeats required case name ${JSON.stringify(name)}`);
    seen.add(name);
  }
  return value;
}

function hasNonBlankFixtureName(value) {
  for (const character of value) {
    const codePoint = character.codePointAt(0);
    if (codePoint !== undefined && !isGoWhitespace(codePoint)) return true;
  }
  return false;
}

// isGoWhitespace enumerates unicode.IsSpace, which decides what Go's fixture
// loaders treat as a blank name.
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

function assertInventory(cases, manifest, label) {
  assert.ok(manifest.requiredCaseNames.length > 0, `${label}: the manifest names no cases`);
  assert.equal(cases.length, manifest.expectedCaseCount, `${label}: has ${cases.length} cases, want exactly ${manifest.expectedCaseCount} from its manifest`);
  const required = new Set(manifest.requiredCaseNames);
  const actual = new Set();
  for (const [index, testCase] of cases.entries()) {
    assert.ok(typeof testCase.name === "string" && testCase.name.trim() !== "", `${label}: case ${index} has a blank name`);
    assert.ok(!actual.has(testCase.name), `${label}: repeats case name ${JSON.stringify(testCase.name)}`);
    actual.add(testCase.name);
    assert.ok(required.has(testCase.name), `${label}: contains unregistered case ${JSON.stringify(testCase.name)}`);
  }
  for (const name of required) assert.ok(actual.has(name), `${label}: is missing required case ${JSON.stringify(name)}`);
}

function assertManifestRejected(loader, source, label) {
  assert.throws(
    () => loader(source, `mutated manifest (${label})`),
    TypeError,
    `manifest mutation (${label}) must be rejected`,
  );
}

function assertInventoryRejected(source, kind, label) {
  assert.throws(
    () => {
      const cases = kind === "browser" ? loadBrowserCorpus(source).cases : loadPositionSelection(source);
      assertInventory(cases, kind === "browser" ? browserManifest : sharedManifest, `mutated ${kind} corpus`);
    },
    /unregistered/i,
    `${label} must fail the manifest inventory check`,
  );
}

function loadPositionSelection(source) {
  const document = parse(source, { uniqueKeys: true });
  const cases = document?.shape_cases?.cases;
  if (!Array.isArray(cases)) throw new TypeError("mutated shared corpus: shape_cases.cases is missing");
  return cases.filter(carriesDirectoryPositions);
}

// carriesDirectoryPositions is the shared selection predicate, matching Go's
// carriesDirectoryPositions in transcript_segments_test.go: the directory states
// the total key, or a ref states the conversationTurns key. Presence, not a
// value and not a name prefix.
function carriesDirectoryPositions(testCase) {
  if (testCase?.input?.schema !== "TranscriptSegmentDirectory") return false;
  const payload = testCase.input.payload;
  if (!isRecord(payload)) return false;
  if (Object.hasOwn(payload, "totalConversationTurns")) return true;
  return Array.isArray(payload.segments) && payload.segments.some((ref) => isRecord(ref) && Object.hasOwn(ref, "conversationTurns"));
}

function sharedVerdict(expected) {
  if (typeof expected === "boolean") return expected;
  if (isRecord(expected) && typeof expected.valid === "boolean") return expected.valid;
  throw new TypeError("a shared directory case must state expected as a boolean or as {valid: boolean}");
}

function resolvePointer(value, pointer) {
  let current = value;
  for (const token of pointer.slice(1).split("/")) {
    const key = token.replaceAll("~1", "/").replaceAll("~0", "~");
    if (!isRecord(current) && !Array.isArray(current)) throw new TypeError(`pointer ${JSON.stringify(pointer)} does not resolve`);
    if (!Object.hasOwn(current, key)) throw new TypeError(`pointer ${JSON.stringify(pointer)} does not resolve`);
    current = current[key];
  }
  return current;
}

function replaceFixtureText(source, oldText, replacement) {
  assert.equal(source.split(oldText).length - 1, 1, `fixture mutation target ${JSON.stringify(oldText)} must occur exactly once`);
  return source.replace(oldText, replacement);
}

function requireExactRecord(value, path, keys) {
  const record = requireRecord(value, path, keys);
  for (const key of keys) if (!Object.hasOwn(record, key)) throw new TypeError(`${path}.${key}: required key is missing`);
  return record;
}

function requireRecord(value, path, keys) {
  if (!isRecord(value)) throw new TypeError(`${path}: must be a mapping`);
  const record = Object.fromEntries(Object.entries(value));
  for (const key of Object.keys(record)) if (!keys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
  return record;
}

function isRecord(value) {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
