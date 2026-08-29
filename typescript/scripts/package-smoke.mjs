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
const publicExports = parse(await readFile(new URL("../../testdata/typescript/public_exports.yaml", import.meta.url), "utf8"));
const typesVersion = publicExports.constants?.find((constant) => constant.name === "TypesVersion")?.value;
assert.equal(typeof typesVersion, "string", "public_exports.yaml must declare the TypesVersion string used to locate the canonical OpenAPI catalog");
const spec = JSON.parse(await readFile(new URL(`../../generated/types-${typesVersion}.json`, import.meta.url), "utf8"));
const enumCatalog = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));
assert.ok(Array.isArray(enumCatalog.enums) && enumCatalog.enums.length > 0, "enums.yaml must provide at least one enum facade for the packed-package runtime probe");
const villageCollectivesFixture = parse(await readFile(new URL("../../openapi/testdata/village_collectives_operations.yaml", import.meta.url), "utf8"));
assert.ok(Array.isArray(villageCollectivesFixture.operations) && villageCollectivesFixture.operations.length > 0, "village_collectives_operations.yaml must provide operation IDs for the packed-package type probe");
const villageCollectiveOperationIDs = villageCollectivesFixture.operations.map((operation) => {
  assert.equal(typeof operation.operation_id, "string", "each Village collectives operation fixture must declare an operation_id");
  assert.match(operation.operation_id, /^[A-Za-z_$][A-Za-z0-9_$]*$/, "Village collectives operation IDs must be TypeScript property identifiers");
  return operation.operation_id;
});
assert.equal(new Set(villageCollectiveOperationIDs).size, villageCollectiveOperationIDs.length, "Village collectives operation IDs must be unique");
const enumName = enumCatalog.enums[0].name;
assert.match(enumName, /^[A-Za-z_$][A-Za-z0-9_$]*$/, "enums.yaml representative enum name must be a JavaScript identifier");
const enumNames = new Set(enumCatalog.enums.map((enumCase) => enumCase.name));
const schemaName = Object.keys(spec.components?.schemas ?? {}).find((name) => !enumNames.has(name));
assert.notEqual(schemaName, undefined, "the Types OpenAPI catalog must provide a non-enum component for the packed-package Zod runtime probe");
assert.match(schemaName, /^[A-Za-z_$][A-Za-z0-9_$]*$/, "Types OpenAPI representative component name must be a JavaScript identifier");
const schemaExportName = `z${schemaName}`;

const temp = await mkdtemp(join(tmpdir(), "peasant-labs-schema-package-"));
try {
  const packed = JSON.parse(execFileSync("pnpm", ["pack", "--json", "--pack-destination", temp], { encoding: "utf8" }));
  const tarball = packed.filename;
  const packageRoot = fileURLToPath(new URL("..", import.meta.url));
  const packageManifest = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
  const packageLock = parse(await readFile(new URL("../pnpm-lock.yaml", import.meta.url), "utf8"));
  const compilerSpecifier = `typescript@${packageManifest.devDependencies.typescript}`;
  const lockedRuntimeSpecifiers = Object.keys(packageManifest.dependencies ?? {})
    .sort()
    .map((name) => {
      const resolved = packageLock.importers?.["."]?.dependencies?.[name]?.version;
      assert.equal(typeof resolved, "string", `pnpm-lock.yaml must pin the runtime dependency ${name}`);
      return `${name}@${resolved.replace(/\(.+$/, "")}`;
    });
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
  execFileSync("pnpm", ["add", "--offline", "--ignore-scripts", "--store-dir", storeDir, compilerSpecifier, ...lockedRuntimeSpecifiers, tarball], {
    cwd: consumerDir,
    stdio: "inherit",
  });

  const probe = [
    `import { ${enumName}, ${schemaExportName} } from ${JSON.stringify(packageManifest.name)};`,
    ...fixture.subpaths.map((subpath) => `await import(${JSON.stringify(subpath)});`),
    `if (typeof ${enumName} !== "object" || ${enumName} === null) throw new TypeError(${JSON.stringify(`packed ${packageManifest.name} export ${enumName} is not a runtime enum facade`)});`,
    `if (typeof ${schemaExportName}.safeParse !== "function") throw new TypeError(${JSON.stringify(`packed ${packageManifest.name} export ${schemaExportName} is not a Zod schema`)});`,
  ].join("\n");
  await writeFile(join(consumerDir, "probe.mjs"), `${probe}\n`);
  execFileSync(process.execPath, [join(consumerDir, "probe.mjs")], { cwd: consumerDir, stdio: "inherit" });

  const operationTypeChecks = villageCollectiveOperationIDs
    .map((operationID) => `type __VillageCollectivesOperation_${operationID} = VillageOperations[${JSON.stringify(operationID)}];`)
    .join("\n");
  await writeFile(join(consumerDir, "tsconfig.json"), JSON.stringify(tsconfig));
  await writeFile(join(consumerDir, "consumer.ts"), `${await readFile(new URL("../tests/fixtures/tarball-consumer.ts", import.meta.url), "utf8")}\n${operationTypeChecks}\n`);
  const tsc = join(consumerDir, "node_modules", ".bin", "tsc");
  execFileSync(tsc, ["--project", join(consumerDir, "tsconfig.json")], { cwd: consumerDir, stdio: "inherit" });

  await writeFile(join(consumerDir, "consumer.ts"), await readFile(new URL("../tests/fixtures/invalid-enum-consumer.ts", import.meta.url), "utf8"));
  const invalid = spawnSync(tsc, ["--project", join(consumerDir, "tsconfig.json")], { cwd: consumerDir, encoding: "utf8" });
  assert.notEqual(invalid.status, 0, "invalid enum sentinel unexpectedly compiled");
  assert.match(`${invalid.stdout}\n${invalid.stderr}`, /unknown-role|not assignable/, "invalid enum failure did not explain the rejected value");

  await writeFile(join(consumerDir, "consumer.ts"), await readFile(new URL("../tests/fixtures/invalid-project-hash-consumer.ts", import.meta.url), "utf8"));
  const invalidProjectHash = spawnSync(tsc, ["--project", join(consumerDir, "tsconfig.json")], { cwd: consumerDir, encoding: "utf8" });
  assert.notEqual(invalidProjectHash.status, 0, "plain string unexpectedly compiled as ProjectHash");
  assert.match(`${invalidProjectHash.stdout}\n${invalidProjectHash.stderr}`, /ProjectHash|not assignable/, "ProjectHash brand failure did not explain the rejected value");
} finally {
  await rm(temp, { recursive: true, force: true });
}
