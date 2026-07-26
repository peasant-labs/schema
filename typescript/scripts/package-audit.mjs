import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { readFile } from "node:fs/promises";

import { parse } from "yaml";

const fixture = parse(await readFile(new URL("../tests/fixtures/package-files.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["files", "repository"]);
assert.ok(Array.isArray(fixture.files));
assert.deepEqual(fixture.repository, {
  type: "git",
  url: "https://github.com/peasant-labs/schema",
});

const packageManifest = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
assert.deepEqual(
  packageManifest.repository,
  fixture.repository,
  "typescript/package.json must declare the canonical peasant-labs/schema git repository metadata for npm provenance",
);

const packed = JSON.parse(execFileSync("pnpm", ["pack", "--json", "--dry-run"], { encoding: "utf8" }));
const actual = packed.files.map((entry) => entry.path).sort();
const expected = [...fixture.files].sort();
assert.deepEqual(actual, expected, "packed package contents differ from the audited fixture");
