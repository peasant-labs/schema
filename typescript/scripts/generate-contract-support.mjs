import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import openapiTS, { astToString } from "openapi-typescript";
import ts from "typescript";
import { parse } from "yaml";

import { canonicalOperationAliases } from "./lib/operation-aliases.mjs";

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
const testcaseSource = await readFile(join(moduleRoot, "testcase", "testcase.go"), "utf8");

await mkdir(generatedRoot, { recursive: true });
await writeFile(join(generatedRoot, "versions.gen.ts"), renderVersions(versions));
await writeFile(join(generatedRoot, "enums.gen.ts"), renderEnums(spec, enumCatalog));
await writeFile(join(generatedRoot, "public-contract.gen.ts"), await renderPublicContract(enumCatalog));
await writeFile(join(generatedRoot, "quality-fixtures.gen.ts"), renderQualityFixtures(qualitySource));
await writeFile(join(generatedRoot, "timeline-fixtures.gen.ts"), renderTimelineFixtures(timelineSource));
await writeFile(join(generatedRoot, "testcase.gen.ts"), renderTestcaseModel(testcaseSource));
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
    // Every other generated All* export is the complete member list for its
    // enum; when all_values is a proper subset of members (today, only
    // Harness: AllHarnesses is the ingestion-supported subset, mirroring Go's
    // own doc comment on the same asymmetry in types.go), say so here too, so
    // a TypeScript-only reader does not assume All* is always exhaustive.
    const allDoc = allMembers.length < members.length
      ? `/**\n * ${enumCase.all_name} is the ingestion-supported subset of ${enumCase.name}, not the full\n * canonical set (mirrors types.go's AllHarnesses doc comment). Every ${enumCase.name}\n * member remains individually valid and accepted by is${enumCase.name}.\n */\n`
      : "";
    blocks.push(`export type ${enumCase.name} = ${enumCase.name}Contract;\nexport const ${enumCase.name} = Object.freeze({\n${renderedMembers}\n} as const);\n${allDoc}export const ${enumCase.all_name} = Object.freeze([${allMembers.join(", ")}]) as readonly ${enumCase.name}[];\nexport function is${enumCase.name}(value: unknown): value is ${enumCase.name} {\n  return z${enumCase.name}.safeParse(value).success;\n}`);
  }
  return `${header()}${imports.join("\n")}\n\n${blocks.join("\n\n")}\n`;
}

async function renderPublicContract(catalog) {
  const source = await readFile(join(generatedRoot, "contract", "zod.gen.ts"), "utf8");
  const schemaNames = [...source.matchAll(/^export const (z[A-Za-z0-9_]+)\s*=/gm)].map((match) => match[1]);
  const typeNames = [...source.matchAll(/^export type ([A-Za-z0-9_]+)\s*=/gm)].map((match) => match[1]);
  const enumNames = new Set(catalog.enums.map((enumCase) => enumCase.name));
  const contractTypes = typeNames.filter((name) => !enumNames.has(name));

  // The Types OpenAPI catalog (not the just-generated zod.gen.ts text) is the
  // authoritative expected export set: every catalog component must produce
  // exactly one zSchema const and, for non-enum components, exactly one
  // exported type. Cross-checking against it (rather than only asserting the
  // regex-scraped set is non-empty) catches a future Hey API output-format
  // change that silently narrows or duplicates the facade while still
  // producing well-formed, non-empty source.
  const catalogNames = Object.keys(spec?.components?.schemas ?? {});
  if (catalogNames.length === 0) {
    throw new Error(`TypeScript contract generation found no components.schemas in ${specPath}; the public contract facade cannot be checked for completeness; regenerate the Types OpenAPI spec from Go.`);
  }
  const expectedSchemaNames = new Set(catalogNames.map((name) => `z${name}`));
  const expectedTypeNames = new Set(catalogNames.filter((name) => !enumNames.has(name)));
  const actualSchemaNames = new Set(schemaNames);
  const actualTypeNames = new Set(contractTypes);
  const missingSchemas = [...expectedSchemaNames].filter((name) => !actualSchemaNames.has(name));
  const extraSchemas = [...actualSchemaNames].filter((name) => !expectedSchemaNames.has(name));
  const missingTypes = [...expectedTypeNames].filter((name) => !actualTypeNames.has(name));
  const extraTypes = [...actualTypeNames].filter((name) => !expectedTypeNames.has(name));
  if (missingSchemas.length > 0 || extraSchemas.length > 0 || missingTypes.length > 0 || extraTypes.length > 0) {
    throw new Error(`TypeScript contract generation found src/internal/generated/contract/zod.gen.ts's exports do not exactly match the ${catalogNames.length}-component Types OpenAPI catalog at ${specPath}; the public contract facade would silently drop or add names; missing schema export(s): [${missingSchemas.join(", ") || "none"}]; unexpected schema export(s): [${extraSchemas.join(", ") || "none"}]; missing type export(s): [${missingTypes.join(", ") || "none"}]; unexpected type export(s): [${extraTypes.join(", ") || "none"}]; inspect the pinned Hey API output before updating the generator parser.`);
  }

  return `${header()}export { ${schemaNames.join(", ")} } from "./contract/zod.gen.js";\nexport type { ${contractTypes.join(", ")} } from "./contract/zod.gen.js";\n`;
}

