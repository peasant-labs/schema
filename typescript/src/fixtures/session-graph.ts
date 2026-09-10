import { canonicalSessionGraphFixtures } from "../internal/generated/session-graph-fixtures.gen.js";

export function sessionGraphFixtures() {
  return structuredClone(canonicalSessionGraphFixtures);
}
