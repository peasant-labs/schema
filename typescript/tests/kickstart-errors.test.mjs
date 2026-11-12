import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";
import * as localApi from "../dist/local-api.js";

// The Go suite (local_kickstart_errors_test.go) holds the closed error-code set,
// the status each code answers with, and the kickstart route table to the
// generated Local API document. This suite holds the TypeScript binding to the
// same two fixture files, so neither language can drift from the other:
// the codes a client parses, the statuses a route declares, the exported
// dropped-label bound, and the declared scan notice. The error body, the scan
// bound and the notice shapes are replayed from the shared corpus in
// testdata/local-api/kickstart.yaml, so they are not repeated here.

const errors = parse(await readFile(new URL("../../testdata/local-api/kickstart_errors.yaml", import.meta.url), "utf8"));
const sets = parse(await readFile(new URL("../../testdata/local-api/kickstart_sets.yaml", import.meta.url), "utf8"));
const localDocument = JSON.parse(await readFile(new URL(`../../generated/peasantlocal-api-${schema.PeasantLocalAPIVersion}.json`, import.meta.url), "utf8"));

test("the closed error-code set is exactly the codes the fixture pins to a status", () => {
  assert.deepEqual([...schema.AllLocalKickstartErrorCodes].sort(), Object.keys(errors.code_status).sort());
  for (const code of Object.keys(errors.code_status)) {
    assert.equal(schema.isLocalKickstartErrorCode(code), true, `${code} is pinned by the fixture but refused by the guard`);
  }
  assert.equal(schema.isLocalKickstartErrorCode("kickstart_made_up"), false);
  assert.equal(schema.isLocalKickstartErrorCode(undefined), false);
  assert.deepEqual(schema.LocalKickstartErrorCode.FieldHidden, "kickstart_field_hidden");
});

test("the kickstart route table is exactly the document's kickstart operations", () => {
  const declared = [];
  for (const [path, item] of Object.entries(localDocument.paths)) {
    if (!path.startsWith("/api/v1/kickstart/")) continue;
    for (const method of Object.keys(item)) declared.push(`${method} ${path}`);
  }
  const pinned = errors.routes.map((route) => `${route.method} ${route.path}`);
  assert.deepEqual(declared.sort(), pinned.sort());
  assert.deepEqual(errors.routes.map((route) => route.operation_id).sort(), [...errors.required_routes].sort());
});

for (const route of errors.routes) {
  test(`${route.operation_id} declares its error statuses and codes`, () => {
    const operation = localDocument.paths[route.path][route.method];
    assert.equal(operation.operationId, route.operation_id);
    const failures = Object.keys(operation.responses).filter((status) => !status.startsWith("2")).sort();
    assert.deepEqual(failures, [...Object.keys(route.errors), ...route.other_statuses].sort());
    for (const [status, codes] of Object.entries(route.errors)) {
      const response = operation.responses[status];
      assert.match(response.content["application/json"].schema.$ref, /LocalKickstartErrorResponse$/, `${route.operation_id} ${status} body`);
      const marker = "one of: ";
      const at = response.description.lastIndexOf(marker);
      assert.notEqual(at, -1, `${route.operation_id} ${status} description ${JSON.stringify(response.description)} does not list its codes after ${JSON.stringify(marker)}; the generator states the codes there and this test reads them from it`);
      const listed = response.description.slice(at + marker.length).replace(/\.$/, "").split(", ");
      assert.deepEqual(listed, codes, `${route.operation_id} ${status} described codes`);
      for (const code of codes) {
        assert.equal(errors.code_status[code], Number(status), `${code} is pinned to ${errors.code_status[code]} but ${route.operation_id} answers it with ${status}`);
      }
    }
  });
}

test("the dropped-label bound is exported once and the scan schema enforces the same number", () => {
  assert.equal(schema.LocalKickstartDroppedLabelsMax, sets.dropped_labels_max);
  assert.equal(localApi.LocalKickstartDroppedLabelsMax, sets.dropped_labels_max);
  const scan = (labels) => ({ state: "done", done: 1, total: 1, droppedSelections: labels.length, droppedLabels: labels });
  const within = Array.from({ length: schema.LocalKickstartDroppedLabelsMax }, (_, index) => `label ${index}`);
  assert.equal(schema.zLocalKickstartScan.safeParse(scan(within)).success, true);
  assert.equal(schema.zLocalKickstartScan.safeParse(scan([...within, "one more"])).success, false);
});

test("a partial scan reports through the declared notice field", () => {
  const field = sets.scan_notice_field;
  assert.ok(Object.hasOwn(schema.zLocalKickstartScan.shape, field), `zLocalKickstartScan declares no ${field} field`);
  const scan = { state: "done", done: 4, total: 4, [field]: "Claude Code could not be read: the folder is not readable" };
  assert.equal(schema.zLocalKickstartScan.safeParse(scan).success, true);
  assert.equal(schema.zLocalKickstartScan.safeParse({ ...scan, phase: "reading" }).success, true);
});
