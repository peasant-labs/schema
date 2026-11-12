import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";

// The Go suites (local_host_refusal_test.go and cross_origin_writes_test.go)
// build the Local API document from source and hold it to
// testdata/local-api/host_refusal.yaml and cross_origin_writes.yaml. This suite
// holds the committed generated document, the one TypeScript consumers
// receive, to the same files, so a token changed in the generator and
// regenerated cannot reach a client without a fixture changing too. The
// operation description must end with the pinned sentence and the 403
// description must equal the pinned one, so a longer token that merely
// contains the pinned one fails.

const fixture = parse(await readFile(new URL("../../testdata/local-api/host_refusal.yaml", import.meta.url), "utf8"));
const writes = parse(await readFile(new URL("../../testdata/local-api/cross_origin_writes.yaml", import.meta.url), "utf8"));
const document = JSON.parse(await readFile(new URL(`../../generated/peasantlocal-api-${schema.PeasantLocalAPIVersion}.json`, import.meta.url), "utf8"));

function endsWithSentence(description, sentence) {
  return description.endsWith(` ${sentence}`);
}

function refBody(response) {
  return response.content["application/json"].schema.$ref;
}

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
    assert.ok(endsWithSentence(operation.description, fixture.operation_sentence), `${read.operation_id} operation description does not end with ${JSON.stringify(fixture.operation_sentence)}: ${JSON.stringify(operation.description)}`);
    const response = operation.responses[fixture.status];
    assert.match(refBody(response), new RegExp(`/${fixture.component}$`), `${read.operation_id} ${fixture.status} body`);
    assert.equal(response.description, fixture.response_description);
  });
}

test("every declared write names both refusal codes in its 403", () => {
  const refusal = writes.refusal;
  for (const code of refusal.codes) {
    assert.ok(refusal.operation_sentence.includes(code) && refusal.response_description.includes(code), `the fixture lists ${code} but a sentence omits it`);
  }
  for (const write of writes.write_operations) {
    const operation = document.paths[write.path][write.method];
    assert.equal(operation.operationId, write.operation_id);
    assert.ok(endsWithSentence(operation.description, refusal.operation_sentence), `${write.operation_id} operation description does not end with the write refusal sentence`);
    const response = operation.responses[refusal.status];
    assert.match(refBody(response), new RegExp(`/${refusal.component}$`), `${write.operation_id} ${refusal.status} body`);
    assert.equal(response.description, refusal.response_description);
  }
});
