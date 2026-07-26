import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { parse } from "yaml";

import { CANONICAL_REPOSITORY, assertPackageRepositoryMetadata } from "./lib/package-repository-metadata.mjs";

const fixture = parse(await readFile(new URL("../tests/fixtures/package-files.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["files", "repository"]);
assert.ok(Array.isArray(fixture.files));
assert.deepEqual(fixture.repository, CANONICAL_REPOSITORY);

const temp = await mkdtemp(join(tmpdir(), "peasant-labs-schema-package-audit-"));
try {
  const packed = JSON.parse(execFileSync("pnpm", ["pack", "--json", "--pack-destination", temp], { encoding: "utf8" }));
  assert.equal(typeof packed.filename, "string", "pnpm pack did not report a package tarball filename");
  assert.ok(Array.isArray(packed.files), "pnpm pack did not report packed file paths");
  const actual = packed.files.map((entry) => entry.path).sort();
  const expected = [...fixture.files].sort();
  assert.deepEqual(actual, expected, "packed package contents differ from the audited fixture");

  const packedManifest = JSON.parse(execFileSync("tar", ["-xOzf", packed.filename, "package/package.json"], { encoding: "utf8" }));
  assertPackageRepositoryMetadata(packedManifest, fixture.repository);
} finally {
  await rm(temp, { recursive: true, force: true });
}
