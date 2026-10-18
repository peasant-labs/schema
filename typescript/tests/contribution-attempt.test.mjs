import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";
import { reconcileVillageContributionAttempt, villageContributionAttemptActor, zVillageShareEvent } from "../dist/index.js";

const fixtures = parse(await readFile(new URL("../../testdata/collective_contribution_attempts.yaml", import.meta.url), "utf8"), { merge: true, maxAliasCount: 1000 });
const cases = fixtures.corpus.cases;
assert.deepEqual(new Set(cases.map(c => c.name)), new Set(fixtures.required_names));
assert.equal(cases.length, new Set(cases.map(c => c.name)).size, "duplicate fixture name");

function value(input) {
  if (input === undefined) return undefined;
  return {
    collectiveId: input.collective_id, transcriptId: input.transcript_id, firstSeenAt: input.first_seen_at,
    attempt: { event_num: input.event_num, status: input.status, recorded_at: input.recorded_at, decided_at: input.decided_at, decided_by_actor: input.decided_by_actor },
  };
}

for (const c of cases) {
  test(`contribution reconciliation follows shared Go corpus: ${c.name}`, () => {
    const saved = value(c.input.saved), incoming = value(c.input.incoming);
    const before = structuredClone({ saved, incoming });
    if (c.classification === "must-fail") {
      assert.throws(() => reconcileVillageContributionAttempt(saved, incoming));
    } else {
      const result = reconcileVillageContributionAttempt(saved, incoming);
      assert.equal(result.disposition, c.expected.disposition);
      assert.equal(result.attempt.attempt.status, c.expected.status);
      const expected = saved !== undefined && c.expected.disposition !== "updated" ? saved : { ...incoming, firstSeenAt: saved?.firstSeenAt ?? incoming.firstSeenAt };
      assert.deepEqual(result.attempt, expected);
      assert.notEqual(result.attempt, incoming);
      assert.notEqual(result.attempt.attempt, incoming.attempt);
      if (saved !== undefined) assert.notEqual(result.attempt.attempt, saved.attempt);
      assert.equal(villageContributionAttemptActor(result.attempt.attempt.status), result.attempt.attempt.decided_by_actor);
      assert.equal(zVillageShareEvent.safeParse(result.attempt.attempt).success, true);
    }
    assert.deepEqual({ saved, incoming }, before, "reconciliation must not mutate evidence");
  });
}
