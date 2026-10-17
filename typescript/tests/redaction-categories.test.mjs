import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse, stringify } from "yaml";
import * as api from "../dist/index.js";
import { loadCorpus } from "../dist/testcase.js";
import { renderRedactionCategories } from "../scripts/lib/redaction-categories.mjs";

const fixture = parse(await readFile(new URL("../../testdata/typescript/redaction_categories.yaml", import.meta.url), "utf8"), { merge: true });
function corpus(arm, manifest) {
  const rows = loadCorpus(stringify(fixture[arm]), {
    decodeInput(value, path) {
      if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: expected an input mapping`);
      return value;
    },
    decodeExpected(value, path) {
      if (typeof value !== "boolean") throw new TypeError(`${path}: expected a boolean verdict`);
      return value;
    },
  }).cases;
  assert.deepEqual(rows.map(row => row.name).sort(), [...fixture[manifest]].sort());
  assert.equal(new Set(rows.map(row => row.name)).size, rows.length);
  for (const row of rows) {
    assert.ok(row.provenance.ref && row.mutation.description);
    assert.ok(row.classification === "must-pass" || row.classification === "must-fail");
  }
  return rows;
}

test("public named bindings preserve every engine pair and frozen inventory", () => {
  const rows = corpus("pairs", "required_pairs");
  assert.deepEqual(Object.keys(api.RedactionCategory).sort(), rows.map(row => row.input.member).sort());
  assert.deepEqual(Object.keys(api.RedactionCategoryLabel).sort(), rows.map(row => row.input.label_member).sort());
  assert.deepEqual([...api.AllRedactionCategories].sort(), rows.map(row => row.input.category).sort());
  assert.deepEqual([...api.AllRedactionCategoryLabels].sort(), rows.map(row => row.input.label).sort());
  assert.ok(Object.isFrozen(api.RedactionCategory) && Object.isFrozen(api.RedactionCategoryLabel));
  assert.ok(Object.isFrozen(api.AllRedactionCategories) && Object.isFrozen(api.AllRedactionCategoryLabels));
  for (const { input } of rows) {
    assert.equal(api.RedactionCategory[input.member], input.category);
    assert.equal(api.RedactionCategoryLabel[input.label_member], input.label);
    assert.equal(api.parseRedactionCategory(input.category), input.category);
    assert.equal(api.parseRedactionCategoryLabel(input.label), input.label);
    assert.equal(api.isRedactionCategory(input.category), true);
    assert.equal(api.redactionCategoryLabel(input.category), input.label);
    assert.equal(api.redactionCategoryFromLabel(input.label), input.category);
    assert.equal(api.redactionCategoryFromLabel(api.redactionCategoryLabel(input.category)), input.category);
    assert.equal(api.redactionCategoryLabel(api.redactionCategoryFromLabel(input.label)), input.label);
  }
});

test("public parse and typed conversion reject untyped JS misuse without echoing input", () => {
  for (const { input } of corpus("invalid", "required_invalid")) {
    assert.equal(api.isRedactionCategory(input.key), false);
    assert.throws(() => api.parseRedactionCategory(input.key), TypeError);
    assert.throws(() => api.parseRedactionCategoryLabel(input.label), TypeError);
    assert.throws(() => api.redactionCategoryLabel(input.key), /@peasant-labs\/schema redactionCategoryLabel.*case-sensitive.*named binding.*Allowed values:/);
    assert.throws(() => api.redactionCategoryFromLabel(input.label), TypeError);
  }
  assert.equal(api.isRedactionCategory(undefined), false);
  assert.equal(api.isRedactionCategory(Symbol()), false);
  assert.throws(() => api.parseRedactionCategoryLabel(undefined), TypeError);
  assert.throws(() => api.parseRedactionCategory({ toString() { throw Error("private"); } }), error => error instanceof TypeError && !error.message.includes("private"));
});

test("released engine projection is pinned, deterministic and rejects invalid projections", async () => {
  const options = { cwd: new URL("../../tools/redaction-categories/", import.meta.url), env: { ...process.env, GOWORK: "off" }, encoding: "utf8", timeout: 120000 };
  const first = execFileSync("go", ["run", "-mod=readonly", "."], options);
  assert.equal(execFileSync("go", ["run", "-mod=readonly", "."], options), first);
  const projection = JSON.parse(first);
  const mod = await readFile(new URL("../../tools/redaction-categories/go.mod", import.meta.url), "utf8");
  const sum = await readFile(new URL("../../tools/redaction-categories/go.sum", import.meta.url), "utf8");
  assert.match(mod, /require github.com\/peasant-labs\/redact v0\.1\.6/);
  assert.ok(sum.includes(`${projection.module} ${projection.version} ${projection.sum}`));
  assert.equal(await readFile(new URL("../src/internal/generated/redaction-categories.gen.ts", import.meta.url), "utf8"), renderRedactionCategories(projection));
  for (const { input } of corpus("projection", "required_projection")) {
    const mutated = structuredClone(projection);
    switch (input.operation) {
      case "missing-key": mutated.categories.pop(); break;
      case "duplicate-label": mutated.labels[1].value = mutated.labels[0].value; break;
      case "empty-label": mutated.labels[0].value = ""; break;
      case "missing-pair": mutated.pairs.pop(); break;
      case "duplicate-pair": mutated.pairs[1] = mutated.pairs[0]; break;
      case "wrong-pin": mutated.version = "v0.0.0"; break;
      case "replaced-module": mutated.module = "local"; break;
      default: assert.fail("unknown projection mutation");
    }
    assert.throws(() => renderRedactionCategories(mutated), /projection validation failed/);
  }
});
