import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";

import ts from "typescript";
import { parse } from "yaml";
import { TypesVersion } from "../dist/index.js";

const fixture = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));
const spec = JSON.parse(await readFile(new URL(`../../generated/types-${TypesVersion}.json`, import.meta.url), "utf8"));
const probePath = fileURLToPath(new URL("./enum-member-aliases.probe.ts", import.meta.url));
const probeSource = [
  `import type { TurnDetail } from "@peasant-labs/schema";`,
  `import { ${fixture.enums.map((row) => `z${row.name}`).join(", ")} } from "@peasant-labs/schema";`,
  `export const turnRole = null as unknown as TurnDetail["role"];`,
  ...fixture.enums.map((row) => `export const schema${row.name} = null as unknown as (typeof z${row.name})["_zod"]["output"];`),
].join("\n");

const config = ts.getParsedCommandLineOfConfigFile(fileURLToPath(new URL("../tsconfig.test.json", import.meta.url)), {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: (diagnostic) => { throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, "\n")); } });
const host = ts.createCompilerHost(config.options);
const readSource = host.readFile.bind(host);
host.readFile = (path) => (path === probePath ? probeSource : readSource(path));
host.fileExists = ((exists) => (path) => path === probePath || exists(path))(host.fileExists.bind(host));
const getSourceFile = host.getSourceFile.bind(host);
host.getSourceFile = (path, ...rest) => (path === probePath ? ts.createSourceFile(path, probeSource, config.options.target ?? ts.ScriptTarget.ES2022, true) : getSourceFile(path, ...rest));
const program = ts.createProgram([probePath], config.options, host);
const checker = program.getTypeChecker();
const probe = program.getSourceFile(probePath);

function declaredType(name) {
  for (const statement of probe.statements) {
    for (const declaration of ts.isVariableStatement(statement) ? statement.declarationList.declarations : []) {
      if (declaration.name.getText() === name) return checker.getTypeAtLocation(declaration.name);
    }
  }
  throw new Error(`enum alias probe has no ${name} declaration; the probe source is incoherent`);
}

// A typed consumer resolves a literal through the receiver's declared alias and a
// same-named const table beside it (eslint-plugin-unstring closed-set-literal).
function constTableBesideAlias(alias) {
  const declaration = alias.declarations?.[0];
  const value = checker.resolveName(alias.getName(), declaration, ts.SymbolFlags.Value, false);
  return value?.valueDeclaration?.parent !== undefined && (value.valueDeclaration.parent.flags & ts.NodeFlags.Const) !== 0;
}

test("generated enum schemas and fields report the named set type", async (t) => {
  const fixtureNames = fixture.enums.map((row) => row.name);
  const catalogNames = Object.entries(spec.components.schemas)
    .filter(([, schema]) => Array.isArray(schema.enum))
    .map(([name]) => name);
  assert.equal(new Set(fixtureNames).size, fixtureNames.length, "enum fixture names must be unique");
  assert.deepEqual([...fixtureNames].sort(), catalogNames.sort(), "enum fixture names must match the generated OpenAPI catalog");
  assert.deepEqual(ts.getPreEmitDiagnostics(program, probe).map((d) => ts.flattenDiagnosticMessageText(d.messageText, "\n")), [], "alias probe must compile against the built package");

  await t.test("TurnDetail role", () => {
    const type = declaredType("turnRole");
    assert.equal(type.aliasSymbol?.getName(), "Role", `TurnDetail["role"] is ${checker.typeToString(type)}, not the named Role set`);
    assert.equal(constTableBesideAlias(type.aliasSymbol), true, "the Role alias has no const member table beside it");
  });

  // A one-member set is a single literal type, which TypeScript cannot alias.
  for (const row of fixture.enums.filter((candidate) => candidate.members.length > 1)) {
    await t.test(`z${row.name} output`, () => {
      const type = declaredType(`schema${row.name}`);
      assert.equal(type.aliasSymbol?.getName(), row.name, `z${row.name} reports ${checker.typeToString(type)}, not the named ${row.name} set`);
      assert.equal(constTableBesideAlias(type.aliasSymbol), true, `the ${row.name} alias has no const member table beside it`);
    });
  }
});

test("drift type tests use the generated closed-set helper", async () => {
  const generated = await readFile(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url), "utf8");
  const tests = await readFile(new URL("./closed-set-drift.ts", import.meta.url), "utf8");
  for (const prefix of ["type ClosedSet", "const closedSet"]) {
    const line = (source) => source.split("\n").find((item) => item.startsWith(prefix));
    assert.ok(line(generated), `generated ${prefix} definition must exist`);
    assert.equal(line(tests), line(generated), `the drift test must use the generated ${prefix} definition`);
  }
});
