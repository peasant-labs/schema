import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import openapiTS, { astToString } from "openapi-typescript";
import ts from "typescript";
import { parse } from "yaml";

const packageRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const moduleRoot = join(packageRoot, "..");
const generatedRoot = join(packageRoot, "src", "internal", "generated");

const versionsSource = await readFile(join(moduleRoot, "versions.go"), "utf8");
const metadataSource = await readFile(join(moduleRoot, "metadata.go"), "utf8");
const versions = {
  VillageAPIVersion: requiredMatch(versionsSource, /VillageAPIVersion\s*=\s*"([^"]+)"/, "VillageAPIVersion", "versions.go"),
  PeasantLocalAPIVersion: requiredMatch(versionsSource, /PeasantLocalAPIVersion\s*=\s*"([^"]+)"/, "PeasantLocalAPIVersion", "versions.go"),
  TypesVersion: requiredMatch(versionsSource, /TypesVersion\s*=\s*"([^"]+)"/, "TypesVersion", "versions.go"),
  MetadataSchemaVersion: Number(requiredMatch(metadataSource, /MetadataSchemaVersion\s*=\s*(\d+)/, "MetadataSchemaVersion", "metadata.go")),
};

const specPath = join(moduleRoot, "generated", `types-${versions.TypesVersion}.json`);
const spec = JSON.parse(await readFile(specPath, "utf8"));
const enumCatalog = parse(await readFile(join(moduleRoot, "testdata", "typescript", "enums.yaml"), "utf8"));
const qualitySource = parse(await readFile(join(moduleRoot, "testdata", "quality", "sessions.yaml"), "utf8"));
const timelineSource = parse(await readFile(join(moduleRoot, "testdata", "local-api", "timeline.yaml"), "utf8"));

await mkdir(generatedRoot, { recursive: true });
await writeFile(join(generatedRoot, "versions.gen.ts"), renderVersions(versions));
await writeFile(join(generatedRoot, "enums.gen.ts"), renderEnums(spec, enumCatalog));
await writeFile(join(generatedRoot, "public-contract.gen.ts"), await renderPublicContract(enumCatalog));
await writeFile(join(generatedRoot, "quality-fixtures.gen.ts"), renderQualityFixtures(qualitySource));
await writeFile(join(generatedRoot, "timeline-fixtures.gen.ts"), renderTimelineFixtures(timelineSource));
await generateOperationContracts("local", `peasantlocal-api-${versions.PeasantLocalAPIVersion}.json`, "local-api.ts");
await generateOperationContracts("village", `village-api-${versions.VillageAPIVersion}.json`, "village-api.ts");

function renderVersions(values) {
  return `${header()}export const VillageAPIVersion = ${JSON.stringify(values.VillageAPIVersion)} as const;\nexport const PeasantLocalAPIVersion = ${JSON.stringify(values.PeasantLocalAPIVersion)} as const;\nexport const TypesVersion = ${JSON.stringify(values.TypesVersion)} as const;\nexport const MetadataSchemaVersion = ${values.MetadataSchemaVersion} as const;\n`;
}

