import type {
  PublishRequest,
  ProjectResolutionPayload,
  ReviewListPayload,
  SessionDetailPayload,
  TimelineSessionRef,
} from "@peasant-labs/schema";
import type { ReviewListPayload as LocalReviewListPayload } from "@peasant-labs/schema/local-api";
import type { PublishRequest as VillagePublishRequest } from "@peasant-labs/schema/village-api";
import type { SessionEntry } from "@peasant-labs/schema/types";
import type { Case, Corpus } from "@peasant-labs/schema/testcase";
import type { TimelineFixtureCorpus } from "@peasant-labs/schema/fixtures/timeline";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
// Root types mirror Go JSON presence. Operation-specific projections retain
// their API-policy requiredness and therefore need not be structurally exact.
const reviewProjectionIsDistinct: Same<ReviewListPayload, LocalReviewListPayload> = false;
const publishProjectionIsDistinct: Same<PublishRequest, VillagePublishRequest> = false;

declare const detail: SessionDetailPayload;
declare const entry: SessionEntry;
declare const resolution: ProjectResolutionPayload;
declare const timelineSession: TimelineSessionRef;
declare const timelineFixtures: TimelineFixtureCorpus;
const testCase: Case<SessionEntry, SessionDetailPayload> = {
  name: "compile-only",
  input: entry,
  expected: detail,
  classification: "must-pass",
  provenance: { source: "manual", ref: "public type smoke" },
  mutation: { description: "imports public package types" },
};
const corpus: Corpus<SessionEntry, SessionDetailPayload> = { cases: [testCase] };
void reviewProjectionIsDistinct;
void publishProjectionIsDistinct;
void corpus;
void resolution;
void timelineSession;
void timelineFixtures;
