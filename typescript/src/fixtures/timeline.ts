import type { CommitRef, SessionAssociation, SessionID, TimelineSessionRef } from "../index.js";
import { canonicalTimelineFixtures } from "../internal/generated/timeline-fixtures.gen.js";
import type { Case } from "../testcase.js";

export type TimelineFixtureSessionInput = Omit<TimelineSessionRef, "harness"> & { harness: string };
// SessionAssociation's kind/confidence/evidence fields are widened to plain
// string so the fixture corpus can carry must-fail negatives (a value
// outside the closed set) without violating the strict generated union.
export type TimelineFixtureAssociationInput = Omit<SessionAssociation, "kind" | "confidence" | "evidence"> & {
  kind: string;
  confidence: string;
  evidence: string;
};
export type TimelineFixtureCommitInput = Omit<CommitRef, "sessionIds" | "associations"> & {
  sessionIds: SessionID[] | null;
  // null models a fixture case that omits the required associations array
  // entirely (the non-nil Associations invariant), mirroring
  // how sessionIds models the null-sessionIds negative case above.
  associations: TimelineFixtureAssociationInput[] | null;
};
export interface TimelineFixtureInput { sessions: TimelineFixtureSessionInput[]; commits: TimelineFixtureCommitInput[]; }
export interface TimelineFixtureExpected { errorContains?: string; }
export type TimelineFixtureCase = Case<TimelineFixtureInput, TimelineFixtureExpected> & { family: string };
export interface TimelineFixtureCorpus { cases: TimelineFixtureCase[]; }

export function loadTimelineFixtures(): TimelineFixtureCorpus { return structuredClone(canonicalTimelineFixtures); }
