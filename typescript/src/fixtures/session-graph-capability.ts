import { canonicalSessionGraphCapabilityFixtures } from "../internal/generated/session-graph-capability-fixtures.gen.js";

export function sessionGraphCapabilityFixtures() {
  return structuredClone(canonicalSessionGraphCapabilityFixtures);
}
