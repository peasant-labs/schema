export * from "./internal/generated/public-contract.gen.js";
export * from "./internal/generated/enums.gen.js";
export * from "./internal/generated/versions.gen.js";

import { zProjectHash, type ProjectHash } from "./internal/generated/contract/zod.gen.js";

function assertProjectHash(value: unknown, operation: "newProjectHash" | "validateProjectHash"): asserts value is ProjectHash {
  if (!isProjectHash(value)) {
    const rendered = renderProjectHashInput(value);
    throw new TypeError("ProjectHash validation failed for " + rendered + " at @peasant-labs/schema ProjectHash during " + operation + ": the value is not a 64-character lowercase hexadecimal string; callers cannot use it as a canonical project identity; pass the lowercase SHA-256 hex digest of the project origin URL or local path.");
  }
}

function renderProjectHashInput(value: unknown): string {
  try {
    const json = JSON.stringify(value);
    if (json !== undefined) return json;
  } catch {
    // Fall through to String for values whose JSON hooks throw.
  }
  try {
    return String(value);
  } catch {
    return "<unrenderable value>";
  }
}

export function isProjectHash(value: unknown): value is ProjectHash {
  return zProjectHash.safeParse(value).success;
}

export function validateProjectHash(value: unknown): asserts value is ProjectHash {
  assertProjectHash(value, "validateProjectHash");
}

export function newProjectHash(raw: string): ProjectHash {
  assertProjectHash(raw, "newProjectHash");
  return raw;
}

export type PushContractVersion = import("./internal/generated/contract/zod.gen.js").ContractVersion;
