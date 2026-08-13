import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import * as root from "../dist/index.js";
import * as localApi from "../dist/local-api.js";
import * as villageApi from "../dist/village-api.js";
import { loadCorpus } from "../dist/testcase.js";

// This suite is the facade-level counterpart to the Local API operation surface
// proof in project-hash-locations.test.mjs: it guards the hand-maintained root
// (index.ts), /local-api, and /village-api export wiring itself, not the
// generated contract those facades re-export. Losing one of these re-export
// lines is invisible to typecheck/build (a dropped `export` is not a type
// error) and invisible to a non-empty package listing, so this is the only
// mechanism that would fail if the wiring silently narrowed.

const fixtureSource = await readFile(new URL("../../testdata/typescript/public_exports.yaml", import.meta.url), "utf8");
const fixture = parse(fixtureSource);
assert.deepEqual(Object.keys(fixture).sort(), ["aliases", "constants", "forbidden", "functions"]);

const typesVersion = fixture.constants.find((constant) => constant.name === "TypesVersion")?.value;
assert.equal(typeof typesVersion, "string", "public_exports.yaml must declare the TypesVersion string used to locate the canonical OpenAPI catalog");
const specSource = await readFile(new URL(`../../generated/types-${typesVersion}.json`, import.meta.url), "utf8");
const spec = JSON.parse(specSource);
const catalogNames = Object.keys(spec.components?.schemas ?? {});
assert.ok(catalogNames.length > 0, "the Types OpenAPI catalog backing the canonical type entries is empty");

const enumCatalogSource = await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8");
const enumCatalog = parse(enumCatalogSource);
const enumNames = enumCatalog.enums.map((enumCase) => enumCase.name);
assert.ok(enumNames.length > 0, "the closed-enum catalog backing the canonical value entries is empty");

// The known content-capability inventory is a runtime facade that is NOT an
// enum (the discovery wire alias is an open string), so it is not in the enum
// catalog. Fold its three runtime exports into the canonical set the same way,
// so the exact-match assertion below still guards them.
const contentCapabilityCatalog = parse(await readFile(new URL("../../testdata/typescript/content_capabilities.yaml", import.meta.url), "utf8"));
const knownContentCapability = contentCapabilityCatalog.known;
assert.equal(typeof knownContentCapability, "object", "content_capabilities.yaml must provide a known inventory mapping");
const contentCapabilityRuntimeNames = [knownContentCapability.name, knownContentCapability.all_name, knownContentCapability.guard];
for (const name of contentCapabilityRuntimeNames) {
  assert.equal(typeof name, "string", "content_capabilities.yaml known inventory must name its facade const, inventory, and guard");
}

const indexSource = await readFile(new URL("../src/index.ts", import.meta.url), "utf8");
const localApiSource = await readFile(new URL("../src/local-api.ts", import.meta.url), "utf8");
const villageApiSource = await readFile(new URL("../src/village-api.ts", import.meta.url), "utf8");

const canonicalTypeEntries = toEntries(catalogNames.filter((name) => !enumNames.includes(name)));
const canonicalRuntimeEntries = [
  ...toEntries(catalogNames.map((name) => `z${name}`)),
  ...toEntries(enumCatalog.enums.flatMap((enumCase) => [enumCase.name, enumCase.all_name, `is${enumCase.name}`])),
  ...toEntries(contentCapabilityRuntimeNames),
  ...toEntries(fixture.functions.map((fn) => fn.name)),
  ...fixture.constants.map((constant) => ({ name: constant.name, target: String(constant.value) })),
];
const canonicalRuntimeNames = canonicalRuntimeEntries.map((entry) => entry.name).sort();

test("root, local-api, and village-api runtime exports match the strict shared fixture", async (t) => {
  for (const constant of fixture.constants) {
    await t.test(`constant ${constant.name}`, () => {
      assert.equal(String(root[constant.name]), String(constant.value));
      for (const surface of constant.surfaces) {
        if (surface === "root") continue;
        const facade = surfaceModule(surface);
        assert.equal(Object.hasOwn(facade, constant.name), true, `${constant.name} is not re-exported from the /${surface} facade`);
        assert.equal(String(facade[constant.name]), String(constant.value), `${constant.name} re-exported from /${surface} does not match the root value`);
      }
    });
  }

  for (const fn of fixture.functions) {
    await t.test(`function ${fn.name}`, () => {
      assert.equal(typeof root[fn.name], "function", `${fn.name} is not exported as a function from the root facade`);
    });
  }

  await t.test("forbidden identifiers never ship in the root facade", () => {
    for (const forbidden of fixture.forbidden) {
      assert.equal(Object.hasOwn(root, forbidden), false, `${forbidden} unexpectedly exported as a runtime value from the root facade`);
      assert.equal(Object.hasOwn(root, `z${forbidden}`), false, `z${forbidden} unexpectedly exported as a runtime schema from the root facade`);
    }
  });
});