// renderTestcaseModel generates the closed Classification and ProvenanceSource
// sets from testcase/testcase.go's AllClassifications/AllProvenanceSources, so
// the TypeScript meta-testing vocabulary can never silently drift from the Go
// source it mirrors (a real regression class: this package's own review
// history landed and then lost this generation once already). These are test
// classification vocabulary, not wire types, so they are generated directly
// from Go source text rather than routed through the Types OpenAPI catalog.
function renderTestcaseModel(source) {
  const models = [
    { goType: "Classification", allVar: "AllClassifications", tsPrefix: "" },
    { goType: "ProvenanceSource", allVar: "AllProvenanceSources", tsPrefix: "Source" },
  ];
  const blocks = models.map((model) => renderTestcaseClosedSet(source, model));
  return `${header()}${blocks.join("\n\n")}\n`;
}

function renderTestcaseClosedSet(source, model) {
  const constPattern = new RegExp(`(\\w+)\\s+${model.goType}\\s*=\\s*"([^"]+)"`, "g");
  const values = new Map();
  for (const match of source.matchAll(constPattern)) values.set(match[1], match[2]);
  if (values.size === 0) {
    throw new Error(`TypeScript testcase generation found no "<Name> ${model.goType} = \"...\"" const declarations in testcase/testcase.go; the ${model.goType} closed set would be empty; keep the canonical Go consts in the expected declaration form or update typescript/scripts/generate-contract-support.mjs.`);
  }

  const allPattern = new RegExp(`var\\s+${model.allVar}\\s*=\\s*\\[\\]${model.goType}\\{([^}]*)\\}`, "s");
  const allMatch = allPattern.exec(source);
  if (allMatch === null) {
    throw new Error(`TypeScript testcase generation could not find "var ${model.allVar} = []${model.goType}{...}" in testcase/testcase.go; the ${model.goType} member order cannot be trusted; keep the canonical Go slice in the expected declaration form or update typescript/scripts/generate-contract-support.mjs.`);
  }
  const order = allMatch[1].split(",").map((entry) => entry.trim()).filter((entry) => entry.length > 0);
  if (order.length === 0) {
    throw new Error(`TypeScript testcase generation found an empty ${model.allVar} in testcase/testcase.go; the ${model.goType} closed set would be empty; add at least one member.`);
  }

  const members = order.map((goName) => {
    const value = values.get(goName);
    if (value === undefined) {
      throw new Error(`TypeScript testcase generation found ${model.allVar} member ${goName} in testcase/testcase.go without a matching "${goName} ${model.goType} = \"...\"" const; the runtime inventory cannot be rendered; add the const or remove the unsupported member.`);
    }
    if (model.tsPrefix !== "" && !goName.startsWith(model.tsPrefix)) {
      throw new Error(`TypeScript testcase generation found ${model.goType} member ${goName} in testcase/testcase.go without the expected "${model.tsPrefix}" prefix; the TypeScript member name cannot be derived; keep the canonical Go naming convention or update typescript/scripts/generate-contract-support.mjs.`);
    }
    const tsName = model.tsPrefix === "" ? goName : goName.slice(model.tsPrefix.length);
    return { tsName, value };
  });

  const renderedMembers = members.map((member) => `  ${member.tsName}: ${JSON.stringify(member.value)},`).join("\n");
  const allMembers = members.map((member) => `${model.goType}.${member.tsName}`).join(", ");
  return `export const ${model.goType} = Object.freeze({\n${renderedMembers}\n} as const);\nexport type ${model.goType} = (typeof ${model.goType})[keyof typeof ${model.goType}];\nexport const ${model.allVar} = Object.freeze([${allMembers}]) as readonly ${model.goType}[];\nexport function is${model.goType}(value: unknown): value is ${model.goType} {\n  return typeof value === "string" && (${model.allVar} as readonly string[]).includes(value);\n}`;
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
      if (!("associations" in commit)) commit.associations = null;
    }
  }
  return `${header()}import type { TimelineFixtureCorpus } from "../../fixtures/timeline.js";\n\nexport const canonicalTimelineFixtures: TimelineFixtureCorpus = ${JSON.stringify(fixtures, null, 2)};\n`;
}

async function generateOperationContracts(surface, filename, outputName) {
  const apiPath = join(moduleRoot, "generated", filename);
  const apiSpec = JSON.parse(await readFile(apiPath, "utf8"));
  const aliases = canonicalOperationAliases(surface, apiSpec, spec);
  const ast = await openapiTS(apiSpec, {
    inject: 'import type * as Schema from "../../index.js";',
    silent: true,
    transform(_schemaObject, metadata) {
      const prefix = "#/components/schemas/";
      if (!metadata.path.startsWith(prefix) || metadata.path.slice(prefix.length).includes("/")) return;
      const rootName = aliases.get(metadata.path.slice(prefix.length));
      return rootName === undefined ? undefined : schemaTypeReference(rootName);
    },
  });
  await writeFile(join(generatedRoot, outputName), astToString(ast));
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
