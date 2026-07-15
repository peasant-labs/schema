import type {
  PublishRequest,
  ProjectResolutionPayload,
  ReviewListPayload,
  SessionEntry as RootSessionEntry,
  SessionDetailPayload,
  TimelineSessionRef,
} from "@peasant-labs/schema";
import type { operations as LocalOperations } from "@peasant-labs/schema/local-api";
import type { operations as VillageOperations } from "@peasant-labs/schema/village-api";
import type { SessionEntry as CompatibilitySessionEntry } from "@peasant-labs/schema/types";
import type { Case, Corpus } from "@peasant-labs/schema/testcase";
import type { TimelineFixtureCase, TimelineFixtureCorpus } from "@peasant-labs/schema/fixtures/timeline";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type LocalReviewListPayload = LocalOperations["listReviewChanges"]["responses"][200]["content"]["application/json"];
type VillagePublishRequest = NonNullable<VillageOperations["publishTranscript"]["requestBody"]>["content"]["application/json"];
const reviewOperationUsesCanonicalRoot: Same<ReviewListPayload, LocalReviewListPayload> = true;
const compatibilitySubpathUsesCanonicalRoot: Same<RootSessionEntry, CompatibilitySessionEntry> = true;
declare const publishOperationRequest: VillagePublishRequest;

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
void reviewOperationUsesCanonicalRoot;
void compatibilitySubpathUsesCanonicalRoot;
void publishOperationRequest;
void corpus;
void resolution;
void timelineSession;
void timelineFixtures;
void timelineFixture;
