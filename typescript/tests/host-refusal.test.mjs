import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";

// The Go suite (local_host_refusal_test.go) builds the Local API document from
// source and holds it to testdata/local-api/host_refusal.yaml. This suite holds
// the committed generated document, the one TypeScript consumers receive, to
// the same file, so a token changed in the generator and regenerated cannot
// reach a client without this fixture changing too.

const fixture = parse(await readFile(new URL("../../testdata/local-api/host_refusal.yaml", import.meta.url), "utf8"));
const document = JSON.parse(await readFile(new URL(`../../generated/peasantlocal-api-${schema.PeasantLocalAPIVersion}.json`, import.meta.url), "utf8"));

test("the reads that declare the host refusal are exactly the fixture's", () => {
  assert.deepEqual(fixture.read_operations.map((read) => read.operation_id).sort(), [...fixture.required_operations].sort());
  const declared = [];
  for (const [path, item] of Object.entries(document.paths)) {
    if (item.get?.responses?.[fixture.status] !== undefined) declared.push(`get ${path}`);
  }
  assert.deepEqual(declared.sort(), fixture.read_operations.map((read) => `${read.method} ${read.path}`).sort());
});

for (const read of fixture.read_operations) {
  test(`${read.operation_id} declares the host refusal with its code`, () => {
    const operation = document.paths[read.path][read.method];
    assert.equal(operation.operationId, read.operation_id);
    assert.ok(operation.description.includes(fixture.code), `${read.operation_id} operation description does not name ${fixture.code}`);
    const response = operation.responses[fixture.status];
    assert.match(response.content["application/json"].schema.$ref, new RegExp(`${fixture.component}$`), `${read.operation_id} ${fixture.status} body`);
    assert.ok(response.description.includes(fixture.code), `${read.operation_id} ${fixture.status} description does not name ${fixture.code}`);
  });
}
