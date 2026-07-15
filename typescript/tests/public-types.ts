import type {
  PublishRequest,
  ReviewListPayload,
  SessionDetailPayload,
} from "@peasant-labs/schema";
import type { ReviewListPayload as LocalReviewListPayload } from "@peasant-labs/schema/local-api";
import type { PublishRequest as VillagePublishRequest } from "@peasant-labs/schema/village-api";
import type { SessionEntry } from "@peasant-labs/schema/types";
import type { Case, Corpus } from "@peasant-labs/schema/testcase";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
const reviewMatches: Same<ReviewListPayload, LocalReviewListPayload> = true;
const publishMatches: Same<PublishRequest, VillagePublishRequest> = true;

declare const detail: SessionDetailPayload;
declare const entry: SessionEntry;
const testCase: Case<SessionEntry, SessionDetailPayload> = {
  name: "compile-only",
  input: entry,
  expected: detail,
  classification: "must-pass",
  provenance: { source: "manual", ref: "public type smoke" },
  mutation: { description: "imports public package types" },
};
const corpus: Corpus<SessionEntry, SessionDetailPayload> = { cases: [testCase] };
void reviewMatches;
void publishMatches;
void corpus;
