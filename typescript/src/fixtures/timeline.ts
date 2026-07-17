import type { CommitRef, SessionID, TimelineSessionRef } from "../index.js";
import { canonicalTimelineFixtures } from "../internal/generated/timeline-fixtures.gen.js";
import type { Case } from "../testcase.js";

export type TimelineFixtureSessionInput = Omit<TimelineSessionRef, "harness"> & { harness: string };
export type TimelineFixtureCommitInput = Omit<CommitRef, "sessionIds"> & { sessionIds: SessionID[] | null };
export interface TimelineFixtureInput { sessions: TimelineFixtureSessionInput[]; commits: TimelineFixtureCommitInput[]; }
export interface TimelineFixtureExpected { errorContains?: string; }
export type TimelineFixtureCase = Case<TimelineFixtureInput, TimelineFixtureExpected> & { family: string };
export interface TimelineFixtureCorpus { cases: TimelineFixtureCase[]; }

export function loadTimelineFixtures(): TimelineFixtureCorpus { return structuredClone(canonicalTimelineFixtures); }
