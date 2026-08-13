import assert from "node:assert/strict";
import fs from "node:fs/promises";
import test from "node:test";
import YAML from "yaml";

import * as schema from "../dist/index.js";
import { validateCorpus } from "../dist/testcase.js";

const fixtureURL = new URL("../../openapi/testdata/ui_capabilities_contract.yaml", import.meta.url);
const fixture = YAML.parse(await fs.readFile(fixtureURL, "utf8"));

test("generated UI capability Zod follows the strict shared response corpus", async (t) => {
  validateCorpus(fixture.cases, { minCases: 7 });
  for (const row of fixture.cases.cases) {
    await t.test(row.name, () => {
      const body = JSON.parse(row.input);
      assert.equal(schema.zUICapabilitiesResponse.safeParse(body).success, row.expected);
    });
  }
});
