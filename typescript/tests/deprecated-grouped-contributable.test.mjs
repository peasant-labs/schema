import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse, stringify } from "yaml";
import * as api from "../dist/index.js";
import { loadCorpus } from "../dist/testcase.js";

const fixture = parse(await readFile(new URL("../../testdata/deprecated_grouped_contributable.yaml", import.meta.url), "utf8"));
const cases = loadCorpus(stringify(fixture.envelopes, { aliasDuplicateObjects: false }), {
  decodeInput(value) { assert.ok(value && typeof value === "object" && !Array.isArray(value)); return value; },
  decodeExpected(value) { assert.equal(typeof value?.valid, "boolean"); return value; },
}).cases;
const village = JSON.parse(await readFile(new URL(`../../generated/village-api-${api.VillageAPIVersion}.json`, import.meta.url), "utf8"));
test("grouped contribution transition retains its exact named cases", () => {
  assert.deepEqual(new Set(fixture.required_names), new Set(cases.map(row => row.name)));
  assert.equal(new Set(fixture.required_names).size, fixture.required_names.length);
  assert.equal(new Set(cases.map(row => row.name)).size, cases.length);
});
test("each route retains its own grouped union and the common omitted-view arm", () => {
  const legacyComponent = village.components.schemas.SchemaVillageDeprecatedGroupedContributableResponse;
  assert.equal(legacyComponent.deprecated, true);
  assert.deepEqual(Object.keys(legacyComponent.properties).sort(), ["groupId", "transcriptList"]);
  assert.deepEqual([...legacyComponent.required].sort(), ["groupId", "transcriptList"]);
  const canonicalComponent = village.components.schemas.SchemaVillageGroupedContributableResponse;
  assert.deepEqual(Object.keys(canonicalComponent.properties).sort(), ["group_id", "transcriptList"]);
  assert.deepEqual([...canonicalComponent.required].sort(), ["group_id", "transcriptList"]);
  assert.deepEqual(new Set(fixture.required_routes), new Set(fixture.routes.map(row => row.name)));
  assert.equal(new Set(fixture.routes.map(row => row.name)).size, fixture.routes.length);
  assert.equal(new Set(fixture.required_routes).size, fixture.required_routes.length);
  for (const row of fixture.routes) {
    const operation = village.paths[row.path].get;
    assert.equal(operation.operationId, row.operation_id);
    assert.equal(operation.deprecated ?? false, row.deprecated);
    const arms = operation.responses[200].content["application/json"].schema.oneOf;
    const names = arms.map(arm => arm.$ref.replace("#/components/schemas/Schema", ""));
    const grouped = row.path.includes("/groups/") ? "VillageDeprecatedGroupedContributableResponse" : "VillageGroupedContributableResponse";
    assert.deepEqual(names, ["VillageContributableResponse", grouped]);
    const payload = row.view === "grouped" ? cases.find(value => value.expected.valid && value.input.schema === row.selected).input.payload : fixture.ungrouped;
    assert.equal(api[`z${row.selected}`].safeParse(payload).success, true, row.name);
    assert.equal(names.some(name => api[`z${name}`].safeParse(payload).success), true, row.name);
    if (row.rejected) {
      const other = cases.find(value => value.expected.valid && value.input.schema === row.rejected).input.payload;
      assert.equal(names.some(name => api[`z${name}`].safeParse(other).success), false, `${row.name} must not advertise the opposite grouped wire`);
    }
  }
});
for (const row of cases) test(`public grouped contribution transition: ${row.name}`, () => {
  const result = api[`z${row.input.schema}`].safeParse(row.input.payload);
  assert.equal(result.success, row.expected.valid, result.error?.message);
  if (result.success) {
    assert.deepEqual(result.data, row.input.payload);
    assert.deepEqual(Object.keys(result.data).sort(), [...row.input.keys].sort());
  }
});
