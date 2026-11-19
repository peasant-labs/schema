import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile, readdir } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

// This is the pre-change reader, not a second implementation of its rule.
// The fixture directory holds the original source files, including their
// transitive imports, at the recorded Fairtrade commit. Their digests are
// checked before loading them so this replay works offline in a clean checkout
// and fails if the oracle is substituted or edited.
const oracleDirectory = new URL("./fixtures/numbering-oracle/", import.meta.url);
const oracle = parse(await readFile(new URL("./fixtures/numbering-oracle.yaml", import.meta.url), "utf8"), { uniqueKeys: true });
const numbering = parse(await readFile(new URL("../../testdata/transcript_numbering.yaml", import.meta.url), "utf8"), { uniqueKeys: true });
const manifest = parse(await readFile(new URL("../../testdata/transcript_numbering_manifest.yaml", import.meta.url), "utf8"), { uniqueKeys: true });

assert.equal(typeof oracle.capturedFrom, "string", "the source commit must be named");
assert.match(oracle.capturedFrom, /^[0-9a-f]{40}$/);
assert.equal(numbering.capturedFrom, oracle.capturedFrom, "the corpus must name the pinned pre-change source commit");
assert.ok(oracle.files && Object.keys(oracle.files).length > 0, "the pinned source inventory must not be empty");
async function sourcePaths(directory, prefix = "") {
  const result = [];
  for (const entry of await readdir(new URL(directory, oracleDirectory), { withFileTypes: true })) {
    const path = `${prefix}${entry.name}`;
    if (entry.isDirectory()) result.push(...await sourcePaths(`${directory}${entry.name}/`, `${path}/`));
    else if (entry.isFile()) {
      // The README explains the capture; it is not a captured source file.
      if (path !== "README.md") result.push(path);
    }
    else throw new TypeError(`oracle source must not be a symlink or special file: ${path}`);
  }
  return result;
}
assert.deepEqual(
  (await sourcePaths("./")).sort(),
  Object.keys(oracle.files).sort(),
  "every oracle source must have a pinned digest and every digest must name a source",
);
for (const [path, expectedHash] of Object.entries(oracle.files)) {
  assert.match(path, /^src\/ui\/(?:transcript\/)?[a-z.-]+\.js$/, `invalid oracle source path: ${path}`);
  assert.match(expectedHash, /^[0-9a-f]{64}$/, `invalid oracle hash for ${path}`);
  const source = await readFile(new URL(path, oracleDirectory));
  assert.equal(createHash("sha256").update(source).digest("hex"), expectedHash, `${path}: pre-change source changed`);
}

const { prefilterTurns } = await import(new URL("src/ui/transcript/adapter.js", oracleDirectory));
const { computeTurnLabels } = await import(new URL("src/ui/transcript/analytics.js", oracleDirectory));

const cases = numbering.versions?.c1?.cases;
assert.ok(Array.isArray(cases) && cases.length > 0, "the original numbering corpus must be present");
assert.ok(Array.isArray(manifest.requiredNames) && manifest.requiredNames.length > 0, "the numbering manifest must be present");
const caseNames = cases.map((row) => row.name);
assert.equal(new Set(caseNames).size, caseNames.length, "numbering cases must be unique");
assert.deepEqual([...caseNames].sort(), [...manifest.requiredNames].sort(), "the oracle must visit the required numbering corpus");
assert.ok(caseNames.includes("numbering-golden-seeded-from-prechange-js"), "the capture provenance case must remain present");
assert.ok(Array.isArray(oracle.constructorOnly), "constructor-only exclusions must be named");
assert.deepEqual(
  cases.filter((row) => row.expect?.stage === "constructor").map((row) => row.name).sort(),
  [...oracle.constructorOnly].sort(),
  "only the named duplicate-key constructor refusals are excluded from the browser replay",
);

test("the pinned pre-change reader reproduces the numbering capture", async (t) => {
  for (const row of cases) {
    if (row.expect.stage === "constructor") continue;
    await t.test(row.name, () => {
      const turns = JSON.parse(row.input.turnsJSON);
      const position = new Map(turns.map((turn, i) => [turn, i]));
      let kept;
      let labels;
      let failure;
      try {
        kept = row.input.harness === "pi" ? turns : prefilterTurns(turns);
        labels = computeTurnLabels(kept);
      } catch (error) {
        failure = error;
      }
      if (row.expect.outcome === "refused") {
        assert.equal(row.expect.stage, "cook", "browser refusals must be cook refusals");
        assert.ok(failure, "the original browser rule must refuse this case");
        return;
      }
      assert.equal(row.expect.outcome, "cooked", "every case must name its outcome");
      assert.ifError(failure);
      const served = new Map(kept.map((turn, i) => [position.get(turn), { position: i, label: labels[i] }]));
      const actual = turns.map((_, i) => {
        const item = served.get(i);
        return {
          position: i,
          keep: item !== undefined,
          servedPosition: item?.position ?? -1,
          displayNumber: item?.label ?? "",
        };
      });
      // A byte comparison of the canonical JSON encoding prevents a replay
      // from passing while an expected position or number differs.
      assert.equal(JSON.stringify(actual), JSON.stringify(row.expected), `${row.name}: pre-change numbering bytes`);
    });
  }
});

test("reference-bearing rows remain in the pinned pre-change reader", async () => {
  const locator = parse(await readFile(new URL("../../testdata/transcript_entry_locate.yaml", import.meta.url), "utf8"), { uniqueKeys: true });
  const row = locator.scan.cases.find((item) => item.name === "locator-references-protect-their-rows");
  assert.ok(row, "the reference protection fixture must exist");
  const turns = JSON.parse(row.input.detail).turns;
  assert.deepEqual(prefilterTurns(turns).map((turn) => turn.index), row.expected.legacyKept);
  assert.deepEqual(turns.map((turn) => turn.index), row.expected.piKept);
  for (const index of row.expected.referenced) {
    assert.ok(row.expected.legacyKept.includes(index), `referenced row ${index} must be kept`);
  }
});
