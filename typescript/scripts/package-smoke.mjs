import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

import { parse } from "yaml";

const fixture = parse(await readFile(new URL("../tests/fixtures/public-subpaths.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["subpaths"]);
assert.ok(Array.isArray(fixture.subpaths));

const temp = await mkdtemp(join(tmpdir(), "peasant-labs-schema-package-"));
try {
  const packed = JSON.parse(execFileSync("pnpm", ["pack", "--json", "--pack-destination", temp], { encoding: "utf8" }));
  const tarball = packed.filename;
  const packageRoot = fileURLToPath(new URL("..", import.meta.url));
  const packageManifest = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
  const compilerSpecifier = `typescript@${packageManifest.devDependencies.typescript}`;
  const tsconfig = {
    compilerOptions: { strict: true, noEmit: true, target: "ES2022", module: "NodeNext", moduleResolution: "NodeNext" },
    include: ["consumer.ts"],
  };

  const storeDir = join(temp, "store");
  const modulesDir = join(temp, "modules");
  const virtualStoreDir = join(temp, "virtual-store");
  const consumerDir = join(temp, "consumer");
  await mkdir(consumerDir, { recursive: true });

  execFileSync("pnpm", ["fetch", "--frozen-lockfile", "--ignore-scripts", "--store-dir", storeDir, "--modules-dir", modulesDir, "--virtual-store-dir", virtualStoreDir], {
    cwd: packageRoot,
    stdio: "inherit",
  });
  await writeFile(join(consumerDir, "package.json"), JSON.stringify({ private: true, type: "module", packageManager: "pnpm@11.5.3" }));
  execFileSync("pnpm", ["add", "--offline", "--ignore-scripts", "--store-dir", storeDir, compilerSpecifier, tarball], {
    cwd: consumerDir,
    stdio: "inherit",
  });

  const probe = fixture.subpaths.map((subpath) => `await import(${JSON.stringify(subpath)});`).join("\n");
  await writeFile(join(consumerDir, "probe.mjs"), `${probe}\n`);
  execFileSync(process.execPath, [join(consumerDir, "probe.mjs")], { cwd: consumerDir, stdio: "inherit" });

  await writeFile(join(consumerDir, "tsconfig.json"), JSON.stringify(tsconfig));
  await writeFile(join(consumerDir, "consumer.ts"), await readFile(new URL("../tests/fixtures/tarball-consumer.ts", import.meta.url), "utf8"));
  const tsc = join(consumerDir, "node_modules", ".bin", "tsc");
  execFileSync(tsc, ["--project", join(consumerDir, "tsconfig.json")], { cwd: consumerDir, stdio: "inherit" });

  await writeFile(join(consumerDir, "consumer.ts"), await readFile(new URL("../tests/fixtures/invalid-enum-consumer.ts", import.meta.url), "utf8"));
  const invalid = spawnSync(tsc, ["--project", join(consumerDir, "tsconfig.json")], { cwd: consumerDir, encoding: "utf8" });
  assert.notEqual(invalid.status, 0, "invalid enum sentinel unexpectedly compiled");
  assert.match(`${invalid.stdout}\n${invalid.stderr}`, /unknown-role|not assignable/, "invalid enum failure did not explain the rejected value");
} finally {
  await rm(temp, { recursive: true, force: true });
}
