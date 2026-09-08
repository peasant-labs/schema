// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
import type { SessionDetailPayload, TurnDetail } from "./contract/zod.gen.js";

/**
 * KnownContentCapability is the closed inventory of content-capability tokens known to this
 * pinned schema release. It is intentionally distinct from the OPEN discovery
 * wire alias ContentCapability (an arbitrary string): SchemaVersionResponse's
 * contentCapabilities parses unknown future tokens so an older client keeps
 * working, while these constants give strongly typed access to known tokens.
 * isContentCapability narrows an arbitrary string to a KnownContentCapability, so consumers can filter a
 * discovered list down to the tokens they understand without stringly typing.
 */
export const KnownContentCapability = Object.freeze({
  DetailedUsageV1: "detailed_usage_v1",
  NativeMetadataV1: "native_metadata_v1",
  ObservedModelV1: "observed_model_v1",
  SessionGraphProvenanceV1: "session_graph_provenance_v1",
} as const);
export type KnownContentCapability = (typeof KnownContentCapability)[keyof typeof KnownContentCapability];
export const AllContentCapabilities = Object.freeze([KnownContentCapability.DetailedUsageV1, KnownContentCapability.NativeMetadataV1, KnownContentCapability.ObservedModelV1, KnownContentCapability.SessionGraphProvenanceV1]) as readonly KnownContentCapability[];
export function isContentCapability(value: unknown): value is KnownContentCapability {
  return typeof value === "string" && (AllContentCapabilities as readonly string[]).includes(value);
}

export function knownContentCapabilities(values: readonly string[]): KnownContentCapability[] {
  return [...new Set(values.filter(isContentCapability))].sort();
}

export function missingContentCapabilities(advertised: readonly string[], required: readonly KnownContentCapability[]): KnownContentCapability[] {
  const present = new Set(knownContentCapabilities(advertised));
  return [...new Set(required)].filter((token) => !present.has(token)).sort();
}

export function validateContentCapabilityAdvertisements(values: readonly string[]): void {
  for (let index = 0; index < values.length; index++) {
    const token = values[index];
    if (token === undefined || !isContentCapability(token)) throw new TypeError(`content capability producer validation failed at @peasant-labs/schema validateContentCapabilityAdvertisements: unknown token ${JSON.stringify(token)}; callers cannot advertise unimplemented preservation; emit only AllContentCapabilities after deployment support is proven`);
    const previous = values[index - 1];
    if (previous !== undefined && previous === token) throw new TypeError("content capability producer validation failed at @peasant-labs/schema validateContentCapabilityAdvertisements: token is duplicated; callers cannot emit a canonical advertisement; deduplicate and sort the known inventory");
    if (previous !== undefined && previous > token) throw new TypeError("content capability producer validation failed at @peasant-labs/schema validateContentCapabilityAdvertisements: tokens are not in canonical lexicographic order; callers cannot emit a canonical advertisement; sort the unique known inventory");
  }
}

function visitTurns(turns: readonly TurnDetail[] | null | undefined, found: Set<KnownContentCapability>): void {
  for (const turn of turns ?? []) {
    if ((turn.observedModel ?? "") !== "") found.add(KnownContentCapability.ObservedModelV1);
    if (turn.usage != null || turn.toolCalls?.some((tool) => tool.usage != null)) found.add(KnownContentCapability.DetailedUsageV1);
    if (turn.provenance != null || turn.toolCalls?.some((tool) => tool.callProvenance != null || tool.resultProvenance != null)) found.add(KnownContentCapability.SessionGraphProvenanceV1);
  }
}

export function requiredContentCapabilities(detail: SessionDetailPayload): KnownContentCapability[] {
  const found = new Set<KnownContentCapability>();
  if (detail.inputSubmissionCount !== undefined || detail.rootSessionId != null || (detail.purpose ?? "") !== "" || (detail.relationships?.length ?? 0) > 0 || (detail.earlierHistory?.length ?? 0) > 0) found.add(KnownContentCapability.SessionGraphProvenanceV1);
  if ((detail.nativeMetadata?.length ?? 0) > 0) found.add(KnownContentCapability.NativeMetadataV1);
  visitTurns(detail.turns, found);
  for (const section of detail.earlierHistory ?? []) {
    if ((section.nativeMetadata?.length ?? 0) > 0) found.add(KnownContentCapability.NativeMetadataV1);
    visitTurns(section.turns, found);
  }
  return [...found].sort();
}
