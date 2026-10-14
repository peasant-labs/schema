import assert from "node:assert/strict";
import test from "node:test";

import { zSessionDetailPayload, zTurnDetail } from "../dist/index.js";
import { issuePaths, parity, replayedCases, turnCorpus } from "./thinking-level-parity.mjs";

const sessionThinkingFields = zSessionDetailPayload.pick({ thinkingLevel: true, thinkingLevelRaw: true }).strict();

test("generated TurnDetail and session detail agree with the Go turn thinking level corpus", async (t) => {
  for (const fixtureCase of replayedCases("turn corpus", turnCorpus, parity.turn)) {
    await t.test(fixtureCase.name, () => {
      const turn = { ...parity.turn.baseTurn, ...JSON.parse(fixtureCase.input.turn) };
      const turnResult = zTurnDetail.safeParse(turn);
      const sessionResult = fixtureCase.input.session === undefined ? { success: true, data: {} } : sessionThinkingFields.safeParse(JSON.parse(fixtureCase.input.session));
      const accepted = turnResult.success && sessionResult.success;
      assert.equal(accepted, fixtureCase.expected.accepted, `${fixtureCase.name}: generated verdict differs from Go (${[...issuePaths(turnResult), ...issuePaths(sessionResult)].join(", ")})`);
      if (!accepted) {
        const paths = [...issuePaths(turnResult), ...issuePaths(sessionResult)];
        assert.ok(paths.every((path) => path === "thinkingLevel" || path === "thinkingLevelRaw"), `${fixtureCase.name}: the rejection must come from the thinking level fields, got ${paths.join(", ")}`);
        return;
      }
      for (const key of ["thinkingLevel", "thinkingLevelRaw"]) {
        const want = fixtureCase.expected[key];
        assert.equal(Object.hasOwn(turnResult.data, key), want !== undefined, `${fixtureCase.name}: turn ${key} presence changed`);
        if (want !== undefined) assert.equal(turnResult.data[key], want, `${fixtureCase.name}: turn ${key} bytes changed`);
      }
      if (fixtureCase.expected.sessionThinkingLevel !== undefined) {
        assert.equal(sessionResult.data.thinkingLevel, fixtureCase.expected.sessionThinkingLevel, `${fixtureCase.name}: session seed bytes changed`);
      }
    });
  }
});
