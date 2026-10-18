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
const retainedCollectiveAliases = parse(await readFile(new URL("../../testdata/typescript/collective_compatibility.yaml", import.meta.url), "utf8"));
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
    include: ["consumer.ts", "sync-pages.ts"],
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
  await writeFile(join(consumerDir, "package.json"), JSON.stringify({ private: true, type: "module", packageManager: "pnpm@11.24.0" }));
  execFileSync("pnpm", ["add", "--offline", "--ignore-scripts", "--store-dir", storeDir, compilerSpecifier, ...lockedRuntimeSpecifiers, tarball], {
    cwd: consumerDir,
    stdio: "inherit",
  });

  const probe = [
    `const envelopeAPI = await import(${JSON.stringify(packageManifest.name)});`,
    ...parse(await readFile(new URL("../../testdata/deprecated_grouped_contributable.yaml", import.meta.url), "utf8")).envelopes.cases.map(row => `if (envelopeAPI[${JSON.stringify(`z${row.input.schema}`)}].safeParse(${JSON.stringify(row.input.payload)}).success !== ${row.expected.valid}) throw new Error(${JSON.stringify(`packed grouped contribution envelope differs for ${row.name}`)});`),
    `const redaction = await import(${JSON.stringify(packageManifest.name)});`,
    `for (const [category, label] of ${JSON.stringify(parse(await readFile(new URL("../../testdata/typescript/redaction_categories.yaml", import.meta.url), "utf8")).pairs.cases.map(row => [row.input.category, row.input.label]))}) { if (redaction.redactionCategoryLabel(category) !== label || redaction.redactionCategoryFromLabel(label) !== category) throw new Error("packed redaction conversion differs"); }`,
    `import { ${enumName}, ${schemaExportName}, parseSessionDetailPayloadText, requiredContentCapabilities, zPublicRevisionRef, zSourceEntryRef, zSubmissionRef } from ${JSON.stringify(packageManifest.name)};`,
    `import { sessionGraphCapabilityFixtures } from ${JSON.stringify(packageManifest.name + "/fixtures/session-graph-capability")};`,
    `import * as publicSchema from ${JSON.stringify(packageManifest.name)};`,
    ...retainedCollectiveAliases.aliases.map(row => `if (publicSchema.z${row.name} !== publicSchema.z${row.canonical}) throw new Error("packed legacy validator ${row.name} lost canonical identity");`),
    ...retainedCollectiveAliases.aliases.filter(row => row.all_name !== undefined).map(row => `if (publicSchema.${row.name} !== publicSchema.${row.canonical} || publicSchema.${row.all_name} !== publicSchema.${row.canonical_all_name} || publicSchema.is${row.name} !== publicSchema.is${row.canonical}) throw new Error("packed legacy enum ${row.name} changed");`),
    `import { sessionGraphFixtures } from ${JSON.stringify(packageManifest.name + "/fixtures/session-graph")};`,
    ...fixture.subpaths.map((subpath) => `await import(${JSON.stringify(subpath)});`),
    `if (typeof ${enumName} !== "object" || ${enumName} === null) throw new TypeError(${JSON.stringify(`packed ${packageManifest.name} export ${enumName} is not a runtime enum facade`)});`,
    `if (typeof ${schemaExportName}.safeParse !== "function") throw new TypeError(${JSON.stringify(`packed ${packageManifest.name} export ${schemaExportName} is not a Zod schema`)});`,
    `for (const validator of [zPublicRevisionRef, zSourceEntryRef, zSubmissionRef]) { if (!validator.safeParse("界".repeat(32)).success || validator.safeParse("界".repeat(32) + "a").success) throw new TypeError("packed public reference validator does not enforce the 96-byte UTF-8 boundary"); }`,
    `const countCase = sessionGraphCapabilityFixtures().derivation.cases.find((row) => row.name === "count-only-zero"); if (countCase === undefined || requiredContentCapabilities(parseSessionDetailPayloadText(countCase.input.detailJSON))[0] !== "session_graph_provenance_v1") throw new TypeError("packed capability fixtures and runtime derivation are incoherent");`,
    `for (const row of sessionGraphFixtures().grouped_read.cases.cases) { const result = publicSchema["z" + row.input.schema].safeParse(row.input.payload); if (result.success !== row.expected.valid) throw new TypeError("packed grouped read validator differs for " + row.name + ": " + result.error); }`,
    `for (const row of sessionGraphFixtures().raw_durable.cases) {
      const detail = JSON.parse(row.input.rawJSON);
      const calls = [
        () => publicSchema.parseSessionDetailPayloadText(row.input.rawJSON),
        () => publicSchema.parseTranscriptContentText(JSON.stringify({kind: "session_detail", contractVersion: "1", sessionDetail: detail})).sessionDetail,
        () => publicSchema.parseServerMessageRaw(JSON.stringify({type: "session_detail", data: detail})).data,
      ];
      const results = [];
      for (const [index, call] of calls.entries()) {
        let value, error; try { value = call(); } catch (caught) { error = caught; }
        const valid = row.classification === "must-pass" || (index === 2 && row.name === "raw-forbidden-relationship-navigation");
        if ((error === undefined) !== valid) throw new TypeError("packed raw parser boundary differs for " + row.name + " at exit " + index + ": " + error);
        results.push(value);
      }
      if (row.classification === "must-pass" && results.some(result => JSON.stringify(result) !== JSON.stringify(results[0]))) throw new TypeError("packed parsers return different normalized evidence for " + row.name);
    }`,
  ].join("\n");
  await writeFile(join(consumerDir, "probe.mjs"), `${probe}\n`);
  execFileSync(process.execPath, [join(consumerDir, "probe.mjs")], { cwd: consumerDir, stdio: "inherit" });

  const operationTypeChecks = villageCollectiveOperationIDs
    .map((operationID) => `type __VillageCollectivesOperation_${operationID} = VillageOperations[${JSON.stringify(operationID)}];`)
    .join("\n");
  await writeFile(join(consumerDir, "tsconfig.json"), JSON.stringify(tsconfig));
  await writeFile(join(consumerDir, "sync-pages.ts"), await readFile(new URL("../tests/sync-pages.ts", import.meta.url), "utf8"));
  const legacyTypeChecks = [
    `type Same<A,B> = (<T>()=>T extends A ? 1 : 2) extends (<T>()=>T extends B ? 1 : 2) ? true : false; type MustBeTrue<T extends true> = T;`,
    ...retainedCollectiveAliases.aliases.map(row => `import type { ${row.name} as Old${row.name}, ${row.canonical} as New${row.name} } from ${JSON.stringify(packageManifest.name)}; type Check${row.name} = MustBeTrue<Same<Old${row.name},New${row.name}>>;`),
  ].join("\n");
  await writeFile(join(consumerDir, "consumer.ts"), `${await readFile(new URL("../tests/fixtures/tarball-consumer.ts", import.meta.url), "utf8")}\n${operationTypeChecks}\n${legacyTypeChecks}\n${await readFile(new URL("../tests/redaction-categories.ts", import.meta.url), "utf8")}\n`);
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
