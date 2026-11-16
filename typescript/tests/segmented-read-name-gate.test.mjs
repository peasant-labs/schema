import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import { extname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { parse } from "yaml";
import { loadCorpus } from "../dist/testcase.js";

const root = fileURLToPath(new URL("../..", import.meta.url));
const readFixture = async (name) => parse(await readFile(join(root, "testdata", "typescript", name), "utf8"), {uniqueKeys: true});
const names = (await readFixture("retired_segmented_read_names.yaml")).names;
const required = (await readFixture("retired_segmented_read_names_manifest.yaml")).requiredNames;
assert.deepEqual([...names].sort(), [...required].sort(), "every retired declaration must remain protected");

// Match whole retired declarations, including generated validator/component
// prefixes. A checksum is opaque provenance, not a declaration or prose.
const alternatives = names.map((name) => name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).sort((a, b) => b.length - a.length).join("|");
const retired = new RegExp(`(?:^|[^A-Za-z0-9_$])(?:zSchema|Schema|z)?(?:${alternatives})(?:Schema)?(?=$|[^A-Za-z0-9_$])`);
const lowerIdentifier = /\b(?=[A-Za-z_$][A-Za-z0-9_$]*\b)(?=[A-Za-z0-9_$]*(?:transcript|segment|directory|window|turn|ref|context))(?=[A-Za-z0-9_$]*v2)[A-Za-z0-9_$]+\b/i;
const lowerProse = /\b(?:v2\s+(?:transcript|segment|directory|window|turn|row|ref|read|protocol|positions?)|(?:transcript|segment|directory|window|turn|row|ref|read|protocol)\s+v2)\b/i;
const checksum = /(?:h1:|sha256[-:])[A-Za-z0-9+/=]+/g;

function hasRetiredName(source) {
  const text = source.replace(checksum, "");
  return retired.test(text) || lowerIdentifier.test(text) || lowerProse.test(text);
}

const mutationSource = await readFile(join(root, "testdata", "typescript", "segmented_read_name_gate.yaml"), "utf8");
const mutations = loadCorpus(mutationSource, {decodeInput: (value) => value, decodeExpected: (value) => value});
const mutationManifest = (await readFixture("segmented_read_name_gate_manifest.yaml")).requiredNames;
assert.deepEqual(mutations.cases.map((row) => row.name).sort(), [...mutationManifest].sort());

test("segmented read name gate distinguishes retired declarations from unrelated text", async (t) => {
  for (const row of mutations.cases) await t.test(row.name, () => {
    assert.equal(hasRetiredName(row.input.source), row.expected);
  });
  for (const name of names) await t.test(`all generated forms of ${name}`, () => {
    for (const form of [name, `z${name}`, `Schema${name}`, `zSchema${name}`, `${name}Schema`]) {
      assert.equal(hasRetiredName(form), true, `${form} is not guarded`);
    }
  });
});

const sourceExtensions = new Set([".go", ".ts", ".mjs", ".json", ".yaml", ".yml", ".html", ".md"]);
const excludedDirs = new Set([".git", "node_modules", "dist", "testdata", "fixtures"]);
const excludedFiles = new Set(["go.sum", "pnpm-lock.yaml", "flake.nix"]);

async function collectSources(dir) {
  const paths = [];
  for (const entry of await readdir(dir, {withFileTypes: true})) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      if (!excludedDirs.has(entry.name) && !entry.name.startsWith(".")) paths.push(...await collectSources(path));
    } else if (entry.isFile() && !excludedFiles.has(entry.name) && sourceExtensions.has(extname(entry.name))) {
      paths.push(path);
    }
  }
  return paths;
}

test("retired segmented read names do not appear on active source or generated surfaces", async () => {
  const paths = await collectSources(root);
  // These maintained fixture rows are export and component wiring, not case
  // data; the rest of the fixture corpus keeps its stable names unchanged.
  for (const name of [
    "openapi/testdata/typescript_catalog.yaml",
    "openapi/testdata/village_transcript_reads.yaml",
    "testdata/typescript/public_exports.yaml",
    "typescript/tests/fixtures/package-files.yaml",
  ]) paths.push(join(root, name));
  const findings = [];
  for (const path of paths) {
    const text = await readFile(path, "utf8");
    if (hasRetiredName(text)) findings.push(relative(root, path));
  }
  assert.deepEqual(findings.sort(), [], "retired segmented read identifiers or prose remain");
});
