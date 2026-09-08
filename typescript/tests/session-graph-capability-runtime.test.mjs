import assert from "node:assert/strict";
import test from "node:test";

import {
  knownContentCapabilities,
  missingContentCapabilities,
  parseSchemaVersionAdvertisement,
  parseSessionDetailPayloadText,
  requiredContentCapabilities,
  validateContentCapabilityAdvertisements,
  zSessionDetailReadPayload,
} from "../dist/index.js";
import { sessionGraphCapabilityFixtures } from "../dist/fixtures/session-graph-capability.js";

const fixtures = sessionGraphCapabilityFixtures();

test("generated runtime derives capabilities from the canonical graph corpus", async (t) => {
  for (const row of fixtures.derivation.cases) await t.test(row.name, () => {
    const detail = parseSessionDetailPayloadText(row.input.detailJSON);
    assert.deepEqual(requiredContentCapabilities(detail), row.expected.capabilities ?? []);
    if (row.input.readJSON !== undefined) {
      zSessionDetailReadPayload.parse(JSON.parse(row.input.readJSON));
      assert.throws(() => parseSessionDetailPayloadText(row.input.rejectedDurableJSON));
    }
  });
});

test("generated runtime reads and validates canonical advertisements", async (t) => {
  for (const row of fixtures.reader.cases) await t.test(row.name, () => {
    let advertised;
    try { advertised = parseSchemaVersionAdvertisement(row.input.json); }
    catch (error) { assert.equal(row.classification, "must-fail", String(error)); return; }
    assert.equal(row.classification, "must-pass");
    assert.deepEqual(knownContentCapabilities(advertised), row.expected.known ?? []);
    assert.deepEqual(missingContentCapabilities(advertised, ["session_graph_provenance_v1"]), row.expected.missing ?? []);
  });
});

test("generated runtime validates canonical producer advertisements", async (t) => {
  for (const row of fixtures.producer.cases) await t.test(row.name, () => {
    const run = () => validateContentCapabilityAdvertisements(row.input.tokens);
    if (row.classification === "must-pass") assert.doesNotThrow(run);
    else assert.throws(run, new RegExp(row.expected.errorContains, "i"));
  });
});
