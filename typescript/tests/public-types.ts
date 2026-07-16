import type {
  PublishRequest,
  ProjectHash,
  ProjectResolutionPayload,
  ReviewListPayload,
  SessionEntry as RootSessionEntry,
  SessionDetailPayload,
  TimelineSessionRef,
} from "@peasant-labs/schema";
import { isProjectHash, newProjectHash, validateProjectHash } from "@peasant-labs/schema";
import type { Case, Corpus } from "@peasant-labs/schema/testcase";
import type { TimelineFixtureCase, TimelineFixtureCorpus } from "@peasant-labs/schema/fixtures/timeline";

const brandedProjectHash: ProjectHash = newProjectHash("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2");
const brandedProjectHashIsAString: string = brandedProjectHash;
declare let unknownProjectHash: unknown;
validateProjectHash(unknownProjectHash);
const validatedProjectHash: ProjectHash = unknownProjectHash;
declare const guardedProjectHash: unknown;
if (isProjectHash(guardedProjectHash)) {
  const narrowedProjectHash: ProjectHash = guardedProjectHash;
  void narrowedProjectHash;
}
// @ts-expect-error A plain string must pass through newProjectHash or validateProjectHash.
const plainStringIsNotAProjectHash: ProjectHash = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";

declare const detail: SessionDetailPayload;
declare const entry: RootSessionEntry;
declare const resolution: ProjectResolutionPayload;
declare const timelineSession: TimelineSessionRef;
declare const timelineFixtures: TimelineFixtureCorpus;
declare const timelineFixture: TimelineFixtureCase;
const testCase: Case<RootSessionEntry, SessionDetailPayload> = {
  name: "compile-only",
  input: entry,
  expected: detail,
  classification: "must-pass",
  provenance: { source: "manual", ref: "public type smoke" },
  mutation: { description: "imports public package types" },
};
const corpus: Corpus<RootSessionEntry, SessionDetailPayload> = { cases: [testCase] };
declare const reviewList: ReviewListPayload;
declare const publishRequest: PublishRequest;
void reviewList;
void publishRequest;
void brandedProjectHashIsAString;
void validatedProjectHash;
void plainStringIsNotAProjectHash;
void corpus;
void resolution;
void timelineSession;
void timelineFixtures;
void timelineFixture;
