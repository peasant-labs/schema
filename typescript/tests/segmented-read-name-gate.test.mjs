import assert from "node:assert/strict";
import { mkdir, mkdtemp, readdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, extname, join, relative, sep } from "node:path";
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
const legacyGeneration = ["v", "2"].join("");
const lowerIdentifier = new RegExp(`\\b(?=[A-Za-z_$][A-Za-z0-9_$]*\\b)(?=[A-Za-z0-9_$]*(?:transcript|segment|directory|window|turn|ref|context))(?=[A-Za-z0-9_$]*${legacyGeneration})[A-Za-z0-9_$]+\\b`, "i");
const readSubject = "(?:transcript|segmented|segments?|directory|window|turns?|rows?|refs?|reads?|protocol|positions?)";
const lowerProse = new RegExp(`\\b${legacyGeneration}\\b(?!-)[^\\n]{0,80}\\b${readSubject}\\b|\\b${readSubject}\\b[^\\n]{0,80}\\b${legacyGeneration}\\b(?!-)`, "i");
const checksum = /(?:h1:|sha256[-:])[A-Za-z0-9+/=]+/g;

function hasRetiredIdentifier(source) {
  const text = source.replace(checksum, "");
  return retired.test(text) || lowerIdentifier.test(text);
}

function hasRetiredName(source) {
  return hasRetiredIdentifier(source) || lowerProse.test(source.replace(checksum, ""));
}

const mutationSource = await readFile(join(root, "testdata", "typescript", "segmented_read_name_gate.yaml"), "utf8");
const mutations = loadCorpus(mutationSource, {decodeInput: (value) => value, decodeExpected: (value) => value});
const mutationManifest = (await readFixture("segmented_read_name_gate_manifest.yaml")).requiredNames;
assert.deepEqual(mutations.cases.map((row) => row.name).sort(), [...mutationManifest].sort());

test("segmented read name gate distinguishes retired declarations from unrelated text", async (t) => {
  for (const row of mutations.cases.filter((entry) => entry.input.path === undefined)) await t.test(row.name, () => {
    assert.equal(hasRetiredName(row.input.source), row.expected);
  });
  for (const name of names) await t.test(`all generated forms of ${name}`, () => {
    for (const form of [name, `z${name}`, `Schema${name}`, `zSchema${name}`, `${name}Schema`]) {
      assert.equal(hasRetiredName(form), true, `${form} is not guarded`);
    }
  });
});

const sourceExtensions = new Set([".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".json", ".yaml", ".yml", ".html", ".md"]);
const excludedDirs = new Set([".git", "node_modules", "dist"]);
const excludedFiles = new Set(["go.sum", "pnpm-lock.yaml", "flake.nix"]);
// These exact case inputs intentionally contain the retired names being tested.
// Their other fields and every other fixture value still go through the scan.
const fixtureControlNames = new Set([
  "name-gate-retired-type",
  "name-gate-retired-zod",
  "name-gate-retired-schema",
  "name-gate-retired-schema-suffix",
  "name-gate-retired-function",
  "name-gate-old-lowercase-identifier",
  "name-gate-old-embedded-identifier",
  "name-gate-fixture-retired-declaration",
  "name-gate-fixture-positional-identifier",
  "name-gate-control-file-mutation",
  "name-gate-inventory-entry-excluded",
]);
const wiringFixtures = new Set([
  "openapi/testdata/typescript_catalog.yaml",
  "openapi/testdata/village_transcript_reads.yaml",
  "testdata/typescript/public_exports.yaml",
  "typescript/tests/fixtures/package-files.yaml",
]);

function hasRetiredNameAtPath(scanRoot, path, source) {
  const name = relative(scanRoot, path).split(sep).join("/");
  const fixture = name.split("/").some((part) => part === "testdata" || part === "fixtures");
  if (name === "testdata/typescript/retired_segmented_read_names.yaml" ||
      name === "testdata/typescript/retired_segmented_read_names_manifest.yaml") {
    const document = parse(source, {uniqueKeys: true});
    for (const field of ["names", "requiredNames"]) {
      if (Array.isArray(document[field])) document[field] = document[field].filter((value) => !required.includes(value));
    }
    source = JSON.stringify(document);
  } else if (name === "testdata/typescript/segmented_read_name_gate.yaml") {
    const document = parse(source, {uniqueKeys: true});
    for (const row of document.cases ?? []) {
      if (fixtureControlNames.has(row.name) && row.input) delete row.input.source;
    }
    source = JSON.stringify(document);
  }
  return fixture && !wiringFixtures.has(name) ? hasRetiredIdentifier(source) : hasRetiredName(source);
}

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

test("the source walk applies the fixture-specific identifier rule", async (t) => {
  for (const row of mutations.cases.filter((entry) => entry.input.path !== undefined)) await t.test(row.name, async () => {
    const temp = await mkdtemp(join(tmpdir(), "segmented-name-gate-"));
    try {
      const path = join(temp, row.input.path);
      await mkdir(dirname(path), {recursive: true});
      await writeFile(path, row.input.source);
      const paths = await collectSources(temp);
      assert.ok(paths.includes(path), `fixture was not scanned: ${row.input.path}`);
      assert.equal(hasRetiredNameAtPath(temp, path, await readFile(path, "utf8")), row.expected);
    } finally {
      await rm(temp, {recursive: true, force: true});
    }
  });
});

test("retired segmented read names do not appear on active source or generated surfaces", async () => {
  const paths = await collectSources(root);
  const findings = [];
  for (const path of paths) {
    const text = await readFile(path, "utf8");
    if (hasRetiredNameAtPath(root, path, text)) findings.push(relative(root, path));
  }
  assert.deepEqual(findings.sort(), [], "retired segmented read identifiers or prose remain");
});
