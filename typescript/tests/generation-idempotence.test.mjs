import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";

const typescriptRoot = fileURLToPath(new URL("..", import.meta.url));
const zodPath = fileURLToPath(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url));

test("the real TypeScript generator preserves root Zod bytes on its second run", async () => {
  await runGenerate();
  const firstGeneration = await readFile(zodPath);
  await runGenerate();
  const secondGeneration = await readFile(zodPath);
  assert.deepEqual(secondGeneration, firstGeneration, "the second pnpm run generate output must match the first root Zod output byte-for-byte");
});

async function runGenerate() {
  const command = "pnpm run generate";
  let result;
  try {
    result = await new Promise((resolve, reject) => {
      const child = spawn("pnpm", ["run", "generate"], { cwd: typescriptRoot, stdio: ["ignore", "pipe", "pipe"] });
      let stdout = "";
      let stderr = "";
      child.stdout.setEncoding("utf8");
      child.stderr.setEncoding("utf8");
      child.stdout.on("data", (chunk) => { stdout += chunk; });
      child.stderr.on("data", (chunk) => { stderr += chunk; });
      child.once("error", reject);
      child.once("close", (code, signal) => resolve({ code, signal, stdout, stderr }));
    });
  } catch (error) {
    const reason = error instanceof Error ? error.message : String(error);
    throw new Error(`TypeScript generator command could not start: command=${JSON.stringify(command)} cwd=${JSON.stringify(typescriptRoot)} stderr="" reason=${JSON.stringify(reason)}`, { cause: error });
  }
  if (result.code !== 0) {
    throw new Error(`TypeScript generator command failed: command=${JSON.stringify(command)} cwd=${JSON.stringify(typescriptRoot)} exitCode=${result.code} signal=${result.signal ?? "none"} stderr=${JSON.stringify(result.stderr)} stdout=${JSON.stringify(result.stdout)}`);
  }
}
