import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
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
  const packed = JSON.parse(execFileSync("pnpm", ["pack", "--json", "--pack-destination", temp], { encoding: "utf8" }));
  const tarball = packed.filename;
  await writeFile(join(temp, "package.json"), JSON.stringify({ private: true, type: "module", packageManager: "pnpm@11.5.3" }));
  const storeDir = process.env.PNPM_STORE_DIR ?? execFileSync("pnpm", ["store", "path", "--silent"], { encoding: "utf8" }).trim();
  execFileSync("pnpm", ["add", "--offline", "--ignore-scripts", "--store-dir", storeDir, tarball], {
    cwd: temp,
    stdio: "inherit",
  });
  const probe = fixture.subpaths.map((subpath) => `await import(${JSON.stringify(subpath)});`).join("\n");
  await writeFile(join(temp, "probe.mjs"), `${probe}\n`);
  execFileSync(process.execPath, [join(temp, "probe.mjs")], { cwd: temp, stdio: "inherit" });

  const tsconfig = {
    compilerOptions: { strict: true, noEmit: true, target: "ES2022", module: "NodeNext", moduleResolution: "NodeNext" },
    include: ["consumer.ts"],
  };
  await writeFile(join(temp, "tsconfig.json"), JSON.stringify(tsconfig));
  await writeFile(join(temp, "consumer.ts"), await readFile(new URL("../tests/fixtures/tarball-consumer.ts", import.meta.url), "utf8"));
  const tsc = new URL("../node_modules/.bin/tsc", import.meta.url).pathname;
  execFileSync(tsc, ["--project", join(temp, "tsconfig.json")], { cwd: temp, stdio: "inherit" });

  await writeFile(join(temp, "consumer.ts"), await readFile(new URL("../tests/fixtures/invalid-enum-consumer.ts", import.meta.url), "utf8"));
  const invalid = spawnSync(tsc, ["--project", join(temp, "tsconfig.json")], { cwd: temp, encoding: "utf8" });
  assert.notEqual(invalid.status, 0, "invalid enum sentinel unexpectedly compiled");
  assert.match(`${invalid.stdout}\n${invalid.stderr}`, /unknown-role|not assignable/, "invalid enum failure did not explain the rejected value");
} finally {
  await rm(temp, { recursive: true, force: true });
}
