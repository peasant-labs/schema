import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { readFile } from "node:fs/promises";

import { parse } from "yaml";

const fixture = parse(await readFile(new URL("../tests/fixtures/package-files.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture), ["files"]);
assert.ok(Array.isArray(fixture.files));

const packed = JSON.parse(execFileSync("pnpm", ["pack", "--json", "--dry-run"], { encoding: "utf8" }));
const actual = packed.files.map((entry) => entry.path).sort();
const expected = [...fixture.files].sort();
assert.deepEqual(actual, expected, "packed package contents differ from the audited fixture");
