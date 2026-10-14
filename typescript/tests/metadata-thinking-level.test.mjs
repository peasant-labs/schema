import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { zThinkingLevelRaw, zUnifiedMetadata } from "../dist/index.js";
import { applyThinkingLevelZodRefinements } from "../scripts/lib/thinking-level-zod-refinements.mjs";
import { issuePaths, metadataCorpus, parity, rawCorpus, replayedCases } from "./thinking-level-parity.mjs";

const baseMetadata = JSON.parse(metadataCorpus.baseMetadata);

test("generated UnifiedMetadata agrees with the Go thinking level metadata corpus", async (t) => {
  for (const fixtureCase of replayedCases("metadata corpus", metadataCorpus, parity.metadata)) {
    await t.test(fixtureCase.name, () => {
      assert.equal(fixtureCase.input.operation, "decode", `${fixtureCase.name}: only decode rows replay through the wire schema`);
      const result = zUnifiedMetadata.safeParse({ ...baseMetadata, ...JSON.parse(fixtureCase.input.json) });
      assert.equal(result.success, fixtureCase.expected.accepted, `${fixtureCase.name}: generated zUnifiedMetadata verdict differs from Go (${issuePaths(result).join(", ")})`);
      if (!result.success) {
        assert.ok(issuePaths(result).includes(fixtureCase.expected.errorContains), `${fixtureCase.name}: the rejection must name ${fixtureCase.expected.errorContains}, got ${issuePaths(result).join(", ")}`);
        return;
      }
      for (const key of ["thinkingLevel", "thinkingLevelRaw"]) {
        const want = fixtureCase.expected[key];
        assert.equal(Object.hasOwn(result.data, key), want !== undefined, `${fixtureCase.name}: ${key} presence changed (absence must stay absent)`);
        if (want !== undefined) assert.equal(result.data[key], want, `${fixtureCase.name}: ${key} bytes changed`);
      }
    });
  }
});

test("generated ThinkingLevelRaw agrees with the Go bytes-exact raw corpus", async (t) => {
  for (const fixtureCase of replayedCases("raw corpus", rawCorpus, parity.raw)) {
    await t.test(fixtureCase.name, () => {
      assert.equal(typeof fixtureCase.input.value, "string", `${fixtureCase.name}: replayed raw rows carry a string value`);
      assert.equal(zThinkingLevelRaw.safeParse(fixtureCase.input.value).success, fixtureCase.expected.accepted, `${fixtureCase.name}: generated zThinkingLevelRaw verdict differs from Go`);
    });
  }
});

test("generated ThinkingLevelRaw applies the Unicode edge-whitespace and byte rules of the Go constructor", async (t) => {
  const names = parity.typescriptCases.map((fixtureCase) => fixtureCase.name);
  assert.deepEqual([...names].sort(), [...parity.requiredTypescriptCaseNames].sort(), "TypeScript-only cases must match requiredTypescriptCaseNames exactly");
  for (const fixtureCase of parity.typescriptCases) {
    await t.test(fixtureCase.name, () => {
      const value = JSON.parse(fixtureCase.valueJSON);
      assert.equal(zThinkingLevelRaw.safeParse(value).success, fixtureCase.accepted, `${fixtureCase.name}: generated zThinkingLevelRaw verdict differs from the Go Unicode rule`);
      assert.equal(zUnifiedMetadata.safeParse({ ...baseMetadata, thinkingLevelRaw: value }).success, fixtureCase.accepted, `${fixtureCase.name}: UnifiedMetadata must enforce the same raw rule`);
    });
  }
});

test("thinking level refinement fails closed when the generated declaration drifts", async () => {
  const generated = await readFile(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url), "utf8");
  const declaration = generated.split("\n").find((line) => line.startsWith("export const zThinkingLevelRaw = "));
  assert.ok(declaration?.includes("new TextEncoder().encode(value).length <= 128"), "the committed zThinkingLevelRaw must carry the byte refine");
  assert.ok(declaration?.includes('text !== "\\u0085"'), "the committed zThinkingLevelRaw must carry the edge refine");
  assert.ok(generated.includes("ThinkingLevelChange must carry level or raw"), "the committed zThinkingLevelChange must carry the at-least-one-field refine");

  const raw = 'export const zThinkingLevelRaw = z.string().min(1).max(128).regex(/^x$/);\n';
  const change = 'export const zThinkingLevelChange = z.intersection(z.unknown(), z.object({\n    level: zThinkingLevel.optional(),\n    raw: zThinkingLevelRaw.optional()\n}));\n';
  const refined = applyThinkingLevelZodRefinements(raw + change);
  assert.match(refined, /\.regex\(\/\^x\$\/\)\.refine\(.*\.refine\(/);
  assert.match(refined, /zThinkingLevelChange = z\.intersection\(z\.unknown\(\), z\.object\(\{[\s\S]*\}\)\)\.refine\(/);
  assert.throws(() => applyThinkingLevelZodRefinements(raw.replace(".max(128)", ".max(129)") + change), /could not apply the UTF-8 byte limit and edge-whitespace rule to zThinkingLevelRaw/);
  assert.throws(() => applyThinkingLevelZodRefinements(raw + 'export const zThinkingLevelChange = z.object({});\n'), /could not apply the at-least-one-field rule to zThinkingLevelChange/);
  assert.throws(() => applyThinkingLevelZodRefinements("export const zOther = z.string();\n"), /update scripts\/lib\/thinking-level-zod-refinements\.mjs/);
});