function renderEnums(openapi, catalog) {
  if (!Array.isArray(catalog?.enums)) {
    throw new Error("TypeScript enum generation could not load testdata/typescript/enums.yaml: the root enums sequence is missing; no runtime closed-set facade was generated; restore the canonical fixture shape.");
  }
  const schemas = openapi?.components?.schemas;
  if (typeof schemas !== "object" || schemas === null) {
    throw new Error(`TypeScript enum generation could not find components.schemas in ${specPath}; no contract definitions can be generated; regenerate the Types OpenAPI spec from Go.`);
  }

  const imports = [];
  const blocks = [];
  for (const enumCase of catalog.enums) {
    const schema = schemas[enumCase.name];
    const actualValues = schema?.enum;
    const members = enumCase.members;
    if (!Array.isArray(actualValues)) {
      throw new Error(`TypeScript enum generation found no OpenAPI enum for ${enumCase.name} in ${specPath}; the Go closed set would degrade to string in TypeScript; add JSONSchema enum metadata to the canonical Go type and regenerate.`);
    }
    if (!Array.isArray(members) || !Array.isArray(enumCase.all_values)) {
      throw new Error(`TypeScript enum generation found an invalid ${enumCase.name} entry in testdata/typescript/enums.yaml; members and all_values must both be sequences; fix the fixture before regenerating.`);
    }
    const expectedValues = members.map((member) => member.value);
    if (JSON.stringify(actualValues) !== JSON.stringify(expectedValues)) {
      throw new Error(`TypeScript enum generation found ${enumCase.name} drift between Go/OpenAPI (${JSON.stringify(actualValues)}) and testdata/typescript/enums.yaml (${JSON.stringify(expectedValues)}); TypeScript output was not written; update the canonical Go closed set and its fixture together.`);
    }
    const memberByValue = new Map(members.map((member) => [member.value, member.name]));
    const allMembers = enumCase.all_values.map((value) => {
      const name = memberByValue.get(value);
      if (name === undefined) {
        throw new Error(`TypeScript enum generation found ${JSON.stringify(value)} in ${enumCase.name}.all_values without a matching member; the runtime inventory cannot be rendered; add the member or remove the unsupported value.`);
      }
      return `${enumCase.name}.${name}`;
    });
    imports.push(`import { z${enumCase.name}, type ${enumCase.name} as ${enumCase.name}Contract } from "./contract/zod.gen.js";`);
    const renderedMembers = members.map((member) => `  ${member.name}: z${enumCase.name}.parse(${JSON.stringify(member.value)}),`).join("\n");
    blocks.push(`export type ${enumCase.name} = ${enumCase.name}Contract;\nexport const ${enumCase.name} = Object.freeze({\n${renderedMembers}\n} as const);\nexport const ${enumCase.all_name} = Object.freeze([${allMembers.join(", ")}]) as readonly ${enumCase.name}[];\nexport function is${enumCase.name}(value: unknown): value is ${enumCase.name} {\n  return z${enumCase.name}.safeParse(value).success;\n}`);
  }
  return `${header()}${imports.join("\n")}\n\n${blocks.join("\n\n")}\n`;
}

async function renderPublicContract(catalog) {
  const source = await readFile(join(generatedRoot, "contract", "zod.gen.ts"), "utf8");
  const schemaNames = [...source.matchAll(/^export const (z[A-Za-z0-9_]+)\s*=/gm)].map((match) => match[1]);
  const typeNames = [...source.matchAll(/^export type ([A-Za-z0-9_]+)\s*=/gm)].map((match) => match[1]);
  const enumNames = new Set(catalog.enums.map((enumCase) => enumCase.name));
  const contractTypes = typeNames.filter((name) => !enumNames.has(name));
  if (schemaNames.length === 0 || contractTypes.length === 0) {
    throw new Error("TypeScript contract generation could not discover Hey API Zod exports in src/internal/generated/contract/zod.gen.ts; the public contract facade would be empty; inspect the pinned Hey API output before updating the generator parser.");
  }
  return `${header()}export { ${schemaNames.join(", ")} } from "./contract/zod.gen.js";\nexport type { ${contractTypes.join(", ")} } from "./contract/zod.gen.js";\n`;
}

function renderQualityFixtures(source) {
  const fixtures = {
    sessions: source.quality_sessions,
    sets: source.quality_fixture_sets,
    variations: {
      outcomes: source.quality_variations?.outcomes,
      projects: source.quality_variations?.projects,
      scopes: source.quality_variations?.scopes,
      taskTitles: source.quality_variations?.task_titles,
      tokenRatios: source.quality_variations?.token_ratios,
      metrics: {
        retryLoops: source.quality_variations?.metrics?.retry_loops,
        signalDensity: source.quality_variations?.metrics?.signal_density,
        specQualityScore: source.quality_variations?.metrics?.spec_quality_score,
        filesTouched: source.quality_variations?.metrics?.files_touched,
        linesChanged: source.quality_variations?.metrics?.lines_changed,
      },
    },
  };
  return `${header()}import type { QualityFixtures } from "../../fixtures/quality.js";\n\nexport const canonicalQualityFixtures: QualityFixtures = ${JSON.stringify(fixtures, null, 2)};\n`;
}

function renderTimelineFixtures(source) {
  const fixtures = mapKeys(source, { error_contains: "errorContains" });
  for (const testCase of fixtures.cases ?? []) {
    for (const commit of testCase.input?.commits ?? []) {
      if (!("sessionIds" in commit)) commit.sessionIds = null;
    }
  }
  return `${header()}import type { TimelineFixtureCorpus } from "../../fixtures/timeline.js";\n\nexport const canonicalTimelineFixtures: TimelineFixtureCorpus = ${JSON.stringify(fixtures, null, 2)};\n`;
}

