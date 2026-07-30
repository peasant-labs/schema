import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import * as schema from "../dist/index.js";

// This suite runs the SAME owner-update corpus the Go tests run, against the
// BUILT root exports rather than the source. It exists because the two language
// bindings can disagree in exactly the way this operation cares about: the Go
// decoder rejects unknown fields and explicit nulls, while a plainly generated
// Zod object would strip unknown keys and report success. A consumer installing
// the npm package would then get a validator that says "fine" to a body the
// contract refuses.
const fixtureSource = await readFile(new URL("../../testdata/publish/owner_update.yaml", import.meta.url), "utf8");
const fixture = parse(fixtureSource);

const rows = fixture?.request_validations?.cases;
assert.ok(Array.isArray(rows) && rows.length > 0, "the owner-update corpus has no request_validations cases; this suite would pass while asserting nothing");

// The corpus is shared with Go, so it carries rows about enum membership as well
// as rows about wire shape. Both are meaningful here: the generated Zod contract
// enforces the enums too, so every row should agree across the two bindings.
test("built root Zod owner-update schema agrees with the Go contract on every corpus row", async (t) => {
  let refusals = 0;
  for (const row of rows) {
    await t.test(row.name, () => {
      const accepted = row.expected.accepted;
      assert.equal(
        row.classification,
        accepted ? "must-pass" : "must-fail",
        `${row.name}: classification must agree with the asserted verdict`,
      );
      if (!accepted) refusals += 1;

      const parsed = JSON.parse(row.input);
      const result = schema.zTranscriptUpdateRequest.safeParse(parsed);
      assert.equal(
        result.success,
        accepted,
        `${row.name}: the built package must reach the same verdict as the Go contract for ${row.input}`,
      );

      // A refusal must actually reject, not quietly succeed after discarding the
      // offending key. Asserting on the parsed OUTPUT is what distinguishes
      // "rejected" from "stripped and accepted", which is the failure this whole
      // suite exists to catch.
      if (accepted) {
        assert.deepEqual(
          result.data,
          parsed,
          `${row.name}: an accepted body must survive parsing unchanged; a dropped key would mean the validator silently discarded part of the request`,
        );
      }
    });
  }
  assert.ok(refusals > 0, "the corpus supplied no refusal rows, so this suite could not have caught a permissive validator");
});

test("the owner-update body rejects unknown keys rather than stripping them", () => {
  // Stated directly as well as via the corpus, because this is the property the
  // village defect makes concrete: `tags` is accepted and dropped server-side,
  // and a stripping validator would reproduce that silent no-op client-side.
  const withUnknown = { visibility: "public", tags: ["a"] };
  const result = schema.zTranscriptUpdateRequest.safeParse(withUnknown);
  assert.equal(result.success, false, "an unknown key must be refused; stripping it would hand the caller a success for a change that will not happen");

  const valid = schema.zTranscriptUpdateRequest.safeParse({ visibility: "public" });
  assert.equal(valid.success, true, "the same body without the unknown key must still be accepted, or the assertion above would prove nothing");
});