test("root, local-api, and village-api type-only aliases are wired in source", async (t) => {
  for (const alias of fixture.aliases) {
    await t.test(`type-only alias ${alias.name}`, () => {
      for (const surface of alias.surfaces) {
        assertTypeAliasWiring(surfaceSource(surface), alias.name, alias.target, `typescript/src/${surface === "root" ? "index" : surface + "-api"}.ts`);
      }
    });
  }
});

test("root runtime facade exactly matches the independent contract catalogs", () => {
  assert.equal(new Set(canonicalRuntimeNames).size, canonicalRuntimeNames.length, "catalog-derived root runtime export names must not collide");
  assert.deepEqual(Object.keys(root).sort(), canonicalRuntimeNames);
});

test("root public export catalog rejects the strict shared mutation corpus", async (t) => {
  const mutationSource = await readFile(new URL("../../testdata/typescript/public_export_mutations.yaml", import.meta.url), "utf8");
  const mutationFixture = loadCorpus(mutationSource, {
    decodeInput(value, path) {
      if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
      const allowedKeys = ["kind", "namespace", "name", "target"];
      for (const key of Object.keys(value)) {
        if (!allowedKeys.includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
      }
      if (typeof value.kind !== "string") throw new TypeError(`${path}.kind: must be a string`);
      if (typeof value.namespace !== "string") throw new TypeError(`${path}.namespace: must be a string`);
      if (typeof value.name !== "string") throw new TypeError(`${path}.name: must be a string`);
      if (value.target !== undefined && typeof value.target !== "string") throw new TypeError(`${path}.target: must be a string when present`);
      return { kind: value.kind, namespace: value.namespace, name: value.name, target: value.target };
    },
    decodeExpected(value, path) {
      if (typeof value !== "boolean") throw new TypeError(`${path}: must be a boolean`);
      return value;
    },
  });
  assert.equal(mutationFixture.cases.length, 6);
  assert.equal(new Set(mutationFixture.cases.map((testCase) => testCase.name)).size, mutationFixture.cases.length);

  for (const testCase of mutationFixture.cases) {
    await t.test(testCase.name, () => {
      const canonical = testCase.input.namespace === "type" ? canonicalTypeEntries : canonicalRuntimeEntries;
      const mutated = applyMutation(canonical, testCase.input);
      assert.equal(validateExportEntries(mutated, canonical), testCase.expected);
    });
  }
});

function toEntries(names) {
  return names.map((name) => ({ name, target: name }));
}

function surfaceModule(surface) {
  switch (surface) {
    case "root": return root;
    case "local": return localApi;
    case "village": return villageApi;
    default: throw new Error(`public_exports.yaml named an unknown surface ${JSON.stringify(surface)}; only root/local/village are wired facades`);
  }
}

function surfaceSource(surface) {
  switch (surface) {
    case "root": return indexSource;
    case "local": return localApiSource;
    case "village": return villageApiSource;
    default: throw new Error(`public_exports.yaml named an unknown surface ${JSON.stringify(surface)}; only root/local/village are wired facades`);
  }
}

function assertTypeAliasWiring(source, name, target, location) {
  const match = source.match(new RegExp(`^export type ${name}\\s*=\\s*(.+);$`, "m"));
  assert.notEqual(match, null, `${location}: no "export type ${name} = ...;" statement found`);
  assert.match(match[1], new RegExp(`\\b${target}\\b`), `${location}: "export type ${name}" does not reference ${target}`);
}

function applyMutation(canonicalEntries, mutation) {
  const entries = canonicalEntries.map((entry) => ({ ...entry }));
  switch (mutation.kind) {
    case "remove":
      return entries.filter((entry) => entry.name !== mutation.name);
    case "add":
      return [...entries, { name: mutation.name, target: mutation.target ?? mutation.name }];
    case "duplicate": {
      const existing = entries.find((entry) => entry.name === mutation.name);
      assert.notEqual(existing, undefined, `public_export_mutations.yaml duplicate mutation references unknown canonical name ${mutation.name}`);
      return [...entries, { ...existing }];
    }
    case "redirect":
      return entries.map((entry) => (entry.name === mutation.name ? { ...entry, target: mutation.target } : entry));
    default:
      throw new Error(`public_export_mutations.yaml used an unknown mutation kind ${JSON.stringify(mutation.kind)}; only remove/add/duplicate/redirect are supported`);
  }
}

// validateExportEntries is the mutation-proof oracle: it fails closed on a
// missing name, an added name, a duplicated name, or a name redirected to the
// wrong target, mirroring the exact-identity contract the deleted
// cmd/schema-gen/typescript_exports_test.go enforced for the bespoke
// generator. actualEntries may contain duplicates or drift from
// canonicalEntries by construction (mutation fixtures build both); a
// well-formed catalog is the only shape that returns true.
function validateExportEntries(actualEntries, canonicalEntries) {
  const canonical = new Map(canonicalEntries.map((entry) => [entry.name, entry.target]));
  const seen = new Map();
  for (const entry of actualEntries) {
    if (seen.has(entry.name)) return false;
    seen.set(entry.name, entry.target);
  }
  if (seen.size !== canonical.size) return false;
  for (const [name, target] of canonical) {
    if (seen.get(name) !== target) return false;
  }
  return true;
}
