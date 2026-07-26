import type { CommitRef, RewrittenCommit, SessionAssociation, SessionID, TimelineSessionRef } from "../index.js";
import { canonicalTimelineFixtures } from "../internal/generated/timeline-fixtures.gen.js";
import type { Case } from "../testcase.js";

export type TimelineFixtureSessionInput = Omit<TimelineSessionRef, "harness"> & { harness: string };
// SessionAssociation's conclusion/confidence/evidence fields are widened to plain
// string so the fixture corpus can carry must-fail negatives (a value
// outside the closed set) without violating the strict generated union.
export interface TimelineFixtureEvidenceObservation {
  kind: string;
  recordedCommitHash?: string;
  touchedFilePath?: string;
  branchName?: string;
  windowStartMs?: number;
  windowEndMs?: number;
}
export type TimelineFixtureAssociationInput = Omit<SessionAssociation, "conclusion" | "confidence" | "evidence"> & {
  conclusion: string;
  confidence: string;
  evidence: TimelineFixtureEvidenceObservation[] | null;
};
export type TimelineFixtureCommitInput = Omit<CommitRef, "sessionIds" | "associations"> & {
  sessionIds: SessionID[] | null;
  // null models a fixture case that omits the required associations array
  // entirely (the non-nil Associations invariant), mirroring
  // how sessionIds models the null-sessionIds negative case above.
  associations: TimelineFixtureAssociationInput[] | null;
};
export type TimelineFixtureRepairKind = "set_session_binding_true" | "replace_successor_association";
export interface TimelineFixtureSetSessionBindingRepair {
  kind: "set_session_binding_true";
  sessionId: SessionID;
  postMutationValid: boolean;
}
export interface TimelineFixtureReplaceSuccessorAssociationRepair {
  kind: "replace_successor_association";
  ghostHash: string;
  successorHash: string;
  associationId: SessionAssociation["id"];
  postMutationValid: boolean;
}
export type TimelineFixtureRepair = TimelineFixtureSetSessionBindingRepair | TimelineFixtureReplaceSuccessorAssociationRepair;
export interface TimelineFixtureInput { sessions: TimelineFixtureSessionInput[]; commits: TimelineFixtureCommitInput[]; rewrittenCommits?: RewrittenCommit[]; }
export interface TimelineFixtureExpected { errorContains?: string; repair?: TimelineFixtureRepair; }
export type TimelineFixtureCase = Case<TimelineFixtureInput, TimelineFixtureExpected> & { family: string };
export interface TimelineFixtureCorpus {
  cases: TimelineFixtureCase[];
  successorAssociationMirrorCases: TimelineFixtureCase[];
}

export function loadTimelineFixtures(): TimelineFixtureCorpus { return structuredClone(canonicalTimelineFixtures); }
