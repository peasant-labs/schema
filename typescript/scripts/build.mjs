import { rm } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

await rm(new URL("../dist", import.meta.url), { recursive: true, force: true });
const compiler = fileURLToPath(new URL("../node_modules/.bin/tsc", import.meta.url));
const result = spawnSync(compiler, ["-p", "tsconfig.json"], { stdio: "inherit" });
if (result.error !== undefined) {
  throw new Error(`TypeScript build could not start tsc while compiling @peasant-labs/schema: ${result.error.message}; no package output is available; install the locked pnpm dependencies and retry.`, { cause: result.error });
}
if (result.status !== 0) {
  throw new Error(`TypeScript build failed in tsc with exit status ${String(result.status)}; the package output is incomplete; fix the reported compiler errors and rerun pnpm --dir typescript build.`);
}
