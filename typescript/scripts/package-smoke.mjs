import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { pathToFileURL } from "node:url";

import { parse } from "yaml";

const fixture = parse(await readFile(new URL("../tests/fixtures/public-subpaths.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["subpaths"]);
assert.ok(Array.isArray(fixture.subpaths));

const temp = await mkdtemp(join(tmpdir(), "peasant-labs-schema-package-"));
try {
  const packed = JSON.parse(execFileSync("npm", ["pack", "--json", "--pack-destination", temp], { encoding: "utf8" }));
  assert.equal(packed.length, 1);
  const tarball = join(temp, packed[0].filename);
  await writeFile(join(temp, "package.json"), JSON.stringify({ private: true, type: "module" }));
  execFileSync("npm", ["install", "--offline", "--ignore-scripts", "--no-audit", "--no-fund", tarball], {
    cwd: temp,
    stdio: "inherit",
  });
  const probe = fixture.subpaths.map((subpath) => `await import(${JSON.stringify(subpath)});`).join("\n");
  await writeFile(join(temp, "probe.mjs"), `${probe}\n`);
  execFileSync(process.execPath, [join(temp, "probe.mjs")], { cwd: temp, stdio: "inherit" });
} finally {
  await rm(temp, { recursive: true, force: true });
}
