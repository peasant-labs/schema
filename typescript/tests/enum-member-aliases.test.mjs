import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";

import ts from "typescript";
import { parse } from "yaml";

const fixture = parse(await readFile(new URL("../../testdata/typescript/enums.yaml", import.meta.url), "utf8"));
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