async function generateOperationContracts(surface, filename, outputName) {
  const apiPath = join(moduleRoot, "generated", filename);
  const apiSpec = JSON.parse(await readFile(apiPath, "utf8"));
  const aliases = canonicalOperationAliases(surface, apiSpec, spec);
  const projectHashPattern = "^[0-9a-f]{64}$";
  const ast = await openapiTS(apiSpec, {
    inject: 'import type * as Schema from "../../index.js";',
    silent: true,
    transform(schemaObject, metadata) {
      if (schemaObject.pattern === projectHashPattern) return schemaTypeReference("ProjectHash");
      const prefix = "#/components/schemas/";
      if (!metadata.path.startsWith(prefix) || metadata.path.slice(prefix.length).includes("/")) return;
      const rootName = aliases.get(metadata.path.slice(prefix.length));
      return rootName === undefined ? undefined : schemaTypeReference(rootName);
    },
  });
  await writeFile(join(generatedRoot, outputName), astToString(ast));
}

function canonicalOperationAliases(surface, apiSpec, rootSpec) {
  const apiSchemas = apiSpec?.components?.schemas;
  const rootSchemas = rootSpec?.components?.schemas;
  if (typeof apiSchemas !== "object" || apiSchemas === null || typeof rootSchemas !== "object" || rootSchemas === null) {
    throw new Error(`TypeScript ${surface} operation generation could not compare components.schemas with the canonical Types catalog; operation payload identities cannot be trusted; regenerate both OpenAPI documents from Go.`);
  }
  const rootCanonical = new Map(Object.entries(rootSchemas).map(([name, schema]) => [name, stableSchema("types", schema)]));
  const aliases = new Map();
  for (const [rawName, schema] of Object.entries(apiSchemas)) {
    const rootName = canonicalTypeScriptName(surface, rawName);
    const rootSchema = rootCanonical.get(rootName);
    if (rootSchema === undefined) continue;
    const apiSchema = stableSchema(surface, schema);
    if (apiSchema !== rootSchema) {
      throw new Error(`TypeScript ${surface} operation generation found API component ${rawName} normalizing to canonical root type ${rootName} with an unequal schema; aliasing it would misrepresent the HTTP contract; rename the operation-specific component or make it exactly equal to the canonical Go/OpenAPI definition.`);
    }
    aliases.set(rawName, rootName);
  }
  return aliases;
}

function canonicalTypeScriptName(surface, name) {
  const unprefixed = surface === "types" ? name : name.replace(/^Schema/, "");
  return unprefixed === "BestiaryHarness" || unprefixed === "Provider" ? "Harness" : unprefixed;
}

function stableSchema(surface, schema) {
  return JSON.stringify(normalizeSchemaRefs(surface, structuredClone(schema)));
}

function normalizeSchemaRefs(surface, value) {
  if (Array.isArray(value)) return value.map((item) => normalizeSchemaRefs(surface, item));
  if (typeof value !== "object" || value === null) return value;
  return Object.fromEntries(Object.entries(value).sort(([left], [right]) => left.localeCompare(right)).map(([key, item]) => {
    if (key === "$ref" && typeof item === "string") {
      const prefix = "#/components/schemas/";
      return [key, item.startsWith(prefix) ? prefix + canonicalTypeScriptName(surface, item.slice(prefix.length)) : item];
    }
    return [key, normalizeSchemaRefs(surface, item)];
  }));
}

function schemaTypeReference(name) {
  return ts.factory.createTypeReferenceNode(
    ts.factory.createQualifiedName(ts.factory.createIdentifier("Schema"), ts.factory.createIdentifier(name)),
  );
}

function mapKeys(value, mapping) {
  if (Array.isArray(value)) return value.map((item) => mapKeys(item, mapping));
  if (typeof value !== "object" || value === null) return value;
  return Object.fromEntries(Object.entries(value).map(([key, item]) => [mapping[key] ?? key, mapKeys(item, mapping)]));
}

function requiredMatch(source, pattern, field, location) {
  const match = pattern.exec(source);
  if (match?.[1] !== undefined) return match[1];
  throw new Error(`TypeScript contract generation could not read ${field} from ${location}; generated version metadata would drift from Go; keep the canonical Go constant in the expected assignment form or update typescript/scripts/generate-contract-support.mjs.`);
}

function header() {
  return "// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.\n";
}
