import {
  zLocalRemoteContributionAttempt,
  type LocalRemoteContributionAttempt,
  type VillageShareEventActor,
  type VillageShareStatus,
} from "./internal/generated/contract/zod.gen.js";

export type VillageContributionAttemptReconciliation = "inserted" | "unchanged" | "updated" | "superseded";

// Capacity is not a moderator UUID or proof of a manual decision.
export function villageContributionAttemptActor(status: VillageShareStatus): VillageShareEventActor {
  switch (status) {
    case "pending": return "";
    case "approved": case "rejected": return "moderator";
    case "retracted": return "owner";
    case "revoked": return "collective";
    default: throw conflict("status is outside the share-attempt menu");
  }
}

// Select the verified storage scope before calling, and reconcile under the
// write transaction against latest evidence. Different ordinals are distinct
// rows; pass undefined saved evidence for a previously unseen identity.
export function reconcileVillageContributionAttempt(
  saved: LocalRemoteContributionAttempt | undefined,
  incoming: LocalRemoteContributionAttempt,
): { attempt: LocalRemoteContributionAttempt; disposition: VillageContributionAttemptReconciliation } {
  const next = zLocalRemoteContributionAttempt.parse(incoming);
  if (saved === undefined) return { attempt: next, disposition: "inserted" };
  const old = zLocalRemoteContributionAttempt.parse(saved);
  if (old.collectiveId !== next.collectiveId || old.transcriptId !== next.transcriptId || old.attempt.event_num !== next.attempt.event_num) {
    throw conflict("the incoming attempt has a different collective, transcript, or ordinal identity");
  }
  if (!equalTime(old.attempt.recorded_at, next.attempt.recorded_at)) throw conflict("recorded_at changed for an existing attempt identity");
  if (old.attempt.status === next.attempt.status && old.attempt.decided_by_actor === next.attempt.decided_by_actor && equalTime(old.attempt.decided_at, next.attempt.decided_at)) {
    return { attempt: old, disposition: "unchanged" };
  }
  if (old.attempt.status === "pending") {
    if (next.attempt.decided_at === null) throw conflict("an in-place pending decision requires a newly recorded decided_at");
    next.firstSeenAt = old.firstSeenAt;
    return { attempt: next, disposition: "updated" };
  }
  if (next.attempt.status === "pending") return { attempt: old, disposition: "superseded" };
  throw conflict("terminal status or decision evidence changed for an existing attempt identity");
}

// Do not use Date.parse: millisecond rounding would erase terminal timestamp
// conflicts present in the Go nanosecond contract. Zod has validated UTC form.
function equalTime(left: string | null, right: string | null): boolean {
  const canonical = (value: string | null): string | null => value === null ? null : value.replace(/\.([0-9]*?)0+Z$/, (_match, fraction: string) => fraction === "" ? "Z" : `.${fraction}Z`);
  return canonical(left) === canonical(right);
}

function conflict(reason: string): TypeError {
  return new TypeError(`Collective contribution attempt reconciliation failed at @peasant-labs/schema during history validation: ${reason}; retain prior evidence without installing this observation and fetch validated owner history before retrying.`);
}
