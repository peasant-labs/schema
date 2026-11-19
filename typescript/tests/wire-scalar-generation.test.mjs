import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { z } from "zod";
import { parse } from "yaml";

import { applyWireScalarZodRefinements } from "../scripts/lib/wire-scalar-zod-refinements.mjs";
import { Classification, loadCorpus } from "../dist/testcase.js";

const base = "../../testdata/typescript/wire_scalar_generation";
const raw = await readFile(new URL(`${base}.yaml`, import.meta.url), "utf8");
const manifest = parse(await readFile(new URL(`${base}_manifest.yaml`, import.meta.url), "utf8"), { uniqueKeys: true });
const cases = loadCorpus(raw, {
  decodeInput(value) { return value; },
  decodeExpected(value) { return value; },
}).cases;
assert.deepEqual(Object.keys(manifest), ["requiredNames"]);
assert.ok(manifest.requiredNames.length > 0);
assert.equal(new Set(manifest.requiredNames).size, manifest.requiredNames.length);
assert.deepEqual(cases.map((row) => row.name).sort(), [...manifest.requiredNames].sort());

test("wire scalar generation follows spec bounds and fails on lost markers", async (t) => {
  for (const row of cases) {
    await t.test(row.name, () => {
      const source = `export const zFixture = z.object({\n    value: ${row.input.before}\n});\n\nexport type Fixture = z.infer<typeof zFixture>;`;
      const spec = { components: { schemas: { Fixture: { properties: { value: row.input.property } } } } };
      if (row.input.preserve === "Fixture") {
        spec.components.schemas.PublicationRoot = { properties: { nested: { $ref: "#/components/schemas/Fixture" } } };
      }
      const options = row.input.preserve === undefined ? {} : {
        preserveComponentsReachableFrom: [row.input.preserve === "Fixture" ? "PublicationRoot" : row.input.preserve],
      };
      if (row.classification === Classification.MustFail) {
        assert.throws(() => applyWireScalarZodRefinements(source, spec, options), new RegExp(row.expected.error));
        return;
      }
      const output = applyWireScalarZodRefinements(source, spec, options);
      const expected = source.replace(row.input.before, row.expected.after);
      assert.equal(output, expected);
      // The fixture contains trusted generator expressions, not user data.
      const schema = Function("z", `return ${row.expected.after}`)(z);
      for (const sample of row.expected.values) assert.equal(schema.safeParse(sample.value).success, sample.valid);
    });
  }
});
