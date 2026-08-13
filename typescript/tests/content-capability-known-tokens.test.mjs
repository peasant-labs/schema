import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";
import * as schema from "../dist/index.js";

// This suite proves the two properties of the content-capability contract
// SIMULTANEOUSLY, from one strict fixture:
//
//   1. the strongly typed, CLOSED known-token inventory (KnownContentCapability
//      / AllContentCapabilities / isContentCapability) exposes exactly the Go
//      inventory, and its first token equals "observed_model_v1"; and
//   2. the discovery wire stays OPEN — a SchemaVersionResponse carrying unknown
//      future tokens (and duplicates) still parses, and the typed guard filters
//      the parsed list down to only the tokens this release understands.
//
// Losing either property is a regression: a narrowed wire would break older
// clients, and a lost inventory would force consumers back to stringly typed
// tokens. Combinatorial discovery cases live in the fixture, not inline here.

const fixture = parse(await readFile(new URL("../../testdata/typescript/content_capabilities.yaml", import.meta.url), "utf8"));
assert.deepEqual(Object.keys(fixture).sort(), ["discovery", "known"]);

const known = fixture.known;
assert.equal(known.name, "KnownContentCapability");
assert.equal(known.wire_alias, "ContentCapability");
assert.equal(known.all_name, "AllContentCapabilities");
assert.equal(known.guard, "isContentCapability");
assert.ok(Array.isArray(known.members) && known.members.length > 0, "known.members must list at least one token");
assert.ok(Array.isArray(known.all_values) && known.all_values.length > 0, "known.all_values must list at least one token");

assert.ok(Array.isArray(fixture.discovery) && fixture.discovery.length > 0, "discovery cases must be a non-empty sequence");
assert.equal(new Set(fixture.discovery.map((discoveryCase) => discoveryCase.name)).size, fixture.discovery.length, "discovery case names must be unique");

function baseResponse(contentCapabilities) {
  const response = {
    annotationSchemaVersion: "1",
    supportedTargetKinds: [],
    supportedTypeIds: [],
    pushContractVersion: "1",
    minPushContractVersion: "1",
  };
  if (contentCapabilities !== undefined) response.contentCapabilities = contentCapabilities;
  return response;
}

test("the typed closed inventory equals the Go known-token set", () => {
  assert.equal(schema.KnownContentCapability.ObservedModelV1, "observed_model_v1");
  assert.deepEqual([...schema.AllContentCapabilities], known.all_values);
  for (const member of known.members) {
    assert.equal(schema.KnownContentCapability[member.name], member.value, `KnownContentCapability.${member.name} must equal ${member.value}`);
    assert.equal(schema.isContentCapability(member.value), true, `isContentCapability must accept the known token ${member.value}`);
  }
  assert.equal(schema.isContentCapability("future_feature_v9"), false, "isContentCapability must reject an unknown future token");
  assert.equal(schema.isContentCapability(42), false, "isContentCapability must reject a non-string value");
});

test("discovery stays open while the typed guard filters known tokens", async (t) => {
  for (const discoveryCase of fixture.discovery) {
    await t.test(discoveryCase.name, () => {
      const parsed = schema.zSchemaVersionResponse.safeParse(baseResponse(discoveryCase.advertised));
      assert.equal(parsed.success, true, `discovery advertising ${JSON.stringify(discoveryCase.advertised)} must parse on the open wire`);
      // The open wire preserves every advertised token verbatim, including
      // unknown and duplicate ones (no narrowing, no dedupe at parse time).
      assert.deepEqual(parsed.data.contentCapabilities, discoveryCase.advertised);
      // The typed guard is what narrows an arbitrary discovered list down to
      // the closed inventory the consumer can act on.
      const guarded = parsed.data.contentCapabilities.filter((token) => schema.isContentCapability(token));
      const deduped = [...new Set(guarded)];
      assert.deepEqual(deduped, discoveryCase.knownFiltered, `guarded+deduped tokens for ${discoveryCase.name} must equal the expected known set`);
      for (const token of deduped) {
        assert.ok(schema.AllContentCapabilities.includes(token), `${token} must belong to the closed AllContentCapabilities inventory`);
      }
    });
  }
});

test("null discovery is rejected while omitted and empty are accepted", () => {
  assert.equal(schema.zSchemaVersionResponse.safeParse(baseResponse(undefined)).success, true);
  assert.equal(schema.zSchemaVersionResponse.safeParse(baseResponse([])).success, true);
  assert.equal(schema.zSchemaVersionResponse.safeParse(baseResponse(null)).success, false);
});
